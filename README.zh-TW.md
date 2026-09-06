# asgard-fde-cli

[English](README.md) | **繁體中文**

Asgard FDE 的命令列工具（`asgard-cli`）。

## 安裝

**這個 repo 是 private，它的每一個 release 也是。** GitHub 的 release 完全繼承 repo 的可見性：頁面、release notes、每一個附件都只有有讀取權限的帳號拿得到，未登入請求資產 URL 拿到的是 **404 而不是 403**。這裡的東西不會發佈到任何其他地方 —— 沒有 Homebrew tap、沒有 Scoop bucket、沒有套件庫（理由見[發佈](#發佈)）。

所以每一條安裝路徑都需要憑證，而最短的那條用的是 `gh` 已經握著的：

```bash
gh release download --repo asgard-ai-partners/asgard-fde-cli \
  --pattern '*_darwin_arm64.tar.gz' --output - | tar xz asgard-cli
sudo mv asgard-cli /usr/local/bin/
asgard-cli doctor          # 告訴你 helm 在不在 PATH 上
```

把 pattern 換成你的平台 —— 資產涵蓋 darwin / linux / windows 的 amd64 / arm64，Debian 或 RPM 主機可以改用 `.deb` / `.rpm`。**用 CLI 下載的檔案不會像瀏覽器下載那樣被 Gatekeeper 隔離**，所以不需要 `xattr` 那一套。

如果你本來就會編 Go，只要 git 進得去這個 private repo，module 直接可用：

```bash
GOPRIVATE=github.com/asgard-ai-partners/* \
  go install github.com/asgard-ai-partners/asgard-fde-cli/cmd/asgard-cli@latest
```

`asgard-cli version` 報的是 release 編進去的值；不帶 ldflags 的 `go build` 會退回 module 與 VCS metadata，而不是宣稱一個它沒有的版號。

## 開發

```bash
go build -o asgard-cli ./cmd/asgard-cli   # build
./asgard-cli version                      # run
```

版面：

```
cmd/asgard-cli/     main；只做 signal handling 與 exit code
internal/cli/       cobra 指令樹，一個子指令一個檔
internal/repo/      客戶 repo 由什麼構成 —— 用看的，不用存的
internal/auth/      OAuth 流程與憑證儲存，那是這支 CLI 唯一放在 repo 外的檔案
internal/work/      客戶 repo 自己的工作紀錄（request、task spec、open question、決議紀錄）
internal/kb/        列出、讀取、評分、出處與連結圖的單一實作，所有語料共用
internal/wiki/      平台 wiki，以及旁邊的客戶詞彙表
internal/usecase/   部署形狀的擷取檔
internal/stage/     onboarding 的提示文，對 repo 的狀態渲染
internal/scaffold/  寫出非客戶特有的骨架，並供應裡面的 skill
internal/generate/  CR 骨架，接上 chart 已經宣告的東西
internal/size/      部署形狀，從 production 數出來的
internal/brief/     一個活動會怎麼做錯，依意圖而非依位置編排
internal/check/     repo 結構：索引、日期檔名、連結、孤兒頁
internal/gate/      對已渲染 chart 的不變量檢查（xref、agent 拆分、enum）
internal/render/    透過 helm 渲染一個 release 的 chart，注入佔位的 asgard 值
internal/binding/   讀寫 .asgard-cli.yaml，這個 checkout 的平台綁定
internal/pipelineconfig/ 讀 .asgard-pipeline.yaml，部署宣告
internal/chart/     讀一個 project **未渲染**的 template，取 (kind, name)
internal/tool/      解析 helm/kubectl/python3，以及怎麼安裝
internal/version/   build 資訊（GoReleaser 用 ldflags 注入）
```

要新增子指令：在 `internal/cli/` 寫一個 `newXxxCmd()`，並在 `root.go` 用 `addTo(cmd, group..., ...)` 註冊。**分組是必填的** —— cobra 對父層沒有的 `GroupID` 會 panic，所以一個指令不可能在沒決定它屬於哪一組的情況下被加進來。

- [STRUCTURE.md](STRUCTURE.md) —— 每個目錄是做什麼的，包含四份內嵌語料以及一個改動屬於哪一份。
- [AGENTS.md](AGENTS.md) —— 這個 repo 遵循的慣例，以及在沒有測試套件的情況下 gate 是什麼。
- [TASK.md](TASK.md) —— 這個 repo 是為什麼存在的，以及還沒完成的部分。

## Coding Agent 會怎麼用它

這一節不在英文版裡，它是給正在盤點的人看的。

**這支 CLI 有兩種讀者，而且兩種都不是「坐在終端機前打字的人」為主。**

第一種是 **FDE 本人**，通常在客戶會議前後：`brief`、`guide`、`find`、`size`、`wiki`、`usecase` 都**不需要 repo、不需要網路、不需要登入**，因為那些問題是在會議室裡被問的，那時候還沒有目錄。

第二種、也是設計時真正瞄準的，是**在客戶 repo 裡工作的 Coding Agent**。它的迴圈長這樣：

```
1. asgard-cli init             一次把骨架、綁定、平台參考素材都建好
2. 讀 .agents/skills/          asgard-cr-shapes（這座叢集的 CRD 欄位）
                               asgard-workflow-processors（這座 runtime 的 processor config）
                               asgard-cr-verification（平台會怎麼檢查、每個規則碼是什麼意思）
3. asgard-cli add <kind>       寫一份 CR 骨架，會失敗得無聲無息的部分是對的
4. asgard-cli gate             本機能檢的全部，一個指令
5. git push <tag>              推上去
6. asgard-cli pipeline runs watch   讀回平台的 plan report
```

其中三件事是刻意這樣設計的：

- **第 4 步是一個指令，不是一張清單。** 一個全是 CR 的 repo 沒有編譯器；寫編譯語言的 agent 不會漏檢查，是因為「不管改了什麼都要跑 build」而且 build 是一個指令。散在 `AGENTS.md`／提示文／agent 記憶裡的清單會過期，而且真的過期過。
- **第 2 步的素材是抓下來的，不是編在 binary 裡的。** 客戶的 server 版本跟他裝的 CLI release 無關（On-Prem），所以「某個欄位叫什麼」是關於**那座 server** 的事實，不是關於某個 CLI 版本的。
- **第 4 步不重做第 6 步的檢查。** CR 能不能被接受是 apiserver 決定的，而**平台永遠不發叢集憑證給 client**。本機綠燈的意思是「值得推上去」，永遠不是「這會部署成功」。

**每一個讀取類指令都吃 `--format json`**，因為 agent 要解析而不是要看。`check` 與 `verify` 這一對最重要：文字模式下警告與失敗只差左邊一個字，而只有一個是致命的；JSON 模式下它們是兩個不同的陣列。

## 指令總覽

```
Ask —— 平台是什麼、一個形狀怎麼組起來
  find <terms>           一次搜四份語料，並交出對應的另一半
  wiki [page]            平台由什麼構成，每一塊是給誰的
  usecase [name]         一種部署形狀怎麼一個欄位一個欄位組起來
  brief [what]           你正要做的事，前人在哪裡做錯過
  guide [name]           一個決策的指引；不帶參數列出全部
  size [shape]           一個能力寫出來之前，它由什麼構成、要多少
  reading                這個 engagement 讀過什麼、沒讀過什麼
  issue-report           這支工具哪裡錯了或不知道，怎麼回報

Build —— 寫出 repo 與裡面的 CR
  init                   一個指令完成 onboarding：骨架 + 綁定 + 參考素材
  scaffold               只寫骨架
  project                每份 chart 宣告了什麼
    add <slug>           寫一個 project 的 chart 骨架
  add <kind> <name>      寫一份 CR 骨架，接上 chart 已宣告的東西
  question / add / answered      沒人回答的問題，以及答案是什麼
  request / add / target / ready / done    客戶要的東西
  task / add / ready / start / done        task spec
  decision add           寫一份帶日期的決議紀錄
  reference add          歸檔客戶自己的素材，連同出處

Check —— 這台機器能檢查的全部
  gate [release ...]     ★ 一個指令跑完下面所有能跑的
  check [project]        repo 的結構不變量
  render <release>       用原生 helm 渲染，注入佔位的平台值
  verify [release]       渲染後的 CR 交叉引用與不變量
  doctor                 外部工具在不在，不在的話怎麼裝

Deploy —— 平台，以及它知道的事
  login / logout / whoami
  workspace  list / use <id> / show
  pipeline   connect / connections / repos / create / list / use <id> / show
             projects / release create|show / releases / deliveries
             runs list|get|log|watch|approve|reject|cancel
             variables list|set|unset|sync-declared
             manifest
  skill      status / update

（另有隱藏指令 audit-material，是給維護這份語料的人用的，不是給客戶 onboarding 用的。）
```

## 指令

### `init`

一個指令完成 repo 的 onboarding：骨架、綁定、以及描述它要部署到哪座平台的參考素材。

```bash
asgard-cli workspace list
asgard-cli pipeline list --workspace <id>
asgard-cli init --workspace <id> --pipeline <id>
```

```
workspace   1878677014576107520  (JohnWS)
pipeline    2095845443282931712  (iac-test -> asgard-ai-platform/asgard-iac-test)
directory   /path/to/acme-asgard-kube

1/3 skeleton
  created      .asgard-pipeline.yaml
  ...
2/3 binding
  wrote        .asgard-cli.yaml
3/3 reference material
  ...
```

它跑 `scaffold`、把 `.asgard-cli.yaml` 寫在 scaffold 產生的宣告檔旁邊、再跑 `skill update` —— 順序是唯一可行的那個，因為綁定要放在一個必須先存在的宣告檔旁邊。三者各自仍是獨立指令、各自可以單獨重跑；合在一起是因為**一份寫在散文裡的三步清單會過期**，而這一份真的過期過。

**它什麼都不猜。** 兩個 id 要嘛給、要嘛已經記著；兩個都沒有時它列出候選然後停下來，而且**一個候選跟五個候選的做法完全一樣**。在一個已經記著兩者的 repo 上，兩個旗標都不需要 —— 所以升級之後重跑一次是把 repo 帶到最新的安全做法。

**它不建立 pipeline。** 那件事會在 provider 上綁定一個 repo、需要一個 VCS connection，是關於平台的決定而不是關於這個 checkout 的：`asgard-cli pipeline create` 做那件事，這裡只記錄結果。

`--force` 覆寫骨架；`--skip-skills` 跳過第 3 步並**報成 skipped，那不是 pass**。

以前有另一個 `init`，它寫 `.asgard-config.json`。那個檔沒了：project 清單直接讀 repo、客戶名字跟平台問、而它記的部署「形狀」是一個沒有工具能檢查的意圖。

### `project add`

在 `projects/<slug>/chart/app` 底下寫一個 project 的 chart 骨架。

```bash
asgard-cli project add internal
```

一個 project 就是一份 Helm chart。**哪些 Release 部署它、部署到哪個平台 Project，是在 `.asgard-pipeline.yaml` 裡宣告的** —— 這個指令寫 chart，替它宣告一個 Release 是另一件事，這裡不做。

**沒有任何地方另外記錄這個 project。** 它存在是因為目錄存在、以及宣告檔指名了它的 chart；沒有第三份清單要同步，也就沒有任何一份會跟其他兩份不一致。既有檔案不動，所以重跑是安全的。

slug 會出現在 chart 渲染出來的物件名稱裡，所以要短：Kubernetes 的名稱上限是 63 字元，而衍生出來的名稱會繼承它的長度。

### `scaffold`

寫出 repo 骨架，包含其他東西都掛在上面的那份宣告檔 —— **給還沒綁定 pipeline 的 repo 用**：

```bash
asgard-cli scaffold
```

平台上已經有這個 repo 的 pipeline 時，要跑的是 [`init`](#init)：它會跑這個、記下綁定、再抓參考素材。**這個指令就是同一件事把平台拿掉**，而那正是它能做而 `init` 不能做的唯一一件事 —— `init` 兩個 id 都必填、而且都不猜，一個還沒有 pipeline 的 repo 兩個都給不出來。真的長這樣的情境有兩種：**還沒開戶**（提案、spike、有人正在建 workspace 的同時先寫 repo），以及**既有 repo 要遷移過來**（已經有 chart 和自己的 CI，缺的是骨架與 pipeline，順序也是這樣）。

pipeline 建好之後跑 `asgard-cli init` 記下它，重跑是安全的，骨架多半會被報成已存在。

它寫的是每個 engagement 都一樣的那部分：

| | |
|---|---|
| `AGENTS.md` | 平台契約，客戶特有的段落標成 TODO |
| `docs/` | 四層模型（meeting-notes / decisions / living spec）與 SDD 規則 |
| `requirements/` | task 與 request 的索引 |
| `.agents/skills/` | 對任何 Asgard 都成立的七個設計期 skill，`db-query` 是其中之一；描述某一座 server 的那些來自 `asgard-cli skill update` |
| `common/` | runtime skill 目錄 |
| `.asgard-pipeline.yaml` | 部署宣告，每個 project 一個 release 待填 |
| `projects/<slug>/` | 每個 project 一份 chart 骨架 |

它**不寫**的是客戶自己的知識：有哪些系統、工作怎麼切分、CR 長什麼樣。那是 onboarding 產出的，沒有模板生得出來。

重跑是安全的。既有檔案不動並計為已存在，所以加了 project 之後、或某個檔被手動刪掉之後可以再跑一次。`--force` 覆寫，那會丟掉對骨架的本機修改 —— 而它也是**升級後重新取用 shipped 素材**的方式，否則那些檔會被報成 `stale` 而不是被安靜換掉。

產生出來的骨架第一次跑就過得了自己的 gate：

```bash
asgard-cli gate
```

**不要手動跑 `helm lint`。** 平台會在每次渲染時注入保留的 `.Values.asgard` 區塊，而 chart **不可以**在自己的 `values.yaml` 宣告它 —— 所以不帶 `-f` 裸跑會在每個讀 `.Values.asgard.projectEnvironmentId` 的 chart 上失敗，也就是每個會貼 label 的 chart。`gate` 只供給那一個檔案，其他都不給。

### `guide`

**`guide` 讀一個決策。** 沒有任何指令會說「這個 engagement 走到哪裡」，而且那是刻意的 —— 曾經有兩個這樣的指令，都拿掉了。

```bash
asgard-cli guide                   # 全部的指引
asgard-cli guide requirements      # 其中一份，隨時
asgard-cli find "<terms>"          # 依主題到達其中任何一份
```

```
  init           開始 onboarding
  scaffold       寫出 repo 骨架
  requirements   把客戶說的話變成一個 request
  projects       決定工作怎麼切成 project
  data-sources   接上客戶的資料庫
  read-path      決定每個 project 的讀取路徑
  entry-point    決定每個 project 的入口
  knowledge      決定非結構化知識放哪裡
  verify         跑驗收 gate
  deploy         部署
  enhance        在已經上線的 repo 上加一個能力
  idle           手上沒有進行中的事
```

**這些沒有一個是「你會抵達的步驟」。** 它們曾經被編號，而 `next` 從「最早缺的那個 CR kind」推導出「stage 4 of 9」。那兩個方向都錯：它一次只能指名一件事，所以其中三個除非你已經知道名字否則到達不了；而一個位置是無法被反駁的，於是一個用不同順序在做事的 engagement 會被告知它落後了。

拿掉編號還不夠。`status` 改成從條件升起同樣那些階梯 —— `switch` 裡的 `!p.Has("DataConnector")`，所以一份缺兩樣東西的 chart 只會被告知第一樣 —— 那是把算術藏起來的位置。**現在沒有任何東西會升起指引**，它只能用 `guide` 依名字、用 `find` 依主題到達。

**`read-path`、`entry-point`、`knowledge` 會把錯的答案印在對的答案旁邊。** 那是這個 engagement 曾經做錯又反轉的三個決策，而且每一次錯的那個都是看起來理所當然的那個。

**一份 chart 不一定以入口作結**，而且這裡沒有任何指令會說它完成了沒有。一個沒掛任何東西的 SemanticLayer，可能是一份做完的 Mimir 交付物，也可能是一個還沒人寫的 agent，而檔案分不出來。`asgard-cli size <shape>` 列出一個形狀由什麼構成給人比對；沒有任何地方記錄一份 chart 打算長成什麼樣，因為「某人本來要建什麼」不是這支工具能檢查的東西。

### `project`、`request`、`task`、`question`

**四個指令把 repo 讀回來給你**，一個指令一個檔，每個都吃 `--format json`。它們**不從彼此推導任何東西**。

```bash
asgard-cli question    # 沒人回答的問題，以及每一個要問誰
asgard-cli request     # 客戶要的、還沒做完的
asgard-cli task        # 開著的 task spec
asgard-cli project     # 每份 chart 宣告了什麼
```

**先讀 `question`。** 在別人開的 repo 裡做出傷害最快的方式，就是繞過一個他們早就知道還開著的問題去設計。

```
Projects:

  insight              DataConnector, SemanticLayer
  helpdesk             chart is empty
```

**它說每份 chart「有」什麼，完全不說它缺什麼。** 那件事以前是對著每個 project 記錄的一個「形狀」去量的，而報出「這個形狀要 X 但 X 不在」等於把某人記下的意圖當成一份這支工具可以檢查的規格。清單本身就是 repo —— 宣告檔指名的 chart 路徑，加上 `projects/` 底下的目錄 —— 所以沒有第二份會漂移。

### `request`、`task`、`question`、`decision` —— 寫紀錄

工作以 **request** 進來：客戶想要、而 agent 今天做不到的一件事。其他所有東西都掛在它下面。

```bash
asgard-cli request add "倉管人員想在聊天裡問庫存"
asgard-cli request target REQ-001 erp
asgard-cli request ready REQ-001

asgard-cli task add "把庫存查詢開出來" --request REQ-001 --project erp --complexity M
asgard-cli task ready TASK-001
asgard-cli task start TASK-001
asgard-cli task done TASK-001

asgard-cli question add "哪一個庫存數字才算數" --blocks REQ-001 --ask "倉管主管"
asgard-cli question answered 1 "只算 608 儲位" --decision 2026-09-04-safety-stock.md

asgard-cli decision add "官網一律走固定查詢工具讀取" --module architecture.md
```

**這些狀態沒有一個放在 CLI 裡。** 每個指令都在客戶的 repo 裡寫一個檔，因為那個 repo 才是下一個 agent 會打開的東西：

| 紀錄 | 檔案 | 帶什麼 |
|---|---|---|
| request | `requirements/requests/REQ-xxx-<name>.md` ＋ registry 的一列 | 提出日期、狀態、目標 project、客戶自己的用字 |
| task spec | `requirements/tasks/TASK-xxx-<name>.md` ＋ queue 的一列 | 建立日期、狀態、SDD 各節、每次狀態轉移的日期紀錄 |
| open question | `docs/open-questions.md` 裡的一列 | 提出日期、它擋住什麼、誰能回答 |
| decision | `docs/decisions/YYYY-MM-DD-<topic>.md` ＋ 一列追溯 | 檔名裡的日期、它改動的模組 |

**這些之所以是指令而不是「請你寫一個檔」的指示**，是因為每一筆紀錄都活在不只一個地方。一個 task 的狀態在 queue 表格裡、在 spec 自己的 `Meta` 裡、也在 spec 的執行紀錄裡；一個 request 的目標 project 在 registry 的 Spec 欄與它的 `Meta` 裡。手動搬一個要改三到四處，而一個其中兩處互相矛盾的 repo，會讓下一個讀的人沒辦法判斷哪一個是現況。這裡每個指令都一次搬完全部，而且日期是蓋上去的、不是問來的。

### `add`

在一個 project 的 chart 裡寫一份 CR 骨架，**會失敗得無聲無息的那些部分是對的**。

```bash
asgard-cli add                                   # 列出所有 kind
asgard-cli add dataconnector erp --db-class postgres --project erp
asgard-cli add flowagent support --bot-class line --project site
```

十種 kind：`dataconnector`、`semanticlayer`、`agent`、`httptool`、`querytool`、`skillset`、`trigger`、`knowledgedrive`、`plugin`、`flowagent`。

**它產生的是骨架**：結構與陷阱是對的，內容標成 TODO。值得產生的正是那些沒有任何東西會抓到的部分 —— 缺一個顯示用 annotation 會讓 UI 出現一個沒有名字的資源、一個沒有 set label 的 Workflow 在那裡是隱形的、一個沒有自己那兩個 label 的 Trigger 打開是一張白畫布、而一個上游改過名的欄位用舊名字仍然 lint 得乾乾淨淨。這些沒有一個會被 `helm lint`、CRD 驗證或 server-side dry-run 抓到。

它會先讀 chart 再往裡面寫，所以它產出的引用指得到真的存在的東西：chart 裡只有一個 SemanticLayer 就掛上去、有好幾個就要求指名、有 SkillSet 才引用。第二個查詢工具不會重新產出第一個已經寫好的 Toolset。

### `wiki`、`usecase`

兩份參考語料，**編在 binary 裡而不是寫進客戶 repo**：一份放在某個 engagement 裡的副本會在沒人看的地方過期，而一頁放在這裡的過期內容，一次發版就替所有 engagement 修好了。

```bash
asgard-cli wiki                       # 平台由什麼構成
asgard-cli wiki agents
asgard-cli wiki --search "匿名 訪客"
asgard-cli wiki --conventions         # 這份 wiki 怎麼維護

asgard-cli usecase                    # 每種部署形狀怎麼組起來
asgard-cli usecase flow-agent-single
asgard-cli usecase --search schedule
```

| | 回答什麼 | 寫自哪裡 |
|---|---|---|
| `wiki` | 平台是什麼、每一塊是給誰的、以及 UI 的名字從哪裡開始對不上 chart 宣告的資源 | 產品文件 [asgard-docs](https://github.com/asgard-ai-platform/asgard-docs)，對著 CRD 校過 |
| `usecase` | 一種部署形狀怎麼一個欄位一個欄位組起來，以及填錯一個值的代價 | 已經在 production 跑的部署 |

**要查東西請用 [`find`](#find)**，它同時搜兩邊並指出命中那一邊的對應另一半。

### `find`

**這是入口。** 它一次搜四份語料 —— wiki、擷取檔、階段指引、設計期 skill —— 因為答案在哪一份，通常在搜之前並不明顯。它**不需要 repo**。

```bash
asgard-cli find schedule
asgard-cli find anonymous visitor
asgard-cli find 儀表板                    # 搜尋前先翻譯
asgard-cli find schedule --format json
```

**用問題被問出來的那個語言去問。** 語料是英文的，而跟客戶的對話不是，所以索引會在搜尋前套用，而且會把改寫印出來：

```
$ asgard-cli find 電商
This material is in English. "電商" was read as:

    commerce marketplace channel
```

索引是 `asgard-cli wiki --aliases`，而且它**不是一頁**。它跟頁面並排放，因為一份放在被搜語料裡的索引會跟它指向的東西競爭：那張表列出所有別名，於是它可靠地命中翻譯後查詢的每一個詞，讀者拿到的是字表而不是頁面。

**一列「只是路由」的紀錄讀起來跟一列「有答案」的紀錄一模一樣**，所以 `find` 會講出來是哪一種：

```
$ asgard-cli find 綠界
**Nothing here names 綠界.** What follows is the shape it belongs to, which
is what this material has - not material about the product.
```

**一個詞在這份語料裡被佔用過，會在結果之前先被標出來，不是之後。** 一次什麼都沒找到的搜尋會被記錄下來並告知讀者；而一次找到**錯誤語意**的搜尋，看起來跟一個答案一模一樣，畫面上沒有任何一處是紅的。那張表是 `asgard-cli wiki glossary`，而且它是被套用在查詢上、不只是給人讀的。

**走進死路時它問一個問題，而不是猜。** 一次什麼都沒找到的搜尋會被記錄在 engagement 裡、寫進一個不進版控的檔 —— 一次查詢帶著的是客戶用的字。

```bash
asgard-cli reading --misses      # 這個 engagement 搜過但沒找到的
asgard-cli issue-report --new    # 缺陷回報，證據已經在裡面了
```

### `brief`、`size`、`reading`、`issue-report`

四個入口，沒有一個是位置。除了 `reading` 之外都不需要 repo。

**`brief`** 回答「我正要做的這件事 —— 我會在哪裡做錯」，那是任何 repo 報告都答不出來的：風險最高的活動在 repo 裡不留痕跡，因為**跟客戶講話不會改任何檔案**，而會議在每個階段都會發生。

**`size`** 是一個能力被寫出來之前由什麼構成 —— 提案被問的第一個問題，也是報價的基礎。數字來自 production 的部署而不是推理，那在直覺答案錯的地方最重要：**flow-agent 那幾種形狀裡完全沒有 `Agent` CR。**

**`reading`** 報告這個 engagement 開過哪些頁、以及從來沒開過哪些。它只記頁面名稱、不記任何關於客戶的東西，所以**可以安全地進版控** —— 而且值得，因為六個月後它會說出前一個人知道什麼。**代價最高的是那些對的、但沒人打開過的頁。**

**`issue-report`** 是這支工具的缺口怎麼被回報。那個缺口不屬於客戶 repo：寫在一個 engagement 裡的筆記，就只有那一個 engagement 有，下一個從頭開始。

### `check`

驗收 gate 的第一步，也是唯一不需要外部工具的一步。它驗證的是 chart 渲染看不到的不變量：

```bash
asgard-cli check                    # 整個 repo
asgard-cli check erp                # project 範圍的檢查只限 erp
asgard-cli check --format json      # error 與 warning 是兩個陣列
```

- 根 README 的 project 表格跟 `projects/` 底下的目錄一致
- `.asgard-pipeline.yaml` 解析得開、沒有重複的 release 名、每個宣告的 release 都指向一個有 `Chart.yaml` 的 chart 目錄
- `common/skills/` 底下的 runtime skill 帶著 `name` 與 `description` frontmatter，且名字與目錄相符
- `requirements/` 底下的 SDD 入口都在
- `docs/` 的 spec 層完整：必要檔案、living spec 的模組索引與磁碟上的檔案相符、日期檔名、以及 `docs/` 裡每個相對連結都解得開
- **沒有孤兒頁** —— `docs/` 或 `requirements/` 底下一份沒有任何東西連到的文件不會被讀到，而寫它的人永遠不會發現，因為檔案還在。這是警告而非錯誤：一份今天記下、還沒被套用的決議，在被套用之前本來就是孤兒。

### `render`、`verify`、`doctor`

這三個是驗收 gate 現在能在 Windows 上跑的原因。

```bash
asgard-cli render internal-dev         # manifest 到 stdout，摘要到 stderr
asgard-cli verify                      # 渲染每個 release，檢查不變量
asgard-cli verify --rendered file.yaml # 檢查一份已經渲染好的串流
asgard-cli doctor                      # 哪些外部工具在這裡，不在的話怎麼拿
```

`render` 吃的是一個 **release**，不是 project 加環境。一份 chart 部署到哪裡是 `.asgard-pipeline.yaml` 裡的一個 Release，而一份 chart 可以有好幾個。

**`check` 與 `verify` 是 agent 最常用力的一對**，因為那是它想要弄綠的 gate —— 所以兩個都吃 `--format json`。

舊的鏈是 `bash render.sh → yq → helm template → python3 + PyYAML`：四個外部相依，其中**三個在沒有 WSL 或 Git Bash 的 Windows 上不能跑**。現在的前置只有 `helm` 一個，加上一個 binary。

**以前的第 4 步已經不是本機的步驟了。** 它跑 `kubectl` 與一支 `check_crd_fidelity.py` 對著叢集；兩者都在 Pipeline 切換時移到平台的 plan 上，因為最有價值的檢查 —— apiserver 自己的 CEL、pattern 與必填驗證，以及 dry run 藏起來的 unknown-field 修剪 —— **需要一座叢集，而任何 client 都拿不到叢集憑證**。本機那一半是 `asgard-cli gate`，權威是 plan report。

### helm 與 kubectl 是前置條件，不是相依

asgard-cli 的散佈方式沒有一種裝得了它們。tar.gz、zip 與 `go install` 都不帶相依 metadata、也永遠不會帶；而宣告在 Homebrew tap 或 Scoop bucket 上的相依只涵蓋用那種方式安裝的人 —— 而那是沒有人，因為對一個 private repo 來說那些通道是**刻意關掉的**。宣告一個不成立的保證，讀起來就像一個保證。

所以 binary 本身就是那個機制。每個需要 helm 的指令都先透過 `internal/tool` 解析它，不在就帶著這台機器的安裝指令拒絕；`asgard-cli doctor` 一次報告全部。

`init`、`scaffold`、`project`、`request`、`task`、`question`、`decision`、`check` **都不需要這些工具**。`render`、`verify` 與 `gate` 的三個 chart 步驟需要 helm。**kubectl 是選用的**：這個 binary 沒有任何一處跟叢集講話，`doctor` 列它是因為一個在除錯部署的人仍然想知道它在不在。

### 三個檔案

四個，而且每一個都是別人對不同問題的答案。

| 檔案 | 誰寫 | 誰讀 | 進版控 |
|---|---|---|---|
| `.asgard-pipeline.yaml` | 人 | **平台**，每次 run | 是 |
| `.asgard-cli.yaml` | `asgard-cli` | 只有 `asgard-cli` | 是 |
| `os.UserConfigDir()/asgard-cli/credentials.json` | `asgard-cli login` | `asgard-cli` | **絕不** |
| `os.UserConfigDir()/asgard-cli/profiles.json` | `asgard-cli profile set` | `asgard-cli` | **絕不**（但可以直接交給同事） |

**`.asgard-pipeline.yaml` 是宣告檔**，也是一次部署唯一依賴的檔案：有哪些 release、每個部署哪份 chart、什麼觸發它、吃哪些 key。

**`.asgard-cli.yaml` 是綁定檔**：這個 checkout 對到哪個 workspace、哪條 pipeline，就這兩件事。**兩欄都必填、都不推導。**

**`credentials.json` 是這支 CLI 唯一放在 repo 外的東西。** 0600、所有 profile 共用一個檔、旁邊什麼都沒有。那裡以前還有一個 `config.json`，存預設 profile、每個 profile 的預設 workspace、以及自訂 profile 的 map；它沒了。每個欄位都是某個旗標或環境變數已經表達過的偏好，而每一個都是升級時要繼續理解的東西。**憑證是唯一真的非放那裡不可的**：它是機密、屬於人而不是屬於 repo、而且推導不出來。

殘留的 `config.json` 是**錯誤而不是警告**。退休的 `defaultProfile` 多半是 `dev`，靜默忽略它會讓每個指令移到 `prod`（客戶的平台）。第一個要解析 profile 的指令會拒絕執行，逐條列出退休的 key 與替代品，並叫你刪掉那個檔。

**刻意不存在任何地方的東西**：這個 repo 有哪些 project（讀宣告檔的 chart 路徑與 `projects/*/`）、客戶的顯示名稱（模板需要時跟平台問）、以及一份 chart「本來要長成什麼樣」（一個沒有工具能檢查的意圖）。它們各自違反的判準是：*一個值該留在設定檔，只有在磁碟上沒有任何東西推得出它、而且平台問不到的時候。*

### `login`、`logout`、`whoami`

登入 Asgard 平台，讓 `pipeline` 指令能以你的身分行動。

```bash
asgard-cli login                     # 登入 prod
asgard-cli login --profile dev       # 登入 dev
asgard-cli login --no-browser        # 印出 URL 而不是開瀏覽器
asgard-cli whoami                    # 問平台這個 session 是誰
asgard-cli logout --all              # 忘掉每一個 session
```

OAuth 2.0 authorization code ＋ PKCE，走 loopback redirect，那是 RFC 8252 對一台有瀏覽器的機器給的答案。binary 不帶 client secret。session 存在這個使用者帳號底下 —— **絕不放在客戶 repo 裡** —— 位置是 `os.UserConfigDir()/asgard-cli/`，0600。

一個 **profile 就是一座 Asgard 安裝**。`--profile` 逐指令指定、`ASGARD_PROFILE` 對一個 shell 生效；兩個都沒有時是 `default`，也就是**代管平台**。**沒有任何地方記錄「現在是哪個 profile」** —— `login --set-default` 以前可以，它沒了：一台機器上的偏好是另外那個人不會有的偏好。

代管平台以外的 profile 用 [`asgard-cli profile`](#profile) 寫。`ASGARD_PLATFORM_API`、`ASGARD_ISSUER`、`ASGARD_CLIENT_ID` 仍然可以逐欄位蓋在生效的 profile 上，那是給一次性用的。

沒有瀏覽器的時候（CI、容器、agent sandbox）改設 `ASGARD_TOKEN`。它**完全繞過儲存**，不讀磁碟也不寫磁碟。

### `profile`

**如果你用的是代管的 Asgard 平台，這一整節你都不需要。** 完全沒有這個檔的時候，每個指令都會連到它 —— 那就是 `default` 的意思，也是它之所以是預設的理由。這些指令是為了兩種**這個 binary 不可能知道**的安裝：**地端部署**，以及**跑在本機的 stack**。

```bash
asgard-cli profile list              # 設了哪些、現在生效的是哪個
asgard-cli profile show [name]       # 三個值，以及每個是哪來的
asgard-cli profile set onprem --platform-api https://asgard.acme.internal \
    --issuer https://iam.acme.internal --client-id abc123
asgard-cli profile remove onprem
```

一個 profile 裝三個值，而且**每一個各自 fallback** 到代管平台的：

| | |
|---|---|
| `--platform-api` | Asgard Platform API 在哪裡 |
| `--issuer` | 替它發 token 的那座 Casdoor |
| `--client-id` | 這支 CLI 用哪個 application 出面（**不是機密** —— 每次登入都會出現在瀏覽器） |

**地端安裝三個都要設。** 它的 API 與替它發 token 的 Casdoor 是**同一座部署**，而其中一座發的 token 另一座不會接受 —— 所以只設 API 的結果是：你對著**代管的** Casdoor 登入，然後把那顆 token 送去別人的伺服器。`profile show` 會印出每個值的來源，並在兩者不同源時警告：

```
profile        onprem
platform api   https://asgard.acme.internal    profiles.json
issuer         https://iam.asgard-ai.com       the hosted platform (this profile does not set it)
client id      r21ntx0eb5igyokl3px4            the hosted platform (this profile does not set it)

WARNING: profile "onprem" takes its Platform API from profiles.json and its
identity provider from the hosted platform ...
```

它**警告而不拒絕**，因為「本機的 Platform API 配真的 Casdoor」是一種合理的開發方式。

**沒有 `profile use`。** 記錄「現在是哪個 profile」是退休的 `config.json` 裡唯一不會回來的那個欄位 —— 它在設了它的那台機器上看不見，在其他每一台上都不存在。選哪個 profile 就是 `--profile`、`ASGARD_PROFILE`、或 `default`。

`profile set` 是**唯一**會建立 `profiles.json` 的指令，而且只在你跑它的時候。沒有任何東西會把它當成別的動作的副作用寫出來。它不含機密，所以可以直接交給要架同一座安裝的同事；**憑證是另一個檔，不行**。

#### 對我們自己的開發平台工作

`dev` **不是內建名字**。我們的開發平台是這支工具會遇到的其中一座安裝，不是第二種東西；把它編進去等於在每個客戶的 binary 裡塞一個內部端點。

跟其他 profile 一樣寫出來，三個值去看內部的設定筆記 —— **它們不在這個 repo 裡**：

```bash
asgard-cli profile set dev \
    --issuer       <內部>  \
    --client-id    <內部>  \
    --platform-api <內部>
asgard-cli login --profile dev
export ASGARD_PROFILE=dev        # 對一個 shell 生效
```

### `workspace`、`pipeline use`

這個 checkout 對到哪個 workspace、哪條 pipeline —— **repo 裡沒有任何東西蘊含、平台也問不到**的那兩件事。

```bash
asgard-cli workspace list            # 這個帳號看得到哪些
asgard-cli workspace use <id>        # 記下 workspace
asgard-cli pipeline list             # 那個 workspace 有哪些 pipeline
asgard-cli pipeline use <id>         # 記下 pipeline
asgard-cli workspace show            # 這裡適用哪一個、為什麼是它
```

兩者都寫進 `.asgard-cli.yaml`，放在它們所屬的宣告檔旁邊，而**那個檔要進版控**：clone 這個 repo 的人、以及在裡面工作的 agent，之後都不需要帶旗標。平台永遠不讀它。

**任何東西都不會被推導，包含只有一個候選的時候。** 一個沒有記錄的指令會列出候選然後拒絕執行。一條只在清單只有一筆時才成立的規則，會在它變成兩筆的那天開始解析到一個沒人選過的對象 —— 而那天沒有人在看。

**不把任何 git remote 當作身分。** pipeline 以前是拿 checkout 的 `origin` 去比對 workspace 裡的 pipeline 找出來的；一個 checkout 可以有任意多個 remote，哪一個叫那個名字是它主人的事。那樣留下的缺口 —— 整個 repo 被複製到**同一個 workspace** 的另一個 repo，pipeline id 仍然解得開 —— 寫在 `.asgard-cli.yaml` 自己的檔頭裡，而不是用一條會對錯誤輸入開火的規則去擋。複製 repo 之後請跑 `pipeline use`。

**換 workspace 會清掉 pipeline 那一行**，因為一條 pipeline 隸屬於一個 workspace。之後每個 pipeline 指令都會拒絕執行並指名補救指令，`asgard-cli gate` 的 `binding` 步驟會紅燈 —— **失敗被前置到下一個指令，而不是被推遲到某個具破壞性的指令。兩行要一起 commit。**

解析順序，高到低：`--workspace`、`ASGARD_WORKSPACE`、checkout 的 `.asgard-cli.yaml`。**就這三個。**

### `pipeline`

透過平台的 IaC pipeline 部署這個 repo。

```bash
asgard-cli pipeline connect                # 把 GitHub 接到這個 workspace
asgard-cli pipeline connections            # 已經接上的 installation
asgard-cli pipeline repos --connection X   # 其中一個看得到哪些 repo
asgard-cli pipeline create --name p --connection <id> --repo R
asgard-cli pipeline list                   # 這個 workspace 有哪些 pipeline
asgard-cli pipeline use <id>               # 記下這個 checkout 用哪一條
asgard-cli pipeline show                   # 這個 checkout 記著的那一條
asgard-cli pipeline releases               # 它的 release，以及只被宣告的幽靈列
asgard-cli pipeline variables list --release <name>
asgard-cli pipeline runs watch --release <name> --commit $(git rev-parse HEAD)
```

這些指令是平台 API 的包裝，**自己不持有任何規則**。**檢核跑在平台上**，因為最有價值的檢查 —— apiserver 自己對每個渲染出來的 CR 做的 CEL、pattern 與必填驗證 —— 需要一座叢集，而**任何 client 都不會拿到叢集憑證**。所以迴圈是：改 chart、用 `asgard-cli gate` 檢查本機能檢查的、推上去、把 plan 讀回來。

一次 run 有六步：`checkout → lint → variables → render_dry_run → review → apply`。`review` 是一個人。它之前的全部是 plan，而 plan 找到的一切都在它的報告裡。

**一次推送什麼都沒產生時，`pipeline deliveries` 是唯一會解釋自己的地方。** 一個從未被建立的 run 不會留下自己的紀錄，所以當一個 tag 看起來被忽略了，原因只在那裡。

## 發佈

發佈由 [GoReleaser](https://goreleaser.com) 驅動。推一個 tag 觸發 `.github/workflows/release.yml`：

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

一次執行產出 linux / darwin / windows × amd64 / arm64 的 binary、`.deb` / `.rpm` / `.apk` 套件與 checksum，全部掛在 GitHub Release 上。changelog 從 conventional commit（`feat:`、`fix:`）自動分組。

本機驗證而不發佈任何東西（產出落在 `dist/`）：

```bash
goreleaser check
goreleaser release --snapshot --clean --skip=publish
```

### private，以及「通道關掉」是從它推導出來的

每一個 release 都是 private，因為 repo 是。在使用者還是內部的階段，那**不是要繞過的限制而是重點**，但它決定了安裝路徑 —— 所以 release notes 帶的是一行 `gh release download` 而不是 `brew install`。

`.goreleaser.yaml` 底部有現成的 **Homebrew tap** 與 **Scoop bucket** 區塊，而它們維持註解狀態。tap 或 bucket 是第二個 repo，安裝的人必須讀得到；把它也設成 private，等於每個使用者都要對一個需要憑證的 repo 跑 `brew tap`，比它要取代的那一行下載指令更麻煩，而服務的對象又比 tap 存在的理由更小。使用者定為內部限定，2026-09-06。**如果將來要公開，開啟它們是第一件要回頭做的事** —— 區塊與它們的前置條件（`HOMEBREW_TAP_TOKEN`、`SCOOP_BUCKET_TOKEN`）都留在原地。

其他值得知道的事：

- **CGO**：build 以 `CGO_ENABLED=0` 執行，讓交叉編譯裝得進一台 runner。引入一個 cgo 相依（sqlite 那一類）就得換成 zig cc 或每個平台各一台 runner。
- **macOS 簽章**：binary 沒有簽章。那之所以撐得住，**正是因為安裝路徑是 CLI 下載** —— `gh` 與 `curl` 不會設 `com.apple.quarantine`，瀏覽器會。把一個 release URL 丟給人點才是會壞的情境，而 `anchore/quill` 是那時候的答案。
- **每個 PR 都 build 一次 release**：`ci.yml` 的 `build` job 跑 `goreleaser release --snapshot --clean --skip=publish`，所以設定或交叉編譯壞掉會在它變成一個失敗的 tag 之前被抓到。
