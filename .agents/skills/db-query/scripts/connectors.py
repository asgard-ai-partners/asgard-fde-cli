"""connectors.py -- db-query 的共用層:class 定義、讀 .env、建立連線、跑查詢。

**這支工具的工作是「連上一個資料庫、看清楚它」,不是「跟 DataConnector CRD 保持一致」。**
它中立地為 local 連 DB 這件事服務,而每個 class 的欄位在這裡是一筆 SPECS。

.env key 的字尾**現在**跟 CRD 欄位名的大寫底線寫法一樣:

    DataConnectorSpec.postgres.sslMode   ->   <PREFIX>SSL_MODE
    DataConnectorSpec.athena.accessKeyId ->   <PREFIX>ACCESS_KEY_ID

**那是一個方便,不是一個契約。** 好處是填 .env 跟填 CR 少一次翻譯;而 CRD 之後長出
新欄位、或這裡為了連得上而多要一個東西,**兩邊分岔是可以的,沒有東西在比對它們**。
需要知道平台實際會收什麼的時候,權威是 `.agents/skills/asgard-cr-shapes/`(那份是從
活著的叢集讀出來的,`asgard-cli skill update` 抓),不是這裡。

**前綴由使用中的 agent 自己取,不是這裡規定的。** 一個 engagement 可能同時接兩個
PostgreSQL,它們在 .env 裡本來就是兩群 key-value,不可能共用名字。所以每次查詢都要
指名 --prefix,而這支工具唯一的約定是「同一群用同一個前綴」。

    python query.py --class postgres --prefix UOF_DB_ "select 1"
    python query.py --class postgres --prefix ERP_DB_ "select 1"

環境變數優先於 .env(同名覆蓋),所以臨時要換一個值不必動檔案。
"""
from __future__ import annotations

import os
import pathlib
import re
from typing import Callable, NamedTuple


class Field(NamedTuple):
    """一個連線欄位。suffix 是 .env key 的字尾,也是 CRD 欄位名。"""

    suffix: str          # <PREFIX> 後面接的部分。現在跟 CRD 欄位同名,見檔頭
    required: bool = True
    secret: bool = False  # 真值不進版控、不印在 transcript
    default: str = ""
    note: str = ""


class Spec(NamedTuple):
    """一個 DataConnector class 在 design time 需要什麼。"""

    fields: tuple[Field, ...]
    connect: str         # 這個 class 走哪條路:見 query.py 的分派
    package: str         # 缺 driver 時要裝什麼
    exactly_one: tuple[str, ...] = ()   # CRD 的 ExactlyOneOf,兩個都填是錯的
    at_least_one: tuple[str, ...] = ()  # 認證方式擇一,一個都沒有連不上


_HOST = Field("HOST", note="裸 host,不含 scheme")
_USER = Field("USER")
_PASSWORD = Field("PASSWORD", secret=True)
_DATABASE = Field("DATABASE")


def _port(default: str) -> Field:
    return Field("PORT", required=False, default=default,
                 note=f"預設 {default};CR 的 spec.*.port 仍必須明寫")


