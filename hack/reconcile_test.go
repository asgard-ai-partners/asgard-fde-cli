package main

import (
	"bytes"
	"github.com/asgard-ai-partners/asgard-fde-cli/hack/internal/src"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

// **The property every test here defends is that a recorded reading never
// covers a document it did not read.** The record is a claim a person makes and
// nothing can check it; what can be checked is that the record means what the
// person meant - that it moves when the target does, that it does not move when
// something the reader never read does, that it covers the document it names
// and no other, and that a name nobody has is refused rather than absorbed.
//
// `reconcileRecord` and `reconcileReport` are driven directly. **Nothing here
// goes near `source/reconciled.json`**: the two paths that touch it,
// `loadReconciled` and `saveReconciled`, take a root and are driven against
// `t.TempDir()`, and `runReconcile` - which resolves the repository root from
// the working directory - is called by no test at all.

// fixture is a corpus small enough to reason about: two documents pointing at
// three, with one pointer shared, so a record written for one document can be
// held against what the other's is still saying.
func fixture() (bodies map[string]string, out map[string][]string) {
	bodies = map[string]string{
		"wiki/alpha":  "# Alpha\n\nthe alpha body\n",
		"wiki/beta":   "# Beta\n\nthe beta body\n",
		"wiki/gamma":  "# Gamma\n\nthe gamma body\n",
		"guide/delta": "# Delta\n\nthe delta body\n",
	}
	out = map[string][]string{
		"usecase/one": {"wiki/alpha", "wiki/beta"},
		"brief/two":   {"wiki/beta", "wiki/gamma"},
	}
	return bodies, out
}

// recordAll is what a reader who has read everything leaves behind, so a test
// about one target moving is not also a test about the rest being unrecorded.
func recordAll(t *testing.T, bodies map[string]string, out map[string][]string) reconciled {
	t.Helper()
	r := reconciled{}
	marking := sortedKeys(out)
	if _, err := reconcileRecord(r, bodies, out, marking); err != nil {
		t.Fatalf("reconcileRecord(%v): %v", marking, err)
	}
	return r
}

func report(t *testing.T, record reconciled, bodies map[string]string, out map[string][]string) (string, []stalePointer, []stalePointer) {
	t.Helper()
	var buf bytes.Buffer
	never, moved := reconcileReport(&buf, record, bodies, out)
	return buf.String(), never, moved
}

func pointerKeys(ps []stalePointer) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.from+" -> "+p.to)
	}
	sort.Strings(out)
	return out
}

// 1. The digest is what makes a record expire, so it has to move when the body
// does and stand still when it does not. A digest that did not move is a
// recorded reading that goes on covering a document it never read; one that
// moved on its own would send the reader back to an unchanged page every run,
// which is the same report with the names removed.
func TestBodyDigestMovesWithTheBodyAndNotOtherwise(t *testing.T) {
	const body = "# Alpha\n\nthe alpha body\n"
	first := bodyDigest(body)
	if again := bodyDigest(body); again != first {
		t.Errorf("one body digested twice gave %q then %q: every recorded reading expires on the next run", first, again)
	}
	if other := bodyDigest(body + "and a sentence somebody added\n"); other == first {
		t.Errorf("adding a sentence left the digest at %q, so a reading recorded before it still reads as current", first)
	}
	// The edit that changes nothing about the length is the one a digest over a
	// weaker summary would miss.
	if swapped := bodyDigest(strings.Replace(body, "alpha", "omega", 1)); swapped == first {
		t.Errorf("changing a word in place left the digest at %q", first)
	}
}

