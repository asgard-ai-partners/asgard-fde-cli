package chart

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, project, name, body string) {
	t.Helper()
	path := filepath.Join(TemplatesDir(root, project), name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// TestScanReadsUnrenderedTemplates is the point of the package: these files are
// Helm templates, so they are not valid YAML and cannot be parsed as such - and
// `asgard-cli add` needs to know what a chart holds before helm is involved.
func TestScanReadsUnrenderedTemplates(t *testing.T) {
	root := t.TempDir()
	write(t, root, "erp", "sl.yaml", `apiVersion: asgard-ai.com/v1alpha1
kind: SemanticLayer
metadata:
  name: sl-erp
  labels:
    {{- include "erp.labels" . | nindent 4 }}
spec:
  dataConnectorName: dc-erp
  defaultCompletionModelName: {{ .Values.defaultCompletionModelName | quote }}
`)
	// Several documents in one file, which is how a chain of related CRs is
	// written.
	write(t, root, "erp", "chain.yaml", `apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: wf-stock
---
apiVersion: asgard-ai.com/v1alpha1
kind: Toolset
metadata:
  name: ts-catalog
`)

	refs, err := Scan(root, "erp")
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(refs) != 3 {
		t.Fatalf("found %d refs, want 3: %+v", len(refs), refs)
	}

	if got := NamesOf(refs, "SemanticLayer"); len(got) != 1 || got[0] != "sl-erp" {
		t.Errorf("SemanticLayer names = %v", got)
	}
	if got := NamesOf(refs, "Toolset"); len(got) != 1 || got[0] != "ts-catalog" {
		t.Errorf("Toolset names = %v", got)
	}
	counts := Counts(refs)
	if counts["Workflow"] != 1 || counts["Toolset"] != 1 || counts["SemanticLayer"] != 1 {
		t.Errorf("counts = %v", counts)
	}
}

func TestScanOfAProjectWithNoChart(t *testing.T) {
	// Normal early on, and not an error.
	refs, err := Scan(t.TempDir(), "erp")
	if err != nil || len(refs) != 0 {
		t.Errorf("Scan = %v, %v; want none and no error", refs, err)
	}
}

func TestNamesOfSortsSoPickingTheOnlyOneIsDeterministic(t *testing.T) {
	refs := []Ref{
		{Kind: "SemanticLayer", Name: "sl-b"},
		{Kind: "SemanticLayer", Name: "sl-a"},
		{Kind: "Agent", Name: "ag-x"},
		// A templated name cannot be resolved, so it is skipped rather than
		// reported as a name that does not exist.
		{Kind: "SemanticLayer", Name: ""},
	}
	got := NamesOf(refs, "SemanticLayer")
	if len(got) != 2 || got[0] != "sl-a" || got[1] != "sl-b" {
		t.Errorf("NamesOf = %v, want [sl-a sl-b]", got)
	}
}
