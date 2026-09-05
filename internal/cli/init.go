package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/auth"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/binding"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
)

// `asgard-cli init` is the composition, and the composition is the point.
//
// **It is `gate` for the beginning of an engagement.** Onboarding a repository
// has always been three commands - write the skeleton, record which platform
// objects it deploys through, fetch the material describing the server - and
// the list of three lived in prose, which is where a list goes to rot. The one
// this replaces wrote `.asgard-config.json`, a file that recorded four things
// the tool could have derived or asked for, and it is gone with that file.
//
// **It chooses nothing.** Both ids have to be given or already recorded; with
// neither, it lists what there is and stops, and it does that for a list of one
// exactly as for a list of five. That is one step of friction for a customer
// with a single workspace, and it is bought deliberately: a rule that resolves
// while there is one candidate stops resolving - or resolves elsewhere - on the
// day there are two, and nobody is watching that day.
func newInitCmd() *cobra.Command {
	var (
		workspaceID string
		pipelineID  string
		force       bool
		skipSkills  bool
		profile     string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Onboard a repository: skeleton, binding, and the platform's reference material",
		Long: `Onboard a repository, in one command.

    asgard-cli workspace list
    asgard-cli pipeline list --workspace <id>
    asgard-cli init --workspace <id> --pipeline <id>

It runs the three things a new customer repository needs, in the only order
they work in:

  1  ` + "`asgard-cli scaffold`" + ` writes the skeleton, including ` + "`" + pipelineconfig.FileName + "`" + `
  2  the binding is recorded in ` + "`" + binding.FileName + "`" + ` beside that declaration
  3  ` + "`asgard-cli skill update`" + ` fetches what this platform accepts

Each is still its own command and each can be re-run alone. This exists because
the list of three was prose, and prose goes stale: the onboarding instructions
this replaces named a step that ran a script deleted a month earlier.

**Nothing is guessed.** --workspace and --pipeline are required the first time,
and the errors list the candidates. That holds even when an account can reach
exactly one workspace or a workspace holds exactly one pipeline: a list of one
is still a list, and a rule that only works while it is short changes behaviour
silently the day it grows. On a repository that already records both, they are
read from ` + "`" + binding.FileName + "`" + ` and neither flag is needed - so re-running this after
an upgrade is a safe way to bring a repository up to date.

**The pipeline is not created here.** Creating one binds a repository on the
provider, needs a VCS connection, and is a decision about the platform rather
than about this checkout - ` + "`asgard-cli pipeline create`" + ` does it, and this records
the result. Run it in the directory that is to become the repository.

Existing files are left alone; --force overwrites the skeleton, which discards
local edits to it. --skip-skills leaves step 3 out, for a machine with no
session that is expected to run ` + "`asgard-cli skill update`" + ` later - a skip, not a
pass, and the output says so.

When it finishes, commit what it wrote and run ` + "`asgard-cli gate`" + `.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()

			root, err := scaffoldRoot()
			if err != nil {
				return err
			}

			// Both ids are settled before anything is written. A half-done
			// onboarding that stopped to ask a question is worse than one that
			// has not started: the skeleton is committed either way, and the
			// binding beside it would be missing without saying so.
			t, err := resolveInitTargets(cmd, root, profile, workspaceID, pipelineID)
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "%-11s %s  (%s)\n", "workspace", t.Workspace.ID, t.Workspace.Name)
			fmt.Fprintf(out, "%-11s %s  (%s -> %s)\n", "pipeline", t.Pipeline.PipelineId, t.Pipeline.Name, t.Pipeline.RepoFullName)
			fmt.Fprintf(out, "%-11s %s\n\n", "directory", root)

			fmt.Fprintf(out, "1/3 skeleton\n")
			if err := runScaffold(cmd, root, repo.Workspace{Name: t.Workspace.Name}, force, false); err != nil {
				return err
			}

			fmt.Fprintf(out, "\n2/3 binding\n")
			rel, err := writeInitBinding(root, t.Workspace.ID, t.Pipeline.PipelineId)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "  wrote        %s\n", rel)

			if skipSkills {
				fmt.Fprintf(out, "\n3/3 reference material  SKIPPED (--skip-skills)\n")
				fmt.Fprintf(out, "  An agent working here has no statement of what this server accepts.\n")
				fmt.Fprintf(out, "  A skip is not a pass:\n\n      asgard-cli skill update\n")
			} else {
				fmt.Fprintf(out, "\n3/3 reference material\n")
				if err := runSkillUpdate(cmd, skillUpdateOptions{Profile: profile}); err != nil {
					return err
				}
			}

			fmt.Fprintf(out, `
Commit all of it, %s included: whoever clones this repository, and whatever
agent works in it, then needs no --workspace and no --pipeline.

Then, after changing anything under a chart or the declaration:

    asgard-cli gate
`, binding.FileName)
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&workspaceID, "workspace", "", "workspace id; \"asgard-cli workspace list\" shows them")
	cmd.Flags().StringVar(&pipelineID, "pipeline", "", "pipeline id or name in that workspace; \"asgard-cli pipeline list\" shows them")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite skeleton files that already exist")
	cmd.Flags().BoolVar(&skipSkills, "skip-skills", false, "do not fetch the platform's reference material; report it as skipped")
	return cmd
}

// resolveInitTargets settles which workspace and which pipeline, or reports
// what the choices are.
//
// The order is flag, then environment, then what the checkout already records.
// **The machine-wide default is deliberately not a source here.** Everywhere
// else it decides one command's behaviour and is gone afterwards; this writes a
// committed file, and a per-machine setting becoming a repository's recorded
// choice is precisely the drift the binding file exists to avoid.
func resolveInitTargets(
	cmd *cobra.Command,
	root, profile, wantWorkspace, wantPipeline string,
) (initTargets, error) {
	var none initTargets

	pc, err := resolveContext(cmd, contextOptions{Profile: profile})
	if err != nil {
		return none, err
	}

	workspaces, err := pc.Client.ListWorkspaces(cmd.Context())
	if err != nil {
		return none, err
	}

	id := wantWorkspace
	if id == "" {
		id = os.Getenv(auth.EnvWorkspace)
	}
	if id == "" {
		if f, err := binding.LoadFrom(root); err == nil {
			id = f.Workspace
		}
	}
	if id == "" {
		return none, initChoiceError(
			fmt.Sprintf("--workspace is required, and this account can reach %s on %s:",
				plural(len(workspaces), "workspace"), pc.Session.Profile.Name),
			workspaceRows(workspaces),
			"asgard-cli init --workspace <id> --pipeline <id>",
		)
	}

	var ws platform.Workspace
	for _, w := range workspaces {
		if w.ID == id {
			ws = w
			break
		}
	}
	if ws.ID == "" {
		return none, initChoiceError(
			fmt.Sprintf("workspace %s is not one this account can reach on %s. It can reach %s:",
				id, pc.Session.Profile.Name, plural(len(workspaces), "workspace")),
			workspaceRows(workspaces),
			"asgard-cli workspace list",
		)
	}

	client := platform.New(pc.Session, ws.ID)
	pipelines, err := client.ListPipelines(cmd.Context())
	if err != nil {
		return none, err
	}

	want := wantPipeline
	if want == "" {
		// Only from a binding whose workspace is the one being used. A
		// pipeline id recorded against a different workspace says nothing
		// about this one, which is why `workspace use` clears the line.
		if f, err := binding.LoadFrom(root); err == nil && f.Workspace == ws.ID {
			want = f.Pipeline
		}
	}
	if want == "" {
		if len(pipelines) == 0 {
			return none, fmt.Errorf(
				"--pipeline is required, and workspace %s (%s) has none.\n\n"+
					"    asgard-cli pipeline connections\n"+
					"    asgard-cli pipeline create --name <name> --connection <id> --repo <owner/name>",
				ws.ID, ws.Name)
		}
		return none, initChoiceError(
			fmt.Sprintf("--pipeline is required, and workspace %s (%s) has %s:",
				ws.ID, ws.Name, plural(len(pipelines), "pipeline")),
			pipelineRows(pipelines),
			fmt.Sprintf("asgard-cli init --workspace %s --pipeline <id>", ws.ID),
		)
	}

	for _, p := range pipelines {
		if p.PipelineId == want || p.Name == want {
			return initTargets{Workspace: ws, Pipeline: p}, nil
		}
	}
	return none, initChoiceError(
		fmt.Sprintf("no pipeline %q in workspace %s (%s), which has %s:",
			want, ws.ID, ws.Name, plural(len(pipelines), "pipeline")),
		pipelineRows(pipelines),
		"asgard-cli pipeline list --workspace "+ws.ID,
	)
}

// initTargets is what `init` had to be told before it writes anything.
type initTargets struct {
	Workspace platform.Workspace
	Pipeline  *platform.Pipeline
}

// initChoiceError is the shape every "which of these?" here takes: the
// question, the candidates, and the command that answers it. Listing the
// candidates is what keeps refusing to guess from being an obstruction.
func initChoiceError(question string, rows []string, remedy string) error {
	var b strings.Builder
	b.WriteString(question)
	b.WriteString("\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "  %s\n", r)
	}
	fmt.Fprintf(&b, "\n    %s\n", remedy)
	fmt.Fprintf(&b, "\nNone is assumed, and that includes a list of one.\n")
	return fmt.Errorf("%s", b.String())
}

func workspaceRows(workspaces []platform.Workspace) []string {
	out := make([]string, 0, len(workspaces))
	for _, w := range workspaces {
		out = append(out, fmt.Sprintf("%-22s %s", w.ID, w.Name))
	}
	return out
}

func pipelineRows(pipelines []*platform.Pipeline) []string {
	out := make([]string, 0, len(pipelines))
	for _, p := range pipelines {
		out = append(out, fmt.Sprintf("%-22s %-20s %-52s %s", p.PipelineId, p.Name, p.RepoFullName, p.ConfigPath))
	}
	return out
}

// writeInitBinding records both ids beside the declaration the scaffold has
// just written, and reports the path relative to the repository.
func writeInitBinding(root, workspaceID, pipelineID string) (string, error) {
	declPath, bindPath, err := binding.Locate(root)
	if err != nil {
		return "", err
	}
	if declPath == "" {
		// The scaffold writes the declaration, so reaching here means it did
		// not - and a binding with nothing to belong to is not worth writing.
		return "", fmt.Errorf("no %s in %s after scaffolding, so there is nothing for a binding to belong to",
			pipelineconfig.FileName, root)
	}

	f := &binding.File{Version: 1}
	if existing, err := binding.Load(bindPath); err == nil {
		f = existing
	}
	f.Workspace, f.Pipeline = workspaceID, pipelineID
	if err := binding.Save(bindPath, f); err != nil {
		return "", err
	}
	if rel, err := filepath.Rel(root, bindPath); err == nil {
		return rel, nil
	}
	return bindPath, nil
}
