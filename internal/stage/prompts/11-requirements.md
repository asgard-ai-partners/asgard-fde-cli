This is the interview, and it produces a **request** - not a design, not a task,
not a chart. Read it before the working session with the customer, and keep it
open during one.

## What this stage produces is a file

Everything below is how to think during the interview. This is what to do with
it, and it comes first because the thinking is the part that goes well on its
own. Three commands, run **as the answers arrive**, not at the end:

    asgard-cli request add "<what they asked for, in their words>"
    asgard-cli request target <<.RequestID>> <project>
    asgard-cli question add "<what blocks it>" --blocks <<.RequestID>> --ask "<who>"

**One request per capability they asked for.** A document with three scenarios
is three requests, because two capabilities in one record cannot be given
different target projects, and the target project is the decision this whole
stage exists to reach.

The way this stage fails is not a bad analysis. It is a good one that stays in
the conversation: the material gets read, the questions get filtered well, the
answer is delivered to whoever asked, and the repository ends the day looking
exactly as it did before. The next run of `asgard-cli next` then says nothing is
in flight, and it is right. Anything worth telling someone is worth `request
add` first - **the message is the summary of the record, not a substitute for
it.**

If you are answering a question rather than running a meeting - "list the open
questions in this document" - that is still this stage. Write the records, then
answer from them.

**You will also need something to take into the room.** The deck is the only
thing in the repository the customer reads, and its shape is not this page's:
the `proposal-deck` skill in `.agents/skills/` owns it, including the one to
build while the questions are still open. Read it before writing slides, not
after.

<<with .Requests>>Open requests:

<<range .>>  <<.ID>>  <<printf "%-8s" (printf "%s" .Status)>>  <<.Title>>
<<end>>
<<else>><<if .References>>**<<.References>> file(s) of customer material are filed in `references/`, and no
request records any of it.** The interview has started and the repository has
no record of it. Write the request before going further - `asgard-cli check`
reports this state until one exists.
<<else>>Nothing is recorded yet, so start with `asgard-cli request add`.
<<end>><<end>>
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

This decides how they reach it. **2b decides what "it" is**, and the two are
asked together.

Not a preference and not a later refinement: the two paths share neither an
entry point nor a read path, so a request that mixes both audiences is two
requests. Same audience as an existing project means this request goes into that
project; a new audience means a new project.

Ask it concretely. "Do they log in to something today, and is it ours?" gets an
answer; "are they internal users?" gets a yes that means nothing, because a
contractor with a company address is internal to the person answering and
anonymous to the platform.

**2b. What do they do with the answer?**

Ask it in the same breath as question 2, because it decides **which product this
is**, and everything from question 3 down assumes the answer.

    look one thing up, in the moment     an agent
    watch the same numbers every day     a Dashboard - this is Mimir
    both, for different people           both, and they are separate deliveries

`asgard-cli wiki product-suite` puts it plainly: **Mimir is often what the
customer actually wants.** "I want an AI that answers stock questions" is not a
statement about an agent - it is a statement about stock questions, and the two
products answer it differently. Glancing at a figure each morning is a
Dashboard; looking one thing up when a customer is on the phone is an agent.

The failure this prevents is the expensive one and it is silent: an agent gets
built, it works, and the customer keeps asking the same three questions every
morning because what they actually needed was a page that was already open. Both
read the same Semantic Model, so the modelling work is not wasted - but the
delivery is, and so is the meeting where it is demonstrated.

Ask it concretely, the way question 2 is asked: "when you have this number, what
happens next - does somebody act on it there and then, or is it something you
check?" A recurring report is a Dashboard, whatever words they used to ask for
it. Anything they want **pushed** to them - mailed, posted to a group - is a
third answer again, and a schedule cannot run anything needing approval.

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

**3a. Each answer names a shape, and you can say so in the room.**

The mapping is mechanical and the same every engagement, so there is no reason
to leave the customer's answer sitting as a fact when it is already a design:

