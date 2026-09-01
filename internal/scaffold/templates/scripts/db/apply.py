#!/usr/bin/env python3
"""apply.py — 把各產業的 .sql 套用到 Postgres(dev/prod)。

.sql 與來源同放:
    {industry}/systems/db/{system}/{table}.sql   # OLTP
    {industry}/systems/warehouse/{table}.sql           # OLAP

每產業一個交易(全表 DDL+COPY 原子化,commit 後進下一產業),失敗自動 rollback。
表檔含 `DROP TABLE … CASCADE` + `CREATE`,重跑即冪等整包置換。透過 psycopg 執行
(連線取單一 ASGARD_PG_DSN,讀專案根 .env 或環境變數),不需系統 psql。
dev / prod 靠所在環境切分,不在此選。

用法:.venv/bin/python scripts/db/apply.py [industry ...] [--layer oltp|olap|both] [--dry-run]
      切到別的 DB:ASGARD_PG_DSN=... .venv/bin/python scripts/db/apply.py
"""
from __future__ import annotations

import argparse
import pathlib
import sys

import pgenv  # 同目錄

ROOT = pathlib.Path(__file__).resolve().parents[2]


def industry_sql_files(industry: pathlib.Path, layers: list[str]) -> list[pathlib.Path]:
    """該產業在指定層的 .sql,依層序、路徑排序(DDL 自含 CREATE SCHEMA,順序無關)。"""
    files: list[pathlib.Path] = []
    if "oltp" in layers:
        files += sorted((industry / "systems" / "db").rglob("*.sql"))
    if "olap" in layers:
        files += sorted((industry / "systems" / "warehouse").glob("*.sql"))
    return files


def industries(argv: list[str], layers: list[str]) -> list[pathlib.Path]:
    return sorted(
        (p for p in ROOT.iterdir()
         if p.is_dir() and (p / "stories").is_dir()
         and (not argv or p.name in argv)
         and industry_sql_files(p, layers)),
        key=lambda p: p.name,
    )


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(description="Apply per-industry .sql to Postgres via psycopg")
    ap.add_argument("industries", nargs="*", help="限定產業(預設全部)")
    ap.add_argument("--layer", choices=["oltp", "olap", "both"], default="both")
    ap.add_argument("-d", "--db", help="PostgreSQL target(省略走 legacy ASGARD_PG_DSN)")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args(argv)

    if args.db and pgenv.DB_TARGETS.get(args.db, {}).get("driver") != "postgres":
        print(f"✗ apply 僅支援 PostgreSQL target;'{args.db}' 非 postgres", file=sys.stderr)
        return 1

    layers = ["oltp", "olap"] if args.layer == "both" else [args.layer]
    inds = industries(args.industries, layers)
    if not inds:
        print("no matching industries with .sql(systems/db 或 systems/warehouse 下無 .sql?)", file=sys.stderr)
        return 1

    if args.dry_run:
        print("=== dry-run ===")
        for industry in inds:
            print(f"  {industry.name}: {len(industry_sql_files(industry, layers))} tables")
        return 0

    print(f"=== {pgenv.describe(args.db)} ===")
    rc = 0
    with pgenv.connect(args.db) as conn:
        for industry in inds:
            files = industry_sql_files(industry, layers)
            try:
                with conn.transaction():
                    for f in files:
                        pgenv.run_sql_file(conn, f)
                print(f"  ✓ {industry.name}: {len(files)} tables")
            except Exception as exc:  # noqa: BLE001  (回報並續跑下一產業)
                rc = 1
                print(f"  ✗ {industry.name}: {type(exc).__name__}: {exc}", file=sys.stderr)
    return rc


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
