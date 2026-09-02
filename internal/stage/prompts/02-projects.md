<<with .InterviewRequests>><<range .>>  <<.ID>>  <<.Title>>
<<end>>
That request has no target project yet, and deciding it is this stage. The output
is a decision, not code.<<else>>No request is open, so this page is being read out of order. This stage is the
interview that decides how the work splits into projects, and it starts from
something a customer asked for: `asgard-cli request add "<what they asked for>"`.

The interview that produces one is `asgard-cli next --stage requirements`.<<end>>

A **project** is one Helm chart deployed to one namespace, and it always lives
under this workspace at `projects/<slug>/` - `asgard-cli project add` puts it
there. There is nothing to decide about where it goes.

The split itself is not about tidiness - it follows the audience, because two
audiences need different entry points and different read paths, and those cannot
be shared:

  internal, authenticated callers   -> the platform's agent hub, semantic layers
  public, anonymous visitors        -> your own BotProvider, fixed query tools

So the question that decides the split is **who is on the other end**, asked once
per capability the customer wants.

**A second question decides whether it is a project at all.** What do they do
with the answer - ask it in the moment, or watch the same numbers every day? The
second is a **Mimir dashboard**: a chart of DataConnector and SemanticLayer with
no entry point, no Agent and nothing to publish, and the deliverable after that
is built by the customer in the product rather than by us in a chart.

    ask     -> a project, in the sense this page means
    watch   -> a read surface, and the Views and Dashboards are theirs to make
    both    -> one model, two deliveries. Do not merge them into one estimate

It still lives under `projects/<slug>/` and still deploys to a namespace, so it
is a project mechanically. What it is not is a project shaped like the rest of
this walk: stages 5 and 6 have nothing to say about it, and `asgard-cli verify`
will report an Agent with no capability only because there is no Agent.
`asgard-cli usecase mimir-dashboard` is the shape.

## What to ask the customer

  - Which business systems hold the data an agent would need to read?
    Get the system's name, what it is for, and **how it can be reached** - this
    last one decides the whole integration, so do not leave it vague:

        "Is this one of ours, and can we read its database directly?"
        "Does it have an API? Who has the credentials and the docs?"
        "If neither - is the only way in a person clicking through a screen?"

    Ask it about **every** system, including the ones that sound obvious. A
    warehouse system that turns out to have a REST API is a different design
    from one that only has a database, and both are different from one that
    only has a web console.
  - For each one: who asks the questions? Employees who log in to something, or
    anonymous visitors on a public page?
  - Is any of the knowledge unstructured - product documents, FAQs, pages on a
    website - rather than rows in a database?
  - Is there anything the agent should be able to **change**, not just read?
    Every write path needs a spec and human approval, so find out early.

## Build this table as you ask

It is the whole output of the interview, and the split falls straight out of it:

| system | what it holds | how to reach it | who asks | what they want | read or write |
|---|---|---|---|---|---|
| ERP | 料件庫存、採購單 | MSSQL, ours | 內部,登入 | 現有量、在途量 | read |
| a marketplace | 該通路的庫存與訂單 | REST API, theirs | 內部,登入 | 跨通路比較 | read + write |
| 官網型錄 | 產品、分類、據點 | PostgreSQL, ours | 匿名訪客 | 產品查詢 | read |

**"How to reach it" has a preference order, and it is not a matter of taste:**

    1. a database we can read      most capable: the agent composes its own
                                   queries and joins across tables
    2. an API                      a fixed set of calls, but a real contract,
                                   reviewable and gateable
    3. a screen a person clicks    last resort: brittle, slow, and it breaks
                                   whenever the vendor changes their UI

Take the highest one available **per system**, and expect a mix. Ask whether the
customer already has something that has done this consolidation for them - a
middleware layer, an OMS, a warehouse that already pulls the channels in. If they
do, several external systems collapse into one database, and that changes the
design more than any other single answer.

**Group the rows by "who asks", not by system.** Each distinct audience is a
project, because an audience determines the entry point and the read path and
those cannot be shared. Two systems read by the same audience belong in one
project; one system read by both audiences is read twice, through two shapes.

A row wanting **write** is not a project decision but a warning: every write path
needs its own spec and human approval, and the standing architecture is
read-only. Note it and move on - `asgard-cli usecase write-path` is where the
approval gate is described, when it comes to designing one.

## What to write down

**The table goes into the request**, section 3 of
`requirements/requests/<<.RequestID>>-*.md`, and the audience you settled
on goes into its Meta as the target project. That file already carries the date
and the status, which is why the interview output belongs there rather than in a
loose note.

Then, in this order:

  1. **Anything the interview did not settle**, one question each:

         asgard-cli question add "<question>" --blocks <<.RequestID>> --ask "<who>"

     Do it the moment a question blocks a decision. Not in your head, and not
     buried in a task spec - a task's open questions disappear when it reaches
     `done`, while `asgard-cli next` prints this file on every run.

  2. **The split, as a decision record:**

         asgard-cli decision add "how the work splits into projects" --module architecture.md

     Fill in **why the rejected split was rejected**. That sentence is the only
     part of the document nobody can reconstruct later. The command stamps the
     date and links the record from the living spec, so neither is yours to
     remember.

  3. **The raw discussion**, in `docs/meeting-notes/YYYY-MM-DD-<topic>.md`, if it
     was long enough to be worth keeping. Optional: a decision record can cite a
     GitHub issue or a direct instruction instead.

  4. **The first module of the living spec**, `docs/spec/<<.SpecSlug>>/architecture.md`:
     the table, the audience of each project, and the invariants that hold from
     day one (which systems are read-only, where any write path points). This is
     what the next person reads to understand the system.

     **Add it to the module index in `docs/spec/<<.SpecSlug>>/README.md` in the
     same change.** `asgard-cli check` compares that index against the files
     actually present, so a module written without being indexed turns the gate
     red.

## Then register each project, and point the request at it

    asgard-cli project add <slug> --env dev
    asgard-cli scaffold

Then point the request at it. That is what moves this stage on: until the request
names a project this repository has, `asgard-cli next` reads it as an interview
that has not finished.

    asgard-cli request target <<.RequestID>> <slug>
    asgard-cli request ready <<.RequestID>>          once the audience and the scope are settled

Keep the slug short. It becomes part of every namespace
(asgard-<<.Workspace.Slug>>-<slug>-<env>), and names derived from a namespace
inherit its length.

Done when: projects/ has a directory per project, the root README table lists
them, and asgard-cli check is green.