| what they answer | the shape it is |
|---|---|
| a read-only database account, and the audience is internal | `semantic-layer` |
| a read-only database account, and the audience is anonymous | `fixed-query-tools` |
| an HTTP API | `external-api`; `api-oauth` when a token has to be fetched first |
| only a web back office | `browser-operation`, and a different order of cost |
| manuals, FAQs, a site | `knowledge-drive` |
| it has to change something, not only read | `write-path`, split into prepare and execute |
| it has to run on a schedule | `trigger`, and it cannot share a Toolset with anything gated |

`asgard-cli usecase <name>` is each one in full, and `asgard-cli size <shape>`
turns it into a count.

**Say it out loud as you go.** "If the answer is a read-only account we build A;
if it is only the web console it becomes B, and B costs considerably more" turns
an interview into a design conversation, and it makes the cost of the last row
visible while they can still do something about it. It also gives the deck its
right-hand column: every question paired with what answering it produces.

**3b. For each system we will actually connect to, get the coordinates.**

Ask in the meeting, not by email afterwards. Every one of these has turned an
integration from days into weeks by being discovered late:

  - host, port, database or schema, and the account name
  - **is the account read-only?** Ask explicitly. The one offered first usually
    is not, and finding out later means going back for a second credential
  - **is it reachable from outside their network?** Asgard is a hosted cloud
    service - it does not run on the customer's network and cannot be put
    there - so an internal system stays unreachable until they allowlist our
    four outbound addresses or bring us on over a VPN. Ask who can approve a
    firewall change and how long one takes there; it is a ticket and a window
    in most companies, not something the person in the meeting can do that
    afternoon. `asgard-cli wiki operations` has the addresses to hand over in
    the meeting. This is the single most expensive thing to discover in week
    three, and it costs one sentence to ask in week one
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

**3f. Listen for the sentences that are a skill.**

Section 3 asks where the data is. This asks for something the customer will say
in passing and never volunteer, because to them it is not a fact about a system -
it is just how things are:

    "這個代碼的意思是⋯"                a status vocabulary
    "我們內部把 A 和 B 算成同一件事"     a cross-system mapping
    "這個數字要這樣加總"                an aggregation convention
    "那一欄我們只在退貨的時候填"          a field's real meaning

**Any of those is a skill, and the moment to write it down is when they say it.**
Nobody can reconstruct it later from the schema, because it is not in the schema.
An agent without it does not fail visibly - it answers confidently and wrongly,
having interpreted a code that meant something else.

It goes to `common/skills/<name>/SKILL.md`, which is synced to the platform. In
`references/` it is invisible to the running agent, and the failure then looks
like a model ignoring instructions rather than a file in the wrong place.

This is also the row a proposal forgets, because it is knowledge rather than a
system: there is no credential to ask for, so it never comes up in the access
conversation. `asgard-cli usecase skill-layers` is what one looks like at full
size, and what its layers are for.

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

**6b. What is the smallest version they would accept as proof?**

Their answer, not yours. You will have one in mind and it will be the one that
is easiest to build; theirs is the one that gets judged.

    "if it only did ___, would that be worth putting in front of someone?"
    "of everything here, which one would you want to see working first?"
    "what would make you say this is not going to work?"

If they handed over a verification list, read it back and ask which item they
would keep if they could only keep one. That item is the MVP, whatever it costs
to build - a first delivery that skips the item being judged has failed however
fast it shipped.

**Read it back means arrive with a reading.** A customer who wrote a document
listing what they want tested has already answered most of this, and asking them
cold - "so which of your three would you like first?" - hands them our
sequencing problem and reads as though we cannot do all of it. Their document is
the answer; bring your reading of it and ask them to correct it:

    no    "三個情境要先做哪一個?"
    yes   "你們驗證項目裡寫了 X。我們讀下來,情境二最快能證明它,因為 ___。
           這樣對嗎?"

