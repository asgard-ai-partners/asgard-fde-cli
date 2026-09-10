#!/usr/bin/env python3
"""Every path this repository's own documents name has to exist.

`audit-material --paths` is the mirror of this and runs the other way round:
it reads the material that LANDS in a customer repository and fails on a path
only we have. This reads the documents that never land - Goal, README, AGENTS,
STRUCTURE, APPROACH, TASK - where naming our own paths is the whole point, and
fails when one of them is not there.

**It exists because AGENTS.md told a reader to run `hack/imperatives.py`**, a
script that has never been in this repository, and because five documents were
still describing the corpus as living under `pages/`. Nothing looked: those
files are read from a checkout rather than shipped, so `--paths` never sees
them, and a wrong path in the file that states the rules is worse than a wrong
path anywhere else.

Needs the checkout, which is why it is here and not in the binary.

    hack/check-doc-paths.py
"""

import pathlib
import re
import sys

DOCS = ["Goal.md", "README.md", "README.zh-TW.md", "AGENTS.md", "STRUCTURE.md",
        "APPROACH.md", "TASK.md"]

# A path inside this repository: a directory we own, then a file or a directory
# under it. Trailing `/` is a directory reference and is checked as one.
#
# **Backticks are not required.** The gate section writes its commands inside a
# fenced block, where a path carries none - and that is where the first version
# of this script failed to see `hack/imperatives.py`, the reference it was
# written for. So the match is on the shape of the path, anchored to a
# directory this repository owns, wherever it appears.
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


def main() -> int:
    root = pathlib.Path(__file__).resolve().parent.parent
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
    print(f"\n{checked} repository path(s) named, {bad} that are not there.")
    if bad:
        print("\nA document that names a file this repository does not have sends a\n"
              "reader to look for it. Fix the path, or delete the sentence.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
