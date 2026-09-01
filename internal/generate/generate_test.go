package generate

import (
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/chart"
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
		Projects:  []config.Project{{Slug: "ops", Name: "ops", Environments: []config.Env{config.EnvDev}}},
	}
}

func repo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := scaffold.Write(dir, cfg(), false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	return dir
}

// generated writes one kind and returns the contents of every file it produced.
func generated(t *testing.T, dir, kindName string, opts Options) map[string]string {
	t.Helper()
	kind, ok := Find(kindName)
	if !ok {
		t.Fatalf("no kind %q", kindName)
	}
	opts.Project = "ops"

	results, err := Write(dir, cfg(), kind, opts)
	if err != nil {
		t.Fatalf("Write %s: %v", kindName, err)
	}

	// Only the CRs. A kind may also append to values.yaml, which the values
	// tests below read directly.
	out := map[string]string{}
	for _, r := range results {
		if r.Values {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, r.Path))
		if err != nil {
			t.Fatalf("ReadFile %s: %v", r.Path, err)
		}
		out[r.Path] = string(data)
	}
	return out
}

// TestEveryKindCarriesWhatFailsSilently is the reason this package exists. Each
// of these is invisible to helm lint, to CRD validation and to a server-side
// dry-run, so a hand-written CR loses them without any signal.
func TestEveryKindCarriesWhatFailsSilently(t *testing.T) {
	tests := []struct {
		kind string
		opts Options
		want map[string]string // substring -> why it matters
	}{
		{"dataconnector", Options{Name: "wms", DBClass: "mssql"}, map[string]string{
			"asgard-ai.com/data-connector-name:": "without it the CR is nameless in the UI",
			"secretKeyRef":                       "a password must never be a value",
			"dataConnectorClass: mssql":          "the class has to match the flag",
		}},
		{"semanticlayer", Options{Name: "wms", Connector: "dc-wms"}, map[string]string{
			"asgard-ai.com/semantic-layer-name:": "nameless in the UI otherwise",
			"completionModelName:":               "required on a SemanticLayer, unlike an Agent",
			"effort:":                            "omitting it is not the same as medium",
			"dataConnectorName: dc-wms":          "the connector has to be wired",
			"{CUBE}":                             "the dimension sql form",
		}},
		{"agent", Options{Name: "wms", Layer: "sl-wms"}, map[string]string{
			"asgard-ai.com/agent-name:":      "nameless in the UI otherwise",
			"asgard-ai.com/agent-published:": "the on/off switch for delegation",
			"aliasName: wms":                 "the delegation name",
			"allowWrite: false":              "source systems stay read-only",
			"byte-identical":                 "the constraint the gate enforces",
		}},
		{"httptool", Options{Name: "publish", Toolset: "ts-x", Write: true}, map[string]string{
			"requestConsent: true":             "a write must be gated",
			"type: update-context":             "parameters must be read before the call",
			"prevPayload is the HTTP response": "the trap that eats the arguments",
			"Content-Type":                     "headers are configs",
		}},
		{"querytool", Options{Name: "list", Connector: "dc-x", Toolset: "ts-x"}, map[string]string{
			"requestConsent: false": "a read needs no gate",
			`"properties": {}`:      "zero parameters is the security property",
		}},
		{"skillset", Options{Name: "base", Repo: "https://example.com/x.git"}, map[string]string{
			"asgard-ai.com/managed-by: skill-set": "needed on the SourceSet and the Syncer",
			"destinationPath:":                    "the current field name",
			"syncer-suspend":                      "CD triggers it, not the scheduler",
			"kind: SourceSet":                     "a SkillSet is a trio",
			"kind: Syncer":                        "a SkillSet is a trio",
		}},
		{"trigger", Options{Name: "daily", Layers: []string{"sl-a", "sl-b"}}, map[string]string{
			"asgard-ai.com/workflow-set-id:":        "or the editor is a blank canvas",
			"asgard-ai.com/project-environment-id:": "or the editor is a blank canvas",
			`"type" "trigger"`:                      "belongs to its Trigger, not to Automations",
			`{"name": "sl-a", "allowQuery": true}`:  "layers are a JSON array",
			"Do NOT write a BotProvider":            "the reconciler provisions it",
		}},
		{"flowagent", Options{Name: "shop", Public: true, Toolset: "ts-x"}, map[string]string{
			"authMode: none":         "an anonymous audience",
			"sandboxBlueprint":       "the capability link, one CR further out",
			"kind: BotProvider":      "three CRs",
			"kind: SandboxBlueprint": "three CRs",
			`"type" "bot"`:           "it appears under Flow Agents",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			files := generated(t, repo(t), tt.kind, tt.opts)
			all := strings.Join(values(files), "\n")
			for want, why := range tt.want {
				if !strings.Contains(all, want) {
					t.Errorf("missing %q - %s", want, why)
				}
			}
		})
	}
}

