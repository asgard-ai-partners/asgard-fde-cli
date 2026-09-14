# Rules the schema enforces, and one it cannot

The CRDs carry validation beyond required-and-type, written as CEL expressions
the apiserver evaluates. **`helm lint` does not run them and a server-side
dry-run does not report them faithfully**, so the first sight of one is usually a
red deploy.

They are here because they are read off the Go type definitions in
`asgard-kube/pkg/apis/`, which carry the reasoning as comments - and this
material had only ever read the generated CRD YAML, where the rule survives and
the reason does not.

## Immutable once set

Changing one of these is a new CR, not an edit. `helm upgrade` produces a
rejection, not a replacement.

| what | on |
|---|---|
| `botProviderClass` | `BotProvider` - see [`integration.md`](../wiki/integration.md) |
| `completionModelClass` | `CompletionModel` - see [`settings.md`](../wiki/settings.md) |
| the indexer | a `Source`'s indexer |
| `leaseId` | once set, on the resource that carries it |

## Toolset: which fields apply is decided by the class

Not a blanket exactly-one rule, and the asymmetry is deliberate:

    toolsetClass: mcp-server     mcpServerConfig REQUIRED, and tools must
                                 stay empty
    any other class              tools optional and may be empty; no
                                 mcpServerConfig

**A workflow-tooling Toolset with zero tools is legal**, because that is the
state between "the MCP server exists, with a name and a description" and the
first tool added to it. So an empty `tools` is not a sign of an unfinished
chart, and a check that treats it as one is wrong.

## SemanticLayer: two rules that fail late

**A Measure needs `sql` unless its type is `count`.** Counting is the one
aggregate the platform can compose itself.

**A Join's `from` and `to` must name the same number of dimensions.** A join
written against a two-column key with one column listed passes review by looking
reasonable and is refused by the apiserver.

Both are worth checking before a tag, because a chart with 31 cubes has enough
of them that one is usually wrong, and neither `asgard-cli check` nor `helm
lint` looks.

## The builtin model holds no key, deliberately

`CompletionModel` of class `builtin` carries only a tier alias - no API key and
no concrete model name. The model router maps the alias to a real model and
supplies its own managed key, **so no key lands in a CR or in a per-namespace
Secret**.

That is the reason to prefer a builtin tier when the customer has no view: it is
one less credential in the engagement, and a concrete model name is one more
thing to come back and fix when that model is retired.

## Which of them a render can be held against

**There are 79 CEL rules written and 231 enforced, and the difference is not a
rounding.** 79 is the number of `XValidation` markers in asgard-kube's Go
types; the generated CRDs carry 231 rule instances, 52 of them distinct,
because one marker on a struct several kinds embed lands in every CRD that
embeds it. **Hold a render against the CRDs and not against the markers**: the
generated schema is the contract and the Go types are only its source, so the
two counts answer different questions and the smaller one is not the platform's.

**41 of the enforced rules are exactly `self == oldSelf`** - 40 of the markers -
and they compare a proposed object against the one already on the cluster, so a
render, which is one object with no history, cannot see any of them. Nothing
offline can tell you a chart will be refused at apply. **What can be said
offline is which fields they are**, and that is the useful half:

### Which fields are chosen once

**Every resource's class is immutable.** `agentClass`, `botProviderClass`,
`completionModelClass`, `dataConnectorClass`, `embeddingModelClass`,
`imageGenerationModelClass`, `knowledgeBaseClass`, `loaderClass`,
`sourceClass`, `syncerClass`, `transcriptionModelClass` - eleven of them, one
per kind that has a class. So `botProviderClass` is not a special case, it is
the instance of this rule an FDE meets first: **changing what kind of thing a
resource is means a new resource with a new name**, and the old one's references
have to move.

**The Syncer is where this costs the most: 21 of the 40.** Not just
`syncerClass` but **where it reads from and where it writes to**:

    sourceSetName            which store it fills
    destinationPath          the path inside that store
    statePath                where it keeps its cursor
    ftp.host  ftp.remotePath
    sftp.host sftp.remotePath
    smb.host  smb.remotePath
    dropbox.folderPath       and its oAuthCredentialName
    googleDrive.folderId     and its oAuthCredentialName
    oneDrive.folderId oneDrive.folderPath  and its oAuthCredentialName
    bot.botProviderName
    database.columns

