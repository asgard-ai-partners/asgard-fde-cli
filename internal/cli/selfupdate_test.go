package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/selfupdate"
)

// The background check asks on an ordinary command and stays out of the way of
// the three runs where an answer is somebody else's or nobody's.
//
// **`version` is the one that matters.** It carries `--check`, which asks the
// same question with its own wording, and a second answer printed underneath
// that one reads as a second question rather than as the same one twice.
func TestWhichRunsAskWhetherANewerReleaseExists(t *testing.T) {
	root := NewRootCmd()

	// cobra adds `help` and `completion` on the first run rather than when the
	// tree is built, so they are asked for by name here.
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()

	find := func(path ...string) *cobra.Command {
		c := root
		for _, name := range path {
			next, _, err := c.Find([]string{name})
			if err != nil || next == c {
				t.Fatalf("no command %q under %q", name, c.Name())
			}
			c = next
		}
		return c
	}

	cases := []struct {
		name string
		cmd  *cobra.Command
		want bool
	}{
		{"an ordinary command asks", find("render"), true},
		{"a subcommand of one asks", find("project", "add"), true},
		{"version does not, because --check is its own answer", find("version"), false},
		{"help does not", find("help"), false},
		{"a completion run is a shell's, not a person's", find("completion"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := wantsUpdateCheck(tc.cmd); got != tc.want {
				t.Fatalf("wantsUpdateCheck(%s) = %v, want %v", tc.cmd.Name(), got, tc.want)
			}
		})
	}

	t.Run("the off switch stops it on every command", func(t *testing.T) {
		t.Setenv(selfupdate.EnvDisable, "1")
		if wantsUpdateCheck(find("render")) {
			t.Fatalf("%s set and the check still wanted to run", selfupdate.EnvDisable)
		}
	})
}

// A check that has not answered by the time the command has is dropped rather
// than waited for. The record it writes is what carries the answer to the next
// run, so nothing is lost by not waiting - and the command the user actually
// ran is not held up by a question they did not ask.
func TestAnUnfinishedCheckDoesNotHoldTheCommandUp(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetErr(discard{})

	// Never written to, and never closed: the goroutine is still in flight.
	stalled := updateCheck{
		result:   make(chan selfupdate.Result),
		deadline: time.Now().Add(backgroundLeash),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		warnIfNewerRelease(cmd, stalled)
	}()

	select {
	case <-done:
	case <-time.After(backgroundLeash + 2*time.Second):
		t.Fatal("warnIfNewerRelease waited past its own leash")
	}
}

// A zero updateCheck is what startUpdateCheck returns when it decided not to
// ask, and reading its nil channel blocks for ever. The caller has to
// recognise it - and so does a deadline of the zero time, which has passed.
func TestNoCheckWasStartedAndNothingIsWaitedFor(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetErr(discard{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		warnIfNewerRelease(cmd, updateCheck{})
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("warnIfNewerRelease blocked on a check that was never started")
	}
}

// The help screen quotes the interval and the off switch rather than repeating
// them, so moving either moves what the screen says.
//
// **A number typed beside a constant is the drift this repository removes
// everywhere else.** The interval was written out as "two hours" in the Long,
// in the flag's usage and in both READMEs; the two the program renders now read
// the declaration.
func TestTheHelpScreenQuotesTheConstantsRatherThanRepeatingThem(t *testing.T) {
	root := NewRootCmd()
	cmd, _, err := root.Find([]string{"version"})
	if err != nil {
		t.Fatalf("no version command: %v", err)
	}
	screen := cmd.Long + "\n" + cmd.Flags().FlagUsages()

	for _, want := range []string{selfupdate.IntervalText(), selfupdate.EnvDisable} {
		if !strings.Contains(screen, want) {
			t.Errorf("`version --help` does not say %q:\n%s", want, screen)
		}
	}
	if strings.Contains(screen, "two hours") {
		t.Errorf("`version --help` writes the interval out by hand as well as reading it:\n%s", screen)
	}
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
