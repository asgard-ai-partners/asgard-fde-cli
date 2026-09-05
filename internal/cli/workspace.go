package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/binding"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Choose which workspace the platform commands act in",
		Long: `Choose which workspace the platform commands act in.

A workspace is the customer. Everything under "pipeline" is scoped to one, and
the choice is recorded so that it is made once rather than typed on every
command.

    asgard-cli workspace list          what this account can reach
    asgard-cli workspace use <id>      record it for this checkout
    asgard-cli workspace show          which one applies here, and why

It is recorded in .asgard-cli.yaml, beside the declaration it belongs to, and
that file is committed: whoever clones the repository, and whatever agent works
in it, then needs no --workspace. The platform never reads it - a run reads the
declaration and the chart, and nothing else - so nothing there can make a
deployment succeed or fail.

**Changing it clears the pipeline recorded beside it**, because a pipeline
belongs to one workspace and means nothing in another. ` + "`asgard-cli pipeline use`" + `
is what fills the line back in, and until it does, every pipeline command
refuses to run and says so. That is deliberate: refusing is the safe half-state,
and the alternative - carrying on against whichever pipeline of the new
workspace looked closest - is how something gets deployed that nobody chose.

Which workspace and which pipeline are the two things neither the repository nor
the platform can answer alone; the releases, their keys and their triggers are
all in the declaration. Neither is guessed, and neither is guessed from a list
of one.

Overriding is --workspace or ASGARD_WORKSPACE, both of which outrank the file.
That direction is the safe one: acting on a test workspace when the customer's
was meant costs a confusing error, and the reverse deploys to a customer.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newWorkspaceListCmd(), newWorkspaceUseCmd(), newWorkspaceShowCmd())
	return cmd
}

func newWorkspaceListCmd() *cobra.Command {
	var (
		profile string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the workspaces this account can reach",
		Long: `List the workspaces this account can reach on the platform.

This is the one platform call that needs no workspace, which makes it both the
way to find an id and the way to check that a session works for the API rather
than only for the sign-in service.

It marks the one that currently applies, so a wrong binding is visible here
without running anything that would act on it. Nothing is marked when nothing
has been chosen, including when this account can reach exactly one: a list of
one is still a list.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := checkFormat(format); err != nil {
				return err
			}
			pc, err := resolveContext(cmd, contextOptions{Profile: profile})
			if err != nil {
				return err
			}
			workspaces, err := pc.Client.ListWorkspaces(cmd.Context())
			if err != nil {
				return err
			}

			// Which one applies is a second question, and one that must not
			// fail this command: not having chosen yet is exactly when somebody
			// runs this.
			current := ""
			if id, _, err := resolveWorkspace(cmd.Context(), pc.Session, pc, ""); err == nil {
				current = id
			}

			out := cmd.OutOrStdout()
			if format == formatJSON {
				rows := make([]map[string]any, 0, len(workspaces))
				for _, w := range workspaces {
					rows = append(rows, map[string]any{"id": w.ID, "name": w.Name, "current": w.ID == current})
				}
				return writeJSON(out, map[string]any{"profile": pc.Session.Profile.Name, "workspaces": rows})
			}

			if len(workspaces) == 0 {
				fmt.Fprintf(out, "No workspaces on the %s platform for this account.\n", pc.Session.Profile.Name)
				return nil
			}
			for _, w := range workspaces {
				mark := "  "
				if w.ID == current {
					mark = "* "
				}
				fmt.Fprintf(out, "%s%-22s %s\n", mark, w.ID, w.Name)
			}
			if current == "" {
				fmt.Fprintf(out, "\nNone chosen yet. `asgard-cli workspace use <id>` records one for this checkout.\n")
			}
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	return cmd
}

