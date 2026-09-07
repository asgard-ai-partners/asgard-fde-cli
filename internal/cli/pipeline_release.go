package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

func newPipelineProjectsCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List the workspace's projects, which a release deploys into",
		Long: `List the workspace's projects.

A project is what a release deploys into: it decides the namespace, and the
platform injects that namespace and the project's main environment id into every
run as ` + "`.Values.asgard.namespace`" + ` and ` + "`.Values.asgard.projectEnvironmentId`" + `. A
chart never writes either of them down.

This is the list ` + "`pipeline release create --project`" + ` takes a value from, so it
also reports whether each project has a main environment. One without is refused
at create time, and finding that out after typing the id is the wrong end of the
mistake - a project created before environments were mandatory can easily have
none.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			projects, err := pc.Client.ListProjects(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, projects)
			}
			if len(projects) == 0 {
				fmt.Fprintf(out, "No projects in workspace %s.\n", pc.Workspace)
				return nil
			}
			// A project with no main environment is refused at create time,
			// and that refusal is a bad way to find out: the id was already
			// typed. Asking each project costs one call, and this list exists
			// precisely to be chosen from.
			for _, p := range projects {
				status := "ready"
				env, err := pc.Client.MainEnvironment(cmd.Context(), p.ID)
				switch {
				case err != nil:
					status = "unknown (" + err.Error() + ")"
				case env == nil:
					status = "NO MAIN ENVIRONMENT - a release cannot be created here"
				default:
					status = "main env " + env.Value
				}
				fmt.Fprintf(out, "%-38s %-24s %-46s %s\n", p.ID, p.Name, p.Namespace, status)
			}
			return nil
		},
	}
	f.register(cmd, false)
	return cmd
}

func newPipelineReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Create and inspect a pipeline's releases",
		Long: `Create and inspect a pipeline's releases.

A release is one chart deployed into one project's namespace, triggered by the
rule its entry in ` + "`.asgard-pipeline.yaml`" + ` declares. The declaration names it; the
platform holds its values; creating it here is what makes the declaration take
effect, because an event matching a release that was never created produces no
run at all.

    asgard-cli pipeline projects
    asgard-cli pipeline release create internal-dev --project <id>
    asgard-cli pipeline release show internal-dev`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newReleaseCreateCmd(), newReleaseShowCmd())
	return cmd
}

func newReleaseCreateCmd() *cobra.Command {
	var (
		f         pipelineFlags
		project   string
		autoApply bool
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a release, bound to one project",
		Long: `Create a release under this pipeline, bound to one project.

The name has to match the one the declaration uses: the platform matches a
trigger to a release by name, so a release called something else is a release no
event ever reaches. A name the declaration does not have is allowed and is
marked "not declared" - which is a warning, not a refusal, because a declaration
can be added afterwards.

--project takes an id, a project name or a namespace from ` + "`pipeline projects`" + `.
The project decides the namespace and cannot be changed afterwards.

--auto-apply skips the review stop: a plan that succeeds applies immediately.
Off by default, and worth leaving off for anything that reaches a cluster
somebody cares about.

WHAT IT CREATES. The platform prepares the namespace side straight away: this
release's own Secret and ConfigMap (empty at first), its deploy identity and its
RBAC. The helm release itself does not exist until a run applies one.

