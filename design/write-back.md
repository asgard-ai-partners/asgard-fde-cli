# The write-back path

A design note. `TASK.md` states the problem: the corpus is compiled into the
binary, so an engagement that learns something has nowhere to write it where it
will be read. `issue-report` is the way back, and the slow step - a person -
is the right one.

This note argues that the route is not broken, that the unsolved part is
narrower than "write-back", and that the narrow part has a shape this
repository already uses elsewhere.

---

## 1. What an engagement actually learns, and where each half belongs

Four kinds, and the split is not by subject. It is by **who can re-read the
evidence**.

### A. The platform behaved unexpectedly - belongs here

The engagement is the only party that has seen it, and the next engagement will
hit the same thing. This is the half the corpus exists for, and it is what
Goal's fourth point is a route for.

  - **The platform-minted resource api key can be read by cross-object
    reference.** `preset-agent-hub` / `api_key`, referenced directly from a
    SourceSet, a Toolset or a BotProvider. Nothing documented it; a deploy was
    blocked on obtaining a value nobody could obtain. It is now
    `internal/corpus/usecase/conventions.md` and it narrowed
    `internal/corpus/wiki/platform-unknowns.md` P13.
  - **A `~/.claude` mount kills the driver**, and **a writable store needs its
    own Syncer-less SourceSet** - both read off deployment clones rather than
    off any document.
  - **SMTP cannot be reached at all**, so the answer a customer actually hands
    over is credentials, not an endpoint. `internal/stage/prompts` was asking
    the wrong question and the answer never fits.
  - **`prevToolCalls` is in scope in an expression and appears in no
    documentation page.** It was found by reading a chart, and it turned "the
    list of six" into a floor rather than a set - P11.

None of these is a defect report. **Each is an answer**, and that matters for
section 3.

### B. This customer's system is shaped oddly - belongs there

Their table names, their back office, their approval queue, their test plan.
This half is solved and the mechanism already exists: `asgard-cli question`,
`decision`, `request`, `reference` write into the customer's repository, and
`issue-report --new` deliberately reports them as **counts** rather than
contents, because their rows are the customer's own system names.

The boundary is not "platform versus customer" but **generalisable versus
not**. A customer's test plan requiring per-user scoping stayed in their
repository; the question it raised - can one user be allowed to see only some
resources - came here as P1 with the customer erased. The corpus already has
the form for anonymised provenance: `**Seen in:**`, which names a shape and
never a name.

### C. The tool was wrong - an issue, and this class works

