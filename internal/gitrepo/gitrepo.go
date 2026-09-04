// Package gitrepo reads the checkout the CLI is standing in.
//
// This is not the git integration TASK.md rules out. That non-goal is about
// provisioning - `git init`, adding a remote, authenticating to one - because
// Asgard grows its own mechanism for handing a customer a repository, and an
// agent that offers to set one up gets in the way of it. Reading what a
// checkout already says is a different thing, and it is what keeps platform
// identifiers out of the customer's repository: the pipeline a command acts on
// is found by matching the origin remote against the workspace's pipelines,
// rather than by a file somebody has to keep in sync with the platform.
//
// Every function here degrades rather than fails. A directory that is not a
// checkout, a checkout with no origin, a machine with no git binary: each is a
// reason a command cannot resolve context on its own and has to be told, not a
// reason for it to stop.
package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ErrNotARepository reports that the directory is not inside a git checkout.
var ErrNotARepository = errors.New("not a git repository")

// ErrNoOrigin reports a checkout with no origin remote.
var ErrNoOrigin = errors.New("no origin remote")

// ErrNoGit reports that the git binary was not found. It is separated from the
// other two because the remedy is entirely different: install git, rather than
// run the command somewhere else.
var ErrNoGit = errors.New("git is not installed or not on PATH")

// Root returns the top level of the checkout containing dir.
func Root(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		if errors.Is(err, ErrNoGit) {
			return "", err
		}
		return "", fmt.Errorf("%s: %w", dir, ErrNotARepository)
	}
	return out, nil
}

// OriginURL returns the origin remote's URL.
func OriginURL(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "remote", "get-url", "origin")
	if err != nil {
		if errors.Is(err, ErrNoGit) {
			return "", err
		}
		return "", ErrNoOrigin
	}
	return out, nil
}

// HeadSHA returns the full commit SHA of HEAD.
func HeadSHA(ctx context.Context, dir string) (string, error) {
	return run(ctx, dir, "rev-parse", "HEAD")
}

// remotePattern matches the two URL shapes a provider hands out, capturing
// owner and repository:
//
//	git@github.com:owner/name.git
//	https://github.com/owner/name(.git)
//
// It is deliberately not a general URL parser. The only thing this has to
// produce is the "owner/name" a pipeline records, and a shape it does not
// recognise is reported as such rather than guessed at - a wrong guess would
// silently match no pipeline, which reads like "you have not created one".
var remotePattern = regexp.MustCompile(`^(?:[a-z]+://)?(?:[^@/]+@)?[^/:]+[:/]([^/]+)/(.+?)(?:\.git)?/?$`)

// FullName reduces a remote URL to "owner/name", or returns false when the URL
// is not a shape it recognises.
func FullName(remoteURL string) (string, bool) {
	m := remotePattern.FindStringSubmatch(strings.TrimSpace(remoteURL))
	if m == nil {
		return "", false
	}
	owner, name := m[1], m[2]
	if owner == "" || name == "" {
		return "", false
	}
	return owner + "/" + name, true
}

// OriginFullName is the composition the callers actually want: the origin
// remote of the checkout containing dir, as "owner/name".
func OriginFullName(ctx context.Context, dir string) (string, error) {
	url, err := OriginURL(ctx, dir)
	if err != nil {
		return "", err
	}
	full, ok := FullName(url)
	if !ok {
		return "", fmt.Errorf("origin %q is not a shape this understands, so the repository it names cannot be matched against a pipeline", url)
	}
	return full, nil
}

// run executes one git command and returns its trimmed stdout.
func run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return "", ErrNoGit
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
