#!/usr/bin/env python3
"""Hold the four points of Goal.md against what the binary actually does.

**Everything else here checks the material; nothing checked the capability.**
A change that made `asgard-cli init` require a session would pass every audit,
every script and every Go test in this repository - the corpus would still be
consistent, every pointer would still resolve, and Goal's first point would be
gone.

So this runs the tool the way Goal.md says it has to work:

    hack/check-goal.py [binary]        default: .out/asgard-cli

It scaffolds a repository in a temporary directory with **no network, no
account and no git repository**, and then asks the four questions. It is an
end-to-end check rather than a consistency check, which is why it is the only
thing here that writes files.

`$ASGARD_*` clones are not needed. Nothing is fetched; a step that would need
the network is the failure, not a skip.
"""

import os
import pathlib
import re
import shutil
import subprocess
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parent.parent
CORPUS = ".agents/skills/asgard-platform"

# Goal's first point names four bodies of knowledge. A landed tree missing one
# is that point half-delivered, and `--links` cannot see it: a directory that
# is not written has no pointers to go dead.
KINDS = ["wiki", "usecase", "needs", "brief", "guide"]

# Goal's second point: the deck is the only thing a customer reads, so it has
# rules of its own. These are the three it names.
DECK_RULES = ["screenshot", "never appear", "worth"]

# Terms an agent would grep for, in the customer's words or the platform's.
# Retrieval is the whole engineering problem in Goal's closing section, and
# "the material landed" is not the same as "a grep finds it".
GREPS = ["allowlist", "botProviderClass", "immutable", "read-only"]


def run(binary, args, cwd, env=None):
    e = dict(os.environ)
    # **No network.** A proxy that resolves nowhere is the cheapest way to
    # make a fetch fail rather than succeed slowly, and Goal's first point is
    # that none is needed.
    e.update({"HTTP_PROXY": "http://127.0.0.1:1", "HTTPS_PROXY": "http://127.0.0.1:1",
              "ALL_PROXY": "http://127.0.0.1:1", "NO_PROXY": "",
              "ASGARD_PROFILE": "", "HOME": cwd})
    if env:
        e.update(env)
    return subprocess.run([binary] + args, cwd=cwd, env=e,
                          capture_output=True, text=True, timeout=180)


