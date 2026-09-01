package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Workspace: config.Workspace{ID: "ws_1", Slug: "unitech-e", Name: "UnitechE"},
		Projects: []config.Project{
			{Slug: "internal", Name: "internal", Environments: []config.Env{config.EnvDev, config.EnvProd}},
			{Slug: "website", Name: "official site", Environments: []config.Env{config.EnvDev}},
		},
	}
}

// written returns the set of paths a Write produced, and the results by path.
func written(t *testing.T, results []Result) map[string]Status {
	t.Helper()
	byPath := make(map[string]Status, len(results))
	for _, r := range results {
		if _, dup := byPath[r.Path]; dup {
			t.Errorf("path %q written twice", r.Path)
		}
		byPath[r.Path] = r.Status
	}
	return byPath
}

func TestWriteProducesTheGateItMustPass(t *testing.T) {
	dir := t.TempDir()

	results, err := Write(dir, testConfig(), false)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	paths := written(t, results)

	// `asgard-cli check` requires every one of these to exist. If this list
	// shrinks, the skeleton stops passing its own gate - which
	// internal/check's TestScaffoldedRepoPasses would also catch.
	required := []string{
		"docs/README.md",
		"docs/spec-driven-development.md",
		"docs/spec/README.md",
		"docs/spec/unitech-e-asgard/README.md",
		"docs/decisions/README.md",
		"docs/decisions/_decision-template.md",
		"docs/meeting-notes/README.md",
		"docs/meeting-notes/_template.md",
		"requirements/_index.md",
		"requirements/requests/_index.md",
		"requirements/tasks/_index.md",
		"README.md",
		"AGENTS.md",
		"CLAUDE.md",
		"scripts/check_crd_fidelity.py",
		"scripts/db/requirements.txt",
		".agents/skills/spec-workflow/SKILL.md",
		".agents/skills/semantic-layer-modeling/SKILL.md",
		".agents/skills/asgard-cr-verification/SKILL.md",
		".agents/skills/asgard-fde-onboarding/SKILL.md",
		".github/workflows/main.yaml",
	}
	for _, want := range required {
		if _, ok := paths[want]; !ok {
			t.Errorf("%s was not written", want)
		}
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("stat %s: %v", want, err)
		}
	}
}

func TestWriteExpandsProjectsAndEnvironments(t *testing.T) {
	dir := t.TempDir()

	results, err := Write(dir, testConfig(), false)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	paths := written(t, results)

	for _, want := range []string{
		"projects/internal/deploy.yaml",
		"projects/internal/chart/app/Chart.yaml",
		"projects/internal/chart/app/values.yaml",
		"projects/internal/chart/app/templates/_helpers.tpl",
		"projects/internal/chart/values-dev.yaml",
		"projects/internal/chart/values-prod.yaml",
		"projects/website/deploy.yaml",
		"projects/website/chart/values-dev.yaml",
	} {
		if _, ok := paths[want]; !ok {
			t.Errorf("%s was not written", want)
		}
	}

	// website declares dev only, so a prod values file would be a file nobody
	// can explain, and deploy.yaml would not reference it.
	if _, ok := paths["projects/website/chart/values-prod.yaml"]; ok {
		t.Error("values-prod.yaml written for a project that declares dev only")
	}
}

func TestWriteRendersTheWorkspaceIn(t *testing.T) {
	dir := t.TempDir()

	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	tests := []struct {
		path string
		want []string
	}{
		{"README.md", []string{"unitech-e-asgard-kube", "UnitechE", "asgard-unitech-e-<slug>-<env>"}},
		{"AGENTS.md", []string{
			"unitech-e-asgard-kube",
			"docs/spec/unitech-e-asgard/",
			// The agent's entry point has to say the CLI exists, or the context
			// provider is never reached.
			"asgard-cli next",
			"asgard-cli usecase",
			"asgard-cli project add",
			"asgard-cli scaffold",
			// The three reversals are the reason this file is worth shipping.
			"The three decisions that get answered wrong",
			"Answered wrong once",
		}},
		{"projects/internal/deploy.yaml", []string{
			"asgard-unitech-e-internal-dev",
			"asgard-unitech-e-internal-prod",
			"chart/values-dev.yaml",
		}},
		{"projects/website/deploy.yaml", []string{"asgard-unitech-e-website-dev"}},
		{"projects/internal/chart/app/templates/_helpers.tpl", []string{
			`define "internal.name"`,
			`define "internal.workflowSetLabels"`,
		}},
		{"projects/internal/chart/app/Chart.yaml", []string{"name: internal"}},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, tt.path))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			for _, want := range tt.want {
				if !strings.Contains(string(data), want) {
					t.Errorf("%s does not contain %q", tt.path, want)
				}
			}
		})
	}
}

