package auth

// `profiles.json` is where a person names a platform this binary does not know.
//
// **It came back deliberately.** A `config.json` beside the credentials was
// deleted in TASK-036, and two of its three fields deserved it: `defaultProfile`
// and `defaultWorkspaces` were per-machine preferences, invisible on the machine
// that had them and absent on every other, so the same command in the same
// checkout did different things for two people. The third field was a map of
// named platforms, and deleting that one was wrong.
//
// It is wrong because of the test the deletion was made under: *a value belongs
// in a config file only when nothing on disk implies it and the platform cannot
// be asked.* An on-prem installation's issuer, client id and API pass that
// cleanly - nothing in a repository implies them, and the platform cannot be
// asked because **which platform is the question**. It is a bootstrap value, not
// a preference.
//
// The argument made at the time was that the three environment overrides cover
// it. They do for one platform. They do not for an on-prem customer running
// their own dev and their own prod, because you cannot hold two sets of three
// variables at once and switch between them by name. That is what a profile is.
//
// **Nothing that resembles a preference came back with it.** There is no
// recorded "current profile"; that is still --profile, ASGARD_PROFILE, and
// `default`. The file holds named platforms and nothing else.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ProfilesFileName is the file, beside the credential store.
//
// It is not called config.json. That name is retired, refused on sight by
// checkRetiredSettings, and meant "a bag of preferences" - which is the thing
// this deliberately is not.
const ProfilesFileName = "profiles.json"

// ProfilesFile is the whole of it.
type ProfilesFile struct {
	// Version is 1.
	Version int `json:"version"`
	// Profiles are platforms by name. A field left empty falls back to the
	// hosted platform, per field.
	Profiles map[string]Profile `json:"profiles"`
}

// Names lists the profiles the file holds, sorted, with `default` first when it
// is there - it is the one somebody is most likely to want to see.
func (f *ProfilesFile) Names() []string {
	if f == nil {
		return nil
	}
	out := make([]string, 0, len(f.Profiles))
	for name := range f.Profiles {
		if name != DefaultProfileName {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	if _, ok := f.Profiles[DefaultProfileName]; ok {
		out = append([]string{DefaultProfileName}, out...)
	}
	return out
}

// ProfilesPath is the file's location, for a command that has to name it.
func ProfilesPath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ProfilesFileName), nil
}

// LoadProfiles reads the file. A missing one is not an error and not a state to
// fix: it means every profile is the hosted platform, which is what a customer
// of the hosted platform should never have to configure.
func LoadProfiles() (*ProfilesFile, error) {
	path, err := ProfilesPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &ProfilesFile{Version: 1, Profiles: map[string]Profile{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var f ProfilesFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON: %w\n"+
			"It holds named platforms and nothing secret, so it is safe to open and fix by hand,\n"+
			"or to delete - every profile then resolves to the hosted platform.", path, err)
	}
	if f.Version != 0 && f.Version != 1 {
		return nil, fmt.Errorf("%s declares version %d, which this build does not understand; upgrade asgard-cli", path, f.Version)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{}
	}
	// The name is the map key. Carrying it in the value too would be a second
	// copy that a hand edit can make disagree with the first.
	for name, p := range f.Profiles {
		p.Name = name
		f.Profiles[name] = p
	}
	return &f, nil
}

// SaveProfiles writes the file, creating Home when it is not there.
//
// **It is only ever called by `profile set` and `profile remove`.** Nothing
// writes this file as a side effect of doing something else: a file that
// appears because you ran an unrelated command is a file nobody remembers
// agreeing to, and this directory just finished being emptied of those.
func SaveProfiles(f *ProfilesFile) error {
	path, err := ProfilesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	if f.Version == 0 {
		f.Version = 1
	}
	// The name lives in the key; drop it from the value before writing.
	out := ProfilesFile{Version: f.Version, Profiles: map[string]Profile{}}
	for name, p := range f.Profiles {
		p.Name = ""
		out.Profiles[name] = p
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("encode the profiles: %w", err)
	}
	// 0644, not 0600: this holds no secret, and somebody setting up an on-prem
	// installation should be able to hand the file to a colleague.
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
