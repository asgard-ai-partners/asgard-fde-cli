// Package stage works out how far an onboarding has got and what to do next.
//
// The stage is derived from the repository itself - which files exist, which
// CR kinds are present - rather than stored in the config. A stored counter
// would be a second source of truth, and it would go stale the moment someone
// does a step by hand.
package stage

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/chart"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

//go:embed prompts
var prompts embed.FS

// Name identifies one step of an onboarding.
type Name string

const (
	Init        Name = "init"
	Scaffold    Name = "scaffold"
	Projects    Name = "projects"
	DataSources Name = "data-sources"
	ReadPath    Name = "read-path"
	EntryPoint  Name = "entry-point"
	Knowledge   Name = "knowledge"
	Verify      Name = "verify"
	Deploy      Name = "deploy"
	Enhance     Name = "enhance"
	Idle        Name = "idle"

	Requirements Name = "requirements"
)

// Stage is one step, in the order they have to happen.
type Stage struct {
	Name    Name
	Number  int
	Title   string
	promptF string
}

// Stages lists every stage in order.
var Stages = []Stage{
	{Init, 0, "Start the onboarding", "00-init.md"},
	{Scaffold, 1, "Write the repository skeleton", "01-scaffold.md"},
	{Projects, 2, "Decide how the work splits into projects", "02-projects.md"},
	{DataSources, 3, "Wire up the customer's databases", "03-data-sources.md"},
	{ReadPath, 4, "Decide each project's read path", "04-read-path.md"},
	{EntryPoint, 5, "Decide each project's entry point", "05-entry-point.md"},
	{Knowledge, 6, "Decide where unstructured knowledge lives", "06-knowledge.md"},
	{Verify, 7, "Run the acceptance gate", "07-verify.md"},
	{Deploy, 8, "Deploy", "08-deploy.md"},
	// Enhance is not part of the linear walk: Current never returns it, because
	// "the onboarding is finished" is not something the repo can tell you - a
	// deployed repo and one waiting for its first tag look the same on disk.
	// Reach it with --stage enhance once the repo is live.
	{Enhance, 9, "Add a capability to a repo that is already live", "09-enhance.md"},
}

// IdleStage is what Current returns when nothing is in flight: no request, no
// task, and no project waiting for a step. It sits outside the numbered walk on
// purpose, because it is not a step of an onboarding - it is the state between
// two of them, and the only question it can ask is what the customer wants next.
var IdleStage = Stage{Idle, -1, "Nothing in flight", "10-idle.md"}

// RequirementsStage is the interview that turns what a customer said into a
// request. It is outside the numbered walk for the same reason IdleStage is,
// but the opposite way round: not a state between two steps, but the one an FDE
// is in before the repository can show anything at all. Current never returns
// it, because a conversation leaves no trace on disk until it is written down -
// so it is read deliberately, with `next --stage requirements`, and it is worth
// reading again at every later request rather than only the first.
var RequirementsStage = Stage{Requirements, -1, "Turn what the customer said into a request", "11-requirements.md"}

// Readable is every stage a person can ask for by name, which is more than the
// walk: the two outside it are reached only with `next --stage`, and anything
// iterating stages (rendering, --list, tests) has to see them too.
var Readable = append(append([]Stage{}, Stages...), RequirementsStage, IdleStage)

// Find returns the stage with the given name.
func Find(name string) (Stage, bool) {
	for _, s := range Readable {
		if string(s.Name) == name {
			return s, true
		}
	}
	return Stage{}, false
}

// State is what the repository looks like right now.
type State struct {
	Scaffolded bool
	Projects   []ProjectState
	Requests   []work.Request
	Tasks      []work.Task
	Questions  []work.Question
}

// InFlight reports whether the repository records any work not yet done. It is
// what separates "the onboarding is between two pieces of work" from "somebody
// is halfway through one".
func (s State) InFlight() bool {
	return len(work.ActiveRequests(s.Requests)) > 0 || len(work.ActiveTasks(s.Tasks)) > 0
}

