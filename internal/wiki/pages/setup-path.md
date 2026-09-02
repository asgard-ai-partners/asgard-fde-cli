# From a credential to an agent someone can talk to

Every other page here describes one object. This one describes the **order**,
because the order is not written down anywhere else - not in the product
documentation, which has a page per screen, and not in the extracts, which have a
shape per CR. Assembling it from three pages on the spot is how it was done
before this page existed, and everyone assembled it slightly differently.

Read it when a customer has just handed over access and the question is "so what
happens now", and when writing a deck or a handover that has to show them.

## The first fork: where does this credential even go?

The most common wrong turn is at the very first step, because the console has a
place called **Data Source** and people put everything there.

| what the customer handed over | where it goes |
|---|---|
| host, port, database, user, password | **Data Source**. Nine DB providers only - see [`settings.md`](settings.md) |
| an OAuth app for Dropbox / Google Drive / OneDrive / Google Sheets | **Connection** - [`settings.md`](settings.md) |
| **an API key or token for an HTTP API** | **neither.** It is a tool's config, not a data source |
| a chat platform's channel secret and access token | a `BotProvider` - [`integration.md`](integration.md) |

**An HTTP API has no home under Settings, and this surprises people.** Data
Source takes nine database providers and nothing else; Connection is OAuth to
five named services. A REST API with a bearer token is reached by a tool that
carries its own credentials - either an `http-request` step inside a Workflow, or
an MCP Server started with environment variables. `asgard-cli usecase external-api`
is the whole shape, including which of the two to take and why.

So "we have the API credentials, let's add the data source" is a sentence that
cannot be completed, and an hour is usually lost finding that out.

## The order

Five steps. Each one is finished and testable before the next, which is what
makes it worth doing in this order rather than starting from the agent.

**1. Land the credential.** Data Source, Connection, or a tool's own config, per
the table above. A Data Source has **Test Connection** on the form - use it
before Save. This is also where a network path failure shows up first, and it is
the failure that has nothing to do with the credential: see
[`operations.md`](operations.md), because a hosted platform reaching an internal
system needs the customer's firewall opened first.

**2. Turn access into something the agent can call.** Which one depends on what
the source is, and they are not interchangeable:

| the source | what to build | page |
|---|---|---|
| a database the agent should query freely | a **Semantic Model** | [`semantic-model.md`](semantic-model.md) |
| a database, but only a fixed set of answers | query tools in a Workflow | `asgard-cli usecase fixed-query-tools` |
| an HTTP API | a Workflow with `http-request`, wrapped as an **MCP Server** | `asgard-cli usecase external-api` |
| an existing MCP server somebody already wrote | **MCP Server**, From Existing | [`tools.md`](tools.md) |
| documents, manuals, FAQs | a **Drive** with a Context Index | [`knowledge.md`](knowledge.md) |
| procedural knowledge the agent needs while running | a **Skillset** | [`tools.md`](tools.md) |

**New MCP Server has two entries and they lead to different work.** From
Workflow asks only for Name and Description, then drops you into an empty
Workflow editor where the actual tool is assembled - the form being short does
not mean the step is. From Existing connects to a server that already exists,
over STDIO (Asgard starts a local process with a Command, Arguments and
Environment Variables) or Streamable HTTP (an endpoint already running).

**3. Configure the agent.** A Managed Agent is where the pieces meet - see
[`agents.md`](agents.md) for the fields. Two of them decide more than the rest:

  - **Description** is not a self-introduction. It is the routing text the
    orchestrator reads to decide whether to delegate this question to this
    agent, and it is the only thing that decision can see. Write it as "ask me
    when ...", not as "I am a helpful assistant for ..."
  - **Prompt** is four separate fields - Persona, Task, Context, Format - not one
    box. Putting everything in Persona works until the agent has two tasks

Then mount what step 2 produced: MCP Servers, Skillsets, Drives, Semantic Model,
and the Browser Configuration switch.

Five built-in templates exist and are worth starting from rather than an empty
form: 簡易線上客服, Customer Support, Knowledge Base Q&A, Data Analyst,
General Assistant.

**4. There is no step 4 for Sindri.** This is the question people ask most often
and the answer is that it is not a step:

> **Every Managed Agent is published to Sindri.** There is no import, no
> install, no "add agent to hub". An enabled agent serves immediately; a
> disabled one stops serving and keeps its configuration.

What a user sees in Sindri is the Available Agents list on the home page, each
card showing a name and a line about when to delegate to it - that line being the
Description from step 3. The user does not have to pick the right agent: the
selector defaults to Sindri, which routes on those descriptions. See
[`sindri.md`](sindri.md).

