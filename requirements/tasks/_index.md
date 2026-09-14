# Task Index

**一個 task 是一份開工前要先寫好的實作規格**,是對 living spec 的 delta。每一列一個
`TASK-*.md`,由 `asgard-cli task` 開與移動狀態。

**這裡不是客戶的要求。** 那些是 request(`../requests/`),用客戶自己的話記。
**這裡也不是行為的事實來源** —— 一份 task 到 `done` 之後就不再被讀,所以行為變了要把
delta 套回 `docs/spec/asgard/`,那份才是下一個人會讀的。

**這張表由 `asgard-cli` 維護。**

## Config
- SPEC_DIR: `docs/spec/asgard/` (living spec — 開工前先讀,收工後套回 delta)
- DECISIONS: `docs/decisions/YYYY-MM-DD-<topic>.md` (dated、不可變)
- CHART_DIR: `projects/<project>/chart/app`
- VALUES: declared as keys in `.asgard-pipeline.yaml`, set on the platform per release (`asgard-cli pipeline variables set`); the namespace comes from the platform project the release binds
`asgard-cli verify <project>`
- Mode: spec

## Next Task

Nothing in flight. The next task usually opens once a request has a target project.

## Task Queue

| Task ID | Title | Owner | Complexity | Status |
|---------|-------|-------|-----------|--------|

## Conventions

- **開工前先讀 living spec**:`docs/spec/asgard/<module>.md` 描述系統**現在**的行為,
  一份 task spec 是對它的 delta。收工時把 delta 套回去 —— 行為變了卻沒有進 living spec 的 task
  **不算 done**;順帶定案的決策要寫一份 `docs/decisions/YYYY-MM-DD-<topic>.md`。
  四層模型見 `docs/README.md`。
- Task specs: `TASK-xxx-short-name.md` — one file per task, sections per
  `docs/spec-driven-development.md`(`Meta` / `1) Requirements` / `2) Design` /
  `3) Implementation Tasks` / `4) Execution Log / Change Log`)。
- Complexity: `S` / `M` / `L`.
- Status: `draft` → `ready` → `in-progress` → `done`(不使用 `in_progress`)。
- **狀態轉換用 CLI,不要手改兩個地方。** `asgard-cli task add`、`task ready`、
  `task start`、`task done` 會同時改這張表、spec 的 `Meta` 狀態、以及 spec 的
  Execution Log(帶日期),並更新上面的 Next Task。手改只會改到其中一個,
  而兩邊不一致時沒人分得出哪一邊是真的。
- **Task ID 全域唯一,跨 project 共用同一條號。** 開新 task 前先看這張表的最大號,
  不要只看自己 project 的 —— 兩個 project 在不同分支上同時開號會撞號。
