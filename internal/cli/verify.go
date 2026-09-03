package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/deploy"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/gate"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/render"
)

func newVerifyCmd() *cobra.Command {
	var (
		rendered string
		layers   []string
		tools    bool
	)

	cmd := &cobra.Command{
		Use:   "verify [project ...]",
		Short: "Render each project and check the invariants a render cannot see",
		Long: `Render each project and check the invariants that helm lint, CRD validation and a
server-side dry run all pass:

  - every CR that the Platform UI lists carries its <kind>-name annotation,
    without which it applies cleanly and appears with no name
  - every reference between CRs resolves, both halves of it: an entrypoint is
    (workflow, entry), and a wrong entry is as dead as a wrong workflow
  - a Workflow has its full set of workflow-set labels, and each set has exactly
    one main; a Trigger's workflow-set-id matches its entrypoint Workflow's, or
    its editor opens blank. The other label an editor needs,
    project-environment-id, is a warning rather than a failure - it comes from
    platformMainEnvironmentId, which does not exist until tf-asgard has created
    the namespace, so it is legitimately empty through the middle of an
    onboarding
  - a Syncer's paths obey the CRD's relative-path rules
  - the agent split: at most one semantic layer per Agent, no layer bound twice,
    no allowedCubes, sampleQuestions on anything published, and prompt.task and
    prompt.format identical across every Agent in one render

It also warns about what CD requires and the apiserver does not: a project with
no Syncer at all is refused by CD after 180 seconds even when helm upgrade
succeeded, and an empty platformMainEnvironmentId renders CRs with no
project-environment-id label. Both are correct during an onboarding and fatal
once someone tags, so they are warnings and do not fail the gate.

A layer that feeds Data Insight rather than a chat agent belongs in the
olapOnlyLayers list of .asgard-config.json, so the rule applies on every run
instead of only when somebody remembers the flag.

With no arguments it does every project in the config, once per environment that
project's deploy.yaml declares. It renders in process, so there is no pipeline
and no temporary file:

    asgard-cli verify
    asgard-cli verify erp
    asgard-cli verify --rendered .out/rendered.yaml
    asgard-cli verify --tools               every tool description, side by side

**--tools prints and checks nothing.** ` + "`tooling.description`" + ` is the single
field where a wrong value makes a model call the wrong tool, and what makes one wrong
is that it does not distinguish itself from the tool beside it - a property of
the set, not of any entry, and so not something a rule can read. A person
reading all of them at once can, and there was nowhere that put them together.
Read them as the model does: in one list, with no other context, deciding which
one answers the question.

This is steps 2 and 3 of the acceptance gate, and it needs only helm on PATH.
Step 1 is "asgard-cli check", and step 4 needs a cluster - see
"asgard-cli next --stage verify". Exits non-zero on any problem.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if rendered != "" {
				if len(args) > 0 {
					return fmt.Errorf("--rendered checks a stream that is already rendered, so it takes no project arguments")
				}
				docs, err := readRendered(cmd, rendered)
				if err != nil {
					return err
				}
				// A pre-rendered stream may come from outside a repository, so
				// the config is read when there is one and not required.
				olap := layers
				if _, cfg, err := loadRepo(); err == nil {
					olap = olapLayers(cfg, layers)
				}
				if tools {
					printTools(out, rendered, docs)
					return nil
				}
				if !runGates(out, rendered, docs, gate.Options{OLAPOnlyLayers: olap}) {
					return fmt.Errorf("verification failed")
				}
				return nil
			}

			root, cfg, err := loadRepo()
			if err != nil {
				return err
			}

			projects, err := projectsToVerify(cfg, args)
			if err != nil {
				return err
			}
			olap := olapLayers(cfg, layers)
			if err := recordOLAPLayers(root, cfg, layers, out); err != nil {
				return err
			}

			ok := true
			checked := 0
			for _, project := range projects {
				file, err := deploy.Load(root, project)
				if err != nil {
					return err
				}
				// Only the declared environments: not declaring one is a
				// deliberate decision not to deploy there, so there is nothing
				// to verify.
				for _, env := range config.Envs {
					if _, declared := file.Environments[string(env)]; !declared {
						continue
					}

					var buf bytes.Buffer
					if _, err := render.Run(cmd.Context(), render.Options{
						Root: root, Project: project, Env: string(env),
					}, &buf, cmd.ErrOrStderr()); err != nil {
						return err
					}

					docs, err := gate.Read(&buf)
					if err != nil {
						return fmt.Errorf("%s/%s: %w", project, env, err)
					}

					checked++
					if tools {
						printTools(out, fmt.Sprintf("%s/%s", project, env), docs)
						checked++
						continue
					}
					if !runGates(out, fmt.Sprintf("%s/%s", project, env), docs,
						gate.Options{Project: project, OLAPOnlyLayers: olap}) {
						ok = false
					}
				}
			}

			if checked == 0 {
				fmt.Fprintf(out, "Nothing to verify: no project declares an environment yet.\n")
				return nil
			}
			if !ok {
				return fmt.Errorf("verification failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&tools, "tools", false, "print every tool name and description instead of checking; the one review a rule cannot do")
	cmd.Flags().StringVar(&rendered, "rendered", "", "check a file of already-rendered manifests, or - for stdin (defaults to rendering each project)")
	cmd.Flags().StringSliceVar(&layers, "olap-only-layer", nil, "semantic layer deliberately bound to no Agent because it feeds Data Insight, repeatable; adds to olapOnlyLayers in "+config.FileName+" (defaults to whatever that records)")

	return cmd
}

// recordOLAPLayers writes what --olap-only-layer named into the config, so the
// rule holds on every later run rather than only on the run that passed the
// flag.
//
// The flag's help has always said it does this and it never did, which mattered
// little while the only way to learn about the list was to read that help.
// R11 now tells a reader to run this exact command to record a layer as
// deliberately unbound - and an instruction that silences one run and forgets is
// worse than no instruction, because the reader believes the fact is recorded.
func recordOLAPLayers(root string, cfg *config.Config, layers []string, out io.Writer) error {
	have := map[string]bool{}
	for _, name := range cfg.OLAPOnlyLayers {
		have[name] = true
	}
	var added []string
	for _, name := range layers {
		if name == "" || have[name] {
			continue
		}
		have[name] = true
		added = append(added, name)
	}
	if len(added) == 0 {
		return nil
	}

	cfg.OLAPOnlyLayers = append(cfg.OLAPOnlyLayers, added...)
	sort.Strings(cfg.OLAPOnlyLayers)
	if err := config.Save(filepath.Join(root, config.FileName), cfg); err != nil {
		return err
	}
	fmt.Fprintf(out, "Recorded %s in %s as bound to no Agent on purpose. R10 refuses any later attempt to bind %s to one.\n\n",
		strings.Join(added, ", "), config.FileName, pluralThem(added))
	return nil
}

// pluralThem keeps the sentence above readable for one layer and for several.
func pluralThem(names []string) string {
	if len(names) == 1 {
		return "it"
	}
	return "them"
}

// olapLayers combines what the config records with what the flag added, into a
// new slice: appending onto the config's own would let one run's flag leak into
// whatever reads the config next.
func olapLayers(cfg *config.Config, extra []string) []string {
	out := make([]string, 0, len(cfg.OLAPOnlyLayers)+len(extra))
	out = append(out, cfg.OLAPOnlyLayers...)
	return append(out, extra...)
}

// readRendered loads manifests from a file or stdin.
func readRendered(cmd *cobra.Command, path string) ([]gate.Doc, error) {
	var r io.Reader = cmd.InOrStdin()
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path, err)
		}
		defer f.Close()
		r = f
	}
	return gate.Read(r)
}

// projectsToVerify resolves the arguments against the config, so a typo is
// reported rather than silently verifying nothing.
func projectsToVerify(cfg *config.Config, args []string) ([]string, error) {
	if len(args) == 0 {
		names := make([]string, 0, len(cfg.Projects))
		for _, p := range cfg.Projects {
			names = append(names, p.Slug)
		}
		return names, nil
	}
	for _, name := range args {
		if _, ok := cfg.Project(name); !ok {
			return nil, fmt.Errorf("no project %q in %s", name, config.FileName)
		}
	}
	return args, nil
}

// wrap indents a continuation line under the label it belongs to, so a long
// message with a command in it stays readable in a terminal.
func wrap(s string) string {
	const width = 72
	var out strings.Builder
	col := 0
	for _, word := range strings.Fields(s) {
		if col > 0 && col+1+len(word) > width {
			out.WriteString("\n       ")
			col = 0
		} else if col > 0 {
			out.WriteString(" ")
			col++
		}
		out.WriteString(word)
		col += len(word)
	}
	return out.String()
}

// runGates runs every check over one render and reports them under one heading.
func runGates(out io.Writer, label string, docs []gate.Doc, opts gate.Options) bool {
	fmt.Fprintf(out, "%s\n", label)

	ok := true
	for _, r := range []gate.Result{
		gate.Xref(docs, opts),
		gate.AgentSplit(docs, opts),
		gate.Processors(docs, opts),
		gate.Enums(docs, opts),
		gate.Constraints(docs, opts),
		gate.Deployability(docs, opts),
	} {
		for _, p := range r.Problems {
			fmt.Fprintf(out, "  FAIL %s\n", wrap(p))
		}
		// A warning does not fail the gate: it is a condition that is correct
		// now and fatal at deploy time, and failing on it would leave the gate
		// red through the middle of every onboarding.
		for _, w := range r.Warnings {
			fmt.Fprintf(out, "  warn %s\n", wrap(w))
		}
		switch {
		case !r.OK():
			fmt.Fprintf(out, "  %d problem(s) - %s\n", len(r.Problems), r.Summary)
			ok = false
		case len(r.Warnings) > 0:
			// Not "ok": the check passed and something in it still needs doing,
			// and "ok 0 syncer(s)" under a warning about having none reads as a
			// contradiction.
			fmt.Fprintf(out, "  --   %s\n", r.Summary)
		default:
			fmt.Fprintf(out, "  ok   %s\n", r.Summary)
		}
	}
	return ok
}

// printTools lists every tool a render exposes, with its whole description.
//
// It is deliberately not a check. What makes a tooling.description wrong is
// that it does not separate itself from the tool beside it, which is a fact
// about the set - so the useful thing a program can do is put the set in front
// of somebody, in the order and with the context the model gets, and stop
// there. Truncating would defeat it: the sentence that disambiguates two tools
// is usually not the first one.
func printTools(out io.Writer, label string, docs []gate.Doc) {
	tools := gate.Tools(docs)
	fmt.Fprintf(out, "%s\n", label)
	if len(tools) == 0 {
		fmt.Fprintf(out, "  no entry in this render is exposed as a tool.\n\n")
		return
	}
	for _, t := range tools {
		fmt.Fprintf(out, "\n  %s  (%s, entry %s)\n", t.Name, t.Workflow, t.Entry)
		for _, line := range strings.Split(strings.TrimRight(t.Description, "\n"), "\n") {
			fmt.Fprintf(out, "      %s\n", line)
		}
	}
	fmt.Fprintf(out, "\n  %d tool(s). Read them together, as the model does - one list, no other\n"+
		"  context, deciding which answers the question. The failure this catches is\n"+
		"  two descriptions that are each accurate and do not say which to prefer.\n\n", len(tools))
}
