package gate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// displayNameAnnotation maps a kind to the annotation the Platform UI reads its
// name from. Without it the CR applies cleanly and shows up nameless in the UI,
// so nothing before this gate can catch it.
//
// Agent is not required to carry a description: that comes from
// spec.managed.description.
var displayNameAnnotation = map[string]string{
	"Agent":            "agent-name",
	"SemanticLayer":    "semantic-layer-name",
	"DataConnector":    "data-connector-name",
	"SkillSet":         "skill-set-name",
	"SourceSet":        "source-set-name",
	"Syncer":           "syncer-name",
	"Toolset":          "toolset-name",
	"Workflow":         "workflow-name",
	"KnowledgeBase":    "knowledge-base-name",
	"Loader":           "loader-name",
	"BotProvider":      "bot-provider-name",
	"Trigger":          "trigger-name",
	"CompletionModel":  "completion-model-name",
	"Plugin":           "plugin-name",
	"SandboxBlueprint": "sandbox-blueprint-name",
}

// workflowSetTypes are the legal values of the workflow-set-type label, per
// asgard-workflow-service internal/shared.go. `trigger` is the newest of the
// three, so an older CR is not a reliable thing to copy from.
var workflowSetTypes = map[string]bool{
	"bot":             true,
	"automation_tool": true,
	"trigger":         true,
}

// workflowSetLabels is the set a Workflow needs to be visible at all. The list
// endpoint always queries with main-workflow-set=true, and the detail endpoint
// locates a set by workflow-set-id. A Workflow missing these is entirely legal -
// lint, CRD validation and a server dry run all pass - and the only symptom is
// that the Platform UI has nothing to list.
var workflowSetLabels = []string{
	"main-workflow-set",
	"workflow-set-id",
	"workflow-key",
	"workflow-set-type",
}

const (
	managedByKey       = "asgard-ai.com/managed-by"
	managedBySkillSet  = "skill-set"
	annotationPrefix   = "asgard-ai.com/"
	workflowSetIDLabel = annotationPrefix + "workflow-set-id"
)

// Xref checks that every reference between rendered CRs resolves, and that the
// metadata the Platform UI depends on is present.
func Xref(docs []Doc, opts Options) Result {
	ix := newIndex(docs)
	x := &xref{index: ix, opts: opts}

	x.buildLookups()
	x.checkNoDuplicateNames()
	for _, d := range ix.docs {
		x.checkDisplayName(d)
		x.checkKind(d)
	}
	x.checkOneMainPerSet()

	sort.Strings(x.problems)
	return Result{Problems: x.problems, Summary: ix.summary()}
}

type xref struct {
	*index
	opts     Options
	problems []string

	// entriesByWorkflow is what makes a two-part reference checkable: an
	// entrypoint is (workflow, entry), and a wrong entry is as fatal as a wrong
	// workflow while apply still succeeds.
	entriesByWorkflow map[string]map[string]bool
	labelsByWorkflow  map[string]map[string]string
	labelsBySourceSet map[string]map[string]string

	// syncerDestsBySourceSet is the fact a SkillSet's searchPaths are checked
	// against. A SourceSet has no member registry any more, so where a Syncer
	// puts content is the only declaration that content is there.
	syncerDestsBySourceSet map[string]map[string]bool

	// skillSetsBySourceSet finds a SourceSet shared by two SkillSets, which
	// leaves the UI unable to show either one's git configuration.
	skillSetsBySourceSet map[string][]string
}

func (x *xref) errf(format string, args ...any) {
	x.problems = append(x.problems, fmt.Sprintf(format, args...))
}