// TestHyphenatedNameIsAddressableInValues covers a bug found by helm, not by
// review: a name with a hyphen cannot be reached with dot notation, and the
// chart fails to parse.
func TestHyphenatedNameIsAddressableInValues(t *testing.T) {
	files := generated(t, repo(t), "trigger", Options{Name: "daily-report"})
	all := strings.Join(values(files), "\n")

	if strings.Contains(all, ".Values.triggers.daily-report") {
		t.Error("a hyphenated key is not reachable with dot notation; helm fails to parse the chart")
	}
	if !strings.Contains(all, ".Values.triggers.dailyReport") {
		t.Error("the values key should be camel case")
	}
	// The CR name keeps the hyphen; only the values key changes.
	if !strings.Contains(all, "name: tr-daily-report") {
		t.Error("the CR name should keep its hyphens")
	}
}

func TestWriteFlagFlipsTheGate(t *testing.T) {
	read := generated(t, repo(t), "httptool", Options{Name: "get", Toolset: "ts-x"})
	if !strings.Contains(strings.Join(values(read), ""), "requestConsent: false") {
		t.Error("a read tool should not be gated")
	}

	write := generated(t, repo(t), "httptool", Options{Name: "put", Toolset: "ts-x", Write: true})
	joined := strings.Join(values(write), "")
	if !strings.Contains(joined, "requestConsent: true") {
		t.Error("a write tool must be gated")
	}
	if !strings.Contains(joined, "must NOT be reachable from a Trigger") {
		t.Error("a gated tool has to warn that a scheduled run cannot approve")
	}
}

func TestFlowAgentWritesThreeFilesInOneDirectory(t *testing.T) {
	files := generated(t, repo(t), "flowagent", Options{Name: "shop", Public: true})
	if len(files) != 3 {
		t.Fatalf("wrote %d files, want 3: %v", len(files), keys(files))
	}
	for _, want := range []string{
		filepath.Join("projects", "ops", "chart", "app", "templates", "shop", "bot_provider.yaml"),
		filepath.Join("projects", "ops", "chart", "app", "templates", "shop", "workflow.yaml"),
		filepath.Join("projects", "ops", "chart", "app", "templates", "shop", "sandbox_blueprint.yaml"),
	} {
		if _, ok := files[want]; !ok {
			t.Errorf("did not write %s; got %v", want, keys(files))
		}
	}
}

func TestRefusesToClobber(t *testing.T) {
	dir := repo(t)
	generated(t, dir, "agent", Options{Name: "wms"})

	kind, _ := Find("agent")
	_, err := Write(dir, cfg(), kind, Options{Project: "ops", Name: "wms"})
	if err == nil {
		t.Fatal("want an error rather than overwriting")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error %q should mention --force", err)
	}
}

// TestMultiFileKindIsAllOrNothing: a partial write leaves a chain that renders
// but does not resolve, which is worse than not writing at all.
func TestMultiFileKindIsAllOrNothing(t *testing.T) {
	dir := repo(t)
	generated(t, dir, "flowagent", Options{Name: "shop"})

	// Remove one file, then try again without --force.
	victim := filepath.Join(dir, "projects", "ops", "chart", "app", "templates", "shop", "workflow.yaml")
	if err := os.Remove(victim); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	kind, _ := Find("flowagent")
	if _, err := Write(dir, cfg(), kind, Options{Project: "ops", Name: "shop"}); err == nil {
		t.Fatal("want a refusal: the other two files still exist")
	}
	if _, err := os.Stat(victim); !os.IsNotExist(err) {
		t.Error("the refused run should not have written anything")
	}
}

func TestRejectsUnknownProjectAndBadName(t *testing.T) {
	dir := repo(t)
	kind, _ := Find("agent")

	if _, err := Write(dir, cfg(), kind, Options{Project: "nope", Name: "x"}); err == nil {
		t.Error("want an error for a project that does not exist")
	}
	if _, err := Write(dir, cfg(), kind, Options{Project: "ops", Name: "Not_A_Slug"}); err == nil {
		t.Error("want an error for a name that cannot be a CR name")
	}
}

