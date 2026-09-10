# How it is built

**Why each mechanism is the way it is.** What the tool is for is
[Goal.md](Goal.md); where it stands is [TASK.md](TASK.md); what the commands do
is [README.md](README.md); what lives in which directory is
[STRUCTURE.md](STRUCTURE.md); and the rules for changing any of it are
[AGENTS.md](AGENTS.md). This file is the layer none of those covers: the
mechanisms the tool is actually made of, and the failure each one was built
from.

Read it before changing retrieval, the pointer form, the audits, or what `init`
writes. Every one of those has a reason that is not visible from the code.

## 1. One corpus abstraction, six bodies of material

`kb.Corpus` is the only reader. Four packages declare one -
`internal/wiki`, `internal/usecase`, `internal/stage`, `internal/scaffold` -
and `needs` and `brief` render documents from Go structs into the same shapes.

    kb.Doc     Name, Title, Summary, Checked, Unchecked, Links, Sources
    kb.Link    Kind, Name, Path, Deliberate
    kb.Corpus  FS, Dir, Unlisted, Noun, Command

**The rule is that there is no second reader.** A body of material with its own
reader, its own search and its own match type is a fifth thing that drifts, and
this has already happened in small: `needs` and `brief` each grew a private
copy of the rule that turns a recorded source into a pointer, and the copies
disagreed about how many kinds existed - so the same `From` rendered as a path
in one and an invocation in the other, months after the kind it named began
landing. That rule is `kb.Landed` now, in one place, because **the set of kinds
is a fact about the material.**

`Unlisted` is the corpus's own bookkeeping - an index, a log, a README.
Readable by name, absent from a listing. Two things went wrong there and both
are worth knowing: `List()` hid the index from the export, so the shipped
`SKILL.md` pointed at a map that had not been written; and `material()` audited
`List()`, so five references to a deleted command sat in the three unlisted
files while the audit reported zero. **A listing decision is not an auditing
decision and not an export decision** - `All`, `Landing` and `List` are three
answers because there are three questions.

## 2. A pointer is data, and it has two forms

A cross-reference is parsed when the document is parsed, not when a result is
printed. Before that it was two regular expressions at two points of use - one
in `find` to name a counterpart, one in the audit to check the same pointer -
which could disagree, and left the corpus with no link graph at all. Without
one the lint the whole design asks for cannot be written: **material nothing
points at is not read, and the writer never finds out, because the file is
there.**

Which form a pointer takes depends on whether its target is written into a
repository:

    ../wiki/<name>.md            a document that lands
    asgard-cli guide <name>      one that does not

Everything lands today except the wiki's `log`, so the second form is nearly
gone. What decides the `../` is where the pointing document sits, and the
readers are not in one place: a document inside a directory writes `../`,
`aliases.md` and the map are at the root and write none, a design-time skill
beside the corpus writes `../asgard-platform/`, and a file deeper in a
customer repository writes the path from the root, because a chain of `../`
from there is not something anybody should have to count. What is fixed is the
tail - the kind, a lower-case name, `.md` - and the lower case is what keeps a
prose mention of `wiki/README.md` from becoming a pointer.

**Paths were bought with a repository move.** `asgard-cli usecase write-path`
means the same thing from anywhere; a path does not. The two trees had to agree
first, which is why `internal/corpus/` holds the material in the layout a
repository receives it, and it is the only reason that package exists.

**A path is a claim that the target lands**, and the audit enforces it. The
wiki's `log` is deliberately not written into a repository, so converting its
pointer produced a link that resolved perfectly here and went nowhere in the
repository the material had been written into. Nothing saw it, because the
audit resolves against the corpus, where log exists. `kb.Link.Path` and the
check in `checkLinks` exist for that one case, and it has caught it twice.

## 3. Four mechanical audits, each catching what the others cannot

    audit-material --links      every pointer resolves
    audit-material --orphans    what nothing points at
    audit-material --bare       a document named without a way to reach it
    audit-material --commands   every command named exists

**`--links` and `--orphans` are two halves of one thing.** A pointer that goes
nowhere is loud: the reader follows it and finds nothing. A document nothing
points at is silent, and costs more - it is there, it is correct, and it is
never read. The index deliberately does not count as a pointer in the second,
because `wiki operations` sat in it under the title Connectivity while an FDE
spent a day on connectivity and never opened it.

**`--commands` is `--links` pointed at the tool.** A document that tells
somebody to run something makes a checkable claim, and six documents once named
a command nobody had built. Its coverage is the thing to watch: it reads the
material and the scaffold templates, and **not this CLI's own Go strings** - so
when the reader commands were deleted, fourteen printed strings kept naming
them and the audit reported zero dead. That is recorded in TASK.md rather than
fixed here.

**A finding is not noise because it is inconvenient.** Thirty-six gate findings
on production charts were dismissed as configuration once; reading them found
that R1b was a real bug, counting a semantic layer and a Toolset as capability
sources and not a SkillSet, so every subagent of a flow-agent supervisor was
told it had none.

## 4. Retrieval: translate, search, then warn

`find` is the way in, and it does four things a grep cannot.

**It translates the query.** The corpus is English and a customer conversation
usually is not, so a term taken from what somebody actually said matches
nothing - and a search that finds nothing reads exactly like a subject the
material lacks. `internal/corpus/aliases.md` is two tables: words that replace
the query term, and names that are added to it. A row gets there because
somebody searched for it and it landed nowhere, which `find` records to
`docs/.find-misses`; **that file is never committed**, because a query carries
whatever words the customer used.

