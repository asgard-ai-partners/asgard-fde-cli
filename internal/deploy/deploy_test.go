package deploy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, project, body string) {
	t.Helper()
	path := Path(root, project)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestLoadAndTarget(t *testing.T) {
	root := t.TempDir()
	write(t, root, "erp", `environments:
  dev:
    namespace: asgard-acme-erp-dev
    values: chart/values-dev.yaml
`)

	file, err := Load(root, "erp")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	target, err := file.Target("erp", "dev")
	if err != nil {
		t.Fatalf("Target: %v", err)
	}
	if target.Namespace != "asgard-acme-erp-dev" || target.Values != "chart/values-dev.yaml" {
		t.Errorf("target = %+v", target)
	}
}

// TestUndeclaredEnvIsADecisionNotAnError pins the distinction: not declaring an
// environment is a deliberate choice not to deploy there, so it needs its own
// error type rather than reading as a broken file.
func TestUndeclaredEnvIsADecisionNotAnError(t *testing.T) {
	root := t.TempDir()
	write(t, root, "erp", "environments:\n  dev:\n    namespace: ns\n    values: v.yaml\n")

	file, err := Load(root, "erp")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	_, err = file.Target("erp", "prod")
	var notDeclared *ErrNotDeclared
	if !errors.As(err, &notDeclared) {
		t.Fatalf("error is %T, want *ErrNotDeclared", err)
	}
	if !strings.Contains(err.Error(), "deliberately does not deploy") {
		t.Errorf("error should say it is a decision: %q", err)
	}
}

func TestLoadReportsWhatIsWrong(t *testing.T) {
	root := t.TempDir()

	if _, err := Load(root, "missing"); err == nil {
		t.Error("a project with no deploy.yaml should be reported")
	} else if !strings.Contains(err.Error(), "single source of truth") {
		t.Errorf("error = %q, want it to say why the file matters", err)
	}

	write(t, root, "broken", "environments: [this is a list\n")
	if _, err := Load(root, "broken"); err == nil {
		t.Error("invalid YAML should be reported")
	}
}

func TestTargetRequiresBothFields(t *testing.T) {
	root := t.TempDir()
	write(t, root, "erp", "environments:\n  dev:\n    namespace: ns\n")

	file, err := Load(root, "erp")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// A namespace with no values file renders nothing useful, and CI would fail
	// on it later; saying so here is cheaper.
	if _, err := file.Target("erp", "dev"); err == nil || !strings.Contains(err.Error(), "values file") {
		t.Errorf("error = %v, want it to name the missing values file", err)
	}
}
