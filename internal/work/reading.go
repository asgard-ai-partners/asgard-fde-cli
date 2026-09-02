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
		key := parts[1] + "/" + parts[2]
		r, ok := byKey[key]
		if !ok {
			r = &Read{Kind: parts[1], Name: parts[2], First: parts[0]}
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
