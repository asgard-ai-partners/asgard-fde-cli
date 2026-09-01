package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/render"
)

func newRenderCmd() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use:   "render <project> <dev|prod>",
		Short: "Render a project's chart the way CD will",
		Long: `Render a project's chart to stdout, the way CD will.

The namespace, the release name, the values files and the order they overlay in
all come from projects/<project>/deploy.yaml - the same declaration CI reads - so
what this prints is what a tag would apply. A project that does not declare the
environment is refused rather than rendered, because not declaring it is a
deliberate decision not to deploy there.

Everything except the manifests goes to stderr, so it pipes:

    asgard-cli render internal dev
    asgard-cli render internal dev > .out/rendered.yaml
    asgard-cli verify internal

It needs helm on PATH, and nothing else. This replaces common/render.sh, which
was bash and called yq, so neither ran on Windows without WSL - while helm
itself has a native Windows build. Everything here is one binary and one helm.

**It renders only, and there is no install path.** Deployment is CD-only: a
Syncer pins its revision to the chart's appVersion, and only CI stamps the
release tag in. A local helm upgrade writes the placeholder version as a git ref
that does not exist, and the Syncer then fails to clone on every single run.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, env := args[0], args[1]

			if !config.Env(env).Valid() {
				return fmt.Errorf("env must be dev or prod, not %q", env)
			}

			root, cfg, err := loadRepo()
			if err != nil {
				return err
			}
			if _, ok := cfg.Project(project); !ok {
				return fmt.Errorf("no project %q in %s; add it with `asgard-cli project add %s`",
					project, config.FileName, project)
			}

			errOut := cmd.ErrOrStderr()
			res, err := render.Run(cmd.Context(), render.Options{
				Root:    root,
				Project: project,
				Env:     env,
			}, cmd.OutOrStdout(), errOut)
			if err != nil {
				return err
			}

			// The summary goes to stderr so that stdout stays exactly the
			// manifests, which is what makes the pipe forms above work.
			if !quiet {
				fmt.Fprintf(errOut, "rendered %s (%s) -> namespace %s, release %s\n",
					project, env, res.Namespace, res.Release)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&quiet, "quiet", false, "do not print the summary line to stderr (defaults to printing it)")

	return cmd
}
