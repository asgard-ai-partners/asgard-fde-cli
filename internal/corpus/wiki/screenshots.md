# Screenshots, and which situation each one is for

Every screenshot of this platform lives in the product documentation. There are
none in this binary and none in a customer repository, and that is deliberate:
an image copied into an engagement goes stale where nobody is looking, and a
folder of PNGs teaches nobody anything. What is worth carrying is **which picture
answers which question**, which is this page.

Both places serve the same file:

    https://docs.asgard-ai.com/img/docs/<path>          the published site
    <clone of asgard-docs>/static/img/docs/<path>       a checkout

The paths below are that `<path>`. The captions are the documentation's own alt
text, in zh-TW; **the console itself is in English**, so a deck in either
language can use them and write its own caption.

**Open one before it goes on a slide.** Nothing records when any of these was
captured, so a form that has since changed looks exactly like a current one - and
a stale screen in front of the customer who uses that screen daily costs more
than having no picture.

Three things to look for, and the third is the one that catches people:

  - **a credential** - a hostname, an account, a connection string
  - **another customer** - a name, a logo, anything recognisable
  - **our own implementation nouns, inside the picture.** A CR name, a `ts-` /
    `sl-` / `dc-` / `ss-` prefix, an internal tool name, the admin console's own
    navigation. The `proposal-deck` skill bans these in the deck's text and the
    ban applies just as much to a screenshot of them - **and several images below
    contain them.** They are marked, with what to crop

Cropping is normal and usually costs nothing: the part that carries the argument
is rarely the part carrying the noun.

## For a proposal: what it looks like to use

**This is the set most proposals need and most decks miss**, because the obvious
instinct is to show the console - which is our side, not theirs. These four are
one continuous story from a user's question to a completed action, and they carry
the governance gate, which is the hardest thing to explain in words.

| path | what it shows |
|---|---|
| `sindri-retail-stockout-transfer/00-available-agents.png` | the hub with five agents, before anything is asked |
| `sindri-retail-stockout-transfer/01-allocation-answer.png` | a real answer: current state, available quantity, three options to choose from |
| `sindri-retail-stockout-transfer/02-approval-gate.png` | **the approval dialog** - the tool it wants to call, and nothing running until a person allows it. **CROP FIRST:** the dialog names `工具集:ts-wms` and `工具:create_transfer_order`. Keep the title line and the three buttons, drop the two rows between them - the argument is entirely in "it stopped and asked", which survives the crop |
| `sindri-retail-stockout-transfer/03-execution-result.png` | after approval: what was created, and the resulting stock forecast |

The third one is worth a slide on its own whenever the requirement has a write in
it. `../usecase/write-path.md` is the argument; this is the picture of it.

For a customer-service shape rather than an internal one:

| path | what it shows |
|---|---|
| `retail-ai-customer-service/21-flow-agent-canvas.png` | the five-node default flow: Entry, Init, Agent, Listen, and a push on Agent's Failure branch |
| `retail-ai-customer-service/23-agent-node-open.png` | the prompt deciding what a **logged-out** visitor is told, versus a logged-in one |

And the one that shows Description doing its job in public:

| path | what it shows |
|---|---|
| `sindri-home/home-available-agents.png` | each card's line is that agent's Description, as routing text |

## For a handover: the setup path

The order these belong in is [`setup-path.md`](../wiki/setup-path.md). Use them when the
audience is the people who will operate the thing - never in a proposal.

| step | path | what it shows |
|---|---|---|
| 1 | `settings/data-source/data-source-create-provider-open.png` | the Provider list open: nine, all databases. **The screen that settles where an API key does not go** |
| 1 | `settings/data-source/data-source-create.png` | the form, with Test Connection before Save |
| 1 | `settings/connection/connection-create-options-open.png` | Connection types, grouped by syncer / For Loader / For Trigger |
| 2 | `mcp-servers/new-menu.png` | From Workflow against From Existing - the fork that decides how much work step 2 is |
| 2 | `mcp-servers/mcp-from-existing.png` | Transport Type, Command, Environment Variables - where an API key actually lives |
| 2 | `skillsets/new-menu.png` | From Scratch against From Git |
| 2 | `drive/drive-syncer-wizard-step1.png` | the five syncer sources: Google Drive, OneDrive, Git, Web Crawler, Data Source |
| 2 | `drive/drive-detail-context-index.png` | Context Index, and the switch that turns it on |
| 2 | `data-insight-semantic-model/table-setting-select.png` | picking tables, with live Data Preview |
| 2 | `data-insight-semantic-model/modeling-built.png` | ten tables modelled, one expanded to Description / Primary Key / Columns |
| 3 | `agent-hub-managed-agent/list.png` | the five built-in templates - show this when they think it starts from nothing |
| 3 | `agent-hub-managed-agent/create.png` | **the one screen most handovers need**: Alias, Description, and Prompt as Persona / Task / Context / Format, with the live preview |
| 3 | `agent-hub-managed-agent/resources-section.png` | mounting what step 2 produced |
| 5 | `agent-hub-flow-agent/create.png` | Name, Description, Advanced Sandbox Settings |

