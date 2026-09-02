// Package wiki serves the platform wiki: what Asgard is made of and who each
// piece is for, distilled from the product documentation and the CRD contract.
//
// It answers a different question from internal/usecase. An extract there says
// how one shape of deployment is assembled, field by field, and assumes you
// already know the platform has that shape. These pages are what an FDE needs
// before that - in the first hour with a customer, where the question is which
// product the customer's sentence even lands in.
//
// Embedded rather than written into a customer repo, for the same reason the
// extracts are: a copy in one engagement goes stale where nobody is looking,
// while a stale page here is fixed for every engagement in one release.
package wiki

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed pages README.md
var content embed.FS

const dir = "pages"

// Page is one wiki page.
type Page struct {
	Name    string
	Title   string
	Summary string

	// Unchecked is what this page has NOT been held against a deployment. The
	// wiki is distilled from product documentation, which describes a UI, and
	// a lot of what it describes is not in any chart at all - so its pages are
	// checked less deeply than the usecase extracts and must say so.
	//
	// A page carrying no such line is UNKNOWN rather than fine.
	//
	// Both markers are the ones internal/usecase uses. The two bodies of
	// material are read together and a reader should not have to learn where
	// the provenance is written twice.
	Checked   string
	Unchecked string
}

// Verified reports whether this page records what was left unchecked. False
// means nobody wrote it down, which is not the same as nothing being left.
func (p Page) Verified() bool { return p.Unchecked != "" }

// parse pulls the first "# ..." line and the first paragraph under it. Every
// page opens that way, which is what makes a listing possible without a
// separate registry to keep in sync.
func parse(name string, data []byte) Page {
	p := Page{Name: name}

	// A separate pass: the line sits at the very bottom, below the sources.
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(line, "**Unchecked:**"); ok {
			p.Unchecked = strings.TrimSpace(rest)
		}
		if rest, ok := strings.CutPrefix(line, "**Checked:**"); ok {
			p.Checked = strings.TrimSpace(rest)
		}
	}

	var summary []string
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "# ") && p.Title == "":
			p.Title = strings.TrimSpace(line[2:])
		case p.Title == "":
			continue
		case strings.HasPrefix(line, "#"):
			if len(summary) > 0 {
				p.Summary = strings.Join(summary, " ")
				return p
			}
		case strings.TrimSpace(line) == "":
			if len(summary) > 0 {
				p.Summary = strings.Join(summary, " ")
				return p
			}
		default:
			summary = append(summary, strings.TrimSpace(line))
		}
	}
	p.Summary = strings.Join(summary, " ")
	return p
}

// List returns every page, sorted by name.
func List() ([]Page, error) {
	entries, err := fs.ReadDir(content, dir)
	if err != nil {
		return nil, fmt.Errorf("read pages: %w", err)
	}

	var out []Page
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".md")
		if name == "index" || name == "log" {
			// The index and the log are the wiki's own bookkeeping, not pages
			// about the platform. They are readable by name, not listed.
			continue
		}
		data, err := content.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		out = append(out, parse(name, data))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Read returns one page in full.
func Read(name string) (string, error) {
	data, err := content.ReadFile(dir + "/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("no wiki page named %q; list them with `asgard-cli wiki`", name)
	}
	return string(data), nil
}

// Conventions returns the wiki's own README: the three layers, the three
// operations, and the rules a page has to follow.
func Conventions() (string, error) {
	data, err := content.ReadFile("README.md")
	if err != nil {
		return "", fmt.Errorf("read conventions: %w", err)
	}
	return string(data), nil
}

// Match is one page that matched a search, with the lines that matched.
type Match struct {
	Page
	Lines []string
}

// Search finds pages mentioning all of the given terms. It is how somebody gets
// from a customer's words to the page that covers them, without knowing what
// the pages are called.
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
	for _, p := range all {
		data, err := Read(p.Name)
		if err != nil {
			return nil, err
		}
		lower := strings.ToLower(data)

		// Every term has to appear, so an extra word narrows rather than
		// widens - the opposite would make a long query useless.
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

		m := Match{Page: p}
		for _, line := range strings.Split(data, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			lowerLine := strings.ToLower(trimmed)
			for _, term := range terms {
				if strings.Contains(lowerLine, term) {
					m.Lines = append(m.Lines, trimmed)
					break
				}
			}
			if len(m.Lines) == 3 {
				break
			}
		}
		matches = append(matches, m)
	}
	return matches, nil
}
