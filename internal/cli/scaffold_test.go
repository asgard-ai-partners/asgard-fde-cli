package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldWritesTheSkeleton(t *testing.T) {
	dir := initialised(t, "acme")
	if _, err := runCLI(t, dir, "project", "add", "app"); err != nil {
		t.Fatalf("project add: %v", err)
	}

	out, err := runCLI(t, dir, "scaffold")
	if err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	for _, path := range []string{"AGENTS.md", "docs/README.md", "scripts/check_crd_fidelity.py", "projects/app/deploy.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Errorf("stat %s: %v", path, err)
		}
	}
	if !strings.Contains(out, "created") {
		t.Errorf("output %q should report what it wrote", out)
	}
	if !strings.Contains(out, "asgard-cli check") {
		t.Errorf("output %q should point at the verification step", out)
	}
}

func TestScaffoldRerunSkips(t *testing.T) {
	dir := initialised(t, "acme")
	if _, err := runCLI(t, dir, "scaffold"); err != nil {
		t.Fatalf("first scaffold: %v", err)
	}

	out, err := runCLI(t, dir, "scaffold")
	if err != nil {
		t.Fatalf("second scaffold: %v", err)
	}
	if !strings.Contains(out, "already present") {
		t.Errorf("output %q should report the files it left alone", out)
	}
	if strings.Contains(out, "  created  ") {
		t.Errorf("output %q should not claim to have created anything", out)
	}
}

// TestScaffoldWritesNextToTheConfig covers the case where an agent runs the
// command from wherever it happens to be: the skeleton belongs at the project
// root, not in the subdirectory.
func TestScaffoldWritesNextToTheConfig(t *testing.T) {
	root := initialised(t, "acme")
	nested := dirNamedIn(t, root, "sub")

	if _, err := runCLI(t, nested, "scaffold"); err != nil {
		t.Fatalf("scaffold from a subdirectory: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md should be at the project root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("nothing should have been written into the subdirectory")
	}
}

func TestScaffoldWithoutConfig(t *testing.T) {
	dir := dirNamed(t, "acme")

	_, err := runCLI(t, dir, "scaffold")
	if err == nil {
		t.Fatal("want failure when there is no config, got success")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error %q should tell the user to run init first", err)
	}
}
