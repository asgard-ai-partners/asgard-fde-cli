package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

// runCLI runs the command line in dir and returns the combined output.
func runCLI(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	t.Chdir(dir)

	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)

	err := root.Execute()
	return out.String(), err
}

// runInit runs init in dir.
func runInit(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	return runCLI(t, dir, append([]string{"init"}, args...)...)
}

// dirNamed makes a directory with a predictable name, since t.TempDir() embeds
// the test name and init derives the slug from the directory.
func dirNamed(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	return dir
}

// dirNamedIn makes a named directory under an existing parent.
func dirNamedIn(t *testing.T, parent, name string) string {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	return dir
}

// writeConfigJSON writes raw JSON, to build broken configs init would never produce.
func writeConfigJSON(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// loadConfig reads the config a command just wrote.
func loadConfig(t *testing.T, dir string) *config.Config {
	t.Helper()
	cfg, err := config.Load(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

func TestInitCreatesConfig(t *testing.T) {
	dir := dirNamed(t, "demo")

	out, err := runInit(t, dir, "--workspace-id", "ws_test",
		"--workspace-slug", "unitech-e", "--workspace-name", "unitech-e")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	cfg := loadConfig(t, dir)
	want := config.Workspace{ID: "ws_test", Slug: "unitech-e", Name: "unitech-e"}
	if cfg.Workspace != want {
		t.Errorf("workspace = %+v, want %+v", cfg.Workspace, want)
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("projects = %+v, want none", cfg.Projects)
	}

	// The derived repository name is what the FDE needs next, so it has to be
	// in the output rather than only in the config.
	if !strings.Contains(out, "unitech-e-asgard-kube") {
		t.Errorf("output %q should show the derived repository name", out)
	}
	if !strings.Contains(out, "project add") {
		t.Errorf("output %q should point at the next command", out)
	}
}

func TestInitDerivesSlugFromDirectory(t *testing.T) {
	// Running inside the repository the onboarding produces should recover the
	// slug, not repeat the suffix.
	dir := dirNamed(t, "unitech-e-asgard-kube")

	if _, err := runInit(t, dir, "--workspace-id", "ws_1"); err != nil {
		t.Fatalf("init: %v", err)
	}

	cfg := loadConfig(t, dir)
	if cfg.Workspace.Slug != "unitech-e" {
		t.Errorf("workspace.slug = %q, want %q", cfg.Workspace.Slug, "unitech-e")
	}
	if cfg.RepoName() != "unitech-e-asgard-kube" {
		t.Errorf("RepoName() = %q, want %q", cfg.RepoName(), "unitech-e-asgard-kube")
	}
}

func TestInitDefaultsNameToSlug(t *testing.T) {
	dir := dirNamed(t, "acme")

	if _, err := runInit(t, dir, "--workspace-id", "ws_1"); err != nil {
		t.Fatalf("init: %v", err)
	}

	cfg := loadConfig(t, dir)
	if cfg.Workspace.Slug != "acme" || cfg.Workspace.Name != "acme" {
		t.Errorf("workspace = %+v, want slug and name both %q", cfg.Workspace, "acme")
	}
}

func TestInitRejectsSlugThatCannotBeANamespace(t *testing.T) {
	dir := dirNamed(t, "Not_A_Slug")

	_, err := runInit(t, dir, "--workspace-id", "ws_1")
	if err == nil {
		t.Fatal("want failure for a directory name that is not a valid slug, got success")
	}
	if !strings.Contains(err.Error(), "workspace.slug") {
		t.Errorf("error %q should name the offending field", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, config.FileName)); !os.IsNotExist(statErr) {
		t.Error("an invalid config must not be written")
	}
}

func TestInitProjectShortcut(t *testing.T) {
	dir := dirNamed(t, "acme")

	out, err := runInit(t, dir, "--workspace-id", "ws_1",
		"--project", "internal", "--project", "website")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	cfg := loadConfig(t, dir)
	if len(cfg.Projects) != 2 {
		t.Fatalf("projects = %+v, want two", cfg.Projects)
	}
	for _, p := range cfg.Projects {
		if len(p.Environments) != 1 || p.Environments[0] != config.EnvDev {
			t.Errorf("project %q environments = %v, want dev only", p.Slug, p.Environments)
		}
	}
	if !strings.Contains(out, "asgard-acme-internal-dev") {
		t.Errorf("output %q should show the derived namespace", out)
	}
}

func TestInitRequiresWorkspaceID(t *testing.T) {
	dir := dirNamed(t, "acme")

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
	if _, err := os.Stat(filepath.Join(dir, config.FileName)); !os.IsNotExist(err) {
		t.Errorf("%s should not have been created", config.FileName)
	}
}

func TestInitPrintsExistingConfig(t *testing.T) {
	dir := dirNamed(t, "acme")

	if _, err := runInit(t, dir, "--workspace-id", "ws_first"); err != nil {
		t.Fatalf("first init: %v", err)
	}

	out, err := runInit(t, dir)
	if err != nil {
		t.Fatalf("rerun init: %v", err)
	}
	if !strings.Contains(out, "already exists") {
		t.Errorf("output %q should report that the config already exists", out)
	}
	if !strings.Contains(out, "ws_first") {
		t.Errorf("output %q should print the current workspace", out)
	}
}

func TestInitLeavesExistingConfigUntouched(t *testing.T) {
	dir := dirNamed(t, "acme")

	if _, err := runInit(t, dir, "--workspace-id", "ws_first"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runInit(t, dir, "--workspace-id", "ws_second"); err != nil {
		t.Fatalf("rerun init: %v", err)
	}

	if got := loadConfig(t, dir).Workspace.ID; got != "ws_first" {
		t.Errorf("workspace.id = %q, the existing config was modified", got)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	dir := dirNamed(t, "acme")

	if _, err := runInit(t, dir, "--workspace-id", "ws_first"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runInit(t, dir, "--force", "--workspace-id", "ws_second", "--workspace-slug", "other"); err != nil {
		t.Fatalf("--force init: %v", err)
	}

	cfg := loadConfig(t, dir)
	if cfg.Workspace.ID != "ws_second" || cfg.Workspace.Slug != "other" {
		t.Errorf("workspace = %+v, want the forced values", cfg.Workspace)
	}
}

func TestInitForceStillRequiresWorkspaceID(t *testing.T) {
	dir := dirNamed(t, "acme")

	if _, err := runInit(t, dir, "--workspace-id", "ws_first"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runInit(t, dir, "--force"); err == nil {
		t.Fatal("want failure for --force without an id, got success")
	}

	if got := loadConfig(t, dir).Workspace.ID; got != "ws_first" {
		t.Errorf("workspace.id = %q, a failed --force must not touch the existing config", got)
	}
}

func TestInitRejectsIncompleteConfig(t *testing.T) {
	dir := dirNamed(t, "acme")
	writeConfigJSON(t, dir, `{"workspace":{}}`)

	_, err := runInit(t, dir, "--workspace-id", "ws_new")
	if err == nil {
		t.Fatal("want failure for an incomplete existing config, got success")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error %q should point at --force", err)
	}
	if !strings.Contains(err.Error(), "workspace.id") {
		t.Errorf("error %q should list the problems it found", err)
	}
}

func TestInitDoesNotClobberBrokenConfig(t *testing.T) {
	dir := dirNamed(t, "acme")
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
