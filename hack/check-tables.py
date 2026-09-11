#!/usr/bin/env python3
"""Hold internal/gate's pinned tables against the CRDs the apiserver enforces.

The tables in `internal/gate` were extracted from asgard-kube's Go types, and
that is not the contract - the generated CRDs are. The difference is not
academic: `status` carries three values in the Asgard types and six in the CRD,
because Kubernetes' own condition schema uses the same field name, and the
entry sat in the enum table for a day before this check existed.

**Run it whenever asgard-kube moves**, and move the read markers in
`internal/gate` in the same change. It takes `$ASGARD_KUBE` with no argument,
which `hack/sources.py` resolves:

    hack/check-tables.py

**One command on purpose.** A pinned table goes stale in both directions - the
platform can delete a constraint as readily as add one - and a check that
needs three lines of shell to set up is a check that does not get run.

Exits 1 on a disagreement. A field this reports as absent from the CRD is not
necessarily a bug - `baseAgentName` lives inside a JSON string rather than in
the schema - but it is always something to explain rather than leave.
"""
import json
import pathlib
import re
import subprocess
import sys

CONSTRAINT_KEYS = ("pattern", "minLength", "maxLength", "minimum", "maximum", "minItems", "maxItems")


def crd_properties(crdjson):
    """Every property in every CRD, with its enum and its constraints."""
    enums, cons = {}, {}

    def walk(node):
        if isinstance(node, dict):
            props = node.get("properties")
            if isinstance(props, dict):
                for name, schema in props.items():
                    if isinstance(schema, dict):
                        if "enum" in schema:
                            enums.setdefault(name, set()).update(schema["enum"])
                        if any(k in schema for k in CONSTRAINT_KEYS):
                            cons.setdefault(name, set()).add(
                                json.dumps({k: schema[k] for k in CONSTRAINT_KEYS if k in schema}, sort_keys=True))
                    walk(schema)
            for key, value in node.items():
                if key != "properties" and isinstance(value, (dict, list)):
                    walk(value)
        elif isinstance(node, list):
            for entry in node:
                walk(entry)

    for path in sorted(pathlib.Path(crdjson).glob("*.json")):
        walk(json.load(open(path)))
    return enums, cons


def as_json(src: pathlib.Path) -> pathlib.Path:
    """Take a directory of CRDs and return one of the same as JSON.

    **Here rather than in `README.md` as three lines of shell**, because a
    ritual that is not one command is a ritual that does not happen.
    """
    if next(src.glob("*.json"), None):
        return src
    out = pathlib.Path(".out/crdjson")
    out.mkdir(parents=True, exist_ok=True)
    for f in sorted(src.glob("*.yaml")):
        target = out / (f.stem + ".json")
        if subprocess.run(["yq", "-o=json", str(f)],
                          stdout=target.open("w")).returncode != 0:
            sys.exit(f"yq failed on {f}")
    if not next(out.glob("*.json"), None):
        sys.exit(f"no CRDs found in {src}")
    return out


# Every place this repository states how many CEL rules there are. **Two numbers
# that are easy to write for each other**: 79 is the `XValidation` markers in
# asgard-kube's Go types, 231 is what the generator emits from them, because one
# marker on a struct several kinds embed lands in every CRD that embeds it. This
# material had the marker count written down as the CRDs' own for a week, which
# is the exact confusion this whole file exists to catch - the Go types are the
# source and the generated CRDs are the contract.
CEL_CLAIMS = [
    (r"(\d+) CEL rules written and (\d+) enforced", ("markers", "rules")),
    (r"(\d+) of the CRDs' (\d+) enforced CEL rules", ("oldself_rules", "rules")),
    (r"(\d+) of the CRDs' (\d+) enforced CEL rules are exactly", ("oldself_rules", "rules")),
    (r"(\d+) of the enforced rules are exactly `self == oldSelf`", ("oldself_rules",)),
    (r"(\d+) of the (\d+) `XValidation` markers", ("oldself_markers", "markers")),
    (r"(\d+) rule instances, (\d+) of them distinct", ("rules", "distinct")),
]


