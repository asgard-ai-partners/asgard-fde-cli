package generate

import (
	"io/fs"
	"strings"
	"testing"
)

// Every template that renders a Syncer has to opt in to being fired by the
// rollout.
//
// The platform's apply step fires only the Syncers of the release that carry
// `asgard-ai.com/auto-fire-on-rollout: "true"`, and the two labels read nothing
// of each other: `syncer-suspend` stops the scheduler and the runner never
// looks at it. So a suspended Syncer without the opt-in never runs at all -
// silently, because the chart renders, the apiserver accepts the CR and the run
// succeeds. The skills or the drive simply resolve to nothing.
//
// If a future Syncer deliberately runs on its schedule only, it belongs in an
// exception here with the reason - not left to be read as an oversight.
func TestEveryGeneratedSyncerOptsInToTheRollout(t *testing.T) {
	const label = `asgard-ai.com/auto-fire-on-rollout: "true"`

	found := 0
	err := fs.WalkDir(templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := fs.ReadFile(templates, path)
		if err != nil {
			return err
		}
		text := string(body)
		if !strings.Contains(text, "kind: Syncer") {
			return nil
		}
		found++
		if !strings.Contains(text, label) {
			t.Errorf("%s renders a Syncer without %s, so nothing will ever run it", path, label)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
	if found == 0 {
		t.Fatal("no template renders a Syncer; this test has stopped checking anything")
	}
	t.Logf("%d Syncer-rendering template(s) checked", found)
}
