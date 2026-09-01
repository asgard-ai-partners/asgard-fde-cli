package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// walk visits cmd and every command below it.
func walk(cmd *cobra.Command, fn func(*cobra.Command)) {
	fn(cmd)
	for _, sub := range cmd.Commands() {
		walk(sub, fn)
	}
}

// args turns a command path such as "asgard-cli init" into the arguments needed
// to reach that command from the root.
func args(cmd *cobra.Command) []string {
	return strings.Fields(cmd.CommandPath())[1:]
}

// TestEveryCommandDocumentsItself is the guard behind the rule in AGENTS.md: a
// command that cannot explain itself is not finished. cobra gives every command
// a --help flag for free, but nothing stops the help text from being empty, so
// this walks the tree and insists on real content.
func TestEveryCommandDocumentsItself(t *testing.T) {
	walk(NewRootCmd(), func(cmd *cobra.Command) {
		t.Run(cmd.CommandPath(), func(t *testing.T) {
			if cmd.Short == "" {
				t.Error("Short is empty; every command needs a one-line summary")
			}
			if cmd.Long == "" {
				t.Error("Long is empty; every command needs a description of what it does")
			}
			// A Long that just repeats Short teaches the reader nothing.
			if cmd.Long != "" && len(cmd.Long) <= len(cmd.Short) {
				t.Errorf("Long (%d chars) is no longer than Short (%d chars); it should explain behaviour, defaults and failure modes",
					len(cmd.Long), len(cmd.Short))
			}

			cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if f.Usage == "" {
					t.Errorf("flag --%s has no usage text", f.Name)
				}
			})
		})
	})
}

// TestHelpRunsForEveryCommand checks that --help actually works everywhere,
// rather than trusting that cobra wired it up.
func TestHelpRunsForEveryCommand(t *testing.T) {
	var paths [][]string
	walk(NewRootCmd(), func(cmd *cobra.Command) {
		paths = append(paths, args(cmd))
	})

	for _, path := range paths {
		name := strings.Join(append([]string{"asgard-cli"}, path...), " ")
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			root := NewRootCmd()
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(append(path, "--help"))

			if err := root.Execute(); err != nil {
				t.Fatalf("--help: %v", err)
			}

			got := out.String()
			if !strings.Contains(got, "Usage:") {
				t.Errorf("help output has no usage section:\n%s", got)
			}
			if !strings.Contains(got, name) {
				t.Errorf("help output does not name the command %q:\n%s", name, got)
			}
		})
	}
}
