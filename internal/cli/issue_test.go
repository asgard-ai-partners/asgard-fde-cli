package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/feedback"
)

// TestIssueReportSend walks the route the help gives: --new, fill it in,
// --send. What --new writes is refused as it stands, and the same report with
// its TODOs answered reaches Sentry.
func TestIssueReportSend(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received = string(b)
	}))
	defer srv.Close()
	old := feedback.DSN
	feedback.DSN = strings.Replace(srv.URL, "http://", "http://k@", 1) + "/1"
	defer func() { feedback.DSN = old }()

	draft, _, err := runCLI(t, "", "issue-report", "--new")
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = runCLI(t, draft, "issue-report", "--send", "-")
	if err == nil || !strings.Contains(err.Error(), "section 1, 3, 4, 5") {
		t.Fatalf("an unfilled report has to be refused, naming its sections: %v", err)
	}
	if received != "" {
		t.Fatal("a refused report reached Sentry")
	}

	filled := strings.ReplaceAll(draft, "TODO - ", "")
	out, stderr, err := runCLI(t, filled, "issue-report", "--send", "-", "--email", "a@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Sent.") || !strings.Contains(stderr, "Sentry User Feedback") {
		t.Errorf("stdout %q, stderr %q", out, stderr)
	}
	if !strings.Contains(received, `"type":"feedback"`) || !strings.Contains(received, "## 2) The state I was in") {
		t.Errorf("Sentry received:\n%s", received)
	}
}

func TestIssueReportNewAndSendExclusive(t *testing.T) {
	if _, _, err := runCLI(t, "", "issue-report", "--new", "--send", "-"); err == nil {
		t.Error("--new with --send has to be refused")
	}
}