def cel_counts(crd: pathlib.Path, kube: pathlib.Path) -> dict:
    """How many CEL rules there are, on both sides of the generator."""
    rules = []
    for f in sorted(crd.glob("*.yaml")):
        rules += [r.strip() for r in re.findall(r'^\s*-?\s*rule:\s*(.*)$', f.read_text(), re.M)]
    types = (kube / "pkg/apis/asgard/v1alpha1/types.go").read_text()
    return {
        "rules": len(rules),
        "distinct": len({r for r in rules}),
        "oldself_rules": len([r for r in rules if r.strip('"\'') == "self == oldSelf"]),
        "markers": len(re.findall(r"XValidation:rule=", types)),
        "oldself_markers": len(re.findall(r'XValidation:rule=`?"?self == oldSelf', types)),
    }


def check_cel(crd: pathlib.Path, kube: pathlib.Path) -> list:
    counts = cel_counts(crd, kube)
    root = pathlib.Path(__file__).resolve().parent.parent
    bodies = []
    for pat in ("internal/corpus/**/*.md", "internal/gate/*.go", "TASK.md", "AGENTS.md"):
        for f in root.glob(pat):
            bodies.append((f.relative_to(root), f.read_text(errors="replace")))
    out, seen = [], 0
    for pattern, names in CEL_CLAIMS:
        for path, text in bodies:
            for m in re.finditer(pattern, text):
                seen += 1
                for i, name in enumerate(names, start=1):
                    if int(m.group(i)) != counts[name]:
                        line = text[:m.start()].count("\n") + 1
                        out.append(f"{path}:{line} says {m.group(i)} for {name}, "
                                   f"and asgard-kube has {counts[name]}")
    if seen == 0:
        out.append("no CEL-rule claim matches any pattern in CEL_CLAIMS, so "
                   f"{counts['rules']} enforced rules and {counts['markers']} markers "
                   "are going unchecked")
    return out


def main():
    if len(sys.argv) > 2:
        sys.exit("usage: check-tables.py [asgard-kube/crd]   (default: $ASGARD_KUBE/crd)")
    if len(sys.argv) == 2:
        crd = pathlib.Path(sys.argv[1])
    else:
        sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
        from sources import resolve
        crd = resolve("kube") / "crd"
    enums, cons = crd_properties(as_json(crd))
    problems = []

    src = pathlib.Path("internal/gate/enums.go").read_text()
    mine = {m.group(1): {v.strip().strip('"') for v in m.group(2).split(",") if v.strip()}
            for m in re.finditer(r'^\t"(\w+)":\s*\{([^}]*)\}', src, re.M)}
    for name, values in sorted(mine.items()):
        if name not in enums:
            problems.append(f"enum {name}: not an enum in any CRD")
        elif values != enums[name]:
            problems.append(f"enum {name}: table {sorted(values)} != CRD {sorted(enums[name])}")

    src = pathlib.Path("internal/gate/constraints.go").read_text()
    fields = set(re.findall(r'^\t"(\w+)":', src, re.M))
    for name in sorted(fields):
        if name not in cons:
            problems.append(f"constraint {name}: no CRD property constrains it")
        elif len(cons[name]) > 1:
            problems.append(f"constraint {name}: {len(cons[name])} different constraint sets in the CRDs, so one table entry cannot be right")

    kube = crd.parent
    if (kube / "pkg/apis/asgard/v1alpha1/types.go").is_file():
        problems += check_cel(crd, kube)
    else:
        print(f"  (the CEL counts need the Go types beside {crd}; skipped)")

    for p in problems:
        print(f"  {p}")
    print(f"\n{len(mine)} enum field(s), {len(fields)} constrained field(s), {len(problems)} disagreement(s)")
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