def main() -> int:
    binary = sys.argv[1] if len(sys.argv) > 1 else str(ROOT / ".out/asgard-cli")
    if not pathlib.Path(binary).exists():
        sys.exit(f"no binary at {binary}\n  go build -o .out/asgard-cli ./cmd/asgard-cli")
    binary = str(pathlib.Path(binary).resolve())

    bad = []
    tmp = tempfile.mkdtemp(prefix="asgard-goal-")
    try:
        # ── Goal 1: the knowledge, offline, with no repository ──────────
        #
        # No `git init`. Goal.md: a tool that needs a directory first will not
        # get asked, and the question is asked in a meeting.
        out = run(binary, ["init"], tmp)
        if out.returncode != 0:
            bad.append(f"1: `init` failed in an empty directory with no network:\n{out.stderr.strip()[:400]}")
        tree = pathlib.Path(tmp)
        for kind in KINDS:
            if not (tree / CORPUS / kind).is_dir():
                bad.append(f"1: `{CORPUS}/{kind}/` did not land, so one of the four bodies is missing")
        landed = list((tree / CORPUS).rglob("*.md")) if (tree / CORPUS).is_dir() else []
        if len(landed) < 60:
            bad.append(f"1: only {len(landed)} corpus documents landed; the corpus is not a handful of files")

        # **TASK.md's "N documents and M words" is the claim this material makes
        # about its own value**, and a floor of 60 does not hold it. The landed
        # tree is the only place the number is true of anything - the input tree
        # has neither the rendered `needs` shapes nor the briefings as files -
        # so it is counted here rather than trusted, whitespace-separated and
        # rounded to the nearest thousand.
        kinded = [p for k in KINDS for p in (tree / CORPUS / k).glob("*.md")]
        claim = re.search(r"\*\*(\d+) documents and ([\d,]+)\s*\n?words\*\*",
                          (ROOT / "TASK.md").read_text())
        if claim is None:
            bad.append("1: TASK.md states no `**<n> documents and <n> words**` claim, so the "
                       "one number this material gives for its own size is unchecked")
        else:
            docs, words = int(claim.group(1)), int(claim.group(2).replace(",", ""))
            got_words = sum(len(p.read_text(errors="replace").split()) for p in kinded)
            if len(kinded) != docs:
                bad.append(f"1: TASK.md says {docs} documents and {len(kinded)} landed "
                           f"across {', '.join(KINDS)}")
            if round(got_words, -3) != words:
                bad.append(f"1: TASK.md says {words:,} words and the landed corpus has "
                           f"{got_words:,}, which rounds to {round(got_words, -3):,}")

        # Retrieval: grep is the way in, so a grep has to find things.
        body = "\n".join(p.read_text(errors="replace") for p in landed)
        for term in GREPS:
            if not re.search(term, body, re.I):
                bad.append(f"1: `grep {term}` finds nothing in the landed corpus, and grep is the only way in")
        if not (tree / CORPUS / "aliases.md").is_file():
            bad.append("1: `aliases.md` did not land, so a question in the customer's words has nothing to translate it")
        # The two-sense warning only reaches a reader if the glossary carries
        # the word - that is the whole mechanism, not the instruction.
        gloss = tree / CORPUS / "wiki/glossary.md"
        if not gloss.is_file():
            bad.append("1: `wiki/glossary.md` did not land")
        elif "payment" not in gloss.read_text(errors="replace"):
            bad.append("1: the glossary does not contain `payment`, so a grep for it will not surface the two senses")

        # ── Goal 2: what to get from the customer, and the deck's rules ──
        for shape in ["semantic-layer", "chat-channel", "write-path"]:
            if not (tree / CORPUS / "needs" / f"{shape}.md").is_file():
                bad.append(f"2: `needs/{shape}.md` did not land, and that file is Goal's second point")
        skills = "\n".join(p.read_text(errors="replace")
                           for p in (tree / ".agents/skills").rglob("SKILL.md")) \
            if (tree / ".agents/skills").is_dir() else ""
        for rule in DECK_RULES:
            if rule not in skills:
                bad.append(f"2: no landed skill mentions {rule!r}, and the deck's own rules are Goal's second point")

        # ── Goal 3: the charts ───────────────────────────────────────────
        subprocess.run(["git", "init", "-q", "."], cwd=tmp, check=True)
        for args in (["project", "add", "app"],
                     ["add", "dataconnector", "erp", "--project", "app"]):
            out = run(binary, args, tmp)
            if out.returncode != 0:
                bad.append(f"3: `{' '.join(args)}` failed:\n{out.stderr.strip()[:300]}")
        chart = tree / "projects/app/chart/app"
        if not (chart / "Chart.yaml").is_file():
            bad.append("3: no chart at projects/app/chart/app, which Goal's third point names")
        if not list(chart.rglob("dc-erp.yaml")):
            bad.append("3: `add dataconnector` wrote no CR")
        out = run(binary, ["check"], tmp)
        if out.returncode != 0:
            bad.append(f"3: `check` failed on a freshly scaffolded repository:\n{out.stdout.strip()[-300:]}")

        # ── Goal 4: the way back in, out of the tool's own output ────────
        #
        # "要從工具自己的輸出講出來" - so the URL has to be in what the tool
        # prints, not only in a document somebody has to think to open.
        out = run(binary, ["issue-report"], tmp)
        if "github.com/asgard-ai-partners/asgard-fde-cli" not in out.stdout:
            bad.append("4: `issue-report` does not print the repository URL")
        out = run(binary, ["issue-report", "--new"], tmp)
        if out.returncode != 0 or len(out.stdout) < 400:
            bad.append("4: `issue-report --new` did not write a report body")
        skill = tree / CORPUS / "SKILL.md"
        if not skill.is_file() or "issue-report" not in skill.read_text(errors="replace"):
            bad.append("4: the landed SKILL.md does not name `issue-report`, so an agent that greps and finds nothing is not told where to file it")
    finally:
        shutil.rmtree(tmp, ignore_errors=True)

    for b in bad:
        print(f"goal  {b}")
    print(f"\nGoal.md's four points, held against the binary: {len(bad)} unmet.")
    if bad:
        print("\n**This is a capability check, not a consistency check.** Every other\n"
              "check here can pass while one of these fails, which is why it exists.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
