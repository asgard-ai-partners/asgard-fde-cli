// Package check verifies the structural invariants of a customer repository -
// the ones a chart render cannot see, and that only surface at deploy time or
// when the next person tries to pick the repo up.
//
// It replaces the check_repo_consistency.py the layout used to carry, so that
// the first gate needs no Python environment and every repo gets the same
// version of the rules.
package check

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

// Level separates a problem that fails the gate from one that is only worth
// mentioning.
type Level int

const (
	// Error fails the check.
	Error Level = iota
	// Warning is reported but does not fail.
	Warning
)

// Finding is one problem.
type Finding struct {
	Level   Level
	Message string
}

// Report is everything one run found.
type Report struct {
	Scope    []string
	Findings []Finding
}

// Errors returns the findings that fail the gate.
func (r Report) Errors() []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Level == Error {
			out = append(out, f)
		}
	}
	return out
}

// Warnings returns the findings that do not fail the gate.
func (r Report) Warnings() []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Level == Warning {
			out = append(out, f)
		}
	}
	return out
}

// OK reports whether the repository passed.
func (r Report) OK() bool {
	return len(r.Errors()) == 0
}

type checker struct {
	root     string
	findings []Finding
}

func (c *checker) errf(format string, args ...any) {
	c.findings = append(c.findings, Finding{Error, fmt.Sprintf(format, args...)})
}

func (c *checker) warnf(format string, args ...any) {
	c.findings = append(c.findings, Finding{Warning, fmt.Sprintf(format, args...)})
}

// Run checks the repository at root. Passing project slugs limits the
// project-scoped checks to those; repo-wide checks always run.
func Run(root string, only ...string) (Report, error) {
	c := &checker{root: root}

	projects, err := c.discoverProjects()
	if err != nil {
		return Report{}, err
	}

	scope := projects
	if len(only) > 0 {
		known := make(map[string]bool, len(projects))
		for _, p := range projects {
			known[p] = true
		}
		scope = nil
		for _, name := range only {
			if !known[name] {
				c.errf("unknown project %q; projects/ has %s", name, strings.Join(projects, ", "))
				continue
			}
			scope = append(scope, name)
		}
	}

	for _, name := range scope {
		if err := c.checkDeploy(name); err != nil {
			return Report{}, err
		}
	}
	if err := c.checkRegistry(projects); err != nil {
		return Report{}, err
	}
	if err := c.checkCommonSkills(projects); err != nil {
		return Report{}, err
	}
	c.checkRequirementIndexes()
	if err := c.checkInterviewRecorded(); err != nil {
		return Report{}, err
	}
	if err := c.checkQuestionsFollowTheDeck(); err != nil {
		return Report{}, err
	}
	if err := c.checkDocs(); err != nil {
		return Report{}, err
	}
	if err := c.checkOrphans(); err != nil {
		return Report{}, err
	}

	return Report{Scope: scope, Findings: c.findings}, nil
}

