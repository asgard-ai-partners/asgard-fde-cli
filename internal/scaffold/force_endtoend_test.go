package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
)

// **`--force` must leave a repository that still renders.**
//
// This is the test that was missing. `.asgard-pipeline.yaml` was made an
// accumulator with a test asserting the accumulator set - the mechanism, not
// the outcome - and `projects/<slug>/chart/app/values.yaml` was not on it.
// `asgard-cli add` appends the keys each CR reads to that file, so after one
// `add` the templates beside it depend on what is in it; regenerating it left a
// chart reading `.Values.dbDB.host` with the block gone, which `helm template`
// fails on with a nil pointer.
//
// `asgard-cli check` reported `ok` throughout, because the structure was
// intact. The failure is one command further on, which is exactly why this
// asserts what survives rather than what the list contains.
func TestForceLeavesWhatAddWroteIntoTheChart(t *testing.T) {
	root := t.TempDir()
	if _, err := scaffold.Write(root, nil, false); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := scaffold.Write(root, []string{"erp"}, false); err != nil {
		t.Fatalf("project add: %v", err)
	}

	var dc generate.Kind
	for _, k := range generate.Kinds {
		if k.Name == "dataconnector" {
			dc = k
		}
	}
	if dc.Name == "" {
		t.Fatal("no dataconnector kind")
	}
	if _, err := generate.Write(root, dc, generate.Options{Project: "erp", Name: "db"}); err != nil {
		t.Fatalf("add dataconnector: %v", err)
	}

	values := filepath.Join(root, "projects", "erp", "chart", "app", "values.yaml")
	before, err := os.ReadFile(values)
	if err != nil {
		t.Fatalf("read values: %v", err)
	}
	if !strings.Contains(string(before), "dbDB") {
		t.Fatalf("add wrote no values block, so this test proves nothing:\n%s", before)
	}

	if _, err := scaffold.Write(root, []string{"erp"}, true); err != nil {
		t.Fatalf("init --force: %v", err)
	}

	after, err := os.ReadFile(values)
	if err != nil {
		t.Fatalf("read values after force: %v", err)
	}
	if !strings.Contains(string(after), "dbDB") {
		t.Errorf("--force discarded the values block `add` wrote, and the CR beside it still reads it:\n%s", after)
	}
}
