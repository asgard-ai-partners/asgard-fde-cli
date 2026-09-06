"""salesforce.py -- Salesforce(SOQL)的傳輸層。

跟 NetSuite 一樣不是資料庫:OAuth 2.0 client credentials 換 token,再打 REST 的
query 端點。走標準函式庫,不需要額外套件。

**client credentials flow 要在 Connected App 上明確啟用**,而且要指定一個 run-as
使用者;沒啟用時 token 端點回 400 `unsupported_grant_type`,而訊息不會說是哪裡沒開。
"""
from __future__ import annotations

import json
import re
import urllib.error
import urllib.parse
import urllib.request

API_VERSION = "v60.0"


def _normalize_host(raw: str) -> str:
    return re.sub(r"^https?://", "", (raw or "").strip(), flags=re.I).rstrip("/")


def _get(url: str, token: str, timeout: int = 180) -> tuple[int, str]:
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", "replace")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", "replace")


def get_token(cfg: dict[str, str]) -> tuple[str, str]:
    """換 token。回傳 (access_token, instance_url) —— 之後的查詢要打 instance_url。"""
    host = _normalize_host(cfg["host"])
    data = urllib.parse.urlencode({
        "grant_type": "client_credentials",
        "client_id": cfg["consumer_key"],
        "client_secret": cfg["consumer_secret"],
    }).encode()
    req = urllib.request.Request(
        f"https://{host}/services/oauth2/token", data=data, method="POST",
        headers={"Content-Type": "application/x-www-form-urlencoded"})
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            body = json.loads(resp.read().decode("utf-8", "replace"))
    except urllib.error.HTTPError as e:
        raise SystemExit(
            f"x 取得 access token 失敗 [{e.code}]:{e.read().decode('utf-8', 'replace')[:600]}\n"
            "  常見原因:Connected App 沒有啟用 client credentials flow 或沒指定 run-as "
            "使用者、consumer key/secret 不成對、host 打錯(sandbox 是 *.sandbox.my.salesforce.com)。")
    token = body.get("access_token")
    if not token:
        raise SystemExit(f"x token 回應缺少 access_token:{json.dumps(body)[:300]}")
    return token, _normalize_host(body.get("instance_url") or host)


def describe(cfg: dict[str, str], sobject: str) -> tuple[list[str], list[list[str]]]:
    """列出一個 sObject 的欄位。SOQL 沒有 information_schema,describe 端點是替代品。"""
    token, instance = get_token(cfg)
    status, body = _get(
        f"https://{instance}/services/data/{API_VERSION}/sobjects/{sobject}/describe", token)
    if status != 200:
        raise SystemExit(f"x describe {sobject} 失敗 [{status}]:{body[:1000]}")
    fields = json.loads(body).get("fields", [])
    cols = ["name", "type", "label", "nillable"]
    return cols, [[str(f.get(c, "")) for c in cols] for f in fields]


def runner(cfg: dict[str, str]):
    token, instance = get_token(cfg)

    def run(soql: str, limit: int) -> tuple[list[str], list[list[str]]]:
        qs = urllib.parse.urlencode({"q": soql.strip()})
        status, body = _get(f"https://{instance}/services/data/{API_VERSION}/query?{qs}", token)
        if status != 200:
            raise SystemExit(f"x SOQL 查詢失敗 [{status}]:{body[:2000]}")
        records = json.loads(body).get("records", [])
        if limit > 0:
            records = records[:limit]
        cols: list[str] = []
        for r in records:
            for k in r:
                # attributes 是每列都附的型別/URL 中繼資料,對查詢沒有意義。
                if k != "attributes" and k not in cols:
                    cols.append(k)
        rows = [["" if r.get(c) is None else str(r.get(c)) for c in cols] for r in records]
        return cols, rows

    return run
