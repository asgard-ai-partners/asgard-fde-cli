# Screenshots, and which situation each one is for

Every screenshot of this platform lives in the product documentation. There are
none in this binary and none in a customer repository, and that is deliberate:
an image copied into an engagement goes stale where nobody is looking, and 1.7MB
of PNGs teaches nobody anything. What is worth carrying is **which picture
answers which question**, which is this page.

Both places serve the same file:

    https://docs.asgard-ai.com/img/docs/<path>          the published site
    <clone of asgard-docs>/static/img/docs/<path>       a checkout

The paths below are that `<path>`. The captions are the documentation's own alt
text, in zh-TW; **the console itself is in English**, so a deck in either
language can use them and write its own caption.

**Open one before it goes on a slide.** Nothing records when any of these was
captured, so a form that has since changed looks exactly like a current one - and
a stale screen in front of the customer who uses that screen daily costs more
than having no picture.

Three things to look for, and the third is the one that catches people:

  - **a credential** - a hostname, an account, a connection string
  - **another customer** - a name, a logo, anything recognisable
  - **our own implementation nouns, inside the picture.** A CR name, a `ts-` /
    `sl-` / `dc-` / `ss-` prefix, an internal tool name, the admin console's own
    navigation. The `proposal-deck` skill bans these in the deck's text and the
    ban applies just as much to a screenshot of them - **and several images below
    contain them.** They are marked, with what to crop

Cropping is normal and usually costs nothing: the part that carries the argument
is rarely the part carrying the noun.

## For a proposal: what it looks like to use

**This is the set most proposals need and most decks miss**, because the obvious
instinct is to show the console - which is our side, not theirs. These four are
one continuous story from a user's question to a completed action, and they carry
the governance gate, which is the hardest thing to explain in words.

| path | what it shows |
|---|---|
| `sindri-retail-stockout-transfer/00-available-agents.png` | the hub with five agents, before anything is asked |
| `sindri-retail-stockout-transfer/01-allocation-answer.png` | a real answer: current state, available quantity, three options to choose from |
| `sindri-retail-stockout-transfer/02-approval-gate.png` | **the approval dialog** - the tool it wants to call, and nothing running until a person allows it. **CROP FIRST:** the dialog names `工具集:ts-wms` and `工具:create_transfer_order`. Keep the title line and the three buttons, drop the two rows between them - the argument is entirely in "it stopped and asked", which survives the crop |
| `sindri-retail-stockout-transfer/03-execution-result.png` | after approval: what was created, and the resulting stock forecast |

The third one is worth a slide on its own whenever the requirement has a write in
it. `asgard-cli usecase write-path` is the argument; this is the picture of it.

For a customer-service shape rather than an internal one:

| path | what it shows |
|---|---|
| `retail-ai-customer-service/21-flow-agent-canvas.png` | the four-node flow: Entry, Init, Agent, Listen, with an error branch |
| `retail-ai-customer-service/23-agent-node-open.png` | the prompt deciding what a **logged-out** visitor is told, versus a logged-in one |

And the one that shows Description doing its job in public:

| path | what it shows |
|---|---|
| `sindri-home/home-available-agents.png` | each card's line is that agent's Description, as routing text |

## For a handover: the setup path

The order these belong in is [`setup-path.md`](setup-path.md). Use them when the
audience is the people who will operate the thing - never in a proposal.

| step | path | what it shows |
|---|---|---|
| 1 | `settings/data-source/data-source-create-provider-open.png` | the Provider list open: nine, all databases. **The screen that settles where an API key does not go** |
| 1 | `settings/data-source/data-source-create.png` | the form, with Test Connection before Save |
| 1 | `settings/connection/connection-create-options-open.png` | Connection types, grouped by syncer / For Loader / For Trigger |
| 2 | `mcp-servers/new-menu.png` | From Workflow against From Existing - the fork that decides how much work step 2 is |
| 2 | `mcp-servers/mcp-from-existing.png` | Transport Type, Command, Environment Variables - where an API key actually lives |
| 2 | `skillsets/new-menu.png` | From Scratch against From Git |
| 2 | `drive/drive-syncer-wizard-step1.png` | the five syncer sources: Google Drive, OneDrive, Git, Web Crawler, Data Source |
| 2 | `drive/drive-detail-context-index.png` | Context Index, and the switch that turns it on |
| 2 | `data-insight-semantic-model/table-setting-select.png` | picking tables, with live Data Preview |
| 2 | `data-insight-semantic-model/modeling-built.png` | ten tables modelled, one expanded to Description / Primary Key / Columns |
| 3 | `agent-hub-managed-agent/list.png` | the five built-in templates - show this when they think it starts from nothing |
| 3 | `agent-hub-managed-agent/create.png` | **the one screen most handovers need**: Alias, Description, and Prompt as Persona / Task / Context / Format, with the live preview |
| 3 | `agent-hub-managed-agent/resources-section.png` | mounting what step 2 produced |
| 5 | `agent-hub-flow-agent/create.png` | Name, Description, Advanced Sandbox Settings |

