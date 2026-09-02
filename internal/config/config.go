// Package config reads and writes .asgard-config.json.
//
// Every command should go through this package instead of assembling JSON of
// its own, so a field change only has to be made in one place.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileName is the fixed name of the config file in the project root.
const FileName = ".asgard-config.json"

// RepoSuffix is appended to the workspace slug to form the repository name.
const RepoSuffix = "-asgard-kube"

// MaxNamespaceLength is the Kubernetes limit for a namespace name. Names derived
// from a namespace inherit its length, which is why the target repo keeps slugs
// deliberately short.
const MaxNamespaceLength = 63

// ErrNotFound reports that no config file exists at the given location or above it.
var ErrNotFound = errors.New("no " + FileName + " found")

// Env is a deployment environment. The platform accepts exactly two symbols, and
// a project may declare either or both.
type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

// Envs lists every valid environment, in deployment order.
var Envs = []Env{EnvDev, EnvProd}

// Valid reports whether e is one of the two accepted symbols.
func (e Env) Valid() bool {
	for _, known := range Envs {
		if e == known {
			return true
		}
	}
	return false
}

// slugPattern is the DNS-1123 label rule. A slug ends up inside a Kubernetes
// namespace, so anything it rejects would be rejected later by the apiserver.
var slugPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// ValidateSlug checks that s can be used as part of a namespace name.
func ValidateSlug(field, s string) error {
	switch {
	case s == "":
		return fmt.Errorf("%s must not be empty", field)
	case !slugPattern.MatchString(s):
		return fmt.Errorf("%s %q must be lowercase letters, digits and hyphens, starting and ending with a letter or digit", field, s)
	}
	return nil
}

// Workspace is the customer this repository serves. The slug is what repository
// and namespace names are built from; the id is issued by the Asgard platform.
//
// The id may be empty. Nothing this CLI generates reads it - namespaces come
// from the slug, and the charts never mention it - so requiring it at init only
// blocked work that had not reached the platform yet. It is recorded because the
// repository should say which workspace it belongs to, and the moment that
// starts to matter is the first project.
type Workspace struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// HasID reports whether the platform's workspace id has been filled in.
func (w Workspace) HasID() bool { return w.ID != "" }

// Project is one Helm chart deployed to one namespace per environment.
type Project struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Environments []Env  `json:"environments"`
}

// Config is the full contents of .asgard-config.json.
type Config struct {
	Workspace Workspace `json:"workspace"`
	Projects  []Project `json:"projects"`

	// OLAPOnlyLayers are semantic layers deliberately bound to no Agent,
	// because they exist to feed Data Insight rather than a chat agent: their
	// cross-schema joins are unrestricted, which is the point of a data lake and
	// exactly the reason not to hang an agent off one.
	//
	// It is per customer, and it lives here rather than as a flag somebody has
	// to remember on every run: `asgard-cli verify` reads it, and a policy that
	// only applies when you pass the right argument is not a policy.
	OLAPOnlyLayers []string `json:"olapOnlyLayers,omitempty"`
}

// RepoName is the repository this onboarding produces, derived rather than
// stored so it cannot drift from the slug.
func (c *Config) RepoName() string {
	return c.Workspace.Slug + RepoSuffix
}

// Namespace is where projectSlug deploys in env, following the convention
// asgard-<workspace>-<project>-<env>.
func (c *Config) Namespace(projectSlug string, env Env) string {
	return fmt.Sprintf("asgard-%s-%s-%s", c.Workspace.Slug, projectSlug, env)
}

// Project returns the project with the given slug.
func (c *Config) Project(slug string) (*Project, bool) {
	for i := range c.Projects {
		if c.Projects[i].Slug == slug {
			return &c.Projects[i], true
		}
	}
	return nil, false
}

// Validate checks every field the rest of the CLI depends on. It joins all the
// problems with errors.Join so a caller can report them one by one rather than
// making the user fix them one round trip at a time.
//
// workspace.id is deliberately not checked. It is not validated for shape,
// because the platform will grow an API to verify it and a guessed pattern would
// reject valid ids today; and it is not required to be present, because nothing
// downstream reads it. Commands that are a good moment to fill it in say so
// instead - see Workspace.HasID.
func (c *Config) Validate() error {
	var errs []error

	if err := ValidateSlug("workspace.slug", c.Workspace.Slug); err != nil {
		errs = append(errs, err)
	}
	if c.Workspace.Name == "" {
		errs = append(errs, errors.New("workspace.name must not be empty"))
	}

	seen := make(map[string]int, len(c.Projects))
	for i, p := range c.Projects {
		where := fmt.Sprintf("projects[%d]", i)

		if err := ValidateSlug(where+".slug", p.Slug); err != nil {
			errs = append(errs, err)
		} else if first, dup := seen[p.Slug]; dup {
			errs = append(errs, fmt.Errorf("%s.slug %q duplicates projects[%d]", where, p.Slug, first))
		} else {
			seen[p.Slug] = i
		}

		if p.Name == "" {
			errs = append(errs, fmt.Errorf("%s.name must not be empty", where))
		}
		errs = append(errs, validateEnvironments(where, p.Environments)...)

		// Namespaces are derived, so an over-long one is only visible here. The
		// alternative is finding out during helm upgrade in CD.
		if c.Workspace.Slug != "" && p.Slug != "" {
			for _, env := range p.Environments {
				ns := c.Namespace(p.Slug, env)
				if len(ns) > MaxNamespaceLength {
					errs = append(errs, fmt.Errorf("%s: namespace %q is %d characters, over the %d limit; shorten the workspace or project slug",
						where, ns, len(ns), MaxNamespaceLength))
				}
			}
		}
	}

	return errors.Join(errs...)
}

func validateEnvironments(where string, envs []Env) []error {
	if len(envs) == 0 {
		return []error{fmt.Errorf("%s.environments must declare at least one of %s", where, envList())}
	}

	var errs []error
	seen := make(map[Env]bool, len(envs))
	for _, env := range envs {
		switch {
		case !env.Valid():
			errs = append(errs, fmt.Errorf("%s.environments has %q, want one of %s", where, env, envList()))
		case seen[env]:
			errs = append(errs, fmt.Errorf("%s.environments lists %q twice", where, env))
		default:
			seen[env] = true
		}
	}
	return errs
}

func envList() string {
	names := make([]string, len(Envs))
	for i, e := range Envs {
		names[i] = string(e)
	}
	return strings.Join(names, ", ")
}

// Load reads the config file at path. When the file is missing, the returned
// error matches errors.Is(err, ErrNotFound).
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s: %w", path, ErrNotFound)
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &cfg, nil
}

// Save writes cfg to path. It writes a temp file in the same directory and then
// renames it, so a failure midway cannot leave a truncated config behind.
func Save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	// Both calls are no-ops once the rename succeeds; they clean up on failure.
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		return fmt.Errorf("chmod %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// Find walks up from dir and returns the path of the first config file it finds,
// so commands still locate the project config when run from a subdirectory.
func Find(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", dir, err)
	}

	for {
		path := filepath.Join(dir, FileName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir { // reached the filesystem root
			return "", fmt.Errorf("searching up from %s: %w", dir, ErrNotFound)
		}
		dir = parent
	}
}
