package scaffold

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// StampName is the record of the shipped material this CLI wrote into a
// repository. It is committed with the files.
//
// **It is not what notices that the material moved.** That is a byte
// comparison against the copy in this binary, and it needs no record at all:
// the embedded skills interpolate one constant, so what this binary renders for
// them is the same in every repository there is. Recording a version and
// comparing that instead would report every repository behind on every release
// that did not touch a skill - which is the failure internal/skills describes
// from the other side, where the platform's version is declared by a person
// rather than derived from the material.
//
// What a byte comparison cannot answer is **who wrote what is there**, and
// three questions need that:
//
//   - **Which way round.** Different is not behind. A repository written by a
//     newer CLI and opened by an older one differs in exactly the way a
//     repository that is behind differs, and `--force` would then hand it the
//     older material. `skills.Behind` exists for this on the platform's side.
//   - **Whether somebody wrote it.** `shipped` asserts by policy that these
//     files are never edited, and the scaffolded AGENTS.md ships a project list
//     whose every row says `TODO: what it does`. An engagement that answers it
//     has edited a shipped file - correctly - and until this record existed the
//     report called that "yours are older" and offered `--force`, which would
//     have deleted the answer.
//   - **What is no longer shipped.** `plan` produces a job per file this binary
//     carries, so a skill removed from the binary is invisible: nothing
//     compares it and it stays in the repository saying whatever it said.
//     `asgard-cr-verification` did that for a month - see the embed comment in
//     scaffold.go. This record is the only thing that can say a file was
//     written here by a CLI that no longer ships it.
//
// **It records only what `shipped` covers.** Recording the rest would claim
// authorship of the files the engagement is meant to edit, and then every
// correct repository would report a wall of edits - a checker that fires on
// correct material, which AGENTS.md says to delete rather than soften.
//
// It is at the repository root rather than under `.agents/skills/` because
// `shipped` reaches AGENTS.md as well, and a record living inside the skills
// directory would have been missing half its subject from the first commit. It
// is a separate file from `.asgard-docs.json` because that one is the
// platform's record of the platform's material: two authorities, two version
// namespaces, and `skills.Apply` rewrites its file list wholesale on every
// fetch.
const StampName = ".asgard-scaffold.json"

// Stamp is the record beside the material.
type Stamp struct {
	// WrittenAt is when the record was last changed, in RFC 3339. It is not
	// part of the identity, and a run that changes nothing leaves it alone
	// rather than churning a committed file.
	WrittenAt string `json:"written_at"`
	// Files maps each shipped path, slash-separated, to what was written
	// there. A path missing from here is a path whose authorship is unknown,
	// which is not the same as one that matches.
	Files map[string]Entry `json:"files"`
}

// Entry is one file's authorship.
//
// **The version is per file and not one field on the Stamp**, because a single
// run can write one shipped file and leave another alone: an engagement that
// answered the project list in AGENTS.md keeps it while the skills beside it
// are replaced. One version for the repository would then claim this CLI wrote
// the file it did not touch - and the next run would read that claim, find the
// versions equal, conclude the render had moved and overwrite the file this
// one deliberately preserved.
type Entry struct {
	// Digest identifies the bytes that were written.
	Digest string `json:"digest"`
	// CLIVersion is which binary wrote them, as `asgard-cli version` reports
	// it. It can be `dev`, and then nothing is ordered against it.
	CLIVersion string `json:"cli_version"`
}

// Writers lists the CLI versions the record says wrote here, sorted and
// deduplicated.
//
// A repository holding files from two versions has two, and it is reported as
// two: picking one would be the summary that hides the thing worth seeing.
func (s *Stamp) Writers() []string {
	if s == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range s.Files {
		if e.CLIVersion == "" || seen[e.CLIVersion] {
			continue
		}
		seen[e.CLIVersion] = true
		out = append(out, e.CLIVersion)
	}
	sort.Strings(out)
	return out
}

// ReadStamp reads the record at a repository root. A missing one is not an
// error: it is a repository scaffolded before this existed, which every caller
// has to handle anyway and which one run converges.
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

// writeStamp replaces the record.
func writeStamp(root string, s Stamp) error {
	body, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, StampName), append(body, '\n'), 0o644)
}

// digest identifies one file's contents in the record.
//
// It is the same arithmetic as `skills.Digest` and deliberately not that
// function. That one is the platform's digest, computed the platform's way so
// that a fetched file's recorded digest and the one the platform sends for the
// same bytes compare equal; this one answers a question no platform is party
// to. Nothing may come to depend on the two agreeing, and importing the
// platform's half of the material into the half that must never describe a
// particular server is how that dependency would start.
func digest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:12]
}

