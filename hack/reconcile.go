package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/brief"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/needs"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

func init() {
	register("reconcile", check{
		Needs:   "this repository",
		Listing: true,
		What:    "which pointers have not been read against their target since the target moved. **Reported, never failed** - a stale pointer may be about a paragraph nothing touched. `<document>...` records that you have read one, and there is deliberately no way to record them all at once",
		Run:     runReconcile,
	})
}

// **`source/SOURCES.md` records the commit an extract was read at; this records
// the same for a pointer inside the corpus.** `--links` says a pointer resolves
// and `go run ./hack related` says who points at a document - neither says
// whether the pointing document has been read against the pointed-at one since
// it last moved, which is the question a term sweep was standing in for.
//
// It is committed, unlike `.out/verified.json`. A recorded pass is a claim
// about one machine's tree; a recorded reading is a claim about the material,
// and the next person inherits it the way they inherit a `**Checked:**` line.
const reconciledFile = "source/reconciled.json"

// reconciled is from -> to -> the digest of `to` when `from` was last read
// against it.
type reconciled map[string]map[string]string

// docBodies returns every document's body, keyed the way a pointer names it.
//
// **Without its frontmatter.** A description is what an index renders, not a
// claim the pointing document could be wrong about, and counting it would send
// somebody re-reading five pages because one index row was reworded.
func docBodies() (map[string]string, error) {
	out := map[string]string{}
	pages, err := wiki.All()
	if err != nil {
		return nil, err
	}
	for _, d := range pages {
		body, err := wiki.Read(d.Name)
		if err != nil {
			return nil, err
		}
		out["wiki/"+d.Name] = kb.Body(body)
	}
	extracts, err := usecase.All()
	if err != nil {
		return nil, err
	}
	for _, d := range extracts {
		body, err := usecase.Read(d.Name)
		if err != nil {
			return nil, err
		}
		out["usecase/"+d.Name] = kb.Body(body)
	}
	for _, d := range needs.Documents() {
		out["needs/"+d.Name] = kb.Body(d.Body)
	}
	for _, d := range brief.Documents() {
		out["brief/"+d.Name] = kb.Body(d.Body)
	}
	guides, err := stage.StaticDocuments()
	if err != nil {
		return nil, err
	}
	for _, d := range guides {
		out["guide/"+d.Name] = kb.Body(d.Body)
	}
	// **The sixth body, and the one that ships.** The design-time skills under
	// `internal/scaffold/templates/.agents/skills/` land in every customer
	// repository and point at the corpus beside them; leaving them out of this
	// graph meant the half a customer reads first was the half nothing asked
	// about.
	skills, err := scaffold.Skills()
	if err != nil {
		return nil, err
	}
	for _, sk := range skills {
		body, err := scaffold.Body(sk.Name)
		if err != nil {
			return nil, err
		}
		out["skill/"+sk.Name] = kb.Body(body)
	}
	return out, nil
}

func bodyDigest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}

// outEdges is the graph the other way round: what each document points at.
func outEdges(root string) (map[string][]string, error) {
	in, err := corpusInEdges(root)
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for to, froms := range in {
		for _, from := range froms {
			out[from] = append(out[from], to)
		}
	}
	// **A document points at itself, and that edge is its own description.**
	// A `description:` is authored, not derived, so it does not follow the page
	// when the page changes - and a row that no longer says what is inside a
	// document is the failure an index exists to prevent, reading exactly like
	// a row that does. Two of them had gone stale in one session before this
	// edge existed. Reconciling a document is therefore also the act of saying
	// its own row still describes it.
	bodies, err := docBodies()
	if err != nil {
		return nil, err
	}
	for name := range bodies {
		out[name] = append(out[name], name)
	}
	for _, tos := range out {
		sort.Strings(tos)
	}
	return out, nil
}

func loadReconciled(root string) (reconciled, error) {
	r := reconciled{}
	data, err := os.ReadFile(filepath.Join(root, reconciledFile))
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	return r, json.Unmarshal(data, &r)
}

