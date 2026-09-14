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

## Consent, and the two things about it that were believed wrong

**`requestConsent` is a field on the Toolset, per tool** - `spec.tools[].requestConsent` -
and not on the Workflow the tool wraps. A Workflow has no say in whether calling
it stops to ask; that decision lives one level up, in the Toolset that exposes
it. So a chart author looking for the gate in the Workflow will not find it, and
adding a second entry to a Toolset silently inherits nothing.

**A `workflow-tooling` Toolset does defer.** The belief that it could not - that
the class itself was a second, independent reason a tool never stopped to ask -
was wrong, and one deployment carried it in a comment for a week. The mechanism,
read at asgard-core `623ceb5`:

    asgard-core internal/constants.go
                                   the built-in safe list "only governs asgard
                                   domain tools; mcp__<toolset>__* honor
                                   RequestConsent"
    processor/driverloop           a toolset's tools reach the CLI as
                                   `mcp__<toolset>__<tool>`, so they carry the
                                   prefix consent gates on
    bpcontroller/server            only a tool with `!RequestConsent` enters
                                   `AllowedToolRefs`
    processor/consentpolicy        an `mcp__` tool that is not in that list is
                                   returned as **defer**

So the single switch is that one field. Everything the agent's own sandbox runs -
Bash, Read, Edit, Grep - is auto-allowed before consent is considered at all,
because those are not Asgard tools.

**An `mcp-server` Toolset cannot ask for consent at all.** There is nowhere to
write it: `requestConsent` exists only on `tools[]`, and asgard-kube's CEL rule
requires `tools` to be **empty** for that class. A design that plans to gate an
external MCP server's calls per tool does not work, and the CRD refuses it
rather than ignoring it.

**One bypass exists and is not a chart field.** `bypass_tool_call_consent` is a
query parameter on the Edge Server's bot-provider endpoint, defaulting to false,
and it treats every tool call in that one request as consented. It is a caller's
switch, so a relay in front of the platform decides whether it is reachable at
all - which is the thing to ask about before promising that a gate cannot be
skipped.

**Checked:** read 2026-09-11 against asgard-core `623ceb5` - its
`internal/constants.go`, asgard-core `internal/processor/consentpolicy/consentpolicy.go`,
asgard-core `internal/bpcontroller/server/bp_controller.go` and asgard-core
`internal/edgeserver/handler/bot_provider.go` - and against asgard-kube
`cbd8d70`, its `pkg/apis/asgard/v1alpha1/types.go`. The correction came from a deployment
chart that had traced it line by line; every step of it was re-read here rather
than taken on trust.

**Unchecked:** what the dialog looks like to the person answering, on a channel
that is not Sindri. `../wiki/platform-unknowns.md` P8 is that question and it is
still open - the deployment above produced its first real consent card in a
staff-facing dashboard, which is the case that was never in doubt.

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
`cbd8d70`, `pkg/apis/asgard/v1alpha1/types.go`): "Deprecated: never implemented
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

## The card tools the platform adds by itself

Separate from a Toolset and from the sandbox's own CLI tools: **the platform
puts these into every conversation**, and no CRD field declares or removes one.
They render something for the user or set conversation metadata, and they are on
the consent safe list - none of them stalls waiting for an approval, because
none has an effect beyond rendering.

| tool | what the user gets |
|---|---|
| `show_result_set_table` | a query's rows as a table |
| `show_vega_visualization` | a chart over a result set |
| `show_channel_home_download_link` | a download card |
| `open_sandbox_file` | a card that opens one file in the file viewer |
| `open_sandbox_folder` | a card that opens a directory in the file explorer's tree |
| `open_sandbox_browser` | a card that hands the sandbox's browser to the user |
| `show_canvas` | an HTML/SVG fragment the model wrote, rendered as a card |
| `update_channel_title` | the conversation's title |

**The file card and the folder card are not interchangeable, and getting it
wrong produces a card that can only fail.** The viewer reads and tails its path
(`fs/file` + `fs/watch`) and the sandbox filesystem API rejects both for a
directory. `open_sandbox_folder` exists because the model had one card and a
folder to show, so it aimed the file card at a directory.

**Both can only address paths inside the working directory**, because that is
where the file explorer is rooted - including for a file derived from a user's
attachment, which lands outside it. An agent asked to unpack an attached `.zip`
extracted beside the attachment and then pushed a card at that directory; the
user tapped a card pointing outside the tree.

**`show_canvas` is the one that is not delivered by a handler.** The fragment is
the tool's own `html` argument, so it streams to the client *before* the tool
executes - which is why consent on it would leave a half-drawn canvas on screen
waiting for an answer about content the user can already see.

## A query tool's rows are truncated, and the full set is a file

`execute_database_query` returns **at most 20 rows** to the model, with
`has_more` when there are more. Two different things to do with the rest, and
the platform's own tool instruction is emphatic about not confusing them:

    to show the user      pass `result_set_id` to show_result_set_table
                          or show_vega_visualization - they get every row
    to use it yourself    read `result_set_path`, a JSON file already written
                          inside the sandbox

`result_set_path` holds `{dataConnectorName, sql, rows}`, where `rows` is keyed
by that query's own output column names. **Point a script at it** - the rows
never enter the model's context, so the size of the result set stops mattering.
Re-running the query with LIMIT/OFFSET to page rows into context is the thing
this exists to avoid, and so is opening the file with a Read tool. The field is
absent when no file was written, and only then is paging the right answer.

**This is worth knowing in front of a customer** who asks whether the agent can
work over a large table: it can, and the mechanism is a file in the sandbox
rather than a bigger context.

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
  `cbd8d70` - `SandboxHookEvent`, `PluginSpec` - and against a deployment
  carrying 29 Plugins
- The card tools, the file/folder distinction, the working-directory rule and
  the result-set file: read 2026-09-11 from
  asgard-core `623ceb50` `internal/constants.go` - `BuiltinToolCallSafeList`,
  the `ToolName*` constants and the `execute_database_query` tool instruction. **No product
  documentation covers any of it**, and no CRD field declares one, which is
  why none of it was here

**Unchecked:** the three roles and the hook events were held against the CRD, and
the shared store against one deployment; the UI form fields come from the product
documentation only. **The card tools were read from source and not from a
deployment or a screen** - the tool names and the two failures they were built
from are the platform's own words, and nobody here has watched a card render.
