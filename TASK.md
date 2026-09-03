# TASK.md

What this repo is for, and what is not finished. Nothing else - how it is built is
[STRUCTURE.md](STRUCTURE.md), how to change it is [AGENTS.md](AGENTS.md), what the
commands do is [README.md](README.md), and the defects it has produced and why
nothing caught them are in `source/FINDINGS.md`.

## Goal

When an FDE onboards a customer, the coding agent working in that customer's repo
needs enough context to do the work. The end state is a repo shaped like
[`unitech-e-asgard-kube`](https://github.com/asgard-ai-platform/unitech-e-asgard-kube).

**That repo is the output. The CLI's job is to emit the prompts that drive an
agent to produce it, plus the parts identical for every customer.** It does not
write the customer's Kubernetes resources, because roughly 90% of that repo's
916-line `AGENTS.md` is knowledge only the engagement can earn.

## Non-goals

- **Anything to do with git.** Not `git init`, not adding a remote, not
  authenticating to one: Asgard is growing its own mechanism for provisioning a
  customer repository. This needed saying **in the prompts**, not just here - the
  CLI never printed a word about git, but an agent that finds the directory is not
  a repository and reads a tag-driven CD workflow offers to set one up on its own
  every time. Stages 1 and 8 tell it not to, and why.
- **helm and kubectl as packaged dependencies.** Nothing about how this is
  distributed can install them: a tar.gz, a zip, a dmg and `go install` carry no
  dependency metadata and never can, and a Homebrew or Scoop dependency would only
  cover people installing that way - nobody, until those repositories exist. They
  are prerequisites, reported by `asgard-cli doctor` with the install line for the
  current machine.
- **Bundling helm or kubectl into the release.** Four platform/arch combinations
  at ~50MB each, and **kubectl has to stay within one minor of the cluster's API
  server**, so a pinned copy goes stale and is worse than none.
- **Validating `workspace.id` against the platform.** Waiting on the API. It is optional as of 2026-09-02, since nothing rendered reads it; `project add` and `check` say when it is still unset.
- **Reading or writing `platformMainEnvironmentId`.** Per project per env, and it
  only exists after tf-asgard has created the namespace, so it belongs to the
  generated repo's values files.
- **Any command that talks to a cluster or the platform.** The CLI stays offline.

## Open questions

### Are the EKS cluster names customer-specific?

`.github/workflows/main.yaml` hardcodes `asgard-ai-eks` (dev) and
`asgard-ai-eks-prod` (prod), and `asgard-cr-verification/SKILL.md` carries the
full ARN including Asgard's AWS account id, which ships into every customer repo.

These read as **Asgard's own clusters, shared by every customer**, with customers
separated by namespace rather than cluster, which would make the CD workflow
portable verbatim. That is inferred from the tag-to-cluster table in `AGENTS.md`,
not verified. If it is wrong the cluster name becomes a fourth thing the scaffold
must parameterise - and the account id becomes something to parameterise either
way. The account id appears once in the embedded material, in the EKS ARN in
`.agents/skills/asgard-cr-verification/SKILL.md`; the scaffolded CD workflow takes
AWS credentials from repository secrets instead. It used to be allowlisted in
`internal/scaffold/secrets_test.go`, which failed the build on any run of 12+
digits in an embedded file - that test is gone, so **nothing now stops a customer
identifier being embedded in a template by accident.**

### Should the helm major version be pinned?

Homebrew and Scoop both ship **Helm 4** now (4.2.4 as of 2026-09-01), while the
generated repo was written for 3. `template` and `lint` are what the CLI uses and
both still exist, so `asgard-cli doctor` prints a note rather than failing. CI
picks its own helm version independently, so the two can differ silently.

### Browser operation, before it can be an extract

Whether the page and operation maps were produced with tooling or by hand, and how
long one takes. Both decide whether an extract is advice or a commitment.

### Platform unknowns

Five are carried in the scaffolded `docs/open-questions.md` (P1-P5) rather than
here, because every engagement hits them: per-user resource scoping, an audit
record of what an agent did, whether approved content can change before it goes
out, non-HTTP protocol reach (answered: yes, via the sandbox), and the cost of
producing a web console's page map.

## The worklist

Everything completed was removed on 2026-09-03; git log is the record of what
was done and why. What is left is open work only, and the rules the finished
work paid for are in AGENTS.md under "Before you say it is done" rather than
here.

**Why this list exists at all.** The pass that produced it was run reactively -
each gap found by walking into it - and the same mistake was made twice:
asserting what the material contains without reading it. `taiwan-channels` first
said nobody had integrated a commerce channel; a middleware deployment
integrates SHOPLINE across two skills, one of them an 88-page back-office map.
**Searching for four names is not reading a repository.**

### In the order an engagement hits them

  1. **71 published asgard-docs pages are not read into any wiki page.**
     Measured 2026-09-03 against a clone: 162 pages, 79 cited, 83 uncited, of
     which 16 are `draft: true` and unpublished - so 71 is the number that means
     anything.

     **The whole of `developer-reference/` has been read.** What is left uncited
     is the part deliberately excluded - the 14 message-template pages, the
     release notes, the site's own redesign plans - plus most of
     `help-community/` and the `superpowers/` directory, which nobody has
     assessed.

     **What would make this worth another pass** is not the count. It is that
     `help-community/faq/` is where a customer's own questions get answered, and
     nothing here has checked whether its answers agree with what the material
     tells an FDE to say.

  2. **`deleteme/` and `asgard-bussiness-plan` are the two repositories nobody
     has assessed.** The rest of the parent directory was triaged 2026-09-03 by
     grepping each for an Asgard CR kind, the `asgard-ai.com/` annotation prefix
     or `apiVersion: asgard`. `just-inference`, `hugin`, `partner-finder` and
     `ppt` have zero hits between them; `asgard-html2img`'s two are a vendored
     skill template; `content-pipeline` yielded the four CD facts now in
     `next --stage deploy`.

     **`deleteme/` is this tool's own scratch output** - every directory in it
     carries a `.asgard-config.json` and the scaffold layout, and one is named
     after a live engagement. That makes it useless as evidence and worth
     knowing about: grepping the parent directory for `botProviderClass: line`
     returns hits, and they are ours. `usecase chat-channel` says so.

     `asgard-bussiness-plan` is partially read - the whitepaper's architecture
     section corroborated `X-API-KEY`. Nobody has read the rest.

  3. **The kami checkers and the deck skill disagree, and the checkers win
     arguments they should lose.**

     - **`--check-content` once induced a content regression.** Its CJK matching
       collapses whitespace, so a cover date running into an eyebrow made the
       eyebrow unfindable. To turn it green the FDE removed the sub-numbering
       from every eyebrow, and every sub-topic slide then claimed the wrong
       level. **A check drove a change it could not itself see.** The rule that
       follows is in the deck skill: never edit what is on a slide to satisfy a
       checker.
     - **It never goes fully green on slides**: `audience` and per-slide
       `layout` are schema fields and are not printed. Read the list; do not
       chase it to zero.
     - **`--check-density` and `--check-rhythm` are wrong about a discovery
       deck.** They read its question pages as sparse and its alternation as
       monotonous, and in both cases the deck is right. So the skill asks for a
       document its own checks will fail. That is smaller than a document the
       typesetting cannot produce, but it is unstated, and an agent that trusts
       the checker will flatten the deck.

     The typesetting traps below are recorded because each one cost a rebuild:
     fixed-height slides with `break-after` silently turn 14 pages into 18 when
     anything joins normal flow, so footers and links must be absolutely
     positioned and the page count re-checked after every edit; `.co` at 12mm
     and `.footer-mark` at 10mm always overlap, and the template's own example
     page uses both; `<b>` does nothing, because the CJK faces embed 400 and 500
     only, so use `font-weight:500`; and `content.json` must be written first
     with the HTML generated from it, because editing HTML and back-filling the
     IR loses things.

  4. **Two asks for asgard-docs, not for this repository.** `asgard-cli
     issue-report` is the channel other agents use to file against this repo;
     these two are the reverse direction and nobody has raised them.

     - **The approval gate has no product documentation page.** Every screenshot
       of it is Sindri's dialog, which is where authenticated staff work.
       `usecase write-path` says a prompt-level "shall I go ahead?" is not a real
       gate and cannot say what one is on an anonymous channel. `wiki
       platform-unknowns` P8.
     - **The Asgard side of LINE has no current screenshot.** The integration
       dialog in `integration/LINE` is the 2024 console. The LINE Developers
       Console image on the same page is usable and current, so the ask is
       narrower than it was: one screen, not the whole flow.

  5. **The first measured reading baseline, from before the log existed.** One
     engagement read about 15 of ~36 pages, and the rule was: the stage it was
     told it was in, plus whatever that stage pointed at. Never opened, and
     needed: `wiki operations` - in the index the whole time under the title
     Connectivity, while that engagement spent a day on connectivity.

     `asgard-cli reading` records this from now on. **That number is the
     baseline to beat**, and the thing to watch is not the ratio but which
     pages sit in the never-opened column while being relevant.

## What is not done

Everything here is known, not discovered - a line being here means somebody
decided it could wait, or could not be done from here, and the reason is next to
it.

### A statement that shipped and was wrong

The scaffolded `AGENTS.md` told every engagement that **there is no
`CompletionModel` CR and therefore no model API key in `app-secret`**. There is
one, the CRD defines it with six classes, and three of seven reference
deployments declare their own with the provider's key as a secretKeyRef. The
same file listed `CompletionModel` among the CRs carrying
`project-environment-id`, eleven lines above - so it contradicted itself and
nobody read the two together.

What it was describing is the `builtin` class, generalised into the absence of
the CR. Corrected in `AGENTS.md.tmpl` and in `wiki settings`, with the two things
the CRD enforces that helm does not: the class is immutable, and exactly one
provider block may be present.

**The lesson is about the sample, not the sentence.** This survived because the
material was written from deployments that all use builtins. Every remaining
`AGENTS.md` claim about what does not exist deserves the same check against the
seven charts, and that has not been done.

### What automated checks cannot see

Six defects in `proposal-deck` were found by an engagement building a real deck,
and **every one passed the layout skill's own checks** - density, rhythm and
content all green. They shared a cause: the skill carried the proposal's rules
and applied them to a discovery deck, where several of them invert.

Titles as assertions become conclusions stated before the questions that would
support them. The three-to-five item bound compresses a customer's document into
something only its author can read. `cap` fills with narration. Two recommended
screenshots carry `ts-` prefixes and the build console's own navigation, which
the same skill's first rule forbids.

All six are fixed. **The pattern is worth keeping**: the rules that produce a
good artefact of one kind silently produce a bad one of another, and a checker
that validates shape cannot tell them apart. Any recipe added here should say
which kind of artefact it is for, and what inverts for the others.

The counterpart is also worth recording: **`demo-generation` is now an extract**,
and the demo generator remains the largest body of Asgard chart material there
is. What it has that nothing else does is an orphan check in both directions -
every asset must be used by a story, not only every reference resolved.

### The solution vocabulary is agent-shaped

Not a missing extract - a missing **kind** of extract, and it is the one that
changes what gets proposed to a customer.

Every shape here assembles an agent. So does the interview, until 2b was added:
its ordered questions go from what they cannot do today, to who is on the other
end, to which systems hold the data, and every one of them assumes the
deliverable is something you talk to. An agent asked what to propose therefore
proposes an agent, and does it fluently, which is what makes this hard to notice.

It has already cost one proposal. A customer's cross-channel inventory question
- exactly the thing Mimir is for, and the subject of one of the product
documentation's own case studies - came back as an agent over a semantic layer,
because nothing in the interview asks what they do with the answer and no shape
existed to propose instead.

`mimir-dashboard` and the interview's 2b close the Mimir case. **Four products
are still unrepresented**: Heimdall, Fehu, the Management Console as work in its
own right, and Knowledge Base as distinct from a Drive. `wiki product-suite`
describes all six in a table; nothing turns any of them into something an
engagement can propose. The next one to hit this will be a customer whose
question is about permissions or about cost.

### Three CRDs nothing covers

`ImageGenerationModel`, `TranscriptionModel` and `SourceSetEditorServer` exist in
the platform contract and appear in **no** product documentation page and **no**
material here. An engagement whose customer needs image generation or
transcription would find the CRD and nothing else - no shape, no traps, no
statement that they are not meant to be used yet.

Which of the two it is matters and is not known: the platform may simply be ahead
of its documentation, or these may not be meant for an engagement to reach for.
Worth one question to the platform team, and cheap to answer.

### Gates that could be stronger

- **`project add`'s terraform prerequisite stays a notice, not a gate.**
  Settled 2026-09-03 rather than left open. The demo generator refuses without
  the namespace and both `platformMainEnvironmentId` values because it runs
  where it can see them; this CLI is offline by design, and at `project add`
  time the id **does not exist yet** - tf-asgard has not run. Refusing would
  block the first step of every engagement on something that is correct to be
  missing. The condition is checkable after rendering and `verify` checks it.

- **The 79 CEL rules are not checked, and 40 of them cannot be.**
  `gate.Enums` (E1) and `gate.Constraints` (C1) cover the CRD's enums, patterns,
  lengths and bounds. What is left is `XValidation`, and **40 of the 79 are
  `self == oldSelf`** - they compare a proposed object against the one already on
  the cluster, and a render is one object with no history. `botProviderClass` is
  the one that bites; `usecase chat-channel` documents it instead.

  The remaining ~39 are checkable and are not checked. The tractable ones are
  the conditional shapes - `toolsetClass != 'mcp-server' || has(mcpServerConfig)`,
  `documentClass != 'video' || video != null` - which are a field implying
  another field's presence. That is a real class and nothing catches it.

### Blocked on access nobody in this loop has

- **The wiki is checked less deeply than the extracts, and that is structural.**
  An extract has a chart to hold it against; the wiki's source is product
  documentation describing a UI, much of which is in no chart at all. Every page
  says how far it got on its `**Unchecked:**` line, and `wiki --unverified`
  lists them.

  Two mechanical passes have been run and found nothing: every enumerated set
  against the CRD enums, and every completeness claim ("the thirteen processor
  types", "the CRD supports ten", "`agentClass` has one value"). A third was
  written and thrown away for calling correct material wrong - see the third
  question under "Before you say it is done" in AGENTS.md.

  **The fix is a login and an afternoon.** `console`, `sindri`, `mimir`, `fehu`
  and `settings` all describe a UI; one person with access could settle all
  five, and nothing short of that will.

- **`wiki setup-path` states an order no source states.** Every step comes from
  the page that owns it and no source puts them in a sequence, which is why the
  page exists. **One fork is verified** - that an HTTP API's credential has no
  home under Settings - and the rest is a reading of what each step needs from
  the one before. Its Sources say so and ask whoever reaches a live console to
  report what is wrong. It is the page most likely to be confidently wrong.

- **`chat-channel` has never run anywhere.** Every BotProvider across every
  reference deployment is `generic`. The credential blocks were checked field by
  field against the CRD on 2026-09-03 and match, and the CRD's `ExactlyOneOf`
  and immutability rules are recorded - but nothing here has been *run*, and the
  first customer on LINE is that page's first test.

  **There is a trap in checking this yourself.** Grepping the parent directory
  for `botProviderClass: line` returns hits, plus `discord`, `slack` and
  `telegram`. Every one is inside a scratch repository this tool scaffolded.
  They are the page's own output, and the extract says so.

### Deferred by the FDE, needs a decision rather than work

- **Homebrew tap and Scoop bucket.** The config is written and guarded by its
  token; enabling either is creating a repository and adding a secret. The FDE
  deferred this deliberately.
- **Linux packaging beyond what nfpm does by default.** Deferred by the FDE.

### Questions that block nothing yet

Listed in full under "Open questions": whether the EKS cluster names are
customer-specific, whether the helm major version should be pinned, and what
producing a web console's page map actually costs. The five platform unknowns
live in the scaffolded `docs/open-questions.md`, because every engagement hits
them and the answers belong where the engagement is.
