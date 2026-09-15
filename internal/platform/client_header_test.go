package platform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
)

// Every call says which client is making it, and the one that matters is three
// hops away from here.
//
// `pipeline connect` ends in a browser, on a page the web console's own connect
// flow ends on too, and that page offers a way back INTO the console. Right for
// somebody who started there, wrong for somebody waiting at this terminal. The
// platform can only tell the two apart because this header said so at the
// start - so a call that stops sending it fails nothing, and quietly walks the
// next person into the wrong flow.
func TestEveryRequestSaysWhichClientIsCalling(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"ok","data":{}}`))
	}))
	defer srv.Close()

	c := &Client{
		profile:   auth.Profile{Name: "test", PlatformAPI: srv.URL},
		token:     "token",
		workspace: "ws-1",
		http:      srv.Client(),
	}
	err := c.do(context.Background(), request{
		method: http.MethodPost,
		path:   "/v1/iac/connections/begin-github-attach",
	})
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if v := got.Get(ClientHeader); v != ClientName {
		t.Errorf("%s = %q, want %q", ClientHeader, v, ClientName)
	}
	// The call that takes no workspace sends it too: what this program is does
	// not depend on which route it is calling.
	got = nil
	if err := c.do(context.Background(), request{
		method:      http.MethodGet,
		path:        "/v1/workspaces",
		noWorkspace: true,
	}); err != nil {
		t.Fatalf("do: %v", err)
	}
	if v := got.Get(ClientHeader); v != ClientName {
		t.Errorf("unscoped %s = %q, want %q", ClientHeader, v, ClientName)
	}
}
