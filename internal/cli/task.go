package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Write and track the executable task specs",
		Long: `Write and track the executable task specs.

A request says what the customer wants; a task spec says how one chart change is
made and how it is verified. Work that touches a ` + "`SemanticLayer`" + ` or a
` + "`DataConnector`" + `, widens which cubes an agent may query, or introduces a write
path needs one before it is implemented - those three point at the customer's
live systems, so they are reviewed before the change, not after.

The record is a file in the customer repository, ` + "`" + work.TaskDir + `/TASK-xxx-<name>.md` + "`" + `,
registered in ` + "`" + work.TaskIndex + "`" + `. Task IDs are global across projects: two
branches numbering from their own project is how a collision happens.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newTaskAddCmd(),
		newTaskStatusCmd("ready", work.Ready,
			"scope and acceptance criteria are complete and every R# maps to an implementation task and a verification entry"),
		newTaskStatusCmd("start", work.InProgress,
			"implementation has begun, which the SDD rules say waits for the user to ask for it"),
		newTaskStatusCmd("done", work.Done,
			"the gate is green, the behaviour delta has reached the living spec, and any decision it settled has its own dated record"),
	)

	return cmd
}

func newTaskAddCmd() *cobra.Command {
	var (
		slug       string
		project    string
		request    string
		complexity string
		owner      string
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Write a task spec and register it",
		Long: `Write a task spec and register it.

The spec is the single-file SDD shape the generated repo documents in
` + "`docs/spec-driven-development.md`" + `: Meta, 1) Requirements, 2) Design,
3) Implementation Tasks, 4) Execution Log. Every part somebody has to answer is
marked TODO, and nothing is filled in on their behalf - a field that looks
decided but never was is worse than an empty one, because the next reader cannot
tell the difference.

It stamps today's date and ` + "`draft`" + `, registers the row, and points the index's
Next Task section at whatever is now most advanced.

    asgard-cli task add "expose stock levels to the warehouse agent" \
      --request REQ-001 --project erp --complexity M`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]

			root, cfg, err := loadRepo()
			if err != nil {
				return err
			}

			if project != "" {
				if _, ok := cfg.Project(project); !ok {
					return fmt.Errorf("no project %q in %s; add it with `asgard-cli project add %s`",
						project, root, project)
				}
			}
			if request != "" {
				requests, err := work.ReadRequests(root)
				if err != nil {
					return err
				}
				known := false
				for _, r := range requests {
					if r.ID == request {
						known = true
						break
					}
				}
				if !known {
					return fmt.Errorf("no %s in %s; open it with `asgard-cli request add`", request, work.RequestIndex)
				}
			}

			if slug == "" {
				slug = work.Slugify(title)
			}

			task := work.Task{
				Title:      title,
				Slug:       slug,
				Status:     work.Draft,
				Owner:      owner,
				Complexity: complexity,
				Project:    project,
				Request:    request,
				Created:    today(),
				SpecSlug:   cfg.Workspace.Slug + "-asgard",
			}

			task, err = work.AddTask(root, task)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Wrote %s\n", task.File())
			fmt.Fprintf(out, "  status   %s\n  created  %s\n", work.Draft, task.Created)
			if project != "" {
				fmt.Fprintf(out, "  project  %s\n", project)
			}
			if request != "" {
				fmt.Fprintf(out, "  request  %s\n", request)
			}
			fmt.Fprintf(out, "Registered in %s\n", work.TaskIndex)

			fmt.Fprintf(out, `
Read `+"`docs/spec/%s/`"+` for the module this touches before filling it in: the
spec is a delta against what that module says the system does today.

    asgard-cli task ready %s     once scope and acceptance criteria are complete
`, task.SpecSlug, task.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "short name for the file (defaults to one derived from the title; required when the title has no ASCII)")
	cmd.Flags().StringVar(&project, "project", "", "project whose chart this changes (defaults to a TODO in the spec)")
	cmd.Flags().StringVar(&request, "request", "", "request this implements, e.g. REQ-001 (defaults to none, for work no customer asked for)")
	cmd.Flags().StringVar(&complexity, "complexity", "", "S, M or L (defaults to a TODO in the spec)")
	cmd.Flags().StringVar(&owner, "owner", "", "who is doing it (defaults to a TODO in the spec)")

	return cmd
}

// newTaskStatusCmd builds one status transition. They are separate commands
// rather than one taking a status argument because each has its own gate, and
// the gate belongs in the help of the command that crosses it.
func newTaskStatusCmd(verb string, to work.Status, gate string) *cobra.Command {
	return &cobra.Command{
		Use:   verb + " <task-id>",
		Short: "Move a task to " + string(to),
		Long: fmt.Sprintf(`Move a task to %s, which means %s.

It rewrites the status in %s, in the task spec's own Meta section, appends a
dated line to the spec's Execution Log, and refreshes the index's Next Task
section. That is four places, and doing them by hand is how a repo ends up
saying two different things about the same task with no way to tell which is
current.

    asgard-cli task %s TASK-001`, to, gate, work.TaskIndex, verb),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			root, _, err := loadRepo()
			if err != nil {
				return err
			}
			if err := work.SetTaskStatus(root, id, to, today()); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s is now %s, as of %s\n", id, to, today())
			if to == work.Done {
				fmt.Fprintf(out, `
Two obligations outlive the status, and a task that skipped them is not really
done: apply the behaviour delta to the living spec, and give any decision it
settled its own dated record under docs/decisions/.
`)
			}
			return nil
		},
	}
}