// 2. Frontmatter is not part of the digest. `description:` is what an index row
// renders and no pointing document rests on it, so rewording one must not
// report every pointer at that page as owed a re-read - a report that cries
// wolf is one somebody turns off.
//
// The raw strings are digested too, to show the stripping is what does it: if
// `docBodies` dropped its `kb.Body` call, the first comparison is the one that
// fails and the second is what it would fail to.
func TestFrontmatterIsNotPartOfTheDigest(t *testing.T) {
	const body = "# Words with one meaning here\n\nA word that means two things.\n"
	before := "---\ngroup: In practice\ndescription: \"what a word means here\"\n---\n" + body
	after := "---\ngroup: In practice\ndescription: \"what a word means here, and what the other senses are called\"\n---\n" + body

	if a, b := bodyDigest(kb.Body(before)), bodyDigest(kb.Body(after)); a != b {
		t.Errorf("rewording a description moved the digest from %q to %q, so every pointer at that page is reported as owed a re-read for an index row", a, b)
	}
	if a, b := bodyDigest(before), bodyDigest(after); a == b {
		t.Fatalf("the two fixtures digest the same (%q) before stripping, so this test cannot tell whether stripping happened", a)
	}
	// And the body a document actually carries is what survives, not a prefix
	// of it: stripping must not eat the title as well.
	if got := kb.Body(before); got != body {
		t.Errorf("kb.Body returned %q, want the body without its frontmatter", got)
	}

	// **The comparison above cannot see `docBodies`.** The two strings are
	// built here and stripped here, so a `docBodies` that dropped its `kb.Body`
	// call passes every assertion so far - which was this test's own defect
	// before this block existed. What is digested is what `docBodies` returns,
	// so that is what has to be held against the rule.
	bodies, err := docBodies()
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range bodies {
		if strings.HasPrefix(got, "---\n") {
			t.Errorf("docBodies keeps %s's frontmatter, so rewording its description reports every pointer at it as owed a re-read", name)
		}
	}
	// Meaningful only if a document on disk has frontmatter to strip.
	pages, err := wiki.All()
	if err != nil {
		t.Fatal(err)
	}
	framed := 0
	for _, p := range pages {
		raw, err := wiki.Read(p.Name)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(raw, "---\n") {
			framed++
		}
	}
	if framed == 0 {
		t.Fatal("no wiki page on disk carries frontmatter, so the check above proved nothing")
	}
}

