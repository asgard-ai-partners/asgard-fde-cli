package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// valid returns a config that passes Validate, for tests that want to break one
// thing at a time.
func valid() *Config {
	return &Config{
		Workspace: Workspace{ID: "ws_1", Slug: "unitech-e", Name: "unitech-e"},
		Projects: []Project{
			{Slug: "internal", Name: "internal", Environments: []Env{EnvDev, EnvProd}},
		},
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	want := valid()

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
	if len(got.Projects) != 1 || got.Projects[0].Slug != "internal" {
		t.Errorf("projects = %+v, want one project %q", got.Projects, "internal")
	}
	if len(got.Projects[0].Environments) != 2 {
		t.Errorf("environments = %v, want dev and prod", got.Projects[0].Environments)
	}
}

func TestSaveWritesReadableJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	cfg := &Config{
		Workspace: Workspace{ID: "ws_1", Slug: "demo", Name: "demo"},
		Projects:  []Project{{Slug: "web", Name: "web", Environments: []Env{EnvDev}}},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// People edit this file by hand, so the indentation and the trailing newline
	// are part of the spec.
	want := `{
  "workspace": {
    "id": "ws_1",
    "slug": "demo",
    "name": "demo"
  },
  "projects": [
    {
      "slug": "web",
      "name": "web",
      "environments": [
        "dev"
      ]
    }
  ]
}
`
	if got := string(data); got != want {
		t.Errorf("file contents =\n%s\nwant\n%s", got, want)
	}
}

func TestSaveOmitsNothingWhenProjectsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	cfg := &Config{Workspace: Workspace{ID: "ws_1", Slug: "demo", Name: "demo"}}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Projects) != 0 {
		t.Errorf("projects = %+v, want none", got.Projects)
	}
}

func TestSaveOverwritesAndLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	first := valid()
	first.Workspace.ID = "ws_old"
	if err := Save(path, first); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	second := valid()
	second.Workspace.ID = "ws_new"
	if err := Save(path, second); err != nil {
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
	if err := Save(path, valid()); err != nil {
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

func TestRepoNameAndNamespace(t *testing.T) {
	cfg := valid()

	if got, want := cfg.RepoName(), "unitech-e-asgard-kube"; got != want {
		t.Errorf("RepoName() = %q, want %q", got, want)
	}
	if got, want := cfg.Namespace("internal", EnvDev), "asgard-unitech-e-internal-dev"; got != want {
		t.Errorf("Namespace() = %q, want %q", got, want)
	}
	if got, want := cfg.Namespace("website", EnvProd), "asgard-unitech-e-website-prod"; got != want {
		t.Errorf("Namespace() = %q, want %q", got, want)
	}
}

func TestProjectLookup(t *testing.T) {
	cfg := valid()

	if p, ok := cfg.Project("internal"); !ok || p.Slug != "internal" {
		t.Errorf("Project(%q) = %+v, %v; want the internal project", "internal", p, ok)
	}
	if _, ok := cfg.Project("nope"); ok {
		t.Error("Project(\"nope\") reported a project that does not exist")
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		slug string
		ok   bool
	}{
		{"unitech-e", true},
		{"internal", true},
		{"a", true},
		{"web2", true},
		{"", false},
		{"UpperCase", false},
		{"trailing-", false},
		{"-leading", false},
		{"under_score", false},
		{"has space", false},
		{"dot.dot", false},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			err := ValidateSlug("slug", tt.slug)
			if (err == nil) != tt.ok {
				t.Errorf("ValidateSlug(%q) error = %v, want ok=%v", tt.slug, err, tt.ok)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string // substring of the expected problem, empty means valid
	}{
		{"valid", func(*Config) {}, ""},
		{"missing id", func(c *Config) { c.Workspace.ID = "" }, "workspace.id"},
		{"missing slug", func(c *Config) { c.Workspace.Slug = "" }, "workspace.slug"},
		{"bad slug", func(c *Config) { c.Workspace.Slug = "Bad_Slug" }, "workspace.slug"},
		{"missing name", func(c *Config) { c.Workspace.Name = "" }, "workspace.name"},
		{"bad project slug", func(c *Config) { c.Projects[0].Slug = "Bad" }, "projects[0].slug"},
		{"missing project name", func(c *Config) { c.Projects[0].Name = "" }, "projects[0].name"},
		{"no environments", func(c *Config) { c.Projects[0].Environments = nil }, "at least one"},
		{"unknown environment", func(c *Config) { c.Projects[0].Environments = []Env{"staging"} }, `"staging"`},
		{"duplicate environment", func(c *Config) { c.Projects[0].Environments = []Env{EnvDev, EnvDev} }, "twice"},
		{"duplicate project", func(c *Config) {
			c.Projects = append(c.Projects, Project{Slug: "internal", Name: "again", Environments: []Env{EnvDev}})
		}, "duplicates"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid()
			tt.mutate(cfg)

			err := cfg.Validate()
			if tt.want == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want a problem mentioning %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Validate() = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestValidateReportsEveryProblemAtOnce(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want problems")
	}
	// An empty config is wrong in three ways; reporting one at a time would make
	// the user fix it in three rounds.
	for _, want := range []string{"workspace.id", "workspace.slug", "workspace.name"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Validate() = %q, missing %q", err, want)
		}
	}
}

func TestValidateRejectsOverlongNamespace(t *testing.T) {
	cfg := valid()
	// asgard- + slug + - + project + -prod must stay within 63 characters.
	cfg.Workspace.Slug = strings.Repeat("a", 40)
	cfg.Projects[0].Slug = strings.Repeat("b", 20)

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want a namespace length problem")
	}
	if !strings.Contains(err.Error(), "63") {
		t.Errorf("Validate() = %q, want it to mention the 63 character limit", err)
	}

	ns := cfg.Namespace(cfg.Projects[0].Slug, EnvProd)
	if len(ns) <= MaxNamespaceLength {
		t.Fatalf("test is not exercising the limit: namespace is %d characters", len(ns))
	}
}

func TestValidateAcceptsNamespaceAtTheLimit(t *testing.T) {
	cfg := valid()
	cfg.Projects[0].Environments = []Env{EnvProd}
	// asgard- (7) + slug + - (1) + project + -prod (5) == 63
	cfg.Workspace.Slug = strings.Repeat("a", 25)
	cfg.Projects[0].Slug = strings.Repeat("b", 25)

	if got := len(cfg.Namespace(cfg.Projects[0].Slug, EnvProd)); got != MaxNamespaceLength {
		t.Fatalf("test setup produced a %d character namespace, want exactly %d", got, MaxNamespaceLength)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() = %v, want a namespace of exactly the limit to be accepted", err)
	}
}
