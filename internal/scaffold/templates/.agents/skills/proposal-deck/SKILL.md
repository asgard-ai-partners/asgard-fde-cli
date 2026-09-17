---
name: proposal-deck
description: Use when deciding what to propose to the customer, or building any deck they will see - a discovery deck taken while the questions are still open, a proposal, a scope review, a phase kick-off, or a handover after the work lands. Covers how the shape is chosen and what phase 1 is, where a deck lives in this repo, which material it is allowed to be built from, the order a proposal argues in, how to say it in the customer's own words without claiming further than the evidence goes, what must never appear on a customer's screen, and what a link on a slide has to point at. The outline that records what each page rests on is `outline.md` beside it, and the design language and the slide contract are in references/.
---

# Proposal decks

A deck is the only thing in this repository the customer reads. Everything else
records what we think; a deck records what they were told, and they agree to
scope on the strength of it.

That is the whole reason this skill exists. A deck assembled from the chart is
accurate and useless. A deck assembled from an ambition is readable and turns
into a broken promise. What works is a deck assembled from **their own material**,
saying back to them what they said, and then one paragraph of what we will do
about it.

## Where it goes

A deck belongs to the meeting it was presented at, so it is filed as one. There
is no new directory for decks: a meeting note that has attachments becomes a
directory instead of a file.

    docs/meeting-notes/YYYY-MM-DD-<topic>/
      README.md    the meeting note - who was there, what they said
      deck.md      what we presented
      deck.pdf     the file they were given
      assets/

The meeting note is what makes the deck readable later. Without it nobody can
tell whether a deck is what was proposed or what was agreed, and those are
different things - what was agreed goes to `docs/decisions/`.

Add four lines to the note, because a deck is a claim about where its content
came from:

    - 對象:營運主管 + IT 窗口(4 人)
    - 材料來源:references/erp-manual/、requirements/requests/REQ-003-*.md
    - 當天結果:一階段範圍確認;退貨流程延到二階段
    - 講出去之後改過嗎:沒有。要改就開新的一場日期。

**A presented deck is not edited afterwards.** The customer's copy does not
change with ours, and a repository version that differs from the one in their
hands is worse than no copy at all. Changed your mind: new date, new directory.

## Step 1 - find the material, and refuse to invent any

A proposal is built from what the customer gave us and what they said, in this
order:

| where | what it gives you |
|---|---|
| `references/` | their own manuals, schema dumps, field dictionaries, screenshots |
| `docs/meeting-notes/` | what they said, before we tidied it |
| `requirements/requests/` | what they asked for, in their words |
| `docs/open-questions.md` | what is still unknown - and **this one is also an output**, see below |

**`docs/open-questions.md` is both the source and the product.** Working a deck
through with somebody rewrites the questions: one engagement's revisions
reworded about half of them, retired three and found six more, and none of it
went back because nothing said it should.

    reworded a question on a slide   change the row, in the same edit
    dropped one                      retire the row, with why
    found one                        add it, with who can answer

**In the same edit, not afterwards.** The FDE who reported this intended to
reconcile at the end; after many rounds the two had diverged far enough that
the reconciliation kept being deferred.

**Why this one matters more than ordinary staleness:** `asgard-cli question`
prints this file back, and it is the first thing whoever picks the repository up
next is told to read. A stale one sends them into a meeting with questions
already abandoned - and worse, it preserves a judgement that has since been
overturned, still argued convincingly. In that engagement a security reasoning
the deck had corrected was sitting in the file, written well.

`asgard-cli check` warns when a file under `docs/meeting-notes/` is dated after
**the newest date inside the questions' own rows** - a raised date, or the date
an answer was written beside one. Deliberately not the questions file's
timestamp, which any edit silences - including one that touches only the prose.
**So reformatting the file does not clear this warning, and answering a
question does.**

**Not** `docs/spec/` and **not** `projects/*/chart/`. Those describe our
implementation. A customer looking at a slide made from a chart sees a diagram of
something they did not ask about.

If a claim you want to make is in none of the four, it is not yet a claim. Two
ways out: ask, or put it on the open-questions slide. Do not smooth it over -
this repository records three decisions that were made confidently, built, and
reversed after contact with reality. Any of the three, written into a proposal,
would have been a commitment we could not keep.

Every number carries where it came from:

- a number the customer gave us - say so, and say when they gave it
- a number we measured - say what was measured and when
- a number nobody has - leave it out, or put a question on the slide

**And a number that passes all three can still be in the wrong place. A title
names things; it does not count them.** 「六個指令」 and
「12 個 CR,14 次部署,4 個坑」 are both accurate, and both spend the three
seconds a title gets without naming a subject: a reader who has not seen the
slide learns nothing, and one who has learns nothing new. What they needed was
「init、project add、add、push、tag、approve」. The quantity goes in the body,
where it is read rather than counted.

The test takes no judgement and no context - **does the title contain a
quantity** - which is why it is worth having beside the ones that need a reader.
**The exception is a discovery deck**, whose titles are the customer's own
headings unedited: if they counted, keep their count, because the title is
theirs to have written. Step 5.

## Step 2 - decide what we are proposing

A deck cannot be written from open questions alone. Between what they said and
what goes on slide 4 there is a decision - **what shape we are proposing** - and
it is made here, deliberately, rather than arrived at while typesetting. It is
the step most often skipped, because the material from step 1 reads like it
already contains an answer. It does not: it contains a problem.

These rules are the interview's, restated here because a deck gets built weeks
after the interview and often by somebody else. `asgard-cli guide
requirements` is where they live in full, with the worked examples.

**The shape follows the audience, not the capability.**

    internal, authenticated callers   -> the platform's agent hub, semantic layers
    public, anonymous visitors        -> your own BotProvider, fixed query tools

The two paths share neither an entry point nor a read path. A proposal covering
both audiences is two proposals, or one with two phases - never one shape
stretched over both. This repository records three decisions that were made the
other way, built, and reversed, and all three were reversed by the audience.

**Take a shape that already exists.** `../asgard-platform/usecase/` holds the ones taken
from deployments in production, and each says what it costs to assemble. A
proposal built on one of them can be estimated, because somebody has already
paid for it. A proposal built on a shape nobody has assembled is an estimate of
a guess.

**For what one looks like end to end, take a worked one rather than inventing a
scenario.** `../asgard-platform/wiki/case-studies.md` has one event described from three
angles, which is where the division of work between the three products stops
being abstract - and a scenario written from it is one somebody has already
checked, rather than one this deck is the first test of.

**If they ask to see it working and we have none of their systems**, that is a
shape too, and it has been built: `../asgard-platform/usecase/demo-generation.md` - no
data, no credentials, and still a real deployment rather than a slide. Decide
whether the deck is promising one before writing a slide that implies it.

**First, is it an agent at all?** The whole vocabulary below is agent-shaped, and
so is most of what this tool carries, which is exactly why this question gets
skipped. Something they want to **watch** - the same numbers every morning, a
figure trending the wrong way - is a **Mimir dashboard**, not an agent. Something
they want to **ask** in the moment is an agent. Many customers want both, and
they are two deliveries over one Semantic Model, not one delivery.

A proposal that answers "we want to see stock across all our channels" with an
agent has answered a question they did not ask. `../asgard-platform/wiki/mimir.md` is the
product; `../asgard-platform/wiki/product-suite.md` is the fork.

