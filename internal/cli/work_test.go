package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

// scaffolded makes a repository the work commands can write into: a config, the
// full skeleton, and one project. The skeleton is the real one, because these
// commands write into files the scaffold produced and a fixture of my own would
// prove nothing about that.
func scaffolded(t *testing.T, projects ...string) string {
	t.Helper()

	dir := initialised(t, "acme")
	if _, err := runCLI(t, dir, "scaffold"); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	for _, slug := range projects {
		if _, err := runCLI(t, dir, "project", "add", slug); err != nil {
			t.Fatalf("project add %s: %v", slug, err)
		}
	}
	if len(projects) > 0 {
		if _, err := runCLI(t, dir, "scaffold"); err != nil {
			t.Fatalf("scaffold after project add: %v", err)
		}
	}
	return dir
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

// TestTheWholeRecordWalk runs the commands in the order an engagement does, and
// checks the repository afterwards. It is one test rather than several because
// the point is the sequence: each command reads what the one before it wrote.
func TestTheWholeRecordWalk(t *testing.T) {
	dir := scaffolded(t, "erp")
	date := today()

	if _, err := runCLI(t, dir, "request", "add", "staff want to ask about stock", "--slug", "stock"); err != nil {
		t.Fatalf("request add: %v", err)
	}
	if _, err := runCLI(t, dir, "question", "add", "which stock figure is authoritative",
		"--blocks", "REQ-001", "--ask", "the warehouse lead"); err != nil {
		t.Fatalf("question add: %v", err)
	}
	if _, err := runCLI(t, dir, "request", "target", "REQ-001", "erp"); err != nil {
		t.Fatalf("request target: %v", err)
	}
	if _, err := runCLI(t, dir, "request", "ready", "REQ-001"); err != nil {
		t.Fatalf("request ready: %v", err)
	}
	if _, err := runCLI(t, dir, "task", "add", "expose stock levels",
		"--slug", "stock-levels", "--request", "REQ-001", "--project", "erp", "--complexity", "M"); err != nil {
		t.Fatalf("task add: %v", err)
	}
	if _, err := runCLI(t, dir, "task", "start", "TASK-001"); err != nil {
		t.Fatalf("task start: %v", err)
	}

	requests, err := work.ReadRequests(dir)
	if err != nil {
		t.Fatalf("ReadRequests: %v", err)
	}
	if len(requests) != 1 || requests[0].Status != work.Ready || requests[0].Project != "erp" {
		t.Fatalf("requests = %+v, want one ready and targeting erp", requests)
	}

	tasks, err := work.ReadTasks(dir)
	if err != nil {
		t.Fatalf("ReadTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Status != work.InProgress {
		t.Fatalf("tasks = %+v, want one in progress", tasks)
	}

	questions, err := work.ReadQuestions(dir)
	if err != nil {
		t.Fatalf("ReadQuestions: %v", err)
	}
	if len(questions) != 1 || questions[0].Raised != date {
		t.Fatalf("questions = %+v, want one raised today", questions)
	}

	// The dates are stamped, not typed, which is the whole reason these are
	// commands rather than a note telling somebody to write the file.
	spec := readFile(t, dir, filepath.Join(work.TaskDir, "TASK-001-stock-levels.md"))
	if !strings.Contains(spec, "- Created: "+date) {
		t.Errorf("task spec has no created date:\n%s", spec)
	}
	if !strings.Contains(spec, date+" status `draft` -> `in-progress`") {
		t.Errorf("task spec did not log the transition:\n%s", spec)
	}

	// And the gate still passes: none of this writes something check rejects.
	if _, err := runCLI(t, dir, "check"); err != nil {
		t.Errorf("check after the walk: %v", err)
	}
}

// TestNextAsksForARequirementWhenNothingIsOpen is the defect this round fixes:
// a freshly scaffolded repo used to have the whole stage-2 interview dumped at
// it, when the honest answer is that nobody has asked for anything yet.
func TestNextAsksForARequirementWhenNothingIsOpen(t *testing.T) {
	dir := scaffolded(t)

	out, err := runCLI(t, dir, "next")
	if err != nil {
		t.Fatalf("next: %v", err)
	}

	if !strings.Contains(out, "nothing in flight") {
		t.Errorf("next should report the idle state:\n%s", out)
	}
	if !strings.Contains(out, "asgard-cli request add") {
		t.Errorf("next should ask for a requirement:\n%s", out)
	}
	if strings.Contains(out, "What to ask the customer") {
		t.Errorf("next should not print the interview before anything was asked for:\n%s", out)
	}

	// One request, and the same command becomes the interview - scoped to it.
	if _, err := runCLI(t, dir, "request", "add", "staff want stock in chat", "--slug", "stock"); err != nil {
		t.Fatalf("request add: %v", err)
	}
	out, err = runCLI(t, dir, "next")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if !strings.Contains(out, "What to ask the customer") {
		t.Errorf("next should now be the interview:\n%s", out)
	}
	if !strings.Contains(out, "REQ-001") {
		t.Errorf("the interview should name the request it is about:\n%s", out)
	}
}

func TestDecisionAddStampsTheDateAndLinksIt(t *testing.T) {
	dir := scaffolded(t)
	date := today()

	out, err := runCLI(t, dir, "decision", "add", "the work splits by audience",
		"--slug", "project-split", "--module", "architecture.md")
	if err != nil {
		t.Fatalf("decision add: %v", err)
	}

	name := date + "-project-split.md"
	if !strings.Contains(out, name) {
		t.Errorf("output %q should name the dated file it wrote", out)
	}

	record := readFile(t, dir, filepath.Join(work.DecisionDir, name))
	if !strings.Contains(record, "# the work splits by audience") {
		t.Errorf("the topic did not reach the heading:\n%s", record)
	}
	if !strings.Contains(record, date) {
		t.Errorf("the date was not filled in:\n%s", record)
	}

	// The living spec has to link back to it, or the record is unreachable from
	// the thing it explains.
	index := readFile(t, dir, filepath.Join("docs", "spec", "acme-asgard", "README.md"))
	if !strings.Contains(index, name) {
		t.Errorf("the traceability table was not updated:\n%s", index)
	}

	if _, err := runCLI(t, dir, "check"); err != nil {
		t.Errorf("check after decision add: %v", err)
	}
}

func TestQuestionAnsweredMovesTheRow(t *testing.T) {
	dir := scaffolded(t)

	if _, err := runCLI(t, dir, "question", "add", "who owns the API credentials"); err != nil {
		t.Fatalf("question add: %v", err)
	}
	if _, err := runCLI(t, dir, "question", "answered", "1", "the integrations team"); err != nil {
		t.Fatalf("question answered: %v", err)
	}

	questions, err := work.ReadQuestions(dir)
	if err != nil {
		t.Fatalf("ReadQuestions: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("questions = %+v, want none open", questions)
	}

	body := readFile(t, dir, work.QuestionFile)
	answered := body[strings.Index(body, "## Answered"):]
	if !strings.Contains(answered, "who owns the API credentials") {
		t.Errorf("the row should move rather than be deleted:\n%s", answered)
	}
}

func TestTheWorkCommandsRejectWhatWouldGoWrongLater(t *testing.T) {
	dir := scaffolded(t, "erp")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			// A typo'd project would leave the request looking targeted while
			// `next` kept reading it as an unfinished interview forever.
			"an unknown project on request target",
			[]string{"request", "target", "REQ-001", "warehouse"},
			"no project",
		},
		{
			"an unknown project on task add",
			[]string{"task", "add", "a thing", "--project", "warehouse"},
			"no project",
		},
		{
			"an unknown request on task add",
			[]string{"task", "add", "a thing", "--project", "erp", "--request", "REQ-009"},
			"REQ-009",
		},
		{
			"an unknown id on a status move",
			[]string{"task", "done", "TASK-009"},
			"TASK-009",
		},
		{
			// A title in the customer's own words is the normal case and has no
			// ASCII to name a file with. Asking is better than writing
			// REQ-001-.md.
			"a title with no ASCII and no slug",
			[]string{"request", "add", "客戶要能在官網問庫存"},
			"--slug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runCLI(t, dir, tt.args...)
			if err == nil {
				t.Fatalf("%v should have failed; output was %q", tt.args, out)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q should mention %q", err, tt.want)
			}
		})
	}
}

