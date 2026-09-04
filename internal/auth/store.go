package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// ErrNotSignedIn reports that a profile has no usable credential. Every command
// that needs the platform turns it into the one instruction that fixes it, so
// the message here stays short.
var ErrNotSignedIn = errors.New("not signed in")

// Credential is one profile's session.
//
// The refresh token is the long-lived half - 30 days against Casdoor's
// 24-hour access token - so this file is the thing worth protecting, and it is
// written 0600 and never inside a customer repository.
type Credential struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time `json:"expiresAt"`

	// Issuer and ClientID record which application issued this, so that
	// repointing a profile at another Casdoor invalidates the session it holds
	// instead of sending the old token to the new host and reporting whatever
	// that host says about it.
	Issuer   string `json:"issuer"`
	ClientID string `json:"clientId"`

	// Subject, Email and Name are what userinfo said at sign-in. They are held
	// so `whoami` can answer without a round trip, and refreshed whenever one
	// happens anyway.
	Subject string `json:"subject,omitempty"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
}

// expiryGrace is how long before the stated expiry a token is treated as
// already expired. A request that leaves with four minutes left can still
// arrive after the token has, and the platform reports that as a plain 401 with
// nothing to distinguish it from a revoked session.
const expiryGrace = 5 * time.Minute

// Expired reports whether the access token is past using.
func (c Credential) Expired() bool {
	return c.AccessToken == "" || !time.Now().Add(expiryGrace).Before(c.ExpiresAt)
}

// CanRefresh reports whether a new access token can be obtained without a
// browser.
func (c Credential) CanRefresh() bool { return c.RefreshToken != "" }

// matches reports whether the credential was issued by the profile's current
// application.
func (c Credential) matches(p Profile) bool {
	return c.Issuer == p.Issuer && c.ClientID == p.ClientID
}

// store is the credentials file: profile name to credential.
//
// One file for every profile, rather than one per profile, because signing out
// of everything has to be a single operation somebody can verify - `logout
// --all` writing one file is checkable, and deleting a directory of files it
// discovered is not.
type store struct {
	Credentials map[string]Credential `json:"credentials"`
}

// storePath is the credentials file inside Home.
func storePath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "credentials.json"), nil
}

func loadStore() (*store, string, error) {
	path, err := storePath()
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &store{Credentials: map[string]Credential{}}, path, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}
	var s store
	if err := json.Unmarshal(data, &s); err != nil {
		// A corrupt credentials file must not wedge the CLI: the fix is to sign
		// in again, and saying so beats an unexplained parse error.
		return nil, "", fmt.Errorf("parse %s: %w; delete it and run `asgard-cli login` again", path, err)
	}
	if s.Credentials == nil {
		s.Credentials = map[string]Credential{}
	}
	return &s, path, nil
}

func saveStore(s *store, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}
	return writeFilePrivate(path, append(data, '\n'))
}

// LoadCredential reads the stored session for a profile.
//
// A credential issued by a different application than the profile now names is
// reported as absent rather than returned, because the only thing that could be
// done with it is to send it somewhere it will not work.
func LoadCredential(p Profile) (Credential, error) {
	s, _, err := loadStore()
	if err != nil {
		return Credential{}, err
	}
	cred, ok := s.Credentials[p.Name]
	if !ok || cred.AccessToken == "" {
		return Credential{}, ErrNotSignedIn
	}
	if !cred.matches(p) {
		return Credential{}, ErrNotSignedIn
	}
	return cred, nil
}

// SaveCredential records a profile's session, leaving every other profile's
// alone.
func SaveCredential(p Profile, cred Credential) error {
	s, path, err := loadStore()
	if err != nil {
		return err
	}
	cred.Issuer = p.Issuer
	cred.ClientID = p.ClientID
	s.Credentials[p.Name] = cred
	return saveStore(s, path)
}

// DeleteCredential forgets one profile's session. Forgetting one that is not
// there is not an error: `logout` run twice should report the same thing.
func DeleteCredential(name string) (bool, error) {
	s, path, err := loadStore()
	if err != nil {
		return false, err
	}
	if _, ok := s.Credentials[name]; !ok {
		return false, nil
	}
	delete(s.Credentials, name)
	return true, saveStore(s, path)
}

// DeleteAllCredentials forgets every session and reports how many there were.
func DeleteAllCredentials() (int, error) {
	s, path, err := loadStore()
	if err != nil {
		return 0, err
	}
	n := len(s.Credentials)
	if n == 0 {
		return 0, nil
	}
	s.Credentials = map[string]Credential{}
	return n, saveStore(s, path)
}

// SignedInProfiles lists the profiles that hold a credential, sorted.
func SignedInProfiles() ([]string, error) {
	s, _, err := loadStore()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(s.Credentials))
	for name := range s.Credentials {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// writeFilePrivate writes data 0600 through a temporary file in the same
// directory, so an interrupted write cannot leave a half-written credential
// where a whole one was.
//
// The temporary file is created 0600 from the start rather than chmodded after:
// between an 0644 create and the chmod there is a window in which another user
// on the machine can read a refresh token, and the window is exactly as long as
// the write.
func writeFilePrivate(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("create a temporary file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0o600); err != nil && runtime.GOOS != "windows" {
		tmp.Close()
		return fmt.Errorf("restrict %s: %w", tmpName, err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
