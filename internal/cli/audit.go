package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/brief"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/needs"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

// The instructions in this material are spread across 67 documents - 27 wiki
// pages, 22 extracts, 12 stage prompts and 6 skills - and three
// contradictions have reached a customer - every one found by somebody walking
// into it. The cause is that no two opposing instructions are ever in front of
// the same pair of eyes.
//
// This command makes that moment, and it lives in the binary rather than beside
// the source for one reason: **an audit that only runs on the maintainer's
// machine only finds what the maintainer can see.** What the maintainer finds
// by reading is inconsistency; what matters is the moment somebody follows an
// instruction into a wall, and that person has the binary and not the
// repository. It also reads the embedded copy, which is what an engagement
// actually gets - a script over the source files audits the input instead.
//
// Hidden, because its reader edits this material and the help output belongs to
// whoever is onboarding a customer.
var (
	boldSpan   = regexp.MustCompile(`(?s)\*\*(.+?)\*\*`)
	imperative = regexp.MustCompile(`^(?i)(ask|do not|don't|never|always|say|write|read|` +
		`check|get|take|use|put|send|give|keep|make|treat|assume|start|stop|leave|` +
		`prefer|avoid|confirm|count|decide|name|record|file|run|open)\b`)
	marked = regexp.MustCompile(`(?i)\bfor the (tracking|follow-up|row)\b|` +
		`\bin the (row|meeting|handover)\b|\bthis line is for\b|\baloud\b|` +
		`\bon a slide\b|\bin ` + "`" + `docs/`)
	asking = regexp.MustCompile(`(?i)\bask\b|\bcustomer\b|\bmeeting\b|\bthey\b`)
	// A sentence naming another command. A bare pointer claims nothing and
	// cannot be wrong this way; one that says what the other command reports
	// can be, and one did - a stage prompt described `check` as saying the
	// opposite of what it says.
	crossSentence = regexp.MustCompile("[^.!?\n]*`asgard-cli[^`]*`[^.!?]*[.!?]")
	claiming      = regexp.MustCompile(`(?i)\b(says?|said|reports?|tells?|warns?|prints?|` +
		`lists?|names?|carries|describes?|covers?|gives?|answers?|states?|has|have|` +
		`holds?|explains?|already|until)\b`)
)

type source struct {
	label string
	name  string
	body  string

	// links is every document this one points at, read when it was parsed.
	// The pointer graph is a fact about the material, so it is not rebuilt
	// here from the prose - see kb.Link.
	links []kb.Link
}

// everything is material() plus the files a repository is actually built from.
//
// The two are kept apart on purpose. The instruction audits read prose, because
// an instruction is a sentence and a template has none. The term sweep has to
// read both, because that is the whole failure it exists for: a platform field
// is taught in a template that writes it, an extract that explains it and a
// stage prompt that mentions it, and a rename caught in one leaves the other
// two teaching a field that no longer exists.
func everything() ([]source, error) {
	out, err := material()
	if err != nil {
		return nil, err
	}
	skills, err := scaffold.Skills()
	if err != nil {
		return nil, err
	}
	crs, err := generate.TemplateBodies()
	if err != nil {
		return nil, err
	}
	for name, body := range crs {
		out = append(out, source{label: "template", name: name, body: body})
	}
	for name, body := range generate.ValuesBlocks() {
		out = append(out, source{label: "template", name: name, body: body})
	}
	files, err := scaffold.TemplateBodies()
	if err != nil {
		return nil, err
	}
	// The design-time skills are already in material() as prose, and including
	// them again would double every hit in them.
	//
	// **The prefix alone was too broad.** `.agents/skills/` also holds the
	// platform corpus this CLI generates - `asgard-platform/SKILL.md` and its
	// index - and those are in no other source, so skipping the whole prefix
	// meant nothing audited them. It shipped a `SKILL.md` naming
	// `asgard-cli scaffold`, a command this build does not answer to, and
	// `--commands` reported 0 dead the whole time; the repo-side check found
	// it, in a scaffolded repository, which is a later and more expensive
	// place to find it. So the skip names the skills material() covered.
	covered := map[string]bool{}
	for _, sk := range skills {
		covered[path.Dir(scaffold.Path(sk.Name))+"/"] = true
	}
	for name, body := range files {
		if covered[path.Dir(name)+"/"] {
			continue
		}
		out = append(out, source{label: "scaffold", name: name, body: body})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].label != out[j].label {
			return out[i].label < out[j].label
		}
		return out[i].name < out[j].name
	})
	return out, nil
}

