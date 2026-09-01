#!/usr/bin/env python3
"""query.py — 對設定的 DB 跑任意 SQL 並印出結果(PostgreSQL / MSSQL)。

供臨時查詢/巡檢/驗證用,不需系統 psql。用 -d/--db 選具名 target(見 pgenv.DB_TARGETS);
省略 -d 則走 legacy ASGARD_PG_DSN。

用法:
    .venv/bin/python scripts/db/query.py -d <target> "select count(*) from ..."   # PostgreSQL
    .venv/bin/python scripts/db/query.py -d <target> "select top 1 * from ..."       # MSSQL
    .venv/bin/python scripts/db/query.py -d <target> -f some.sql
    echo "select 1" | .venv/bin/python scripts/db/query.py -d <target>            # 從 stdin 讀
"""
from __future__ import annotations

import argparse
import sys

import pgenv  # 同目錄


def render(cur) -> str:
    if cur.description is None:
        return f"({cur.rowcount} rows affected)" if cur.rowcount >= 0 else "(ok)"
    # psycopg 的 Column 有 .name;pymssql 的 description 是 tuple,欄名在 [0]
    cols = [getattr(d, "name", None) or d[0] for d in cur.description]
    rows = cur.fetchall()
    widths = [len(c) for c in cols]
    srows = [[("" if v is None else str(v)) for v in r] for r in rows]
    for r in srows:
        widths = [max(w, len(c)) for w, c in zip(widths, r)]
    line = lambda cells: " | ".join(c.ljust(w) for c, w in zip(cells, widths))
    out = [line(cols), "-+-".join("-" * w for w in widths)]
    out += [line(r) for r in srows]
    out.append(f"({len(rows)} row{'s' if len(rows) != 1 else ''})")
    return "\n".join(out)


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(description="Run SQL against the configured DB and print results")
    ap.add_argument("sql", nargs="?", help="SQL 字串(省略則從 -f 或 stdin 讀)")
    ap.add_argument("-f", "--file", help="從檔案讀 SQL")
    ap.add_argument("-d", "--db", help=f"DB target:{', '.join(pgenv.DB_TARGETS)}(省略走 legacy DSN)")
    args = ap.parse_args(argv)

    if args.file:
        sql = open(args.file, encoding="utf-8").read()
    elif args.sql:
        sql = args.sql
    else:
        sql = sys.stdin.read()
    if not sql.strip():
        print("no SQL provided", file=sys.stderr)
        return 1

    print(f"=== {pgenv.describe(args.db)} ===", file=sys.stderr)
    with pgenv.connect(args.db, autocommit=True) as conn:
        with conn.cursor() as cur:
            cur.execute(sql)  # type: ignore[arg-type]
            print(render(cur))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
