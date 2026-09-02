# Asgard platform wiki

    asgard-cli find <terms>         look something up - searches here AND the
                                    extracts, and names the counterpart it finds
    asgard-cli wiki                 list every page
    asgard-cli wiki <page>          read one
    asgard-cli wiki --unverified    what each page has NOT been held against
    asgard-cli wiki --conventions   this file

`find` is the way in. `wiki --search` exists for when you already know the answer
is on the platform side rather than in a deployment shape.

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
| `asgard-cli usecase` | how one deployment shape is assembled, field by field |
| [asgard-kube's `crd/*.yaml`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) | whether a field is legal, and whether it is required |

Both of those assume the reader already knows the platform has the thing. This
wiki is that assumption.

**Do not repeat them.** Chart-writing cautions belong to `usecase` and field
rules to the CRD; point at them instead.

> These pages ship inside the binary and are never written into a customer
> repository - the same reason `internal/usecase/extracts/` is not. A copy inside
> one engagement goes stale where nobody is looking, while a stale page here is
> fixed for every engagement in one release.

## Three layers

The llm-wiki split. What separates the layers is which one may be rewritten.

| layer | contents | may be edited |
|---|---|---|
| raw sources | asgard-docs and asgard-kube (URLs below) | read-only. Never copied in; only the commit is recorded |
| the wiki | `pages/` | rewritten continuously, and only ever describes the present |
| the schema | this file | changed deliberately, by a person |

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
| CRD definitions | https://github.com/asgard-ai-platform/asgard-kube | `15ded0f` |

Neither lives in this repository. `git pull` before writing against them.

## Three operations

### ingest - a new source arrives

1. Read it, and confirm the reading with the person before writing
2. Write or rewrite the matching page under `pages/`
3. Update `index.md`
4. **Update the other pages it touches**
5. Append a line to `log.md`

Step 4 is the one that gets skipped.

### query - answering a question

Search `pages/` first. If an answer needs three pages assembled on the spot, that
assembly is new knowledge: write it into a page, or the next reader repeats it.

### lint - the audit

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
- Checked 2026-09-02 against asgard-kube `15ded0f`
```

Link the rendered page on docs.asgard-ai.com rather than the file path: a URL a
reader can open beats a path only somebody with the clone can.

**Every page also ends with an `**Unchecked:**` line**, saying which parts of it
were never held against a real deployment. The wiki is checked less deeply than
`usecase` by nature - an extract has a chart to compare against, while the wiki's
source is product documentation describing a UI, much of which is in no chart at
all (permissions, billing, the chat interface). Left unsaid, a reader assumes the
two are equally reliable.

Those markers are the same ones `internal/usecase` uses. The two bodies are read
together and a reader should not have to learn where the provenance is written
twice.

**A page with a corresponding chart shape must point at its extract**; a page
without one must say so rather than leaving a blank. `console`, `fehu`,
`operations` and `product-suite` say so, because billing, permissions and scoping
produce no CRs at all - that is a decision, not an omission.

Extracts point at a wiki page without exception: an extract assumes the reader
knows the platform has that shape, and this is where the assumption comes from.

Four more:

- **One fact, one home; everywhere else links.** Turn a paragraph you were about
  to copy into a link.
- **One word, one meaning.** `pages/glossary.md` lists the terms that already
  mean something specific here. Check it before introducing a word, and before
  using one of those for something else - a word with two senses in one body
  produces the failure nobody can see, because nothing contradicts anything and
  the reader simply takes the wrong one.
- **Anything telling a reader to ask a customer something has to pass filter 0.**
  `asgard-cli next --stage requirements` carries it: does the answer change what
  we build? Two pages have told an FDE to ask a question that filter rejects, and
  both times they followed the page in front of them rather than the rule.
- **Say which layer a statement comes from.** Product documentation describes
  objects in an interface, the CRD describes resources, and the vocabulary is not
  one to one. Where they diverge, say so on the page - `pages/agents.md` has the
  table.
- **Mark what is uncertain.** "The documentation does not say" is a useful entry;
  a guess is not.
- **English.** The sources are zh-TW and the pages are not; an agent asked in
  Chinese queries in English, so the corpus does not have to carry both. Product
  labels keep their own names - Managed Agent, Drive, Context Index are what the
  UI says.

## Coverage

`index.md` has a second table listing what has not been written. It is maintained
alongside the pages that have.
