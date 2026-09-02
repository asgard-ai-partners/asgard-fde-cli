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

## The worklist, 2026-09-02

Written down because this pass was run reactively - each gap found by walking
into it - and the same mistake was made twice: asserting what the material
contains without reading it. **`taiwan-channels` first said nobody had
integrated a commerce channel. A middleware deployment integrates SHOPLINE
across two skills, one of them an 88-page back-office map.** Searching for four
names is not reading a repository.

### Done this pass

  - `mimir-dashboard`, `demo-generation`, `taiwan-channels`, `setup-path`,
    `screenshots` - five new pieces of material
  - `asgard-cli size` - what a capability is made of, before it is written
  - the interview's 2b (which product), and 6b turned from a blocker into a
    read-back
  - the discovery deck, and the six rules of the proposal's that invert for it
  - `CompletionModel` corrected in `AGENTS.md.tmpl` and `wiki settings`
  - `find` searches the design-time skills; `scaffold` reports stale material;
    a record can be named by its ID; a question number is not reused

### Open, in the order an engagement hits them

  1. **83 of asgard-docs' 162 pages are not read into any wiki page**, and the
     hole is one section rather than scattered.

     Two wrong numbers preceded this one and both were believed rather than
     checked. The index said 130 cited, "100% of what is in scope" - measuring
     one source of nine. Then this entry said 93, from a comparison that was
     case-sensitive and matched paths against pages that cite URLs, so it
     counted every channel page as missing when `integration.md` is built from
     all of them. **Do not quote a coverage number here without saying how it
     was measured.** Three have been wrong in one day.

     What is actually missing is **the whole of `developer-reference/`**:

     | uncited | why it matters |
     |---|---|
     | ~~`processor/` - 16 pages, one per processor~~ | **done 2026-09-02** - `wiki processors`. This was the largest single gap: a chart author writing a Workflow had a type list and no fields |
     | ~~`api-doc/send-message/sse-response/**` - 11 event pages~~ | **done 2026-09-02** - the envelope, the `fact` tagged union and `runError.location`, which names the Workflow and processor that failed. `api.md` had the event list and none of the payloads |
     | `api-doc/send-message/` - the four endpoint pages | `api.md` covers the endpoint, actions and `customChannelId` accurately. What is unread is the file path: `blob` upload as multipart, and append-file-and-send |
     | ~~`asgard-builtin/**` - 18 pages~~ | **rechecked 2026-09-02, and the exclusion was wrong for four of them.** The expression language every processor field is written in - three value types, six variables, seven functions, the Blob shape, and that Expression is ECMA5 with no optional chaining. Now in `wiki processors`. The 14 message-template pages stay excluded and that half of the judgement holds |
     | `examples/` - webhook-integration, knowledge-base-query, streaming-response | worked examples, which nothing here has |
     | `sdk/`, `asgard-sdk`, `others/channel-log` | the front-end path, which `api.md` covers from the platform side only |
     | `overview/asgard-concepts`, `why-asgard`, `product-suite/*/intro` | the vocabulary a customer will have read before meeting us |

     **`examples/` and the file path next.** Four worked examples - webhook
     integration, knowledge-base query, streaming response - and the `blob`
     upload endpoint, which is the only part of the API contract `api.md` still
     describes without having read.

     **And a habit to keep**: three inherited judgements were checked today and
     three were wrong - the coverage percentage, the `asgard-builtin` exclusion,
     and this repo's own claim that there is no `CompletionModel` CR. An
     exclusion or a summary statistic is somebody's conclusion, not a fact, and
     the ones that exclude a whole directory are where nobody has looked.

  2. **Heimdall has one documentation page and it is a link to a marketing
     site.** The deployment behind it - a content pipeline with 28 Plugins, the
     only `KnowledgeBase` in any chart, and a scheduled web crawl - is here and
     mined for two extracts. So the product an engagement might be asked about
     has no material, while its deployment is one of the best-documented. Decide
     whether Heimdall is in scope for an FDE at all, and write that down either
     way.

  3. ~~**`size` does not count ConfigMaps.**~~ **Done 2026-09-02** - one per
     Workflow, labelled in the output as what it is, since a reader who has
     never seen one will otherwise assume the estimate is wrong.

  4. **The per-processor config definitions in `asgard-core` are still not
     carried, and one attempt was discarded.** `ProcessorDefinitions` holds
     name, type, `IsRequired` and `DefaultValue` for every config key - the only
     place the defaults exist, since the documentation pages give none.

     A pattern-based extraction on 2026-09-02 misaligned, attributing one
     processor's fields to the next, and was thrown away rather than published.
     Two were then read individually and are in `wiki processors`:
     `retrieve-knowledge.sampleK` is required and defaults to **20**, and
     `validate-payload.path` defaults to `"$"`.

     **Do it with a real parse** - `go/ast` over the file, or a small Go program
     importing the package - not with a regex. A wrong table here is worse than
     none, because a default is exactly the kind of fact nobody re-checks.

  5. **Fehu and the Management Console reached the walk today; neither has an
     extract.** The Console's fact is a go-live blocker - build a resource, skip
     the grant, the customer sees nothing - and it is now in `08-deploy.md`.
     Whether either needs more than that is unjudged.

  6. ~~**Three reference charts have been counted and not read.**~~ **Their
     AGENTS.md files are read, 2026-09-02.** Five things came out that nothing
     here had, all now written up: the platform sends no mail at all; a
     brand-new channel's first message may never be answered; a join can render
     fine and match zero rows; `joins[].relationship` has no `many_to_many` and
     forcing one inflated sums 18.2x; and the sandbox's own CLI tools cannot be
     disabled by any CRD field. Also that `Toolset.spec.instruction` is gone
     from the live CRD.

     **What is still unread in those three is the charts themselves** - the
     AGENTS.md files say what each deployment learned, and the templates say
     what it actually emits. Most of the older extracts came from these repos,
     and nothing has checked whether those extracts still match what is there.

  6b. **And the repository set was scoped from memory and was wrong.** Every
     count of "the seven reference deployments" in this file and in the material
     excluded the platform's own code. Under `projects/asgard/` there are also
     `asgard-freyr-api` (873 Go files - the application behind the Freyr
     skills), `asgard-router` (154), `content-pipeline` (1674 Python files),
     `deleteme` (130 CRs, 1131 Python files, despite the name), plus
     `asgard-bussiness-plan`, `partner-finder`, `ppt`, `hugin`,
     `just-inference` and `asgard-html2img`. **None has been opened.**

     `asgard-kube/pkg/apis/` was in the same blind spot until today and produced
     `wiki crd-rules` in one pass, so the expected yield from the rest is not
     low.

     **`asgard-freyr-api` read 2026-09-02, and the yield was one thing.** It is
     a conventional Go service - Gin, fx, GORM, layered handlers/services/
     repositories - and carries almost no platform knowledge, so nobody should
     re-read it hoping. What it does carry is the shape for **a credential the
     customer's own users supply**: AES-256-GCM sealed in a column, key from the
     environment, masked for display, the boundary isolated behind one package
     the build stops other layers importing. That is now in the interview at 3b,
     because it is the point where an engagement stops being a chart and needs
     somewhere to run code, and it should be said early rather than discovered.

     **`asgard-router` read 2026-09-02**, and it answered a question the wiki
     had been treating as opaque. `settings` said the builtin tiers are
     "semantic aliases rather than specific model names" and stopped there. The
     router is what resolves one: a logical model backed by several
     provider-model pairs, with weighted-random, round-robin or ordered-fallback
     selection, and **automatic failover on a 5xx or a timeout**. So a builtin
     tier is a pool, and choosing a custom `CompletionModel` gives that up for
     one provider and one point of failure - which is the trade to state when a
     customer asks for a named model. Now in `wiki settings` and in
     `brief customer-meeting`.

     **`deleteme/` is this tool's own test scaffolds**, not deployments -
     `coldstart`, `coldstart2`, `coldstart3`, `final`, `final2`, `classes`,
     `classes2`, `precommit`, plus one named for a customer engagement. It
     carries no platform knowledge and should not be read as a reference chart;
     its 130 CRs inflated every count in this file until today.

     It is worth one thing: **somebody ran a cold start three times.** The three
     attempts differ in which projects they ended with - `erp`, then `erp` and
     `site`, then `site` alone - which reads as the project split being redone
     rather than a command failing. That is the decision stage 2 exists to make,
     and it suggests the split was being discovered by trying it. Not
     actionable on its own; worth remembering if the same shape appears again.

     Still unopened: `content-pipeline` (1674 Python files),
     `asgard-bussiness-plan`, `partner-finder`, `ppt`, `hugin`,
     `just-inference`, `asgard-html2img`.

  7. ~~**The kami formatting report is unactioned.**~~ **Written into step 6 of
     the deck skill, 2026-09-02.** Kept below because the first item is a rule
     about checkers in general, not about this one.

     ~~The substance:~~ An FDE built the deck and
     reported back; nothing has been written from it yet. The substance, so it
     is not lost if the message is:

     - **`--check-content` induced a content regression.** Its CJK matching
       collapses whitespace, so a cover date running into an eyebrow made the
       eyebrow unfindable. To turn it green the FDE removed the sub-numbering
       from every eyebrow - and every sub-topic slide then claimed the wrong
       level. **A check drove a change it could not itself see.** The rule that
       follows: never edit what is on a slide to satisfy a checker.
     - `--check-content` never goes fully green on slides: `audience` and
       per-slide `layout` are schema fields and are not printed. Read the list;
       do not chase it to zero.
     - Fixed-height slides with `break-after`: anything added to normal flow
       silently turned 14 pages into 18. Footers and links must be absolutely
       positioned, and page count re-checked after every edit.
     - `.co` at 12mm and `.footer-mark` at 10mm always overlap, and the
       template's own example page uses both.
     - `<b>` does nothing: the CJK faces embed 400 and 500 only, so bold
       silently falls back. Use `font-weight:500`.
     - Write `content.json` first and generate the HTML from it. Editing HTML
       and back-filling the IR loses things.
     - **The content rules did not fight the template - they fight the
       checkers.** `--check-density` reads a discovery deck's question pages as
       sparse and `--check-rhythm` reads its alternation as monotonous, and in
       both cases the deck is right. So the skill asks for a document its own
       checks will fail, which is a smaller problem than one the typesetting
       cannot produce, but it is unstated.

  8. **Issues are for other agents to file, not for me.** `asgard-cli
     issue-report` is the channel and #9 is the worked example. The five things
     drafted as issues are maintainer work and stay in this file: `next`
     reporting one state for a repo whose capabilities are in several states;
     `verify` and `deploy` disagreeing on an empty `platformMainEnvironmentId`;
     question numbers colliding across branches; the approval gate having no
     product documentation page; LINE having no current screenshot. The last two
     are asks for asgard-docs rather than for this repository.

  9. **No engagement has been walked end to end with the current material.**
     Everything above was found by watching one, at stage 2. Stages 3 to 9 have
     been read but not exercised since any of this changed.

