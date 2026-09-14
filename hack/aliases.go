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
	register("aliases", check{
		Needs: "this repository",
		What:  "**every search term `aliases.md` sends a reader to appears somewhere in the corpus** - the half `--links` cannot do, because a row routes by word rather than by path, and a word that lands nowhere reads exactly like a subject the material does not cover",
		Run:   runAliases,
	})
}

// aliasRow is a table row: what somebody said, and what to search for.
var aliasRow = regexp.MustCompile(`(?m)^\|\s*([^|]+?)\s*\|\s*(.+?)\s*\|\s*$`)

// pointer drops the `path/to.md` half of a cell. Those are checked by
// `audit-material --links`, and checking them twice reports one defect as two.
var pointer = regexp.MustCompile("`[^`]*`")

// aliasTerms pulls the searchable words out of one cell.
//
// **Everything after "and" is a pointer or a skill name**, not a word to grep,
// so the cell is cut there - "commerce, marketplace, channel - and
// `wiki/taiwan-channels.md`" contributes three terms.
func aliasTerms(cell string) []string {
	cell = pointer.ReplaceAllString(cell, "")
	if i := regexp.MustCompile(`\band\b`).FindStringIndex(cell); i != nil {
		cell = cell[:i[0]]
	}
	var out []string
	for _, t := range strings.Split(cell, ",") {
		t = strings.TrimSpace(strings.Trim(strings.TrimSpace(t), "-"))
		if t == "" || strings.HasPrefix(t, "the ") {
			continue
		}
		out = append(out, t)
	}
	return out
}

// corpusBody is everything a reader's grep would reach, minus the table itself.
// A term found only in `aliases.md` is the failure this check exists for: the
// row was the one document carrying every word of a translated query, and a
// search for a subject returned the word list rather than the page.
func corpusBody(root string) (string, error) {
	var b strings.Builder
	err := filepath.Walk(filepath.Join(root, "internal/corpus"), func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		if filepath.Base(p) == "aliases.md" {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		b.WriteString(strings.ToLower(string(data)))
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		return "", err
	}
	// The needs lists and the briefings land beside the corpus and a grep
	// reaches them, so they count as somewhere the word can be found.
	for _, rel := range []string{"internal/needs/needs.go", "internal/brief/brief.go"} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err == nil {
			b.WriteString(strings.ToLower(string(data)))
			b.WriteString("\n")
		}
	}
	// **And the design-time skills**, which land in the same repository and
	// which several rows point at by name. Leaving them out reported `投影片 ->
	// slides` as dead when the deck skill is full of the word and the row says
	// so - a check firing on correct material, which is the one kind this
	// repository deletes rather than tunes.
	err = filepath.Walk(filepath.Join(root, "internal/scaffold/templates/.agents/skills"),
		func(p string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			b.WriteString(strings.ToLower(string(data)))
			b.WriteString("\n")
			return nil
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func runAliases(args []string) error {
	root, err := src.Root()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(root, "internal/corpus/aliases.md"))
	if err != nil {
		return err
	}
	body, err := corpusBody(root)
	if err != nil {
		return err
	}

	seen := map[string]string{} // term -> the first row that sends somebody to it
	for _, m := range aliasRow.FindAllStringSubmatch(string(data), -1) {
		said := strings.TrimSpace(m[1])
		if said == "they said" || strings.HasPrefix(said, "---") {
			continue
		}
		for _, t := range aliasTerms(m[2]) {
			if _, ok := seen[t]; !ok {
				seen[t] = said
			}
		}
	}

	var dead []string
	for _, t := range sortedKeys(seen) {
		if !strings.Contains(body, strings.ToLower(t)) {
			dead = append(dead, fmt.Sprintf("  %q sends a reader to search for %q, which the corpus never uses", seen[t], t))
		}
	}
	sort.Strings(dead)
	for _, d := range dead {
		fmt.Println(d)
	}
	fmt.Printf("\n%d search term(s) in the alias table, %d that land nowhere.\n", len(seen), len(dead))
	if len(dead) > 0 {
		fmt.Println("\nA translated query that comes back empty reads as a subject nobody covered.")
		fmt.Println("Point the row at the word a page actually writes, or write the page.")
		return errFailed
	}
	return nil
}
