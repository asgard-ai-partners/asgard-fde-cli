// Package generate writes CR skeletons into a project's chart.
//
// A reference document tells an agent what a CR should look like; this writes
// one that already is. The difference matters for the parts that fail silently:
// a missing display annotation shows up as a nameless resource in the UI, a
// workflow without its set labels is invisible there, and a field renamed
// upstream still lints clean under its old name. None of those are caught by
// lint, by CRD validation, or by a server dry-run.
package generate

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/chart"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

//go:embed templates
var templates embed.FS

// File is one file a Kind writes.
type File struct {
	Dir      string // directory under the chart's templates/
	Template string // template file name
	Suffix   string // appended to the CR name, for a multi-file kind
}

// Kind is one thing that can be generated.
type Kind struct {
	Name    string // what the user types
	Summary string
	Prefix  string // CR name prefix
	Files   []File
	Needs   []string
	Extract string // which usecase extract covers it

	// AlsoRead are extracts that cover the mechanism rather than the shape.
	// Without these a reader is told how the shape is arranged and not how the
	// pieces inside it pass values to each other, which is where the silent
	// failures are.
	AlsoRead []string
	After    []string // what has to be true before this can pass the gate
	Values   string   // template for the values.yaml keys this CR reads
}

// Kinds lists everything that can be generated, in the order they are usually
// created: a connector before a layer, a layer before an agent.
var Kinds = []Kind{
	{
		Name: "dataconnector", Summary: "connection to one database",
		Prefix: "dc-", Files: []File{{Dir: "data_connector", Template: "dataconnector.yaml.tmpl"}},
		Needs:   []string{"--db-class postgres|mssql"},
		Extract: "semantic-layer",
		Values: `
# <<.DisplayName>>
<<.ValuesKey>>DB:
  host: ""
  port: <<if eq .DBClass "mssql">>1433<<else>>5432<<end>>
  user: ""
  database: ""
<<if eq .DBClass "postgres">>  sslMode: "disable"
<<end>>`,
	},
	{
		Name: "semanticlayer", Summary: "read surface over one system, for an internal audience",
		Prefix: "sl-", Files: []File{{Dir: "semantic_layer", Template: "semanticlayer.yaml.tmpl"}},
		Needs:   []string{"--connector dc-<name>"},
		Extract: "semantic-layer",
		After: []string{
			"Introspect the real database before filling in cubes. Do not guess a",
			"  schema: load .agents/skills/semantic-layer-modeling/ and query it.",
		},
		Values: `
# Reasoning depth for semantic layers. Omitting the field is NOT the same as
# medium: it sends no effort at all and the model's own default applies.
defaultSemanticLayerEffort: "medium"
`,
	},
	{
		Name: "agent", Summary: "one specialist, reached through the platform's agent hub",
		Prefix: "ag-", Files: []File{{Dir: "agent", Template: "agent.yaml.tmpl"}},
		Needs:   []string{"--layer sl-<name> (optional)"},
		Extract: "agent-hub",
		After: []string{
			"a project heading for a deploy needs at least one Syncer, and a SkillSet",
			"  brings one - CD fails a deployed project that has no Syncer at all:",
			"    asgard-cli add skillset base --repo <git url>",
			"  Then reference it from skillSetNames. Nothing is referenced by default,",
			"  because a name that does not exist is a dangling reference.",
			"prompt.task and prompt.format must be byte-identical across every agent",
			"  in this chart - the gate diffs them. Copy from a sibling rather than",
			"  writing fresh ones.",
		},
	},
	{
		Name: "httptool", Summary: "a tool that calls an external HTTP API, plus its Toolset",
		Prefix: "wf-", Files: []File{{Dir: "tool", Template: "httptool.yaml.tmpl"}},
		Needs:    []string{"--toolset ts-<name>", "--write for a gated write path"},
		Extract:  "external-api",
		AlsoRead: []string{"api-oauth", "workflow-chain"},
		After: []string{
			"Measure the request and response against the real API. The body field",
			"  names are ours until someone checks them against theirs.",
			"Add the auth header only once infra has created the key - a secretKeyRef",
			"  to a key that does not exist deploys fine and fails on first call.",
		},
		Values: `
# <<.DisplayName>>
<<.ValuesKey>>:
  endpoint: ""
`,
	},
	{
		Name: "querytool", Summary: "a zero-parameter database query tool, plus its Toolset",
		Prefix: "wf-", Files: []File{{Dir: "tool", Template: "querytool.yaml.tmpl"}},
		Needs:    []string{"--connector dc-<name>", "--toolset ts-<name>"},
		Extract:  "fixed-query-tools",
		AlsoRead: []string{"workflow-chain"},
	},
	{
		Name: "skillset", Summary: "SkillSet with its own SourceSet and git Syncer",
		Prefix: "sk-", Files: []File{{Dir: "skill_set", Template: "skillset.yaml.tmpl"}},
		Needs:   []string{"--repo <git url>", "--private if it needs a PAT"},
		Extract: "skill-set",
		After: []string{
			"searchPaths must name one directory per skill. A parent directory",
			"  resolves to nothing, and no check catches it - the symptom is an agent",
			"  with fewer skills than expected.",
		},
	},
	{
		Name: "trigger", Summary: "scheduled run, with its entrypoint Workflow",
		Prefix: "tr-", Files: []File{{Dir: "trigger", Template: "trigger.yaml.tmpl"}},
		Needs:    []string{"--layer sl-<name> (repeatable)"},
		Extract:  "trigger",
		AlsoRead: []string{"workflow-chain"},
		After: []string{
			"platformMainEnvironmentId must be a real value in chart/values-<env>.yaml.",
			"  A Trigger without it fails the gate, and on a cluster it fires correctly",
			"  while its editor opens as a blank canvas. The id only exists after",
			"  tf-asgard has created the namespace and the platform has reconciled it.",
		},
		Values: `
# <<.DisplayName>>
triggers:
  <<.ValuesKey>>:
    schedule: "0 9 * * *"
    timeZone: "Asia/Taipei"
    # Suspended until the run has been exercised once by hand.
    suspend: "true"
`,
	},
	{
		Name: "knowledgedrive", Summary: "a SourceSet Drive with a knowledge graph, for documents",
		Prefix:  "ss-",
		Files:   []File{{Dir: "source_set", Template: "knowledgedrive.yaml.tmpl"}},
		Needs:   []string{"--connector dc-<name> for the database Syncer"},
		Extract: "knowledge-drive",
		After: []string{
			"Mount it read-only from the blueprint, and tell the agent in its prompt",
			"  to query the graph first and then read only the files it points at -",
			"  otherwise it crawls the whole Drive.",
			"Documents nobody can sync have to be uploaded after deploy, and the",
			"  index has to run once. Until then the knowledge answers are poor: put",
			"  it in the chart README as a post-deploy step with an owner.",
		},
		Values: `
# <<.DisplayName>>
<<.ValuesKey>>:
  timeZone: "Asia/Taipei"
  dbSync:
    schedule: "0 9 * * *"
    suspend: "false"
  contextIndex:
    schedule: "0 10 * * *"
    suspend: "false"
`,
	},
	{
		Name: "plugin", Summary: "a capability bundle a blueprint loads by name, and can pick per request",
		Prefix:  "pg-",
		Files:   []File{{Dir: "plugin", Template: "plugin.yaml.tmpl"}},
		Needs:   []string{"--connector ss-<name> for the skill store (defaults to ss-skill-repos)"},
		Extract: "plugin",
		After: []string{
			"Load it from a blueprint: pluginNames is a comma-separated string, or",
			"  an expression that computes the list from the caller's payload.",
		},
	},
	{
		Name: "flowagent", Summary: "a self-hosted entry point: BotProvider, Workflow and SandboxBlueprint",
		Prefix: "bp-",
		Files: []File{
			{Dir: "", Template: "flowagent-botprovider.yaml.tmpl", Suffix: "bot_provider"},
			{Dir: "", Template: "flowagent-workflow.yaml.tmpl", Suffix: "workflow"},
			{Dir: "", Template: "flowagent-blueprint.yaml.tmpl", Suffix: "sandbox_blueprint"},
		},
		Needs:    []string{"--public for an anonymous audience", "--toolset / --layer for its capabilities", "--bot-class generic|line|telegram|discord|slack (defaults to generic)"},
		Extract:  "flow-agent-single",
		AlsoRead: []string{"workflow-chain", "chat-channel"},
		After: []string{
			"a project heading for a deploy needs at least one Syncer, and a SkillSet",
			"  brings one - CD fails a deployed project that has no Syncer at all:",
			"    asgard-cli add skillset base --repo <git url>",
			"  Then reference it from skillSetNames. Nothing is referenced by default,",
			"  because a name that does not exist is a dangling reference.",
			"An anonymous entry point has no auth to protect it, so the protection is",
			"  on the capability side: read-only chain, zero-parameter tools, read-only",
			"  mounts. That argument stops holding the moment a parameterised or",
			"  write-capable tool is added.",
			"The prompt lives on the workflow's processor. Several specialists means",
			"  adding subagents to the blueprint - see asgard-cli usecase",
			"  flow-agent-supervisor.",
		},
		Values: `
# <<.DisplayName>>
botProviders:
  <<.ValuesKey>>:
    # Flip to take the public endpoint down without deleting anything.
    disabled: false
`,
	},
}

