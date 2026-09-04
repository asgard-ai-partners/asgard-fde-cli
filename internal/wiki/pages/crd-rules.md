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
| `botProviderClass` | `BotProvider` - see [`integration.md`](integration.md) |
| `completionModelClass` | `CompletionModel` - see [`settings.md`](settings.md) |
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

The CRDs carry 79 CEL rules. **Forty are `self == oldSelf`** - they compare a
proposed object against the one already on the cluster, so a render, which is
one object with no history, cannot see them. `botProviderClass` is the one that
bites; `asgard-cli usecase chat-channel` says why.

The rest are two families, and `asgard-cli verify` checks both as of
2026-09-04:

    exactly one of [...]        a credential that is neither a literal nor a
                                reference, or both; a class block that is
                                missing or doubled
    class implies its block     `toolsetClass: mcp-server` without
                                `mcpServerConfig`; a `documentClass` without
                                the block named after it

Every one of those renders, lints and passes a server-side dry-run, and is
refused at apply. Run over the six reference deployments - 107 CRs - the check
reports nothing, which is what a rule the platform already enforces should do.

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
  - asgard-kube `15ded0f`
- The generated CRDs carry the same rules without the reasoning:
  [asgard-kube `crd/`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd)

**Unchecked:** none of these has been seen to fire. They are read off the
declarations rather than from a deployment that hit one, so what is confirmed is
that the rule exists and what it says - not what the failure looks like in CD.
The evaluation-time one is the exception worth treating as urgent anyway, since
by construction it cannot fail at deploy.
