package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/binding"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/gate"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/platform"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/render"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/skills"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/tool"
)

// `asgard-cli gate` is the one command to run after changing anything.
//
// **It exists because a checklist in prose is not a gate.** Everything it runs
// was already here - `check`, `helm lint`, `render`, `verify`, and now the
// freshness of the reference material - and the only thing that assembled them
// was an AGENTS.md section, a stage prompt and whatever an agent happened to
// remember. Every one of those is a place the list can go stale, and one of
// them did: the scaffolded gate described four steps, the fourth of which ran a
// python script that had been deleted a month earlier.
//
// An agent working in a compiled language does not have this problem. Whatever
// it changed, it knows to run the build, and the build is one command whose
// definition lives with the code. This is that command.
//
// **What it deliberately does not do is reproduce the platform's checks.** A
// CR's admission is decided by an apiserver, and no client is ever given
// credentials for one; a second copy of those rules here would disagree with
// the server the first time either changed, while still missing the ones that
// matter most. The last line of output says so and names the plan.

type stepStatus string

const (
	stepPass stepStatus = "pass"
	stepFail stepStatus = "fail"
	stepWarn stepStatus = "warn"
	stepSkip stepStatus = "skip"
)

// stepResult is one step's outcome.
//
// A skip is not a pass, and the distinction is the whole reason this is a
// struct rather than a boolean: "kubectl was not installed so the cluster check
// did not run" was reported as green for long enough to reach a customer, in
// the gate this replaces.
type stepResult struct {
	Name    string     `json:"name"`
	Status  stepStatus `json:"status"`
	Summary string     `json:"summary"`
	Details []string   `json:"details,omitempty"`
	// Remedy is the command that fixes it, when one command does.
	Remedy string `json:"remedy,omitempty"`
}

