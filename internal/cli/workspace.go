package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
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

Which workspace is the one thing neither the repository nor the platform can
answer alone. The pipeline follows from the origin remote, and the releases,
their keys and their triggers are all in the declaration.

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

func newWorkspaceUseCmd() *cobra.Command {
	var (
		profile  string
		asGlobal bool
	)

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
    asgard-cli workspace use 1862431170889781248 --default

--default records it for this machine instead, for commands run where there is
no declaration to write beside. It is per profile.

Anything already in the file is kept, so recording a workspace never drops a
pipeline id.

The id is not checked against the platform here. "workspace list" is where one
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

			// Anything already recorded is kept: writing a workspace must not
			// silently drop a pipeline id somebody relied on.
			f := &binding.File{Version: 1}
			if existing, err := binding.Load(bindPath); err == nil {
				f = existing
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
			fmt.Fprintf(out, "\nCommit it: whoever clones this repository, and whatever agent works in it,\nthen needs no --workspace.\n")
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().BoolVar(&asGlobal, "default", false,
		"record it for this machine, for commands run where there is no declaration to write beside")
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
--workspace, then ASGARD_WORKSPACE, then the checkout's .asgard-cli.yaml, then
the default recorded on this machine, and finally the only workspace the account
can reach when there is exactly one.

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
