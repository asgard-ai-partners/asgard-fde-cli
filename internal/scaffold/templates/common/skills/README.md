# common/skills — RUNTIME skills

一個技能一個目錄,裡面一份 `SKILL.md`。**這些是給部署後的 Asgard agent 用的**,
不是給在這個 repo 工作的 coding agent 用的(那些在 `.agents/skills/`,兩者不可混用,
分界見根目錄 `README.md`)。

## 怎麼生效

```
commit → SourceSet 同步這個 repo → SkillSet.searchPaths 選取 → Agent 或 SandboxBlueprint 綁定
```

`searchPaths` 要**逐個技能目錄**列出,不能只列父目錄 —— 平台以「一個 searchPath =
一個技能目錄」解析,指向父目錄會靜靜地解析出零個技能。

```yaml
searchPaths:
  - git/common/skills/<skill>      # 對,一個技能一條
  # - git/common/skills            # 錯,會解析成零個技能
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
