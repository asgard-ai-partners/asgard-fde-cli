# semantic-layer: what to get from the customer

**a database we can read, with an agent asking questions of it**

**This is theirs to provide, not ours to design.** Ask for exactly the thing
named - offering options invites the other side to pick one that does not
apply, and the week it takes to find that out is the week you were saving.

## host, port, database or schema, and the account name

there is no connection without them

Stated in `../guide/requirements.md`.

## **is that account read-only?** Ask explicitly

the one offered first usually is not, and finding out later means going back for a second credential

Stated in `../guide/requirements.md`.

## **that Asgard's four outbound addresses go on their allowlist** - ask their network team for exactly that, not for "a VPN, an allowlist or a jump host"

Asgard is hosted and the agent runs in the platform's own cloud; there is nothing of ours to put on their network. **This is the single most expensive thing to discover in week three**, and it is a ticket, an approval and a window in most companies

Stated in `../wiki/operations.md`.

## whether the allowlist change can be done, and roughly when

a date changes the plan. Do not ask who approves it - a name changes nothing we build

Stated in `../guide/requirements.md`.

## a description of every cube, dimension and measure, in the customer's own words

the CRD requires a description on each, and it is what the model matches on - not the column name

Stated in `../usecase/semantic-layer.md`.

**Checked:** each row above names the document that owns its claim, and
`asgard-cli audit-material --links` resolves those. That is the whole of the
checking: a row is as good as the document it cites.

**Unchecked:** the list itself. Nothing holds it against a finished engagement,
so a shape can be missing something every one of its rows is right about.
