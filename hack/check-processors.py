#!/usr/bin/env python3
"""Hold `wiki/processors.md`'s two tables against the repositories they distil.

That page is the most claim-dense one in the corpus - thirteen processors, their
required keys, their defaults, their outputs, and which keys an author may set -
and every one of those claims belongs to somebody else's source file. It was
written by reading, and reading is what "never re-walked" in TASK.md meant.

**The two tables have two different owners, and that is the point of the page.**

    the definitions table   asgard-core `internal/constants.go`,
                            `ProcessorDefinitions` - what the runtime validates
    the palette table       asgard-docs `.../docs/processor/*/metadata.json` -
                            what the builder lets an author type

They disagree, on purpose, and the page says where. So this checks each table
against its own owner and never against the other.

**The literal is walked by brace depth, not matched by pattern.** An earlier
pattern-based extraction of this same literal misaligned - it attributed one
processor's fields to the next - and a table that is confidently wrong about
`allowWrite` is worse than no table. Every identifier must resolve to a string
or this exits; an unresolved one means the literal grew a shape this does not
understand, which is exactly when a reader must not trust the output.

    hack/check-processors.py            against the clones as they stand
    hack/check-processors.py --dump     print what upstream says, and stop

Needs $ASGARD_CORE and $ASGARD_DOCS. See hack/sources.py.
"""

import argparse
import json
import pathlib
import re
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from sources import commit, resolve  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
PAGE = ROOT / "internal/corpus/wiki/processors.md"

CORE_FILE = "internal/constants.go"
KUBE_FILE = "pkg/apis/asgard/v1alpha1/types.go"
PALETTE_DIR = "content-generator/services/developer-reference/docs/processor"

# How many there are. Stated here so that a definition appearing or vanishing is
# a failure rather than a quietly shorter table - the page says thirteen in four
# places, and a fourteenth processor is the single most consequential thing that
# can happen to it.
EXPECTED = 13

# A `const`/`var` line binding a name to a string literal. The key constants,
# the relation keys and the processor types are all this shape, in two
# repositories.
DECL = re.compile(r'^\s*(?:var\s+|const\s+)?([A-Za-z0-9_]+)\s+(?:[A-Za-z0-9_.]+\s+)?=\s*"([^"]*)"\s*$', re.M)

NUMERIC = re.compile(r'(?:int32|int64|float32|float64)\((-?[0-9.]+)\)')


def constants(core: str, kube: str) -> dict:
    """Every string constant either repository binds, by name.

    asgard-core's literal names asgard-kube's processor types through the
    `v1alpha1.` qualifier and its own keys bare, so both spellings are stored.
    """
    out = {}
    for src, prefix in ((core, ""), (kube, "v1alpha1.")):
        for name, value in DECL.findall(src):
            out[prefix + name] = value
            if prefix:
                out[name] = value
    return out


def literal(src: str, decl: str) -> str:
    """The body of a brace-delimited literal, by depth rather than by pattern."""
    at = src.index(decl)
    start = src.index("{", at)
    depth = 0
    for i in range(start, len(src)):
        if src[i] == "{":
            depth += 1
        elif src[i] == "}":
            depth -= 1
            if depth == 0:
                return src[start + 1:i]
    sys.exit(f"{decl} is not closed in {CORE_FILE}, so this cannot be walked")


def elements(body: str) -> list:
    """The top-level `{...}` elements of a slice literal."""
    out, depth, start = [], 0, 0
    for i, ch in enumerate(body):
        if ch == "{":
            if depth == 0:
                start = i
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                out.append(body[start + 1:i])
    return out


