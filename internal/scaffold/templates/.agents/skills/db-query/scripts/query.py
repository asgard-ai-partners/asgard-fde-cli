#!/usr/bin/env python3
"""query.py -- 對客戶的來源系統跑唯讀查詢並印出結果。

一支 CLI 蓋住八個 DataConnector class。連線設定放在 repo 根的 .env,每一組用自己
的前綴;前綴由你決定,因為一個 engagement 可能同時接兩個 PostgreSQL。

    # 這個 class 需要哪些鍵(把輸出接到 .env,再請使用者填值)
    python query.py --class postgres --prefix UOF_DB_ --keys

    # 查詢
    python query.py --class postgres --prefix UOF_DB_ "select 1"
    python query.py --class mssql    --prefix ERP_DB_ -f some.sql
    echo "select 1" | python query.py --class postgres --prefix UOF_DB_

    # 這張表有哪些欄位(建 SemanticLayer 的第一步)
    python query.py --class postgres --prefix UOF_DB_ --columns sales.orders

連線摘要(不含任何機密)印到 stderr,結果表格印到 stdout,所以可以直接接管線。

**唯讀。** 這些是客戶的正式業務系統,任何會改動資料的語句都會被擋下來。
"""
from __future__ import annotations

import argparse
import sys

import connectors  # 同目錄

DEFAULT_LIMIT = 200


def render(cols: list[str], rows: list[list[str]], width: int = 48) -> str:
    if not cols:
        return "(no result set)"
    widths = [len(c) for c in cols]
    for r in rows:
        widths = [max(w, len(v)) for w, v in zip(widths, r)]
    widths = [min(w, width) for w in widths]
    # rstrip:結果常被貼進 docs/,行尾空白在 diff 裡是噪音。
    line = lambda cells: " | ".join(str(c)[:w].ljust(w) for c, w in zip(cells, widths)).rstrip()
    out = [line(cols), "-+-".join("-" * w for w in widths)]
    out += [line(r) for r in rows]
    out.append(f"({len(rows)} row{'s' if len(rows) != 1 else ''})")
    return "\n".join(out)


def suggest_prefixes(cls: str) -> list[str]:
    """.env 裡看起來屬於這個 class 的前綴。用來把「忘了給 --prefix」變成一個提示。"""
    spec = connectors.SPECS[cls]
    anchors = [f.suffix for f in spec.fields if f.required]
    found: list[str] = []
    for key in connectors.load_env():
        for a in anchors:
            if key.endswith("_" + a) or key == a:
                p = key[: len(key) - len(a)]
                if p and p not in found:
                    found.append(p)
    return sorted(found)


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(
        description="Run a read-only query against a customer source system",
        formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("sql", nargs="?", help="SQL/SOQL(省略則從 -f 或 stdin 讀)")
    ap.add_argument("-c", "--class", dest="cls", metavar="CLASS",
                    help="DataConnector class:" + ", ".join(connectors.CLASSES))
    ap.add_argument("-p", "--prefix", help="這組連線在 .env 的鍵前綴,例如 UOF_DB_")
    ap.add_argument("-f", "--file", help="從檔案讀查詢")
    ap.add_argument("--keys", action="store_true", help="印出這個 class 需要的 .env 鍵,不連線")
    ap.add_argument("--columns", metavar="TABLE", help="列出一張表的欄位(可寫 schema.table)")
    ap.add_argument("--limit", type=int, default=DEFAULT_LIMIT,
                    help=f"最多取幾列(預設 {DEFAULT_LIMIT};0 表示不限)")
    ap.add_argument("--classes", action="store_true", help="列出支援的 class 後結束")
    args = ap.parse_args(argv)

    if args.classes:
        for name in connectors.CLASSES:
            spec = connectors.SPECS[name]
            need = " ".join(f.suffix for f in spec.fields if f.required)
            print(f"{name:<11} {need}")
        return 0

    if not args.cls:
        ap.error("--class 是必要的;`--classes` 列出可用的值")
    if args.cls not in connectors.SPECS:
        ap.error(f"未知的 class '{args.cls}';可用:{', '.join(connectors.CLASSES)}")
    if not args.prefix:
        hint = suggest_prefixes(args.cls)
        ap.error("--prefix 是必要的(一個 repo 可以接不只一個同 class 的系統)。"
                 + (f"\n.env 裡看起來像的前綴:{', '.join(hint)}" if hint
                    else f"\n.env 裡還沒有 {args.cls} 的鍵;先跑 --keys 產生它們。"))

    if args.keys:
        print(connectors.env_block(args.cls, args.prefix))
        return 0

    sample = False
    if args.columns:
        if args.cls == "salesforce":
            import salesforce  # 同目錄
            cfg = connectors.config(args.cls, args.prefix)
            print(f"=== {connectors.describe(args.cls, args.prefix)} ===", file=sys.stderr)
            cols, rows = salesforce.describe(cfg, args.columns)
            print(render(cols, rows))
            return 0
        sql, sample = connectors.columns_query(args.cls, args.columns)
    elif args.file:
        sql = open(args.file, encoding="utf-8").read()
    elif args.sql:
        sql = args.sql
    else:
        sql = sys.stdin.read()

    if not sql.strip():
        print("no query provided", file=sys.stderr)
        return 1
    connectors.assert_read_only(sql)

    print(f"=== {connectors.describe(args.cls, args.prefix)} ===", file=sys.stderr)
    run = connectors.runner(args.cls, args.prefix)
    cols, rows = run(sql, args.limit)
    if sample:
        # 取樣模式:欄名才是答案。NULL 欄位在回應裡整個 key 會被省略,所以一列
        # 取樣會少報 —— 挑一列真的有值的再問一次。
        cols, rows = ["column"], [[c] for c in cols]
    print(render(cols, rows))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
