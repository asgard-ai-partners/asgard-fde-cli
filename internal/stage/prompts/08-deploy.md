# Deploy

The gate is green. Deployment is **CD-only**.

## Before the first deploy of an environment

This order cannot be reversed:

  1. tf-asgard creates the namespace and its app-secret for that env.
  2. The platform reconciles the namespace into a Project, which produces the
     platformMainEnvironmentId.
  3. You put that ID in projects/<project>/chart/values-<env>.yaml.
  4. Only then does deploy.yaml declare environments.<env>.

Declaring the env first means the next tag fails at helm upgrade.

**Whether a project needs at least one Syncer depends on its own CD, so read
the workflow rather than assuming.** After helm upgrade, CD triggers each
project's Syncers and waits, polling for CronJobs labelled syncer-name; it exits
1 after 180 seconds if none appear, even though the upgrade itself succeeded.

The difference between the reference repositories is one `if`, and **it is an
even split**: of the six that run this step, three check how many Syncers the
chart declares and skip it when that is zero, and three wait unconditionally and
fail any project that has none. Counted 2026-09-04. A production chart is
running today with zero Syncers under the first kind.

An even split is why this is written as "read your workflow" rather than as a
rule. There is no majority to assume, and the two behaviours are one line apart
in a file nobody opens after the repository is created.

    grep -n 'syncer-name' -A15 .github/workflows/*.y*ml

If what you find waits unconditionally, the project needs a Syncer before its
first tag - which a SkillSet or a knowledge drive creates. If it counts first,
it does not. **This tool warns either way**, because it cannot see your
workflow, and the warning says which.

## After the deploy, and before saying it is live

**A green deploy is not a working deployment**, and the step between them is not
in this repository at all.

Building a semantic model or an agent does **not** make it visible to anyone.
Somebody has to go to the **Management Console**, open the page for that product,
use *Manage Accounts in* to select the resource, and add the people. Every new
resource repeats it - **permissions do not inherit**, and nothing in a chart, in
`check`, in `verify` or in CD can see that the step was skipped.

    the symptom     "we deployed it and the customer says there is nothing there"
    the cause       not a chart problem, and looking for one wastes a day
    the fix         `asgard-cli wiki console` - which page, and which scope

Two things to settle before the day it goes live rather than on it:

  - **who in the customer's organisation can do this.** It is their Console and
    their Workspace. If nobody has been named, the deployment waits on an
    introduction
  - **who should see each resource.** Not everyone, usually - and the Console is
    where that is decided, because the platform has no answer for scoping what a
    caller may reach after they are in. `asgard-cli wiki platform-unknowns` P1

## Deploying

This is the first step that needs a remote at all. **Getting the repository onto
one is not your job**: Asgard is growing its own mechanism for provisioning a
customer repository and authenticating to it. If there is no remote yet, say so
and stop - do not `git init`, do not add one, and do not offer to.

    git tag -a dev-0.1.0 -m "dev-0.1.0"
    git push origin dev-0.1.0

  dev-x.y.z -> dev
  x.y.z     -> prod

One tag rolls out every project that declares that env; the rest are skipped.

**The prefix is the whole of the decision, and prod is the fallback.** CD tests
whether the ref starts with `dev-` and sends everything else to the production
cluster - so the pattern list in the workflow's `on.tags` is the only thing
standing between a stray tag and production. A tag named for a person, a date or
a ticket does not match the patterns and does nothing at all, which is safe; one
that happens to look like `1.2.3` is a production deploy.

**A tag deploys whatever commit it points at, not what is on the branch.** CD
checks out the tag ref. So a `dev-` tag on an unmerged branch head is a real way
to preview that branch on the dev cluster - and a bare-semver tag placed on the
wrong commit ships that commit to production, with a tag name that says nothing
about which one. Tag the commit you have just verified, and check what it points
at before pushing it:

    git tag -a dev-0.1.0 -m "dev-0.1.0" && git show --stat dev-0.1.0 | head -3

**A green deploy does not mean anything reconciled.** The chart is entirely
Asgard CRs with no Deployment or Pod, so CD runs `helm upgrade` without `--wait`
- there is no workload for it to wait on. What it waits for instead is the
Syncer: up to 180 seconds for the CronJob to appear, then up to 600 for one sync
to finish. That single sync is the only thing in the pipeline that proves the
platform accepted any of it - **so a project with no Syncer, on a CD that skips
the step, has nothing checking it at all.** Green there means helm returned.

**Values layer, and the tenant's file wins.** CD passes the shared
`common/values-<env>.yaml` first and the project's own second, so a key set in
both takes the project's. Putting something in the shared file and not seeing it
means some project overrode it, rather than that it was ignored.

## Do not helm upgrade from a laptop

A Syncer that pins revision to the chart's appVersion needs the release tag CI
stamps in. A local install renders the placeholder version as a git ref that does
not exist, and the Syncer then fails to clone every single run. asgard-cli render renders
only, and has no install path, for this reason.

## After deploying

  - A failed helm upgrade can leave the namespace **partially** applied. The
    previous revision stays "deployed", so the next upgrade proceeds without a
    rollback - do not read "the deploy failed" as "nothing changed".
  - Anything the platform created before helm did needs a one-time adoption, or
    helm refuses with invalid ownership metadata.
  - Record what still needs a human: uploading documents, triggering the first
    index run, pointing a front end at the new endpoint.

## Then close the loop

A behaviour change is not done until docs/spec/<<.SpecSlug>>/ carries it, and
each decision that got settled has its own dated record under docs/decisions/.
A task spec stops being read once it is done; the living spec is what the next
person reads.

## This is not the end of the work

The onboarding is what got the repo to its first deployment. **Everything after
that is enhancement**, and it has its own loop - which project a new capability
belongs to, whether it needs a spec first, and the closing step that gets
skipped:

    asgard-cli guide enhance

A new capability for a **new audience** is a new project, and it walks stages
3-8 again on its own.

**Between two pieces of work there is a state, and it is not a gap.** No request
open, no task open, nothing missing from any chart: what moves the work on then
is the customer, and the only question left to ask is what they want next.
`asgard-cli guide idle` is that state, and being in it is not being behind.

**Checked:** 2026-09-04 against the six reference repositories that run the
Syncer step, read directly rather than from memory: three guard on the Syncer
count and three do not, which corrects a statement written from a sample of two.
The label the step polls for is `asgard-ai.com/syncer-name` in all six. The
ordering above - namespace, then platformMainEnvironmentId, then values, then
deploy.yaml - is what tf-asgard and this tool each require, and declaring the
environment first is what makes the next tag fail at helm upgrade.

**Unchecked:** anything about a cluster. Nothing here has been run against one
from this repository, and the 180-second timeout, the reconcile into a Project
and the CronJob the platform creates are all read off the workflows and the
values files rather than observed. **The first deploy of a new environment is
this document's first real test**, and if the order is wrong the symptom is a
failed helm upgrade rather than anything subtle.
