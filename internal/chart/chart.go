// Package chart reads what a project's chart already declares.
//
// It reads the templates as text rather than parsing them as YAML, because a
// Helm template is not valid YAML until it is rendered. That is enough for the
// question anything here asks - which CRs exist, and what they are called - and
// it works before helm is installed, which `asgard-cli add` needs.
package chart

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Ref is one custom resource a chart declares.
type Ref struct {
	Kind string
	Name string
}

// TemplatesDir is where a project's CR templates live.
func TemplatesDir(root, project string) string {
	return filepath.Join(root, "projects", project, "chart", "app", "templates")
}

// kindPattern matches the kind line of a manifest.
var kindPattern = regexp.MustCompile(`(?m)^kind:\s*"?([A-Za-z]+)"?`)

// namePattern matches the metadata name that follows it. The name may be
// quoted, and in a generated skeleton it is always a literal - a templated name
// is not something this can resolve, and is reported as empty rather than
// guessed at.
var namePattern = regexp.MustCompile(`(?m)^\s{2}name:\s*"?([A-Za-z0-9][A-Za-z0-9._-]*)"?`)

// Scan returns every CR the project's chart declares, in file order.
//
// A project with no chart yet has none; that is a state, not an error.
func Scan(root, project string) ([]Ref, error) {
	dir := TemplatesDir(root, project)

	var refs []Ref
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return fs.SkipAll
			}
			return err
		}
		if d.IsDir() || !isManifest(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		refs = append(refs, parse(string(content))...)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return refs, nil
}

// parse pulls the (kind, name) pairs out of one file, which may hold several
// documents separated by ---.
func parse(content string) []Ref {
	var refs []Ref
	for doc := range strings.SplitSeq(content, "\n---") {
		kind := kindPattern.FindStringSubmatch(doc)
		if kind == nil {
			continue
		}
		ref := Ref{Kind: kind[1]}
		if name := namePattern.FindStringSubmatch(doc); name != nil {
			ref.Name = name[1]
		}
		refs = append(refs, ref)
	}
	return refs
}

func isManifest(path string) bool {
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		return true
	default:
		return false
	}
}

// NamesOf returns the names of every CR of one kind, sorted so that a caller
// picking "the only one" is deterministic.
func NamesOf(refs []Ref, kind string) []string {
	var out []string
	for _, r := range refs {
		if r.Kind == kind && r.Name != "" {
			out = append(out, r.Name)
		}
	}
	sort.Strings(out)
	return out
}

// Counts returns how many of each kind the chart declares.
func Counts(refs []Ref) map[string]int {
	out := map[string]int{}
	for _, r := range refs {
		out[r.Kind]++
	}
	return out
}