# 對應 asgard-kube DataConnectorClass 的九個值。hana 沒有 design-time 解方
# (見 references/connectors.md 的說明),其餘八個都在這裡。
SPECS: dict[str, Spec] = {
    "postgres": Spec(
        fields=(_HOST, _port("5432"), _USER, _PASSWORD, _DATABASE,
                Field("SSL_MODE", required=False,
                      note="CRD 只收 disable|require|verify-ca|verify-full;"
                           "留空則用 libpq 自己的預設")),
        connect="dbapi", package="psycopg[binary]>=3",
    ),
    "mysql": Spec(
        fields=(_HOST, _port("3306"), _USER, _PASSWORD, _DATABASE),
        connect="dbapi", package="PyMySQL>=1.1",
    ),
    "mssql": Spec(
        fields=(_HOST, _port("1433"), _USER, _PASSWORD, _DATABASE,
                Field("INSTANCE", required=False, note="具名執行個體,多數部署留空")),
        connect="dbapi", package="pymssql>=2.3",
    ),
    "oracle": Spec(
        fields=(_HOST, _port("1521"), _USER, _PASSWORD,
                Field("SERVICE_NAME", required=False),
                Field("SID", required=False)),
        connect="dbapi", package="oracledb>=2",
        exactly_one=("SERVICE_NAME", "SID"),
    ),
    "salesforce": Spec(
        fields=(_HOST, Field("CONSUMER_KEY", secret=True),
                Field("CONSUMER_SECRET", secret=True)),
        connect="salesforce", package="(無;走標準函式庫的 REST)",
    ),
    "netsuite": Spec(
        fields=(_HOST, Field("CONSUMER_KEY", secret=True),
                Field("SIGNATURE_ALGORITHM", required=False, default="ES256",
                      note="ES256|ES512|PS256"),
                Field("CERTIFICATE_ID"),
                Field("PRIVATE_KEY_PEM", secret=True,
                      note="PKCS#8 PEM;可單行、換行寫成字面 \\n")),
        connect="netsuite", package="pyjwt[crypto]>=2.8",
    ),
    "trino": Spec(
        fields=(Field("SCHEME", required=False, default="https", note="http|https"),
                _HOST, _port("443"), _USER,
                Field("PASSWORD", required=False, secret=True),
                Field("JWT_ACCESS_TOKEN", required=False, secret=True),
                Field("SSL_CERTIFICATE_PEM", required=False, secret=True)),
        connect="dbapi", package="trino>=0.330",
    ),
    "athena": Spec(
        fields=(Field("REGION", note="例:ap-northeast-1"),
                Field("ACCESS_KEY_ID", secret=True),
                Field("SECRET_ACCESS_KEY", secret=True),
                Field("OUTPUT_LOCATION", note="s3:// 開頭,Athena 寫結果的位置"),
                Field("WORK_GROUP", required=False)),
        connect="dbapi", package="PyAthena>=3",
    ),
}

CLASSES = tuple(SPECS)


def repo_root() -> pathlib.Path:
    """往上找到 repo 根(有 .asgard-pipeline.yaml 或 .git 的那層)。

    不用 parents[N] 數層數:這支檔案搬過一次目錄,而數字不會在搬的時候報錯,
    只會安靜地讀錯一個 .env。
    """
    here = pathlib.Path(__file__).resolve()
    for d in here.parents:
        if (d / ".asgard-pipeline.yaml").is_file() or (d / ".git").exists():
            return d
    return here.parents[4] if len(here.parents) > 4 else here.parent


def env_file() -> pathlib.Path:
    return repo_root() / ".env"


def rel(path: pathlib.Path) -> str:
    """訊息裡一律印 repo 相對路徑。

    絕對路徑在 transcript 裡會洩漏本機目錄結構,而且讀的人要的是「在這個 repo
    的哪裡」,不是「在你的 home 底下哪裡」。
    """
    try:
        return str(path.resolve().relative_to(repo_root()))
    except ValueError:
        return str(path)


def parse_value(raw: str) -> str:
    """把 `.env` 一行的等號右邊解成值。

    **行內註解要拿掉。** 少了這一步,`KEY=   # 說明` 會被讀成值 `# 說明`,於是一個
    還沒填的鍵看起來像填好了 —— 接著就是拿註解去當 host 連線,或把它當成 sslMode
    餵給 driver。

    規則跟其他 dotenv 讀取器一致,兩條:
      - `#` 只有在**前面有空白**(或整行開頭)時才開始註解。所以 `abc#123` 是一個
        完整的密碼,不是 `abc`。
      - 引號內的 `#` 一律是值的一部分。值本身要保留前後空白時,就用引號包起來。
        單引號裡一切照字面;雙引號裡認 `\\\\` `\\"` `\\n` `\\r` `\\t`。

    **Go 那邊有第二份實作**(`internal/localenv` 的 `ParseValue`),因為 local-env
    的 UI 讀寫同一個檔。兩邊照同一段描述寫,改一邊就要改另一邊。
    """
    raw = raw.strip()
    if not raw:
        return ""
    if raw[0] == "'":
        end = raw.find("'", 1)
        return raw[1:end] if end != -1 else raw[1:]
    if raw[0] == '"':
        # 雙引號裡認 \\ \" \n \r \t —— 這是 PEM 能寫在一行裡的原因。
        out, i, body = [], 0, raw[1:]
        while i < len(body):
            ch = body[i]
            if ch == '"':
                break
            if ch == "\\" and i + 1 < len(body):
                nxt = body[i + 1]
                # 認不得的跳脫兩個字元都留著:值裡的 Windows 路徑不是筆誤,
                # 安靜吃掉它的反斜線才是。
                out.append({"n": "\n", "r": "\r", "t": "\t",
                            "\\": "\\", '"': '"'}.get(nxt, "\\" + nxt))
                i += 2
                continue
            out.append(ch)
            i += 1
        return "".join(out)
    cut = re.search(r"(?:^|\s)#", raw)
    if cut:
        raw = raw[: cut.start()]
    return raw.strip()