`add httptool` generating no relationships, so the tool never makes its call
(#115). `--force` discarding the values every `add` had written. R1 refusing a
shape the CRD and the platform both accept (#97). `pipeline project create`
returning 500 after creating the project (#94). Reproducible, and the
maintainer can stand where the reporter stood.

### D. The material was unfindable - an issue about retrieval, not content

Different repair, and the report has exactly one place that tells the two
apart: section 4, "where the answer actually was".

  - `稽核 -> logging` routed every reader translating that term to a search the
    corpus has never matched. The page existed; the row pointed past it.
  - The marketplace names were in `asgard-freyr-skills` the whole time - the
    example `internal/cli/issue.go` uses in its own help.
  - Issue 43: the empty-result fallback told a reader to open a question and
    not what makes one worth asking.

**A missing page and an unfindable page need opposite repairs**, and an issue
that omits section 4 cannot be sorted into either.

---

## 2. What is wrong with the route today

**Not that nobody uses it.** That was the available guess and the record
refuses it:

    37 issues filed, 23 of them by an FDE who is not the maintainer
    33 of 37 carry the report's section headings
    11 carry the `--new` footer verbatim
    every one closed inside a day, most inside ten hours

So the mechanism is used, the generated body is used, and the turnaround is not
the bottleneck. Three things are wrong, and they are narrower than "write-back".

**2a. The template is a bug report, and a discovery is not a bug.** The five
sections are: what I was trying to do, the state I was in, what I ran and what
came back, where the answer actually was, what it cost. There is **no section
for what I now know.** An engagement that proved the api key is readable by
cross-object reference has to file it as a complaint - which is what issue 109
is: "the console page the material names does not exist." True, and it is the
smaller half of what that engagement had. The claim, the shape it was proved
on, and how far it was proved were then **re-derived by the maintainer** and
written into `conventions.md` afterwards. That re-derivation is the cost, and
it is paid every time the most valuable kind of learning arrives in a form
built for the least valuable.

**2b. The claim an engagement can make is one the corpus has no slot for.**
`**Checked:**` names a source anybody can re-read: a commit, a CRD, a page.
An engagement's evidence is a namespace that no longer exists and that nobody
here ever had. `internal/corpus/usecase/conventions.md` already carries one,
hand-written:

    **Proved on a deployed namespace, 2026-09-14, and how far differs by kind.**
    A SourceSet is proved end to end ... a Toolset is proved to the server-side
    dry run ... a BotProvider's `adminApiKey` ... has not been exercised at all.
    Treat the first as settled, the second as very likely, and the third as
    reasoned.

That paragraph is the design, invented once under pressure, in prose, in one
page. It is reachable by nothing, comparable to nothing, and there is no way to
ask which other claims in the corpus are of this kind. `kb` parses two markers
and this is neither.

**2c. The tool cannot observe the moment of failure, by design.** Retrieval is
`grep` - there is no search command, deliberately, and `APPROACH.md` says so.
So nothing can print "nothing matched, file it" at the moment nothing matched.
The route is printed at the foot of the landed `index.md`, once in
`AGENTS.md.tmpl`, and in one stage prompt: **all places read at the start of a
session, none read at the moment of the gap.**

The filing record shows the consequence. Twelve issues arrived in one batch on
2026-09-08 and four more in one batch on 09-13, all closed together. Write-back
happens at the end of a session, from memory - and sections 4 and 5, the two
only answerable at the moment, are the ones that thin out down a batch: #96
carries section 2 alone, and #97, #98 and #103 carry no sections at all.

**2d, and it is free.** The tool already knows one write-back signal and does
not use it. `scaffold.classify` reports `edited` for a shipped file an
engagement changed in place - which is an engagement disagreeing with shipped
material, in writing, with the diff on disk. `gate`'s `shipped` step sees it
and `issue-report --new` does not mention it.

---

## 3. Three candidates

### A. Keep `issue-report`, and make it carry a discovery

**What a person does.** Files as they do today, and where they have an answer
rather than a complaint, fills a sixth section: what I now know, the shape it
was seen on (never the customer), and **how far it was proved** - settled, very
likely, or reasoned, the three grades `conventions.md` already invented.

**What the tool does.**

  - `--new` emits the sixth section, and a `--learned` shape that puts it first
    and marks the bug-report sections optional. The grade is prompted as three
    named choices, not free text, because a grade with no vocabulary is a
    sentence nobody can compare to another sentence.
  - Section 2 gains the one fact it already has and omits: which shipped files
    are `edited`, by path, from `scaffold.InspectShipped`. **Paths only, never
    contents** - the same rule that makes `question` and `request` arrive as
    counts.
  - The route is printed where the gap is met rather than where the session
    starts. Every document already carries `**Unchecked:**` naming the surface
    it was not held against; `audit-material --unchecked` can end with the
    filing line, and the landed `index.md` can say that meeting an `Unchecked`
    marker is the moment to file rather than the moment to work around.

**Cost.** One more section that will sometimes arrive as TODO, and a grade
vocabulary that has to be kept to three words or it becomes prose.
**What it breaks.** Past issues and future issues stop having the same shape,
which matters because `issue.go`'s help promises "a reader of the repository's
issues should not have to learn two shapes." Two shapes is what this proposes,
and the argument for it is that a defect and a discovery are already two
things.

### B. A landed capture file the engagement appends to, harvested at filing

**What a person does.** `asgard-cli note add "..."` at the moment, appending a
dated row to a scaffolded `docs/platform-notes.md`. At the end of the session,
`issue-report --new` includes the notes and empties nothing.

**What the tool does.** Owns a sixth work file, its template, its parse, and
its own directory README saying what it is not.

**Cost.** A sixth work file and a second parse.
**What it breaks.** Two rules, both load-bearing. `issue.go`'s first paragraph
forbids exactly this - "a note in one engagement's docs is a note one
engagement has, and the next one starts over" - and a file that can hold an
unfiled note is a file that will, which is the rot the compiled corpus exists
to prevent. It also gives the material a body with its own reader, which
`AGENTS.md` refuses. **It does solve 2c**, and it is the only candidate that
does, which is why it is here rather than dismissed in a line.

### C. Change nothing about the route; give the corpus a slot for the claim

**What a person does.** Nothing new. The maintainer, writing an
engagement-sourced claim into a page, writes a third marker instead of a
paragraph:

    **Proved:** a SourceSet reading `preset-agent-hub` / `api_key` by
    cross-object reference, on a deployed namespace, 2026-09-14 - applied,
    Ready, Syncers ran. Settled. The Toolset form is proved to the server-side
    dry run only: very likely. The BotProvider form is reasoned.

**What the tool does.** `kb.ParseDoc` gains one entry in the map it already
has - `**Proved:**` beside `**Checked:**` and `**Unchecked:**`. One reader, one
schema, one line of code. `audit-material --proved` lists every such claim with
its grade, the way `--unchecked` lists every surface.

**Cost.** Almost nothing.
**What it breaks.** Nothing - and alone **it routes nothing**. It is a landing
shape, not a path. It belongs to A rather than competing with it.

---

## 4. Recommendation

**A, with C as the half that lands.** Extend `issue-report` with a discovery
shape and a three-word grade, print the route at `**Unchecked:**` rather than
only at the start of a session, put the `edited` paths into section 2 - and add
`**Proved:**` to `kb` so the claim lands as a marker every reader meets rather
than as a paragraph one page happens to carry. Reject B.

The argument, in three steps.

**The route is not the problem, so do not build a second one.** The evidence
above is unambiguous: filed, used, closed same-day, mostly by somebody who is
not the maintainer. A new mechanism here would be the thing this repository
keeps deleting - and B in particular would be a mechanism that reintroduces the
per-repository note the compiled corpus exists to abolish, to solve a capture
problem that has not been shown to be the binding one.

**The binding constraint is the shape of the claim, not the shape of the
pipe.** The one engagement discovery that reached the corpus this week reached
it as a complaint and had to be reconstructed. That is a template problem and a
schema problem, and both are small.

**The repository's own `reconcile` pattern applies exactly, and says what the
mechanism's job is.** `reconcile` records a claim nobody can check, because the
honest alternative - deriving one - would be false. An engagement's "proved on
a deployed namespace" is the same shape: the evidence is gone, no future reader
can re-read it, and marking the page `**Unchecked:**` instead would be **false**,
because somebody did check it. So the mechanism's job is to make the claim
cheap to record and visible, **not to validate it**. Two corollaries follow
straight from `reconcile` and should be built in:

  - **No bulk record.** `reconcile` has no `--all` because that would write a
    claim nobody made. There must be no way to mark an engagement's learnings
    proved in one action.
  - **The grade is written by whoever proved it, never derived.** Three words,
    and a page saying "reasoned" is more useful than a page saying nothing,
    because it tells the next engagement what is left to do.

Every stated constraint holds: the claim still passes a person (a GitHub issue,
reviewed); nothing customer-specific enters, because the discovery section asks
for the shape and `**Seen in:**` is the existing precedent for that; there is
one schema and one reader, because `**Proved:**` is one entry in a map that
already exists; `issue-report` keeps working with no repository, no network and
no login, since nothing proposed here reads anything it does not already read.

---

## 5. What would tell us this is wrong

Three observations, in the order they would arrive. Each is countable, which is
the point - none of them is a judgement call.

**The falsifier for the recommendation.** Over the next two engagements, count
the issues whose discovery section is filled against those where it arrives as
TODO. **If most arrive as TODO while the bug-report sections are filled**, the
problem was never the template's shape - it was that write-back happens at the
end of a session from memory, and a section added to a form somebody fills in
an hour later cannot fix that. B's premise would then be the right one, and the
answer would be capture at the moment, in the repository, with the rot accepted
as a cost rather than refused as a principle. The burst pattern in the existing
record is already weak evidence for this, and it is the honest risk in the
recommendation.

**The falsifier for the grade.** If `**Proved:**` markers accumulate and **no
grade ever moves** - if no claim goes from reasoned to settled across two
engagements that both touched that shape - then the grade is decoration. It
should be deleted rather than softened, by this repository's own rule about a
check that fires on correct material.

**The falsifier for the whole diagnosis.** If the maintainer still has to
re-derive the claim from the filed issue - if the commit that lands a discovery
says materially more than the issue did, the way `conventions.md` says more
than issue 109 - then the section did not carry the knowledge, and the reason
is something other than a missing heading. Diff one landed discovery against
its issue and the answer is visible in a minute.
