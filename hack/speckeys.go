package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template/parse"

	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
	"gopkg.in/yaml.v3"
)

func init() {
	register("spec-key-gap", check{
		Needs: "the clones, helm and a built binary",
		What:  "**how much of a production chart `add` never writes** - the number behind \"the chart half is the least finished\", rendered on both sides rather than quoted, with `add` run across the flag combinations it supports rather than one per kind. --missing lists the keys",
		Run:   runSpecKeyGap,
	})
}

// **What `add` can write is not one run per kind.** Several templates branch on
// a flag - `<<if .Private>>`, `<<if .Supervisor>>`, the `--db-class` blocks
// - and a key that exists only inside one of those branches is still a key
// `add` writes. One combination per kind counted every one of them as never
// written: 48 of the 165 keys this check reported as missing were already
// generated, which is the expensive direction to be wrong in, because the
// number is what somebody reads before implementing one of them.
//
// So the combinations are **derived from the generator**, in three steps, and
// nothing below lists a flag that a new branch would have to be added to:
//
//  1. every flag `add` registers, and the Options field each one sets, read
//     off the cobra calls in internal/cli/add.go;
//  2. per kind, the Options fields its own templates **branch** on, taken from
//     the template's parse tree. A field that is only interpolated -
//     `<<.DisplayName>>` - changes a value and never a key path, so it is not a
//     combination;
//  3. for a flag whose vocabulary is closed, the vocabulary itself, read from
//     `generate` - so a DataConnector class added upstream is probed here
//     without an edit.
//
// **A flag nobody would type in a real chart still counts as written.** What
// this check measures is what `add` can produce, not what an FDE usually asks
// for: a key behind `--db-class netsuite` is one an FDE gets by typing that, and
// listing it as missing would be telling somebody to implement what exists.

// requiredArgs is what `add` refuses to run at all without, so that a variant
// meant to exercise the other branch of a flag fails loudly rather than
// quietly. Two sources: `requiredFlags` in internal/cli/add.go, and `Resolve`
// in internal/generate, which is what rejects a plugin with no store.
//
// Everything else a kind takes is enumerated, so this names only what would
// otherwise turn a variant into an error.
var requiredArgs = map[string][]string{
	"semanticlayer": {"connector"},
	"httptool":      {"toolset"},
	"querytool":     {"connector", "toolset"},
	"skillset":      {"repo"},
	"plugin":        {"connector"},
}

// sampleValue is what to pass to a flag that takes one. The value decides which
// name appears inside a key's value and never which keys are written, so one
// placeholder per flag covers every kind that takes it - a plugin's store is a
// `ss-` SourceSet rather than a `dc-` DataConnector, and the rendered key paths
// are identical either way.
var sampleValue = map[string]string{
	"connector": "dc-probe",
	"layer":     "sl-probe",
	"layers":    "sl-probe",
	"toolset":   "ts-probe",
	"repo":      "https://github.com/example/skills",
}

// enumDomain is the closed vocabulary of a flag, where it has one. The values
// are the generator's own; only the two flag names are written here, because
// nothing in the generator declares which of its flags are enumerated.
func enumDomain(flag string) []string {
	switch flag {
	case "db-class":
		return generate.DBClasses()
	case "bot-class":
		return generate.BotClasses
	}
	return nil
}

// structureField is the exception to step 2 above, and the reason it is
// written down rather than derived: a field that is only interpolated normally
// changes a value inside a key, but `<<.DBSpec>>` interpolates **a whole spec
// block**, so each `--db-class` value is a different set of keys with no branch
// anywhere in the template. Nothing in a parse tree distinguishes a
// field that carries a scalar from one that carries YAML.
//
// `DBNote` is the same flag's note, and is a branch; naming it here keeps both
// halves of the class attributed to the flag that decides them.
var structureField = map[string]string{
	"DBSpec": "db-class",
	"DBNote": "db-class",
}

// stateSeed names the template fields that are **what the chart already
// contains** rather than anything typed, and what has to be in the chart for
// each to be true. A kind that branches on one is probed twice: once in an
// empty project and once in a project seeded with these.
//
// Both branches matter. An Agent in a chart with no SkillSet writes a comment
// where one with a SkillSet writes `managed.skillSetNames`, and a chart is only
// ever in one of those two states at a time.
var stateSeed = map[string][]string{
	"SkillSets":     {"add", "skillset", "seed", "--repo", "https://github.com/example/skills"},
	"ToolsetExists": {"add", "httptool", "seedtool", "--toolset", "ts-probe"},
}

