This is the interview, and it produces a **request** - not a design, not a task,
not a chart. Read it before the working session with the customer, and keep it
open during one.

<<with .Requests>>Open requests:

<<range .>>  <<.ID>>  <<printf "%-8s" (printf "%s" .Status)>>  <<.Title>>
<<end>>
<<else>>Nothing is recorded yet, so start with `asgard-cli request add`.
<<end>>
## Why the interview is a stage of its own

Of the target repo's thirteen task specs, **three were superseded and one was
reverted**, and they were the three most expensive decisions in the engagement.
None of them flipped because the code was wrong. They flipped because the
question that decides them was answered from the design already in mind rather
than asked out loud.

All three are decided by the same fact, and it is the cheapest thing to ask for:

    who is on the other end

Ask it once per capability the customer wants, before anything else.

## The order to ask in

The order is not a style preference. Each answer narrows the next question, and
asking them out of order means designing against an audience nobody confirmed.

**1. What can the agent not do today?**

In their words, before translation. Write down the sentence they actually said,
including the parts that sound imprecise - "the warehouse people keep phoning to
ask about stock" carries who, where and why, and "a stock query API" carries
none of it. The translation into platform terms is the request's own section 2,
and keeping the original is the only way a later reader can check that the
translation was right.

One request per thing they asked for. Two capabilities in one record cannot be
given different target projects, and the target project is the decision this
whole stage exists to reach.

**2. Who is on the other end?**

    internal, authenticated callers   -> the platform's agent hub, semantic layers
    public, anonymous visitors        -> your own BotProvider, fixed query tools

Not a preference and not a later refinement: the two paths share neither an
entry point nor a read path, so a request that mixes both audiences is two
requests. Same audience as an existing project means this request goes into that
project; a new audience means a new project.

Ask it concretely. "Do they log in to something today, and is it ours?" gets an
answer; "are they internal users?" gets a yes that means nothing, because a
contractor with a company address is internal to the person answering and
anonymous to the platform.

**3. Which systems hold the data, and how can each one be reached?**

One row per system, including the ones that sound obvious.

| system | what it holds | how to reach it | who asks | what they want | read or write |
|---|---|---|---|---|---|

"How to reach it" has a preference order, and it is not a matter of taste:

    a database we can read   >   an API   >   a screen a person clicks

Take the highest available per system, and ask about every system rather than
inferring. A warehouse system with a REST API is a different design from one
with only a database, and both are different from one with only a web console -
the last of which is browser operation, and costs more than the other two
combined.

Then ask the question that saves the most work in the whole interview:

    "Is there already something that pulls these together for you?"

A middleware layer, an OMS, a warehouse that consolidates the channels. If one
exists, several rows collapse into a single database, and that single answer
changes the design more than anything else on this page.

**3b. For each system we will actually connect to, get the coordinates.**

Ask in the meeting, not by email afterwards. Every one of these has turned an
integration from days into weeks by being discovered late:

  - host, port, database or schema, and the account name
  - **is the account read-only?** Ask explicitly. The one offered first usually
    is not, and finding out later means going back for a second credential
  - **is it reachable from the cluster**, or does it need an IP allowlist, a
    VPN, or a bastion? This is the single most expensive thing to discover in
    week three, and it costs one sentence to ask in week one
  - who issues the credential, by name or role. A credential with no owner is
    not a dependency, it is a delay
  - for an API instead of a database: the auth scheme, who holds the client id
    and secret, whether there is a sandbox, and the rate limit - the rate limit
    decides whether a Syncer can backfill at all

They go in the request's section 4, one block per system.

**Passwords do not.** The request record is committed to git, and a secret that
reaches a commit is not fixed by deleting the line - the history keeps it. It is
fixed by rotating the credential, which means going back to the customer to ask
for a new one, having just told them we leaked the last one.

    coordinates  -> the request record, and chart/values-<env>.yaml
    passwords    -> .env locally (gitignored), app-secret in the cluster

Write the *name* of the key in the record - `<TARGET>_DB_PASSWORD` locally,
`<target>_db_password` in app-secret - and never its value. `.env.example` at
the repo root has the full pattern, including the three places a new database
has to be registered before it works end to end.

If the customer wants to hand over a password during the meeting, take it into
`.env` there and then and say why it is not going in the notes. Doing that once
in front of them is usually the last time they paste one into a chat.

**3c. What does it have to fit into, not just read from?**

