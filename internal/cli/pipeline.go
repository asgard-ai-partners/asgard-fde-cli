package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/binding"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Deploy this repository through the platform's IaC pipeline",
		Long: `Deploy this repository through the platform's IaC pipeline.

A pipeline binds one repository to a set of releases; a release deploys one
chart into one project's namespace when a tag or branch matches its rule; each
trigger produces a run that walks Plan, Review and Apply. What a repository
declares is the one ` + "`.asgard-pipeline.yaml`" + ` at its root; the values are held on
the platform, never in the repository.

    asgard-cli pipeline connect                connect GitHub to this workspace
    asgard-cli pipeline connections            the installations already connected
    asgard-cli pipeline repos --connection X   what one of them can reach
    asgard-cli pipeline create --name p --repo R
    asgard-cli pipeline show                   the pipeline bound to this checkout
    asgard-cli pipeline releases               its releases, and the ghost rows

These commands are a wrapper over the platform's API and hold no rules of their
own. **The checking runs on the platform**, because the checks worth the most -
the apiserver's own CEL, pattern and required validation of every rendered CR -
need a cluster, and no cluster credential is ever issued to a client. So the
loop is: change the chart, check what can be checked locally with ` + "`helm lint`" + `
and ` + "`asgard-cli verify`" + `, push, and read the plan back.

Which pipeline a command acts on is worked out from the checkout's origin
remote, so no platform identifier is written into the repository. Which
workspace is ` + "`asgard-cli workspace`" + `.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(
		newPipelineConnectCmd(),
		newPipelineConnectionsCmd(),
		newPipelineDeliveriesCmd(),
		newPipelineReposCmd(),
		newPipelineCreateCmd(),
		newPipelineListCmd(),
		newPipelineShowCmd(),
		newPipelineManifestCmd(),
		newPipelineProjectsCmd(),
		newPipelineReleaseCmd(),
		newPipelineReleasesCmd(),
		newPipelineRunsCmd(),
		newPipelineVariablesCmd(),
	)
	return cmd
}

// pipelineFlags are the ones every subcommand of `pipeline` carries.
type pipelineFlags struct {
	profile   string
	workspace string
	format    string
	// pipeline names one explicitly, by id or by name, instead of deriving it
	// from the checkout.
	pipeline string
}

func (f *pipelineFlags) register(cmd *cobra.Command, withPipeline bool) {
	addProfileFlag(cmd, &f.profile)
	cmd.Flags().StringVar(&f.workspace, workspaceFlag, "",
		"workspace to act in; defaults to what `asgard-cli workspace use` recorded for this repository")
	cmd.Flags().StringVar(&f.format, formatFlag, formatText, formatUsage)
	if withPipeline {
		cmd.Flags().StringVar(&f.pipeline, "pipeline", "",
			"pipeline id or name; defaults to the one bound to this checkout's origin remote")
	}
}

func (f *pipelineFlags) context(cmd *cobra.Command) (*platformContext, error) {
	if err := checkFormat(f.format); err != nil {
		return nil, err
	}
	return resolveContext(cmd, contextOptions{
		Profile:       f.profile,
		Workspace:     f.workspace,
		NeedWorkspace: true,
	})
}

// resolvePipeline finds the pipeline a command acts on.
//
// Named explicitly, it is looked up by id or name. Otherwise it is the one
// bound to this checkout's origin remote - which is what keeps platform
// identifiers out of the repository, and is unambiguous unless a monorepo binds
// the same repository twice at different config paths.
func resolvePipeline(ctx context.Context, pc *platformContext, nameOrID string) (*platform.Pipeline, error) {
	pipelines, err := pc.Client.ListPipelines(ctx)
	if err != nil {
		return nil, err
	}

	if nameOrID != "" {
		for _, p := range pipelines {
			if p.PipelineId == nameOrID || p.Name == nameOrID {
				return p, nil
			}
		}
		return nil, fmt.Errorf("no pipeline %q in workspace %s; `asgard-cli pipeline list` shows what is there", nameOrID, pc.Workspace)
	}

	if pc.Binding != nil && pc.Binding.Pipeline != "" {
		return pipelineFromBinding(pc, pipelines)
	}

	if pc.RepoFullName == "" {
		return nil, fmt.Errorf("not in a git checkout with a recognisable origin remote, so there is no repository to match a pipeline against; name one with --pipeline")
	}

	var matches []*platform.Pipeline
	for _, p := range pipelines {
		if p.RepoFullName == pc.RepoFullName {
			matches = append(matches, p)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("no pipeline in workspace %s binds %s; create one with `asgard-cli pipeline create`", pc.Workspace, pc.RepoFullName)
	}

	// A monorepo binding the same repository at two config paths is legitimate,
	// and the config path is what tells them apart. Recording one in the
	// checkout's binding is how that is answered once rather than every time.
	var b strings.Builder
	fmt.Fprintf(&b, "%d pipelines bind %s, so which one has to be said:\n", len(matches), pc.RepoFullName)
	for _, p := range matches {
		fmt.Fprintf(&b, "  %-22s %-20s %s\n", p.PipelineId, p.Name, p.ConfigPath)
	}
	fmt.Fprintf(&b, "\nPass --pipeline, or record the right one in %s beside its declaration.\n", binding.FileName)
	return nil, fmt.Errorf("%s", b.String())
}

// pipelineFromBinding uses the pipeline the checkout's `.asgard-cli.yaml`
// records, after checking that it is still this repository's.
//
// The check is the point. A binding file is committed, so the way it goes wrong
// is by being copied into another repository along with everything else - and a
// pipeline id that still resolves, against a repository it no longer describes,
// would deploy the wrong chart to a real namespace without a word.
func pipelineFromBinding(pc *platformContext, pipelines []*platform.Pipeline) (*platform.Pipeline, error) {
	want := pc.Binding.Pipeline
	for _, p := range pipelines {
		if p.PipelineId != want {
			continue
		}
		if pc.RepoFullName != "" && p.RepoFullName != pc.RepoFullName {
			return nil, fmt.Errorf(
				"%s records pipeline %s, but that pipeline is bound to %s and this checkout is %s.\n"+
					"The file was most likely copied from another repository. Fix it, or pass --pipeline.",
				pc.Binding.Path, want, p.RepoFullName, pc.RepoFullName)
		}
		return p, nil
	}
	return nil, fmt.Errorf(
		"%s records pipeline %s, which does not exist in workspace %s.\n"+
			"Either the wrong workspace is in effect (`asgard-cli workspace show` says which and why),\n"+
			"or the pipeline was deleted. `asgard-cli pipeline list` shows what is there.",
		pc.Binding.Path, want, pc.Workspace)
}

func newPipelineConnectionsCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "connections",
		Short: "List the workspace's VCS connections",
		Long: `List the workspace's VCS connections.

