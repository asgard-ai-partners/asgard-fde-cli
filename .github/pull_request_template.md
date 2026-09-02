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

## What else says the same thing

<!--
  A fact lives in a template, an extract, a stage prompt and a wiki page, and a
  diff usually touches one of them. `asgard-cli find <the thing>` finds the rest.

  Name the other copies you checked, or say the change has none.
-->

## What this does NOT do

<!--
  Required. What was left out, what was not verified, what is known to be
  incomplete. "Nothing" is a valid answer and an unusual one.

  If something belongs in TASK.md's "What is not done", put it there too - this
  box is read once, that file is read by whoever picks the work up.
-->