def load_env() -> dict[str, str]:
    """讀 repo 根的 .env,再讓真正的環境變數覆蓋同名的鍵。"""
    env: dict[str, str] = {}
    path = env_file()
    if path.is_file():
        for raw in path.read_text(encoding="utf-8").splitlines():
            line = raw.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, val = line.split("=", 1)
            env[key.strip()] = parse_value(val)
    env.update(os.environ)
    return env


def keys(cls: str, prefix: str) -> list[tuple[str, Field]]:
    """這個 class 在這個前綴下的完整鍵名清單,依宣告順序。"""
    return [(prefix + f.suffix, f) for f in SPECS[cls].fields]


def env_block(cls: str, prefix: str) -> str:
    """印一段可以直接接到 .env 後面的佔位鍵。值一律留空,由使用者填。

    說明寫在鍵的**上一行**,不是等號後面。等號後面的註解會被當成值 —— 不只是被
    這支工具,任何 dotenv 讀取器都可能 —— 而一個「填了註解」的鍵看起來就是填好了。
    """
    lines = [f"# --- {prefix.rstrip('_')} ({cls}) ---"]
    for name, f in keys(cls, prefix):
        tags = []
        if not f.required:
            tags.append("選填")
        if f.secret:
            tags.append("機密")
        if f.note:
            tags.append(f.note)
        if tags:
            lines.append("# " + ";".join(tags))
        lines.append(f"{name}=")
    # 結尾空一行:`--keys >> .env` 連續接兩組的時候,兩組之間才不會黏在一起。
    return "\n".join(lines) + "\n"


def config(cls: str, prefix: str) -> dict[str, str]:
    """組出連線設定,鍵是小寫的 CRD 欄位名。缺什麼就明講缺哪個環境變數。"""
    if cls not in SPECS:
        raise SystemExit(f"x 未知的 class '{cls}';可用:{', '.join(CLASSES)}")
    spec = SPECS[cls]
    env = load_env()
    cfg: dict[str, str] = {}
    missing: list[str] = []
    for name, f in keys(cls, prefix):
        # 不再 strip 一次:load_env 已經處理過行尾空白與引號,而**用引號包住**正是
        # 保留前後空白的方式。在這裡二次 strip 會把它又拿掉。
        val = env.get(name, "") or f.default
        if not val.strip() and f.required:
            missing.append(name)
        cfg[f.suffix.lower()] = val
    if missing:
        raise SystemExit(
            f"x {cls} 前綴 '{prefix}' 缺少 {', '.join(missing)}\n"
            f"  這些鍵要填在 {rel(env_file())};要看完整清單:\n"
            f"    python {rel(pathlib.Path(__file__).with_name('query.py'))} "
            f"--class {cls} --prefix {prefix} --keys"
        )
    for group, need_one in ((spec.exactly_one, True), (spec.at_least_one, False)):
        if not group:
            continue
        have = [g for g in group if cfg.get(g.lower())]
        names = ", ".join(prefix + g for g in group)
        if need_one and len(have) != 1:
            raise SystemExit(f"x {cls} 需要 {names} 剛好一個(收到 {len(have)} 個)")
        if not need_one and not have:
            raise SystemExit(f"x {cls} 需要 {names} 至少一個")
    return cfg


def describe(cls: str, prefix: str) -> str:
    """給人看的連線摘要。**不含任何機密欄位** —— 打錯帳號要看得出來,密碼不能外流。"""
    cfg = config(cls, prefix)
    if cls == "athena":
        return f"{prefix} [athena] {cfg['region']} -> {cfg['output_location']}"
    if cls == "netsuite":
        return (f"{prefix} [netsuite/suiteql] {cfg['host']} "
                f"cert={cfg['certificate_id'][:8]}... alg={cfg['signature_algorithm']}")
    if cls == "salesforce":
        return f"{prefix} [salesforce/soql] {cfg['host']}"
    who = f"{cfg.get('user', '')}@" if cfg.get("user") else ""
    where = cfg["host"] + (f":{cfg['port']}" if cfg.get("port") else "")
    tail = "/" + cfg["database"] if cfg.get("database") else ""
    return f"{prefix} [{cls}] {who}{where}{tail}"


