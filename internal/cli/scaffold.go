package cli

import (
	"fmt"
	"os"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
)

func newScaffoldCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Write the repository skeleton next to " + pipelineconfig.FileName,
		Long: `Write the repository skeleton next to ` + pipelineconfig.FileName + `.

This is the part of a customer repo that is the same for every engagement: the
four-layer docs model, the SDD rules, the six design-time skills that hold for
any Asgard, and an AGENTS.md carrying the platform contract. What it does not write
is the customer's own knowledge - which systems exist, how the projects split,
what the CRs look like. That is what the onboarding produces.

**It also does not write anything that describes a particular Asgard server.**
The CR shapes, the processor catalogue and the verification skill come from
` + "`asgard-cli skill update`" + `, which asks the platform. A customer's server can be
several versions from this binary, and a file saying what a field is called is
only true of one of them.

Running it again is safe: existing files are left alone and reported as skipped,
so it can be re-run after adding a project or when a file was deleted by hand.
--force overwrites, which discards local edits to the skeleton.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := scaffoldRoot()
			if err != nil {
				return err
			}
			ws, err := workspaceForTemplates(cmd)
			if err != nil {
				return err
			}
			return runScaffold(cmd, root, ws, force, true)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite files that already exist")

	return cmd
}

// scaffoldRoot is where the skeleton goes.
//
// **The skeleton is written where you are.** Every other command finds the
// repository by its declaration, and this is the command that writes one - so
// it cannot require one to already exist.
func scaffoldRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}
	if found := repo.Root(dir); found != "" {
		return found, nil
	}
	return dir, nil
}

// runScaffold writes the skeleton and reports it. Split out of the command so
// that `init`, which is the composition somebody actually runs first, is one
// call to each of the three things it composes rather than a copy of them.
//
// standalone says whether this is the whole of what was run. `init` does the
// next two steps itself, and printing "then run skill update" in the middle of
// a command that is about to run it reads as an instruction rather than as a
// report.
func runScaffold(cmd *cobra.Command, root string, ws repo.Workspace, force, standalone bool) error {
	projects, err := repo.Projects(root)
	if err != nil {
		return err
	}

	results, err := scaffold.Write(root, ws, projects, force)
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

	if standalone && (created > 0 || overwritten > 0) {
		fmt.Fprintf(out, `
Verify the skeleton before writing any CRs:

  asgard-cli check

Then fetch the material that describes the server this repository deploys to -
the CR shapes, the processor catalogue and the verification skill. It is not in
this binary, because what a field is called is a fact about a server and not
about a CLI release:

  asgard-cli skill update

Then "asgard-cli project" for what each chart declares, and
AGENTS.md for how to change the repo.
`)
	}
	return nil
}