// stampKey is how a repo-relative path is spelled in the record. Slashes,
// always: a record written on Windows is read on macOS by whoever clones the
// repository next.
func stampKey(target string) string { return filepath.ToSlash(target) }

// direction says which of two CLI versions is the newer, for the one question
// a byte comparison cannot answer.
type direction int

const (
	// indeterminate: the two cannot be ordered. Not an error and not a
	// fallback - it is the honest answer for a `go build` binary and for a
	// pre-release, and it is reported as "different" rather than as a
	// direction this cannot know.
	indeterminate direction = iota
	// same version, so whatever moved is not the binary.
	same
	// older: the recorded CLI is older than the one running.
	older
	// newer: the recorded CLI is newer than the one running. With a binary
	// that updates itself this is a downgrade, and it is the case `--force`
	// must not act on.
	newer
)

// compare orders two CLI versions.
//
// **A version that is not a release is not ordered.** A plain `go build` leaves
// it at `dev`, `debug.ReadBuildInfo` can append `-dirty`, and a pseudo-version
// is not three numbers - guessing on any of those would claim a direction this
// cannot know, which is worse than the "different" it replaces.
// `skills.Behind` refuses on the same grounds and for the same reason.
//
// Textual equality is checked first, so two `dev` builds and two `-dirty` ones
// are the same version. That is what makes a maintainer's edited template
// reach a scratch repository on a plain re-run, and it is right for a release
// too: equal versions with different bytes means what moved is what the render
// reads, not the binary.
func compare(recorded, running string) direction {
	if recorded == running {
		return same
	}
	rec, recOK := releaseVersion(recorded)
	run, runOK := releaseVersion(running)
	if !recOK || !runOK {
		return indeterminate
	}
	for i := range rec {
		switch {
		case rec[i] < run[i]:
			return older
		case rec[i] > run[i]:
			return newer
		}
	}
	return same
}

// releaseVersion parses vMAJOR.MINOR.PATCH and nothing else. A pre-release or
// build suffix makes a version not comparable rather than comparable with a
// guess: the suffix is what orders two of them, and this has no reason to know
// how.
func releaseVersion(v string) ([3]int, bool) {
	fields := strings.Split(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".")
	if len(fields) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}

// classify says what one shipped file's state is: whether it is what this
// binary carries, and when it is not, who wrote what is there.
//
// rec is what the record holds for the path, zero when it holds nothing.
// running is this binary's own version.
func classify(disk, rendered []byte, rec Entry, running string) Status {
	if bytes.Equal(disk, rendered) {
		return Skipped
	}

	// **In a file with a managed region, the region is the shipped material.**
	// Everything outside it is the engagement's, and this CLI never rewrites
	// such a file whole: the moment AGENTS.md gained a region, taking it whole
	// wrote away the answers somebody had filled into the TODO sections above
	// the marker. So the comparison that decides this file's state is region
	// against region, and a difference outside one is the engagement having
	// written something - which is neither an edit to report nor a repository
	// that is behind.
	//
	// **Both callers come through here, which is the point.** Write carried
	// this rule and InspectShipped did not, so `asgard-cli init` reported
	// nothing to do about AGENTS.md while `asgard-cli gate` reported that it
	// would be updated - the same file, the same repository, two answers.
	//
	// A render that has no region while the file on disk does is left alone
	// rather than taken whole: the marker is what the engagement's half exists
	// behind, and a template that drops one is not a licence to delete what
	// accumulated behind it.
	if here := managedRegion.Find(disk); here != nil {
		if want := managedRegion.Find(rendered); want != nil && !bytes.Equal(here, want) {
			return Updated
		}
		return Skipped
	}

	if rec.Digest == "" {
		// Nobody recorded writing this: a repository scaffolded before the
		// record existed, or a file brought in by hand. It differs, and which
		// way round is not knowable.
		return Stale
	}
	if rec.Digest != digest(disk) {
		return Edited
	}
	switch compare(rec.CLIVersion, running) {
	case older:
		return Behind
	case newer:
		return Ahead
	case same:
		// This CLI wrote what is on disk and now renders something else, so
		// what moved is what the render reads - the repository's own directory
		// name, or its project list. Taking the new one discards nothing,
		// because the old one is provably ours.
		return Updated
	}
	return Stale
}
