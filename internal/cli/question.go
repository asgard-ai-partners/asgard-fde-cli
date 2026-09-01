package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newQuestionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "question",
		Short: "Record what nobody has answered yet, and what the answer turned out to be",
		Long: `Record what nobody has answered yet, and what the answer turned out to be.

An unanswered question has nowhere else to live. A decision record is for
something settled. A task spec's open questions vanish when that task reaches
done. The living spec describes what is, not what nobody knows. So without
` + "`" + work.QuestionFile + "`" + ` the question is rediscovered by the next person, usually by
making the wrong assumption first.

` + "`asgard-cli next`" + ` prints everything still open, on every run, before anything
else it has to say.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newQuestionAddCmd(), newQuestionAnsweredCmd())

	return cmd
}

func newQuestionAddCmd() *cobra.Command {
	var (
		blocks string
		ask    string
	)

	cmd := &cobra.Command{
		Use:   "add <question>",
		Short: "Add an open question",
		Long: `Add an open question, numbered after the highest one already there and stamped
with today's date.

Add it the moment it blocks or shapes a decision, not later. A question you can
already answer is not one - answer it instead.

    asgard-cli question add "which of the two stock figures is authoritative" \
      --blocks REQ-001 --ask "the warehouse lead"

--ask matters more than it looks: a question with no owner is a wish. --blocks is
what makes it findable from the work it is holding up.

This is not the place to park a question to avoid asking it. If the customer can
answer it in the next meeting, it belongs in that agenda, not in a table.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, _, err := loadRepo()
			if err != nil {
				return err
			}

			question, err := work.AddQuestion(root, work.Question{
				Text:   args[0],
				Blocks: blocks,
				Owner:  ask,
				Raised: today(),
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Added question %s to %s, raised %s\n", question.Number, work.QuestionFile, question.Raised)
			if ask == "" {
				fmt.Fprintf(out, "\nNo owner recorded. A question with nobody to ask does not get answered;\n"+
					"add one by editing the row, or with --ask next time.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&blocks, "blocks", "", "what this holds up - a REQ or TASK id, or a decision in words (defaults to empty)")
	cmd.Flags().StringVar(&ask, "ask", "", "who can answer it (defaults to empty, and a question with no owner is a wish)")

	return cmd
}

func newQuestionAnsweredCmd() *cobra.Command {
	var decision string

	cmd := &cobra.Command{
		Use:   "answered <number> <answer>",
		Short: "Move a question to the answered table",
		Long: `Move a question to the answered table, with the answer and today's date.

The row moves rather than being deleted. That a question was once open is what
explains the shape of the design that answered it, and deleting the row leaves
the design looking arbitrary.

    asgard-cli question answered 3 "location 608 only, the row's own is stale" \
      --decision 2026-09-04-safety-stock-source.md

If the answer settled anything, write the dated record under docs/decisions/ and
name it with --decision. The answer here is one line; the record is where the
reasoning goes.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			number, answer := args[0], args[1]

			root, _, err := loadRepo()
			if err != nil {
				return err
			}
			if err := work.AnswerQuestion(root, number, answer, decision, today()); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Question %s answered, as of %s\n", number, today())
			if decision == "" {
				fmt.Fprintf(out, "\nNo decision record named. If this settled how something is built, it needs\n"+
					"one under docs/decisions/ - the answer above is a line, not the reasoning.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&decision, "decision", "", "file name of the dated record under docs/decisions/ (defaults to empty)")

	return cmd
}
