package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	want := &Config{Workspace: Workspace{ID: "ws_abc123", Name: "demo"}}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Workspace != want.Workspace {
		t.Errorf("workspace = %+v, want %+v", got.Workspace, want.Workspace)
	}
}

func TestSaveWritesReadableJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	cfg := &Config{Workspace: Workspace{ID: "ws_abc123", Name: "demo"}}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(data)

	// People edit this file by hand, so the indentation and the trailing newline
	// are part of the spec.
	want := "{\n  \"workspace\": {\n    \"id\": \"ws_abc123\",\n    \"name\": \"demo\"\n  }\n}\n"
	if got != want {
		t.Errorf("file contents =\n%q\nwant\n%q", got, want)
	}
}

func TestSaveOverwritesAndLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	if err := Save(path, &Config{Workspace: Workspace{ID: "ws_old", Name: "old"}}); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	if err := Save(path, &Config{Workspace: Workspace{ID: "ws_new", Name: "new"}}); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Workspace.ID != "ws_new" {
		t.Errorf("workspace.id = %q, want %q", got.Workspace.ID, "ws_new")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("leftover files %v, want only %s", names, FileName)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), FileName))
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	if err := os.WriteFile(path, []byte("{ not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("want a parse error, got success")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not mention path %q", err, path)
	}
}

func TestFindWalksUp(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, FileName)
	if err := Save(path, &Config{Workspace: Workspace{ID: "ws_1", Name: "n"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := Find(nested)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	// macOS puts TempDir behind symlinks such as /var -> /private/var.
	wantResolved, _ := filepath.EvalSymlinks(path)
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != wantResolved {
		t.Errorf("Find = %q, want %q", gotResolved, wantResolved)
	}
}

func TestFindNotFound(t *testing.T) {
	_, err := Find(t.TempDir())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"complete", Config{Workspace{ID: "ws_1", Name: "n"}}, false},
		{"missing id", Config{Workspace{Name: "n"}}, true},
		{"missing name", Config{Workspace{ID: "ws_1"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