// flagRegistration reads a cobra flag registration: the variable it binds, and
// the flag name. `--layers` binds a local rather than an Options field, which is
// why the `opts.` is optional and the match is made case-insensitively against
// the field a template names.
var flagRegistration = regexp.MustCompile(`cmd\.Flags\(\)\.(String|StringSlice|Bool)Var\(&(?:opts\.)?(\w+), "([\w-]+)"`)

type addFlag struct {
	name   string // as typed, without the dashes
	isBool bool
}

// addFlags maps an Options field, lowercased, to the flag that sets it.
func addFlags(root string) (map[string]addFlag, error) {
	path := filepath.Join(root, "internal/cli/add.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]addFlag{}
	for _, m := range flagRegistration.FindAllStringSubmatch(string(body), -1) {
		out[strings.ToLower(m[2])] = addFlag{name: m[3], isBool: m[1] == "Bool"}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no flag registration found in %s, so every combination below would be the default one", path)
	}
	return out, nil
}

// fieldsUsed returns the fields a template branches on - the pipelines of its
// if, range and with actions - and, separately, every field it names at all.
// The first is what changes which keys are written; the second is only read for
// structureField below.
//
// It reads the parse tree rather than matching text, because `<<if eq .BotClass
// "generic">>` and `<<if .Write>>POST<<else>>GET<<end>>` are both branches and
// neither is the shape a regular expression over `<<if .X>>` catches.
func fieldsUsed(path string) (branch, all map[string]bool, err error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	tree := parse.New(filepath.Base(path))
	// The generator registers its own functions; SkipFuncCheck means a
	// function added there is not a parse error here.
	tree.Mode = parse.SkipFuncCheck
	if _, err := tree.Parse(string(body), "<<", ">>", map[string]*parse.Tree{}); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", path, err)
	}
	branch, all = map[string]bool{}, map[string]bool{}
	var pipe func(*parse.PipeNode, map[string]bool)
	pipe = func(p *parse.PipeNode, into map[string]bool) {
		if p == nil {
			return
		}
		for _, cmd := range p.Cmds {
			for _, arg := range cmd.Args {
				switch a := arg.(type) {
				case *parse.FieldNode:
					if len(a.Ident) > 0 {
						into[a.Ident[0]] = true
						all[a.Ident[0]] = true
					}
				case *parse.PipeNode:
					pipe(a, into)
				}
			}
		}
	}
	var walk func(parse.Node)
	walk = func(n parse.Node) {
		switch t := n.(type) {
		case *parse.ListNode:
			if t == nil {
				return
			}
			for _, c := range t.Nodes {
				walk(c)
			}
		case *parse.ActionNode:
			pipe(t.Pipe, all)
		case *parse.IfNode:
			pipe(t.Pipe, branch)
			walk(t.List)
			walk(t.ElseList)
		case *parse.RangeNode:
			pipe(t.Pipe, branch)
			walk(t.List)
			walk(t.ElseList)
		case *parse.WithNode:
			pipe(t.Pipe, branch)
			walk(t.List)
			walk(t.ElseList)
		}
	}
	walk(tree.Root)
	return branch, all, nil
}

// probe is one scratch project: what to create in it before anything else, and
// the flag combinations to run in it.
//
// One project per kind rather than one for everything, because the chart's own
// contents decide two of the branches and one kind's leftovers would settle
// another kind's: an Agent added to a chart holding one SemanticLayer mounts it
// without being asked, and `add agent` refuses outright once there are two.
type probe struct {
	kind  string
	slug  string
	seeds [][]string // whole `add` argv, minus --project
	runs  [][]string // flag arguments, one entry per run
}