A connection is one GitHub App installation, and it is what decides which
repositories a pipeline can bind. A connection whose installation was removed on
GitHub shows as revoked: its pipelines stop receiving events and manual runs
fail at Checkout until it is installed again.

Creating one is ` + "`asgard-cli pipeline connect`" + `.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			conns, err := pc.Client.ListConnections(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, conns)
			}
			if len(conns) == 0 {
				fmt.Fprintf(out, "No connections in workspace %s.\n\nConnect one:\n\n    asgard-cli pipeline connect\n", pc.Workspace)
				return nil
			}
			for _, c := range conns {
				fmt.Fprintf(out, "%-22s %-24s %-8s %s (%d pipeline(s))\n",
					c.ConnectionId, c.AccountLogin, c.Status, c.Provider, c.PipelineCount)
			}
			return nil
		},
	}
	f.register(cmd, false)
	return cmd
}

func newPipelineReposCmd() *cobra.Command {
	var (
		f          pipelineFlags
		connection string
	)

	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List the repositories a connection can reach",
		Long: `List the repositories a connection can reach.

The list comes from GitHub's own view of the installation, so a repository
missing from it is a repository the App was not granted - fix that on GitHub,
where the installation's repository selection lives, and it appears here on the
next call.

A repository already bound by another pipeline of this workspace is marked.
That is not a refusal: one repository may carry several pipelines as long as
their config paths differ, which is how a monorepo holds two independent sets of
releases.

--connection defaults to the only one when the workspace has exactly one.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			id, err := resolveConnection(cmd.Context(), pc, connection)
			if err != nil {
				return err
			}
			repos, err := pc.Client.ListRepositories(cmd.Context(), id)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, repos)
			}
			if len(repos) == 0 {
				fmt.Fprintf(out, "Connection %s can reach no repositories.\n", id)
				return nil
			}
			for _, r := range repos {
				note := ""
				if r.AlreadyBoundByPipelineId != "" {
					note = "  (already bound by pipeline " + r.AlreadyBoundByPipelineId + ")"
				}
				fmt.Fprintf(out, "%-22s %-52s %s%s\n", r.RepositoryId, r.FullName, r.DefaultBranch, note)
			}
			return nil
		},
	}
	f.register(cmd, false)
	cmd.Flags().StringVar(&connection, "connection", "",
		"connection id; defaults to the only one when the workspace has exactly one")
	return cmd
}

// resolveConnection picks the connection to use, defaulting only when there is
// no choice to make.
func resolveConnection(ctx context.Context, pc *platformContext, want string) (string, error) {
	if want != "" {
		return want, nil
	}
	conns, err := pc.Client.ListConnections(ctx)
	if err != nil {
		return "", err
	}
	switch len(conns) {
	case 1:
		return conns[0].ConnectionId, nil
	case 0:
		return "", fmt.Errorf("workspace %s has no VCS connection; `asgard-cli pipeline connect` creates one", pc.Workspace)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "workspace %s has %d connections, so --connection has to say which:\n", pc.Workspace, len(conns))
	for _, c := range conns {
		fmt.Fprintf(&b, "  %-22s %-24s %s\n", c.ConnectionId, c.AccountLogin, c.Status)
	}
	return "", fmt.Errorf("%s", b.String())
}

func newPipelineCreateCmd() *cobra.Command {
	var (
		f          pipelineFlags
		name       string
		connection string
		repo       string
		configPath string
		configRef  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Bind a repository to a new pipeline",
		Long: `Bind a repository to a new pipeline.

--repo takes either the repository id from ` + "`pipeline repos`" + ` or its "owner/name",
which is resolved through the connection. With neither --repo nor a value, the
checkout's own origin remote is used, which is usually what was meant.

    asgard-cli pipeline create --name iac-test
    asgard-cli pipeline create --name iac-test --repo asgard-ai-platform/asgard-iac-test

--config-ref pins the declaration source: the branch or tag the declared-config
snapshot is read from, which is what draws the ghost rows and the trigger
column. Leaving it empty follows the repository's default branch, so a branch
renamed on the provider is followed rather than breaking the sync. It never
decides what is deployed - a run always reads the yaml at its own commit - which
is exactly why it can be pointed at a working branch.

A missing or invalid ` + "`.asgard-pipeline.yaml`" + ` does not fail the create; it is
reported on the pipeline as a failed config sync, and fixing the ref or the path
recovers it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if name == "" {
				return fmt.Errorf("--name is required; it is the pipeline's name in the workspace and has to be unique there")
			}
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			connID, err := resolveConnection(cmd.Context(), pc, connection)
			if err != nil {
				return err
			}
			repoID, repoName, err := resolveRepository(cmd.Context(), pc, connID, repo)
			if err != nil {
				return err
			}

			created, err := pc.Client.CreatePipeline(cmd.Context(), platform.CreatePipelineInput{
				Name:         name,
				ConnectionId: connID,
				RepositoryId: repoID,
				ConfigPath:   configPath,
				ConfigRef:    configRef,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, created)
			}
			fmt.Fprintf(out, "Created pipeline %s (%s) binding %s.\n", created.Name, created.PipelineId, repoName)
			printPipelineConfigState(out, created)

			// Record it beside the declaration, so the commands that follow do
			// not have to be told which pipeline - and so a repository that
			// later grows a second one does not become ambiguous.
			if path, err := recordPipeline(pc, created.PipelineId); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "\nCould not record the pipeline in %s: %v\n", binding.FileName, err)
			} else if path != "" {
				fmt.Fprintf(out, "\nRecorded it in %s. Commit that.\n", path)
			}

			// A ref that was asked for and did not come back means the gateway
			// dropped it, which is silent otherwise - the pipeline then reads
			// the default branch and nothing says so.
			if configRef != "" && created.ConfigRef != configRef {
				fmt.Fprintf(out, "\nWARNING: --config-ref %q was not recorded (the pipeline reports %q).\n"+
					"The platform gateway may predate the field; declarations will be read from the default branch.\n",
					configRef, created.ConfigRef)
			}
			return nil
		},
	}
	f.register(cmd, false)
	cmd.Flags().StringVar(&name, "name", "", "pipeline name, unique within the workspace (required)")
	cmd.Flags().StringVar(&connection, "connection", "", "connection id; defaults to the only one when the workspace has exactly one")
	cmd.Flags().StringVar(&repo, "repo", "", "repository id or \"owner/name\"; defaults to this checkout's origin remote")
	cmd.Flags().StringVar(&configPath, "config-path", "", "path of the declaration relative to the repository root; empty means .asgard-pipeline.yaml")
	cmd.Flags().StringVar(&configRef, "config-ref", "", "branch or tag to read the declaration from; empty follows the repository's default branch")
	return cmd
}

