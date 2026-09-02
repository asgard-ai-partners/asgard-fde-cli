This is the interview, and it produces a **request** - not a design, not a task,
not a chart. Read it before the working session with the customer, and keep it
open during one.

## What this stage produces is a file

Everything below is how to think during the interview. This is what to do with
it, and it comes first because the thinking is the part that goes well on its
own. Three commands, run **as the answers arrive** - not at the end, and not
before they start:

    asgard-cli request add "<what they asked for, in their words>"
    asgard-cli request target <<.RequestID>> <project>
    asgard-cli question add "<what blocks it>" --blocks <<.RequestID>> --ask "<who>"

**One request per capability they asked for.** A document with three scenarios
is three requests, because two capabilities in one record cannot be given
different target projects, and the target project is the decision this whole
stage exists to reach.

**And not before the interview.** A customer's own document arriving ahead of
the meeting is normal - their internal approval comes before they will book one -
and it is not a reason to open a request. Only section 1 exists at that point,
and it already exists, in `references/`. The other six - audience, target
project, how each system is reached, writes, success, scope - are what the
interview decides, so a request written first is one copied section and six
TODOs. It shows in the index as progress, it cannot pass any of `request
ready`'s checks, and it puts the customer's words and our translation in one
file while the translation is the request's whole reason to exist.

Before the interview the material has three correct homes and needs no fourth:

    references/              their document, as they wrote it
    docs/open-questions.md   what to ask, and who can answer it
    docs/meeting-notes/      the meeting, and what is taken into it

**An empty `requirements/requests/` before the interview is the right state, not
a gap.** `asgard-cli check` treats it as one.

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
## Read the later stages before this meeting, not after

Every other page here is written to be read when you arrive at it, and that is
the right arrangement for them. **It is wrong for this one**, for a reason that
is about the output rather than the content:

    stage 3 to 9 produce files      a wrong one is edited, re-rendered, reverted
    this stage produces speech      a wrong one is in the customer's notes

You get one interview. At that point you have not read stages 3 to 6, and **half
of what you will say out loud is settled there**. So read them first - at least
`data-sources`, `read-path`, `entry-point` and `knowledge`. An hour before the
meeting is cheaper than a correction after it.

### The five that get said wrong

They live in `asgard-cli brief customer-meeting`, not here, because meetings
happen at every stage and a briefing reachable only from this page is
unreachable to an engagement at stage 5 with a meeting tomorrow.

**Run it before the meeting.** The short version of why: four of the five fail
in the same direction - the intuitive answer undersells the platform or
overstates a limit - so when a customer asks whether something is possible and
the honest-sounding answer is "no" or "not yet", that is the moment to check
rather than to be modest. Being careful and being wrong look identical from
their side.

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

**0. What have they already read or been told?**

One sentence, before describing anything. A customer who read the product site
arrives believing their own staff will build workflows in Odin, carrying a
vocabulary - Basic Function, Template - that maps to nothing in the product as
documented anywhere else, and expecting Mimir to forecast.

`asgard-cli wiki what-they-read` has the three and what is actually true. You
cannot correct a picture you have not seen, and describing the platform over the
top of a different one produces a customer who nods and disagrees later.

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

The mapping is mechanical, and **the walk is where it lives** - one decision per
stage, each with the case that got it wrong:

    the systems and how each is reached      asgard-cli next --stage data-sources
    the read surface, per audience           asgard-cli next --stage read-path
    the entry point, per audience            asgard-cli next --stage entry-point
    where unstructured knowledge goes        asgard-cli next --stage knowledge

Read them **before** the meeting rather than when you arrive at the stage. They
are written as build-time decisions, but every one of them is settled by an
answer the customer gives here, and knowing which answer produces which shape is
what lets you say it out loud:

    "if that is a read-only account we build A; if it is only the web console it
     becomes B, and B costs considerably more than everything else together"

That turns an interview into a design conversation, makes the expensive answer
visible while they can still change it, and gives a discovery deck its
right-hand column - every question paired with what answering it produces.
`asgard-cli size <shape>` turns the shape into a count.

**3b. For each system we will actually connect to, get the coordinates.**

