package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/binding"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

// Every subcommand here that changes something prints one line to stderr first,
// naming the Platform API it is about to change and how that profile came to be
// the one in effect. See actingOn: it is a receipt, and a receipt has no
// conditions - the line is there whether or not --profile was typed, because
// its absence would otherwise have to mean two different things.
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
    asgard-cli pipeline list                   the pipelines this workspace has
    asgard-cli pipeline use <id>               record which one this checkout uses
    asgard-cli pipeline show                   the pipeline this checkout records
    asgard-cli pipeline releases               its releases, and the ghost rows

These commands are a wrapper over the platform's API and hold no rules of their
own. **The checking runs on the platform**, because the checks worth the most -
the apiserver's own CEL, pattern and required validation of every rendered CR -
need a cluster, and no cluster credential is ever issued to a client. So the
loop is: change the chart, check what can be checked locally with ` + "`helm lint`" + `
and ` + "`asgard-cli verify`" + `, push, and read the plan back.

Which pipeline a command acts on is the one recorded in ` + "`.asgard-cli.yaml`" + `, which
` + "`asgard-cli pipeline use`" + ` writes and which is committed. **It is never derived.**
It used to be read off the origin remote whenever exactly one pipeline of the
workspace bound it, and that guessed twice: that a remote called ` + "`origin`" + ` is this
repository's identity - a repository may have several remotes, and which one
carries that name is nobody's business but its owner's - and that one candidate
means no choice had to be made. A command with nothing recorded now lists the
pipelines and stops. Which workspace is ` + "`asgard-cli workspace`" + `.`,
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
		newPipelineUseCmd(),
		newPipelineShowCmd(),
		newPipelineManifestCmd(),
		newPipelineProjectsCmd(),
		newPipelineProjectCmd(),
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
		"workspace to act in; defaults to what \"asgard-cli workspace use\" recorded for this repository")
	cmd.Flags().StringVar(&f.format, formatFlag, formatText, formatUsage)
	if withPipeline {
		cmd.Flags().StringVar(&f.pipeline, "pipeline", "",
			"pipeline id or name; defaults to the one recorded in "+binding.FileName)
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
// There are exactly two answers: the one named on the command line, and the one
// recorded in the checkout's binding. **There is no third**, and there was - a
// pipeline of the workspace whose repository matched the checkout's origin
// remote was used when exactly one did. That is gone, for two separate reasons
// that happened to sit in the same expression:
//
//   - `origin` is not an identity. A checkout may have any number of remotes,
//     and whether the one somebody named `origin` is the repository a pipeline
//     binds is a convention this tool has no standing to enforce. Somebody
//     whose main remote is `upstream` was refused; somebody with no `origin`
//     at all skipped the check entirely.
//   - one candidate is not a decision. A rule that resolves while a workspace
//     holds one matching pipeline stops resolving - or worse, resolves to
//     something else - on the day it holds two.
//
// With nothing recorded this fails and lists what there is. That failure is the
// point: it lands on the first command run, rather than on whichever later one
// happened to be destructive.
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
	return nil, needPipelineError(pc, pipelines)
}

// needPipelineError is the "which of these?" for a checkout that records no
// pipeline, including the half-bound state `workspace use` leaves behind.
func needPipelineError(pc *platformContext, pipelines []*platform.Pipeline) error {
	var b strings.Builder
	where := "this checkout"
	if pc.Binding != nil {
		where = pc.Binding.Path
	}

	if len(pipelines) == 0 {
		fmt.Fprintf(&b, "%s records no pipeline, and workspace %s has none.\n\n", where, pc.Workspace)
		fmt.Fprintf(&b, "    asgard-cli pipeline create --name <name> --connection <id> --repo <owner/name>\n")
		return fmt.Errorf("%s", b.String())
	}

	fmt.Fprintf(&b, "%s records no pipeline, and workspace %s has %s:\n",
		where, pc.Workspace, plural(len(pipelines), "pipeline"))
	for _, p := range pipelines {
		fmt.Fprintf(&b, "  %-22s %-20s %-52s %s\n", p.PipelineId, p.Name, p.RepoFullName, p.ConfigPath)
	}
	fmt.Fprintf(&b, "\nRecord one, which is committed so nobody has to choose again:\n\n")
	fmt.Fprintf(&b, "    asgard-cli pipeline use <id>\n")
	fmt.Fprintf(&b, "\nOr name one for this run with --pipeline. None is assumed, and that\nincludes a list of one.\n")
	return fmt.Errorf("%s", b.String())
}

// pipelineFromBinding uses the pipeline the checkout's `.asgard-cli.yaml`
// records.
//
// **It checks that the pipeline exists in this workspace, and nothing else.**
// It used to also compare the pipeline's repository against the checkout's
// origin remote, to catch a binding copied wholesale into another repository -
// and that check is gone, because it made a correct setup fail or pass
// depending on what somebody named their remotes.
//
// The gap that leaves is real and worth naming: a repository copied into
// another repository of the SAME workspace keeps a pipeline id that still
// resolves, and nothing says so. Copied into a different workspace it fails
// here, because the pipeline is not in the list. That gap is written into the
// file's own header, where somebody copying a repository will read it, rather
// than defended by a rule that guesses at remote names.
func pipelineFromBinding(pc *platformContext, pipelines []*platform.Pipeline) (*platform.Pipeline, error) {
	want := pc.Binding.Pipeline
	for _, p := range pipelines {
		if p.PipelineId == want {
			return p, nil
		}
	}
	return nil, fmt.Errorf(
		"%s records pipeline %s, which does not exist in workspace %s.\n"+
			"Either the wrong workspace is in effect (`asgard-cli workspace show` says which and why),\n"+
			"the repository was copied from somewhere else, or the pipeline was deleted.\n"+
			"`asgard-cli pipeline list` shows what is there, and `asgard-cli pipeline use <id>` records one.",
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

--connection is required. A workspace with one connection today has two the day
somebody connects a second GitHub organisation, and a default that quietly
stopped applying is worse than one that never did. ` + "`asgard-cli pipeline connections`" + `
lists them, and so does the error.`,
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
	cmd.Flags().StringVar(&connection, "connection", "", "connection id (required); \"asgard-cli pipeline connections\" lists them")
	return cmd
}

