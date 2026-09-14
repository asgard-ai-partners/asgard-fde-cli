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

**Two shapes of this URL are in circulation and they differ by one segment.**
The API reference gives the path above, with `/generic/`. The SDK overview and a
production tenant's own chart README both give it **without**:

    https://api.asgard-ai.com/ns/<namespace>/bot-provider/<name>/message/sse

The second is not a documentation slip - it is what a live relay sends to, and
the SDK derives its SSE URL by appending `/message/sse` to whatever
`botProviderEndpoint` you hand it, adding nothing. Which of the two the platform
actually routes has not been established here, and it is a 404 either way if you
guess wrong in front of a front-end team.

**So do not hand anyone a URL from this page. Get it from the deployment**: the
BotProvider's name and its namespace are in the chart, and the tenant charts
that already have a front end put the exact working URL in
`projects/<project>/chart/README.md`. Confirm it with one request before it goes
in a document. `../wiki/platform-unknowns.md` P9 tracks the question.

**The header is `X-API-KEY`** - that part is consistent across the API
reference, the authentication page and the platform's own architecture. If a
customer's existing relay uses something else, that is the relay's own
convention and not the platform's; one deployment authenticates its relay with a
webhook-token header of its own. Do not generalise from a customer's front end
to what the platform expects.

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

**`RESET_CHANNEL` is not "open the chat" - it is "throw the history away".**
The two actions above are the whole of what this page used to offer, so the
obvious front end sends `RESET_CHANNEL` when the widget mounts, and every page
reload then destroys the conversation the customer was having. That is a bug
nobody reports as one: it looks like the agent forgetting.

Rejoining an existing channel is a different path. The SDK's `Channel.restore`
asks `GET /channel/metadata` whether the channel exists and, if it does,
replays the server's transcript rather than resetting - a cold-start rejoin is a
`GET .../message/sse` carrying the channel id and no message. `Channel.reset` is
for a conversation the user has deliberately started over.

So the question to ask a front-end team is not "how do you open a channel" but
**"what happens on reload"** - and if they are not using the SDK, they need the
rejoin call as well as the send.

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
[`processors.md`](../wiki/processors.md). **Every one of those variables can be absent**,
so each access is written defensively or throws on the turn somebody sends no
file.

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

### What the client does besides opening a channel

Worth knowing before a front-end team asks, because each of these is otherwise
a REST endpoint somebody has to find. All of them derive their endpoint from
`botProviderEndpoint`, and without it the error names which derivation failed:

    uploadFile                  a file, returning blob info to carry into a message
    downloadChannelHomeFile     one of that channel's own files
    channelMetadata             whether a channel exists; null when it does not
    suspendChannel              stop the run in flight on that channel
    deleteChannel               remove it
    sendMessageFeedback         the user's thumb up or down on one reply

**`detach` and `close` are not the same.** `close` shuts the client down;
`detach` stops accepting new requests and waits for the ones in flight, which
is the one to call as a page unloads.

**A feedback comment is capped in bytes, not characters.** `@asgard-js/core`
exports `FEEDBACK_COMMENT_MAX_BYTES` and `feedbackCommentByteLength` for that
reason - a Chinese comment reaches the cap in a third of the characters an
English one does.

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

**This is not the credential a CR uses.** `SourceSet.apiKey`, `Toolset.apiKey`
and `BotProvider.adminApiKey` take a platform resource credential, which the
platform mints per namespace and no console page issues -
`../usecase/conventions.md` has where that one comes from. The two are read as
one easily, and the only sentences saying where a key comes from used to be
these, about this one.

**The console route above is the documentation's and is unconfirmed.** An
engagement holding the account could not find "Integration", "App settings" or
anything issuing a key, on 2026-09-14. Either the page moved or it is named
something else now; `../wiki/platform-unknowns.md` P13 carries it.

## Before writing the chart

`../usecase/external-api.md` covers an HTTP tool, `../usecase/api-oauth.md` the two-call
token chain, and `../usecase/workflow-chain.md` what passes between processors.

## Sources

- Every file under
  [send-message](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/introduction)
  - asgard-docs `f00e0ee`
- [Integration guide](https://docs.asgard-ai.com/docs/developer-reference/integration-guide)
  - asgard-docs `f00e0ee`
- [SDK overview](https://docs.asgard-ai.com/docs/developer-reference/sdk/overview),
  the JavaScript and React guides, and `asgard-sdk`
  - asgard-docs `f00e0ee`
- [Migration guide](https://docs.asgard-ai.com/docs/developer-reference/migration)
  - asgard-docs `f00e0ee`
- The examples under
  [examples](https://docs.asgard-ai.com/docs/developer-reference/examples)
  - asgard-docs `f00e0ee`
- [Authentication](https://docs.asgard-ai.com/docs/developer-reference/authentication)
  - asgard-docs `f00e0ee`

- The two-call file path and `blobIds`:
  [upload file](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/upload-file-api)
  and [append file and send](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/append-file-and-send-message-api)
  - asgard-docs `f00e0ee`, read 2026-09-02
- The event envelope, the `fact` union and `runError.location`: the pages
  under `developer-reference/api-doc/send-message/sse-response/` - one page per
  event, starting at
  [run-init](https://docs.asgard-ai.com/docs/developer-reference/api-doc/send-message/sse-response/run-init).
  **The directory itself has no landing page and 404s**, so cite the pages
  rather than the directory. Read 2026-09-02

- The client's channel and file methods, `detach` against `close`, and the
  byte-capped feedback comment: asgard-docs `23409b3` `docs/developer-reference/sdk/javascript.mdx`,
  read 2026-09-11. **Not held against a front end** - no reference deployment
  uses the SDK

**Unchecked:** everything here comes from the product documentation; no actual
integration was examined.