Then, what their answers turn into - and it is not one row per system:

| what they have | what we build |
|---|---|
| numbers somebody checks on a schedule | a **Dashboard** in Mimir, over the model below |
| a database we can read | a **Semantic Model** - the agent composes its own queries |
| a database, but only a few answers should be reachable | **fixed query tools** |
| an HTTP API | a **Workflow** wrapped as an **MCP Server** |
| documents, manuals, FAQs, a support site | a **Drive** with a Context Index |
| only a web console | **browser operation** - costs more than everything else combined |
| **systems whose concepts do not line up with each other** | a **Skill** |

**That last row is the one that gets left out of every proposal**, because it is
not a system to connect - it is knowledge about the systems, and it has no
credential to ask for, so it never comes up in the access conversation.

It is what you need whenever the customer's own systems disagree: four
marketplaces with four sets of order-status codes, a SKU that is written three
ways, "shipped" meaning something different in the WMS than on the channel, a
returns flow with vocabulary only their staff knows. **Somebody has to write
down how those correspond**, or the agent guesses, and it guesses plausibly -
which is worse than failing, because the answer looks right.

Two things about it worth saying out loud to the customer, because both are
counter-intuitive:

  - **This is the part that needs them most, and it is not an IT task.** A
    credential comes from their network admin; this comes from whoever actually
    knows that channel B's "processing" is channel A's "paid". It is an hour of
    an operations person's time, and there is no substitute for it
  - **It is the cheapest thing on the list and the one that decides whether the
    rest is usable.** Nothing to procure, nothing to provision - a document

It goes to `assets/skills/<name>/SKILL.md` and is synced to the platform. In
`references/` it is invisible to the running agent, and the failure then looks
like a model ignoring instructions rather than a file in the wrong place -
`../asgard-platform/guide/requirements.md` has the full rule.

An integration across several channels that proposes only the connections has
proposed half of it.

The cost ladder for reaching a system is worth putting in front of the customer
in words, because it is what decides the timeline:

    a database we can read   >   an API   >   a screen a person clicks

The last one is browser operation, and it costs more than the other two
combined - not because it is technically hard, but because somebody has to
document every page and every dialog the menu does not show first.

**Phase 1 is the smallest thing that proves what THEY said they are testing.**
Not the smallest thing that is easy to build. If they handed over a verification
list, phase 1 has to reach the hardest item on it; a first delivery that avoids
the thing being judged has failed however fast it shipped. What phase 1 cuts is
a **mechanism, not a capability** - "for now the user tells us which record they
mean" is a cut, "we will not read your system yet" is a failure.

**A document they give us is a deliverable line, not an input.** The commonest
omission in the right-hand column is the skill: "we take your RMA manual and turn
its fields, validation rules and status codes into something the agent knows" is
a concrete thing we do with what we asked for, and it is the answer to "why do
you need our documentation". Every capability reading a system the customer
documented has one.

**Decide what we are not doing, in the same sitting.** Not as an afterthought at
slide 7: a scope line written after the shape is chosen is a line someone
remembers, and a line written while choosing is a line with a reason attached.
Each one gets a note of what would have to be answered before it comes back.

Two things this step produces before anything is typeset:

  - a `docs/decisions/YYYY-MM-DD-<topic>.md` for the shape, and why the
    alternative was not taken. The deck states the choice; the decision record
    is where the reasoning survives being presented
  - the request's own scope section, updated - what is in phase 1, what is
    deferred, and what is out

If you cannot write the decision record, the shape is not decided yet - and a
**proposal** built on it would be a slide-shaped version of the same
uncertainty. That is not the same as having nothing to say. Run this step
conditionally instead - "given a read-only account and a route to it, this
becomes X" - and take a discovery deck, which is step 5's first shape. What you
must not do is state the conditional shape as a decided one.

## Step 3 - the order a proposal argues in

Nine slides, in this order, is a complete proposal. Longer decks are usually a
short one with the same slide repeated. **This is the proposal's order.** A
discovery deck - the one taken while the questions are still open - has its own,
in step 5; forcing this one when the facts are missing is what produces slides
4, 5 and 6 full of a shape nobody has confirmed.

| # | slide | the job it does |
|---|---|---|
| 1 | cover | who this is for, and the date |
| 2 | what you told us | their words, close to verbatim. Earns the rest |
| 3 | what it costs today | the number they gave us, attributed to them |
| 4 | what we propose | one sentence a non-technical reader repeats correctly |
| 5 | what it looks like to use | the interaction, not the architecture |
| 6 | what it reads | which systems, and who can see what |
| 7 | what is NOT in scope | the slide that prevents the argument in month three |
| 8 | how we get there | phases with dates, and what we need from them |
| 9 | what we still do not know | **the customer's unknowns only** - see below |

Slide 2 is the one that makes the deck work, and the one most often skipped.
A customer who hears their own problem stated back accurately will accept a
rough solution; one who hears a polished solution to a problem they do not
recognise will argue with every slide after it.

Slide 7 is the one that saves the engagement. Scope stated narrowly is a proposal
that can be delivered; scope left open is a proposal that grows until it fails.

**Slide 9 holds one kind of unknown, and the two kinds are opposite.**

| whose unknown | on the slide | why |
|---|---|---|
| **theirs** - how their CRM is read, which network their devices are on, whether a channel has an API | **yes**, and it is most of the deck | only they can answer. Listing it asks for their help |
| **ours** - whether the platform can scope what a user may reach, what it logs and for how long | **no** | it is our product. "We do not know what we can do" is not honesty, it is not having prepared |

`../asgard-platform/wiki/platform-unknowns.md` is the second kind, and its own instruction
is **ask the platform team when a requirement touches one** - before the meeting.
Printing that list onto a slide does the opposite of what the page says, and
announces that we have not read our own product.

If an answer has not come back in time, say it in the room, once, and put the
written reply in the follow-up. It does not get a slide.

So the closing slide is **what each side does next**: what we reply on and by
when, which accounts and documents they provide, and the date you meet again.
Both sides have an action on it. That is a different page from a confession.

## Step 4 - say it in their words, and no further than the evidence goes

Two failures, and they pull in opposite directions, which is why they are one
step. A deck written to avoid the first usually commits the second.

### Speak the way they do

**The test:** can someone who was not in the room repeat the sentence to a
colleague and get it right? If the sentence needs one of our nouns to survive,
it has not been translated yet.

    no    Workflow 會呼叫 Toolset 查詢 CRM,再由 Managed Agent 組出回覆
    yes   客戶在 LINE 問訂單到哪了,它去 CRM 查,查到就回,查不到就說查不到

The second sentence is longer and worse-sounding and it is the right one, because
the customer can check it. The first cannot be checked by anyone who does not
work here, and a claim nobody can check is not a claim - it is a mood.

Three habits do most of the work:

  - **name the person, not the component.** Who is at the keyboard, what they
    typed, what came back. Slide 5 is an interaction, not an architecture
  - **keep their vocabulary, even when ours is more precise.** If they say 報修
    單, the deck says 報修單, not "RMA ticket entity". Correcting a customer's
    own word for their own thing reads as not having listened
  - **say what it does when it fails.** A deck that only describes the happy path
    is the single strongest signal that nobody has built it yet. "查不到就說查
    不到" is worth a line

### Claim no further than the evidence goes

