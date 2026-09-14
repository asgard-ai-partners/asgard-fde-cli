package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
)

func init() {
	register("sources", check{
		Needs: "the clones",
		What:  "what each upstream resolves to, how far behind it is, and **whether a reading TASK.md records has gone behind the clone it was held against**; --extracts is how far each extract's source chart has moved",
		Run:   runSources,
	})
}

// Two commits per deployment, and they answer different questions. `written
// from` is the version an extract describes; `held against` is the version its
// claims were last read against, which is a weaker act and a later commit.
var writtenFrom = regexp.MustCompile(`(?m)^\|\s*([a-z0-9-]+)\s*\|\s*` + "`" + `([0-9a-f]{7,})` + "`")
var heldAgainst = regexp.MustCompile(`(?m)^\|\s*([a-z0-9-]+)\s*\|\s*` + "`" + `[0-9a-f]{7,}` + "`" + `[^|]*\|\s*` + "`" + `([0-9a-f]{7,})` + "`")

// The readings TASK.md records, and the clone each was held against. **This is
// the only checkable thing about a reading**: not that it happened - nobody but
// the reader can say that - but whether the thing it was held against has moved
// since. A `never` row has nothing to go stale.
var readings = map[string]string{
	"extracts-vs-charts":    "deployments",
	"wiki-vs-docs":          "docs",
	"processors-vs-palette": "docs",
}

func sourcesDoc(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "source/SOURCES.md"))
	return string(b), err
}

// referenceDeployments reads the eight reference deployments off
// `source/SOURCES.md`.
//
// **Not a list in this file.** That document is the only one allowed to name a
// customer's repository, and a second copy here is the drift this whole
// directory exists to catch. The two contract repositories are excluded by name
// because they are declared as sources in their own right.
func referenceDeployments(doc string) []string {
	re := regexp.MustCompile(`\[([a-z0-9-]+)\]\(https://github\.com/asgard-ai-platform/[a-z0-9-]+\)`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(doc, -1) {
		switch m[1] {
		case "asgard-kube", "asgard-docs", "asgard-core":
		default:
			seen[m[1]] = true
		}
	}
	return sortedKeys(seen)
}

func runSources(args []string) error {
	root, err := src.Root()
	if err != nil {
		return err
	}
	doc, err := sourcesDoc(root)
	if err != nil {
		return err
	}

	if len(args) > 0 && args[0] == "--extracts" {
		return extractsReport(root, doc)
	}

	width := 0
	for _, s := range src.Sources {
		if len(s.Env) > width {
			width = len(s.Env)
		}
	}
	for _, name := range src.Order {
		s := src.Sources[name]
		dir, err := src.Resolve(name)
		if err != nil {
			fmt.Printf("%-*s  not found   %s\n", width, s.Env, s.What)
			continue
		}
		at := src.Commit(dir)
		if at == "" {
			n := 0
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if fi, err := os.Stat(filepath.Join(dir, e.Name(), ".git")); err == nil && fi != nil {
					n++
				}
			}
			fmt.Printf("%-*s  %-10s %d clone(s) under it        %s\n", width, s.Env, "-", n, dir)
			continue
		}
		state := "(current)"
		if b := behind(dir); b != "" {
			state = "(" + b + ")"
		}
		fmt.Printf("%-*s  %-10s %-34s %s\n", width, s.Env, at, state, dir)
	}
	fmt.Println("\nNothing here pulls. `git -C <path> pull` before a reading that matters.")

	stale := staleReadings(root, doc)
	if len(stale) == 0 {
		fmt.Println("\nEvery reading TASK.md records is against a source that has not moved.")
		return nil
	}
	fmt.Println("\nreadings TASK.md records that are now behind their source:")
	for _, s := range stale {
		fmt.Printf("  %s\n", s)
	}
	fmt.Println("\nThat is not a failure - it is the size of what re-reading would cover.")
	return nil
}

// behind reports how far a clone is behind its own remote, as a phrase.
//
// **Read, never fetched.** This reports the clone as it stands; if the answer
// matters, pull first. A check that fetched would be answering a different
// question each time it ran.
func behind(dir string) string {
	head := src.Commit(dir)
	for _, ref := range []string{"origin/HEAD", "origin/main", "origin/master"} {
		out, err := src.Git(dir, "rev-parse", "--short", ref)
		if err != nil {
			continue
		}
		up := strings.TrimSpace(out)
		if up == head {
			return ""
		}
		n, err := src.Git(dir, "rev-list", "--count", "HEAD.."+ref)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s commit(s) behind %s", strings.TrimSpace(n), up)
	}
	return ""
}

