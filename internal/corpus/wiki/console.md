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

Everything hangs off the Workspace, and every per-resource page hangs off a
platform Project inside it:

    /workspace/<workspaceId>/overview
    /workspace/<workspaceId>/pipelines/<pipelineId>/releases/<releaseId>
    /workspace/<workspaceId>/project/<projectId>
    /workspace/<workspaceId>/project/<projectId>/app-toolsets/<name>/detail
    /workspace/<workspaceId>/project/<projectId>/app-skillsets/<name>/detail
    /workspace/<workspaceId>/project/<projectId>/drive/<name>/detail
    /workspace/<workspaceId>/project/<projectId>/chat-agent/<name>/detail

**The name in a resource path is the CR's own `metadata.name`**, which is why
these are worth having: an engagement holds that name already, in the chart it
wrote. `app-toolsets` is an MCP Server, `drive` is a SourceSet, and `chat-agent`
is a Managed Agent - the path segments are the UI's vocabulary rather than the
CRD's, so this table and `index.md`'s UI-name mapping answer different halves.

**The host is not derivable and is not here.** The Console and the Platform API
are different hosts, and a URL assembled from the wrong one resolves to nothing
while looking correct. `asgard-cli links` prints these for the checkout it is
run in, and `asgard-cli profile set --console` is where an installation records
its own.

**What an engagement does not hold** is the platform project id and a release
id: both are in the paths above, neither is written into a chart or a binding,
and a checkout that names its project does so in a comment. So a per-resource
link is assembled once by hand from the Console and then pasted, which is what
`links` says rather than guesses.

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

- The path shapes above are the hrefs of a partner deck built against a
  **production Console**, read 2026-09-17. Two were corroborated against the
  ids that engagement's `.asgard-cli.yaml` records - the workspace and the
  pipeline - which is what makes them shapes rather than one page's URLs. No
  published source states any of them.

**Checked:** 2026-09-15 for the path shapes only, against the Console they were
observed in - which is a screen rather than a document, so it carries no commit
and nothing here can tell you it has not moved since.

**Unchecked:** everything else here comes from the product documentation.
Permissions are not in any chart, so none of it could be held against one. **No
path here was exercised by this tool against a Console** - they were read off a
document that was, so a segment the product renames breaks them silently.