// TestNextHonoursStageWithoutAConfig pins a flag that used to do nothing: with
// no config, next took an early exit and printed stage 0 whatever --stage said.
// A flag the code silently ignores is worse than no flag at all.
func TestNextHonoursStageWithoutAConfig(t *testing.T) {
	dir := dirNamed(t, "not-an-onboarding")

	out, err := runCLI(t, dir, "next", "--stage", "scaffold")
	if err != nil {
		t.Fatalf("next --stage scaffold: %v", err)
	}
	if !strings.Contains(out, "stage 1 of") {
		t.Errorf("--stage was ignored:\n%s", out)
	}

	// And a typo is reported rather than falling back to stage 0.
	if _, err := runCLI(t, dir, "next", "--stage", "nonsense"); err == nil {
		t.Error("an unknown stage should be an error, not a silent default")
	}

	// With no flag it still explains how to begin, which is the whole point of
	// running next in an empty directory.
	out, err = runCLI(t, dir, "next")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if !strings.Contains(out, "not started yet") {
		t.Errorf("next should still explain how to begin:\n%s", out)
	}
}

// TestRenderAndVerifyNeedNoShell is what this round is for: the gate used to be
// bash calling yq, piped into Python that needed PyYAML in a virtualenv, and so
// none of it ran on Windows. These two commands replace all of it.
func TestRenderAndVerifyNeedNoShell(t *testing.T) {
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm is not installed; asgard-cli doctor is what reports that to a user")
	}
	dir := scaffolded(t, "erp")

	// A CR with a display annotation and no dangling reference passes.
	if _, err := runCLI(t, dir, "add", "dataconnector", "erp", "--project", "erp"); err != nil {
		t.Fatalf("add dataconnector: %v", err)
	}

	out, err := runCLI(t, dir, "render", "erp", "dev")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out, "kind: DataConnector") {
		t.Errorf("render did not emit the manifest:\n%s", out)
	}

	if out, err = runCLI(t, dir, "verify"); err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	if !strings.Contains(out, "erp/dev") || !strings.Contains(out, "DataConnector=1") {
		t.Errorf("verify output = %q", out)
	}

	// Not declaring an environment is a decision, and rendering it is refused
	// rather than guessed at.
	if _, err := runCLI(t, dir, "render", "erp", "prod"); err == nil {
		t.Error("rendering an undeclared env should be refused")
	}
	if _, err := runCLI(t, dir, "render", "erp", "staging"); err == nil {
		t.Error("only dev and prod are environments")
	}
}

