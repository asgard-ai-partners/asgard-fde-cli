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

**Only Odin lets the customer bring their own model.** Sindri and Mimir use the
platform's designated models, billed as Asgard Credit, and **the LLM cannot be
swapped there**.

    Odin Studio          a customised agent may use the customer's own LLM key
    Sindri Agent Hub     platform model, Credit-billed, not swappable
    Mimir Data Insight   the same

So "we will use our own Claude account" is answerable with yes or no depending on
which product the capability lands in - and that is decided at
`../guide/requirements.md` question 2b, before anyone has thought
about billing. A customer with a model contract, or a compliance rule about
where inference happens, needs to hear this while the shape is still open.

It also bounds `../wiki/settings.md`: a custom `CompletionModel` CR is an Odin-side
thing. It does not make a hub agent or a dashboard use the customer's key.

## What is billed, and in what units

The shape is durable even though the numbers are not, and the units are what
connect a design to a cost:

    Project              per project, per day
    Processor            per node, per day
    Loader               per loader, per day
    Data Indexer         per indexer, per day
    Knowledge storage    per GB, per day
    Semantic model generation      per generation
    Semantic relation generation   per generation

**A Loader and an Indexer are each several times a Project, per day.** With ten
Loaders allowed per Workspace, the document path is the one where a design
decision shows up on a bill - see [`knowledge.md`](knowledge.md), and count
sources rather than documents.

**A processor is billed per node per day**, so a workflow's node count is a
standing cost rather than a per-run one. A router-and-subworkflow design that
could have been one prompt costs every day it exists.

**The prices themselves are not copied here.** They change, and a stale price in
front of a customer is worse than none - the same reason this material carries no
screenshots and no outbound IP addresses. They are in the business-plan
repository's pricing reference; read them there when asked, and let sales quote.

## How cost is broken down

Overview's detail table has Service, Item, Usage / Unit and Cost.

- **Service** - Platform, Knowledge Base, Data Insight, Agent Hub, Heimdall
- **Item** - Project Usage, Processor Usage, Seat and so on
- **Usage / Unit** - Units-Days, GB-Days, Times

The trend chart colours the sources apart.

Units of the form **quantity times duration** mean a thing is billed for existing,
not only for being called. That is worth saying first when explaining cost to a
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
  Mimir cannot swap the LLM: an internal pricing reference, read 2026-09-02, not
  published anywhere an engagement can reach.
  **Deliberately not copied**: the prices, the subscription tiers, and the
  one-off project fees

- [Overview](https://docs.asgard-ai.com/docs/product-suite/fehu/overview),
  [usage reports](https://docs.asgard-ai.com/docs/product-suite/fehu/usage-reports),
  [billing reports](https://docs.asgard-ai.com/docs/product-suite/fehu/billing-reports),
  [plan](https://docs.asgard-ai.com/docs/product-suite/fehu/plan),
  [payment settings](https://docs.asgard-ai.com/docs/product-suite/fehu/payment-settings),
  [credit wallet](https://docs.asgard-ai.com/docs/product-suite/fehu/credit-wallet),
  [quickstart](https://docs.asgard-ai.com/docs/product-suite/fehu/quickstart),
  [quota limits](https://docs.asgard-ai.com/docs/product-suite/fehu/quota-limits)
  - asgard-docs `f00e0ee`
- `fehu/features.mdx` has been folded into quickstart and now only redirects
- The billing unit's definition:
  [glossary](https://docs.asgard-ai.com/docs/help-community/glossary)

**Unchecked:** everything here comes from the product documentation.
