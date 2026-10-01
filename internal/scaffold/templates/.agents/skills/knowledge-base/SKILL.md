---
name: knowledge-base
description: How to maintain this repo as a knowledge base rather than a pile of files - what the layers are for, filing an answer you worked out so the next reader gets it for free, and auditing for the rot no linter can see (pages that contradict each other, claims a newer source superseded, concepts nobody owns). Use when picking up a repo somebody else worked on, before a handover, after a burst of decisions, when you answered a question by reading a handful of files, or when an answer you gave from the docs turned out to be wrong.
---

# This repo is a knowledge base

Most of what is in here is not code. It is what an engagement learned about one
customer's systems, and the repository is the only place it exists - the next
agent to open it starts from these files and nothing else.

So maintaining the repo includes keeping it navigable. Accurate documents that
nobody can navigate are of little use. A repo usually gets there gradually: a
decision that was recorded and never applied, an answer somebody worked out
twice, a term five files define slightly differently.

The layers and how facts move between them are in
[`docs/README.md`](../../../docs/README.md); the loop that lands a decision is
in the `spec-workflow` skill. Do not restate either here. This skill covers
what neither of them does: the habits that keep the repo readable, and the
audit that finds where it stopped being readable.

## Who does what

The bookkeeping is yours: updating the index, fixing the cross-reference,
noticing that three other pages now say something slightly wrong. You can touch
many files in one pass. The person you are working with curates, directs, and
questions - which page matters, whether a claim is true, what the customer
actually meant. When you land a change, land the bookkeeping with it in the same
pass, instead of reporting that an index needs updating.

The exception is judgement that belongs to a person: which of two contradicting
pages is right, whether a stale claim is worth re-verifying, whether a decision
should be reversed. Bring those back rather than resolving them quietly.

## Filing what you worked out

When you answer a question by reading several pages, the answer is new
knowledge. If it stays in the chat, the next person asks the same question and
reads the same files again. This is the most common way effort is wasted here.

So after answering a question that took real assembly, ask: would the next
reader have to redo this? If yes, file it, in whichever layer owns it:

  - it describes how the system behaves -> a living spec module, or a
    paragraph added to one
  - it explains why something is the way it is and you had to reconstruct
    that from several sources -> a decision record only if a decision was
    actually taken; otherwise it belongs in the spec module as context
  - it is how to work in this repo -> `AGENTS.md`
  - the running agent needs it -> `assets/skills/<skill>/SKILL.md`, because
    nothing else reaches the cluster
  - you could not settle it -> `docs/open-questions.md`, with what it blocks and
    who can answer

Not every answer earns a page. A one-line lookup does not; a rule you had to
infer by comparing three files does. Decide by how much assembly it took, not by
length.

## The audit

`asgard-cli check` finds the structural rot: a page with no inbound link, a
module missing from its index, a filename that is not dated, a link that does
not resolve. Run it first, and fix its findings before reading for
contradictions.

    asgard-cli check

What is left needs somebody to read the pages against each other. Three targets.

### 1. Contradictions

Two pages that cannot both be true. Read for these in this order, cheapest
first:

1. A decision record against the living spec module it changed. A decision
   is applied by rewriting the module; if the module still describes the old
   behaviour, the decision was recorded and never landed. This is the most
   common one.
2. A `done` task spec against the living spec. A task that changed behaviour
   is not done until its delta reached the spec, but the status was moved by
   hand and the delta was not.
3. Two decision records on the same topic. Both are immutable and both stay,
   which is fine as long as one says which one won. If neither does, write a
   third that supersedes, instead of editing either.
4. `AGENTS.md` against the living spec. These answer different questions -
   how to change the repo, versus what the system does - so overlap between them
   means a fact has two homes, and the two copies will drift.
5. A `assets/skills/` skill against the living spec. The most expensive kind,
   because the skill is what the running agent believes. A stale skill changes
   production behaviour.

For each one, report which side you believe and why. For example: "the decision
is dated later and the spec module was last touched before it, so the spec is
stale". Reporting only "these disagree" leaves the reader to redo the work.

### 2. Superseded claims

A statement that was true when written and has been overtaken. Unlike a
contradiction, nothing in the repo disagrees with it; what it describes changed
and the page did not.

  - A claim about the platform that a newer CRD or deployment contradicts.
    Platform vocabulary moves: a field is retired, a default flips, a resource
    is deprecated. Nothing on a page shows that a field it names no longer
    exists.
  - A claim about the customer's systems with no provenance, or with
    provenance old enough to doubt. `requirements/README.md` requires every
    claim about a real system to say how it was established and when; anything
    undated is a finding on its own.
  - An open question that a later decision answered, and that was never moved
    to Answered in `docs/open-questions.md`.
  - A `references/` document presented as current that the customer supplied
    long enough ago that nobody should rely on it unchecked.

Verify before reporting where verifying is cheap: a schema claim against the
database, a CR field against the platform's CRD. Where it is not cheap, say what
would settle it and leave it as an open question with an owner. Do not record a
guess as a finding.

### 3. Concepts nobody owns

These are the hardest to see, because nothing is wrong on any single page. A term appears
across five documents, everybody explains it in passing, and no page owns it.
The explanations drift, and a new reader assembles a definition from whichever
ones they happened to read.

Find them by noticing repetition: a term explained more than twice in different
files, or explained differently in two of them, needs a home. One fact, one
home, everywhere else links - the same rule `requirements/README.md` states.

When you propose a page, say which sentences in which files it replaces, and
remove them. A new page that replaces nothing adds another copy.

## Reporting

Write findings; do not fix contradictions silently. Which side is right is
frequently a question for a person - the FDE who was in the room, or the
customer.

Order by the cost of being wrong, not by how obvious it is:

    1. anything under assets/skills/     the running agent believes it today
    2. the living spec                   the next person will believe it
    3. decisions and requirements        found only when somebody goes looking

For each: the pages, which one you believe, the evidence, and the smallest fix.
Land the settled ones through the normal loop rather than by editing pages
directly - a new dated decision record, a delta applied to the living spec
module, a row moved to Answered. Do not edit a decision record to make it agree
with the current state; decision records are a history and stay as written.

End by re-running `asgard-cli check`. Applying a delta often leaves a new page
that nothing links to yet, and the check finds that.

**Checked:** no platform claims to check. This page is about
keeping a body of written material healthy, and the three kinds of rot it names
are properties of documents rather than of Asgard. Its own subject is checkable
by running the audits it describes.

**Unchecked:** that the three kinds of rot named here are the ones that matter in
a customer's repository. They come from maintaining this tool's corpus, where a
fourth - upstream moved and the corpus did not - is recorded on the wiki's own
README; a reader maintaining a different corpus should expect their own fourth.
