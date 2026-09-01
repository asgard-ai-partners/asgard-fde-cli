// Package deploy reads projects/<project>/deploy.yaml.
//
// That file is the single source of truth for where a project deploys: CI builds
// its matrix from it, and everything local that renders a chart has to read the
// same namespace and the same values overlay, or a local render is not what CD
// will apply. One reader here is what keeps them the same.
package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileName is the fixed name inside a project directory.
const FileName = "deploy.yaml"

// File is the subset of deploy.yaml anything outside CI needs.
type File struct {
	Environments map[string]Environment `yaml:"environments"`
}

// Environment is one deployment target.
type Environment struct {
	Namespace string `yaml:"namespace"`
	Values    string `yaml:"values"`
}

// ErrNotDeclared reports that a project deliberately does not deploy to an
// environment. It is not a defect: the two environments are independent, and a
// project declares only the ones it wants.
type ErrNotDeclared struct {
	Project string
	Env     string
}

func (e *ErrNotDeclared) Error() string {
	return fmt.Sprintf("projects/%s/%s does not declare %s, so this project deliberately does not deploy there; add %s under environments (namespace + values) to change that",
		e.Project, FileName, e.Env, e.Env)
}

// Path is the file's location under the repository root.
func Path(root, project string) string {
	return filepath.Join(root, "projects", project, FileName)
}

// Load reads and parses a project's deploy.yaml.
func Load(root, project string) (*File, error) {
	path := Path(root, project)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("no projects/%s/%s; it is the single source of truth for deployment targets. Run `asgard-cli scaffold` to write it", project, FileName)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var parsed File
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("projects/%s/%s is not valid YAML: %w", project, FileName, err)
	}
	return &parsed, nil
}

// Target returns the environment's declaration, or ErrNotDeclared.
func (f *File) Target(project, env string) (Environment, error) {
	target, ok := f.Environments[env]
	if !ok {
		return Environment{}, &ErrNotDeclared{Project: project, Env: env}
	}
	if target.Namespace == "" {
		return Environment{}, fmt.Errorf("projects/%s/%s: %s has no namespace", project, FileName, env)
	}
	if target.Values == "" {
		return Environment{}, fmt.Errorf("projects/%s/%s: %s has no values file", project, FileName, env)
	}
	return target, nil
}