A worked example of step 3 filled in, rather than empty:

| path | what it shows |
|---|---|
| `retail-stockout-transfer/12-allocation-top.png` | a real agent's Alias, Description and four Prompt sections |
| `retail-stockout-transfer/13-allocation-resources.png` | its sample questions and everything it mounts |
| `retail-stockout-transfer/01-managed-agents.png` | five agents, all `Released` |

## For explaining a capability

| subject | path |
|---|---|
| the governance gate, as a list | `retail-stockout-transfer/04-automation-tools.png` - seven approval-gated tools |
| what a gated tool declares | `retail-stockout-transfer/11-transfer-order-tool.png` - its input schema |
| how many systems one agent reads | `retail-stockout-transfer/03-semantic-models.png` - CRM, ERP, e-commerce, POS, supplier, OLAP, WMS. **CROP FIRST:** the whole Odin console navigation is down the left - Agent Hub, MCP Servers, Skillsets, Plugins, Settings. Keep the card area only |
| conversational analysis | `mimir-thread/answer.png`, `mimir-thread/chart-result.png` |
| a dashboard that gets shared | `mimir-dashboard/share-dashboard.png` |
| teaching the model the business's own questions | `mimir-knowledge/add-question-sql-pair.png` |
| a Knowledge Base loading documents | `knowledge-base-knowledge/auto-load-form.png` |
| scheduled runs | `automation-trigger/new-trigger.png` - Schedule, Timezone, Model, Prompt |

## For a cost or governance conversation

| subject | path |
|---|---|
| what a bill looks like | `fehu/fehu-billing-reports-list.png`, `fehu/fehu-bill-detail-service.png` |
| usage against quota | `fehu/fehu-quota-limits-top.png`, `fehu/fehu-usage-reports.png` |
| who can reach which product | `console-product-permission/agent-hub-accounts.png` |
| workspace membership | `console-workspace-settings/workspace-accounts.png` |

**These four are the ones to be careful with.** Permissions and billing are where
a customer most easily reads a screenshot as a promise about what can be
restricted. `../wiki/platform-unknowns.md` still has scope control as
unanswered: the Console decides who reaches a product, and what a caller can
touch **after** that is the part no source settles. A permissions screenshot on a
slide about limiting access is the overstatement this tool warns about most.

## The rest of what the documentation uses

The sections above are the curated set - grouped by the situation they answer,
and some of them opened. **This is everything else the documentation references**,
added 2026-09-03 so that "not in this page" stops meaning "does not
exist". They are grouped by product rather than by situation, because nobody has
worked out which situation each one answers, and pretending otherwise would make
this look more considered than it is.

**Every caption below is the documentation's own alt text, and none of these has
been opened.** Everything the page says at the top applies with more force here:
no capture date exists, a form that has since changed looks exactly like a
current one, and the three things to check - a credential, another customer, our
own implementation nouns - have not been checked for any of them.

**A caption saying a screen is empty is useful.** Several read 空狀態 - the empty
state. Those are the wrong picture for a proposal, where the customer is trying
to imagine the thing full, and the right one for a handover, where somebody needs
to recognise the screen they are about to fill in.

### Odin: building the thing