// discoverProjects finds the project directories. A project is a directory
// under projects/ that has a chart - the same rule the layout documents.
func (c *checker) discoverProjects() ([]string, error) {
	dir := filepath.Join(c.root, "projects")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		c.errf("projects/ does not exist; every project is projects/<name>/")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	var projects []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		chart := filepath.Join(dir, e.Name(), "chart", "app", "Chart.yaml")
		if _, err := os.Stat(chart); err == nil {
			projects = append(projects, e.Name())
		}
	}
	sort.Strings(projects)

	// A project the config declares but that has no chart on disk is the
	// failure this check exists for: everything else here inspects files, so
	// without this the repo passes and `render` is the first thing to say the
	// project is not real.
	if cfg, err := config.Load(filepath.Join(c.root, config.FileName)); err == nil {
		on := make(map[string]bool, len(projects))
		for _, p := range projects {
			on[p] = true
		}
		for _, p := range cfg.Projects {
			if !on[p.Slug] {
				c.errf("%s declares project %q but projects/%s/chart/app has no Chart.yaml; "+
					"run `asgard-cli scaffold` to write it", config.FileName, p.Slug, p.Slug)
			}
		}

		// The other direction: a chart on disk that the config does not
		// declare. `render` and CD both work from the config, so such a project
		// is dead weight nobody deploys, and until this check existed the only
		// symptom was a directory that never appeared in any output.
		declared := make(map[string]bool, len(cfg.Projects))
		for _, p := range cfg.Projects {
			declared[p.Slug] = true
		}
		for _, p := range projects {
			if !declared[p] {
				c.errf("projects/%s has a chart but %s does not declare it, so nothing renders or deploys it; "+
					"add it with `asgard-cli project add %s` or delete the directory", p, config.FileName, p)
			}
		}

		// A warning and not an error: nothing rendered from this repository
		// reads the id, so a repo without one is not broken. It is worth saying
		// once a project exists, because that is what the platform deploys and
		// an id nobody ever fetched is easy to carry all the way to a handover.
		if !cfg.Workspace.HasID() && len(cfg.Projects) > 0 {
			c.warnf("%s has no workspace.id, and %d project(s) are declared; set it with "+
				"`asgard-cli init --workspace-id ws_xxxxxxxx`", config.FileName, len(cfg.Projects))
		}
	}
	return projects, nil
}

// deployFile is the subset of deploy.yaml this checks.
type deployFile struct {
	Environments map[string]struct {
		Namespace string `yaml:"namespace"`
		Values    string `yaml:"values"`
	} `yaml:"environments"`
}

// checkDeploy verifies the deployment declaration: CI builds its matrix from
// these files, so a values path that does not exist is a failed deploy rather
// than a lint error.
func (c *checker) checkDeploy(project string) error {
	dir := filepath.Join(c.root, "projects", project)
	path := filepath.Join(dir, "deploy.yaml")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		c.errf("%s: missing projects/%s/deploy.yaml; it is the single source of truth for deployment targets, and is required even when no env is declared", project, project)
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	var parsed deployFile
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		c.errf("%s: projects/%s/deploy.yaml is not valid YAML: %v", project, project, err)
		return nil
	}

	if len(parsed.Environments) == 0 {
		c.warnf("%s: deploy.yaml declares no environment, so CI and asgard-cli render both skip this project", project)
		return nil
	}

	envs := make([]string, 0, len(parsed.Environments))
	for env := range parsed.Environments {
		envs = append(envs, env)
	}
	sort.Strings(envs)

	for _, env := range envs {
		spec := parsed.Environments[env]

		if !config.Env(env).Valid() {
			c.errf("%s: deploy.yaml declares env %q; only dev and prod are valid", project, env)
			continue
		}
		if spec.Namespace == "" {
			c.errf("%s: deploy.yaml %s has no namespace", project, env)
		}
		if spec.Values == "" {
			c.errf("%s: deploy.yaml %s has no values file", project, env)
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, spec.Values)); err != nil {
			c.errf("%s: deploy.yaml %s points at projects/%s/%s, which does not exist", project, env, project, spec.Values)
		}
		shared := filepath.Join(c.root, "common", fmt.Sprintf("values-%s.yaml", env))
		if _, err := os.Stat(shared); err != nil {
			c.errf("%s: %s is missing the shared common/values-%s.yaml, which CI overlays first", project, env, env)
		}
	}
	return nil
}

// projectRow matches the second column of the root README's project table,
// with or without the projects/ prefix.
var projectRow = regexp.MustCompile("(?m)^\\|[^|]*\\|\\s*`(?:projects/)?([a-z0-9-]+)/?`\\s*\\|")

