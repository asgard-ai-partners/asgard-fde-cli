// Package cli assembles the asgard-cli command tree.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

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
		Long: `asgard-cli drives the onboarding of a customer onto the Asgard platform.

START WITH:

    asgard-cli next

It reports which stage the onboarding is at and what that stage requires,
working it out from the repository itself rather than from anything remembered.
Run it again after each step instead of guessing which command comes next - it
answers even in an empty directory, where the answer is how to begin.

Work arrives as a request: one thing the customer wants that the agent cannot do
today. "asgard-cli request add" opens one, and every status the engagement keeps
lives in the customer's repository, never in this tool, so the agent that opens
that repo next can read where the work stands.

An onboarding produces one repository per customer: a Helm chart of Asgard
custom resources per project, each deployed to its own namespace. The repository
root is the workspace, and every project lives inside it.

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

	// --template-dir is persistent because a prompt is read by `next`, `find`
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
		newNextCmd(),
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

	return cmd
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
