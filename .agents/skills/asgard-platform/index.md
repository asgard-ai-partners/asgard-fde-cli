# The Asgard platform: the map

Five kinds of document, and a pointer from one to another is a path you can
follow:

    wiki/       what the platform has, and which CR a UI name maps to
    usecase/    how ONE deployment shape is assembled, field by field
    needs/      what to get from the customer before a shape can be built
    brief/      what has actually been got wrong, before you do the thing
    guide/      which decision to make now, and what it costs to change later
    aliases.md  what a customer said -> what to search for

They answer different questions and it is worth knowing which you have.
"Can Asgard do X" is `wiki/`; "what goes in this field" is `usecase/`; "what do I
have to ask them for" is `needs/`; "where does this go wrong" is `brief/`; and
"which decision am I making" is `guide/`.

**A pointer is a path, relative to the document it is written in:**

    ../wiki/<name>.md     from a document inside one of the directories
    wiki/<name>.md        from `aliases.md` or this file, which are at the root

Written with `../` even between two documents in the same directory, so that a
pointer carries which kind it points at. Following one is opening a file, and
this lists everything a document points at:

    grep -o '\.\./[a-z]*/[a-z0-9-]*\.md' wiki/agents.md

**Read `aliases.md` first if the question did not arrive in English.** The corpus
is English and a customer conversation usually is not, so a term taken from what
somebody actually said matches nothing - and grep reports that identically to a
subject the material genuinely lacks.

## `wiki/` - what the platform has, and which CR a UI name maps to

Grouped by the question each answers in [`wiki/index.md`](wiki/index.md), which is
worth reading first. This is the flat list.

| document | covers |
|---|---|
| [`wiki/README.md`](wiki/README.md) | Asgard platform wiki |
| [`wiki/agents.md`](wiki/agents.md) | Flow Agent and Managed Agent |
| [`wiki/api.md`](wiki/api.md) | API, SSE and the SDK |
| [`wiki/automation.md`](wiki/automation.md) | Trigger and API |
| [`wiki/case-studies.md`](wiki/case-studies.md) | Case study: a stockout and a cross-store transfer |
| [`wiki/console.md`](wiki/console.md) | Management Console - permissions and Workspace |
| [`wiki/coverage.md`](wiki/coverage.md) | Which shapes each reference deployment actually uses |
| [`wiki/crd-rules.md`](wiki/crd-rules.md) | Rules the schema enforces, and one it cannot |
| [`wiki/fehu.md`](wiki/fehu.md) | Fehu - billing and usage |
| [`wiki/glossary.md`](wiki/glossary.md) | Words with one meaning here |
| [`wiki/index.md`](wiki/index.md) | Index |
| [`wiki/integration.md`](wiki/integration.md) | Reaching an agent from outside |
| [`wiki/knowledge.md`](wiki/knowledge.md) | Drive and Knowledge Base |
| [`wiki/mimir.md`](wiki/mimir.md) | Mimir - Data Insight |
| [`wiki/operations.md`](wiki/operations.md) | Connectivity, model capability, vocabulary |
| [`wiki/platform-unknowns.md`](wiki/platform-unknowns.md) | What the platform's documentation does not answer |
| [`wiki/processors.md`](wiki/processors.md) | The processors, and the fields that decide behaviour |
| [`wiki/product-suite.md`](wiki/product-suite.md) | Six products |
| [`wiki/screenshots.md`](wiki/screenshots.md) | Screenshots, and which situation each one is for |
| [`wiki/semantic-model.md`](wiki/semantic-model.md) | Semantic Model |
| [`wiki/settings.md`](wiki/settings.md) | Models, data sources and connections |
| [`wiki/setup-path.md`](wiki/setup-path.md) | From a credential to an agent someone can talk to |
| [`wiki/sindri.md`](wiki/sindri.md) | Sindri - Agent Hub |
| [`wiki/taiwan-channels.md`](wiki/taiwan-channels.md) | The commerce channels a customer will name, and what we have |
| [`wiki/tools.md`](wiki/tools.md) | MCP Server, Skillset and Plugin |
| [`wiki/what-they-read.md`](wiki/what-they-read.md) | What the customer read before meeting us |
| [`wiki/workflow.md`](wiki/workflow.md) | Workflow and Processor |

## `usecase/` - how one deployment shape is assembled, field by field

Grouped by the question each answers in [`usecase/README.md`](usecase/README.md), which is
worth reading first. This is the flat list.