// checkRegistry keeps the root README's project table and the directories in
// step. The table is how a reader learns what exists, so a stale one is worse
// than none.
func (c *checker) checkRegistry(projects []string) error {
	path := filepath.Join(c.root, "README.md")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		c.errf("README.md does not exist")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	listed := map[string]bool{}
	for _, m := range projectRow.FindAllSubmatch(data, -1) {
		listed[string(m[1])] = true
	}
	actual := map[string]bool{}
	for _, p := range projects {
		actual[p] = true
	}

	for _, p := range projects {
		if !listed[p] {
			c.errf("README.md project table is missing %q, which exists under projects/", p)
		}
	}
	for name := range listed {
		if !actual[name] {
			c.errf("README.md project table lists %q, which has no directory under projects/", name)
		}
	}
	return nil
}

var frontmatterBlock = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n`)

// frontmatter returns the top-level scalar fields of a markdown file's
// frontmatter, or nil when there is none.
func frontmatter(data []byte) map[string]string {
	m := frontmatterBlock.FindSubmatch(data)
	if m == nil {
		return nil
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(m[1]), "\n") {
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		out[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return out
}

// checkCommonSkills validates the runtime skills. The platform resolves one
// directory as one skill, so the directory name and the declared name have to
// agree or the SkillSet's searchPath silently resolves to nothing.
//
// The "none yet" warning is held back until there is a project. A freshly
// scaffolded repo has no agent to carry a skill, so warning there fires on
// every run of every new repo and teaches the reader that a warn from this
// command means nothing - which matters, because the interview check below
// reports something worth acting on through the same channel.
func (c *checker) checkCommonSkills(projects []string) error {
	dir := filepath.Join(c.root, "common", "skills")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		if len(projects) > 0 {
			c.warnf("common/skills/ does not exist; it is where runtime skills live, and there are none yet")
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}

	found := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		found = true
		path := filepath.Join(dir, e.Name(), "SKILL.md")
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			c.errf("common/skills/%s/ has no SKILL.md; one directory is one skill", e.Name())
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		fm := frontmatter(data)
		if fm == nil {
			c.errf("common/skills/%s/SKILL.md has no frontmatter; it needs name and description", e.Name())
			continue
		}
		for _, key := range []string{"name", "description"} {
			if fm[key] == "" {
				c.errf("common/skills/%s/SKILL.md frontmatter is missing %s", e.Name(), key)
			}
		}
		if name := fm["name"]; name != "" && name != e.Name() {
			c.errf("common/skills/%s/SKILL.md declares name %q, which does not match its directory", e.Name(), name)
		}
	}
	if !found && len(projects) > 0 {
		c.warnf("common/skills/ has no skill directories yet")
	}
	return nil
}

// checkInterviewRecorded reports customer material that nothing has been done
// with - read, and left in the repository with no trace of the reading.
//
// Every other check here asks whether a file is well formed. This one asks
// whether work left anything behind, because that failure is not a malformed
// record - it is no record at all. Material is read, the analysis is done well,
// it is delivered in conversation, and the repository ends the day looking
// exactly as it did before.
//
// **Filed material is not evidence that an interview happened.** The first
// version of this check assumed it was, and warned through several hours of a
// perfectly normal state: a customer sent their own document ahead of the
// meeting, which is the usual order here because their internal approval comes
// before they will book one. Nothing was being lost - the material had been
// read and sixteen open questions written from it - and the tool said otherwise
// on every run. A warning that fires while somebody is doing the right thing
// teaches them to stop reading warnings.
//
// So open questions count as a record. They are the trace the reading leaves
// before there is anything to request.
func (c *checker) checkInterviewRecorded() error {
	filed, err := work.FiledReferences(c.root)
	if err != nil {
		return err
	}
	if filed == 0 {
		return nil
	}

	requests, err := work.ReadRequests(c.root)
	if err != nil {
		return err
	}
	if len(requests) > 0 {
		return nil
	}

	questions, err := work.ReadQuestions(c.root)
	if err != nil {
		return err
	}
	if len(questions) > 0 {
		return nil
	}

	scheduled, err := meetingScheduled(c.root)
	if err != nil {
		return err
	}
	if scheduled {
		return nil
	}

	c.warnf("references/ holds %d file(s) of customer material, and nothing records that it "+
		"was read - no question in %s, no meeting filed, no request. Read it into questions "+
		"first: `asgard-cli question add \"<what blocks it>\" --ask \"<who can answer>\"`. "+
		"**A request comes after the interview, not before it** - six of its seven sections "+
		"are what the interview decides",
		filed, filepath.ToSlash(work.QuestionFile))
	return nil
}