// TestVerifyCatchesWhatHelmCannotSee renders a chart whose CR is missing the
// annotation the Platform UI reads its name from. helm lint and a server dry run
// both accept it.
func TestVerifyCatchesWhatHelmCannotSee(t *testing.T) {
	dir := scaffolded(t)

	rendered := filepath.Join(dir, "rendered.yaml")
	body := "apiVersion: asgard-ai.com/v1alpha1\nkind: DataConnector\nmetadata:\n  name: dc-erp\n"
	if err := os.WriteFile(rendered, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	out, err := runCLI(t, dir, "verify", "--rendered", rendered)
	if err == nil {
		t.Fatalf("verify should have failed:\n%s", out)
	}
	if !strings.Contains(out, "asgard-ai.com/data-connector-name") {
		t.Errorf("verify did not name the missing annotation:\n%s", out)
	}
}

// TestEveryExtractIsReachable is the guard against what this repo has now done
// three times: writing material and leaving nothing that points at it. Three
// extracts sat unreferenced by any generator and any prompt, and the note that
// they existed was a line in a table nobody read.
//
// Reachable means a `generate.Kind` sends the reader there, or a stage prompt
// does. Discoverable through `asgard-cli usecase` alone does not count: nobody
// browses a list of sixteen for a mechanism they do not know they need.
func TestEveryExtractIsReachable(t *testing.T) {
	available, err := usecase.List()
	if err != nil {
		t.Fatalf("usecase.List: %v", err)
	}

	referenced := map[string]bool{}
	for _, kind := range generate.Kinds {
		if kind.Extract == "" {
			t.Errorf("kind %q names no extract", kind.Name)
		}
		for _, name := range append([]string{kind.Extract}, kind.AlsoRead...) {
			referenced[name] = true
		}
	}

	// The prompts, rendered, because a reference may sit inside a conditional.
	cfg := &config.Config{
		Workspace: config.Workspace{ID: "ws_1", Slug: "acme", Name: "acme"},
		Projects:  []config.Project{{Slug: "app", Name: "app", Environments: []config.Env{config.EnvDev}}},
	}
	for _, s := range append(stage.Stages, stage.IdleStage) {
		body, err := s.Prompt(cfg, stage.State{})
		if err != nil {
			t.Fatalf("prompt %s: %v", s.Name, err)
		}
		for _, e := range available {
			if strings.Contains(body, "usecase "+e.Name) {
				referenced[e.Name] = true
			}
		}
	}

	for _, e := range available {
		// conventions is what the shapes assume, not somewhere a reader is sent.
		if e.Name == "conventions" || referenced[e.Name] {
			continue
		}
		t.Errorf("nothing points at %q: no generate.Kind names it and no stage prompt mentions it, "+
			"so the only way to find it is to browse the list", e.Name)
	}

	// And nothing points at an extract that is not there.
	known := map[string]bool{}
	for _, e := range available {
		known[e.Name] = true
	}
	for name := range referenced {
		if !known[name] {
			t.Errorf("something sends the reader to %q, which does not exist", name)
		}
	}
}
