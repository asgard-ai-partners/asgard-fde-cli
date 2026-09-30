package platform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func stampFile(t *testing.T) string {
	t.Helper()
	// A directory that does not exist yet: the CLI has to make it.
	path := filepath.Join(t.TempDir(), "missing", ".asgard", "side-effect-at")
	t.Setenv(EnvSideEffectFile, path)
	fixed := time.Date(2026, 10, 1, 8, 0, 0, 123_000_000, time.UTC)
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = time.Now })
	return path
}

func okServer(t *testing.T) *Client {
	return workbenchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"number":3,"token":"x"}}`))
	})
}

// A change the platform accepted is stamped - as a JSON object, in a directory
// the CLI creates when it is not there.
func TestSideEffectIsStamped(t *testing.T) {
	path := stampFile(t)
	c := okServer(t)
	if _, err := c.CreateWorkbenchIssue(context.Background(), CreateWorkbenchIssueRequest{Title: "t", Type: "task"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no stamp after a write: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("stamp is not a JSON object: %q", raw)
	}
	if got["at"] != "2026-10-01T08:00:00.123Z" {
		t.Errorf("at = %v, want the call's time in UTC with milliseconds", got["at"])
	}

	// A second change replaces it, even when the file was deleted meanwhile.
	os.RemoveAll(filepath.Dir(path))
	now = func() time.Time { return time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC) }
	if _, err := c.CreateWorkbenchIssue(context.Background(), CreateWorkbenchIssueRequest{Title: "t", Type: "task"}); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(path); string(raw) != "{\"at\":\"2026-10-01T09:00:00.000Z\"}\n" {
		t.Errorf("second stamp = %q", raw)
	}
}

// Reads are not stamped, and neither are the POSTs that only read: minting a
// repository token (git asks for one on every fetch) and an audit query.
func TestReadsAreNotStamped(t *testing.T) {
	path := stampFile(t)
	c := okServer(t)
	ctx := context.Background()
	if _, err := c.GetWorkbenchIssue(ctx, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := c.MintRepositoryToken(ctx, "acme/repo", true); err != nil {
		t.Fatal(err)
	}
	rc, err := c.QueryAuditLog(ctx, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	rc.Close()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("a read wrote the stamp: %v", err)
	}
}

// A refused change changed nothing, so nothing is stamped.
func TestRefusedChangeIsNotStamped(t *testing.T) {
	path := stampFile(t)
	c := workbenchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"message":"no"}`))
	})
	if _, err := c.CreateWorkbenchIssue(context.Background(), CreateWorkbenchIssueRequest{Title: "t", Type: "task"}); err == nil {
		t.Fatal("expected the 403")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("a refused write stamped: %v", err)
	}
}

// Unset, nothing is written anywhere; unwritable, the command still succeeds.
func TestStampNeverFailsTheCommand(t *testing.T) {
	t.Setenv(EnvSideEffectFile, "")
	c := okServer(t)
	if _, err := c.CreateWorkbenchIssue(context.Background(), CreateWorkbenchIssueRequest{Title: "t", Type: "task"}); err != nil {
		t.Fatal(err)
	}

	// The target is a directory, so it cannot be replaced by a file.
	dir := t.TempDir()
	t.Setenv(EnvSideEffectFile, dir)
	if _, err := c.CreateWorkbenchIssue(context.Background(), CreateWorkbenchIssueRequest{Title: "t", Type: "task"}); err != nil {
		t.Errorf("an unwritable stamp failed the command: %v", err)
	}
}

// An upload goes through the raw path and is stamped like any other change.
func TestUploadIsStamped(t *testing.T) {
	path := stampFile(t)
	c := workbenchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"a1"}}`))
	})
	meta := WorkbenchAttachmentUpload{What: "minutes", From: "customer PM", Dated: "2026-10-01"}
	if _, err := c.UploadWorkbenchAttachment(context.Background(), 3, "m.pdf", strings.NewReader("x"), meta); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("an upload was not stamped: %v", err)
	}
}