| path | the documentation's caption |
|---|---|
| `agent-hub-configuration/add-model.png` | Add Model 視窗，提示沒有可加入的 Completion Model |
| `agent-hub-configuration/global-directory.png` | Configuration 頁面的 Global Directory 頁籤，目前是空的檔案總管介面 |
| `agent-hub-configuration/list.png` | Configuration 頁面的 Models 頁籤，列出四個內建模型 |
| `agent-hub-flow-agent/list.png` | Flow Agent 列表，目前沒有任何 agent |
| `applications-customized-integration/list.png` | Customized Integration 頁面，目前沒有已建立的整合 |
| `applications-overview/list.png` | Data Insight & Agent Hub 頁面，目前沒有已發布的應用 |
| `automation-api/list.png` | API 列表，目前沒有任何 API |
| `automation-api/new-api.png` | New API 建立表單，包含 Name 與 Description 兩個欄位 |
| `automation-trigger/list.png` | Trigger 列表，目前沒有任何 Trigger |
| `mcp-servers/list.png` | MCP Servers 列表，目前是空的 |
| `mcp-servers/mcp-from-workflow.png` | Add new MCP server from workflow 表單，只有 Name 與 Description 兩個欄位 |
| `plugins/list.png` | Plugins 空狀態畫面 |
| `settings/completion-model/completion-model-create.png` | New Completion Model 表單，Name／Model Name／API Key 標為必填 |
| `settings/completion-model/completion-model-list.png` | Completion Model 列表，內建四個分級模型加上兩個自行加入的具名模型 |
| `settings/connection/connection-create.png` | New Connection 對話框，Name 跟 Type 兩個欄位 |
| `settings/connection/connection-list.png` | Connection 空狀態畫面，尚未建立任何連線 |
| `settings/data-source/data-source-list.png` | Data Source 列表，目前有一個名為 Retail DW (RDS) 的 PostgreSQL 連線 |
| `settings/embedding-model/embedding-model-create.png` | New Embedding Model 表單，預設供應商是 Azure OpenAI Embedding Model，六個欄位皆為必填 |
| `settings/embedding-model/embedding-model-list.png` | Embedding Model 列表，內建一個 Balanced 分級模型 |
| `skillsets/list.png` | Skillsets 空狀態畫面 |
| `skillsets/skillset-detail.png` | Skillset 詳情頁的 Files 頁籤，目前是空的檔案總管，右上角有 Open in Advance Editor 連結 |
| `skillsets/skillset-from-scratch.png` | New Skillsets / From Scratch 視窗，包含 Name 與 Search Paths 兩個欄位 |

### Mimir: asking the data

| path | the documentation's caption |
|---|---|
| `data-insight-semantic-model/create.png` | New Semantic Model 建立表單第一步 Basic Information，包含 Name、Completion Model、Effort、Language、Timezone、Data Source 等欄位 |
| `data-insight-semantic-model/list.png` | Semantic Model 空狀態畫面 |
| `data-insight-semantic-model/modeling-empty.png` | Modeling 步驟剛進入時，10 張 Table 都還在 Waiting to be modeled，右側對話框列出三個快速動作 |
| `mimir-dashboard/add-view-wizard.png` | Add View 精靈：Select Views / Preview / Complete 三個步驟 |
| `mimir-dashboard/dashboard-empty.png` | 剛建立、還沒有任何圖表的儀表板 |
| `mimir-dashboard/name-your-dashboard.png` | Name your dashboard 對話框：輸入儀表板名稱 |
| `mimir-data-model/data-model-panel.png` | Data Model 面板：Summary、Model Source、資料表、欄位、關聯與度量，以及 Preview Data |
| `mimir-data-model/preview-data.png` | Data Preview 展開後：十筆實際資料，含 SKU、門市、等待會員數與高階會員數 |
| `mimir-intro/project-home.png` | Mimir 專案首頁：側邊的 Threads／Views／Dashboards，中央的提問框與建議問題 |
| `mimir-knowledge/add-instruction.png` | Add an instruction 表單：Global 與 Matched 兩種生效範圍，以及規則內容欄位 |
| `mimir-knowledge/instructions.png` | 已儲存的 Instruction：規則內容、生效範圍與建立時間 |
| `mimir-knowledge/question-sql-pairs.png` | 已儲存的 Question-SQL pair 清單：問題、SQL 與建立時間 |
| `mimir-knowledge/row-menu.png` | 清單列的 Edit 與 Delete 選單 |
| `mimir-retail-stockout-demand/save-as-view.png` | Save current view：把呈現方式選為長條圖，連同 SQL 一起存下來 |
| `mimir-retail-stockout-demand/surplus-chart.png` | 長條圖：各門市現有庫存由多到少排列，藍色為高於安全庫存、橘色為低於安全庫存的信義旗艦店 |
| `mimir-retail-stockout-demand/surplus-tables.png` | Mimir 的回答：可調撥門市與缺口門市兩張表，以及調撥建議 |
| `mimir-thread/sql-editor.png` | 提問框旁的 SQL 按鈕開啟的 SQL 編輯器 |
| `mimir-thread/view-data-panel.png` | View data 面板：Data Preview 與 SQL Query 分頁，以及 Pin to dashboard、Create View 兩個動作 |
| `mimir-view/convert-to-dynamic-view.png` | Convert to Dynamic View 面板：Prompt 描述欄位、Convert 動作，以及下方的圖表對照 |
| `mimir-view/save-current-view.png` | Save current view 對話框：Name、Description、SQL 與 Visualization 呈現方式 |
| `mimir-view/view-detail.png` | View 詳細頁：View SQL、Download、Add to Dashboard、Convert to Dynamic View 四個動作 |
| `mimir-view/visualization-options.png` | Visualization 選單：Table 之外多了這次回答產生的水平長條圖 |

