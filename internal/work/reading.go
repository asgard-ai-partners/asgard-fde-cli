// Reading records which pages of this tool's material an engagement actually
// opened.
//
// It exists to measure one thing that nothing else can see. Every real defect
// found in this material has been found by somebody walking into it, and the
// most expensive ones were not pages that were wrong - they were pages that
// were right and never opened. `wiki operations` sat in the index under the
// title Connectivity while an FDE spent a day on connectivity and never
// followed it, because nothing pointed there at the moment it was needed.
//
// **Discovery is by pointer, not by browsing.** A page nobody points at when it
// is needed does not exist. The gap between what this tool ships and what an
// engagement opened is that problem, measured, and it needs no judgement from
// anybody - only a log.
//
// What it cannot see is a page opened and misread. That still needs a person.
package work

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ReadingLog is where the record lives, relative to the repository root. It is
// the engagement's own file: it names pages of this tool and nothing about the
// customer, so it is safe to commit and useful to read six months later.
var ReadingLog = filepath.Join("docs", ".reading-log")

// Recall notes that one page was opened. A failure is ignored - a missing log
// must never break a command whose job is to print a page.
func Recall(root, kind, name string) {
	if root == "" || name == "" {
		return
	}
	line := fmt.Sprintf("%s\t%s\t%s\n", time.Now().UTC().Format("2006-01-02"), kind, name)
	path := filepath.Join(root, ReadingLog)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}

// legacyGuideKind is what `asgard-cli guide` wrote before it was renamed from
// `next --stage <name>`. Rows under it are read as `guide` and nothing writes
// it any more.
const legacyGuideKind = "stage"

// Read is one page and how often it was opened.
type Read struct {
	Kind  string
	Name  string
	Count int
	First string
	Last  string
}

// Readings returns what this engagement has opened, most-read first.
func Readings(root string) ([]Read, error) {
	data, err := readOptional(filepath.Join(root, ReadingLog))
	if err != nil || data == "" {
		return nil, err
	}

	byKey := map[string]*Read{}
	for _, line := range strings.Split(data, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		kind := parts[1]
		if kind == legacyGuideKind {
			// A log is a historical record and this file is committed, so
			// every engagement scaffolded before `next --stage` became
			// `guide` carries rows under the old name. Folding them here is
			// what keeps six months of somebody else's reading answerable by
			// the command name they would type today.
			kind = "guide"
		}
		key := kind + "/" + parts[2]
		r, ok := byKey[key]
		if !ok {
			r = &Read{Kind: kind, Name: parts[2], First: parts[0]}
			byKey[key] = r
		}
		r.Count++
		r.Last = parts[0]
	}

	out := make([]Read, 0, len(byKey))
	for _, r := range byKey {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// MissLog is where a search that landed nowhere is recorded, relative to the
// repository root.
//
// **It is a different file from ReadingLog because it has the opposite
// contract.** The reading log names pages of this tool and nothing else, so it
// is committed. A miss is the query as it was typed, which is whatever words
// the customer used - their system names, their people, their vocabulary - so
// it is written where nothing publishes it, and the scaffold's .gitignore keeps
// it out of the repository.
//
// It is written at all because the query that found nothing is the one piece of
// a defect report nobody has to be believed about. Everything else in a report
// is somebody's account of what happened; this is the tool's own record that a
// search was run and the corpus had no answer. `asgard-cli reading --misses`
// reads it back, and `asgard-cli issue-report --new` puts it in the report.
var MissLog = filepath.Join("docs", ".find-misses")

const missHeader = `# Searches this engagement ran that the material did not answer.
#
# NOT COMMITTED. These are queries as they were typed, which may carry the
# customer's own words. The scaffold's .gitignore excludes this file.
#
# date	kind	query
#   miss      nothing in the material carried any term
#   unplaced  results came back, but these terms appeared in none of them
#
# Each line is a candidate row for the alias index, which asgard-cli init
# writes to .agents/skills/asgard-platform/aliases.md
`

// Miss notes a search the material did not answer. Failures are ignored: a log
// that cannot be written must never break a search.
func Miss(root, kind, query string) {
	if root == "" || strings.TrimSpace(query) == "" {
		return
	}
	path := filepath.Join(root, MissLog)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_, statErr := os.Stat(path)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	if os.IsNotExist(statErr) {
		_, _ = f.WriteString(missHeader)
	}
	_, _ = f.WriteString(fmt.Sprintf("%s\t%s\t%s\n",
		time.Now().UTC().Format("2006-01-02"), kind, strings.ReplaceAll(query, "\t", " ")))
}

// AMiss is one recorded search the material did not answer.
type AMiss struct {
	Date  string
	Kind  string
	Query string
}

// Misses returns what this engagement searched for and did not find, newest
// last, in the order they happened. Order is the point: a run of misses on one
// afternoon is one subject somebody could not reach, and collapsing them by
// count would hide that.
func Misses(root string) ([]AMiss, error) {
	data, err := readOptional(filepath.Join(root, MissLog))
	if err != nil || data == "" {
		return nil, err
	}
	var out []AMiss
	for _, line := range strings.Split(data, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		out = append(out, AMiss{Date: parts[0], Kind: parts[1], Query: parts[2]})
	}
	return out, nil
}
