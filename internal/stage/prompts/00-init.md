# Start the onboarding

No .asgard-pipeline.yaml here, so this directory is not an onboarding yet.

## The shape, before anything else

**One workspace is one repository.** The workspace is the customer, and the
repository root *is* the workspace. Every project lives inside it:

    <slug>-asgard-kube/          <- the workspace. One customer, one repo.
      .asgard-pipeline.yaml      <- what deploys, and the only file the platform reads
      .asgard-cli.yaml           <- which workspace and pipeline this checkout is bound to
      projects/<project>/        <- every project, always here
      docs/ requirements/ .agents/

**One production deployment does it differently, and knowing why matters.** A
commerce middleware repository serves several **tenants** from one repo -
`tenants/<tenant>/chart/` - so a tenant can be prod-only and the environments
are fully asymmetric. That shape is a set of releases now: one declaration lists
them all, each binding a chart to the platform project whose namespace it
deploys into, and nothing derives a namespace from a folder name.

That shape is right when the tenants are **instances of one product** - the same
chart, the same skills, different customers of a platform we built. It is wrong
for what this tool is for: a customer whose systems, audience and read path are
their own shares nothing with the next one, and putting two of them in one
repository means every change is reviewed against a customer it does not affect.

**Take the exception only when the second tenant is the same product.** If you
are asking the question, it is not.

So there is no question of "where does a project go" - projects only ever exist
under the workspace root, and `asgard-cli project add` puts them there. The only
thing to decide up front is the workspace itself.

## What you need

**Two ids, and nothing else.** Which workspace this checkout deploys into and
which pipeline it deploys through are the only two facts no file in the
repository implies and the platform cannot be asked for on your behalf:

    asgard-cli workspace list
    asgard-cli pipeline list --workspace <id>

Everything else about the repository - which projects it has, what each chart
declares, where each chart deploys - is read off the repository itself or held
on the platform.

**Neither id is ever guessed**, including when there is only one candidate. A
command with nothing recorded lists what there is and stops. That costs a step
on a customer with one workspace, and it buys the property that matters: a rule
which resolves while a list holds one entry starts resolving to something nobody
chose on the day it holds two.

An id is a long decimal number (around 19 digits), **not** a UUID - if what you
have looks like `7ab7f523-3cd9-...`, it is the wrong value. **Do not copy a
workspace id from another customer's repository or from an example**: it is a
live production identifier, and a wrong one binds this repository to somebody
else's workspace.

There is no pipeline yet on a genuinely new customer. `asgard-cli pipeline
connections` and `asgard-cli pipeline create` make one; that is a decision about
the platform, taken before the repository exists.

**Do not ask the customer about projects yet.** How the work splits is decided by interviewing
the customer - `asgard-cli guide projects`. Starting with none is the
normal case.

## Start

**Run it in the directory that is to become the repository.** In the common case
that is the directory you are already in - an empty one, or a repository just
cloned for this customer:

    asgard-cli init --workspace <id> --pipeline <id>

**Do not create another directory level.** Check where you are before assuming.

That one command does the three things this step used to list separately, in
the only order they work in:

    1  asgard-cli scaffold        the skeleton, including .asgard-pipeline.yaml
    2  the binding                .asgard-cli.yaml beside that declaration
    3  asgard-cli skill update    what this platform's CRs actually accept

Each is still its own command and each can be re-run alone. They are composed
because a list of three steps kept in prose is a list that goes stale - and this
one did.

Run it again after an upgrade: on a repository that already records both ids,
neither flag is needed, existing files are left alone, and step 3 brings the
reference material up to whatever the server now serves.

## The repository name

Convention is `<slug>-asgard-kube`, and **nothing reads it**. Namespaces come
from the platform project a release binds to, not from any name on disk, and no
file records a slug. Rename the directory whenever you like; nothing has to be
corrected afterwards.

## Then

`asgard-cli guide scaffold` is what the skeleton step is deciding - what it
contains, what it deliberately does not, and the one thing to read ahead of
rather than on arrival.

    asgard-cli gate

That is the one command to run after changing anything, and its `binding` step
is what catches a checkout that names a workspace and no pipeline.

    asgard-cli project

    asgard-cli project

That reads what each chart declares off the repository itself, so it stays right
no matter who did what. **It reports no step**, because there is none.

**Checked:** 2026-09-05 - nothing here to check against a source. This document
makes **no claim about the platform**: it is the repository shape this tool
writes. Three claims it used to make are retired: that namespaces carry a
workspace slug, that a slug is recorded at all, and that the pipeline follows
from the checkout's origin remote. `.asgard-config.json` is gone, and with it
the slug, the project list and the per-project shape; the pipeline is a recorded
choice now, and nothing here reads a git remote.

**Unchecked:** that one workspace is one repository, and that the tenant-per-
directory shape one production repository uses is right for a platform's
customers and wrong here. That is a judgement about **which engagements this
tool is for**, taken from the one it was built in. Nothing enforces it, and a
reader who is onboarding several tenants of one product should read the section
above rather than the rule.
