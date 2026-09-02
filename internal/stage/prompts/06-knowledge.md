Optional stage. Do this only if the customer has knowledge that is **not rows in
a database** - product documents, FAQs, pages on a website.

**This is the third decision that gets answered wrong**, and it has one answer:

    SourceSet + spec.contextIndex, mounted read-only, queried with graphify

KnowledgeBase is deprecated platform-side. If you find it in an older chart or a
CR dump, it is not a template to copy.

> Answered wrong once: knowledge was built on KnowledgeBase + Loader + retrieval
> workflows, and the whole mechanism was later removed.

    asgard-cli usecase knowledge-drive

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
    an edit. (The member registry these used to name - members on the SourceSet,
    destinationMemberKey and stateMemberKey on the Syncer - was retired: the
    paths a Syncer writes to are now the whole truth about what is in a Drive.)
  - On a database Syncer, isMaxValueColumn and isIdentifier sit on a column, not
    on the database block, and both query and batchSize are required. A flag
    written one level up is an unknown field the apiserver drops in silence,
    leaving a Syncer that re-reads the whole table every run.
  - After deploying, someone has to upload the manual documents and let the
    Syncers and the index run once. Until then the knowledge answers are poor.
    Say so in the chart README rather than letting a demo discover it.

Done when: the Drive syncs, or you have decided the customer has no unstructured
knowledge and recorded that.