# --- 唯讀閘門 -------------------------------------------------------------
#
# 這些是客戶的正式業務系統。db-query 存在的理由是「看清楚現況」,不是改動它。
# 兩層防護:開頭必須是讀取語句,而且整段語句裡不得出現寫入關鍵字 —— 只擋開頭
# 是不夠的,PostgreSQL 的 `WITH x AS (INSERT ... RETURNING *) SELECT ...` 開頭
# 就是 WITH,而它會寫進去。

_READ_HEAD = re.compile(r"^\s*(select|with|show|describe|desc)\b", re.I)
_WRITE_WORDS = re.compile(
    r"\b(insert|update|delete|merge|upsert|truncate|drop|create|alter|"
    r"grant|revoke|call|exec|execute|commit|rollback|set|copy|load)\b", re.I)
_REPLACE_INTO = re.compile(r"\breplace\s+into\b", re.I)


def _strip_noise(sql: str) -> str:
    """拿掉字串字面與註解,免得 'deleted' 這種資料值被誤判成寫入。"""
    sql = re.sub(r"/\*.*?\*/", " ", sql, flags=re.S)
    sql = re.sub(r"--[^\n]*", " ", sql)
    sql = re.sub(r"'(?:''|[^'])*'", "''", sql)
    sql = re.sub(r'"(?:""|[^"])*"', '""', sql)
    return sql


def assert_read_only(sql: str) -> None:
    """擋掉任何會改動來源系統的語句。"""
    bare = _strip_noise(sql)
    if not _READ_HEAD.match(bare):
        raise SystemExit(
            "x 只接受 SELECT / WITH / SHOW / DESCRIBE。db-query 對來源系統一律唯讀。\n"
            "  收到的開頭:" + " ".join(sql.split())[:80])
    hit = _WRITE_WORDS.search(bare) or _REPLACE_INTO.search(bare)
    if hit:
        raise SystemExit(
            f"x 語句裡出現 '{hit.group(0).upper()}',db-query 不執行任何會改動資料的查詢。\n"
            "  真的需要寫入時,由使用者自己用他們的工具跑,並記在 docs/decisions/。")


# --- 連線 -----------------------------------------------------------------
#
# 每個 class 回傳一個 run(sql, limit) -> (欄名, 列),讓 CLI 只需要一種輸出路徑。
# driver 一律 lazy import:只用到 postgres 的人不必裝 oracle 的套件。

Runner = Callable[[str, int], tuple[list[str], list[list[str]]]]

# 由 query.py 依 --traceback 設定。診斷訊息蓋掉的是「哪個憑證錯、哪一欄不存在」
# 那一層;真的要看 driver 的 stack 時,這個開關讓原例外原樣往上丟。
SHOW_TRACEBACK = False


def _need(cls: str, *modules: str):
    """載入 driver,缺了就講清楚要裝什麼。回傳第一個模組。"""
    import importlib
    loaded = []
    for m in modules:
        try:
            loaded.append(importlib.import_module(m))
        except ModuleNotFoundError:
            raise SystemExit(
                f"x {cls} 需要 {SPECS[cls].package},但這個 python 沒有。\n"
                f"  安裝:.venv/bin/pip install '{SPECS[cls].package}'\n"
                f"  或一次裝齊:.venv/bin/pip install -r "
                f"{rel(pathlib.Path(__file__).with_name('requirements.txt'))}")
    return loaded[0]


def _pem_file(pem: str) -> str:
    """把 .env 裡的 PEM 字串寫成暫存檔,回傳路徑。

    .env 是一行一個值,所以 PEM 多半寫成單行、換行是字面 `\\n` —— 跟 NetSuite
    的私鑰同一個處理。行程結束時刪掉。
    """
    import atexit
    import tempfile
    body = (pem.replace("\\r\\n", "\n").replace("\\n", "\n")
               .replace("\\r", "").replace("\r", "").strip()) + "\n"
    fd, path = tempfile.mkstemp(suffix=".pem")
    with os.fdopen(fd, "w", encoding="utf-8") as f:
        f.write(body)
    atexit.register(lambda: os.path.exists(path) and os.remove(path))
    return path