| document | covers |
|---|---|
| [`usecase/README.md`](usecase/README.md) | The extracts |
| [`usecase/agent-hub.md`](usecase/agent-hub.md) | Agent hub |
| [`usecase/api-oauth.md`](usecase/api-oauth.md) | An API that needs a token first |
| [`usecase/browser-operation.md`](usecase/browser-operation.md) | Operating a system through its web UI |
| [`usecase/chat-channel.md`](usecase/chat-channel.md) | Reaching the agent from a chat platform |
| [`usecase/conventions.md`](usecase/conventions.md) | Conventions |
| [`usecase/demo-generation.md`](usecase/demo-generation.md) | A demo, and the pipeline that generates one |
| [`usecase/external-api.md`](usecase/external-api.md) | External HTTP APIs |
| [`usecase/fixed-query-tools.md`](usecase/fixed-query-tools.md) | Fixed query tools |
| [`usecase/flow-agent-single.md`](usecase/flow-agent-single.md) | Single-agent flow agent |
| [`usecase/flow-agent-supervisor.md`](usecase/flow-agent-supervisor.md) | Supervisor with subagents |
| [`usecase/knowledge-base.md`](usecase/knowledge-base.md) | KnowledgeBase, Loader and Source - the older knowledge path |
| [`usecase/knowledge-drive.md`](usecase/knowledge-drive.md) | Knowledge drive |
| [`usecase/mimir-dashboard.md`](usecase/mimir-dashboard.md) | A read surface for dashboards, with no agent on it |
| [`usecase/per-turn-credentials.md`](usecase/per-turn-credentials.md) | A credential the caller supplies, per turn |
| [`usecase/plugin.md`](usecase/plugin.md) | Plugin |
| [`usecase/semantic-layer.md`](usecase/semantic-layer.md) | SemanticLayer |
| [`usecase/skill-layers.md`](usecase/skill-layers.md) | A skill library at full size, and what the layers are for |
| [`usecase/skill-set.md`](usecase/skill-set.md) | SkillSet, SourceSet, Syncer |
| [`usecase/trigger.md`](usecase/trigger.md) | Trigger |
| [`usecase/workflow-chain.md`](usecase/workflow-chain.md) | Chaining processors, and the graph that wires them |
| [`usecase/write-path.md`](usecase/write-path.md) | Write paths and the approval gate |

## `needs/` - what to get from the customer before a shape can be built

| document | covers |
|---|---|
| [`needs/browser-operation.md`](needs/browser-operation.md) | browser-operation: what to get from the customer |
| [`needs/chat-channel.md`](needs/chat-channel.md) | chat-channel: what to get from the customer |
| [`needs/external-api.md`](needs/external-api.md) | external-api: what to get from the customer |
| [`needs/knowledge-drive.md`](needs/knowledge-drive.md) | knowledge-drive: what to get from the customer |
| [`needs/semantic-layer.md`](needs/semantic-layer.md) | semantic-layer: what to get from the customer |
| [`needs/skill-set.md`](needs/skill-set.md) | skill-set: what to get from the customer |
| [`needs/write-path.md`](needs/write-path.md) | write-path: what to get from the customer |

## `brief/` - what has actually been got wrong, before you do the thing

| document | covers |
|---|---|
| [`brief/customer-meeting.md`](brief/customer-meeting.md) | Before customer-meeting |
| [`brief/connect.md`](brief/connect.md) | Before connect |
| [`brief/write-chart.md`](brief/write-chart.md) | Before write-chart |
| [`brief/handover.md`](brief/handover.md) | Before handover |

## `guide/` - which decision to make now, and what it costs to change later

| document | covers |
|---|---|
| [`guide/projects.md`](guide/projects.md) | Decide how the work splits into projects |
| [`guide/data-sources.md`](guide/data-sources.md) | Wire up the customer's databases |
| [`guide/read-path.md`](guide/read-path.md) | Decide each project's read path |
| [`guide/entry-point.md`](guide/entry-point.md) | Decide each project's entry point |
| [`guide/knowledge.md`](guide/knowledge.md) | Decide where unstructured knowledge lives |
| [`guide/verify.md`](guide/verify.md) | Run the acceptance gate |
| [`guide/deploy.md`](guide/deploy.md) | Deploy |
| [`guide/enhance.md`](guide/enhance.md) | Add a capability to a repo that is already live |
| [`guide/requirements.md`](guide/requirements.md) | Turn what the customer said into a request |
| [`guide/idle.md`](guide/idle.md) | Nothing in flight |

## What is not here

Everything the material points at is a path you can follow. There is one
exception, and it needs the `asgard-cli` binary:

| pointer | why it is not a file |
|---|---|
| `asgard-cli guide <name>` | **half of it is here.** A guide renders this repository's own state into its guidance - which projects exist, what is still open - and that half cannot be a file, because a file would freeze one moment of it. The decisions are in `guide/`; run the command for where this repository actually stands |

The commands that answer that half directly, when it is all you want:

    asgard-cli project     what each chart declares, and still lacks
    asgard-cli question    what nobody has answered yet
    asgard-cli request     what the customer asked for
    asgard-cli task        the open task specs

## Staleness

These files came from one binary. A newer one may carry different pages, and
nothing in here can tell you which:

    asgard-cli init

It reports what is here against what the running binary carries, and replaces
this directory outright when the version has moved. Read it as the authority
over these files rather than anything written inside them.
