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
// Nothing here is reachable from the half of the tool that answers a question.
// `asgard-cli init` writes the material and `asgard-cli guide` and
// `asgard-cli size` read it, all with no network, no repository and no login,
// and that has to stay true - the question they answer is asked in a meeting,
// before there is an engagement to log in to.
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
	// Name is the map key in profiles.json; it is not written into the value,
	// where a second copy could disagree with the key after a hand edit.
	Name string `json:"name,omitempty"`
	// Issuer is the Casdoor base URL, with no trailing slash.
	Issuer string `json:"issuer"`
	// ClientID is the Casdoor application's client id. It is not a secret:
	// a public client's id is disclosed to the browser on every sign-in.
	ClientID string `json:"clientId"`
	// PlatformAPI is the Asgard Platform API's base URL, with no trailing
	// slash. Paths this CLI calls are appended to it, so it includes no
	// version segment.
	//
	// **Named for the service, not for "the API".** This tool talks to one
	// Asgard service today and is expected to grow into others - the Control
	// Center API at cc-api.asgard-ai.com is the next one - and a field called
	// `API` would have taken the general word for the first arrival. Adding a
	// second endpoint then means either an awkward name for it or renaming
	// this one after somebody has already scripted against it.
	//
	// Issuer, ClientID and ASGARD_TOKEN are deliberately NOT split the same
	// way: every Asgard service authenticates against the same Casdoor and
	// accepts the same token, so those three genuinely are about all of them
	// and the general name is the correct one.
	PlatformAPI string `json:"platformApi"`
}

// DefaultProfileName is the profile used when nothing selects one.
//
// **It is a name, not an environment.** It used to be "prod", with "dev" beside
// it as the other built-in, and that spent two general words on two particular
// installations of ours. An on-prem customer runs their own dev and their own
// prod, and neither is either of those.
//
// A profile that does not exist, or exists and sets none of the three fields,
// resolves to the hosted platform below - so a customer who never opens the
// file reaches the right place, and everybody else names their own.
const DefaultProfileName = "default"

// The hosted platform. **These are a fallback, not a profile.**
//
// They are compiled in because a hostname somebody has to type is a hostname
// somebody gets wrong, and the failure surfaces as an authentication error
// rather than as a typo. They are the value of each field that a profile leaves
// empty - per field, so a profile that names only its own API keeps everything
// else - and they are what makes `default` work with no file at all.
//
// There is deliberately no compiled-in entry for our development platform. It
// is one installation among the ones this tool will meet, not a second kind of
// thing, and baking it in would put an internal endpoint in every customer's
// binary. Ours is written with `asgard-cli profile set`, like anybody else's.
const (
	hostedIssuer      = "https://iam.asgard-ai.com"
	hostedClientID    = "r21ntx0eb5igyokl3px4"
	hostedPlatformAPI = "https://platform-api.asgard-ai.com"
)

// Origin says where a resolved field came from, so `profile show` can print it.
//
// **The provenance is the feature.** A profile that sets one field and inherits
// the rest looks identical to a complete one until something fails at the far
// end, and the two mixes that matter are silent by construction - see
// mixedOrigin.
type Origin string

const (
	FromProfileFile Origin = "profiles.json"
	FromHosted      Origin = "the hosted platform (this profile does not set it)"
	FromEnv         Origin = "environment"

	// The three ways the profile *name* gets chosen, as opposed to where one
	// of its fields came from. They read as the thing a person would look at
	// to change the answer, because that is what somebody does with this line
	// when it says something they did not expect.
	FromFlag       Origin = "--profile"
	FromEnvProfile Origin = EnvProfile
	FromDefault    Origin = "the default: no --profile, no " + EnvProfile
)

// Resolved is a profile with a record of where each field came from.
type Resolved struct {
	Profile
	// Exists reports whether profiles.json actually holds this name.
	Exists bool
	// NameFrom is how this profile came to be the one in effect, which is a
	// different question from where its fields came from.
	//
	// **It is the half that decides whether somebody can be wrong about it.**
	// A name typed on the command line cannot be a surprise; one inherited from
	// the environment was decided once, possibly in another shell, and the
	// default is the hosted platform reached by configuring nothing.
	NameFrom     Origin
	IssuerFrom   Origin
	ClientIDFrom Origin
	APIFrom      Origin
}

