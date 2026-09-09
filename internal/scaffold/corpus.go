package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

// corpusSkillDir is where the platform corpus is written in a repository.
//
// **Under `.agents/skills/` on purpose**: that is where an agent in a customer
// repository finds what it knows, and a copy anywhere else is one it has to be
// told about. `shipped` already covers that prefix, so these files get the
// stamp and the five states with no rule of their own.
//
// Why a copy exists at all, and why that reverses what the corpus used to say,
// is in `internal/wiki/README.md` beside the rule it replaced. Not repeated
// here.
const corpusSkillDir = ".agents/skills/asgard-platform"

// corpusJobs writes the wiki and the extracts into a repository as plain files.
//
// **Nothing here is interpolated.** A version string in any of these files
// would change their bytes on every release, and what finds a stale repository
// is a byte comparison - so every repository in the world would report behind
// on a release that touched no page. Same failure as the one at the top of
// stamp.go, from the other side.
func corpusJobs() ([]job, error) {
	var jobs []job
	add := func(target string, body string) {
		jobs = append(jobs, job{target: filepath.Join(corpusSkillDir, target), body: []byte(body)})
	}

	// Not List: a corpus hides its own bookkeeping from a listing, and some of
	// that bookkeeping is exactly what an exported copy needs. Writing List's
	// view of the wiki produced a SKILL.md pointing at an index that was not
	// there.
	//
	// **But not all of it, either.** `wiki.Landing` drops the log, because the
	// two unlisted wiki documents are not alike: the index is the map and has
	// to travel, while the log is provenance for whoever maintains this
	// repository - the wiki's own index says an FDE looking for an answer
	// should never land there. Landing it also put two warnings in every
	// customer repository, for historical entries that name a command the tool
	// has since removed and are correct to.
	for _, half := range []struct {
		dir  string
		all  func() ([]kb.Doc, error)
		read func(string) (string, error)
	}{
		{"wiki", wiki.Landing, wiki.Read},
		{"usecase", usecase.All, usecase.Read},
	} {
		docs, err := half.all()
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", half.dir, err)
		}
		for _, d := range docs {
			body, err := half.read(d.Name)
			if err != nil {
				return nil, fmt.Errorf("read %s %s: %w", half.dir, d.Name, err)
			}
			add(filepath.Join(half.dir, d.Name+".md"), body)
		}
	}

	// Two files the wiki keeps outside `pages/`, so they are not in All. The
	// alias index goes to the skill root because it applies to both halves and
	// is the first file to read: a query in the customer's own words matches
	// nothing in an English corpus, and grep reports that identically to a
	// subject the material genuinely lacks.
	for _, f := range []struct {
		target string
		read   func() (string, error)
	}{
		{"aliases.md", wiki.Index},
		{filepath.Join("wiki", "README.md"), wiki.Conventions},
	} {
		body, err := f.read()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.target, err)
		}
		add(f.target, body)
	}

	add("SKILL.md", corpusSkill)
	return jobs, nil
}

