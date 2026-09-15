# How it is built

**How the main capabilities are implemented.** What the tool is for is
[Goal.md](Goal.md); where it stands is [TASK.md](TASK.md); what the commands do
is [README.md](README.md); what lives in which directory is
[STRUCTURE.md](STRUCTURE.md); the rules for changing it are
[AGENTS.md](AGENTS.md).

Present tense. Why something changed is `git log`.

## The corpus

Every body of material, one reader. `kb.Corpus` is that reader and all of them
declare one; `needs` and `brief` have no files, so they render documents from
Go structs into an in-memory FS and get the same treatment.

    kb.Doc     what a document says about itself
    kb.Link    one pointer out of it
    kb.Corpus  one body of material, and how to read it

**The fields are not listed here.** They are a list of what exists, the struct
is the source, and the copy that used to be here had already lost one field
before it lost two more - `go doc ./internal/kb Doc` prints them, correct at the
commit you are standing on.

| body | package | kind |
|---|---|---|
| wiki pages | `internal/wiki` over `internal/corpus/wiki/` | `wiki` |
| deployment extracts | `internal/usecase` over `internal/corpus/usecase/` | `usecase` |
| what to obtain from a customer | `internal/needs` | `needs` |
| what an activity gets wrong | `internal/brief` | `brief` |
| stage guidance | `internal/stage` over `internal/stage/prompts/` | `guide` |
| design-time skills | `internal/scaffold` over `internal/scaffold/templates/.agents/skills/` | — |

**Add material to one of these, not beside them.** A body with its own reader
and its own parse drifts from the others and nothing mechanical notices. If new
material does not fit `kb.Corpus`, change `kb`.

**Almost everything a document declares about itself is parsed from its prose**
- a `# ` title, a summary paragraph, the two provenance markers. Frontmatter
carries the exceptions, and there are two: `group:` and `description:`, which is
what `go run ./hack index` renders both indexes from - the wiki's and the
extracts'.

**Both are there because they cannot be derived, and nothing else is.** A
grouping is the question a section asks, and no parse recovers it. A description
is what somebody would come to the page FOR, which is a different claim from the
page's opening thesis - deriving one from the other was tried and gives rows like
"They are not two of the same thing". Everything an index would otherwise repeat
- the title, the path - is computed, so the only hand-written part of a row is
the part that is a judgement, and it lives on the page rather than in the index.

`Docs` and `ParseDoc` are how a body that is not one `<name>.md` per document
joins anyway: a stage is a numbered prompt file, a skill is a directory with
YAML frontmatter. Both supply their own, and everything downstream reads
`kb.Doc`.

Two questions, and they are not interchangeable:

    List()      what a reader listing documents should see
    All()       plus the corpus's own bookkeeping - an index, a README

`Unlisted` is what separates them. `wiki.Landing()` - what `init` writes - is
`All()`, because a map that does not travel with the material leaves the copy
without one. The audit's source set is the same: **a document that ships is a
document that is checked**, and using `List()` there hides every dead reference
in the files it skips.

## Pointers

A cross-reference is parsed with the document into `kb.Link`, and everything
downstream reads that one graph — `--links`, `--orphans`.
**Every document pointer is a path**, because every kind is written into a
repository:

    ../wiki/<name>.md

`--links` enforces both directions. A path whose target `init` does not write
resolves here, where the corpus is whole, and goes nowhere in the repository
the material was written into. And a pointer written as an invocation is
refused in the material, because a document pointing at a document points at a
file — `kb.Link.Path` is what tells the two apart. A help screen is not
material: `brief` and `guide` are commands, and naming one there is telling
somebody to run it.

What precedes the kind depends on where the pointing document sits:

| written in | form |
|---|---|
| a document inside one of the directories | `../wiki/x.md` |
| `aliases.md` or `index.md`, at the landed root | `wiki/x.md` |
| a design-time skill beside the corpus | `../asgard-platform/wiki/x.md` |
| deeper in a customer repository | `.agents/skills/asgard-platform/wiki/x.md` |

