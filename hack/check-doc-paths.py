#!/usr/bin/env python3
"""Every path and Go symbol this repository's own documents name has to exist.

`audit-material --paths` is the mirror of this and runs the other way round:
it reads the material that LANDS in a customer repository and fails on a path
only we have. This reads the documents that never land - Goal, README, AGENTS,
STRUCTURE, APPROACH, TASK - where naming our own paths is the whole point, and
fails when one of them is not there.

**Those files are read from a checkout rather than shipped, so `--paths` never
sees them** - and a wrong path in the file that states the rules is worse than
a wrong path anywhere else.

Needs the checkout, which is why it is here and not in the binary.

    hack/check-doc-paths.py
"""

import pathlib
import re
import sys

# The seven root documents, plus this directory's own - `hack/README.md` and the
# scripts, which tell somebody what to run and had the last stale reference:
# `verify-references.sh` described where an exemption lives in a file deleted
# from all four reference repositories.
#
# **This file is not in the list.** It names the deleted symbols it was written
# to catch, in the paragraphs explaining why it resolves a symbol inside its own
# package - so checking itself reports six findings, all of them its own
# subject. Same reason `internal/cli/self.go` parses Go source instead of
# grepping it: a note recording that something was removed must not read as
# naming it.
DOCS = ["Goal.md", "README.md", "README.zh-TW.md", "AGENTS.md", "STRUCTURE.md",
        "APPROACH.md", "TASK.md", "CLAUDE.md",
        "hack/README.md", "hack/check-tables.py", "hack/sources.py",
        "hack/validate-crs.py", "hack/extract-crs.py", "hack/verify-references.sh",
        ".env.example", ".agents/skills/consistency-checks/SKILL.md"]

# A path inside this repository: a directory we own, then a file or a directory
# under it. Trailing `/` is a directory reference and is checked as one.
#
# **Backticks are not required.** The gate section writes its commands inside a
# fenced block, where a path carries none, so the match is on the shape of the
# path - anchored to a directory this repository owns - wherever it appears.
PATH = re.compile(
    r"(?:^|[^A-Za-z0-9_./`-])"
    r"((?:source|hack|internal|cmd|prompts|projects|docs|assets|plugins|\.github)/"
    r"[A-Za-z0-9_./*<>-]*)")

# Placeholders the scaffold expands at write time, and globs. Neither is a path
# on disk here, and both are correct in prose.
def skip(p: str) -> bool:
    return ("<" in p or "*" in p or "__" in p
            # A path inside a customer's repository, not ours.
            or p.startswith(("projects/", "docs/", "assets/")))


# A Go symbol written `pkg.Symbol`. The docs cite these as the place a rule is
# implemented, and a deleted one sends a reader looking for it: `config.Find`
# outlived the whole `config` package, and `kb.Rank` outlived the search engine.
SYMBOL = re.compile(r"`([a-z][a-z0-9]*)\.([A-Z][A-Za-z0-9_]*)`")


def go_files(root: pathlib.Path):
    return [f for f in root.rglob("*.go") if ".out" not in f.parts]


def go_packages(root: pathlib.Path):
    """The package names a `pkg.Symbol` reference could legitimately name:
    every directory under internal/, plus the last segment of every import
    path in the tree, plus the module root's own package.

    A reference to a package outside that set names nothing - which is how
    `config.Find` survived the deletion of the whole `config` package.
    """
    names = {d.name for d in (root / "internal").iterdir() if d.is_dir()}
    names.add("selfsrc")
    for f in go_files(root):
        for line in f.read_text(errors="replace").split("\n"):
            line = line.strip().strip("_ ").strip()
            if line.startswith('"') and line.endswith('"'):
                names.add(line.strip('"').split("/")[-1])
    return names


DECL = "func|type|var|const"


def main() -> int:
    root = pathlib.Path(__file__).resolve().parent.parent
    pkgs = go_packages(root)
    internal = {d.name: d for d in (root / "internal").iterdir() if d.is_dir()}
    internal["selfsrc"] = root
    bad = 0
    checked = 0
    for name in DOCS:
        f = root / name
        if not f.exists():
            print(f"missing document  {name}")
            bad += 1
            continue
        for i, line in enumerate(f.read_text().split("\n"), 1):
            for m in PATH.finditer(line):
                p = m.group(1)
                if skip(p):
                    continue
                checked += 1
                if not (root / p.rstrip("/")).exists():
                    print(f"gone  {name}:{i} -> `{p}`")
                    bad += 1
            for m in SYMBOL.finditer(line):
                pkg, sym = m.group(1), m.group(2)
                checked += 1
                if pkg not in pkgs:
                    print(f"gone  {name}:{i} -> `{pkg}.{sym}` (no such package)")
                    bad += 1
                    continue
                if pkg not in internal:
                    # Somebody else's package. Whether it has that symbol is
                    # their business and their version's.
                    continue
                # **Resolved inside the package that owns it.** A search of the
                # whole tree cannot tell `usecase.Index` from `strings.Index`,
                # and passed the first for months after it was deleted.
                own = "\n".join(f.read_text(errors="replace")
                                for f in internal[pkg].glob("*.go"))
                if not re.search(rf"\b(?:{DECL})\s+(?:\([^)]*\)\s*)?{re.escape(sym)}\b", own) \
                        and not re.search(rf"^\t{re.escape(sym)}\s", own, re.M):
                    print(f"gone  {name}:{i} -> `{pkg}.{sym}`")
                    bad += 1
    print(f"\n{checked} path(s) and symbol(s) named, {bad} that are not there.")
    if bad:
        print("\nA document that names a file or a symbol this repository does not have\n"
              "sends a reader to look for it. Fix it, or delete the sentence.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
