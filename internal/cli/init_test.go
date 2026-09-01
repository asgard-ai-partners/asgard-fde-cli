package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

// runInit runs init in dir and returns the combined output.
func runInit(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	t.Chdir(dir)

	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"init"}, args...))

	err := root.Execute()
	return out.String(), err
}

func TestInitCreatesConfig(t *testing.T) {
	dir := t.TempDir()

	out, err := runInit(t, dir, "--workspace-id", "ws_test", "--workspace-name", "my-workspace")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Workspace.ID != "ws_test" {
		t.Errorf("workspace.id = %q, want %q", cfg.Workspace.ID, "ws_test")
	}
	if cfg.Workspace.Name != "my-workspace" {
		t.Errorf("workspace.name = %q, want %q", cfg.Workspace.Name, "my-workspace")
	}
	if !strings.Contains(out, "Created") || !strings.Contains(out, "ws_test") {
		t.Errorf("output %q should report creation and include the workspace id", out)
	}
}

func TestInitDefaultsNameToDir(t *testing.T) {
	// t.TempDir() embeds the test name, so nest one level for a predictable name.
	dir := filepath.Join(t.TempDir(), "my-project")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	if _, err := runInit(t, dir, "--workspace-id", "ws_1"); err != nil {
		t.Fatalf("init: %v", err)
	}

	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Workspace.Name != "my-project" {
		t.Errorf("workspace.name = %q, want %q", cfg.Workspace.Name, "my-project")
	}
}

func TestInitRequiresWorkspaceID(t *testing.T) {
	dir := t.TempDir()

	out, err := runInit(t, dir)
	if err == nil {
		t.Fatal("want failure without --workspace-id, got success")
	}
	if !strings.Contains(err.Error(), "--workspace-id") {
		t.Errorf("error %q should mention the missing --workspace-id", err)
	}
	if out != "" {
		t.Errorf("a failed run should print nothing to stdout, got %q", out)
	}

	// A failed run must not leave a half-written config behind.
	if _, err := os.Stat(filepath.Join(dir, config.FileName)); !os.IsNotExist(err) {
		t.Errorf("%s should not have been created", config.FileName)
	}
}

func TestInitPrintsExistingConfig(t *testing.T) {
	dir := t.TempDir()

	if _, err := runInit(t, dir, "--workspace-id", "ws_first", "--workspace-name", "first"); err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Already initialised: a rerun should succeed and only report the state.
	out, err := runInit(t, dir)
	if err != nil {
		t.Fatalf("rerun init: %v", err)
	}
	if !strings.Contains(out, "already exists") {
		t.Errorf("output %q should report that the config already exists", out)
	}
	if !strings.Contains(out, "ws_first") || !strings.Contains(out, "first") {
		t.Errorf("output %q should print the current workspace", out)
	}
}

func TestInitLeavesExistingConfigUntouched(t *testing.T) {
	dir := t.TempDir()

	if _, err := runInit(t, dir, "--workspace-id", "ws_first", "--workspace-name", "first"); err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Even with different values, nothing may change without --force.
	if _, err := runInit(t, dir, "--workspace-id", "ws_second", "--workspace-name", "second"); err != nil {
		t.Fatalf("rerun init: %v", err)
	}

	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Workspace.ID != "ws_first" {
		t.Errorf("workspace.id = %q, the existing config was modified", cfg.Workspace.ID)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()

	if _, err := runInit(t, dir, "--workspace-id", "ws_first", "--workspace-name", "first"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runInit(t, dir, "--force", "--workspace-id", "ws_second", "--workspace-name", "second"); err != nil {
		t.Fatalf("--force init: %v", err)
	}

	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Workspace.ID != "ws_second" {
		t.Errorf("workspace.id = %q, want %q", cfg.Workspace.ID, "ws_second")
	}
}

func TestInitForceStillRequiresWorkspaceID(t *testing.T) {
	dir := t.TempDir()

	if _, err := runInit(t, dir, "--workspace-id", "ws_first", "--workspace-name", "first"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runInit(t, dir, "--force"); err == nil {
		t.Fatal("want failure for --force without an id, got success")
	}

	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Workspace.ID != "ws_first" {
		t.Errorf("workspace.id = %q, a failed --force must not touch the existing config", cfg.Workspace.ID)
	}
}

func TestInitRejectsIncompleteConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfigJSON(t, dir, `{"workspace":{}}`)

	_, err := runInit(t, dir, "--workspace-id", "ws_new")
	if err == nil {
		t.Fatal("want failure for an incomplete existing config, got success")
	}
	if !strings.Contains(err.Error(), "lint") || !strings.Contains(err.Error(), "--force") {
		t.Errorf("error %q should point at lint and --force", err)
	}
}

func TestInitDoesNotClobberBrokenConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfigJSON(t, dir, "{ not json")

	if _, err := runInit(t, dir, "--workspace-id", "ws_new"); err == nil {
		t.Fatal("want failure for broken JSON, got success")
	}

	// The contents must be untouched, leaving the call to the user.
	data, err := os.ReadFile(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "{ not json" {
		t.Errorf("file was modified: %q", data)
	}
}
