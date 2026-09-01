package gate

import (
	"strings"
	"testing"
)

// docs parses a rendered stream the way verify does.
func docs(t *testing.T, yaml string) []Doc {
	t.Helper()
	parsed, err := Read(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return parsed
}

// wants asserts on the problems a check found, by substring: the exact wording
// is allowed to improve, the rule that fired is not.
func wants(t *testing.T, r Result, substrings ...string) {
	t.Helper()
	if len(r.Problems) != len(substrings) {
		t.Fatalf("found %d problems, want %d:\n  %s", len(r.Problems), len(substrings),
			strings.Join(r.Problems, "\n  "))
	}
	for i, want := range substrings {
		if !strings.Contains(r.Problems[i], want) {
			t.Errorf("problem %d = %q, want it to mention %q", i, r.Problems[i], want)
		}
	}
}

func TestReadSkipsWhatIsNotAManifest(t *testing.T) {
	// helm writes a leading separator, and a stderr line captured into the same
	// file must not fail the gate for the wrong reason.
	parsed := docs(t, `
---
rendering something to stderr
---
apiVersion: asgard-ai.com/v1alpha1
kind: DataConnector
metadata:
  name: dc-erp
  annotations:
    asgard-ai.com/data-connector-name: erp
---
`)
	if len(parsed) != 1 {
		t.Fatalf("parsed %d docs, want 1: %+v", len(parsed), parsed)
	}
	if parsed[0].Kind != "DataConnector" || parsed[0].Name != "dc-erp" {
		t.Errorf("doc = %+v", parsed[0])
	}
}

func TestXrefWantsTheDisplayAnnotation(t *testing.T) {
	// The whole reason this rule exists: the CR below applies cleanly, passes
	// lint and passes a server dry run, and shows up nameless in the UI.
	r := Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: DataConnector
metadata:
  name: dc-erp
`), Options{Project: "erp"})
	wants(t, r, "missing annotation asgard-ai.com/data-connector-name")
}

func TestXrefResolvesReferences(t *testing.T) {
	r := Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: Agent
metadata:
  name: ag-erp
  annotations:
    asgard-ai.com/agent-name: erp
spec:
  managed:
    toolsetNames: [ts-missing]
    skillSetNames: [sk-missing]
    semanticLayers:
      - name: sl-missing
`), Options{Project: "erp"})
	wants(t, r,
		"Agent/ag-erp.semanticLayers: no SemanticLayer/sl-missing",
		"Agent/ag-erp.skillSetNames: no SkillSet/sk-missing",
		"Agent/ag-erp.toolsetNames: no Toolset/ts-missing",
	)
}

// TestXrefChecksBothHalvesOfAnEntrypoint is the rule that catches the failure a
// wrong workflow name and a wrong entry name share: apply accepts both.
func TestXrefChecksBothHalvesOfAnEntrypoint(t *testing.T) {
	manifests := `
apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: wf-site
  annotations:
    asgard-ai.com/workflow-name: site
    asgard-ai.com/workflow-set-name: site
  labels:
    asgard-ai.com/main-workflow-set: "true"
    asgard-ai.com/workflow-set-id: set-1
    asgard-ai.com/workflow-key: main
    asgard-ai.com/workflow-set-type: bot
spec:
  entries:
    - name: chat
---
apiVersion: asgard-ai.com/v1alpha1
kind: BotProvider
metadata:
  name: bp-site
  annotations:
    asgard-ai.com/bot-provider-name: site
spec:
  entrypoint:
    workflow: wf-site
    entry: %s
`
	if r := Xref(docs(t, strings.Replace(manifests, "%s", "chat", 1)), Options{Project: "erp"}); !r.OK() {
		t.Fatalf("the correct entry should pass: %v", r.Problems)
	}
	r := Xref(docs(t, strings.Replace(manifests, "%s", "chatt", 1)), Options{Project: "erp"})
	wants(t, r, `Workflow/wf-site has no entry "chatt" (it has: chat)`)
}