// resolveConnection returns the connection named on the command line, or an
// error listing the candidates.
//
// **It does not default to the only one.** That is not a judgement about how
// likely a second connection is; it is that the day a second appears, every
// command which had been resolving silently starts resolving to something a
// person never chose - and a connection decides which repositories a pipeline
// can bind at all.
func resolveConnection(ctx context.Context, pc *platformContext, want string) (string, error) {
	if want != "" {
		return want, nil
	}
	conns, err := pc.Client.ListConnections(ctx)
	if err != nil {
		return "", err
	}
	if len(conns) == 0 {
		return "", fmt.Errorf("workspace %s has no VCS connection; `asgard-cli pipeline connect` creates one", pc.Workspace)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "--connection is required, and workspace %s has %s:\n",
		pc.Workspace, plural(len(conns), "connection"))
	for _, c := range conns {
		fmt.Fprintf(&b, "  %-22s %-24s %s\n", c.ConnectionId, c.AccountLogin, c.Status)
	}
	fmt.Fprintf(&b, "\nNone is assumed, and that includes a list of one.\n")
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
which is resolved through the connection. Without it, this checkout's origin
remote is used and the command says so on stderr before creating anything -
which is the one place a remote name is read, at the moment a person is naming
a new thing and reads the confirmation. Nothing afterwards looks at a remote.

    asgard-cli pipeline create --name iac-test --connection <id>
    asgard-cli pipeline create --name iac-test --connection <id> --repo asgard-ai-platform/asgard-iac-test

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
			actingOn(cmd, pc.Session)
			connID, err := resolveConnection(cmd.Context(), pc, connection)
			if err != nil {
				return err
			}
			// Falling back to the origin remote is a convenience at the moment
			// a pipeline is being created, not a claim that `origin` is this
			// repository's identity - so it is said out loud rather than
			// applied silently, and the confirmation below names what was
			// bound either way.
			if repo == "" && pc.RepoFullName != "" {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"no --repo, so this checkout's origin remote is used: %s\n", pc.RepoFullName)
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
	cmd.Flags().StringVar(&connection, "connection", "", "connection id (required); \"asgard-cli pipeline connections\" lists them")
	cmd.Flags().StringVar(&repo, "repo", "", "repository id or \"owner/name\"; without it, this checkout's origin remote is used and named in the output")
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

The starred row is the one this checkout records in ` + "`.asgard-cli.yaml`" + `, so a wrong
binding is visible here without running anything that acts on it.

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
				// Marked by what this checkout records, not by what its
				// origin remote happens to be. The star used to mean "this
				// pipeline's repository matches your origin", which is a
				// different claim - and one that could star a pipeline no
				// command here would ever act on.
				mark := " "
				if pc.Binding != nil && pc.Binding.Pipeline == p.PipelineId {
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

func newPipelineUseCmd() *cobra.Command {
	var f pipelineFlags

	cmd := &cobra.Command{
		Use:   "use <pipeline-id>",
		Short: "Record which pipeline this checkout deploys through",
		Long: `Record which pipeline this checkout deploys through.

    asgard-cli pipeline list        what the workspace has
    asgard-cli pipeline use 1862431170889781248

It writes the ` + "`pipeline`" + ` line of ` + "`.asgard-cli.yaml`" + `, beside the declaration it
belongs to, and that file is committed: whoever clones this repository, and
whatever agent works in it, then needs no --pipeline.

**It is checked against the platform and against nothing else.** The pipeline
has to exist in the workspace currently in effect - that is one call, and a
typo is worth catching here rather than three commands later. What is NOT
checked is the repository: this never looks at a git remote. A checkout may
have several, and which one is called ` + "`origin`" + ` says nothing about what it
deploys, so a rule built on that name refuses correct setups and passes broken
ones depending on how somebody happened to name things.

Run it after copying a repository. A copy carries the previous repository's
pipeline id, and if the copy lives in the same workspace that id still resolves
- which is the one silent way this file can be wrong, and is written into its
own header for the same reason.

Which workspace a pipeline belongs to is ` + "`asgard-cli workspace use`" + `, and changing
that clears this line, because a pipeline of one workspace means nothing in
another.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			want := args[0]
			actingLocally(cmd, f.profile, binding.FileName)
			pc, err := f.context(cmd)
			if err != nil {
				return err
			}

			pipelines, err := pc.Client.ListPipelines(cmd.Context())
			if err != nil {
				return err
			}
			var chosen *platform.Pipeline
			for _, p := range pipelines {
				if p.PipelineId == want || p.Name == want {
					chosen = p
					break
				}
			}
			if chosen == nil {
				var b strings.Builder
				fmt.Fprintf(&b, "no pipeline %q in workspace %s", want, pc.Workspace)
				if len(pipelines) == 0 {
					fmt.Fprintf(&b, ", which has none.\n\n    asgard-cli pipeline create --name <name> --connection <id> --repo <owner/name>\n")
					return fmt.Errorf("%s", b.String())
				}
				fmt.Fprintf(&b, ". It has %s:\n", plural(len(pipelines), "pipeline"))
				for _, p := range pipelines {
					fmt.Fprintf(&b, "  %-22s %-20s %-52s %s\n", p.PipelineId, p.Name, p.RepoFullName, p.ConfigPath)
				}
				return fmt.Errorf("%s", b.String())
			}

			// Written by id even when a name was typed: a name is the
			// workspace's to change, and a file that survives a rename is
			// worth more than one that reads slightly better.
			rel, err := recordPipeline(pc, chosen.PipelineId)
			if err != nil {
				return err
			}
			if rel == "" {
				return fmt.Errorf("no %s at or above this directory, so there is nothing for a binding to belong to.\n"+
					"A pipeline is recorded beside the declaration it deploys; `asgard-cli scaffold` writes one.",
					pipelineconfig.FileName)
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, map[string]any{
					"path":      rel,
					"workspace": pc.Workspace,
					"pipeline":  chosen.PipelineId,
					"name":      chosen.Name,
					"repo":      chosen.RepoFullName,
				})
			}
			fmt.Fprintf(out, "Recorded pipeline %s (%s) in %s.\n", chosen.PipelineId, chosen.Name, rel)
			fmt.Fprintf(out, "%-14s %s\n", "repository", chosen.RepoFullName)
			fmt.Fprintf(out, "%-14s %s\n", "config path", chosen.ConfigPath)
			fmt.Fprintf(out, "\nCommit it, together with the workspace line beside it.\n")
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
// whatever else it holds, and filling in the workspace when the file did not
// have one - a binding with a pipeline and no workspace is not a state worth
// being able to produce.
//
// It reports the path written, or an empty one when there is no declaration to
// write beside. For `pipeline create` that is not a failure - a pipeline can
// legitimately be created from outside its repository - and `pipeline use`
// turns it into one, because recording a choice that lands nowhere is not
// recording it.
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
