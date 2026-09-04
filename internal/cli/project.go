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
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/size"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
)

func newProjectCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "project",
		Short: "List the projects, what each chart declares and what it still lacks",
		Long: `List the projects recorded in ` + config.FileName + `, what each chart declares, and
what it still lacks for the shape it is being built to.

A project is the unit of deployment: one Helm chart, one namespace per
environment. An onboarding usually starts before the split is known, so projects
are added as the engagement discovers them.

**What is missing is arithmetic, not a step.** A shape asks for a set of CR
kinds; this reports the ones the chart does not declare. It carries no claim
about the order they get written in, and a project with a gap is not behind - a
chart is built in whatever order the engagement finds the answers. Where no
shape is declared, nothing is claimed at all: record one with
` + "`asgard-cli project shape <slug> <shape>`" + `.

"Complete for its shape" is not "deployed". Whether a finished chart is waiting
for its first tag or has been live for a month is not a fact about files.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkFormat(format); err != nil {
				return err
			}
			state, err := loadState()
			if err != nil {
				return err
			}
			if format == formatJSON {
				return writeJSON(cmd.OutOrStdout(), projectReport(state))
			}
			printProjects(cmd.OutOrStdout(), state)
			return nil
		},
	}

	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	cmd.AddCommand(newProjectAddCmd(), newProjectShapeCmd())

	return cmd
}

// projectJSON is one project's chart, as a record rather than as a column.
//
// There is deliberately no "stage" field. The chart has kinds it declares and
// kinds its shape asks for and it does not have; a field naming where it is in
// a sequence is a position waiting to be reintroduced.
type projectJSON struct {
	Slug  string `json:"slug"`
	Shape string `json:"shape,omitempty"`
	// Complete is absent, not false, for a project with no declared shape.
	// There is nothing to hold such a chart against, and `false` would be a
	// claim that it is incomplete.
	Complete *bool          `json:"chartComplete,omitempty"`
	Missing  []string       `json:"missing,omitempty"`
	Kinds    map[string]int `json:"kinds"`
}

func projectReport(state stage.State) []projectJSON {
	gaps := map[string]stage.Gap{}
	for _, g := range stage.Gaps(state) {
		gaps[g.Slug] = g
	}
	out := []projectJSON{}
	for _, p := range state.Projects {
		r := projectJSON{Slug: p.Slug, Shape: p.Shape, Kinds: p.Kinds}
		if g, ok := gaps[p.Slug]; ok {
			done := g.Done
			r.Complete, r.Missing = &done, g.Missing
		}
		out = append(out, r)
	}
	return out
}

// printProjects reports what each chart declares and what it still lacks.
//
// This is per project on purpose. One stage was once reported for a whole
// repository, so an engagement with one project live and a second just started
// was told the whole repository was at the second one's step - and the guidance
// said "you have no DataConnector" while a DataConnector had been in production
// for a fortnight.
func printProjects(out io.Writer, state stage.State) {
	if len(state.Projects) == 0 {
		fmt.Fprintf(out, "No projects yet. A project is one chart in one namespace per environment:\n\n    asgard-cli project add <slug>\n")
		return
	}

	gaps := map[string]stage.Gap{}
	for _, g := range stage.Gaps(state) {
		gaps[g.Slug] = g
	}

	fmt.Fprintf(out, "Projects:\n\n")
	for _, p := range state.Projects {
		shape := p.Shape
		if shape == "" {
			shape = "shape not declared"
		}
		fmt.Fprintf(out, "  %-20s %s\n", p.Slug, shape)
		fmt.Fprintf(out, "  %-20s %s\n", "", p.Summary())
		g, held := gaps[p.Slug]
		switch {
		case !held:
			// With no shape declared there is nothing to hold the chart
			// against, and guessing what it ought to have is the ladder.
			fmt.Fprintf(out, "  %-20s no shape declared, so nothing is claimed about what it lacks\n", "")
			fmt.Fprintf(out, "  %-20s `asgard-cli project shape %s <shape>`, or `asgard-cli size` for the list\n", "", p.Slug)
		case g.Done:
			fmt.Fprintf(out, "  %-20s complete for its shape - which is not the same as deployed\n", "")
		default:
			fmt.Fprintf(out, "  %-20s still declares no %s\n", "", strings.Join(g.Missing, ", and no "))
		}
		fmt.Fprintln(out)
	}
}

func newProjectAddCmd() *cobra.Command {
	var (
		name  string
		envs  []string
		shape string
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
				Shape:        shape,
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

    asgard-cli guide requirements

Connection coordinates belong in the request record and in
chart/values-<env>.yaml. Passwords belong in .env, which is gitignored, and in
app-secret, which infra provisions. Never in a file this repo commits.
`)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the project (defaults to the slug)")
	cmd.Flags().StringSliceVar(&envs, "env", []string{string(config.EnvDev)}, "environment this project deploys to, repeatable")
	// No back-quotes in this usage string: pflag reads a back-quoted word as the
	// flag's argument placeholder, so one here rendered the flag as
	// "--shape asgard-cli project shape".
	cmd.Flags().StringVar(&shape, "shape", "",
		"deployment shape this chart is built to ("+strings.Join(size.Names(), ", ")+"); optional, settable later with 'asgard-cli project shape'")

	return cmd
}

