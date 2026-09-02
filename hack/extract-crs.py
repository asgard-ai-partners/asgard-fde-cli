"""Pull the CR skeletons out of the extracts, as documents a schema check can read.

    python3 scripts/extract-crs.py .out/extracts.ndjson

The skeletons in internal/usecase/extracts/ are chart fragments written for a
person to copy, so they are not parseable YAML: they carry Helm actions and
<placeholder> text. Both are replaced with sentinels here. validate-crs.py knows
the sentinels and does not report a pattern or enum failure against one, because
what the chart really supplies is not visible from the extract.

Writes NDJSON, plus a .src file mapping each document back to its extract.
"""
import glob, re, json, subprocess, sys

HELM = "__HELM__"          # a value the chart supplies
PLACE = "__PLACEHOLDER__"  # a value the reader fills in

def defuse(block):
    out = []
    for line in block.split('\n'):
        s = line.strip()
        if s.startswith('{{') and s.endswith('}}'):
            indent = line[:len(line) - len(line.lstrip())]
            control = any(s.startswith(p) for p in ('{{- if', '{{ if', '{{- end', '{{ end', '{{- range', '{{ range', '{{- with', '{{ with'))
            # A control action becomes a comment; an injecting one becomes a map
            # entry, which is enough for the parser and is required by nothing.
            out.append(f'{indent}# {s}' if control else f'{indent}helmInjected: "x"')
            continue
        line = re.sub(r'\{\{.*?\}\}', HELM, line)
        line = re.sub(r'(?<!\|)<[^<>\n]*>', PLACE, line)
        out.append(line)
    return '\n'.join(out)

def main(outpath):
    docs, blocks, bad = [], 0, []
    for p in sorted(glob.glob('internal/usecase/extracts/*.md')):
        for b in re.findall(r'```ya?ml\n(.*?)```', open(p).read(), re.S):
            if 'apiVersion: asgard-ai.com' not in b:
                continue
            blocks += 1
            r = subprocess.run(['yq', '-o=json', '-I=0', '.'], input=defuse(b), capture_output=True, text=True)
            if r.returncode != 0:
                bad.append((p, r.stderr.strip().splitlines()[0][:100]))
                continue
            for line in r.stdout.strip().split('\n'):
                if line.strip() and line.strip() != 'null':
                    d = json.loads(line)
                    if str(d.get('apiVersion', '')).startswith('asgard-ai.com'):
                        docs.append((p, d))
    for p, e in bad:
        print(f"PARSE FAIL {p}: {e}", file=sys.stderr)
    print(f"blocks: {blocks}  documents: {len(docs)}  unparseable: {len(bad)}", file=sys.stderr)
    with open(outpath, 'w') as f:
        for _, d in docs:
            f.write(json.dumps(d) + '\n')
    with open(outpath + '.src', 'w') as f:
        for p, d in docs:
            f.write(f"{d.get('kind')}/{(d.get('metadata') or {}).get('name')}\t{p}\n")
    return 1 if bad else 0

if __name__ == '__main__':
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else '.out/extracts.ndjson'))
