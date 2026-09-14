# Deploy

The local gate is green. **The platform runs the rollout**, and it is the half
that decides whether the change is deployable at all.

## Before the first deploy of a release

Two things have to exist on the platform, and neither is in the repository:

  1. **The release**, bound to the project whose namespace it deploys into:

         asgard-cli pipeline release create <name> --project <id>

     Creating it provisions that release's own Secret and ConfigMap, its deploy
     identity and its RBAC. A tag matching a release nobody created produces **no
     run at all** - `asgard-cli pipeline deliveries` says so, and nothing else
     will.

  2. **The values the declaration names.** A required key with no value fails the
     plan with `vars/required-missing`:

         asgard-cli pipeline variables list --release <name>

The ordering trap that used to live here - namespace first, then the environment
id, then the values file, then the declaration - is gone. The namespace comes
from the project the release binds, and the environment id is injected as
`.Values.asgard.projectEnvironmentId`; neither is a value anybody fetches and
pastes any more.

**Whether the rollout waits on a Syncer is a property of the chart.** Apply
triggers the Syncers this release deployed that carry
`asgard-ai.com/auto-fire-on-rollout: "true"`, and waits for them on one shared
budget. A chart is running today with zero of them.

**A repository this tool writes has no CD workflow of its own**, and no cluster
credential to run one with - the rollout is the platform's, and there is no
`.github/workflows/` here to read. So the question is not what your workflow
does with a Syncer count; it is whether the chart has one at all, and what a
release with none actually proves.

With none, nothing runs after the dry run: a succeeded run means helm returned.
`asgard-cli gate` warns about exactly that, and a SkillSet or a knowledge drive
is what creates the first Syncer. When the shape genuinely has none - a
DataConnector and a SemanticLayer is one - the warning is correct and the
read-back is `asgard-cli pipeline manifest --release <name> --status`, which the
warning itself names.

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
    the fix         `../wiki/console.md` - which page, and which scope

Two things to settle before the day it goes live rather than on it:

  - **who in the customer's organisation can do this.** It is their Console and
    their Workspace. If nobody has been named, the deployment waits on an
    introduction
  - **who should see each resource.** Not everyone, usually - and the Console is
    where that is decided, because the platform has no answer for scoping what a
    caller may reach after they are in. `../wiki/platform-unknowns.md` P1

## Deploying

This is the first step that needs a remote at all. **Getting the repository onto
one is not your job**: Asgard is growing its own mechanism for provisioning a
customer repository and authenticating to it. If there is no remote yet, say so
and stop - do not `git init`, do not add one, and do not offer to.

    git tag -a dev-0.1.0 -m "dev-0.1.0"
    git push origin dev-0.1.0
    asgard-cli pipeline runs watch --release <name> --ref dev-0.1.0

**Which tag reaches which release is in `.asgard-pipeline.yaml`, one RE2 pattern
per release.** There is no global convention any more and no fallback: a tag that
matches nothing produces nothing, and the patterns are readable in the file
rather than in a workflow's `on.tags`. One tag can match several releases and
produce a run for each.

**Read the patterns before tagging.** A release whose pattern is a bare semver
deploys to whatever project that release binds, and the tag name says nothing
about which. `asgard-cli pipeline show` lists every release with its trigger.

**A tag deploys whatever commit it points at, not what is on the branch**, and
the run reads its declaration from that commit too. So a tag on an unmerged
branch head is a real way to preview that branch - and a tag placed on the wrong
commit ships that commit. Tag the commit you have just verified, and check what
it points at before pushing it:

    git tag -a dev-0.1.0 -m "dev-0.1.0" && git show --stat dev-0.1.0 | head -3

**A succeeded run does not mean anything reconciled.** The chart is entirely
Asgard CRs with no Deployment or Pod, so there is no workload to wait on. What
Apply waits for instead is the Syncers marked `auto-fire-on-rollout`: it fires
them all first and then waits on one shared budget rather than restarting the
clock per job, so N stuck syncers cost one budget and not N. A release with none
has nothing checking it beyond the dry run, and green there means helm returned.

**The plan is where a change is actually judged.** It reports the resource diff
per CR, the variable changes, and the server-side dry run's verdict on every one
of them. Reading it is not a formality after the fact - a run stops at review
precisely so it can be read before anything is written.

**Values are on the platform, not in the repository.** Two releases of one chart
differ by what was set on each. A value that is not taking effect is either not
declared (an orphan, which is never injected and shows in the plan as a warning)
or was set on a different release.

## Do not helm upgrade from a laptop

A Syncer that pins revision to the chart's appVersion needs the ref the platform
stamps in. A local install renders the placeholder as a git ref that does not
exist, and the Syncer then fails to clone every single run. `asgard-cli render`
renders only, with placeholder values, and has no install path, for this reason.

There is also nothing to install with: **no cluster credential is ever issued to
a client.** That is the same reason the plan's dry run cannot be reproduced
locally, and why the local gate checks a different class of thing.

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

    ../guide/enhance.md

A new capability for a **new audience** is a new project, and it walks stages
3-8 again on its own.

**Between two pieces of work there is a state, and it is not a gap.** No request
open, no task open, nothing missing from any chart: what moves the work on then
is the customer, and the only question left to ask is what they want next.
`../guide/idle.md` is that state, and being in it is not being behind.

**Checked:** 2026-09-04 against the platform's own behaviour spec and against a
real rollout on dev: a tag pushed to a bound repository produced a run within
seconds, the plan reported 29 CRs to create and stopped at review, and Apply
wrote them and then failed at the Syncer step with the reason named. The
auto-fire label is `asgard-ai.com/auto-fire-on-rollout` and is read only by the
runner. The ordering trap that this page used to describe - namespace, then
environment id, then values, then declaration - no longer exists: both halves are
injected.

**Unchecked:** anything about a cluster. Nothing here has been run against one
from this repository, and the reconcile into a Project and the CronJob the
platform creates are read off the reference deployments' own values files rather
than observed. Those deployments predate this tool and carry CD workflows of
their own; **a repository this writes does not**, which is why no instruction
here sends a reader to grep one. **The first deploy of a new environment is
this document's first real test**, and if the order is wrong the symptom is a
failed helm upgrade rather than anything subtle.
