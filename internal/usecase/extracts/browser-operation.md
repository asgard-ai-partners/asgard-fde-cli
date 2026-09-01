# Operating a system through its web UI

The last resort, for a system with **no database you can read and no API** - an
appliance with only a web console, a vendor back office, an internal tool nobody
has an integration for.

**Seen in:** a deployment operating a commerce platform's back office, where the
capability is a skill describing 88 pages plus everything the menu cannot see.

## When this shape, and when not

Take it **only after establishing there is no better route**:

    a database you can read   the agent composes its own queries      best
    an HTTP API               a fixed contract, reviewable, gateable  good
    another protocol          the agent runs the client in its
                              sandbox: SNMP, SSH, a vendor CLI        good
    a web UI                  this extract                            last

It is last for reasons that do not go away: it breaks when the vendor changes
their UI, it is slow, and **verifying that the agent did the right thing is hard**
in a way the other two are not.

Ask before designing it:

- Does the vendor have an API that nobody has asked for? Very often the console
  is a client of one - open the browser's network tab and look.
- **Is there a command-line route?** The sandbox can run any client that exists,
  so SSH, SNMP or a vendor CLI beats driving a web console on every axis:
  faster, verifiable, and it does not break when the UI is restyled.
- Is there an internal system that already holds this data - a monitoring
  system, a middleware layer, an existing integration?
- Would the customer accept a person doing the write, with the agent doing only
  the reading and the analysis?

**If the answer is still the UI, it is real work, not a checkbox.** Budget for
producing the maps below, and say so in the plan.

## What has to exist before an agent can operate anything

The capability is **not** "give the agent a browser". It is a set of reference
documents the agent reads, and producing them is the work:

    references/page-map.md         every page the menu can reach:
                                   area | page name | route | menu path |
                                   what you can do here | deeper layers
    references/operation-map.md    everything the menu CANNOT see: in-page tabs,
                                   dialogs and wizards, editor panels, the apps
                                   inside nested iframes
    references/embedded-apps.md    each cross-domain iframe origin, and how to
                                   navigate the inner app directly
    references/self-exploration.md what to do when a request falls outside what
                                   is recorded
    SKILL.md                       the entry point, with the refusal list

Routes in those files are **measured, never derived**. Resource ids are
placeholders taken from the live UI, and the skill stores no customer
identifiers.

## Three ways a crawl misses things, all found the hard way

A user asked to be taken to the announcement setting. The agent guessed a URL,
was bounced to the home page, and eventually found it by clicking into a page
builder. The post-mortem rejected the single obvious cause and found **three
independent mechanisms**. Any map you build will have all three unless you plan
against them:

**1. Extraction shaped like controls throws away the prose.** A crawl that
collects tabs, fields, buttons and column names discards sentences that *name*
deeper objects - including a banner on an already-visited page saying which
controls are global. **The page had the answer printed on it.**

**2. Crawling follows links, and an editor entry is not a link.** The way in was
a `<button>` with no `href` and no accessible name. A link-following crawl
**structurally cannot** reach what is behind it, no matter how long it runs.

**3. "Cannot read the iframe" gets treated as "cannot cover it."** The inner
origins had been recorded all along; nobody tried navigating to them directly,
so every page built that way stayed uncovered.

The lesson generalises: **coverage measured in pages visited is not coverage.**
Track what a user can *do*, and keep a per-page ledger of whether its deeper
layers were reached.

## Designing the maps - the part that is actually the work

### Cover operations, not pages

**Pages visited is not coverage.** Track what a user can *do*, and keep a
per-page ledger with a column for the deeper layers - tabs, dialogs, wizards,
editors, embedded apps. An empty column is an admission, and it is what makes the
gap visible before a user finds it.

### Write the route down as measured, never as derived

Every route in the map came from a link that was clicked. Resource ids stay as
placeholders, replaced from the live UI at run time, and **the map stores no
customer identifiers**.

### Say what each page is for, in the user's words

The map is read by an agent trying to match a request like "where do I set the
announcement" to a place. A row saying "設定 > 網店" helps nobody; a row saying
what a person accomplishes there does.

### Record the exceptions as their own inventory

Pages whose content comes from another origin, and menu items that leave for a
different system, behave differently enough that they need their own list with
the recipe for reaching each one. Otherwise every future crawl rediscovers that
it cannot read them, and stops.

## The discipline that matters most: never guess a route

A route may come from exactly two places:

1. a link visible in the current UI - a menu item, a row, a breadcrumb
2. a route already recorded and measured in the maps