// `workspace use` reaches no platform, so it carries no --profile.
//
// It writes a file, and the file names a workspace id. Which platform that id
// lives on is settled by whatever later command acts on it, and a --profile
// here would have looked like it decided something.
func newWorkspaceUseCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "use <workspace-id>",
		Short: "Record which workspace this checkout's commands act in",
		Long: `Record which workspace this checkout's commands act in.

It writes .asgard-cli.yaml beside the nearest .asgard-pipeline.yaml, and that
file is meant to be committed. Beside the declaration rather than at the
repository root because a repository may carry several: a monorepo with one
declaration per team has one pipeline per team, and a single file at the root
could name only one of them.

    asgard-cli workspace use 1862431170889781248

**It writes a file in the repository and nothing outside it.** A machine-wide
default used to be recordable with --default, and it is gone: it was invisible
on the machine that had it and absent on every other, so the same command in the
same checkout did different things for two people. Outside a checkout, name the
workspace with --workspace or export ASGARD_WORKSPACE - both are visible where
they are set.

**Moving to a different workspace clears the pipeline line.** A pipeline belongs
to one workspace, so the id recorded beside it is not a pipeline of the new one -
it is either absent there, or, worse, an id that happens to exist and points at
somebody else's repository. Clearing it leaves the checkout in a state where
every pipeline command refuses and names the remedy, which is the failure being
brought forward to the next command instead of waiting for a destructive one.
` + "`asgard-cli pipeline use <id>`" + ` fills it in. **Commit both lines together**: a
commit carrying a new workspace and the old pipeline is a state that never
existed on anybody's disk.

Recording the same workspace again changes nothing and clears nothing.

The id is not checked against the platform here. "workspace list" is where one
comes from, and a check would put a network call in the middle of recording a
choice - the first command that acts on it reports a bad one anyway.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID := args[0]
			out := cmd.OutOrStdout()

			// The binding belongs beside the declaration it is for, so a
			// repository with two declarations gets two bindings rather than
			// one that can only name half of it.
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			declPath, bindPath, err := binding.Locate(dir)
			if err != nil {
				return err
			}
			if declPath == "" {
				return errNoDeclarationToBind
			}

			f := &binding.File{Version: 1}
			if existing, err := binding.Load(bindPath); err == nil {
				f = existing
			}

			// A pipeline belongs to one workspace, so a workspace that changes
			// invalidates the pipeline recorded beside it. Dropping the line is
			// the honest outcome: the alternative is a file that names two
			// objects which have nothing to do with each other, and every
			// command downstream believing it.
			previous, cleared := f.Workspace, ""
			if previous != workspaceID && f.Pipeline != "" {
				cleared, f.Pipeline = f.Pipeline, ""
			}
			f.Workspace = workspaceID
			if err := binding.Save(bindPath, f); err != nil {
				return err
			}

			rel := bindPath
			if root, _ := locateRepo(cmd.Context()); root != "" {
				if r, relErr := filepath.Rel(root, bindPath); relErr == nil {
					rel = r
				}
			}
			fmt.Fprintf(out, "Recorded workspace %s in %s.\n", workspaceID, rel)

			if cleared != "" {
				// Said in full, every time. The next command will refuse, and
				// an agent that is told why here does not have to work it out
				// from a failure two steps later.
				fmt.Fprintf(out, "\nCleared the pipeline line, which recorded %s.\n", cleared)
				fmt.Fprintf(out, "That pipeline belongs to workspace %s, not to %s, so it says nothing about\n"+
					"where this checkout now deploys. Nothing was guessed in its place.\n", previous, workspaceID)
				fmt.Fprintf(out, "\nUntil a pipeline is recorded, every `asgard-cli pipeline` command here refuses\nto run and lists the candidates:\n\n")
				fmt.Fprintf(out, "    asgard-cli pipeline list\n    asgard-cli pipeline use <id>\n")
				fmt.Fprintf(out, "\nCommit both lines together. A commit carrying the new workspace and the old\npipeline is a pairing that never existed on anybody's disk.\n")
				return nil
			}

			fmt.Fprintf(out, "\nCommit it: whoever clones this repository, and whatever agent works in it,\nthen needs no --workspace.\n")
			if f.Pipeline == "" {
				fmt.Fprintf(out, "\nNo pipeline is recorded yet. `asgard-cli pipeline list` shows what this\nworkspace has, and `asgard-cli pipeline use <id>` records one.\n")
			}
			return nil
		},
	}

	return cmd
}

func newWorkspaceShowCmd() *cobra.Command {
	var (
		profile string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Report which workspace applies here, and why",
		Long: `Report which workspace applies here, and why that one.

The reason is the useful half. A command that acted in the wrong workspace is
the failure the resolution order exists to prevent, and the order is:
--workspace, then ASGARD_WORKSPACE, then the checkout's .asgard-cli.yaml. Those
three are the whole list. **There is no last resort and no machine-wide
default**: every answer is either on the command line or in a committed file, so
two people in the same checkout get the same one.

The "origin remote" line is a statement about this checkout and nothing more.
Nothing compares it to anything: which remote somebody calls origin is their
business, and a repository may have several.

It names the workspace without acting on it, so it is safe to run first when a
command is about to do something that matters.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := checkFormat(format); err != nil {
				return err
			}
			pc, err := resolveContext(cmd, contextOptions{Profile: profile, NeedWorkspace: true})
			if err != nil {
				return err
			}

			// The name is worth a call: an id says nothing about whether it is
			// the right customer. A failure to get one is not fatal - the
			// binding is still what it is.
			name := ""
			if workspaces, err := platform.New(pc.Session, "").ListWorkspaces(cmd.Context()); err == nil {
				for _, w := range workspaces {
					if w.ID == pc.Workspace {
						name = w.Name
						break
					}
				}
			}

			out := cmd.OutOrStdout()
			if format == formatJSON {
				pipeline := ""
				if pc.Binding != nil {
					pipeline = pc.Binding.Pipeline
				}
				return writeJSON(out, map[string]any{
					"profile":   pc.Session.Profile.Name,
					"api":       pc.Session.Profile.API,
					"workspace": pc.Workspace,
					"name":      name,
					"source":    string(pc.WorkspaceSource),
					"origin":    pc.RepoFullName,
					"pipeline":  pipeline,
				})
			}

			fmt.Fprintf(out, "%-11s %s\n", "profile", pc.Session.Profile.Name)
			fmt.Fprintf(out, "%-11s %s\n", "api", pc.Session.Profile.API)
			if pc.RepoFullName != "" {
				// Labelled for what it is. It used to be called "repository",
				// which read as a claim that this checkout is that repository
				// as far as the platform is concerned - and something did once
				// check a pipeline against it.
				fmt.Fprintf(out, "%-11s %s\n", "origin", pc.RepoFullName)
			}
			fmt.Fprintf(out, "%-11s %s", "workspace", pc.Workspace)
			if name != "" {
				fmt.Fprintf(out, "  (%s)", name)
			}
			fmt.Fprintf(out, "\n%-11s %s\n", "chosen by", pc.WorkspaceSource)

			// The pipeline is the other half of the binding, and a checkout
			// missing it is the state `workspace use` leaves behind.
			switch {
			case pc.Binding == nil:
				fmt.Fprintf(out, "%-11s none recorded (no %s here)\n", "pipeline", binding.FileName)
			case pc.Binding.Pipeline == "":
				fmt.Fprintf(out, "%-11s NONE RECORDED - `asgard-cli pipeline use <id>` records one\n", "pipeline")
			default:
				fmt.Fprintf(out, "%-11s %s\n", "pipeline", pc.Binding.Pipeline)
			}
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	return cmd
}