So the thing that makes an agent findable in the hub is not a publish action. It
is having written the Description well two steps earlier.

**5. Only if someone outside has to reach it**, this is a different path, and
this is where it stops being console work: a Flow Agent as the entry point, and a
`BotProvider` for the channel. An anonymous visitor cannot authenticate to the
hub at all. [`integration.md`](integration.md) has the four routes.

## Console or chart - the same five steps either way

The steps above are the console. An engagement builds them as a chart instead,
and the order is the same - each step still depends on the one before it. What is
not the same is the count: a console object is often several CRs.

| step | console | where the CRs are stated |
|---|---|---|
| 1 | Data Source | `DataConnector` - [`settings.md`](settings.md) |
| 2 | Semantic Model | `SemanticLayer` - [`semantic-model.md`](semantic-model.md) |
| 2 | MCP Server, query tool | a `Workflow` - [`tools.md`](tools.md) |
| 2 | Drive | `SourceSet` (+ `Syncer`) - [`knowledge.md`](knowledge.md) |
| 2 | Skillset | `SkillSet` + `SourceSet` + `Syncer` - [`tools.md`](tools.md) |
| 3 | Managed Agent | one `Agent` - [`agents.md`](agents.md) |
| 5 | Flow Agent | **three** CRs, not one - [`agents.md`](agents.md) |

**A console object is not one CR, and the names do not match.** There is no
`FlowAgent` kind, no `KnowledgeDrive` and no `HttpTool` - those are
`asgard-cli add` template names, and the CRs they write are in the column above.
[`agents.md`](agents.md) has the full UI-to-CR table and is where that fact
lives; this one is only here so the order can be followed in a chart.

**Connection is the exception in the other direction**: OAuth authorisation
happens in the UI and is not declared in a chart at all.

What this means in practice is that a screenshot of the console is a fair
illustration of what a chart does - the customer's admin sees step 3 as a form
whether or not the form is what we filled in. That is what makes the
documentation's screenshots usable in a handover deck.

## Screenshots, for a deck or a handover

[`screenshots.md`](screenshots.md) is the index - every picture, what it shows,
and which situation it is for. It carries the setup path as its own section, so
this is the short version. Fetch them from:

    https://docs.asgard-ai.com/img/docs/<path>

The console itself is in **English**; only the documentation's captions are
zh-TW, so a deck in either language can use them.

The three that carry this page:

| step | path |
|---|---|
| 1 | `settings/data-source/data-source-create-provider-open.png` - nine providers, all databases |
| 3 | `agent-hub-managed-agent/create.png` - the form an admin fills in |
| 4 | `sindri-home/home-available-agents.png` - what their users see |

**Check the file before using it.** These were captured against the product at
some point and nothing here tracks when. A screenshot showing an older form is
worse in a customer deck than no screenshot, because they will compare it to
what they see.

The last of the three is also the best available argument for writing a
Description as routing text - every card on it reads 「當使用者要⋯時,委派給 X
Agent」, which is the field doing its job in public.

## Corresponding extracts

No single extract covers this path, because it is the order rather than a shape.
The shapes it passes through are `asgard-cli usecase agent-hub`,
`external-api`, `semantic-layer`, `fixed-query-tools`, `knowledge-drive` and
`skill-set`.

## Sources

- [Managed Agent](https://docs.asgard-ai.com/docs/product-suite/odin/features/agent-hub-managed-agent),
  [MCP Servers](https://docs.asgard-ai.com/docs/product-suite/odin/features/mcp-servers),
  [Skillsets](https://docs.asgard-ai.com/docs/product-suite/odin/features/skillsets),
  [Configuration](https://docs.asgard-ai.com/docs/product-suite/odin/features/agent-hub-configuration),
  [Data Source](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/data-source),
  [Connection](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/connection)
  - asgard-docs `f00e0ee`
- Screenshot paths read off that same checkout's `static/img/docs/`, 2026-09-02
- Checked 2026-09-02 against asgard-kube `15ded0f`, head at the time: the CR column names only
  kinds the CRDs define. `HttpTool`, `KnowledgeDrive` and `FlowAgent` are not
  among them, which an earlier draft of this table asserted
- **Checked** 2026-09-02 against three of those images, opened rather than
  listed: the Data Source Provider list (nine entries, every one a database, no
  HTTP option), the Managed Agent form (the four Prompt fields and the live
  preview), and the Sindri home page (five Available Agents, each card showing
  its Description as routing text)

**Unchecked:** the order itself. Every step is documented and every claim about a
step comes from the page above it, but **no source states the sequence** - it is
assembled here from the object pages plus what the CR mapping implies, which is
the reason this page exists.
