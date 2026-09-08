package gate

import (
	"strings"
	"testing"
)

// A suspended Syncer with no auto-fire label never runs, and this is the only
// thing that says so. Every other check on that chart passes: it renders, the
// apiserver accepts the CR, the run succeeds, and the skills resolve to zero.
//
// The second and third cases are the reason this is not simply "warn on every
// Syncer without the label". A Syncer left on its schedule runs; not firing it
// at deploy time is a choice, and reporting it would train people to ignore the
// warning that matters.
func TestUnfiredSyncerWarning(t *testing.T) {
	syncer := func(name string, labels map[string]string) Doc {
		return Doc{Kind: "Syncer", Name: name, Labels: labels}
	}
	const (
		suspend  = annotationPrefix + "syncer-suspend"
		autoFire = annotationPrefix + "auto-fire-on-rollout"
	)

	for _, c := range []struct {
		name string
		doc  Doc
		warn bool
	}{
		{
			name: "suspended and not opted in never runs",
			doc:  syncer("syn-base", map[string]string{suspend: "true"}),
			warn: true,
		},
		{
			name: "suspended and opted in runs on every deploy",
			doc:  syncer("syn-base", map[string]string{suspend: "true", autoFire: "true"}),
			warn: false,
		},
		{
			name: "on its schedule and not opted in still runs",
			doc:  syncer("syn-drive", map[string]string{suspend: "false"}),
			warn: false,
		},
		{
			name: "no suspend label at all is the same as scheduled",
			doc:  syncer("syn-drive", nil),
			warn: false,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			res := Deployability([]Doc{c.doc}, Options{})
			got := false
			for _, w := range res.Warnings {
				if strings.Contains(w, "auto-fire-on-rollout") && strings.Contains(w, c.doc.Name) {
					got = true
				}
			}
			if got != c.warn {
				t.Fatalf("warned=%v, want %v; warnings: %q", got, c.warn, res.Warnings)
			}
		})
	}
}