So `pathLinkRe` fixes only the tail — the kind, a lower-case name, `.md`. The
lower case is what stops a prose mention of `wiki/README.md` becoming a
pointer.

`kb.Landed(prefix, invocation)` converts one form to the other and is the only
place that knows the set of kinds. `needs` and `brief` store invocations and
call it when rendering: the stored value is an identity — *this claim belongs
to that document* — and where the text comes out decides the form.

**`internal/corpus/` exists so that a path resolves in both trees.** It holds
the material in the layout a repository receives it, so `../usecase/x.md` is
correct here and there alike.

## The audits

    asgard-cli audit-material --help     every flag, and what each one answers

**The list is not here**, because a list of flags in a document is a second copy
of `--help` that drifts from it. What follows is why the interesting ones are
shaped the way they are.

`--links` and `--orphans` read the same graph from opposite ends. A dead
pointer is loud: the reader follows it and finds nothing. **A document nothing
points at is silent, and costs more** — it is there, it is correct, and it is
never read. An index does not count as a pointer in `--orphans`, because a
document reachable only from a list is reachable only by somebody who already
suspects it.

**`--bare` is what keeps the graph complete, and the graph is only as good as
it.** A reference written without a path — a same-directory markdown link, or a
name on its own — resolves for a reader today and is checked by nothing, so a
renamed page breaks it silently, and an index is where they accumulate because
an index is nothing but references. The rule that makes the check safe is a
hyphen: `agents` is a CR field and `verify` is a command, while a hyphenated
token matching a document name has no other reading.

`--unverified` reports `needs` and `brief` differently, and that is the point.
`needs` and `brief` render every document from one shared provenance string, so
the marker is there by construction and the check cannot fail on them. Counting
them as having passed would say more than was done.

**`--sources` is the version number of every moment of synthesis, collected.**
A page's Sources block, a pinned table's `const` in `internal/gate` and the
raw-sources table in the wiki README all carry one: which commit of which
upstream a claim was read at.

**It does not require them to agree.** Provenance is per claim, so two commits
of one upstream is the ordinary state of a corpus read over time - a rule
requiring one commit everywhere fails the schema for working, and fails it
hardest on the day somebody re-reads a single section.

So what it prints is the spread, which is the useful thing: a sweep that was
meant to move every citation and moved some of them looks exactly like a corpus
read over several days, and nothing can tell those apart. What it **fails** on
is an upstream cited somewhere and missing from the raw-sources table - a
dependency nobody declared, which is the one thing here that cannot be a matter
of timing.

It does not ask whether a commit is current: **nothing inside this repository
can**, which is why the commit is recorded at all. `go run ./hack sources` reads the
clones and says how far behind each is.

`--commands` is `--links` pointed at the tool: it resolves every
`asgard-cli <command>` against the command tree this binary answers to — in the
material, in the scaffold templates, and **in this repository's own string
literals**, which make the same claim and are the half a customer never sees
until the tool prints one. `selfsrc` embeds the source and `internal/cli/self.go`
parses it, so a comment recording that a command was removed does not read as
naming it.

**A command reference is what is written as code**: backticked, fenced,
indented, or inside a double-quoted span that is nothing but a command — the
last because a Go raw string cannot hold a backtick, and the root help is one.
Two things that form cannot see, so they are checked separately: a bare name
with no `asgard-cli` in front of it, swept for over the closed set of removed
names in `replacements`; and a path that is not a document pointer, which is
`--paths`.

`--paths` exists because these documents land in somebody else's repository,
where "this repo" is theirs and `source/SOURCES.md` is not there. The rule is
not "do not name a path" — provenance should name the file it came from — it is
**name the repository the path is inside**, on the same line.

**The audits check the material and not the capability**, and the difference
has teeth: `asgard-cli init` could come to require a session with every audit
still green. `go run ./hack goal` is the other side - it runs the tool in a
temporary directory with no network, no account and no repository, and holds
Goal.md's four points against what happens.

