#!/usr/bin/env python3
"""TASK.md's consistency pass must name every check this repository has.

The skill says to derive the list from `AGENTS.md`'s inventory. That is only as
complete as the inventory, and a list derived faithfully from an incomplete
source is incomplete in the same places - which is the failure the written list
exists to stop, one level up.

So the list is checked against the binary and the directory rather than against
another document: every `audit-material` flag the command tree actually has,
every script in `hack/`, and the four Go steps.

    hack/check-pass-list.py

Also refuses a repository's own maintenance skill appearing in a scaffolded
tree. `.agents/skills/` at the root is ours and never ships; the ones that do
are under `internal/scaffold/templates/`. The two directories have the same
name, which is the whole reason to check.
"""

import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
TASKS = ROOT / "TASK.md"

# Flags that are not checks: they change what a check reads, or take a value.
NOT_A_CHECK = {"template-dir", "help", "format"}

# Flags that have no pass or fail: listings for a person, and one query.
NOT_A_VERDICT = {"orphans", "crossref", "ask", "unmarked", "unchecked", "term"}

# What each script needs to run, because that is the only thing about the list
# that is judgement rather than discovery. A script absent from here still
# appears - under "ungrouped", which is the prompt to say what it needs.
NEEDS = {
    "check-doc-paths.py": "this repository",
    "check-pass-list.py": "this repository",
    "check-goal.py": "this repository - **the capability, not the material**",
    "check-coverage.py": "$ASGARD_DOCS",
        "check-processors.py": "$ASGARD_CORE and $ASGARD_DOCS",
    "check-counts.py": "$ASGARD_DEPLOYMENTS",
    "spec-key-gap.py": "the clones, helm and a built binary",
    "sources.py": "the clones",
    "extract-crs.py": "$ASGARD_KUBE",
    "validate-crs.py": "$ASGARD_KUBE",
    "verify-references.sh": "the clones",
}

# The Go steps. They are not discoverable from a binary, so they are named -
# and named here rather than in TASK.md, so this file is the only list.
GO_STEPS = ["go build", "go vet", "gofmt", "go test"]


def pass_section() -> str:
    text = TASKS.read_text()
    start = text.index("## The consistency pass")
    end = text.index("\n## ", start + 10)
    return text[start:end]


def audit_flags(binary: str) -> set:
    out = subprocess.run([binary, "audit-material", "--help"],
                         capture_output=True, text=True)
    if out.returncode != 0:
        sys.exit(f"{binary} audit-material --help failed:\n{out.stderr}")
    got = set(re.findall(r"^\s+--([a-z][a-z-]*)", out.stdout, re.M))
    # A flag taking a value shows as `--term string`; it is still a check.
    return {f for f in got if f not in NOT_A_CHECK}


def scripts() -> list:
    return sorted(s.name for s in (ROOT / "hack").iterdir() if s.suffix in (".py", ".sh"))


def go_checks() -> list:
    """The subcommands of the Go gate, asked for rather than listed here.

    **`hack/` is Go now** - see AGENTS.md - and each check that moves across
    stops being a file in this directory and becomes a subcommand. Asking the
    binary is the same discipline as asking `audit-material --help` for its
    flags: a list written here would be the copy this file exists to refuse.
    """
    out = subprocess.run(["go", "run", "./hack", "list"], cwd=ROOT,
                         capture_output=True, text=True)
    if out.returncode != 0:
        sys.exit(f"`go run ./hack list` failed, so the Go half of the pass cannot be read:\n{out.stderr}")
    found = []
    for line in out.stdout.split("\n"):
        m = re.match(r"^  (\S+)  ", line)
        if m:
            found.append(m.group(1))
    return found


