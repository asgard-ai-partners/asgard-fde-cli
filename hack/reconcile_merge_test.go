package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// **A recorded reading is append-only, and a read-modify-write is not.**
//
// `saveReconciled` used to load the record, change one document's entry and
// write the whole file back - so two readers working different slices at once,
// which is how a pass is actually done, silently lost whichever finished first.
// Three agents hit this in one sitting. **A record that drops readings is worse
// than no record**: it reports as owed the work somebody did, and the second
// time that happens nobody trusts any of it.
func TestSaveMergesRatherThanOverwrites(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "source"), 0o755); err != nil {
		t.Fatal(err)
	}

	// One reader records a document.
	if err := saveReconciled(root, reconciled{
		"wiki/a": {"usecase/x": "aaaa"},
	}); err != nil {
		t.Fatal(err)
	}
	// Another, holding a record from before the first one wrote, records a
	// different document. Its view of the file is stale - that is the race.
	if err := saveReconciled(root, reconciled{
		"wiki/b": {"usecase/y": "bbbb"},
	}); err != nil {
		t.Fatal(err)
	}

	got := load(t, root)
	if got["wiki/a"]["usecase/x"] != "aaaa" {
		t.Errorf("the first reader's record was lost: %v", got)
	}
	if got["wiki/b"]["usecase/y"] != "bbbb" {
		t.Errorf("the second reader's record was lost: %v", got)
	}

	// **A re-reading of the same pointer replaces it**, because the digest is
	// the whole point: an older digest left in place would report a document
	// as read against content it was not.
	if err := saveReconciled(root, reconciled{
		"wiki/a": {"usecase/x": "cccc"},
	}); err != nil {
		t.Fatal(err)
	}
	got = load(t, root)
	if got["wiki/a"]["usecase/x"] != "cccc" {
		t.Errorf("a re-reading did not replace the digest: %v", got)
	}
	if got["wiki/b"]["usecase/y"] != "bbbb" {
		t.Errorf("merging a re-read dropped an unrelated document: %v", got)
	}

	// Nothing is left beside the record: the write goes to a temporary name and
	// is renamed, so a reader never sees half a file.
	entries, err := os.ReadDir(filepath.Join(root, "source"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("the write left something beside the record: %v", entries)
	}
}

func load(t *testing.T, root string) reconciled {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, reconciledFile))
	if err != nil {
		t.Fatal(err)
	}
	r := reconciled{}
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r
}