## What is not done

Ordered by what an engagement would hit first. Everything here is known, not
discovered - a line being here means somebody decided it could wait, and the
reason is next to it.

### Reference material still missing

- **auto-post's 28 `Plugin` CRs have no extract.** `plugin` covers one; nothing
  covers the shape at that scale, or what governs which plugin an agent may
  reach.
- **`router` at scale.** `workflow-chain` gives the mechanism and a two-branch
  example. auto-post's content pipeline has nine branches across several
  workflows, and the question that shape answers - when a branch belongs in the
  graph rather than in the prompt - is exactly the one an FDE gets wrong.
- ~~**The `execute-script`, `validate-payload`, `generate-embedding` and
  `retrieve-knowledge` processors** appear in the platform's contract and in no
  chart that has been read.~~ **Answered 2026-09-02** against `asgard-core`'s
  `internal/constants.go`, which is the source of truth the CRD is generated
  from: all four are real, defined in `ProcessorDefinitions` with full static
  config definitions. So the answer is the second one - the sample of
  deployments was too small - and the wiki's list of 13 processor types matches
  the source exactly, with no extras and none missing.
- **Which deployments have been mined has never been written down, and the
  suspicion above turned out to be right.** A finance deployment sitting in the
  same parent directory the whole time carries a shape nothing here covered - a
  SemanticLayer deliberately bound to no Agent, whose consumer is Mimir - and it
  was found by a customer asking for it, not by anyone looking. `mimir-dashboard`
  now covers it. What is still missing is the audit: the six reference charts
  between them declare `Toolset`, `Trigger`, `KnowledgeBase`, `Loader`, `Source`,
  `Plugin` and `CompletionModel`, and nobody has checked those uses against the
  extracts that claim to describe them.

  **Do this before writing another extract.** One pass over the charts, listing
  which shapes each one actually uses, turns "we think the sample is too small"
  into a list. Every extract written without it is written from whichever
  deployment somebody happened to remember.

  **That pass has now been run once**, over the seven reference charts, counting
  declared kinds. What it found, beyond the Mimir shape:

  | | what nothing here covers |
  |---|---|
  | the demo generator | 123 SkillSets, 80 Workflows, 64 SemanticLayers and 64 Agents from one parameterised source. Generating a deployment per industry is a shape in itself, and it is the largest body of Asgard chart material in existence |
  | Heimdall (auto-post) | 28 `Plugin` + 28 `SkillSet` one-to-one, and the only `KnowledgeBase` + `Loader` + `Source` in any chart - the older knowledge path the wiki says is still live |
  | Freyr | 5 Agents over 5 SkillSets with **no SemanticLayer at all** - capability entirely from skills and one Toolset |
  | `ConfigMap` | 80 in the demo generator, 2 in Freyr. Not an Asgard CR; no material says what one is doing in these charts or when to reach for it |
  | `CompletionModel` | corrected below, but no extract - the shape is a customer bringing their own model |

  The counts are declared kinds, not shapes. Turning each row into an extract
  still needs the chart read rather than counted.

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

