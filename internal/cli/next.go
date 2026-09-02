package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newNextCmd() *cobra.Command {
	var (
		stageName string
		list      bool
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "Report which stage the onboarding is at, and what to do",
		Long: `Report which stage the onboarding is at, and what to do.

The stage is worked out from the repository itself - which files exist, which CR
kinds each chart declares - not from a counter in the config. So it stays right
when someone does a step by hand, and it never claims work is done that is not.

Stages 4, 5 and 6 are the decisions that have been answered wrong before. Those
stages print the wrong answer as well as the right one, because the wrong one is
what looks obvious.

When nothing is in flight - no request, no task, every chart complete - it does
not print a stage at all. It reports where the repo stands and asks what the
customer wants next, because that is the only thing that can move the work on.

Use --stage to read any stage out of order, and --list to see them all.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if list {
				// Neither requirements nor idle is a step of the walk, so
				// neither carries a number - but requirements is printed where
				// it actually happens, after the skeleton exists and before the
				// split is decided. Listed at the end it read as though it came
				// after deploy.
				for _, s := range stage.Stages {
					fmt.Fprintf(out, "  %d  %-14s %s\n", s.Number, s.Name, s.Title)
					if s.Name == stage.Scaffold {
						fmt.Fprintf(out, "  -  %-14s %s\n",
							stage.RequirementsStage.Name, stage.RequirementsStage.Title)
					}
				}
				fmt.Fprintf(out, "  -  %-14s %s\n", stage.IdleStage.Name, stage.IdleStage.Title)
				fmt.Fprintf(out, "\nRead one with `asgard-cli next --stage <name>`.\n")
				return nil
			}

			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}
			path, err := config.Find(dir)
			if err != nil {
				if errors.Is(err, config.ErrNotFound) {
					// Not an error: "there is no onboarding here yet" is a
					// stage, and it is the one this command exists to explain.
					// --stage still wins, because reading a stage out of order
					// is exactly what someone does before starting one.
					want := stage.Init
					if stageName != "" {
						want = stage.Name(stageName)
					}
					s, ok := stage.Find(string(want))
					if !ok {
						return unknownStage(stageName)
					}
					prompt, err := s.Prompt(&config.Config{}, stage.State{})
					if err != nil {
						return err
					}
					fmt.Fprintf(out, "%s\n\n%s", s, prompt)
					return nil
				}
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			root := filepath.Dir(path)

			state, err := stage.Inspect(root, cfg)
			if err != nil {
				return err
			}

			current := stage.Current(state)
			if stageName != "" {
				found, ok := stage.Find(stageName)
				if !ok {
					return unknownStage(stageName)
				}
				current = found
			}

			prompt, err := current.Prompt(cfg, state)
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "%s\n\n%s", current, prompt)
			printQuestions(out, state)
			printWork(out, current, state)
			return nil
		},
	}

	cmd.Flags().StringVar(&stageName, "stage", "", "read a specific stage instead of the current one")
	cmd.Flags().BoolVar(&list, "list", false, "list every stage")

	return cmd
}

// unknownStage is one message for both places that resolve --stage: without a
// config the stage is all the command has to go on, so a typo there has to be
// reported rather than quietly falling back to stage 0.
func unknownStage(name string) error {
	return fmt.Errorf("unknown stage %q; list them with `asgard-cli next --list`", name)
}

// printWork reports what the repository says is open, after the stage guidance.
// The stage is derived from files; a status is something somebody wrote down.
// Showing both is how a person picking the repo up sees where the work stands,
// and neither one can be inferred from the other.
func printWork(out io.Writer, current stage.Stage, state stage.State) {
	requests := work.ActiveRequests(state.Requests)
	tasks := work.ActiveTasks(state.Tasks)

	if len(requests) > 0 {
		fmt.Fprintf(out, "\nOpen requests, from %s:\n", work.RequestIndex)
		for _, r := range requests {
			target := r.Project
			if target == "" {
				target = "no project yet"
			}
			fmt.Fprintf(out, "  %-9s %-10s %-16s %s\n", r.ID, r.Status, target, r.Title)
		}
	}

	switch {
	case len(tasks) > 0:
		fmt.Fprintf(out, "\nOpen task specs, from %s:\n", work.TaskIndex)
		for _, t := range tasks {
			fmt.Fprintf(out, "  %-9s %-11s %s\n", t.ID, t.Status, t.Title)
		}
		fmt.Fprintf(out, "\nFinish or park these before opening another. Move a status with\n"+
			"`asgard-cli task ready|start|done <id>`, which changes the index, the\n"+
			"spec's Meta and the spec's log together - by hand it is three places.\n")

	case current.NeedsTask() && len(requests) > 0:
		fmt.Fprintf(out, "\nNo task spec is open. Work at this stage touches a chart, so it wants one\n"+
			"first:\n\n    asgard-cli task add \"<title>\" --request %s --project <project>\n",
			requests[0].ID)

	case current.NeedsTask():
		fmt.Fprintf(out, "\nNo task spec is open. Work at this stage touches a chart, so it wants one\n"+
			"first: `asgard-cli task add \"<title>\" --project <project>`.\n")
	}
}

// printQuestions reports what is still unanswered. A returning agent needs this
// before anything else: the fastest way to do damage is to design past a
// question someone already knew was open.
func printQuestions(out io.Writer, state stage.State) {
	if len(state.Questions) == 0 {
		return
	}

	fmt.Fprintf(out, "\n%d open question(s), from %s:\n", len(state.Questions), work.QuestionFile)
	for _, q := range state.Questions {
		fmt.Fprintf(out, "  %s. %s\n", q.Number, q.Text)
		if q.Blocks != "" {
			fmt.Fprintf(out, "       blocks: %s\n", q.Blocks)
		}
		if q.Owner != "" {
			fmt.Fprintf(out, "       ask:    %s\n", q.Owner)
		}
		if q.Raised != "" && q.Raised != "-" {
			fmt.Fprintf(out, "       raised: %s\n", q.Raised)
		}
	}
	fmt.Fprintf(out, "\nDo not design past one of these. Either get the answer, or record the\n"+
		"assumption you are proceeding on and which branch it commits you to.\n")
}