// resolveRepository turns what a person typed into the provider's repository
// id, and reports the full name so the confirmation names something readable.
func resolveRepository(ctx context.Context, pc *platformContext, connectionID, want string) (id, fullName string, err error) {
	if want == "" {
		want = pc.RepoFullName
	}
	if want == "" {
		return "", "", fmt.Errorf("no --repo, and this is not a git checkout with a recognisable origin remote")
	}

	repos, err := pc.Client.ListRepositories(ctx, connectionID)
	if err != nil {
		return "", "", err
	}
	for _, r := range repos {
		if r.RepositoryId == want || r.FullName == want {
			return r.RepositoryId, r.FullName, nil
		}
	}
	return "", "", fmt.Errorf("connection %s cannot reach %q; `asgard-cli pipeline repos` lists what it can, and adding one is a change to the App's repository selection on GitHub", connectionID, want)
}

func newPipelineListCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the workspace's pipelines",
		Long: `List the workspace's pipelines, with the repository each binds and the state of
its last config sync.

A failed sync is worth reading before anything else: a pipeline whose
declaration could not be read declares no releases at all, so its releases tab
looks empty for a reason that has nothing to do with what was created.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			pipelines, err := pc.Client.ListPipelines(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, pipelines)
			}
			if len(pipelines) == 0 {
				fmt.Fprintf(out, "No pipelines in workspace %s.\n", pc.Workspace)
				return nil
			}
			for _, p := range pipelines {
				sync := "never synced"
				if p.LastConfigSync != nil {
					if p.LastConfigSync.Ok {
						sync = "config ok"
					} else {
						sync = "config error"
					}
				}
				mark := " "
				if p.RepoFullName == pc.RepoFullName {
					mark = "*"
				}
				fmt.Fprintf(out, "%s %-22s %-20s %-52s %d release(s)  %s\n",
					mark, p.PipelineId, p.Name, p.RepoFullName, p.ReleaseCount, sync)
			}
			return nil
		},
	}
	f.register(cmd, false)
	return cmd
}

func newPipelineShowCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the pipeline bound to this checkout",
		Long: `Show the pipeline bound to this checkout: its repository, where its declarations
