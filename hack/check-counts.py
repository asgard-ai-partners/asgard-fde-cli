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
import subprocess
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from sources import commit, resolve  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
MATERIAL = [ROOT / "internal/corpus", ROOT / "internal/needs", ROOT / "internal/stage",
            ROOT / "source"]

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


def kind_count(kind: str):
    """How many CRs of one kind a clone holds at one commit.

    **At a commit rather than in the working tree**, because the claim is dated:
    a deployment that gained a Plugin yesterday does not make yesterday's count
    wrong. `check-coverage.py` measures the docs tree the same way and for the
    same reason - a count with no commit beside it cannot be checked twice.
    """
    def count(clone: pathlib.Path, ref: str) -> int:
        # `-c` without `-h`: the count needs its file name to be parseable, and
        # `-h` prints a bare number per file that cannot be told from a path.
        out = subprocess.run(["git", "-C", str(clone), "grep", "-c",
                              f"^kind: {kind}$", ref],
                             capture_output=True, text=True)
        if out.returncode not in (0, 1):
            return -1
        total = 0
        for line in out.stdout.split("\n"):
            if ":" in line:
                total += int(line.rsplit(":", 1)[1])
        return total
    return count


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


# Counts of a deployment AT A NAMED COMMIT. Each claim in the material says
# which commit it was counted at, so the answer never changes - which is what
# separates a dated count from the ones above, recomputed off whatever the clone
# now holds.
PINNED = [
    {
        "slug": "auto-post-plugins",
        "clone": "asgard-auto-post-kube",
        "ref": "edb0ad0",
        "how": kind_count("Plugin"),
        "what": "Plugin CRs",
        "says": [r"(\d+) Plugins and \d+ SkillSets", r"carrying (\d+) Plugins"],
    },
    {
        "slug": "auto-post-skillsets",
        "clone": "asgard-auto-post-kube",
        "ref": "edb0ad0",
        "how": kind_count("SkillSet"),
        "what": "SkillSet CRs",
        "says": [r"\d+ Plugins and (\d+) SkillSets"],
    },
    {
        "slug": "auto-post-plugins-as-read",
        "clone": "asgard-auto-post-kube",
        "ref": "d11b802",
        "how": kind_count("Plugin"),
        "what": "Plugin CRs at the commit its extracts were read at",
        "says": [r"(\d+) Plugin CRs"],
    },
]


def sources_table(base: pathlib.Path) -> list:
    """`source/SOURCES.md`'s CR-file count per deployment, at its read commit.

    Two tables in that file have to be read together: the first gives each
    deployment's read-at commit, the second its CR-file count under a short name.
    **Both columns are pinned to that commit**, so the answer is fixed - which is
    the difference between this and the "since then" column that file used to
    carry and `hack/sources.py --extracts` now computes.

    Returns a list of findings, empty when every row agrees.
    """
    doc = (ROOT / "source/SOURCES.md").read_text()
    read_at = dict(re.findall(r"^\|\s*([a-z0-9-]+)\s*\|\s*`([0-9a-f]{7,})`", doc, re.M))
    out = []
    for short, said in re.findall(r"^\|\s*([a-z0-9-]+)\s*\|[^|]*\|\s*(\d+)\s*\|", doc, re.M):
        # The short name in the second table is a prefix of the clone name in
        # the first - "freyr" against "asgard-freyr-kube" - and the two are
        # written for their own readers rather than to be joined, so match on
        # the clone containing it and require exactly one.
        hits = [n for n in read_at if short in n and not n.endswith("-skills")]
        if len(hits) != 1:
            out.append(f"source/SOURCES.md: {short!r} in the CR-file table matches "
                       f"{len(hits)} deployment(s) in the read-at table, so its count "
                       f"cannot be held against a commit")
            continue
        clone, ref = base / hits[0], read_at[hits[0]]
        got = subprocess.run(["git", "-C", str(clone), "ls-tree", "-r", "--name-only", ref],
                             capture_output=True, text=True)
        if got.returncode != 0:
            out.append(f"source/SOURCES.md: {hits[0]} has no commit {ref}, so its "
                       f"{said} CR files cannot be recounted")
            continue
        n = len([f for f in got.stdout.split("\n") if f.endswith(".yaml") and "/templates/" in f])
        if n != int(said):
            out.append(f"source/SOURCES.md: {short} says {said} CR files and {hits[0]} "
                       f"has {n} under templates/ at {ref}")
    return out


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

    for row in PINNED:
        clone = base / row["clone"]
        want = row["how"](clone, row["ref"])
        if want < 0:
            bad.append(f"{row['slug']}: {row['clone']} has no commit {row['ref']}. "
                       f"`git -C {clone} fetch` first.")
            continue
        counted += 1
        if args.dump:
            print(f"{row['slug']:<26} {want:>4}  {row['what']} at {row['clone']} {row['ref']}")
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
                                   f"and {row['clone']} has {want} at {row['ref']}")
        if found == 0:
            bad.append(f"{row['slug']}: no claim matches any of its patterns, so a count of "
                       f"{want} at {row['ref']} is going unchecked")

    if not args.dump:
        bad.extend(sources_table(base))
        counted += 1

    if args.dump:
        return 0

    print(f"{counted} count(s) recomputed off {base}")
    for b in bad:
        print(f"  {b}")
    print(f"{len(bad)} disagreement(s)")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
