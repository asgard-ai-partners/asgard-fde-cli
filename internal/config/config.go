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

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/size"
)

// FileName is the fixed name of the config file in the project root.
const FileName = ".asgard-config.json"

// RepoSuffix is appended to the workspace slug to form the repository name.
const RepoSuffix = "-asgard-kube"

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

// Workspace is the customer this repository serves. The slug is what the
// repository is named from.
//
// It carries no platform workspace id. Which workspace a checkout deploys into
// is in `.asgard-cli.yaml`, beside the declaration it belongs to - one fact,
// one home, and a repository with two declarations needs two answers rather
// than one field here.
type Workspace struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Project is one chart. Where it deploys is not here: a release binds a chart to
// a platform project, and the declaration is what names them.
type Project struct {
	Slug string `json:"slug"`
	Name string `json:"name"`

	// Shape is the deployment shape this chart is being built to, named after
	// one of `asgard-cli size`. It decides which CR kinds finish the chart, and
	// it is recorded rather than derived because the files cannot say it: a
	// chart with a SemanticLayer and no consumer is either a Mimir deliverable
	// that is finished or an agent nobody has written yet, and those look
	// identical on disk.
	//
	// Empty means undeclared, which is not an error - most projects are the
	// common shape and every repository written before this field existed has
	// none. `internal/stage` says what it assumes when it is empty, and says it
	// out loud rather than silently.
	Shape string `json:"shape,omitempty"`
}

// Config is the full contents of .asgard-config.json.
type Config struct {
	Workspace Workspace `json:"workspace"`
	Projects  []Project `json:"projects"`
}

// RepoName is the repository this onboarding produces, derived rather than
// stored so it cannot drift from the slug.
func (c *Config) RepoName() string {
	return c.Workspace.Slug + RepoSuffix
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

		// A shape decides when the chart is finished, so a typo in one is not
		// cosmetic: it would fall back to the default ladder and the project
		// would be told to build a CR its shape does not have.
		if p.Shape != "" {
			if _, ok := size.Find(p.Shape); !ok {
				errs = append(errs, fmt.Errorf("%s.shape %q is not a shape; one of %s",
					where, p.Shape, strings.Join(size.Names(), ", ")))
			}
		}

	}

	return errors.Join(errs...)
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
