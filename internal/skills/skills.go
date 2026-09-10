// Package skills writes the platform's reference material into a repository,
// and reports what is there against what the platform has.
//
// **The content is the server's and the location is ours.** The platform
// renders the files and gives a path relative to a skills directory; which
// directory that is, and what the stamp beside them looks like, is this
// package's decision. That split is what lets the platform change how it writes
// a reference without anybody shipping a new binary - which is the whole reason
// the material is fetched rather than embedded. See internal/scaffold for the
// long version of why embedding it would break every on-prem customer.
package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// StampName is the record of what was written and from which version.
//
// It is committed with the files. Deriving the version from the files instead
// would work for one file and stop working for two, and it could not answer the
// other question this has to answer: whether a generated file has been edited
// by hand since it was written. The files say "do not edit"; this is what
// notices when somebody did, before an update overwrites their work in silence.
const StampName = ".asgard-docs.json"

// Roots are the directories an agent's tooling reads skills from, in the order
// they are looked for. The first that exists wins; when none does, the first is
// created.
//
// `.agents/skills` is what `asgard-cli init` writes and is not tied to one
// vendor's tool. `.claude/skills` is where Claude Code looks. A repository that
// has chosen one should not acquire the other by running an update.
var Roots = []string{
	filepath.Join(".agents", "skills"),
	filepath.Join(".claude", "skills"),
}

// Stamp is the record beside the files.
type Stamp struct {
	// Version is the platform's version for the material, as it was written.
	Version string `json:"version"`
	// Platform is which API answered - a repository can be pointed at dev
	// while the material it holds came from prod, and that is worth seeing.
	Platform string `json:"platform"`
	// FetchedAt is when, in RFC 3339. It is not part of the identity.
	FetchedAt string `json:"fetched_at"`
	// Sources is each upstream's digest as it was when this was written.
	//
	// **The version alone is not enough, because the version is declared.** A
	// person on the platform side increments it when a change is worth telling
	// everybody about; the material can move without that happening - a field
	// added to a CRD, a processor gaining a config key - and then two equal
	// version numbers describe two different sets of facts. Comparing these
	// says so, and says which half moved, without downloading the material to
	// find out.
	Sources map[string]string `json:"sources,omitempty"`
	// Files maps each written path to the digest of what was written, so a
	// later run can tell a hand edit from an update.
	Files map[string]string `json:"files"`
}

// Root resolves the directory the material is written into.
//
// dir is the repository root. An explicit override wins outright: an agent
// working somewhere unusual should not have to move its files to satisfy this.
//
// **A root that already holds the material wins over one that merely exists**,
// and that order is the whole of this function. The comment on Roots says a
// repository that has chosen one should not acquire the other by running an
// update, and picking the first directory that exists did not implement it:
// `asgard-cli init` creates `.agents/skills` unconditionally, so a repository
// that had fetched into `.claude/skills` - a checkout already using Claude
// Code, where nothing had run `init` yet - silently changed which directory
// the material was read from the moment somebody scaffolded it. Everything
// fetched before that stayed on disk, in the directory the agent's own runtime
// still looks in, describing whatever server it described in August, with
// `skill status` reporting "none" and no command in the tool mentioning it.
func Root(repoRoot, override string) string {
	if override != "" {
		if filepath.IsAbs(override) {
			return override
		}
		return filepath.Join(repoRoot, override)
	}
	for _, candidate := range Roots {
		full := filepath.Join(repoRoot, candidate)
		if _, err := os.Stat(filepath.Join(full, StampName)); err == nil {
			return full
		}
	}
	for _, candidate := range Roots {
		full := filepath.Join(repoRoot, candidate)
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			return full
		}
	}
	return filepath.Join(repoRoot, Roots[0])
}

// Elsewhere lists the candidate roots that hold fetched material and are not
// the one in use, with the version each holds.
//
// **It is what notices a repository holding two copies.** One can only arise
// now from a version of this CLI that chose its root by which directory
// existed, or from somebody passing --dir - but the copy that is not in use is
// read by an agent's runtime exactly as readily as the one that is, and until
// this existed nothing in the tool said it was there. That is the same failure
// as a skill this binary no longer ships: material nobody is comparing against
// anything, in a directory nobody is looking at.
//
// inUse is the absolute path Root returned. A root that holds no record is not
// reported: an empty `.claude/skills` is a repository that uses Claude Code,
// not a second copy of anything.
func Elsewhere(repoRoot, inUse string) []Other {
	var out []Other
	for _, candidate := range Roots {
		full := filepath.Join(repoRoot, candidate)
		if full == inUse {
			continue
		}
		stamp, err := ReadStamp(full)
		if err != nil || stamp == nil {
			continue
		}
		out = append(out, Other{Dir: candidate, Version: stamp.Version, Files: len(stamp.Files)})
	}
	return out
}