def show(binary: str) -> int:
    """Print the pass, derived from the binary and this directory.

    **TASK.md used to carry this as a table.** It was a hand-written copy of
    what this file already computes, which is the one shape of claim this
    repository has decided not to keep: a list that can be generated is not
    written down. So the pass is printed and TASK.md points here.
    """
    flags = sorted(audit_flags(binary))
    print("The consistency pass, most-volatile first. No state is recorded for a")
    print("check anywhere: the answer is its exit code, today.\n")

    print("1. upstream - pull first, or these check a clone rather than the platform")
    for name in go_checks():
        print(f"     go run ./hack {name}")
    for name in scripts():
        need = NEEDS.get(name, "")
        if "$" in need or "clone" in need:
            print(f"     hack/{name:<24} needs {need}")
    print("\n2. the material - build from the working tree first, or you audit an older corpus")
    for f in flags:
        if f not in NOT_A_VERDICT:
            print(f"     asgard-cli audit-material --{f}")
    print("\n3. this repository")
    for name in scripts():
        if NEEDS.get(name, "").startswith("this repository"):
            print(f"     hack/{name:<24} {NEEDS[name][len('this repository'):].lstrip(' -') or ''}")
    for name in scripts():
        if name not in NEEDS:
            print(f"     hack/{name:<24} ungrouped - say what it needs in NEEDS here")
    print("\n4. the compiler")
    for step in GO_STEPS:
        print(f"     {step} ./...")
    print("\n5. the network, last, because it is the only one that needs it")
    print("     asgard-cli audit-material --urls")

    print("\nListings, not checks - they do not fail:")
    print("     " + "  ".join(f"--{f}" for f in flags if f in NOT_A_VERDICT))
    return 0


def main() -> int:
    binary = sys.argv[1] if len(sys.argv) > 1 and not sys.argv[1].startswith("-") \
        else str(ROOT / ".out/asgard-cli")
    if not pathlib.Path(binary).exists():
        sys.exit(f"no binary at {binary}\n"
                 f"  go build -o .out/asgard-cli ./cmd/asgard-cli\n"
                 f"  A pass reads what is embedded, so build from the working tree first.")

    if "--list" in sys.argv:
        return show(binary)

    section = pass_section()
    missing = []

    # **The names are not compared any more**, because the pass is printed from
    # this file rather than copied into TASK.md. What is still compared is the
    # part no program can derive: the prose surfaces, below. A flag or script
    # missing from NEEDS shows up in `--list` as ungrouped instead.
    for name in scripts():
        if name not in NEEDS and name != pathlib.Path(__file__).name:
            missing.append(f"hack/{name} is in this directory and NEEDS does not say what it needs")

    # **The prose surfaces correspond by key, not by wording.** They are
    # derived rather than discovered, so the two lists drift - and comparing
    # their prose drifts with them: the same surface is worded for its own
    # context and filed under the group that suits it. Each row carries a
    # slug, and only the slugs are compared.
    def keys(text, heading):
        sec = text[text.index(heading):]
        return set(re.findall(r"^\| `([a-z][a-z0-9-]*-[a-z0-9-]+)`", sec, re.M))

    agents = (ROOT / "AGENTS.md").read_text()
    inventory = keys(agents, "**Checked by nothing, and verified by reading.**")
    listed = keys(section, "## The consistency pass")
    for k in sorted(inventory - listed):
        missing.append(f"prose surface `{k}` is in AGENTS.md's inventory and not in the pass")
    for k in sorted(listed - inventory):
        missing.append(f"prose surface `{k}` is in the pass and not in AGENTS.md's inventory")

    # The maintenance skill must not be in the scaffolded tree.
    shipped = ROOT / "internal/scaffold/templates/.agents/skills"
    ours = {d.name for d in (ROOT / ".agents/skills").iterdir() if d.is_dir()} \
        if (ROOT / ".agents/skills").is_dir() else set()
    leaked = sorted(n for n in ours if (shipped / n).exists())

    for m in missing:
        if m.startswith("prose surface"):
            print(f"adrift    {m}")
        else:
            print(f"ungrouped {m}")
    for n in leaked:
        print(f"leaked    .agents/skills/{n} is also in the scaffolded tree, so it ships")

    print(f"\n{len(audit_flags(binary))} flag(s), {len(scripts()) - 1} script(s), "
          f"{len(GO_STEPS)} Go step(s); {len(missing)} ungrouped or adrift, {len(leaked)} leaked.")
    print("`hack/check-pass-list.py --list` prints the pass itself.")
    if missing or leaked:
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