// corpusSkill makes the directory discoverable and says what the exported form
// needs that the pages themselves do not carry. What the two halves are for is
// in the READMEs written beside it, so it does not restate them.
const corpusSkill = `---
name: asgard-platform
description: The Asgard platform as greppable files - what the platform has, which CR a UI name maps to, and how each deployment shape is assembled field by field. Use when writing or reading an Asgard CR or Helm chart, when a customer names something and you need to know what it maps to, or before answering any question about what the platform can do. Read aliases.md first when the question came in a language other than English.
---

# The Asgard platform, as files

The platform knowledge an agent in a customer repository does not otherwise
have. **That repository describes one customer's systems and never the platform
those systems run on**; this is the missing half.

    wiki/       what the platform has, and which CR a UI name maps to
    usecase/    how ONE deployment shape is assembled, field by field
    aliases.md  what a customer said -> what to search for

` + "`wiki/index.md`" + ` and ` + "`usecase/README.md`" + ` group their own documents by the question
each answers, and ` + "`wiki/README.md`" + ` says what a page has to carry. Start at one of
those rather than here.

**It is generated. Editing it is meaningless** - ` + "`asgard-cli init`" + ` writes it from
the corpus inside that binary and the next run replaces it, so an edit is a
claim about the platform that no other engagement sees. A page that is wrong is
worth an issue:

    asgard-cli issue-report --new

## Grep it

    grep -ril "<term>" wiki/ usecase/
    grep -n "<term>" wiki/processors.md

**Read ` + "`aliases.md`" + ` first if the question did not arrive in English.** The corpus is
English and a customer conversation usually is not, so a term taken from what
somebody actually said matches nothing - and grep reports that identically to a
subject the material genuinely lacks.

## What grep does not do

` + "`asgard-cli find <terms>`" + ` searches the same material. Shell out to it when any of
these matters:

- **It translates the query** using ` + "`aliases.md`" + ` and prints what it searched for.
  This is the one that decides whether a customer's word lands at all.
- **It warns about a word with two senses here.** ` + "`payment`" + ` matches both the
  billing between Asgard and this customer and the customer's own payment
  gateway. Both hits are correct and nothing contradicts anything, so a reader
  who cannot tell them apart takes the wrong one - a failure grep cannot
  surface, because it looks like a successful search.
- **It names the counterpart** of what it found: the extract for a page, the page
  for an extract. That is parsed from the documents, not written in them.
- **It records a query that landed nowhere**, and ` + "`asgard-cli reading --misses`" + `
  reads those back. A grep that finds nothing is silent, so the gap reaches
  nobody who could fill it.

You know the term, you want every mention, you want to read a whole page: grep
is the better tool and needs no subprocess.

## Staleness

These files came from one binary and a newer one may carry different pages.
Nothing in this directory can tell you which:

    asgard-cli init

That compares what is here against the running binary and reports five states.
**` + "`ahead`" + ` is the one worth knowing**: these files were written by a newer
build of this CLI than the one you are running, so your binary is the stale half
and ` + "`--force`" + ` would be a downgrade.

The platform's own reference material is a separate half with its own
record - ` + "`asgard-cli skill status`" + ` - because a customer's server can be several
versions from this CLI in either direction, and only the server can say what it
accepts.
`

// corpusPrefix is the stamp-key prefix of everything corpusJobs writes.
var corpusPrefix = stampKey(corpusSkillDir) + "/"

// replaceCorpus removes the exported corpus when this repository's copy was
// written by a different version of this CLI, so that the write which follows
// lays it down fresh. It reports whether it removed anything.
//
// **This is the one place a scaffold deletes from a customer's repository**, and
// the exception is narrow on purpose. Everywhere else a file this binary no
// longer ships is reported as `Retired` and left, because somebody may have come
// to rely on it. A wiki page is different in one way that settles it: the whole
// directory is declared generated in its own SKILL.md, so nothing in it is
// anybody's work - and a page that was renamed upstream would otherwise leave
// both names on disk, where grep returns the old one with nothing marking it
// stale. Merging cannot fix that; only replacing can.
//
// Three conditions, all of which must hold:
//
//   - The stamp has records under the prefix. **An unrecorded directory is
//     somebody else's**, and this must not delete a directory it cannot prove
//     it wrote.
//   - Some record's CLI version differs from the running one. Same version means
//     the material is this binary's already, and the byte comparison in the main
//     loop handles an edit to it.
//   - The path is a directory rather than a symlink, so the removal cannot
//     escape the repository by following one out.
func replaceCorpus(root string, recorded map[string]Entry, running string) (bool, error) {
	var ours, differ int
	for key, rec := range recorded {
		if !strings.HasPrefix(key, corpusPrefix) {
			continue
		}
		ours++
		if rec.CLIVersion != running {
			differ++
		}
	}
	if ours == 0 || differ == 0 {
		return false, nil
	}

	dir := filepath.Join(root, filepath.FromSlash(corpusSkillDir))
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect %s: %w", corpusSkillDir, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory, so it will not be replaced", corpusSkillDir)
	}
	if err := os.RemoveAll(dir); err != nil {
		return false, fmt.Errorf("replace %s: %w", corpusSkillDir, err)
	}
	return true, nil
}