func TestEveryKindNamesItsExtract(t *testing.T) {
	// The decision comes before the YAML, so every kind has to point at the
	// shape that explains when to use it.
	for _, k := range Kinds {
		if k.Extract == "" {
			t.Errorf("%s names no extract", k.Name)
		}
		if len(k.Files) == 0 {
			t.Errorf("%s writes no files", k.Name)
		}
	}
}

func values(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestValuesAreDeclaredForWhatTheCRReads covers the gap the end-to-end run
// found: a generated CR that reads .Values.x while nothing declares x renders
// fine with an env file overlaid, and nil-pointers for anyone running plain
// helm template. The bare helm lint is the only step that catches it.
func TestValuesAreDeclaredForWhatTheCRReads(t *testing.T) {
	dir := repo(t)

	for _, tc := range []struct {
		kind string
		opts Options
		want []string
	}{
		{"dataconnector", Options{Name: "wms", DBClass: "mssql"}, []string{"wmsDB:", "port: 1433"}},
		{"semanticlayer", Options{Name: "wms", Connector: "dc-wms"}, []string{"defaultSemanticLayerEffort:"}},
		{"httptool", Options{Name: "pub", Toolset: "ts-w"}, []string{"pub:", "endpoint:"}},
		{"trigger", Options{Name: "daily-sales", Layers: []string{"sl-wms"}}, []string{"triggers:", "dailySales:"}},
		{"flowagent", Options{Name: "shop", Public: true}, []string{"botProviders:", "shop:"}},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			generated(t, dir, tc.kind, tc.opts)

			data, err := os.ReadFile(filepath.Join(dir, "projects", "ops", "chart", "app", "values.yaml"))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			for _, want := range tc.want {
				if !strings.Contains(string(data), want) {
					t.Errorf("values.yaml does not declare %q", want)
				}
			}
		})
	}
}

func TestValuesAreNotDuplicated(t *testing.T) {
	dir := repo(t)

	// Two connectors, then two more tools: the shared blocks must appear once.
	generated(t, dir, "dataconnector", Options{Name: "a", DBClass: "postgres"})
	generated(t, dir, "dataconnector", Options{Name: "b", DBClass: "postgres"})
	generated(t, dir, "semanticlayer", Options{Name: "a", Connector: "dc-a"})
	generated(t, dir, "semanticlayer", Options{Name: "b", Connector: "dc-b"})

	data, err := os.ReadFile(filepath.Join(dir, "projects", "ops", "chart", "app", "values.yaml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if n := strings.Count(string(data), "defaultSemanticLayerEffort:"); n != 1 {
		t.Errorf("defaultSemanticLayerEffort declared %d times, want 1", n)
	}
	// Per-CR blocks are distinct and both belong there.
	for _, want := range []string{"aDB:", "bDB:"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("missing %q", want)
		}
	}
}

// TestPrefixedNameIsNotPrefixedTwice pins the fix for a name typed in the form
// this CLI's own guidance uses everywhere else. "--connector dc-<name>" and
// "--layer sl-<name>" are prefixed, so `add dataconnector dc-erp` is what an
// FDE types; before this, it produced metadata.name: dc-dc-erp, which helm lint
// and a server-side dry-run both accept and only asgard-cli verify catches -
// and only once something references the name that was meant.
func TestPrefixedNameIsNotPrefixedTwice(t *testing.T) {
	for _, kindName := range []string{"dataconnector", "semanticlayer", "agent"} {
		t.Run(kindName, func(t *testing.T) {
			kind, ok := Find(kindName)
			if !ok {
				t.Fatalf("no kind %q", kindName)
			}
			bare := generated(t, repo(t), kindName, Options{Name: "erp", Connector: "dc-erp", Layer: "sl-erp", DBClass: "postgres"})
			pre := generated(t, repo(t), kindName, Options{Name: kind.Prefix + "erp", Connector: "dc-erp", Layer: "sl-erp", DBClass: "postgres"})

			for path, want := range bare {
				got, ok := pre[path]
				if !ok {
					t.Fatalf("%s: %q wrote %v, %q wrote %v", kindName, "erp", keys(bare), kind.Prefix+"erp", keys(pre))
				}
				if got != want {
					t.Errorf("%s: %s differs between the bare and the prefixed name", kindName, path)
				}
			}
			for _, body := range pre {
				if strings.Contains(body, kind.Prefix+kind.Prefix) {
					t.Errorf("%s: doubled prefix %q in output:\n%s", kindName, kind.Prefix+kind.Prefix, body)
				}
			}
		})
	}
}