**Not** derived from another route, not by changing an id or a version segment,
not "let me see if this loads".

The reason is precise, and it is not about tidiness:

> Guessing wrong and being bounced to the home page is the lucky outcome.
> Guessing a page that is **semantically similar but not the one you wanted**,
> and then operating on it - that is the failure mode nobody detects.

When the map has no route, the answer is to walk the visible menu again, not to
invent one.

### Pin it with schema, not with prose

Where a tool takes a path, make every legal value an `enum` in its
`inputSchema`, and say in the description that ids must be substituted from the
live UI rather than invented. Schema is enforced; a prompt is advice.

## Exploration beyond what is recorded

The records will always lag the system. Grade what the agent may do:

| level | what | who authorises |
|---|---|---|
| read only | navigate and report. Look, do not click things whose effect is unknown | nobody, it is the default |
| reversible action | a setting that can be set back | the user, in the conversation |
| significant action | affects others, or is awkward to undo | the approval gate |
| **refused at every level** | deletion, publishing or unpublishing, bulk writes, payments, permission changes | **nobody. Not reachable by exploring** |

That last row does not relax because the agent is "just exploring". Write it in
the skill, and write it in the prompt.

Apply the customer's data-masking rules at **every** level, including read-only.

## Generate it

There is no CR specific to this shape - it is an ordinary skill, plus a flag on
whatever holds the capability:

    asgard-cli add skillset <system>-ops --repo <where the skill files live>

Then set `browser.enabled: true` on the Agent or the blueprint that binds it,
and bind the SkillSet there. The generated SkillSet already carries the trio and
the field names; what it cannot write is the reference documents below, and
**those are the work**.

## The skeleton

Two sides, in two places.

The CR side goes where the capability is bound - `templates/agent/ag-<name>.yaml`
for the hub shape, or the blueprint under
`templates/supervisor/<name>/sandbox_blueprint.yaml` for a flow agent - plus a
`templates/skill_set/sk-<system>-ops.yaml` for the trio:

```yaml
# Agent, or the SandboxBlueprint for a flow agent
spec:
  managed:
    browser:
      enabled: true
    skillSetNames:
      - sk-<system>-ops
```

The skill side is a normal SkillSet trio (`asgard-cli usecase skill-set`)
pointing at wherever the skill files live:

```yaml
kind: SkillSet
metadata:
  name: sk-<system>-ops
spec:
  sourceSetName: ss-<name>
  searchPaths:
    - <member>/<system>-backoffice
```

Credentials do not go in the skill. A runtime config file written into the
sandbox by a hook is how a session gets its base URL and the user's token -
see `asgard-cli usecase flow-agent-supervisor` for the hook, including why it
must be `user-prompt-submit` rather than `session-start`.

**Where a login cannot be automated, hand the browser to the person.** A real
user completing the login in the sandbox is a legitimate step, and better than
storing a long-lived credential.

## Skills as their own repository

Once the reference documents get real, they outgrow the chart repo. The shape
that works is three repos with an explicit contract:

    <name>-skills     the skills, versioned and released on their own
    <name>-api        the API contract, if the system has one, as its source
                      of truth. A script vendors it in and generates the
                      operation pages - which are never hand-edited
    <name>-kube       injects runtime config and loads the skills via SkillSet

Keep skills that depend on each other in the same repo, and say so in the
frontmatter. **The platform does not resolve dependencies between skills** - a
SkillSet lists paths, and a skill whose prerequisite is not also bound simply
does not work.

## Verify

```bash
asgard-cli check
asgard-cli verify <project>
```

Those confirm the SkillSet resolves. **Nothing verifies the maps**, which is the
part that decides whether the capability works, so verification is its own task:

- **a per-page ledger**, not a page count. For every page: were its tabs,
  dialogs, editors and embedded apps reached, or is that column empty?
- **spot-check the recorded routes** by navigating them from a cold session.
- **try the request that started it all**: ask for something the map does not
  name, and see whether the agent walks the menu or invents a URL.

An automated check on the maps themselves - required columns, cross-file
references, no orphan entries - is worth writing once the format settles.

## Two things not established here

**How the maps were produced** - tooling, or a person working through the system -
is not recorded in the material this extract came from. That decides whether
producing one is a day or a week, so establish it before committing to a date.

**Whether the maps can be regenerated** when the vendor ships a UI change, or
whether it is manual work each time, is likewise unknown. It decides the
maintenance cost of every integration built this way.
