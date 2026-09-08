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
    4. no match, so the app's install page is opened for that account

Nothing prompts. --account defaults to the owner of this checkout's origin
remote, which is the account whose repositories this engagement is about; name
it explicitly when connecting some other account, or when there is no remote to
read.

Step 4 is what makes the account you asked for reachable, and it is a step
nobody can take unaided: the install page is /apps/<slug>/installations/new, the
slug differs per platform, and a wrong slug and a private app look identical
from outside. So the platform is asked for that URL rather than anybody guessing
it. Whatever gets installed is then matched against --account again - an install
that lands on a different account is reported, not connected.

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
own connection to the same installation, and they do not see each other's.

AND ONE WORKSPACE, AS MANY ACCOUNTS. The other direction is a connection each:
run this once per account, and the workspace ends up holding one connection per
installation, which is what "pipeline create --connection" chooses between.
Holding one already is not a reason this stops - it was, for as long as the only
way in was the list of installations you already reach.`,
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
			// Only when the remote is on the provider being connected. An
			// account name is a name ON a provider: "acme" from a gitlab.com
			// remote is not a GitHub account, and treating it as one would
			// connect whatever GitHub org happens to share the name — silently,
			// and correctly as far as anything here could tell.
			if wanted == "" && pc.RepoFullName != "" && isGitHubHost(pc.RepoHost) {
				// The owner of this checkout's remote is the account this
				// engagement is about. Derived rather than asked, because
				// asking would need somebody at the terminal and this command
				// is run by an agent.
				//
				// Said out loud, like `pipeline create` does with --repo: a
				// value taken from the remote is a suggestion the caller did
				// not make, and one it cannot correct if it never hears it.
				wanted, _, _ = strings.Cut(pc.RepoFullName, "/")
				fmt.Fprintf(cmd.ErrOrStderr(),
					"no --account, so the owner of this checkout's origin remote is used: %s\n", wanted)
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
				created, err = attachNamedAccount(cmd, pc, choices, wanted, before, noBrowser, wait)
			case insufficient:
				// Authorized, and reaches nothing here. Not a failure to report
				// — it is the answer to "which way in", arrived at without
				// asking anybody to know it in advance.
				fmt.Fprintf(msg, "\nYou reach no installation of this app yet, so it has to be installed first.\n")
				created, err = installFirst(cmd, pc, before, wanted, noBrowser, wait)
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
			// The id is on the line above. A suggestion that cannot be run as
			// printed costs a round trip for nothing, and `pipeline repos`
			// refuses without --connection on purpose - nothing is assumed
			// from a list of one, including a list of one connection.
			fmt.Fprintf(out, "\n`asgard-cli pipeline repos --connection %s` lists what it can reach; a\nrepository missing from that list is one the installation was not granted,\nwhich is changed on the provider rather than here.\n", created.ConnectionId)
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
// when there is no such installation, goes and gets one.
//
// Nothing prompts. This command is run by an agent following a skill, not by
// somebody at a terminal — `init` is the one command in this tool with a person
// in front of it — so a question here is not a slower path, it is a dead one.
// The caller declares which account it means and the platform matches it.
//
// **The miss used to end the command, and that was the dead end.** Reaching an
// installation at all sends every call down this path, so holding one — the
// state every returning engagement is in — made a second account unreachable
// from this tool: the remedy on offer was "pick one of the ones you already
// have", which is never the answer when the point is a new one. Installing is
// the step that was missing, and only the platform knows the URL for it, so
// there was nowhere else to get it either.
func attachNamedAccount(
	cmd *cobra.Command,
	pc *platformContext,
	choices *platform.AttachChoices,
	account string,
	before map[string]bool,
	noBrowser bool,
	wait time.Duration,
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
		// Nothing to match and nothing to install towards either: which account
		// this engagement is about is the one fact neither the checkout nor the
		// provider supplied.
		return nil, fmt.Errorf(
			"no --account, and this is not a git checkout with a recognisable origin remote.\n" +
				"Name the account to connect with --account; the ones you reach are listed above")
	}
	if in := findInstallation(choices.Installations, account); in != nil {
		return attach(cmd, pc, choices.AttachState, in)
	}

	// Says what was established without asserting which of the two reasons it
	// is. Both are real: the app may not be installed there, or this identity
	// may not reach it. Installing is the one this can act on — and the only
	// one whose next step nobody can work out for themselves.
	fmt.Fprintf(msg,
		"\n%q is not among the accounts you reach on the provider (listed above).\n"+
			"Either the app is not installed on it, or the account you authorized as\n"+
			"cannot see its repositories. Installing is the fix for the first.\n",
		account)
	return installFirst(cmd, pc, before, account, noBrowser, wait)
}

// installFirst prints where to install the app and waits. Installing is a
// person's action on the provider; there is nothing this can do but say where.
//
// account is what the caller asked for, and it is checked again on the way
// back. The provider's install page asks which account to install on, so an
// install can land somewhere other than where this was heading — and reporting
// that as success would connect an account nobody named. Empty only on the way
// in from "you reach nothing at all", where there is nothing to tell apart.
func installFirst(
	cmd *cobra.Command,
	pc *platformContext,
	before map[string]bool,
	account string,
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
	where := "the account you want"
	if account != "" {
		where = account
	}
	fmt.Fprintf(msg, "\nOpen this URL and install the app on %s:\n\n    %s\n\n", where, install.InstallUrl)
	fmt.Fprintf(msg, "Installing on an organisation may need one of its admins, which is a wait\nrather than a failure.\n")
	fmt.Fprintf(msg, "Waiting...\n")

	created, choices, _, err := waitForConnection(ctx, pc, before, install.State, wait)
	if err != nil {
		return nil, err
	}
	if choices != nil {
		// The provider sent them through authorization on the way, so this
		// ended in a list rather than a connection.
		return attachInstalled(cmd, pc, choices, account)
	}
	if created == nil {
		// Settled, reaching nothing, and no connection: the install and the
		// authorization went to accounts that do not overlap.
		return nil, fmt.Errorf(
			"the app was installed, but the account that authorized reaches no installation of it.\n" +
				"Authorize as an account with repository access to the one it was installed on")
	}
	if account != "" && !strings.EqualFold(created.AccountLogin, account) {
		return nil, fmt.Errorf(
			"the app was installed on %q rather than the %q that was asked for, and %q is\n"+
				"what this workspace is now connected to.\n"+
				"Re-run to install on %s, or `--account %s` to keep what you got",
			created.AccountLogin, account, created.AccountLogin, account, created.AccountLogin)
	}
	return created, nil
}

// attachInstalled connects what the install produced, matched against the same
// --account the caller named.
//
// **It used to connect whichever installation this workspace did not already
// hold**, which is only right when it held none — the one state the install
// path used to be reachable from. Reached from a miss it is not: the caller can
// reach installations that were never connected here, so "not held" names
// several, and taking the first would connect an account nobody asked for.
// Silently, and looking exactly like success.
func attachInstalled(
	cmd *cobra.Command,
	pc *platformContext,
	choices *platform.AttachChoices,
	account string,
) (*platform.VcsConnection, error) {
	if account != "" {
		in := findInstallation(choices.Installations, account)
		if in == nil {
			return nil, fmt.Errorf(
				"%q is still not among the accounts you reach, so nothing was installed on it.\n"+
					"`asgard-cli pipeline connections` shows what this workspace has",
				account)
		}
		return attach(cmd, pc, choices.AttachState, in)
	}

	// No account named, which is the "you reach nothing at all" way in. What
	// this workspace does not hold is what was just installed — and when that
	// is more than one thing, saying so beats picking.
	var fresh []*platform.UserInstallation
	for _, in := range choices.Installations {
		if in.ConnectionId == "" {
			fresh = append(fresh, in)
		}
	}
	switch len(fresh) {
	case 0:
		return nil, fmt.Errorf("nothing new was installed; every account you reach is already connected here")
	case 1:
		return attach(cmd, pc, choices.AttachState, fresh[0])
	}
	names := make([]string, 0, len(fresh))
	for _, in := range fresh {
		names = append(names, in.AccountLogin)
	}
	return nil, fmt.Errorf(
		"more than one account you reach is unconnected here (%s), so which one was just\n"+
			"installed cannot be told apart.\n"+
			"`asgard-cli pipeline connect --account <one of them>` names it",
		strings.Join(names, ", "))
}

// attach records the connection, saying which account it is for first: the
// account is the whole question this command answers, so the one it settled on
// is worth reading before the result.
func attach(cmd *cobra.Command, pc *platformContext, attachState string, in *platform.UserInstallation) (*platform.VcsConnection, error) {
	fmt.Fprintf(cmd.ErrOrStderr(), "\nConnecting %s...\n", in.AccountLogin)
	return pc.Client.AttachInstallation(cmd.Context(), attachState, in.InstallationId)
}

// findInstallation is the account match, in the one place both paths use so
// they cannot come to different answers. Case-insensitive, because provider
// account names are: a caller that typed one differently means the same
// account.
func findInstallation(installations []*platform.UserInstallation, account string) *platform.UserInstallation {
	for _, in := range installations {
		if strings.EqualFold(in.AccountLogin, account) {
			return in
		}
	}
	return nil
}

// isGitHubHost reports whether a remote host is github.com.
//
// Deliberately not "anything that looks like a git host": GitHub Enterprise
// lives on a host this cannot know, so an account there has to be named with
// --account rather than assumed. Refusing to guess costs one flag; guessing
// wrong connects an account nobody asked for.
func isGitHubHost(host string) bool {
	return host == "github.com" || host == "www.github.com"
}
