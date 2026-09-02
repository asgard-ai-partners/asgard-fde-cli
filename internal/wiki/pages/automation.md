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

## Before writing the chart

`asgard-cli usecase trigger` has the rules a scheduled run needs: the cursor and
the cold start, why not to write a BotProvider yourself, and why a schedule
cannot use a tool that asks for consent. Not repeated here.

## Sources

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
