package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/localenv"
)

func newLocalEnvCmd() *cobra.Command {
	var (
		focus     []string
		noBrowser bool
		terminal  bool
		timeout   time.Duration
	)

	cmd := &cobra.Command{
		Use:   "local-env",
		Short: "Fill in this repository's design-time credentials, without typing them at an agent",
		Long: `Open a form for the ` + "`.env`" + ` at the repository root, so that whoever holds a
credential can type it in themselves.

**A coding agent must never ask anybody to say a password to it.** Not in the
conversation, not "paste it and I will remove it after" - a credential that has
been through a transcript has to be treated as disclosed. But the alternative
has been to ask a person who may not be an engineer to open a dotfile, find the
right line, and mind the whitespace, which is a request that fails.

So the agent writes the keys it needs, with the values left empty, and this
opens a form to fill them in:

    asgard-cli local-env
    asgard-cli local-env --focus UOF_DB_HOST,UOF_DB_PASSWORD

It serves one page on 127.0.0.1 on a random port, opens a browser at it, and
closes as soon as the form is saved. The URL carries a one-time token, the
server answers to no other host name, and the page is served under a policy
that lets it talk to nothing but the process that served it.

**What comes back here is a list of key names.** Never a value - not on save,
not in an error, not in the summary. That is the whole point of the command:
the values reach the file and stop there.

` + "`--focus`" + ` highlights the keys you are waiting for. **It does not hide the
others**, deliberately: the person filling this in may know about a second
database nobody has mentioned yet, and a form that shows them only what was
asked for is a form that cannot tell you about it. They can add keys too, which
is why the agent should re-read ` + "`.env`" + ` afterwards rather than assume it got
back exactly what it asked for.

**The kinds of credential, and this is only one of them:**

    design time         this .env, on this machine       the customer, or you
    pipeline variables  the platform, per release        asgard-cli pipeline variables set
    runtime secret      a Kubernetes Secret in a cluster the platform provisions it

The value is often the same string, because it is the same database. How each
is set is not, and reading one to obtain another - a cluster Secret to get a
design-time password, say - is how a production credential ends up somewhere
nobody can withdraw it from.

**With no browser** - a container, a locked-down server - the URL is printed for
you to open, over an SSH port forward if that is what it takes. Where even that
is not possible, ` + "`--terminal`" + ` asks for each value at the prompt instead, and
does not echo the secrets.`,
		Args:    cobra.NoArgs,
		GroupID: groupBuild,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := envRoot()
			if err != nil {
				return err
			}
			file, err := localenv.Load(root)
			if err != nil {
				return err
			}

			errOut := cmd.ErrOrStderr()
			if err := ensureIgnored(root); err != nil {
				fmt.Fprintf(errOut, "warning: %v\n", err)
			}

			focus = splitFocus(focus)
			if missing := localenv.MissingFocus(file, focus); len(missing) > 0 {
				// A typo in --focus would otherwise highlight nothing and look
				// like the form simply had nothing to point at.
				fmt.Fprintf(errOut, "warning: --focus names %s that %s not in %s yet, so nothing is highlighted for %s\n",
					plural(len(missing), "key"), isAre(len(missing)), localenv.FileName,
					strings.Join(missing, ", "))
			}

			opts := localenv.Options{
				File: file, Focus: focus, Timeout: timeout, NoBrowser: noBrowser,
			}

			var res localenv.Result
			if terminal {
				res, err = localenv.Terminal(opts, os.Stdin, errOut)
			} else {
				res, err = localenv.Serve(cmd.Context(), opts, errOut)
			}
			if err != nil {
				return err
			}

			// stdout, and key names only: this is what an agent reads back.
			res.Report(cmd.OutOrStdout())
			return nil
		},
	}

	f := cmd.Flags()
	f.StringSliceVar(&focus, "focus", nil,
		"highlight these keys in the form; it never hides the rest")
	f.BoolVar(&noBrowser, "no-browser", false,
		"do not open a browser; the URL is printed either way")
	f.BoolVar(&terminal, "terminal", false,
		"ask for each value at the prompt, for a machine that cannot reach a browser at all")
	f.DurationVar(&timeout, "timeout", 15*time.Minute,
		"give up if nothing is saved by then, so a tab holding credentials does not outlive the task")

	return cmd
}

// envRoot is the repository the .env belongs to.
//
// The declaration is what marks a repository, the same rule `render` uses. A
// .env somewhere above or below it would be a second answer to where the
// credentials live.
func envRoot() (string, error) {
	root, err := repoRoot()
	if err == nil {
		return root, nil
	}
	dir, wdErr := os.Getwd()
	if wdErr != nil {
		return "", err
	}
	return "", fmt.Errorf("%w\n  local-env edits the %s beside it, and %s is not in one",
		err, localenv.FileName, dir)
}

// ensureIgnored keeps .env out of git.
//
// It is a warning rather than an error, and it appends rather than asking:
// somebody is about to type a customer's password into this file, and a
// .gitignore that does not mention it is one `git add -A` away from putting
// that password in a repository's history, where removing it is not a delete.
func ensureIgnored(root string) error {
	path := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		switch strings.TrimSpace(line) {
		case ".env", "/.env", "*.env", ".env*":
			return nil
		}
	}
	body := string(data)
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += "\n# Design-time credentials. Added by `asgard-cli local-env`.\n.env\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("%s did not cover %s, and adding it failed: %w",
			path, localenv.FileName, err)
	}
	return fmt.Errorf("%s did not cover %s; added it", path, localenv.FileName)
}

// splitFocus accepts both --focus A,B and --focus A --focus B, and drops the
// empty strings a trailing comma leaves behind.
func splitFocus(in []string) []string {
	var out []string
	for _, item := range in {
		for _, k := range strings.Split(item, ",") {
			if k = strings.TrimSpace(k); k != "" {
				out = append(out, k)
			}
		}
	}
	return out
}

func isAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}
