package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
)

func cfg() *config.Config {
	return &config.Config{
		Workspace: config.Workspace{ID: "ws_1", Slug: "acme", Name: "Acme"},
		Projects: []config.Project{
			{Slug: "api", Name: "api", Environments: []config.Env{config.EnvDev, config.EnvProd}},
		},
	}
}

// scaffolded builds the repo the CLI produces, which is the baseline every
// other case in this file breaks in one way.
func scaffolded(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := scaffold.Write(dir, cfg(), false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	return dir
}

// TestScaffoldedRepoPasses is the invariant that matters most: what the CLI
// writes must satisfy the checks the CLI enforces. If these two drift, every
// new customer repo starts red.
func TestScaffoldedRepoPasses(t *testing.T) {
	report, err := Run(scaffolded(t))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.OK() {
		for _, f := range report.Errors() {
			t.Errorf("unexpected error: %s", f.Message)
		}
	}
	if len(report.Scope) != 1 || report.Scope[0] != "api" {
		t.Errorf("scope = %v, want [api]", report.Scope)
	}
}

func TestEmptyRepoPasses(t *testing.T) {
	dir := t.TempDir()
	empty := &config.Config{Workspace: config.Workspace{ID: "ws_1", Slug: "acme", Name: "Acme"}}
	if _, err := scaffold.Write(dir, empty, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	// A repo with no projects yet is a normal state - it is where every
	// onboarding sits after the first scaffold - so it must not be red.
	report, err := Run(dir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.OK() {
		for _, f := range report.Errors() {
			t.Errorf("a freshly scaffolded repo should pass, got: %s", f.Message)
		}
	}
}

func TestFindsBrokenRepos(t *testing.T) {
	tests := []struct {
		name   string
		break_ func(t *testing.T, dir string)
		want   string
	}{
		{"deploy.yaml removed", func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "projects", "api", "deploy.yaml"))
		}, "deploy.yaml"},

		{"env that does not exist", func(t *testing.T, dir string) {
			write(t, filepath.Join(dir, "projects", "api", "deploy.yaml"),
				"environments:\n  staging:\n    namespace: ns\n    values: chart/values-dev.yaml\n")
		}, `env "staging"`},

		{"values file missing", func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "projects", "api", "chart", "values-prod.yaml"))
		}, "does not exist"},

		{"shared values missing", func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "common", "values-prod.yaml"))
		}, "common/values-prod.yaml"},

		{"README does not list the project", func(t *testing.T, dir string) {
			write(t, filepath.Join(dir, "README.md"), "# acme-asgard-kube\n\nnothing here\n")
		}, "missing \"api\""},

		{"README lists a project that is gone", func(t *testing.T, dir string) {
			os.RemoveAll(filepath.Join(dir, "projects", "api"))
		}, "no directory"},

		{"runtime skill without frontmatter", func(t *testing.T, dir string) {
			skill := filepath.Join(dir, "common", "skills", "domain")
			os.MkdirAll(skill, 0o755)
			write(t, filepath.Join(skill, "SKILL.md"), "# domain\n\nno frontmatter\n")
		}, "no frontmatter"},

		{"runtime skill name does not match its directory", func(t *testing.T, dir string) {
			skill := filepath.Join(dir, "common", "skills", "domain")
			os.MkdirAll(skill, 0o755)
			write(t, filepath.Join(skill, "SKILL.md"), "---\nname: other\ndescription: d\n---\n\n# x\n")
		}, "does not match its directory"},

		{"requirements index removed", func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "requirements", "tasks", "_index.md"))
		}, "requirements/tasks/_index.md"},

		{"docs template removed", func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "docs", "decisions", "_decision-template.md"))
		}, "_decision-template.md"},

		{"spec module not indexed", func(t *testing.T, dir string) {
			write(t, filepath.Join(dir, "docs", "spec", "acme-asgard", "architecture.md"), "# arch\n")
		}, "does not index architecture.md"},

		{"decision record misnamed", func(t *testing.T, dir string) {
			write(t, filepath.Join(dir, "docs", "decisions", "why-we-did-it.md"), "# why\n")
		}, "YYYY-MM-DD"},

		{"dead link inside docs", func(t *testing.T, dir string) {
			write(t, filepath.Join(dir, "docs", "README.md"), "# docs\n\nsee [gone](nowhere.md)\n")
		}, "does not resolve"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := scaffolded(t)
			tt.break_(t, dir)

			report, err := Run(dir)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if report.OK() {
				t.Fatalf("expected an error mentioning %q, got a clean report", tt.want)
			}

			var messages []string
			for _, f := range report.Errors() {
				messages = append(messages, f.Message)
			}
			joined := strings.Join(messages, "\n")
			if !strings.Contains(joined, tt.want) {
				t.Errorf("errors do not mention %q:\n%s", tt.want, joined)
			}
		})
	}
}

func TestScopeLimitsProjectChecks(t *testing.T) {
	dir := t.TempDir()
	two := cfg()
	two.Projects = append(two.Projects, config.Project{
		Slug: "web", Name: "web", Environments: []config.Env{config.EnvDev},
	})
	if _, err := scaffold.Write(dir, two, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	os.Remove(filepath.Join(dir, "projects", "web", "deploy.yaml"))

	// Naming only the healthy project skips the broken one's checks.
	report, err := Run(dir, "api")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.OK() {
		t.Errorf("scoping to api should not report web's problem: %v", report.Errors())
	}

	// Without a scope, it is found.
	report, err = Run(dir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.OK() {
		t.Error("the whole-repo run should have found web's missing deploy.yaml")
	}
}

func TestUnknownProjectIsAnError(t *testing.T) {
	report, err := Run(scaffolded(t), "nope")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.OK() {
		t.Fatal("naming a project that does not exist should fail")
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