The second takes the same minute and produces a decision instead of a
deliberation. It also surfaces disagreement, which the open version cannot: a
customer correcting your reading tells you something; a customer picking from a
list tells you what was easiest to say.

**So this is a meeting question, not an open-questions row.** It belongs on the
agenda, and it only becomes a tracked row in one case: they handed over nothing
that speaks to it and would not answer when asked. Filing it as a blocker when
their own document answers it puts a question at the top of a list that the
customer can see they already answered, and everything under it inherits that
impression.

Then work out what that one item genuinely needs, and the two filters below turn
the rest into deferred scope rather than open questions.

**7. What is explicitly out of scope?**

Write down what you are NOT building, particularly the things they mentioned in
passing. An unrecorded "we could also..." returns as an assumption three weeks
later, and by then nobody remembers whether it was agreed.

## What they ask us

An interview is not one-directional and the material here has been, until now.
A customer who wrote a test plan usually ends it with a list of things they want
**us** to confirm - account permissions, whether a channel can do X, what we
recommend for their existing system. Those have nowhere to live: the
open-questions file is this engagement's own questions, and
`asgard-cli wiki platform-unknowns` is what no source settles.

They go in `docs/open-questions.md`, in its own section, and the file the
scaffold writes now has one.

**Check each against `asgard-cli wiki platform-unknowns` before answering.** A
surprising share of what a customer asks us is already on that list, because
they ask about the same things every engagement hits - what a given user is
allowed to reach, what gets logged and for how long. When one matches:

    say so, plainly, in writing. "We do not have a confirmed answer to this
    yet and are checking with the platform team" is a real answer and an
    honest one

The failure it avoids is the expensive kind: answering from a reasonable
assumption, having it written into their evaluation, and discovering in week
six that the platform does not do it. Their list is usually also their
acceptance criteria.

**Answer in writing, with a date, and put the answer next to the question.** A
verbal answer in a meeting is not traceable to anything, and the next person
cannot tell what we committed to.

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

## Two filters before a question becomes a row

An interview that ends with twenty-five open questions has not narrowed anything.
It has moved the customer's whole document into a table, and the meeting that
follows spends its time on questions nobody needed answered yet.

Apply both to every question before filing it. Each turns a question into
something other than a blocker - not our problem, or not now - and most questions
are one of the two. What survives is short, and short is what the meeting is
for.

### Filter 0 - is this ours to answer at all?

**Ask first, because it removes the most rows.** We are delivering an agent. We
are not designing the customer's support operation, and an interview that drifts
into how their own systems and teams fit together has stopped being a
requirements interview.

The test is narrow: **does the answer change what we build?**

    ours        what we need FROM them to build it - a credential, an endpoint,
                a network path, a document, an account, a decision only they can
                make about our behaviour
    theirs      how they staff a channel, who maintains a document, how their
                two systems relate to each other, what their people do today

A question about their internal arrangements is not an open question. It is
either something to hand back as a note - "this is worth deciding before you go
live, and it is yours" - or nothing at all.

Two ways this goes wrong, and both look like diligence:

- **Doing their integration analysis for them.** How their channel binds to their
  CRM is their business unless we are the thing doing the binding. Asking it
  makes us look thorough and produces a table nobody uses.
- **Turning an operational precondition into a design question.** "Is a person
  already answering on this account" matters, but it is one line in the handover
  - a thing they must sort out before we attach anything - not a row we track
  and chase.

What survives filter 0 is almost always a small set of the same shapes: a
credential, an endpoint, a network path, a document, an account, and the two
answers only they can give (2b and 6b).

**Those last two are agenda items before they are rows.** Both are answered in
the meeting by a person in the room, so a row for either is a note that the
meeting has not happened yet - and 6b in particular has usually been answered
already, in whatever they handed over. Bring a reading and ask them to correct
it. File a row only if you asked and got nothing.

### Filter 1 - the minimum that proves it works (MVP)

**Ask: what is the smallest thing that proves what THEY said they are testing?**