// checkQuestionsFollowTheDeck reports a deck edited more recently than the
// questions it came from.
//
// The deck is not a rendering of `docs/open-questions.md` - it is that file's
// second editor. Working through one with a customer rewrites the questions,
// retires some and finds others, and one engagement's twenty rounds of revision
// changed about half of them, dropped three and added six. None of it went
// back, because nothing said it should.
//
// That matters more than an ordinary staleness because **`next` prints the
// questions before anything else**. The next person to pick the repository up
// reads the superseded file first, and walks into a meeting with questions
// already abandoned. Worse, a judgement that was overturned survives there
// looking considered - in that engagement, a security reasoning the deck had
// corrected was still sitting in the file, argued well.
//
// Modification time is a weak signal and the right one here: it is exactly the
// question being asked, it needs no parsing, and a false positive costs one
// glance.
func (c *checker) checkQuestionsFollowTheDeck() error {
	questions := filepath.Join(c.root, work.QuestionFile)
	qi, err := os.Stat(questions)
	if err != nil {
		return nil
	}

	dir := filepath.Join(c.root, "docs", "meeting-notes")
	var newer []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == dir {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch ext := strings.ToLower(filepath.Ext(d.Name())); ext {
		case ".html", ".md", ".json", ".pdf":
		default:
			return nil
		}
		if d.Name() == "README.md" || strings.HasPrefix(d.Name(), "_") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.ModTime().After(qi.ModTime()) {
			rel, _ := filepath.Rel(c.root, path)
			newer = append(newer, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	if len(newer) == 0 {
		return nil
	}

	sort.Strings(newer)
	c.warnf("%s was edited after %s - %d file(s), the newest being %s. **Working a deck "+
		"rewrites the questions**, so if any was reworded, retired or discovered while "+
		"building it, that belongs back in the row. `asgard-cli next` prints those "+
		"questions before anything else, and a stale one does not merely lag: it keeps "+
		"a judgement that has been overturned, argued well",
		filepath.ToSlash(filepath.Join("docs", "meeting-notes")),
		filepath.ToSlash(work.QuestionFile), len(newer), newer[len(newer)-1])
	return nil
}

// meetingScheduled reports whether a meeting has been filed under
// docs/meeting-notes/. A directory or a dated file there means the material has
// not only been read but turned into an agenda, which is the whole of what the
// warning above is asking for.
func meetingScheduled(root string) (bool, error) {
	dir := filepath.Join(root, "docs", "meeting-notes")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", dir, err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "_") || e.Name() == "README.md" {
			continue
		}
		if dateNamed.MatchString(e.Name()) || (e.IsDir() && len(e.Name()) > 10) {
			return true, nil
		}
	}
	return false, nil
}

// checkRequirementIndexes verifies the SDD entry points exist.
func (c *checker) checkRequirementIndexes() {
	for _, rel := range []string{
		filepath.Join("requirements", "_index.md"),
		filepath.Join("requirements", "requests", "_index.md"),
		filepath.Join("requirements", "tasks", "_index.md"),
	} {
		if _, err := os.Stat(filepath.Join(c.root, rel)); err != nil {
			c.errf("missing %s; it is an SDD entry point, see docs/spec-driven-development.md", filepath.ToSlash(rel))
		}
	}
}

// dateNamed matches the filename convention for dated records.
var dateNamed = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-[a-z0-9][a-z0-9-]*\.md$`)

// moduleLink matches a link to a module file from a spec index.
var moduleLink = regexp.MustCompile(`\]\(([a-z0-9][a-z0-9-]*\.md)\)`)

// markdownLink matches a relative markdown link, ignoring any anchor.
var markdownLink = regexp.MustCompile(`\]\(([^)\s#]+\.md)(?:#[^)\s]*)?\)`)

// checkDocs validates the spec layer: the files the four-layer model needs, the
// living spec's module index matching what is on disk, dated filenames, and
// that every relative link inside docs/ resolves. Rotten links are how this
// layer decays.
func (c *checker) checkDocs() error {
	docs := filepath.Join(c.root, "docs")
	if _, err := os.Stat(docs); os.IsNotExist(err) {
		c.errf("docs/ does not exist; it is the spec and decision layer, see docs/README.md")
		return nil
	}

	for _, rel := range []string{
		"README.md",
		"spec-driven-development.md",
		filepath.Join("spec", "README.md"),
		filepath.Join("decisions", "README.md"),
		filepath.Join("decisions", "_decision-template.md"),
		filepath.Join("meeting-notes", "README.md"),
		filepath.Join("meeting-notes", "_template.md"),
	} {
		if _, err := os.Stat(filepath.Join(docs, rel)); err != nil {
			c.errf("missing docs/%s; see the converge loop in docs/README.md", filepath.ToSlash(rel))
		}
	}

	if err := c.checkLivingSpec(filepath.Join(docs, "spec")); err != nil {
		return err
	}
	if err := c.checkDatedNames(docs); err != nil {
		return err
	}
	return c.checkDocLinks(docs)
}

// checkLivingSpec compares each spec slug's module index against the files
// present. A module written but never indexed is invisible to the next reader.
func (c *checker) checkLivingSpec(specDir string) error {
	entries, err := os.ReadDir(specDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", specDir, err)
	}

	slugs := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		slugs++
		slug := e.Name()
		index := filepath.Join(specDir, slug, "README.md")

		data, err := os.ReadFile(index)
		if os.IsNotExist(err) {
			c.errf("missing docs/spec/%s/README.md, the living spec's module index", slug)
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", index, err)
		}

		listed := map[string]bool{}
		for _, m := range moduleLink.FindAllSubmatch(data, -1) {
			listed[string(m[1])] = true
		}

		modules, err := os.ReadDir(filepath.Join(specDir, slug))
		if err != nil {
			return fmt.Errorf("read %s: %w", filepath.Join(specDir, slug), err)
		}
		actual := map[string]bool{}
		for _, m := range modules {
			if m.IsDir() || m.Name() == "README.md" || !strings.HasSuffix(m.Name(), ".md") {
				continue
			}
			actual[m.Name()] = true
			if !listed[m.Name()] {
				c.errf("docs/spec/%s/README.md does not index %s, which exists", slug, m.Name())
			}
		}
		for name := range listed {
			if !actual[name] {
				c.errf("docs/spec/%s/README.md indexes %s, which does not exist", slug, name)
			}
		}
	}
	if slugs == 0 {
		c.warnf("docs/spec/ has no living spec yet; it should be docs/spec/<slug>/")
	}
	return nil
}

// checkDatedNames enforces YYYY-MM-DD-<topic>.md on the immutable layers. The
// date is what makes a record findable later.
func (c *checker) checkDatedNames(docs string) error {
	for _, sub := range []string{"decisions", "meeting-notes"} {
		dir := filepath.Join(docs, sub)
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || name == "README.md" || strings.HasPrefix(name, "_") || !strings.HasSuffix(name, ".md") {
				continue
			}
			if !dateNamed.MatchString(name) {
				c.errf("docs/%s/%s is not named YYYY-MM-DD-<topic>.md with a lowercase kebab-case topic", sub, name)
			}
		}
	}
	return nil
}

