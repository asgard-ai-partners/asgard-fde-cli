# Sindri - Agent Hub

Where published agents run. **Sindri offers no creating or editing** - agents are
built in Odin and published here.

## Project

The switcher at the top of the screen, standing for one real business situation.
A Workspace can hold several Projects - retail, semiconductor, financial
investment - each publishing its own agents.

A Project is decided at build time. In Sindri a user can only switch between the
Projects they have access to, not create one. Switching changes the Available
Agents, Directory and My Chats.

## The home page and delegation

The centre of the home page is a prompt box, with the published Available Agents
listed below it. Each card shows a name and a line about when to delegate to it.

**A user does not have to pick the right agent first.** The selector above the
prompt box defaults to Sindri rather than any one agent, and Sindri routes the
question using each agent's delegation description.

That is why a Managed Agent's `description` is routing text rather than a
self-introduction: it is the only thing the routing decision can see.

After sending, the screen shows, in order: "Thought for a moment", an
automatically generated title, a Subagents panel naming the agents delegated to
and the steps they are running, and finally the reply.

Below the prompt box are two more selectors - the model tier (the Builtin series)
and Reasoning effort - and a paperclip for attaching files.

### The governance gate

When an agent decides it needs to change data in a system, it stops and asks
which option the user wants, then raises an approval dialog. Nothing runs until
the user allows or refuses.

This is what `requestConsent` looks like to the person. A scheduled run has
nobody to press the button.

**There is no product documentation page for the approval gate.** Not under
Sindri, not under Odin - the sources at the foot of this page cover Project, the
home page, My Chat, Directory and settings, and none of them is about this. So
the platform's most-asked-about mechanism, and the one hardest to explain in
words, has nothing to link when a customer asks for documentation - and they do
ask, because "important actions can be confirmed by a person before they run" is
the kind of line that appears on their own acceptance list.

What there is:

    the screen        `../wiki/screenshots.md`, the four-image sequence.
                      Crop the dialog first - it names a `ts-` prefix and an
                      internal tool name
    the mechanism     `../usecase/write-path.md`
    the limit         a schedule cannot approve anything, so a gated tool and a
                      scheduled run cannot share a Toolset

**Do not spend time looking for the page.** It was looked for.

## My Chat and Directory

Two ways of organising conversations:

| | My Chat | Directory |
|---|---|---|
| structure | each conversation stands alone | a folder holding one topic's conversations and files |
| suited to | one-off questions | a long-running subject returned to repeatedly |

A Directory has Chats and Files tabs. Conversations opened inside it stay under
that topic, and files uploaded to Files are available to whichever agent is
delegated to.

Each conversation's menu offers Share (add an email to the access list, or copy a
link), Rename and Delete.

## The Sandbox and its files

Every conversation has a Sandbox of its own behind it, and that is where the
agent's queries, computation and file access happen. The Panels menu beside the
conversation title opens a Files panel, which is that Sandbox's file browser.

A Directory's Files tab and a conversation's Files panel show **the same data** -
the conversation's Sandbox working directory is mounted into that Directory's
shared file space.

## User settings

Settings, bottom left, has two tabs.

**General** - Language (the interface's own language), Spoken Language (the
user's preferred language, which decides what Sindri answers in; one not on the
list can still be detected automatically) and Appearance. Account and password
changes happen in the Management Console; this tab only shows the current account
and last sign-in, with an Edit in Console link.

**Personalization** - Base Tone, Characteristics (Warm, Enthusiastic, Headers &
Lists, Emoji, each with a Default) and Memory.

Memory's "Reference saved memories" toggle lets Sindri store and draw on earlier
interactions, with a Manage link for reviewing them. This is a **per-user**
memory, distinct from an agent's prompt and from the Global Directory.

## Global Directory

Configured in Odin under Agent Hub -> Configuration -> Global Directory. Its
contents are shared with every Agent Hub conversation in that Project and look
the same to everyone. Agents can read but not change them; editing happens only
on that page.

The convention is to put durable working rules in a top-level `AGENTS.md` and
file other reference material however suits.

## Corresponding extracts

Sindri only runs published agents and produces no CRs of its own. Building them
is `../usecase/agent-hub.md` and `../usecase/flow-agent-supervisor.md`.

## Sources

- [Project](https://docs.asgard-ai.com/docs/product-suite/sindri/features/project),
  [home](https://docs.asgard-ai.com/docs/product-suite/sindri/features/home),
  [My Chat](https://docs.asgard-ai.com/docs/product-suite/sindri/features/my-chat),
  [Directory](https://docs.asgard-ai.com/docs/product-suite/sindri/features/directory),
  [general](https://docs.asgard-ai.com/docs/product-suite/sindri/features/settings/general)
  and [personalization](https://docs.asgard-ai.com/docs/product-suite/sindri/features/settings/personalization)
  - asgard-docs `f00e0ee`
- [Configuration](https://docs.asgard-ai.com/docs/product-suite/odin/features/agent-hub-configuration)
  - asgard-docs `f00e0ee`

**Unchecked:** everything here comes from the product documentation. The
delegation logic and the governance gate's actual screens were not held against a
deployment.
