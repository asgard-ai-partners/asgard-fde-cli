package selfupdate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// **A wrong "you are behind" is acted on; a missed one costs a delay.** So the
// comparison reports nothing whenever it is not sure, and these are the cases
// where being sure is the hard part.
func TestNewerIsSilentWhenItCannotBeSure(t *testing.T) {
	for _, c := range []struct {
		running, latest string
		want            bool
		why             string
	}{
		{"0.1.2", "0.1.3", true, "an ordinary upgrade"},
		{"0.1.2", "0.2.0", true, "minor bump"},
		{"0.9.9", "1.0.0", true, "major bump"},
		{"0.1.3", "0.1.3", false, "already current"},
		{"0.1.4", "0.1.3", false, "running ahead of the last release"},
		{"0.1.10", "0.1.9", false, "10 is after 9, not before it - a string compare gets this wrong"},
		{"0.1.9", "0.1.10", true, "and the same case the other way"},
		// The one that would otherwise nag on every run of a dev build.
		{"0.1.3-0.20260922031239-8e73d59+dirty", "0.1.2", false, "a development build is not behind a release"},
		{"0.1.3-0.20260922031239-8e73d59+dirty", "0.1.9", false, "even when the numbers say so"},
		{"", "0.1.3", false, "no running version"},
		{"0.1.2", "", false, "no answer from the last check"},
		{"weird", "0.1.3", false, "unparseable"},
		{"0.1.2", "not.a.version", false, "unparseable answer"},
		{"0.1", "0.1.3", false, "too few parts to compare"},
	} {
		if got := newer(c.running, c.latest); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v - %s", c.running, c.latest, got, c.want, c.why)
		}
	}
}

// **The point of the record is that the question is not asked again**, and the
// test that matters is the one proving a fresh record stops the call rather
// than the one proving a call happens.
func TestAFreshRecordAsksNothing(t *testing.T) {
	home := t.TempDir()
	if err := save(Path(home), Record{CheckedAt: time.Now(), Latest: "9.9.9"}); err != nil {
		t.Fatal(err)
	}

	res := Check(context.Background(), home, "0.1.2", false)
	if res.Asked {
		t.Error("a record written seconds ago was treated as stale")
	}
	if !res.Newer || res.Latest != "9.9.9" {
		t.Errorf("the cached answer was not used: %+v", res)
	}

	// Older than the interval, and it asks again. No network here, so what is
	// asserted is that it tried and still returned rather than failing.
	old := Record{CheckedAt: time.Now().Add(-Interval - time.Minute), Latest: "9.9.9"}
	if err := save(Path(home), old); err != nil {
		t.Fatal(err)
	}
	if res := Check(context.Background(), home, "0.1.2", false); !res.Asked {
		t.Error("a record older than the interval was treated as fresh")
	}
}

// **A failed check is recorded too.** Otherwise a machine with no network asks
// on every single command, which is the cost this file exists to remove - and
// that machine is exactly the one that can least afford it.
func TestAFailedCheckStillStampsTheTime(t *testing.T) {
	home := t.TempDir()
	t.Setenv(EnvDisable, "")

	// No record, and the network call will fail or succeed - either way the
	// time has to move, and a subsequent call must not ask again.
	_ = Check(context.Background(), home, "0.1.2", false)
	rec, err := load(Path(home))
	if err != nil {
		t.Fatalf("no record was written: %v", err)
	}
	if rec.CheckedAt.IsZero() {
		t.Error("the check did not stamp a time, so the next command asks again")
	}
	if res := Check(context.Background(), home, "0.1.2", false); res.Asked {
		t.Error("it asked twice in a row")
	}
}

// The off switch is off, and `--check` still works for somebody who asks
// directly - the setting is about background calls, not about refusing to
// answer a question.
func TestTheOffSwitchStopsTheBackgroundCheckOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv(EnvDisable, "1")

	if res := Check(context.Background(), home, "0.1.2", false); res.Asked {
		t.Error("disabled, and it still asked")
	}
	if _, err := os.Stat(filepath.Join(home, FileName)); err == nil {
		t.Error("disabled, and it still wrote a record")
	}
	if res := Check(context.Background(), home, "0.1.2", true); !res.Asked {
		t.Error("asked directly with --check, and it refused")
	}
}
