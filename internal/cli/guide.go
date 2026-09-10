package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
)

// newGuideCmd reads one piece of decision guidance.
//
// This was `next --stage <name>`: reading a document, behind a flag, on a
// command about repository state. The two shared a name because the guidance
// used to be steps of a walk that `next` reported a position in. It is not, and
// the corpus's other parts are read by `wiki <page>` and `usecase <name>`, so
// this is the third of three.
func newGuideCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guide [name]",
		Short: "Read the guidance for one decision",
		Long: `Read the guidance for one decision. With no argument, list all of it.

Each of these is a decision somebody has to make, with the answer that has been
got wrong before printed next to the right one. **None of them is a step you
arrive at.** Read whichever the question in front of you reaches, in any order,
as many times as it is useful.

    asgard-cli guide                 all of it
    asgard-cli guide read-path       one
    grep -ril "<term>" .agents/skills/asgard-platform/
                                     reach any of it by subject, alongside the
                                     wiki, the extracts and the skills

Nothing raises one of these for you. There was a command that did - it read the
repository, matched its shape against a fixed set of conditions and named the
guidance each raised - and it was the walk with the numbers taken off: the
conditions were the old rungs, in the old order. What replaced it is this list,
and by grepping the material.

Guidance that names actual projects and requests is rendered against the
repository when there is one. Read outside a repository it still reads, with
those parts empty.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if len(args) == 0 {
				// Printed where it actually happens. It used to follow
				// `scaffold`, which was the step before it; with the two
				// procedural pages gone it is the first thing, because the
				// interview is what everything below is decided from.
				fmt.Fprintf(out, "  %-14s %s\n",
					stage.RequirementsStage.Name, stage.RequirementsStage.Title)
				for _, s := range stage.Stages {
					fmt.Fprintf(out, "  %-14s %s\n", s.Name, s.Title)
				}
				fmt.Fprintf(out, "  %-14s %s\n", stage.IdleStage.Name, stage.IdleStage.Title)
				fmt.Fprintf(out, "\nRead one with `asgard-cli guide <name>`.\n")
				return nil
			}

			found, ok := stage.Find(args[0])
			if !ok {
				return fmt.Errorf("no guidance named %q; list it with `asgard-cli guide`", args[0])
			}

			// Reference material reads with no repository - that is the point
			// of it - so being outside one renders against an empty state
			// rather than failing.
			state := stage.State{}
			if root := repo.Root("."); root != "" {
				if s, err := stage.Inspect(root); err == nil {
					state = s
				}
				// The kind is the command somebody types, not the package the
				// pages come from. This logged "stage" for years after `next
				// --stage <name>` became `guide <name>`, so grepping the log
				// for what a predecessor read, by the name they would have
				// typed, missed every one of them.
			}

			prompt, err := found.Prompt(state)
			if err != nil {
				return err
			}
			fmt.Fprint(out, prompt)
			return nil
		},
	}

	return cmd
}