// addMatrix plans every run, from the generator rather than from a list.
func addMatrix(root string) ([]probe, error) {
	flags, err := addFlags(root)
	if err != nil {
		return nil, err
	}
	var out []probe
	for _, k := range generate.Kinds {
		branch, used := map[string]bool{}, map[string]bool{}
		for _, f := range k.Files {
			b, a, err := fieldsUsed(filepath.Join(root, "internal/generate/templates", f.Template))
			if err != nil {
				return nil, err
			}
			for name := range b {
				branch[name] = true
			}
			for name := range a {
				used[name] = true
			}
		}

		required := toSet(requiredArgs[k.Name])
		var base []string
		for _, name := range requiredArgs[k.Name] {
			arg, err := flagArgs(addFlag{name: name}, k.Name)
			if err != nil {
				return nil, err
			}
			base = append(base, arg...)
		}

		// What to vary: the flags this kind's own templates react to.
		combine := map[string]bool{}
		var seeds [][]string
		for _, field := range sortedKeys(used) {
			if flag, ok := structureField[field]; ok {
				combine[flag] = true
				continue
			}
			// A field that is only interpolated puts a name inside a value and
			// leaves the key paths alone, so it is not worth a run.
			if !branch[field] {
				continue
			}
			if seed, ok := stateSeed[field]; ok {
				seeds = append(seeds, seed)
				continue
			}
			f, ok := flags[strings.ToLower(field)]
			if !ok {
				// **The one thing that would quietly shrink this side again.**
				// A template that grows a branch on something no flag sets, and
				// that is not chart state, is a branch no run here ever takes.
				return nil, fmt.Errorf("a %s template branches on .%s, which no `add` flag sets\n"+
					"and which is neither chart state nor a structure block. Add it to stateSeed\n"+
					"or structureField in hack/speckeys.go, or the keys behind that branch will be\n"+
					"reported as never written", k.Name, field)
			}
			combine[f.name] = true
		}

		var toggles, refs, enums [][]string
		for _, name := range sortedKeys(combine) {
			if required[name] {
				continue
			}
			if domain := enumDomain(name); domain != nil {
				for _, v := range domain {
					enums = append(enums, []string{"--" + name, v})
				}
				continue
			}
			f, ok := flagByName(flags, name)
			if !ok {
				return nil, fmt.Errorf("%s needs --%s and `add` no longer registers it", k.Name, name)
			}
			arg, err := flagArgs(f, k.Name)
			if err != nil {
				return nil, err
			}
			if f.isBool {
				toggles = append(toggles, arg)
			} else {
				refs = append(refs, arg)
			}
		}

		// Not the cartesian product: the base, each optional flag on its own,
		// and all of them at once, once per value of an enumerated flag. That
		// covers every key behind one flag and every key behind the whole set;
		// a key needing exactly two of them and not the rest would be missed,
		// and no template has one.
		var all []string
		for _, a := range append(append([][]string{}, toggles...), refs...) {
			all = append(all, a...)
		}
		var runs [][]string
		add := func(args ...[]string) {
			var run []string
			run = append(run, base...)
			for _, a := range args {
				run = append(run, a...)
			}
			runs = append(runs, run)
		}
		add()
		for _, e := range enums {
			add(e)
			if len(all) > 0 {
				add(e, all)
			}
		}
		if len(enums) == 0 && len(all) > 0 {
			add(all)
		}
		for _, a := range append(append([][]string{}, toggles...), refs...) {
			add(a)
		}
		runs = dedupeRuns(runs)

		out = append(out, probe{kind: k.Name, slug: k.Name, runs: runs})
		if len(seeds) > 0 {
			out = append(out, probe{kind: k.Name, slug: k.Name + "-seeded", seeds: seeds, runs: runs})
		}
	}
	return out, nil
}

// flagByName finds a registration by the flag as typed.
func flagByName(flags map[string]addFlag, name string) (addFlag, bool) {
	for _, f := range flags {
		if f.name == name {
			return f, true
		}
	}
	return addFlag{}, false
}

// flagArgs is how one flag is typed, with its placeholder where it takes one.
func flagArgs(f addFlag, kind string) ([]string, error) {
	if f.isBool {
		return []string{"--" + f.name}, nil
	}
	v, ok := sampleValue[f.name]
	if !ok {
		return nil, fmt.Errorf("--%s takes a value and hack/speckeys.go has no placeholder for it,\n"+
			"so %s cannot be probed with it", f.name, kind)
	}
	return []string{"--" + f.name, v}, nil
}

func dedupeRuns(runs [][]string) [][]string {
	seen := map[string]bool{}
	var out [][]string
	for _, r := range runs {
		k := strings.Join(r, "\x00")
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, r)
	}
	return out
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

// **The claim, not the count.** TASK.md used to carry the four numbers and this
// held them digit for digit - so every change to what `add` writes turned the
// check red and the repair was to retype a number a script had just computed.
// That is the shape this repository removed everywhere else: a number a script
// can derive does not belong in prose.
//
// What TASK.md states is the judgement - that the chart half is the least
// finished, because `add` writes a starting point rather than a chart - and
// what this holds is whether that is still true. The numbers are printed by the
// run above, where they cannot go stale.
var specKeyClaim = regexp.MustCompile(`the least finished`)

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

