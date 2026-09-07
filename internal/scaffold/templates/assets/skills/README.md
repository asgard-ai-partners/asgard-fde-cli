# assets/skills — RUNTIME 技能

一個技能一個目錄,裡面一份 `SKILL.md`。

## 這裡是什麼

**給部署後的 Asgard agent 用的技能。** 它們會被同步進 SourceSet 的 volume、由
`SkillSet` CR 挑選、再掛到 Agent 上。

**這個目錄是空的,而且是刻意先立在這裡的。** runtime 技能這個需求太常見,所以先把位置
定下來,免得每個 engagement 各自發明一個路徑 —— 路徑一旦不同,`SkillSet` 的
`searchPaths` 就每個 repo 長得不一樣。

## 這裡**不是**什麼

**不是給在這個 repo 工作的 coding agent 用的。** 那些在 `.agents/skills/`,而兩個目錄
只差一個字:

|  | `assets/skills/` | `.agents/skills/` |
|---|---|---|
| 誰讀 | **部署後**的 Asgard agent | **在這個 repo 工作**的 coding agent |
| 怎麼到讀者手上 | Syncer 同步進 volume → SkillSet 挑選 → Agent 綁定 | 就在磁碟上,agent 自己載入 |
| 放什麼 | 客戶的領域知識、業務口徑 | 怎麼建 CR、怎麼查資料庫、怎麼寫中文 |
| 會進平台嗎 | **會** | **永不** |

**放錯邊的代價**:把 design-time 的操作指引同步進 runtime(部署後的 agent 拿到一份教它
改這個 repo 的文件),或反過來讓部署後的 agent 拿不到它真正需要的領域知識。

**寫技能前先確定要的是哪一種。** 部署後 agent 需要的領域知識放這裡;開發流程的操作指引放
`.agents/skills/`。分界見根目錄 `README.md`。

## 怎麼生效

```
commit → SourceSet 同步這個 repo → SkillSet.searchPaths 選取 → Agent 或 SandboxBlueprint 綁定
```

`searchPaths` 要**逐個技能目錄**列出,不能只列父目錄 —— 平台以「一個 searchPath =
一個技能目錄」解析,指向父目錄會靜靜地解析出零個技能。

```yaml
searchPaths:
  - git/assets/skills/<skill>      # 對,一個技能一條
  # - git/assets/skills            # 錯,會解析成零個技能
```

## 寫一個技能

`SKILL.md` 需要 frontmatter,而且 `name` 必須等於資料夾名
(`asgard-cli check` 會驗):

```markdown
---
name: <skill-directory-name>
description: 什麼時候該用這個技能。agent 是靠這句話決定要不要載入它的。
---

# <技能名>

領域知識放這裡 —— cube 名與欄位名傳達不了的東西:狀態碼的實際語義、
跨系統的實體對應、業務口徑、哪些資料表不可信。
```

> **不要放空骨架。** 一個只有標題、沒有內容的技能,結果是 agent 被告知「別用這個」,
> 那不如不要建。等到真的有領域知識要寫的時候再開。
