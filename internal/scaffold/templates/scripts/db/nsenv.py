"""nsenv.py — NetSuite(SuiteQL)工具共用層:讀 .env、換 token、下 SuiteQL 查詢。

`pgenv.py` 的姊妹檔。四個 DB target 走 DBAPI(psycopg / pymssql),NetSuite 沒有 DBAPI ——
它是 REST 服務,所以自成一層:OAuth 2.0 client credentials + JWT client assertion 換 token,
再 POST 到 SuiteQL 端點。

    .venv/bin/python scripts/db/nsquery.py "SELECT 1 AS ok FROM DUAL"

認證流程與另外兩個實作逐字對應,三邊必須一致:
  - Go:  asgard-loader/pkg/suiteql/client.go → ensureAccessToken
  - JS:  a reference NetSuite token implementation → signClientAssertion / getAccessToken

三個踩過的坑,改這個檔前先讀:

1. **scope 必須是 `rest_webservices`。** SuiteQL 與 REST Web Services 用這個;
   RESTlet 用的是 `restlets`。給錯會拿到 401 INVALID_LOGIN_ATTEMPT,而錯誤訊息不會告訴你
   是 scope 的問題(見 nsToken.js 的註解,那是實測出來的)。
2. **ES256 的簽章必須是 IEEE-P1363 的原始 R||S**,不是 DER。PyJWT 的 ES256 預設就是
   P1363(JWS 規範要求),所以這裡不用特別處理;但若改用 `cryptography` 手刻就要注意
   (JS 版要顯式指定 `dsaEncoding: "ieee-p1363"`)。
3. **私鑰在 .env 裡是單行、用字面 `\\n` 代表換行。** 必須還原成真正的換行才解析得動,
   見 `_normalize_pem`(對應 JS 的 `normalizePem` 與 Go 的 `loadNSPrivateKey`)。

> **唯讀。** SuiteQL 本身只是查詢服務,但這裡仍然明確擋掉非 SELECT/WITH 的語句 ——
> 與 `semantic-layer-modeling` skill 的唯讀規則一致,而且 NetSuite 是正式帳。
"""
from __future__ import annotations

import json
import os
import pathlib
import re
import time
import urllib.error
import urllib.parse
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[2]  # repo root
ENV_FILE = ROOT / ".env"

ENV_PREFIX = "NS_"
TOKEN_PATH = "services/rest/auth/oauth2/v1/token"
QUERY_PATH = "services/rest/query/v1/suiteql"
# SuiteQL / REST Web Services 用這個 scope。RESTlet 才是 "restlets"(見檔頭坑 1)。
SCOPE = "rest_webservices"
CLIENT_ASSERTION_TYPE = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
SUPPORTED_ALGORITHMS = ("ES256", "ES512", "PS256")

# NetSuite 單次回應的硬上限。要更多就翻頁(見 query_all)。
MAX_PAGE_LIMIT = 1000

_REQUIRED = ("HOST", "CONSUMER_KEY", "CERTIFICATE_ID", "PRIVATE_KEY_PEM")


def _load_env() -> dict[str, str]:
    """讀專案根 .env 為 dict。環境變數優先(同名覆蓋),與 pgenv 的行為一致。"""
    env: dict[str, str] = {}
    if ENV_FILE.is_file():
        for raw in ENV_FILE.read_text(encoding="utf-8").splitlines():
            line = raw.strip()
            if line.startswith("#") or "=" not in line:
                continue
            key, val = line.split("=", 1)
            env[key.strip()] = val.strip().strip('"').strip("'")
    for k, v in os.environ.items():
        if k.startswith(ENV_PREFIX):
            env[k] = v
    return env


def _normalize_host(raw: str) -> str:
    """host 一律裸 host(不含 scheme)。防呆:去掉誤貼的 scheme 與尾斜線。"""
    return re.sub(r"^https?://", "", (raw or "").strip(), flags=re.I).rstrip("/")


def _normalize_pem(raw: str) -> str:
    """把 .env 單行 PEM 的字面 `\\n` / `\\r\\n` 還原成真正換行,並統一成 LF。

    本來就是多行 PEM 的話,這些替換都是 no-op。
    """
    return (
        (raw or "")
        .replace("\\r\\n", "\n")
        .replace("\\n", "\n")
        .replace("\\r", "")
        .replace("\r", "")
        .strip()
    )


def load_config() -> dict[str, str]:
    """組出 NetSuite 連線設定。缺變數就明確報錯,不要靜默連到錯的帳號。"""
    env = _load_env()
    missing = [ENV_PREFIX + k for k in _REQUIRED if not env.get(ENV_PREFIX + k, "").strip()]
    if missing:
        raise SystemExit(
            f"✗ NetSuite 缺少 {', '.join(missing)}(填 {ENV_FILE} 或環境變數;參考 .env.example)"
        )
    alg = env.get(ENV_PREFIX + "SIGNATURE_ALGORITHM", "").strip() or "ES256"
    if alg not in SUPPORTED_ALGORITHMS:
        raise SystemExit(
            f"✗ 不支援的 {ENV_PREFIX}SIGNATURE_ALGORITHM '{alg}'"
            f"(僅支援 {', '.join(SUPPORTED_ALGORITHMS)})"
        )
    return {
        "host": _normalize_host(env[ENV_PREFIX + "HOST"]),
        "consumer_key": env[ENV_PREFIX + "CONSUMER_KEY"].strip(),
        "certificate_id": env[ENV_PREFIX + "CERTIFICATE_ID"].strip(),
        "private_key_pem": _normalize_pem(env[ENV_PREFIX + "PRIVATE_KEY_PEM"]),
        "signature_algorithm": alg,
    }


