#!/usr/bin/env python3
"""nsquery.py — 對 NetSuite 跑任意 SuiteQL 並印出結果(唯讀)。

`query.py` 的 NetSuite 版。那支走 DBAPI 連四個 DB target;NetSuite 沒有 DBAPI,
所以走 `nsenv.py` 的 REST 路徑(JWT → token → POST suiteql)。

用法:
    .venv/bin/python scripts/db/nsquery.py "SELECT 1 AS ok FROM DUAL"
    .venv/bin/python scripts/db/nsquery.py -f some.sql
    echo "SELECT ..." | .venv/bin/python scripts/db/nsquery.py
    .venv/bin/python scripts/db/nsquery.py --limit 5 "SELECT ..."
    .venv/bin/python scripts/db/nsquery.py --all "SELECT ..."     # 翻頁撈完
    .venv/bin/python scripts/db/nsquery.py --json "SELECT ..."    # 原始 JSON

探索 schema 用 `--columns`:SuiteQL 沒有 information_schema,但 `SELECT * ... WHERE ROWNUM=1`
可以把某張表的欄位名(含 NetSuite 實際使用的大小寫)全部列出來 —— 這是建 SemanticLayer 時
確認 dimension `name` 的手段。

    .venv/bin/python scripts/db/nsquery.py --columns transaction
"""
from __future__ import annotations

import argparse
import json
import sys

import nsenv  # 同目錄


def render(items: list[dict], meta: str = "") -> str:
    if not items:
        return f"(0 rows){(' ' + meta) if meta else ''}"
    # NetSuite 各列的 key 可能不齊(值為 null 時整個 key 會被省略),故聯集所有 key。
    # `links` 是 NetSuite 給每列附的 HATEOAS 欄位,對查詢結果沒有意義,直接濾掉。
    cols: list[str] = []
    for it in items:
        for k in it:
            if k != "links" and k not in cols:
                cols.append(k)
    widths = {c: len(c) for c in cols}
    srows = [{c: ("" if it.get(c) is None else str(it.get(c))) for c in cols} for it in items]
    for r in srows:
        for c in cols:
            widths[c] = max(widths[c], len(r[c]))
    widths = {c: min(w, 48) for c, w in widths.items()}
    line = lambda cells: " | ".join(str(cells[c])[: widths[c]].ljust(widths[c]) for c in cols)
    out = [line({c: c for c in cols}), "-+-".join("-" * widths[c] for c in cols)]
    out += [line(r) for r in srows]
    out.append(f"({len(items)} row{'s' if len(items) != 1 else ''}){(' ' + meta) if meta else ''}")
    return "\n".join(out)


def show_columns(table: str) -> int:
    """列出一張表的欄位名與一列樣本值。SuiteQL 沒有 information_schema,只能靠取一列反推。"""
    payload = nsenv.query(f"SELECT * FROM {table} WHERE ROWNUM = 1", limit=1)
    items = payload.get("items", [])
    if not items:
        print(f"-- {table}:查得到表但沒有資料列,無法反推欄位", file=sys.stderr)
        return 1
    row = {k: v for k, v in items[0].items() if k != "links"}
    print(f"-- {table}:{len(row)} 個欄位(NetSuite 一律以小寫回傳欄名)")
    print("   註:值為 NULL 的欄位整個 key 會被省略,所以這裡看不到的欄位不代表不存在 ——")
    print("       換一列取樣(例如 WHERE mainline='F')可能會多出欄位。")
    width = max(len(k) for k in row)
    for k in sorted(row):
        v = row[k]
        print(f"  {k.ljust(width)}  {type(v).__name__:5s}  {str(v)[:60]}")
    return 0


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(description="Run SuiteQL against NetSuite and print results")
    ap.add_argument("sql", nargs="?", help="SuiteQL 字串(省略則從 -f 或 stdin 讀)")
    ap.add_argument("-f", "--file", help="從檔案讀 SuiteQL")
    ap.add_argument("--limit", type=int, default=200,
                    help=f"單頁列數上限(預設 200,NetSuite 硬上限 {nsenv.MAX_PAGE_LIMIT})")
    ap.add_argument("--offset", type=int, default=0, help="起始位移(翻頁用)")
    ap.add_argument("--all", action="store_true", help="翻頁把結果撈完(上限 10000 列)")
    ap.add_argument("--json", action="store_true", help="印出原始 JSON 而非表格")
    ap.add_argument("--columns", metavar="TABLE",
                    help="列出某張表的欄位名(SuiteQL 無 information_schema,靠取一列反推)")
    args = ap.parse_args(argv)

    print(f"=== {nsenv.describe()} ===", file=sys.stderr)

    if args.columns:
        return show_columns(args.columns)

    if args.file:
        sql = open(args.file, encoding="utf-8").read()
    elif args.sql:
        sql = args.sql
    else:
        sql = sys.stdin.read()
    if not sql.strip():
        print("no SuiteQL provided", file=sys.stderr)
        return 1

    if args.all:
        items = nsenv.query_all(sql)
        print(json.dumps(items, ensure_ascii=False, indent=2) if args.json
              else render(items, "(翻頁撈完)"))
        return 0

    payload = nsenv.query(sql, limit=args.limit, offset=args.offset)
    if args.json:
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 0
    meta = f"totalResults={payload.get('totalResults')} hasMore={payload.get('hasMore')}"
    print(render(payload.get("items", []), meta))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
