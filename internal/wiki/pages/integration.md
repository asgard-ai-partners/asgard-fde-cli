# Reaching an agent from outside

Four ways.

| route | CR | when |
|---|---|---|
| a chat platform (LINE / Slack / Discord / Telegram) | `BotProvider`, that class | the customer already lives there |
| your own front end or API | `BotProvider` class `generic` | the customer's own site or app |
| the SDK (JS / React) | as above, plus the packages | embedding a ready-made chat UI |
| an MCP Server | `Toolset` | letting another agent call it |

## Chat platforms

Each platform needs different credentials, all supplied by the customer:

| platform | needs | extra step |
|---|---|---|
| LINE | Channel Secret, Channel Access Token | Asgard produces a Webhook URL that has to be pasted back into the LINE Developers Console, verified, and Use webhook enabled |
| Slack | Client ID, Client Secret, Signing Secret, Permission Scopes | a Slack app has to exist and subscribe to Bot events |
| Discord | Bot Token | an app and bot in the Developer Portal, authorised and invited to the server |
| Telegram | Bot Token | created through BotFather |

LINE is the only one needing **two-way** setup - Asgard gives a URL that has to go
back into LINE. The rest only take credentials inward. For Taiwanese customers it
is usually the first one asked about.

`botProviderClass` is immutable in the CRD, so it has to be right the first time.

## What the platform does NOT own: the conversation

This is the question every customer service engagement asks, and the answer is
the same one every time.

**Handing over to a human, pausing the agent while a person replies, resuming
afterwards, counting how many questions one user has asked - none of these are
platform features.** There is no CR for any of them, and no field: searching the
CRDs for a handoff, a takeover, a suspend or a per-user quota finds nothing.

What the platform quota does cover is capacity, not people. All eight numbers,
and **they apply to the Workspace - projects inside it share them**:

| per request | per workspace |
|---|---|
| 5 RPS per endpoint | 40 Projects |
| 3 minutes | 300 GB of Knowledge Base |
| 30 steps | 500 Processors |
| | **10 Loaders** |
| | 150 Indexers |

**Ten Loaders is the one that bites first.** A Loader is one recurring pull, so a
customer with a dozen document sources exceeds it before anything else on this
list - see [`knowledge.md`](knowledge.md).

**These are defaults, not ceilings, and they are per plan.** They are raised by
contacting sales or writing to service@asgard-ai.com, which is a different
sentence in a meeting than "that is the limit". The overview says a Workspace
has a price plan and that how many Projects it may hold depends on it - so 40 is
one plan's number rather than the platform's. What is not documented is which
plan gives what. See [`what-they-read.md`](what-they-read.md).

A multi-system troubleshooting conversation can reach 30 steps,
which is worth saying out loud before somebody designs one.

The mechanism the platform's own case study describes puts the conversation
somewhere else entirely. A retail site's support desk receives the customer's
message, writes it into its own conversation log and answers the customer
immediately; only then does it forward the message to the Flow Agent in the
background, with a scope-limited credential. A human can join that same thread at
any time, because the thread was never the agent's to begin with.

So the shape is:

    the customer     ->  something that owns the conversation  ->  Asgard
                         (a support desk, a site, a relay)

"Pause the AI" is that middle layer deciding not to forward. "Three strikes then
a human" is that middle layer counting. "Ten questions a day" is that middle
layer counting too.

**With a website the middle layer is obvious - the site itself. With LINE it is
not, and that is the question to ask.**

What is settled is the part above: none of it is ours. What is *not* settled is
how much LINE gives you for free, and the difference decides whether the customer
needs a support desk or a small piece of state.

- A person replying in LINE Official Account Manager is a takeover surface LINE
  may already provide. Older LINE accounts had a 回應模式 that was either chat or
  bot; whether a current account can run both at once was **not confirmed** -
  LINE's own Messaging API page on building a bot says nothing either way, and
  the chat-handling page could not be read.
- Even if it does, LINE gives the bot **no signal that a human took over**. So
  the pause/resume state and the counters still have to live somewhere, and that
  somewhere is still not the platform.

So do not tell a customer this cannot be built. Ask who owns the LINE Official
Account, get them to say what their agents use today, and check LINE's current
documentation for the account they actually have. The answer moves the
requirement between "needs a support desk in front" and "needs a small piece of
state" - a very different conversation, and not one to have from memory.

## Two pages in Odin

- **Applications -> Data Insight & Agent Hub** - browse and open the Mimir and
  Sindri applications published to this Workspace, filtered by All, Data Insight
  or Agent Hub
- **Applications -> Customized Integration** - manage integrations connecting an
  agent to outside channels and applications, filtered by All, Bot, API or MCP
  Servers

## API and SSE

```
POST {base_url}/generic/ns/{namespace}/bot-provider/{bot_provider_name}/message/sse
```

Authenticated with an `X-API-KEY` header. The key comes from the project's
Integration -> App settings page.

