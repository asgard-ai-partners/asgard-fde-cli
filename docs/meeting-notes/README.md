# docs/meeting-notes/ — 原始輸入

收斂迴路的**最上游**:raw 會議筆記、逐字稿、issue 討論摘要。內容未經整理,
它是寫決議紀錄時的材料。

> 這一層是「原始輸入」,**不是決議**。決議寫在 [`../decisions/`](../decisions/);
> 當下行為寫在 [`../spec/`](../spec/)。三層分清楚,逐字稿才不會污染 spec。

## 命名

`YYYY-MM-DD-<topic>.md`,例如 `2026-09-02-mail-endpoint-review.md`。
複製 [`_template.md`](_template.md) 開始。

**這場會議有附件的時候,改成同名資料夾**,筆記本身叫 `README.md`:

```
docs/meeting-notes/2026-09-02-phase-1-proposal/
  README.md      筆記
  deck.pdf       當天給客戶的簡報
  deck.md        簡報原始檔
```

給客戶的提案簡報就是這樣存的 —— 它是某一天講出去的東西,跟這一層的性質一樣:
**只增不改**。客戶手上那份不會跟著我們改,所以要改就開新的一場日期。
怎麼做一份見 `proposal-deck` skill(`.agents/skills/`)。

## 規則

- raw 筆記**只增不改**(它是歷史輸入,保留原樣)。要修正理解 → 在決議紀錄那一層處理。
- **這一層可以省略。** 來源是 GitHub issue 或使用者的一句直接指示時,直接在決議紀錄裡引用即可,
  不必為了流程而抄一份筆記。**它存在是為了容納「討論很長、結論還沒定」的那種輸入。**

## 收斂流程

```
docs/meeting-notes/YYYY-MM-DD-<topic>.md   （raw 筆記貼這裡)
   └─ 收斂 ─→ docs/decisions/YYYY-MM-DD-<topic>.md   (決議紀錄)
                └─ 套用 ─→ docs/spec/asgard/<module>.md   (改寫 living spec)
                            └─→ requirements/tasks/TASK-xxx-*.md    (要動 chart 才開)
```
