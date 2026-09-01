package cli

import (
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

// initialised makes a directory with a valid config, ready for project add.
func initialised(t *testing.T, name string) string {
	t.Helper()
	dir := dirNamed(t, name)
	if _, err := runInit(t, dir, "--workspace-id", "ws_1"); err != nil {
		t.Fatalf("init: %v", err)
	}
	return dir
}

func TestProjectAdd(t *testing.T) {
	dir := initialised(t, "acme")

	out, err := runCLI(t, dir, "project", "add", "internal")
	if err != nil {
		t.Fatalf("project add: %v", err)
	}

	cfg := loadConfig(t, dir)
	p, ok := cfg.Project("internal")
	if !ok {
		t.Fatalf("projects = %+v, want one named internal", cfg.Projects)
	}
	if p.Name != "internal" {
		t.Errorf("name = %q, want it to default to the slug", p.Name)
	}
	if len(p.Environments) != 1 || p.Environments[0] != config.EnvDev {
		t.Errorf("environments = %v, want dev only", p.Environments)
	}
	if !strings.Contains(out, "asgard-acme-internal-dev") {
		t.Errorf("output %q should show the namespace this project deploys to", out)
	}
}

func TestProjectAddWarnsAboutOrdering(t *testing.T) {
	dir := initialised(t, "acme")

	out, err := runCLI(t, dir, "project", "add", "internal")
	if err != nil {
		t.Fatalf("project add: %v", err)
	}

	// Both of these fail in CD rather than here, so the command has to say them.
	if !strings.Contains(out, "app-secret") {
		t.Errorf("output %q should mention the tf-asgard prerequisite", out)
	}
	if !strings.Contains(out, "syncer-name") {
		t.Errorf("output %q should mention the CD zero-Syncer trap", out)
	}
}

func TestProjectAddWithEnvsAndName(t *testing.T) {
	dir := initialised(t, "acme")

	out, err := runCLI(t, dir, "project", "add", "website",
		"--env", "dev", "--env", "prod", "--name", "official site")
	if err != nil {
		t.Fatalf("project add: %v", err)
	}

	p, _ := loadConfig(t, dir).Project("website")
	if p.Name != "official site" {
		t.Errorf("name = %q, want the value passed to --name", p.Name)
	}
	if len(p.Environments) != 2 {
		t.Fatalf("environments = %v, want dev and prod", p.Environments)
	}
	for _, ns := range []string{"asgard-acme-website-dev", "asgard-acme-website-prod"} {
		if !strings.Contains(out, ns) {
			t.Errorf("output %q should list %q", out, ns)
		}
	}
}

func TestProjectAddRejectsDuplicate(t *testing.T) {
	dir := initialised(t, "acme")

	if _, err := runCLI(t, dir, "project", "add", "internal"); err != nil {
		t.Fatalf("first add: %v", err)
	}

	_, err := runCLI(t, dir, "project", "add", "internal", "--env", "prod")
	if err == nil {
		t.Fatal("want failure for a duplicate slug, got success")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q should say the project already exists", err)
	}

	// The failed add must not have changed the existing entry.
	p, _ := loadConfig(t, dir).Project("internal")
	if len(p.Environments) != 1 || p.Environments[0] != config.EnvDev {
		t.Errorf("environments = %v, the existing project was modified", p.Environments)
	}
}

func TestProjectAddRejectsBadSlug(t *testing.T) {
	dir := initialised(t, "acme")

	_, err := runCLI(t, dir, "project", "add", "Not_A_Slug")
	if err == nil {
		t.Fatal("want failure for an invalid slug, got success")
	}
	if len(loadConfig(t, dir).Projects) != 0 {
		t.Error("an invalid project must not be written")
	}
}

func TestProjectAddRejectsUnknownEnv(t *testing.T) {
	dir := initialised(t, "acme")

	_, err := runCLI(t, dir, "project", "add", "internal", "--env", "staging")
	if err == nil {
		t.Fatal("want failure for an unknown environment, got success")
	}
	if !strings.Contains(err.Error(), "staging") {
		t.Errorf("error %q should name the rejected environment", err)
	}
	if !strings.Contains(err.Error(), "dev") {
		t.Errorf("error %q should say what is accepted", err)
	}
	if len(loadConfig(t, dir).Projects) != 0 {
		t.Error("an invalid project must not be written")
	}
}

func TestProjectAddRejectsOverlongNamespace(t *testing.T) {
	dir := initialised(t, "acme")

	_, err := runCLI(t, dir, "project", "add", strings.Repeat("b", 60))
	if err == nil {
		t.Fatal("want failure for a namespace over the limit, got success")
	}
	if !strings.Contains(err.Error(), "63") {
		t.Errorf("error %q should mention the namespace length limit", err)
	}
}

func TestProjectAddWithoutConfig(t *testing.T) {
	dir := dirNamed(t, "acme")

	_, err := runCLI(t, dir, "project", "add", "internal")
	if err == nil {
		t.Fatal("want failure when there is no config, got success")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error %q should tell the user to run init first", err)
	}
}

func TestProjectAddFindsConfigInParentDir(t *testing.T) {
	root := initialised(t, "acme")
	nested := dirNamedIn(t, root, "sub")

	if _, err := runCLI(t, nested, "project", "add", "internal"); err != nil {
		t.Fatalf("project add from a subdirectory: %v", err)
	}

	if _, ok := loadConfig(t, root).Project("internal"); !ok {
		t.Error("the project should have been written to the config above")
	}
}

func TestProjectRequiresSubcommand(t *testing.T) {
	dir := initialised(t, "acme")

	out, err := runCLI(t, dir, "project")
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if !strings.Contains(out, "add") {
		t.Errorf("output %q should list the add subcommand", out)
	}
}