**It warns on a word with two senses here.** `find payment` returns Fehu, which
is billing between Asgard and the customer, to somebody asking about the
customer's own payment gateway. Both hits are correct, nothing contradicts
anything, and the reader takes the wrong one - which is the failure the miss log
is blind to, because a search that lands is not a miss. The senses come from the
glossary's first table and fire on success.

**It names the counterpart**, parsed from the documents rather than written in
them, and **it records a dead query**, which is the entry to the issue path.

**It reaches four of the six bodies.** `needs` and `brief` became documents
after `parts()` was written and were never added, so `find allowlist` misses
the row about Asgard's outbound addresses - the second point of Goal.md
unreachable from the first. TASK.md has what the fix costs; the reason it is
not one line is that `part` wants a search and a list over documents that have
no files, because they are rendered from Go.

No embeddings and no vector index. The synthesis happens once, into a document,
rather than on every query - that is the whole of what
[llm-wiki](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
buys and the only reason the form was chosen.

## 5. Landing: the corpus becomes files in the customer's repository

`asgard-cli init` writes the material into
`.agents/skills/asgard-platform/` - five kinds, a generated `index.md`, and the
alias index - because that is where an agent working in a customer repository
finds what it knows, and `grep` answers "which document says this" with no
subprocess and no ranking pass.

**This reverses a rule, and the reason it stopped applying is the point.** The
corpus said these pages are never written into a customer repository, because
"a copy inside one engagement goes stale where nobody is looking". True for as
long as nothing could see the staleness. `scaffold.Stamp` can: a digest and a
CLI version **per file**, committed beside the material, so a report can tell a
repository that is behind from one somebody edited from one written by a newer
binary than the one now reading it - and `--force` refuses that last case
rather than downgrading it. The objection was to an invisible stale copy, not
to a copy.

Five states follow from that record: `behind`, `edited`, `ahead`, `stale`,
`retired`. `edited` is the one the record was built for - before it existed the
report called an engagement's own answers "yours are older" and offered
`--force`, which would have deleted them.

The corpus subtree has one rule of its own: **a version change replaces it
outright.** It is the only delete this tool performs, and it is narrow -
`replaceCorpus` requires that the stamp has records under that prefix, that
some record's version differs, and that the path is a directory rather than a
symlink. It is right there and nowhere else because the whole directory is
generated, and because a page renamed upstream would otherwise leave both names
on disk, where grep returns the old one with nothing marking it stale.

**`AGENTS.md` is the file that is half ours.** It ships sections an engagement
is told, in the file, to fill in. A managed region marks the half this CLI
owns, and that region is replaced on every run even in a file somebody edited.
Adding it introduced a data-loss bug worth remembering: the merge records a
digest, the next run reads that as proof the whole file is ours, and the
whole-file branch wrote away every answer above the marker. **A file with a
managed region is never rewritten whole** - the record was not wrong, it just
did not mean what the branch took it to mean.

Nothing written is interpolated. A version string in any landed document would
change its bytes on every release, and staleness is found by byte comparison,
so every repository in the world would report behind on a release that touched
no page.

## 6. Verification: the half a client can do

`asgard-cli gate` runs everything this machine can check, in order: `tools`,
`repo`, `shipped`, `binding`, `skills`, `lint`, `render`, `verify`. **A skip is
not a pass**, and the two are printed differently, because two steps need the
platform.

`verify` reads the rendered CRs against each other and against tables pinned
from the CRDs - enums in `internal/gate/enums.go`, field constraints in
`constraints.go`. Both can go stale in one direction only, so both are
warnings.

**Those tables are keyed by kind and path, not by field name, and that was a
bug.** A name is not a location: `Loader.spec.schedule` is an unconstrained
string and was being held against `Trigger`'s cron pattern, so a production
Loader running `00 09 * * *` was reported. Nine entries were over-broad with
one shape - the constraint had been read off the runtime kind and applied to
the authoring kind. `hack/check-tables.py` exists to hold the tables against
the generated CRDs and could not see it, because it compared a name's
*constrained* occurrences against each other and never counted the bare ones.

**It does not reproduce the platform's checks and must not.** Whether a CR is
admitted is decided by an apiserver and no client is ever issued credentials
for one, so a copy of those rules here would drift from the server the first
time either changed while still missing the two that matter most: a field the
CRD silently prunes, and a rejection only the apiserver can produce. Forty of
the seventy-nine CEL rules are `self == oldSelf`, comparing a proposal against
the object already on the cluster; a render is one object with no history, so
nothing offline can see them.

A green gate means "worth pushing", never "this will deploy". The authority is
the plan, and reading it back is
`asgard-cli pipeline runs watch --release <name> --ref <tag>`.

## 7. Chart authoring

`asgard-cli add <kind> <name>` writes a CR skeleton into a project's chart -
ten kinds, declared in `generate.Kinds`, each naming the wiki page and the
extract that explain it. Those two names are pointers like any other and
`--links` resolves them, which is the half a prose search cannot reach: a kind
pointing at a renamed extract goes unnoticed until somebody runs `add` and
follows it.

What the generator writes is the conventions applied - naming, the display
annotations, what belongs in values and what stays in the template - so an edit
that departs from them is the half a rendered chart still passes.

Deliberately out of scope, and Goal.md says why: the namespace and the
environment id, which the platform injects as `.Values.asgard.*` on every run,
and whether the thing deploys at all.
