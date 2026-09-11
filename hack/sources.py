#!/usr/bin/env python3
"""Where the upstream clones are, resolved from the environment.

**The URL is the source of truth and a local path is not**, which is the rule
`STRUCTURE.md` states and the reason nothing upstream is vendored in. But every
check in this directory needs a clone to read, and the paths were written into
the scripts and into `README.md` - true on one machine, wrong on every other,
and the reason `hack/check-tables.py` went eight upstream commits without being
run.

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
"""

import os
import pathlib
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


def main() -> int:
    """Print what each source resolves to, which is the thing to run first."""
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
    return 0


if __name__ == "__main__":
    sys.exit(main())