Section 3 asks where the data is. This asks what already exists around it, and
it is a separate question because customers do not volunteer the answer - the
systems they mention are the ones holding data, not the ones the agent will have
to live alongside.

  - Is there **already a bot or a helpdesk** the staff use? A new channel that
    duplicates one is a channel nobody opens. Sometimes the right answer is to
    become a tool inside theirs rather than a front end beside it
  - Is there an **internal portal or admin console** this should sit inside? That
    decides the entry point as firmly as the audience does
  - Has anyone **tried to build this before**? Ask directly. A previous attempt
    tells you which part turned out to be hard, and they will not mention it
    unless asked, because it did not work
  - Is there a system with **no API and no database** - only a web console a
    person clicks? That is browser operation, and it costs more than every other
    integration on the list combined. Establish it now, not in week three
  - **Who owns each system internally?** Not the credential owner - the person
    whose approval is needed before anything touches it. Integration work stalls
    on this more often than on anything technical

**3d. If they are reached through a chat platform, which one?**

Question 2 decides whether the entry point is the platform's hub or one of your
own. This is the separate question of **which channel**, and it is worth asking
in the same breath because `botProviderClass` is **immutable once created** -
changing it later means a new BotProvider, not an edit.

    LINE / Telegram          the platform posts a webhook; one CR
    Slack / Discord          a connector pod holds a socket; infra provisions it
    your own front end       generic, and you own the appearance

Ask where their users already are, not where it would be convenient to put them.
An official account with a following is a distribution channel a widget cannot
reproduce, and asking those people to visit a web page instead loses most of
them.

**3e. Is there anything between the channel and us?**

Ask this whenever the answer to 3d is a chat platform, and ask it early, because
a whole class of requirement depends on it and the customer will not raise it
themselves.

    the customer  ->  ???  ->  Asgard

Whatever sits in that middle - a support desk, a helpdesk product, their own
relay, or nothing at all - is what owns the conversation. The platform does not:
there is no CR for handing over to a human, for pausing while a person replies,
for resuming afterwards, or for counting how many questions one user has asked.
`asgard-cli wiki integration` has the detail and the sources.

So every requirement of this shape belongs to that middle layer, not to us:

    "transfer to a real agent"           the desk takes the thread
    "pause the AI while a human replies"  the desk stops forwarding
    "three failures then a human"         the desk counts
    "ten questions per user per day"      the desk counts

With a website the middle layer is obvious, because the site is already there.
With LINE it often does not exist - but **how much LINE gives you for free is
unresolved**, so do not walk in saying it cannot be done. What is certain is that
LINE gives the bot no signal that a person has taken over, so the pause/resume
state and the counters live outside the platform either way. Ask who owns the
LINE Official Account and what their agents use today, and check LINE's current
documentation for that account. See `asgard-cli wiki integration`.

If the answer is "nothing", say so plainly rather than designing around it. The
choice is theirs: put a desk in front, or drop the requirement. A proposal that
promises handoff with nothing in the middle is a promise nobody can keep.

Two platform limits worth handing over in the same conversation, because they
shape what can be asked for: one request gets **30 steps and 3 minutes**, and an
endpoint serves **5 requests per second**. A troubleshooting conversation that
consults a knowledge base, then a CRM, then a ticket system, then asks a
follow-up, can reach 30 steps.

LINE also needs **two-way setup** - Asgard issues a webhook URL that somebody has
to paste back into the LINE console and verify - so it needs an owner on their
side, not just a credential. See `asgard-cli wiki integration`.

**4. Is any of it unstructured?**

Documents, FAQs, pages on a website - things a query cannot answer exactly.
Those become a Drive with a knowledge graph, and they are a different shape from
rows in a database. Ask separately; customers rarely volunteer documents when
the conversation has been about systems.

Anything a query DOES answer exactly - counts, prices, stock levels, contact
details - belongs to a query tool, not to a Drive. Putting a number in a Drive
makes the agent paraphrase a figure it should have read.

**5. Is there anything it should change, and not just read?**

The standing architecture is read-only. A write is not a project decision to be
absorbed into this request: it is a warning, it needs its own spec and human
approval, and it cannot be reached from a scheduled run, because a schedule has
nobody to approve it.

Record it in the row and say so out loud in the meeting. A write path discovered
after the read path is built is the most expensive rework in this repo's
history.

**6. How will they know it worked?**

Ask for the sentence they would use to tell a colleague it was working, then
keep asking until it names something observable. "It answers stock questions" is
not verifiable; "the warehouse lead stops phoning about location 608" is.

This becomes the acceptance criteria of the task specs, and a request whose
success nobody can describe produces tasks nobody can close.

**7. What is explicitly out of scope?**

Write down what you are NOT building, particularly the things they mentioned in
passing. An unrecorded "we could also..." returns as an assumption three weeks
later, and by then nobody remembers whether it was agreed.

## Material they hand you

Ask for it in the meeting, before you need it: API documentation, an operation
manual, a schema dump, an ERD, a field dictionary, a status-code table, the
screenshots someone made for training new staff. Customers usually have more
written down than they think, and none of it arrives unless asked for.