func TestXrefWantsTheWorkflowSetMetadata(t *testing.T) {
	r := Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: wf-site
  annotations:
    asgard-ai.com/workflow-name: site
  labels:
    asgard-ai.com/main-workflow-set: "true"
    asgard-ai.com/workflow-set-id: set-1
    asgard-ai.com/workflow-key: main
    asgard-ai.com/workflow-set-type: chatbot
`), Options{Project: "erp"})
	wants(t, r,
		"Workflow/wf-site.workflow-set-type=\"chatbot\" is not legal",
		"missing annotation asgard-ai.com/workflow-set-name",
	)
}

// TestXrefWantsExactlyOneMainPerSet covers both directions: zero means the list
// endpoint cannot see the set, two means the front end's find() picks by list
// order.
func TestXrefWantsExactlyOneMainPerSet(t *testing.T) {
	workflow := func(name, main string) string {
		return `
apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: ` + name + `
  annotations:
    asgard-ai.com/workflow-name: x
    asgard-ai.com/workflow-set-name: x
  labels:
    asgard-ai.com/main-workflow-set: "` + main + `"
    asgard-ai.com/workflow-set-id: set-1
    asgard-ai.com/workflow-key: ` + name + `
    asgard-ai.com/workflow-set-type: bot
`
	}

	r := Xref(docs(t, workflow("wf-a", "false")+"---"+workflow("wf-b", "false")), Options{Project: "erp"})
	wants(t, r, `workflow-set "set-1" has 0 workflows with main-workflow-set=true`)

	r = Xref(docs(t, workflow("wf-a", "true")+"---"+workflow("wf-b", "true")), Options{Project: "erp"})
	wants(t, r, `workflow-set "set-1" has 2 workflows with main-workflow-set=true, and needs exactly 1: wf-a, wf-b`)

	if r := Xref(docs(t, workflow("wf-a", "true")+"---"+workflow("wf-b", "false")), Options{Project: "erp"}); !r.OK() {
		t.Errorf("one main should pass: %v", r.Problems)
	}
}

func TestXrefChecksSyncerPaths(t *testing.T) {
	syncer := func(dest, state string) string {
		return `
apiVersion: asgard-ai.com/v1alpha1
kind: Syncer
metadata:
  name: syn-x
  annotations:
    asgard-ai.com/syncer-name: x
spec:
  destinationPath: ` + dest + `
  statePath: ` + state + `
`
	}

	tests := []struct {
		dest, state, want string
	}{
		{"/git/", "state.json", "must not start with /"},
		{"git//x/", "state.json", "must not contain consecutive slashes"},
		{"git/../x/", "state.json", "must not contain a . or .. component"},
		{"git/x", "state.json", "the destination is a directory and must end in /"},
		{"git/x/", "state/", "the state path is a file and must not end in /"},
	}
	for _, tt := range tests {
		t.Run(tt.dest+" "+tt.state, func(t *testing.T) {
			wants(t, Xref(docs(t, syncer(tt.dest, tt.state)), Options{Project: "erp"}), tt.want)
		})
	}

	if r := Xref(docs(t, syncer("git/x/", "state.json")), Options{Project: "erp"}); !r.OK() {
		t.Errorf("legal paths should pass: %v", r.Problems)
	}
}

// TestXrefParsesTheJSONStringFields pins the SandboxBlueprint shape: the CRD
// declares these as strings, so a name inside one is invisible to CRD
// validation, and a typo silently costs a subagent at runtime.
func TestXrefParsesTheJSONStringFields(t *testing.T) {
	r := Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: SandboxBlueprint
metadata:
  name: sbp-site
  annotations:
    asgard-ai.com/sandbox-blueprint-name: site
spec:
  toolsetNames:
    value: "ts-one, ts-two"
  agents:
    value: '[{"baseAgentName":"ag-missing"}]'
  sourceSetMounts:
    value: '[{"sourceSetName":"ss-missing"}]'
`), Options{Project: "erp"})
	wants(t, r,
		"SandboxBlueprint/sbp-site.agents[].baseAgentName: no Agent/ag-missing",
		"SandboxBlueprint/sbp-site.sourceSetMounts[].sourceSetName: no SourceSet/ss-missing",
		"SandboxBlueprint/sbp-site.toolsetNames: no Toolset/ts-one",
		"SandboxBlueprint/sbp-site.toolsetNames: no Toolset/ts-two",
	)

	r = Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: SandboxBlueprint
