// Package usecase serves the extracts: how each shape of Asgard deployment is
// built, taken from ones already in production.
//
// An agent about to author a chart reaches for these. They are embedded rather
// than written into a customer repo, because an example nobody is using becomes
// stale boilerplate there, while a stale one here is fixed for every engagement
// in a single release.
package usecase

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed extracts
var extracts embed.FS

const dir = "extracts"

// Extract is one shape.
type Extract struct {
	Name    string
	Title   string
	Summary string

	// Checked and Unchecked are how far this page has been held against a
	// source, and what has not been. They exist because the two are not the
	// same claim and a reader cannot tell them apart from the prose: a field
	// name is checkable against the CRD, a shape is checkable against a
	// deployment, and "why it was designed this way" is checkable against
	// nothing at all.
	//
	// An extract carrying neither line is UNKNOWN, not fine. That is the
	// honest default and Verified() reports it as such.
	Checked   string
	Unchecked string
}

// Verified reports whether this extract records having been held against a
// source. False means nobody has written down that it was - which is different
// from it being wrong, and different from it being right.
func (e Extract) Verified() bool { return e.Checked != "" }

// heading pulls the first "# ..." line and the paragraph under it, which is how
// every extract opens.
func parse(name string, content []byte) Extract {
	e := Extract{Name: name}

	// A separate pass: these lines sit below the attribution, by which point
	// the summary loop has already returned.
	for _, line := range strings.Split(string(content), "\n") {
		switch {
		case strings.HasPrefix(line, "**Checked:**"):
			e.Checked = strings.TrimSpace(strings.TrimPrefix(line, "**Checked:**"))
		case strings.HasPrefix(line, "**Unchecked:**"):
			e.Unchecked = strings.TrimSpace(strings.TrimPrefix(line, "**Unchecked:**"))
		}
	}

	var summary []string
	for _, line := range strings.Split(string(content), "\n") {
		switch {
		case strings.HasPrefix(line, "# ") && e.Title == "":
			e.Title = strings.TrimSpace(line[2:])
		case e.Title == "":
			continue
		case strings.HasPrefix(line, "**Seen in:**"), strings.HasPrefix(line, "#"):
			// The attribution line and the next heading both end the summary.
			if len(summary) > 0 {
				e.Summary = strings.Join(summary, " ")
				return e
			}
		case strings.TrimSpace(line) == "":
			if len(summary) > 0 {
				e.Summary = strings.Join(summary, " ")
				return e
			}
		default:
			summary = append(summary, strings.TrimSpace(line))
		}
	}
	e.Summary = strings.Join(summary, " ")
	return e
}

// List returns every extract, README first because it explains how to read one.
func List() ([]Extract, error) {
	entries, err := fs.ReadDir(extracts, dir)
	if err != nil {
		return nil, fmt.Errorf("read extracts: %w", err)
	}

	var out []Extract
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".md")
		if name == "README" {
			continue
		}
		content, err := extracts.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		out = append(out, parse(name, content))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Read returns one extract in full.
func Read(name string) (string, error) {
	content, err := extracts.ReadFile(dir + "/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("no extract named %q; list them with `asgard-cli usecase`", name)
	}
	return string(content), nil
}

// Index returns the README, which explains how the extracts are organised.
func Index() (string, error) {
	content, err := extracts.ReadFile(dir + "/README.md")
	if err != nil {
		return "", fmt.Errorf("read index: %w", err)
	}
	return string(content), nil
}

// Match is one extract that matched a search, with the lines that matched.
type Match struct {
	Extract
	Lines []string
	Score int
}

// Search finds extracts mentioning all of the given terms. It is how an agent
// gets from a customer's words ("stock across platforms", "daily report") to
// the shape that answers them, without knowing what the shapes are called.
func Search(query string) ([]Match, error) {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil, fmt.Errorf("search needs at least one term")
	}

	all, err := List()
	if err != nil {
		return nil, err
	}

	var matches []Match
	for _, e := range all {
		content, err := Read(e.Name)
		if err != nil {
			return nil, err
		}
		lower := strings.ToLower(content)

		// Every term has to appear somewhere, so an extra word narrows rather
		// than widens - the opposite would make a long query useless.
		hitsAll := true
		for _, term := range terms {
			if !strings.Contains(lower, term) {
				hitsAll = false
				break
			}
		}
		if !hitsAll {
			continue
		}

		m := Match{Extract: e}
		for _, line := range strings.Split(content, "\n") {
			l := strings.ToLower(line)
			for _, term := range terms {
				if strings.Contains(l, term) {
					m.Score++
					if len(m.Lines) < 3 && len(strings.TrimSpace(line)) > 20 {
						m.Lines = append(m.Lines, strings.TrimSpace(line))
					}
					break
				}
			}
		}
		matches = append(matches, m)
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].Score > matches[j].Score })
	return matches, nil
}
