---
group: While building
description: Trigger and API, and why only cron is left
---
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

### The cron field is an ordinary crontab, and the CRD checks nothing

`Trigger.spec.cron.schedule` is a five-field cron expression or an
`@descriptor`, copied verbatim into the derived CronJob, so the grammar is
whatever that field accepts - ranges, comma lists and descriptors included.

**The CRD carries no pattern on it, deliberately.** One used to be there and
had copied cron wrong in both directions: it rejected `0 8,13 * * *`, ranges,
step-on-range, month and day names and every `@descriptor`, all of which the
API server accepts, while admitting `*/0 * * * *`, which it does not. What
checks the expression now is a parse, run by the admission webhook on write and
again at the public API boundary, so a malformed one comes back with a message
rather than deploying and never firing. `helm lint` still looks at none of it.

**This is the expensive direction a pinned copy goes stale in.** A constraint
that is deleted upstream leaves a warning against exactly what the platform now
accepts, and three places here taught the old one - the same trap is on the
SourceSet syncer's schedule, which lost the same regex in the same change.

## API

Publishes a workflow as a callable HTTP API. The list shows Name, Group,
Description, Active, Last Modified. Creating one needs only a Name and a
Description.

### As a webhook, which is what a customer usually means

"When an order arrives, do X" is this, not a Trigger: an external system calls
in when something happens rather than us checking on a schedule. The shape is
three stages and each is a processor - see [`processors`](../wiki/processors.md):

    the external system  ->  the endpoint
                               |
      1. validate-payload      the Payload Schema, a Secret Signature to prove
                               where it came from, and the allowed Content-Type
                               |
      2. whatever it does      query, model, http-request - the ordinary middle
                               |
      3. push-message          what the caller gets back. The documentation
                               calls this page Response, at
                               `processor/automation-tool-response`; there is
                               no `response` type, and a chart writes
                               `push-message` scoped to `automation_tool`

**Enable the Secret Signature.** Without it the endpoint runs whatever anyone who
finds the URL sends it, and a webhook URL travels: it is pasted into somebody's
CI, their vendor console, a ticket.

**The endpoint is the same URL a chat message goes to**, and this surprises
people:

    POST {{base_url}}/generic/ns/{{namespace}}/bot-provider/{{name}}/message/sse

So an inbound webhook and a person typing reach the platform the same way, and
`../wiki/api.md` describes the request and its SSE response for both. What
differs is what is on the other end of the Workflow, not the route in. **Do not
hand that URL to the system that will call it** - two shapes of it are in
circulation and `../wiki/api.md` says which one to take from the deployment
instead.

**A webhook and a schedule are not interchangeable** even though both start a run
with nobody watching. A webhook fires when their system decides; a Trigger fires
when we decide. If the customer cannot make their system call out, a schedule is
the fallback and it will always be later than the event.

## Before writing the chart

`../usecase/trigger.md` has the rules a scheduled run needs: the cursor and
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
- Cron being the only class left: checked 2026-09-02, re-read 2026-09-11
  against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube)
  `cbd8d70` - `TriggerClass`, `TriggerCronSpec`. **The `schedule` pattern is
  gone** as of that commit: the grammar is whatever `batch/v1` CronJob accepts,
  validated by a real cron parse in an admission webhook rather than by a regex
  in the schema

**Unchecked:** `TriggerClass` was held against the CRD and one deployment; the
UI fields come from the product documentation only.
