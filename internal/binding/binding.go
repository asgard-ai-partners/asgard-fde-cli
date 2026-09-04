// Package binding reads and writes `.asgard-cli.yaml`, the record of which
// platform objects a checkout deploys through.
//
// IT IS CLIENT-SIDE ONLY. The platform never reads it. What a run reads from a
// repository is the declaration at the pipeline's config path and the chart it
// names, and nothing else — so this file cannot make a deployment succeed or
// fail, and nothing in it is validated by anything but this CLI.
//
// That matters beyond tidiness. A repository may carry several pipelines, one
// per declaration, and the platform's uniqueness rule is (workspace, repo,
// config path). A file the server enforced would close that door; a file only
// the client reads leaves it open.
//
// WHAT IT IS FOR is the one thing neither the repository nor the platform can
// answer on its own: which workspace. The pipeline follows from the origin
// remote in every case but a repository carrying more than one, and the
// releases, their keys and their triggers are all in the declaration. So this
// holds two ids and nothing else — a third field would be a second source of
// truth for something the platform already owns.
//
// ITS SECOND READER IS A CODING AGENT. It sits in the working tree, so an agent
// landing in a fresh clone can see what this checkout is pointed at without
// being told and without a call, which is why the file carries a comment header
// explaining itself rather than being the smallest possible YAML.
package binding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/pipelineconfig"
)

// FileName is the binding's name, beside the declaration it belongs to.
const FileName = ".asgard-cli.yaml"

// ErrNotFound reports that a checkout records no binding.
var ErrNotFound = errors.New("no " + FileName)

// File is the whole of it.
type File struct {
	// Version is 1.
	Version int `yaml:"version"`
	// Workspace is the platform workspace id this checkout deploys into. It is
	// the field that has to exist: nothing in the repository implies it.
	Workspace string `yaml:"workspace"`
	// Pipeline is the pipeline id. It is derivable from the origin remote
	// whenever the repository carries only one, and recorded anyway so that a
	// repository which grows a second does not become ambiguous.
	Pipeline string `yaml:"pipeline,omitempty"`

	// Path is where it was read from. Not part of the file.
	Path string `yaml:"-"`
}

// header is written above the data. It is most of the file, deliberately: the
// second reader is an agent that has never seen one before.
const header = `# .asgard-cli.yaml
#
# Which platform objects this checkout deploys through. Written by
# ` + "`asgard-cli workspace use`" + `, and committed: whoever clones this repository, and
# whatever agent works in it, should not have to be told again.
#
#   asgard-cli workspace show    which workspace is in effect, and why that one
#   asgard-cli pipeline show     what the pipeline reads as its declaration
#
# THE PLATFORM NEVER READS THIS FILE. A run reads the declaration at the
# pipeline's config path and the chart it names; nothing here can make one
# succeed or fail. It exists so that the commands below need no --workspace.
#
# It records only what nothing else can answer. Which releases exist, which keys
# they take and when they deploy are all in .asgard-pipeline.yaml, which is the
# declaration and the only file that decides anything.
`

// Locate finds the declaration governing dir and the binding beside it.
//
// The binding lives next to the declaration rather than at the repository root
// because a repository may hold several of each: a monorepo with one
// declaration per team has one pipeline per team, and a single file at the root
// could name only one of them.
//
// declPath is empty when no declaration was found. bindPath is where the
// binding is or would be written, which is why it is returned either way.
func Locate(dir string) (declPath, bindPath string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("resolve %s: %w", dir, err)
	}

	for {
		candidate := filepath.Join(dir, pipelineconfig.FileName)
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return candidate, filepath.Join(dir, FileName), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", nil
		}
		dir = parent
	}
}

// Load reads the binding at path.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s: %w", path, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("%s is not valid YAML: %w", path, err)
	}
	if f.Version != 0 && f.Version != 1 {
		return nil, fmt.Errorf("%s declares version %d, which this build does not understand; upgrade asgard-cli", path, f.Version)
	}
	f.Path = path
	return &f, nil
}

// LoadFrom finds and reads the binding governing dir.
//
// A checkout with no declaration, and one with a declaration but no binding,
// both report ErrNotFound: neither is a failure, and both mean the same thing
// to a caller - nothing here says which workspace.
func LoadFrom(dir string) (*File, error) {
	declPath, bindPath, err := Locate(dir)
	if err != nil {
		return nil, err
	}
	if declPath == "" {
		return nil, ErrNotFound
	}
	return Load(bindPath)
}

// Save writes the binding, header and all.
func Save(path string, f *File) error {
	if f.Version == 0 {
		f.Version = 1
	}
	body, err := yaml.Marshal(struct {
		Version   int    `yaml:"version"`
		Workspace string `yaml:"workspace"`
		Pipeline  string `yaml:"pipeline,omitempty"`
	}{Version: f.Version, Workspace: f.Workspace, Pipeline: f.Pipeline})
	if err != nil {
		return fmt.Errorf("encode the binding: %w", err)
	}

	// Written through a temporary file in the same directory so an interrupted
	// write cannot leave half a binding where a whole one was.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, FileName+".*")
	if err != nil {
		return fmt.Errorf("create a temporary file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(header + "\n" + string(body)); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpName, err)
	}
	// 0644 rather than 0600: it is committed and read by everyone working in
	// the repository, and it holds no secret.
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("set the mode of %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
