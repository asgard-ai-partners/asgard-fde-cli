# Before connect

before binding a fresh checkout to the platform, and before adding a second account or pipeline to one

**Every item below is a question for a person, and every one of them has an
answer the tool will appear to have already made.** That is the failure mode
this exists for: not a wrong answer, an unasked question.

A list of one reads as a default no matter what the prose next to it says, and
saying "nothing is guessed, including from a list of one" in six places did not
stop an agent binding the only pipeline in a workspace - its name resembled the
directory, and the person wanted a new pipeline against a different repository.
Read this as a checklist and carry an answer back for each line.

## which workspace

**Said:** the one whose name looks like the customer's

**True:** **Ask, always.** Nothing in the checkout says which workspace it deploys into - it is one of the two facts a repository cannot supply about itself, which is why `.asgard-cli.yaml` records it. An account can reach fifteen, and several will be plausible

Read `asgard-cli workspace list`.

## which provider account

**Said:** the one the workspace already has a connection to

**True:** **The account whose repositories this engagement is about**, which is often not the one connected months ago for something else. A workspace holds one connection per account and may hold several; `--account` names the one you mean, and defaults to the owner of this checkout's origin remote when there is one

Read `asgard-cli pipeline connections`.

## the app is not installed on that account yet

**Said:** pick an account from the list you already reach

**True:** **Installing is the answer, and it is a person's action on GitHub.** `pipeline connect --account <account>` prints the install URL for it - the app's slug is per-platform and a wrong slug is indistinguishable from a private app, so it comes from the platform rather than being guessed. On an organisation it may need one of its admins, which is a wait rather than a failure: say so and stop

Read `asgard-cli pipeline connect --help`.

## which repository

**Said:** the one this directory is named after

**True:** **Ask.** A fresh `init` checkout often has no remote at all, and a directory name is not a repository. `pipeline repos` lists what the installation actually grants - a repository missing from it was not granted, which is fixed on GitHub and not here

Read `asgard-cli pipeline repos`.

## a new pipeline, or one the workspace already has

**Said:** there is one pipeline and its name resembles the directory, so that is it

**True:** **This is the one that has actually gone wrong.** A workspace's existing pipeline may be for an entirely different repository, and one repository may carry several pipelines as long as their config paths differ. `pipeline use` and `pipeline create` are equally normal, and which one applies is not derivable from anything the tool can see

Read `asgard-cli pipeline list`.

## which platform project a release deploys into

**Said:** a workspace with no projects means something is broken

**True:** **An empty workspace is the normal start.** A release deploys into a platform project, and the project decides the namespace, so nothing deploys until one exists. `pipeline project create <name>` makes one with the default environment that `release create` requires - and it consumes account quota, so a refusal there is a subscription limit rather than a bad name

Read `asgard-cli pipeline projects`.

## how many releases the chart needs

**Said:** one release, because the chart is one chart

**True:** **Usually one per environment.** `<slug>-dev` and `<slug>-prod` name the same chart directory, differ by `on.pattern`, and are each created against a **different** platform project - which is what gives them different namespaces. One release is the shape for a POC nobody will maintain, and worth recording as a decision rather than arriving at by not asking

Read `../guide/projects.md`.

**Nothing here prompts.** These commands are run by an agent following a
skill, not by somebody at a terminal - `init` is the one command in this tool with
a person in front of it - so a question the tool asks is a dead path, and a
question it does not ask is one the agent has to. That is what this list is.

`asgard-cli gate` says which of these have been recorded, at any point.

**Checked:** every entry above is here because it actually happened, and each
names the document that carries the right version.

**Unchecked:** whether the list is complete. It grows when somebody gets
something new wrong, so an activity with few entries is not a safe one.
