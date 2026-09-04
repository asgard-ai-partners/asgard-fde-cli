// Package auth signs the CLI in to an Asgard platform and holds the result.
//
// The flow is OAuth 2.0 authorization code with PKCE against Casdoor, over a
// loopback redirect - RFC 8252's answer for a program running on a machine that
// has a browser. The alternative, the device authorization grant, is not
// available: Casdoor has no device_code endpoint, so a CLI that wanted one
// would have to be waited on rather than written.
//
// The client is public and ships no secret. Casdoor generates a client secret
// for every application whether or not one is wanted, and honours PKCE without
// it: a token request carrying a code_challenge and no client_secret is
// accepted. So the binary sends none, and there is nothing in a release anybody
// could lift a credential out of.
//
// Nothing here is reachable from the knowledge commands. `wiki`, `usecase`,
// `find`, `brief`, `size` and `guide` answer with no network, no repository and
// no login, and that has to stay true - the question they answer is asked in a
// meeting, before there is an engagement to log in to.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Profile is one Asgard environment: where its Casdoor lives, which
// application the CLI presents itself as, and where its platform API is.
//
// The two built-in profiles are the two environments Asgard runs. A profile is
// a set rather than three separate flags because they are only ever correct
// together: a dev token sent to the prod API is rejected, and the failure reads
// like a permission problem rather than a mismatch.
type Profile struct {
	// Name is what --profile takes.
	Name string `json:"name"`
	// Issuer is the Casdoor base URL, with no trailing slash.
	Issuer string `json:"issuer"`
	// ClientID is the Casdoor application's client id. It is not a secret:
	// a public client's id is disclosed to the browser on every sign-in.
	ClientID string `json:"clientId"`
	// API is the platform API base URL, with no trailing slash. Paths this CLI
	// calls are appended to it, so it includes no version segment.
	API string `json:"api"`
}

// DefaultProfileName is the profile used when nothing selects one.
//
// It is prod, because the people this binary is built for are customers running
// against the platform, and a tool whose default is the vendor's test
// environment is a tool that fails for everybody except the vendor. We are the
// exception, and the exception is the one that carries the flag: ASGARD_PROFILE
// or --profile switches to dev, and ASGARD_API points at anything else.
const DefaultProfileName = "prod"

// builtinProfiles are the two environments the platform runs. They are compiled
// in rather than configured because a hostname somebody has to type is a
// hostname somebody gets wrong, and the failure surfaces as an authentication
// error rather than as a typo.
var builtinProfiles = map[string]Profile{
	"dev": {
		Name:     "dev",
		Issuer:   "https://iam.dev.asgard-ai.com",
		ClientID: "q0bnivgjxhvtg2poqmle",
		API:      "https://platform-api.dev.asgard-ai.com",
	},
	"prod": {
		Name:     "prod",
		Issuer:   "https://iam.asgard-ai.com",
		ClientID: "r21ntx0eb5igyokl3px4",
		API:      "https://platform-api.asgard-ai.com",
	},
}

// BuiltinProfileNames lists the built-in profiles, the default first.
func BuiltinProfileNames() []string { return []string{"prod", "dev"} }

// Environment variables that override a profile's fields, one field each.
//
// They exist for developing against a platform running somewhere else - a local
// stack, a preview environment - and not for switching between dev and prod,
// which is what --profile is for. Each is applied on top of the resolved
// profile, so overriding the API alone keeps the rest of it.
const (
	EnvProfile  = "ASGARD_PROFILE"
	EnvIssuer   = "ASGARD_ISSUER"
	EnvClientID = "ASGARD_CLIENT_ID"
	EnvAPI      = "ASGARD_API"
	// EnvToken supplies an access token directly, for CI and for an agent
	// sandbox where no browser can be opened. It bypasses the store entirely:
	// nothing is read from disk and nothing is written to it.
	EnvToken = "ASGARD_TOKEN"
	// EnvWorkspace names the workspace to act in, ahead of anything recorded.
	EnvWorkspace = "ASGARD_WORKSPACE"
	// EnvHome relocates the whole configuration directory, which is what makes
	// this testable without touching a real one.
	EnvHome = "ASGARD_CLI_HOME"
)

// ErrUnknownProfile reports a --profile that names nothing.
type ErrUnknownProfile struct {
	Name  string
	Known []string
}

func (e *ErrUnknownProfile) Error() string {
	return fmt.Sprintf("unknown profile %q; one of %s", e.Name, strings.Join(e.Known, ", "))
}

