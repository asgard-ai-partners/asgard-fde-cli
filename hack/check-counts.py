#!/usr/bin/env python3
"""Recompute the counts this material asserts about a reference deployment.

**A number copied out of a document that states its own count is the cheapest
thing here to get wrong**, and the most expensive to notice: nothing about "88"
reads differently from "93". One pass that set out to recount SHOPLINE's
back-office map took a figure off a different tally and wrote it into seven
places, where it sat for a week looking exactly as authoritative as the truth.

So these are computed. Each row below says where the number lives upstream, how
to arrive at it, and every regular expression in this material that states it -
**and a claim whose wording has drifted out of every pattern is a failure**, not
a pass, because that is how a count stops being checked without anybody
deciding to stop checking it.

`hack/check-coverage.py` is the same idea for the asgard-docs coverage row.

    hack/check-counts.py           against the clones as they stand
    hack/check-counts.py --dump    print what upstream counts, and stop

Needs $ASGARD_DEPLOYMENTS. See hack/sources.py.
"""

import argparse
import pathlib
import re
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from sources import commit, resolve  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
MATERIAL = [ROOT / "internal/corpus", ROOT / "internal/needs", ROOT / "internal/stage"]

SEPARATOR = re.compile(r'^\|[\s:|-]+\|$')


def table_rows(text: str) -> list:
    """Data-row counts of every markdown table in a document, in order.

    A table is what follows a separator row, which is the only line a markdown
    table is obliged to have and the only one whose shape is unambiguous.
    Counting `|`-prefixed lines instead counts headers and separators too, and
    counting them across a whole file merges the tables - both of which turn a
    ledger of 88 into some other number that still looks like a count.
    """
    out, current = [], None
    for line in text.split("\n"):
        line = line.rstrip()
        if SEPARATOR.match(line):
            current = 0
            out.append(current)
            continue
        if line.startswith("|"):
            if out and current is not None:
                out[-1] += 1
        else:
            current = None
    return out


def biggest_table(path: pathlib.Path) -> int:
    """The largest ledger in a document. The explanatory tables are smaller."""
    return max(table_rows(path.read_text()), default=0)


def files_in(path: pathlib.Path) -> int:
    return len(sorted(p for p in path.iterdir() if p.suffix == ".md"))


# Each count: where it comes from, how, and every way this material states it.
COUNTS = [
    {
        "slug": "shopline-l1-pages",
        "clone": "asgard-freyr-skills",
        "of": "shopline-backoffice/references/page-map.md",
        "how": biggest_table,
        "what": "L1 page entry points in the back-office page ledger",
        "says": [
            r"(\d+) L1 page entry points",
            r"That (\d+) is the count",
            r"That (\d+)-page map",
            r"describing (\d+) pages",
            r"covering (\d+) pages",
            r"costs: (\d+) pages",
            r"costs: (\d+) menu-level page entry points",
        ],
    },
    {
        "slug": "shopline-operations",
        "clone": "asgard-freyr-skills",
        "of": "shopline-backoffice/references/operation-map.md",
        "how": biggest_table,
        "what": "rows in the back-office operation ledger",
        "says": [r"(\d+) rows of operations"],
    },
    {
        "slug": "shopline-api-domains",
        "clone": "asgard-freyr-skills",
        "of": "shopline-backoffice/references/api",
        "how": files_in,
        "what": "back-office API domains recorded",
        "says": [r"(\d+) API\s+domains recorded", r"(\d+)\s+API domains recorded"],
    },
]


def material() -> list:
    out = []
    for root in MATERIAL:
        for p in sorted(root.rglob("*")):
            if p.suffix in (".md", ".go"):
                out.append((p, p.read_text()))
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--dump", action="store_true", help="print what upstream counts, and stop")
    args = ap.parse_args()

    base = resolve("deployments")
    docs = material()
    bad = []
    counted = 0

    for row in COUNTS:
        clone = base / row["clone"]
        target = clone / row["of"]
        if not target.exists():
            bad.append(f"{row['slug']}: {row['clone']}/{row['of']} is not in the clone, "
                       f"so this count cannot be recomputed. `git -C {clone} fetch` first.")
            continue
        want = row["how"](target)
        counted += 1
        if args.dump:
            print(f"{row['slug']:<22} {want:>4}  {row['what']}")
            print(f"{'':<22}       {row['clone']}/{row['of']} at {commit(clone) or '?'}")
            continue

        found = 0
        for path, text in docs:
            for pattern in row["says"]:
                for m in re.finditer(pattern, text):
                    found += 1
                    if int(m.group(1)) != want:
                        rel = path.relative_to(ROOT)
                        line = text[:m.start()].count("\n") + 1
                        bad.append(f"{rel}:{line} says {m.group(1)} {row['what']}, "
                                   f"and {row['clone']} has {want}")
        if found == 0:
            bad.append(f"{row['slug']}: no claim in this material matches any of its patterns, "
                       f"so a count of {want} is going unchecked. Either the wording moved and "
                       f"the pattern has to move with it, or the claim is gone and so should "
                       f"this row be.")

    if args.dump:
        return 0

    print(f"{counted} count(s) recomputed off {base}")
    for b in bad:
        print(f"  {b}")
    print(f"{len(bad)} disagreement(s)")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
