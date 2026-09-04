// Package cli assembles the asgard-cli command tree.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

// NewRootCmd builds the root command. Every call returns a fresh tree so tests
// cannot interfere with one another.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asgard-cli",
		Short: "Command line tool for Asgard FDE",
		Long: `asgard-cli is what an agent asks about integrating with Asgard.

It does two things, and they are reached differently.

ASKING - what the platform has, which CR a UI name maps to, how one shape is
assembled field by field, and where each has been got wrong before:

    asgard-cli find <terms>

That searches all four parts of the material at once - the platform wiki, the
deployment extracts, the guidance for each decision, and the skills the agent in
a customer repo loads - and hands over the counterpart of whatever it finds. Ask
in Chinese if that is the language the question was asked in; the glossary
carries the translation. **This works with no repository**, which is the point:
the question gets asked in a meeting, before there is a directory.

    asgard-cli wiki <page>     the platform
    asgard-cli usecase <name>  one deployment shape, field by field
    asgard-cli guide <name>    one decision, and how it has been got wrong
    asgard-cli brief <what>    the thing you are about to do

BUILDING - a chart of Asgard custom resources per project, each deployed to its
own namespace:

    asgard-cli project       every chart: its shape, what it declares, what it lacks
    asgard-cli question      what nobody has answered yet, and who each is with
    asgard-cli request       what the customer asked for and is not done
    asgard-cli task          the task specs that are open
    asgard-cli add <kind>    a CR skeleton, wired to what the chart declares
    asgard-cli check         the structure; "verify" is the rendered chart

Each of those four reads a file in the customer's repository back to you, and
each takes ` + "`--format json`" + `. **None of them says where the engagement is.** There
is no such command and there was: it derived one position from the earliest
missing CR kind, and an onboarding is not linear - three of the most expensive
decisions in the engagement this was built from were made, built and reversed.
What replaced it is the records themselves, and guidance reached by subject
through "find" or by name through "guide", without arriving anywhere to be
handed it.

Work arrives as a request: one thing the customer wants that the agent cannot do
today. "asgard-cli request add" opens one, and every status the engagement keeps
lives in the customer's repository, never in this tool, so the agent that opens
that repo next can read where the work stands.

With no repository yet, "asgard-cli guide init" says how one begins and
"asgard-cli init" writes the config. The repository root is the workspace - one
customer, one repository - and every project lives inside it.

Run "asgard-cli <command> --help" for details on an individual command.`,
		Version: version.Get().String(),

		// On failure print just the error, not a full page of usage; main is the
		// single place that prints it and picks the exit code.
		SilenceUsage:  true,
		SilenceErrors: true,

		// With no subcommand, show help rather than silently succeeding.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

	// --template-dir is persistent because a prompt is read by `guide`, `find`
	// and `audit-material` alike, and an override that applied to only one of
	// them would make the three disagree about what the material says.
	var templateDir string
	cmd.PersistentFlags().StringVar(&templateDir, "template-dir", "",
		"read stage prompts from this directory instead of the embedded copies, per file; for iterating on prompt text")
	cmd.PersistentPreRunE = func(c *cobra.Command, _ []string) error {
		if templateDir == "" {
			stage.SetOverrideDir("")
			return nil
		}
		info, err := os.Stat(templateDir)
		if err != nil {
			return fmt.Errorf("--template-dir %s: %w", templateDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("--template-dir %s is not a directory", templateDir)
		}
		stage.SetOverrideDir(templateDir)
		// Say it on every run. A prompt that is not the released one is the
		// first thing to suspect when guidance looks wrong, and an override
		// nobody can see is worse than no override.
		fmt.Fprintf(c.ErrOrStderr(), "reading stage prompts from %s where they exist, embedded otherwise\n", templateDir)
		return nil
	}

	cmd.AddCommand(
		newAddCmd(),
		newCheckCmd(),
		newDecisionCmd(),
		newReferenceCmd(),
		newDoctorCmd(),
		newAuditCmd(),
		newBriefCmd(),
		newReadingCmd(),
		newFindCmd(),
		newIssueCmd(),
		newSizeCmd(),
		newInitCmd(),
		newGuideCmd(),
		newProjectCmd(),
		newQuestionCmd(),
		newRenderCmd(),
		newRequestCmd(),
		newScaffoldCmd(),
		newTaskCmd(),
		newUsecaseCmd(),
		newVerifyCmd(),
		newVersionCmd(),
		newWikiCmd(),
	)

	// What this build answers to, handed to `check` so it can report a command
	// name in a customer repository that no longer exists. The list has to come
	// from the tree rather than from a constant: a constant is a second copy
	// that goes stale exactly when a command is renamed, which is the failure
	// this exists to catch.
	check.SetKnownCommands(commandNames(cmd))

	return cmd
}

// commandNames returns every name and alias in the tree, one level deep.
//
// One level is deliberate. `asgard-cli request add` names the command
// `request`, and whether `add` is one of its subcommands is a different
// question - a wrong subcommand is a typo, a wrong command is a rename nobody
// was told about.
func commandNames(root *cobra.Command) []string {
	var out []string
	for _, c := range root.Commands() {
		out = append(out, c.Name())
		out = append(out, c.Aliases...)
	}
	return out
}

// recallHere notes a page as opened, when the command was run inside an
// engagement. Both `wiki` and `usecase` work with no repository at all - that
// is deliberate, they are reference material - so this finds one if there is
// one and does nothing if there is not.
func recallHere(kind, name string) {
	path, err := config.Find(".")
	if err != nil {
		return
	}
	work.Recall(filepath.Dir(path), kind, name)
}