Not the smallest thing that is easy to build. A customer who hands over a test
plan has already written down what counts as success, and a first delivery that
avoids the item they most want to see has failed, however quickly it shipped.

So read their verification list first, then find the smallest slice that reaches
the hardest item on it. What that slice does not need is not an open question: it
is a line in the request's section 5, Scope, with a note of what would have to be
answered before it comes back.

**What gets cut is a mechanism, not a capability.** That distinction is the whole
filter, and it is the one that gets it wrong in both directions.

    cutting a capability     "we will not read your system yet"
                             fails their test. It is the thing being judged
    cutting a mechanism      "for now the user tells us which record they mean"
                             passes it. The integration is still proved

The pattern that keeps recurring: the hardest question in an engagement is
usually **how the agent knows who it is talking to**, and it is almost always
cuttable, because the user can be asked. A lookup keyed on something the user
types proves the same integration as a lookup keyed on a recognised identity, and
the identity question moves to phase two without the delivery losing anything the
customer is measuring.

What that leaves blocking the first delivery is normally something duller and far
more useful to raise in a meeting - whether the system is reachable from a
cluster at all, and who can grant an account this month.

**Apply the filter per question, not per row.** Some rows are two questions
wearing one sentence, and the filter then defers the half that should have
stayed. "Is there anything in front of the channel" is the recurring one: the
half about identity defers cleanly, while the half about **whether that channel
is already staffed today** does not. Attaching a webhook to an account real
people are already answering on changes their experience on day one, before any
of the deferred machinery exists. That is an operational precondition of the
first delivery, not a phase-two design. Split the row and keep that half.

Also note what "we already have that system" does not tell you. It says the data
exists. It says nothing about a network path, a read replica, or an account.

**The MVP is not a demo.** It runs against their real channel with their real
documents and a real person can use it. A demo on sample data proves nothing and
the questions it defers all come back at once.

### What is left after both

Whatever neither removes. Those are the real questions, and there are
usually two or three:

  - **how a system is reached** - the one that blocks the MVP nearly every time,
    and where "we have that system" means the data exists and nothing more. A
    hosted platform reaching an internal system needs a firewall change only
    they can make, with an approver and a lead time
  - anything the customer must do before we can - provision an account, paste a
    webhook URL back, open a network path
  - anything where two of their answers contradict each other

If a surviving row is not one of those shapes, run it through filter 0 again.
Most of what gets past these filters and still turns out to be noise is a
question about the customer's own arrangements that felt too important to drop.

## Write it down as you go

Not afterwards. A record written after the meeting is a record of what you
remember, and what you remember is the design you were already forming. The
three commands are at the top of this page.

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
  - what they do with each answer is decided - an agent, a dashboard, or both -
    because the product follows from it and every section below assumes one
  - every system has a row, and every row says how it is reached
  - the open questions have been through both filters, so what is left blocks the
    MVP rather than describing everything still unknown
  - the customer has said which single item they would keep, or corrected the
    reading you brought them. It is theirs to settle and cannot be decided for
    them - but arriving without a reading is not neutrality, it is asking them
    to do the work of the meeting
  - every system we will connect to has a section 4 block, with a named
    credential owner and an answer on network reach - and no secret in it
  - unstructured knowledge is either listed or explicitly ruled out
  - every write is marked as a write
  - success is described in terms somebody could check
  - nothing that blocks the work is missing an owner

Not ready is a normal state to be in. A `draft` that names its open questions is
more useful than a `ready` that guessed at them, and `asgard-cli next` will keep
the request in front of you either way.

What usually comes before the task specs is saying it back to them: what we
propose to do, what phase 1 is, and what we are not doing. That is the
`proposal-deck` skill in `.agents/skills/` - it owns choosing the shape as well
as the deck, and it is where the rule about claiming no further than the
evidence goes lives. Do not decide the shape here and write the deck from
memory afterwards.

Then split it into task specs:

    asgard-cli task add "<title>" --request <<.RequestID>> --project <project> --complexity M
