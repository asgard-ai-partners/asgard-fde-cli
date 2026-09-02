// Package kb serves a body of embedded reference material.
//
// There are two of them - the platform wiki and the deployment extracts - and
// they answer different questions, so they stay separate as content. What they
// share is everything else: the same file shape, the same provenance markers,
// the same listing, the same search. That was two copies of one implementation
// until they drifted, which is the failure this package exists to prevent.
//
// The files themselves stay with their own package, because go:embed can only
// reach a directory it sits in. What each package keeps is its corpus; what it
// gets from here is what to do with one.
package kb

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Doc is one page or extract.
type Doc struct {
	Name    string
	Title   string
	Summary string

	// Checked and Unchecked are how far this document has been held against a
	// source, and what has not been. They are separate fields because they are
	// separate claims and a reader cannot tell them apart from the prose: a
	// field name is checkable against the CRD, a shape against a deployment,
	// and "why it was designed this way" against nothing at all.
	//
	// A document carrying neither is UNKNOWN, not fine. That is the honest
	// default, and Verified reports it as such.
	Checked   string
	Unchecked string
}

// Verified reports whether this document records having been held against a
// source. False means nobody wrote down that it was - which is different from
// it being wrong, and different from it being right.
func (d Doc) Verified() bool { return d.Checked != "" || d.Unchecked != "" }

// Match is one document a search hit, with the lines that hit and how often.
type Match struct {
	Doc
	Lines []string

	// Score is how many lines mentioned a term. It orders the results, so the
	// document that talks about the subject most comes first rather than the
	// one whose name sorts earliest.
	Score int

	// Terms are the query terms this document actually contains. A caller needs
	// them to say which half of a query landed: a search that quietly matched
	// on one word of six reads as an answer to the whole question, and the
	// reader acts on material about something else.
	Terms []string
}

// Corpus is one body of material: where the files are, and which of them are
// readable by name without appearing in a listing.
type Corpus struct {
	FS  fs.FS
	Dir string

	// Unlisted are files that belong to the corpus's own bookkeeping rather
	// than being material about the subject - an index, a log, a README.
	// Readable by name, absent from List.
	Unlisted map[string]bool

	// What names the material, for an error a reader can act on.
	Noun    string // "extract", "wiki page"
	Command string // "asgard-cli usecase", "asgard-cli wiki"
}

// marker matches a provenance or attribution line - the bold-prefixed labels a
// document carries below its opening paragraph. They end the summary, and two
// of them are read as fields.
func marker(line string) bool {
	return strings.HasPrefix(line, "**") && strings.Contains(line, ":**")
}

// parse reads a document's opening - the "# " title and the paragraph under it -
// and its provenance markers. Every file opens that way, which is what makes a
// listing possible without a separate registry to keep in step.
func parse(name string, data []byte) Doc {
	d := Doc{Name: name}
	lines := strings.Split(string(data), "\n")

	// The markers sit below the opening paragraph, by which point the summary
	// loop below has already returned, so they need a pass of their own.
	for _, line := range lines {
		if rest, ok := strings.CutPrefix(line, "**Checked:**"); ok {
			d.Checked = strings.TrimSpace(rest)
		}
		if rest, ok := strings.CutPrefix(line, "**Unchecked:**"); ok {
			d.Unchecked = strings.TrimSpace(rest)
		}
	}

	var summary []string
	done := func() Doc {
		d.Summary = strings.Join(summary, " ")
		return d
	}
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "# ") && d.Title == "":
			d.Title = strings.TrimSpace(line[2:])
		case d.Title == "":
			continue
		case marker(line):
			// Nothing below a marker is summary, including a marker's own
			// wrapped continuation lines. Ending here unconditionally is what
			// stops a document whose marker follows the title immediately -
			// conventions has no opening paragraph - from reporting the
			// marker's second line as its summary.
			return done()
		case strings.HasPrefix(line, "#"), strings.TrimSpace(line) == "":
			// A heading or a blank line ends it too, but only once there is a
			// summary, so the gap under the title is skipped.
			if len(summary) > 0 {
				return done()
			}
		default:
			summary = append(summary, strings.TrimSpace(line))
		}
	}
	return done()
}

