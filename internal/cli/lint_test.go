package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

// runLint runs lint in dir and returns its output and error.
func runLint(t *testing.T, dir string) (string, error) {
	t.Helper()
	t.Chdir(dir)

	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"lint"})

	err := root.Execute()
	return out.String(), err
}

// writeConfigJSON writes raw JSON, to build broken configs init would never produce.
func writeConfigJSON(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestLintPasses(t *testing.T) {
	dir := t.TempDir()
	writeConfigJSON(t, dir, `{"workspace":{"id":"ws_1","name":"demo"}}`)

	out, err := runLint(t, dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if !strings.HasPrefix(out, "ok  ") {
		t.Errorf("output %q should start with the success marker", out)
	}
}

func TestLintFailsWhenConfigMissing(t *testing.T) {
	// t.TempDir() sits under the filesystem root, so Find walks all the way up
	// without finding anything.
	out, err := runLint(t, t.TempDir())
	if err == nil {
		t.Fatal("want failure when the config is missing, got success")
	}
	if !strings.Contains(err.Error(), config.FileName) {
		t.Errorf("error %q should mention %s", err, config.FileName)
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error %q should tell the user to run init first", err)
	}
	if out != "" {
		t.Errorf("a failed run should print nothing to stdout, got %q", out)
	}
}

func TestLintFailsWhenWorkspaceMissing(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantFields  []string
		wantProblem int
	}{
		{
			name:        "no workspace field at all",
			json:        `{}`,
			wantFields:  []string{"workspace.id", "workspace.name"},
			wantProblem: 2,
		},
		{
			name:        "empty workspace object",
			json:        `{"workspace":{}}`,
			wantFields:  []string{"workspace.id", "workspace.name"},
			wantProblem: 2,
		},
		{
			name:        "id missing",
			json:        `{"workspace":{"name":"demo"}}`,
			wantFields:  []string{"workspace.id"},
			wantProblem: 1,
		},
		{
			name:        "name missing",
			json:        `{"workspace":{"id":"ws_1"}}`,
			wantFields:  []string{"workspace.name"},
			wantProblem: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeConfigJSON(t, dir, tt.json)

			out, err := runLint(t, dir)
			if err == nil {
				t.Fatalf("want failure, got success; output %q", out)
			}
			for _, field := range tt.wantFields {
				if !strings.Contains(out, field) {
					t.Errorf("output %q should point at %s", out, field)
				}
			}
			if got := countProblemLines(out); got != tt.wantProblem {
				t.Errorf("problem count = %d, want %d; output %q", got, tt.wantProblem, out)
			}
		})
	}
}

func TestLintFailsOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeConfigJSON(t, dir, "{ not json")

	if _, err := runLint(t, dir); err == nil {
		t.Fatal("want failure for broken JSON, got success")
	}
}

func TestLintFindsConfigInParentDir(t *testing.T) {
	root := t.TempDir()
	writeConfigJSON(t, root, `{"workspace":{"id":"ws_1","name":"demo"}}`)

	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if _, err := runLint(t, nested); err != nil {
		t.Fatalf("lint from a subdirectory should find the config above it: %v", err)
	}
}

// countProblemLines counts the "- " prefixed lines lint prints, one per problem.
func countProblemLines(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "- ") {
			n++
		}
	}
	return n
}
