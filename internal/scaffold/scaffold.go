// Package scaffold writes the parts of a customer repository that are the same
// for every engagement: the four-layer docs model, the SDD rules, the acceptance
// gate scripts, the design-time skills and the CD workflow.
//
// What it deliberately does not write is the customer's own knowledge - which
// systems exist, how the projects split, what the CRs look like. That is what
// the onboarding is for, and templates cannot produce it.
package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
)

// The tree is embedded with all: so that dot-prefixed paths (.agents, .github,
// .gitignore, .env.example) are included; the default pattern skips them.
//
//go:embed all:templates
var templates embed.FS

const (
	templateRoot = "templates"

	// tmplSuffix marks a file that is rendered rather than copied. Templates use
	// << >> because the files they produce contain Helm's {{ }}.
	tmplSuffix = ".tmpl"
	leftDelim  = "<<"
	rightDelim = ">>"

	// These path segments are placeholders expanded at render time. A path
	// containing projectDir is written once per project, and one containing
	// envDir once per environment that project declares.
	specSlugDir = "__SPEC_SLUG__"
	projectDir  = "__PROJECT__"
	envDir      = "__ENV__"

	// specSuffix turns the workspace slug into the living spec's slug.
	specSuffix = "-asgard"
)

// Project wraps config.Project with the presentation the templates need.
type Project struct {
	config.Project
}

// EnvList renders the declared environments the way the README table shows them.
func (p Project) EnvList() string {
	names := make([]string, len(p.Environments))
	for i, e := range p.Environments {
		names[i] = string(e)
	}
	return strings.Join(names, " + ")
}

// Data is what every template is rendered with. Project and Env are set only
// while rendering a file that lives under a __PROJECT__ or __ENV__ path.
type Data struct {
	Workspace config.Workspace
	Projects  []Project
	RepoName  string
	SpecSlug  string

	Project *Project
	Env     config.Env
}

// Namespace is the namespace the current project and environment deploy to.
// Only meaningful while rendering a per-project-per-env file.
func (d Data) Namespace() string {
	if d.Project == nil || d.Env == "" {
		return ""
	}
	return fmt.Sprintf("asgard-%s-%s-%s", d.Workspace.Slug, d.Project.Slug, d.Env)
}

// NewData derives the render data from a config.
func NewData(cfg *config.Config) Data {
	projects := make([]Project, len(cfg.Projects))
	for i, p := range cfg.Projects {
		projects[i] = Project{Project: p}
	}
	return Data{
		Workspace: cfg.Workspace,
		Projects:  projects,
		RepoName:  cfg.RepoName(),
		SpecSlug:  SpecSlug(cfg.Workspace.Slug),
	}
}

// SpecSlug is the directory name under docs/spec/ for this workspace.
func SpecSlug(workspaceSlug string) string {
	return workspaceSlug + specSuffix
}

// Status is what happened to one file.
type Status int

const (
	// Created means the file was written.
	Created Status = iota
	// Skipped means a file was already there and force was not set.
	Skipped
	// Overwritten means an existing file was replaced because force was set.
	Overwritten
	// Updated means only the file's managed region was refreshed, leaving
	// everything the engagement wrote around it untouched.
	Updated
	// Preserved means force was set and the file was left alone anyway,
	// because it is one the CLI's own commands write into and it no longer
	// matches the template it started as.
	Preserved
)

func (s Status) String() string {
	switch s {
	case Created:
		return "created"
	case Overwritten:
		return "overwritten"
	case Updated:
		return "updated"
	case Preserved:
		return "preserved"
	default:
		return "skipped"
	}
}

// Result reports one file's outcome, with paths relative to the repo root.
type Result struct {
	Path   string
	Status Status
}

// Write renders the skeleton into root. It never removes anything, and without
// force it leaves existing files alone, so it can be run again after a project
// is added or when a file was deleted by hand.
func Write(root string, cfg *config.Config, force bool) ([]Result, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	jobs, err := plan(NewData(cfg))
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(jobs))
	for _, j := range jobs {
		target := filepath.Join(root, j.target)

		exists, err := fileExists(target)
		if err != nil {
			return nil, err
		}
		content, err := render(j.source, j.data)
		if err != nil {
			return nil, err
		}

		// An accumulator is a file the CLI's other commands write into after
		// scaffold has run - an index, the open-questions table. --force means
		// "discard local edits to the skeleton", and these stopped being
		// skeleton the first time `request add` or `question add` touched them.
		// Overwriting one silently destroys an engagement's interview, and in a
		// repo with no commits there is nothing to recover from. Untouched ones
		// still match their template, so leaving those to the normal path costs
		// nothing.
		if exists && force && accumulator(j.target) {
			current, err := os.ReadFile(target)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", target, err)
			}
			if !bytes.Equal(current, content) {
				results = append(results, Result{Path: j.target, Status: Preserved})
				continue
			}
		}

		if exists && !force {
			// A managed region is derived from the config, so leaving it stale
			// would put the file out of step with the repo - the project table
			// in README.md is the case that matters, because the gate compares
			// it against the directories on disk.
			merged, updated, err := mergeManaged(target, content)
			if err != nil {
				return nil, err
			}
			if !updated {
				results = append(results, Result{Path: j.target, Status: Skipped})
				continue
			}
			if err := writeFile(target, merged, executable(j.target)); err != nil {
				return nil, err
			}
			results = append(results, Result{Path: j.target, Status: Updated})
			continue
		}
		if err := writeFile(target, content, executable(j.target)); err != nil {
			return nil, err
		}

		status := Created
		if exists {
			status = Overwritten
		}
		results = append(results, Result{Path: j.target, Status: status})
	}

	return results, nil
}