// mixedOrigin reports a profile whose Platform API and identity provider come
// from different places, which is almost always a half-written profile.
//
// Both directions are wrong and neither announces itself:
//
//   - an explicit API with an inherited issuer sends you to sign in against the
//     hosted Casdoor and then presents that token to somebody else's server;
//   - an explicit issuer with an inherited API signs you in against your own
//     Casdoor and then sends that token to OUR hosted production API, which is
//     the direction worth being loud about.
//
// The client id is not part of this: a different application on the same issuer
// is an ordinary thing to want.
//
// It is a warning rather than a refusal because a local Platform API against a
// real Casdoor is a legitimate way to develop - it is just never the thing an
// on-prem installation wants.
func (r Resolved) mixedOrigin() bool {
	return r.APIFrom != r.IssuerFrom
}

// Warning returns what is worth saying about this resolution, or "".
func (r Resolved) Warning() string {
	if !r.mixedOrigin() {
		return ""
	}
	return fmt.Sprintf(
		"profile %q takes its Platform API from %s and its identity provider from %s.\n"+
			"An API and the Casdoor that issues tokens for it have to be the same installation:\n"+
			"a token from one is not accepted by the other, and the failure reads as a permission\n"+
			"problem rather than as a mismatch. Set both, or neither.\n\n"+
			"    asgard-cli profile set %s --platform-api <url> --issuer <url> --client-id <id>",
		r.Name, r.APIFrom, r.IssuerFrom, r.Name)
}

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
	// EnvPlatformAPI overrides the Platform API's base URL. It is named for
	// that service rather than for "the API" so that a second one - the
	// Control Center API is the expected next - can be added without renaming
	// this or giving that an awkward name. ASGARD_ISSUER, ASGARD_CLIENT_ID and
	// ASGARD_TOKEN stay general on purpose: every Asgard service uses the same
	// Casdoor and the same token, so they are about all of them.
	EnvPlatformAPI = "ASGARD_PLATFORM_API"
	// EnvRetiredAPI is what EnvPlatformAPI used to be called. It is read only
	// to refuse, never to apply - see checkRetiredEnv.
	EnvRetiredAPI = "ASGARD_API"
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
	var b strings.Builder
	fmt.Fprintf(&b, "no profile named %q.\n", e.Name)

	// The two names this binary used to have compiled in. Somebody who typed
	// one is not making a typo - they are running a habit from an older build,
	// and the answer is different for each.
	switch e.Name {
	case "prod":
		fmt.Fprintf(&b, "\n%q was a built-in name and is now just %q, which is the hosted platform\nwith nothing configured. Drop the flag.\n", "prod", DefaultProfileName)
	case "dev":
		fmt.Fprintf(&b, "\n%q was a built-in name. Our development platform is one installation among\nthe ones this tool meets, not a second kind of thing, so it is written like any\nother - see the internal setup notes for the three values:\n\n    asgard-cli profile set dev --issuer <url> --client-id <id> --platform-api <url>\n", "dev")
	}

	if len(e.Known) == 0 {
		fmt.Fprintf(&b, "\nNo profiles are configured, so every name resolves to the hosted platform\nexcept the one that says so:\n\n    asgard-cli --profile %s ...\n    asgard-cli profile set <name> --platform-api <url> --issuer <url> --client-id <id>\n", DefaultProfileName)
		return b.String()
	}
	fmt.Fprintf(&b, "\nConfigured: %s\n\n    asgard-cli profile list\n    asgard-cli profile set %s --platform-api <url> --issuer <url> --client-id <id>\n",
		strings.Join(e.Known, ", "), e.Name)
	return b.String()
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

