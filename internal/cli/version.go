package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
)

func newVersionCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and build information",
		Long: `Print version and build information.

The version, commit and build date come from the ldflags GoReleaser sets when it
builds a release. A binary from a plain go build or go install reports "dev" and
falls back to the module and VCS metadata the Go toolchain embeds.

Use --json when a script needs to read the values.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()
			out := cmd.OutOrStdout()

			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			fmt.Fprintf(out, "asgard-cli %s\n", info.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")

	return cmd
}