func (x *xref) buildLookups() {
	x.entriesByWorkflow = map[string]map[string]bool{}
	x.labelsByWorkflow = map[string]map[string]string{}
	x.labelsBySourceSet = map[string]map[string]string{}
	x.syncerDestsBySourceSet = map[string]map[string]bool{}
	x.skillSetsBySourceSet = map[string][]string{}

	for _, d := range x.docs {
		switch d.Kind {
		case "Workflow":
			entries := map[string]bool{}
			for _, e := range digList(d.Spec, "entries") {
				if name := digStr(mapOf(e), "name"); name != "" {
					entries[name] = true
				}
			}
			x.entriesByWorkflow[d.Name] = entries
			x.labelsByWorkflow[d.Name] = d.Labels

		case "SourceSet":
			x.labelsBySourceSet[d.Name] = d.Labels

		case "Syncer":
			ss := digStr(d.Spec, "sourceSetName")
			dest := digStr(d.Spec, "destinationPath")
			if ss != "" && dest != "" {
				if x.syncerDestsBySourceSet[ss] == nil {
					x.syncerDestsBySourceSet[ss] = map[string]bool{}
				}
				x.syncerDestsBySourceSet[ss][dest] = true
			}

		case "SkillSet":
			if ref := digStr(d.Spec, "sourceSetName"); ref != "" {
				x.skillSetsBySourceSet[ref] = append(x.skillSetsBySourceSet[ref], d.Name)
			}
		}
	}
}

func (x *xref) checkDisplayName(d Doc) {
	want, ok := displayNameAnnotation[d.Kind]
	if !ok {
		return
	}
	if d.Annotations[annotationPrefix+want] == "" {
		x.errf("%s/%s: missing annotation %s%s, so the Platform UI shows it with no name",
			d.Kind, d.Name, annotationPrefix, want)
	}
}

// checkEntrypoint verifies a (workflow, entry) pair. The entry half is the one
// that gets missed: a wrong entry name is a dead end that apply accepts.
func (x *xref) checkEntrypoint(where, workflow, entry string) {
	if workflow == "" {
		return
	}
	if !x.has("Workflow", workflow) {
		x.errf("%s: no Workflow/%s%s", where, workflow, x.fix("Workflow", workflow))
		return
	}
	if entry != "" && !x.entriesByWorkflow[workflow][entry] {
		x.errf("%s: Workflow/%s has no entry %q (it has: %s)",
			where, workflow, entry, sorted(x.entriesByWorkflow[workflow]))
	}
}

// remedy names the command that creates a missing CR, per kind. A gate that
// says only what is wrong leaves an agent working from this output alone to
// guess, and the guess is usually to delete the reference rather than create
// what it points at.
var remedy = map[string]string{
	"Agent":            "asgard-cli add agent %s --project %s",
	"DataConnector":    "asgard-cli add dataconnector %s --project %s --db-class postgres|mssql",
	"KnowledgeBase":    "asgard-cli add knowledgedrive %s --project %s",
	"Plugin":           "asgard-cli add plugin %s --project %s",
	"SandboxBlueprint": "asgard-cli add flowagent %s --project %s",
	"SemanticLayer":    "asgard-cli add semanticlayer %s --project %s --connector dc-<name>",
	"SkillSet":         "asgard-cli add skillset %s --project %s --repo <git url>",
	"SourceSet":        "asgard-cli add knowledgedrive %s --project %s",
	"Toolset":          "asgard-cli add querytool %s --project %s",
	"Workflow":         "asgard-cli add flowagent %s --project %s",
}

// ref checks a single reference and reports it against the field it came from.
func (x *xref) ref(where, kind, name string) {
	if name == "" || x.has(kind, name) {
		return
	}
	x.errf("%s: no %s/%s%s", where, kind, name, x.fix(kind, name))
}

// fix appends the command that would create the missing CR, with the prefix
// trimmed off the name because `add` takes the bare form.
func (x *xref) fix(kind, name string) string {
	tmpl, ok := remedy[kind]
	if !ok {
		return ""
	}
	bare := name
	if i := strings.Index(bare, "-"); i > 0 && i <= 4 {
		bare = bare[i+1:]
	}
	return fmt.Sprintf(" - create it with `"+tmpl+"`, or drop the reference", bare, x.opts.ProjectOr())
}