// checkDocLinks resolves every relative markdown link inside docs/.
func (c *checker) checkDocLinks(docs string) error {
	return filepath.Walk(docs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		rel, _ := filepath.Rel(c.root, path)
		for _, m := range markdownLink.FindAllSubmatch(data, -1) {
			target := string(m[1])
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "/") {
				continue
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), target)); err != nil {
				c.errf("%s links to %s, which does not resolve", filepath.ToSlash(rel), target)
			}
		}
		return nil
	})
}

// wikiPage reports whether a file under docs/ or requirements/ is a page that
// somebody is expected to reach by following a link. READMEs are entry points
// and are reached by opening a directory; a leading underscore marks an index
// or a template, which is linked TO rather than linked FROM.
func wikiPage(rel string) bool {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if !strings.HasPrefix(dir, "docs") && !strings.HasPrefix(dir, "requirements") {
		return false
	}
	// A living spec module is owned by checkLivingSpec, which compares the slug
	// README against the files on disk and says exactly which is missing. That
	// is the same defect stated more precisely, so reporting it here as well
	// would print two findings for one problem.
	if strings.HasPrefix(dir, "docs/spec/") {
		return false
	}
	name := filepath.Base(rel)
	return strings.HasSuffix(name, ".md") &&
		name != "README.md" &&
		!strings.HasPrefix(name, "_")
}