**The checks that need somebody else's repository are not in the binary**, and
are Go under `hack/` instead: `go run ./hack tables` holds the gate's pinned
tables against the generated CRDs, `processors` and `counts` hold the material's
own tables and figures against what they were distilled from, and
`hack/verify-references.sh` runs the gate over the reference deployments.
`go run ./hack list` says what each one needs. They do not ship because the
repositories they read are not vendored - see AGENTS.md for why that directory
is compiled rather than scripted.

## Retrieval

**There is no search command.** `asgard-cli init` writes the material into the
repository and `grep` is the way in. Two things a grep does not do for itself,
so the landed `SKILL.md` and `index.md` say them:

**`aliases.md` is applied before searching, not after failing.** The corpus is
English and a customer conversation usually is not, so a term taken from what
somebody said matches nothing — which reads identically to a subject the
material lacks. Three tables, and the difference between them is what a row
does to the query:

| table | the row | why |
|---|---|---|
| what a customer says | *replaces* the term | a Chinese term appears nowhere in an English corpus, so keeping it only adds a term that lands nowhere |
| names the material covers | *adds* to the term | SHOPLINE is written verbatim in a page, and that page is the best answer there is |
| names it only routes | *adds*, and says nothing answers it | the material does not name 綠界; what it has is the shape the thing belongs to |

The third table is the one that earns its heading. **A row that routes reads
exactly like a row that answers**, and a reader who cannot tell them apart
takes results about a shape as results about the product they asked for. Moving
a row up a table means somebody did the search and wrote down what came back.

**`glossary.md`'s first table is the word with two senses here.** `payment` is
billing between Asgard and the customer, and also the customer's own payment
gateway. Both sets of results are correct and nothing contradicts anything, so
the wrong one reads exactly like an answer — the failure a search cannot report,
because it found something.

**What makes that table reachable is that it contains the word.** A grep for
`payment` returns `wiki/glossary.md` among its hits, by construction, for every
word with a row — so the warning arrives in the same result set as the
ambiguity rather than depending on the reader having read an instruction first.
That is the whole of the mechanism, and it is the reason the table is a table of
words rather than prose about them.

When the material has no answer, `asgard-cli issue-report --new` writes the
report with what the tool already knows filled in. That is the only way back
in, and it is Goal's fourth point.

