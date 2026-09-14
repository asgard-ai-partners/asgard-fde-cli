# Before handover

before telling anyone it is live, and before a training session

A green deploy is not a working deployment, and the step between them is
not in this repository.

## why the customer says there is nothing there

**Said:** a chart problem, so look at the chart

**True:** Building a resource does not make it visible. Somebody opens the **Management Console**, finds the product's page, uses *Manage Accounts in* and adds the people - **per resource, because permissions do not inherit**. Nothing in a chart, in `check`, in `verify` or in CD can see that it was skipped

Read `../wiki/console.md`.

## what to walk them through

**Said:** the chart, because that is what we built

**True:** The order from a credential to an agent somebody can talk to - and there is no Sindri step in it, because every Managed Agent is published there already

Read `../wiki/setup-path.md`.

## which screens to show

**Said:** whatever is in the documentation

**True:** Two of the recommended images are marked **CROP FIRST** and for different reasons: one dialog names a toolset by its `ts-` prefix, and one card view has the whole Odin console navigation down its left side. Open every one - nothing records when any was captured

Read `../wiki/screenshots.md`.

A handover deck may show the console; a proposal may not. They are
different documents - `proposal-deck` in `.agents/skills/` has the table.

**Checked:** every entry above is here because it actually happened, and each
names the document that carries the right version.

**Unchecked:** whether the list is complete. It grows when somebody gets
something new wrong, so an activity with few entries is not a safe one.