def _dbapi_connect(cls: str, cfg: dict[str, str]):
    if cls == "postgres":
        psycopg = _need(cls, "psycopg")
        kw = dict(host=cfg["host"], port=cfg["port"], user=cfg["user"],
                  password=cfg["password"], dbname=cfg["database"])
        if cfg.get("ssl_mode"):
            kw["sslmode"] = cfg["ssl_mode"]
        return psycopg.connect(**kw)

    if cls == "mysql":
        pymysql = _need(cls, "pymysql")
        return pymysql.connect(host=cfg["host"], port=int(cfg["port"]), user=cfg["user"],
                               password=cfg["password"], database=cfg["database"])

    if cls == "mssql":
        pymssql = _need(cls, "pymssql")
        # 具名執行個體走 server\instance,不是另一個埠。
        server = cfg["host"] + ("\\" + cfg["instance"] if cfg.get("instance") else "")
        return pymssql.connect(server=server, port=str(cfg["port"]), user=cfg["user"],
                               password=cfg["password"], database=cfg["database"])

    if cls == "oracle":
        # thin 模式:不需要 Oracle Instant Client,這是選 oracledb 而非 cx_Oracle 的理由。
        oracledb = _need(cls, "oracledb")
        dsn = oracledb.makedsn(cfg["host"], int(cfg["port"]),
                               service_name=cfg.get("service_name") or None,
                               sid=cfg.get("sid") or None)
        return oracledb.connect(user=cfg["user"], password=cfg["password"], dsn=dsn)

    if cls == "trino":
        # dbapi 與 auth 都要顯式載入,import trino 不會把它們帶進來。
        dbapi = _need(cls, "trino.dbapi", "trino.auth")
        from trino import auth as trino_auth
        scheme = cfg.get("scheme") or "https"
        kw = dict(host=cfg["host"], port=int(cfg["port"]), user=cfg["user"],
                  http_scheme=scheme)

        # 三個憑證在 CRD 裡都是選填的,因為 Trino 可能擋在 gateway 後面、自己
        # 不驗證。所以「沒有憑證」是一個合法的連法,不是設定漏填。
        if scheme == "http":
            # trino 的 client 會直接拒絕 basic over http,而 JWT 走明文等於把
            # token 送出去。兩個都不送,並且講明為什麼。
            if cfg.get("password") or cfg.get("jwt_access_token"):
                raise SystemExit(
                    "x trino 的 SCHEME 是 http,但同時給了 PASSWORD 或 "
                    "JWT_ACCESS_TOKEN。\n"
                    "  明文連線不送憑證:要嘛把 SCHEME 改成 https,要嘛把憑證留空"
                    "(未驗證的 Trino 是合法的連法)。")
        elif cfg.get("jwt_access_token"):
            kw["auth"] = trino_auth.JWTAuthentication(cfg["jwt_access_token"])
        elif cfg.get("password"):
            kw["auth"] = trino_auth.BasicAuthentication(cfg["user"], cfg["password"])

        if cfg.get("ssl_certificate_pem"):
            # requests 要的是一個 CA bundle 檔案路徑,不是 PEM 字串。
            kw["verify"] = _pem_file(cfg["ssl_certificate_pem"])
        return dbapi.connect(**kw)

    if cls == "athena":
        pyathena = _need(cls, "pyathena")
        kw = dict(region_name=cfg["region"],
                  aws_access_key_id=cfg["access_key_id"],
                  aws_secret_access_key=cfg["secret_access_key"],
                  s3_staging_dir=cfg["output_location"])
        if cfg.get("work_group"):
            kw["work_group"] = cfg["work_group"]
        return pyathena.connect(**kw)

    raise SystemExit(f"x '{cls}' 沒有 DBAPI 路徑")