func (x *xref) checkKind(d Doc) {
	switch d.Kind {
	case "Agent":
		x.checkAgent(d)
	case "Toolset":
		for _, t := range digList(d.Spec, "tools") {
			ep := mapOf(dig(mapOf(t), "entrypoint"))
			x.checkEntrypoint(fmt.Sprintf("Toolset/%s.tools.entrypoint", d.Name),
				digStr(ep, "workflow"), digStr(ep, "entry"))
		}
	case "Trigger":
		x.checkTrigger(d)
	case "SemanticLayer":
		x.ref(fmt.Sprintf("SemanticLayer/%s.dataConnectorName", d.Name),
			"DataConnector", digStr(d.Spec, "dataConnectorName"))
		// A SemanticLayer may also bind Toolsets. This rule was missing once,
		// and a layer carried a Toolset that did not exist while the gate
		// stayed green.
		for _, ref := range strList(d.Spec["toolsetNames"]) {
			x.ref(fmt.Sprintf("SemanticLayer/%s.toolsetNames", d.Name), "Toolset", ref)
		}
	case "SkillSet":
		x.checkSkillSet(d)
	case "Syncer":
		x.checkSyncer(d)
	case "Workflow":
		x.checkWorkflow(d)
	case "SandboxBlueprint":
		x.checkSandboxBlueprint(d)
	case "BotProvider":
		ep := mapOf(dig(d.Spec, "entrypoint"))
		x.checkEntrypoint(fmt.Sprintf("BotProvider/%s.entrypoint", d.Name),
			digStr(ep, "workflow"), digStr(ep, "entry"))
	case "Loader":
		x.ref(fmt.Sprintf("Loader/%s.knowledgeBaseName", d.Name),
			"KnowledgeBase", digStr(d.Spec, "knowledgeBaseName"))
		x.ref(fmt.Sprintf("Loader/%s.database.dataConnectorName", d.Name),
			"DataConnector", digStr(d.Spec, "database", "dataConnectorName"))
	}
}

func (x *xref) checkAgent(d Doc) {
	managed := mapOf(d.Spec["managed"])
	if managed == nil {
		return
	}
	for _, ref := range strList(managed["toolsetNames"]) {
		x.ref(fmt.Sprintf("Agent/%s.toolsetNames", d.Name), "Toolset", ref)
	}
	for _, ref := range strList(managed["skillSetNames"]) {
		x.ref(fmt.Sprintf("Agent/%s.skillSetNames", d.Name), "SkillSet", ref)
	}
	for _, sl := range digList(managed, "semanticLayers") {
		x.ref(fmt.Sprintf("Agent/%s.semanticLayers", d.Name),
			"SemanticLayer", digStr(mapOf(sl), "name"))
	}
}

func (x *xref) checkTrigger(d Doc) {
	ep := mapOf(dig(d.Spec, "entrypoint"))
	workflow := digStr(ep, "workflow")
	// A wrong entrypoint means the Trigger fires on schedule and fails every
	// time, while apply and lint stay green: only the CronJob log shows it.
	x.checkEntrypoint(fmt.Sprintf("Trigger/%s.entrypoint", d.Name), workflow, digStr(ep, "entry"))

	// The front end opens a Trigger's editor using these two labels on the
	// Trigger itself. Creating a Trigger from the UI copies them off the
	// entrypoint workflow; a hand-written CR has nobody to copy them.
	setID := d.Labels[workflowSetIDLabel]
	if setID == "" {
		x.errf("Trigger/%s: missing label %s, so its editor opens as a blank canvas",
			d.Name, workflowSetIDLabel)
	} else if want := x.labelsByWorkflow[workflow][workflowSetIDLabel]; want != "" && setID != want {
		x.errf("Trigger/%s: workflow-set-id %q does not match its entrypoint Workflow/%s, which has %q",
			d.Name, setID, workflow, want)
	}
	if d.Labels[annotationPrefix+"project-environment-id"] == "" {
		x.errf("Trigger/%s: missing label %sproject-environment-id, so its editor opens as a blank canvas",
			d.Name, annotationPrefix)
	}
}

