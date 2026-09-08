package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
)

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

// runScaffold writes the skeleton and reports it.
//
// There used to be a `scaffold` command that was this and nothing else, beside
// an `init` that was this plus a binding plus a skills fetch. Once init stopped
// needing a session the two did the same thing, and two commands doing the same
// thing is a question - "which of these do I run?" - that gets asked, and was,
// twice. This is what is left: one command, one internal writer.
func runScaffold(cmd *cobra.Command, root string, force bool) error {
	projects, err := repo.Projects(root)
	if err != nil {
		return err
	}

	results, err := scaffold.Write(root, projects, force)
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
		fmt.Fprintf(out, "\nTake the newer ones with `asgard-cli init --force`. Nothing an\n"+
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

	return nil
}
