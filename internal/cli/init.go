package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

func newInitCmd() *cobra.Command {
	var (
		workspaceID   string
		workspaceName string
		force         bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create " + config.FileName + " in the current directory",
		Long: `Create ` + config.FileName + ` in the current directory, binding it to a workspace.

If the config file already exists nothing is changed and the current settings are
printed; pass --force to rebind. Creating a new config requires --workspace-id.
--workspace-name defaults to the directory name.`,
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
						return fmt.Errorf("%s is incomplete; run `asgard-cli lint` to see the problems, or pass --force to reinitialise", config.FileName)
					}
					fmt.Fprintf(out, "%s already exists\n", config.FileName)
					printWorkspace(out, existing)
					return nil
				case !errors.Is(err, config.ErrNotFound):
					// The file is there but unreadable (broken JSON, permissions);
					// never silently overwrite it.
					return err
				}
			}

			if workspaceID == "" {
				return fmt.Errorf("creating %s requires --workspace-id, for example:\n  asgard-cli init --workspace-id ws_xxxxxxxx", config.FileName)
			}
			if workspaceName == "" {
				workspaceName = filepath.Base(dir)
			}

			cfg := &config.Config{
				Workspace: config.Workspace{
					ID:   workspaceID,
					Name: workspaceName,
				},
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}

			fmt.Fprintf(out, "Created %s\n", path)
			printWorkspace(out, cfg)
			return nil
		},
	}

	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "workspace id (required when creating a new config)")
	cmd.Flags().StringVar(&workspaceName, "workspace-name", "", "workspace name (defaults to the directory name)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite an existing "+config.FileName)

	return cmd
}

func printWorkspace(out io.Writer, cfg *config.Config) {
	fmt.Fprintf(out, "  workspace.id   %s\n", cfg.Workspace.ID)
	fmt.Fprintf(out, "  workspace.name %s\n", cfg.Workspace.Name)
}
