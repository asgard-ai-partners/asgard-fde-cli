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


def main() -> int:
    binary = sys.argv[1] if len(sys.argv) > 1 else str(ROOT / ".out/asgard-cli")
    if not pathlib.Path(binary).exists():
        sys.exit(f"no binary at {binary}\n"
                 f"  go build -o .out/asgard-cli ./cmd/asgard-cli\n"
                 f"  A pass reads what is embedded, so build from the working tree first.")

    section = pass_section()
    missing = []

    for flag in sorted(audit_flags(binary)):
        if f"--{flag}" not in section:
            missing.append(f"--{flag}")

    for script in sorted((ROOT / "hack").iterdir()):
        if script.suffix not in (".py", ".sh"):
            continue
        if script.name == pathlib.Path(__file__).name:
            continue
        if f"hack/{script.name}" not in section:
            missing.append(f"hack/{script.name}")

    for step in GO_STEPS:
        if step not in section:
            missing.append(step)

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
            print(f"unlisted  {m} is a check here and TASK.md's pass does not name it")
    for n in leaked:
        print(f"leaked    .agents/skills/{n} is also in the scaffolded tree, so it ships")

    print(f"\n{len(audit_flags(binary))} flag(s), "
          f"{len([s for s in (ROOT / 'hack').iterdir() if s.suffix in ('.py', '.sh')]) - 1} script(s), "
          f"{len(GO_STEPS)} Go step(s); {len(missing)} unlisted, {len(leaked)} leaked.")
    if missing or leaked:
        print("\nAdd the row, or - if it is not a check - say so in NOT_A_CHECK here.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
