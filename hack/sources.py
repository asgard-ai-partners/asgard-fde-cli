#!/usr/bin/env python3
"""Where the upstream clones are, resolved from the environment.

**The URL is the source of truth and a local path is not**, which is the rule
`STRUCTURE.md` states and the reason nothing upstream is vendored in. But every
check in this directory needs a clone to read, and a path written into a script
is true on one machine and wrong on every other - which is the same reason
nothing upstream is vendored in.

So each source has an environment variable and a default. The default is one
person's layout and is documented as such; the variable is the contract.

    ASGARD_KUBE         the CRDs, the platform contract
    ASGARD_DOCS         the product documentation
    ASGARD_CORE         the processor definitions the CRDs are generated from
    ASGARD_DEPLOYMENTS  the directory holding the reference deployment clones

Nothing here clones or pulls. **`git pull` is the reader's act**: a script that
fetches turns "read at this commit" into "read at whatever was there when the
script ran", which is the one thing the provenance rule exists to prevent.

    from sources import resolve
    kube = resolve("kube")          # exits with a usable message if absent
    docs = resolve("docs", need=False)   # None if absent

    hack/sources.py              what each source resolves to, and what is stale
    hack/sources.py --extracts   how far each extract's source chart has moved
"""

import os
import pathlib
import re
import subprocess
import sys

# name -> (env var, default under $HOME, what it is)
SOURCES = {
    "kube": ("ASGARD_KUBE", "projects/asgard/asgard-kube",
             "the CRDs: https://github.com/asgard-ai-platform/asgard-kube"),
    "docs": ("ASGARD_DOCS", "projects/asgard-docs",
             "the product documentation: https://github.com/asgard-ai-platform/asgard-docs"),
    "core": ("ASGARD_CORE", "projects/asgard/asgard-core",
             "the processor definitions: https://github.com/asgard-ai-platform/asgard-core"),
    "deployments": ("ASGARD_DEPLOYMENTS", "projects/asgard",
                    "the directory holding the reference deployment clones"),
}


def dotenv() -> dict:
    """`.env` at the repository root, as a dict. Empty when there is none.

    **So the template is load-bearing.** `.env.example` documents these, and a
    template nobody loads is decoration: the environment still wins, this fills
    in what it does not set, and the default fills in the rest.
    """
    f = pathlib.Path(__file__).resolve().parent.parent / ".env"
    out = {}
    if not f.is_file():
        return out
    for line in f.read_text().split("\n"):
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        out[k.strip()] = v.strip().strip('"').strip("'")
    return out


def resolve(name: str, need: bool = True):
    """The clone for one source, or None. Exits with a message when need.

    Environment first, then `.env`, then the default.
    """
    env, default, what = SOURCES[name]
    raw = os.environ.get(env) or dotenv().get(env) or str(pathlib.Path.home() / default)
    p = pathlib.Path(raw).expanduser()
    if p.is_dir():
        return p
    if not need:
        return None
    sys.exit(f"{env} is not a directory: {p}\n"
             f"  {what}\n"
             f"  Clone it wherever you like, then set {env} - in your shell, or in\n"
             f"  a `.env` at the repository root; `.env.example` is the template.\n"
             f"  Nothing here pulls for you.")


def commit(path: pathlib.Path) -> str:
    """The short commit a clone is on, or "" when the path is not a repository.

    `ASGARD_DEPLOYMENTS` is normally the parent of several clones rather than a
    clone, so an empty answer here is the ordinary case and not a failure.
    """
    r = subprocess.run(["git", "-C", str(path), "rev-parse", "--short", "HEAD"],
                       capture_output=True, text=True)
    return r.stdout.strip() if r.returncode == 0 else ""


