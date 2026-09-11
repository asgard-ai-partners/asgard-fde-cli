#!/usr/bin/env python3
"""Recompute the asgard-docs coverage row, and fail when the page disagrees.

`internal/corpus/wiki/index.md` states how much of the product documentation
this material is written from. That row is the one place a reader looks to
decide whether an absence means "not covered" or "not there", **so a wrong
denominator makes the material read better than it is** - and the two easiest
to confuse are the number of LINKS it writes and the number of PAGES there are.

So it is computed rather than counted. The page names the commit it was
measured at, and this measures at that same commit - `git ls-tree` on the
clone, so pulling does not change the answer.

    hack/check-coverage.py          measures at the commit the page names
    hack/check-coverage.py --head   also measures at the clone's HEAD

Needs $ASGARD_DOCS. See hack/sources.py.
"""

import pathlib
import re
import subprocess
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from sources import commit, resolve  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
INDEX = ROOT / "internal/corpus/wiki/index.md"

# The three directories the page excludes on purpose. Kept here because the
# check has to count them, and stated on the page with the reason for each.
EXCLUDED = re.compile(r"asgard-builtin/message-template|help-community/release-notes/|superpowers/")

# The row this checks, and the four numbers in it.
ROW = re.compile(
    r"\|\s*asgard-docs\s*\|[^|]*\|\s*(\d+)\s*/\s*(\d+)\s*cited at `([0-9a-f]{7,})`"
    r";\s*(\d+)\s*published and unread;\s*(\d+)\s*deliberately excluded")


def pages(docs: pathlib.Path, ref: str) -> set:
    """Every published page at one commit, as a slug.

    **`.md` and `.mdx` both.** Counting only `.mdx` gives 148 where the page
    says 162, which is what sent one reading of this down a false trail.
    """
    out = subprocess.run(["git", "-C", str(docs), "ls-tree", "-r", "--name-only", ref, "--", "docs"],
                         capture_output=True, text=True)
    if out.returncode != 0:
        sys.exit(f"{docs} has no commit {ref}. `git -C {docs} fetch` first.")
    got = set()
    for f in out.stdout.split():
        for ext in (".mdx", ".md"):
            if f.endswith(ext):
                got.add(f[len("docs/"):-len(ext)])
                break
    return got


def cited_slugs() -> set:
    """Every docs.asgard-ai.com page the material links to, as a slug."""
    bodies = []
    for pat in ("internal/corpus/**/*.md", "internal/stage/prompts/*.md",
                "internal/scaffold/templates/**/SKILL.md*"):
        for f in ROOT.glob(pat):
            bodies.append(f.read_text(errors="replace"))
    urls = set()
    for b in bodies:
        for u in re.findall(r"https://docs\.asgard-ai\.com/[A-Za-z0-9/_.#-]*[A-Za-z0-9/_-]", b):
            urls.add(u.split("#")[0].rstrip("/"))
    out = set()
    for u in urls:
        s = u.split("docs.asgard-ai.com/", 1)[1]
        out.add(s[len("docs/"):] if s.startswith("docs/") else s)
    return out


def measure(docs: pathlib.Path, ref: str):
    P = pages(docs, ref)
    cited = cited_slugs() & P
    return len(cited), len(P), len(P) - len(cited), len([p for p in P if EXCLUDED.search(p)])


def main() -> int:
    docs = resolve("docs")
    text = INDEX.read_text()
    m = ROW.search(text)
    if not m:
        sys.exit("the asgard-docs row in internal/corpus/wiki/index.md does not match the shape\n"
                 "this checks: `<cited> / <total> cited at `<commit>`; <n> published and unread;\n"
                 "<n> deliberately excluded`. Keep the shape or update this script.")
    claim = (int(m.group(1)), int(m.group(2)), int(m.group(4)), int(m.group(5)))
    ref = m.group(3)

    got = measure(docs, ref)
    names = ("cited", "total", "unread", "excluded")
    print(f"at {ref}, the commit the page names:")
    bad = 0
    for name, c, g in zip(names, claim, got):
        flag = "" if c == g else "   <-- page says " + str(c)
        bad += c != g
        print(f"  {name:<10} {g}{flag}")

    if "--head" in sys.argv:
        head = commit(docs)
        print(f"\nat {head}, the clone's HEAD:")
        for name, g in zip(names, measure(docs, head)):
            print(f"  {name:<10} {g}")
        print("\nA difference here is not a defect: the page records a reading, and the\n"
              "clone has moved since. It is the size of what re-reading would cover.")

    if bad:
        print(f"\n{bad} number(s) in the coverage row do not match a measurement at its own\n"
              "commit. Fix the row, or say what else it counts.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
