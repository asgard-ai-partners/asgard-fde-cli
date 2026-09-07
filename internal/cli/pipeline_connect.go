package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/browser"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
)

// connectPollInterval is how often the workspace is asked whether the new
// connection has arrived.
//
// The wait is a person clicking through a provider's own pages, so this is
// paced for a person rather than for a machine: often enough that the terminal
// reacts as soon as they are done, rarely enough that a five-minute wait is
// under a hundred calls.
const connectPollInterval = 3 * time.Second

// connectTimeout bounds the whole wait. It matches the sign-in flow's, and for
// the same reason: a command that never returns leaves a terminal nobody can
// tell apart from a hung one.
const connectTimeout = 5 * time.Minute

func newPipelineConnectCmd() *cobra.Command {
	var (
		f         pipelineFlags
		noBrowser bool
		account   string
		wait      time.Duration
	)

	cmd := &cobra.Command{
		Use:   "connect [provider]",
		Short: "Connect a version control provider to this workspace",
		Long: `Connect a version control provider to this workspace, so that pipelines can bind
its repositories.

    asgard-cli pipeline connect            connect GitHub
    asgard-cli pipeline connect github     the same, said explicitly

The provider is an argument rather than part of the command name because more
are expected: a connection records which provider it is, and adding one later
should be a new value here, not a new command to learn. Today github is the only
one, and naming another says so rather than pretending.

WHAT HAPPENS. This identifies you on the provider first, and only then works out
what to connect. That order is deliberate: whether the app is already installed
on the account is a fact the caller may not have — somebody else, or another
workspace, may have installed it months ago — and the two ways in fail very
differently if you guess wrong. So it is never asked.

    1. the provider says who you are
    2. it lists the installations you can reach
    3. the one matching --account is connected

Nothing prompts. --account defaults to the owner of this checkout's origin
remote, which is the account whose repositories this engagement is about; name
it explicitly when connecting some other account, or when there is no remote to
read. When the account you asked for is not among the ones you reach, the app
has to be installed on it first, and the installation URL is printed rather than
guessed at.

The state is sealed rather than stored, so it stays usable for its whole
lifetime rather than being spent on first use. The provider redirects back to
the platform - not to this machine - which is what completes the connection.
What this waits on is therefore two things at once: a connection appearing, and
the platform reporting how the attempt ended. Both are needed: the callbacks
never reach this process, so from here "nothing yet" and "it failed a minute
ago" look identical.

    asgard-cli pipeline connect --no-browser

prints the URL instead of opening one, for a session where the browser is
somewhere else. The URL is printed either way.

ONE ACCOUNT, AS MANY WORKSPACES AS NEED IT. GitHub issues one installation per
account, so a rule that one installation belongs to one workspace would have
meant a GitHub organisation could serve one workspace. Each workspace holds its
own connection to the same installation, and they do not see each other's.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := platform.ProviderGitHub
			if len(args) == 1 {
				provider = strings.ToLower(args[0])
			}
			if provider != platform.ProviderGitHub {
				return fmt.Errorf("no provider %q; this platform connects %s",
					provider, strings.Join(platform.Providers(), ", "))
			}

			pc, err := f.context(cmd)
			if err != nil {
				return err
			}
			actingOn(cmd, pc.Session)
			ctx := cmd.Context()

			// Snapshot first. The new connection is identified by not having
			// been there, which needs no cooperation from the provider's flow
			// and works the same for the next provider.
			before, err := connectionIDs(ctx, pc)
			if err != nil {
				return err
			}

			wanted := account
			if wanted == "" {
				// The owner of this checkout's remote is the account this
				// engagement is about. Derived rather than asked, because
				// asking would need somebody at the terminal and this command
				// is run by an agent.
				wanted, _, _ = strings.Cut(pc.RepoFullName, "/")
			}

			// Identify first, always. Which way in is right depends on
			// whether the app is already installed on the account, and that is
			// a fact the caller may not have.
			install, err := pc.Client.BeginGitHubAttach(ctx)
			if err != nil {
				return err
			}

			// Progress on stderr keeps a --format json run's stdout to the
			// answer alone.
			msg := cmd.ErrOrStderr()
			fmt.Fprintf(msg, "Connecting %s to workspace %s.\n\n", provider, pc.Workspace)
			if !noBrowser {
				if err := browser.Open(install.InstallUrl); err != nil {
					fmt.Fprintf(msg, "Could not open a browser (%v).\n", err)
				}
			}
			fmt.Fprintf(msg, "Open this URL and authorize, so the provider can say what you can connect:\n\n    %s\n\n", install.InstallUrl)
			if install.ExpiresAt != nil {
				fmt.Fprintf(msg, "The link expires %s.\n", install.ExpiresAt.Local().Format("15:04"))
			}
			fmt.Fprintf(msg, "Waiting for the connection to appear...\n")

			if wait <= 0 {
				wait = connectTimeout
			}
			created, choices, insufficient, err := waitForConnection(ctx, pc, before, install.State, wait)
			if err != nil {
				return err
			}
			switch {
			case choices != nil:
				created, err = attachNamedAccount(cmd, pc, choices, wanted)
			case insufficient:
				// Authorized, and reaches nothing here. Not a failure to report
				// — it is the answer to "which way in", arrived at without
				// asking anybody to know it in advance.
				fmt.Fprintf(msg, "\nYou reach no installation of this app yet, so it has to be installed first.\n")
				created, err = installFirst(cmd, pc, before, noBrowser, wait)
			}
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if f.format == formatJSON {
				return writeJSON(out, created)
			}
			fmt.Fprintf(out, "Connected %s (%s), installation %s, connection %s.\n",
				created.AccountLogin, created.AccountType, created.InstallationId, created.ConnectionId)
			fmt.Fprintf(out, "\n`asgard-cli pipeline repos` lists what it can reach; a repository missing from\nthat list is one the installation was not granted, which is changed on the\nprovider rather than here.\n")
			return nil
		},
	}

	f.register(cmd, false)
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false,
		"do not open a browser; the URL is printed either way")
	cmd.Flags().StringVar(&account, "account", "",
		"which provider account to connect; defaults to the owner of this checkout's origin remote")
	cmd.Flags().DurationVar(&wait, "wait", connectTimeout,
		"how long to wait for the connection to appear before giving up")
	return cmd
}

// connectionIDs is the set of connection ids the workspace has now.
func connectionIDs(ctx context.Context, pc *platformContext) (map[string]bool, error) {
	conns, err := pc.Client.ListConnections(ctx)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]bool, len(conns))
	for _, c := range conns {
		ids[c.ConnectionId] = true
	}
	return ids, nil
}

// waitForConnection ends as soon as EITHER signal lands: a connection that was
// not there before, or the platform reporting how this flow ended.
//
// Watching only for the connection is what this used to do, and it is why every
// failure took the full timeout and then produced a guess. The provider's
// callbacks never reach this process, so "nothing yet" and "it failed four
// minutes ago" look identical from here — the status is the only thing that
// separates them.
//
// A transient failure while polling is not fatal: the installation may well
// have succeeded, and giving up on one bad response would report a failure that
// did not happen. Only the deadline, or a settled flow, ends the wait.
func waitForConnection(
	ctx context.Context,
	pc *platformContext,
	before map[string]bool,
	state string,
	wait time.Duration,
) (*platform.VcsConnection, *platform.AttachChoices, bool, error) {
	deadline := time.Now().Add(wait)
	ticker := time.NewTicker(connectPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			return nil, nil, false, ctx.Err()
		case <-ticker.C:
		}

		conns, err := pc.Client.ListConnections(ctx)
		if err != nil {
			lastErr = err
		} else {
			lastErr = nil
			for _, c := range conns {
				if !before[c.ConnectionId] {
					return c, nil, false, nil
				}
			}
		}

		// Asked second, so a connection that exists is reported as success
		// even if the status write lost a race with it.
		if status, serr := pc.Client.GetInstallStatus(ctx, state); serr == nil {
			if status.AttachState != "" {
				// The attach path ended in a choice, not a connection.
				return nil, &platform.AttachChoices{
					Installations: status.Installations,
					AttachState:   status.AttachState,
				}, false, nil
			}
			if status.Settled() {
				if status.ConnectionId != "" {
					// It connected, and the listing has not caught up. Fetch it
					// rather than reporting a bare id.
					if c, gerr := pc.Client.GetConnection(ctx, status.ConnectionId); gerr == nil {
						return c, nil, false, nil
					}
				}
				if status.Outcome == "authorization_insufficient" {
					// Reported rather than raised: it is not a failure of this
					// attempt, it is which way in the person needs.
					return nil, nil, true, nil
				}
				return nil, nil, false, fmt.Errorf("%s", strings.TrimSpace(status.Message))
			}
		}

		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, nil, false, fmt.Errorf("gave up after %s, and the last check failed: %w", wait, lastErr)
			}
			return nil, nil, false, fmt.Errorf(
				"gave up after %s: workspace %s reports no outcome for this installation yet.\n"+
					"The provider page may still be open, or it was closed without finishing.\n"+
					"`asgard-cli pipeline connections` shows what this workspace has.", wait, pc.Workspace)
		}
	}
}

// attachNamedAccount connects the installation whose account was named, and
// says what it saw when it cannot.
//
// Nothing prompts. This command is run by an agent following a skill, not by
// somebody at a terminal — `init` is the one command in this tool with a person
// in front of it — so a question here is not a slower path, it is a dead one.
// The caller declares which account it means and the platform matches it.
func attachNamedAccount(
	cmd *cobra.Command,
	pc *platformContext,
	choices *platform.AttachChoices,
	account string,
) (*platform.VcsConnection, error) {
	msg := cmd.ErrOrStderr()

	fmt.Fprintf(msg, "\nYou can reach:\n\n")
	for _, in := range choices.Installations {
		held := ""
		if in.ConnectionId != "" {
			held = "  (already connected here)"
		}
		fmt.Fprintf(msg, "  %s (%s) - %s%s\n", in.AccountLogin, in.AccountType, repositoryScope(in.RepositorySelection), held)
	}

	if account == "" {
		return nil, fmt.Errorf(
			"no --account, and this is not a git checkout with a recognisable origin remote.\n" +
				"Name the account to connect with --account; the ones you reach are listed above")
	}
	for _, in := range choices.Installations {
		if strings.EqualFold(in.AccountLogin, account) {
			fmt.Fprintf(msg, "\nConnecting %s...\n", in.AccountLogin)
			return pc.Client.AttachInstallation(cmd.Context(), choices.AttachState, in.InstallationId)
		}
	}
	// Says what was established — the account is not among the ones reached —
	// and what would change that, without asserting which of the two reasons it
	// is. Both are real: the app may not be installed there, or this identity
	// may not reach it.
	return nil, fmt.Errorf(
		"%q is not among the accounts you reach on the provider (listed above).\n"+
			"Either the app is not installed on it, or the account you authorized as cannot see its repositories.\n"+
			"`asgard-cli pipeline connect --account <one of the above>` connects one of them",
		account)
}

// installFirst prints where to install the app, for an account that has no
// installation yet. Installing is a person's action on the provider; there is
// nothing this can do but say where and wait.
func installFirst(
	cmd *cobra.Command,
	pc *platformContext,
	before map[string]bool,
	noBrowser bool,
	wait time.Duration,
) (*platform.VcsConnection, error) {
	ctx := cmd.Context()
	msg := cmd.ErrOrStderr()

	install, err := pc.Client.BeginGitHubInstall(ctx)
	if err != nil {
		return nil, err
	}
	if !noBrowser {
		if berr := browser.Open(install.InstallUrl); berr != nil {
			fmt.Fprintf(msg, "Could not open a browser (%v).\n", berr)
		}
	}
	fmt.Fprintf(msg, "\nOpen this URL and install the app on the account you want:\n\n    %s\n\n", install.InstallUrl)
	fmt.Fprintf(msg, "Waiting...\n")

	created, choices, _, err := waitForConnection(ctx, pc, before, install.State, wait)
	if err != nil {
		return nil, err
	}
	if choices != nil {
		// The provider sent them through authorization on the way. Whatever
		// they installed is the one to connect, so there is still nothing to
		// ask: a fresh installation is the only one that was not reachable a
		// moment ago.
		return attachFreshest(cmd, pc, choices, before)
	}
	return created, nil
}

// attachFreshest connects the installation this workspace does not already
// hold, which after an install is the one just made.
func attachFreshest(cmd *cobra.Command, pc *platformContext, choices *platform.AttachChoices, _ map[string]bool) (*platform.VcsConnection, error) {
	for _, in := range choices.Installations {
		if in.ConnectionId == "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "\nConnecting %s...\n", in.AccountLogin)
			return pc.Client.AttachInstallation(cmd.Context(), choices.AttachState, in.InstallationId)
		}
	}
	return nil, fmt.Errorf("nothing new was installed; every account you reach is already connected here")
}