def _dbapi_runner(conn) -> Runner:
    """把一個 DBAPI 連線包成 Runner。

    不用 `with conn.cursor()`:各家 driver 對 cursor 的 context manager 支援不一,
    trino 的就沒有,而那個差異在這裡只會表現成一個看不懂的 AttributeError。
    """
    def run(sql: str, limit: int) -> tuple[list[str], list[list[str]]]:
        cur = conn.cursor()
        try:
            try:
                cur.execute(sql)
            except Exception as e:
                if SHOW_TRACEBACK:
                    raise
                # driver 的訊息通常帶著行號與位置,那是最有用的部分。
                raise SystemExit(f"x 查詢失敗:{type(e).__name__}: {e}")
            if cur.description is None:
                return [], []
            # psycopg 的 Column 有 .name;其他 driver 的 description 是 tuple。
            cols = [getattr(d, "name", None) or d[0] for d in cur.description]
            rows = cur.fetchmany(limit) if limit > 0 else cur.fetchall()
            return cols, [["" if v is None else str(v) for v in r] for r in rows]
        finally:
            try:
                cur.close()
            finally:
                conn.close()
    return run


def runner(cls: str, prefix: str) -> Runner:
    """取得這個 class 的查詢函式。"""
    cfg = config(cls, prefix)
    how = SPECS[cls].connect
    if how == "dbapi":
        try:
            conn = _dbapi_connect(cls, cfg)
        except SystemExit:
            raise
        except Exception as e:
            if SHOW_TRACEBACK:
                raise
            # driver 的原文留著(它才知道是密碼錯還是連不上),但不要 traceback:
            # 八行 stack 把唯一有用的那行推到最下面,而且會印出這台機器的
            # 絕對路徑。
            raise SystemExit(
                f"x 連不上 {cls} 前綴 '{prefix}':{type(e).__name__}: {e}\n"
                f"  依序確認:host 與 port 通不通、帳密對不對、這個帳號看不看得到"
                f"這個 database。憑證跟系統的擁有者要,不要去讀叢集的 Secret。")
        return _dbapi_runner(conn)
    if how == "netsuite":
        import netsuite  # 同目錄
        return netsuite.runner(cfg)
    if how == "salesforce":
        import salesforce  # 同目錄
        return salesforce.runner(cfg)
    raise SystemExit(f"x '{cls}' 沒有連線路徑")


# --- 欄位內省 -------------------------------------------------------------
#
# 建 SemanticLayer 的第一步永遠是「這張表到底有哪些欄位、什麼型別」。六個走
# information_schema,Oracle 走 all_tab_columns,NetSuite 兩者都沒有、只能取樣。

def _split(table: str) -> tuple[str, str]:
    """把 schema.table 拆開。沒給 schema 就回空字串。"""
    if "." in table:
        schema, _, name = table.rpartition(".")
        return schema.strip('"'), name.strip('"')
    return "", table.strip('"')


def columns_query(cls: str, table: str) -> tuple[str, bool]:
    """回傳 (SQL, 結果是否為取樣列)。

    取樣列的意思是:查回來的**欄名**才是答案,列的內容不是 —— NetSuite 只能這樣問。
    """
    schema, name = _split(table)
    if cls == "netsuite":
        import netsuite  # 同目錄
        return netsuite.columns_sql(table), True
    if cls == "salesforce":
        raise SystemExit("x salesforce 走 describe 端點,不是 SQL(query.py 會自己處理)")
    if cls == "trino":
        # Trino 是聯邦查詢:information_schema 是**每個 catalog 一份**,不帶
        # catalog 的話 apiserver 回 MISSING_CATALOG_NAME,而那個訊息不會提到
        # 你少打的是什麼。所以這裡要三段式。
        parts = [x.strip('"') for x in table.split(".")]
        if len(parts) != 3:
            raise SystemExit(
                f"x trino 的表要寫成 catalog.schema.table(收到 '{table}')。\n"
                "  它把好幾個來源掛在一起,catalog 就是掛哪一個。\n"
                "  列出有哪些:--class trino --prefix <P> \"show catalogs\"")
        catalog, schema, name = parts
        return (f"select column_name, data_type, is_nullable "
                f"from {catalog}.information_schema.columns "
                f"where table_schema = '{schema}' and table_name = '{name}' "
                f"order by ordinal_position"), False

    if cls == "oracle":
        # Oracle 的字典把識別字存成大寫。
        where = f"table_name = '{name.upper()}'"
        if schema:
            where += f" and owner = '{schema.upper()}'"
        return (f"select column_name, data_type, nullable, data_length "
                f"from all_tab_columns where {where} order by column_id"), False
    where = f"table_name = '{name}'"
    if schema:
        where += f" and table_schema = '{schema}'"
    return (f"select column_name, data_type, is_nullable "
            f"from information_schema.columns where {where} order by ordinal_position"), False
