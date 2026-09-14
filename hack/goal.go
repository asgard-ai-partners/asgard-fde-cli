package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
)

func init() {
	register("goal", check{
		Needs: "this repository - **the capability, not the material**",
		What:  "Goal.md's four points held against the binary, in a temp directory with no network, no account and no git repository",
		Run:   runGoal,
	})
}

const corpusDir = ".agents/skills/asgard-platform"

// Goal's first point names four bodies of knowledge. A landed tree missing one
// is that point half-delivered, and `--links` cannot see it: a directory that is
// not written has no pointers to go dead.
var goalKinds = []string{"wiki", "usecase", "needs", "brief", "guide"}

// Goal's second point: the deck is the only thing a customer reads, so it has
// rules of its own. These are the three it names.
var deckRules = []string{"screenshot", "never appear", "worth"}

// Terms an agent would grep for, in the customer's words or the platform's.
// Retrieval is the whole engineering problem in Goal's closing section, and
// "the material landed" is not the same as "a grep finds it".
var greps = []string{"allowlist", "botProviderClass", "immutable", "read-only"}

// runCLI runs the binary in a scratch directory with nothing available to it.
//
// **No network.** A proxy that resolves nowhere is the cheapest way to make a
// fetch fail rather than succeed slowly, and Goal's first point is that none is
// needed.
func runCLI(binary, dir string, args ...string) (string, string, error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY=http://127.0.0.1:1", "HTTPS_PROXY=http://127.0.0.1:1",
		"ALL_PROXY=http://127.0.0.1:1", "NO_PROXY=", "ASGARD_PROFILE=", "HOME="+dir)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	return out.String(), errb.String(), err
}