Ask in the meeting, not by email afterwards. Every one of these has turned an
integration from days into weeks by being discovered late:

  - host, port, database or schema, and the account name
  - **is the account read-only?** Ask explicitly. The one offered first usually
    is not, and finding out later means going back for a second credential
  - **is it reachable from outside their network?** Asgard is a hosted cloud
    service and the agent runs in a sandbox the platform starts, in that cloud.
    There is nothing of ours to put on their network. So the ask has one shape:
    **they add Asgard's four outbound addresses to their allowlist.**

    **Ask for that, not for "a VPN, an allowlist or a jump host".** Offering
    options invites their network team to choose one that does not apply, and
    the week it takes to find that out is the week you were trying to save.
    If their policy needs a VPN, that is their side's business about how the
    allowlist gets implemented; what we need from them is unchanged.

    Ask whether it can be done and roughly when - a date changes our plan. **Do
    not ask who approves it**: a name does not change what we build, and filter 0
    below names this exact case. It is a ticket, an approval and a window in most
    companies rather than something done that afternoon, and that is their queue
    to manage, not ours to chase.

    **Ask for the person; do not hand out the addresses.** They go to whoever
    makes the change, once, read fresh from `asgard-cli wiki operations` - not
    into this repository, not onto a slide, not into a thread that gets
    forwarded. They can change and a copy will not, and a stale allowlist is the
    customer's connection dropping. Coordinates in a committed record are the
    thing section 4 already refuses; ours are the same class as theirs. This is the single most expensive thing
    to discover in week three, and it costs one sentence to ask in week one
  - who issues the credential, by name or role. A credential with no owner is
    not a dependency, it is a delay
  - for an API instead of a database: the auth scheme, who holds the client id
    and secret, and the rate limit - the rate limit decides whether a Syncer can
    backfill at all
  - **is there a test environment for this system?** Ask it of every system, not
    only of APIs. It decides what the first delivery can actually do:

        there is one        the test period runs against it, and the whole path
                            is genuinely proved - fields, validation, status
                            codes, all of it
        there is not, read  ask whether they permit reading production data
                            during a test. Some security policies do not, and
                            that is a go/no-go for the first meeting rather
                            than for week three
        there is not, write drafting or a mock, and say so now

    **Do not call it a sandbox in front of anyone.** In this material a sandbox
    is the thing the platform starts to run an agent in - a different subject
    that appears a few paragraphs above this one.

    And when there is one: **production and test schemas differ**, in shape and
    in volume. A query built against a test database is not proved against
    production, which is the same problem as guessing a schema.

**For a system we will write into, that list asks nothing useful.** Host and
port and read-only do not describe creating a record. Ask instead:

  - **what is the token allowed to do?** Not how many credentials they will
    issue and not whose name it sits under - what the one we get may do
  - **what does creating one of these require?** The mandatory fields and the
    validation rules, and ask for the document rather than the answer - see the
    asking ladder above
  - **which field on the record says which of their customers this is for?**
    The one genuinely ours, because the agent fills it. Everything else about
    that record is their system's business

**Their side of that record is not ours to design.** Whether the token is a
service account or sits under a named person, whether their system routes or
counts SLA by creator, how they notify - **do not ask**. It changes nothing we
build, and asking it in front of a customer is designing their permissions for
them. Filter 0 below is the test; this is the case it catches most often,
because a write makes the questions feel responsible.

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

**A credential the customer's own users supply is a different problem again**,
and it comes up whenever the agent acts on behalf of individual people rather
than as one service account: each user's token for a third-party platform has to
be stored and replayed, so it cannot live in `app-secret` and cannot live in
`.env` either.

That is a service with a database, not a chart - one existing deployment holds
them AES-256-GCM sealed in a column, with the key from its own environment,
masked for display, and the whole boundary isolated behind one package that the
build refuses to let other layers import. **If a requirement implies this, say
so early**: it is the point at which the engagement stops being a chart and
needs somewhere to run code. `asgard-cli usecase per-turn-credentials` is the
lighter alternative - the caller supplies the credential each turn and nothing
is stored - and it is worth checking whether that is enough before agreeing to
hold anything.

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

There are two halves to this and the second is easier to miss:

    what they say out loud     write a skill there and then
    a document they hand you   it becomes a skill - it does not stay in references/

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

