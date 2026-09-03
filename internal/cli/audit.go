package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/brief"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
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
	crs, err := generate.TemplateBodies()
	if err != nil {
		return nil, err
	}
	for name, body := range crs {
		out = append(out, source{"template", name, body})
	}
	for name, body := range generate.ValuesBlocks() {
		out = append(out, source{"template", name, body})
	}
	files, err := scaffold.TemplateBodies()
	if err != nil {
		return nil, err
	}
	for name, body := range files {
		// The skills are already in material() as prose; including them again
		// would double every hit in them.
		if strings.HasPrefix(name, ".agents/skills/") {
			continue
		}
		out = append(out, source{"scaffold", name, body})
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
		fmt.Fprintf(out, "\nNothing mentions it. If you were checking before a rename, there is\nnothing to rename; if you expected hits, check the spelling against\n`asgard-cli wiki glossary`.\n")
		return nil
	}
	fmt.Fprintf(out, "\nA rename has to touch all of them. The templates are the half that a\nprose-only search misses, and the half a customer's repository is built\nfrom - a stale field there is written into every new chart.\n")
	return nil
}

func material() ([]source, error) {
	var out []source
	pages, err := wiki.List()
	if err != nil {
		return nil, err
	}
	for _, p := range pages {
		body, err := wiki.Read(p.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, source{"wiki", p.Name, body})
	}
	extracts, err := usecase.List()
	if err != nil {
		return nil, err
	}
	for _, e := range extracts {
		body, err := usecase.Read(e.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, source{"usecase", e.Name, body})
	}
	for _, s := range stage.List() {
		body, err := s.Raw()
		if err != nil {
			return nil, err
		}
		out = append(out, source{"stage", string(s.Name), body})
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
		out = append(out, source{"skill", sk.Name, body})
	}
	return out, nil
}

func newAuditCmd() *cobra.Command {
	var onlyAsk, onlyUnmarked, cross, links, urls bool
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
    asgard-cli audit-material --term <s>   every line mentioning <s>, in the
                                           templates as well as the prose
    asgard-cli audit-material --urls       fetch every docs link; exits 1 on a
                                           404. Needs the network

**--ask is the set to read whole.** All three incidents were in it, and it is
short enough for one sitting.

**--links is the only part that fails.** Everything else here is for a person to
read; this one resolves every ` + "`asgard-cli wiki <page>`" + `, ` + "`usecase <extract>`" + `,
` + "`brief <activity>`" + ` and ` + "`next --stage <name>`" + ` the material writes, and exits 1
on one that resolves to nothing. A renamed page leaves the pointers to it
behind, and nobody finds out until a reader follows one - which is the same
failure as a stale instruction, except that it can be checked mechanically. Run
it before a release.

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
				return checkLinks(out, sources)
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

// pointer matches the ways this material sends a reader somewhere else. Each
// captures the target in group 2, after a keyword in group 1.
var pointer = regexp.MustCompile(`asgard-cli (wiki|usecase|brief) ([a-z0-9][a-z0-9-]*)|asgard-cli next --stage ([a-z0-9][a-z0-9-]*)`)

// checkLinks resolves every pointer the material writes and fails on one that
// goes nowhere.
//
// This is the one thing here that can be decided mechanically. A page renamed
// or an extract merged leaves every pointer to the old name behind, reading
// perfectly and resolving to nothing, and the reader who follows one has no way
// to tell a dead pointer from a page they failed to find. The checks that did
// this lived in a test suite that no longer exists, which is why it is a flag on
// a command that ships rather than a test on the maintainer's machine.
func checkLinks(out io.Writer, sources []source) error {
	known := map[string]map[string]bool{
		"wiki":    {},
		"usecase": {},
		"brief":   {},
		"stage":   {},
	}
	pages, err := wiki.List()
	if err != nil {
		return err
	}
	for _, p := range pages {
		known["wiki"][p.Name] = true
	}
	extracts, err := usecase.List()
	if err != nil {
		return err
	}
	for _, e := range extracts {
		known["usecase"][e.Name] = true
	}
	for _, n := range brief.Names() {
		known["brief"][n] = true
	}
	for _, st := range stage.List() {
		known["stage"][string(st.Name)] = true
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

	type dead struct{ where, kind, name string }
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
				found = append(found, dead{"generate " + k.Name, ref.kind, ref.name})
			}
		}
		for _, name := range k.AlsoRead {
			checked++
			if !known["usecase"][name] {
				found = append(found, dead{"generate " + k.Name, "usecase", name})
			}
		}
	}
	for _, s := range sources {
		seen := map[string]bool{}
		for _, m := range pointer.FindAllStringSubmatch(s.body, -1) {
			kind, name := m[1], m[2]
			if kind == "" {
				kind, name = "stage", m[3]
			}
			key := kind + "/" + name
			if seen[key] {
				continue
			}
			seen[key] = true
			checked++
			if !known[kind][name] {
				found = append(found, dead{s.label + " " + s.name, kind, name})
			}
		}
	}

	for _, d := range found {
		target := "asgard-cli " + d.kind + " " + d.name
		if d.kind == "stage" {
			target = "asgard-cli next --stage " + d.name
		}
		fmt.Fprintf(out, "dead  %s -> `%s`\n", d.where, target)
	}
	fmt.Fprintf(out, "\n%d pointer(s) resolved, %d dead.\n", checked, len(found))
	if len(found) > 0 {
		return fmt.Errorf("%d pointer(s) go nowhere", len(found))
	}
	return nil
}

// docsURL matches the documentation links the material cites. Only this host:
// a link to anywhere else is somebody else's uptime, and a checker that fails
// the build because a third-party blog moved is a checker people turn off.
var docsURL = regexp.MustCompile(`https://docs\.asgard-ai\.com/[A-Za-z0-9/_.-]*[A-Za-z0-9/_-]`)

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
			for _, u := range docsURL.FindAllString(line, -1) {
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