// 3. `outEdges` is `corpusInEdges` transposed. Driven against the real corpus
// once, because an inversion here is invisible in every other test - the
// fixtures above are symmetric enough to pass either way - and reports the
// wrong document as owing the reading, which is a worklist somebody finishes
// without reading anything they needed to.
func TestOutEdgesIsTheInEdgeGraphTransposed(t *testing.T) {
	in, err := corpusInEdges(mustRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err := outEdges(mustRoot(t))
	if err != nil {
		t.Fatal(err)
	}

	edges := func(g map[string][]string, flip bool) map[string]bool {
		set := map[string]bool{}
		for a, bs := range g {
			for _, b := range bs {
				if flip {
					set[b+" -> "+a] = true
				} else {
					set[a+" -> "+b] = true
				}
			}
		}
		return set
	}
	// in is to -> [from], so flipping it gives from -> to, which is out.
	want, got := edges(in, true), edges(out, false)
	for e := range want {
		if !got[e] {
			t.Errorf("%s is in the in-edge graph and not in outEdges", e)
		}
	}
	// **Plus one self-edge per document, deliberately.** A document's own
	// `description:` is authored rather than derived, so it does not follow the
	// page when the page changes; the self-edge is what makes reconciling a
	// document also the act of saying its own index row still describes it.
	// Two rows had gone stale in one session before it existed.
	bodies, err := docBodies()
	if err != nil {
		t.Fatal(err)
	}
	for e := range got {
		if want[e] {
			continue
		}
		from, to, ok := strings.Cut(e, " -> ")
		if !ok || from != to {
			t.Errorf("%s is in outEdges, is not in the in-edge graph, and is not a self-edge", e)
			continue
		}
		if _, known := bodies[from]; !known {
			t.Errorf("%s is a self-edge for something that is not a document", e)
		}
	}
	for name := range bodies {
		if !got[name+" -> "+name] {
			t.Errorf("%s has no self-edge, so nothing ever asks whether its description still fits", name)
		}
	}
	if len(got) == 0 {
		t.Fatal("outEdges is empty, so this test proved nothing")
	}
}

// 4. A pointer nobody has recorded is `never`, and is not folded in with the
// ones that are recorded and unchanged. The distinction is the whole of what
// this file adds to `--links`: "it resolves" is what `--links` says, and "and
// somebody read the one against the other" is what only a record can say.
func TestAPointerNeverRecordedIsReportedAsNever(t *testing.T) {
	bodies, out := fixture()

	text, never, moved := report(t, reconciled{}, bodies, out)
	if len(moved) != 0 {
		t.Errorf("an empty record reported %d pointer(s) as moved, want none: nothing has been read, so nothing can have moved since", len(moved))
	}
	want := []string{
		"brief/two -> wiki/beta", "brief/two -> wiki/gamma",
		"usecase/one -> wiki/alpha", "usecase/one -> wiki/beta",
	}
	if got := pointerKeys(never); !reflect.DeepEqual(got, want) {
		t.Errorf("never = %v, want every pointer %v", got, want)
	}
	if !strings.Contains(text, "4 pointer(s): 4 never recorded, 0 whose target has moved since.") {
		t.Errorf("the report does not count the unrecorded pointers as never:\n%s", text)
	}
	if !strings.Contains(text, "never   usecase/one") {
		t.Errorf("the report does not name an unrecorded pointer:\n%s", text)
	}

	// **And a read pointer is not reported at all**, which is what makes the
	// list finite. If `never` covered both, finishing it would be impossible.
	record := recordAll(t, bodies, out)
	text, never, moved = report(t, record, bodies, out)
	if len(never) != 0 || len(moved) != 0 {
		t.Errorf("after recording everything: %d never, %d moved, want none", len(never), len(moved))
	}
	if !strings.Contains(text, "4 pointer(s): 0 never recorded, 0 whose target has moved since.") {
		t.Errorf("a fully recorded corpus is not reported as such:\n%s", text)
	}
}

// 5. A recorded pointer whose target has moved is `moved`, and one whose target
// has not is neither. The digest in the report is the one it was read at, not
// the one it has now - that is what tells a reader which version they last
// held the pointing document against.
func TestARecordedPointerIsMovedOnlyWhenItsTargetMoved(t *testing.T) {
	bodies, out := fixture()
	record := recordAll(t, bodies, out)
	was := record["usecase/one"]["wiki/alpha"]

	bodies["wiki/alpha"] = bodies["wiki/alpha"] + "\na paragraph somebody added later\n"

	text, never, moved := report(t, record, bodies, out)
	if len(never) != 0 {
		t.Errorf("never = %v, want none: every pointer was recorded", pointerKeys(never))
	}
	if got, want := pointerKeys(moved), []string{"usecase/one -> wiki/alpha"}; !reflect.DeepEqual(got, want) {
		t.Errorf("moved = %v, want %v: only the pointer whose target moved", got, want)
	}
	if moved[0].was != was {
		t.Errorf("moved pointer was read at %q, want the digest it was recorded at, %q", moved[0].was, was)
	}
	if !strings.Contains(text, "moved   usecase/one") || !strings.Contains(text, "read at "+was) {
		t.Errorf("the report does not name the moved pointer and what it was read at:\n%s", text)
	}
	// wiki/beta is pointed at by both documents and did not move; neither
	// pointer at it may be reported, or "moved" degrades into "something in
	// this corpus changed".
	if strings.Contains(text, "wiki/beta") || strings.Contains(text, "wiki/gamma") {
		t.Errorf("the report names a pointer whose target did not move:\n%s", text)
	}
}

// 6. Recording one document records that document's pointers and nothing else.
// **This is where a record would come to cover a document nobody read**: a
// reader says they have read `usecase/one`, and every other document's claim
// has to be exactly what it was before they said it.
func TestRecordingOneDocumentLeavesEveryOtherRecordAlone(t *testing.T) {
	bodies, out := fixture()

	// brief/two was read against an older wiki/beta, and this reading is not
	// about brief/two.
	record := reconciled{"brief/two": {"wiki/beta": "0000deadbeef", "wiki/gamma": "1111deadbeef"}}

	marked, err := reconcileRecord(record, bodies, out, []string{"usecase/one"})
	if err != nil {
		t.Fatal(err)
	}
	if marked != 2 {
		t.Errorf("marked = %d, want 2: usecase/one's two pointers", marked)
	}
	if got, want := sortedKeys(record), []string{"brief/two", "usecase/one"}; !reflect.DeepEqual(got, want) {
		t.Errorf("record covers %v, want only the document read and the one already there", got)
	}
	for to, want := range map[string]string{"wiki/beta": "0000deadbeef", "wiki/gamma": "1111deadbeef"} {
		if got := record["brief/two"][to]; got != want {
			t.Errorf("reading usecase/one changed brief/two -> %s from %q to %q: it now claims a reading nobody did", to, want, got)
		}
	}
	for _, to := range []string{"wiki/alpha", "wiki/beta"} {
		if got, want := record["usecase/one"][to], bodyDigest(bodies[to]); got != want {
			t.Errorf("usecase/one -> %s recorded as %q, want the target's digest now, %q", to, got, want)
		}
	}

	// brief/two still owes its re-read, and usecase/one does not.
	_, never, moved := report(t, record, bodies, out)
	if got, want := pointerKeys(moved), []string{"brief/two -> wiki/beta", "brief/two -> wiki/gamma"}; !reflect.DeepEqual(got, want) {
		t.Errorf("moved = %v, want %v", got, want)
	}
	if len(never) != 0 {
		t.Errorf("never = %v, want none", pointerKeys(never))
	}
}

// 7. A missing record file is an empty record rather than an error - the
// starting state of this file is that nobody has recorded anything, and a check
// that refused to run until somebody created one would be a check nobody runs.
// And what a save writes is what the next load reads: the record survives
// version control, so a round trip that lost a document would silently reset
// its readings to never.
func TestLoadReconciledOnAMissingFileAndTheRoundTrip(t *testing.T) {
	root := t.TempDir()

	r, err := loadReconciled(root)
	if err != nil {
		t.Fatalf("loadReconciled on a root with no %s: %v", reconciledFile, err)
	}
	if len(r) != 0 {
		t.Errorf("loadReconciled on a missing file returned %v, want an empty record", r)
	}
	// Empty rather than nil, so the caller can write into it.
	r["usecase/one"] = map[string]string{"wiki/alpha": bodyDigest("the alpha body")}

	want := reconciled{
		"usecase/one": {"wiki/alpha": "aaaa11112222", "wiki/beta": "bbbb11112222"},
		"brief/two":   {"wiki/gamma": "cccc11112222"},
	}
	if err := saveReconciled(root, want); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, reconciledFile)); err != nil {
		t.Fatalf("saveReconciled wrote nothing at %s: %v", reconciledFile, err)
	}
	got, err := loadReconciled(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip gave %v, want %v", got, want)
	}
}

