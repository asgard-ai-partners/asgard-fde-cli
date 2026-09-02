# API, SSE and the SDK

The whole path for reaching an agent from your own front end. Maps to a
`BotProvider` of class `generic`.

## The endpoint

Starting a channel, sending a message, uploading a file and sending a file all go
to the same endpoint:

```
POST {base_url}/generic/ns/{namespace}/bot-provider/{bot_provider_name}/message/sse
Header: X-API-KEY
```

They differ by the `action` field:

| action | `action` | `text` |
|---|---|---|
| start or reset the channel | `RESET_CHANNEL` | empty |
| send a message | `NONE` | the message |

Request parameters:

| field | |
|---|---|
| `customChannelId` | required, and yours to choose. **Conversation memory is keyed on it** |
| `customMessageId` | optional, for tracing and debugging |
| `text` | the message |
| `action` | see above |

`customChannelId` is the only thing tying a conversation together. A front end
that generates a fresh id per message makes every message a new conversation.

## SSE events

The response is a Server-Sent Events stream and the connection stays open.

| event | when |
|---|---|
| `asgard.run.init` | the service initialises |
| `asgard.message.start` | a message begins, possibly with initial text |
| `asgard.message.delta` | content arrives progressively, possibly many times; `idx` gives the order |
| `asgard.message.complete` | the full text plus its Template form |
| `asgard.run.done` | the run ends |
| `asgard.run.error` | only on failure |
| `asgard.process.start` / `.complete` | a process begins and ends |
| `asgard.tool_call.start` / `.complete` | **only when a Toolset is in use** |

The ordinary sequence is `run.init`, `message.start`, one or more
`message.delta`, `message.complete`, then either back to `message.start` or on to
`run.done`.

## Sending a file is two calls

A file does not travel with the message. Upload it, keep the id, send the
message referencing it:

    POST .../bot-provider/<name>/blob            multipart/form-data
      customChannelId, file                      -> a blobId

    POST .../bot-provider/<name>/message/sse     the ordinary send
      customChannelId, text, blobIds: [ ... ]    an array, several files

**`customChannelId` has to match across both**, as it does for everything else -
it is what ties a conversation together, and an upload that used a different one
is attached to a conversation nobody is having.

Inside the workflow the files arrive as `prevBlobs`, an array of Blob with
`blobId`, `fileType`, `fileName`, `size` and `mime` - see
[`processors.md`](processors.md), including that ECMA5 has no optional chaining,
so every access to it is written defensively or throws on the turn somebody sends
no file.

## What an event actually looks like

Every event carries the same envelope, and the payload is a **tagged union**:

```json
{ "eventType": "asgard.message.delta",
  "requestId": "...", "eventId": "...",
  "namespace": "proj-...", "botProviderName": "bp-...",
  "customChannelId": "...",
  "fact": { "runInit": null, "runDone": null, "runError": null,
            "messageStart": null,
            "messageDelta": { "message": { "text": "...", "idx": 3, ... } },
            "messageComplete": null } }
```

**`fact` has one key per event type and only the one matching `eventType` is
populated; the rest are `null`.** A front end reads `fact.<name>` rather than
inspecting the shape - which matters because the keys present in `fact` vary
between events, so pattern-matching on shape breaks the first time the platform
adds one.

`requestId` groups every event of one run and is what to quote in a bug report.
`eventId` orders them; on a delta, `idx` orders within the message.

## The error event names the processor that failed

`asgard.run.error`'s `fact.runError.error` is the most useful thing in this
stream and nothing else here mentioned it:

    message   human-readable
    code      INVALID_ARGUMENT and friends
    inner     the underlying error, often empty
    location  namespace, workflowName, processorName,
              processorType, processorConfigName, processId

**`location` is which node of which Workflow failed.** For a chart of any size
that is the difference between reading a stream and reading a diagram. The fields
are empty when the failure happens before a processor runs - the example in the
documentation is an empty message rejected at the channel - and an empty
`workflowName` is therefore itself information: it did not reach the flow.

Surface it. A front end that shows only `message` throws away the location, and
whoever debugs it later has to reproduce the failure to get it back.

## Four integration patterns

The documentation groups integrations into four patterns, and the choice governs
how auth, history and memory are handled:

- **Hosted Embed**
- **Direct Connect**
- **Workflow Auth**
- **Backend Relay**

The details are on each pattern's own page. The choice decides where the API key
lives - connecting from the browser means the key is in the browser, and Backend
Relay exists to avoid that.

## SDK

```bash
npm install @asgard-js/core     # framework agnostic
npm install @asgard-js/react    # chat UI components and hooks
```

`@asgard-js/react` takes `@asgard-js/core` as a peer dependency.

The SDK wraps the REST requests and authentication, handles the SSE stream, and
persists a session so multi-turn conversation works.

### Version migration

The breaking change from 0.1.x to 0.2.x is **new SSE events** -
`asgard.process.start/complete` and `asgard.tool_call.start/complete` - which
need handlers adding. The `<Chatbot>` component's `config` should be checked too.

## From nothing to live

1. Pick an integration pattern
2. Build the Bot and Workflow in Odin
3. Configure a Completion Model and the provider's API key
4. Publish as a Generic Integration and collect the `botProviderEndpoint` and API
   key
5. Manage the API key - issuing, rotating, rate limits
6. Connect through the SDK
7. Understand the SSE sequence
8. Design how replies render; message templates are rendered by the front end

The API key comes from the project's Integration -> App settings page. **Keep it
out of front-end code, version control and anywhere public.**

## Before writing the chart

`asgard-cli usecase external-api` covers an HTTP tool, `api-oauth` the two-call
token chain, and `workflow-chain` what passes between processors.

## Sources

- All 17 files under
  [send-message](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/introduction)
  - asgard-docs `f00e0ee`
- [Integration guide](https://docs.asgard-ai.com/docs/developer-reference/integration-guide)
  - asgard-docs `f00e0ee`
- [SDK overview](https://docs.asgard-ai.com/docs/developer-reference/sdk/overview),
  the JavaScript and React guides, and `asgard-sdk`
  - asgard-docs `f00e0ee`
- [Migration guide](https://docs.asgard-ai.com/docs/developer-reference/migration)
  - asgard-docs `f00e0ee`
- The four examples under
  [examples](https://docs.asgard-ai.com/docs/developer-reference/examples)
  - asgard-docs `f00e0ee`
- [Authentication](https://docs.asgard-ai.com/docs/developer-reference/authentication)
  - asgard-docs `f00e0ee`

- The two-call file path and `blobIds`:
  [upload file](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/upload-file-api)
  and [append file and send](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/append-file-and-send-message-api)
  - asgard-docs `f00e0ee`, read 2026-09-02
- The event envelope, the `fact` union and `runError.location`: the eleven pages
  under [send-message/sse-response](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/sse-response),
  read 2026-09-02. They had not been read into this material before then - this
  page had the event list and not the payloads

**Unchecked:** everything here comes from the product documentation; no actual
integration was examined.
