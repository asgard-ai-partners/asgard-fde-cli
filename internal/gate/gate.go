// Package gate runs the invariant checks on a rendered chart.
//
// These are the checks that helm lint, CRD validation and a server-side dry run
// all pass: a missing display annotation, a reference to a CR that does not
// exist, an entry name that is not declared, a Workflow with no set labels. Each
// one applies cleanly and then fails at runtime or shows up as a blank page in
// the Platform UI, which is why they need their own gate.
//
// They were two Python scripts, and the port is what makes the gate work on
// Windows: the scripts needed PyYAML in a virtualenv, and were reached through a
// bash pipeline. Nothing here needs anything but the binary.
package gate

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Doc is one rendered manifest.
type Doc struct {
	Kind        string
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Spec        map[string]any
}

// Options is what a check needs that the manifests do not say.
type Options struct {
	// Project the manifests were rendered from. It fills in the commands a
	// problem's remedy names, so that what is printed can be run as printed;
	// empty becomes a placeholder.
	Project string

	// Release the manifests were rendered for. The Platform derives this
	// release's Secret and ConfigMap names from it, and a credential reference
	// has to point at one of those two - so a check needs the name to know what
	// the references SHOULD say. Empty when checking a stream rendered
	// elsewhere, and the credential check then has nothing to compare against.
	Release string
}

// ProjectOr returns the project name, or a placeholder when it is not known -
// as it is not when checking a stream that was rendered elsewhere.
func (o Options) ProjectOr() string {
	if o.Project == "" {
		return "<project>"
	}
	return o.Project
}

// Result is what one check found.
type Result struct {
	// Problems fail the gate.
	Problems []string

	// Warnings do not. They are for a condition that is correct now and fatal
	// later - a project with no Syncer yet, an environment id the platform has
	// not issued - where failing would make the gate red through the whole
	// middle of an onboarding, and staying silent means finding out from a red
	// tag.
	Warnings []string

	Summary string
}

// OK reports whether the check passed. Warnings do not affect it.
func (r Result) OK() bool { return len(r.Problems) == 0 }

// Read parses a stream of rendered manifests.
//
// Anything that is not a mapping is skipped rather than rejected: helm writes a
// leading `---` and a trailing newline, and a stderr line captured into the same
// file would otherwise fail the whole gate for the wrong reason.
func Read(r io.Reader) ([]Doc, error) {
	dec := yaml.NewDecoder(r)

	var docs []Doc
	for {
		// Decoded as any rather than as a map: a document that is not a mapping
		// has to be skipped, and decoding straight into a map turns it into an
		// error instead. That is the whole case this tolerance exists for.
		var doc any
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse rendered manifests: %w", err)
		}
		raw := mapOf(doc)
		if raw == nil {
			continue
		}

		kind := str(raw["kind"])
		meta := mapOf(raw["metadata"])
		if kind == "" || meta == nil {
			continue
		}

		docs = append(docs, Doc{
			Kind:        kind,
			Name:        str(meta["name"]),
			Labels:      strMap(meta["labels"]),
			Annotations: strMap(meta["annotations"]),
			Spec:        mapOf(raw["spec"]),
		})
	}
	return docs, nil
}

// index groups the documents the way both checks need them.
type index struct {
	docs         []Doc
	namesByKind  map[string]map[string]bool
	countsByKind map[string]int
}

func newIndex(docs []Doc) *index {
	ix := &index{
		docs:         docs,
		namesByKind:  map[string]map[string]bool{},
		countsByKind: map[string]int{},
	}
	for _, d := range docs {
		if ix.namesByKind[d.Kind] == nil {
			ix.namesByKind[d.Kind] = map[string]bool{}
		}
		ix.namesByKind[d.Kind][d.Name] = true
		ix.countsByKind[d.Kind]++
	}
	return ix
}

// has reports whether a CR of this kind and name was rendered.
func (ix *index) has(kind, name string) bool {
	return ix.namesByKind[kind][name]
}

// of returns every doc of one kind.
func (ix *index) of(kind string) []Doc {
	var out []Doc
	for _, d := range ix.docs {
		if d.Kind == kind {
			out = append(out, d)
		}
	}
	return out
}

// summary counts the rendered CRs, in a fixed order.
func (ix *index) summary() string {
	kinds := make([]string, 0, len(ix.countsByKind))
	for k := range ix.countsByKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)

	if len(kinds) == 0 {
		return "no CRs yet"
	}

	parts := make([]string, 0, len(kinds))
	total := 0
	for _, k := range kinds {
		parts = append(parts, fmt.Sprintf("%s=%d", k, ix.countsByKind[k]))
		total += ix.countsByKind[k]
	}
	return fmt.Sprintf("CRs: %d (%s)", total, strings.Join(parts, ", "))
}

// The accessors below walk the generic maps yaml.v3 produces. They return the
// zero value for anything missing or of the wrong type, because a rendered
// manifest that does not match the CRD is the apiserver's business, not this
// gate's - and a nil check at every level would bury the rules.

func str(v any) string {
	s, _ := v.(string)
	return s
}

func mapOf(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func listOf(v any) []any {
	l, _ := v.([]any)
	return l
}

func strMap(v any) map[string]string {
	out := map[string]string{}
	for k, val := range mapOf(v) {
		if s, ok := val.(string); ok {
			out[k] = s
		}
	}
	return out
}

// dig walks a path of map keys.
func dig(m map[string]any, path ...string) any {
	var cur any = m
	for _, key := range path {
		next := mapOf(cur)
		if next == nil {
			return nil
		}
		cur = next[key]
	}
	return cur
}

// digStr walks a path and returns the string at the end.
func digStr(m map[string]any, path ...string) string {
	return str(dig(m, path...))
}

// digList walks a path and returns the list at the end.
func digList(m map[string]any, path ...string) []any {
	return listOf(dig(m, path...))
}

// strList reads a list of strings, skipping anything that is not one.
func strList(v any) []string {
	var out []string
	for _, item := range listOf(v) {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// sorted returns a set's members in order, or "none".
func sorted(set map[string]bool) string {
	if len(set) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