// sweep prints every line mentioning term, across prose and templates alike.
//
// A rename is the one vocabulary change that is cheap to do and expensive to do
// halfway, and nothing here could answer "where else does this word appear".
// The answer has to include the templates, so this is the only audit that reads
// them.
func sweep(out io.Writer, sources []source, term string) error {
	needle := strings.ToLower(term)
	files, hits := 0, 0
	for _, s := range sources {
		var lines []string
		for i, line := range strings.Split(s.body, "\n") {
			if !strings.Contains(strings.ToLower(line), needle) {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 120 {
				trimmed = trimmed[:117] + "..."
			}
			lines = append(lines, fmt.Sprintf("  %5d  %s", i+1, trimmed))
		}
		if len(lines) == 0 {
			continue
		}
		files++
		hits += len(lines)
		fmt.Fprintf(out, "\n%s %s\n", s.label, s.name)
		for _, l := range lines {
			fmt.Fprintln(out, l)
		}
	}
	fmt.Fprintf(out, "\n%q: %d line(s) in %d file(s).\n", term, hits, files)
	if files == 0 {
		fmt.Fprintf(out, "\nNothing mentions it. If you were checking before a rename, there is\nnothing to rename; if you expected hits, check the spelling against\n`.agents/skills/asgard-platform/wiki/glossary.md`.\n")
		return nil
	}
	fmt.Fprintf(out, "\nA rename has to touch all of them. The templates are the half that a\nprose-only search misses, and the half a customer's repository is built\nfrom - a stale field there is written into every new chart.\n")
	return nil
}

// **The corpus's own bookkeeping is audited too, and for a long time it was
// not.** `List` hides the wiki's index and README and the extracts' README,
// which is right for somebody listing pages and wrong here: they are documents
// this material ships, they carry pointers and they name commands. Five
// references to the deleted `asgard-cli usecase` sat in those three files
// while `--commands` reported 0 dead - and the check that found them was the
// one this tool writes into a customer repository, which is a later and more
// expensive place to find anything.
//
// So the source set is what lands, not what lists.
func material() ([]source, error) {
	var out []source
	pages, err := wiki.Landing()
	if err != nil {
		return nil, err
	}
	for _, p := range pages {
		body, err := wiki.Read(p.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, source{label: "wiki", name: p.Name, body: body, links: p.Links})
	}
	extracts, err := usecase.All()
	if err != nil {
		return nil, err
	}
	for _, e := range extracts {
		body, err := usecase.Read(e.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, source{label: "usecase", name: e.Name, body: body, links: e.Links})
	}
	guides, err := stage.Docs()
	if err != nil {
		return nil, err
	}
	links := map[string][]kb.Link{}
	for _, g := range guides {
		links[g.Name] = g.Links
	}
	for _, s := range stage.List() {
		body, err := s.Raw()
		if err != nil {
			return nil, err
		}
		out = append(out, source{label: "stage", name: string(s.Name), body: body, links: links[string(s.Name)]})
	}
	// **The needs rows were claiming to be audited and were not.** The package
	// comment says each row carries the document that owns it and that
	// `--links` resolves those "the way it resolves every other pointer here" -
	// which was true of the intent and false of the code: needs was in no
	// source, so a From naming an extract that does not exist passed with 0
	// dead. They are documents now, written into a repository as `needs/`, so
	// they join the graph as themselves.
	for _, d := range needs.Documents() {
		links, _ := kb.Links(d.Body)
		out = append(out, source{label: "needs", name: d.Name, body: d.Body, links: links})
	}
	// Same for the briefs, and the same gap: a `Where` naming a page that does
	// not exist passed with 0 dead until they became documents.
	for _, d := range brief.Documents() {
		links, _ := kb.Links(d.Body)
		out = append(out, source{label: "brief", name: d.Name, body: d.Body, links: links})
	}
	skills, err := scaffold.Skills()
	if err != nil {
		return nil, err
	}
	for _, sk := range skills {
		body, err := scaffold.Body(sk.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, source{label: "skill", name: sk.Name, body: body, links: sk.Links})
	}
	return out, nil
}

// bare matches a document named without the command that opens it - “ `guide
// projects` “ rather than “ `asgard-cli guide projects` “.
//
// **A pointer written that way is invisible to everything.** `kb.Link` reads a
// pointer as an invocation on purpose: a bare name is not actionable, because a
// reader cannot follow it without knowing which command takes it. The cost is
// that a writer who drops the prefix writes a pointer nothing can follow and
// nothing reports - seven of them were in the material, and one of them was the
// only route to `guide projects`, which became an orphan the moment the
// document that carried the other route was deleted.
var bare = regexp.MustCompile("`(wiki|usecase|guide|brief) ([a-z][a-z0-9-]*)`")

// checkBare reports a document named without its command, where the name
// resolves to a real document.
//
// Only where it resolves: “ `guide projects` “ names something, and a
// backticked phrase that happens to start with one of those words does not.
// The log is skipped - it is append-only, and a line in it records what was
// written at the time rather than sending anybody anywhere.
func checkBare(out io.Writer, sources []source) error {
	known, err := targets()
	if err != nil {
		return err
	}

	type hit struct{ where, kind, name string }
	var found []hit
	for _, s := range sources {
		seen := map[string]bool{}
		for _, m := range bare.FindAllStringSubmatch(s.body, -1) {
			key := m[1] + "/" + m[2]
			if seen[key] || !known[m[1]][m[2]] {
				continue
			}
			seen[key] = true
			found = append(found, hit{s.label + " " + s.name, m[1], m[2]})
		}
	}

	fmt.Fprintf(out, "Documents named without the command that opens them.\n\n"+
		"**A pointer written this way is invisible.** `kb.Link` reads a pointer as an\n"+
		"invocation on purpose - a bare name is not actionable, because a reader cannot\n"+
		"follow it without knowing which command takes it - so `--links` never checks\n"+
		"one and `--orphans` never counts one. Seven were in the material, and one was\n"+
		"the only route to `asgard-cli guide projects`.\n\n")

	for _, h := range found {
		fmt.Fprintf(out, "bare  %s -> `%s %s` should be `asgard-cli %s %s`\n", h.where, h.kind, h.name, h.kind, h.name)
	}
	fmt.Fprintf(out, "\n%d bare name(s).\n", len(found))
	if len(found) > 0 {
		return fmt.Errorf("%d document(s) named without a command", len(found))
	}
	return nil
}

// bookkeeping is the corpus's own navigation - the index of pages, the alias
// index and the log. None of it is material about the platform.
//
// It is separated from material() rather than left out, because the two checks
// want opposite things from it. **--links has to read it**: its rows carry
// pointers, and while the alias table lived on the glossary page those pointers
// were checked, so moving the index out without this would have quietly stopped
// checking eight of them - and `index.md` itself had never been checked at all,
// because it is Unlisted and so was never a source. **--orphans must not count
// it**: a list that names every page makes every page reachable, and the whole
// finding is that a document reachable only from a list is not reached.
func bookkeeping() []source {
	var out []source
	add := func(label, name, body string) {
		links, _ := kb.Links(body)
		out = append(out, source{label: label, name: name, body: body, links: links})
	}
	if body, err := wiki.Aliases(); err == nil {
		add("index", "aliases", body)
	}
	for _, name := range []string{"index"} {
		if body, err := wiki.Read(name); err == nil {
			add("index", "wiki "+name, body)
		}
	}
	if body, err := usecase.Read("index"); err == nil {
		add("index", "usecase index", body)
	}
	return out
}

// helpText is every command's own help, as a link source.
//
// **It is a body of material like any other.** Sixty-odd pointers into the
// corpus live in Long and Short strings, and a page renamed out from under one
// of them goes dead exactly the way a page's own pointer does. It is also
// where an agent is sent from before it has read anything, so a document
// reached only from here is reached, which the orphan count would otherwise
// get wrong.
//
// It is not part of material(): the instruction audits count sentences somebody
// wrote as guidance, and a usage string is not one.
func helpText(cmd *cobra.Command) []source {
	var out []source
	var walk func(c *cobra.Command, path string)
	walk = func(c *cobra.Command, path string) {
		if c.Hidden && c.Name() != "audit-material" {
			return
		}
		name := strings.TrimSpace(path + " " + c.Name())
		body := c.Short + "\n" + c.Long
		links, _ := kb.Links(body)
		if len(links) > 0 {
			out = append(out, source{label: "help", name: name, body: body, links: links})
		}
		for _, sub := range c.Commands() {
			walk(sub, name)
		}
	}
	walk(cmd, "")
	return out
}

func newAuditCmd() *cobra.Command {
	var onlyAsk, onlyUnmarked, cross, links, commands, orphans, bareNames, urls, unverified, paths bool
	var term string

	cmd := &cobra.Command{
		Use:    "audit-material",
		Short:  "Read every instruction in this tool's material at once",
		Hidden: true,
		Long: `Every instruction this tool ships, on one screen.

Three contradictions in this material have reached a customer and every one was
found by somebody walking into it. The cause is structural: the instructions are
spread across 67 documents, so **no two opposing ones are ever in front of the
same reader.** This makes that moment.

It detects nothing, deliberately. Matching opposing verbs over prose produces
noise, and a checker that cries wolf teaches people to change what it can see
rather than what is wrong - a failure this material has already caused once, in
a customer deck.

    asgard-cli audit-material              every instruction, by page
    asgard-cli audit-material --ask        only those about asking a customer
    asgard-cli audit-material --unmarked   only those not saying who they are for
    asgard-cli audit-material --crossref   only sentences claiming what another
                                           command says
    asgard-cli audit-material --links      resolve every pointer, and exit 1 on
                                           one that goes nowhere
    asgard-cli audit-material --commands   resolve every command this material
                                           names, and exit 1 on one that does
                                           not exist
    asgard-cli audit-material --paths      a landed document naming a path only
                                           this repository has
    asgard-cli audit-material --unverified what says nothing about having been
                                           checked, across every body
    asgard-cli audit-material --orphans    documents nothing points at. The
                                           index does not count as a pointer
    asgard-cli audit-material --bare       documents named without the command
                                           that opens them, which nothing sees
    asgard-cli audit-material --term <s>   every line mentioning <s>, in the
                                           templates as well as the prose
    asgard-cli audit-material --urls       fetch every docs link; exits 1 on a
                                           404. Needs the network

**--ask is the set to read whole.** All three incidents were in it, and it is
short enough for one sitting.

**--links is the only part that fails.** Everything else here is for a person to
read; this one resolves every ` + "`../wiki/<page>.md`" + ` and ` + "`../usecase/<extract>.md`" + `,
` + "`brief <activity>`" + ` and ` + "`guide <name>`" + ` the material writes, and exits 1
on one that resolves to nothing. A renamed page leaves the pointers to it
behind, and nobody finds out until a reader follows one - which is the same
failure as a stale instruction, except that it can be checked mechanically. Run
it before a release.

**--commands is --links for the tool itself.** --links resolves the documents
this material points at; this resolves the COMMANDS it tells somebody to run,
against the tree this binary actually answers to. It shipped without one:
` + "`asgard-cli pipeline deliveries`" + ` was named in six documents as the one place a
push that produced no run explains itself, and no such command had ever been
built - found by a person re-reading a provenance line, weeks later, which is a
terrible mechanism for a claim a program can resolve instantly. It reads the
scaffold templates too, because a scaffolded README is where a customer meets
these names first.

**--orphans is the other half of --links.** A pointer that goes nowhere is
caught by --links; a document nothing points at is not caught by anything, and
costs more - material nobody links to is not read, and the writer never finds
out, because the file is there. The index is deliberately not counted: one
engagement had ` + "`wiki operations`" + ` sitting in it under the title Connectivity while
an FDE spent a day on connectivity and never opened it. It does not fail the
build, because search answers for some of them.

**--paths is --links for everything that is not a document pointer.** These
files are written into a customer's repository, where "this repo" means theirs
and ` + "`source/SOURCES.md`" + ` is not there. A wiki page cited it in a Sources
block and another cited ` + "`APPROACH.md`" + `; both read perfectly here and neither
resolves where they land. A path inside a repository has to name the repository
it is inside, on the same line.

**--urls is the one that needs the network**, which is why it is not in --links.
Six of the 82 documentation links in this material were 404s when this was first
run, and nothing had ever checked: two directory URLs with no landing page, and
four pages marked ` + "`draft: true`" + `, which the site does not publish. A draft is the
one worth knowing about - the file is readable in a checkout, so the material is
sound and only the link is broken, and it looks identical to a link that was
never right.

**--term is for a rename.** When a platform field is renamed or retired, it is
taught in three places - a template that writes it, an extract that explains it,
a stage prompt that mentions it - and fixing one leaves the other two teaching a
field that no longer exists. This is the only audit here that reads the
templates, because they are the half a prose search misses and the half a
customer's repository is built from.

**--crossref catches the shape nothing else can.** A page saying "` + "`check`" + ` will
report X" while ` + "`check`" + ` reports the opposite: both pages read correctly alone.
Open each command a sentence names and confirm it says what the sentence claims.

It reads the embedded material - what an engagement actually gets - rather than
the source files, and it is in the binary rather than beside the source because
an audit that only runs on the maintainer's machine only finds what the
maintainer can see.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			sources, err := material()
			if err != nil {
				return err
			}
			if urls {
				all, err := everything()
				if err != nil {
					return err
				}
				return checkURLs(cmd.Context(), out, all)
			}
			if term != "" {
				all, err := everything()
				if err != nil {
					return err
				}
				return sweep(out, all, term)
			}
			if links {
				// **Templates carry pointers too**, and they are the half a
				// customer's chart is built from: a CR skeleton sends the
				// reader to the extract that explains the field it is about
				// to ask them to fill in. A path there that goes nowhere is
				// found in the customer's repository or not at all.
				all, err := everything()
				if err != nil {
					return err
				}
				all = append(all, bookkeeping()...)
				return checkLinks(out, append(all, helpText(cmd.Root())...))
			}
			if commands {
				// everything(), not material(): a scaffolded README is where
				// half of these are written, and it is the half a customer
				// reads first.
				all, err := everything()
				if err != nil {
					return err
				}
				all = append(all, helpText(cmd.Root())...)

				// **And this package's own strings**, which make the same
				// claim as a document and were the half nothing checked.
				strs, err := goStrings()
				if err != nil {
					return err
				}
				for name, body := range strs {
					all = append(all, source{label: "source", name: name, body: body})
				}

				// **And this repository's own documentation.** An agent
				// working here reads AGENTS.md before it reads anything else,
				// and a command named there is one it is about to run.
				docs, err := repoDocs()
				if err != nil {
					return err
				}
				for name, body := range docs {
					all = append(all, source{label: "repo", name: name, body: body})
				}
				return checkCommands(out, cmd.Root(), all)
			}
			if bareNames {
				return checkBare(out, append(sources, bookkeeping()...))
			}
			if unverified {
				return checkUnverified(out)
			}
			if paths {
				return checkPaths(out, append(sources, bookkeeping()...))
			}
			if orphans {
				// Help counts as a pointer and the index does not. A command's
				// help is read at the moment somebody is deciding what to run;
				// a catalogue is read by somebody who already suspects the
				// document exists.
				return checkOrphans(out, append(sources, helpText(cmd.Root())...))
			}
			if cross {
				return crossref(out, sources)
			}

			total, shown := 0, 0
			for _, s := range sources {
				var lines []string
				for _, m := range boldSpan.FindAllStringSubmatch(s.body, -1) {
					line := strings.Join(strings.Fields(m[1]), " ")
					if len(line) < 8 || !imperative.MatchString(line) {
						continue
					}
					total++
					if onlyAsk && !asking.MatchString(line) {
						continue
					}
					if onlyUnmarked && marked.MatchString(line) {
						continue
					}
					lines = append(lines, line)
				}
				if len(lines) == 0 {
					continue
				}
				fmt.Fprintf(out, "\n%s %s\n", s.label, s.name)
				for _, l := range lines {
					fmt.Fprintf(out, "  - %s\n", truncate(l, 150))
					shown++
				}
			}
			fmt.Fprintf(out, "\n%d shown of %d instructions.\n", shown, total)
			if onlyUnmarked {
				fmt.Fprintf(out, "\nEach tells somebody to do something without saying who it is for or\n"+
					"where the answer goes. Most instructions have one obvious reader; this is\n"+
					"the list to read when asking which do not.\n")
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.BoolVar(&onlyAsk, "ask", false, "only instructions about asking a customer")
	f.BoolVar(&onlyUnmarked, "unmarked", false, "only those with no reader or destination stated")
	f.BoolVar(&cross, "crossref", false, "only sentences claiming what another command says")
	f.BoolVar(&links, "links", false, "resolve every pointer in the material; exits 1 on a dead one")
	f.BoolVar(&commands, "commands", false, "resolve every `asgard-cli <command>` this material names, against the command tree; exits 1 on one that does not exist")
	f.BoolVar(&bareNames, "bare", false, "documents named without the command that opens them, which nothing else can see; exits 1 on one")
	f.BoolVar(&orphans, "orphans", false, "documents nothing else points at; the index does not count as a pointer")
	f.BoolVar(&unverified, "unverified", false, "documents carrying no record of having been held against anything")
	f.BoolVar(&paths, "paths", false, "a landed document naming a file only this repository has")
	f.StringVar(&term, "term", "", "every line mentioning this word, templates included - for a rename")
	f.BoolVar(&urls, "urls", false, "fetch every docs.asgard-ai.com link in the material; exits 1 on a 404. Needs the network")
	return cmd
}

func crossref(out io.Writer, sources []source) error {
	n := 0
	for _, s := range sources {
		var lines []string
		for _, m := range crossSentence.FindAllString(s.body, -1) {
			line := strings.Join(strings.Fields(m), " ")
			if len(line) < 60 || !claiming.MatchString(line) {
				continue
			}
			lines = append(lines, line)
		}
		if len(lines) == 0 {
			continue
		}
		sort.Strings(lines)
		fmt.Fprintf(out, "\n%s %s\n", s.label, s.name)
		for _, l := range lines {
			fmt.Fprintf(out, "  - %s\n", truncate(l, 190))
			n++
		}
	}
	fmt.Fprintf(out, "\n%d sentences claim what another command says.\n\n"+
		"Open each command named and confirm it says what the sentence claims.\n"+
		"Nothing textual catches this shape: both pages read correctly alone, and\n"+
		"one describes the other wrongly. It has shipped once - a stage prompt said\n"+
		"`check` would report a state until a request existed, while `check` said a\n"+
		"request comes after the interview.\n", n)
	return nil
}

// checkLinks resolves every pointer the material writes and fails on one that
// goes nowhere.
//
// This is the one thing here that can be decided mechanically. A page renamed
// or an extract merged leaves every pointer to the old name behind, reading
// perfectly and resolving to nothing, and the reader who follows one has no way
// to tell a dead pointer from a page they failed to find. The checks that did
// this lived in a test suite that no longer exists, which is why it is a flag on
// a command that ships rather than a test on the maintainer's machine.
// targets is every document a pointer can legitimately resolve to, by kind.
func targets() (map[string]map[string]bool, error) {
	known := map[string]map[string]bool{
		"wiki":    {},
		"usecase": {},
		"needs":   {},
		"brief":   {},
		"guide":   {},
	}
	pages, err := wiki.List()
	if err != nil {
		return nil, err
	}
	for _, p := range pages {
		known["wiki"][p.Name] = true
	}
	extracts, err := usecase.List()
	if err != nil {
		return nil, err
	}
	for _, e := range extracts {
		known["usecase"][e.Name] = true
	}
	for _, n := range needs.Names() {
		known["needs"][n] = true
	}
	for _, n := range brief.Names() {
		known["brief"][n] = true
	}
	for _, st := range stage.List() {
		known["guide"][string(st.Name)] = true
	}
	return known, nil
}

// checkOrphans reports the documents nothing else points at.
//
// **The index does not count as a pointer, and that is the whole check.** In
// one engagement `wiki operations` sat in `index.md` under the title
// Connectivity while an FDE spent a day on connectivity and never opened it -
// discovery is by pointer at the moment of need, not by browsing a list, and a
// document reachable only from the index is reachable only by somebody who
// already suspects it exists. So the index and the log are excluded as sources
// here: counting them would mark every page reachable and report nothing.
//
// It does not fail. An orphan is not a defect the way a dead pointer is - a
// document can be answered for by search alone - it is a reading list, and the
// judgement about each one is a person's.
func checkOrphans(out io.Writer, sources []source) error {
	known, err := targets()
	if err != nil {
		return err
	}

	pointedAt := map[string]bool{}
	for _, s := range sources {
		for _, l := range s.links {
			// A document pointing at itself is not somebody else finding it.
			if l.Kind == s.label || (l.Kind == "guide" && s.label == "stage") {
				if l.Name == s.name {
					continue
				}
			}
			pointedAt[l.Kind+"/"+l.Name] = true
		}
	}
	for _, k := range generate.Kinds {
		if k.Wiki != "" {
			pointedAt["wiki/"+k.Wiki] = true
		}
		if k.Extract != "" {
			pointedAt["usecase/"+k.Extract] = true
		}
		for _, n := range k.AlsoRead {
			pointedAt["usecase/"+n] = true
		}
	}

	fmt.Fprintf(out, "Documents nothing else points at.\n\n"+
		"**The index does not count.** `wiki operations` sat in it under the title\n"+
		"Connectivity while an FDE spent a day on connectivity and never opened it:\n"+
		"discovery is by pointer at the moment it is needed, and a document reachable\n"+
		"only from a list is reachable only by somebody who already suspects it.\n\n"+
		"Not a defect list. Search answers for some of these, and for some the right\n"+
		"fix is a sentence in the document that should have sent a reader here.\n")

	var total, orphaned int
	for _, kind := range []string{"wiki", "usecase", "needs", "brief", "guide"} {
		names := make([]string, 0, len(known[kind]))
		for n := range known[kind] {
			names = append(names, n)
		}
		sort.Strings(names)

		var bare []string
		for _, n := range names {
			total++
			if !pointedAt[kind+"/"+n] {
				bare = append(bare, n)
				orphaned++
			}
		}
		fmt.Fprintf(out, "\n%s\n  %d of %d\n", strings.ToUpper(kind), len(bare), len(names))
		for _, n := range bare {
			fmt.Fprintf(out, "    %s\n", n)
		}
	}
	fmt.Fprintf(out, "\n%d of %d documents are reached by no pointer.\n", orphaned, total)
	return nil
}

func checkLinks(out io.Writer, sources []source) error {
	known, err := targets()
	if err != nil {
		return err
	}

	// Sub-pages a command takes that are not corpus entries. `wiki index` and
	// the searches are real invocations and would otherwise read as dead.
	for _, extra := range []struct{ kind, name string }{
		{"wiki", "index"},
		{"wiki", "README"},
		{"usecase", "index"},
	} {
		known[extra.kind][extra.name] = true
	}

	// **Every document pointer is a path**, because every kind is written into
	// a repository. A path is a claim that the target is there, so it is
	// checked against what `init` writes rather than against the corpus, which
	// is whole here and not there.
	//
	// The invocation form is refused **in the material**: a document that
	// points at another document points at a file. A help screen is different
	// - `brief` and `guide` are commands, and a help screen naming one is
	// telling somebody to run it.
	lands := map[string]map[string]bool{"wiki": {}, "usecase": {}, "needs": {}, "brief": {}, "guide": {}}
	for _, n := range needs.Names() {
		lands["needs"][n] = true
	}
	for _, n := range brief.Names() {
		lands["brief"][n] = true
	}
	for _, st := range stage.List() {
		lands["guide"][string(st.Name)] = true
	}
	landing, err := wiki.Landing()
	if err != nil {
		return err
	}
	for _, p := range landing {
		lands["wiki"][p.Name] = true
	}
	extracts, err := usecase.All()
	if err != nil {
		return err
	}
	for _, e := range extracts {
		lands["usecase"][e.Name] = true
	}

	type dead struct{ where, kind, name, why string }
	var found []dead
	checked := 0

	// The generator's kinds name a wiki page and an extract in struct fields
	// rather than in prose, so no amount of reading the material finds them.
	// A kind pointing at an extract that was renamed goes unnoticed until
	// somebody runs `add` and follows the pointer, which is the failure this
	// whole check exists for and the half of it a prose search cannot reach.
	for _, k := range generate.Kinds {
		for _, ref := range []struct{ kind, name string }{
			{"wiki", k.Wiki},
			{"usecase", k.Extract},
		} {
			if ref.name == "" {
				continue
			}
			checked++
			if !known[ref.kind][ref.name] {
				found = append(found, dead{where: "generate " + k.Name, kind: ref.kind, name: ref.name})
			}
		}
		for _, name := range k.AlsoRead {
			checked++
			if !known["usecase"][name] {
				found = append(found, dead{where: "generate " + k.Name, kind: "usecase", name: name})
			}
		}
	}
	for _, s := range sources {
		// A source that arrived without its graph gets one read here. The
		// templates are the case: they are collected for the term sweep,
		// which reads bodies, and a pointer in one was invisible to this
		// until the day somebody followed it in a customer's chart.
		if s.links == nil {
			s.links, _ = kb.Links(s.body)
		}
		seen := map[string]bool{}
		for _, l := range s.links {
			key := l.Kind + "/" + l.Name
			if seen[key] {
				continue
			}
			seen[key] = true
			checked++
			if !known[l.Kind][l.Name] {
				found = append(found, dead{where: s.label + " " + s.name, kind: l.Kind, name: l.Name})
				continue
			}
			if l.Path && !lands[l.Kind][l.Name] {
				found = append(found, dead{where: s.label + " " + s.name, kind: l.Kind, name: l.Name,
					why: "written as a path, and `asgard-cli init` does not write that document into a repository"})
				continue
			}
			// Only in the material. `brief` and `guide` are still commands, so
			// a help screen naming one is an invocation on purpose - it is
			// telling somebody to run it, not pointing at a document.
			if !l.Path && s.label != "help" {
				found = append(found, dead{where: s.label + " " + s.name, kind: l.Kind, name: l.Name,
					why: "written as an invocation; every document lands, so a pointer is a path - `../" + l.Kind + "/" + l.Name + ".md`"})
			}
		}
	}

	for _, d := range found {
		if d.why != "" {
			fmt.Fprintf(out, "dead  %s -> `../%s/%s.md`: %s\n", d.where, d.kind, d.name, d.why)
			continue
		}
		fmt.Fprintf(out, "dead  %s -> `asgard-cli %s %s`\n", d.where, d.kind, d.name)
	}
	fmt.Fprintf(out, "\n%d pointer(s) resolved, %d dead.\n", checked, len(found))
	if len(found) > 0 {
		return fmt.Errorf("%d pointer(s) go nowhere", len(found))
	}
	return nil
}

// checkCommands resolves every command reference in this material against the
// command tree, and fails on one that names something this build does not
// answer to.
//
// **The failure it exists for shipped.** `asgard-cli pipeline deliveries` was
// named in six documents - the verification skill, a scaffolded AGENTS.md and
// README, two stage prompts - as the one place a push that produced no run
// explains itself, and no such command had ever been built. It was caught by a
// person re-reading a provenance line, weeks later. Nothing mechanical was
// looking, even though the tree is already enumerated at startup for `check`.
//
// So this is the same enumeration turned inward. `check` asks whether a
// CUSTOMER'S repository names a command this build no longer has; this asks
// whether OUR OWN material does, which is the half that writes the customer's
// repository in the first place.
//
// It reads the templates as well as the prose, for the reason the term sweep
// does: a scaffolded README is the half a prose-only search misses and the half
// every new engagement is built from.
// bareNamesCount is where a bare command name can only be a command: the
// material an engagement reads and the help it is printed. Not `source`, not
// `repo` - both talk about the packages those words also name.
var bareNamesCount = map[string]bool{
	"wiki": true, "usecase": true, "needs": true, "brief": true,
	"stage": true, "skill": true, "scaffold": true, "template": true,
	"help": true,
}

// removedNames is the single-word keys of `replacements`, which is the list a
// removal is already required to update. Multi-word keys are left out: they
// name a subcommand, and `asgard-cli project shape` is what the invocation
// check already resolves.
func removedNames() []string {
	var out []string
	for name := range replacements {
		if !strings.Contains(name, " ") {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func checkCommands(out io.Writer, root *cobra.Command, sources []source) error {
	type dead struct{ where, at, word string }
	var found []dead
	checked := 0

	for _, s := range sources {
		seen := map[string]bool{}
		report := func(line int, at, word string) {
			key := at + "/" + word
			if seen[key] {
				return
			}
			seen[key] = true
			found = append(found, dead{fmt.Sprintf("%s %s:%d", s.label, s.name, line), at, word})
		}
		// A bare removed name claims the same thing an invocation does. It is
		// resolved through the same tree so a name later reinstated stops
		// being reported without anybody remembering to take the row out.
		//
		// **Only where a bare word can only be a command.** In Go source a
		// body is one string literal per line, so `Dir: "needs"` is
		// indistinguishable from a claim; in this repository's own
		// documentation the same words are package names, and `| `brief` |`
		// opens a row of a table about `internal/brief`. Both are swept for
		// invocations, which are unambiguous, and the rendered `help` covers
		// what a reader actually sees of those strings.
		refs := kb.Invocations(s.body)
		if bareNamesCount[s.label] {
			refs = append(refs, kb.RemovedNames(s.body, removedNames())...)
		}
		for _, inv := range refs {
			if len(inv.Words) == 0 {
				continue
			}
			node, at, word, ok := resolveInvocation(root, inv.Words)
			checked++
			if !ok {
				report(inv.Line, at, word)
				continue
			}
			for _, f := range inv.Flags {
				checked++
				if hasFlag(node, f) {
					continue
				}
				report(inv.Line, strings.Join(inv.Words[:min(len(inv.Words), depthOf(node))], " "), "--"+f)
			}
		}
	}

	sort.Slice(found, func(i, j int) bool { return found[i].where < found[j].where })
	for _, d := range found {
		where := strings.TrimSpace("asgard-cli " + d.at)
		fmt.Fprintf(out, "dead  %s -> `%s %s`", d.where, where, d.word)
		if r, ok := replacements[d.word]; ok && d.at == "" {
			fmt.Fprintf(out, "  (removed: %s)", truncate(r, 90))
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintf(out, "\n%d command reference(s) resolved, %d dead.\n", checked, len(found))
	if len(found) > 0 {
		return fmt.Errorf("%d command reference(s) name something this build does not answer to", len(found))
	}
	// A checker that finds nothing to check passes everything. There is no
	// test suite here - it was removed deliberately - so the only thing
	// standing between a narrowed parser and a gate that silently stops
	// looking is this line.
	if checked == 0 {
		return fmt.Errorf("no command reference resolved at all, so this checked nothing: kb.Invocations stopped matching")
	}
	return nil
}

// resolveInvocation walks the words of one reference down the command tree.
//
// It stops being a command name the moment the current node has no child by
// that name. Whether that is a defect depends on where it stopped: a node with
// subcommands was expecting one, so an unknown word there is a dead reference;
// a leaf command was expecting an argument, so `asgard-cli guide onboarding`
// and `asgard-cli render internal-dev` resolve and stop.
//
// That distinction is the whole rule, and it is exactly the shape of the one
// that shipped - `pipeline` has subcommands and `deliveries` was not among
// them - while `asgard-cli add <kind>` and `asgard-cli check xref` are not.
//
// Hidden commands count: `audit-material` is hidden and the material names it.
func resolveInvocation(root *cobra.Command, words []string) (node *cobra.Command, at, word string, ok bool) {
	node = root
	var path []string
	for _, w := range words {
		if child := findChild(node, w); child != nil {
			node = child
			path = append(path, w)
			continue
		}
		if len(node.Commands()) > 0 {
			return node, strings.Join(path, " "), w, false
		}
		return node, "", "", true
	}
	return node, "", "", true
}

// depthOf is how many words it took to reach this command, so a report can
// name the command a flag was written against rather than the whole line.
func depthOf(node *cobra.Command) int {
	n := 0
	for c := node; c != nil && c.HasParent(); c = c.Parent() {
		n++
	}
	return n
}

// hasFlag asks whether this command accepts a long flag by that name, its own
// or one inherited from a parent.
//
// --help and --version are cobra's, added at execution rather than at
// construction, so they are not in either set when this walks the tree.
func hasFlag(node *cobra.Command, name string) bool {
	if name == "help" || name == "version" {
		return true
	}
	if node.Flags().Lookup(name) != nil {
		return true
	}
	return node.InheritedFlags().Lookup(name) != nil
}

func findChild(node *cobra.Command, name string) *cobra.Command {
	for _, c := range node.Commands() {
		if c.Name() == name || slices.Contains(c.Aliases, name) {
			return c
		}
	}
	return nil
}

// ourFiles are paths that exist in this repository and are never written into
// a customer's. A landed document naming one of them points at nothing.
//
// The material cites files in other repositories all the time and should: an
// extract that says which asgard-core file a constant came from is doing
// provenance properly. What separates the two is whether the line says which
// repository. So the rule is not "do not name a path" - it is **name the
// repository the path is inside**, and this reports the lines that do not.
//
// `pages/` and `extracts/` are here for a different reason: they are the
// layout this material used to have, and they resolve in neither tree now.
// Five documents still described themselves in those terms, one of them the
// file `SKILL.md` says to read first. A renamed directory leaves prose behind
// exactly the way a renamed page leaves a pointer behind, and only one of the
// two had a check.
var ourFiles = regexp.MustCompile(`(?:^|[^A-Za-z0-9_./-])((?:source|hack|internal|cmd|prompts|pages|extracts)/[A-Za-z0-9_./*-]*|(?:Goal|TASK|STRUCTURE|APPROACH)\.md|selfsrc\.go)`)

// knownRepos are the repository names the material is allowed to cite a path
// inside. **Add one when the material starts drawing on another repository**,
// the same contract as `replacements`: the list is what makes this check able
// to tell provenance from a dead pointer.
var knownRepos = []string{
	"asgard-fde-cli", "asgard-core", "asgard-kube", "asgard-docs",
	"asgard-ai-partners", "asgard-ai-platform", "workflow-service",
}

// checkPaths reports a landed document naming a file only this repository has,
// on a line that does not say which repository it is in.
func checkPaths(out io.Writer, sources []source) error {
	var found int
	checked := 0
	for _, s := range sources {
		for i, line := range strings.Split(s.body, "\n") {
			ms := ourFiles.FindAllStringSubmatch(line, -1)
			if ms == nil {
				continue
			}
			checked += len(ms)
			named := false
			for _, r := range knownRepos {
				if strings.Contains(line, r) {
					named = true
					break
				}
			}
			if named {
				continue
			}
			for _, m := range ms {
				found++
				fmt.Fprintf(out, "unrooted  %s %s:%d -> `%s`\n", s.label, s.name, i+1, m[1])
			}
		}
	}
	fmt.Fprintf(out, "\n%d repository path(s) checked, %d naming no repository.\n", checked, found)
	if found > 0 {
		return fmt.Errorf("%d path(s) resolve only in this repository, and these documents land in somebody else's", found)
	}
	return nil
}

// docsURL is kb's: a document's outbound documentation links are a fact about
// the document, read where every other one is. Only this host - a link to
// anywhere else is somebody else's uptime, and a checker that fails the build
// because a third-party blog moved is a checker people turn off.
var docsURL = kb.SourceURLs

// checkURLs fetches every documentation link and reports the ones that are not
// there.
//
// It is separate from --links because it needs the network, and a gate that
// only works online is a gate that fails on a plane. Run it before a release.
//
// The failure it was written for is not a typo. Four of the six it found on its
// first run cite pages marked "draft: true", which exist in a checkout of
// asgard-docs and are not published - so the material's content is sound, only
// the link is broken, and it is indistinguishable from a link that was never
// right without fetching it.
func checkURLs(ctx context.Context, out io.Writer, sources []source) error {
	seen := map[string][]string{}
	disclosed := map[string]bool{}
	var order []string
	for _, s := range sources {
		lines := strings.Split(s.body, "\n")
		for i, line := range lines {
			for _, u := range docsURL(line) {
				u = strings.TrimSuffix(u, ".")
				// A URL ending in / is a prose template - the pages write
				// `https://docs.asgard-ai.com/img/docs/<path>` to say how a
				// path becomes a URL. Fetching the prefix proves nothing.
				if strings.HasSuffix(u, "/") {
					continue
				}
				if _, ok := seen[u]; !ok {
					order = append(order, u)
				}
				where := s.label + " " + s.name
				if !slices.Contains(seen[u], where) {
					seen[u] = append(seen[u], where)
				}
				// A citation that already says the link 404s is not a defect,
				// it is a citation doing its job: the page is a draft, the
				// file is readable in a checkout, and the material says so.
				// Look at the citing line and the two after it, which is where
				// such a note goes.
				for _, near := range lines[i:min(i+3, len(lines))] {
					l := strings.ToLower(near)
					if strings.Contains(l, "404") || strings.Contains(l, "draft: true") {
						disclosed[u] = true
					}
				}
			}
		}
	}
	sort.Strings(order)

	client := &http.Client{Timeout: 15 * time.Second}
	dead, known := 0, 0
	for _, u := range order {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(out, "err   %s\n      cited by %s\n", u, strings.Join(seen[u], ", "))
			dead++
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 400 {
			continue
		}
		if disclosed[u] {
			known++
			fmt.Fprintf(out, "%d   %s  (the citation says so)\n", resp.StatusCode, u)
			continue
		}
		fmt.Fprintf(out, "%d   %s\n      cited by %s\n", resp.StatusCode, u, strings.Join(seen[u], ", "))
		dead++
	}

	fmt.Fprintf(out, "\n%d link(s) fetched, %d dead, %d dead and disclosed.\n", len(order), dead, known)
	if dead > 0 {
		fmt.Fprintf(out, "\nA 404 here is usually one of two things: a directory URL with no landing\npage, or a page marked `draft: true`, which asgard-docs does not publish. For\na draft, keep the citation and say the link 404s - the file is readable in a\ncheckout, and the content behind it is still where the material came from.\n")
		return fmt.Errorf("%d documentation link(s) are dead", dead)
	}
	return nil
}

// checkUnverified lists what says nothing about having been checked.
//
// A document carrying neither a Checked nor an Unchecked line is UNKNOWN, not
// fine - see kb.Doc.Verified. **Do not close it by writing the lines.** An
// `Unchecked:` line written to satisfy a listing converts UNKNOWN into a
// claim, which is worse than the silence it replaces.
func checkUnverified(out io.Writer) error {
	fmt.Fprintf(out, "What carries no record of having been held against anything.\n\n"+
		"Neither line present means UNKNOWN - not that the document is wrong, and\nnot that it is right.\n\n")

	bodies := []struct {
		label string
		list  func() ([]kb.Doc, error)
	}{
		{"wiki", wiki.List},
		{"usecase", usecase.List},
		{"needs", needs.List},
		{"brief", brief.List},
		{"guide", stage.Docs},
		{"skills", scaffold.List},
	}
	var total, bare int
	for _, b := range bodies {
		docs, err := b.list()
		if err != nil {
			return err
		}
		var unverified []string
		for _, d := range docs {
			if !d.Verified() {
				unverified = append(unverified, d.Name)
			}
		}
		total += len(docs)
		bare += len(unverified)
		fmt.Fprintf(out, "%-9s %d of %d\n", b.label, len(unverified), len(docs))
		for _, n := range unverified {
			fmt.Fprintf(out, "    %s\n", n)
		}
	}
	fmt.Fprintf(out, "\n%d of %d document(s) say nothing either way.\n", bare, total)
	if bare > 0 {
		return fmt.Errorf("%d document(s) carry no provenance marker, and the rule is that every one does", bare)
	}
	// A checker that finds nothing to check passes everything.
	if total == 0 {
		return fmt.Errorf("no document was read at all, so this checked nothing")
	}
	return nil
}
