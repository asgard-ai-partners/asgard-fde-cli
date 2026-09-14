package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// A CRD read far enough to answer what this gate asks of it. Everything below
// the schema is `any`, because the questions here are about shapes that repeat
// at every depth rather than about a fixed set of fields.
type crd struct {
	Spec struct {
		Names struct {
			Kind string `yaml:"kind"`
		} `yaml:"names"`
		Versions []struct {
			Schema struct {
				OpenAPIV3Schema map[string]any `yaml:"openAPIV3Schema"`
			} `yaml:"schema"`
		} `yaml:"versions"`
	} `yaml:"spec"`
}

// loadCRDs reads every CRD in a directory.
//
// **Unmarshalled rather than shelled out to.** These used to be converted with
// `yq` and read back as JSON, which made the check need a tool the repository
// does not otherwise use and turned every schema question into string handling.
func loadCRDs(dir string) ([]crd, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no *.yaml in %s, so there is no contract to check against", dir)
	}
	out := make([]crd, 0, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var c crd
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		if c.Spec.Names.Kind == "" || len(c.Spec.Versions) == 0 {
			return nil, fmt.Errorf("%s has no kind or no version, so it is not a CRD this can read", p)
		}
		out = append(out, c)
	}
	return out, nil
}

// spec is the `spec` property's schema, which is the only part a chart writes.
func (c crd) spec() map[string]any {
	props, _ := c.Spec.Versions[0].Schema.OpenAPIV3Schema["properties"].(map[string]any)
	s, _ := props["spec"].(map[string]any)
	return s
}

// walkProps calls fn for every named property at every depth, with the dotted
// path to it. A list's `items` carries the path of the list, because a
// constraint on an element is a constraint on the field.
func walkProps(node any, path string, fn func(path string, schema map[string]any)) {
	m, ok := node.(map[string]any)
	if !ok {
		if list, ok := node.([]any); ok {
			for _, e := range list {
				walkProps(e, path, fn)
			}
		}
		return
	}
	if props, ok := m["properties"].(map[string]any); ok {
		names := make([]string, 0, len(props))
		for n := range props {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			child := n
			if path != "" {
				child = path + "." + n
			}
			if schema, ok := props[n].(map[string]any); ok {
				fn(child, schema)
				walkProps(schema, child, fn)
			}
		}
	}
	if items, ok := m["items"]; ok {
		walkProps(items, path, fn)
	}
}

// leaf is the last segment of a dotted path, which is how the pinned tables in
// internal/gate are keyed: they match on a json field name wherever it appears.
func leaf(path string) string {
	if i := strings.LastIndex(path, "."); i >= 0 {
		return path[i+1:]
	}
	return path
}

// walkAll calls fn for every map in a schema, at every depth and by every key.
//
// walkProps answers questions keyed by property name and descends only where a
// property can be; this answers questions about anything a schema node can
// carry, which includes the validations attached to an object rather than to
// one of its fields.
func walkAll(node any, fn func(map[string]any)) {
	switch t := node.(type) {
	case map[string]any:
		fn(t)
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			walkAll(t[k], fn)
		}
	case []any:
		for _, e := range t {
			walkAll(e, fn)
		}
	}
}