// TestWriteLeavesNoPlaceholders is the check that catches a template variable
// that was never substituted, which would otherwise ship as literal text.
func TestWriteLeavesNoPlaceholders(t *testing.T) {
	dir := t.TempDir()

	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	markers := []string{specSlugDir, projectDir, envDir, leftDelim + ".", "<<if", "<<range", "<<end>>"}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		for _, marker := range markers {
			if strings.Contains(rel, marker) {
				t.Errorf("path %s still contains %q", rel, marker)
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, marker := range markers {
			if strings.Contains(string(data), marker) {
				t.Errorf("%s still contains %q", rel, marker)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
}

// TestWriteCarriesNoCustomerContent guards the generalisation work: the
// templates came from one customer's repo, and anything left behind would show
// up in every other customer's repo.
func TestWriteCarriesNoCustomerContent(t *testing.T) {
	dir := t.TempDir()

	cfg := testConfig()
	cfg.Workspace = config.Workspace{ID: "ws_1", Slug: "acme", Name: "Acme"}
	cfg.Projects = []config.Project{{Slug: "app", Name: "app", Environments: []config.Env{config.EnvDev}}}
	if _, err := Write(dir, cfg, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Every reference deployment, not only the one the templates came from.
	//
	// A generated repo naming another customer's repository is the same leak as
	// naming their workspace id: customer A's repo should not tell them who
	// customers B and C are. This list caught `common/README.md` recommending
	// two other engagements' layouts, in a paragraph that also hardcoded
	// "exactly two projects" from the repo it was copied out of.
	leaked := []string{
		"unitech", "UnitechE", "台新", "sl-data-lake", "ag-website", "bp-website",
		"xxentria", "finance-ai", "freyr", "buy123", "auto-post",
		"industry-demo", "shopline",
	}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		lower := strings.ToLower(string(data))
		for _, name := range leaked {
			if strings.Contains(lower, strings.ToLower(name)) {
				t.Errorf("%s mentions %q, which belongs to the repo the templates came from", rel, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
}

func TestWriteIsIdempotent(t *testing.T) {
	dir := t.TempDir()

	first, err := Write(dir, testConfig(), false)
	if err != nil {
		t.Fatalf("first Write: %v", err)
	}

	// An edit to a generated file must survive a re-run: the skeleton is a
	// starting point, and AGENTS.md in particular is meant to be filled in.
	agents := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("edited by the engagement\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	second, err := Write(dir, testConfig(), false)
	if err != nil {
		t.Fatalf("second Write: %v", err)
	}
	if len(second) != len(first) {
		t.Errorf("second run planned %d files, first planned %d", len(second), len(first))
	}
	for _, r := range second {
		if r.Status != Skipped {
			t.Errorf("%s = %s on a re-run, want skipped", r.Path, r.Status)
		}
	}

	data, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "edited by the engagement\n" {
		t.Error("a re-run overwrote an edited file")
	}
}

func TestWriteForceOverwrites(t *testing.T) {
	dir := t.TempDir()

	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	agents := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("edited\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	results, err := Write(dir, testConfig(), true)
	if err != nil {
		t.Fatalf("force Write: %v", err)
	}
	for _, r := range results {
		if r.Status != Overwritten {
			t.Errorf("%s = %s with force, want overwritten", r.Path, r.Status)
		}
	}

	data, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) == "edited\n" {
		t.Error("--force did not overwrite")
	}
}

func TestWriteSetsExecutableBits(t *testing.T) {
	dir := t.TempDir()

	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// The gate scripts carry a shebang and are run directly.
	for _, path := range []string{"scripts/check_crd_fidelity.py", "scripts/db/query.py"} {
		info, err := os.Stat(filepath.Join(dir, path))
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Errorf("%s has mode %v, want the execute bit set", path, info.Mode().Perm())
		}
	}

	info, err := os.Stat(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("stat AGENTS.md: %v", err)
	}
	if info.Mode().Perm()&0o111 != 0 {
		t.Errorf("AGENTS.md has mode %v, want no execute bit", info.Mode().Perm())
	}
}

func TestWriteRejectsInvalidConfig(t *testing.T) {
	dir := t.TempDir()

	_, err := Write(dir, &config.Config{}, false)
	if err == nil {
		t.Fatal("Write with an empty config = nil, want a validation error")
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("ReadDir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("wrote %d entries despite an invalid config", len(entries))
	}
}

func TestSpecSlug(t *testing.T) {
	if got, want := SpecSlug("unitech-e"), "unitech-e-asgard"; got != want {
		t.Errorf("SpecSlug() = %q, want %q", got, want)
	}
}

// TestManagedRegionTracksTheConfig covers the bug this was written for: running
// scaffold before any project exists, then adding one, used to leave the README
// project table empty forever - and the gate compares that table against the
// directories on disk, so the repo failed its own check.
func TestManagedRegionTracksTheConfig(t *testing.T) {
	dir := t.TempDir()

	empty := &config.Config{Workspace: config.Workspace{ID: "ws_1", Slug: "acme", Name: "Acme"}}
	if _, err := Write(dir, empty, false); err != nil {
		t.Fatalf("first Write: %v", err)
	}

	readme := filepath.Join(dir, "README.md")
	before, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(before), "projects/app/") {
		t.Fatal("the table lists a project that does not exist yet")
	}

	// Something the engagement wrote, outside the managed region.
	edited := string(before) + "\n## Notes from the engagement\n\nkeep me\n"
	if err := os.WriteFile(readme, []byte(edited), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	withProject := &config.Config{
		Workspace: empty.Workspace,
		Projects:  []config.Project{{Slug: "app", Name: "app", Environments: []config.Env{config.EnvDev}}},
	}
	results, err := Write(dir, withProject, false)
	if err != nil {
		t.Fatalf("second Write: %v", err)
	}

	after, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(after), "`projects/app/`") {
		t.Errorf("the managed table was not refreshed:\n%s", after)
	}
	if !strings.Contains(string(after), "keep me") {
		t.Error("an edit outside the managed region was lost")
	}

	byPath := written(t, results)
	if got := byPath["README.md"]; got != Updated {
		t.Errorf("README.md = %s, want updated", got)
	}
	// A file with no managed region is still left completely alone.
	if got := byPath["AGENTS.md"]; got != Skipped {
		t.Errorf("AGENTS.md = %s, want skipped", got)
	}
}

func TestManagedRegionRespectsRemoval(t *testing.T) {
	dir := t.TempDir()
	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Taking the markers out is a deliberate act: the file becomes the
	// engagement's to maintain, and scaffold must stop touching it.
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# hand-written\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	results, err := Write(dir, testConfig(), false)
	if err != nil {
		t.Fatalf("second Write: %v", err)
	}
	if got := written(t, results)["README.md"]; got != Skipped {
		t.Errorf("README.md = %s, want skipped once the markers are gone", got)
	}

	data, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "# hand-written\n" {
		t.Errorf("file was modified: %q", data)
	}
}

// TestOnboardingSkillIsDesignTime guards the split that is easy to get wrong:
// .agents/skills is read by the coding agent from the working tree and never
// reaches the cluster, while common/skills is synced into the platform and
// bound to the deployed agent. Routing guidance belongs in the first; putting it
// in the second would attach "how to onboard a customer" to an agent whose job
// is answering that customer's questions.
func TestOnboardingSkillIsDesignTime(t *testing.T) {
	dir := t.TempDir()
	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	designTime := filepath.Join(dir, ".agents", "skills", "asgard-fde-onboarding", "SKILL.md")
	data, err := os.ReadFile(designTime)
	if err != nil {
		t.Fatalf("the skill should ship in .agents/skills: %v", err)
	}
	if !strings.Contains(string(data), "asgard-cli next") {
		t.Error("the skill must route to the CLI")
	}
	if !strings.Contains(string(data), "never synced into the platform") {
		t.Error("the skill must say which of the two kinds it is")
	}

	// The frontmatter name has to match the directory name, which
	// `asgard-cli check` enforces for runtime skills, and which the tools
	// reading a design-time tree rely on too.
	if !strings.Contains(string(data), "name: asgard-fde-onboarding\n") {
		t.Error("frontmatter name must match the directory name")
	}

	runtime := filepath.Join(dir, "common", "skills", "asgard-fde-onboarding")
	if _, err := os.Stat(runtime); !os.IsNotExist(err) {
		t.Error("a design-time skill must not be written into common/skills")
	}
}

// referencedPath matches a repository-relative path to a script or a data file
// in the generated documentation. Only common/ and scripts/ are checked: those
// hold the things a reader is told to *run*, so a stale one there is an
// instruction that fails rather than a broken link.
var referencedPath = regexp.MustCompile(`\b(?:common|scripts)/[A-Za-z0-9_./-]+\.(?:sh|py|yaml|yml|txt|md)\b`)

// TestGeneratedDocsOnlyNameFilesThatExist is a guard against a specific way this
// repo has already broken once: a template was deleted, the documentation that
// told people to run it was updated everywhere except one file, and the
// generated repo shipped instructions for `common/render.sh` months after
// `asgard-cli render` replaced it. Nothing else would have caught it - the file
// is prose, so no test read it and no gate compiled it.
func TestGeneratedDocsOnlyNameFilesThatExist(t *testing.T) {
	dir := t.TempDir()
	if _, err := Write(dir, testConfig(), false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(dir, path)
		for _, ref := range referencedPath.FindAllString(string(body), -1) {
			// A path with a placeholder in it is describing a shape, not
			// naming a file.
			if strings.ContainsAny(ref, "<>*") || strings.Contains(ref, "xxx") {
				continue
			}
			if _, err := os.Stat(filepath.Join(dir, ref)); err != nil {
				t.Errorf("%s names %s, which the scaffold does not write", rel, ref)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
}
