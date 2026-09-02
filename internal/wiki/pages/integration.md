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

**Unchecked:** the per-platform credentials come from the product documentation
only, and **no deployment uses a non-generic class** - every BotProvider across
every reference deployment is `generic`.