are read from, and what that declaration currently says.

The declarations shown here are the pipeline's snapshot, which is what draws the
ghost rows. A release that has already run shows the declarations of its own
most recent run instead, and ` + "`pipeline releases`" + ` says which.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			p, err := resolvePipeline(cmd.Context(), pc, f.pipeline)
			if err != nil {
				return err
			}
			// The list row carries no declarations; the single get does.
			full, err := pc.Client.GetPipeline(cmd.Context(), p.PipelineId)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, full)
			}
			fmt.Fprintf(out, "%-14s %s\n", "pipeline", full.Name)
			fmt.Fprintf(out, "%-14s %s\n", "id", full.PipelineId)
			fmt.Fprintf(out, "%-14s %s\n", "repository", full.RepoFullName)
			fmt.Fprintf(out, "%-14s %s\n", "connection", full.ConnectionId+"  ("+full.ConnectionStatus+")")
			printPipelineConfigState(out, full)

			if len(full.DeclaredReleases) == 0 {
				fmt.Fprintf(out, "\nThe declaration names no releases.\n")
				return nil
			}
			fmt.Fprintf(out, "\ndeclared releases\n")
			for _, r := range full.DeclaredReleases {
				trigger := "no trigger"
				if r.On != nil {
					trigger = r.On.Type + " " + r.On.Pattern
				}
				fmt.Fprintf(out, "  %-24s %-40s %s\n", r.Name, trigger, r.Chart)
			}
			return nil
		},
	}
	f.register(cmd, true)
	return cmd
}