def describe() -> str:
    """給人看的連線摘要(不含任何密鑰),供 CLI 顯示 —— 免得打錯帳號還不自知。"""
    c = load_config()
    return (
        f"netsuite [suiteql] {c['host']} "
        f"cert={c['certificate_id'][:8]}… alg={c['signature_algorithm']}"
    )


def _post(url: str, *, data: bytes, headers: dict[str, str], timeout: int = 180) -> tuple[int, str]:
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", "replace")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", "replace")


def get_token(cfg: dict[str, str] | None = None) -> str:
    """簽 JWT client assertion 並換 access token。"""
    try:
        import jwt  # lazy:只有 NetSuite 需要(pyjwt[crypto])
    except ModuleNotFoundError:
        raise SystemExit(
            "✗ 缺少 pyjwt[crypto];安裝:.venv/bin/pip install -r scripts/db/requirements.txt"
        )
    c = cfg or load_config()
    aud = f"https://{c['host']}/{TOKEN_PATH}"
    now = int(time.time())
    try:
        assertion = jwt.encode(
            {"iss": c["consumer_key"], "scope": SCOPE, "aud": aud, "iat": now, "exp": now + 3600},
            c["private_key_pem"],
            algorithm=c["signature_algorithm"],
            headers={"kid": c["certificate_id"]},
        )
    except Exception as e:  # 私鑰格式或演算法不合
        raise SystemExit(
            f"✗ JWT 簽章失敗(請確認 {ENV_PREFIX}PRIVATE_KEY_PEM 是 PKCS#8 PEM 且與 "
            f"{c['signature_algorithm']} 相符):{type(e).__name__}: {e}"
        )
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
            f"✗ 取得 access token 失敗 [{status}]:{body[:600]}\n"
            f"  常見原因:憑證已失效/整合被停用(500 server_error)、"
            f"{ENV_PREFIX}CERTIFICATE_ID 與私鑰不成對、帳號 host 打錯。"
        )
    token = json.loads(body).get("access_token")
    if not token:
        raise SystemExit(f"✗ token 回應缺少 access_token:{body[:300]}")
    return token


_READ_ONLY_RE = re.compile(r"^\s*(?:--[^\n]*\n|/\*.*?\*/|\s)*(select|with)\b", re.I | re.S)


def assert_read_only(sql: str) -> None:
    """擋掉非 SELECT/WITH 的語句。NetSuite 是正式帳,唯讀是本 repo 的架構前提。"""
    if not _READ_ONLY_RE.match(sql):
        raise SystemExit(
            "✗ 只允許 SELECT / WITH 查詢(本 repo 對來源系統一律唯讀)。\n"
            "  收到的語句開頭:" + " ".join(sql.split())[:80]
        )


def query(sql: str, *, limit: int = 200, offset: int = 0,
          token: str | None = None, cfg: dict[str, str] | None = None) -> dict:
    """跑一次 SuiteQL,回傳 NetSuite 的原始 JSON(items / hasMore / totalResults …)。"""
    assert_read_only(sql)
    c = cfg or load_config()
    tok = token or get_token(c)
    if limit > MAX_PAGE_LIMIT:
        raise SystemExit(f"✗ limit 上限為 {MAX_PAGE_LIMIT}(要更多請翻頁,見 --all)")
    qs = urllib.parse.urlencode({"limit": limit, "offset": offset})
    status, body = _post(
        f"https://{c['host']}/{QUERY_PATH}?{qs}",
        data=json.dumps({"q": sql.strip()}).encode(),
        headers={
            "Content-Type": "application/json",
            # SuiteQL 要求 Prefer: transient,少了它會被拒。
            "Prefer": "transient",
            "Authorization": f"Bearer {tok}",
        },
    )
    if status != 200:
        raise SystemExit(f"✗ SuiteQL 查詢失敗 [{status}]:{body[:2000]}")
    return json.loads(body)


def query_all(sql: str, *, page: int = MAX_PAGE_LIMIT, max_rows: int = 10_000,
              cfg: dict[str, str] | None = None) -> list[dict]:
    """翻頁把結果撈完(上限 max_rows,避免手滑把整張 transaction 拉下來)。"""
    c = cfg or load_config()
    tok = get_token(c)
    rows: list[dict] = []
    offset = 0
    while True:
        payload = query(sql, limit=page, offset=offset, token=tok, cfg=c)
        rows.extend(payload.get("items", []))
        if not payload.get("hasMore") or len(rows) >= max_rows:
            return rows[:max_rows]
        offset += page
