// Package feedback sends a report about this tool to its maintainers, as a
// Sentry User Feedback.
//
// It is the channel behind `asgard-cli issue-report --send`, and it is not the
// repository's issue tracker because that repository is public: a report is
// written by somebody in the middle of an engagement, and a Sentry User
// Feedback is read by the maintainers alone.
//
// sentry-go has no call for a feedback item, so the envelope is built here and
// sentry-go supplies what is Sentry's to define: how a DSN is parsed, which
// endpoint it names, and the client string the auth header carries.
package feedback

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
)

// DSN is the maintainers' Sentry project. A DSN is a public identifier by
// design - it is what every Sentry client ships - so it is a constant rather
// than a secret, and a variable only so a test can point it elsewhere.
var DSN = "https://eda4c7b15041a7effe1a5208ef686f30@o4511098939506688.ingest.us.sentry.io/4512179102875648"

// MaxMessage is Sentry's limit on a feedback's message, in code points. A
// report longer than that is cut in the message and sent whole as an
// attachment, so nothing written is lost.
const MaxMessage = 4096

// AttachmentName is the file the whole report arrives as.
const AttachmentName = "report.md"

// Report is one feedback.
type Report struct {
	Body    string // the report, as `issue-report --new` writes it
	Email   string // optional: where the maintainers can answer
	Name    string // optional
	Release string // `asgard-cli@<version>`: Sentry drops a release with a "/" in it, so not the version line
}

// Send delivers r and returns the event id Sentry files it under.
func Send(ctx context.Context, client *http.Client, r Report) (string, error) {
	dsn, err := sentry.NewDsn(DSN)
	if err != nil {
		return "", fmt.Errorf("parse feedback DSN: %w", err)
	}
	id, err := eventID()
	if err != nil {
		return "", err
	}
	envelope, err := encode(id, dsn.String(), r, time.Now().UTC())
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dsn.GetAPIURL().String(), bytes.NewReader(envelope))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-sentry-envelope")
	req.Header.Set("X-Sentry-Auth", fmt.Sprintf("Sentry sentry_version=7, sentry_client=sentry.go/%s, sentry_key=%s",
		sentry.SDKVersion, dsn.GetPublicKey()))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send feedback: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("send feedback: Sentry answered %s: %s", resp.Status, bytes.TrimSpace(msg))
	}
	return id, nil
}

// encode builds the envelope: its header, the feedback item, and the whole
// report as an attachment. The format is
// https://develop.sentry.dev/sdk/telemetry/feedbacks/.
func encode(id, dsn string, r Report, now time.Time) ([]byte, error) {
	feedback := map[string]any{"message": truncate(r.Body, MaxMessage), "source": "asgard-cli issue-report"}
	if r.Email != "" {
		feedback["contact_email"] = r.Email
	}
	if r.Name != "" {
		feedback["name"] = r.Name
	}
	event := map[string]any{
		"event_id":  id,
		"timestamp": float64(now.UnixMilli()) / 1000,
		"platform":  "go",
		"level":     "info",
		"release":   r.Release,
		"contexts":  map[string]any{"feedback": feedback},
		"tags":      map[string]string{"os": runtime.GOOS, "arch": runtime.GOARCH},
	}

	var b bytes.Buffer
	line := func(v any) error {
		j, err := json.Marshal(v)
		if err != nil {
			return err
		}
		b.Write(j)
		b.WriteByte('\n')
		return nil
	}
	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	for _, v := range []any{
		map[string]string{"event_id": id, "sent_at": now.Format(time.RFC3339), "dsn": dsn},
		map[string]any{"type": "feedback", "length": len(body)},
	} {
		if err := line(v); err != nil {
			return nil, err
		}
	}
	b.Write(body)
	b.WriteByte('\n')
	if err := line(map[string]any{"type": "attachment", "length": len(r.Body),
		"filename": AttachmentName, "content_type": "text/markdown"}); err != nil {
		return nil, err
	}
	b.WriteString(r.Body)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

// truncate cuts s to at most n code points, saying so at the end when it did.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	note := []rune("\n\n[cut at Sentry's limit - the whole report is attached as " + AttachmentName + "]")
	return string(runes[:n-len(note)]) + string(note)
}

func eventID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("event id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
