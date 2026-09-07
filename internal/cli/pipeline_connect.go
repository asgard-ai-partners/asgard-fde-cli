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
		existing  bool
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

WHAT HAPPENS. The platform mints a state bound to this workspace and returns the
provider's installation URL carrying it. The state is sealed rather than stored,
so it stays usable for its whole lifetime rather than being spent on first use.
This opens that URL, and the provider redirects back to the platform - not to
this machine - which is what completes the connection. So there is nothing here
to catch.

What this waits on is therefore two things at once: a new connection appearing,
and the platform reporting how the attempt ended. Both are needed. The callbacks
never reach this process, so from here "nothing yet" and "it failed a minute ago"
look identical - which is why a refused installation used to sit until the
timeout and then produce a guess. It now says what happened, and why, as soon as
the platform knows.

    asgard-cli pipeline connect --no-browser

prints the URL instead of opening one, for a session where the browser is
somewhere else. The URL is printed either way.

ONE ACCOUNT, AS MANY WORKSPACES AS NEED IT. GitHub issues one installation per
account, so a rule that one installation belongs to one workspace would have
meant a GitHub organisation could serve one workspace. Each workspace holds its
own connection to the same installation, and they do not see each other's.

An installation that already exists on the provider still has to be connected
here once: existing on GitHub and being bound to a workspace are different
things, and the provider's page for an already-installed App is a "configure"
screen rather than an "install" one.

That screen has a catch worth knowing before you meet it: its save button is
disabled while there is nothing to save, which is exactly the case when the
repository access you want is already granted. There is then no button to click
and no redirect back here.

    asgard-cli pipeline connect --existing

takes the other way in for that case: it identifies you on the provider first,
then offers the installations it says you can reach, and connects the one you
pick. Nothing on the provider is changed, which matters because the repository
access on an installation is now shared by every workspace connected to it.`,
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

			begin := pc.Client.BeginGitHubInstall
			if existing {
				begin = pc.Client.BeginGitHubAttach
			}
			install, err := begin(ctx)
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
			what := "complete the installation"
			if existing {
				what = "authorize, so the provider can say which installations you reach"
			}
			fmt.Fprintf(msg, "Open this URL and %s:\n\n    %s\n\n", what, install.InstallUrl)
			if install.ExpiresAt != nil {
				fmt.Fprintf(msg, "The link expires %s.\n", install.ExpiresAt.Local().Format("15:04"))
			}
			fmt.Fprintf(msg, "Waiting for the connection to appear...\n")

			if wait <= 0 {
				wait = connectTimeout
			}
			created, choices, err := waitForConnection(ctx, pc, before, install.State, wait)
			if err != nil {
				return err
			}
			if choices != nil {
				created, err = pickAndAttach(cmd, pc, choices)
				if err != nil {
					return err
				}
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
	cmd.Flags().BoolVar(&existing, "existing", false,
		"connect an account the app is already installed on: identify yourself on the provider, then pick from what it says you reach")
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
) (*platform.VcsConnection, *platform.AttachChoices, error) {
	deadline := time.Now().Add(wait)
	ticker := time.NewTicker(connectPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-ticker.C:
		}

		conns, err := pc.Client.ListConnections(ctx)
		if err != nil {
			lastErr = err
		} else {
			lastErr = nil
			for _, c := range conns {
				if !before[c.ConnectionId] {
					return c, nil, nil
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
				}, nil
			}
			if status.Settled() {
				if status.ConnectionId != "" {
					// It connected, and the listing has not caught up. Fetch it
					// rather than reporting a bare id.
					if c, gerr := pc.Client.GetConnection(ctx, status.ConnectionId); gerr == nil {
						return c, nil, nil
					}
				}
				return nil, nil, fmt.Errorf("%s", strings.TrimSpace(status.Message))
			}
		}

		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, nil, fmt.Errorf("gave up after %s, and the last check failed: %w", wait, lastErr)
			}
			return nil, nil, fmt.Errorf(
				"gave up after %s: workspace %s reports no outcome for this installation yet.\n"+
					"The provider page may still be open, or it was closed without finishing.\n"+
					"`asgard-cli pipeline connections` shows what this workspace has.", wait, pc.Workspace)
		}
	}
}

// pickAndAttach shows what the provider says this person reaches and connects
// the one they choose.
//
// One installation is still shown rather than taken silently: the person is
// about to give a workspace access to somebody's repositories, and seeing which
// account and how much of it is the last moment that is cheap.
func pickAndAttach(cmd *cobra.Command, pc *platformContext, choices *platform.AttachChoices) (*platform.VcsConnection, error) {
	msg := cmd.ErrOrStderr()
	if len(choices.Installations) == 0 {
		return nil, fmt.Errorf("the provider reports no installation you can reach")
	}

	fmt.Fprintf(msg, "\nThe provider says you can reach:\n\n")
	for i, in := range choices.Installations {
		scope := in.RepositorySelection
		if scope == "all" {
			scope = "all repositories"
		} else if scope == "selected" {
			scope = "selected repositories"
		}
		held := ""
		if in.ConnectionId != "" {
			held = "  (already connected here)"
		}
		fmt.Fprintf(msg, "  %d) %s (%s) - %s%s\n", i+1, in.AccountLogin, in.AccountType, scope, held)
	}

	choice := 1
	if len(choices.Installations) > 1 {
		fmt.Fprintf(msg, "\nWhich one? [1-%d]: ", len(choices.Installations))
		if _, err := fmt.Fscanln(cmd.InOrStdin(), &choice); err != nil {
			return nil, fmt.Errorf("no choice made: %w", err)
		}
		if choice < 1 || choice > len(choices.Installations) {
			return nil, fmt.Errorf("%d is not one of the choices", choice)
		}
	}
	picked := choices.Installations[choice-1]
	fmt.Fprintf(msg, "\nConnecting %s...\n", picked.AccountLogin)
	return pc.Client.AttachInstallation(cmd.Context(), choices.AttachState, picked.InstallationId)
}
