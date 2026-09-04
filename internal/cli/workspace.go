package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Choose which workspace the platform commands act in",
		Long: `Choose which workspace the platform commands act in.

A workspace is the customer. Everything under "pipeline" is scoped to one, and
the choice is remembered per repository so that it is made once rather than
typed on every command.

    asgard-cli workspace list          what this account can reach
    asgard-cli workspace use <id>      bind this repository to one
    asgard-cli workspace show          which one applies here, and why

Where the choice is kept is deliberate. It is not written into the repository:
the declaration contract for a customer repository is a chart and one
.asgard-pipeline.yaml, and a platform identifier would be a third file it does
not have. It is not on the platform either - a workspace does not know which of
your directories holds its repository. So it lives beside the credentials, keyed
by profile and by the repository's origin remote, and a repository legitimately
bound in both dev and prod keeps one binding for each.

A repository scaffolded by "asgard-cli init" that already records workspace.id
is read from there first, because that file is committed and a teammate cloning
it should not have to choose again.`,
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
without running anything that would act on it.`,
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
			if id, _, err := resolveWorkspace(cmd.Context(), pc.Session, pc.RepoFullName, pc.RepoRoot, ""); err == nil {
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
				fmt.Fprintf(out, "\nNone chosen yet. `asgard-cli workspace use <id>` binds one to this repository.\n")
			}
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	return cmd
}

func newWorkspaceUseCmd() *cobra.Command {
	var (
		profile  string
		asGlobal bool
	)

	cmd := &cobra.Command{
		Use:   "use <workspace-id>",
		Short: "Record which workspace this repository's commands act in",
		Long: `Record which workspace this repository's commands act in.

The binding is keyed by the repository's origin remote and by the profile, so
the same repository bound in dev and in prod keeps both, and a command meant for
one never reaches the other.

    asgard-cli workspace use 1862431170889781248
    asgard-cli workspace use 1862431170889781248 --profile dev
    asgard-cli workspace use 1862431170889781248 --default

--default records it for commands run outside a repository instead of binding it
to one, and is the only form available when there is no origin remote to key on.

The id is not checked against the platform here. "workspace list" is where an id
comes from, and a check would put a network call in the middle of recording a
choice - the first command that acts on it reports a bad one anyway.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID := args[0]

			p, err := auth.ResolveProfile(profile)
			if err != nil {
				return err
			}
			settings, err := auth.LoadSettings()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			if asGlobal {
				settings.SetFallbackWorkspace(p.Name, workspaceID)
				if err := auth.SaveSettings(settings); err != nil {
					return err
				}
				fmt.Fprintf(out, "Recorded %s as the default workspace on %s.\n", workspaceID, p.Name)
				return nil
			}

			_, repoFullName := locateRepo(cmd.Context())
			if repoFullName == "" {
				return errNoRepoToBind
			}
			settings.BindWorkspace(p.Name, repoFullName, workspaceID)
			if err := auth.SaveSettings(settings); err != nil {
				return err
			}
			fmt.Fprintf(out, "Bound %s to workspace %s on %s.\n", repoFullName, workspaceID, p.Name)
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().BoolVar(&asGlobal, "default", false,
		"record it for commands run outside a repository, rather than binding it to this one")
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
the failure the resolution order exists to prevent, and the order has six steps:
--workspace, then ASGARD_WORKSPACE, then a scaffolded repository's own
workspace.id, then what "workspace use" recorded for this repository, then the
recorded default, and finally the only workspace the account can reach when
there is exactly one.

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
				return writeJSON(out, map[string]any{
					"profile":   pc.Session.Profile.Name,
					"api":       pc.Session.Profile.API,
					"workspace": pc.Workspace,
					"name":      name,
					"source":    string(pc.WorkspaceSource),
					"repo":      pc.RepoFullName,
				})
			}

			fmt.Fprintf(out, "%-11s %s\n", "profile", pc.Session.Profile.Name)
			fmt.Fprintf(out, "%-11s %s\n", "api", pc.Session.Profile.API)
			if pc.RepoFullName != "" {
				fmt.Fprintf(out, "%-11s %s\n", "repository", pc.RepoFullName)
			}
			fmt.Fprintf(out, "%-11s %s", "workspace", pc.Workspace)
			if name != "" {
				fmt.Fprintf(out, "  (%s)", name)
			}
			fmt.Fprintf(out, "\n%-11s %s\n", "chosen by", pc.WorkspaceSource)
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	return cmd
}
