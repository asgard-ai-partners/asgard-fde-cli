#!/usr/bin/env python3
"""How much of a production chart `asgard-cli add` never writes.

**This is the number behind "the chart half is the least finished of the four."**
Goal.md claims it, TASK.md states it, and it decides whether an FDE should treat
what `add` emits as a chart or as a starting point - so it is the one claim in
this repository where being roughly right is not good enough, and it went a week
as a figure nobody could re-derive.

The method, because the number means nothing without it:

    a spec key is a dotted path under `spec`, list indices collapsed
    the production side is every reference chart rendered with its own values
    the `add` side is a scaffolded repository with one of every kind added,
    rendered the same way

Collapsing list indices is what makes the two comparable: `processors.0.configs`
and `processors.7.configs` are the same key, and counting them apart would say
the gap closes as a chart grows.

    hack/spec-key-gap.py             recompute, and check TASK.md's claim
    hack/spec-key-gap.py --missing   list the keys production uses and we never write

Needs helm, yq, $ASGARD_DEPLOYMENTS and a built binary ($ASGARD_CLI).
"""

import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from sources import reference_deployments, resolve  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent

# The kinds `add` writes, and the flags each needs to write anything. A kind
# that grows a required flag and is not listed here would silently shrink the
# `add` side, which would make the gap look wider than it is.
KINDS = [
    ("dataconnector", []),
    ("semanticlayer", ["--connector", "dc-db"]),
    ("agent", []),
    ("httptool", ["--toolset", "ts-http"]),
    ("querytool", ["--connector", "dc-db", "--toolset", "ts-query"]),
    ("skillset", ["--repo", "https://github.com/example/skills"]),
    ("trigger", []),
    ("knowledgedrive", ["--connector", "dc-db"]),
    ("plugin", ["--connector", "ss-skills"]),
    ("flowagent", []),
]

# Values the platform injects per run, which helm cannot know. `asgard-cli
# render` supplies them; this renders with helm directly so that the two sides
# are rendered the same way, so they are supplied here.
INJECTED = """asgard:
  projectEnvironmentId: probe-env
  projectId: probe-project
  appSecretName: probe-secret
  namespace: probe-ns
"""

CLAIM = re.compile(r"(\d+) spec keys and `add` never mentions (\d+) of them")


def spec_keys(text: str) -> set:
    """Every dotted key path under `spec`, with list indices collapsed."""
    out = set()

    def walk(node, path):
        if isinstance(node, dict):
            for k, v in node.items():
                p = f"{path}.{k}" if path else k
                out.add(p)
                walk(v, p)
        elif isinstance(node, list):
            for e in node:
                walk(e, path)

    conv = subprocess.run(["yq", "-o=json", "-I=0", "."], input=text,
                          capture_output=True, text=True)
    for line in conv.stdout.split("\n"):
        if not line.strip():
            continue
        try:
            doc = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(doc, dict) and isinstance(doc.get("spec"), (dict, list)):
            walk(doc["spec"], "")
    return out


def render(chart: pathlib.Path, values: list) -> str:
    cmd = ["helm", "template", "probe", str(chart)]
    for v in values:
        cmd += ["-f", str(v)]
    got = subprocess.run(cmd, capture_output=True, text=True)
    return got.stdout


def production(base: pathlib.Path) -> dict:
    """Each reference chart's spec keys, by the chart's own directory name.

    **Only the deployments `source/SOURCES.md` declares.** `$ASGARD_DEPLOYMENTS`
    is somebody's projects directory: it also holds scratch repositories this
    tool scaffolded, and those pass by construction - one of them showed 0 keys
    not written, which would make the gap look like it had closed.
    """
    out = {}
    known = set(reference_deployments())
    for chart in sorted(base.glob("*/**/chart/app")):
        if "/.git/" in str(chart) or "/charts/" in str(chart):
            continue
        if chart.relative_to(base).parts[0] not in known:
            continue
        holder = chart.parent
        values = [holder / n for n in ("values-prod.yaml", "values-dev.yaml")
                  if (holder / n).is_file()]
        if not values:
            continue
        text = render(chart, values[:1])
        if not text.strip():
            continue
        label = str(holder.relative_to(base))
        out[label] = spec_keys(text)
    return out