- **Two gates disagree about an empty `platformMainEnvironmentId`.**
  `gate/deploy.go` treats it as a warning and says why - the platform issues the
  id after the namespace exists, so it is empty through the whole middle of an
  onboarding. `gate/xref.go` errs on the same condition for a `Trigger`, with a
  message that names neither the cause nor the fix. So a freshly scaffolded
  project that adds a Trigger fails `verify` and is told to look at a label the
  template already writes. Reported, not fixed.
- **Nothing knows that a SemanticLayer may have no consumer on purpose.** That
  is the whole of `mimir-dashboard`, and the deployment it came from says
  plainly that its own cross-reference check does not protect it: the danger is
  a later reader binding the layer to an Agent as a tidy-up, which silently
  gives agents restricted to an API a second path into the database. A gate
  cannot tell that apart from a genuine omission without being told which it is,
  and no CR field says. Worth a marker the chart can carry.

- **`project add`'s terraform prerequisite is a notice, not a gate.** The demo
  generator's equivalent command *refuses* to continue without the namespace and
  both `platformMainEnvironmentId` values. Theirs is stronger. Doing the same here
  needs the CLI to be able to check the condition, which today it cannot -
  it is offline by design.
- **Nothing checks that a `tooling.description` names the tool it could be
  confused with.** It is the single field where a wrong value makes a model pick
  the wrong tool, three extracts say so, and it is unenforceable by anything but
  a person reading it.