// staleReadings reports which recorded readings are behind what they were held
// against.
//
// **Against the commit the reading names, not against the remote.** This used
// to report a clone that was behind its own origin as a reading gone stale,
// which is a different fact about a different thing: the clone had not moved at
// all, its remote had, and the extracts describe the clone.
func staleReadings(root, doc string) []string {
	task, err := os.ReadFile(filepath.Join(root, "TASK.md"))
	if err != nil {
		return []string{err.Error()}
	}
	var out []string
	for _, key := range sortedKeys(readings) {
		row := ""
		for _, line := range strings.Split(string(task), "\n") {
			if strings.HasPrefix(line, "|") && strings.Contains(line, "`"+key+"`") {
				row = line
				break
			}
		}
		if row == "" {
			out = append(out, key+": no row in TASK.md's pass, so nothing records what it was read against")
			continue
		}
		if strings.Contains(row, "**never") {
			continue
		}
		dir, err := src.Resolve(readings[key])
		if err != nil {
			out = append(out, fmt.Sprintf("%s: %s", key, err))
			continue
		}
		if readings[key] != "deployments" {
			if b := behind(dir); b != "" {
				out = append(out, fmt.Sprintf("%s: read against %s, which is now %s",
					key, src.Sources[readings[key]].Env, b))
			}
			continue
		}
		var moved []string
		for _, name := range sortedKeys(toSet(heldAgainstNames(doc))) {
			ref := heldAgainstMap(doc)[name]
			clone := filepath.Join(dir, name)
			if _, err := os.Stat(filepath.Join(clone, ".git")); err != nil {
				out = append(out, fmt.Sprintf("%s: no clone of %s, so its reading cannot be answered", key, name))
				continue
			}
			n := src.Since(clone, ref)
			if n < 0 {
				out = append(out, fmt.Sprintf("%s: %s has no commit %s, the one its claims were read against",
					key, name, ref))
			} else if n > 0 {
				moved = append(moved, fmt.Sprintf("%s by %d", name, n))
			}
		}
		if len(moved) > 0 {
			shown := moved
			suffix := ""
			if len(shown) > 4 {
				shown, suffix = shown[:4], "..."
			}
			out = append(out, fmt.Sprintf("%s: %d clone(s) have moved since the reading: %s%s",
				key, len(moved), strings.Join(shown, ", "), suffix))
		}
	}
	return out
}

func heldAgainstMap(doc string) map[string]string {
	out := map[string]string{}
	for _, m := range heldAgainst.FindAllStringSubmatch(doc, -1) {
		out[m[1]] = m[2]
	}
	return out
}

func heldAgainstNames(doc string) []string {
	var out []string
	for n := range heldAgainstMap(doc) {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func toSet(list []string) map[string]bool {
	out := map[string]bool{}
	for _, s := range list {
		out[s] = true
	}
	return out
}

// extractsReport says how far each extract's source chart has moved.
//
// **`source/SOURCES.md` used to carry these numbers as prose, and both of the
// two that were not zero had rotted within nine days.** Written-down distances
// between two moving things are the one shape of claim that cannot hold, so
// that column points here instead.
func extractsReport(root, doc string) error {
	base, err := src.Resolve("deployments")
	if err != nil {
		return err
	}
	type row struct {
		name, at, head string
		since          int
	}
	var rows []row
	width := 0
	for _, m := range writtenFrom.FindAllStringSubmatch(doc, -1) {
		name, at := m[1], m[2]
		if len(name) > width {
			width = len(name)
		}
		clone := filepath.Join(base, name)
		if _, err := os.Stat(filepath.Join(clone, ".git")); err != nil {
			rows = append(rows, row{name, at, "no clone", -1})
			continue
		}
		n := src.Since(clone, at)
		head, _ := src.Git(clone, "log", "-1", "--format=%h %ad", "--date=short")
		rows = append(rows, row{name, at, strings.TrimSpace(head), n})
	}
	fmt.Println("Each extract's source chart, as the clone stands. Nothing here pulls.")
	fmt.Println()
	moved := 0
	for _, r := range rows {
		switch {
		case r.since < 0:
			fmt.Printf("  %-*s  read at %s  -- %s\n", width, r.name, r.at, r.head)
		case r.since == 0:
			fmt.Printf("  %-*s  read at %s  unmoved\n", width, r.name, r.at)
		default:
			moved++
			fmt.Printf("  %-*s  read at %s  %d commit(s) since, now at %s\n", width, r.name, r.at, r.since, r.head)
		}
	}
	fmt.Printf("\n%d of %d have moved since the extracts were written from them.\n", moved, len(rows))
	fmt.Println("An extract describes one version of one chart; that is the size of the re-read.")
	return nil
}
