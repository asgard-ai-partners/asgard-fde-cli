package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
)

func newAddCmd() *cobra.Command {
	var opts generate.Options
	var layers []string

	cmd := &cobra.Command{
		Use:   "add <kind> <name>",
		Short: "Write a CR skeleton into a project's chart",
		Long: `Write a CR skeleton into a project's chart, correct in the parts that fail
silently.

A missing display annotation shows up as a nameless resource in the UI. A
Workflow without its set labels is invisible there. A Trigger without its own two
labels opens as a blank canvas. A field renamed upstream still lints clean under
its old name. **None of those are caught by helm lint, by CRD validation, or by a
server-side dry-run** - which is why they are worth generating rather than
typing.

What is generated is a skeleton: the structure and the traps are right, and the
content is marked TODO. Read the matching shape first - "asgard-cli add" names
it for each kind - because the decision comes before the YAML.

The name is written without the kind's prefix: "asgard-cli add dataconnector erp"
creates dc-erp. Passing the prefixed form is accepted and means the same thing,
so dc-erp never becomes dc-dc-erp.

A flowagent serves your own front end unless --bot-class names a chat platform:

    asgard-cli add flowagent support --project site --bot-class line

That writes the channel's credential block and says what the channel costs -
which credentials infra has to add, and whether the class needs a connector pod.
The field is immutable on the platform side, so it is worth getting right the
first time. Read "asgard-cli usecase chat-channel" before choosing.

Run "asgard-cli add" with no arguments to list the kinds.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if len(args) == 0 {
				fmt.Fprintf(out, "Kinds, in the order they are usually created:\n\n")
				for _, k := range generate.Kinds {
					fmt.Fprintf(out, "  %-14s %s\n", k.Name, k.Summary)
					fmt.Fprintf(out, "  %-14s what it is:  asgard-cli wiki %s\n", "", k.Wiki)
					fmt.Fprintf(out, "  %-14s how to build: asgard-cli usecase %s\n", "", k.Extract)
					if len(k.Needs) > 0 {
						fmt.Fprintf(out, "  %-14s needs: %s\n", "", strings.Join(k.Needs, ", "))
					}
					fmt.Fprintln(out)
				}
				fmt.Fprintf(out, "  asgard-cli add <kind> <name> --project <slug>\n")
				return nil
			}
			if len(args) < 2 {
				return fmt.Errorf("needs a kind and a name: asgard-cli add <kind> <name> --project <slug>")
			}

			kind, ok := generate.Find(args[0])
			if !ok {
				return fmt.Errorf("unknown kind %q; one of: %s", args[0], generate.Names())
			}
			opts.Name = args[1]
			opts.Layers = layers

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
			root := filepath.Dir(path)

			if opts.Project == "" {
				if len(cfg.Projects) != 1 {
					return fmt.Errorf("--project is required; this repo has %d projects", len(cfg.Projects))
				}
				// One project is unambiguous, so do not make them type it.
				opts.Project = cfg.Projects[0].Slug
			}
			if err := requiredFlags(kind, opts); err != nil {
				return err
			}

			opts, notes, err := generate.Resolve(root, kind, opts)
			if err != nil {
				return err
			}

			results, err := generate.Write(root, cfg, kind, opts)
			if err != nil {
				return err
			}
			for _, note := range notes {
				fmt.Fprintf(out, "%s\n", note)
			}

			for _, r := range results {
				if r.Values {
					fmt.Fprintf(out, "updated %s (added the values it reads)\n", r.Path)
					continue
				}
				fmt.Fprintf(out, "created %s\n", r.Path)
			}

			if len(kind.After) > 0 {
				fmt.Fprintf(out, "\nBefore this works:\n")
				for _, note := range kind.After {
					fmt.Fprintf(out, "  %s\n", note)
				}
			}

			// Both, in reading order: the wiki page says what the thing is,
			// the extract says how it is assembled and assumes you know the
			// first. The listing prints the same pair.
			fmt.Fprintf(out, "\nNext:\n")
			fmt.Fprintf(out, "  1. what it is:      asgard-cli wiki %s\n", kind.Wiki)
			fmt.Fprintf(out, "  2. how to build it: asgard-cli usecase %s\n", kind.Extract)
			// The mechanism extracts, where they apply. A reader who knows the
			// shape and not how values cross between processors writes the
			// silent failures back in.
			for _, also := range kind.AlsoRead {
				fmt.Fprintf(out, "     and:            asgard-cli usecase %s\n", also)
			}
			fmt.Fprintf(out, `  3. fill in the TODOs
  4. verify:          asgard-cli check
                      asgard-cli verify %s
`, opts.Project)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Project, "project", "", "project to add it to (optional when the repo has one)")
	cmd.Flags().StringVar(&opts.DisplayName, "display-name", "", "name shown in the platform UI (defaults to the CR name)")
	cmd.Flags().StringVar(&opts.Connector, "connector", "", "DataConnector this reads through")
	cmd.Flags().StringVar(&opts.Layer, "layer", "", "SemanticLayer to mount")
	cmd.Flags().StringSliceVar(&layers, "layers", nil, "SemanticLayers a scheduled run mounts, repeatable")
	cmd.Flags().StringVar(&opts.Toolset, "toolset", "", "Toolset to create alongside the tool")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "git repository the skills come from")
	cmd.Flags().BoolVar(&opts.Private, "private", false, "the skills repo is private and needs a PAT")
	cmd.Flags().BoolVar(&opts.Write, "write", false, "this tool has a side effect, so gate it with requestConsent")
	cmd.Flags().BoolVar(&opts.Public, "public", false, "the entry point serves anonymous visitors (authMode: none)")
	cmd.Flags().StringVar(&opts.DBClass, "db-class", "postgres", "database class: postgres or mssql")
	cmd.Flags().StringVar(&opts.BotClass, "bot-class", "", "for flowagent, the channel the BotProvider serves: "+strings.Join(generate.BotClasses, ", ")+" (defaults to generic, an HTTP API for your own front end)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "overwrite an existing file")

	return cmd
}

// requiredFlags reports what a kind cannot be generated without, rather than
// emitting a skeleton with a dangling reference the gate would reject later.
func requiredFlags(kind generate.Kind, opts generate.Options) error {
	switch kind.Name {
	case "semanticlayer", "querytool":
		if opts.Connector == "" {
			return fmt.Errorf("%s needs --connector dc-<name>; create one first with `asgard-cli add dataconnector <name>`", kind.Name)
		}
	case "skillset":
		if opts.Repo == "" {
			return fmt.Errorf("skillset needs --repo <git url>")
		}
	}
	if kind.Name == "httptool" || kind.Name == "querytool" {
		if opts.Toolset == "" {
			return fmt.Errorf("%s needs --toolset ts-<name>, the set this tool belongs to", kind.Name)
		}
	}
	if kind.Name == "dataconnector" && opts.DBClass != "postgres" && opts.DBClass != "mssql" {
		return fmt.Errorf("--db-class must be postgres or mssql, got %q", opts.DBClass)
	}
	return nil
}
