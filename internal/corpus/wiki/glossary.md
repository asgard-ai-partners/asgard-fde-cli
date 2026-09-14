# Words with one meaning here

A word that means two things inside one body of material produces the failure
that is hardest to see: nothing contradicts anything, and the reader takes the
wrong sense. **"sandbox" is the pair to know** - the platform's agent runtime,
and the customer's test environment - and a paragraph using both without
saying which is where an FDE reads the second as the first.

So each of these has one meaning here, and the other senses have their own words.
**Check this page before introducing a term, and before using one of these for
something else.**

**And grep it for the word you searched.** The table contains the word, so a
search for an ambiguous term returns this page in the same result set as the
ambiguity - which is why the senses are a table rather than prose about them.

| word | means, here | not |
|---|---|---|
| **sandbox** | the isolated runtime the platform starts to run an agent in | the customer's test environment - call that a **test environment** |
| **project** | one Helm chart deployed to one namespace, under `projects/<slug>/` | the platform's own Project object, which is a division inside a Workspace - say **platform Project** |
| **environment** | `dev` or `prod`: one release in `.asgard-pipeline.yaml`, bound to its own platform Project | the platform's Environment inside a platform Project - say **platform Environment**. There are no per-environment values files; a release's values are variables set on the platform |
| **agent** | a `Agent` CR, which the UI calls a Managed Agent | the flow-agent shape, which contains no `Agent` CR at all; and not the coding agent working in a repository |
| **workspace** | the customer, and the repository root | nothing else. It is also the platform's billing unit, which is the same thing seen from Fehu |
| **template** | a Go template this tool renders | the "Template" node type the product overview describes, which maps to nothing else here - see `../wiki/what-they-read.md` |
| **skill** | `assets/skills/<name>/SKILL.md`, read by the deployed agent at runtime | `.agents/skills/`, which the coding agent reads while authoring. Say **design-time skill** for the second |
| **knowledge-base** | the extract `../usecase/knowledge-base.md`, about `KnowledgeBase` / `Loader` / `Source` - the older knowledge path | the design-time skill of the same name, which is about keeping a repository's own documents honest. A grep for the word returns both, and they have nothing to do with each other |
| **request** | a record under `requirements/requests/` | an HTTP request, and not one run of an agent. For the platform's per-request limits, say **per run** |
| **source** | a `Source` CR under a KnowledgeBase | source code, and not a documentation source. For the material's provenance say **source block**; for code say **source code** |
| **check** | `asgard-cli check`, the structural gate | a layout or content checker in a typesetting skill - say which one |
| **payment** | billing between Asgard and this customer - see `../wiki/fehu.md` | **the customer's own payment gateway**, which is an external system with side effects: `../usecase/write-path.md` and the ladder on `../wiki/taiwan-channels.md`. Say **payment gateway** for theirs |

## Two more that collide with the customer's vocabulary

Not this material's fault, and worth knowing before a meeting.

**"Project"** is the word a customer uses for the engagement. When they say
"how many projects", they mean pieces of work; when this tool says it, it means
charts and namespaces. Ask which they mean rather than answering.

**"Agent"** is the word they will use for the whole thing they are buying. "How
many agents do we need" is a question about capability, and the honest answer
starts by saying the count of `Agent` CRs is not the same number - see
`asgard-cli size`.

## What a customer says

That table is not on this page any more. It is an **index**, not knowledge about
the platform, and while it lived here it competed with the pages it points at.
A table that lists every alias carries every word any translated query is
rewritten into, so it was reliably the one document matching a whole query, and
a reader asking about a subject was handed the word list instead of the page.

It is now beside the pages rather than among them, which is where `index.md`
already lives:

    aliases.md

Apply it to a query before searching, so a question asked in
the customer's own words reaches material written in English. **Add a row when a
search of yours came back empty and the subject turned out to exist under
another name** - that is the only test, and the rules are on that page.

## How a word gets onto this page

When it has been read the wrong way once. Not when it is ambiguous in principle:
this page is only useful while it is short enough to read, and every entry here
was a real misreading.

## Corresponding extracts

None. This is about the material rather than about a deployment.

## Sources

- Each row is a term already defined somewhere in this material; this page adds
  no meanings and only collects the ones that collide. `sandbox` was raised by an
  engagement in 2026-09; the rest were found while writing the row above it

**Unchecked:** nothing enforces this, and a static check cannot: telling which
sense a bare word is in needs a reader. **What was done instead is the reading**,
2026-09-14, across every help screen and every page - and some of these were
live in the tool's own output, each at the moment both senses are in a reader's
hands. The platform commands called a platform Project a "project" while
`projects/<slug>/` means a chart, at `pipeline release create --project`, which
is exactly where somebody is holding both. `pipeline projects` called a main
platform Environment an "environment" beside releases named `dev` and `prod`.
And `asgard-cli skill` fetches design-time skills while the bare word here means
the runtime ones a Syncer feeds. Each of those says which now.

**A new page can still redefine any of these and no command will notice**, so
this stays open - but the failures so far were not new pages. They were the
tool's own help, written by somebody who knew which sense they meant.
