package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeSessionFile(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "session.json")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvSessionFile, p)
	t.Setenv(EnvSandboxMode, "true")
	t.Setenv(EnvHome, t.TempDir())
	t.Setenv(EnvToken, "")
	t.Setenv(EnvPlatformAPI, "")
	t.Setenv(EnvProfile, "")
}

// In the Workbench sandbox the member's identity is the session file the
// platform writes every turn: its token, its workspace, and the platform it
// names - which a fresh sandbox's profile could not know.
func TestResolveReadsTheSandboxSession(t *testing.T) {
	writeSessionFile(t, `{"access_token":"tok-1","workspace_id":"ws-9","platform_api":"https://platform-api.dev.example/",
		"user":{"id":"u-1","display_name":"Ada","email":"ada@example.test"},"channel_id":"wb-fde-x"}`)
	s, err := Resolve(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if s.Token != "tok-1" || s.Source != SourceSandbox || s.SandboxWorkspace != "ws-9" {
		t.Errorf("session %+v", s)
	}
	if s.Profile.PlatformAPI != "https://platform-api.dev.example" {
		t.Errorf("platform API %q, want the session file's, trailing slash trimmed", s.Profile.PlatformAPI)
	}
	if s.Subject != "u-1" || s.Name != "Ada" || s.Email != "ada@example.test" || s.ProfileFrom != FromSandbox {
		t.Errorf("identity %+v", s)
	}

	// ASGARD_PLATFORM_API still overrides, as everywhere.
	t.Setenv(EnvPlatformAPI, "http://localhost:8080")
	s, _ = Resolve(context.Background(), "")
	if s.Profile.PlatformAPI != "http://localhost:8080" {
		t.Errorf("platform API %q, want the environment's", s.Profile.PlatformAPI)
	}

	// ASGARD_TOKEN wins over the file, for trying something by hand.
	t.Setenv(EnvToken, "tok-env")
	s, _ = Resolve(context.Background(), "")
	if s.Token != "tok-env" || s.Source != SourceEnv {
		t.Errorf("with ASGARD_TOKEN: %+v", s)
	}
}

func TestResolveWithoutASandboxSession(t *testing.T) {
	writeSessionFile(t, `{"access_token":""}`)
	if _, err := Resolve(context.Background(), ""); !errors.Is(err, ErrNoSandboxSession) {
		t.Errorf("empty token: %v", err)
	}
	t.Setenv(EnvSessionFile, filepath.Join(t.TempDir(), "absent.json"))
	if _, err := Resolve(context.Background(), ""); !errors.Is(err, ErrNoSandboxSession) {
		t.Errorf("no file: %v", err)
	}
	writeSessionFile(t, `not json, but containing a secret`)
	_, err := Resolve(context.Background(), "")
	if err == nil || errors.Is(err, ErrNoSandboxSession) {
		t.Fatalf("malformed: %v", err)
	}
	if contains(err.Error(), "secret") {
		t.Errorf("the error quotes the file: %v", err)
	}
}

// Anything other than exactly "true" is not sandbox mode.
func TestSandboxModeIsExact(t *testing.T) {
	for v, want := range map[string]bool{"true": true, "": false, "1": false, "TRUE": false, "yes": false} {
		t.Setenv(EnvSandboxMode, v)
		if SandboxMode() != want {
			t.Errorf("%q: %v", v, !want)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
