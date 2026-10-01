---
group: Products and scope
description: the permission layers, the pages that disagree with each other, Workspace Management - and the path each thing sits at, for when the answer is where to click
---
# Management Console - permissions and Workspace

The Console does no business work. It decides two things: where users enter each
product, and who may manage or open which resource in which product - not who
may call it at runtime, which is below.

Its scope is one Workspace (`/workspace/{workspaceId}`), with no Project layer
beneath it.

## Where a thing is, as a path

The rest of this wiki maps a UI name to the CR behind it. This section gives the
route to each page, so that an agent asked "where do I set that?" in a meeting
can answer with somewhere to click rather than only a CR kind.

Everything hangs off the Workspace, and every per-resource page hangs off a
platform Project inside it:

    /workspace/<workspaceId>/overview
    /workspace/<workspaceId>/pipelines/<pipelineId>/releases/<releaseId>
    /workspace/<workspaceId>/project/<projectId>
    /workspace/<workspaceId>/project/<projectId>/app-toolsets/<name>/detail
    /workspace/<workspaceId>/project/<projectId>/app-skillsets/<name>/detail
    /workspace/<workspaceId>/project/<projectId>/drive/<name>/detail
    /workspace/<workspaceId>/project/<projectId>/chat-agent/<name>/detail

The name in a resource path is the CR's own `metadata.name`, which an
engagement already has in the chart it wrote. `app-toolsets` is an MCP Server, `drive` is a SourceSet, and `chat-agent`
is a Managed Agent - the path segments are the UI's vocabulary rather than the
CRD's, so this table and `index.md`'s UI-name mapping answer different halves.

The host cannot be derived and is not listed here. The Console and the Platform
API are different hosts, and a URL assembled from the wrong one looks correct and
resolves to nothing. `asgard-cli links` prints these for the checkout it is
run in, and `asgard-cli profile set --console` is where an installation records
its own.

An engagement does not hold the platform project id or a release
id: both are in the paths above, neither is written into a chart or a binding,
and a checkout that names its project does so in a comment. So a per-resource
link is assembled once by hand from the Console and then pasted. `links` says
so rather than guessing.

## Building a resource does not make it visible

After a semantic model or an agent is built in Odin, other people do not
automatically see it. To let them reach it in Mimir or Sindri, go back to the
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

To see what an account actually holds, click its row in Workspace Accounts: the
member's permission page lists Domain, Group, Product and Role, each expandable
to the underlying permission codes.

Both layers govern who manages resources in the Console, not who may call
them. A request reaching a deployed BotProvider, Toolset, SourceSet or
knowledge base is checked against that resource's own API key and the
Workspace's subscription, and carries no user identity - at asgard-core
`478cf5d6`, asgard-core `internal/edgeserver/middleware/`, where `bot_provider.go`,
`toolset.go`, `sourceset.go` and `knowledge_base.go` compare a header against
the resource's key and `subscription.go` checks the Workspace. So when a
customer needs one user limited to some of what an agent can reach, that is
decided in front of the platform: separate BotProviders with separate keys,
or a front end that decides who may send what. The platform will not tell two
callers holding the same key apart.

## The product permission pages do not match

Product Permission covers Odin, Mimir, Sindri, Heimdall and Fehu. The Console's
own introduction page still labels the same five Studio, Data Insight, Agent Hub,
Heimdall and Billing Portal, so a customer may use either set of names. Every one
manages accounts and roles, but the pages are laid out differently, so steps for
one page do not carry over to another.

| product | tabs | first column | main action | has Roles |
|---|---|---|---|---|
| Odin | Collaborators / Invite | Name | Add Account | yes, at the Project layer only |
| Mimir | Named User / Invite | Shared with | Purchase Named User | yes |
| Sindri | Named User / Invite | Shared with | Add Account | no |
| Heimdall | Named User / Invite | Name | Add Account | yes |
| Fehu | Collaborators / Invite | Name | Add Account | yes |

Differences to watch for:

- Sindri has no Roles. There is no entry point in the Console for editing
  its role definitions.
- Shares exists only for Mimir.
- Only Mimir is purchase-based (Purchase Named User); the rest
  invite directly.