Overstating is not a tone problem. It is a delivery problem that arrives eight
weeks late, at the acceptance meeting, where the deck is read back to you.

Four sources of it, in the order they actually happen:

**1. A capability the platform does not have.** The list is not a matter of
judgement - it is written down:

    `../asgard-platform/wiki/platform-unknowns.md`

Check every claim on slides 4, 5 and 6 against it. Anything that appears there
belongs on **slide 9, phrased as a question**, and never on 4 or 6. The two that
catch people are the two customers most often put on their own verification
list, because they read as table stakes: **限制不同使用者能查到什麼**, and
**操作紀錄可稽核**. Neither is settled. Writing them into a proposal as existing
features is the most expensive sentence in this document.

**2. A capability that belongs to the layer in front of us.** Handoff to a human,
pausing the AI while a person replies, resuming afterwards, counting how many
questions one user has asked in a day - **no CR carries any of these.** They
belong to whatever owns the conversation: a support desk, a helpdesk product,
their own relay. With a website that layer exists; with LINE it usually does not.

So a proposal that promises 轉真人 with nothing in the middle is a promise
nobody can keep. Say plainly which layer would have to own it, and make it the
customer's choice - put a desk in front, or drop the requirement. That sentence
is uncomfortable once, in a meeting. The alternative is uncomfortable for the
rest of the engagement.

**3. A number nobody measured.** Covered in step 1, and it is the easiest to
catch: every figure is theirs, ours-and-measured, or absent.

**4. A phase 8 we could not start on Monday.** Read slide 8 as though they said
yes today. Anything resting on a credential nobody has issued, a firewall change
nobody has approved, or a system nobody has connected to is not a phase - it is
a dependency, and it goes on slide 8 as **what we need from them**, with an owner
and a date, or on slide 9.

### Say the same thing every time

Banning implementation nouns without offering replacements means every FDE
invents their own, and the same customer sees two names for one thing across two
documents. These are the replacements. **Use them rather than a fresh
paraphrase**, and `asgard-cli size <shape>` produces the same wording from a
count - the plain reading is its second output, and it is not optional.

| what it is | what the customer is told |
|---|---|
| a read-only Toolset with four query Workflows | 四個查詢 / four kinds of question it can answer |
| a SkillSet carrying status vocabulary | 一份狀態對照 / a translation of your own codes |
| a SkillSet built from their manual | 手冊裡的欄位與規則會變成 AI 知道的東西 / the manual becomes something it knows |
| SourceSet + Syncer + contextIndex | 產品知識庫 / it reads from your own documents |
| a tool with `requestConsent: true` | 需要人工確認的動作 / an action that stops for a person |
| BotProvider + Workflow + SandboxBlueprint | 對外客服入口 / one way in for people outside |
| an Agent mounting a SemanticLayer | 一個能自己查資料的專員 / it works out its own query |
| a Trigger | 定期自動跑 / it runs on a schedule and only reports |
| a SemanticLayer over a database | 讀得到那套系統 / it can read that system |
| the outbound-IP allowlist request | 把這四個位址加進防火牆白名單 / add these four addresses to the firewall's allowed list. **The sentence, not the values** |
| browser operation | 沒有介接管道,要照著人的操作做 / no way in but the screen |

The browser-operation row is the one to say plainly rather than smooth over. It
is the most expensive answer on the list and the customer is the only person who
can change it, by finding an API.

The allowlist row has a rule of its own: **ask their network team for the allowlist, not for "a VPN
or an allowlist or a jump host".** Asgard is hosted and the agent runs in a
sandbox in that cloud, so there is nothing of ours to place on their network and
the other two options do not apply. A slide offering three invites their network
team to pick the wrong one, and that is discovered a week later.
`../asgard-platform/wiki/operations.md` has the addresses; hand them over in the meeting.

**Two names for one thing is worse than a clumsy name.** If a phrase here is
wrong for a customer's industry, change it once and use the changed one
everywhere, including in the meeting notes.

### Who is pointing at whom - a rule for Chinese decks only

**Do not write 「你們」.** A deck that says it fourteen times has stopped being a
shared document and become one side addressing the other.

This does not transfer from English, and that is why it is easy to miss. English
cannot avoid "you", so it carries no weight and nobody notices it. **Chinese can
drop the subject entirely**, so writing 「你們」 is an active choice, and the
reader feels the choice being made - as a finger pointed across the table rather
than two people reading the same page.

Three rewrites, none of which needs added politeness:

| instead of | write | how |
|---|---|---|
| 你們現在的日報怎麼做 | 現在的日報怎麼做 | drop the subject |
| **你們文件寫的是** | **測試計畫寫的是** | **make their document the subject, not them** |
| 哪些數值算異常由你們定義 | 異常門檻可自行定義 | state it as a capability |

**Every row of the vocabulary table in step 4 was written this way**, and three
of them had to be corrected after this rule existed - which is the evidence that
it is not obvious. 「請你們把我們的四個位址加進白名單」 became 「把這四個位址加
進防火牆白名單」: same request, no finger.

**The second is the one to reach for.** It keeps the whole point of quoting them
- this is what you said - and removes the pointing. It is also the only one that
is correct for the rest of the room: a meeting has their IT lead and their
network admin in it, and 「你們」 addresses them for a document they did not
write.

The worst case is the one that looks most respectful: 「你們文件寫的是⋯」 reads
as raising an old account rather than as citing a source.

### The register

Confidence in a proposal comes from specificity, not from adjectives. 全面、
智慧化、無縫、大幅提升 say nothing and cost nothing to write, which is exactly
why a reader discounts them. One checkable sentence outranks a paragraph of them:

    no    全面提升客服效率,智慧化整合各系統
    yes   客服現在要開四個系統才答得出一張報修單的狀態。第一階段目標是在 LINE
          裡問一句就回答得出來,前提是 CRM 給得到唯讀帳號。

A small promise that lands beats a complete-sounding one that does not. The
customer is going to test this - that is what the test plan they sent is for.

### The AI tells on itself, and the register rule does not catch it

The section above catches empty adjectives. It does not catch **shapes** - the
sentence patterns and section endings that mark a Chinese document as
machine-written while every individual word survives inspection. A customer who
reads vendor decks recognises those in about four seconds.

**Load `.agents/skills/plain-chinese/SKILL.md`.** It owns every rule of that
kind, for this deck and for everything else this engagement writes in 繁體中文,
and it says which of them change on a slide - including the one that inverts,
which will damage the deck if you meet it anywhere else first.

**When: after the deck reads correctly end to end, not while drafting.** You will
write these patterns regardless - they are what the model reaches for - and
hunting them mid-draft costs the argument, which is the thing that actually
matters. The pass is one question, asked page by page: 這一頁哪裡看得出來是
機器寫的?

## Step 5 - what to show, and what must never be shown

### Three kinds of deck, and only one of them is a proposal

| | discovery | proposal | handover |
|---|---|---|---|
| when | **the questions are still open** | before the decision | after it - training, a readout, a kick-off with their IT |
| audience | whoever can answer them | whoever decides | whoever will operate it |
| what it asks for | answers and access | a yes | nothing - it explains |
| may show a console screen | no | no | **yes - it is the point** |
| may name platform parts | no | no | the ones they will click |
| built from | open questions, and step 2 run **conditionally** | steps 1-3 | the same, plus `../asgard-platform/wiki/setup-path.md` |

