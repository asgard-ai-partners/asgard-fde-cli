package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
	"gopkg.in/yaml.v3"
)

func init() {
	register("spec-key-gap", check{
		Needs: "the clones, helm and a built binary",
		What:  "**how much of a production chart `add` never writes** - the number behind \"the chart half is the least finished\", rendered on both sides rather than quoted. --missing lists the keys",
		Run:   runSpecKeyGap,
	})
}

// The kinds `add` writes, and the flags each needs to write anything. A kind
// that grows a required flag and is not listed here would silently shrink the
// `add` side, which would make the gap look wider than it is.
var addKinds = []struct {
	name  string
	flags []string
}{
	{"dataconnector", nil},
	{"semanticlayer", []string{"--connector", "dc-db"}},
	{"agent", nil},
	{"httptool", []string{"--toolset", "ts-http"}},
	{"querytool", []string{"--connector", "dc-db", "--toolset", "ts-query"}},
	{"skillset", []string{"--repo", "https://github.com/example/skills"}},
	{"trigger", nil},
	{"knowledgedrive", []string{"--connector", "dc-db"}},
	{"plugin", []string{"--connector", "ss-skills"}},
	{"flowagent", nil},
}

// Values the platform injects per run, which helm cannot know. `asgard-cli
// render` supplies them; this renders with helm directly so that both sides are
// rendered the same way, so they are supplied here.
const injected = `asgard:
  projectEnvironmentId: probe-env
  projectId: probe-project
  appSecretName: probe-secret
  namespace: probe-ns
`

var specKeyClaim = regexp.MustCompile(`(\d+) spec keys and ` + "`" + `add` + "`" + ` never mentions (\d+) of them`)

// specKeys returns every dotted key path under `spec`, with list indices
// collapsed.
//
// Collapsing indices is what makes the two sides comparable: `processors.0` and
// `processors.7` are the same key, and counting them apart would say the gap
// closes as a chart grows.
func specKeys(text string) map[string]bool {
	out := map[string]bool{}
	var walk func(node any, path string)
	walk = func(node any, path string) {
		switch t := node.(type) {
		case map[string]any:
			for k, v := range t {
				child := k
				if path != "" {
					child = path + "." + k
				}
				out[child] = true
				walk(v, child)
			}
		case []any:
			for _, e := range t {
				walk(e, path)
			}
		}
	}
	// **One document at a time, and a failure skips only that one.** A
	// streaming decoder stops at the first document it cannot parse, which
	// silently abandons every document after it - here that lost two keys from
	// a chart whose later documents were fine, and the gap looked smaller than
	// it is.
	for _, doc := range strings.Split(text, "\n---") {
		var parsed map[string]any
		if err := yaml.Unmarshal([]byte(doc), &parsed); err != nil {
			continue
		}
		if spec, ok := parsed["spec"]; ok {
			walk(spec, "")
		}
	}
	return out
}

func helmRender(chart string, values []string) string {
	args := []string{"template", "probe", chart}
	for _, v := range values {
		args = append(args, "-f", v)
	}
	out, _ := exec.Command("helm", args...).Output()
	return string(out)
}

// production returns each reference chart's spec keys.
//
// **Only the deployments `source/SOURCES.md` declares.** `$ASGARD_DEPLOYMENTS`
// is somebody's projects directory: it also holds scratch repositories this
// tool scaffolded, and those pass by construction - one showed 0 keys not
// written, which would make the gap look like it had closed.
func production(base, root string) (map[string]map[string]bool, error) {
	doc, err := sourcesDoc(root)
	if err != nil {
		return nil, err
	}
	known := toSet(referenceDeployments(doc))
	out := map[string]map[string]bool{}
	_ = filepath.Walk(base, func(p string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() || filepath.Base(p) != "app" {
			return nil
		}
		if filepath.Base(filepath.Dir(p)) != "chart" {
			return nil
		}
		rel, err := filepath.Rel(base, p)
		if err != nil || !known[strings.Split(rel, string(filepath.Separator))[0]] {
			return nil
		}
		holder := filepath.Dir(p)
		var values []string
		for _, n := range []string{"values-prod.yaml", "values-dev.yaml"} {
			if _, err := os.Stat(filepath.Join(holder, n)); err == nil {
				values = append(values, filepath.Join(holder, n))
				break
			}
		}
		if len(values) == 0 {
			return nil
		}
		text := helmRender(p, values)
		if strings.TrimSpace(text) == "" {
			return nil
		}
		label, _ := filepath.Rel(base, holder)
		out[label] = specKeys(text)
		return nil
	})
	return out, nil
}

