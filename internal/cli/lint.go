package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

func newLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Check that " + config.FileName + " exists and is valid",
		Long: `Check that ` + config.FileName + ` exists and is valid.

The config file is looked up from the current directory upwards. If it is missing,
cannot be parsed, or its workspace fields are incomplete, every problem is listed
and the command exits non-zero, which makes it suitable for CI or a pre-commit hook.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}

			path, err := config.Find(dir)
			if err != nil {
				if errors.Is(err, config.ErrNotFound) {
					return fmt.Errorf("no %s found; run `asgard-cli init` first", config.FileName)
				}
				return err
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			shown := displayPath(dir, path)

			if probs := unwrapJoined(cfg.Validate()); len(probs) > 0 {
				for _, p := range probs {
					fmt.Fprintf(out, "- %s\n", p)
				}
				return fmt.Errorf("%s has %s", shown, pluralize(len(probs), "problem"))
			}

			fmt.Fprintf(out, "ok  %s\n", shown)
			return nil
		},
	}

	return cmd
}

// unwrapJoined splits an errors.Join chain into its individual problems so each
// one can be shown on its own line. A non-joined error becomes a single element.
func unwrapJoined(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}
	return []error{err}
}

// displayPath shows path relative to base when it can, to keep output short.
func displayPath(base, path string) string {
	if rel, err := filepath.Rel(base, path); err == nil {
		return rel
	}
	return path
}

// pluralize renders a count with its noun, adding "s" for anything but one.
func pluralize(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