func newGateCmd() *cobra.Command {
	var (
		format  string
		offline bool
		profile string
	)

	cmd := &cobra.Command{
		Use:   "gate [release ...]",
		Short: "Everything this machine can check, in one command",
		Long: `Everything this machine can check about this repository, in one command.

    asgard-cli gate                  every release the declaration names
    asgard-cli gate internal-dev     one of them
    asgard-cli gate --offline        skip the step that needs the platform
    asgard-cli gate --format json    one record per step, for an agent

**Run it after changing anything under a chart or ` + "`" + pipelineconfig.FileName + "`" + `.** It is the
build step of a repository that has no build step: an agent working in a
compiled language knows that whatever it changed, it runs the compiler, and the
compiler is one command whose definition lives with the code. Nothing here was
missing before - ` + "`check`" + `, ` + "`helm lint`" + `, ` + "`render`" + `, ` + "`verify`" + ` and the reference
material's freshness all existed - but the only thing that assembled them was
prose, and prose goes stale. The gate this replaces described four steps, and
the fourth ran a script that had been deleted a month earlier.

What it runs, in order:

  tools    helm is on PATH. Without it the three chart steps cannot run, and
           they are reported as skipped rather than passed
  repo     the structural invariants a chart render cannot see
  binding  whether .asgard-cli.yaml names a workspace and a pipeline that the
           platform still has. It is the step that catches a half-bound
           checkout - ` + "`workspace use`" + ` clears the pipeline line, and this goes
           red rather than waiting for whichever command somebody runs next.
           Platform facts only: no git remote is read here or anywhere else.
           Needs a session; --offline skips the half that asks
  skills   whether the reference material here still describes the server this
           repository deploys to. Needs a session; --offline skips it
  lint     helm lint on each chart, with the reserved asgard block and NOTHING
           else. That is what proves values.yaml declares a default for every
           .Values.* the chart itself owns; overlay an environment file and a
           missing default is masked until somebody runs plain helm template.
           Linting with no values file at all - the old instruction - fails on
           every chart that reads .Values.asgard.*, which a chart must not
           declare and the platform always injects
  render   each release renders, with placeholder platform values
  verify   the rendered CRs against each other: dangling references, both
           halves of every entrypoint, missing display annotations, the
           workflow-set labels, the agent-split invariants

**A skip is not a pass**, and the two are printed differently on purpose.

**Do not run helm by hand here.** The platform injects a reserved ` + "`asgard`" + ` block
into every render, and a chart must not declare it in its own values.yaml - so
` + "`helm lint <chart>`" + ` with no -f fails on every chart that reads
` + "`.Values.asgard.projectEnvironmentId`" + `, which is every chart that labels
anything. That failure looks like the chart is broken and it is not. This
supplies that one file and nothing else, which is why the lint step still
proves that values.yaml defaults everything the chart itself owns.

**It does not reproduce the platform's checks, and it must not.** Whether a CR
is admitted is decided by an apiserver, and no client is ever given credentials
for one - so a copy of those rules here would drift from the server the first
time either changed, while still missing the two that matter most: a field the
CRD silently prunes, and a rejection only the apiserver can produce. A green
gate means "worth pushing", never "this will deploy". The authority is the plan:

    asgard-cli pipeline runs watch --release <name> --commit $(git rev-parse HEAD)

Exits non-zero if any step failed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkFormat(format); err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			// The declaration is what anchors a Pipeline repository, and it is
			// the one file every one of them has.
			root, err := repoRoot()
			if err != nil {
				return err
			}
			releases, err := releasesToVerify(root, args)
			if err != nil {
				return err
			}

			steps := []stepResult{gateTools()}
			helmReady := steps[0].Status == stepPass

			steps = append(steps, gateRepo(root, args))
			steps = append(steps, gateBinding(cmd, root, profile, offline))
			steps = append(steps, gateSkills(cmd, profile, offline))
			steps = append(steps, gateCharts(cmd, root, releases, helmReady)...)

			if format == formatJSON {
				if err := writeJSON(out, map[string]any{
					"ok":    stepsOK(steps),
					"steps": steps,
				}); err != nil {
					return err
				}
				if !stepsOK(steps) {
					return ErrSilent
				}
				return nil
			}
			printSteps(out, steps)
			if !stepsOK(steps) {
				return ErrSilent
			}
			return nil
		},
	}

	addProfileFlag(cmd, &profile)
	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	cmd.Flags().BoolVar(&offline, "offline", false, "skip the two steps that need the platform: binding and skills")
	return cmd
}

func stepsOK(steps []stepResult) bool {
	for _, s := range steps {
		if s.Status == stepFail {
			return false
		}
	}
	return true
}

func printSteps(out io.Writer, steps []stepResult) {
	failed, skipped := 0, 0
	for _, s := range steps {
		fmt.Fprintf(out, "%-9s %-5s %s\n", s.Name, s.Status, s.Summary)
		for _, d := range s.Details {
			fmt.Fprintf(out, "  %s\n", wrapAt(d, 76, 2))
		}
		if s.Remedy != "" {
			fmt.Fprintf(out, "  -> %s\n", s.Remedy)
		}
		switch s.Status {
		case stepFail:
			failed++
		case stepSkip:
			skipped++
		}
	}

	fmt.Fprintln(out)
	switch {
	case failed > 0:
		fmt.Fprintf(out, "%d of %d step(s) failed.\n", failed, len(steps))
	case skipped > 0:
		fmt.Fprintf(out, "Nothing failed, and %d step(s) did not run. A skip is not a pass.\n", skipped)
	default:
		fmt.Fprintf(out, "All %d steps passed.\n", len(steps))
	}
	// Said every time, because a green local gate is the moment somebody is
	// most likely to believe the work is finished.
	fmt.Fprintf(out, "This is what a machine with no cluster can check. The plan checks the rest:\n"+
		"    asgard-cli pipeline runs watch --release <name> --commit $(git rev-parse HEAD)\n")
}

// ── the steps ────────────────────────────────────────────────────────────

func gateTools() stepResult {
	if _, err := tool.Helm.Path(); err != nil {
		var missing *tool.ErrMissing
		remedy := "asgard-cli doctor"
		if errors.As(err, &missing) {
			remedy = missing.Tool.InstallHint()
		}
		return stepResult{
			Name:    "tools",
			Status:  stepFail,
			Summary: "helm is not on PATH, so nothing that touches a chart can run",
			Remedy:  remedy,
		}
	}
	return stepResult{Name: "tools", Status: stepPass, Summary: "helm is on PATH"}
}

func gateRepo(root string, only []string) stepResult {
	// AGENTS.md is what `scaffold` always writes, and it is the marker `stage`
	// already uses for "this repository has a skeleton". Without one there is
	// no skeleton to hold the repository against - the docs layers, the README
	// project table, the requirements indexes are all things `scaffold` writes
	// - and reporting their absence as six failures would be reporting that a
	// repository is not something it never claimed to be.
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		return stepResult{
			Name:    "repo",
			Status:  stepSkip,
			Summary: "no AGENTS.md, so this repository has no skeleton to check",
			Remedy:  "asgard-cli scaffold",
		}
	}
	report, err := check.Run(root, only...)
	if err != nil {
		return stepResult{Name: "repo", Status: stepFail, Summary: err.Error()}
	}
	res := stepResult{Name: "repo", Status: stepPass}
	for _, w := range report.Warnings() {
		res.Details = append(res.Details, "warn  "+w.Message)
	}
	errs := report.Errors()
	for _, e := range errs {
		res.Details = append(res.Details, "FAIL  "+e.Message)
	}
	switch {
	case len(errs) > 0:
		res.Status = stepFail
		res.Summary = fmt.Sprintf("%d problem(s) in the repository's structure", len(errs))
	case len(report.Warnings()) > 0:
		res.Status = stepWarn
		res.Summary = fmt.Sprintf("%d warning(s)", len(report.Warnings()))
	default:
		res.Summary = "structure and declaration are consistent"
	}
	return res
}

// gateBinding checks that this checkout names a workspace and a pipeline the
// platform still has.
//
// **It is here because a half-bound checkout has no other symptom.**
// `workspace use <other>` clears the pipeline line, deliberately - a pipeline
// belongs to one workspace - and if the agent that ran it does not go on to
// `pipeline use`, nothing is wrong until somebody runs a pipeline command,
// which might be `runs approve`. This is a command an agent already runs after
// every change, so the half state surfaces at the next edit instead.
//
// **It reads platform facts and nothing else.** It does not compare the
// pipeline's repository to a git remote: a checkout may have several remotes,
// and which one is called `origin` is not this tool's business. The gap that
// leaves - a repository copied wholesale within one workspace - is named in the
// binding file's own header.
func gateBinding(cmd *cobra.Command, root, profile string, offline bool) stepResult {
	res := stepResult{Name: "binding"}

	f, err := binding.LoadFrom(root)
	switch {
	case errors.Is(err, binding.ErrNotFound):
		// Nothing recorded is the ordinary state of a repository nobody has
		// bound yet, and the same reasoning as the repo step applies: this is
		// not a repository failing to be something it never claimed to be.
		res.Status = stepSkip
		res.Summary = "no " + binding.FileName + ", so this checkout is not bound to a pipeline yet"
		res.Remedy = "asgard-cli init"
		return res
	case err != nil:
		res.Status = stepFail
		res.Summary = err.Error()
		return res
	}

	// The local half runs offline and runs first: a missing field is a fact
	// about the file, and asking the platform about it would be asking the
	// wrong question.
	if missing := f.Missing(); len(missing) > 0 {
		res.Status = stepFail
		res.Summary = fmt.Sprintf("%s records no %s", binding.FileName, strings.Join(missing, " and no "))
		if len(missing) == 1 && missing[0] == "pipeline" {
			res.Details = append(res.Details,
				"a workspace is recorded and a pipeline is not, which is what `asgard-cli workspace use` leaves behind")
			res.Remedy = "asgard-cli pipeline list, then asgard-cli pipeline use <id>"
			return res
		}
		res.Remedy = "asgard-cli workspace use <id>, then asgard-cli pipeline use <id>"
		return res
	}

	if offline {
		res.Status = stepSkip
		res.Summary = "--offline, so the platform was not asked whether these still exist"
		return res
	}

	pc, err := resolveContext(cmd, contextOptions{Profile: profile})
	if err != nil {
		res.Status = stepSkip
		res.Summary = "no session, so the platform was not asked (`asgard-cli login`, or --offline to say so on purpose)"
		return res
	}

	// The recorded workspace, not the resolved one. --workspace and
	// ASGARD_WORKSPACE outrank the file everywhere else on purpose, and here
	// the file is the thing being checked.
	workspaces, err := platform.New(pc.Session, "").ListWorkspaces(cmd.Context())
	if err != nil {
		res.Status = stepSkip
		res.Summary = fmt.Sprintf("the platform did not answer: %v", err)
		return res
	}
	wsName := ""
	found := false
	for _, w := range workspaces {
		if w.ID == f.Workspace {
			wsName, found = w.Name, true
			break
		}
	}
	if !found {
		res.Status = stepFail
		res.Summary = fmt.Sprintf("workspace %s is not one this account can reach on %s", f.Workspace, pc.Session.Profile.Name)
		for _, w := range workspaces {
			res.Details = append(res.Details, fmt.Sprintf("      %-22s %s", w.ID, w.Name))
		}
		res.Remedy = "asgard-cli workspace list"
		return res
	}

	pipelines, err := platform.New(pc.Session, f.Workspace).ListPipelines(cmd.Context())
	if err != nil {
		res.Status = stepFail
		res.Summary = fmt.Sprintf("workspace %s exists, and its pipelines could not be listed: %v", f.Workspace, err)
		return res
	}
	for _, p := range pipelines {
		if p.PipelineId == f.Pipeline {
			res.Status = stepPass
			res.Summary = fmt.Sprintf("%s -> %s (%s), pipeline %s", binding.FileName, f.Workspace, wsName, p.Name)
			return res
		}
	}

	res.Status = stepFail
	res.Summary = fmt.Sprintf("pipeline %s is not in workspace %s (%s)", f.Pipeline, f.Workspace, wsName)
	for _, p := range pipelines {
		res.Details = append(res.Details, fmt.Sprintf("      %-22s %-20s %s", p.PipelineId, p.Name, p.RepoFullName))
	}
	res.Remedy = "asgard-cli pipeline use <id>"
	return res
}

// gateSkills asks whether the reference material in this repository still
// describes the server it deploys to.
//
// **It is a step of the gate rather than a reminder**, because material going
// stale has no symptom: the CR an agent writes against it is wrong in a way
// that reads perfectly, applies cleanly, and fails at run time. That is the
// same class as the checks below it, and it belongs in the same command.
//
// Not having a session is a skip, not a failure. A gate that only works online
// is a gate that fails on a plane, and every other step here works offline.
func gateSkills(cmd *cobra.Command, profile string, offline bool) stepResult {
	res := stepResult{Name: "skills", Remedy: "asgard-cli skill update"}
	if offline {
		res.Status = stepSkip
		res.Summary = "--offline, so the platform was not asked"
		res.Remedy = ""
		return res
	}

	root, _, err := skillRoot(cmd, "")
	if err != nil {
		res.Status = stepSkip
		res.Summary = err.Error()
		return res
	}
	stamp, err := skills.ReadStamp(root)
	if err != nil {
		res.Status = stepFail
		res.Summary = err.Error()
		return res
	}

	pc, err := resolveContext(cmd, contextOptions{Profile: profile})
	if err != nil {
		// Not logged in is the ordinary state of a CI runner and of an agent
		// sandbox, and neither is a reason to call the repository broken.
		res.Status = stepSkip
		res.Summary = "no session, so the platform was not asked (`asgard-cli login`, or --offline to say so on purpose)"
		res.Remedy = ""
		return res
	}
	remote, err := pc.Client.DocsVersionOnly(cmd.Context())
	if err != nil {
		res.Status = stepSkip
		res.Summary = fmt.Sprintf("the platform did not answer: %v", err)
		res.Remedy = ""
		return res
	}

	if stamp == nil {
		res.Status = stepFail
		res.Summary = fmt.Sprintf("nothing fetched here, so an agent has no statement of what %s accepts", pc.Session.Profile.Name)
		return res
	}
	if stamp.Version != remote.Version {
		res.Status = stepFail
		if skills.Behind(stamp.Version, remote.Version) {
			res.Summary = fmt.Sprintf("behind: the platform has published version %s and this repository holds %s",
				remote.Version, stamp.Version)
			return res
		}
		res.Summary = fmt.Sprintf("this repository holds version %s and %s serves %s",
			stamp.Version, pc.Session.Profile.Name, remote.Version)
		return res
	}
	if moved := skills.MovedSources(stamp, sourceDigests(remote.Sources)); len(moved) > 0 {
		res.Status = stepFail
		res.Summary = fmt.Sprintf("same version, different material: %s moved since this was fetched", strings.Join(moved, ", "))
		return res
	}

	res.Status = stepPass
	res.Summary = fmt.Sprintf("version %s, matching %s", stamp.Version, pc.Session.Profile.Name)
	res.Remedy = ""
	return res
}

// gateCharts runs lint, render and verify over the releases, and reports them
// as three steps rather than one: they fail for different reasons and are fixed
// in different places, and a single "charts" line would hide which.
func gateCharts(cmd *cobra.Command, root string, releases []string, helmReady bool) []stepResult {
	lint := stepResult{Name: "lint", Status: stepPass}
	rendered := stepResult{Name: "render", Status: stepPass}
	verified := stepResult{Name: "verify", Status: stepPass}

	if !helmReady {
		for _, s := range []*stepResult{&lint, &rendered, &verified} {
			s.Status = stepSkip
			s.Summary = "helm is not on PATH"
		}
		return []stepResult{lint, rendered, verified}
	}
	if len(releases) == 0 {
		for _, s := range []*stepResult{&lint, &rendered, &verified} {
			s.Status = stepSkip
			s.Summary = pipelineconfig.FileName + " declares no releases yet"
		}
		return []stepResult{lint, rendered, verified}
	}

	decl, err := pipelineconfig.LoadFromRepo(root, "")
	if err != nil {
		lint.Status, lint.Summary = stepFail, err.Error()
		return []stepResult{lint, rendered, verified}
	}

	charts, renders, resources, problems := 0, 0, 0, 0
	warnings := &warningSet{}
	for _, name := range releases {
		release, ok := decl.Release(name)
		if !ok {
			continue
		}

		// With the reserved `asgard` block and NOTHING else.
		//
		// The instruction this replaces said to lint bare, with no values file
		// at all, because that is the only form that proves values.yaml
		// declares a default for every `.Values.*` a template reads - overlay
		// an environment file and a missing default is masked until somebody
		// runs plain `helm template`, at which point it is a nil pointer in
		// somebody else's terminal.
		//
		// **That instruction stopped being right when the platform started
		// injecting values.** A chart must not declare the `asgard` block in
		// its own values.yaml - the platform overwrites it, and declaring it is
		// a warning on every plan - so a bare lint fails on every chart that
		// reads `.Values.asgard.projectEnvironmentId`, which is every chart
		// that labels anything. Verified on a real repository: four of four
		// charts, all for that reason and no other.
		//
		// Supplying that one file and no other keeps the property intact for
		// everything the chart does own.
		chart := filepath.Join(root, filepath.FromSlash(release.Chart))
		if output, err := runHelmLint(cmd, chart, name); err != nil {
			lint.Status = stepFail
			lint.Details = append(lint.Details, fmt.Sprintf("FAIL  %s: %s", name, output))
		} else {
			charts++
		}

		var buf bytes.Buffer
		if _, err := render.Run(cmd.Context(), render.Options{Root: root, Release: name}, &buf, io.Discard); err != nil {
			rendered.Status = stepFail
			rendered.Details = append(rendered.Details, fmt.Sprintf("FAIL  %s: %v", name, err))
			continue
		}
		renders++

		docs, err := gate.Read(&buf)
		if err != nil {
			verified.Status = stepFail
			verified.Details = append(verified.Details, fmt.Sprintf("FAIL  %s: %v", name, err))
			continue
		}
		resources += len(docs)

		for _, r := range gates(docs, gate.Options{Project: projectOfRelease(root, name)}) {
			for _, p := range r.Problems {
				verified.Status = stepFail
				verified.Details = append(verified.Details, fmt.Sprintf("FAIL  %s: %s", name, p))
				problems++
			}
			// A warning is truncated to one line and a failure is not.
			// Several of these rules explain themselves in a paragraph -
			// correctly, because the judgement is the point - and a command
			// somebody runs after every edit cannot print six paragraphs of
			// advice about something that is not blocking them. The summary
			// says where the full text is.
			for _, w := range r.Warnings {
				warnings.add(w, name)
			}
		}
	}

	lint.Summary = fmt.Sprintf("%d of %d chart(s) declare a default for every value they read", charts, len(releases))
	rendered.Summary = fmt.Sprintf("%d of %d release(s) rendered", renders, len(releases))

	lines := warnings.lines()

	switch {
	case verified.Status == stepFail:
		verified.Details = append(verified.Details, lines...)
		verified.Summary = fmt.Sprintf("%d problem(s) across %d resource(s)", problems, resources)
	case renders == 0:
		// Nothing rendered, so nothing was checked. Reporting that as a pass
		// or as a warning would both be claims about resources that do not
		// exist.
		verified.Status = stepSkip
		verified.Summary = "nothing rendered, so nothing was checked"
	case len(lines) > 0:
		verified.Status = stepWarn
		verified.Details = lines
		verified.Summary = fmt.Sprintf("%d resource(s), %d warning(s); `asgard-cli verify` prints them in full", resources, len(lines))
	default:
		verified.Summary = fmt.Sprintf("%d resource(s) reference each other correctly", resources)
	}
	return []stepResult{lint, rendered, verified}
}

// warningSet collapses the same warning raised against several releases.
//
// Two releases of one chart produce identical findings, and a repository with
// four releases printed the same six sentences twice. The finding is one fact
// about one chart; which releases it reached is a suffix, not a second copy.
type warningSet struct {
	order    []string
	releases map[string][]string
}

func (w *warningSet) add(text, release string) {
	if w.releases == nil {
		w.releases = map[string][]string{}
	}
	if _, seen := w.releases[text]; !seen {
		w.order = append(w.order, text)
	}
	if release != "" && !slices.Contains(w.releases[text], release) {
		w.releases[text] = append(w.releases[text], release)
	}
}

func (w *warningSet) lines() []string {
	out := make([]string, 0, len(w.order))
	for _, text := range w.order {
		line := "warn  " + truncate(text, 110)
		if where := w.releases[text]; len(where) > 0 {
			line += "  [" + strings.Join(where, ", ") + "]"
		}
		out = append(out, line)
	}
	return out
}

// runHelmLint returns helm's own output on failure, because that output names
// the template and the line and this cannot improve on it.
func runHelmLint(cmd *cobra.Command, chart, release string) (string, error) {
	if _, err := os.Stat(filepath.Join(chart, "Chart.yaml")); err != nil {
		return "no Chart.yaml in " + chart, err
	}
	values, err := render.AsgardValuesFile(release, "")
	if err != nil {
		return err.Error(), err
	}
	defer os.Remove(values)

	helm, err := tool.Helm.Command(cmd.Context(), "lint", chart, "-f", values)
	if err != nil {
		return err.Error(), err
	}
	var buf bytes.Buffer
	helm.Stdout = &buf
	helm.Stderr = &buf
	if err := helm.Run(); err != nil {
		return firstProblem(buf.String()), err
	}
	return "", nil
}

// firstProblem picks what helm meant out of the banner it prints around it.
//
// helm splits one failure across three lines - the file and position, then the
// expression, then the reason - and the reason is the only one that says what
// is wrong. Taking the first line alone reported "templates/: <file>:28:52",
// which names a location and no problem.
func firstProblem(output string) string {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "[ERROR]") && !strings.HasPrefix(line, "Error:") {
			continue
		}
		if strings.HasPrefix(line, "Error:") && strings.Contains(line, "chart(s) failed") {
			// helm's own tally, printed after the real finding.
			continue
		}
		parts := []string{line}
		for _, next := range lines[i+1:] {
			next = strings.TrimSpace(next)
			if next == "" || strings.HasPrefix(next, "[") || strings.HasPrefix(next, "Error:") {
				break
			}
			parts = append(parts, next)
		}
		return strings.Join(parts, " ")
	}
	return strings.TrimSpace(output)
}
