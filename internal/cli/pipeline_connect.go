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

WHAT HAPPENS. The platform mints a single-use state bound to this workspace and
returns the provider's installation URL carrying it. This opens that URL, and
the provider redirects back to the platform - not to this machine - which is
what completes the connection. So there is nothing here to catch: what this
waits on is a new connection appearing in the workspace, which is the same
thing to watch whatever the provider is.

    asgard-cli pipeline connect --no-browser

prints the URL instead of opening one, for a session where the browser is
somewhere else. The URL is printed either way.

ONE INSTALLATION, ONE WORKSPACE. An installation already held by another
workspace is refused, naming the one that holds it - sharing an installation
would let either side deploy the other's repositories. Releasing it means
deleting the connection that holds it.

An installation that already exists on the provider still has to be connected
here once: existing on GitHub and being bound to a workspace are different
things, and the provider's page for an already-installed App is a "configure"
screen rather than an "install" one. Clicking through it completes the same
flow.`,
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
			ctx := cmd.Context()

			// Snapshot first. The new connection is identified by not having
			// been there, which needs no cooperation from the provider's flow
			// and works the same for the next provider.
			before, err := connectionIDs(ctx, pc)
			if err != nil {
				return err
			}

			install, err := pc.Client.BeginGitHubInstall(ctx)
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
			fmt.Fprintf(msg, "Open this URL and complete the installation:\n\n    %s\n\n", install.InstallUrl)
			if install.ExpiresAt != nil {
				fmt.Fprintf(msg, "The link is single-use and expires %s.\n", install.ExpiresAt.Local().Format("15:04"))
			}
			fmt.Fprintf(msg, "Waiting for the connection to appear...\n")

			if wait <= 0 {
				wait = connectTimeout
			}
			created, err := waitForConnection(ctx, pc, before, wait)
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
		"do not open a browser; the installation URL is printed either way")
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

// waitForConnection polls until a connection that was not in before appears.
//
// A transient failure while polling is not fatal: the installation may well
// have succeeded, and giving up on one bad response would report a failure that
// did not happen. Only the deadline ends the wait.
func waitForConnection(
	ctx context.Context,
	pc *platformContext,
	before map[string]bool,
	wait time.Duration,
) (*platform.VcsConnection, error) {
	deadline := time.Now().Add(wait)
	ticker := time.NewTicker(connectPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}

		conns, err := pc.Client.ListConnections(ctx)
		if err != nil {
			lastErr = err
		} else {
			lastErr = nil
			for _, c := range conns {
				if !before[c.ConnectionId] {
					return c, nil
				}
			}
		}

		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, fmt.Errorf("gave up after %s, and the last check failed: %w", wait, lastErr)
			}
			return nil, fmt.Errorf(
				"gave up after %s: no new connection appeared in workspace %s.\n"+
					"The installation may not have been completed, or it may already be held by another\n"+
					"workspace - one installation binds one workspace, and the platform refuses a second.\n"+
					"`asgard-cli pipeline connections` shows what this workspace has.", wait, pc.Workspace)
		}
	}
}
