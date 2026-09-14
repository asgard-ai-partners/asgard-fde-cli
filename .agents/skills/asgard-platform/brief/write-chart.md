# Before write-chart

before authoring or editing CRs

Three decisions in this repository's history were made the obvious way,
built, and reversed. They are obvious in the same way again each time.

## the read surface for a public audience

**Said:** a SemanticLayer, like everything else

**True:** Five zero-parameter query tools. A layer without `allowedCubes` is arbitrary SQL over every cube, and the surface grows by itself each time one is added

Read `../usecase/fixed-query-tools.md`.

## the entry point for an anonymous visitor

**Said:** the platform's agent hub, like everything else

**True:** Its own BotProvider -> Workflow -> SandboxBlueprint. An anonymous visitor cannot authenticate to the hub, and `BotProvider.entrypoint` takes a Workflow, never an Agent

Read `../usecase/flow-agent-single.md`.

## where unstructured knowledge goes

**Said:** a KnowledgeBase with Loaders and a retrieval workflow

**True:** A SourceSet Drive with `contextIndex`. Both are live; the Drive is the recommendation for new work

Read `../usecase/knowledge-drive.md`.

## a write inside a scheduled run

**Said:** omitting `allowWrite` because it is read-only anyway

**True:** **The definitions give `allowWrite` a default of true and `allowQuery` a default of false** - the safe field is off by default and the dangerous one is on. So configuring a processor by adding only what you want produces a write path, and the rendered chart does not show it. Write `allowWrite: false` out on every entry that has a layer, and note that `query-database` carries the same pair

Read `../wiki/processors.md`.

## what an Expression may use

**Said:** ECMA5 only, because the documentation says ECMA5

**True:** **That limit is `execute-script`'s Engine field and does not reach an Expression.** They are ordinary JavaScript: `prevBlobs.map(b => b.blobId).join(',')` evaluates in a shipped chart, and `const` gets in through an IIFE, which is the form this tool generates. A bare declaration has nowhere to go only because the field holds one expression. Writing an Expression defensively costs nothing; **refusing a shape because of the wrong limit** is the failure, and this material has stated it both ways inside one page

Read `../wiki/processors.md`.

## a SandboxBlueprint's subagents

**Said:** it deployed, so the blueprint is right

**True:** **Exactly one of `baseAgentName` or `aliasName`, and no CRD checks it** - the shape rides inside a JSON string where CEL cannot see it, so the controller enforces it at evaluation time. Both set, or neither, deploys green and fails the first time somebody talks to the agent. The only rule worth reading a blueprint by hand for

Read `../wiki/crd-rules.md`.

## whether the chart is enough

**Said:** a green render means it is done

**True:** A Workflow needs a `ConfigMap` of node positions or its editor opens as a pile, and `project-environment-id` or the editor opens blank. Neither is an Asgard CR, so nothing in the gate mentions them

Read `../wiki/workflow.md`.

`asgard-cli check` and `asgard-cli verify` catch the shape. None of the
above is a shape problem, which is why they are here.

**Checked:** every entry above is here because it actually happened, and each
names the document that carries the right version.

**Unchecked:** whether the list is complete. It grows when somebody gets
something new wrong, so an activity with few entries is not a safe one.
