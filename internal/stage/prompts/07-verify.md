# Run the acceptance gate

Every project has a read path and an entry point. Run the gate before committing
anything, and **stop at the first red step** - do not push past it.

Load the `asgard-cr-verification` skill under .agents/skills/ for the full
procedure. It comes from the platform - `asgard-cli skill update` writes it -
so it states what the server this repository deploys to actually checks, rather
than what some version of it did when this tool was built. The four steps:

  1. asgard-cli check

     The repository's structure: the README project table against the folders on
     disk, .asgard-pipeline.yaml against the charts it names, the docs/ layers,
     the requirements indexes.

  2. helm lint projects/<project>/chart/app

     **Bare, with no -f.** It is the only step that proves values.yaml declares
     a default for every .Values.* a template reads. Overlay a file and a
     missing default is masked, and then it nil-pointers for anyone running
     plain helm template.

  3. asgard-cli verify

     The invariants helm cannot see, because to helm these are opaque CRs: a
     dangling reference, a missing display annotation, a Workflow with no set
     labels, the agent split. A wrong **entry** name is as fatal as a wrong
     workflow name, and apply catches neither.

     With no arguments it does every release the declaration names, rendering
     each one itself. Name a release to narrow it.

  4. The platform's plan. Push, then read it back:

         asgard-cli pipeline runs watch --release <name> --commit $(git rev-parse HEAD)

     **This is the step that cannot be run here, and it is not optional.** The
     plan renders with the release's real values and sends every CR to the
     apiserver with a server-side dry run, so CEL rules, patterns, required
     fields and unknown-field pruning are checked against the real cluster.
     Nothing local can do that: no cluster credential is issued to a client,
     which is exactly why steps 1 to 3 check a different class of thing - what
     a dry run passes and runtime still fails.

     It answers two questions the old local pair used to answer separately:
     `crd/dry-run-rejected` is "will it be accepted", `crd/unknown-field` is
     "**will it be kept**". A deprecated field once passed 25 of 25 plain
     dry-runs and broke the deploy, because CRDs prune what they do not declare
     while helm's server-side apply refuses it.

     If the run was never created, the push matched nothing - no pattern
     matched, or the release was never created on the platform.
     `asgard-cli pipeline deliveries` says which.

     **What this step is catching is written down.** `asgard-cli wiki crd-rules`
     lists the CEL validations the apiserver evaluates, which `helm lint` does
     not run and a dry-run does not report faithfully - and the one rule the
     schema cannot express at all. Read it before deciding a red deploy is a
     mystery.

Steps 1 to 3 need only `helm` on PATH. Step 4 needs a remote and a signed-in
session, and no cluster access at all. `asgard-cli doctor` says whether helm is
installed and how to install it.

Done when: every step is green, and you have said which ones could not be run.

Once every chart is complete, `asgard-cli project` says so for each of them
rather than a next step. After that, new capability is added
with the loop in `asgard-cli guide enhance`.

**Checked:** 2026-09-04 - each of the four steps names a command that exists and
does what is said: `asgard-cli check` is structural, `asgard-cli verify` renders
and checks the invariants a render cannot see, and step 4's dry-run and fidelity
scripts are the two the generated repo ships. The claim that a dry-run reports
success while dropping an undeclared field is the CRD's documented pruning
behaviour and is why step 4 is two commands rather than one.

**Unchecked:** that a step this cannot run is not a step that passed. Nothing
enforces saying so, and the failure it guards against - a green gate that never
reached a cluster - has happened once in the engagement this came from. Also
unchecked: that these four in this order are the whole gate. They are the gate
**this tool implements**; a deployment that fails for a reason none of them
looks at is the case that would disprove it, and there has been one - see
"A statement that shipped and was wrong" in TASK.md.
