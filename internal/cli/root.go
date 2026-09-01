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
		Long: `asgard-cli is the command line tool for Asgard FDE.

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
		newInitCmd(),
		newLintCmd(),
		newVersionCmd(),
	)

	return cmd
}
