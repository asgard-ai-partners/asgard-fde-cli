# MCP Server, Skillset and Plugin

Three things that are easy to confuse. The difference is what each one holds.

| | holds | CRs |
|---|---|---|
| **MCP Server** | callable tools - functions, external APIs, data sources | `Toolset` |
| **Skillset** | instructions the agent reads | `SkillSet` + `SourceSet` + `Syncer` |
| **Plugin** | a bundle of those, plus Drives, Hooks and Managed Agents | `Plugin` |

In one line: **an MCP Server is what the agent can do, a Skillset is what it
should know, and a Plugin is a set of capabilities behind one name.**

## MCP Server

A Model Context Protocol server, letting an agent reach external tools, data
sources and custom functions.

**From Workflow** wraps a Workflow as an MCP Server. It takes a Name and a
Description; creating it produces an empty Workflow where the actual tool logic
goes. Maps to `toolsetClass: workflow-tooling`.

**From Existing MCP Server** connects to one that already runs. Maps to
`toolsetClass: mcp-server`.

| field | required | |
|---|---|---|
| Name, Description | yes | |
| Transport Type | yes | STDIO or Streamable HTTP |
| Command | yes | the launch command; `npx` and `uvx` are supported |
| Arguments | | space separated |
| Environment Variables | | Name/Value pairs, passed over STDIO |
| MCP Server Volume Mounts | | mount a Drive into the server's environment, with Mount Path, Sub Path and Read Only |

With STDIO, Asgard starts a local process and talks to it over standard
input/output. With Streamable HTTP it connects to an endpoint already running.

## Skillset

A reusable set of skills an agent loads at run time.

**From Scratch** takes a Name and optional Search Paths. **From Git** imports
from a repository.

Search Paths are where the agent looks for skill files, comma separated, and
**each path must end in a slash**.

The detail page has Files and Settings tabs. Files is that Skillset's own
browser, and Open in Advance Editor edits the skill files in a separate tab.

A Skillset maps to three CRs: `SkillSet` plus its own `SourceSet` plus the
`Syncer` that fills it, 1:1:1. **A Skillset a Plugin bundles is the exception**:
when the skills all live in one repository, several bundles share one SourceSet
and slice it with searchPaths, at the cost of those Skillsets not being presented
as first-class objects a person picks. For the chart details see
`../usecase/skill-set.md` and `../usecase/plugin.md`.

## Plugin

Packages MCP Servers, Skillsets, Drives, Hooks and Managed Agents into one
reusable sandbox bundle.

It earns its place when **the same agent needs different capabilities on
different turns**, and a blueprint decides per request which bundles to load. If
the capability set is fixed, mount the pieces directly and skip this layer.

Hooks belong to a Plugin and to nothing else.

### Hook events

| event | fires | status |
|---|---|---|
| `session-start` | once, when the container starts | usable |
| `session-end` | on SIGTERM | usable |
| `user-prompt-submit` | before each user message reaches the CLI | usable |
| `pre-tool-call` / `post-tool-call` | - | **never implemented; declaring one is a silent no-op** |

The last row exists only so older CRs stay valid, and nothing catches it for
you: the generated CRD lists both events in the enum with no marking, so a CR
declaring one is accepted and then does nothing. The statement comes from the
API types themselves ([asgard-kube](https://github.com/asgard-ai-platform/asgard-kube)
`15ded0f`, `pkg/apis/asgard/v1alpha1/types.go`): "Deprecated: never implemented
... declaring a hook with either event is a silent no-op". A `session-start` hook's
content has to be stable; anything derived from the turn's payload belongs in
`user-prompt-submit`.

## The sandbox's own tools cannot be switched off

An agent runs in a sandbox that is a coding-agent CLI, and **that CLI's built-in
tools are present in every sandbox** - web search, web fetch, task and schedule
listing. They are not Asgard domain tools, they are not a leak from the agent
hub, and **no Toolset or SandboxBlueprint setting removes them**. The platform
hard-codes its disallow list to two planning tools and there is **no field on any
CRD to opt out**.

So the only control today is the prompt, which is a weak defence and the only
one there is:

    a deployment that needs them off   says so in the prompt, explicitly
    that instruction                   must not be deleted as redundant

**This matters in front of a customer who asks what the agent can reach.** The
honest answer is that it can search the web unless told not to, and that the
instruction not to is a prompt rather than a permission. Anyone answering "it
only sees what you connect" is wrong.

## `Toolset.spec.instruction` is gone

Removed from the live CRD. Tool usage guidance now lives on
`Workflow.entries[].tooling.description`. A chart carrying `spec.instruction` is
carrying a field the apiserver no longer knows, and re-adding it is a common
repair to make when guidance seems to be missing.

## Sources

- [MCP Servers](https://docs.asgard-ai.com/docs/product-suite/odin/features/mcp-servers)
  - asgard-docs `f00e0ee`
- [Skillsets](https://docs.asgard-ai.com/docs/product-suite/odin/features/skillsets)
  - asgard-docs `f00e0ee`
- [Plugins](https://docs.asgard-ai.com/docs/product-suite/odin/features/plugins)
  - asgard-docs `f00e0ee`. That page is short - it covers the list and the New
  Plugin button, and does not document the creation form's fields
- Hook events, the SkillSet trio and the shared-store exception: checked
  2026-09-02 against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube)
  `15ded0f` - `SandboxHookEvent`, `PluginSpec` - and against a deployment
  carrying 28 Plugins

**Unchecked:** the three roles and the hook events were held against the CRD, and
the shared store against one deployment; the UI form fields come from the product
documentation only.