// List returns every document in the corpus, sorted by name.
func (c Corpus) List() ([]Doc, error) {
	entries, err := fs.ReadDir(c.FS, c.Dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", c.Dir, err)
	}

	var out []Doc
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".md")
		if c.Unlisted[name] {
			continue
		}
		data, err := fs.ReadFile(c.FS, c.Dir+"/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		out = append(out, parse(name, data))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Read returns one document in full.
func (c Corpus) Read(name string) (string, error) {
	data, err := fs.ReadFile(c.FS, c.Dir+"/"+name+".md")
	if err != nil {
		return "", fmt.Errorf("no %s named %q; list them with `%s`", c.Noun, name, c.Command)
	}
	return string(data), nil
}

// File returns a file beside the corpus rather than in it - a README the
// corpus's own conventions live in.
func (c Corpus) File(path string) (string, error) {
	data, err := fs.ReadFile(c.FS, path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

// Search finds documents covering the given terms. It is how somebody gets from
// a customer's words to the material that covers them, without knowing what any
// of it is called.
//
// All the terms first, then any of them. Requiring all of them is right when the
// query is well aimed - an extra word should narrow - but it is the wrong answer
// to a query that is a handful of words from a customer's document, where one
// unknown word suppresses everything the other five would have found. So a query
// that matches nothing outright falls back to the documents matching the most
// terms, and Match.Terms records which ones, so the caller can say what did not
// land rather than presenting a partial hit as a whole one.
func (c Corpus) Search(query string) ([]Match, error) {
	terms := Terms(query)
	if len(terms) == 0 {
		return nil, fmt.Errorf("search needs at least one term")
	}

	all, err := c.List()
	if err != nil {
		return nil, err
	}

	var matches []Match
	for _, d := range all {
		body, err := c.Read(d.Name)
		if err != nil {
			return nil, err
		}
		lower := strings.ToLower(body)

		m := Match{Doc: d}
		for _, term := range terms {
			if Covers(lower, term) {
				m.Terms = append(m.Terms, term)
			}
		}
		if len(m.Terms) == 0 {
			continue
		}

		for _, line := range strings.Split(body, "\n") {
			lowerLine := strings.ToLower(line)
			for _, term := range m.Terms {
				if !Covers(lowerLine, term) {
					continue
				}
				m.Score++
				// Show at most three lines, and skip the very short ones - a
				// bare field name quotes badly and says less than a sentence.
				if trimmed := strings.TrimSpace(line); len(m.Lines) < 3 && len(trimmed) > 20 {
					m.Lines = append(m.Lines, trimmed)
				}
				break
			}
		}
		matches = append(matches, m)
	}

	// A document carrying every term outranks one carrying more mentions of
	// fewer, so the exact hit stays on top and the fallback only ever appears
	// underneath it - or alone, when there was no exact hit at all.
	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i].Terms) != len(matches[j].Terms) {
			return len(matches[i].Terms) > len(matches[j].Terms)
		}
		return matches[i].Score > matches[j].Score
	})

	// Once something matches every term, the partial matches are noise: they
	// are what the fallback is for, and the fallback is not needed.
	if len(matches) > 0 && len(matches[0].Terms) == len(terms) {
		for i, m := range matches {
			if len(m.Terms) < len(terms) {
				return matches[:i], nil
			}
		}
	}
	return matches, nil
}

// Terms splits a query the way Search reads it.
func Terms(query string) []string {
	return strings.Fields(strings.ToLower(query))
}

// Covers reports whether the body covers one term. Short terms are held to a
// word boundary: as a substring "ap" is inside "api", "apply" and "happen", so a
// query naming an access point matched almost the whole corpus and the result
// looked like an answer.
func Covers(lowerBody, term string) bool {
	if len(term) > 3 {
		return strings.Contains(lowerBody, term)
	}
	for i := 0; ; {
		j := strings.Index(lowerBody[i:], term)
		if j < 0 {
			return false
		}
		start := i + j
		end := start + len(term)
		if !wordByte(lowerBody, start-1) && !wordByte(lowerBody, end) {
			return true
		}
		i = start + 1
	}
}

// wordByte reports whether the byte at i is one a word can be made of, treating
// anything outside the string as a boundary.
func wordByte(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return c == '_' || c >= 0x80 ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
