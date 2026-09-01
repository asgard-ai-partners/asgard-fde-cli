The gate is green. Deployment is **CD-only**.

## Before the first deploy of an environment

This order cannot be reversed:

  1. tf-asgard creates the namespace and its app-secret for that env.
  2. The platform reconciles the namespace into a Project, which produces the
     platformMainEnvironmentId.
  3. You put that ID in projects/<project>/chart/values-<env>.yaml.
  4. Only then does deploy.yaml declare environments.<env>.

Declaring the env first means the next tag fails at helm upgrade.

**And every deployed project needs at least one Syncer.** CD triggers each
project's Syncers after helm upgrade and waits; it polls for CronJobs labelled
syncer-name and **exits 1 after 180 seconds if it finds none** - even when the
upgrade itself succeeded.

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

    asgard-cli next --stage enhance

A new capability for a **new audience** is a new project, and it walks stages
3-8 again on its own.