- **No gate compares a generated CR against the platform's CRD schema.** The
  first such pass found five violations and two extracts teaching the wrong
  thing, so the class is real and nothing here catches it. Making it a test means
  vendoring the released `asgard.crds.yaml` and validating the rendered output of
  every kind against it - required fields, unknown fields, enums, patterns, CEL.
  What that costs is a pinned copy of a contract that moves: a stale schema fails
  the build for the wrong reason, and pinning it to a platform version this repo
  does not otherwise depend on is a new thing to keep current. Worth doing anyway,
  because the alternative is finding out during someone's CD.
- **Nothing links a retired platform field to every file that mentions it.**
  Vocabulary lives in templates, extracts and stage prompts; a rename caught in
  one leaves the other two teaching a field that no longer exists.
- **Nothing checks the embedded material against itself.** A stage prompt can
  send a reader to a page that does not exist, and a generator kind can name an
  extract that was renamed, without anything noticing until somebody follows the
  pointer. The checks that did this were in the test suite, removed on
  2026-09-02; **the gate is now `go build`, `go vet`, `gofmt -l` and running the
  CLI by hand against a scratch repository.**
- **`internal/usecase` and `internal/wiki` are 190 lines each of the same four
  functions.** Same `parse`, `List`, `Read`, `Search`, over two corpora that now
  share a provenance format and a linking rule. One `internal/kb` with a shelf
  per corpus is the right shape; it was deferred rather than done before a
  release, because both packages are correct today and a refactor is not.

