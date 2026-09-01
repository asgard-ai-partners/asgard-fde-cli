"""pgenv.py — DB 工具共用層:讀 .env、連線(PostgreSQL / MSSQL)、執行含 COPY 的 .sql。

連線改採「具名 target」模式:每個 target 對應 .env 的一組 `*_DB_*` 變數,並帶自己的
driver(postgres=psycopg / mssql=pymssql)。target 清單見 DB_TARGETS。

    # 例:對某個 PostgreSQL target 跑查詢
    .venv/bin/python scripts/db/query.py -d <target> "select 1"
    # 例:對某個 MSSQL target 跑查詢
    .venv/bin/python scripts/db/query.py -d <target> "select top 1 * from ..."

環境變數優先於 .env(同名覆蓋)。舊的單一 `ASGARD_PG_DSN` 仍相容:未指定 target 時,
若該變數存在則以 PostgreSQL DSN 連線(對應 legacy 行為)。

.sql 採 psql 慣例(`COPY … FROM stdin … ;` + inline CSV + `\\.`),由本層用 psycopg 的
`cursor.copy()` 執行 → 不需系統 psql。COPY 匯入僅支援 PostgreSQL target。
"""
from __future__ import annotations

import os
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[2]    # repo root
ENV_FILE = ROOT / ".env"
LEGACY_DSN_VAR = "ASGARD_PG_DSN"

# 具名 target → .env 變數前綴 + driver。新增 DB 時在此登記即可。
#
# 接上一個資料庫要做三件事,缺一不可:
#   1. 在這裡登記 target
#   2. 在 .env.example 補一組變數(並在自己的 .env 填真值)
#   3. 在對應 project 的 chart 裡新增 DataConnector CR
#
# 例:
#   "erp": {"prefix": "ERP_DB_", "driver": "mssql"},
DB_TARGETS: dict[str, dict[str, str]] = {}


def _load_env() -> dict[str, str]:
    """讀專案根 .env 為 dict(不覆蓋既有環境變數;環境變數優先)。"""
    env: dict[str, str] = {}
    if ENV_FILE.is_file():
        for raw in ENV_FILE.read_text(encoding="utf-8").splitlines():
            line = raw.strip()
            if line.startswith("#") or "=" not in line:
                continue
            key, val = line.split("=", 1)
            env[key.strip()] = val.strip().strip('"').strip("'")
    env.update({k: v for k, v in os.environ.items() if k in env or k.endswith("_DB_HOST")
                or k.endswith("_DB_PORT") or k.endswith("_DB_USER") or k.endswith("_DB_NAME")
                or k.endswith("_DB_PASSWORD") or k.endswith("_DB_SSLMODE") or k == LEGACY_DSN_VAR})
    return env


def _cfg(target: str) -> dict[str, str]:
    """組出某 target 的連線設定(host/port/user/password/database/sslmode/driver)。"""
    if target not in DB_TARGETS:
        raise SystemExit(f"✗ 未知 DB target '{target}';可用:{', '.join(DB_TARGETS)}")
    spec = DB_TARGETS[target]
    prefix, driver = spec["prefix"], spec["driver"]
    env = _load_env()
    get = lambda suffix: env.get(prefix + suffix, "")
    cfg = {
        "driver": driver,
        "host": get("HOST"),
        "port": get("PORT"),
        "user": get("USER"),
        "password": get("PASSWORD"),
        "database": get("NAME"),
        "sslmode": get("SSLMODE"),
    }
    missing = [k for k in ("host", "port", "user", "database") if not cfg[k]]
    if missing:
        raise SystemExit(
            f"✗ target '{target}' 缺少 {', '.join(prefix + m.upper() for m in missing)}"
            f"(填 {ENV_FILE} 或環境變數)"
        )
    return cfg


def describe(target: str | None) -> str:
    """給人看的連線摘要(不含密碼),供 CLI 顯示。"""
    if target is None:
        return f"legacy DSN ({LEGACY_DSN_VAR}) → " + dsn().rsplit("@", 1)[-1]
    c = _cfg(target)
    return f"{target} [{c['driver']}] {c['user']}@{c['host']}:{c['port']}/{c['database']}"


def dsn() -> str:
    """legacy:取單一 PostgreSQL DSN(環境變數優先,否則讀 .env)。"""
    env = _load_env()
    val = os.environ.get(LEGACY_DSN_VAR) or env.get(LEGACY_DSN_VAR)
    if not val:
        raise SystemExit(
            f"✗ {LEGACY_DSN_VAR} 未設定;請改用具名 target(-d {'/'.join(DB_TARGETS)})"
            f"或在 {ENV_FILE} 設 {LEGACY_DSN_VAR}"
        )
    return val


def connect(target: str | None = None, autocommit: bool = False):
    """回傳 DBAPI 連線。target=None 走 legacy PostgreSQL DSN;否則依 DB_TARGETS 分派 driver。"""
    if target is None:
        import psycopg  # lazy:僅 PostgreSQL 需要
        return psycopg.connect(dsn(), autocommit=autocommit)

    c = _cfg(target)
    if c["driver"] == "postgres":
        import psycopg  # lazy
        return psycopg.connect(
            host=c["host"], port=c["port"], user=c["user"],
            password=c["password"], dbname=c["database"],
            sslmode=c["sslmode"] or "prefer", autocommit=autocommit,
        )
    if c["driver"] == "mssql":
        import pymssql  # lazy:僅 MSSQL 需要
        return pymssql.connect(
            server=c["host"], port=str(c["port"]), user=c["user"],
            password=c["password"], database=c["database"], autocommit=autocommit,
        )
    raise SystemExit(f"✗ 不支援的 driver '{c['driver']}'(target={target})")


def parse_sql_file(path: pathlib.Path) -> tuple[str, str | None, str, str]:
    """切出 (DDL 區塊, COPY 語句, COPY 資料, post-COPY SQL)。無 COPY 時回 (全文, None, "", "")。"""
    lines = path.read_text(encoding="utf-8").splitlines()
    copy_idx = next((i for i, l in enumerate(lines) if l.startswith("COPY ")), None)
    if copy_idx is None:
        return "\n".join(lines), None, "", ""
    ddl = "\n".join(lines[:copy_idx])
    copy_stmt = lines[copy_idx].rstrip().rstrip(";")
    data_lines = lines[copy_idx + 1:]
    end = data_lines.index("\\.") if "\\." in data_lines else len(data_lines)
    data = "\n".join(data_lines[:end])
    post = "\n".join(data_lines[end + 1:]) if end < len(data_lines) else ""
    return ddl, copy_stmt, (data + "\n" if data else ""), post


def run_sql_file(conn, path: pathlib.Path) -> None:
    """在既有連線/交易內執行一個表檔(DDL + COPY + post-COPY SQL)。COPY 僅 PostgreSQL。"""
    ddl, copy_stmt, data, post = parse_sql_file(path)
    with conn.cursor() as cur:
        if ddl.strip():
            cur.execute(ddl)
        if copy_stmt is not None:
            if not hasattr(cur, "copy"):
                raise SystemExit("✗ COPY 匯入僅支援 PostgreSQL target(MSSQL 的 cursor 無 copy())")
            with cur.copy(copy_stmt) as cp:
                if data:
                    cp.write(data)
        if post.strip():
            cur.execute(post)
