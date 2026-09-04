// Package pipelineconfig reads the repository's `.asgard-pipeline.yaml`.
//
// It reads and it does not judge. Whether a config is valid is the platform's
// answer - the same parse that decides a run also computes `vars/required-
// missing` and matches trigger patterns, and a second opinion here would be a
// second opinion that disagrees the first time either changes. What this
// package is for is the questions the CLI has to answer before it can call the
// platform at all: which releases this repository declares, and which chart
// directory each one names.
//
// So a field this package does not need is carried without interpretation, and
// a file it cannot parse is reported as unreadable rather than as invalid.
package pipelineconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileName is the declaration's default path, relative to the repository root.
const FileName = ".asgard-pipeline.yaml"

// ErrNotFound reports that no declaration exists at the path looked at.
var ErrNotFound = errors.New("no " + FileName)

// Trigger is a release's `on:` block.
type Trigger struct {
	// Type is "tag" or "branch".
	Type string `yaml:"type"`
	// Pattern is an RE2 regular expression matched against the whole tag or
	// branch name. It is not compiled here: the platform matches it, and a
	// pattern this CLI rejected but the platform accepted would be a lie.
	Pattern string `yaml:"pattern"`
}

// Key is one declared key of any of the three kinds.
type Key struct {
	Key string `yaml:"key"`
	// Type applies to chart values only: "auto" (the default), "string" or
	// "json".
	Type string `yaml:"type,omitempty"`
	// Required defaults to true when the field is absent, which is why it is a
	// pointer: false and unset are different declarations.
	Required *bool `yaml:"required,omitempty"`
}

// IsRequired reports the effective requirement, applying the yaml default.
func (k Key) IsRequired() bool { return k.Required == nil || *k.Required }

// Release is one entry of `releases:`.
type Release struct {
	Name string   `yaml:"name"`
	On   *Trigger `yaml:"on,omitempty"`
	// Chart is a directory containing Chart.yaml, relative to the repository
	// root.
	Chart        string `yaml:"chart"`
	ChartValues  []Key  `yaml:"chartValues,omitempty"`
	AppSecret    []Key  `yaml:"appSecret,omitempty"`
	AppConfigMap []Key  `yaml:"appConfigMap,omitempty"`
}

// Config is the whole declaration.
type Config struct {
	Version  int       `yaml:"version"`
	Releases []Release `yaml:"releases"`

	// Path is where it was read from, relative to the repository root. It is
	// not part of the file.
	Path string `yaml:"-"`
}

// Release returns the declared release with the given name.
func (c *Config) Release(name string) (*Release, bool) {
	for i := range c.Releases {
		if c.Releases[i].Name == name {
			return &c.Releases[i], true
		}
	}
	return nil, false
}

// Names lists every declared release, in declaration order.
func (c *Config) Names() []string {
	out := make([]string, 0, len(c.Releases))
	for _, r := range c.Releases {
		out = append(out, r.Name)
	}
	return out
}

// Load reads the declaration at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s: %w", path, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// The platform is the judge of a config's validity, but a file that
		// does not parse as YAML cannot be read by anything, so saying so here
		// beats sending it to the platform to be told the same.
		return nil, fmt.Errorf("%s is not valid YAML: %w", path, err)
	}
	cfg.Path = path
	return &cfg, nil
}

// LoadFromRepo reads the declaration of a repository, at configPath when one is
// given and at the default otherwise.
func LoadFromRepo(root, configPath string) (*Config, error) {
	rel := configPath
	if rel == "" {
		rel = FileName
	}
	cfg, err := Load(filepath.Join(root, rel))
	if err != nil {
		return nil, err
	}
	cfg.Path = filepath.ToSlash(rel)
	return cfg, nil
}

// Charts lists the distinct chart directories the config names, in declaration
// order. Two releases of the same chart - the usual dev and prod pair - are one
// entry.
func (c *Config) Charts() []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range c.Releases {
		dir := strings.TrimSuffix(filepath.ToSlash(r.Chart), "/")
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		out = append(out, dir)
	}
	return out
}