// Other is one candidate root holding material that is not in use.
type Other struct {
	// Dir is the path relative to the repository root, as Roots spells it.
	Dir string
	// Version is what that copy's record says it is.
	Version string
	// Files is how many the record covers.
	Files int
}

// ReadStamp reads the record in a skills directory. A missing one is not an
// error: it means the material has never been fetched here, which is a state
// every caller has to handle anyway.
func ReadStamp(root string) (*Stamp, error) {
	body, err := os.ReadFile(filepath.Join(root, StampName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s Stamp
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("read %s: %w", filepath.Join(root, StampName), err)
	}
	return &s, nil
}

// Status is what one file's state is, against what the platform would write.
type Status string

const (
	// Created: the file is not there.
	Created Status = "created"
	// Updated: the file is there, matches what was written last time, and the
	// platform has different content now.
	Updated Status = "updated"
	// Unchanged: byte for byte what the platform has.
	Unchanged Status = "unchanged"
	// Edited: on disk, differs from the platform's, and also differs from what
	// was written here last time. Somebody edited a generated file.
	Edited Status = "edited"
)

// File is one file's plan.
type File struct {
	Path   string
	Status Status
}

// Plan reports what applying these files would do, without doing it.
func Plan(root string, files []Content, stamp *Stamp) ([]File, error) {
	out := make([]File, 0, len(files))
	for _, f := range files {
		full := filepath.Join(root, filepath.FromSlash(f.Path))
		existing, err := os.ReadFile(full)
		if os.IsNotExist(err) {
			out = append(out, File{Path: f.Path, Status: Created})
			continue
		}
		if err != nil {
			return nil, err
		}
		switch {
		case string(existing) == f.Content:
			out = append(out, File{Path: f.Path, Status: Unchanged})
		case stamp != nil && stamp.Files[f.Path] != "" && stamp.Files[f.Path] != Digest(string(existing)):
			out = append(out, File{Path: f.Path, Status: Edited})
		default:
			out = append(out, File{Path: f.Path, Status: Updated})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// Content is one file to write: the platform's path and bytes.
type Content struct {
	Path    string
	Content string
}

// Behind compares two versions.
//
// The platform's version is a declared number that only goes up, so "behind" is
// a real answer rather than "differs" - which is what a content digest could
// have said. A version that is not a number is not compared: this has to keep
// working against a platform that changes how it spells one, and guessing there
// would report every repository as behind at once.
func Behind(local, remote string) bool {
	l, lok := strconv.Atoi(local)
	r, rok := strconv.Atoi(remote)
	if lok != nil || rok != nil {
		return false
	}
	return l < r
}

// MovedSources returns the upstreams whose digest differs from what was
// recorded here, in a stable order.
//
// This is what notices a change the declared version did not announce. An empty
// result with a stamp that has no sources at all means the stamp predates them,
// not that nothing moved - the caller has to tell those apart, so this returns
// nothing for both and Stamp.Sources is what says which case it is.
func MovedSources(stamp *Stamp, remote map[string]string) []string {
	if stamp == nil || len(stamp.Sources) == 0 {
		return nil
	}
	var out []string
	for name, digest := range remote {
		if recorded, ok := stamp.Sources[name]; ok && recorded != digest {
			out = append(out, name)
		}
	}
	for name := range stamp.Sources {
		if _, ok := remote[name]; !ok {
			out = append(out, name+" (gone)")
		}
	}
	sort.Strings(out)
	return out
}

// Apply writes the files and the stamp beside them.
//
// It writes every file rather than only the changed ones. The alternative is a
// partial write that leaves a repository holding two versions at once, and the
// stamp would then be true of neither.
func Apply(root string, files []Content, stamp Stamp) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	written := map[string]string{}
	for _, f := range files {
		if err := safePath(f.Path); err != nil {
			return err
		}
		full := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(f.Content), 0o644); err != nil {
			return err
		}
		written[f.Path] = Digest(f.Content)
	}
	stamp.Files = written

	body, err := json.MarshalIndent(stamp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, StampName), append(body, '\n'), 0o644)
}

// safePath refuses a path that would write outside the skills directory.
//
// The platform is not an adversary, but this writes files into a customer's
// repository from a network response, and a path that escapes the directory is
// the one mistake in that shape that cannot be undone by running the command
// again.
func safePath(p string) error {
	if p == "" {
		return fmt.Errorf("the platform sent a file with no path")
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "/") {
		return fmt.Errorf("the platform sent an absolute path (%s), which would write outside the skills directory", p)
	}
	clean := filepath.Clean(filepath.FromSlash(p))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("the platform sent a path that leaves the skills directory (%s)", p)
	}
	return nil
}
