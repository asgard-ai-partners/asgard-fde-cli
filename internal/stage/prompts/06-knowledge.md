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

  - destinationMemberKey must end with / for database and web Syncers, and both
    it and stateMemberKey must name declared members. The CRD only checks this at
    runtime: a typo is a CronJob that fails every run while apply stays green.
  - After deploying, someone has to upload the manual documents and let the
    Syncers and the index run once. Until then the knowledge answers are poor.
    Say so in the chart README rather than letting a demo discover it.

Done when: the Drive syncs, or you have decided the customer has no unstructured
knowledge and recorded that.