The response is Server-Sent Events; the connection stays open and carries agent
messages, system events and end-user messages.

The endpoint, the event sequence, the four integration patterns and the SDK are
in [`api.md`](api.md).

## SDK

```bash
npm install @asgard-js/core     # framework agnostic, Node.js or browser
npm install @asgard-js/react    # ready-made chat UI components and hooks
```

`@asgard-js/react` takes `@asgard-js/core` as a peer dependency; install both.

The SDK wraps REST requests and authentication, handles the SSE stream, and
persists a session for multi-turn conversation.

## Platform architecture

```
client (web / SDK / REST API / chat platform)
  -> API Gateway (X-API-KEY auth, SSE streaming)
    -> Asgard Core Engine (Workflow Engine, Processor Manager, Channel Manager)
      -> Processors
        -> AI resources (LLM providers, Knowledge Base, Data Source, MCP Server)
```

## Before writing the chart

`asgard-cli usecase chat-channel` has each class's credential block and what a
channel costs - which credentials infra has to provide, and whether the class
needs a connector pod.

## What the platform cannot send

**There is no mail capability anywhere in the platform** - no SMTP, no preset
mail Toolset, nothing in the core. "Email me when it happens" is one of the most
common things a customer asks for, and the answer is not "yes, of course".

    they have an HTTP endpoint that sends mail   an external-api call
    they have no endpoint                        it cannot be built yet

**A deployment that mocks it owes a disclosure**, and this is worth copying. One
does: `wf-send-mail` is a single `push-message` returning
`{ok: true, mocked: true, to, subject, body}`, so the whole pipeline runs and
the drafted mail lands in the invocation record for review. **`ok: true` is
deliberate** - a false would stop a Trigger's cursor and the path would never be
exercised.

Which makes disclosure the entire safety property. That deployment requires both
the tool's `tooling.description` and the agent prompt to lead every summary with
"MOCK - not actually sent", and to never say "notified".

**A log that reads as though people were emailed is the real damage a mock can
do.** The same applies to any mocked outward action - a ticket not created, an
order not placed.

## Sources

- [LINE](https://docs.asgard-ai.com/docs/integration/line),
  [Slack](https://docs.asgard-ai.com/docs/integration/slack),
  [Telegram](https://docs.asgard-ai.com/docs/integration/telegram),
  [SDK](https://docs.asgard-ai.com/docs/integration/sdk)
  - asgard-docs `f00e0ee`. `integration/Discord.mdx` is an empty file; the
  Discord content is under `integration-with-asgard/`
- The four pages under `integration-with-asgard/` (api, line, slack, discord)
  - asgard-docs `f00e0ee`. All four are marked `draft`, and they describe an
  interface called "Published -> add integrated" inside a Project, which does not
  match Odin's current Applications -> Customized Integration. **Possibly stale**
- [Authentication](https://docs.asgard-ai.com/docs/developer-reference/authentication)
  and [architecture](https://docs.asgard-ai.com/docs/developer-reference/architecture)
  - asgard-docs `f00e0ee`
- [Applications overview](https://docs.asgard-ai.com/docs/product-suite/odin/features/applications-overview)
  and [Customized Integration](https://docs.asgard-ai.com/docs/product-suite/odin/features/applications-customized-integration)
  - asgard-docs `f00e0ee`
- `botProviderClass` being immutable: checked against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` -
  `BotProviderSpec`
- **The platform owning no handoff, takeover, suspend or per-user quota**:
  checked 2026-09-02 against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` -
  no CRD and no field carries any of those concepts
- The quota numbers, all eight, that they are Workspace-level and shared, and
  that they are raised through sales:
  [Quota and limits](https://docs.asgard-ai.com/docs/help-community/quota-limits)
  - asgard-docs `f00e0ee`, read in full 2026-09-02
- The support desk owning the conversation:
  [AI customer service answering order enquiries](https://docs.asgard-ai.com/docs/product-suite/odin/case-studies/retail-ai-customer-service)
  - asgard-docs `f00e0ee`
- LINE's own behaviour: [Building a bot](https://developers.line.biz/en/docs/messaging-api/building-bot/)
  read 2026-09-02, which does not mention response modes or any exclusivity;
  `messaging-api/handling-chats/` returned 403 and was not read. asgard-docs
  `f00e0ee` covers only the webhook setup steps and says nothing about a human
  replying in the same account

**Unchecked:** the per-platform credentials come from the product documentation
only, and **no deployment uses a non-generic class** - every BotProvider across
every reference deployment is `generic`. That the support desk owns the
conversation is read from one case study and has not been held against a
deployment; **whether a current LINE Official Account can serve both a human
in Official Account Manager and a webhook at the same time is unresolved** - it
is LINE's behaviour rather than Asgard's, no source here settles it, and it
decides how much of a customer's handoff requirement is buildable. Get it from
LINE's documentation for the account in question before designing around it.