No embeddings, no vector index. Synthesis happens once, into a document, rather
than on every query — the [llm-wiki](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
form, and the only reason for it.

## Landing

`asgard-cli init` writes the material into `.agents/skills/asgard-platform/`,
so an agent in a customer repository reaches it with `cat` and `grep`:

    index.md    generated from the jobs actually written
    aliases.md  the alias index
    wiki/ usecase/ needs/ brief/ guide/

`internal/scaffold/corpus.go` builds those as jobs in `scaffold`'s own plan, so
they inherit the existing contract instead of getting rules of their own:
`shipped()` covers the prefix, the stamp records them, `gate`'s `shipped` step
checks them.

`guide` lands as its static half only. A stage renders this repository's own
state — which projects exist, what is still open — and that half stays a
command, because a file would freeze one moment of it. The split is by
paragraph: placeholders are substituted first, then any paragraph still
carrying a template action is dropped.

**Nothing written is interpolated.** Staleness is found by byte comparison, so
a version string in a landed document would make every repository report behind
on a release that touched no page.

### The record

`.asgard-scaffold.json` holds a digest and a CLI version **per file**,
committed beside the material. A byte comparison says a file differs; only the
record says which way round, which is what these states report:

| state | meaning |
|---|---|
| `behind` | this binary is newer, nobody here edited it — `--force` takes it |
| `edited` | somebody here changed it; `--force` would discard that |
| `ahead` | written by a **newer** CLI than the one running; `--force` refuses |
| `stale` | differs, and which way round is not knowable |
| `retired` | this binary no longer ships it |

### Two rules of its own

**A version change replaces the corpus subtree outright.** It is the only
delete the scaffold performs, and `replaceCorpus` requires all three: the stamp
has records under that prefix, some record's version differs, and the path is a
directory rather than a symlink. The whole directory is generated, and a page
renamed upstream would otherwise leave both names on disk where grep returns
the old one.

**A file with a managed region is never rewritten whole.** `AGENTS.md` ships
sections an engagement fills in; markers bound the half this CLI owns and only
that half is replaced. The digest recorded after a merge says this CLI wrote
those bytes — true, and not the same as having written all of them.

## Keeping it current

Three things a corpus needs that no static check answers, and each is a listing
rather than a check because each reports a question.

**What a change owes a re-read.** `go run ./hack related` reads the in-edges of
every document a change touched, off the same `kb.Link` graph `--links` walks
from the other end. `--orphans` asks whether anything points at a document; this
asks what does. A change to one page names about a dozen documents, which is a
reading somebody does, where "read the corpus again" is a reading they abandon.

**Which of those were ever read against it.** `go run ./hack reconcile` records
the digest of a target at the moment somebody says they read a pointer against
it, in `source/reconciled.json`, and reports the pointers whose target has moved
since. **A document points at itself**, and that edge is its own
`description:` - an index row is authored rather than derived, so it does not
follow the page when the page changes, and a row that no longer says what is
inside a document reads exactly like one that does.

**The record is committed, and the one in `.out/` is not**, which is the whole
distinction: a reading is a claim about the material and the next person
inherits it, the way they inherit a `**Checked:**` line; a recorded pass is a
claim about one machine's tree. There is deliberately no way to record every
pointer at once - that would write a claim nobody made, which is the
verdict-in-a-table this repository refuses everywhere else.

**What has already been answered.** `go run ./hack verified` keys each check on
the digests of the classes of input it reads - the corpus, the Go source, the
root documents, each upstream clone at its commit - and skips the ones whose
inputs have not moved. The classes are coarse deliberately: naming the exact
files a check reads is a second copy of what the check does, kept by hand, and
it goes stale in the direction that matters. **A skip that lies is worse than
re-running everything**, so a check that declares no class always runs, and a
skip is printed differently from a pass.

## Verification

`asgard-cli gate` runs, in order: `tools`, `repo`, `shipped`, `binding`,
`skills`, `lint`, `render`, `verify`. Two steps need the platform, and **a skip
is printed differently from a pass**.

`verify` reads rendered CRs against each other and against tables pinned from
the CRDs — enums in `internal/gate/enums.go`, field constraints in
`constraints.go`. Both are warnings, and **not because a pinned copy can only go
stale in the safe direction**: it cannot. The platform deleted a cron pattern
its own regex had copied wrong, and the pinned row then reported four schedules
an FDE would obviously want as violations while admitting one the apiserver
refuses. `constraints.go` carries that reasoning; `go run ./hack tables` is what
holds both against the CRDs.

**Those tables are keyed by kind and path, not by field name.** A json field
name is not a location: `Loader.spec.schedule` is an unconstrained string while
`Trigger.spec.cron.schedule` carries a pattern, and a table keyed on the name
alone holds one against the other's rule.

**It does not reproduce the platform's checks.** Whether a CR is admitted is an
apiserver's decision and no client is issued cluster credentials, so a copy of
those rules here would drift, and would still miss the two that matter: a field
the CRD silently prunes, and a rejection only the apiserver produces. **41 of
the CRDs' 231 enforced CEL rules are `self == oldSelf`**, comparing a proposal
against the object already on the cluster — a render is one object with no
history. (A marker in the Go types and a rule on the cluster are different
questions with different answers; `internal/corpus/wiki/crd-rules.md` has both,
and `go run ./hack tables` recomputes them.)

A green gate means *worth pushing*. The authority is the plan:

    asgard-cli pipeline runs watch --release <name> --ref <tag>

## Chart authoring

`asgard-cli add <kind> <name>` writes a CR skeleton into a project's chart.
`generate.Kinds` is the set, each naming the wiki page and the extract that
explain it — those names are pointers like any other and `--links` resolves
them, which is the half a prose search cannot reach.

What the generator writes is the conventions applied: naming, the display
annotations, what belongs in `values.yaml` and what stays in the template.

Out of scope deliberately, and Goal.md says why: the namespace and the
environment id, which the platform injects as `.Values.asgard.*` on every run,
and whether the thing deploys at all.
