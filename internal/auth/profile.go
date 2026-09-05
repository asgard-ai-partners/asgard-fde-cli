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
	"fmt"
	"os"
	"path/filepath"
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
//
// It is the whole list. A settings file used to be able to add profiles and
// override a built-in by name; that went with the file, and the three field
// overrides below do the same job for the case it was for - pointing at a
// platform that has moved, or at a local stack - without anything on disk that
// a later release has to keep understanding.
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

// Home is the directory this CLI keeps its own state in: **the credential
// store, and nothing else.**
//
// It held a settings file too, and that file is gone. Every field it carried
// was a preference that an environment variable or a flag already expressed,
// and each one was a thing this binary had to keep understanding across
// upgrades - a breaking change waiting on a file nobody remembers writing. A
// credential is the one thing that genuinely has to live here: it is a secret,
// it is per-person rather than per-repository, and it cannot be re-derived.
//
// See asgard-odin-pm docs/decisions/2026-09-05-asgard-cli-config-surface.md.
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

// retiredSettingsName is the file this package used to keep beside the
// credentials, and no longer reads.
const retiredSettingsName = "config.json"

// retiredSettings names the leftover settings file, when one is there.
func retiredSettings() (string, bool) {
	home, err := Home()
	if err != nil {
		return "", false
	}
	path := filepath.Join(home, retiredSettingsName)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return "", false
	}
	return path, true
}

// ErrRetiredSettings reports a settings file this build does not read.
//
// **It is an error and not a warning, and the reason is which way it fails.**
// The retired `defaultProfile` was most often `dev`; with the file ignored, the
// default becomes `prod`, which is a customer's platform. A warning printed
// after the fact is a warning printed after the command has already run there.
// So the first command that would resolve a profile refuses, and says what to
// type instead.
//
// It is reached only from ResolveProfile, so the commands that answer with no
// network and no login - `wiki`, `find`, `brief`, `guide`, `usecase`, `size` -
// are unaffected, as they must be.
type ErrRetiredSettings struct {
	Path string
	Keys []string
}

func (e *ErrRetiredSettings) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is a settings file this version of asgard-cli does not read.\n\n", e.Path)
	for _, k := range e.Keys {
		switch k {
		case "defaultProfile":
			fmt.Fprintf(&b, "  defaultProfile     now %s, per command or exported once:\n", EnvProfile)
			fmt.Fprintf(&b, "                       export %s=dev\n", EnvProfile)
			fmt.Fprintf(&b, "                       asgard-cli <command> --profile dev\n")
		case "defaultWorkspaces", "workspaces":
			fmt.Fprintf(&b, "  %-18s now --workspace or %s. A checkout's own workspace\n", k, EnvWorkspace)
			fmt.Fprintf(&b, "                       belongs in its committed .asgard-cli.yaml:\n")
			fmt.Fprintf(&b, "                       asgard-cli workspace use <id>\n")
		case "profiles":
			fmt.Fprintf(&b, "  profiles           now %s, %s and %s, applied on top of\n", EnvIssuer, EnvClientID, EnvAPI)
			fmt.Fprintf(&b, "                       dev or prod\n")
		}
	}
	// Only where a profile was recorded. That is the case where ignoring the
	// file quietly would move commands onto a customer's platform, and it is
	// the whole reason this refuses instead of warning.
	for _, k := range e.Keys {
		if k == "defaultProfile" {
			fmt.Fprintf(&b, "\n**Until one of those is said, %s applies** - which is a customer's\n", DefaultProfileName)
			fmt.Fprintf(&b, "platform. That is why this refuses rather than warns.\n")
			break
		}
	}
	fmt.Fprintf(&b, "\nThen delete it:\n\n    rm %s\n", e.Path)
	return b.String()
}

// checkRetiredSettings refuses to run while a settings file this build ignores
// is still on disk.
//
// A file with none of the retired keys - an empty object somebody left behind -
// is deleted quietly rather than reported: there is nothing to migrate, and an
// error somebody cannot act on is an error that teaches them to ignore errors.
func checkRetiredSettings() error {
	path, ok := retiredSettings()
	if !ok {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		// Unreadable and unread. Nothing can be migrated out of it.
		return os.Remove(path)
	}
	var keys []string
	for _, k := range []string{"defaultProfile", "defaultWorkspaces", "workspaces", "profiles"} {
		if v, ok := raw[k]; ok && len(v) > 0 && string(v) != "null" && string(v) != "{}" && string(v) != `""` {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return os.Remove(path)
	}
	return &ErrRetiredSettings{Path: path, Keys: keys}
}

// ResolveProfile returns the profile named by want, or the default when want is
// empty.
//
// Precedence, highest first: the --profile argument, ASGARD_PROFILE, and prod.
// **There is no fourth**, and there was: a `defaultProfile` recorded by `login
// --set-default` in a file under the user's config directory. It was one more
// thing an upgrade had to keep understanding, and one more way for two machines
// running the same command to do different things. The three field overrides
// are applied last, on whichever profile that produced.
func ResolveProfile(want string) (Profile, error) {
	if err := checkRetiredSettings(); err != nil {
		return Profile{}, err
	}

	name := want
	if name == "" {
		name = os.Getenv(EnvProfile)
	}
	if name == "" {
		name = DefaultProfileName
	}

	p, ok := builtinProfiles[name]
	if !ok {
		return Profile{}, &ErrUnknownProfile{Name: name, Known: BuiltinProfileNames()}
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