**So a Syncer is not repointed, it is replaced.** "Sync from this folder
instead" is a new Syncer and a deleted one, not an edit - and the cursor goes
with it, so the new one re-reads from the beginning unless `statePath` is
handed over deliberately. `../usecase/skill-set.md` writes one; nothing in
that page said this.

**The Loader is the same shape, smaller.** `knowledgeBaseName` is immutable, so
a Loader cannot be pointed at a different knowledge base, along with its
`loaderClass`, its Drive folder ids and its credential names.

**And one that reads like a typo and is not:** `Indexer.spec.xlsx` is immutable.
Whether a Drive's index treats spreadsheets as tables is decided when the
Indexer is created.

**Checked:** 2026-09-11, by walking every `self == oldSelf` rule in asgard-kube
`cbd8d70` `crd/` back to the property that carries it - 40 properties across
twelve kinds. The count of rules is 41 because one kind carries the same rule
at two paths.

The rest are two families, and `asgard-cli verify` checks both as of
2026-09-04:

    exactly one of [...]        a credential that is neither a literal nor a
                                reference, or both; a class block that is
                                missing or doubled
    class implies its block     `toolsetClass: mcp-server` without
                                `mcpServerConfig`; a `documentClass` without
                                the block named after it

Every one of those renders, lints and passes a server-side dry-run, and is
refused at apply. Run over the six renderable reference charts - 119 CRs as of
2026-09-11 - the check reports nothing on either family, which is what a rule
the platform already enforces should do.

## An undeclared field is pruned, and a dry run says success

**This is the one that breaks a release after everything passed.** A CRD
discards a field its schema does not declare, silently:

    kubectl apply --dry-run=server     reports success, field already discarded
    helm's server-side apply, in CD    fails with `field not declared in schema`

So the two are not the same check, and the cheap one answers a different
question. **`crd/dry-run-rejected` is "will it be accepted"; `crd/unknown-field`
is "will it be kept."** The platform's plan report runs both; nothing local
runs the second, because pruning is an apiserver behaviour and no client is
issued cluster credentials.

It has happened. `Toolset.spec.instruction` was removed from the CRD and added
back by hand; the chart passed **25 of 25 dry runs** and the deploy failed.
`../usecase/write-path.md` and `../usecase/fixed-query-tools.md` both carry it
against the field they concern.

**So a green local gate is not evidence a field survives.** `asgard-cli gate`
says as much, and the authority is the plan:

    asgard-cli pipeline runs watch --release <name> --ref <tag>

## The one the schema cannot enforce

**A `SandboxBlueprint`'s subagent must set exactly one of `baseAgentName` or
`aliasName`, and nothing in the CRD checks it.** The shape rides inside a
JSON-string value, so CEL cannot see it. It is enforced by the blueprint
controller **at evaluation time**.

That is a different failure from every other rule on this page:

    a CEL rule          the apiserver refuses the CR. You find out at deploy
    this one            the CR is accepted, and the run fails when it is used

So a blueprint carrying both, or neither, deploys green and breaks the first
time somebody talks to the agent. It is the only rule here worth reading a
blueprint by hand for.

## Sources

- `asgard-kube/pkg/apis/asgard/v1alpha1/types.go`, read 2026-09-02 - the type
  definitions the CRDs are generated from, with the reasoning in comments
  - asgard-kube `cbd8d70`
- The generated CRDs carry the same rules without the reasoning:
  [asgard-kube `crd/`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd)
- The pruning behaviour: read off the two extracts that carry it against the
  field they concern, `../usecase/write-path.md` and
  `../usecase/fixed-query-tools.md`, which took it from a deployment. The two
  plan-report codes are the platform's own

**Unchecked:** none of the CEL rules has been seen to fire. They are read off the
declarations rather than from a deployment that hit one, so what is confirmed is
that the rule exists and what it says - not what the failure looks like in CD.
The evaluation-time one is the exception worth treating as urgent anyway, since
by construction it cannot fail at deploy.