Ask which one you are making before writing a slide. "Show them how the agent
gets set up" in a proposal is a request for slide 5 - the interaction - not for
the console. In a handover it is the whole deck.

### The discovery deck, because it is the one that gets made wrong

A first meeting usually happens with most of the interview still unanswered, and
the instinct is to force the nine-slide proposal anyway. Slides 4, 5 and 6 then
get filled with a shape nobody has the facts for - which the pre-send checklist
correctly rejects, and the deck collapses back into a list of questions.

That is the wrong conclusion. A list of questions is a bad meeting: the customer
has to guess why each one matters, and the ones that sound like bureaucracy get
deferred. **Pair every question with what it unlocks**, and the same meeting asks
for access rather than for patience:

    cover
    capability A          their heading and their description. Asks nothing
      A, sub-topic 1      the question | what it produces, one page
      A, sub-topic 2      same, one page per sub-heading of theirs
    capability B          and so on
    close                 see below - one half of this is general, one is not

**How to close.** The general half: **do not close on our own unknowns.** Anything
we cannot answer about our own product is asked of the platform team before the
meeting and answered aloud, never printed - that is the same rule as elsewhere on
this page.

What to close on instead **depends on the engagement, and the one worked example
is not general.** That deck closed on what would be delivered that round, which
worked because there were three concrete things to name. An engagement whose
round delivers "an agent" would produce a thin page, and a thin closing page is
worse than none. If there are concrete deliverables, close there. If not, close
on what each side does next, which at least has an action on both sides.

**How much this one is worth, stated exactly**, because the rest of this page is
worth more and a reader cannot tell them apart otherwise: the deck that worked
closed on deliverables, and **that page's content was specified by the reviewer,
not arrived at.** So it is not one engagement's judgement - it is one person's
one instruction, on one deck, with no comparison against any other ending. The
agent that built it has since said so unprompted. Weigh it a level below
everything else here.

**Three page kinds and no others.** The deck that worked had cover, context,
subject and close - and every attempt to invent a fifth kind, or to group
subjects under headings, was removed. Eight different grouping labels were tried
across many pages and all were deleted.

**The pairing is the shape, and both halves are on one page, side by side.**

This page previously said they were separate slides. **That was wrong and it was
tried**: the deck was split and the reaction was that the split version
was worse. The reason is worth keeping - **two slides separate what you are
asking for from what it buys, and the trade disappears.** A customer looking at a
full page of requests has finished counting the cost before turning to the
return. Side by side, both arrive at once.

    left column    the question, as a plain sentence in a heading
    right column   產出 - what that answer produces

There is a working skeleton beside this file - `discovery-deck.html`, eight
pages, every customer noun neutralised and **the question wording left exactly as
it ended up**. Replace the content; do not redesign the page shapes.

The second of each pair is step 2 run conditionally, and it is the whole reason
this deck works. **State the condition in the same sentence as the capability** - "given a
read-only account on the CRM and a route to it, this becomes X" - so the customer
sees the access request and the thing they get for it as one trade, which is what
it is. `../asgard-platform/usecase/` is where those shapes come from, so what you promise
is something somebody has already built.

Two things this deck must still refuse, and they are the ones the shape tempts:

  - **A conditional is not a promise, and it is not licence to skip step 4.**
    "Given an API, we can integrate all five channels" is a claim about five
    APIs nobody has seen. Name the condition per system, not once for the slide
  - **A question the platform cannot answer stays a question.**
    `../asgard-platform/wiki/platform-unknowns.md` does not become answerable by being
    written as a conditional

What it does not have is a phase plan with dates. That is the proposal, and it
comes after the answers.

### If you are the one reviewing this

The corrections in this section were all caught by a reviewer, every time, and
that is the only part of this with an unbroken record. It cost twenty rounds and
by the fifth the reviewer was asking how many more pages this would take.

Two things make that cheaper, and both are yours rather than the writer's:

  - **say which of the six it is**, not just that the line is wrong. "That is a
    question whose answer changes nothing" transfers; "take that out" has to be
    rediscovered on the next page
  - **watch the next page for the overshoot.** A correction lands as a
    replacement rather than an addition, so the page immediately after "describe
    the scene concretely" is where scene-describing appears in a place it does
    not belong. You are the only one positioned to see that, because from the
    inside it feels like compliance

### The correction count tracks how much of the page you wrote

The single most predictive thing observed across twenty rounds on one deck:

    the customer's own words          zero corrections
    question sentences I wrote        three to five rounds each
    grouping labels I invented        eight rounds, then deleted entirely

**The context pages needed no correction at all, and not one word on them was
ours.** Their heading, their description, their own flow diagram, their own list
of what they wanted tested. The only contribution was the frame.

**That rule has a precondition nobody wrote down, and it bites.** Those pages
worked because that customer's document was thick and carried a diagram they had
drawn. Against a document with two sentences and three sub-items, the same page
fills a third of the sheet - and **the predictable next move is to fill the
space, with the only material available, which is ours.** The rule is then
defeated by its own layout consequence.

**Leave it thin.** The emptiness is true and worth saying in the room: they have
not worked that capability out either, and a page that is thin because their
thinking is thin has told everybody something. If they drew a flow, use it. If
they did not, do not draw one for them here - that is the one thing this page
must not contain.

So when a page is being reworked for the third time, the question is not how to
word it better. It is **how much of it is ours, and whether that part needs to
exist.**

One thing we write that also never changed: **the last line of the outcome
column, saying what happens when it fails.** "It says it cannot find it rather
than guessing one." "It says it cannot reach the system rather than giving you
yesterday's number." Those survived untouched, and the likely reason is that
failure behaviour has one correct answer while a capability can be phrased twenty
ways - **so every phrasing of a capability invites another round, and the failure
line does not.**

### Late rounds catch a different kind of thing

The six below are early-round errors - something added that should not be there.
The last few rounds are the opposite: **things that are correct and still do not
belong.**

    "this question decides everything after it"   true. But it is emphasis
    "without this, nothing here is possible"      true. Adding drama to a fact
    "have they registered on the account?"        a correct restatement of the
                                                  question already above it
    eight grouping labels                         each accurate for its group

**The early test is "is this right". The late test is "does this make the page
heavier without making it clearer".** Emphasis, restatement and categorisation
are all true and all surplus, and they are harder to see precisely because
nothing about them is wrong.

### Six ways a question page goes wrong

From one deck that took about twenty rounds of correction. **They share a root,
and the root is worth more than the list:** every one of them was writing
something that *looks like having done the homework* rather than something whose
answer changes what happens. The first is short and each line can say why it is
there; the second is long and reassuring.

**How solid each of these is.** They came from one deck and one FDE
reconstructing his own decisions, and they are not equally established. Where a
line below is one recollection or a guess, it says so. **Treat an untested one as
a suggestion and not as a rule** - and if you find it wrong in practice, that is
the more useful outcome.

**0a. Every outcome ends with what happens when it fails.**

First because it is the cheapest and the most certain. "It says it cannot find
that number rather than guessing one." "It says it cannot reach the system rather
than giving you yesterday's figure."

**Those lines were written once and never touched**, through twenty rounds that
rewrote half of everything else - because **failure behaviour has one correct
answer and a capability has twenty phrasings.** A rule with no room for choice
cannot be worded wrong, so it costs one line and never comes back.

