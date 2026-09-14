# Before customer-meeting

before any conversation with the customer, at any stage

**The intuitive answer is wrong in both directions, and they cost
differently.** Undersell - "it cannot tell who they are", "we would have to
build an approval step", "only the systems you connect to it" - gives away
scope, and from the customer's side being careful and being wrong look
identical. Oversell - a ticket number treated as authentication, a notification
the platform cannot send - is a promise with no way to build it. Read every
entry before the meeting rather than the ones that sound like your topic: the
direction you are about to get wrong is not the one you expect.

This stage's output is speech. Every other stage produces a file, which is
edited and re-rendered; a wrong sentence is in their notes.

## how we reach a system inside their network

**Said:** "a VPN, an allowlist or a jump host - whichever suits you"

**True:** One shape only: they add our outbound addresses to their allowlist. Asgard is hosted and the agent runs in a sandbox in that cloud, so there is nothing of ours to place on their network. Offering options gets one chosen that does not apply. **Ask who can change the firewall - do not hand out the addresses**

Read `../wiki/operations.md`.

## whether an anonymous channel knows who is asking

**Said:** "it cannot tell who they are, so identity has to wait"

**True:** An anonymous channel answers "where is MY order" perfectly well. The caller supplies identity server-side every turn; only the **model** supplies nothing. **LINE's webhook carries a userId**, which is `../guide/requirements.md`'s rather than this pointer's. Deferring this gives away something you already had

Read `../guide/read-path.md`.

## how many agents

**Said:** "one agent, then"

**True:** A flow-agent shape renders **zero `Agent` CRs** - the prompt is on a Workflow processor and the capabilities on the blueprint. A quote built on the instinct is priced for work that does not exist

Read `asgard-cli size flow-agent-single`.

## how an anonymous caller is identified

**Said:** "they can give us their ticket number and we look it up"

**True:** **A ticket number is not authentication.** They are usually sequential, so a lookup keyed on one alone lets anybody enumerate other people's cases. Any self-service query on a public channel has to say what identifies the person - and if the answer is a number they type, there is no answer yet

Read `../guide/read-path.md`.

## how a write is tested

**Said:** "during the test, shall we really create the ticket or just draft it?"

**True:** **Ask whether they have a test environment first.** That question assumes they only have production, and it gives away the strongest version of the first delivery: writing into a test system proves the fields, the validation rules and the status codes, and a mock proves none of them. A mock is the fallback when there is no test environment - or when they will not let us write to production, which is their call

Read `../usecase/write-path.md`.

## whether it can do things, not only answer

**Said:** "we would have to build an approval step for that"

**True:** The approval gate is **what the platform is built around**. A write stops, names the tool, and nothing runs until a person allows it. It is not custom work

Read `../usecase/write-path.md`.

## whether it can email or message people

**Said:** "sure, we can have it send you a summary"

**True:** **The platform sends nothing.** No SMTP, no mail toolset, nothing in the core. It can call an endpoint of theirs that sends mail - if they have one. Promising a notification without asking that first is promising something with no way to build it

Read `../wiki/integration.md`.

## what the agent can reach

**Said:** "only the systems you connect to it"

**True:** **Not true.** Every sandbox carries the coding CLI's own tools - web search and fetch among them - and no Toolset or blueprint setting removes them. The only control is an instruction in the prompt

Read `../wiki/tools.md`.

## whether they can use their own model or their own key

**Said:** "yes, we can point it at your account"

**True:** **Only on Odin.** Sindri and Mimir use the platform's designated models and the LLM cannot be swapped there. So the answer depends on which product the capability lands in - which is question 2b, decided before anyone thinks about billing. A customer with a model contract or a rule about where inference happens needs this while the shape is still open

Read `../wiki/fehu.md`.

## whether they can have a specific model

**Said:** "of course, we will configure that one"

**True:** You can, and it costs something worth saying: a builtin tier is a **logical model backed by several providers with automatic failover**, and a custom `CompletionModel` is one provider, one key, one point of failure. A good reason - compliance, an existing contract, a model they tested against - is fine. A preference usually is not

Read `../wiki/settings.md`.

## what the limits are

**Said:** "30 steps and 3 minutes per request is the limit"

**True:** Those are **defaults**, raised by contacting sales or service@asgard-ai.com. And **do not quote the step count at all**: nothing defines what a step is, so the next question has no answer

Read `../wiki/integration.md`.

**A slide is not an agenda and not a tracking list.** "Who is
responsible for this?" is right to ask in the room and wrong to print - asking it
says you intend to follow through, printing it reads as managing their
organisation. The carrier is the test, not the content: is this answer somebody
we chase after the meeting, or something we build? Six of one deck's twenty
corrections were this, more than any other kind, and three of those were copied
out of this tool's own material.

**Ask what they have already read**, before describing anything. The
product site tells them Odin is for non-technical staff, gives them a vocabulary
- Basic Function, Template - that maps to nothing here, and says Mimir simulates
the future. `../wiki/what-they-read.md` has the three, and it costs
one sentence to find out which of them you are correcting.

**Before the meeting, not during it:** anything on
`../wiki/platform-unknowns.md` that this engagement touches is ours to
chase, not theirs to hear about. Standing in front of a customer saying we do
not know what our own product does is not honesty.

**Never on their screen:** implementation nouns, another customer, coordinates -
theirs *or* our own outbound addresses - and this repository's own bookkeeping,
question numbers included. The `proposal-deck` skill has the rest.

**Checked:** every entry above is here because it actually happened, and each
names the document that carries the right version.

**Unchecked:** whether the list is complete. It grows when somebody gets
something new wrong, so an activity with few entries is not a safe one.
