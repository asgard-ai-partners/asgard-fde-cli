Every project has a read path and an entry point. Run the gate before committing
anything, and **stop at the first red step** - do not push past it.

Load the `asgard-cr-verification` skill under .agents/skills/ for the full
procedure. The four steps:

  1. asgard-cli check

     The repository's structure: the README project table against the folders on
     disk, deploy.yaml against the values files it names, the docs/ layers, the
     requirements indexes.

  2. helm lint projects/<project>/chart/app
     helm lint projects/<project>/chart/app -f projects/<project>/chart/values-dev.yaml

     **Run the bare one too.** It is the only step that proves values.yaml
     declares a default for every .Values.* a template reads. With an env file
     overlaid a missing default is masked, and then it nil-pointers for anyone
     running plain helm template.

  3. asgard-cli verify

     The invariants helm cannot see, because to helm these are opaque CRs: a
     dangling reference, a missing display annotation, a Workflow with no set
     labels, the agent split. A wrong **entry** name is as fatal as a wrong
     workflow name, and apply catches neither.

     With no arguments it does every project once per environment its deploy.yaml
     declares, rendering each one itself. Name a project to narrow it.

  4. Against the cluster you deploy to, read-only:

         asgard-cli render <project> <env> | kubectl apply --dry-run=server -n <ns> -f -
         asgard-cli render <project> <env> | python3 scripts/check_crd_fidelity.py - --context <ctx>

     **Both, because they answer different questions.** Dry-run answers "will it
     be accepted"; fidelity answers "**will it be kept**". CRDs silently prune
     fields they do not declare, so dry-run reports success while the field is
     discarded - and then helm's server-side apply fails in CD, which is far too
     late. A deprecated field once passed 25 of 25 dry-runs and broke the deploy.

     If no cluster is reachable, this step is **not run**. That is not the same
     as passing. Say so.

Steps 1 to 3 need only `helm` on PATH; step 4 also needs `kubectl` and python3.
`asgard-cli doctor` says which of those are installed, and how to install one
that is not.

Done when: every step is green, and you have said which ones could not be run.

Then `asgard-cli next` moves on to deploying. After that, new capability is added
with the loop in `asgard-cli next --stage enhance`.
