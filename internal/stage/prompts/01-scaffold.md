# Write the repository skeleton

The repository skeleton has not been written yet.

**Read ahead before the interview, though.** The rest of this walk is written to
be read on arrival, and that breaks at the interview: it is the one stage whose
output is spoken rather than written, so a mistake there ends up in the
customer's notes rather than in a file you can edit. `asgard-cli brief customer-meeting`
lists the five things that reach a customer wrong and where the right version
lives - and it is addressed by what you are about to do rather than by which
stage you are at, so it is the same command the day before any meeting.

    asgard-cli scaffold

That writes AGENTS.md (the platform contract), the acceptance gate under
scripts/, the four-layer docs structure, the requirements/ and references/
entry points, the design-time skills, and the CD workflow. It is safe to re-run
at any point: it never overwrites a file that already exists.

Two of those decide where everything later ends up, and are worth reading once
before writing anything into the repo rather than looking up when it is already
in the wrong place:

    requirements/README.md   what may live in requirements/, and the tense test
                             that decides what belongs in the living spec instead
    docs/README.md           the four layers, and why each one exists

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
until a remote exists. Nothing is: everything up to the gate is local work, and none of
them needs one.

**Checked:** 2026-09-04 - **no platform claims to check.** Everything here
describes what `asgard-cli scaffold` writes and what it deliberately leaves out,
and the tool is its own source: run it in an empty directory and compare.

**Unchecked:** that the skeleton contains the right things. Which files a
customer repository needs on day one is this engagement's answer, and the parts
it leaves as TODO are the parts it decided cannot be guessed. Read the list as
one repository's shape rather than a standard, and the note about git as a
statement of scope - see the non-goals in TASK.md for why nothing here touches
a remote.
