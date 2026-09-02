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