// ours is what `add` writes: one project with one of every kind in it.
func ours(binary string) (map[string]bool, error) {
	tmp, err := os.MkdirTemp("", "asgard-keys-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	if err := exec.Command("git", "-C", tmp, "init", "-q", ".").Run(); err != nil {
		return nil, err
	}
	for _, argv := range [][]string{{"init"}, {"project", "add", "site"}} {
		if _, stderr, err := runCLI(binary, tmp, argv...); err != nil {
			return nil, fmt.Errorf("`%s` failed in a scratch repository:\n%s",
				strings.Join(argv, " "), stderr)
		}
	}
	for _, k := range addKinds {
		argv := append([]string{"add", k.name, "probe-" + k.name, "--project", "site"}, k.flags...)
		if _, stderr, err := runCLI(binary, tmp, argv...); err != nil {
			return nil, fmt.Errorf("`add %s` failed, so the `add` side would be short a kind:\n%s\n"+
				"If it grew a required flag, add it to addKinds", k.name, trim(stderr, 400))
		}
	}
	vals := filepath.Join(tmp, "injected.yaml")
	if err := os.WriteFile(vals, []byte(injected), 0o644); err != nil {
		return nil, err
	}
	text := helmRender(filepath.Join(tmp, "projects/site/chart/app"), []string{vals})
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("the scaffolded chart rendered empty, so there is nothing to compare")
	}
	return specKeys(text), nil
}

func runSpecKeyGap(args []string) error {
	for _, tool := range []string{"helm"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s is not on PATH, and both sides have to be rendered", tool)
		}
	}
	root, err := src.Root()
	if err != nil {
		return err
	}
	base, err := src.Resolve("deployments")
	if err != nil {
		return err
	}
	binary := os.Getenv("ASGARD_CLI")
	if binary == "" {
		binary = filepath.Join(root, ".out/asgard-cli")
	}
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("no binary at %s\n  go build -o .out/asgard-cli ./cmd/asgard-cli\n"+
			"  Set $ASGARD_CLI to use another one.", binary)
	}
	binary, _ = filepath.Abs(binary)

	prod, err := production(base, root)
	if err != nil {
		return err
	}
	if len(prod) == 0 {
		return fmt.Errorf("no reference chart under %s rendered, so there is nothing to measure", base)
	}
	mine, err := ours(binary)
	if err != nil {
		return err
	}

	every := map[string]bool{}
	widest, widestN := "", -1
	for label, keys := range prod {
		for k := range keys {
			every[k] = true
		}
		if len(keys) > widestN {
			widest, widestN = label, len(keys)
		}
	}

	if len(args) > 0 && args[0] == "--missing" {
		var missing []string
		for k := range every {
			if !mine[k] {
				missing = append(missing, k)
			}
		}
		sort.Strings(missing)
		for _, k := range missing {
			fmt.Printf("  %s\n", k)
		}
		fmt.Printf("\n%d key(s) production uses and `add` never writes.\n", len(missing))
		return nil
	}

	labels := sortedKeys(prod)
	// Widest first, and ties by name so the order is the same on every run.
	sort.Slice(labels, func(i, j int) bool {
		if len(prod[labels[i]]) != len(prod[labels[j]]) {
			return len(prod[labels[i]]) > len(prod[labels[j]])
		}
		return labels[i] < labels[j]
	})
	for _, label := range labels {
		fmt.Printf("  %-44s%4d spec key(s), %4d not written\n",
			label, len(prod[label]), countMissing(prod[label], mine))
	}
	fmt.Printf("\n  %-44s%4d spec key(s), %4d not written\n",
		"every reference chart together", len(every), countMissing(every, mine))
	fmt.Printf("  %-44s%4d spec key(s)\n", "what `add` writes", len(mine))

	task, err := os.ReadFile(filepath.Join(root, "TASK.md"))
	if err != nil {
		return err
	}
	m := specKeyClaim.FindStringSubmatch(string(task))
	if m == nil {
		fmt.Println("\nTASK.md states no `<n> spec keys and `add` never mentions <n> of them` claim,")
		fmt.Println("so this measures and checks nothing. Either restore the claim or delete this.")
		return errFailed
	}
	wantKeys, wantMissing := len(prod[widest]), countMissing(prod[widest], mine)
	if atoi(m[1]) != wantKeys || atoi(m[2]) != wantMissing {
		fmt.Printf("\nTASK.md says %s spec keys and %s never mentioned;\n", m[1], m[2])
		fmt.Printf("the widest reference chart has %d and %d.\n", wantKeys, wantMissing)
		return errFailed
	}
	fmt.Printf("\nTASK.md's claim matches the widest chart: %d keys, %d never written.\n",
		wantKeys, wantMissing)
	return nil
}

func countMissing(from, have map[string]bool) int {
	n := 0
	for k := range from {
		if !have[k] {
			n++
		}
	}
	return n
}
