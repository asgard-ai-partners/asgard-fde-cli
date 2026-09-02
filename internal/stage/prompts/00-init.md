No .asgard-config.json here, so this directory is not an onboarding yet.

## The shape, before anything else

**One workspace is one repository.** The workspace is the customer, and the
repository root *is* the workspace - that is what `init` fixes in place. Every
project lives inside it:

    <slug>-asgard-kube/          <- the workspace. One customer, one repo.
      .asgard-config.json        <- the workspace id and the project list
      projects/<project>/        <- every project, always here
      docs/ requirements/ common/ scripts/ .agents/

**One production deployment does it differently, and knowing why matters.** A
commerce middleware repository serves several **tenants** from one repo -
`tenants/<tenant>/chart/`, each with its own namespace and its own
`deploy.yaml`, and the CI matrix is built by scanning those files, so a tenant
can be prod-only and dev/prod are fully asymmetric. Namespaces are declared
there rather than derived from folder names.

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

**The slug**, and that is all. Short, lowercase, hyphens only. It is load
bearing:

    repository   <slug>-asgard-kube
    namespace    asgard-<slug>-<project>-<env>

Kubernetes caps a namespace at 63 characters and names derived from a namespace
inherit its length, so keep it short. Prefer the customer's own short name.

**The workspace id can wait.** It comes from the Asgard platform and nothing this
repository renders reads it - namespaces come from the slug - so not having one
should not stop the work that comes before it. Add it whenever it arrives:

    asgard-cli init --workspace-id <id>

On an already-initialised repository that fills the field in and changes nothing
else, so it needs no `--force`. `asgard-cli project add` and `asgard-cli check`
both say when it is still unset, because a project is what the platform deploys
and so is the point at which it is worth chasing.

When you do get it: it is a long decimal number (around 19 digits), **not** a
UUID - if what you have looks like `7ab7f523-3cd9-...`, it is the wrong value.
Ask for it, or read it off the platform. **Do not copy one from another
customer's repository or from an example**: it is a live production identifier,
and a wrong one binds this repository to somebody else's workspace.

**Do not ask about projects yet.** How the work splits is decided by interviewing
the customer, which is stage 2. Starting with none is the normal case.

## Start

**Run init in the directory that is to become the workspace.** In the common
case that is the directory you are already in - an empty one, or a repository
just cloned for this customer:

    asgard-cli init

**Do not create another directory level.** Check where you are before assuming.

Only if you are sitting in a *parent* directory - the place other
`*-asgard-kube` repos live - create the workspace first:

    mkdir <slug>-asgard-kube && cd <slug>-asgard-kube
    asgard-cli init

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
