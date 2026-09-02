package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

func newInitCmd() *cobra.Command {
	var (
		workspaceID   string
		workspaceSlug string
		workspaceName string
		projects      []string
		force         bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bind the current directory to a workspace",
		Long: `Bind the current directory to a workspace by writing ` + config.FileName + `.

RUN THIS IN THE DIRECTORY THAT IS TO BECOME THE WORKSPACE. The repository root
is the workspace - one customer, one repository - and every project lives inside
it. Running it one level up, in the directory where other repositories live,
binds that directory instead and takes the slug from its name.

Check the slug it prints before building anything on top of it.

The workspace is the customer. Its slug is what the repository name and every
namespace are derived from:

  repository   <slug>` + config.RepoSuffix + `
  namespace    asgard-<slug>-<project>-<env>

--workspace-slug defaults to the directory name with a trailing ` + config.RepoSuffix + `
removed, so running this inside acme` + config.RepoSuffix + ` yields acme.
--workspace-name defaults to the slug.

--workspace-id is optional. Nothing this CLI generates reads it - namespaces come
from the slug - so waiting for the platform to issue one should not block the
work that comes before it. Fill it in whenever you have it:

  asgard-cli init --workspace-id ws_xxxxxxxx

On an already-initialised repository that sets the id and changes nothing else,
so it needs no --force. Changing an id that is already recorded does, because
that is a different workspace rather than a missing fact.

Projects are usually added later with "asgard-cli project add", once the
engagement knows how the work splits. --project is a shortcut for when it is
already known; it may be repeated, and each project starts with the dev
environment only.

If the config file already exists nothing is changed and the current settings are
printed; pass --force to rebind. A rebind keeps the projects already recorded,
because their charts are on disk either way - pass --project to replace the list
instead.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}
			path := filepath.Join(dir, config.FileName)
			out := cmd.OutOrStdout()

			// Already initialised: report the current state, keeping init idempotent.
			if !force {
				existing, err := config.Load(path)
				switch {
				case err == nil:
					if err := existing.Validate(); err != nil {
						return fmt.Errorf("%s is incomplete, pass --force to reinitialise:\n%w", config.FileName, err)
					}
					// Filling in a workspace id that was left empty is not a
					// rebind, so it does not need --force. Replacing one that is
					// already there is, because that names a different
					// workspace.
					switch {
					case workspaceID == "" || workspaceID == existing.Workspace.ID:
						fmt.Fprintf(out, "%s already exists\n", config.FileName)
					case existing.Workspace.HasID():
						return fmt.Errorf("%s already records workspace.id %q; pass --force to rebind this repository to %q",
							config.FileName, existing.Workspace.ID, workspaceID)
					default:
						existing.Workspace.ID = workspaceID
						if err := config.Save(path, existing); err != nil {
							return err
						}
						fmt.Fprintf(out, "Set workspace.id in %s\n", path)
					}
					printConfig(out, existing)
					return nil
				case !errors.Is(err, config.ErrNotFound):
					// The file is there but unreadable (broken JSON, permissions);
					// never silently overwrite it.
					return err
				}
			}

			if workspaceSlug == "" {
				workspaceSlug = defaultSlug(dir)
			}
			if workspaceName == "" {
				workspaceName = workspaceSlug
			}

			cfg := &config.Config{
				Workspace: config.Workspace{
					ID:   workspaceID,
					Slug: workspaceSlug,
					Name: workspaceName,
				},
			}

			// Rebinding the workspace does not delete the projects. They exist
			// on disk as charts, and dropping them from the config leaves a
			// repository whose charts nothing declares - which `render` will not
			// touch and `check` did not used to notice. Carry them over unless
			// --project is given, which is an explicit new list.
			if force && len(projects) == 0 {
				if existing, err := config.Load(path); err == nil {
					cfg.Projects = existing.Projects
					cfg.OLAPOnlyLayers = existing.OLAPOnlyLayers
				}
			}
			for _, slug := range projects {
				cfg.Projects = append(cfg.Projects, config.Project{
					Slug:         slug,
					Name:         slug,
					Environments: []config.Env{config.EnvDev},
				})
			}

			if err := cfg.Validate(); err != nil {
				return err
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}

			fmt.Fprintf(out, "Created %s\n", path)
			printConfig(out, cfg)
			return nil
		},
	}

	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "workspace id issued by the Asgard platform (optional; can be set later)")
	cmd.Flags().StringVar(&workspaceSlug, "workspace-slug", "", "slug used to derive the repository and namespace names (defaults to the directory name without "+config.RepoSuffix+")")
	cmd.Flags().StringVar(&workspaceName, "workspace-name", "", "display name for the customer (defaults to the slug)")
	cmd.Flags().StringSliceVar(&projects, "project", nil, "project slug to create up front, repeatable (defaults to none)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite an existing "+config.FileName)

	return cmd
}

// defaultSlug derives the workspace slug from the directory name. The onboarding
// produces a repository called <slug>-asgard-kube, so running init inside one
// should recover the slug rather than repeat the suffix.
func defaultSlug(dir string) string {
	return strings.TrimSuffix(filepath.Base(dir), config.RepoSuffix)
}

// printConfig reports what a command just wrote or found, including the two
// derived names, which are the facts the FDE needs next.
func printConfig(out io.Writer, cfg *config.Config) {
	if cfg.Workspace.HasID() {
		fmt.Fprintf(out, "  workspace.id    %s\n", cfg.Workspace.ID)
	} else {
		fmt.Fprintf(out, "  workspace.id    (not set yet)\n")
	}
	fmt.Fprintf(out, "  workspace.slug  %s\n", cfg.Workspace.Slug)
	fmt.Fprintf(out, "  workspace.name  %s\n", cfg.Workspace.Name)
	fmt.Fprintf(out, "  repository      %s\n", cfg.RepoName())

	if !cfg.Workspace.HasID() {
		fmt.Fprintf(out, "\nThe workspace id is what the platform knows this customer by. Nothing here\n"+
			"needs it yet - namespaces come from the slug - so it can wait until the\n"+
			"platform has issued one:\n\n    asgard-cli init --workspace-id ws_xxxxxxxx\n")
	}

	if len(cfg.Projects) == 0 {
		fmt.Fprintf(out, "\nNo projects yet. Add one with `asgard-cli project add <slug>`.\n")
		return
	}

	fmt.Fprintf(out, "\n%d project(s):\n", len(cfg.Projects))
	for _, p := range cfg.Projects {
		fmt.Fprintf(out, "  %s\n", p.Slug)
		for _, env := range p.Environments {
			fmt.Fprintf(out, "    %-5s %s\n", env, cfg.Namespace(p.Slug, env))
		}
	}
}
