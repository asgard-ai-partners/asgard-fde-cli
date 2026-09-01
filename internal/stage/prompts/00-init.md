No .asgard-config.json here, so this directory is not an onboarding yet.

## The shape, before anything else

**One workspace is one repository.** The workspace is the customer, and the
repository root *is* the workspace - that is what `init` fixes in place. Every
project lives inside it:

    <slug>-asgard-kube/          <- the workspace. One customer, one repo.
      .asgard-config.json        <- the workspace id and the project list
      projects/<project>/        <- every project, always here
      docs/ requirements/ common/ scripts/ .agents/

So there is no question of "where does a project go" - projects only ever exist
under the workspace root, and `asgard-cli project add` puts them there. The only
thing to decide up front is the workspace itself.

## What you need

**The workspace id**, from the Asgard platform. It is a long decimal number
(around 19 digits), **not** a UUID - if what you have looks like
`7ab7f523-3cd9-...`, it is the wrong value. It identifies the customer, and it is
the one value that cannot be derived or guessed. Get it before starting.

Ask for it, or read it off the platform. **Do not copy one from another
customer's repository or from an example**: it is a live production identifier,
and a wrong one binds this repository to somebody else's workspace.

**The slug.** Short, lowercase, hyphens only. It is load bearing:

    repository   <slug>-asgard-kube
    namespace    asgard-<slug>-<project>-<env>

Kubernetes caps a namespace at 63 characters and names derived from a namespace
inherit its length, so keep it short. Prefer the customer's own short name.

**Do not ask about projects yet.** How the work splits is decided by interviewing
the customer, which is stage 2. Starting with none is the normal case.

## Start

**Run init in the directory that is to become the workspace.** In the common
case that is the directory you are already in - an empty one, or a repository
just cloned for this customer:

    asgard-cli init --workspace-id <id>

**Do not create another directory level.** Check where you are before assuming.

Only if you are sitting in a *parent* directory - the place other
`*-asgard-kube` repos live - create the workspace first:

    mkdir <slug>-asgard-kube && cd <slug>-asgard-kube
    asgard-cli init --workspace-id <id>

### The slug comes from the directory name

`init` takes it from the current directory, dropping a trailing `-asgard-kube`,
so `acme-asgard-kube` and `acme` both give the slug `acme`. Pass
`--workspace-slug <slug>` when the directory is named something else - a generic
checkout directory, or a name with characters a namespace cannot take.

Check what it chose: `init` prints the slug and the namespaces derived from it
before you build anything on top.

## Then

    asgard-cli next

Ask it after every step. It works out where the onboarding is from the
repository itself, so it stays right no matter who did what.

## If you picked the slug wrong

Before any namespace exists, it is cheap: `asgard-cli init --force
--workspace-slug <new>` then `asgard-cli scaffold --force`, and rename the
directory. **After tf-asgard has created namespaces from it, it is not** - the
namespaces carry the slug, so settle it before that step.
