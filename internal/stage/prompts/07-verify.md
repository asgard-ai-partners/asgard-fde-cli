# Run the acceptance gate

Every project has a read path and an entry point. Run the gate before committing
anything, and **stop at the first red step** - do not push past it.

Run it with one command:

    asgard-cli gate

That is the whole local half, and it is one command on purpose: **a checklist
in prose is not a gate** - it goes stale where the binary cannot, and a reader
has no way to tell. A step it could not run is reported as skipped, which is
not a pass.

Load the `asgard-cr-verification` skill under .agents/skills/ for what the
PLATFORM checks, which is the other half and the authoritative one. It comes
from the platform - `asgard-cli skill update` writes it - so it states what the
server this repository deploys to actually checks, rather than what some
version of it did when this tool was built.

What `asgard-cli gate` runs, and what each step is for:

  tools    helm on PATH. Without it the chart steps cannot run, and the
           gate says skipped rather than passed.

  repo     the repository's structure: the README project table against the
           folders on disk, .asgard-pipeline.yaml against the charts it names,
           the docs/ layers, the requirements indexes. Alone: asgard-cli check

  shipped  whether the files this CLI wrote here still match the ones it
           carries - a page renamed upstream, or one edited here. It reads
           .asgard-scaffold.json, which records a digest and a CLI version per
           file. Alone: asgard-cli init, which reports and does not overwrite

  binding  whether this checkout records which workspace and which pipeline it
           deploys through. Nothing recorded is the ordinary state of a
           repository nobody has connected yet, and it is reported as skipped.
           Alone: asgard-cli pipeline show

  skills   whether the reference material here still describes the server this
           repository deploys to. Alone: asgard-cli skill status

  lint     helm lint on each chart, **with the reserved asgard block and
           nothing else**. That is what proves values.yaml declares a default
           for every .Values.* the chart itself owns; overlay an environment
           file and a missing default is masked until it nil-pointers for
           somebody running plain helm template. Linting with no -f at all
           fails on every chart that reads .Values.asgard.*, and a chart must
           not declare that block.

  render   each release renders, with placeholder platform values.

  verify   the invariants helm cannot see, because to helm these are opaque
           CRs: a dangling reference, a missing display annotation, a Workflow
           with no set labels, the agent split. A wrong **entry** name is as
           fatal as a wrong workflow name, and apply catches neither.
           Alone: asgard-cli verify [release]

Then the step that cannot be run here. Push, and read the plan back:

         asgard-cli pipeline runs watch --release <name> --ref <tag>

     **This is the step that cannot be run here, and it is not optional.** The
     plan renders with the release's real values and sends every CR to the
     apiserver with a server-side dry run, so CEL rules, patterns, required
     fields and unknown-field pruning are checked against the real cluster.
     Nothing local can do that: no cluster credential is issued to a client,
     which is exactly why steps 1 to 3 check a different class of thing - what
     a dry run passes and runtime still fails.

     It answers two questions the old local pair used to answer separately:
     `crd/dry-run-rejected` is "will it be accepted", `crd/unknown-field` is
     "**will it be kept**". A deprecated field once passed every plain
     dry run and broke the deploy, because CRDs prune what they do not declare
     while helm's server-side apply refuses it.

     If the run was never created, the push matched nothing - no pattern
     matched, or the release was never created on the platform.
     `asgard-cli pipeline deliveries` says which.

     **What this step is catching is written down.** `../wiki/crd-rules.md`
     lists the CEL validations the apiserver evaluates, which `helm lint` does
     not run and a dry-run does not report faithfully - and the one rule the
     schema cannot express at all. Read it before deciding a red deploy is a
     mystery.

`asgard-cli gate` needs only `helm` on PATH, plus a session for the one step
that asks the platform (`--offline` skips it). Reading the plan needs a remote
and a signed-in session, and no cluster access at all. `asgard-cli doctor` says whether helm is
installed and how to install it.

Done when: every step is green, and you have said which ones could not be run.

**No command says a chart is complete, and `asgard-cli project` refuses to.**
It lists what each chart declares and says nothing about what it lacks: a chart
with a SemanticLayer and no Agent may be finished or unfinished, and the files
cannot tell the two apart. Completeness is your judgement against the request,
and after it new capability is added with the loop in `../guide/enhance.md`.

**Checked:** 2026-09-11 - the steps above are `gate`'s own, in its own
order, read off asgard-fde-cli's own `internal/cli/gate.go`; each names the command that runs it
alone and each of those exists. This line described a four-step checklist that
the body above had already replaced, and named dry-run and fidelity scripts as
"the two the generated repo ships" when the scaffold ships neither. The claim
that a dry-run reports success while dropping an undeclared field is the CRD's
documented pruning behaviour.

**Unchecked:** what the platform's own half reports. The steps that need a
session are the ones nobody has held this page against.

**Unchecked:** that a step this cannot run is not a step that passed. Nothing
enforces saying so, and the failure it guards against - a green gate that never
reached a cluster - has happened once in the engagement this came from. Also
unchecked: that these eight in this order are the whole gate. They are the gate
**this tool implements**; a deployment that fails for a reason none of them
looks at is the case that would disprove it, and there has been one: a chart
that passed every step here and failed in the platform's own dry run, on a
field the CRD prunes rather than rejects.