// saveReconciled merges what this run read into whatever is on disk now, and
// only then writes.
//
// **A recorded reading is append-only, and a read-modify-write is not.** This
// loaded the record, changed a document's entry and wrote the whole file back -
// so two readers working different slices at once, which is how a pass is
// actually done, silently lost whichever finished first. A record that drops
// readings is worse than no record: it reports work as owed that somebody did,
// and the second time that happens nobody trusts any of it.
//
// Merging by document rather than by file is what makes the write safe. This
// still races against a writer between the read below and the rename, which no
// amount of merging fixes - a lock would, and the cost of one is not worth the
// window. What it does remove is the certainty of loss.
func saveReconciled(root string, r reconciled) error {
	path := filepath.Join(root, reconciledFile)
	onDisk, err := loadReconciled(root)
	if err != nil {
		return err
	}
	for from, tos := range r {
		if onDisk[from] == nil {
			onDisk[from] = map[string]string{}
		}
		for to, digest := range tos {
			onDisk[from][to] = digest
		}
	}
	data, err := json.MarshalIndent(onDisk, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Written beside and renamed, so a reader never sees half a file.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func runReconcile(args []string) error {
	root, err := src.Root()
	if err != nil {
		return err
	}
	bodies, err := docBodies()
	if err != nil {
		return err
	}
	out, err := outEdges(root)
	if err != nil {
		return err
	}
	record, err := loadReconciled(root)
	if err != nil {
		return err
	}

	// **There is no `--all`, and its absence is the design.** A command that
	// records every pointer at once exists only to make this report look
	// clean, and what it records is a claim that somebody read 348 documents.
	// The backlog being visible is the point: it is finite, it closes one
	// reading at a time, and an empty record is an honest starting state where
	// a full one is a lie at scale.
	var marking []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			return fmt.Errorf("unknown flag %q; this takes the names of documents you have read", a)
		}
		marking = append(marking, a)
	}

	if len(marking) > 0 {
		marked, err := reconcileRecord(record, bodies, out, marking)
		if err != nil {
			return err
		}
		if err := saveReconciled(root, record); err != nil {
			return err
		}
		// **Print the claim, not a count of it.** This records rather than
		// reports, so somebody who runs it to see what is owed has instead
		// asserted they read every pointer out of that document - which
		// happened, and was only caught because the person then went and did
		// the reading. A number cannot be checked against memory; a list can.
		for _, from := range marking {
			for _, to := range out[from] {
				if _, ok := bodies[to]; ok {
					fmt.Printf("  read  %-34s against  %s\n", from, to)
				}
			}
		}
		fmt.Printf("\nrecorded %d pointer(s) across %d document(s) in %s.\n", marked, len(marking), reconciledFile)
		fmt.Println("\n**Every line above is a claim that you read it.** Nothing here can check")
		fmt.Println("that, which is why it is written down rather than derived - the same")
		fmt.Println("honesty a `**Checked:**` line asks for. If one of them is not true, read")
		fmt.Println("it now, or this record is worth less than none to whoever comes next.")
		return nil
	}

	reconcileReport(os.Stdout, record, bodies, out)
	return nil
}

// stalePointer is one pointer the report names, and the digest it was read at.
type stalePointer struct{ from, to, was string }

// reconcileRecord writes, for each document in marking, the digest every target
// it points at has right now, and returns how many pointers that covered.
//
// **Separated from the file so it can be driven.** What a record has to get
// right - that one document's reading covers its own pointers and nobody
// else's, and that a name this corpus does not have is refused rather than
// absorbed - is a property of this function and of nothing on disk, and a test
// that drove it through `runReconcile` would be a test that writes
// `source/reconciled.json`.
//
// **Every name is resolved before anything is written.** A reading is one claim
// about a set of documents; half of one, left behind by a typo in the second
// name, is a claim nobody made.
func reconcileRecord(record reconciled, bodies map[string]string, out map[string][]string, marking []string) (int, error) {
	for _, from := range marking {
		if _, known := out[from]; !known {
			return 0, fmt.Errorf("%q points at nothing, or is not a document - `go run ./hack related` lists the names", from)
		}
	}
	marked := 0
	for _, from := range marking {
		if record[from] == nil {
			record[from] = map[string]string{}
		}
		for _, to := range out[from] {
			if body, ok := bodies[to]; ok {
				record[from][to] = bodyDigest(body)
				marked++
			}
		}
	}
	return marked, nil
}

// reconcileReport writes the pointers a re-read is owed on, and returns them
// split the way it counts them: never recorded, and recorded against a target
// that has moved since.
//
// **The two are not one list.** "Nobody has read this against that" and "read,
// and the target has not moved" are the entire answer this file exists to hold,
// so a report that merged them into "not known good" would give back what
// `--links` already says.
func reconcileReport(w io.Writer, record reconciled, bodies map[string]string, out map[string][]string) (never, moved []stalePointer) {
	pointers := 0
	for _, from := range sortedKeys(out) {
		for _, to := range out[from] {
			body, ok := bodies[to]
			if !ok {
				continue
			}
			pointers++
			now := bodyDigest(body)
			switch was := record[from][to]; {
			case was == "":
				never = append(never, stalePointer{from, to, ""})
			case was != now:
				moved = append(moved, stalePointer{from, to, was})
			}
		}
	}

	for _, s := range moved {
		fmt.Fprintf(w, "moved   %-34s -> %-34s read at %s\n", s.from, s.to, s.was)
	}
	if len(never) > 0 && len(moved) == 0 {
		for _, s := range never[:min(len(never), 10)] {
			fmt.Fprintf(w, "never   %-34s -> %s\n", s.from, s.to)
		}
		if len(never) > 10 {
			fmt.Fprintf(w, "        ... and %d more never recorded\n", len(never)-10)
		}
	}

	fmt.Fprintf(w, "\n%d pointer(s): %d never recorded, %d whose target has moved since.\n",
		pointers, len(never), len(moved))
	fmt.Fprintln(w, "\n**This cannot fail, and a clean run is not a verdict.** A target moving")
	fmt.Fprintln(w, "does not make the pointing document wrong - the change may be in a")
	fmt.Fprintln(w, "paragraph it never relied on. What it says is which re-reads are owed,")
	fmt.Fprintln(w, "and that is a list somebody can finish, where \"read the corpus again\"")
	fmt.Fprintln(w, "is not. Record one with `go run ./hack reconcile <document>` after")
	fmt.Fprintln(w, "reading it, never before.")
	return never, moved
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