It is also the most persuasive thing on the page. In a support scenario "it says
it cannot find it, it does not invent one" carries more weight with a customer
than any description of what it can do.

**0. A slide is not an agenda, and it is not a tracking list.**

The rule the other six are downstream of, and the one an FDE said would actually
have stopped him where a test would not:

> **"Who is responsible for this?" is right to ask in the room and wrong to
> print.** Asking it is professional - it says you intend to follow through, and
> it tells you who to chase afterwards. It is not a bad habit. It is a good habit
> on the wrong carrier.

    ask aloud     who issues the account, who maintains it, who does it today
    put on a page only what changes what we build

**The test is the carrier, not the content.** Is this answer somebody we chase
after the meeting, or something we build? The first is spoken; only the second
earns a page.

*(Established. Six occurrences in one deck, and the FDE identified this as the
rule that would have stopped him where a correctness test would not.)*

This matters more than it sounds because **it does not require deciding the
question is wrong.** Half the times this went wrong, the FDE was copying an
instruction from this material rather than making a judgement - and a test you
have to remember to apply does not fire when you do not think you are choosing.
A rule about where a right thing goes does fire, because it does not ask you to
overrule anything.

**0b. What a question that works looks like.**

Everything else here is a counter-example. This one survived twenty rounds
untouched:

> 現在要看各平台庫存,是一個一個後台登,還是有一個地方全看得到?
> *(To see stock across your channels today, do you log into each back office,
> or is there one place that shows all of it?)*

Three things, and the third is what makes it worth a page:

  - **it asks what they do now**, not what they have. People answer their own
    behaviour immediately and accurately; they answer "do you have an OMS" with
    "no", because their name for it is not ours
  - **it offers two concrete possibilities**, so the answer arrives in one breath
    rather than after a pause
  - **the two answers fork the design.** One consolidating system is one
    integration; separate back offices are six

**The counter-example is what makes this usable**, because the rule alone still
produces the wrong one:

    no    有沒有 OMS 或電商中台?
          do you have an OMS or a middleware layer?
    yes   現在要看各平台庫存,是一個一個後台登,還是有一個地方全看得到?

The first collects a "no" - not because they have nothing, but because their
name for it is not ours.

*(Established: this exact sentence went through twenty rounds of correction
without being touched.)*

**1. Asking what changes nothing.** Six of the twenty. Who maintains the
documents, who approves the firewall change, who issues the account, whose name
the ticket goes under, who does the daily report and how long it takes.

**The two columns will not be the same length, and should not be.** The left is
what they have to give; the right is what they get for it. There is no reason
those are the same size, and a page with three lines on the left and five on the
right means this item needs little, not that something is missing.

**Two columns create an obligation to align that nothing else on a page does.**
That much is a property of the layout rather than of one deck, and it is why this
sentence exists at all.

*(The observation that a three-line column beside a five-line one reads as
unfinished is one person's account of one deck. If it does not feel that way to
you, the paragraph below still stands and this part does not.)*

Two things follow, and the second is the sharper one:

  - **if you are adding a line and cannot say what its answer changes, the page
    is finished, not short**
  - *(One case, from recollection. If a short left column does not in fact feel
    unfinished to you, ignore this.)* **the trigger is a short column plus an
    available cheap truth.** The pages
    that were not padded had short left columns too - one had a single line. The
    difference was that no cheap sentence existed for them. "Who is responsible"
    is the cheapest true sentence available about any system: always true, always
    sensible, and it sounds like concern

**The test below is for reviewing, not for writing.** It fires when you know you
are making a choice - which was two of the six times, both while adding a line.
The other four were copying or complying, and nothing you have to remember to run
fires then. Use it when re-reading a page, and see the review section above.

**The test: assume they give you the most specific answer possible, and ask what
you would do differently.** "Ming issues it" and "Ming spends five hours a
day on it" are perfect answers that change nothing. **Three of the six were copied out of this material**, not reasoned into
existence - from filter 0 saying to ask who issues an account, from an
instruction to get the name of whoever approves a firewall change, and from
the LINE integration needing an owner on their side. The first and the third are
correct **for tracking**; the second is not, and
`../asgard-platform/wiki/operations.md` refuses it outright - what it tracks is
whether the allowlist can be changed and whether it has been, never who signs.
None is slide content, and each now says so where it stands, because somebody
copying reads one place and not the canonical one.

**And a third came from over-correcting.** Told a minute earlier to describe the
scene concretely, the FDE added "whose name does this ticket go under" - which in
many ticket systems genuinely is a design consequence, and therefore felt exactly
like the scene-thinking that had just been asked for. That one is the hardest to
catch, because the self-image at the time is "I am fixing this".

**What helps is one habit**, and it works because it fires while you are not
defending anything:

> **After acting on a correction, re-run the check you had before it - not the
> correction.**

The failure is that a correction *replaces* the check rather than joining it.
"Am I describing the scene? Yes" ran; "does each line's answer change anything?"
never ran again. Re-running the older one takes a moment and happens before
anything is shown to anybody.

*(Untested. Nobody has run this habit deliberately; it is reconstructed from one
incident.)*

**Do not expect it to prevent this.** It reduces rounds. The reviewer caught all
six of these, every time, so that path works - at the cost of patience, and by
the fifth the reviewer was asking how many more pages this would take. Judge
this habit by whether the count comes down, not by whether the error stops.

**2. Asking, then supplying the consequence yourself.** Five of the twenty:

    is there a test environment?   ...and if not, may we read production?
    is there an existing binding?  ...if not, they can give us an order number
    is there a controller?         ...if not, this costs weeks more

**Ask the customer, in the meeting, and stop talking.** The consequence is theirs to state. Supplying it rehearses
their answer, and every one of those additions invented a branch that does not
exist - "no test environment" does not mean no integration.

**3. Inventing groupings, then bending content to fit them.** Eight different
group labels across many pages, and content distorted to sit under them: three
things that are all required presented as a ladder, a question filed under
"things you can give us".

**The fix is not better labels, it is not needing any.** Write the question as a
complete sentence in the title, put supporting detail under it, and **leave the
detail empty when there is none**. A title that is a real question stands on its
own. The labels only existed because the content had been cut into fragments
that then needed grouping.

**4. Printing the narration.** Nine caption lines, all deleted - and then the
same sentence again as a label above the title, as a line under it, and as the
third column of a table with no class on it at all. **Removing the slot does not remove the
sentence**, which is why the rule is not a list of banned containers: it is
`references/design.md`'s "the title is the argument", and it is stated where a
slide is written rather than here.

**4b. Counting what we deliver instead of describing how it behaves.** The same
number, rejected one way and accepted the other:

    no    四個查詢:客戶資料、已登錄產品、報修案件、維修進度
          four queries: customer record, registered product, ticket, repair status
    yes   它只查這四件事,不會自己去翻別的資料
          it looks up these four things and nothing else in your systems

Same four, same systems. The first is an inventory of what we hand over; the
second is what it does in front of their customer - **and it answers the question
they were actually holding**, which was whether the thing will go rummaging
through their data.

This is the "who is the subject" rule applied to a number: **the subject has to
be theirs, and the quantity has to be expressed as behaviour.**

**5. Getting the subject wrong.** A page described the approval screen as
showing "the customer's name, product and fault description" - which is the
shape of an internal supervisor reviewing somebody else's record. The actual
scenario was the customer, in their own chat, being asked whether to open the
ticket.

