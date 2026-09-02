// Package cli assembles the asgard-cli command tree.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
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

	cmd.AddCommand(
		newAddCmd(),
		newCheckCmd(),
		newDecisionCmd(),
		newDoctorCmd(),
		newFindCmd(),
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
