package gitrepo

import "testing"

// The shapes a remote actually comes in, and the ones that must be refused.
//
// Refusal is the point. A remote this cannot reduce is reported as such; a
// remote it reduces WRONGLY does not fail, it succeeds at the wrong thing —
// and every caller downstream then acts on a name nobody chose.
func TestFullNameAndHost(t *testing.T) {
	for _, c := range []struct {
		url      string
		wantFull string
		wantHost string
	}{
		{"git@github.com:acme/app.git", "acme/app", "github.com"},
		{"https://github.com/acme/app.git", "acme/app", "github.com"},
		{"https://github.com/acme/app", "acme/app", "github.com"},
		{"ssh://git@github.com/acme/app.git", "acme/app", "github.com"},
		{"https://user:tok@github.com/acme/app", "acme/app", "github.com"},
		{"https://GitHub.com/acme/app.git", "acme/app", "github.com"},
		{"git@bitbucket.org:acme/app.git", "acme/app", "bitbucket.org"},

		// A port is a port. The old pattern read the colon as a separator and
		// answered "2222" as the owner.
		{"ssh://git@ghe.internal:2222/acme/app.git", "acme/app", "ghe.internal"},

		// Deeper than owner/name: a GitLab subgroup is a real remote and a
		// shape a pipeline cannot bind, so it is refused rather than folded.
		{"https://gitlab.com/group/subgroup/app.git", "", "gitlab.com"},

		// Not provider remotes at all.
		{"/srv/git/app.git", "", ""},
		{"../sibling.git", "", ""},
		{"file:///srv/git/acme/app.git", "", ""},
		{"", "", ""},
		{"git@github.com:acme", "", "github.com"},
		{"https://github.com/acme", "", "github.com"},
	} {
		gotFull, ok := FullName(c.url)
		if c.wantFull == "" {
			if ok {
				t.Errorf("FullName(%q) = %q, want refused", c.url, gotFull)
			}
		} else if !ok || gotFull != c.wantFull {
			t.Errorf("FullName(%q) = %q,%v; want %q", c.url, gotFull, ok, c.wantFull)
		}
		if got := RemoteHost(c.url); got != c.wantHost {
			t.Errorf("RemoteHost(%q) = %q, want %q", c.url, got, c.wantHost)
		}
	}
}