The helm release name is ` + "`iac-<name>`" + ` and has to be unique within the project, so
a second pipeline declaring the same release name against the same project is
refused here - that is the lock that stops two pipelines overwriting each
other's deployment.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			actingOn(cmd, pc.Session)
			p, err := resolvePipeline(cmd.Context(), pc, f.pipeline)
			if err != nil {
				return err
			}
			projectID, projectLabel, err := resolveProject(cmd.Context(), pc, project)
			if err != nil {
				return err
			}

			// Warn before creating, not after: a name that matches nothing in
			// the declaration is almost always a typo, and it is far cheaper to
			// see it here than to wonder later why a tag produced no run.
			full, err := pc.Client.GetPipeline(cmd.Context(), p.PipelineId)
			if err != nil {
				return err
			}
			declared := false
			var names []string
			for _, d := range full.DeclaredReleases {
				names = append(names, d.Name)
				if d.Name == name {
					declared = true
				}
			}
			if !declared {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"WARNING: %q is not declared in %s at %s. It will be created and marked\n"+
						"\"not declared\", and no push will ever trigger it until the declaration names it.\n"+
						"Declared: %s\n\n",
					name, full.ConfigPath, declaredRef(full), strings.Join(names, ", "))
			}

			created, err := pc.Client.CreateRelease(cmd.Context(), p.PipelineId, platform.CreateReleaseInput{
				Name:      name,
				ProjectId: projectID,
				AutoApply: autoApply,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, created)
			}
			fmt.Fprintf(out, "Created release %s in project %s.\n", created.Name, projectLabel)
			printRelease(out, created)
			fmt.Fprintf(out, "\nFill in the values it declares, then trigger it:\n\n"+
				"    asgard-cli pipeline variables list --release %s\n", created.Name)
			return nil
		},
	}
	f.register(cmd, true)
	cmd.Flags().StringVar(&project, "project", "", "project id, name or namespace to deploy into (required)")
	cmd.Flags().BoolVar(&autoApply, "auto-apply", false, "apply a successful plan without stopping for review")
	return cmd
}

func newReleaseShowCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show one release",
		Long: `Show one release: where it deploys, what it declares, and which config that
declaration was read from.

That last one matters. A release that has run reads the declarations of its own
most recent run; one that has not falls back to the pipeline's snapshot. It is
the only explanation for why a variable is marked Orphan, so it is printed
rather than assumed.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			p, err := resolvePipeline(cmd.Context(), pc, f.pipeline)
			if err != nil {
				return err
			}
			rel, err := resolveRelease(cmd.Context(), pc, p, args[0])
			if err != nil {
				return err
			}
			full, err := pc.Client.GetRelease(cmd.Context(), rel.ReleaseId)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, full)
			}
			printRelease(out, full)
			return nil
		},
	}
	f.register(cmd, true)
	return cmd
}

// resolveProject turns an id, a name or a namespace into a project id.
func resolveProject(ctx context.Context, pc *platformContext, want string) (id, label string, err error) {
	projects, err := pc.Client.ListProjects(ctx)
	if err != nil {
		return "", "", err
	}
	if want == "" {
		var b strings.Builder
		fmt.Fprintf(&b, "--project is required; workspace %s has %d (`asgard-cli pipeline projects`\nsays which of them have a main environment, which a release needs):\n", pc.Workspace, len(projects))
		for _, p := range projects {
			fmt.Fprintf(&b, "  %-38s %-24s %s\n", p.ID, p.Name, p.Namespace)
		}
		return "", "", fmt.Errorf("%s", b.String())
	}
	for _, p := range projects {
		if p.ID == want || p.Name == want || p.Namespace == want {
			return p.ID, p.Name + " (" + p.Namespace + ")", nil
		}
	}
	return "", "", fmt.Errorf("no project %q in workspace %s; `asgard-cli pipeline projects` lists them", want, pc.Workspace)
}

// declaredRef names the ref a pipeline's declarations were read from, for a
// message that has to say where a name was looked for.
func declaredRef(p *platform.Pipeline) string {
	if p.ConfigRef != "" {
		return p.ConfigRef
	}
	if p.LastConfigSync != nil && p.LastConfigSync.Ref != "" {
		return p.LastConfigSync.Ref
	}
	return p.DefaultBranch
}

func printRelease(out interface{ Write([]byte) (int, error) }, r *platform.Release) {
	fmt.Fprintf(out, "%-16s %s\n", "release", r.Name)
	fmt.Fprintf(out, "%-16s %s\n", "id", r.ReleaseId)
	fmt.Fprintf(out, "%-16s %s\n", "namespace", r.Namespace)
	fmt.Fprintf(out, "%-16s %s\n", "helm release", r.HelmReleaseName)
	fmt.Fprintf(out, "%-16s %s\n", "app secret", r.AppSecretName)
	fmt.Fprintf(out, "%-16s %s\n", "app configmap", r.AppConfigMapName)
	fmt.Fprintf(out, "%-16s %v\n", "auto apply", r.AutoApply)
	fmt.Fprintf(out, "%-16s %s\n", "state", r.State)

	if r.Declaration == nil {
		fmt.Fprintf(out, "%-16s NOT DECLARED\n", "declaration")
	} else if r.Declaration.On != nil {
		fmt.Fprintf(out, "%-16s %s %s -> %s\n", "trigger",
			r.Declaration.On.Type, r.Declaration.On.Pattern, r.Declaration.Chart)
	}
	if r.DeclarationSource != nil {
		fmt.Fprintf(out, "%-16s %s\n", "declared", r.DeclarationSource.Describe())
	}
	if r.PendingDeploy {
		// True on a release nobody has edited yet as well as on one whose
		// values were just changed: what it says is "the platform holds
		// something the cluster does not", and a release that has never run
		// qualifies. So it must not claim somebody saved an edit.
		fmt.Fprintf(out, "\nThe platform holds values the cluster does not have yet. A run is what sends them.\n")
	}
}

func newPipelineManifestCmd() *cobra.Command {
	var (
		f                    pipelineFlags
		release              string
		includeManagedFields bool
		summary              bool
	)

	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Read back what a release has on the cluster right now",
		Long: `Read back the objects a release's last helm revision deployed, as the cluster