**And ask whether they need to see where an answer came from.** Citations are
available and are not automatic: the sources arrive on the completion event
inside the message's `template`, but only if the Workflow is built to return them
and the front end reads that field. It is decided when the chart is written, and
retrofitting it means changing the workflow and the front end together - so it
belongs in the request rather than in a later conversation. Regulated industries
and anything replacing a human who cites a manual will want it.

**Retrieval quality depends on how their people ask.** Specific keywords beat
vague ones, one topic per question, context helps. A customer whose staff ask
"tell me everything about X" will judge the knowledge base as bad and will be
describing their questions. Sample questions on the agent are where this gets
taught without anyone reading a guide.

Anything a query DOES answer exactly - counts, prices, stock levels, contact
details - belongs to a query tool, not to a Drive. Putting a number in a Drive
makes the agent paraphrase a figure it should have read.

**4b. Does something have to happen when their system does something?**

"When an order comes in", "when a ticket is escalated", "when the stock drops
below" - that is a **webhook**, not a schedule, and the two get confused because
both run with nobody watching:

    their system calls us when it happens    a webhook. `asgard-cli wiki automation`
    we look on a timer                       a Trigger, and always later than the event

**The question that decides it is whether their system can call out at all.**
Many cannot - an old ERP, a vendor SaaS with no outbound hooks - and then a
schedule is the fallback, with a delay the customer should hear about now rather
than at acceptance.

Ask who can configure that on their side. It is usually a different person from
whoever gives you a database account.

**4c. Do they expect it to send anything outward?**

Mail, SMS, a message into a group. Customers ask for this constantly and it
sounds trivial next to reading a database, so it gets nodded through.

**The platform cannot send mail.** No SMTP, no preset mail toolset, nothing in
the core. So:

    they have an HTTP endpoint that sends mail   we can call it
    they do not                                  it cannot be built yet, and
                                                 that is a question for them

**Do not let it be mocked silently.** A mocked send that returns success and
writes "notified" into a log is worse than nothing - somebody later reads that
log and believes people were told. If a mock is right for a test phase, say now
that every summary will lead with "not actually sent". `asgard-cli wiki
integration` has how one deployment does it.

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

**6c. Do they need to know what it will cost to run?**

Not our fee - the platform's usage billing, which is a separate question and one
a customer with a procurement process will ask before signing anything.

`asgard-cli wiki fehu` has how it is broken down: by **Service** (Platform,
Knowledge Base, Data Insight, Agent Hub, Heimdall) and by **Item** (Project
Usage, Processor Usage, Seat), in units of Units-Days, GB-Days and Times.

Two things worth knowing before answering:

  - **the Workspace is the billing unit.** So how the work splits into projects
    and workspaces has a cost consequence, and the split is decided at stage 2 -
    before anybody has asked this question. Ask it now
  - **a seat is a line item.** "Everyone in the company can use it" is a
    sentence with a price, and the customer usually has not connected the two
  - **only Odin lets them bring their own model.** Sindri and Mimir use the
    platform's models and the LLM cannot be swapped, so "we will use our own
    Claude account" has a different answer per product - and which product this
    is was decided at 2b. A customer with a model contract or a rule about where
    inference happens has to hear it there, not here
  - **a Loader and an Indexer each cost several times a Project, per day**, and
    a Processor is billed per node per day. So a workflow's node count is a
    standing cost, and a design that routes through sub-workflows where one
    prompt would do pays for it every day it exists

If they do not raise it, say the shape of it anyway, once. A cost discovered
after a pilot is the reason a pilot does not convert.

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

Ask for it in the meeting, before you need it. Customers usually have more
written down than they think, and none of it arrives unless asked for.

**Asking has an order too, and it is not the one in question 3.** The two look
alike and are opposite - one is how we reach a system once we have it, the other
is what to ask them for first:

    question 3, how we READ a system   a database  >  an API  >  a screen
    here, what we ASK THEM FOR         docs > source > API spec > the DB > the UI

| ask for | what it gives you |
|---|---|
| **system documentation / operating manual** | best. Fields, validation rules, status codes and the process are all in it, and it explains what things *mean* |
| **the source code** | better than a spec, and people forget to ask. The code is what the system does; a spec is what somebody wrote down about it once. For a system they built themselves this is usually available and usually decisive |
| **the API spec** | a clear contract, and it drifts. Good for shape, weak on business meaning |
| **the database** | the data without the rules. You can see every field and not what any of them means |
| **the back office screen** | last. You can look at it and cannot quote it |