// newProjectShapeCmd records what a project is being built to be.
//
// This exists because the chart cannot say it. A SemanticLayer with nothing
// mounted on it is either a finished Mimir deliverable or an agent nobody has
// written yet, and those are the same files on disk - so the tool guessed, and
// guessed the same way every time.
func newProjectShapeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shape <slug> [shape]",
		Short: "Record which deployment shape a project is being built to",
		Long: `Record which deployment shape a project is being built to.

With no shape argument this prints the project's current one and the choices.

The shape decides **when the chart is finished**, which is the thing the files
cannot say. A chart holding a SemanticLayer and nothing else is either a
finished Mimir deliverable or an agent nobody has written yet; they are
identical on disk. Undeclared, a project is assumed to need an entry point,
because most do.

    asgard-cli project shape insight mimir-dashboard

The one that changes behaviour today is ` + "`mimir-dashboard`" + `: it declares that the
customer reaches this through Data Insight rather than through anything the
chart contains, so no Agent, Toolset, BotProvider or entry point is written and
nothing reports the chart as lacking one. Read the shape and the trap that comes
with it before declaring it:

    asgard-cli usecase mimir-dashboard
    asgard-cli guide read-path

` + "`asgard-cli size`" + ` describes every shape and what each costs before anything is
added to it.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			out := cmd.OutOrStdout()

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
			project, ok := cfg.Project(slug)
			if !ok {
				return fmt.Errorf("no project %q in %s; add it with `asgard-cli project add %s`", slug, config.FileName, slug)
			}

			if len(args) == 1 {
				current := project.Shape
				if current == "" {
					current = "(not declared - assumed to need an entry point)"
				}
				fmt.Fprintf(out, "%-20s %s\n\n", slug, current)
				fmt.Fprintf(out, "The shapes, from deployments in production:\n\n")
				for _, sh := range size.Shapes {
					entry := "declares its own entry point"
					if !sh.HasEntryPoint() {
						entry = "**no entry point** - reached through the product"
					}
					fmt.Fprintf(out, "  %-22s %s\n  %-22s %s\n\n", sh.Name, sh.Audience, "", entry)
				}
				fmt.Fprintf(out, "Read one in full with `asgard-cli size <shape>`.\n")
				return nil
			}

			want := args[1]
			if _, ok := size.Find(want); !ok {
				return fmt.Errorf("%q is not a shape; one of %s", want, strings.Join(size.Names(), ", "))
			}

			was := project.Shape
			project.Shape = want
			if err := cfg.Validate(); err != nil {
				return err
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}

			if was == "" {
				fmt.Fprintf(out, "%s is now %s\n", slug, want)
			} else {
				fmt.Fprintf(out, "%s is now %s, was %s\n", slug, want, was)
			}

			if sh, _ := size.Find(want); !sh.HasEntryPoint() {
				fmt.Fprintf(out, `
This shape has no entry point, so `+"`asgard-cli project`"+` will not report one as
missing; this chart is complete once its read path exists.

**Say so in the chart, next to the layer.** A later reader finds a SemanticLayer
that nothing references, reads it as a missed connection, and binds an Agent to
it - which hands agents deliberately restricted to an API a second path into the
database. Nothing in the gate catches that: cross-reference checking validates
references that exist, never one that should not.

    asgard-cli usecase mimir-dashboard
`)
			}
			return nil
		},
	}

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
