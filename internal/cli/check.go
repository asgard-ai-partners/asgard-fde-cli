package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [project ...]",
		Short: "Verify the repository's structural invariants",
		Long: `Verify the repository's structural invariants - the ones a chart render cannot
see, and that otherwise surface at deploy time or when the next person picks the
repo up:

  - the root README's project table matches the directories under projects/
  - every project has a deploy.yaml, its envs are dev or prod, and the values
    files it names exist, along with the shared common/values-<env>.yaml
  - runtime skills under common/skills/ carry name and description frontmatter,
    with the name matching the directory
  - the SDD entry points under requirements/ are present
  - the docs/ spec layer is intact: required files, the living spec's module
    index matching the files on disk, dated filenames, and every relative link
    inside docs/ resolving

Naming projects limits the project-scoped checks to those; the repo-wide checks
always run. Exits non-zero when anything fails.

This is the first step of the acceptance gate. The remaining steps work on a
rendered chart and still need helm: see "asgard-cli next --stage verify".`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}
			path, err := config.Find(dir)
			if err != nil {
				if errors.Is(err, config.ErrNotFound) {
					return fmt.Errorf("no %s found; run `asgard-cli init` first", config.FileName)
				}
				return err
			}
			root := filepath.Dir(path)

			report, err := check.Run(root, args...)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, f := range report.Warnings() {
				fmt.Fprintf(out, "warn   %s\n", f.Message)
			}
			for _, f := range report.Errors() {
				fmt.Fprintf(out, "error  %s\n", f.Message)
			}

			scope := "whole repo"
			if len(report.Scope) > 0 {
				scope = fmt.Sprintf("%d project(s): %v", len(report.Scope), report.Scope)
			}

			if !report.OK() {
				return fmt.Errorf("%d problem(s) in %s", len(report.Errors()), scope)
			}
			fmt.Fprintf(out, "ok  structure is consistent (%s)\n", scope)
			return nil
		},
	}

	return cmd
}