def behind(path: pathlib.Path) -> str:
    """How far a clone is behind its own remote, as a phrase or "" when current.

    **Read, never fetched.** This reports the clone as it stands; if the answer
    matters, pull first. A check that fetched would be answering a different
    question each time it ran.
    """
    head = commit(path)
    for ref in ("origin/HEAD", "origin/main", "origin/master"):
        r = subprocess.run(["git", "-C", str(path), "rev-parse", "--short", ref],
                           capture_output=True, text=True)
        if r.returncode == 0:
            up = r.stdout.strip()
            if up == head:
                return ""
            n = subprocess.run(["git", "-C", str(path), "rev-list", "--count", f"HEAD..{ref}"],
                               capture_output=True, text=True).stdout.strip()
            return f"{n} commit(s) behind {up}"
    return ""


# The readings TASK.md records, and the clone each was held against. **This is
# the only checkable thing about a reading**: not that it happened - nobody but
# the reader can say that - but whether the thing it was held against has moved
# since. A `never` row has nothing to go stale.
READINGS = {
    "extracts-vs-charts": "deployments",
    "wiki-vs-docs": "docs",
    "processors-vs-palette": "docs",
}


def stale() -> list:
    """Which recorded readings are now behind their source."""
    task = (pathlib.Path(__file__).resolve().parent.parent / "TASK.md").read_text()
    out = []
    for key, src in READINGS.items():
        row = next((l for l in task.split("\n") if "`" + key + "`" in l and l.startswith("|")), None)
        if row is None:
            out.append(f"{key}: no row in TASK.md's pass, so nothing records what it was read against")
            continue
        if "**never" in row:
            continue
        p = resolve(src, need=False)
        if p is None:
            out.append(f"{key}: ${SOURCES[src][0]} is not set, so this cannot be answered")
            continue
        if src == "deployments":
            # **Against the commit the reading names, not against the remote.**
            # This used to report a clone that was behind its own origin as a
            # reading gone stale, which is a different fact about a different
            # thing: the clone had not moved at all, its remote had, and the
            # extracts describe the clone. One deployment was reported as six
            # commits of drift while sitting exactly where its extracts were
            # read.
            doc = (pathlib.Path(__file__).resolve().parent.parent
                   / "source/SOURCES.md").read_text()
            moved = []
            for name, ref in held_against(doc).items():
                if not (p / name / ".git").exists():
                    out.append(f"{key}: no clone of {name}, so its reading cannot be answered")
                    continue
                n = since(p / name, ref)
                if n < 0:
                    out.append(f"{key}: {name} has no commit {ref}, the one its claims "
                               f"were read against")
                elif n:
                    moved.append((name, n))
            if moved:
                out.append(f"{key}: {len(moved)} clone(s) have moved since the reading: "
                           + ", ".join(f"{n} by {c}" for n, c in moved[:4])
                           + ("..." if len(moved) > 4 else ""))
            continue
        if b := behind_of(p):
            out.append(f"{key}: read against {SOURCES[src][0]}, which is now {b}")
    return out


def extracts() -> list:
    """Each extract's source chart, and how far the clone has moved since.

    **`source/SOURCES.md` used to carry these numbers as prose, and both of the
    two that were not zero had rotted within nine days.** One named a commit the
    clone had already moved past; the other reported six commits of drift on a
    clone sitting exactly where the extract was read. Written-down distances
    between two moving things are the one shape of claim that cannot hold, so
    that column points here instead.

    Read, never fetched: this is the clone as it stands, which is what a reader
    would open. `git -C <path> pull` first if the answer matters.
    """
    doc = (pathlib.Path(__file__).resolve().parent.parent / "source/SOURCES.md").read_text()
    base = resolve("deployments", need=False)
    out = []
    for name, at in written_from(doc).items():
        if base is None or not (base / name / ".git").exists():
            out.append((name, at, None, "no clone"))
            continue
        p = base / name
        n = subprocess.run(["git", "-C", str(p), "rev-list", "--count", f"{at}..HEAD"],
                           capture_output=True, text=True)
        head = subprocess.run(["git", "-C", str(p), "log", "-1", "--format=%h %ad", "--date=short"],
                              capture_output=True, text=True).stdout.strip()
        if n.returncode != 0:
            # The recorded commit is not in this clone. Either it was never
            # fetched or the branch was rewritten, and both mean the extract's
            # source cannot be opened - which is worse than being behind.
            out.append((name, at, None, "not in the clone"))
            continue
        out.append((name, at, int(n.stdout.strip()), head))
    return out