**Check who the subject is on every line.** On an anonymous channel the person
approving a write is usually the person in the conversation, not staff.

**5b. A write with no field saying who it is for.** The counterpart to 6 below,
and the same boundary confused in the other direction.

    ours     which field on the record says which of their customers it is
             for. The agent fills it, so it is a data question
    theirs   whose account the record sits under, whether their system routes
             or counts SLA by creator. Do not ask

Getting this backwards is easy because a write makes every question feel
responsible. It was asked wrongly the first time.

**6. A self-service lookup with no identity.** The deck offered "if there is no
account binding, the customer can give us their ticket number" - and presented
it as a clean scope cut.

**Ticket numbers are usually sequential.** A lookup keyed on one alone lets
anybody enumerate other people's cases, and it is not authentication. **Any
self-service query on an anonymous channel has to say what identifies the
person.** If the answer is "a number they type", that is not an answer -
`../asgard-platform/guide/read-path.md` has what a real one looks like, and it
carries the failure mode: a query that forgets to filter on the injected
identity fails only on the anonymous path, which is the path nobody tests.

### The discovery deck's own rules, which are not the proposal's

Three rules elsewhere on this page belong to the **proposal** and produce a bad
discovery deck when applied. Each has been made and rejected by the person who
had to present it.

**A page you created must say where it came from.** Splitting one of their items
into two pages is often right - one of them carried a security question that
deserved its own - but a title that is not theirs makes the page look like a
demand we invented. The reaction to one was "stop drifting the titles".

    the title    a phrase that does appear in their document, even when the
                 split is ours
    the lead     "your document's item 2 says: ..." - so the page traces back

**Titles: the customer's own section names, unedited.** The ghost deck test -
titles as assertions - assumes you have facts and a conclusion. In discovery you
have neither, and an asserted title does two kinds of damage:

    no    「你們要測的三件事,卡點都不在 AI,而在能不能讀到你們的系統」
    no    「情境一能不能做,取決於 CRM 和 RMA 能用什麼方式讀」
    yes   「測試情境一:AI 客服與 CRM / RMA 整合」   ← their document's heading

The first states a conclusion **before asking the questions that would support
it**, on slide 2, and contradicts the customer's framing to their face - while
pre-empting exactly what the question slides exist to elicit. The second is worse
in a quieter way: 情境一 is our numbering. Their document says the whole heading,
and using it verbatim costs nothing and buys a shared vocabulary for the meeting.
They recognise the slide instead of learning what our number refers to.

**A customer who wrote a document has already classified their own problem**, in
the words they use internally. Take the classification.

**Their own words go on the slide, at whatever length they wrote them.** This is
the positive rule, and it is the one the density rules actively prevent.

A proposal is **us** talking, and every word we cut is respect for the reader's
time. A discovery deck is **the customer** talking - we are spreading their own
document on the table so everyone can look at it together - and compressing it
is not economy, it is distortion.

The room is the reason. The person who wrote the plan is in it, but so are their
IT lead, their network admin, their e-commerce manager, and **none of those read
the document**. A five-bullet summary is legible only to its author; everyone
else sees questions with no provenance. Their own 情境說明, verbatim, is a page
every person in the room recognises as theirs.

So: **give each capability a context slide before any question slide.** Their
description as they wrote it, their sub-items, their diagram if they drew one.
That page asks nothing. It exists so that the pages after it have somewhere to
stand.

    the one-line-per-bullet rule    a proposal rule. Does not apply here
    the three-to-five items bound   a proposal rule. Does not apply here
    a paragraph of theirs           goes on whole

**The page list is not something you design.** Their document already decided
it: three capabilities with three or four sub-items each are those pages. Time
spent working out how to group them or how many pages is right is time spent
re-deciding something already settled, and it was several rounds in one deck.

**Density: one sub-topic per slide, following their sub-headings.** Three to five
items is the proposal's density rule. Applied here it compresses a whole
capability's questions into one table, and five questions with no context read as
a questionnaire - three such tables read as an audit.

Their document is already broken down: a capability with four sub-items is four
slides, each saying what we need and what it unlocks. **More slides is not the
problem; a compressed one is.** This deck is worked through line by line in the
room, and a page per subject is what makes that possible.

**A sentence about the slide is not content, wherever you put it.** There is no
caption, callout or lead field to fill any more - `references/design.md` says why
- but the sentence does not need a field, and these are the three shapes it
takes:

    no    「以上引號內文字出自你們 8/31 的測試計畫」     where the material came from
    no    「今天要談的,是每個情境要拿到什麼才做得起來」   what we are about to do
    no    「這一頁說明三個情境的差異」                  what this slide is

All three are things you say out loud. Printing what you are about to say wastes
the line and tells the room you are reading it. **If a reader would not copy it
into their notes, delete it** - and delete it rather than moving it, because the
next container takes it just as willingly as the last one did.

The last two slides are the two filters made visible: what came out as deferred
scope, with what has to be answered before it comes back, and the questions that
are ours to chase rather than theirs - the ones
`../asgard-platform/wiki/platform-unknowns.md` says nobody has settled. Both are written as sentences a
reader outside this repository can follow, not as counts of rows.

### What a handover deck actually walks through

`../asgard-platform/wiki/setup-path.md` is the order, written for exactly this: where a
credential goes, what gets built from it, the agent's configuration, and why
there is no import step on the Sindri side. Two things on that page save a
handover deck from being wrong:

  - **An HTTP API's credential has no home under Settings.** Data Source is nine
    database providers; Connection is OAuth to five named services. A deck that
    shows their API key going into Data Source is teaching them something that
    does not work
  - **Nothing is published to Sindri manually.** Every Managed Agent is there
    already; enabled serves, disabled does not. A slide with a "publish to the
    hub" step is a slide about a button that does not exist

### Screenshots

**A discovery deck almost never wants one, and you can know that before starting.**
Its pages fill with the customer's own words, and one evidence shape per page
then leaves no room. One deck's images were downloaded, cropped and checked one
by one, and not one was used.

**Decide whether the deck takes images before doing any of that work**, and
default to no for a discovery deck. Cropping and checking are real work, they are
the slow part, and they are not recoverable.

There are none in this repository and none in `asgard-cli`. They live in the
product documentation and are fetched by URL:

    https://docs.asgard-ai.com/img/docs/<path>

**`../asgard-platform/wiki/screenshots.md` is the index** - every picture, what it shows,
and which situation it is for. Read it rather than browsing the docs site: it is
grouped by what you are trying to say, which the documentation is not.

Two things in it are worth knowing before you go looking.

**A proposal usually wants the case-study set, not the console.** The instinct is
to screenshot the admin screens, and those are our side of the work. The four
`sindri-retail-stockout-transfer/` images are one continuous story from a user's
question to a completed action, and the third of them is the approval dialog -
which is the only way anybody has found to explain the governance gate on a slide
without a paragraph.

**A handover wants the setup path**, and `../asgard-platform/wiki/setup-path.md` is the
narration for it. `agent-hub-managed-agent/create.png` is the one screen most
handovers need.

The console is in **English**; only the documentation's captions are zh-TW, so a
zh-TW deck can carry these and write its own captions.

**Open every one before it goes on a slide.** Nothing records when any of them
was captured, so a form that has since changed looks exactly like a current one -
and a stale screen in front of the customer who uses that screen daily is worse
than no picture. A screenshot is also the most common way a credential or another
customer's name reaches a deck, in a corner nobody read.