// Settings is the CLI's own configuration, held outside any customer
// repository.
//
// It is deliberately not `.asgard-config.json`: that file describes a customer
// and is committed, and a default profile is a property of the person running
// the tool. Two FDEs sharing a repository do not share a session.
type Settings struct {
	// DefaultProfile is used when neither --profile nor ASGARD_PROFILE says.
	DefaultProfile string `json:"defaultProfile,omitempty"`
	// Profiles are additional environments, by name. A name that matches a
	// built-in replaces it, which is how a platform that has moved can be
	// pointed at without a new release.
	Profiles map[string]Profile `json:"profiles,omitempty"`
	// Workspaces records which workspace a repository's pipeline commands act
	// in: profile, then the repository's "owner/name". See workspace.go for
	// why it is held here rather than in the repository or on the platform.
	Workspaces map[string]RepoWorkspaces `json:"workspaces,omitempty"`
	// DefaultWorkspaces is the per-profile fallback for commands run outside a
	// repository.
	DefaultWorkspaces map[string]string `json:"defaultWorkspaces,omitempty"`
}

// Home is the directory this CLI keeps its own state in: the settings file and
// the credential store, and nothing else.
//
// os.UserConfigDir is used rather than a hand-rolled ~/.asgard so that the
// Windows build lands in %AppData% instead of a dot-directory in the user's
// profile, which is where a Windows user would never look for it.
func Home() (string, error) {
	if dir := os.Getenv(EnvHome); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate the user configuration directory: %w", err)
	}
	return filepath.Join(base, "asgard-cli"), nil
}

// settingsPath is the settings file inside Home.
func settingsPath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config.json"), nil
}

// LoadSettings reads the settings file. A missing file is not an error: it
// means every default applies, which is the state a first run is in.
func LoadSettings() (*Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Settings{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &s, nil
}

// SaveSettings writes the settings file, creating Home when it is not there.
func SaveSettings(s *Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return writeFilePrivate(path, append(data, '\n'))
}

// ResolveProfile returns the profile named by want, or the configured default
// when want is empty.
//
// Precedence, highest first: the --profile argument, ASGARD_PROFILE, the
// settings file's defaultProfile, and dev. The three field overrides are
// applied last, on whichever profile that produced.
func ResolveProfile(want string) (Profile, error) {
	s, err := LoadSettings()
	if err != nil {
		return Profile{}, err
	}
	return s.Resolve(want)
}

// Resolve is ResolveProfile against already-loaded settings.
func (s *Settings) Resolve(want string) (Profile, error) {
	name := want
	if name == "" {
		name = os.Getenv(EnvProfile)
	}
	if name == "" {
		name = s.DefaultProfile
	}
	if name == "" {
		name = DefaultProfileName
	}

	p, ok := s.Profiles[name]
	if !ok {
		p, ok = builtinProfiles[name]
	}
	if !ok {
		return Profile{}, &ErrUnknownProfile{Name: name, Known: s.Names()}
	}
	p.Name = name

	if v := os.Getenv(EnvIssuer); v != "" {
		p.Issuer = v
	}
	if v := os.Getenv(EnvClientID); v != "" {
		p.ClientID = v
	}
	if v := os.Getenv(EnvAPI); v != "" {
		p.API = v
	}

	p.Issuer = strings.TrimRight(p.Issuer, "/")
	p.API = strings.TrimRight(p.API, "/")

	if err := p.validate(); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// Names lists every profile that resolves, built-in and configured, sorted.
func (s *Settings) Names() []string {
	seen := map[string]bool{}
	for name := range builtinProfiles {
		seen[name] = true
	}
	for name := range s.Profiles {
		seen[name] = true
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// validate rejects a profile that cannot be used, naming the field. A profile
// assembled from environment overrides can be missing any of the three, and the
// failure without this is a request to an empty URL.
func (p Profile) validate() error {
	var missing []string
	if p.Issuer == "" {
		missing = append(missing, "issuer ("+EnvIssuer+")")
	}
	if p.ClientID == "" {
		missing = append(missing, "client id ("+EnvClientID+")")
	}
	if p.API == "" {
		missing = append(missing, "api ("+EnvAPI+")")
	}
	if len(missing) > 0 {
		return fmt.Errorf("profile %q has no %s", p.Name, strings.Join(missing, ", no "))
	}
	return nil
}

// AuthorizeURL is Casdoor's sign-in page. It is the front-end route, not an
// API path, which is why it carries no /api prefix.
func (p Profile) AuthorizeURL() string { return p.Issuer + "/login/oauth/authorize" }

// TokenURL is Casdoor's token endpoint, used for both the authorization code
// exchange and the refresh.
func (p Profile) TokenURL() string { return p.Issuer + "/api/login/oauth/access_token" }

// UserinfoURL is the OIDC userinfo endpoint. It is also what the platform's own
// IAM calls to verify a bearer token, so a token this endpoint accepts is a
// token the platform accepts.
func (p Profile) UserinfoURL() string { return p.Issuer + "/api/userinfo" }
