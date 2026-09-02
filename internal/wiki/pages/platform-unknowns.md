# What the platform's documentation does not answer

Questions no source settles - not the CRDs, not the product documentation. Each
one is here because an engagement hit it and had to proceed without an answer.

**Ask the platform team when a requirement touches one. Do not assume.** Then
write the answer into the customer repo's `docs/decisions/`, cite it wherever the
design depends on it, and tell whoever maintains `asgard-cli` so the next
engagement starts with the answer rather than rediscovering it.

**Checked:** 2026-09-02 against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` and asgard-docs `f00e0ee` - each row was re-checked to confirm neither source answers it.

**Unchecked:** whether the platform team has since answered any of these outside those two repositories.

## The list

| # | Question | When it matters |
|---|---|---|
| P1 | **Can one user be allowed to see or operate only some of the resources, while another sees others?** The entry point authenticates the caller, but nothing found so far scopes what a given user may then reach. Asked because a customer's own test plan required limiting which equipment different users could query or operate, and there was no answer. | Any customer with more than one team. A security review asks it early |
| P2 | **Is there a record of what an agent did - who asked, what was called, what came back - and can it be shown to an auditor?** CRs carry a label the platform echoes into audit events, so something is recorded, but not what or for how long. Asked because a customer's test plan required keeping a record of equipment operations. | Anything with a side effect, and every regulated customer |
| P3 | **Between a human approving a call and that call going out, can the content change?** Asked because an engagement had to write "the approved listing is the one that gets published" as an assumption with nothing to cite. | Any case where the approved content **is** the deliverable - a listing, a message, a document |
| ~~P4~~ | ~~Can the platform reach a system over something other than HTTP?~~ **Answered: yes.** The agent gets a sandbox, so any protocol with a client - SNMP, SSH, a vendor CLI - is reachable by running that client. Reads may happen there; **writes still go through a Toolset with `requestConsent`**. See `asgard-cli usecase external-api`. | - |
| P5 | **Before an agent can operate a system that only has a web console, someone has to write down what that console contains** - every page, what each page does, and what is behind the buttons the menu does not show. One deployment has such a document covering 88 pages. **Is producing it tooling-assisted, or does a person work through the system by hand?** And when the vendor restyles their UI, is it redone? | Any web-only system. It is the difference between a day and a week, and it decides the maintenance cost |
| P6 | **Are the per-request limits adjustable per workspace?** One request gets 30 steps and 3 minutes, and an endpoint serves 5 requests per second. A conversation that consults a knowledge base, then a CRM, then a ticket system, then asks a follow-up can reach 30 steps. Whether that ceiling can be raised, and at what tier, is not documented. | Any multi-system agent, and every customer-service scenario that troubleshoots across more than two systems |

## Why this is not in the customer's repository

It used to be, in `docs/open-questions.md`, and that was wrong for the reason
this whole body of material is embedded rather than copied: a copy in one
engagement goes stale where nobody is looking, while a stale one here is fixed
for every engagement in a single release. A repo that filed its first question in
March and one that filed its first in September would have disagreed about this
list forever, and the older one would have been the one still showing an answered
question as open.

The customer repo's `docs/open-questions.md` holds that engagement's own
questions and nothing else.

## Adding a row

Every row here came from an engagement actually hitting it - a customer's test
plan asking for something with no answer, or a design that had to proceed on an
assumption with nothing to cite. Add one the same way.

A question that merely seems like a good thing to know does not belong here. It
dilutes the list, and the list only works because every row is one somebody has
already paid for.

When one gets answered, do not delete it: strike it through and write the answer
in, the way P4 reads now. The fact that it was once unknown explains the shape of
designs that were made without it.

## Sources

- Each row names the engagement or the customer requirement that produced it
- P4's answer: `asgard-cli usecase external-api`
- P6's numbers: [Quota and limits](https://docs.asgard-ai.com/docs/help-community/quota-limits)
  - asgard-docs `f00e0ee`