- The URL slug and the navigation label disagree for Odin, which is `platform`.
  Sindri's slug is `sindri` - the product's own name, which `../wiki/sindri.md`
  is about.

### Bound to a resource, or bound to a product

Mimir and Sindri grant against a single resource, not a Project.
The Manage Accounts in selector at the top lists resources across every Project
in the Workspace - every semantic model for Mimir, every agent for Sindri - so
granting does not require switching Project first.

Odin is the only one with two layers: Platform -> Accounts affects every
Project beneath it, while Project -> Accounts / Roles applies to one Project, and
Roles exists only at the Project layer.

Heimdall and Fehu have no resource selector; permissions apply to the
whole product.

## Workspace Management

The Console's left navigation has three sections: Workspace Management, Group
Management (a single item, Group Access, which the published docs name and do
not describe), and Product Permission.

Workspace Accounts starts with a row of named-user seat cards, one per product
that sells seats (Mimir, Sindri and Heimdall), each showing seats available of
seats purchased; a card's Detail opens that product's seat page at
`accounts/seats/{product}`, where seats are assigned and unassigned. Below it
are Collaborators and Invitations tabs, with columns Name, Email, Groups, Named
User, Owner and Last Active. Each row offers Edit and Revoke, and clicking the
row itself opens the member's permission page at `accounts/{userId}`. Edit opens
a dialog with the member's groups, a checkbox per product seat, and Revoke
access. A Workspace owner is whoever is in the Owner Group - there is no
separate button to make someone an owner. Add people, at the top right, invites
a member.

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

## The audit log

The Console's Explore page queries the audit log. Each audit event carries who and where - product, namespace,
BotProvider, channel, run and session, and a user identity hint rather than a
name - then the workflow, processor, model, subagent, Toolset and tool it
concerns, and the content: the prompt, the reply, or a tool call's arguments
and result. Content is cut to 8 KB, each half of a tool call to 2 KB, and the
full text stays in the channel transcript the event points back at. The event
types are `user_prompt`, `assistant_response`, `tool_call`, `subagent_start`,
`subagent_complete`, `model_usage`, `run_done`, `run_error` and
`user_feedback` - asgard-core `478cf5d6`
asgard-core `internal/shared/auditlog/auditlog.go`. An event is a log line the logging
pipeline copies into the lakehouse the Console queries; how long it is kept is
set in that pipeline, and nothing in asgard-core or the kube repositories
configures it.

Every event is filed under the resource's `asgard-ai.com/product` label, which
`../usecase/conventions.md` covers.

## Corresponding extracts

The Console produces no CRs - permissions are not in any chart - so there is no
extract for it.

## Sources

- [About the Console](https://docs.asgard-ai.com/docs/product-suite/management-console/about-console/intro),
  [product permission](https://docs.asgard-ai.com/docs/product-suite/management-console/features/product-permission),
  - asgard-docs `6261fdff`
- [workspace settings](https://docs.asgard-ai.com/docs/product-suite/management-console/features/workspace-settings)
  - asgard-docs `95a27895`
- [About Odin](https://docs.asgard-ai.com/docs/product-suite/odin/about-odin/about)
  and [manage workspace](https://docs.asgard-ai.com/docs/product-suite/odin/about-odin/introduction/manage-workspace)
  - asgard-docs `6261fdff`

- The path shapes above are the hrefs of a partner deck built against a
  production Console. Two were corroborated against the
  ids that engagement's `.asgard-cli.yaml` records - the workspace and the
  pipeline - which is what makes them shapes rather than one page's URLs.

**Checked:** the path shapes against the Console they were observed in, which
is a screen rather than a document and carries no commit; the `overview`,
`project`, `app-toolsets`, `app-skillsets`, `drive` and `chat-agent/<id>/detail`
segments against the front-end routes named in
asgard-docs `95a27895` `content-generator/services/*/docs/**/metadata.json`; the
seat page path against asgard-docs `95a27895` `docs/product-suite/management-console/features/workspace-settings.mdx`;
that a call is checked by resource key and not by user, and the audit log's
fields, caps and event types, against asgard-core `478cf5d6`, the files named
above.

**Unchecked:** the permission pages, the Workspace pages and the `pipelines`
path come from the product documentation and one deck, and only a Console
account can confirm them; permissions are in no chart.