**Source code above a spec is the row that surprises people.** A spec describes
an intention; the code is the behaviour, including the special cases nobody
documented and the field that means two things depending on another field.
Customers rarely offer it and often will hand it over when asked - and nobody
asks.

**For a system they bought, this row does not exist** - skip it and go to the
API spec. The order does not change; there is simply nothing to ask for, and
asking anyway spends a request on it.

So establish which kind it is before working down the list. If you do not know
yet, ask without assuming: *"the API documentation, and if it is something you
built yourselves, the source as well."*

**A document replaces an interrogation, and that is the point of asking first.**
Do not go through fields, validation rules or status codes one at a time in the
meeting: it is slow, and what you get is the version the person remembers.
Meeting time is for what only they can answer - whether the network reaches it,
who issues the account, who approves a write.

**Say why you want it, because it makes them more willing to give it:** their
manual is not background reading for us, it is what the agent will know. A field
dictionary becomes the thing that stops it inventing a status code.

**Do not design their permissions while asking.** "A read-only account for
queries and a separate writable one" is our implementation preference stated as
a request, and it is not always even possible - plenty of systems issue one
account with different rights. Ask what access they can give and what it allows;
let them tell you how many credentials that is.

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

**Apply it by imagining the most specific answer possible.** Not "would this be
useful to know" - assume they answer perfectly, then ask what you would do
differently. "Ming issues it" and "Ming spends five hours a day on it" are both
perfect answers and neither changes anything. That version of the test catches
in one pass what the categories below catch one at a time.

    ours        what we need FROM them to build it - a credential, an endpoint,
                a network path, a document, an account, a decision only they can
                make about our behaviour
    theirs      how they staff a channel, who maintains a document, how their
                two systems relate to each other, what their people do today

**Filter 0 decides what to track, not what to put on a slide.** These are two
different lists and this section has been read as one - an FDE saw "who issues
the read-only account: ask" and put that question on a customer slide, where it
was rejected on sight.

    tracked      knowing who to chase is project management. It belongs in
                 the open-questions row, with the name in the ask column
    on a slide   only questions whose answer changes the design

Everything below is about the first list.

**Names cut both ways, and the line runs between two kinds of person:**

    the person who will hand us the thing      ask. Without a name, a
                                               dependency is just a delay
    the person who authorises them internally  do not ask. Their org chart,
                                               their queue, and it changes
                                               nothing we build

So: who issues the read-only account - yes, we will be chasing them. Who signs
off the firewall change - no. Ask **whether** it can be done and roughly
**when**, because a date changes our plan; a name in their approval chain does
not, and asking for one in front of a customer reads as managing their internal
process. This has reached a slide.

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
usually **how the agent knows who it is talking to**, and it is often cuttable,
because the user can be asked. A lookup keyed on something the user types proves
the same integration as a lookup keyed on a recognised identity, and the identity
question moves to phase two without the delivery losing anything the customer is
measuring.

**Check whether it needs cutting at all, before you cut it.** This is the
paragraph that most often produces a promise we did not have to make, and the
mistake has already been made from it.

An anonymous channel can answer "where is MY order". What it cannot do is let
the **model** choose whose case to look up. Whatever sits in front - a website, a
chat channel, a support desk - is what knows who is talking, and it passes the
identity through server-side on every turn. **LINE's webhook carries a userId.**
So if the customer's system has the binding stored, per-customer lookup is in the
first delivery and there is nothing to defer.

    is there a layer in front that knows who is speaking?
      yes  -> keep it. Cutting it gives away something you had
      no   -> now it is genuinely cuttable, and that is a question for
              them rather than a design to work around

Getting this backwards costs more than a deferred feature: it is telling a
customer their channel cannot recognise their own customers, on a slide, when it
can. `asgard-cli next --stage read-path` has the mechanism and
`asgard-cli usecase per-turn-credentials` has the shape - read one of them before
promising anything of the form 「查我的⋯」.

**And when it is kept, it brings an acceptance criterion with it**, which belongs
in section 6 now rather than being discovered at verification: a query that
forgets to filter on the injected identity **fails on the anonymous path only,
which is the path nobody tests**.

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
