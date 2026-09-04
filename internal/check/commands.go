package check

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// knownCommands is what this build of the binary answers to, set by the command
// tree before Run. Empty means the check does not run: a build that forgot to
// set it must not report every command in the repository as unknown.
var knownCommands map[string]bool

// replaced maps a command this tool no longer has to what to type instead, set
// by the command tree alongside SetKnownCommands.
var replaced map[string]string

// SetReplacements records what a removed command became.
//
// **This exists so the tool answers the question instead of its maintainer.**
// The first repository this check ran against had twelve references to a
// removed command, and the agent that found them had to ask what to replace
// them with - which meant the mapping lived in a conversation rather than in
// the binary, and the next engagement would have had to ask again.
func SetReplacements(m map[string]string) { replaced = m }

// SetKnownCommands records the commands this binary has, including aliases.
//
// It is a package variable for the same reason `stage.SetOverrideDir` is: the
// list belongs to the cobra tree in internal/cli, and internal/check cannot
// import that without a cycle.
func SetKnownCommands(names []string) {
	knownCommands = make(map[string]bool, len(names))
	for _, n := range names {
		knownCommands[n] = true
	}
}

// invocation matches this tool naming itself and then a word. Only the first
// word after the binary is looked at: `asgard-cli request add` names the
// command `request`, and whether `add` is one of its subcommands is a question
// this check deliberately does not ask - a wrong subcommand is a typo, and a
// wrong command is a rename nobody was told about.
var invocation = regexp.MustCompile(`asgard-cli\s+([a-z][a-z0-9-]*)`)

// scanned are the files scaffold writes that carry command names, plus the two
// records an engagement keeps in prose.
var scanned = []string{
	"AGENTS.md",
	"README.md",
	"docs",
	"requirements",
	"projects",
	".agents",
}

// checkCommands reports a command name in this repository's own documents that
// this binary does not have.
//
// **This is the half of the material nobody could see.** `scaffold` writes
// AGENTS.md, the design-time skills and four READMEs into the customer's
// repository and then never overwrites them - which is right, because an FDE
// edits them - so a command renamed in the tool leaves every repository already
// scaffolded pointing at something that no longer exists. Nothing detected
// that. It was found by an engagement typing a command and getting
// `unknown command`, then guessing what the new name might be.
//
// `audit-material` audits the material this binary ships. This audits the same
// material after it has been written into a repository and edited, which is
// where it is actually read.
//
// A warning: a stale pointer costs a reader one confused minute and nothing in
// the chart, and a repository mid-onboarding should not go red for it.
func (c *checker) checkCommands() error {
	if len(knownCommands) == 0 {
		return nil
	}

	type hit struct {
		file string
		line int
		name string
	}
	var found []hit

	for _, entry := range scanned {
		root := filepath.Join(c.root, entry)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// A missing one is not a defect here: checkDocs and
				// checkRequirementIndexes own which files have to exist.
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(c.root, path)
			if err != nil {
				rel = path
			}
			for i, line := range strings.Split(string(data), "\n") {
				for _, m := range invocation.FindAllStringSubmatch(line, -1) {
					if knownCommands[m[1]] {
						continue
					}
					found = append(found, hit{filepath.ToSlash(rel), i + 1, m[1]})
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if len(found) == 0 {
		return nil
	}

	// Grouped by the command rather than by the file: a rename produces one
	// name in many files, and the reader needs the name to know what to
	// replace it with.
	byName := map[string][]string{}
	for _, h := range found {
		byName[h.name] = append(byName[h.name], fmt.Sprintf("%s:%d", h.file, h.line))
	}
	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, n := range names {
		where := byName[n]
		sort.Strings(where)
		instead := "`asgard-cli --help` lists what this build answers to"
		if r, ok := replaced[n]; ok {
			instead = r
		}
		c.warnf("this repository's documents name `asgard-cli %s` %d time(s) and this build has no such command - %s. "+
			"%s. `scaffold` never overwrites a file it has already written, which is right - you edit them - so a command "+
			"renamed in the tool leaves this repository pointing at the old name and nothing else notices",
			n, len(where), strings.Join(where, ", "), instead)
	}
	return nil
}