def field(text: str, name: str):
    """The value of a top-level `Name:` field inside one literal.

    Depth-tracked, so a `Name:` belonging to a nested `ConfigDefinition` is not
    read as the outer `ProcessorDefinition`'s. That confusion is precisely the
    misalignment this file's docstring warns about.
    """
    depth = 0
    for line in text.split("\n"):
        m = re.match(r'\s*([A-Za-z][A-Za-z0-9_]*):\s*(.*?),?\s*$', line)
        if depth == 0 and m and m.group(1) == name:
            return m.group(2).rstrip(",")
        depth += line.count("{") + line.count("[") - line.count("}") - line.count("]")
    return None


def value(raw, consts: dict):
    """Resolve one Go expression to a Python value, or exit."""
    if raw is None:
        return None
    raw = raw.strip().rstrip(",").lstrip("&")
    if raw.startswith('"'):
        return raw[1:-1]
    if raw in ("nil", "true", "false"):
        return {"nil": None, "true": True, "false": False}[raw]
    if raw in consts:
        return consts[raw]
    m = NUMERIC.fullmatch(raw)
    if m:
        return float(m.group(1)) if "." in m.group(1) else int(m.group(1))
    sys.exit(f"{CORE_FILE}: cannot resolve {raw!r} to a value. The literal has "
             "grown a shape this walk does not understand; do not trust the page "
             "until this is taught it.")


def definitions(core_dir: pathlib.Path, kube_dir: pathlib.Path) -> list:
    """`ProcessorDefinitions`, walked."""
    core = (core_dir / CORE_FILE).read_text()
    kube = (kube_dir / KUBE_FILE).read_text()
    consts = constants(core, kube)

    out = []
    for element in elements(literal(core, "var ProcessorDefinitions = []ProcessorDefinition{")):
        configs = []
        m = re.search(r'StaticConfigs:\s*\[\]ConfigDefinition\{', element)
        if m:
            for one in elements(literal(element[m.start():], "StaticConfigs:")):
                default = field(one, "DefaultValue")
                configs.append({
                    "name": value(field(one, "Name"), consts),
                    "required": value(field(one, "IsRequired"), consts) is True,
                    "default": value(default, consts),
                    "hasDefault": default not in (None, "nil"),
                })
        rels = re.search(r'StaticRelationships:\s*\[\]v1alpha1\.RelationKey\{([^}]*)\}', element)
        out.append({
            "type": value(field(element, "Type"), consts),
            "configs": configs,
            "dynamic": field(element, "AllowDynamicConfig") == "true",
            "dynamicRelationship": field(element, "AllowDynamicRelationship") == "true",
            "relationships": [value(x, consts) for x in rels.group(1).split(",") if x.strip()] if rels else [],
        })
    return out


def palette(docs: pathlib.Path) -> dict:
    """The editor palette per processor type, from asgard-docs' metadata.

    Second hand on purpose: asgard-docs reads the palette out of
    `asgard-ai-platform-web`, which nothing here has a clone of, and records what
    it found beside each page. The page says so.
    """
    out = {}
    for f in sorted((docs / PALETTE_DIR).glob("*/metadata.json")):
        meta = json.loads(f.read_text())
        crd = meta.get("crdType")
        if not crd:
            continue
        out.setdefault(crd, []).append({
            "page": f.parent.name,
            "author": meta.get("authorConfigs", []),
            "platform": meta.get("platformSetKeys", []),
            "dynamic": meta.get("editorPalette", {}).get("dynamicConfig", False),
            "present": meta.get("editorPalette", {}).get("present", False),
            "scope": meta.get("editorPalette", {}).get("workflowSetTypes", []),
        })
    return out


# ── the page ──────────────────────────────────────────────────────────────

def rows(page: str, heading: str) -> dict:
    """The table under one heading, keyed by the `type` in its first cell."""
    body = page.split(heading, 1)[1]
    out = {}
    for line in body.split("\n"):
        if not line.startswith("|"):
            if out:
                break
            continue
        cells = [c.strip() for c in line.strip().strip("|").split("|")]
        m = re.match(r'^`([a-z-]+)`$', cells[0])
        if m:
            out[m.group(1)] = cells[1:]
    return out


