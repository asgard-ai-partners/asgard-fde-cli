package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// connectFake is the platform side of the sandbox connect flow. The member has
// authorized and reaches one installation, acme-org; the workspace holds none.
type connectFake struct {
	begins, installs, attaches atomic.Int32
	attachedInstallation       atomic.Value
}

func (f *connectFake) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/iac/connections") && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"success":true,"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/begin-github-attach"):
			f.begins.Add(1)
			_, _ = io.WriteString(w, `{"success":true,"data":{"authorize_url":"https://github.com/login/oauth/authorize?state=S1","state":"S1"}}`)
		case strings.HasSuffix(r.URL.Path, "/begin-github-install"):
			f.installs.Add(1)
			_, _ = io.WriteString(w, `{"success":true,"data":{"install_url":"https://github.com/apps/x/installations/new?state=S2","state":"S2"}}`)
		case strings.HasSuffix(r.URL.Path, "/install-status"):
			_, _ = io.WriteString(w, `{"success":true,"data":{"outcome":"attach_choices","attach_state":"A1",`+
				`"installations":[{"installation_id":"77","account_login":"acme-org","account_type":"Organization","repository_selection":"all"}]}}`)
		case strings.HasSuffix(r.URL.Path, "/connections/attach"):
			f.attaches.Add(1)
			body, _ := io.ReadAll(r.Body)
			f.attachedInstallation.Store(string(body))
			_, _ = io.WriteString(w, `{"success":true,"data":{"connection_id":"c-new","account_login":"acme-org","account_type":"Organization","status":"active"}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"success":false,"message":"not here"}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// No account and no origin remote: no link is started, and the agent is told
// to ask the member which account - the authorization that used to be started
// here ended in "which account?" after the member had clicked.
func TestSandboxConnectWithoutAnAccountStartsNothing(t *testing.T) {
	f := &connectFake{}
	sandboxEnv(t, f.server(t).URL)
	t.Chdir(t.TempDir())

	out, _, err := runCLI(t, "", "pipeline", "connect")
	if err == nil || !strings.Contains(err.Error(), "Ask the member which GitHub organisation") ||
		!strings.Contains(err.Error(), "pipeline connect --account <login>") {
		t.Fatalf("connect without an account: %q %v", out, err)
	}
	if f.begins.Load() != 0 {
		t.Errorf("an authorization was started without an account")
	}
	if _, err := os.Stat(connectPendingPath()); !os.IsNotExist(err) {
		t.Errorf("a pending flow was recorded: %v", err)
	}
}

// A flow that started without an account (an older CLI, or a state written by
// hand) is finished by --continue --account, on the same authorization: the
// error says exactly that, and following it attaches without a second link.
func TestContinueTakesTheAccount(t *testing.T) {
	f := &connectFake{}
	sandboxEnv(t, f.server(t).URL)
	t.Chdir(t.TempDir())
	if err := saveConnectPending(&connectPending{Workspace: "ws-sb", Stage: "attach", State: "S1", StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := runCLI(t, "", "pipeline", "connect", "--continue", "--wait", "10s")
	if err == nil || !strings.Contains(err.Error(), "authorization still stands") ||
		!strings.Contains(err.Error(), "pipeline connect --continue --account <login>") {
		t.Fatalf("continue without an account: %v", err)
	}
	if !strings.Contains(errOut, "acme-org (Organization)") || !strings.Contains(errOut, "To connect a different") {
		t.Errorf("the list did not say another account is possible: %q", errOut)
	}

	out, _, err := runCLI(t, "", "pipeline", "connect", "--continue", "--account", "acme-org", "--wait", "10s")
	if err != nil {
		t.Fatalf("continue --account: %v", err)
	}
	if !strings.Contains(out, "c-new") || f.attaches.Load() != 1 || f.begins.Load() != 0 {
		t.Errorf("out %q, attaches %d, begins %d", out, f.attaches.Load(), f.begins.Load())
	}
	if body, _ := f.attachedInstallation.Load().(string); !strings.Contains(body, "77") {
		t.Errorf("attached %q, want installation 77", body)
	}
	if _, err := os.Stat(connectPendingPath()); !os.IsNotExist(err) {
		t.Errorf("pending state left behind after connecting: %v", err)
	}
}

// An account the member does not reach yet is not a dead end: --continue
// --account <it> hands over the app's install page, where GitHub asks which
// account, and records that account to check the install against.
func TestContinueWithAnUnlistedAccountInstalls(t *testing.T) {
	f := &connectFake{}
	sandboxEnv(t, f.server(t).URL)
	t.Chdir(t.TempDir())
	if err := saveConnectPending(&connectPending{Workspace: "ws-sb", Stage: "attach", State: "S1", StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	out, _, err := runCLI(t, "", "pipeline", "connect", "--continue", "--account", "other-org", "--wait", "10s")
	if err != nil {
		t.Fatalf("continue --account other-org: %v", err)
	}
	if !strings.Contains(out, "installations/new?state=S2") || f.installs.Load() != 1 || f.attaches.Load() != 0 {
		t.Errorf("out %q, installs %d, attaches %d", out, f.installs.Load(), f.attaches.Load())
	}
	p, err := loadConnectPending()
	if err != nil || p.Stage != "install" || p.Account != "other-org" {
		t.Errorf("pending %+v %v", p, err)
	}
}
