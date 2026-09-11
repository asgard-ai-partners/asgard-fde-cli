#!/usr/bin/env python3
"""Hold internal/gate's pinned tables against the CRDs the apiserver enforces.

The tables in `internal/gate` were extracted from asgard-kube's Go types, and
that is not the contract - the generated CRDs are. The difference is not
academic: `status` carries three values in the Asgard types and six in the CRD,
because Kubernetes' own condition schema uses the same field name, and the
entry sat in the enum table for a day before this check existed.

**Run it whenever asgard-kube moves**, and move the read markers in
`internal/gate` in the same change. Point it at the checkout:

    hack/check-tables.py ~/projects/asgard/asgard-kube/crd

That is the whole ritual now. It was three lines of shell, so it did not get
run: the tables went eight upstream commits unchecked, and in that window the
platform deleted the cron `schedule` pattern this repository was still
enforcing - which made `verify` report four correct schedules as violations
and sent one extract's readers to build five Triggers instead of a range.

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

    **The conversion used to be three lines of shell in `README.md`**, so
    running this meant remembering them, and the tables went eight upstream
    commits without being held against anything - long enough for the platform
    to delete a pattern this repository still enforced, and for an extract to
    go on telling readers to build five Triggers because of it. A ritual that
    is not one command is a ritual that does not happen.
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


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: check-tables.py <asgard-kube/crd, as yaml or as json>")
    enums, cons = crd_properties(as_json(pathlib.Path(sys.argv[1])))
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

    for p in problems:
        print(f"  {p}")
    print(f"\n{len(mine)} enum field(s), {len(fields)} constrained field(s), {len(problems)} disagreement(s)")
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
