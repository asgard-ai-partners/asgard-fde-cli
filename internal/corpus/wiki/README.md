# Asgard platform wiki

    ../wiki/            every page, as files
    ../wiki/<page>.md   read one
    index.md            them grouped by the question each answers
    ../aliases.md       what a customer said -> what to search for

**Read `../aliases.md` first if the question did not arrive in English**, then
grep. The material is English and a customer conversation usually is not, so a
term taken from what somebody actually said matches nothing - which reads
exactly like a subject this wiki does not cover.

What the Asgard platform is made of, and who each piece is for.

**The reader is an agent.** These pages are compiled into the binary so that an
agent working in a customer repository has the platform knowledge - all it
otherwise has is that repository, which describes one customer's systems and
never the platform they run on. An FDE reads the same thing through the same
command.

So write them the way an agent can use: name mappings, limits, the basis for a
choice, what has been retired. UI mechanics only where they change a decision or
tell somebody which button to press.

How this differs from the other two bodies of material:

| material | answers |
|---|---|
| this wiki | what the platform has, who it is for, which CR a UI name maps to |
| `../usecase/` | how one deployment shape is assembled, field by field |
| [asgard-kube's `crd/*.yaml`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) | whether a field is legal, and whether it is required |

Both of those assume the reader already knows the platform has the thing. This
wiki is that assumption.

**Do not repeat them.** Chart-writing cautions belong to `../usecase/` and
field rules to the CRD; point at them instead.

> **These pages are the binary's copy, and `asgard-cli init` also writes them
> into a repository** under `.agents/skills/asgard-platform/`, with the
> extracts beside them. Edit them here: that copy is generated, and the next
> `init` replaces it when the binary's version moves.
>
> The copy is there because an agent in a customer repository finds what it
> knows under `.agents/skills/`, and `grep -r` answers "which document says
> this" with no subprocess. What a copy cannot do is translate a query, so
> `aliases.md` is written beside it, and it does what a grep cannot: it
> translates the customer's words into the material's.
>
> **A copy is safe because staleness is visible.** `.asgard-scaffold.json`
> records a digest and a CLI version per file, so a repository that is behind
> is distinguishable from one somebody edited and from one written by a newer
> binary than the one now reading it - and `--force` refuses that last case
> rather than downgrading it.

## Three layers

The llm-wiki split. What separates the layers is which one may be rewritten.

**These are the corpus's layers, not this wiki's alone.** The extracts, the
stage guidance and the design-time skills are the same middle layer read for a
different question, and `AGENTS.md` states the four rules that apply to all of
them. This file is the longer version, and the part below about a page's source
block is the wiki's own.

| layer | contents | may be edited |
|---|---|---|
| raw sources | asgard-docs and asgard-kube (URLs below) | read-only. Never copied in; only the commit is recorded |
| the corpus | these pages, and the other four parts | rewritten continuously, and only ever describes the present |
| the schema | `AGENTS.md`, and this file for what a page must carry | changed deliberately, by a person |

Raw sources are not vendored, and the reason is not size. A copy stops tracking
upstream, and a stale copy is indistinguishable from a current one by looking at
it. Recording the repository and the commit at least makes the comparison
possible.

### The raw sources

**The source of truth is the URL, not a path on somebody's machine.** Clone them
wherever; what goes into a page is the repository and the commit.

| source | source of truth | read at |
|---|---|---|
| product documentation | https://github.com/asgard-ai-platform/asgard-docs | `f00e0ee` (2026-08-31) |
| CRD definitions | https://github.com/asgard-ai-platform/asgard-kube | `cbd8d70` (2026-09-11) |
| processor definitions | https://github.com/asgard-ai-platform/asgard-core (private) | `62a696b7` (2026-09-11). `internal/constants.go` last changed at `80b70e7c`, 2026-09-01 - before the 2026-09-03 reading - so what was read then is what is there now |

None of them lives in this repository. `git pull` before writing against them.

**Moving a commit here is a claim that somebody re-read.** For `cbd8d70` what
was read is the whole `15ded0f..cbd8d70` diff rather than every page again:
four commits, and exactly one change that reaches a page. The cron `schedule`
pattern was deleted from `Trigger.spec.cron` and from the SourceSet syncer,
because the regex "had copied [cron] wrong in both directions". Nothing else
in the diff touches an enum, a constraint or a field this material describes -
the rest is CI, a lint, and documenting that `pre-tool-call` and
`post-tool-call` do nothing, which `../usecase/per-turn-credentials.md` had
already found from a deployment. **Reading the diff is a real reading when the
diff is small; it is not one when the diff is large, and saying which was done
is the point of this paragraph.**

## Three operations

### ingest - a new source arrives

1. Read it, and confirm the reading with the person before writing
2. Write or rewrite the matching page under `wiki/`
3. Update `index.md`
4. **Update the other pages it touches**

Step 4 is the one that gets skipped.

**The commit you read a source at goes in the page's own source block**, which
is where a reader checking a claim already is. There is no separate ledger:
one would be a second copy of what every page already carries, and it would go
stale in the direction that matters, because nothing points at it.

### query - answering a question

Search `wiki/` first. If an answer needs three pages assembled on the spot, that
assembly is new knowledge: write it into a page, or the next reader repeats it.

### lint - the audit

Three of these are mechanical and ship as flags:

    asgard-cli audit-material --links     every pointer resolves
    asgard-cli audit-material --commands  every command named exists
    asgard-cli audit-material --orphans   what nothing points at

**--commands is --links pointed at the tool.** A document that tells somebody to
run something resolves the same way a pointer does, against the command tree
this binary answers to, and six documents once named a command nobody had
built. It reads the scaffold templates too - a customer meets these names in a
generated README before meeting any of this.

**--links and --orphans are two halves of one thing.** A pointer that goes nowhere is loud - the
reader follows it and finds nothing. A document nothing points at is silent, and
costs more: it is there, it is correct, and it is never read. The index does not
count as a pointer in the second check, because `../wiki/operations.md` sat in
`index.md` under the title Connectivity while an FDE spent a day on connectivity
and never opened it.

Both read the link graph, which is a field on every document rather than a
regular expression over prose - see `kb.Link`.

The three kinds of rot in `.agents/skills/knowledge-base/` apply here too. There
is a fourth that only happens here:

**Upstream moved and the wiki did not.** asgard-docs and asgard-kube both change.
A page can be internally consistent, unorphaned and free of contradictions while
describing a platform from three months ago. Nothing inside the wiki can detect
that; only going back to the source can.

That is why every page carries its sources.

## Rules for a page

**Every page ends with a source block**, in a fixed shape:

```markdown
## Sources

- [Flow Agent](https://docs.asgard-ai.com/docs/product-suite/odin/features/agent-hub-flow-agent)
  - asgard-docs `f00e0ee`
- Checked 2026-09-02, re-read 2026-09-11 against asgard-kube `cbd8d70`
```

Link the rendered page on docs.asgard-ai.com rather than the file path: a URL a
reader can open beats a path only somebody with the clone can.

**Every page also ends with an `**Unchecked:**` line**, saying which parts of it
were never held against a real deployment. The wiki is checked less deeply than
`../usecase/` by nature - an extract has a chart to compare against, while the wiki's
source is product documentation describing a UI, much of which is in no chart at
all (permissions, billing, the chat interface). Left unsaid, a reader assumes the
two are equally reliable.

Those markers are the same ones `../usecase/` uses. The two bodies are read
together and a reader should not have to learn where the provenance is written
twice.

**A page with a corresponding chart shape must point at its extract**; a page
without one must say so rather than leaving a blank. `../wiki/console.md`, `../wiki/fehu.md`,
`../wiki/operations.md` and `../wiki/product-suite.md` say so, because billing, permissions and scoping
produce no CRs at all - that is a decision, not an omission.

Extracts point at a wiki page without exception: an extract assumes the reader
knows the platform has that shape, and this is where the assumption comes from.

Four more:

- **One fact, one home; everywhere else links.** Turn a paragraph you were about
  to copy into a link.
- **One word, one meaning.** `../wiki/glossary.md` lists the terms that already
  mean something specific here. Check it before introducing a word, and before
  using one of those for something else - a word with two senses in one body
  produces the failure nobody can see, because nothing contradicts anything and
  the reader simply takes the wrong one.
- **Anything telling a reader to ask a customer something has to pass filter 0.**
  `../guide/requirements.md` carries it: does the answer change what
  we build? Two pages have told an FDE to ask a question that filter rejects, and
  both times they followed the page in front of them rather than the rule.
- **Say which layer a statement comes from.** Product documentation describes
  objects in an interface, the CRD describes resources, and the vocabulary is not
  one to one. Where they diverge, say so on the page - `agents.md` has the
  table.
- **Mark what is uncertain.** "The documentation does not say" is a useful entry;
  a guess is not.
- **English.** The sources are zh-TW and the pages are not. The corpus carries
  one language because **the mapping lives somewhere else** - the alias table on
  `glossary.md`, which is applied to a query before
  searching, printing what it actually searched for. It is not because the
  reader translates first: a query is given in the customer's own
  words, and translating only after a search came back empty was measurably
  worse - the glossary carries the table, so a Chinese term matched the row
  about itself and the reader got the word list instead of the answer. Product
  labels keep their own names - Managed Agent, Drive, Context Index are what the
  UI says.

## The index

Two files here are not pages and are not searched:

| file | holds |
|---|---|
| `index.md` | the catalogue of pages, by subject, and what is not written yet |
| `aliases.md` | what a customer says, and what to search for |

**An index inside the corpus competes with what it points at.** The alias table
was a section of `glossary.md`, and because it lists every alias it was
reliably the one document carrying every term of a translated query: a search
for a subject returned the word list rather than the page. It sits one level up
now, beside `index.md`.

`aliases.md` is applied to a query before searching, and it carries the rule
for adding a row: a search of yours came back empty and the subject turned out
to exist under another name.

**A row that routes is not a row that answers**, and the index says which it is.
A name under "names the material covers" was searched for against the reference
deployments and the answer written down; one under "names it only routes"
reaches the shape the thing belongs to and nothing more. The table prints the
difference, because a set of results about a shape reads exactly like a set of
results about the product that was asked for. **Moving a row up means somebody
did the search.**

The `../wiki/glossary.md` page's one-meaning-here table is applied the same way. It sat as
prose for a long time while the failure it describes went on happening: a search
for `payment` returns Fehu, which is billing between Asgard and the customer,
and nothing anywhere is red. **A grep for a word in that table has to be read
against both senses**, which is why the table is at the top of the page rather
than in the body.

## Coverage

`index.md` has a second table listing what has not been written. It is maintained
alongside the pages that have.