**It does not go in the request record.** Three directories, and the difference
is who reads them:

    references/            background, for humans and spec-writing agents
    requirements/          the implementation source of truth
    common/skills/         what the RUNNING agent needs, synced to the platform

So: paste the material into `references/`, and have the request **cite** it. A
request that inlines forty pages of API documentation stops being readable as a
request, and the section that matters - what the customer asked for and who is
on the other end - disappears into an appendix.

The rule the repo already enforces: convert reference material into
`requirements/` before implementing, and if the two conflict, `requirements/`
wins and the conflict becomes an open question or a decision record. Nobody
implements straight from `references/`, because material a customer wrote for
their own staff describes the system they believe they have.

**Domain knowledge the agent needs at runtime is the third case, and it is the
one people get wrong.** Status-code meanings, cross-system entity mapping,
aggregation conventions - those belong in `common/skills/<skill>/SKILL.md`,
because they have to be synced into the platform to be usable at all. Left in
`references/`, they are invisible to the running agent, and the failure looks
like a model that ignores instructions rather than a file in the wrong place.

### Filing a large body of material

A manual worth keeping is worth structuring. One shape already carries 300K of a
web console's operation map without becoming unreadable:

    <topic>/SKILL.md                  the router: what this covers, and when
    <topic>/references/conventions.md the cross-cutting rules read first
    <topic>/references/page-map.md    broad and shallow: every entry point
    <topic>/references/api/<domain>.md one file per domain, index table on top

What makes it usable is not the layout but two habits inside it:

  - **Every index row carries its provenance.** That set records, per operation,
    whether the contract was actually observed or is only good enough to
    navigate to. A row that says "not verified" is worth more than a plausible
    one, because the reader knows which to trust
  - **Measured, not inferred, and it says which.** Routes were read off the real
    UI rather than derived from other routes, and the file states that. When a
    later reader finds a mismatch they know it is drift, not a guess

Both are the same discipline the request record asks for in section 4:
**say what was confirmed, how, and when.** Material without provenance reads
identically whether it is accurate or three years stale.

## The three answers that look obvious and were wrong

Each of these was decided one way, built, and reversed. They are not
hypothetical.

| the obvious answer | what it turned out to be | why the obvious one failed |
|---|---|---|
| a public website should read through a **SemanticLayer**, like everything else | **five zero-parameter query tools** | a layer without `allowedCubes` is arbitrary SQL over every cube, and the exposed surface grows by itself every time a table is added. Nobody widens it on purpose |
| a public website is reached through **the platform's agent hub**, like everything else | its **own BotProvider -> Workflow -> SandboxBlueprint** | an anonymous visitor cannot authenticate to the agent hub. `BotProvider.entrypoint` takes a Workflow, never an Agent |
| unstructured knowledge is a **KnowledgeBase** with Loaders and a retrieval workflow | a **SourceSet Drive with `contextIndex`** | the Loader-and-retrieval-workflow path was harder to keep correct than a Context Index over files. `KnowledgeBase` is still live and still shipping, so this one is experience rather than a platform rule |

The pattern in all three: **"like everything else" is the wrong reason**, because
the audience is what decides, and the audience is the one thing "everything
else" does not share.

## Write it down as you go

Not afterwards. A record written after the meeting is a record of what you
remember, and what you remember is the design you were already forming.

    asgard-cli request add "<what they asked for, in their words>"
    asgard-cli request target <<.RequestID>> <project>
    asgard-cli question add "<what blocks it>" --blocks <<.RequestID>> --ask "<who can answer>"

`request add` writes `requirements/requests/REQ-xxx-<name>.md` with today's date
and `draft` on it, and its seven sections are this interview in the same order.
Fill them in the file; the TODOs are the questions above.

**Open questions go to `docs/open-questions.md` as well**, one row each, with
what they block and who can answer. Not only into the request: a question buried
in a spec disappears when that spec reaches `done`, and `asgard-cli next` reads
the open-questions file on every run and prints it before anything else.

Ask who can answer, by name or by role, in the meeting. A question with no owner
is not tracked, it is just written down.

## When the request is ready

`asgard-cli request ready <<.RequestID>>` when all of these hold:

  - the customer's own wording is in section 1, unedited
  - the audience is decided, and the target project follows from it
  - every system has a row, and every row says how it is reached
  - every system we will connect to has a section 4 block, with a named
    credential owner and an answer on network reach - and no secret in it
  - unstructured knowledge is either listed or explicitly ruled out
  - every write is marked as a write
  - success is described in terms somebody could check
  - nothing that blocks the work is missing an owner

Not ready is a normal state to be in. A `draft` that names its open questions is
more useful than a `ready` that guessed at them, and `asgard-cli next` will keep
the request in front of you either way.

Then split it into task specs:

    asgard-cli task add "<title>" --request <<.RequestID>> --project <project> --complexity M