metadata:
  name: sbp-site
  annotations:
    asgard-ai.com/sandbox-blueprint-name: site
spec:
  agents:
    value: 'not json'
`), Options{Project: "erp"})
	wants(t, r, "SandboxBlueprint/sbp-site.agents is not valid JSON")
}

func TestXrefWantsOneSourceSetPerSkillSet(t *testing.T) {
	r := Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: SourceSet
metadata:
  name: ss-shared
  annotations:
    asgard-ai.com/source-set-name: shared
---
apiVersion: asgard-ai.com/v1alpha1
kind: SkillSet
metadata:
  name: sk-a
  annotations:
    asgard-ai.com/skill-set-name: a
spec:
  sourceSetName: ss-shared
---
apiVersion: asgard-ai.com/v1alpha1
kind: SkillSet
metadata:
  name: sk-b
  annotations:
    asgard-ai.com/skill-set-name: b
spec:
  sourceSetName: ss-shared
`), Options{Project: "erp"})
	// Reported once, by the first owner, not once per owner.
	wants(t, r,
		"SourceSet/ss-shared is SkillSet/sk-a's own source, so it needs label asgard-ai.com/managed-by=skill-set",
		"SourceSet/ss-shared is SkillSet/sk-b's own source, so it needs label asgard-ai.com/managed-by=skill-set",
		"SourceSet/ss-shared is used by 2 SkillSets (sk-a, sk-b)",
	)
}

