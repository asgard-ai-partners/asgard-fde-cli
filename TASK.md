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
- **The `execute-script`, `validate-payload`, `generate-embedding` and
  `retrieve-knowledge` processors** appear in the platform's contract and in no
  chart that has been read. Either they are unused in practice, which is worth
  knowing, or the sample of deployments read so far is too small.

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