// ours is what `add` writes, across every flag combination it supports.
//
// It returns the union of the spec keys and how many `add` runs produced them,
// because the count is what says this is a matrix rather than one run per kind -
// the shape that under-reported the `add` side.
//
// Offline by construction: `init`, `project add` and `add` write files, and
// helm renders them. Nothing here reaches a platform, an account or a network.
func ours(binary, root string) (map[string]bool, int, error) {
	matrix, err := addMatrix(root)
	if err != nil {
		return nil, 0, err
	}
	tmp, err := os.MkdirTemp("", "asgard-keys-")
	if err != nil {
		return nil, 0, err
	}
	defer os.RemoveAll(tmp)
	if err := exec.Command("git", "-C", tmp, "init", "-q", ".").Run(); err != nil {
		return nil, 0, err
	}
	if _, stderr, err := runCLI(binary, tmp, "init"); err != nil {
		return nil, 0, fmt.Errorf("`init` failed in a scratch repository:\n%s", stderr)
	}
	vals := filepath.Join(tmp, "injected.yaml")
	if err := os.WriteFile(vals, []byte(injected), 0o644); err != nil {
		return nil, 0, err
	}

	keys := map[string]bool{}
	runs := 0
	for _, p := range matrix {
		if _, stderr, err := runCLI(binary, tmp, "project", "add", p.slug); err != nil {
			return nil, 0, fmt.Errorf("`project add %s` failed:\n%s", p.slug, stderr)
		}
		for _, seed := range p.seeds {
			argv := append(append([]string{}, seed...), "--project", p.slug)
			if _, stderr, err := runCLI(binary, tmp, argv...); err != nil {
				return nil, 0, fmt.Errorf("seeding %s with `%s` failed, so the branch it exists to\n"+
					"reach would be reported as never written:\n%s",
					p.slug, strings.Join(argv, " "), trim(stderr, 400))
			}
		}
		for i, flags := range p.runs {
			name := fmt.Sprintf("%s-%d", p.kind, i)
			argv := append([]string{"add", p.kind, name, "--project", p.slug}, flags...)
			if _, stderr, err := runCLI(binary, tmp, argv...); err != nil {
				return nil, 0, fmt.Errorf("`%s` failed, so the `add` side would be short whatever\n"+
					"that combination writes:\n%s\n"+
					"If the kind grew a flag it refuses to run without, add it to requiredArgs",
					strings.Join(argv, " "), trim(stderr, 400))
			}
			runs++
		}
		text := helmRender(filepath.Join(tmp, "projects", p.slug, "chart/app"), []string{vals})
		if strings.TrimSpace(text) == "" {
			return nil, 0, fmt.Errorf("the scaffolded chart for %s rendered empty, so its keys are missing\n"+
				"from the `add` side and the gap would look wider than it is", p.slug)
		}
		for k := range specKeys(text) {
			keys[k] = true
		}
	}
	return keys, runs, nil
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
	mine, runs, err := ours(binary, root)
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
	fmt.Printf("  %-44s%4d spec key(s), from %3d `add` run(s)\n", "what `add` writes", len(mine), runs)

	task, err := os.ReadFile(filepath.Join(root, "TASK.md"))
	if err != nil {
		return err
	}
	m := specKeyClaim.FindStringSubmatch(string(task))
	if m == nil {
		fmt.Println("\nTASK.md no longer says the chart half is the least finished, so this")
		fmt.Println("measures nothing. Either the claim is back, or this check goes with it.")
		return errFailed
	}
	wantKeys, wantMissing := len(prod[widest]), countMissing(prod[widest], mine)
	if wantMissing == 0 {
		fmt.Printf("\n`add` now writes every one of the widest chart's %d spec keys.\n", wantKeys)
		fmt.Println("TASK.md still says the chart half is the least finished, and that is what is")
		fmt.Println("wrong now - rewrite the paragraph rather than this check.")
		return errFailed
	}
	fmt.Printf("\nThe widest reference chart uses %d spec keys and `add` never mentions %d.\n",
		wantKeys, wantMissing)
	fmt.Println("TASK.md states that as a judgement and carries no number, which is why")
	fmt.Println("neither can go stale. `--missing` lists them.")
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
