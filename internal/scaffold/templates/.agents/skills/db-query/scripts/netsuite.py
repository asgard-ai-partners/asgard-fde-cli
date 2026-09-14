"""netsuite.py -- NetSuite(SuiteQL)的傳輸層。

多數 class 走 DBAPI。NetSuite 不是資料庫,是 REST 服務,所以自成一層:
OAuth 2.0 client credentials + JWT client assertion 換 token,再 POST 到 SuiteQL
端點。connectors.py 已經把 .env 讀好,這裡只負責「怎麼問 NetSuite」。

三個踩過的坑,改這個檔前先讀:

1. **scope 必須是 `rest_webservices`。** SuiteQL 與 REST Web Services 用這個;
   RESTlet 用的是 `restlets`。給錯會拿到 401 INVALID_LOGIN_ATTEMPT,而錯誤訊息
   不會告訴你是 scope 的問題 —— 這是實測出來的。
2. **ES256 的簽章必須是 IEEE-P1363 的原始 R||S**,不是 DER。PyJWT 的 ES256 預設
   就是 P1363(JWS 規範要求),所以這裡不用特別處理;若改用 `cryptography` 手刻
   就要注意。
3. **私鑰在 .env 裡是單行、用字面 `\\n` 代表換行。** 必須還原成真正的換行才解析
   得動,見 `_normalize_pem`。
"""
from __future__ import annotations

import json
import re
import time
import urllib.error
import urllib.parse
import urllib.request

TOKEN_PATH = "services/rest/auth/oauth2/v1/token"
QUERY_PATH = "services/rest/query/v1/suiteql"
# SuiteQL / REST Web Services 用這個 scope。RESTlet 才是 "restlets"(見檔頭坑 1)。
SCOPE = "rest_webservices"
CLIENT_ASSERTION_TYPE = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

# NetSuite 單次回應的硬上限。要更多就翻頁。
MAX_PAGE_LIMIT = 1000


def _normalize_host(raw: str) -> str:
    """host 一律裸 host。防呆:去掉誤貼的 scheme 與尾斜線。"""
    return re.sub(r"^https?://", "", (raw or "").strip(), flags=re.I).rstrip("/")


def _normalize_pem(raw: str) -> str:
    """把 .env 單行 PEM 的字面 `\\n` / `\\r\\n` 還原成真正換行。本來就多行的話是 no-op。"""
    return ((raw or "")
            .replace("\\r\\n", "\n").replace("\\n", "\n")
            .replace("\\r", "").replace("\r", "").strip())


def _post(url: str, *, data: bytes, headers: dict[str, str], timeout: int = 180) -> tuple[int, str]:
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", "replace")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", "replace")


def get_token(cfg: dict[str, str]) -> str:
    """簽 JWT client assertion 並換 access token。"""
    try:
        import jwt  # pyjwt[crypto]
    except ModuleNotFoundError:
        raise SystemExit("x netsuite 需要 pyjwt[crypto];安裝:.venv/bin/pip install 'pyjwt[crypto]>=2.8'")

    host = _normalize_host(cfg["host"])
    aud = f"https://{host}/{TOKEN_PATH}"
    alg = cfg.get("signature_algorithm") or "ES256"
    now = int(time.time())
    try:
        assertion = jwt.encode(
            {"iss": cfg["consumer_key"], "scope": SCOPE, "aud": aud, "iat": now, "exp": now + 3600},
            _normalize_pem(cfg["private_key_pem"]),
            algorithm=alg,
            headers={"kid": cfg["certificate_id"]},
        )
    except Exception as e:  # 私鑰格式或演算法不合
        raise SystemExit(
            f"x JWT 簽章失敗(確認 PRIVATE_KEY_PEM 是 PKCS#8 PEM 且與 {alg} 相符):"
            f"{type(e).__name__}: {e}")

    status, body = _post(
        aud,
        data=urllib.parse.urlencode({
            "grant_type": "client_credentials",
            "client_assertion_type": CLIENT_ASSERTION_TYPE,
            "client_assertion": assertion,
        }).encode(),
        headers={"Content-Type": "application/x-www-form-urlencoded"},
        timeout=60,
    )
    if status != 200:
        raise SystemExit(
            f"x 取得 access token 失敗 [{status}]:{body[:600]}\n"
            "  常見原因:憑證已失效或整合被停用(500 server_error)、"
            "CERTIFICATE_ID 與私鑰不成對、host 打錯。")
    token = json.loads(body).get("access_token")
    if not token:
        raise SystemExit(f"x token 回應缺少 access_token:{body[:300]}")
    return token


def columns_sql(table: str) -> str:
    """SuiteQL 沒有 information_schema,只能抓一列反推欄位名。

    注意:值為 NULL 的欄位整個 key 會被省略,所以單列取樣會少報。要看某些特定
    欄位時,取一列真的有值的(例如 `WHERE mainline = 'F'`)。
    """
    return f"SELECT * FROM {table} WHERE ROWNUM = 1"


def runner(cfg: dict[str, str]):
    """回傳 (sql, limit) -> (欄名, 列) 的查詢函式。"""
    host = _normalize_host(cfg["host"])
    token = get_token(cfg)

    def run(sql: str, limit: int) -> tuple[list[str], list[list[str]]]:
        page = min(limit, MAX_PAGE_LIMIT) if limit > 0 else MAX_PAGE_LIMIT
        qs = urllib.parse.urlencode({"limit": page, "offset": 0})
        status, body = _post(
            f"https://{host}/{QUERY_PATH}?{qs}",
            data=json.dumps({"q": sql.strip()}).encode(),
            headers={
                "Content-Type": "application/json",
                # SuiteQL 要求 Prefer: transient,少了它會被拒。
                "Prefer": "transient",
                "Authorization": f"Bearer {token}",
            },
        )
        if status != 200:
            # 不存在的欄位會回一個沒有訊息的 500 UNEXPECTED_ERROR,GROUP BY 不合它
            # 意時也是。500 的時候逐欄二分找出是哪一欄。
            raise SystemExit(f"x SuiteQL 查詢失敗 [{status}]:{body[:2000]}")
        items = json.loads(body).get("items", [])
        cols: list[str] = []
        for it in items:
            for k in it:
                # links 是 NetSuite 給每列附的 HATEOAS 欄位,對查詢沒有意義。
                if k != "links" and k not in cols:
                    cols.append(k)
        rows = [["" if it.get(c) is None else str(it.get(c)) for c in cols] for it in items]
        return cols, rows

    return run
