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
	"regexp"
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

	// Links is every document this one points a reader at, in the order they
	// appear. See Link.
	Links []Link

	// Sources are the documentation URLs this document cites, in the order they
	// appear, deduplicated. They live in the Sources block at the foot of a
	// page and were only reachable by reading it: an engagement building a
	// customer deck copied nine of them out by hand, one page at a time, and
	// then checked each one itself.
	Sources []string

	// NamesCounterparts is true when the document carries the section where it
	// states its counterparts on purpose. A document that has one is answered
	// by it **including when the answer is none** - `glossary` says
	// "Corresponding extracts: None. This is about the material rather than
	// about a deployment", and reading on into its prose turned that into a
	// confident link to whichever extract a paragraph happened to mention.
	NamesCounterparts bool
}

// Link is one document pointing a reader at another.
//
// It is recovered when the document is parsed, not when a result is printed.
// That was two regular expressions at two points of use - one in `find` to name
// a hit's counterpart, one in `audit-material --links` to check the same
// pointer resolved - which could disagree about what a document pointed at, and
// left the corpus with no link graph at all. Without one, the lint the pattern
// asks for cannot be written: **material nothing points at is not read, and the
// writer never finds out, because the file is there.**
type Link struct {
	Kind string // wiki, usecase, guide, brief
	Name string

	// Path marks a pointer written as a relative path rather than as an
	// invocation. It is the difference between "follow this on disk" and "run
	// this" - and a path is a claim that the target is written into a
	// repository, which `audit-material --links` checks. `wiki log` is the one
	// that is not: it stays an invocation because init does not land it.
	Path bool

	// Deliberate marks a link inside the section where the document names its
	// counterpart on purpose, as against one mentioned in passing. Taking the
	// first pointer anywhere sent a reader to whichever reference happened to
	// appear earliest, which on the knowledge page was a CD note pointing at
	// `skill-set` rather than the page's own `knowledge-drive`.
	Deliberate bool
}

