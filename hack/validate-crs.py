"""Validate CR documents against the platform CRDs.

    for f in <asgard-kube>/crd/*.yaml; do yq -o=json "$f" > .out/crdjson/$(basename $f .yaml).json; done
    python3 scripts/validate-crs.py .out/crdjson <documents.ndjson>

Pull asgard-kube first - validating against an old clone proves nothing. Checks
required fields, fields absent from the schema (the apiserver prunes those
silently), enums, patterns, maxItems, and the ExactlyOneOf CEL rules. None of
that is done by helm lint, by `asgard-cli check`, or by a server-side dry-run.

Two sentinels from extract-crs.py mean "not visible from an extract" and are not
reported: __HELM__ for a value the chart supplies, __PLACEHOLDER__ for one the
reader fills in.
"""
import glob, json, os, re, sys

schemas = {}
for f in glob.glob(os.path.join(sys.argv[1], '*.json')):
    d = json.load(open(f))
    kind = d['spec']['names']['kind']
    schemas[kind] = d['spec']['versions'][0]['schema']['openAPIV3Schema']


SENTINELS = ('__HELM__', '__PLACEHOLDER__')

def sentinel(v):
    return isinstance(v, str) and any(x in v for x in SENTINELS)

EXACTLY_ONE = re.compile(r"exactly one of the fields in \[([^\]]+)\]")

def cel(node, obj, path, errs):
    for r in (node.get('x-kubernetes-validations') or []):
        msg = r.get('message') or ''
        m = EXACTLY_ONE.search(msg)
        if m and isinstance(obj, dict):
            names = m.group(1).split()
            present = [n for n in names if n in obj]
            if len(present) != 1:
                errs.append(f"{path}: {msg} (present: {present or 'none'})")

def check(node, obj, path, errs):
    if not isinstance(node, dict):
        return
    t = node.get('type')
    if t == 'object':
        if not isinstance(obj, dict):
            return
        cel(node, obj, path, errs)
        props = node.get('properties') or {}
        req = node.get('required') or []
        for r in req:
            if r not in obj:
                errs.append(f"{path}.{r}: REQUIRED but missing")
        ap = node.get('additionalProperties')
        if props:
            for k, v in obj.items():
                if k in props:
                    check(props[k], v, f"{path}.{k}", errs)
                elif isinstance(ap, dict):
                    check(ap, v, f"{path}.{k}", errs)
                elif ap is not True:
                    errs.append(f"{path}.{k}: UNKNOWN field (pruned silently by the apiserver)")
        elif isinstance(ap, dict):
            for k, v in obj.items():
                check(ap, v, f"{path}.{k}", errs)
    elif t == 'array':
        if not isinstance(obj, list):
            return
        cel(node, obj, path, errs)
        items = node.get('items') or {}
        mi = node.get('maxItems')
        if mi is not None and len(obj) > mi:
            errs.append(f"{path}: {len(obj)} items exceeds maxItems={mi}")
        for i, v in enumerate(obj):
            check(items, v, f"{path}[{i}]", errs)
    else:
        if obj is None:
            return
        if sentinel(obj):
            return
        if node.get('enum') is not None and obj not in node['enum']:
            errs.append(f"{path}: {obj!r} not in enum {node['enum']}")
        if node.get('pattern') and isinstance(obj, str):
            if not re.search(node['pattern'], obj):
                errs.append(f"{path}: {obj!r} fails pattern {node['pattern']}")
        if node.get('minLength') and isinstance(obj, str) and len(obj) < node['minLength']:
            errs.append(f"{path}: shorter than minLength={node['minLength']}")

total = 0
for line in open(sys.argv[2]):
    line = line.strip()
    if not line:
        continue
    doc = json.loads(line)
    kind = doc.get('kind')
    name = (doc.get('metadata') or {}).get('name')
    if kind not in schemas:
        print(f"\n### {kind}/{name}: NO CRD for this kind")
        continue
    errs = []
    sch = schemas[kind]
    for top in ('spec',):
        if top in doc:
            check(sch['properties'][top], doc[top], top, errs)
        elif top in (sch.get('required') or []):
            errs.append(f"{top}: REQUIRED but missing")
    if errs:
        total += len(errs)
        print(f"\n### {kind}/{name}")
        for e in errs:
            print("   ", e)
print(f"\n==== {total} schema violation(s) ====")
