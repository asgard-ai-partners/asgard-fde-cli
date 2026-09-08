package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/generate"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
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

The rules every generated CR already follows - naming, the display annotations,
what goes in values and what stays in the template - are "asgard-cli usecase
conventions". Read it before writing a CR by hand, or before changing one this
wrote: what it generates is those conventions applied, and an edit that departs
from them is the half a rendered chart still passes.

The name is written without the kind's prefix: "asgard-cli add dataconnector erp"
creates dc-erp. Passing the prefixed form is accepted and means the same thing,
so dc-erp never becomes dc-dc-erp.

**--db-class takes any of the nine DataConnector classes and they share almost
nothing.** salesforce has no port and no user; athena has neither host nor
database, just a region, an S3 output location and an IAM key pair; netsuite
authenticates with a certificate whose PEM is a secret and whose id is not;
oracle takes serviceName **or** sid and the CRD refuses both and refuses
neither. Each skeleton carries that class's fields and the note its shape cannot
say. hana is generated like the rest - the platform reads it - but
the db-query skill cannot reach it, because SAP does not distribute its driver
openly.

A flowagent serves your own front end unless --bot-class names a chat platform:

    asgard-cli add flowagent support --project site --bot-class line

That writes the channel's credential block and says what the channel costs -
which credential keys have to be declared and set, and whether the class needs a connector pod.
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

			root, err := loadRepo()
			if err != nil {
				return err
			}

			// **Not "one project, so use it".** A rule that only holds while
			// there is exactly one changes behaviour silently on the day a
			// second appears, and nobody is watching that day.
			if opts.Project == "" {
				projects, err := repo.Projects(root)
				if err != nil {
					return err
				}
				return fmt.Errorf("--project is required. This repository has: %s",
					strings.Join(projects, ", "))
			}
			if err := requiredFlags(kind, opts); err != nil {
				return err
			}

			opts, notes, err := generate.Resolve(root, kind, opts)
			if err != nil {
				return err
			}

			results, err := generate.Write(root, kind, opts)
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

			// What was just written is what has to be declared. Naming the keys is
			// the half that was missing: `variables set` alone stores a value and
			// never injects it, and every local check stays green while the CR
			// resolves to nothing at runtime.
			secretKeys, configKeys, err := generate.SecretKeysWritten(results)
			if err != nil {
				return err
			}
			if len(secretKeys) > 0 || len(configKeys) > 0 {
				fmt.Fprintf(out, "\nBefore any of it resolves, declare these in %s and set each on the platform:\n", pipelineconfig.FileName)
				if len(secretKeys) > 0 {
					fmt.Fprintf(out, "  appSecret:     %s\n", strings.Join(secretKeys, ", "))
					fmt.Fprintf(out, "    asgard-cli pipeline variables set --release <release> --kind %s %s --from-file <path>\n", kindSecret, secretKeys[0])
				}
				if len(configKeys) > 0 {
					fmt.Fprintf(out, "  appConfigMap:  %s\n", strings.Join(configKeys, ", "))
					fmt.Fprintf(out, "    asgard-cli pipeline variables set --release <release> --kind %s %s <value>\n", kindConfig, configKeys[0])
				}
				fmt.Fprintf(out, "  A value set against no declaration is stored and never injected - `variables list` calls it ORPHAN,\n")
				fmt.Fprintf(out, "  and lint, render and the server dry run all stay green while the CR resolves to nothing.\n")
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
			fmt.Fprint(out, `  3. fill in the TODOs
  4. verify:          asgard-cli check
                      asgard-cli verify
`)
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
	cmd.Flags().BoolVar(&opts.Supervisor, "supervisor", false,
		"for flowagent, write the four-processor conversation loop three deployments share, whose specialists are subagents on the blueprint")
	cmd.Flags().StringVar(&opts.DBClass, "db-class", "postgres",
		"DataConnector class: "+strings.Join(generate.DBClasses(), ", ")+". Each has its own fields - salesforce has no port and no user, athena has neither host nor database")
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
	if kind.Name == "dataconnector" && !generate.ValidDBClass(opts.DBClass) {
		return fmt.Errorf("--db-class %q is not a DataConnectorClass; the CRD has %s",
			opts.DBClass, strings.Join(generate.DBClasses(), ", "))
	}
	return nil
}