### Material that is thinner than it looks

- **`wiki screenshots` describes about a hundred images and five were opened**,
  and two of those five turned out to need cropping before a customer could see
  them - one of which this page had recommended as the single most useful image
  it carries. So the marked ones are not the only ones; they are the ones
  somebody happened to download.
  The rest carry the documentation's own alt text, which describes the file
  honestly but does not say whether the product still looks like that, and
  **nothing anywhere records when any of them was captured**. Alt text also
  cannot see what this pass learned to look for: our own implementation nouns
  inside the picture. Only opening one finds those.
- **`wiki setup-path` states an order no source states.** Every step comes from
  the page that owns it, but the sequence is assembled, and only its first fork
  - that an HTTP API's credential has no home under Settings - was held against
  a real console screen.

- **The wiki is checked less deeply than the extracts, by nature.** An extract has
  a chart to hold it against; the wiki's source is product documentation
  describing a UI, much of which is in no chart at all. Every page says so on its
  `**Unchecked:**` line, and `wiki --unverified` lists them - but the fix is to
  hold each page against a deployment the way the extracts were, and that has not
  been done.
- **`chat-channel` has never run anywhere.** Every BotProvider across every
  reference deployment is `generic`; the credential blocks and the per-class costs
  are read off the contract. The first customer on LINE is that page's first test.
- **The systems-inventory example in `02-projects.md` is modelled on one real
  customer** - their ERP's database, their catalogue's three dimensions. It names
  nobody, but somebody who knows that engagement would recognise it, and this repo
  states elsewhere that nothing shipped names another customer. A real example is
  more persuasive than an invented one, so whether to replace it is a judgement
  nobody has made.

### Deliberately deferred

- **`asgard-cli reference add <file>`.** Filing a customer's document into
  `references/` is a step every engagement takes and none does the same way: the
  agent invents a provenance table each time, and `references/customer-source/`
  is a convention one engagement made up. A command would make provenance
  mechanical. Not built because the shape of the record is a decision, not a
  detail - what it must carry is the argument, and nobody has had it yet.
- **Nothing enforces that a project split follows a recorded requirement.**
  `next` states the rule and `11-requirements.md` explains it; `project add`
  accepts a split with no request on file. The interview check added to `check`
  catches the common case - material filed, nothing recorded - but not this one.
- **The order in `wiki/pages/setup-path.md` is unverified against a live
  console.** Every step is documented and every claim comes from its own page,
  but no source states the sequence, and three screenshots were opened out of
  the hundred-odd the screenshot index names. The claim worth testing first is
  that an HTTP API's credential has no home under Settings.
- **Open-question numbers still collide across branches.** Fixed within one
  file - an answered question keeps its number now - but two branches each
  number one past what their own copy shows, and the merge produces two rows
  with the same number and no conflict marker, because they are different lines.
  A number derived from a count cannot survive that; the raised date is already
  in the row and is the obvious material, and changing the format breaks every
  existing reference, which is why it was not done inside a bug fix.
- **`decision add` still needs an ASCII `--slug`.** A request and a task fall
  back to their ID when the title is Chinese; a decision has no ID, because its
  file name is the date plus the topic and `check` requires that shape. The
  error says so. Giving it a numbered fallback would produce
  `2026-09-02-decision-2.md`, which defeats the naming convention.
- **A Claude Code plugin in the customer repo.**
  `.claude-plugin/marketplace.json` + `plugins/asgard-fde/{commands,skills}`, as
  the demo generator does it. The `asgard-fde-onboarding` design-time skill does
  the routing part today, which was most of the value; the rest waits until the
  plugin format is worth committing to.
- **`--template-dir` to override embedded prompt text.** Prompts are embedded and
  versioned with the binary, which is the right default. This is for when
  iterating on prompt text against a live engagement turns out to be painful, and
  it has not yet.
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