// job is one template rendered to one path with one set of data.
type job struct {
	source string
	target string
	data   Data
}

// accumulators are the files this CLI's own commands append to. The spec
// module index is matched by suffix because its directory carries the spec slug.
var accumulators = map[string]bool{
	filepath.Join("docs", "open-questions.md"):             true,
	filepath.Join("requirements", "requests", "_index.md"): true,
	filepath.Join("requirements", "tasks", "_index.md"):    true,
	filepath.Join("docs", "decisions", "README.md"):        true,
}

func accumulator(target string) bool {
	if accumulators[target] {
		return true
	}
	// docs/spec/<slug>/README.md carries the living spec's module index and its
	// traceability table, both written a row at a time as the work happens.
	dir, file := filepath.Split(target)
	return file == "README.md" && strings.HasPrefix(filepath.ToSlash(dir), "docs/spec/")
}

// plan walks the embedded tree and expands the placeholder path segments. A
// template under __PROJECT__ produces one file per project, and one that also
// carries __ENV__ produces one per environment that project declares - so a
// project with dev only never gets a values-prod.yaml it would have to explain.
func plan(data Data) ([]job, error) {
	var jobs []job

	err := fs.WalkDir(templates, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return fmt.Errorf("resolve template path %s: %w", path, err)
		}
		rel = strings.ReplaceAll(rel, specSlugDir, data.SpecSlug)

		if !strings.Contains(rel, projectDir) {
			jobs = append(jobs, job{source: path, target: trimSuffix(rel), data: data})
			return nil
		}

		for _, project := range data.Projects {
			projectData := data
			projectData.Project = &project
			projectRel := strings.ReplaceAll(rel, projectDir, project.Slug)

			if !strings.Contains(projectRel, envDir) {
				jobs = append(jobs, job{source: path, target: trimSuffix(projectRel), data: projectData})
				continue
			}

			for _, env := range project.Environments {
				envData := projectData
				envData.Env = env
				envRel := strings.ReplaceAll(projectRel, envDir, string(env))
				jobs = append(jobs, job{source: path, target: trimSuffix(envRel), data: envData})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// trimSuffix drops the marker that says a file is rendered rather than copied.
func trimSuffix(rel string) string {
	return strings.TrimSuffix(rel, tmplSuffix)
}

// render returns the file's contents, running it through text/template only
// when it carries the template suffix.
func render(path string, data Data) ([]byte, error) {
	content, err := templates.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", path, err)
	}
	if !strings.HasSuffix(path, tmplSuffix) {
		return content, nil
	}

	tmpl, err := template.New(filepath.Base(path)).
		Delims(leftDelim, rightDelim).
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("render template %s: %w", path, err)
	}
	return []byte(out.String()), nil
}

// executable reports whether the written file needs the execute bit: the gate
// scripts carry a shebang and are run directly.
//
// .sh is still here although no template is one any more. common/render.sh was,
// until it became `asgard-cli render` - it was bash calling yq, so the whole
// acceptance gate was unavailable on Windows while helm itself has a native
// Windows build. Keep the case: the next shell script somebody adds should not
// arrive without its execute bit.
func executable(rel string) bool {
	switch filepath.Ext(rel) {
	case ".sh", ".py":
		return true
	default:
		return false
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
}

func writeFile(path string, content []byte, exec bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}
	mode := os.FileMode(0o644)
	if exec {
		mode = 0o755
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	// WriteFile does not change the mode of a file that already exists.
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}
	return nil
}

// managedRegion matches a block the CLI keeps in step with the config even in a
// file that has otherwise been edited by hand.
var managedRegion = regexp.MustCompile(`(?s)<!-- asgard-cli:managed:start -->.*?<!-- asgard-cli:managed:end -->`)

// mergeManaged replaces the managed region of the file at path with the one from
// freshly rendered content. It reports false when either side has no managed
// region, or when the region is already identical.
func mergeManaged(path string, rendered []byte) ([]byte, bool, error) {
	want := managedRegion.Find(rendered)
	if want == nil {
		return nil, false, nil
	}

	current, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	if managedRegion.Find(current) == nil {
		// The marker was removed deliberately; leave the file alone.
		return nil, false, nil
	}

	merged := managedRegion.ReplaceAllFunc(current, func([]byte) []byte { return want })
	if string(merged) == string(current) {
		return nil, false, nil
	}
	return merged, true, nil
}
