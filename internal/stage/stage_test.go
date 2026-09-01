package stage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

func cfgWith(projects ...config.Project) *config.Config {
	return &config.Config{
		Workspace: config.Workspace{ID: "ws_1", Slug: "acme", Name: "Acme"},
		Projects:  projects,
	}
}

func project(slug string) config.Project {
	return config.Project{Slug: slug, Name: slug, Environments: []config.Env{config.EnvDev}}
}

// repo builds a directory that looks like an onboarding at a given point.
func repo(t *testing.T, scaffolded bool, crs map[string][]string) string {
	t.Helper()
	dir := t.TempDir()

	if scaffolded {
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# AGENTS.md\n"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	for slug, kinds := range crs {
		templates := filepath.Join(dir, "projects", slug, "chart", "app", "templates")
		if err := os.MkdirAll(templates, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		for i, kind := range kinds {
			body := "apiVersion: asgard-ai.com/v1alpha1\nkind: " + kind + "\nmetadata:\n  name: x\n"
			name := filepath.Join(templates, kind+string(rune('a'+i))+".yaml")
			if err := os.WriteFile(name, []byte(body), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
		}
	}
	return dir
}

func TestCurrentFollowsTheRepo(t *testing.T) {
	tests := []struct {
		name       string
		scaffolded bool
		cfg        *config.Config
		crs        map[string][]string
		want       Name
	}{
		{"nothing written", false, cfgWith(), nil, Scaffold},
		// No projects and nothing asked for is not "go and interview": there is
		// no requirement to interview about. This is the case that used to dump
		// the whole interview at someone who had just run init.
		{"scaffolded, nothing asked for", true, cfgWith(), nil, Idle},
		{"project, no connector", true, cfgWith(project("app")), nil, DataSources},
		{"connector, no read path", true, cfgWith(project("app")),
			map[string][]string{"app": {"DataConnector"}}, ReadPath},
		{"semantic layer, no entry", true, cfgWith(project("app")),
			map[string][]string{"app": {"DataConnector", "SemanticLayer"}}, EntryPoint},
		{"toolset counts as a read path", true, cfgWith(project("app")),
			map[string][]string{"app": {"DataConnector", "Toolset"}}, EntryPoint},
		// Every chart complete with nothing open is idle, not "run the gate":
		// there is nothing to gate that has not been gated.
		{"agent completes it, nothing open", true, cfgWith(project("app")),
			map[string][]string{"app": {"DataConnector", "SemanticLayer", "Agent"}}, Idle},
		{"bot provider completes it too", true, cfgWith(project("app")),
			map[string][]string{"app": {"DataConnector", "Toolset", "BotProvider"}}, Idle},
		{"the least finished project decides", true, cfgWith(project("app"), project("web")),
			map[string][]string{
				"app": {"DataConnector", "SemanticLayer", "Agent"},
				"web": {"DataConnector"},
			}, ReadPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := repo(t, tt.scaffolded, tt.crs)

			state, err := Inspect(dir, tt.cfg)
			if err != nil {
				t.Fatalf("Inspect: %v", err)
			}
			if got := Current(state).Name; got != tt.want {
				t.Errorf("Current() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEveryPromptRenders is the check that catches a template error, which would
// otherwise only surface when an FDE reached that stage.
func TestStagesAreOrderedAndUnique(t *testing.T) {
	seen := map[Name]bool{}
	for i, s := range Stages {
		if seen[s.Name] {
			t.Errorf("stage %q listed twice", s.Name)
		}
		seen[s.Name] = true
		if s.Number != i {
			t.Errorf("stage %q has number %d at index %d; Current() and the "+
				"printed \"stage N of M\" both depend on these agreeing", s.Name, s.Number, i)
		}
	}
}

func TestEveryPromptRenders(t *testing.T) {
	cfg := cfgWith(project("app"), project("web"))
	dir := repo(t, true, map[string][]string{"app": {"DataConnector"}})
	state, err := Inspect(dir, cfg)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}

	for _, s := range append(Stages, IdleStage) {
		t.Run(string(s.Name), func(t *testing.T) {
			out, err := s.Prompt(cfg, state)
			if err != nil {
				t.Fatalf("Prompt: %v", err)
			}
			if strings.TrimSpace(out) == "" {
				t.Error("rendered to nothing")
			}
			for _, marker := range []string{"<<", ">>", "<no value>"} {
				if strings.Contains(out, marker) {
					t.Errorf("output still contains %q:\n%s", marker, out)
				}
			}
		})
	}
}

// TestDecisionPromptsCarryTheReversal guards the point of the whole thing: the
// three decisions are worth printing because they were answered wrong once.
func TestDecisionPromptsCarryTheReversal(t *testing.T) {
	cfg := cfgWith(project("app"))
	dir := repo(t, true, nil)
	state, err := Inspect(dir, cfg)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}

	for _, name := range []Name{ReadPath, EntryPoint, Knowledge} {
		s, ok := Find(string(name))
		if !ok {
			t.Fatalf("stage %q not found", name)
		}
		out, err := s.Prompt(cfg, state)
		if err != nil {
			t.Fatalf("Prompt: %v", err)
		}
		if !strings.Contains(out, "wrong") {
			t.Errorf("%s does not say the decision was answered wrong before", name)
		}
	}
}

func TestPromptNamesTheProjectsThatNeedWork(t *testing.T) {
	cfg := cfgWith(project("app"), project("web"))
	dir := repo(t, true, map[string][]string{"app": {"DataConnector"}})
	state, err := Inspect(dir, cfg)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}

	s, _ := Find(string(DataSources))
	out, err := s.Prompt(cfg, state)
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if strings.Contains(out, "  - app\n") {
		t.Error("app already has a DataConnector and should not be listed")
	}
	if !strings.Contains(out, "  - web\n") {
		t.Errorf("web has no DataConnector and should be listed:\n%s", out)
	}
}

func TestNeedsTaskCoversTheChartStages(t *testing.T) {
	// SDD is required for a new SemanticLayer or DataConnector, for widening
	// what an agent may query, and for a write path. Those all land in these
	// stages, and nowhere else.
	wantTask := map[Name]bool{
		DataSources: true, ReadPath: true, EntryPoint: true, Knowledge: true, Enhance: true,
		Init: false, Scaffold: false, Projects: false, Verify: false, Deploy: false,
	}
	for _, s := range Stages {
		if got := s.NeedsTask(); got != wantTask[s.Name] {
			t.Errorf("%s.NeedsTask() = %v, want %v", s.Name, got, wantTask[s.Name])
		}
	}
}

// writeRequests puts a request registry in the repo, in the shape the scaffold
// writes and `asgard-cli request add` appends to.
func writeRequests(t *testing.T, dir string, rows ...string) {
	t.Helper()
	path := filepath.Join(dir, "requirements", "requests", "_index.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "# Requests Index\n\n| Request ID | Title | Priority | Status | Spec |\n|---|---|---|---|---|\n" +
		strings.Join(rows, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// TestCurrentFollowsTheOpenRequest is the reason the interview is not reached by
// "there are no projects": the interview belongs to a request, and a repo with
// no request is not waiting to be interviewed.
func TestCurrentFollowsTheOpenRequest(t *testing.T) {
	complete := map[string][]string{"app": {"DataConnector", "SemanticLayer", "Agent"}}

	tests := []struct {
		name string
		rows []string
		want Name
	}{
		{
			"a request with no project yet is the interview",
			[]string{"| [REQ-001](REQ-001-stock.md) | stock in chat | - | `draft` | TODO |"},
			Projects,
		},
		{
			"a request naming a project the config does not have is still the interview",
			[]string{"| [REQ-001](REQ-001-stock.md) | stock in chat | - | `draft` | projects/warehouse |"},
			Projects,
		},
		{
			"a request on an existing project with a complete chart is ready for the gate",
			[]string{"| [REQ-001](REQ-001-stock.md) | stock in chat | - | `ready` | projects/app |"},
			Verify,
		},
		{
			"a request already done leaves nothing in flight",
			[]string{"| [REQ-001](REQ-001-stock.md) | stock in chat | - | `done` | projects/app |"},
			Idle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := repo(t, true, complete)
			writeRequests(t, dir, tt.rows...)

			state, err := Inspect(dir, cfgWith(project("app")))
			if err != nil {
				t.Fatalf("Inspect: %v", err)
			}
			if got := Current(state).Name; got != tt.want {
				t.Errorf("Current() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAnOpenTaskAlsoCountsAsInFlight pins that the two records are read
// independently: a task opened without a request still means work is underway.
func TestAnOpenTaskAlsoCountsAsInFlight(t *testing.T) {
	dir := repo(t, true, map[string][]string{"app": {"DataConnector", "SemanticLayer", "Agent"}})
	index := filepath.Join(dir, "requirements", "tasks", "_index.md")
	if err := os.MkdirAll(filepath.Dir(index), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "# Task Index\n\n## Task Queue\n\n| Task ID | Title | Owner | Complexity | Status |\n" +
		"|---|---|---|---|---|\n| TASK-001 | a thing | - | M | `in-progress` |\n"
	if err := os.WriteFile(index, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	state, err := Inspect(dir, cfgWith(project("app")))
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if got := Current(state).Name; got != Verify {
		t.Errorf("Current() = %q, want %q", got, Verify)
	}
}