// printPipelineConfigState prints where a pipeline reads its declaration from
// and how that read went.
//
// The ref is stated even when it is the default branch, because "which ref" is
// the question somebody has when the declared releases are not what they
// expected, and leaving it implicit means going to look it up.
func printPipelineConfigState(out interface{ Write([]byte) (int, error) }, p *platform.Pipeline) {
	ref := p.ConfigRef
	suffix := ""
	if ref == "" {
		ref = p.DefaultBranch
		suffix = "  (following the default branch)"
	}
	fmt.Fprintf(out, "%-14s %s%s\n", "declared from", ref, suffix)
	fmt.Fprintf(out, "%-14s %s\n", "config path", p.ConfigPath)

	switch {
	case p.LastConfigSync == nil:
		fmt.Fprintf(out, "%-14s never synced\n", "config")
	case p.LastConfigSync.Ok:
		short := p.LastConfigSync.CommitSha
		if len(short) > 7 {
			short = short[:7]
		}
		fmt.Fprintf(out, "%-14s ok at %s\n", "config", short)
	default:
		fmt.Fprintf(out, "%-14s ERROR %s\n", "config", p.LastConfigSync.Error)
	}
}

func newPipelineReleasesCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "releases",
		Short: "List a pipeline's releases, and the ones only declared",
		Long: `List a pipeline's releases.

Releases the declaration names but the platform has not created are listed
separately, because they behave differently: an event matching one of them
creates no run at all, and the delivery says "release not created". Creating it
is what makes the repository's declaration take effect.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			p, err := resolvePipeline(cmd.Context(), pc, f.pipeline)
			if err != nil {
				return err
			}
			full, err := pc.Client.GetPipeline(cmd.Context(), p.PipelineId)
			if err != nil {
				return err
			}
			releases, err := pc.Client.ListReleases(cmd.Context(), p.PipelineId)
			if err != nil {
				return err
			}

			created := map[string]bool{}
			for _, r := range releases {
				created[r.Name] = true
			}
			var ghosts []string
			for _, d := range full.DeclaredReleases {
				if !created[d.Name] {
					ghosts = append(ghosts, d.Name)
				}
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, map[string]any{
					"pipeline":             full.PipelineId,
					"releases":             releases,
					"declared_not_created": ghosts,
				})
			}

			if len(releases) == 0 {
				fmt.Fprintf(out, "No releases created under %s.\n", full.Name)
			}
			for _, r := range releases {
				last := "no runs"
				if r.LastRun != nil {
					last = fmt.Sprintf("#%d %s", r.LastRun.Number, r.LastRun.State)
				}
				declared := ""
				if r.Declaration == nil {
					declared = "  NOT DECLARED"
				}
				pending := ""
				if r.PendingDeploy {
					pending = "  (the platform holds values the cluster does not)"
				}
				fmt.Fprintf(out, "%-22s %-20s %-28s %-16s %s%s%s\n",
					r.ReleaseId, r.Name, r.Namespace, r.State, last, declared, pending)
			}
			if len(ghosts) > 0 {
				fmt.Fprintf(out, "\ndeclared but not created: %s\n", strings.Join(ghosts, ", "))
				fmt.Fprintf(out, "An event matching one of these creates no run until it is created:\n\n    asgard-cli pipeline release create <name> --project <id>\n\n`asgard-cli pipeline projects` lists the projects to bind one to.\n")
			}
			return nil
		},
	}
	f.register(cmd, true)
	return cmd
}

// resolveRelease finds a release of a pipeline by name.
//
// By name rather than by id because the name is what the repository's own
// `.asgard-pipeline.yaml` declares, so it is the only identifier an agent
// working in the checkout already has.
func resolveRelease(ctx context.Context, pc *platformContext, p *platform.Pipeline, name string) (*platform.Release, error) {
	releases, err := pc.Client.ListReleases(ctx, p.PipelineId)
	if err != nil {
		return nil, err
	}
	known := make([]string, 0, len(releases))
	for _, r := range releases {
		if r.Name == name || r.ReleaseId == name {
			return r, nil
		}
		known = append(known, r.Name)
	}
	return nil, &platform.ErrNoSuchRelease{Name: name, Pipeline: p.Name, Known: known}
}

// recordPipeline writes the pipeline id into the checkout's binding, keeping
// whatever else it holds. It reports the path written, or an empty one when
// there is no declaration to write beside - which is not a failure: a pipeline
// can legitimately be created from outside its repository.
func recordPipeline(pc *platformContext, pipelineID string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	declPath, bindPath, err := binding.Locate(dir)
	if err != nil || declPath == "" {
		return "", err
	}

	f := &binding.File{Version: 1, Workspace: pc.Workspace}
	if existing, err := binding.Load(bindPath); err == nil {
		f = existing
	}
	if f.Workspace == "" {
		f.Workspace = pc.Workspace
	}
	f.Pipeline = pipelineID
	if err := binding.Save(bindPath, f); err != nil {
		return "", err
	}
	if pc.RepoRoot != "" {
		if rel, err := filepath.Rel(pc.RepoRoot, bindPath); err == nil {
			return rel, nil
		}
	}
	return bindPath, nil
}

func newPipelineDeliveriesCmd() *cobra.Command {
	var (
		f     pipelineFlags
		limit int
	)

	cmd := &cobra.Command{
		Use:   "deliveries",
		Short: "Why a push did or did not produce a run",
		Long: `List the events this pipeline received and what came of each.