func TestXrefWantsSearchPathsUnderASyncerDestination(t *testing.T) {
	// A searchPath outside every Syncer destination resolves to zero skills,
	// silently.
	r := Xref(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: SourceSet
metadata:
  name: ss-sk
  annotations:
    asgard-ai.com/source-set-name: sk
  labels:
    asgard-ai.com/managed-by: skill-set
---
apiVersion: asgard-ai.com/v1alpha1
kind: Syncer
metadata:
  name: syn-sk
  annotations:
    asgard-ai.com/syncer-name: sk
  labels:
    asgard-ai.com/managed-by: skill-set
spec:
  sourceSetName: ss-sk
  destinationPath: git/
---
apiVersion: asgard-ai.com/v1alpha1
kind: SkillSet
metadata:
  name: sk-base
  annotations:
    asgard-ai.com/skill-set-name: base
spec:
  sourceSetName: ss-sk
  searchPaths:
    - git/common/skills/one
    - elsewhere/skills/two
`), Options{Project: "erp"})
	wants(t, r, `SkillSet/sk-base.searchPaths["elsewhere/skills/two"] is not under any Syncer destinationPath`)
}

func TestAgentSplitAcceptsNoAgentsAtAll(t *testing.T) {
	// A pure Flow Agent project keeps its prompt in a Workflow and its
	// capability in a SandboxBlueprint, so its chart has no Agent CR.
	r := AgentSplit(nil, Options{})
	if !r.OK() {
		t.Fatalf("zero agents is legal: %v", r.Problems)
	}
	if !strings.Contains(r.Summary, "0 agent(s)") {
		t.Errorf("summary = %q, want it to say 0 agent(s) so a chart that lost them is visible", r.Summary)
	}
}

func TestAgentSplitRules(t *testing.T) {
	agent := func(body string) string {
		return "apiVersion: asgard-ai.com/v1alpha1\nkind: Agent\n" + body
	}

	tests := []struct {
		name  string
		yaml  string
		opts  Options
		wants []string
	}{
		{
			"R1 more than one layer",
			agent(`metadata:
  name: ag-a
spec:
  managed:
    semanticLayers:
      - name: sl-one
      - name: sl-two
`),
			Options{},
			[]string{"R1 ag-a: has 2 semanticLayers"},
		},
		{
			"R1 the same layer twice",
			agent(`metadata:
  name: ag-a
spec:
  managed:
    semanticLayers: [{name: sl-one}]
`) + "---\n" + agent(`metadata:
  name: ag-b
spec:
  managed:
    semanticLayers: [{name: sl-one}]
`),
			Options{},
			[]string{"R1 ag-b: sl-one is already bound by ag-a"},
		},
		{
			"R1b no capability at all",
			agent(`metadata:
  name: ag-a
spec:
  managed: {}
`),
			Options{},
			[]string{"R1b ag-a: has neither semanticLayers nor toolsetNames"},
		},
		{
			"R4 allowedCubes",
			agent(`metadata:
  name: ag-a
spec:
  managed:
    semanticLayers:
      - name: sl-one
        allowedCubes: [orders]
`),
			Options{},
			[]string{"R4 ag-a: sl-one sets allowedCubes"},
		},
		{
			"R7 published without sample questions",
			agent(`metadata:
  name: ag-a
  labels:
    asgard-ai.com/agent-published: "true"
spec:
  managed:
    semanticLayers: [{name: sl-one}]
    sampleQuestions: ["only one"]
`),
			Options{},
			[]string{"R7 ag-a: published but has 1 sampleQuestions"},
		},
		{
			"R10 an OLAP-only layer bound to an agent",
			agent(`metadata:
  name: ag-a
spec:
  managed:
    semanticLayers: [{name: sl-lake}]
`),
			Options{OLAPOnlyLayers: []string{"sl-lake"}},
			[]string{"R10 ag-a: references sl-lake"},
		},
		{
			"R12 prompts that drifted apart",
			agent(`metadata:
  name: ag-a
spec:
  managed:
    toolsetNames: [ts-x]
    prompt:
      task: do the thing
      format: markdown
`) + "---\n" + agent(`metadata:
  name: ag-b
spec:
  managed:
    toolsetNames: [ts-x]
    prompt:
      task: do the thing differently
      format: markdown
`),
			Options{},
			[]string{"R12 prompt.task differs: 2 distinct values"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wants(t, AgentSplit(docs(t, tt.yaml), tt.opts), tt.wants...)
		})
	}
}

// TestAgentSplitPassesAValidPair is the negative control: without it, a rule
// that fires on everything would still make every test above pass.
func TestAgentSplitPassesAValidPair(t *testing.T) {
	r := AgentSplit(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: Agent
metadata:
  name: ag-a
  labels:
    asgard-ai.com/agent-published: "true"
spec:
  managed:
    semanticLayers: [{name: sl-one}]
    sampleQuestions: ["a", "b"]
    prompt:
      task: shared
      format: shared
---
apiVersion: asgard-ai.com/v1alpha1
kind: Agent
metadata:
  name: ag-b
spec:
  managed:
    semanticLayers: [{name: sl-two}]
    prompt:
      task: shared
      format: shared
`), Options{})
	if !r.OK() {
		t.Fatalf("a valid split should pass: %v", r.Problems)
	}
	if !strings.Contains(r.Summary, "2 agent(s) (1 published)") {
		t.Errorf("summary = %q", r.Summary)
	}
	if !strings.Contains(r.Summary, "sl-one, sl-two") {
		t.Errorf("summary = %q, want the layers listed", r.Summary)
	}
}

// TestXrefRejectsTwoCRsWithOneName covers the failure nothing else sees: helm
// renders both, apply reports success twice, and Kubernetes keeps only whichever
// landed last. The generator used to cause exactly this - several fixed query
// tools in one Toolset, each emitting its own copy of that Toolset, and the
// surviving copy carried one tool.
func TestXrefRejectsTwoCRsWithOneName(t *testing.T) {
	toolset := func(tool string) string {
		return `
apiVersion: asgard-ai.com/v1alpha1
kind: Toolset
metadata:
  name: ts-catalog
  annotations:
    asgard-ai.com/toolset-name: catalog
spec:
  tools:
    - name: ` + tool + `
`
	}

	r := Xref(docs(t, toolset("stock")+"---\n"+toolset("prices")), Options{Project: "site"})
	wants(t, r, "Toolset/ts-catalog is declared 2 times in this chart")

	if r := Xref(docs(t, toolset("stock")), Options{Project: "site"}); !r.OK() {
		t.Errorf("one Toolset should pass: %v", r.Problems)
	}
}