func runGoal(args []string) error {
	root, err := src.Root()
	if err != nil {
		return err
	}
	binary := filepath.Join(root, ".out/asgard-cli")
	if len(args) > 0 {
		binary = args[0]
	}
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("no binary at %s\n  go build -o .out/asgard-cli ./cmd/asgard-cli", binary)
	}
	binary, _ = filepath.Abs(binary)

	tmp, err := os.MkdirTemp("", "asgard-goal-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	var bad []string
	add := func(format string, a ...any) { bad = append(bad, fmt.Sprintf(format, a...)) }

	// ── Goal 1: the knowledge, offline, with no repository ───────────────
	//
	// No `git init`. Goal.md: a tool that needs a directory first will not get
	// asked, and the question is asked in a meeting.
	if _, stderr, err := runCLI(binary, tmp, "init"); err != nil {
		add("1: `init` failed in an empty directory with no network:\n%s", trim(stderr, 400))
	}
	corpus := filepath.Join(tmp, corpusDir)
	for _, kind := range goalKinds {
		if fi, err := os.Stat(filepath.Join(corpus, kind)); err != nil || !fi.IsDir() {
			add("1: `%s/%s/` did not land, so one of the four bodies is missing", corpusDir, kind)
		}
	}
	landed := markdownUnder(corpus)
	if len(landed) < 60 {
		add("1: only %d corpus documents landed; the corpus is not a handful of files", len(landed))
	}

	// **TASK.md's "N documents and M words" is the claim this material makes
	// about its own value**, and a floor of 60 does not hold it. The landed tree
	// is the only place the number is true of anything - the input tree has
	// neither the rendered `needs` shapes nor the briefings as files.
	var kinded []string
	for _, k := range goalKinds {
		paths, _ := filepath.Glob(filepath.Join(corpus, k, "*.md"))
		kinded = append(kinded, paths...)
	}
	task, err := os.ReadFile(filepath.Join(root, "TASK.md"))
	if err != nil {
		return err
	}
	// Matched on the words rather than on the markup around them: requiring the
	// bold to open immediately before the number meant a rewrap stopped it
	// matching, which this then reported as an unchecked claim.
	claim := regexp.MustCompile(`(\d+) documents and (?:over\s+)?([\d,]+)\s*\n?\s*words`).
		FindStringSubmatch(string(task))
	if claim == nil {
		add("1: TASK.md states no `**<n> documents and <n> words**` claim, so the " +
			"one number this material gives for its own size is unchecked")
	} else {
		docs, _ := strconv.Atoi(claim[1])
		words, _ := strconv.Atoi(strings.ReplaceAll(claim[2], ",", ""))
		got := 0
		for _, p := range kinded {
			data, _ := os.ReadFile(p)
			got += len(strings.Fields(string(data)))
		}
		if len(kinded) != docs {
			add("1: TASK.md says %d documents and %d landed across %s",
				docs, len(kinded), strings.Join(goalKinds, ", "))
		}
		// **The word count is a floor, not an equality.** Every edit moves it,
		// so an exact figure fails on the ordinary act of writing a paragraph -
		// which teaches whoever hits it to stop believing the check. A floor
		// fails on the thing worth failing on: material that has gone missing.
		switch {
		case got < words:
			add("1: TASK.md says over %s words and the landed corpus has %s. "+
				"Material has gone rather than grown.", comma(words), comma(got))
		case got > words+10000:
			add("1: TASK.md says over %s words and the landed corpus has %s, which is far "+
				"enough past it that the figure understates the material. Raise it.",
				comma(words), comma(got))
		}
	}

	// Retrieval: grep is the way in, so a grep has to find things.
	var body strings.Builder
	for _, p := range landed {
		data, _ := os.ReadFile(p)
		body.Write(data)
		body.WriteString("\n")
	}
	all := body.String()
	for _, term := range greps {
		if !regexp.MustCompile(`(?i)` + term).MatchString(all) {
			add("1: `grep %s` finds nothing in the landed corpus, and grep is the only way in", term)
		}
	}
	if _, err := os.Stat(filepath.Join(corpus, "aliases.md")); err != nil {
		add("1: `aliases.md` did not land, so a question in the customer's words has nothing to translate it")
	}
	// The two-sense warning only reaches a reader if the glossary carries the
	// word - that is the whole mechanism, not the instruction.
	gloss, err := os.ReadFile(filepath.Join(corpus, "wiki/glossary.md"))
	switch {
	case err != nil:
		add("1: `wiki/glossary.md` did not land")
	case !strings.Contains(string(gloss), "payment"):
		add("1: the glossary does not contain `payment`, so a grep for it will not surface the two senses")
	}

	// ── Goal 2: what to get from the customer, and the deck's rules ──────
	for _, shape := range []string{"semantic-layer", "chat-channel", "write-path"} {
		if _, err := os.Stat(filepath.Join(corpus, "needs", shape+".md")); err != nil {
			add("2: `needs/%s.md` did not land, and that file is Goal's second point", shape)
		}
	}
	var skills strings.Builder
	_ = filepath.Walk(filepath.Join(tmp, ".agents/skills"), func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || filepath.Base(p) != "SKILL.md" {
			return nil
		}
		data, _ := os.ReadFile(p)
		skills.Write(data)
		skills.WriteString("\n")
		return nil
	})
	for _, rule := range deckRules {
		if !strings.Contains(skills.String(), rule) {
			add("2: no landed skill mentions %q, and the deck's own rules are Goal's second point", rule)
		}
	}

	// ── Goal 3: the charts ───────────────────────────────────────────────
	if err := exec.Command("git", "-C", tmp, "init", "-q", ".").Run(); err != nil {
		return fmt.Errorf("git init in the scratch repository failed: %w", err)
	}
	for _, argv := range [][]string{
		{"project", "add", "app"},
		{"add", "dataconnector", "erp", "--project", "app"},
	} {
		if _, stderr, err := runCLI(binary, tmp, argv...); err != nil {
			add("3: `%s` failed:\n%s", strings.Join(argv, " "), trim(stderr, 300))
		}
	}
	chart := filepath.Join(tmp, "projects/app/chart/app")
	if _, err := os.Stat(filepath.Join(chart, "Chart.yaml")); err != nil {
		add("3: no chart at projects/app/chart/app, which Goal's third point names")
	}
	if len(find(chart, "dc-erp.yaml")) == 0 {
		add("3: `add dataconnector` wrote no CR")
	}
	if stdout, _, err := runCLI(binary, tmp, "check"); err != nil {
		add("3: `check` failed on a freshly scaffolded repository:\n%s", tail(stdout, 300))
	}

	// ── Goal 4: the way back in, out of the tool's own output ────────────
	//
	// The URL has to be in what the tool prints, not only in a document
	// somebody has to think to open.
	if stdout, _, _ := runCLI(binary, tmp, "issue-report"); !strings.Contains(stdout,
		"github.com/asgard-ai-partners/asgard-fde-cli") {
		add("4: `issue-report` does not print the repository URL")
	}
	if stdout, _, err := runCLI(binary, tmp, "issue-report", "--new"); err != nil || len(stdout) < 400 {
		add("4: `issue-report --new` did not write a report body")
	}
	skill, err := os.ReadFile(filepath.Join(corpus, "SKILL.md"))
	if err != nil || !strings.Contains(string(skill), "issue-report") {
		add("4: the landed SKILL.md does not name `issue-report`, so an agent that greps " +
			"and finds nothing is not told where to file it")
	}

	for _, b := range bad {
		fmt.Printf("goal  %s\n", b)
	}
	fmt.Printf("\nGoal.md's four points, held against the binary: %d unmet.\n", len(bad))
	if len(bad) > 0 {
		fmt.Println("\n**This is a capability check, not a consistency check.** Every other")
		fmt.Println("check here can pass while one of these fails, which is why it exists.")
		return errFailed
	}
	return nil
}

func markdownUnder(dir string) []string {
	var out []string
	_ = filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && strings.HasSuffix(p, ".md") {
			out = append(out, p)
		}
		return nil
	})
	return out
}

func find(dir, name string) []string {
	var out []string
	_ = filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && filepath.Base(p) == name {
			out = append(out, p)
		}
		return nil
	})
	return out
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}

// comma writes a count the way the claim it is compared against writes one.
func comma(n int) string {
	s := strconv.Itoa(n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}