// Find returns the kind by name.
func Find(name string) (Kind, bool) {
	for _, k := range Kinds {
		if k.Name == name {
			return k, true
		}
	}
	return Kind{}, false
}

// Names lists every kind, for error messages.
func Names() string {
	names := make([]string, len(Kinds))
	for i, k := range Kinds {
		names[i] = k.Name
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// Options are the per-kind inputs.
type Options struct {
	Project     string
	Name        string // without the prefix
	DisplayName string
	Connector   string
	Layer       string
	Layers      []string
	Toolset     string
	Repo        string
	Private     bool
	Write       bool
	Public      bool
	DBClass     string
	Force       bool

	// SkillSets the project's chart already declares, filled in by Resolve. A
	// skeleton references only what exists: a name that does not is a dangling
	// reference the gate rejects, written by the tool itself.
	SkillSets []string

	// ToolsetExists reports that --toolset names a Toolset the chart already
	// has, so a template that would otherwise emit its own copy must not.
	ToolsetExists bool

	// BotClass is the BotProvider's channel. It defaults to generic, and it is
	// **immutable after creation** on the platform side, so getting it wrong
	// means a new CR rather than an edit.
	BotClass string
}

// Data is what a template renders with.
type Data struct {
	Options
	Chart     string // the chart's helper prefix
	CRName    string // prefixed
	ValuesKey string // the name as a Helm values key
	Workspace config.Workspace
	SpecSlug  string
}

// valuesKey turns a CR name into something addressable in a values file.
// Helm cannot reach a key containing a hyphen with dot notation, so a name like
// daily-report has to become dailyReport - otherwise the chart fails to parse,
// and only at helm lint, well after it looked fine.
func valuesKey(name string) string {
	parts := strings.Split(name, "-")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// Result reports what was written.
type Result struct {
	Path    string
	Created bool
	Values  bool // an addition to values.yaml rather than a new CR
}

// Write renders the kind into the project's chart. A kind may produce several
// files: some shapes are a chain of CRs that have no life apart from each other.
// BotClassGeneric is the default channel: an HTTP API for your own front end.
const BotClassGeneric = "generic"

// BotClasses is the whole vocabulary of spec.botProviderClass.
var BotClasses = []string{BotClassGeneric, "line", "telegram", "discord", "slack"}

func validBotClass(s string) bool {
	return slices.Contains(BotClasses, s)
}

// botClassNote says what choosing a channel costs, at the moment of choosing.
// Each of these is invisible in the generated YAML.
func botClassNote(class string) string {
	switch class {
	case "line":
		return "botProviderClass line: one LINE official account cannot host two bots, so replacing an existing one is a cutover - deploy with disabled: true first. " +
			"Add line_channel_access_token and line_channel_secret to app-secret before deploying. Read `asgard-cli usecase chat-channel`"
	case "telegram":
		return "botProviderClass telegram: add telegram_bot_token and telegram_webhook_secret to app-secret before deploying. Read `asgard-cli usecase chat-channel`"
	case "discord", "slack":
		return "botProviderClass " + class + ": the operator creates a Connector Pod for this class, because it holds an outbound WebSocket. " +
			"Read `asgard-cli usecase chat-channel`"
	default:
		return ""
	}
}

// Resolve fills in the references a kind needs from what the project's chart
// already declares, and returns what it decided so the caller can say so.
//
// It exists because the generator used to write a skeleton that could not pass
// the gate: an Agent hardcoded a SkillSet nothing creates, and ignored the
// SemanticLayer sitting in the same chart. Both are dangling references the
// moment the file is written, and an agent working from the CLI's output alone
// has no way to know that - it followed the instructions and the gate went red.
func Resolve(root string, kind Kind, opts Options) (Options, []string, error) {
	refs, err := chart.Scan(root, opts.Project)
	if err != nil {
		return opts, nil, err
	}

	var notes []string
	layers := chart.NamesOf(refs, "SemanticLayer")

	if kind.Name == "agent" && opts.Layer == "" {
		switch len(layers) {
		case 0:
			// Legal: an Agent may read through Toolsets instead. The skeleton
			// says so, and the gate rejects one with neither.
		case 1:
			opts.Layer = layers[0]
			notes = append(notes, fmt.Sprintf("mounted the chart's only SemanticLayer, %s", opts.Layer))
		default:
			return opts, nil, fmt.Errorf("this chart has %d semantic layers (%s); one agent takes at most one, so name it with --layer",
				len(layers), strings.Join(layers, ", "))
		}
	}

	if kind.Name == "flowagent" {
		if opts.BotClass == "" {
			opts.BotClass = BotClassGeneric
		}
		if !validBotClass(opts.BotClass) {
			return opts, nil, fmt.Errorf("unknown bot class %q; one of: %s",
				opts.BotClass, strings.Join(BotClasses, ", "))
		}
		if note := botClassNote(opts.BotClass); note != "" {
			notes = append(notes, note)
		}
	}

	opts.SkillSets = chart.NamesOf(refs, "SkillSet")

	// A fixed-query Toolset normally holds several tools, so the second and
	// later ones must not re-emit the set. Writing it twice renders two CRs with
	// one name, and whichever applies last takes the other's tools with it.
	if opts.Toolset != "" {
		for _, name := range chart.NamesOf(refs, "Toolset") {
			if name == opts.Toolset {
				opts.ToolsetExists = true
				notes = append(notes, fmt.Sprintf(
					"%s already exists, so only the tool was written; add its entry to that Toolset's tools list", opts.Toolset))
				break
			}
		}
	}

	return opts, notes, nil
}

func Write(root string, cfg *config.Config, kind Kind, opts Options) ([]Result, error) {
	project, ok := cfg.Project(opts.Project)
	if !ok {
		return nil, fmt.Errorf("no project %q in %s; add it with `asgard-cli project add %s`",
			opts.Project, config.FileName, opts.Project)
	}
	// Every reference in this CLI's own guidance is written prefixed - "--connector
	// dc-<name>", "--layer sl-<name>" - so the prefixed form is the natural thing
	// to type as the name too. Prefixing it again produced dc-dc-erp in
	// metadata.name, which lints clean and is only found by the xref gate, or by
	// nobody. Accept either form.
	opts.Name = strings.TrimPrefix(opts.Name, kind.Prefix)
	if err := config.ValidateSlug("name", opts.Name); err != nil {
		return nil, err
	}

	// Defaulted here as well as in Resolve: an empty class renders a CR with no
	// channel and no credentials, and botProviderClass cannot be edited after it
	// is applied.
	if opts.BotClass == "" {
		opts.BotClass = BotClassGeneric
	}

	data := Data{
		Options:   opts,
		Chart:     project.Slug,
		CRName:    kind.Prefix + opts.Name,
		ValuesKey: valuesKey(opts.Name),
		Workspace: cfg.Workspace,
		SpecSlug:  cfg.Workspace.Slug + "-asgard",
	}
	if data.DisplayName == "" {
		data.DisplayName = data.CRName
	}

	templatesDir := filepath.Join(root, "projects", project.Slug, "chart", "app", "templates")

	// Resolve every path before writing any of them, so a kind that would
	// clobber something does not leave half its files behind.
	targets := make([]string, len(kind.Files))
	for i, f := range kind.Files {
		dir, name := f.Dir, data.CRName+".yaml"
		if f.Suffix != "" {
			// A multi-file kind gets its own directory, named for the chain.
			dir = filepath.Join(f.Dir, opts.Name)
			name = f.Suffix + ".yaml"
		}
		targets[i] = filepath.Join(templatesDir, dir, name)

		if _, err := os.Stat(targets[i]); err == nil && !opts.Force {
			return nil, fmt.Errorf("%s already exists; pass --force to overwrite",
				mustRel(root, targets[i]))
		}
	}

	var results []Result
	for i, f := range kind.Files {
		content, err := render(f.Template, data)
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(targets[i]), 0o755); err != nil {
			return nil, fmt.Errorf("create directory: %w", err)
		}
		if err := os.WriteFile(targets[i], content, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", targets[i], err)
		}
		results = append(results, Result{Path: mustRel(root, targets[i]), Created: true})
	}

	if kind.Values != "" {
		added, err := appendValues(templatesDir, kind, data)
		if err != nil {
			return nil, err
		}
		if added != "" {
			results = append(results, Result{Path: mustRel(root, added), Created: true, Values: true})
		}
	}

	return results, nil
}

// appendValues adds the keys this CR reads to the chart's values.yaml, unless
// they are declared already. values.yaml has to default every .Values.* a
// template reads: the bare helm lint is the only step that proves it, and
// without it a missing default is masked whenever an env file is overlaid, then
// nil-pointers for anyone running plain helm template.
func appendValues(templatesDir string, kind Kind, data Data) (string, error) {
	chartDir := filepath.Dir(templatesDir)
	path := filepath.Join(chartDir, "values.yaml")

	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	snippet, err := renderInline(kind.Values, data)
	if err != nil {
		return "", err
	}

	// A snippet either introduces its own top-level key, or adds an entry under
	// one that several CRs share. Both have to work, and only the first used to.
	//
	// The bug that came of it: `botProviders:` exists after the first flow
	// agent, so the second one's whole snippet was skipped as "already there",
	// and its template then read `.Values.botProviders.<name>.disabled` off a
	// map with no such entry. Bare `helm lint` catches it - which is exactly
	// what bare lint is for - but only after the file is written.
	top, nested := splitSnippet(string(snippet))
	if top == "" {
		return "", nil
	}

	text := string(existing)
	if !strings.Contains("\n"+text, "\n"+top+":") {
		// New block: append the snippet whole.
		if err := os.WriteFile(path, []byte(text+string(snippet)), 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", path, err)
		}
		return path, nil
	}

	// The block exists. Add this CR's entry under it, unless it is already there.
	if len(nested) == 0 {
		return "", nil
	}
	entry := strings.TrimSpace(nested[0])
	if strings.Contains(text, "\n  "+entry) {
		return "", nil
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != top+":" {
			continue
		}
		updated := append([]string{}, lines[:i+1]...)
		updated = append(updated, nested...)
		updated = append(updated, lines[i+1:]...)
		if err := os.WriteFile(path, []byte(strings.Join(updated, "\n")), 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", path, err)
		}
		return path, nil
	}
	return "", nil
}

// splitSnippet separates a values snippet's top-level key from the indented
// lines under it, dropping the leading comment - a comment naming one CR does
// not belong above a block that several share.
func splitSnippet(snippet string) (top string, nested []string) {
	for _, line := range strings.Split(snippet, "\n") {
		switch {
		case strings.TrimSpace(line) == "":
			continue
		case top == "" && strings.HasPrefix(line, "#"):
			continue
		case top == "":
			key, _, found := strings.Cut(line, ":")
			if !found || key == "" || strings.HasPrefix(line, " ") {
				return "", nil
			}
			top = key
		default:
			nested = append(nested, line)
		}
	}
	return top, nested
}

func renderInline(text string, data Data) ([]byte, error) {
	tmpl, err := template.New("values").Delims("<<", ">>").Funcs(funcs).Parse(text)
	if err != nil {
		return nil, fmt.Errorf("parse values snippet: %w", err)
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("render values snippet: %w", err)
	}
	return []byte(out.String()), nil
}

// funcs are what a skeleton template may call. A SandboxBlueprint's name lists
// are comma-separated strings in one field, not YAML lists, so joining is not a
// convenience here - it is the field's format.
var funcs = template.FuncMap{
	"join": func(items []string, sep string) string { return strings.Join(items, sep) },
}

func render(name string, data Data) ([]byte, error) {
	raw, err := templates.ReadFile("templates/" + name)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", name, err)
	}

	// << >> because the output is a Helm template full of {{ }}.
	tmpl, err := template.New(name).Delims("<<", ">>").Funcs(funcs).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("render %s: %w", name, err)
	}
	return []byte(out.String()), nil
}

func mustRel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
