---
group: Products and scope
description: billing and usage, how cost is broken down - and which product lets a customer bring their own model
---
# Fehu - billing and usage

AI usage billing. Rarely touched during onboarding, but a customer asking about
cost asks in detail.

| page | holds |
|---|---|
| Overview | period totals, a trend chart, a cost breakdown |
| Usage Reports | monthly usage and units |
| Billing Reports | look up bills, view them, download invoices |
| Plan | current plan, switching, buying additional products |
| Payment Settings | billing information, address, payment methods |
| Credit Wallet | credit balance and transactions, where credit billing applies |

## The constraint that is not about money

Only Odin lets the customer bring their own model. Sindri and Mimir use the
platform's designated models, billed as Asgard Credit, and the LLM cannot be
swapped there.

    Odin Studio          a customised agent may use the customer's own LLM key
    Sindri Agent Hub     platform model, Credit-billed, not swappable
    Mimir Data Insight   the same

So "we will use our own Claude account" is answerable with yes or no depending on
which product the capability lands in, and that is decided at
`../guide/requirements.md` question 2b, before anyone has thought
about billing. A customer with a model contract, or a compliance rule about
where inference happens, needs to hear this while the shape is still open.

It also bounds `../wiki/settings.md`: a custom `CompletionModel` CR is an Odin-side
thing. It does not make a hub agent or a dashboard use the customer's key.

## What is billed, and in what units

The units change less often than the prices, and they are what connect a design
to a cost:

    Project              per project, per day
    Processor            per node, per day
    Loader               per loader, per day
    Data Indexer         per indexer, per day
    Knowledge storage    per GB, per day
    Semantic model generation      per generation
    Semantic relation generation   per generation

A Loader and an Indexer each cost several times a Project, per day. With ten
Loaders allowed per Workspace, the document path is where a design
decision shows up on a bill - see [`knowledge.md`](../wiki/knowledge.md), and count
sources rather than documents.

A processor is billed per node per day, so a workflow's node count is a
daily cost rather than a per-run one. The platform's meter counts it that way:
a recurring job records, per namespace, the processors declared across every
Workflow, the Loaders, the Indexers and the stored content size, whether or not
anything ran (asgard-core `478cf5d6` `internal/jobrunner/job/meter_job.go`). A router-and-subworkflow design that
could have been one prompt costs more every day it exists.

The prices are not copied here. They change, and a stale price shown to a
customer does more harm than no price. This material carries no screenshots
for the same reason. They are in the business-plan repository's pricing reference; read
them there when asked, and let sales quote.

## How cost is broken down

Overview's detail table has Service, Item, Usage / Unit and Cost.

- **Service** - Platform, Knowledge Base, Data Insight, Agent Hub, Heimdall
- **Item** - Project Usage, Processor Usage, Seat and so on
- **Usage / Unit** - Units-Days, GB-Days, Times

The trend chart colours the sources apart.

Units of the form quantity times duration mean a thing is billed for existing,
not only for being called. Say that first when explaining cost to a
customer.

## Billing Reports

The list shows Billing#, Billing Period, Type (Subscription and so on), Amount,
Payment Status, Invoice Status and Bill Date. Each row offers View Bill and
Download Invoice.

Opening a bill shows the Workspace, the period, the total and the payment status,
with Service listing the cost by item and Invoice holding the invoice and its
download.

## Relation to the billing unit

A Workspace is the smallest unit a plan is billed against, and the number of
Projects is capped by the plan. So how workspaces and projects are divided shapes
the bill directly.

## Corresponding extracts

Billing produces no CRs, so there is no extract for it.

## Sources

- The product-line split on model choice, the billing units, and that Sindri and
  Mimir cannot swap the LLM: an internal pricing reference, not
  published anywhere an engagement can reach.
  Deliberately not copied: the prices, the subscription tiers, and the
  one-off project fees

- [Overview](https://docs.asgard-ai.com/docs/product-suite/fehu/overview),
  [usage reports](https://docs.asgard-ai.com/docs/product-suite/fehu/usage-reports),
  [billing reports](https://docs.asgard-ai.com/docs/product-suite/fehu/billing-reports),
  [plan](https://docs.asgard-ai.com/docs/product-suite/fehu/plan),
  [payment settings](https://docs.asgard-ai.com/docs/product-suite/fehu/payment-settings),
  [credit wallet](https://docs.asgard-ai.com/docs/product-suite/fehu/credit-wallet),
  [quickstart](https://docs.asgard-ai.com/docs/product-suite/fehu/quickstart),
  [quota limits](https://docs.asgard-ai.com/docs/product-suite/fehu/quota-limits)
  - asgard-docs `6261fdff`
- `fehu/features.mdx` has been folded into quickstart and now only redirects
- The billing unit's definition:
  [glossary](https://docs.asgard-ai.com/docs/help-community/glossary)

**Checked:** that processors, Loaders, Indexers and knowledge storage are
metered as standing counts per namespace, against
asgard-core `478cf5d6` `internal/jobrunner/job/meter_job.go`.

**Unchecked:** the Fehu pages come from the product documentation and only a
Console account can show them; the model-choice split and the Loader limit come
from an internal pricing reference with no copy here.