// 8. A name this corpus does not have is an error. **A typo that recorded
// nothing and said so cheerfully is the worst outcome this command has**: the
// reader has done the reading, believes it is written down, and the report goes
// on listing the pointers they read - or worse, they record the name they meant
// later, over a target that has moved again since.
//
// The second half is the one the loop shape decides: a good name followed by a
// typo must leave the record as it was, because the reading was one claim about
// both documents and neither half of it was made.
func TestAnUnknownDocumentNameIsAnError(t *testing.T) {
	bodies, out := fixture()

	record := reconciled{}
	if _, err := reconcileRecord(record, bodies, out, []string{"usecase/one-typo"}); err == nil {
		t.Fatal("recording a name no document has returned no error, so a reader is told their reading was written down when nothing was")
	} else if !strings.Contains(err.Error(), "usecase/one-typo") {
		t.Errorf("the error does not name the argument that was wrong: %v", err)
	}
	if len(record) != 0 {
		t.Errorf("a refused name left %v in the record", record)
	}

	if _, err := reconcileRecord(record, bodies, out, []string{"usecase/one", "brief/typo"}); err == nil {
		t.Fatal("a typo after a good name returned no error")
	}
	if len(record) != 0 {
		t.Errorf("a refused reading recorded %v for the names before the typo: half a reading is a claim nobody made", record)
	}

	// A target is not a source: `wiki/alpha` is pointed at and points at
	// nothing here, so recording it is a name that records nothing, and the
	// message says so rather than reporting a reading of zero pointers.
	if _, err := reconcileRecord(record, bodies, out, []string{"wiki/alpha"}); err == nil {
		t.Error("recording a document with no pointers of its own returned no error, so `reconcile` reports a reading that covered nothing as done")
	}
}

// mustRoot is this checkout, for the two tests that read the real corpus.
func mustRoot(t *testing.T) string {
	t.Helper()
	root, err := src.Root()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// 10. **A never-recorded pointer is named even when something has moved.**
//
// The report used to print the `never` rows only when `moved` was empty, so one
// moved pointer hid every unread one behind a count: the summary said "1 never
// recorded" and named neither which nor where. That is the failure this file is
// otherwise about - the number is right and the thing somebody would act on is
// gone - and it hid a pointer added the same afternoon, which is exactly when a
// pointer is least likely to have been read.
func TestANeverRecordedPointerIsNamedEvenWhenSomethingMoved(t *testing.T) {
	bodies, out := fixture()

	// One document read, the other never. Then the read one's target moves.
	record := reconciled{}
	if _, err := reconcileRecord(record, bodies, out, []string{"usecase/one"}); err != nil {
		t.Fatal(err)
	}
	bodies["wiki/alpha"] = "# Alpha\n\nthe alpha body, rewritten\n"

	text, never, moved := report(t, record, bodies, out)
	if len(moved) == 0 {
		t.Fatal("the moved target was not reported, so this proves nothing about hiding")
	}
	if len(never) == 0 {
		t.Fatal("the unread document was not reported as never")
	}
	for _, p := range never {
		if !strings.Contains(text, "never   "+p.from) {
			t.Errorf("an unrecorded pointer was counted and not named: %s -> %s\n%s", p.from, p.to, text)
		}
	}
	// And the never rows come first, because "nobody has read this" is the
	// stronger claim of the two and a long moved list is what buries it.
	if i, j := strings.Index(text, "never   "), strings.Index(text, "moved   "); i > j {
		t.Errorf("the moved rows are printed before the never rows:\n%s", text)
	}
}