func (x *xref) checkSkillSet(d Doc) {
	ref := digStr(d.Spec, "sourceSetName")
	if ref == "" {
		return
	}
	if !x.has("SourceSet", ref) {
		x.errf("SkillSet/%s.sourceSetName: no SourceSet/%s%s", d.Name, ref, x.fix("SourceSet", ref))
		return
	}

	// A searchPath outside every Syncer destination resolves to zero skills,
	// silently. (The other silent failure - listing a parent directory when the
	// platform reads one searchPath as one skill directory - cannot be seen
	// from the CR at all.)
	dests := x.syncerDestsBySourceSet[ref]
	for _, path := range strList(d.Spec["searchPaths"]) {
		if len(dests) == 0 {
			continue
		}
		under := false
		for dest := range dests {
			if strings.HasPrefix(path, dest) {
				under = true
				break
			}
		}
		if !under {
			x.errf("SkillSet/%s.searchPaths[%q] is not under any Syncer destinationPath feeding SourceSet/%s (they are: %s)",
				d.Name, path, ref, sorted(dests))
		}
	}

	// One SkillSet, one SourceSet, one Syncer. The platform's
	// POST /v1/skill-set/from-git creates the three together and binds them,
	// and the UI shows them that way; sharing a SourceSet leaves the UI unable
	// to find a given skill set's git configuration.
	owners := x.skillSetsBySourceSet[ref]
	if len(owners) > 1 {
		sorted := append([]string(nil), owners...)
		sort.Strings(sorted)
		// Reported by the first owner only, or the same problem is reported
		// once per owner.
		if d.Name == sorted[0] {
			x.errf("SourceSet/%s is used by %d SkillSets (%s); each SkillSet needs its own, or the Platform UI cannot find its git configuration",
				ref, len(owners), strings.Join(sorted, ", "))
		}
	}
	if x.labelsBySourceSet[ref][managedByKey] != managedBySkillSet {
		x.errf("SourceSet/%s is SkillSet/%s's own source, so it needs label %s=%s (the front end uses it to recognise the pairing)",
			ref, d.Name, managedByKey, managedBySkillSet)
	}
}

func (x *xref) checkSyncer(d Doc) {
	sourceSet := digStr(d.Spec, "sourceSetName")
	x.ref(fmt.Sprintf("Syncer/%s.sourceSetName", d.Name), "SourceSet", sourceSet)

	// destinationPath and statePath are relative paths inside the volume, and
	// these are the CRD's own CEL rules. Catching them here reads better than
	// having the apiserver reject the upgrade.
	dest := digStr(d.Spec, "destinationPath")
	if dest == "" {
		x.errf("Syncer/%s: missing destinationPath", d.Name)
	} else if why := badRelPath(dest); why != "" {
		x.errf("Syncer/%s.destinationPath=%q: %s", d.Name, dest, why)
	} else if !strings.HasSuffix(dest, "/") {
		x.errf("Syncer/%s.destinationPath=%q: the destination is a directory and must end in /", d.Name, dest)
	}

	if state := digStr(d.Spec, "statePath"); state != "" {
		if why := badRelPath(state); why != "" {
			x.errf("Syncer/%s.statePath=%q: %s", d.Name, state, why)
		} else if strings.HasSuffix(state, "/") {
			x.errf("Syncer/%s.statePath=%q: the state path is a file and must not end in /", d.Name, state)
		}
	}

	// Only a database-class Syncer has a DataConnector.
	x.ref(fmt.Sprintf("Syncer/%s.database.dataConnectorName", d.Name),
		"DataConnector", digStr(d.Spec, "database", "dataConnectorName"))

	// A Syncer feeding a SkillSet's own SourceSet carries the same label;
	// without it, it appears in the general Syncer list although it belongs to
	// that SkillSet.
	if owners, ok := x.skillSetsBySourceSet[sourceSet]; ok {
		if d.Labels[managedByKey] != managedBySkillSet {
			sortedOwners := append([]string(nil), owners...)
			sort.Strings(sortedOwners)
			x.errf("Syncer/%s feeds SkillSet/%s's own SourceSet/%s, so it needs label %s=%s (or it shows up in the general Syncer list)",
				d.Name, strings.Join(sortedOwners, ", "), sourceSet, managedByKey, managedBySkillSet)
		}
	}
}