A worked example of step 3 filled in, rather than empty:

| path | what it shows |
|---|---|
| `retail-stockout-transfer/12-allocation-top.png` | a real agent's Alias, Description and four Prompt sections |
| `retail-stockout-transfer/13-allocation-resources.png` | its sample questions and everything it mounts |
| `retail-stockout-transfer/01-managed-agents.png` | five agents, all `Released` |

## For explaining a capability

| subject | path |
|---|---|
| the governance gate, as a list | `retail-stockout-transfer/04-automation-tools.png` - seven approval-gated tools |
| what a gated tool declares | `retail-stockout-transfer/11-transfer-order-tool.png` - its input schema |
| how many systems one agent reads | `retail-stockout-transfer/03-semantic-models.png` - CRM, ERP, e-commerce, POS, supplier, OLAP, WMS. **CROP FIRST:** the whole Odin console navigation is down the left - Agent Hub, MCP Servers, Skillsets, Plugins, Settings. Keep the card area only |
| conversational analysis | `mimir-thread/answer.png`, `mimir-thread/chart-result.png` |
| a dashboard that gets shared | `mimir-dashboard/share-dashboard.png` |
| teaching the model the business's own questions | `mimir-knowledge/add-question-sql-pair.png` |
| a Knowledge Base loading documents | `knowledge-base-knowledge/auto-load-form.png` |
| scheduled runs | `automation-trigger/new-trigger.png` - Schedule, Timezone, Model, Prompt |

## For a cost or governance conversation

| subject | path |
|---|---|
| what a bill looks like | `fehu/fehu-billing-reports-list.png`, `fehu/fehu-bill-detail-service.png` |
| usage against quota | `fehu/fehu-quota-limits-top.png`, `fehu/fehu-usage-reports.png` |
| who can reach which product | `console-product-permission/agent-hub-accounts.png` |
| workspace membership | `console-workspace-settings/workspace-accounts.png` |

**These four are the ones to be careful with.** Permissions and billing are where
a customer most easily reads a screenshot as a promise about what can be
restricted. `asgard-cli wiki platform-unknowns` still has scope control as
unanswered: the Console decides who reaches a product, and what a caller can
touch **after** that is the part no source settles. A permissions screenshot on a
slide about limiting access is the overstatement this tool warns about most.

## What has no screenshot

Not everything does, and assuming otherwise wastes a search:

  - the chart, the CRs, the cluster - there is no UI for any of it
  - **LINE: images exist and they are of a product that no longer looks like
    that.** `integration/LINE` carries two, hash-named rather than in a topic
    directory, which is why a survey of the topic folders misses them:

        /img/docs/60e2492a66bf.png    the prerequisites
        /img/docs/f3563b5553dc.png    the integration dialog, LINE selected,
                                      Channel Secret and Access Token fields

    The second one is the screen a customer would want to see. **It is the old
    console** - a left nav of Overview / Workflows / Knowledge / Environment /
    Apps, and a card dated 2024/09/20 - and today's Odin has none of those.
    `integration.md` already flags those four pages as possibly stale for the
    same reason.

    So the answer is still not to use them, but for a checkable reason rather
    than because none exist: **a screenshot of a console the customer will not
    recognise is worse than no screenshot**, and this is the exact failure this
    page warns about, in the one place somebody would most want to ignore it.

    What to do instead, and it is better than it sounds: describe the exchange
    in their words - "your customer types their order number in LINE, it replies
    with the repair status, and says so plainly when it cannot find it". Slide 5
    is an interaction, not a screen. A vendor's own documentation screenshot is
    not an option either: it is someone else's product surface in our proposal.

    **If somebody captures a current one, this is the highest-value screenshot
    missing** - it is the channel a Taiwanese customer asks about first.

  - anything about handoff, pausing, or per-user counters, because the platform
    does not have them - see [`integration.md`](integration.md)

## Sources

- Every path here was read off the `![alt](/img/docs/...)` references in
  asgard-docs' own `.mdx` pages, so a caption is the documentation's rather than
  a guess about what the file contains
  - asgard-docs `f00e0ee`
- **Checked** 2026-09-02, opened rather than listed:
  `settings/data-source/data-source-create-provider-open.png`,
  `agent-hub-managed-agent/create.png`, `sindri-home/home-available-agents.png`,
  and the two marked CROP FIRST above - which is how the crops came to be
  marked. An engagement found the first of them by downloading it, after this
  page had recommended it as the single most useful image here

**Unchecked:** everything else here is described by its alt text, not by having
been looked at, and **no capture date exists for any of them**. Five of roughly a
hundred have been opened, and two of those five needed cropping - so assume an
unopened one does too, rather than that the marked ones are the only ones. The grouping into
situations is this tool's judgement rather than anything the documentation says.
