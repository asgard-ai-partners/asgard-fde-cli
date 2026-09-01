#!/usr/bin/env python3
"""check_crd_fidelity.py — 比對 render 出來的 CR 與**叢集上實際的 CRD schema**,找出會被靜默丟掉的欄位。

    asgard-cli render internal dev | python3 scripts/check_crd_fidelity.py - --context <ctx>

## 為什麼需要這一支(2026-08-24 實際踩過)

`kubectl apply --dry-run=server` **不會**擋下 CRD 沒宣告的欄位 —— CRD 預設 prune 未知欄位,
所以 dry-run 會回報「created」而那個欄位已經被丟掉了。`ts-mail` 帶著已棄用的
`spec.instruction` 通過了 25/25 的 dry-run,然後在 CD 上被 helm 的 server-side apply 擋下:

    .spec.instruction: field not declared in schema

dry-run 驗證的是「會不會被接受」,這一支驗證的是「會不會被完整保留」。兩者都要跑。

## 它做什麼

拉出叢集上每個 `asgard-ai.com` CRD 的 openAPIV3Schema,遞迴走過 render 出來的每個 CR,
回報任何 schema 沒宣告的欄位路徑。遇到 `x-kubernetes-preserve-unknown-fields: true`
的子樹就停止下探(那裡的自由欄位是刻意的)。

需要能讀該叢集的 CRD(唯讀)。沒有叢集可連時本檢查應視為「未執行」,不是通過。
"""
from __future__ import annotations

import argparse
import json
import subprocess
import sys

try:
    import yaml
except ModuleNotFoundError:
    sys.stderr.write("需要 PyYAML:.venv/bin/pip install -r scripts/db/requirements.txt\n")
    sys.exit(2)

GROUP = "asgard-ai.com"
# metadata 由 apiserver 定義,CRD 的 schema 通常只寫 `type: object` 不列子欄位 —— 不是缺漏。
SKIP_ROOTS = {"metadata"}


def fetch_crd_schemas(context: str | None) -> dict[str, dict]:
    """kind → 該 CRD 儲存版本的 openAPIV3Schema。"""
    cmd = ["kubectl"]
    if context:
        cmd += ["--context", context]
    cmd += ["get", "crd", "-o", "json"]
    out = subprocess.run(cmd, capture_output=True, text=True)
    if out.returncode != 0:
        sys.stderr.write(f"✗ 讀不到叢集 CRD:{out.stderr.strip()[:400]}\n")
        sys.exit(2)
    schemas: dict[str, dict] = {}
    for item in json.loads(out.stdout).get("items", []):
        spec = item["spec"]
        if spec["group"] != GROUP:
            continue
        versions = spec.get("versions", [])
        chosen = next((v for v in versions if v.get("storage")), versions[0] if versions else None)
        if chosen and chosen.get("schema", {}).get("openAPIV3Schema"):
            schemas[spec["names"]["kind"]] = chosen["schema"]["openAPIV3Schema"]
    return schemas


def walk(value, schema: dict | None, path: str, problems: list[str]) -> None:
    """遞迴比對。schema 為 None 代表這一層 CRD 沒宣告 → 記錄並停止下探。"""
    if schema is None:
        problems.append(path)
        return
    # 這棵子樹允許自由欄位,不再往下查
    if schema.get("x-kubernetes-preserve-unknown-fields"):
        return
    if isinstance(value, dict):
        props = schema.get("properties")
        addl = schema.get("additionalProperties")
        for key, sub in value.items():
            if props is not None and key in props:
                walk(sub, props[key], f"{path}.{key}", problems)
            elif isinstance(addl, dict):
                walk(sub, addl, f"{path}.{key}", problems)
            elif addl is True or props is None:
                continue  # map 或無型別約束
            else:
                problems.append(f"{path}.{key}")
    elif isinstance(value, list):
        items = schema.get("items")
        for i, sub in enumerate(value):
            walk(sub, items, f"{path}[{i}]", problems)


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(
        description="Find CR fields the cluster's CRDs would silently prune")
    ap.add_argument("path", help="render 輸出的 YAML 檔,或 - 讀 stdin")
    ap.add_argument("--context", help="kubectl context(省略則用當前 context)")
    args = ap.parse_args(argv)

    raw = sys.stdin.read() if args.path == "-" else open(args.path, encoding="utf-8").read()
    docs = [d for d in yaml.safe_load_all(raw) if d]
    schemas = fetch_crd_schemas(args.context)
    if not schemas:
        sys.stderr.write(f"✗ 該叢集沒有任何 {GROUP} CRD —— 是不是連錯叢集?\n")
        return 2

    problems: list[str] = []
    unknown_kinds: set[str] = set()
    checked = 0
    for d in docs:
        kind, name = d["kind"], d["metadata"]["name"]
        schema = schemas.get(kind)
        if schema is None:
            unknown_kinds.add(kind)
            continue
        checked += 1
        found: list[str] = []
        for key, sub in d.items():
            if key in SKIP_ROOTS:
                continue
            props = schema.get("properties", {})
            walk(sub, props.get(key), key, found)
        problems += [f"{kind}/{name} → {p}" for p in found]

    ctx = args.context or "(current context)"
    if unknown_kinds:
        print(f"! 叢集上找不到這些 kind 的 CRD,已跳過:{', '.join(sorted(unknown_kinds))}")
    if problems:
        for p in problems:
            print(f"ERROR 欄位不在 CRD schema 中,apply 時會被靜默丟掉:{p}")
        print(f"\n✗ {len(problems)} 個欄位會被 prune — 檢查了 {checked} 個 CR @ {ctx}")
        print("  註:kubectl apply --dry-run=server 抓不到這類問題(prune 不算錯誤),"
              "但 helm 的 server-side apply 會直接失敗。")
        return 1
    print(f"✓ CRD fidelity OK — {checked} 個 CR 的所有欄位都在 schema 中 @ {ctx}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