// badRelPath applies the CRD's relative-path rules.
func badRelPath(v string) string {
	if strings.HasPrefix(v, "/") {
		return "must not start with / (there is no root inside the volume)"
	}
	segments := strings.Split(strings.TrimSuffix(v, "/"), "/")
	for _, s := range segments {
		switch s {
		case "":
			return "must not contain consecutive slashes"
		case ".", "..":
			return "must not contain a . or .. component"
		}
	}
	return ""
}

func (x *xref) checkWorkflow(d Doc) {
	for _, key := range workflowSetLabels {
		if d.Labels[annotationPrefix+key] == "" {
			x.errf("Workflow/%s: missing label %s%s, so it belongs to no workflow set and the Platform UI has nothing to list",
				d.Name, annotationPrefix, key)
		}
	}
	if t := d.Labels[annotationPrefix+"workflow-set-type"]; t != "" && !workflowSetTypes[t] {
		legal := make([]string, 0, len(workflowSetTypes))
		for k := range workflowSetTypes {
			legal = append(legal, k)
		}
		sort.Strings(legal)
		x.errf("Workflow/%s.workflow-set-type=%q is not legal (%s)", d.Name, t, strings.Join(legal, ", "))
	}
	if d.Annotations[annotationPrefix+"workflow-set-name"] == "" {
		x.errf("Workflow/%s: missing annotation %sworkflow-set-name, so the set shows with no name in the Platform UI",
			d.Name, annotationPrefix)
	}

	// Node display names, default entries, relationship ids and the canvas
	// ConfigMap are deliberately not checked since 2026-08-31: the platform
	// falls back for all of that canvas metadata (asgard-workflow-service
	// #336-#340) and the chart no longer carries it.

	for _, exit := range digList(d.Spec, "exits") {
		e := mapOf(exit)
		hw := mapOf(e["handlingWorkflow"])
		x.checkEntrypoint(fmt.Sprintf("Workflow/%s.exits[%s].handlingWorkflow", d.Name, digStr(e, "name")),
			digStr(hw, "name"), digStr(hw, "entry"))
	}

	for _, p := range digList(d.Spec, "processors") {
		proc := mapOf(p)
		for _, c := range digList(proc, "configs") {
			cfg := mapOf(c)
			// Only static literals are checked: an expression or a template is
			// not known until runtime.
			if digStr(cfg, "name") == "sandboxBlueprint" {
				if v := digStr(cfg, "value"); v != "" {
					x.ref(fmt.Sprintf("Workflow/%s.processors[%s].configs[sandboxBlueprint]",
						d.Name, digStr(proc, "name")), "SandboxBlueprint", v)
				}
			}
		}
	}
}