**This is the only place a push that produced nothing explains itself.** A run
that was never created leaves no record of its own, so when a tag appears to
have been ignored the reason is here and nowhere else:

    release "x" not created      the pattern matched, but nobody created it
    no release matches tag "x"   no pattern in the declaration matched the ref
    config unreadable / invalid  the declaration at that commit could not be used
    repository no longer bound   the pipeline was pointed at another repository
    connection revoked           the installation was removed on the provider

A delivery that did create runs names them, so this is also how one push that
matched several releases is seen as one event rather than as several unrelated
runs.

The same event reaches every pipeline bound to the repository, each reading its
own config path, so a monorepo's two pipelines each keep their own history of
it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			p, err := resolvePipeline(cmd.Context(), pc, f.pipeline)
			if err != nil {
				return err
			}
			deliveries, err := pc.Client.ListDeliveries(cmd.Context(), p.PipelineId, limit)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, deliveries)
			}
			if len(deliveries) == 0 {
				fmt.Fprintf(out, "No deliveries yet for %s. A push to %s is what produces one.\n",
					p.Name, p.RepoFullName)
				return nil
			}
			for _, d := range deliveries {
				when := ""
				if d.ReceivedAt != nil {
					when = d.ReceivedAt.Local().Format("01-02 15:04")
				}
				ref := d.Ref
				if ref == "" {
					ref = "-"
				}
				fmt.Fprintf(out, "%-12s %-10s %-28s %s\n", when, d.Event, ref, d.Outcome)
			}
			return nil
		},
	}
	f.register(cmd, true)
	cmd.Flags().IntVar(&limit, "limit", 20, "how many events to fetch")
	return cmd
}