// checkRetiredEnv refuses to run while a renamed environment variable is set.
//
// **Silently ignoring it is the one outcome to avoid.** ASGARD_API used to
// point this CLI at a platform other than dev or prod - a local stack, a
// preview environment - so a shell that still exports it is a shell whose
// owner believes they are talking to that. Ignore it and every command goes to
// the built-in prod URL instead, which is a customer's platform, and nothing
// says so.
//
// It refuses rather than quietly applying the old name because a fallback is a
// second spelling that has to keep working forever, and this rename exists
// precisely to stop one name meaning two things.
func checkRetiredEnv() error {
	old := os.Getenv(EnvRetiredAPI)
	if old == "" {
		return nil
	}
	if os.Getenv(EnvPlatformAPI) == old {
		// Both set to the same thing: somebody has already migrated and left
		// the old one behind. Nothing is ambiguous, so this is not worth
		// stopping for.
		return nil
	}
	return fmt.Errorf(
		"%s is set, and this version of asgard-cli does not read it.\n\n"+
			"    %s=%s\n\n"+
			"It is %s now. The name changed because this CLI is expected to talk to more than\n"+
			"one Asgard service - the Control Center API is the next one - and %s had taken\n"+
			"the general word for whichever arrived first.\n\n"+
			"    export %s=%s\n    unset %s\n\n"+
			"**This refuses rather than ignoring the old name**: ignored, every command would\n"+
			"go to the built-in %s URL instead of where you pointed it, and nothing would say so.\n\n"+
			"%s, %s and %s are unchanged - every Asgard service uses the same Casdoor and the\n"+
			"same token, so those are about all of them.",
		EnvRetiredAPI, EnvRetiredAPI, old, EnvPlatformAPI, EnvRetiredAPI,
		EnvPlatformAPI, old, EnvRetiredAPI, DefaultProfileName,
		EnvIssuer, EnvClientID, EnvToken)
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
// network and no login - `asgard-cli init`, `asgard-cli guide`,
// `asgard-cli size` - are unaffected, as they must be.
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
			fmt.Fprintf(&b, "  profiles           now %s, %s and %s, applied on top of\n", EnvIssuer, EnvClientID, EnvPlatformAPI)
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
func ResolveProfile(want string) (Profile, error) {
	r, err := ResolveWithOrigin(want)
	if err != nil {
		return Profile{}, err
	}
	return r.Profile, nil
}

// ResolveWithOrigin is ResolveProfile with a record of where each field came from.
//
// The order is: which profile (--profile, then ASGARD_PROFILE, then "default"),
// then each of its three fields (the file if it sets one, otherwise the hosted
// platform), then the three environment overrides on top.
//
// **The per-field fallback is the point.** A profile that names only its own
// Platform API is a complete, working profile for a local stack; one that names
// all three is an on-prem installation. Neither has to restate what it is not
// changing, and `profile show` says which is which.
//
// **A name with no entry in the file is not an error.** `default` is expected
// to have none - that is the hosted platform, reached with nothing configured -
// and any other name resolving to the hosted platform would be a confusing way
// to spell it, so that one is refused. See ErrUnknownProfile.
func ResolveWithOrigin(want string) (Resolved, error) {
	if err := checkRetiredSettings(); err != nil {
		return Resolved{}, err
	}
	if err := checkRetiredEnv(); err != nil {
		return Resolved{}, err
	}

	name, nameFrom := want, FromFlag
	if name == "" {
		name, nameFrom = os.Getenv(EnvProfile), FromEnvProfile
	}
	if name == "" {
		name, nameFrom = DefaultProfileName, FromDefault
	}

	file, err := LoadProfiles()
	if err != nil {
		return Resolved{}, err
	}
	stored, exists := file.Profiles[name]
	if !exists && name != DefaultProfileName {
		return Resolved{}, &ErrUnknownProfile{Name: name, Known: file.Names()}
	}

	r := Resolved{
		Profile:      Profile{Name: name},
		Exists:       exists,
		NameFrom:     nameFrom,
		IssuerFrom:   FromHosted,
		ClientIDFrom: FromHosted,
		APIFrom:      FromHosted,
	}
	r.Issuer, r.ClientID, r.PlatformAPI = hostedIssuer, hostedClientID, hostedPlatformAPI
	if stored.Issuer != "" {
		r.Issuer, r.IssuerFrom = stored.Issuer, FromProfileFile
	}
	if stored.ClientID != "" {
		r.ClientID, r.ClientIDFrom = stored.ClientID, FromProfileFile
	}
	if stored.PlatformAPI != "" {
		r.PlatformAPI, r.APIFrom = stored.PlatformAPI, FromProfileFile
	}

	if v := os.Getenv(EnvIssuer); v != "" {
		r.Issuer, r.IssuerFrom = v, FromEnv
	}
	if v := os.Getenv(EnvClientID); v != "" {
		r.ClientID, r.ClientIDFrom = v, FromEnv
	}
	if v := os.Getenv(EnvPlatformAPI); v != "" {
		r.PlatformAPI, r.APIFrom = v, FromEnv
	}

	r.Issuer = strings.TrimRight(r.Issuer, "/")
	r.PlatformAPI = strings.TrimRight(r.PlatformAPI, "/")

	if err := r.Profile.validate(); err != nil {
		return Resolved{}, err
	}
	return r, nil
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
	if p.PlatformAPI == "" {
		missing = append(missing, "platform api ("+EnvPlatformAPI+")")
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
