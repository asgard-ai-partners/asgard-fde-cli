<!--
  The questions this is derived from are in AGENTS.md, "Before you say it is
  done". They are not repeated here - what is asked for below are the answers.

  The last section is the one that matters. Scope stated narrowly is a review
  that can do its job; scope stated widely is how something broken gets through.
-->

## What this changes

<!-- One paragraph. What is different afterwards, not a list of files. -->

## What was run

<!--
  Not what passes in CI - what you actually exercised.

      go build ./... && go vet ./... && gofmt -l internal/ cmd/

  Anything touching a command also gets driven end to end in a throwaway
  directory: init, scaffold, project add, add <kind>, check, render, verify.
  Anything touching a CR template gets its rendered output validated against
  asgard-kube/crd/. Say which of these you did.
-->

## Contract conformance

<!--
  Delete this section only if nothing here touches a CR template, an extract's
  YAML, or a claim about a platform field.

  FIRST: pull. Validating against a clone from three weeks ago proves nothing,
  and asgard-kube moves without announcing it.

      git -C <asgard-kube> fetch && git -C <asgard-kube> status -sb

  State the commit you validated against and whether it was head when you did.
  "Checked against the CRD" without a commit is not an answer.

  This repo does not define CRDs, it consumes them, so the question is not
  whether a CRD is right - it is whether what we now emit is still accepted by
  the contract as it stands.

  Render every kind and validate the documents against asgard-kube/crd/:
  required fields, fields not in the schema, enums, patterns, ExactlyOneOf. Do
  the same for the YAML skeletons in the extracts, since those are what somebody
  copies by hand. `hack/README.md` is the procedure; both counts come from
  `go run ./hack validate-crs`.

  If the contract HAS moved since the last time this repo looked, say what
  changed and what it means here - a retired field, a flipped default and a new
  required field each land differently.

  helm lint, `asgard-cli check` and a server-side dry-run do NOT do any of this.
  The dry-run is worse than silent: it drops a field it does not recognise and
  reports success, while helm's own server-side apply refuses.
-->

## What else says the same thing

<!--
  A fact lives in a template, an extract, a stage prompt and a wiki page, and a
  diff usually touches one of them. `grep <the thing>` finds the rest.

  Name the other copies you checked, or say the change has none.
-->

## What this does NOT do

<!--
  Required. What was left out, what was not verified, what is known to be
  incomplete. "Nothing" is a valid answer and an unusual one.

  If something belongs in TASK.md's "What is not done", put it there too - this
  box is read once, that file is read by whoever picks the work up.
-->