Download it into the deck's own `assets/` rather than hot-linking the docs site
or a path on somebody's machine: the first breaks when the docs are rebuilt, the
second the moment anyone else opens the deck.

### Links, on a deck that has any

A customer proposal rarely carries one. An internal or partner deck is mostly
links, and all three ways of getting one wrong render identically - nothing
catches them but somebody clicking.

- **Do not assemble one by hand.** `asgard-cli links` prints what this checkout
  is bound to, from the ids already on disk, and prints nothing it would have to
  guess. What it names as not printed is not printed for a reason - take that
  rather than building the URL yourself.
- **A link points at the page being discussed, not at the site it lives on.** A
  root URL is the version somebody writes when they did not look up the real
  one.
- **The name is the link.** A row that already says what the thing is does not
  also spell the URL out beside it.
- **Do not infer who can open it.** A private repository is evidence about the
  repository and about nothing else; partners in the same org have access.
  **Ask who is in the room before removing a link**, because removing one and
  leaving prose in its place is the failure that looks most like care.

### What must never be on a customer's screen

- **Implementation nouns**, in a proposal. No CR kinds, no `sl-` / `dc-` / `ss-`
  prefixes, no namespaces, no Helm, no chart, no CRD. The customer bought an
  outcome. A handover may name what they will click - Managed Agent, Drive,
  Skillset - and still never needs the CR behind it.
- **The second person, in a Chinese deck.** 「你們」 on every page turns a shared
  document into two sides. See step 4 - it is a register rule rather than a leak,
  and it does not apply to a deck in English.
- **Anything only we can resolve.** Question numbers and `REQ-` ids are the
  obvious form, and the rule is wider than numbers: "the five the filters
  removed", "that mechanism mentioned earlier" and "item 3 above" fail the same
  way. **The test is whether a reader who has only this deck can resolve the
  reference.** No question numbers, no `REQ-` or
  `TASK-` ids, no "the items the two filters removed", no "out of scope table".
  These are the easiest words to leak, because they are what you have just
  finished writing and they feel like the content - but they are an index into
  files the customer cannot open, and a slide made of them is unreadable to
  everyone except whoever built the records.

  The test is the same one as for platform nouns: **would a reader who has never
  seen this repository know what the slide says?** Write what the question asks,
  not the row it lives in.

      no    第 14-16 題:平台未解項目
      yes   有三件事要回去跟平台團隊確認,確認後書面回覆

      no    被兩道篩選刷掉的五項
      yes   這五項這一輪不做,以及各自要先有什麼答案才會回來

  This also applies to an outline you show internally before the deck exists.
  An outline written in row numbers cannot be reviewed by the person who has to
  present it.
- **Another customer.** No name, no logo, no screenshot, no "we did this for a
  retailer with 200 stores" that is recognisable. Their engagement is not ours to
  spend.
- **Credentials and coordinates - theirs and ours.** No hostnames, no connection
  strings, no account names, not even in a screenshot's corner. Check the
  screenshots.

  **And not Asgard's outbound addresses either.** They belong in the firewall
  ticket, given to the person making the change - not on a slide, not in the
  repository, not in a mail thread that gets forwarded. They can change, and a
  copy will not; a stale allowlist is the customer's connection dropping.
  `../asgard-platform/wiki/operations.md` is the source to read them from each time. What
  goes in the deck is the request - 把這四個位址加進防火牆白名單 - and never
  the values.
- **A capability nobody has verified.** If it is not in `references/`, in a
  meeting note, or measured, it is not a promise.
- **An implementation noun inside a screenshot.** The rule above is not only
  about the words you type: an approval dialog naming `ts-wms` and
  `create_transfer_order`, or a console screenshot with the whole build-platform
  navigation down its left edge, breaks it exactly as a slide title would. Crop
  before using; `../asgard-platform/wiki/screenshots.md` marks the ones known to need it.
- **Precision we do not have.** "Around 10 minutes, from your own figures" beats
  "11.4 minutes" when 11.4 came from one afternoon's sample.

## Step 6 - write it

**Read this before the first slide, not after the tenth.** Everything below the
next two sections is content; these are the things that cost an FDE a working
afternoon and are invisible until they happen.

### Never edit a slide to satisfy a checker

The most expensive mistake made with this skill so far, and it looks like
diligence while it happens.

A content check reported a scene line as missing. It was there - the checker
collapses whitespace when comparing CJK, so a cover date running straight into
it made the string unfindable. To turn the check green, the FDE removed the
sub-numbering from every scene line. The check went green. **Every sub-topic
slide then claimed the wrong level** - the customer's own numbering says the
scenario, and the slides were now saying it about a sub-topic of it.

The user caught it. The checker could not, because the regression was in
something it does not look at.

    a check that fails on text you can see    the check is wrong. Investigate it
    a change to the slide to make it pass     never

**On a discovery deck, do not run the content checks at all.** They are not
broken; they are the right instrument for a document whose text was fixed before
layout, and this is not that document. Running them produces failures you then
have to decide to ignore - and the section above is what happens on the round
where somebody stops ignoring one.

Where they are run - on a proposal - **they never go fully green on slides**
anyway: `audience` and each slide's `layout` are schema fields and are not
printed anywhere. Read the list, fix what is genuinely missing, and stop.

### Four things about the template that nothing warns you about

  - **Slides are fixed height with `break-after`.** Anything added to normal
    flow can push content onto a new page **with no error** - one deck went from
    14 pages to 18 by gaining a line of links. Footers, links and page marks go
    in absolutely positioned elements, and **re-count the pages after every
    edit**
  - **`<b>` does nothing.** The CJK faces embed weights 400 and 500 only, so
    bold silently falls back to normal. Write `font-weight: 500`
  - **Do not keep a second copy of the content.** For a proposal, whose text is
    settled before layout, writing the structured form first and generating from
    it is right. For a discovery deck it is not - see below - and a half-kept
    second copy is worse than neither

### Borrow the layout language, not the process

    take        the template, the palette, the serif hierarchy, the page
                geometry, the CJK font handling, the PDF output
    do not take the structured-content step, or the content checks built on it

**Not because that process is wrong** - it is a good process, and it is designed
to prevent exactly what went wrong here. Writing the content first, in a file
with no columns and no pages, makes it impossible to add a line because the left
column looks short. The mechanism existed, it was documented, it was skipped,
and then the error it prevents was made over and over.

**The reason not to take it is what a discovery deck is.** Its content is not
settled before layout; **it grows through being rejected.** Many rounds of
revision reworded about half the question sentences - and those sentences are
precisely what a locked content file exists to fix in place. Fixing something and
then changing it repeatedly is not fixing it, and maintaining two files through
those rounds doubles the work of each one.

**And two sources drift on their own.** In that engagement the content file
caught exactly one thing: an inconsistency between itself and the HTML. That
inconsistency could only exist because there were two truths. The single source
is the HTML.

**Say the cost out loud:** without a structured source there is no coverage gate,
so nothing verifies that no fact was dropped while filling the layout. That is
acceptable here for the same reason - with one source there is nothing to drop
between; an edit is the edit.

**Layout checks still run**, and they earn it. Style and placeholder checks and
the page count caught real defects in that deck: a line-height over its bound,
and the silent overflow that turned 14 pages into 18. Those measure the page, not
the words.

