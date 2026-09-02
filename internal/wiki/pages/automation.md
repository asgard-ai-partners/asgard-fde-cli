# Trigger and API

Two entry points that start a run without a user conversation.

| | Trigger | API |
|---|---|---|
| started by | a schedule | an external system calling in |
| CRs | `Trigger` + its entrypoint `Workflow` | `Workflow` + `Toolset` |
| generate with | `asgard-cli add trigger` | `asgard-cli add httptool` / `querytool` |

## Trigger

Starts a conversation with an agent on a schedule. The list shows Name, Schedule,
Active, Last run.

| field | required | |
|---|---|---|
| Name | yes | |
| Description | yes | what this schedule is for |
| Schedule | yes | a cron expression plus a Timezone (Asia/Taipei by default) |
| Model | | a built-in model tier, **not** one of the project's Agents |
| Prompt | yes | the agent's role, and what each run has to finish |

Two collapsible sections follow: Advanced Sandbox Settings and Advanced.

### Only cron is left

`TriggerClass` now has one value, `cron`. The google-sheet, google-drive,
google-mail, onedrive, onedrive-workbook and imap classes were removed - each of
them had the platform decide how to read a customer's data, which meant inventing
conventions the customer then could not hold to. With cron the shape of the data
is the customer's business again, and one firing can work through a whole batch.

Settings -> Connection still lists a "For Trigger" group (Google Drive, Google
Sheets, OneDrive, OneDrive Workbook). Those correspond to the removed classes.
**That group is stale.**

### The cron field takes less than crontab does

The CRD's pattern accepts `*`, one number, or `*/n` per field. Ranges and comma
lists are both rejected, so `0 9-18 * * *` and `0 9,13,17 * * *` do not apply -
business hours are every hour or nothing. `helm lint` does not check the pattern;
the rejection arrives during CD.

## API

Publishes a workflow as a callable HTTP API. The list shows Name, Group,
Description, Active, Last Modified. Creating one needs only a Name and a
Description.

### As a webhook, which is what a customer usually means

"When an order arrives, do X" is this, not a Trigger: an external system calls
in when something happens rather than us checking on a schedule. The shape is
three stages and each is a processor - see [`processors`](processors.md):

    the external system  ->  the endpoint
                               |
      1. validate-payload      the Payload Schema, a Secret Signature to prove
                               where it came from, and the allowed Content-Type
                               |
      2. whatever it does      query, model, http-request - the ordinary middle
                               |
      3. automation-tool-response  what the caller gets back

**Enable the Secret Signature.** Without it the endpoint runs whatever anyone who
finds the URL sends it, and a webhook URL travels: it is pasted into somebody's
CI, their vendor console, a ticket.

**The endpoint is the same URL a chat message goes to**, and this surprises
people:

    POST {{base_url}}/generic/ns/{{namespace}}/bot-provider/{{name}}/message/sse

So an inbound webhook and a person typing reach the platform the same way, and
`asgard-cli wiki api` describes the request and its SSE response for both. What
differs is what is on the other end of the Workflow, not the route in.

**A webhook and a schedule are not interchangeable** even though both start a run
with nobody watching. A webhook fires when their system decides; a Trigger fires
when we decide. If the customer cannot make their system call out, a schedule is
the fallback and it will always be later than the event.

## Before writing the chart

`asgard-cli usecase trigger` has the rules a scheduled run needs: the cursor and
the cold start, why not to write a BotProvider yourself, and why a schedule
cannot use a tool that asks for consent. Not repeated here.

## Sources

- The webhook shape, its three stages, the Secret Signature and that the
  endpoint is the same one:
  [Webhook integration](https://docs.asgard-ai.com/docs/developer-reference/examples/webhook-integration)
  - asgard-docs `f00e0ee`, read 2026-09-02

- [Trigger](https://docs.asgard-ai.com/docs/product-suite/odin/features/automation-trigger)
  - asgard-docs `f00e0ee`
- [API](https://docs.asgard-ai.com/docs/product-suite/odin/features/automation-api)
  - asgard-docs `f00e0ee`
- [Connection](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/connection)
  - asgard-docs `f00e0ee`
- Cron being the only class left, and the pattern's limits: checked 2026-09-02
  against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube)
  `15ded0f` - `TriggerClass`, `TriggerCronSpec`

**Unchecked:** `TriggerClass` and the cron pattern were held against the CRD and
one deployment; the UI fields come from the product documentation only.
