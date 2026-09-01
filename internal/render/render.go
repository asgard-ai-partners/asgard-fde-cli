// Package render renders a project's chart the way CD does.
//
// It replaces common/render.sh, and the reason is Windows: that script is bash
// with BASH_SOURCE and set -euo pipefail, and it shells out to yq for three
// field reads. None of that runs on Windows without WSL or Git Bash, which made
// the whole acceptance gate unavailable there even though helm and kubectl both
// have native Windows builds.
//
// It renders only. Deployment is CD-only, because a Syncer pins its revision to
// the chart's appVersion and only CI stamps the release tag in; a local helm
// upgrade writes the placeholder version as a git ref that does not exist, and
// the Syncer then fails to clone on every run.
package render

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/deploy"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/tool"
)

// Options is one render.
type Options struct {
	Root    string
	Project string
	Env     string
}

// Result describes what was rendered, for the caller to report.
type Result struct {
	Namespace string
	Release   string
	Chart     string
	Values    []string
}

// Run renders the chart to out. Anything helm writes to its own stderr goes to
// errOut, so that `asgard-cli render x dev | asgard-cli check xref -` keeps the
// pipe clean - the same property render.sh had by writing its progress line to
// stderr.
func Run(ctx context.Context, opts Options, out, errOut io.Writer) (Result, error) {
	file, err := deploy.Load(opts.Root, opts.Project)
	if err != nil {
		return Result{}, err
	}
	target, err := file.Target(opts.Project, opts.Env)
	if err != nil {
		return Result{}, err
	}

	projectDir := filepath.Join(opts.Root, "projects", opts.Project)
	res := Result{
		Namespace: target.Namespace,
		// The release name is the project name, the same as CI's main.yaml.
		Release: opts.Project,
		Chart:   filepath.Join(projectDir, "chart", "app"),
		// Shared values first, the project's own second, so the project
		// overrides the shared overlay. This order is CI's order; reversing it
		// renders something CD will never apply.
		Values: []string{
			filepath.Join(opts.Root, "common", fmt.Sprintf("values-%s.yaml", opts.Env)),
			filepath.Join(projectDir, target.Values),
		},
	}

	for _, path := range append([]string{filepath.Join(res.Chart, "Chart.yaml")}, res.Values...) {
		if err := mustExist(path); err != nil {
			return res, err
		}
	}

	args := []string{"template", res.Release, res.Chart, "--namespace", res.Namespace}
	for _, v := range res.Values {
		args = append(args, "-f", v)
	}

	cmd, err := tool.Helm.Command(ctx, args...)
	if err != nil {
		return res, err
	}
	cmd.Stdout = out
	cmd.Stderr = errOut
	cmd.Dir = opts.Root

	if err := cmd.Run(); err != nil {
		// helm has already written its own diagnosis to errOut; repeating it
		// here would print the same thing twice.
		return res, fmt.Errorf("helm template failed for %s/%s", opts.Project, opts.Env)
	}
	return res, nil
}

// mustExist turns a missing input into a sentence naming the file, since the
// alternative is helm's own error, which names a temporary path.
func mustExist(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("missing %s", path)
	}
	return nil
}