### Sindri: using the agent

| path | the documentation's caption |
|---|---|
| `sindri-directory/directory-chats-files.png` | 一個已有內容的 Directory：Chats／Files 分頁，以及一則剛開的新對話 |
| `sindri-directory/directory-empty-state.png` | 尚未建立任何 Directory 時的側邊欄狀態 |
| `sindri-my-chat/chat-menu.png` | 對話項目的選單：Share／Rename／Delete |
| `sindri-my-chat/files-panel-directory-awake.png` | 喚醒後的 Files 面板：Sandbox 名稱、路徑，以及目前的檔案清單 |
| `sindri-my-chat/files-panel-no-sandbox.png` | Sandbox 被回收時的 Files 面板：No sandbox is running，提供 Wake a sandbox 按鈕 |
| `sindri-my-chat/panels-menu.png` | Panels 選單：Files／Quick Menu 兩個可以個別開關的面板 |
| `sindri-my-chat/quick-menu-empty.png` | Quick Menu 面板的空狀態：目前 Project 沒有 Agent 設定快速選單 |
| `sindri-project/project-switcher.png` | Project 切換器：可搜尋的 Project 清單，列出同一 Workspace 底下不同業態的 Project |
| `sindri-settings-general/general-settings.png` | 一般設定面板：帳號、語言、口語、外觀主題 |
| `sindri-settings-personalization/personalization.png` | 個人化設定面板：Base Tone、Characteristics、Memory |

### Knowledge: Drive and Knowledge Base

| path | the documentation's caption |
|---|---|
| `drive/drive-create-step1.png` | New Drive 表單，Basic Information 加上選填的 Reference Paths |
| `drive/drive-detail-files.png` | Drive 詳情頁 Files 分頁，唯讀模式，目錄目前是空的 |
| `drive/drive-detail-settings-tab.png` | Settings 分頁，Basic Information 表單 |
| `drive/drive-detail-syncers.png` | Syncers 分頁，目前是空的，表格欄位是 Folder / Source / Active / Status / Schedule |
| `drive/drive-list.png` | Drive 列表，目前有一個名為 drive 的 Drive |
| `knowledge-base-knowledge/auto-load-crawler-step2.png` | Crawler 來源的排程設定，包含 Name、Frequency（Daily/Weekly）與 Repeat at 時間 |
| `knowledge-base-knowledge/create.png` | New Knowledge Base 視窗，包含 Name 與 Alias Name 兩個欄位 |
| `knowledge-base-knowledge/list.png` | Knowledge Base 空狀態畫面 |
| `knowledge-base-knowledge/manual-upload-add-menu.png` | Knowledge Base 詳情頁，Add 下拉選單提供 Manual Upload 與 Auto Load 兩種來源 |
| `knowledge-base-knowledge/manual-upload-processing.png` | Manual Upload 上傳 CSV 的 Processing 步驟，顯示欄位預覽、Skip Header 開關與 Identifier 欄位 |

### Console: who can see what

| path | the documentation's caption |
|---|---|
| `console-intro/my-products.png` | My Products 首頁：三個區塊與五張產品卡 |
| `console-product-permission/billing-accounts.png` | Billing Accounts：Collaborators 分頁與 Roles 側欄項目 |
| `console-product-permission/data-insight-accounts.png` | Data Insight Accounts：Manage Accounts in 選擇器、Purchase Named User，以及 Shared with 欄位 |
| `console-product-permission/studio-platform-accounts.png` | Studio 的 Platform Accounts，側欄可以看到 Platform 與 Project 兩層 |
| `console-workspace-settings/edit-workspace-modal.png` | Rename Workspace 開啟的 Edit Workspace 對話框，name 欄位已帶入現有名稱 |
| `manage-workspace/add-new-project.png` | Add New Project 視窗，只有 Name of Project 一個欄位 |
| `manage-workspace/management-console-redirect.png` | 點擊 Workspace Member 或 Rename Workspace 後出現的提示，導向 Management Console |
| `manage-workspace/project-list.png` | My Projects 頁面，列出目前 Workspace 底下的所有 Project 卡片，右上角顯示 Used : 19 / 300 用量與 Add New Project 按鈕 |
| `manage-workspace/workspace-menu.png` | 左下角點擊 Workspace 展開的選單，顯示 Workspace Member 與 Rename Workspace |
| `manage-workspace/workspace-overview.png` | 工作空間 Overview 頁面，Analysis 區塊列出 Requests per second、Request Duration、Token Usage、Total Message、Data Source 等圖表 |

