# Management Console - permissions and Workspace

The Console does no business work. It decides two things: where users enter each
product, and who may use which resource in which product.

Its scope is one Workspace (`/workspace/{workspaceId}`), with no Project layer
beneath it.

## Building a resource does not make it visible

After a semantic model or an agent is built in Odin, **other people do not
automatically see it**. To let them reach it in Mimir or Sindri, go back to the
Console's matching page, use Manage Accounts in to select that resource, and add
the people. Each new resource repeats the step - permissions do not inherit
across resources.

When a customer reports that the thing is built but they cannot see it, look here
before looking at the chart.

## Two layers of permission

| layer | governs | scope |
|---|---|---|
| Workspace Accounts | members and owners | every Project in the Workspace |
| Product Permission | one product, or one resource | that product or resource |

The layers are independent. A Workspace collaborator does not thereby see
anything in any product; a Workspace owner has full management rights over
everything beneath it.

The fastest way to see what an account actually holds is View Permissions from
Workspace Accounts, which lists Domain, Role and Product, each expandable to the
underlying permission codes.

## The five product permission pages do not match

Product Permission covers Studio, Data Insight, Agent Hub, Heimdall and Billing
Portal. All five manage accounts and roles, but their composition differs -
carrying steps from one page to another means looking for buttons that are not
there.

| product | tabs | first column | main action | has Roles |
|---|---|---|---|---|
| Studio | Collaborators / Invite | Name | Add Account | yes, at the Project layer only |
| Data Insight | Named User / Invite | Shared with | **Purchase Named User** | yes |
| Agent Hub | Named User / Invite | Shared with | Add Account | **no** |
| Heimdall | Named User / Invite | Name | Add Account | yes |
| Billing Portal | Collaborators / Invite | Name | Add Account | yes |

Things that bite:

- **Agent Hub has no Roles.** There is no entry point in the Console for editing
  its role definitions.
- **Shares exists only for Data Insight.**
- **Only Data Insight is purchase-based** (Purchase Named User); the other four
  invite directly.
- The URL slug and the navigation label disagree: Studio is `platform`, Agent Hub
  is `sindri`.

### Bound to a resource, or bound to a product

Data Insight and Agent Hub grant against **a single resource**, not a Project.
The Manage Accounts in selector at the top lists resources across every Project
in the Workspace - every semantic model for Data Insight, every agent for Agent
Hub - so granting does not require switching Project first.

Studio is the only one with two layers: Platform -> Accounts affects every
Project beneath it, while Project -> Accounts / Roles applies to one Project, and
Roles exists only at the Project layer.

Heimdall and Billing Portal have no resource selector; permissions apply to the
whole product.

## Workspace Settings

**Workspace Accounts** has Collaborators and Invite tabs, with columns Name,
Email, Owner and Last Active. Each row offers View Permissions and Revoke. Add
Owner at the top right adds an owner directly, not an ordinary member.

**Rename Workspace** does not navigate anywhere. It overlays a dialog titled Edit
Workspace with a single name field.

## My Products

The landing page after signing in, split by purpose:

| section | products |
|---|---|
| Enterprise AI Solution | Odin, Mimir, Sindri |
| Public Opinion | Heimdall |
| Utility | Fehu |

More products, beside a section heading, shows what has not been enabled.

## Odin's own Workspace layer

Odin also has a Workspace layer, a different interface from anything inside a
Project, with only two navigation items:

- **Overview** - an Analysis section filterable by Project and time range:
  Requests per second, Request Duration, Token Usage (Completion and Embedding),
  Total Message
- **Project** - the My Projects list, with usage at the top right (for example
  `Used : 19 / 300`) and Add New Project. Creating one needs only a name

Project count is capped by the subscription plan. Agent Hub, Context, Automation,
Knowledge Base, Applications and Settings are all features inside a Project, not
at the Workspace layer.

Account and name management has moved to the Management Console.

## Corresponding extracts

The Console produces no CRs - permissions are not in any chart - so there is no
extract for it.

## Sources

- [About the Console](https://docs.asgard-ai.com/docs/product-suite/management-console/about-console/intro),
  [product permission](https://docs.asgard-ai.com/docs/product-suite/management-console/features/product-permission),
  [workspace settings](https://docs.asgard-ai.com/docs/product-suite/management-console/features/workspace-settings)
  - asgard-docs `f00e0ee`
- [About Odin](https://docs.asgard-ai.com/docs/product-suite/odin/about-odin/about)
  and [manage workspace](https://docs.asgard-ai.com/docs/product-suite/odin/about-odin/introduction/manage-workspace)
  - asgard-docs `f00e0ee`

**Unchecked:** everything here comes from the product documentation. Permissions
are not in any chart, so none of it could be held against one.