def written_from(doc: str) -> dict:
    """Each deployment and the commit its extracts were written from."""
    return dict(re.findall(r"^\|\s*([a-z0-9-]+)\s*\|\s*`([0-9a-f]{7,})`", doc, re.M))


def held_against(doc: str) -> dict:
    """Each deployment and the commit its claims were last read against.

    **The second column of that table, and the one a staleness report wants.**
    An extract describes the version it was written from; whether it is still
    true is a question about the version somebody last checked it against, and
    those are different commits for four of the eight.
    """
    return {n: c for n, _, c in re.findall(
        r"^\|\s*([a-z0-9-]+)\s*\|\s*`([0-9a-f]{7,})`[^|]*\|\s*`([0-9a-f]{7,})`", doc, re.M)}


def since(path: pathlib.Path, ref: str) -> int:
    """Commits between one ref and a clone's HEAD, or -1 when it has no such ref."""
    got = subprocess.run(["git", "-C", str(path), "rev-list", "--count", f"{ref}..HEAD"],
                         capture_output=True, text=True)
    if got.returncode != 0:
        return -1
    return int(got.stdout.strip() or 0)


def reference_deployments() -> list:
    """The eight reference deployments, read off `source/SOURCES.md`.

    **Not a list in this file.** That document is the only one allowed to name
    a customer's repository, and a second copy here is the drift this whole
    directory exists to catch. The two contract repositories are excluded by
    name because they are declared as sources in their own right.
    """
    text = (pathlib.Path(__file__).resolve().parent.parent / "source/SOURCES.md").read_text()
    names = set(re.findall(r"\[([a-z0-9-]+)\]\(https://github\.com/asgard-ai-platform/[a-z0-9-]+\)", text))
    return sorted(names - {"asgard-kube", "asgard-docs", "asgard-core"})


def behind_of(path: pathlib.Path) -> str:
    """behind() for one clone, empty when it is current."""
    return behind(path)


def main() -> int:
    """Print what each source resolves to, which is the thing to run first."""
    if "--extracts" in sys.argv:
        rows = extracts()
        width = max(len(n) for n, _, _, _ in rows)
        print("Each extract's source chart, as the clone stands. Nothing here pulls.\n")
        for name, at, n, head in rows:
            if n is None:
                print(f"  {name:<{width}}  read at {at}  -- {head}")
            elif n == 0:
                print(f"  {name:<{width}}  read at {at}  unmoved")
            else:
                print(f"  {name:<{width}}  read at {at}  {n} commit(s) since, now at {head}")
        moved = [n for _, _, c, n in rows if c]
        print(f"\n{len(moved)} of {len(rows)} have moved since the extracts were written from them.")
        print("An extract describes one version of one chart; that is the size of the re-read.")
        return 0

    width = max(len(e) for e, _, _ in SOURCES.values())
    for name, (env, _, what) in SOURCES.items():
        p = resolve(name, need=False)
        if p is None:
            print(f"{env:<{width}}  not found   {what}")
            continue
        at = commit(p)
        if not at:
            n = len([d for d in p.iterdir() if (d / ".git").exists()])
            print(f"{env:<{width}}  {'-':<10} {n} clone(s) under it        {p}")
            continue
        b = behind(p)
        print(f"{env:<{width}}  {at:<10} {('(' + b + ')') if b else '(current)':<34} {p}")
    print("\nNothing here pulls. `git -C <path> pull` before a reading that matters.")

    # **Whether a recorded reading has gone behind.** A reading cannot be
    # verified; a reading being stale can.
    rows = stale()
    if rows:
        print("\nreadings TASK.md records that are now behind their source:")
        for r in rows:
            print(f"  {r}")
        print("\nThat is not a failure - it is the size of what re-reading would cover.")
    else:
        print("\nEvery reading TASK.md records is against a source that has not moved.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