// checkOrphans reports pages nothing links to.
//
// It is the one part of a knowledge base's decay that is mechanical. A page
// with no inbound link is not read, and the person who wrote it never finds
// out, because the file is still sitting there: a decision nobody applied, a
// module missing from its index, a task spec that fell out of the queue.
//
// A warning, not an error. A decision recorded today and not yet applied to the
// living spec is an orphan for as long as that takes, and that is a normal
// state to pass through - failing the gate on it would leave the gate red in
// the middle of ordinary work. The three kinds of rot that are NOT mechanical -
// two pages that contradict each other, a claim a newer source superseded, and
// a concept discussed everywhere but owned by no page - need a reader, and the
// knowledge-base skill under .agents/skills/ is how that pass is run.
func (c *checker) checkOrphans() error {
	linked := map[string]bool{}
	var pages []string

	err := filepath.Walk(c.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			switch info.Name() {
			case ".git", ".out", "node_modules", ".venv":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Every markdown file in the repo counts as a source of links, not just
		// the ones under docs/: AGENTS.md is where an agent starts, and a page
		// reachable only from a skill is still reachable.
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		for _, m := range markdownLink.FindAllSubmatch(data, -1) {
			target := string(m[1])
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "/") {
				continue
			}
			linked[filepath.Clean(filepath.Join(filepath.Dir(path), target))] = true
		}

		if rel, err := filepath.Rel(c.root, path); err == nil && wikiPage(rel) {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(pages)
	for _, path := range pages {
		if linked[filepath.Clean(path)] {
			continue
		}
		rel, _ := filepath.Rel(c.root, path)
		c.warnf("%s has no inbound link, so nothing leads a reader to it: %s",
			filepath.ToSlash(rel), orphanHint(rel))
	}
	return nil
}

// orphanHint names the index that was supposed to carry the link, because the
// fix differs by layer and "add a link" does not say where.
func orphanHint(rel string) string {
	switch dir := filepath.ToSlash(filepath.Dir(rel)); {
	case strings.HasPrefix(dir, "docs/decisions"):
		return "link it from the living spec module it changed, and from docs/decisions/README.md"
	case strings.HasPrefix(dir, "docs/meeting-notes"):
		return "link it from the decision it converged into, or from docs/meeting-notes/README.md"
	case strings.HasPrefix(dir, "docs/spec"):
		return "add it to that slug's README module index"
	case strings.HasPrefix(dir, "requirements/requests"):
		return "register it in requirements/requests/_index.md"
	case strings.HasPrefix(dir, "requirements/tasks"):
		return "register it in requirements/tasks/_index.md"
	default:
		return "link it from the index of its layer"
	}
}