def keys(cell: str) -> list:
    """The backticked keys in one cell, without the `=default` suffixes."""
    return [k for k in re.findall(r'`([A-Za-z][A-Za-z0-9_.-]*)`', cell)]


def defaults(cell: str) -> dict:
    """`key` =value pairs, as the definitions table writes them."""
    out = {}
    for k, v in re.findall(r'`([A-Za-z][A-Za-z0-9_.]*)`\s*=(\S+)', cell):
        out[k] = v
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--dump", action="store_true", help="print upstream and stop")
    args = ap.parse_args()

    core, kube, docs = resolve("core"), resolve("kube"), resolve("docs")

    defs = {d["type"]: d for d in definitions(core, kube)}
    pal = palette(docs)

    if args.dump:
        for name in sorted(defs):
            d = defs[name]
            print(f"{name}  outputs={','.join(d['relationships']) or 'none'}  "
                  f"dynamic={d['dynamic']}  dynamicRelationship={d['dynamicRelationship']}")
            for c in d["configs"]:
                mark = "required" if c["required"] else "optional"
                shown = f" = {c['default']!r}" if c["hasDefault"] else ""
                print(f"    {c['name']:<38} {mark}{shown}")
            for row in pal.get(name, []):
                print(f"    palette({row['page']}): author={' '.join(row['author']) or 'none'}"
                      f" platform={' '.join(row['platform']) or 'none'}"
                      f" dynamic={row['dynamic']} scope={','.join(row['scope']) or 'not in the palette'}")
        return 0

    page = PAGE.read_text()
    bad = []

    if len(defs) != EXPECTED:
        bad.append(f"asgard-core declares {len(defs)} processors and the page is written for {EXPECTED}. "
                   "Every count on that page, and the CRD enum, has to be re-read.")

    # ── the definitions table ────────────────────────────────────────────
    table = rows(page, "| processor | outputs | extra keys | required keys, with any default |")
    for name in sorted(set(defs) | set(table)):
        if name not in table:
            bad.append(f"{name} is in asgard-core and has no row in the definitions table")
            continue
        if name not in defs:
            bad.append(f"the definitions table has a row for {name}, which asgard-core does not declare")
            continue
        d, cells = defs[name], table[name]
        outputs, extra, required = cells[0], cells[1], cells[2]

        want = {"success": "Success", "failure": "Failure", "else": "Else"}
        said = {w for w in want.values() if w in outputs} or set()
        has = {want[r] for r in d["relationships"] if r in want}
        if said != has:
            bad.append(f"{name} outputs: the page says {' + '.join(sorted(said)) or 'none'}, "
                       f"asgard-core declares {' + '.join(sorted(has)) or 'none'}")

        if d["dynamic"] != ("yes" in extra):
            bad.append(f"{name} extra keys: the page says {extra!r}, "
                       f"asgard-core has AllowDynamicConfig={d['dynamic']}")

        want_req = {c["name"] for c in d["configs"] if c["required"]}
        said_req = set(keys(required))
        if want_req != said_req:
            missing = ", ".join(sorted(want_req - said_req))
            extra_k = ", ".join(sorted(said_req - want_req))
            bad.append(f"{name} required keys: the page {'omits ' + missing if missing else ''}"
                       f"{' and ' if missing and extra_k else ''}"
                       f"{'claims ' + extra_k if extra_k else ''}")

        # `=x` in a cell says the key is required AND has a default, which is the
        # distinction the page's `=` legend exists for.
        want_def = {c["name"]: c["default"] for c in d["configs"] if c["required"] and c["hasDefault"]}
        said_def = defaults(required)
        for k, v in want_def.items():
            shown = str(v).lower() if isinstance(v, bool) else ('""' if v == "" else str(v))
            if k not in said_def:
                bad.append(f"{name} `{k}` has a default of {shown} in asgard-core and the page marks none")
            elif said_def[k].strip('"') not in (shown.strip('"'),):
                bad.append(f"{name} `{k}`: the page says ={said_def[k]}, asgard-core says {shown}")
        for k in said_def:
            if k not in want_def:
                bad.append(f"{name} `{k}`: the page marks a default asgard-core does not give it")

    # ── the palette table ────────────────────────────────────────────────
    ptable = rows(page, "| type | author sets | platform sets | dynamic | scope |")
    for name, entries in sorted(pal.items()):
        if name not in ptable:
            bad.append(f"{name} is in the asgard-docs palette metadata and has no row in the palette table")
            continue
        cells = ptable[name]
        author_cell, platform_cell, dynamic_cell, scope_cell = cells[0], cells[1], cells[2], cells[3]

        # One CRD type can own two pages - `push-message` is reached as a bot
        # reply and as an Automation Tool's response - and the palette entry is
        # the same for both. Any of them may answer.
        if not any(e["dynamic"] == ("yes" in dynamic_cell) for e in entries):
            bad.append(f"{name} dynamic: the palette table says {dynamic_cell!r}, "
                       f"asgard-docs records dynamicConfig={entries[0]['dynamic']}")

        for e in entries:
            if not e["present"] and "not in the palette" not in scope_cell:
                bad.append(f"{name} is absent from the palette per asgard-docs, "
                           f"and the page's scope cell says {scope_cell!r}")
            for s in e["scope"]:
                if s not in scope_cell.replace("_", "_"):
                    bad.append(f"{name} scope: asgard-docs says {s!r}, "
                               f"which the page's scope cell {scope_cell!r} does not name")

        # The author and platform cells are checked per key in one direction:
        # a key asgard-docs calls the platform's must not sit in the author
        # column, because that is the error that gets a chart to write a key it
        # does not own. Rows written as a delta off the row above ("the same,
        # plus ...") carry no key list of their own and are skipped here.
        if "the same" not in author_cell:
            said_author, said_platform = set(keys(author_cell)), set(keys(platform_cell))
            for e in entries:
                for k in e["platform"]:
                    if k in said_author:
                        bad.append(f"{name} `{k}`: the page has it as the author's, "
                                   f"asgard-docs records it as the platform's")
                    if k not in said_platform and platform_cell != "-":
                        bad.append(f"{name} `{k}`: asgard-docs records it as a platform key "
                                   f"and the page's platform cell does not name it")
                for k in e["author"]:
                    if k not in said_author and "one per branch" not in author_cell and author_cell != "*none*":
                        bad.append(f"{name} `{k}`: asgard-docs records it as an author key "
                                   f"and the page's author cell does not name it")

    # ── the extra-key table ──────────────────────────────────────────────
    #
    # What a dynamic key MEANS is not in the definitions - it is in the loop
    # that reads it, one file per processor in asgard-core. So this checks only
    # the thing the definitions can answer: that every processor accepting
    # dynamic config has a row saying what its keys are for. A processor that
    # becomes dynamic and gets no row is an author writing keys into a void,
    # which `llm-completion` already does.
    etable = rows(page, "| processor | an extra key is | the shape |")
    for name, d in sorted(defs.items()):
        if d["dynamic"] and name not in etable:
            bad.append(f"{name} accepts dynamic config and has no row saying what an extra key means there")
    for name in sorted(etable):
        if name in defs and not defs[name]["dynamic"]:
            bad.append(f"the extra-key table has a row for {name}, which asgard-core does not accept dynamic config on")

    print(f"processors: {len(defs)} definitions, {len(table)} rows in the definitions table, "
          f"{len(ptable)} in the palette table, {len(etable)} in the extra-key table")
    print(f"  asgard-core {commit(core) or '?'}, asgard-docs {commit(docs) or '?'}, "
          f"asgard-kube {commit(kube) or '?'}")
    for b in bad:
        print(f"  {b}")
    print(f"{len(bad)} disagreement(s)")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
