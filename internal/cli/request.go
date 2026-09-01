package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "Record what the customer asked for, and move its status",
		Long: `Record what the customer asked for, and move its status.

A request is the unit of work an engagement actually receives: one thing the
customer wants that the agent cannot do today. It is what ` + "`asgard-cli next`" + `
walks - an onboarding is the first request, and everything after it arrives the
same way.

The record is a file in the customer repository, ` + "`" + work.RequestDir + `/REQ-xxx-<name>.md` + "`" + `,
registered in ` + "`" + work.RequestIndex + "`" + `. Nothing is stored in this CLI: the repo is
where the next agent looks, so the repo is where the state lives.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newRequestAddCmd(),
		newRequestTargetCmd(),
		newRequestStatusCmd("ready", work.Ready,
			"the background, the audience and the scope are settled and it can be broken into tasks"),
		newRequestStatusCmd("done", work.Done,
			"every task it spawned is done and the delta has reached the living spec"),
	)

	return cmd
}

func newRequestAddCmd() *cobra.Command {
	var (
		slug     string
		audience string
		project  string
		priority string
	)

	cmd := &cobra.Command{
		Use:   "add <what the customer asked for>",
		Short: "Open a request for something the customer asked for",
		Long: `Open a request for something the customer asked for.

Write the title in the customer's own words. The translation into a project, a
read path and an entry point is the request's own job, and keeping the original
wording is what lets the next reader check that the translation was right.

One request per thing they asked for. Two capabilities in one file is how a
half-finished request ends up marked done.

It writes the spec, registers it, and stamps today's date and ` + "`draft`" + ` on both.
The ID is the first unused REQ number in ` + "`" + work.RequestDir + "`" + `, read from the file
names rather than the index, so a spec written without its row still owns its
number.

    asgard-cli request add "warehouse staff need to ask about stock in chat"
    asgard-cli request add "let visitors ask about products" --project site

--project is optional and usually unknown at this point: it is decided by who is
on the other end, which is section 2 of the spec. Until it is set,
` + "`asgard-cli next`" + ` treats the request as the interview it is.`,
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

			if slug == "" {
				slug = work.Slugify(title)
			}
			if slug == "" {
				return fmt.Errorf("cannot derive a file name from %q; pass --slug <short-name>", title)
			}

			request := work.Request{
				Title:    title,
				Slug:     slug,
				Status:   work.Draft,
				Priority: priority,
				Audience: audience,
				Project:  project,
				Raised:   today(),
				SpecSlug: cfg.Workspace.Slug + "-asgard",
			}

			request, err = work.AddRequest(root, request)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Wrote %s\n", request.File())
			fmt.Fprintf(out, "  status   %s\n  raised   %s\n", work.Draft, request.Raised)
			if project != "" {
				fmt.Fprintf(out, "  project  %s\n", project)
			}
			fmt.Fprintf(out, "Registered in %s\n", work.RequestIndex)

			fmt.Fprintf(out, `
Fill in sections 2 and 3 of that file - who is on the other end, and which
systems it has to read. Those two answers decide the project, the read path and
the entry point, and nothing downstream can be designed without them.

    asgard-cli next     walks this request from here
`)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "short name for the file (defaults to one derived from the title; required when the title has no ASCII)")
	cmd.Flags().StringVar(&audience, "audience", "", "who is on the other end, if it is already known (defaults to a TODO in the spec)")
	cmd.Flags().StringVar(&project, "project", "", "project this lands in, if the audience is already known (defaults to a TODO, decided later)")
	cmd.Flags().StringVar(&priority, "priority", "", "how urgent it is, in whatever words the engagement uses (defaults to a TODO)")

	return cmd
}

// newRequestStatusCmd builds one status transition. They are separate commands
// rather than one taking a status argument because each transition has its own
// gate, and the gate belongs in the help of the command that crosses it.
func newRequestStatusCmd(verb string, to work.Status, gate string) *cobra.Command {
	return &cobra.Command{
		Use:   verb + " <request-id>",
		Short: "Move a request to " + string(to),
		Long: fmt.Sprintf(`Move a request to %s, which means %s.

It rewrites the status in %s, in the request spec's own Meta section, and appends
a dated line to the spec's log. Those are three places, and they are done
together here because a repo where two of them disagree gives the next reader no
way to tell which one is current.

    asgard-cli request %s REQ-001`, to, gate, work.RequestIndex, verb),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			root, _, err := loadRepo()
			if err != nil {
				return err
			}
			if err := work.SetRequestStatus(root, id, to, today()); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s is now %s, as of %s\n", id, to, today())
			return nil
		},
	}
}

func newRequestTargetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target <request-id> <project>",
		Short: "Record which project a request lands in",
		Long: `Record which project a request lands in.

This is the answer the interview produces, and it is what moves the onboarding
on: until a request names a project this repository has, ` + "`asgard-cli next`" + ` reads
it as an interview that has not finished.

The project follows the audience, not the data. Same audience as an existing
project means it goes in that project; a new audience means a new project, which
walks its own way through the remaining stages. Putting a public capability into
an internal project because the data happens to be nearby is how a semantic layer
becomes reachable from a public endpoint.

It writes the Spec column of ` + "`" + work.RequestIndex + "`" + `, the Target project line of the
request's Meta, and a dated line in its log.

    asgard-cli request target REQ-001 erp`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, project := args[0], args[1]

			root, cfg, err := loadRepo()
			if err != nil {
				return err
			}
			if _, ok := cfg.Project(project); !ok {
				return fmt.Errorf("no project %q in %s; add it with `asgard-cli project add %s`",
					project, config.FileName, project)
			}
			if err := work.SetRequestProject(root, id, project, today()); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s targets project %s, as of %s\n", id, project, today())
			return nil
		},
	}

	return cmd
}
