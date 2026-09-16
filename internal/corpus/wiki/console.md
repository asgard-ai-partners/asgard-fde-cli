---
group: Products and scope
description: the permission layers, the pages that disagree with each other, Workspace settings - and the path each thing sits at, for when the answer is where to click
---
# Management Console - permissions and Workspace

The Console does no business work. It decides two things: where users enter each
product, and who may use which resource in which product.

Its scope is one Workspace (`/workspace/{workspaceId}`), with no Project layer
beneath it.

## Where a thing is, as a path

**A page an FDE cannot name a route to is a page they describe instead of
opening.** The rest of this wiki maps a UI name to the CR behind it; this is the
other half of the same question, and without it an agent asked "where do I set
that?" in a meeting can answer with a CR kind and not with somewhere to click.

**Observed, not documented.** These were read off a production Console while
building a partner deck, 2026-09-15, and no source states them - so treat a
shape that does not resolve as this page being behind rather than as the reader
being wrong:

    /workspace/<workspaceId>/overview     the Workspace, and what `.asgard-cli.yaml`
                                          records is exactly this id
    app-toolsets/<toolset>/detail         an MCP Server
    app-skillsets/<skillset>/detail       a Skillset
    drive/<sourceSet>/detail              a Drive
    chat-agent/<agent>/detail             a Managed Agent

**The prefix those four hang off is not known**, and neither is the path to a
Pipeline's Variables tab - that one is reached as Pipelines, the repository's
name, the release, then the Variables tab, which is a route to describe rather
than a URL to build. `platform-unknowns.md` P16 carries both, because a link
assembled from a guessed prefix is the failure that renders identically to a
correct one and is caught only by somebody clicking it.

**The object name in the path is the CR's own `metadata.name`**, which is why
these are worth having: an engagement holds that name already, in the chart it
wrote.

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

## The product permission pages do not match

Product Permission covers Studio, Data Insight, Agent Hub, Heimdall and Billing
Portal. Every one manages accounts and roles, but their composition differs -
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
- **Only Data Insight is purchase-based** (Purchase Named User); the rest
  invite directly.
- The URL slug and the navigation label disagree: Studio is `platform`, Agent Hub
  is `sindri` - the product's own name, which `../wiki/sindri.md` is about.

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

- The path shapes above were read off a **production Console** on 2026-09-15,
  while building a partner deck that linked to each of those pages, and the
  workspace shape was confirmed against the id `.asgard-cli.yaml` records for
  that engagement. No published source states any of them.

**Checked:** 2026-09-15 for the path shapes only, against the Console they were
observed in - which is a screen rather than a document, so it carries no commit
and nothing here can tell you it has not moved since.

**Unchecked:** everything else here comes from the product documentation.
Permissions are not in any chart, so none of it could be held against one. **The
prefix the four resource paths hang off was not captured**, so those four are a
shape to recognise rather than a URL to build - `platform-unknowns.md` P16.