def ours(cli: pathlib.Path) -> set:
    """What `add` writes, rendered: one project with one of every kind in it."""
    with tempfile.TemporaryDirectory() as tmp:
        repo = pathlib.Path(tmp) / "probe"
        repo.mkdir()
        env = dict(os.environ, HOME=str(repo))
        subprocess.run(["git", "init", "-q", "."], cwd=repo, check=True)
        for args in ([str(cli), "init"], [str(cli), "project", "add", "site"]):
            got = subprocess.run(args, cwd=repo, capture_output=True, text=True, env=env)
            if got.returncode != 0:
                sys.exit(f"{' '.join(args)} failed in a scratch repository:\n{got.stderr}")
        for kind, flags in KINDS:
            got = subprocess.run([str(cli), "add", kind, f"probe-{kind}", "--project", "site"] + flags,
                                 cwd=repo, capture_output=True, text=True, env=env)
            if got.returncode != 0:
                sys.exit(f"`add {kind}` failed, so the `add` side would be short a kind:\n"
                         f"{got.stderr.strip()[:400]}\n"
                         f"If it grew a required flag, add it to KINDS in this file.")
        injected = repo / "injected.yaml"
        injected.write_text(INJECTED)
        chart = repo / "projects/site/chart/app"
        text = render(chart, [injected])
        if not text.strip():
            sys.exit("the scaffolded chart rendered empty, so there is nothing to compare")
        return spec_keys(text)


def main() -> int:
    for tool in ("helm", "yq"):
        if subprocess.run(["which", tool], capture_output=True).returncode != 0:
            sys.exit(f"{tool} is not on PATH, and both sides have to be rendered")
    base = resolve("deployments")
    cli = pathlib.Path(os.environ.get("ASGARD_CLI", ROOT / ".out/asgard-cli")).expanduser()
    if not cli.is_file():
        sys.exit(f"no binary at {cli}\n"
                 f"  go build -o .out/asgard-cli ./cmd/asgard-cli\n"
                 f"  Set $ASGARD_CLI to use another one.")

    prod = production(base)
    if not prod:
        sys.exit(f"no reference chart under {base} rendered, so there is nothing to measure")
    mine = ours(cli)
    every = set().union(*prod.values())
    widest = max(prod, key=lambda k: len(prod[k]))

    if "--missing" in sys.argv:
        for k in sorted(every - mine):
            print(f"  {k}")
        print(f"\n{len(every - mine)} key(s) production uses and `add` never writes.")
        return 0

    for label in sorted(prod, key=lambda k: -len(prod[k])):
        print(f"  {label:<44}{len(prod[label]):>4} spec key(s), {len(prod[label] - mine):>4} not written")
    print(f"\n  {'every reference chart together':<44}{len(every):>4} spec key(s), "
          f"{len(every - mine):>4} not written")
    print(f"  {'what `add` writes':<44}{len(mine):>4} spec key(s)")

    # The claim is about one chart - the widest, which is the honest one to
    # quote - and TASK.md states it as a pair.
    task = (ROOT / "TASK.md").read_text()
    m = CLAIM.search(task)
    if not m:
        print("\nTASK.md states no `<n> spec keys and `add` never mentions <n> of them` claim,\n"
              "so this measures and checks nothing. Either restore the claim or delete this.")
        return 1
    want = (len(prod[widest]), len(prod[widest] - mine))
    said = (int(m.group(1)), int(m.group(2)))
    if said != want:
        line = task[:m.start()].count("\n") + 1
        print(f"\nTASK.md:{line} says {said[0]} spec keys and {said[1]} never mentioned;\n"
              f"the widest reference chart has {want[0]} and {want[1]}.")
        return 1
    print(f"\nTASK.md's claim matches the widest chart: {want[0]} keys, {want[1]} never written.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