// ProjectState is what one project's chart contains.
type ProjectState struct {
	config.Project
	Kinds map[string]int
}

// Has reports whether the project's chart declares any CR of these kinds.
func (p ProjectState) Has(kinds ...string) bool {
	for _, kind := range kinds {
		if p.Kinds[kind] > 0 {
			return true
		}
	}
	return false
}

// summaryOrder is the order CR kinds are reported in. It is fixed so that the
// same repository prints the same line twice, which a map range would not do.
var summaryOrder = []string{
	"DataConnector", "SemanticLayer", "Toolset", "Workflow", "Agent",
	"BotProvider", "SandboxBlueprint", "SkillSet", "SourceSet", "Syncer",
	"Trigger", "Plugin", "CompletionModel",
}

// Summary names the CR kinds this project's chart declares. Anything not in the
// known order is listed after them, sorted, so a kind added to the platform
// shows up rather than vanishing.
func (p ProjectState) Summary() string {
	if len(p.Kinds) == 0 {
		return "chart is empty"
	}

	seen := map[string]bool{}
	var parts []string
	add := func(kind string) {
		if n := p.Kinds[kind]; n > 1 {
			parts = append(parts, fmt.Sprintf("%s x%d", kind, n))
		} else if n == 1 {
			parts = append(parts, kind)
		}
	}
	for _, kind := range summaryOrder {
		seen[kind] = true
		add(kind)
	}

	var rest []string
	for kind := range p.Kinds {
		if !seen[kind] {
			rest = append(rest, kind)
		}
	}
	sort.Strings(rest)
	for _, kind := range rest {
		add(kind)
	}
	return strings.Join(parts, ", ")
}

// Inspect reads the repository at root and reports its state.
func Inspect(root string, cfg *config.Config) (State, error) {
	state := State{}

	// AGENTS.md is the marker: scaffold always writes it, and it is the file an
	// agent is told to read first.
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err == nil {
		state.Scaffolded = true
	} else if !os.IsNotExist(err) {
		return state, fmt.Errorf("stat AGENTS.md: %w", err)
	}

	for _, project := range cfg.Projects {
		ps := ProjectState{Project: project, Kinds: map[string]int{}}

		refs, err := chart.Scan(root, project.Slug)
		if err != nil {
			return state, err
		}
		ps.Kinds = chart.Counts(refs)

		state.Projects = append(state.Projects, ps)
	}

	// The three records the repository keeps of its own work. They are read
	// rather than derived: a stage is a fact about files, a status is a fact
	// somebody wrote down, and neither can stand in for the other.
	requests, err := work.ReadRequests(root)
	if err != nil {
		return state, err
	}
	tasks, err := work.ReadTasks(root)
	if err != nil {
		return state, err
	}
	questions, err := work.ReadQuestions(root)
	if err != nil {
		return state, err
	}
	state.Requests, state.Tasks, state.Questions = requests, tasks, questions

	return state, nil
}

// Current returns the first stage that is not finished yet. The stages are
// ordered by dependency, so the first unfinished one is what to do next.
func Current(state State) Stage {
	// Look stages up by name rather than by index: the list has an entry for
	// "not started yet" at the front, and indices would silently shift again
	// the next time one is inserted.
	if !state.Scaffolded {
		return mustFind(Scaffold)
	}

	// A request whose target project is not decided yet is the split interview,
	// whatever else the repo already contains. The interview belongs to a
	// request rather than to the repository, which is why it is not reached by
	// "there are no projects": a repo with no projects and no request in flight
	// is not waiting to be interviewed, it is waiting for a requirement.
	for _, r := range work.ActiveRequests(state.Requests) {
		if !state.hasProject(r.Project) {
			return mustFind(Projects)
		}
	}

	// A project reads before it answers, and answers before it is deployed, so
	// the earliest project still missing a step decides the stage.
	for _, s := range []struct {
		stage Name
		kinds []string
	}{
		{DataSources, []string{"DataConnector"}},
		{ReadPath, []string{"SemanticLayer", "Toolset"}},
		{EntryPoint, []string{"Agent", "BotProvider"}},
	} {
		for _, p := range state.Projects {
			if !p.Has(s.kinds...) {
				return mustFind(s.stage)
			}
		}
	}

	// Every chart is complete. Whether that is "gate it" or "there is nothing
	// to do" is not a fact about the files - a deployed repo and one waiting for
	// its first tag look identical on disk - so the answer comes from whether
	// the repo records any work still open.
	if state.InFlight() {
		return mustFind(Verify)
	}
	return IdleStage
}