The rest are reachable by path under `https://docs.asgard-ai.com/img/docs/`;
what is known about them is below, under `## What has no screenshot`.

## What has no screenshot

Not everything does, and assuming otherwise wastes a search:

  - the chart, the CRs, the cluster - there is no UI for any of it
  - **LINE: three images.** `integration/LINE` carries them hash-named rather
    than in a topic directory, which is why a survey of the topic folders
    misses them. Counted and opened 2026-09-03:

        /img/docs/60e2492a66bf.png    the prerequisites
        /img/docs/f3563b5553dc.png    the integration dialog, LINE selected,
                                      Channel Secret and Access Token fields
        /img/docs/7a4b3a60eb28.png    **LINE's own Developers Console**, the
                                      Webhook settings panel: Webhook URL,
                                      Verify, and the Use webhook toggle

    **The third one is usable.** It is not Asgard's console at all - it is
    LINE's, so Odin's redesign does not
    date it, and it carries a placeholder `https://example.com/webhook` rather
    than anyone's real endpoint. It is exactly the picture for explaining the
    half of the setup the customer does themselves: paste the Webhook URL we
    give them, press Verify, turn Use webhook on.

    The second one is the screen a customer would want to see. **It is the old
    console** - a left nav of Overview / Workflows / Knowledge / Environment /
    Apps, and a card dated 2024/09/20 - and today's Odin has none of those.
    `integration.md` already flags those pages as possibly stale for the
    same reason.

    So the answer for the first two is still not to use them, but for a
    checkable reason rather than because none exist: **a screenshot of a console
    the customer will not recognise is worse than no screenshot**, and this is
    the exact failure this page warns about, in the one place somebody would
    most want to ignore it.

    What to do instead, and it is better than it sounds: describe the exchange
    in their words - "your customer types their order number in LINE, it replies
    with the repair status, and says so plainly when it cannot find it". Slide 5
    is an interaction, not a screen. A vendor's own documentation screenshot is
    not an option either: it is someone else's product surface in our proposal.

    **If somebody captures a current one, this is the highest-value screenshot
    missing** - it is the channel a Taiwanese customer asks about first.

  - anything about handoff, pausing, or per-user counters, because the platform
    does not have them - see [`integration.md`](../wiki/integration.md)

## Sources

- Every path here was read off the `![alt](/img/docs/...)` references in
  asgard-docs' own `.mdx` pages, so a caption is the documentation's rather than
  a guess about what the file contains
  - asgard-docs `f00e0ee`
- **Checked** 2026-09-02, opened rather than listed:
  `settings/data-source/data-source-create-provider-open.png`,
  `agent-hub-managed-agent/create.png`, `sindri-home/home-available-agents.png`,
  and the two marked CROP FIRST above - which is how the crops came to be
  marked. An engagement found the first of them by downloading it, after this
  page had recommended it as the single most useful image here

**Counted 2026-09-03. How it was measured, because the number here has been
wrong twice:** the repository holds **679** `.png` files, but only **168** of
them are referenced by any documentation page - the rest are orphaned assets
from an older structure, and 204 of those sit under `user-guide/`, which no
longer has pages. So the denominator that means anything is 168, not 679.

This page names **116** of that 168. Every one of them resolves at `https://docs.asgard-ai.com/img/docs/<path>`, exists in the
repository, and is referenced by a live page, so nothing here points at a
missing file or at an orphan. **Coverage is 69% of the images the documentation
uses**, and the paragraph that used to sit here said "roughly a hundred", which
was a guess about a number seven times too small.

**That number moves whenever somebody adds an image to this page**, which is an
ordinary edit and not one anybody remembers to recount after. So it is computed:
asgard-fde-cli's `go run ./hack counts` intersects the paths named here with the
images a live page references at `f00e0ee`, and adding one moves the figure or
fails the check.

**The 52 not named here carry no alt text**, and most are hash-named files from
an older documentation structure. There is nothing useful to say about them
from a listing - the LINE three under `## What has no screenshot` are the
exception, and they were found by reading the page that uses them rather than by
listing files. So **if the picture you want is not here, the next move is to
read the docs page for that feature**, not to conclude the platform has no
screen for it.

**Unchecked:** everything else here is described by its alt text, not by having
been looked at, and **no capture date exists for any of them**. Only the ones
the Checked block names have been opened, and two of those needed cropping - so
assume an unopened one does too, rather than that the marked ones are the only
ones. The grouping
into situations is this tool's judgement rather than anything the documentation
says.