// TestPrefixAloneIsRejected: trimming the prefix must not turn a typo into an
// unnamed CR.
func TestPrefixAloneIsRejected(t *testing.T) {
	kind, _ := Find("dataconnector")
	if _, err := Write(repo(t), cfg(), kind, Options{Project: "ops", Name: kind.Prefix, DBClass: "postgres"}); err == nil {
		t.Fatalf("Write with name %q: want an error, got none", kind.Prefix)
	}
}

// TestResolveWiresTheSkeletonToWhatExists is the defect this closes: the
// generator wrote skeletons that could not pass the gate. An Agent hardcoded a
// SkillSet nothing creates and ignored the SemanticLayer in the same chart, so
// an agent that followed the CLI's own instructions ended with two dangling
// references and no idea why.
func TestResolveWiresTheSkeletonToWhatExists(t *testing.T) {
	root := t.TempDir()
	cfg := cfg()
	kind, _ := Find("agent")

	// With nothing else in the chart, an Agent mounts nothing: it is allowed to
	// read through Toolsets instead, and referencing a layer that is not there
	// would be the very problem this avoids.
	opts, notes, err := Resolve(root, kind, Options{Project: "ops", Name: "erp"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if opts.Layer != "" || len(notes) != 0 {
		t.Errorf("empty chart: layer=%q notes=%v, want neither", opts.Layer, notes)
	}

	// One layer is the documented shape - one system, one layer, one Agent - so
	// the choice is forced and taking it is not a guess.
	slKind, _ := Find("semanticlayer")
	if _, err := Write(root, cfg, slKind, Options{Project: "ops", Name: "erp", Connector: "dc-erp"}); err != nil {
		t.Fatalf("Write semanticlayer: %v", err)
	}
	opts, notes, err = Resolve(root, kind, Options{Project: "ops", Name: "erp"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if opts.Layer != "sl-erp" {
		t.Errorf("layer = %q, want sl-erp", opts.Layer)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "sl-erp") {
		t.Errorf("notes = %v, want one saying what was mounted", notes)
	}

	// Two layers is a decision, so it is refused rather than picked.
	if _, err := Write(root, cfg, slKind, Options{Project: "ops", Name: "crm", Connector: "dc-crm"}); err != nil {
		t.Fatalf("Write second semanticlayer: %v", err)
	}
	if _, _, err = Resolve(root, kind, Options{Project: "ops", Name: "erp"}); err == nil {
		t.Error("two layers should be refused, not picked between")
	} else if !strings.Contains(err.Error(), "--layer") {
		t.Errorf("error %q should name the flag that decides it", err)
	}
}

// TestResolveDoesNotReEmitAnExistingToolset covers the silent one: several fixed
// query tools belong to one Toolset, and each writing its own copy renders two
// CRs with one name. Whichever applies last replaces the other and takes its
// tools with it, while helm and apply both report success.
func TestResolveDoesNotReEmitAnExistingToolset(t *testing.T) {
	root := t.TempDir()
	cfg := cfg()
	kind, _ := Find("querytool")

	base := Options{Project: "ops", Connector: "dc-site", Toolset: "ts-catalog"}

	first, notes, err := Resolve(root, kind, withName(base, "stock"))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if first.ToolsetExists || len(notes) != 0 {
		t.Fatalf("the first tool writes the set: exists=%v notes=%v", first.ToolsetExists, notes)
	}
	if _, err := Write(root, cfg, kind, first); err != nil {
		t.Fatalf("Write: %v", err)
	}

	second, notes, err := Resolve(root, kind, withName(base, "prices"))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !second.ToolsetExists {
		t.Error("the second tool must not write the set again")
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "ts-catalog") {
		t.Errorf("notes = %v, want one naming the set", notes)
	}

	if _, err := Write(root, cfg, kind, second); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// One Toolset across the whole chart, which is the invariant.
	refs, err := chart.Scan(root, "ops")
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got := chart.Counts(refs)["Toolset"]; got != 1 {
		t.Errorf("chart declares %d Toolsets, want 1", got)
	}
}

func withName(o Options, name string) Options {
	o.Name = name
	return o
}

// TestBotClassWritesTheChannelsCredentials covers the gap that existed until a
// customer wanted LINE: every extract and this template hardcoded
// `botProviderClass: generic`, so the CLI had nothing at all for the channel a
// Taiwanese customer is most likely to ask for.
func TestBotClassWritesTheChannelsCredentials(t *testing.T) {
	kind, _ := Find("flowagent")

	tests := []struct {
		class  string
		want   []string
		absent []string
	}{
		{
			"generic",
			[]string{"botProviderClass: generic", "authMode:"},
			[]string{"line:", "telegram:"},
		},
		{
			"line",
			[]string{"botProviderClass: line", "line_channel_access_token", "line_channel_secret",
				"CANNOT HOST TWO BOTS", "immutable after creation"},
			// A chat platform owns its own presentation in its own console, so
			// the widget appearance annotation has no business here. The prose
			// does mention embedConfig, to say where presentation went.
			[]string{"additional-annotation", "authMode:"},
		},
		{
			"telegram",
			[]string{"botProviderClass: telegram", "telegram_bot_token", "telegram_webhook_secret",
				"inbound webhook"},
			[]string{"Connector Deployment", "line:"},
		},
		{
			"slack",
			[]string{"botProviderClass: slack", "slack_app_token", "slack_bot_token",
				"outbound WebSocket", "Connector"},
			[]string{"line:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			root := t.TempDir()
			opts, notes, err := Resolve(root, kind, Options{
				Project: "ops", Name: "support", BotClass: tt.class, Public: true,
			})
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if opts.BotClass != tt.class {
				t.Fatalf("class = %q", opts.BotClass)
			}
			// Anything but generic costs the engagement something invisible in
			// the YAML, so choosing one has to say so.
			if tt.class != "generic" && len(notes) == 0 {
				t.Errorf("choosing %s should report what it costs", tt.class)
			}

			results, err := Write(root, cfg(), kind, opts)
			if err != nil {
				t.Fatalf("Write: %v", err)
			}

			var body string
			for _, r := range results {
				if strings.Contains(r.Path, "bot_provider") {
					data, err := os.ReadFile(filepath.Join(root, r.Path))
					if err != nil {
						t.Fatalf("ReadFile: %v", err)
					}
					body = string(data)
				}
			}
			if body == "" {
				t.Fatal("no bot_provider file was written")
			}

			for _, want := range tt.want {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q:\n%s", want, body)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(body, absent) {
					t.Errorf("should not contain %q:\n%s", absent, body)
				}
			}
		})
	}

	// An unknown class is refused rather than silently written into the CR,
	// where it would be immutable.
	if _, _, err := Resolve(t.TempDir(), kind, Options{Project: "ops", Name: "x", BotClass: "whatsapp"}); err == nil {
		t.Error("an unknown bot class should be refused")
	}
}

// TestASecondCRSharingAValuesBlockGetsItsEntry covers the bug bare `helm lint`
// exists to catch: `botProviders:` exists after the first flow agent, so the
// second one's whole values snippet was skipped as "already there", and its
// template then read `.Values.botProviders.<name>.disabled` off a map with no
// such entry. The rendered chart was fine with an env file overlaid, and nil
// pointered for anyone running plain `helm template`.
func TestASecondCRSharingAValuesBlockGetsItsEntry(t *testing.T) {
	root := repo(t)
	kind, _ := Find("flowagent")

	for _, name := range []string{"first", "second", "third"} {
		opts, _, err := Resolve(root, kind, Options{Project: "ops", Name: name})
		if err != nil {
			t.Fatalf("Resolve %s: %v", name, err)
		}
		if _, err := Write(root, cfg(), kind, opts); err != nil {
			t.Fatalf("Write %s: %v", name, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, "projects", "ops", "chart", "app", "values.yaml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	values := string(data)

	// One block, three entries.
	if got := strings.Count(values, "\nbotProviders:"); got != 1 {
		t.Errorf("botProviders appears %d times, want 1", got)
	}
	for _, name := range []string{"first", "second", "third"} {
		if !strings.Contains(values, "\n  "+name+":") {
			t.Errorf("no entry for %s:\n%s", name, values)
		}
	}
}

// TestAValuesBlockIsNotDuplicated is the other half: a snippet whose top-level
// key is its own still gets appended once, and re-adding the same CR does not
// append it twice.
func TestAValuesBlockIsNotDuplicated(t *testing.T) {
	root := repo(t)
	kind, _ := Find("dataconnector")

	opts := Options{Project: "ops", Name: "erp", DBClass: "postgres"}
	if _, err := Write(root, cfg(), kind, opts); err != nil {
		t.Fatalf("Write: %v", err)
	}
	opts.Force = true
	if _, err := Write(root, cfg(), kind, opts); err != nil {
		t.Fatalf("Write again: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "projects", "ops", "chart", "app", "values.yaml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got := strings.Count(string(data), "\nerpDB:"); got != 1 {
		t.Errorf("erpDB appears %d times, want 1", got)
	}
}