// Counterpart returns the document of this kind that d names as its
// counterpart, or "" when it names none.
func (d Doc) Counterpart(kind string) string {
	for _, l := range d.Links {
		if l.Kind == kind && l.Deliberate {
			return l.Name
		}
	}
	if d.NamesCounterparts {
		return ""
	}
	for _, l := range d.Links {
		if l.Kind == kind {
			return l.Name
		}
	}
	return ""
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

// Ref is one document's name and the file holding it. A corpus whose files are
// not one `<name>.md` under Dir supplies these itself.
type Ref struct {
	Name string
	Path string
}

// Corpus is one body of material: where the files are, and which of them are
// readable by name without appearing in a listing.
//
// Two fields are optional and exist because the four bodies do not agree at the
// byte level and should not be forced to. A stage prompt is
// `prompts/04-read-path.md` and is called `read-path`; a skill is
// `<name>/SKILL.md` and carries YAML frontmatter, which is the Agent Skills
// contract the runtime discovers it by and is not ours to change. What they are
// made to agree on is Doc - a name, a title, a summary, and what the document
// has and has not been held against - because that is what a reader and a
// search need, and it is the only part any of them can share.
type Corpus struct {
	FS  fs.FS
	Dir string

	// Docs lists the corpus's documents when they are not one `<name>.md`
	// directly under Dir. Optional.
	Docs func() ([]Ref, error)

	// Parse reads a document's metadata when it does not open with a "# "
	// title and a paragraph. Optional; Parse is the default.
	ParseDoc func(name string, data []byte) Doc

	// Scan builds the scanner for one document's body, when what is searched
	// and what is quotable are not the same text. Optional.
	Scan func(body string) Scanner

	// Unlisted are files that belong to the corpus's own bookkeeping rather
	// than being material about the subject - an index, a log, a README.
	// Readable by name, absent from List.
	Unlisted map[string]bool

	// What names the material, for an error a reader can act on.
	Noun    string // "extract", "wiki page"
	Command string // "asgard-cli usecase", "asgard-cli wiki"
}

// A pointer is how this material sends a reader to another document, and there
// are two forms because the two halves of the material are not in the same
// place.
//
// **A path, for anything that lands.** `asgard-cli init` writes the wiki and
// the extracts into a customer repository as `wiki/` and `usecase/` side by
// side, and internal/corpus holds them in that same layout, so
// `../usecase/write-path.md` resolves in both trees. It is written `../` even
// from inside the half it points into - `../wiki/tools.md` from a wiki page
// resolves back into `wiki/` - because one form that carries its own kind beats
// two forms that need to know where the reader is standing. That is also what
// lets this stay a function of the body alone.
//
// **An invocation, for anything that does not land yet.** `brief` and `guide`
// are still only in the binary, so a path to them would resolve nowhere;
// `asgard-cli guide requirements` is what a reader can act on. Those convert to
// paths when they land - see the sequence in TASK.md - and until then a
// document pointer is one form or the other depending on where its target
// lives, which is a state to get out of rather than a design.
//
// **A bare page name is not a pointer in either form.** A reader cannot act on
// `write-path` without knowing which command opens it or which directory it is
// in, and `audit-material --bare` reports one.
//
// The invocation form may wrap. This material is hard wrapped at about 78
// columns, so a pointer near the right margin is split across two lines, and a
// pattern expecting a single space did not see it - six real pointers were
// invisible to both `find`'s counterpart and `--links`, reading perfectly to a
// person the whole time. **Exactly one space, or a line break.** Not "any run
// of whitespace": a help screen aligns its columns with spaces, so
// `asgard-cli guide` followed by padding and the words "all of it" would
// resolve to a document called "all". A path carries no internal whitespace, so
// it has no equivalent problem.
//
// Both yield the same two groups - kind, then name - so everything downstream
// reads one shape: kb.Link, Counterpart, --links, --orphans, and what `find`
// prints.
var pathLinkRe = regexp.MustCompile(`\.\./(wiki|usecase|needs|brief)/([a-z0-9][a-z0-9-]*)\.md`)

var linkRe = regexp.MustCompile(`asgard-cli(?: |[ \t]*\n[ \t]*)(wiki|usecase|brief|guide)(?: |[ \t]*\n[ \t]*)([a-z0-9][a-z0-9-]*)`)

// counterpartSection is where a document states its counterparts on purpose.
var counterpartSection = regexp.MustCompile(
	`(?s)##+ (?:Before writing the chart|Corresponding extracts|Read the platform side first)[^` + "\n" + `]*` + "\n" + `(.*?)(?:` + "\n" + `##|\z)`)

// docsURL matches the documentation links this material cites. Only this host:
// a link anywhere else is somebody else's, and a page's Sources block is where
// the platform's own documentation is named.
var docsURL = regexp.MustCompile(`https://docs\.asgard-ai\.com/[A-Za-z0-9/_.-]*[A-Za-z0-9/_-]`)

// SourceURLs returns the documentation links this body cites, deduplicated, in
// the order they appear.
func SourceURLs(body string) []string {
	var out []string
	seen := map[string]bool{}
	for _, u := range docsURL.FindAllString(body, -1) {
		if seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

// Links returns every document this body points at, deduplicated, in the order
// they first appear, with the ones inside the counterpart section marked.
func Links(body string) ([]Link, bool) {
	deliberate := map[string]bool{}
	named := false
	if sec := counterpartSection.FindStringSubmatch(body); sec != nil {
		named = true
		for _, m := range pointers(sec[1]) {
			deliberate[m.key()] = true
		}
	}

	var out []Link
	seen := map[string]bool{}
	for _, m := range pointers(body) {
		if seen[m.key()] {
			continue
		}
		seen[m.key()] = true
		out = append(out, Link{Kind: m.kind, Name: m.name, Path: m.path, Deliberate: deliberate[m.key()]})
	}
	return out, named
}

// pointers returns every pointer in s, in both forms, as {whole, kind, name}.
//
// Order matters for Deliberate but not for correctness: a document naming the
// same target twice in one section is one link either way, and the dedupe in
// Links keeps the first.
func pointers(s string) []pointer {
	var out []pointer
	for _, m := range pathLinkRe.FindAllStringSubmatch(s, -1) {
		out = append(out, pointer{kind: m[1], name: m[2], path: true})
	}
	for _, m := range linkRe.FindAllStringSubmatch(s, -1) {
		out = append(out, pointer{kind: m[1], name: m[2]})
	}
	return out
}

type pointer struct {
	kind, name string
	path       bool
}

func (p pointer) key() string { return p.kind + "/" + p.name }

// ── Command references ────────────────────────────────────────────────────
//
// A document that tells somebody to run `asgard-cli pipeline deliveries` is
// making a checkable claim: that this build answers to that command. Six
// documents made exactly that claim about a command nobody had built, and it
// was found by a person re-reading a provenance line - which is a terrible
// mechanism for a claim a program can resolve in a millisecond.
//
// Links() above resolves the four DOCUMENT pointers (`wiki`, `usecase`,
// `brief`, `guide`). This resolves the invocation itself, against the command
// tree, and the two are deliberately separate: a dead document pointer is a
// rename nobody swept, a dead command is a command that does not exist.

// invocationRe matches a mention of this tool and the plain lower-case words
// after it. It stops where a command name cannot continue: a placeholder
// (`<name>`, `[project]`), a flag (`--links`), a redirection, a comment, a
// closing backtick, or end of line.
//
// **Exactly one space between words**, the same rule Links() settled on and
// for the same reason: these documents are full of two-column lists, where a
// command is padded out to a margin and the description follows.
//
//	asgard-cli project          every chart, and what each still lacks
//
// A separator of "one or more spaces" reads that as `project every`, and the
// first run of this check reported twelve of them.
//
// **Same line only.** Links() tolerates a pointer that wraps, because a
// document name at the right margin really does get split. This must not: the
// word after a line break is usually the next sentence, and consuming it would
// invent a subcommand out of prose. A wrapped invocation resolves the part
// before the break, which is still checked - a false negative here costs a
// missed reference, and a false positive costs a failing build over a sentence
// that reads perfectly.
var invocationRe = regexp.MustCompile(`asgard-cli((?: [a-z][a-z0-9-]*)*)`)

// fence matches the start or end of a fenced code block.
var fence = regexp.MustCompile("^\\s*```")

// flagRe matches a long flag. Only long ones: `-f` is also a diff marker, a
// bullet and half of `-force`, and a short flag carries no name to check.
var flagRe = regexp.MustCompile(`--([a-z][a-z0-9-]*)`)

// handoff matches where one command's arguments end and another program's
// begin. Without it, `asgard-cli render x | kubectl apply --dry-run=server`
// resolves `--dry-run` against `render`, and the first run of this check
// reported exactly that.
//
// A redirection has to be a `>` that follows a space. Every placeholder in this
// material ends in one - `asgard-cli project add <slug> --env dev` - and a bare
// `>` ended the span before the flag, which silently un-checked five of the six
// stale flags this found.
var handoff = regexp.MustCompile(`\|\||\||&&|;|(?:^|[ \t])>`)

// Invocation is one mention of this tool with the words that follow it.
type Invocation struct {
	// Words are the tokens after `asgard-cli`, in order. Empty for a bare
	// mention of the tool, which claims nothing.
	Words []string
	// Flags are the long flags written after those words, without the dashes.
	// A flag is the other half of the same claim - `--env may be repeated` sat
	// in `project add`'s own help for a day after the flag was deleted - and
	// it resolves against the same node the words reached.
	Flags []string
	// Line is the 1-based line it was written on, so a report can point at it.
	Line int
}

// Invocations returns every command reference this body makes.
//
// **Only the ones written as code.** Prose says "what asgard-cli uses" and
// "how asgard-cli is distributed", and a scanner that read those would resolve
// `uses` and `is` against the command tree and fail the build over correct
// English. A real instruction is written the way it is typed: inside backticks,
// inside a fenced block, or on an indented line of its own. Those three forms
// cover all 551 invocations in this material and none of the prose.
func Invocations(body string) []Invocation {
	var out []Invocation
	inFence := false
	for i, line := range strings.Split(body, "\n") {
		if fence.MatchString(line) {
			inFence = !inFence
			continue
		}
		for _, seg := range codeSegments(line, inFence) {
			ms := invocationRe.FindAllStringSubmatchIndex(seg, -1)
			for j, m := range ms {
				// A flag belongs to the invocation it follows, so one
				// invocation owns the text up to the next one - which is what
				// `asgard-cli render x | asgard-cli check xref -` needs.
				end := len(seg)
				if j+1 < len(ms) {
					end = ms[j+1][0]
				}
				if h := handoff.FindStringIndex(seg[m[3]:end]); h != nil {
					end = m[3] + h[0]
				}
				inv := Invocation{Words: strings.Fields(seg[m[2]:m[3]]), Line: i + 1}
				for _, f := range flagRe.FindAllStringSubmatch(seg[m[3]:end], -1) {
					inv.Flags = append(inv.Flags, f[1])
				}
				out = append(out, inv)
			}
		}
	}
	return out
}

// codeSegments returns the parts of a line that are written as code.
//
// Inside a fence the whole line is. Outside one, it is what the backticks
// enclose, plus the line itself when it begins with the tool's name - the
// indented example form, which the help screens and the wiki both use and
// which carries no backticks at all.
func codeSegments(line string, inFence bool) []string {
	if inFence {
		return []string{line}
	}
	var out []string
	// The indented-example form. **Indented** is the whole test: an example
	// block is set in from the margin, and a paragraph that opens with the
	// tool's name is prose - the root help begins "asgard-cli is what an agent
	// asks about integrating with Asgard", and reading that as code resolves
	// `is` against the command tree.
	if indent := len(line) - len(strings.TrimLeft(line, " \t")); indent > 0 {
		if trimmed := strings.TrimLeft(line, " \t$"); strings.HasPrefix(trimmed, "asgard-cli") {
			out = append(out, trimmed)
		}
	}
	// Backtick spans. An odd trailing backtick opens nothing, so a line with
	// one is read up to it and no further.
	parts := strings.Split(line, "`")
	for i := 1; i < len(parts); i += 2 {
		out = append(out, parts[i])
	}
	return out
}

// marker matches a provenance or attribution line - the bold-prefixed labels a
// document carries below its opening paragraph. They end the summary, and two
// of them are read as fields.
func marker(line string) bool {
	return strings.HasPrefix(line, "**") && strings.Contains(line, ":**")
}

// Parse reads a document's opening - the "# " title and the paragraph under it -
// and its provenance markers. Every file opens that way, which is what makes a
// listing possible without a separate registry to keep in step.
func Parse(name string, data []byte) Doc {
	d := Doc{Name: name}
	d.Links, d.NamesCounterparts = Links(string(data))
	d.Sources = SourceURLs(string(data))
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

// refs returns the corpus's documents and the files holding them.
func (c Corpus) refs() ([]Ref, error) {
	if c.Docs != nil {
		return c.Docs()
	}
	entries, err := fs.ReadDir(c.FS, c.Dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", c.Dir, err)
	}
	var out []Ref
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		out = append(out, Ref{
			Name: strings.TrimSuffix(e.Name(), ".md"),
			Path: c.Dir + "/" + e.Name(),
		})
	}
	return out, nil
}

// path resolves one document name to its file.
func (c Corpus) path(name string) (string, error) {
	refs, err := c.refs()
	if err != nil {
		return "", err
	}
	for _, r := range refs {
		if r.Name == name {
			return r.Path, nil
		}
	}
	return "", fmt.Errorf("no %s named %q; list them with `%s`", c.Noun, name, c.Command)
}

func (c Corpus) parse(name string, data []byte) Doc {
	if c.ParseDoc != nil {
		return c.ParseDoc(name, data)
	}
	return Parse(name, data)
}

// List returns every document in the corpus, sorted by name.
func (c Corpus) List() ([]Doc, error) { return c.list(false) }

// All returns every document, including the corpus's own bookkeeping - an
// index, a log, a README. List hides those because somebody listing the
// material does not want them; anything writing the corpus out needs them,
// because an export missing its index has no map. `asgard-cli init` writes the
// wiki and the extracts into a repository and hit exactly that: the shipped
// SKILL.md told the reader to start at the index, which had not been written.
func (c Corpus) All() ([]Doc, error) { return c.list(true) }

func (c Corpus) list(bookkeeping bool) ([]Doc, error) {
	refs, err := c.refs()
	if err != nil {
		return nil, err
	}

	var out []Doc
	for _, r := range refs {
		if c.Unlisted[r.Name] && !bookkeeping {
			continue
		}
		data, err := fs.ReadFile(c.FS, r.Path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", r.Path, err)
		}
		out = append(out, c.parse(r.Name, data))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Read returns one document in full.
func (c Corpus) Read(name string) (string, error) {
	path, err := c.path(name)
	if err != nil {
		return "", err
	}
	data, err := fs.ReadFile(c.FS, path)
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
		sc := Scanner{}
		if c.Scan != nil {
			sc = c.Scan(body)
		}
		h := sc.Scan(body, terms)
		if !h.Found() {
			continue
		}
		matches = append(matches, Match{Doc: d, Lines: h.Lines, Score: h.Score, Terms: h.Terms})
	}

	// A document carrying every term outranks one carrying more mentions of
	// fewer, so the exact hit stays on top and the fallback only ever appears
	// underneath it - or alone, when there was no exact hit at all.
	Rank(matches, func(m Match) Hit { return Hit{Terms: m.Terms, Score: m.Score} })

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

// Hit is what one document's body gave a query: which terms it carries, a few
// of the lines that carry them, and how often.
type Hit struct {
	Terms []string
	Lines []string
	Score int
}

// Found reports whether the body carried any term at all.
func (h Hit) Found() bool { return len(h.Terms) > 0 }

// Scan is the one implementation of "what did this document give the query".
//
// It exists because there were three. Corpus.Search had it, and so did
// internal/stage and internal/scaffold, which hold material that is not a
// Corpus - a stage has a number and a prompt template, a skill is a directory
// with frontmatter - and so could not share the rest of this file. What they
// share is the scoring, and three copies of it is the drift this package was
// written to prevent, one level down from where it was prevented.
//
// Line selection is part of the contract, not a detail: at most three, and
// nothing under 20 characters, because a bare field name quotes badly and says
// less than a sentence.
func Scan(body string, terms []string) Hit { return Scanner{}.Scan(body, terms) }

// Scanner is Scan with the two things one corpus needs differently.
//
// A skill's SKILL.md is searched whole - its frontmatter carries the
// description somebody queries on - but quoting a line out of that frontmatter,
// or out of an unrendered `<< >>` placeholder, puts broken text in a result. So
// what is matched and what is quotable are not always the same text.
type Scanner struct {
	// Quotable is the text shown lines are taken from, when it is not the
	// whole body. Empty means the body itself.
	Quotable string

	// Keep drops a candidate line before it is shown. Nil keeps everything the
	// length rule already allows.
	Keep func(line string) bool
}

// Scan reports what body gave the query under this scanner's rules.
func (sc Scanner) Scan(body string, terms []string) Hit {
	var h Hit
	lower := strings.ToLower(body)
	for _, term := range terms {
		if Covers(lower, term) {
			h.Terms = append(h.Terms, term)
		}
	}
	if len(h.Terms) == 0 {
		return h
	}

	quotable := sc.Quotable
	if quotable == "" {
		quotable = body
	}
	for _, line := range strings.Split(quotable, "\n") {
		lowerLine := strings.ToLower(line)
		for _, term := range h.Terms {
			if !Covers(lowerLine, term) {
				continue
			}
			h.Score++
			trimmed := strings.TrimSpace(line)
			if len(h.Lines) < 3 && len(trimmed) > 20 && (sc.Keep == nil || sc.Keep(trimmed)) {
				h.Lines = append(h.Lines, trimmed)
			}
			break
		}
	}
	return h
}

// Rank orders matches the way every part of the corpus orders them: a document
// carrying more of the query outranks one carrying more mentions of fewer, so
// an exact hit stays above a partial one rather than being buried by a document
// that repeats a single word.
func Rank[T any](items []T, hit func(T) Hit) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := hit(items[i]), hit(items[j])
		if len(a.Terms) != len(b.Terms) {
			return len(a.Terms) > len(b.Terms)
		}
		return a.Score > b.Score
	})
}
