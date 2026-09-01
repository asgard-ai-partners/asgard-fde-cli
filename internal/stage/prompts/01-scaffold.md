The repository skeleton has not been written yet.

    asgard-cli scaffold

That writes AGENTS.md (the platform contract), the acceptance gate under
scripts/, the four-layer docs structure, the design-time skills, and the CD
workflow. It is safe to re-run at any point: it never overwrites a file that
already exists.

Then confirm it landed:

    asgard-cli check

Done when that prints that the structure is consistent.

## Leave git alone

The directory not being a git repository yet is **not a problem to solve here**,
and neither is the missing remote. Asgard is growing its own mechanism for
provisioning a customer repository and authenticating to it, so do not run
`git init`, do not add a remote, and do not offer to - a repo wired up by hand
now is one that has to be unpicked when that mechanism lands.

The scaffolded CD workflow is tag-driven, so it looks like something is missing
until a remote exists. Nothing is: stages 2 to 7 are all local work, and none of
them needs one.
