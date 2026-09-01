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
)

// FileName is the fixed name of the config file in the project root.
const FileName = ".asgard-config.json"

// ErrNotFound reports that no config file exists at the given location or above it.
var ErrNotFound = errors.New("no " + FileName + " found")

// Workspace is the workspace this directory is bound to.
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Config is the full contents of .asgard-config.json.
type Config struct {
	Workspace Workspace `json:"workspace"`
}

// Validate checks that the required fields are present. It joins every problem
// with errors.Join so callers such as lint can unwrap and report them one by one.
func (c *Config) Validate() error {
	var errs []error
	if c.Workspace.ID == "" {
		errs = append(errs, errors.New("workspace.id must not be empty"))
	}
	if c.Workspace.Name == "" {
		errs = append(errs, errors.New("workspace.name must not be empty"))
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
