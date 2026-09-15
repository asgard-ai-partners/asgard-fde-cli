---
description: a Drive with contextIndex rather than a KnowledgeBase, the immutable Syncer paths, what a person still has to do after the deploy
---
# Decide where unstructured knowledge lives

Optional stage. Do this only if the customer has knowledge that is **not rows in
a database** - product documents, FAQs, pages on a website.

**Whether they need to see where an answer came from is a customer question, and
by this stage the interview is over** - so it should already be in the request.
If it is not, it is a phone call rather than an assumption. Citations are
available - the sources arrive on the completion event inside the message's
`template` - but **only if the Workflow is built to return them**, and only if
the front end reads that field. It is a decision made here, and retrofitting it
means changing the workflow and the front end together.
`../wiki/knowledge.md` has the shape.

**This is the third decision that gets answered wrong**, and it has one answer:

    SourceSet + spec.contextIndex, mounted read-only, queried with graphify

`KnowledgeBase` is still a live CRD and still a shipping console feature, so
this is a recommendation rather than a platform rule. It rests on one engagement
that built the other way and reversed.

> Answered wrong once: knowledge was built on KnowledgeBase + Loader + retrieval
> workflows, and was later rebuilt on a Drive with a Context Index (TASK-013).
> An older chart containing one is a shape somebody chose before that, not
> proof the mechanism is going away.

    ../usecase/knowledge-drive.md

## The shape

A SourceSet declares members, each fed by something:

  a database Syncer   incremental, cursor in the Syncer status
  a web Syncer        page list in version control, not a deep crawl
  manual uploads      documents nobody can sync automatically

Setting spec.contextIndex is the whole switch: the reconciler derives three
same-named CRs that mount the Drive writable and run the indexer on a cron.

**To pause it, label the SourceSet context-index-suspend. Do not remove the
field** - clearing contextIndex tears the three CRs down and renames the index
aside.

Mount it read-only on the SandboxBlueprint. The agent queries the knowledge graph
with the sandbox's builtin graphify skill and then reads only the few files the
graph points at - tell it that in the prompt, or it will crawl the whole Drive.

## Two things that will bite

  - destinationPath must end with / for database and web Syncers, and statePath
    must not. Both are immutable once set, so a typo is a new Syncer rather than
    an edit. The CRD enforces both, plus that neither is absolute and neither
    contains `//` or `../`.
  - The member registry these used to name is retired **as the mechanism, not as
    a field**: `members` is gone from the SourceSet, but `destinationMemberKey`
    and `stateMemberKey` are still on the Syncer, marked deprecated and kept so
    pre-rename objects stay readable. **A chart that still sets them lints clean
    and dry-runs clean**, which is why this is written down rather than left to
    be discovered. The paths a Syncer writes to are now the whole truth about
    what is in a Drive.
  - **Exactly one column may carry isMaxValueColumn**, and the CRD refuses a
    second. On a database Syncer, isMaxValueColumn and isIdentifier sit on a column, not
    on the database block, and both query and batchSize are required. A flag
    written one level up is an unknown field the apiserver drops in silence,
    leaving a Syncer that re-reads the whole table every run.
  - After deploying, someone has to upload the manual documents and let the
    Syncers and the index run once. Until then the knowledge answers are poor.
    Say so in the chart README rather than letting a demo discover it.

Done when: the Drive syncs, or you have decided the customer has no unstructured
knowledge and recorded that.

**Checked:** 2026-09-04, re-read 2026-09-11 against asgard-kube `cbd8d70`. `SourceSet.spec` carries
`contextIndex` and no `members`; `Syncer` enforces destinationPath ending in `/`,
statePath not ending in one, both immutable, neither absolute and neither
containing `//` or `../`; one column at most may set `isMaxValueColumn`;
`KnowledgeBase` is still a CRD in the contract, so "still live" is current. One
statement was too strong and is corrected: the member keys are **deprecated and
still accepted**, not removed, so a stale chart passes every mechanical check.

**Unchecked:** that a Drive plus a Context Index is the right answer and
KnowledgeBase is not, and that a graph the prompt does not point at gets crawled
whole. Both come from the engagement this was written in - the first was built
the other way and rebuilt (TASK-013). **Neither has a source**, and the console's
own behaviour is what would settle the second.
