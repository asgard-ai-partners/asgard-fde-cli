# docs/ — 文件目錄索引

這裡放**規格與決策**。實作用的工作單位在 [`requirements/`](../requirements/),背景資料在
[`references/`](../references/),而「怎麼在這個 repo 動手」在 [`AGENTS.md`](../AGENTS.md)。

> **最常找的兩個**:現在講定的行為 → [`spec/asgard/`](spec/asgard/);
> 當初為什麼這樣定 → [`decisions/`](decisions/)。

## 四層心智模型

這個 repo 的文件分四層,**時態不同**是它們分家的唯一理由:

```
docs/meeting-notes/YYYY-MM-DD-<topic>.md   raw 原始輸入(會議 / issue 討論,收斂前)
        │ 收斂
        ▼
docs/decisions/YYYY-MM-DD-<topic>.md       dated 決議紀錄(不可變,回答「當初為什麼」)
        │ 套用 delta
        ▼
docs/spec/asgard/<module>.md     ★ living spec「當下唯一」(持續改寫,只反映現狀)
        │ 拆成工作單位
        ▼
requirements/tasks/TASK-xxx-*.md           可執行任務規格(EARS 驗收條件 + 實作步驟 + 執行紀錄)
        │ 實作
        ▼
projects/<project>/chart/                  Helm chart 裡的 Asgard CR
```

| 層 | 時態 | 誰是它的讀者 |
|---|---|---|
| `meeting-notes/` | 收斂**前**的某個時間點 | 想知道原始討論長什麼樣的人 |
| `decisions/` | 某時間點的快照,**不可變** | 想知道「當初為什麼這樣定」的人 |
| `spec/` | **永遠是現狀** | 想知道「這套系統現在的行為是什麼」的人 —— 包含 agent |
| `requirements/tasks/` | 一件工作的生命週期(`draft`→`done`) | 要動手做這件事的 agent |

**living spec 只寫現狀,不累積歷史。** 「從 A 改成 B」這件事屬於決議紀錄;living spec 只寫 B。
反過來,決議紀錄一旦寫下就不改 —— 改變主意就寫新的一份。

## 未解的問題

[`open-questions.md`](open-questions.md) 是**還沒有答案**的問題的家 —— 卡住決策的、
或會改變設計形狀的。四層都不收它:決議紀錄寫的是已定案的,任務規格的開放問題會隨
任務 `done` 一起消失,living spec 只寫現狀。

**接手這個 repo 時先讀它**,`asgard-cli question` 也會把還開著的問題印出來。
答案到手時把該列移到「Answered」並補上決議紀錄的連結 —— 不要刪掉那一列,
「這件事曾經沒有答案」本身就解釋了設計為什麼長這樣。

## 子資料夾

| 路徑 | 是什麼 | 何時看 |
|------|--------|--------|
| [`spec/<slug>/`](spec/) | ★ **living spec** —— 當下唯一的系統行為規格 | 「現在這套系統講定的行為是什麼」 |
| [`decisions/`](decisions/) | dated **決議紀錄**(不可變) | 「當初為什麼這樣決定」 |
| [`meeting-notes/`](meeting-notes/) | raw 會議 / issue 輸入(收斂前) | 「這次討論的原始內容」 |

## 單篇文件

| 檔案 | 內容 |
|------|------|
| [`spec-driven-development.md`](spec-driven-development.md) | SDD 規範:什麼時候要寫規格、狀態流、任務規格格式、驗收閘門 |

## 與 `AGENTS.md` 的分界

兩份文件都是「事實來源」,但**回答的是不同問題**,不要把內容搬來搬去:

| | [`AGENTS.md`](../AGENTS.md) | `docs/spec/` |
|---|---|---|
| 回答 | **怎麼改這個 repo** | **這套系統的行為是什麼** |
| 內容 | 版面、模板慣例、Helm/平台踩雷、驗收閘門怎麼跑、CRD 欄位陷阱 | 能力邊界、資料口徑、路由規則、狀態機、安全論證 |
| 讀者 | 在這個 repo 工作的 coding agent | 任何要理解或驗證這套系統的人 |
| 一句話判準 | 「不看它就會把 chart 寫壞」 | 「不看它就會把行為做錯」 |

同一件事只寫一邊。例:`workflow-set-id` 這個 label **要怎麼寫**在 `AGENTS.md`;
**沒有它使用者會看到什麼**在 living spec 的平台後設資料模組。

## 收斂迴路(怎麼維護)

新的決策進來時,照這個順序走一輪 —— 詳細步驟見
[`.agents/skills/spec-workflow/SKILL.md`](../.agents/skills/spec-workflow/SKILL.md) 的
「Living Spec 收斂迴路」:

1. raw 討論 → `docs/meeting-notes/YYYY-MM-DD-<topic>.md`(可省略;來源是 issue 就直接引用連結)
2. 寫決議 → `docs/decisions/YYYY-MM-DD-<topic>.md`(複製 `_decision-template.md`)
3. 套用 delta → 改寫 `docs/spec/asgard/<module>.md`,並更新該 slug README 的模組索引
4. 要動 chart → 開 `requirements/tasks/TASK-xxx-*.md`,登記進 `requirements/tasks/_index.md`
5. 實作完 → 跑驗收閘門(`AGENTS.md` → Acceptance Gate),回填任務規格的執行紀錄

`asgard-cli check` 會驗這層結構:模組索引與實際檔案一致、決議 /
會議紀錄的檔名合法、範本與 README 齊全、`docs/` 裡的相對連結都指得到東西。
