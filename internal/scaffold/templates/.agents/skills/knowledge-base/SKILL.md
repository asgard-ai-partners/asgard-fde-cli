---
name: knowledge-base
description: How to maintain this repo as a knowledge base rather than a pile of files - what the layers are for, filing an answer you worked out so the next reader gets it for free, and auditing for the rot no linter can see (pages that contradict each other, claims a newer source superseded, concepts nobody owns). Use when picking up a repo somebody else worked on, before a handover, after a burst of decisions, when you answered a question by reading four files, or when an answer you gave from the docs turned out to be wrong.
---

# This repo is a knowledge base

Most of what is in here is not code. It is what an engagement learned about one
customer's systems, and the repository is the only place it exists - the next
agent to open it starts from these files and nothing else.

That changes what "maintaining the repo" means. A pile of accurate documents
that nobody can navigate is worth about as much as no documents, and the way it
gets to that state is never a single bad commit. It is a decision that was
recorded and never applied, an answer somebody worked out twice, a term five
files define slightly differently.

**The layers and how facts move between them are in
[`docs/README.md`](../../../docs/README.md); the loop that lands a decision is
in the `spec-workflow` skill. Do not restate either here.** This skill is the
part neither of them covers: the habits that keep the thing readable, and the
audit that finds where it stopped being readable.

## Who does what

The tedious half of a knowledge base is not the reading or the thinking. It is
the bookkeeping: updating the index, fixing the cross-reference, noticing that
three other pages now say something slightly wrong.

That half is yours. You do not get bored, you do not forget a cross-reference,
and you can touch fifteen files in one pass. The person you are working with
curates, directs, and questions - which page matters, whether a claim is true,
what the customer actually meant. **When you land a change, land the bookkeeping
with it in the same pass**, rather than reporting that an index needs updating.

The exception is judgement that belongs to a person: which of two contradicting
pages is right, whether a stale claim is worth re-verifying, whether a decision
should be reversed. Bring those back rather than resolving them quietly.

## Filing what you worked out

**When you answer a question by reading several pages, the answer is new
knowledge and it is about to be thrown away.**

This is the most common way effort is wasted here, and it does not feel like
waste at the time - the question got answered, so the work looks finished. But
the next person asks the same question and reads the same four files, because
the synthesis lived in a chat and the files are unchanged.

So after answering a question that took real assembly, ask one thing: *would the
next reader have to redo this?* If yes, file it, in whichever layer owns it:

  - it describes **how the system behaves** -> a living spec module, or a
    paragraph added to one
  - it explains **why something is the way it is** and you had to reconstruct
    that from several sources -> a decision record only if a decision was
    actually taken; otherwise it belongs in the spec module as context
  - it is **how to work in this repo** -> `AGENTS.md`
  - the **running agent** needs it -> `common/skills/<skill>/SKILL.md`, because
    nothing else reaches the cluster
  - you could not settle it -> `docs/open-questions.md`, with what it blocks and
    who can answer

Not every answer earns a page. A one-line lookup does not; a rule you had to
infer by comparing three files does. The test is the assembly, not the length.

## The audit

`asgard-cli check` finds the structural rot: a page with no inbound link, a
module missing from its index, a filename that is not dated, a link that does
not resolve. **Run it first** - there is no point reading for contradictions in
a repo whose indexes are already wrong.

    asgard-cli check

What is left needs somebody to read the pages against each other. Three targets.

### 1. Contradictions

Two pages that cannot both be true. Read for these in the order that finds them
cheapest:

1. **A decision record against the living spec module it changed.** A decision
   is applied by rewriting the module; if the module still describes the old
   behaviour, the decision was recorded and never landed. The most common one,
   because writing the decision feels like finishing.
2. **A `done` task spec against the living spec.** A task that changed behaviour
   is not done until its delta reached the spec, but the status was moved by
   hand and the delta was not.
3. **Two decision records on the same topic.** Both are immutable and both stay,
   so this is not a defect by itself - the defect is when neither says which one
   won. Fix it by writing a third that supersedes, not by editing either.
4. **`AGENTS.md` against the living spec.** These answer different questions -
   how to change the repo, versus what the system does - so overlap between them
   is a sign a fact has two homes, and two homes drift.
5. **A `common/skills/` skill against the living spec.** The most expensive kind,
   because the skill is what the *running agent* believes. A stale skill is a
   production behaviour, not a documentation problem.

For each one, report **which side you believe and why**. "These disagree" sends
the reader back to do the work again; "the decision is dated later and the spec
module was last touched before it, so the spec is stale" does not.

### 2. Superseded claims

A statement that was true when written and has been overtaken. Unlike a
contradiction, nothing in the repo disagrees with it - the world moved and the
page did not.

  - **A claim about the platform that a newer CRD or deployment contradicts.**
    Platform vocabulary moves: a field is retired, a default flips, a resource
    is deprecated. A page naming a field that no longer exists reads exactly
    like one naming a field that does.
  - **A claim about the customer's systems with no provenance**, or with
    provenance old enough to doubt. `requirements/README.md` requires every
    claim about a real system to say how it was established and when; anything
    undated is a finding on its own.
  - **An open question that has quietly been answered** by a later decision and
    never moved to Answered in `docs/open-questions.md`.
  - **A `references/` document presented as current** that the customer supplied
    long enough ago that nobody should rely on it unchecked.

Verify before reporting where verifying is cheap: a schema claim against the
database, a CR field against the platform's CRD. Where it is not cheap, say what
would settle it and leave it as an open question with an owner - a guess
recorded as a finding is the same defect one layer up.

### 3. Concepts nobody owns

The hardest to see, because nothing is wrong on any single page. A term appears
across five documents, everybody explains it in passing, and no page owns it.
The explanations drift, and a new reader assembles a definition from whichever
ones they happened to read.

Find them by noticing repetition: a term explained more than twice in different
files, or explained differently in two of them, needs a home. **One fact, one
home, everywhere else links** - the same rule `requirements/README.md` states.

When you propose a page, say which sentences in which files it replaces. A new
page that deletes nothing has made the problem worse.

## Reporting

Write findings; do not fix contradictions silently. Which side is right is
frequently a question for a person - the FDE who was in the room, or the
customer.

Order by the cost of being wrong, which is not the same as how obvious it is:

    1. anything under common/skills/     the running agent believes it today
    2. the living spec                   the next person will believe it
    3. decisions and requirements        found only when somebody goes looking

For each: the pages, which one you believe, the evidence, and the smallest fix.
Land the settled ones through the normal loop rather than by editing pages
directly - a new dated decision record, a delta applied to the living spec
module, a row moved to Answered. **Editing a decision record to make it agree
with reality destroys the one thing that layer exists for.**

End by re-running `asgard-cli check`. Applying a delta commonly leaves a new
page nothing links to yet, and that is exactly the half you should not have to
find by reading.

**Checked:** 2026-09-04 - **no platform claims to check.** This page is about
keeping a body of written material healthy, and the three kinds of rot it names
are properties of documents rather than of Asgard. Its own subject is checkable
by running the audits it describes.

**Unchecked:** that those three are the kinds that matter. They come from
maintaining this tool's corpus, where a fourth turned out to exist and is
recorded on the wiki's own README - upstream moved and the corpus did not,
which nothing inside a body of material can detect. A reader maintaining a
different corpus should expect their own fourth.