holds them right now.

    asgard-cli pipeline manifest --release internal-dev --summary
    asgard-cli pipeline manifest --release internal-dev --format json

THIS IS LIVE STATE, NOT A COMPARISON. The platform has no diff endpoint on
purpose: comparing the cluster with the IaC source needs the source, and the
source is here. So this hands back the objects and the comparing is yours -
render the chart locally with helm and compare, and the three-way answer is
worth more than the two-way one:

    live vs the render of the deployed commit   -> somebody changed the cluster
    the render of HEAD vs the deployed commit   -> the repository is ahead

The object list comes from the helm release record's stored manifest, not a
label selector, because helm records ownership in an annotation that cannot be
selected on - and because a selector silently omits an object somebody deleted,
while the manifest reports it with found=false. Being deleted out of band is as
important as being edited.

managedFields is stripped unless --include-managed-fields. It is large and
mostly noise, and it is also the only record of which field manager owns which
field, which is what separates "the pipeline set this" from "somebody changed it
in the UI".

A release that has never deployed has no manifest, and says so.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if release == "" {
				return fmt.Errorf("--release is required")
			}
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			p, err := resolvePipeline(cmd.Context(), pc, f.pipeline)
			if err != nil {
				return err
			}
			rel, err := resolveRelease(cmd.Context(), pc, p, release)
			if err != nil {
				return err
			}
			live, err := pc.Client.GetLiveManifest(cmd.Context(), rel.ReleaseId, includeManagedFields)
			if err != nil {
				if platform.NotFound(err) {
					return fmt.Errorf("release %s has never deployed, so it has nothing on the cluster yet", rel.Name)
				}
				return err
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, live)
			}

			fmt.Fprintf(out, "helm revision %d, %s", live.HelmRevision, live.HelmStatus)
			if live.DeployedAt != nil {
				fmt.Fprintf(out, ", deployed %s", live.DeployedAt.Local().Format("2006-01-02 15:04"))
			}
			fmt.Fprintf(out, "\n\n")

			var missing, failed int
			for _, o := range live.Objects {
				switch {
				case o.Error != "":
					failed++
				case !o.Found:
					missing++
				}
			}
			for _, o := range live.Objects {
				state := "ok"
				switch {
				case o.Error != "":
					state = "READ FAILED: " + o.Error
				case !o.Found:
					state = "DELETED OUT OF BAND"
				}
				fmt.Fprintf(out, "%-24s %-40s %s\n", o.Kind, o.Name, state)
			}
			fmt.Fprintf(out, "\n%d object(s)", len(live.Objects))
			if missing > 0 {
				fmt.Fprintf(out, ", %d deleted out of band", missing)
			}
			if failed > 0 {
				fmt.Fprintf(out, ", %d unreadable", failed)
			}
			fmt.Fprintln(out)
			if summary {
				return nil
			}
			fmt.Fprintf(out, "\nFull objects are in --format json; each carries the live YAML verbatim.\n")
			return nil
		},
	}
	f.register(cmd, true)
	cmd.Flags().StringVar(&release, "release", "", "release whose deployed objects to read back (required)")
	cmd.Flags().BoolVar(&includeManagedFields, "include-managed-fields", false,
		"keep metadata.managedFields, which says which field manager owns which field")
	cmd.Flags().BoolVar(&summary, "summary", false, "list the objects without the closing note")
	return cmd
}
