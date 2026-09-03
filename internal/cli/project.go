package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage the projects recorded in " + config.FileName,
		Long: `Manage the projects recorded in ` + config.FileName + `.

A project is the unit of deployment: one Helm chart, one namespace per
environment. An onboarding usually starts before the split is known, so projects
are added as the engagement discovers them.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newProjectAddCmd())

	return cmd
}

func newProjectAddCmd() *cobra.Command {
	var (
		name string
		envs []string
	)

	cmd := &cobra.Command{
		Use:   "add <slug>",
		Short: "Add a project to " + config.FileName,
		Long: `Add a project to ` + config.FileName + `.

The slug becomes part of every namespace this project deploys to
(asgard-<workspace>-<slug>-<env>), so keep it short: names derived from a
namespace inherit its length, and Kubernetes caps a namespace at 63 characters.

--env may be repeated and defaults to dev. The two environments are independent:
a project may declare dev only, prod only, or both. Adding an environment later
means running this command again with --force, or editing the config.

If the workspace id is still unset this says so, because a project is what the
platform deploys and so is the point at which the id is worth chasing. It is a
reminder, not a gate - nothing rendered from this repository reads it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}
			path, err := config.Find(dir)
			if err != nil {
				if errors.Is(err, config.ErrNotFound) {
					return fmt.Errorf("no %s found; run `asgard-cli init` first", config.FileName)
				}
				return err
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			if _, exists := cfg.Project(slug); exists {
				return fmt.Errorf("project %q already exists in %s", slug, config.FileName)
			}

			project := config.Project{
				Slug:         slug,
				Name:         name,
				Environments: parseEnvs(envs),
			}
			if project.Name == "" {
				project.Name = slug
			}

			cfg.Projects = append(cfg.Projects, project)
			if err := cfg.Validate(); err != nil {
				return err
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Added project %q to %s\n", slug, path)
			for _, env := range project.Environments {
				fmt.Fprintf(out, "  %-5s %s\n", env, cfg.Namespace(slug, env))
			}

			// Write the project's chart skeleton now rather than leaving the
			// repo in a state where the config declares a project that has no
			// deploy.yaml. Adding a project used to require remembering to run
			// scaffold again, and forgetting produced a repo where `check`
			// passed and `render` failed. Existing files are left alone, so
			// this is safe on a repo that already has the project.
			root := filepath.Dir(path)
			if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err == nil {
				written, err := scaffold.Write(root, cfg, false)
				if err != nil {
					return fmt.Errorf("write the chart skeleton for %q: %w", slug, err)
				}
				var created int
				for _, r := range written {
					if r.Status == scaffold.Created {
						created++
					}
				}
				fmt.Fprintf(out, "\nWrote %d file(s) for it, including projects/%s/deploy.yaml.\n", created, slug)
			} else {
				fmt.Fprintf(out, "\nThe repository skeleton is not written yet, so this project has no chart:\n"+
					"    asgard-cli scaffold\n")
			}

			// A project is the first point where the workspace id stops being a
			// formality: it names the workspace the platform will deploy this
			// into. Still not a blocker - nothing rendered reads it - but this
			// is the moment to go and get it.
			if !cfg.Workspace.HasID() {
				fmt.Fprintf(out, "\nworkspace.id is still unset in %s. Nothing here needs it - the\n"+
					"namespaces above come from the slug - but a project is what the platform\n"+
					"deploys, so this is the point to ask for it:\n\n"+
					"    asgard-cli init --workspace-id ws_xxxxxxxx\n", config.FileName)
			}

			// Both of these are ordering traps rather than things to look up
			// later: getting either wrong fails in CD, not here.
			fmt.Fprintf(out, `
Before the first deploy of each environment:
  1. tf-asgard must create the namespace and its app-secret first. Declaring an
     environment before they exist makes the next tag fail at helm upgrade.
     platformMainEnvironmentId is written empty in chart/values-<env>.yaml and
     the platform only issues it once that namespace exists, so an empty one is
     correct today and fatal at the first tag. "asgard-cli verify" warns until
     it is filled in; that warning is the reminder, not this line.
  2. the project may need at least one Syncer, and whether it does is one "if"
     in your own CD - some workflows count what the chart declares and skip the
     step at zero, some wait 180s for a CronJob labelled
     asgard-ai.com/syncer-name and exit 1. A production chart runs today with
     none. Check before the first tag:
       grep -n syncer-name -A15 .github/workflows/*.y*ml

A new project means a new audience, which is a thing to have asked rather than
assumed - along with which systems it reads and how each one is reached:

    asgard-cli next --stage requirements

Connection coordinates belong in the request record and in
chart/values-<env>.yaml. Passwords belong in .env, which is gitignored, and in
app-secret, which infra provisions. Never in a file this repo commits.
`)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the project (defaults to the slug)")
	cmd.Flags().StringSliceVar(&envs, "env", []string{string(config.EnvDev)}, "environment this project deploys to, repeatable")

	return cmd
}

// parseEnvs converts flag strings to Env values without judging them; Validate
// reports an unknown symbol along with everything else that is wrong.
func parseEnvs(values []string) []config.Env {
	envs := make([]config.Env, len(values))
	for i, v := range values {
		envs[i] = config.Env(v)
	}
	return envs
}
