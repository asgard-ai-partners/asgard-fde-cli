package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/gitrepo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

// workspaceFlag is the name of the flag that names a workspace.
const workspaceFlag = "workspace"

// WorkspaceSource says where a resolved workspace id came from, so that a
// command can print it. It is not decoration: a command that acted in the wrong
// workspace is the failure this whole resolution order exists to make visible,
// and "which workspace, and why that one" is one line.
type WorkspaceSource string

const (
	fromFlag       WorkspaceSource = "--workspace"
	fromEnv        WorkspaceSource = auth.EnvWorkspace
	fromRepoConfig WorkspaceSource = config.FileName
	fromBinding    WorkspaceSource = "recorded for this repository"
	fromFallback   WorkspaceSource = "recorded as the default"
	fromOnlyOne    WorkspaceSource = "the only workspace you can reach"
)

// platformContext is what every command that talks to the platform needs: a
// session, the workspace to act in, and - when there is one - the repository
// the command was run inside.
type platformContext struct {
	Session   *auth.Session
	Client    *platform.Client
	Workspace string
	// WorkspaceSource is why that workspace, for the line a command prints.
	WorkspaceSource WorkspaceSource
	// RepoFullName is the origin remote as "owner/name"; empty outside a
	// checkout, or when the remote is a shape this does not recognise.
	RepoFullName string
	// RepoRoot is the checkout's top level; empty outside one.
	RepoRoot string
}

// contextOptions are the per-command overrides of the resolution.
type contextOptions struct {
	Profile   string
	Workspace string
	// NeedWorkspace is false for the handful of calls that take none - listing
	// workspaces being the one that has to work before any is chosen.
	NeedWorkspace bool
}

// resolveContext produces everything a platform command needs, or an error that
// says what to do about it.
//
// The workspace is looked for in this order, and the first answer wins:
//
//  1. --workspace
//  2. ASGARD_WORKSPACE
//  3. .asgard-config.json's workspace.id, when the command is inside a
//     scaffolded repository - it is committed, so a teammate inherits it
//  4. what `workspace use` recorded for this repository under this profile
//  5. what `workspace use --default` recorded for this profile
//  6. the only workspace the session can reach, when there is exactly one
//
// Six is deliberately last and deliberately present: a customer with one
// workspace should never have to name it, and anybody with two must.
func resolveContext(cmd *cobra.Command, opts contextOptions) (*platformContext, error) {
	ctx := cmd.Context()

	session, err := auth.Resolve(ctx, opts.Profile)
	if err != nil {
		return nil, err
	}

	pc := &platformContext{Session: session}
	pc.RepoRoot, pc.RepoFullName = locateRepo(ctx)

	if !opts.NeedWorkspace {
		pc.Client = platform.New(session, "")
		return pc, nil
	}

	ws, source, err := resolveWorkspace(ctx, session, pc.RepoFullName, pc.RepoRoot, opts.Workspace)
	if err != nil {
		return nil, err
	}
	pc.Workspace, pc.WorkspaceSource = ws, source
	pc.Client = platform.New(session, ws)
	return pc, nil
}

// locateRepo reports the checkout the command was run in, or two empty strings.
// Not being in one is normal - half this tool answers questions in a meeting -
// so nothing here is an error.
func locateRepo(ctx context.Context) (root, fullName string) {
	dir, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	root, err = gitrepo.Root(ctx, dir)
	if err != nil {
		return "", ""
	}
	fullName, err = gitrepo.OriginFullName(ctx, root)
	if err != nil {
		return root, ""
	}
	return root, fullName
}

func resolveWorkspace(
	ctx context.Context,
	session *auth.Session,
	repoFullName, repoRoot, flag string,
) (string, WorkspaceSource, error) {
	if flag != "" {
		return flag, fromFlag, nil
	}
	if env := os.Getenv(auth.EnvWorkspace); env != "" {
		return env, fromEnv, nil
	}

	if repoRoot != "" {
		if id, ok := workspaceFromRepoConfig(repoRoot); ok {
			return id, fromRepoConfig, nil
		}
	}

	settings, err := auth.LoadSettings()
	if err != nil {
		return "", "", err
	}
	if repoFullName != "" {
		if id, ok := settings.WorkspaceFor(session.Profile.Name, repoFullName); ok {
			return id, fromBinding, nil
		}
	}
	if id, ok := settings.FallbackWorkspace(session.Profile.Name); ok {
		return id, fromFallback, nil
	}

	// Nothing recorded. One workspace needs no choice; more than one does, and
	// the error lists them rather than making somebody go and look.
	workspaces, err := platform.New(session, "").ListWorkspaces(ctx)
	if err != nil {
		return "", "", err
	}
	switch len(workspaces) {
	case 0:
		return "", "", fmt.Errorf("the %s platform reports no workspaces for this account", session.Profile.Name)
	case 1:
		return workspaces[0].ID, fromOnlyOne, nil
	}
	return "", "", &needWorkspaceError{Profile: session.Profile.Name, Repo: repoFullName, Workspaces: workspaces}
}

// workspaceFromRepoConfig reads a scaffolded repository's recorded workspace.
func workspaceFromRepoConfig(root string) (string, bool) {
	path, err := config.Find(root)
	if err != nil {
		return "", false
	}
	cfg, err := config.Load(path)
	if err != nil || !cfg.Workspace.HasID() {
		return "", false
	}
	return cfg.Workspace.ID, true
}

// needWorkspaceError is the "which of these?" that a first run in a second
// workspace produces.
type needWorkspaceError struct {
	Profile    string
	Repo       string
	Workspaces []platform.Workspace
}

func (e *needWorkspaceError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "no workspace chosen for the %s platform, and this account can reach %d:\n", e.Profile, len(e.Workspaces))
	for _, w := range e.Workspaces {
		fmt.Fprintf(&b, "  %-22s %s\n", w.ID, w.Name)
	}
	b.WriteString("\nRecord one and every later command in this repository uses it:\n\n")
	if e.Repo != "" {
		fmt.Fprintf(&b, "    asgard-cli workspace use <id>%s\n", profileArgFor(e.Profile))
	} else {
		fmt.Fprintf(&b, "    asgard-cli workspace use <id> --default%s\n", profileArgFor(e.Profile))
		b.WriteString("\n(--default because this is not a git repository, so there is nothing to bind it to.)\n")
	}
	return b.String()
}

// profileArgFor is profileArg for a bare profile name.
func profileArgFor(name string) string {
	if name == auth.DefaultProfileName {
		return ""
	}
	return " --profile " + name
}

// errNoRepoToBind is what `workspace use` reports outside a checkout, where
// there is no repository to attach a binding to.
var errNoRepoToBind = errors.New(
	"not in a git repository with an origin remote, so there is nothing to bind a workspace to; " +
		"pass --default to record it for every command run outside a repository instead")
