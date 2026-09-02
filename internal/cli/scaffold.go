package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
)

func newScaffoldCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Write the repository skeleton next to " + config.FileName,
		Long: `Write the repository skeleton next to ` + config.FileName + `.

This is the part of a customer repo that is the same for every engagement: the
four-layer docs model, the SDD rules, the four acceptance gate scripts, the three
design-time skills, the CD workflow, and an AGENTS.md carrying the platform
contract. What it does not write is the customer's own knowledge - which systems
exist, how the projects split, what the CRs look like. That is what the
onboarding produces.

Running it again is safe: existing files are left alone and reported as skipped,
so it can be re-run after adding a project or when a file was deleted by hand.
--force overwrites, which discards local edits to the skeleton.`,
		Args: cobra.NoArgs,
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
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			// The skeleton belongs next to the config, not in whatever
			// subdirectory the command happened to be run from.
			root := filepath.Dir(path)

			results, err := scaffold.Write(root, cfg, force)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			var created, overwritten, skipped int
			var preserved, stale []string
			for _, r := range results {
				switch r.Status {
				case scaffold.Created:
					created++
					fmt.Fprintf(out, "  created      %s\n", r.Path)
				case scaffold.Overwritten:
					overwritten++
					fmt.Fprintf(out, "  overwritten  %s\n", r.Path)
				case scaffold.Preserved:
					preserved = append(preserved, r.Path)
					fmt.Fprintf(out, "  preserved    %s\n", r.Path)
				case scaffold.Stale:
					stale = append(stale, r.Path)
					fmt.Fprintf(out, "  stale        %s\n", r.Path)
				default:
					skipped++
				}
			}

			fmt.Fprintf(out, "\n%d created", created)
			if overwritten > 0 {
				fmt.Fprintf(out, ", %d overwritten", overwritten)
			}
			if skipped > 0 {
				fmt.Fprintf(out, ", %d already present", skipped)
			}
			if len(stale) > 0 {
				fmt.Fprintf(out, ", %d stale", len(stale))
			}
			fmt.Fprintf(out, " in %s\n", root)

			// "already present" reads as "up to date", and that reading has
			// been acted on: an agent re-ran scaffold after an upgrade, saw
			// nothing to do, told the user the repo was current, and went on to
			// work from a skill three versions old. These files are the ones an
			// engagement never edits, so a difference in them is this CLI having
			// moved, not the engagement having written something.
			if len(stale) > 0 {
				fmt.Fprintf(out, "\n%d file(s) are shipped material this CLI has since changed. Yours are\n"+
					"older, and were left alone:\n\n", len(stale))
				for _, p := range stale {
					fmt.Fprintf(out, "  %s\n", p)
				}
				fmt.Fprintf(out, "\nTake the newer ones with `asgard-cli scaffold --force`. Nothing an\n"+
					"`asgard-cli` command writes into is touched by that - indexes, the\n"+
					"open-questions table and the living spec are preserved either way.\n")
			}

			if len(preserved) > 0 {
				fmt.Fprintf(out, "\n%d file(s) preserved despite --force, because `asgard-cli`\n"+
					"writes into them and they no longer match the template they started as:\n\n", len(preserved))
				for _, p := range preserved {
					fmt.Fprintf(out, "  %s\n", p)
				}
				fmt.Fprintf(out, "\n--force discards local edits to the skeleton, and each of these stopped\n"+
					"being skeleton the first time an `asgard-cli` command wrote to it. To reset\n"+
					"one deliberately, delete it and run scaffold again.\n")
			}

			if created > 0 || overwritten > 0 {
				fmt.Fprintf(out, `
Verify the skeleton before writing any CRs:

  asgard-cli check

Then "asgard-cli next" for what this stage requires, and AGENTS.md for how to
change the repo.
`)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite files that already exist")

	return cmd
}
