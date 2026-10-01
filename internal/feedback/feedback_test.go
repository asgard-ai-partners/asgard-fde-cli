package feedback

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestSend checks the envelope Sentry receives: a feedback item whose message
// carries the report, and the whole report as an attachment.
func TestSend(t *testing.T) {
	var got []byte
	var auth, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		auth, path = r.Header.Get("X-Sentry-Auth"), r.URL.Path
		w.Write([]byte(`{"id":"x"}`))
	}))
	defer srv.Close()
	old := DSN
	DSN = strings.Replace(srv.URL, "http://", "http://publickey@", 1) + "/42"
	defer func() { DSN = old }()

	body := "## 1) What I was trying to do\n\nbuild a deck\n"
	id, err := Send(context.Background(), srv.Client(), Report{Body: body, Email: "a@example.com", Release: "asgard-cli v1"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/42/envelope/" || !strings.Contains(auth, "sentry_key=publickey") {
		t.Errorf("posted to %q with auth %q", path, auth)
	}

	lines := bufio.NewScanner(bytes.NewReader(got))
	var items []map[string]any
	for lines.Scan() {
		var m map[string]any
		if err := json.Unmarshal(lines.Bytes(), &m); err == nil {
			items = append(items, m)
		}
	}
	if len(items) != 4 {
		t.Fatalf("want envelope header, feedback header, event, attachment header; got %d JSON lines:\n%s", len(items), got)
	}
	if items[0]["event_id"] != id || items[1]["type"] != "feedback" || items[3]["type"] != "attachment" {
		t.Errorf("envelope shape: %v", items)
	}
	fb := items[2]["contexts"].(map[string]any)["feedback"].(map[string]any)
	if fb["message"] != body || fb["contact_email"] != "a@example.com" {
		t.Errorf("feedback context: %v", fb)
	}
	if !bytes.Contains(got, []byte(body)) {
		t.Error("the attachment does not carry the report")
	}
}

func TestSendRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "project disabled", http.StatusForbidden)
	}))
	defer srv.Close()
	old := DSN
	DSN = strings.Replace(srv.URL, "http://", "http://k@", 1) + "/1"
	defer func() { DSN = old }()

	if _, err := Send(context.Background(), srv.Client(), Report{Body: "x"}); err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("a refusal has to surface with its status: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	long := strings.Repeat("字", MaxMessage+10)
	cut := truncate(long, MaxMessage)
	if n := utf8.RuneCountInString(cut); n != MaxMessage {
		t.Errorf("cut to %d code points, want %d", n, MaxMessage)
	}
	if !strings.Contains(cut, AttachmentName) {
		t.Error("a cut message has to say where the rest is")
	}
	if truncate("short", MaxMessage) != "short" {
		t.Error("a short message is not cut")
	}
}
