#!/usr/bin/env python3
"""List every instruction in the embedded material, on one screen.

Three self-contradictions have reached a customer, and all three were found by
somebody walking into one. They share a cause that is structural rather than
careless: **the instructions are scattered, so there is no moment at which two
opposing ones are in front of the same pair of eyes.**

This does not detect contradictions. Verb-pair matching over prose produces
noise, and a checker that cries wolf teaches people to change what it can see
rather than what is wrong - which is a failure this repository has already had.

What it does is make a reviewable surface: every imperative, with its file, so
that reading them in one pass is possible at all. A few hundred lines is a
sitting; three bodies of material spread over ninety files is not.

Two things to look for while reading:

  1. **Pairs.** Two lines telling a reader opposite things about one subject.
     Both times this happened, the pages were `wiki/pages/operations.md` and
     `stage/prompts/11-requirements.md`, and the FDE followed whichever was in
     front of them.

  2. **Instructions with no reader.** A line that is right for one reader and
     wrong for another, with nothing saying which. `Get the name of whoever
     approves it` is correct for someone doing an integration and was copied
     onto a customer slide. The repair is to embed the reader or the destination
     in the sentence itself - adjacency was tried and it failed, because an
     imperative is what a scanner lifts and prose beside one reads as
     elaboration.

    python3 hack/imperatives.py            every imperative, grouped by file
    python3 hack/imperatives.py --ask      only those about asking a customer
    python3 hack/imperatives.py --unmarked only those with no reader or
                                           destination stated in the sentence
"""

import argparse
import pathlib
import re
import sys

ROOTS = [
    "internal/stage/prompts",
    "internal/wiki/pages",
    "internal/usecase/extracts",
    "internal/scaffold/templates/.agents/skills",
]

# An instruction is a bolded imperative, which is the form this material uses
# and - established the hard way - the span a reader copies.
BOLD = re.compile(r"\*\*(.+?)\*\*", re.S)
IMPERATIVE = re.compile(
    r"^(ask|do not|don't|never|always|say|write|read|check|get|take|use|put|send|"
    r"give|keep|make|treat|assume|start|stop|leave|prefer|avoid|confirm|count|"
    r"decide|name|record|file|run|open)\b",
    re.I,
)
# A sentence that already says who it is for, or where the answer goes.
MARKED = re.compile(
    r"\bfor the (tracking|follow-up|row)\b|\bin the (row|meeting|handover)\b|"
    r"\bthis line is for\b|\baloud\b|\bon a slide\b|\bin `docs/|\bnot for a slide\b",
    re.I,
)
ASKING = re.compile(r"\bask\b|\bcustomer\b|\bmeeting\b|\bthey\b", re.I)


def instructions(root: pathlib.Path):
    for path in sorted(root.rglob("*.md")) + sorted(root.rglob("*.tmpl")):
        text = path.read_text(encoding="utf-8", errors="replace")
        for m in BOLD.finditer(text):
            line = " ".join(m.group(1).split())
            if len(line) < 8 or not IMPERATIVE.match(line):
                continue
            yield path, line


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--ask", action="store_true", help="only instructions about asking a customer")
    ap.add_argument("--unmarked", action="store_true", help="only those with no reader or destination")
    args = ap.parse_args()

    repo = pathlib.Path(__file__).resolve().parent.parent
    total = shown = 0
    for rel in ROOTS:
        root = repo / rel
        if not root.is_dir():
            continue
        rows = []
        for path, line in instructions(root):
            total += 1
            if args.ask and not ASKING.search(line):
                continue
            if args.unmarked and MARKED.search(line):
                continue
            rows.append((path.relative_to(repo), line))
        if not rows:
            continue
        print(f"\n{'=' * 78}\n{rel}\n{'=' * 78}")
        current = None
        for path, line in rows:
            if path != current:
                current = path
                print(f"\n  {path}")
            print(f"    - {line[:150]}")
            shown += 1

    print(f"\n{shown} shown of {total} instructions.")
    if args.unmarked:
        print(
            "\nEach of these tells somebody to do something without saying who it is\n"
            "for or where the answer goes. That is not automatically wrong - most\n"
            "instructions have one obvious reader. It is the list to read when\n"
            "asking which ones do not."
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