### The rest

**The design language is `references/design.md` beside this file**, and the
content contract is `references/slides.json`. Both are here so that building a
deck needs nothing but this repository: the palette, the type scale, the slide
classes, the print rules and the layouts, taken from the `kami` skill's design
system and reduced to the half a deck uses.

**Take the layout language, not an authoring process.** If your environment has
`kami` installed you may render with it, and the decision above still holds -
which rules apply differs between a proposal and a discovery deck, and a layout
skill's own density checks do not know the difference. This skill has already
decided what goes on each slide.

For a discovery deck there is a working starting point beside this file:
`discovery-deck.html`, plain HTML with no separate content file,
every customer noun neutralised and the question wording left as it ended up.

The four rules worth knowing before you write, which `references/design.md`
carries in full:

- **The ghost deck test**, for a **proposal**. Reading only the titles, in order,
  must tell the whole argument. If a title is a topic label - "Current
  situation", "Architecture" - it fails. Titles are assertions: "Support answers
  one ticket by opening four systems". **This inverts for a discovery deck**,
  where you have no facts to assert and the customer's own section headings are
  the right titles - see step 5.
- **One evidence shape per slide.** A slide is a claim plus one thing that backs
  it: three to five items, or a chart, or a quote. Not two of them. **Both the
  three-to-five bound and "trim each bullet to one line" are proposal rules**, and
  a layout skill's own density checks will enforce them on a discovery deck where
  they do not belong - the checks go green on a deck that has compressed the
  customer's document into something only its author can read. See step 5.
- **There is no caption, callout or lead field.** A line about the slide is not
  content wherever it sits, and `references/design.md`'s "the title is the
  argument" is the test: cover the title, and if the sentence still says
  something the slide did not show, it stays. A slide that looks empty is not a
  reason to write one.
- The content contract is `references/slides.json`, and the layouts are
  `cover`, `chapter`, `content`, `quote`, `metrics`, `close`.

Write the deck in the customer's language. **A PDF is what gets handed over**;
produce an editable file only when the customer has said they want to edit it -
and `references/design.md` says why HTML to PDF rather than a presentation
format, which is about CJK rendering rather than preference.

**With no typesetting skill installed the deck is still typeset**, because the
design language is `references/design.md` and `discovery-deck.html` is it
applied. Plain Marp markdown - `---` between slides - is the fallback below
that, and then say in the handover that it has not been typeset. Either way, **do
not invent a second house style**: a plain deck that says what it means beats a
decorated one that does not.

## Before you send it

**Run question 0 over every slide, one sentence at a time.** It is the only one
that has to be run against each sentence rather than each deck, and it is the
one this skill has been corrected on most: the same sentence has come back as a
label above a title, as a line under it, as a caption under a screenshot, and as
the third column of a table that had no class on it at all. The containers were
deleted; the sentence was not, because it never needed one.

Eleven questions, and the last five are the ones that get skipped:

0. **Cover the title and read the sentence.** Does it still tell you something
   the slide did not already show? If not, delete it. Every sentence, including
   one in a table cell, a closing line or a parenthesis - and including one that
   reads to you as argument rather than as narration, which is the exemption
   that has failed every time, because a writer files their own reasoning under
   argument. **Delete it rather than moving it.** Moving it is what produced
   four of the recurrences above.
1. Do the titles alone tell the argument?
1b. Which of the three decks is this? A proposal made while the interview is
    still open is a discovery deck wearing the wrong slide order.
2. Is every number attributed - their figure, our measurement, or absent?
2b. Does any title we wrote count something instead of naming it?
3. Does slide 7 say what is not in scope, specifically enough to point at later?
4. Is there an implementation noun anywhere, including in a screenshot?
5. Is another customer recognisable anywhere, including in a screenshot?
6. Could we deliver everything on slide 8 if they said yes today? If any part
   rests on something nobody has built or verified, **first ask whose unknown it
   is**: theirs moves to slide 9, ours goes back to the platform team before the
   meeting and is answered out loud, never printed.
7. Is there a decision record for the shape, written before the deck?
8. Has every claim on slides 4, 5 and 6 been held against
   `../asgard-platform/wiki/platform-unknowns.md`? Anything on that list is a slide 9
   question, not a feature.
9. Read slide 4 aloud to someone who does not work here. Do they repeat it back
   correctly? That is the whole plain-language test, and it takes a minute.
9b. Is there a question number, a `REQ-`/`TASK-` id, or a phrase like "the items
    the filters removed" anywhere on a slide - or in the outline you showed
    internally? Those name rows in files the reader cannot open.
9c. Chinese deck: count 「你們」. Every one of them is a choice, because the
    subject could have been dropped. Most should be the document instead.
10. If it shows a console screen: was every screenshot opened, is it current,
    and is this a handover rather than a proposal? A proposal showing admin
    screens has usually skipped `../asgard-platform/wiki/screenshots.md`, where the
    customer-facing set is.

## The outline beside the deck

**Write `outline.md` beside the deck, and write it before you edit a slide.**
`outline.md` beside this skill is the worked one, for the deck beside it. A
deck without one is unusable six months later, because nobody can tell whether it
is what was proposed or what was agreed - but the label is the smaller half. The
outline is where a page's claim is held against the thing it came from, and a
slide is the one artefact here that carries no provenance of its own.

It opens with who it was for and on what date, then **the argument in its two or
three beats** - not the page list. If the beats do not survive being read alone,
the deck has no argument yet and the page list is decoration.

Then one row per page: **what that page does, and what it rests on.** The last
column is the one that earns the file:

    a diagram      what it actually draws, in enough words to redraw it
    a screenshot   the exact path it was taken from - product, page, tab - and
                   what is visible on it
    a link         where it points
    a chart file   **the path in this repository, and the rule that the slide
                   changes when that file does**

That last row is the whole point. A slide that copies a chart's own logic -
the five outcomes a tool's `proc-response` sorts into, the fields an Agent
carries - is a copy, and a copy with no pointer back is the thing this
repository removes everywhere else. Write the path, and write that editing the
chart means editing the slide.

Two sections close it:

- **How each screenshot was captured, and when to retake it.** A platform
  release ages every screen, and **an out-of-date screen looks exactly like a
  current one** - so the rule is to open those pages once before you present.
- **What is not done**, each item pointing at the open question it is waiting
  on. A deck shown with a blank box in a diagram is honest; the same deck with
  nothing recording why is a question nobody asks again.

**Checked:** 2026-09-04. The CR vocabulary in the translation table is real at
asgard-kube `cbd8d70` - `SkillSet`, `SourceSet` + `Syncer` + `contextIndex`,
`BotProvider` -> `Workflow` -> `SandboxBlueprint`, an `Agent` mounting a
`SemanticLayer`, and `requestConsent` on a gated tool all exist and mean what
the right-hand column says. Every shape it tells a deck to take is one
`../asgard-platform/usecase/` assembles from a deployment in production, which is what
makes an estimate on it an estimate rather than a guess.

**Unchecked:** the whole argument about what persuades. The order a proposal
argues in, what must never be shown, and which questions are worth a customer's
time come from **one deck, for one customer, in one industry**. Three of that
deck's six worst questions were copied out of this material rather than reasoned
into existence, which is the evidence this page exists on - and it is evidence
of a failure, not of a method that works. Read the prohibitions as earned and
the positive advice as untested.