// TestXrefNamesTheCommandThatFixesIt is what makes the gate usable by an agent
// working from this output alone. Without it the reading is ambiguous, and the
// cheap resolution - deleting the reference - is the wrong one.
func TestXrefNamesTheCommandThatFixesIt(t *testing.T) {
	manifests := `
apiVersion: asgard-ai.com/v1alpha1
kind: Agent
metadata:
  name: ag-erp
  annotations:
    asgard-ai.com/agent-name: erp
spec:
  managed:
    semanticLayers:
      - name: sl-missing
`
	r := Xref(docs(t, manifests), Options{Project: "erp"})
	wants(t, r, "asgard-cli add semanticlayer missing --project erp")

	// Without a project - a stream rendered elsewhere - the command still reads,
	// with the one part that cannot be known left as a placeholder.
	r = Xref(docs(t, manifests), Options{})
	wants(t, r, "--project <project>")
}

// TestDeployabilityWarnsRatherThanFails is the design, not a detail: both of
// these conditions are correct during an onboarding and fatal once somebody
// tags. Failing on them would leave the gate red through the middle of every
// engagement, which teaches people to ignore it.
func TestDeployabilityWarnsRatherThanFails(t *testing.T) {
	// The shape a fresh chart has: a read path, an entry point, no skills.
	r := Deployability(docs(t, `
apiVersion: asgard-ai.com/v1alpha1
kind: Agent
metadata:
  name: ag-erp
`), Options{Project: "erp"})

	if !r.OK() {
		t.Errorf("this must not fail the gate: %v", r.Problems)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "no Syncer yet") {
		t.Fatalf("warnings = %v, want one about the missing Syncer", r.Warnings)
	}
	// The warning has to carry the command, and the reason it matters.
	if !strings.Contains(r.Warnings[0], "asgard-cli add skillset base --project erp") {
		t.Errorf("warning does not name the command: %q", r.Warnings[0])
	}
	if !strings.Contains(r.Warnings[0], "exits 1 after 180s") {
		t.Errorf("warning does not say what fails: %q", r.Warnings[0])
	}
}

func TestDeployabilityChecksTheEnvironmentIdLabel(t *testing.T) {
	// An empty platformMainEnvironmentId renders no label at all, deliberately,
	// which is exactly why nothing else sees it.
	withSyncer := `
apiVersion: asgard-ai.com/v1alpha1
kind: Syncer
metadata:
  name: syn-x
---
apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: wf-site
`
	r := Deployability(docs(t, withSyncer), Options{Project: "site"})
	if !r.OK() {
		t.Errorf("this must not fail: %v", r.Problems)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "Workflow/wf-site") {
		t.Fatalf("warnings = %v, want one naming the unlabelled Workflow", r.Warnings)
	}

	// With both in place there is nothing to say.
	labelled := strings.Replace(withSyncer, `  name: wf-site`,
		"  name: wf-site\n  labels:\n    asgard-ai.com/project-environment-id: env-1", 1)
	r = Deployability(docs(t, labelled), Options{Project: "site"})
	if len(r.Warnings) != 0 {
		t.Errorf("warnings = %v, want none", r.Warnings)
	}
	if !strings.Contains(r.Summary, "environment id set") {
		t.Errorf("summary = %q", r.Summary)
	}
}

func TestDeployabilityOfAnEmptyChart(t *testing.T) {
	// Nothing rendered is not a deployability problem; it is stage 3.
	r := Deployability(nil, Options{Project: "erp"})
	if !r.OK() || len(r.Warnings) != 0 {
		t.Errorf("empty chart = %+v, want nothing to report", r)
	}
}
