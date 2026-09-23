//go:build windows

package selfupdate

import (
	"os"
	"path/filepath"
	"testing"
)

// **Windows cannot overwrite or delete a file that is executing, but it can
// rename one**, so replacing the binary is two renames rather than the single
// one every other platform needs. This is the test of that path, and it only
// exists because the path only exists here: the Unix half is exercised by every
// other run and this one was never exercised anywhere until CI grew a Windows
// job.
func TestReplaceMovesTheOldBinaryAsideAndSweepsItLater(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "asgard-cli.exe")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(dir, ".staged")
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replace(staged, target); err != nil {
		t.Fatalf("replace: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("target holds %q, want %q", got, "new")
	}

	// The displaced one is still there, because a running .exe cannot be
	// deleted. That is the cost of the two-step, and it is swept rather than
	// left for ever.
	old := target + oldSuffix
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("the old binary was not kept at %s: %v", old, err)
	}
	if got, _ := os.ReadFile(old); string(got) != "old" {
		t.Errorf("%s holds %q, want %q", old, got, "old")
	}

	SweepOld(target)
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("%s survived SweepOld, so it would accumulate on every update", old)
	}
}

// A second update must not trip over the file the first one left behind.
func TestASecondUpdateReplacesTheFileTheFirstLeftBehind(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "asgard-cli.exe")
	for _, body := range []string{"one", "two", "three"} {
		staged := filepath.Join(dir, ".staged")
		if err := os.WriteFile(staged, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(target); os.IsNotExist(err) {
			if err := os.WriteFile(target, []byte("first"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := replace(staged, target); err != nil {
			t.Fatalf("replace to %q: %v", body, err)
		}
		if got, _ := os.ReadFile(target); string(got) != body {
			t.Fatalf("target holds %q, want %q", got, body)
		}
	}
}
