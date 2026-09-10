# How it is built

**How the main capabilities are implemented.** What the tool is for is
[Goal.md](Goal.md); where it stands is [TASK.md](TASK.md); what the commands do
is [README.md](README.md); what lives in which directory is
[STRUCTURE.md](STRUCTURE.md); the rules for changing it are
[AGENTS.md](AGENTS.md).

Present tense. Why something changed is `git log`.

## The corpus

Six bodies of material, one reader. `kb.Corpus` is that reader; four packages
declare one, and `needs` and `brief` render documents from Go structs into the
same shapes.

    kb.Doc     Name, Title, Summary, Checked, Unchecked, Links, Sources
    kb.Link    Kind, Name, Path, Deliberate
    kb.Corpus  FS, Dir, Unlisted, Noun, Command

| body | package | kind |
|---|---|---|
| wiki pages | `internal/wiki` over `internal/corpus/wiki/` | `wiki` |
| deployment extracts | `internal/usecase` over `internal/corpus/usecase/` | `usecase` |
| what to obtain from a customer | `internal/needs` | `needs` |
| what an activity gets wrong | `internal/brief` | `brief` |
| stage guidance | `internal/stage` over `prompts/` | `guide` |
| design-time skills | `internal/scaffold` over `templates/.agents/skills/` | — |

**Add material to one of these, not beside them.** A body with its own reader,
its own search and its own match type drifts from the others and nothing
mechanical notices. If new material does not fit `kb.Corpus`, change `kb`.

Three questions, three answers, not interchangeable:

    List()      what a reader listing documents should see
    All()       plus the corpus's own bookkeeping - an index, a log, a README
    Landing()   what `asgard-cli init` writes into a repository

`Unlisted` is what separates them. The wiki's index is bookkeeping that has to
travel; its log is bookkeeping that must not. The audit's source set is what
lands: a document that ships is a document that is checked.

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

    audit-material --links      every pointer resolves, and a path's target lands
    audit-material --orphans    what nothing points at
    audit-material --bare       a document named with no way to reach it
    audit-material --commands   every command named exists
    audit-material --paths      a landed document naming a file only we have
    audit-material --urls       every documentation link is live
    audit-material <term>       every line mentioning a term, prose and templates

`--links` and `--orphans` read the same graph from opposite ends. A dead
pointer is loud: the reader follows it and finds nothing. **A document nothing
points at is silent, and costs more** — it is there, it is correct, and it is
never read. An index does not count as a pointer in `--orphans`, because a
document reachable only from a list is reachable only by somebody who already
suspects it.

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

`hack/check-tables.py` holds the gate's pinned tables against the generated
CRDs; `hack/verify-references.sh` runs the gate over the reference deployments.
Neither ships in the binary — both need repositories that are not vendored.

## Retrieval

**There is no search command.** `asgard-cli init` writes the material into the
repository and `grep` is the way in. Two things a grep does not do for itself,
so the landed `SKILL.md` and `index.md` say them:

**`aliases.md` is applied before searching, not after failing.** The corpus is
English and a customer conversation usually is not, so a term taken from what
somebody said matches nothing — which reads identically to a subject the
material lacks. It holds two tables: words that *replace* a query term, because
a Chinese term appears nowhere in an English corpus, and names that are *added*
to it, because a product name may be written verbatim in a page.

**`glossary.md`'s first table is the word with two senses here.** `payment` is
billing between Asgard and the customer, and also the customer's own payment
gateway. Both sets of results are correct and nothing contradicts anything, so
the wrong one reads exactly like an answer — the failure a search cannot report,
because it found something.

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
record says which way round, which is what the five states report:

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

## Verification

`asgard-cli gate` runs, in order: `tools`, `repo`, `shipped`, `binding`,
`skills`, `lint`, `render`, `verify`. Two steps need the platform, and **a skip
is printed differently from a pass**.

`verify` reads rendered CRs against each other and against tables pinned from
the CRDs — enums in `internal/gate/enums.go`, field constraints in
`constraints.go`. Both can only go stale in the direction of the platform
adding something, so both are warnings.

**Those tables are keyed by kind and path, not by field name.** A json field
name is not a location: `Loader.spec.schedule` is an unconstrained string while
`Trigger.spec.cron.schedule` carries a pattern, and a table keyed on the name
alone holds one against the other's rule.

**It does not reproduce the platform's checks.** Whether a CR is admitted is an
apiserver's decision and no client is issued cluster credentials, so a copy of
those rules here would drift, and would still miss the two that matter: a field
the CRD silently prunes, and a rejection only the apiserver produces. Forty of
the seventy-nine CEL rules are `self == oldSelf`, comparing a proposal against
the object already on the cluster — a render is one object with no history.

A green gate means *worth pushing*. The authority is the plan:

    asgard-cli pipeline runs watch --release <name> --ref <tag>

## Chart authoring

`asgard-cli add <kind> <name>` writes a CR skeleton into a project's chart. Ten
kinds in `generate.Kinds`, each naming the wiki page and the extract that
explain it — those names are pointers like any other and `--links` resolves
them, which is the half a prose search cannot reach.

What the generator writes is the conventions applied: naming, the display
annotations, what belongs in `values.yaml` and what stays in the template.

Out of scope deliberately, and Goal.md says why: the namespace and the
environment id, which the platform injects as `.Values.asgard.*` on every run,
and whether the thing deploys at all.