func (x *xref) checkSandboxBlueprint(d Doc) {
	// Every top-level field here is a ValueExprTemplate: the name lists are a
	// comma-separated string in `value`, and the mount lists are a JSON string
	// in `value`. Only static values are checked. (These were once iterated as
	// YAML lists; nothing used them then, so nothing broke.)
	names := func(field string) []string {
		raw := digStr(d.Spec, field, "value")
		var out []string
		for part := range strings.SplitSeq(raw, ",") {
			if s := strings.TrimSpace(part); s != "" {
				out = append(out, s)
			}
		}
		return out
	}

	for _, ref := range names("toolsetNames") {
		x.ref(fmt.Sprintf("SandboxBlueprint/%s.toolsetNames", d.Name), "Toolset", ref)
	}
	for _, ref := range names("skillSetNames") {
		x.ref(fmt.Sprintf("SandboxBlueprint/%s.skillSetNames", d.Name), "SkillSet", ref)
	}
	for _, ref := range names("pluginNames") {
		x.ref(fmt.Sprintf("SandboxBlueprint/%s.pluginNames", d.Name), "Plugin", ref)
	}

	if raw := digStr(d.Spec, "sourceSetMounts", "value"); raw != "" {
		var mounts []struct {
			SourceSetName string `json:"sourceSetName"`
		}
		if err := json.Unmarshal([]byte(raw), &mounts); err != nil {
			x.errf("SandboxBlueprint/%s.sourceSetMounts is not valid JSON: %v", d.Name, err)
		}
		for _, m := range mounts {
			x.ref(fmt.Sprintf("SandboxBlueprint/%s.sourceSetMounts[].sourceSetName", d.Name),
				"SourceSet", m.SourceSetName)
		}
	}

	// spec.agents is a JSON string holding an array of SandboxBlueprintAgent -
	// that is how the CRD defines it - so baseAgentName is only visible after
	// parsing the string. This is the joint of the Flow Agent chain
	// (BotProvider -> Workflow -> SandboxBlueprint -> Agent), and a typo here
	// fails as badly as a wrong entrypoint: lint passes, the CRD passes because
	// agents is just a string, and at runtime a subagent is silently absent so
	// the agent merely looks like it cannot call its tools.
	if raw := digStr(d.Spec, "agents", "value"); raw != "" {
		var agents []struct {
			BaseAgentName string `json:"baseAgentName"`
		}
		if err := json.Unmarshal([]byte(raw), &agents); err != nil {
			x.errf("SandboxBlueprint/%s.agents is not valid JSON: %v", d.Name, err)
		}
		for _, a := range agents {
			x.ref(fmt.Sprintf("SandboxBlueprint/%s.agents[].baseAgentName", d.Name),
				"Agent", a.BaseAgentName)
		}
	}
}

// checkNoDuplicateNames rejects two CRs of the same kind and name in one render.
//
// Kubernetes has no way to hold both: whichever applies second replaces the
// first, so the content of one of them is silently gone. Nothing else catches
// it - helm renders both happily, and apply reports success twice.
//
// The way this happens in practice is a generator writing a shared CR into more
// than one file: several fixed query tools belonging to one Toolset each emitted
// their own copy of that Toolset, and the surviving one carried a single tool.
func (x *xref) checkNoDuplicateNames() {
	seen := map[string]int{}
	for _, d := range x.docs {
		seen[d.Kind+"/"+d.Name]++
	}

	keys := make([]string, 0, len(seen))
	for k, n := range seen {
		if n > 1 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		x.errf("%s is declared %d times in this chart; Kubernetes keeps only the one that applies last, so the others are silently discarded", k, seen[k])
	}
}

// checkOneMainPerSet requires exactly one main per workflow set. Zero means the
// list endpoint cannot see the set at all; two means the front end picks with a
// find() that takes the first, so which one wins depends on list order.
func (x *xref) checkOneMainPerSet() {
	mains := map[string][]string{}
	for name, labels := range x.labelsByWorkflow {
		setID := labels[workflowSetIDLabel]
		if setID == "" {
			continue
		}
		if _, ok := mains[setID]; !ok {
			mains[setID] = nil
		}
		if labels[annotationPrefix+"main-workflow-set"] == "true" {
			mains[setID] = append(mains[setID], name)
		}
	}

	setIDs := make([]string, 0, len(mains))
	for id := range mains {
		setIDs = append(setIDs, id)
	}
	sort.Strings(setIDs)

	for _, id := range setIDs {
		found := mains[id]
		if len(found) == 1 {
			continue
		}
		sort.Strings(found)
		detail := ""
		if len(found) > 0 {
			detail = ": " + strings.Join(found, ", ")
		}
		x.errf("workflow-set %q has %d workflows with main-workflow-set=true, and needs exactly 1%s",
			id, len(found), detail)
	}
}
