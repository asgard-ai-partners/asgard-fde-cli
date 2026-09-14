# knowledge-drive: what to get from the customer

**documents the agent reads - manuals, FAQs, pages**

**This is theirs to provide, not ours to design.** Ask for exactly the thing
named - offering options invites the other side to pick one that does not
apply, and the week it takes to find that out is the week you were saving.

## the documents themselves, or the place they live and access to it

a Drive syncs from somewhere; without the source there is nothing to index

Stated in `../usecase/knowledge-drive.md`.

## who keeps them current, and how often they change

it decides the Syncer's schedule, and whether a stale answer is a real risk

Stated in `../guide/requirements.md`.

**Checked:** each row above names the document that owns its claim, and
`asgard-cli audit-material --links` resolves those. That is the whole of the
checking: a row is as good as the document it cites.

**Unchecked:** the list itself. Nothing holds it against a finished engagement,
so a shape can be missing something every one of its rows is right about.