// hasProject reports whether slug names a project this repository has. An empty
// slug never does: it is what a request carries before its audience is decided.
func (s State) hasProject(slug string) bool {
	if slug == "" {
		return false
	}
	for _, p := range s.Projects {
		if p.Slug == slug {
			return true
		}
	}
	return false
}

// mustFind panics on an unknown stage, which can only be a programming error:
// the names come from constants in this file.
func mustFind(name Name) Stage {
	s, ok := Find(string(name))
	if !ok {
		panic("stage: unknown stage " + name)
	}
	return s
}

// Data is what a stage prompt is rendered with.
type Data struct {
	Workspace config.Workspace
	Projects  []ProjectState
	Requests  []work.Request
	RepoName  string
	SpecSlug  string
	Stage     Stage
}

// InterviewRequests are the open requests whose target project is not decided
// yet. They are what stage 2 is about, and a prompt read out of order has none.
func (d Data) InterviewRequests() []work.Request {
	var out []work.Request
	for _, r := range d.Requests {
		if r.Project == "" {
			out = append(out, r)
		}
	}
	return out
}

// RequestID names the request a prompt should refer to in the commands it
// prints: the first one still waiting on its project, otherwise the first open
// one, otherwise a placeholder, so that reading a stage out of order still
// produces a command somebody can adapt rather than a broken one.
func (d Data) RequestID() string {
	if waiting := d.InterviewRequests(); len(waiting) > 0 {
		return waiting[0].ID
	}
	if len(d.Requests) > 0 {
		return d.Requests[0].ID
	}
	return "REQ-xxx"
}

// Prompt renders the stage's guidance.
func (s Stage) Prompt(cfg *config.Config, state State) (string, error) {
	content, err := prompts.ReadFile("prompts/" + s.promptF)
	if err != nil {
		return "", fmt.Errorf("read prompt %s: %w", s.promptF, err)
	}

	tmpl, err := template.New(s.promptF).Delims("<<", ">>").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parse prompt %s: %w", s.promptF, err)
	}

	var out strings.Builder
	err = tmpl.Execute(&out, Data{
		Workspace: cfg.Workspace,
		Projects:  state.Projects,
		Requests:  work.ActiveRequests(state.Requests),
		RepoName:  cfg.RepoName(),
		SpecSlug:  cfg.Workspace.Slug + "-asgard",
		Stage:     s,
	})
	if err != nil {
		return "", fmt.Errorf("render prompt %s: %w", s.promptF, err)
	}
	return out.String(), nil
}

func (s Stage) String() string {
	// Switch on the name, not the number: the stages outside the walk share the
	// number -1, so matching on it labelled the requirements interview "nothing
	// in flight" - the opposite of what a reader is being told at that point.
	switch s.Name {
	case Idle:
		return "nothing in flight"
	case Requirements:
		return "before the walk: the interview"
	case Init:
		return "not started yet"
	}
	return fmt.Sprintf("stage %d of %d: %s", s.Number, len(Stages)-1, s.Title)
}

// NeedsTask reports whether work at this stage should be written up as a task
// spec first. The stages that change a chart are the ones the SDD rules cover:
// a new SemanticLayer or DataConnector, widening what an agent may query, or
// introducing a write path.
func (s Stage) NeedsTask() bool {
	switch s.Name {
	case DataSources, ReadPath, EntryPoint, Knowledge, Enhance:
		return true
	default:
		return false
	}
}
