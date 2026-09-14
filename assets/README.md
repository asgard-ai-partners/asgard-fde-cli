# assets — runtime 會用到的資源

**這裡放的是「跑起來之後」會用到的素材。** 最常見的用途是被 Syncer 在 runtime 同步進
SourceSet(又稱 Drive)的 volume —— 有時是單純的知識素材,有時是 agent 要裝備的技能。

**但那只是最常見的用法,不是唯一的用法。** 任何 runtime 要用到的資源素材都可以放這裡。

> **它以前叫 `common/`,因為當時那些素材是各個 project 共用的。** 共不共用是某一個
> engagement 當下的處境,不是判準 —— 判準是**這東西是不是給跑起來之後用的**。

```
assets/
  skills/<skill>/SKILL.md    # runtime 技能(見下)
```

## 這裡**不是**什麼

反面比正面重要,因為每一條都有人踩過:

- **不是 chart。** CR 在 `projects/<project>/chart/`。這裡放的是**素材**,不是資源定義。
- **不是給 coding agent 讀的。** 在這個 repo 裡工作的 agent 讀 `.agents/skills/`、
  `AGENTS.md` 與 `docs/`。**放錯邊的代價是把 design-time 的東西同步進 runtime**,
  或反過來讓部署後的 agent 拿不到它需要的知識。
- **不放值,也不放密鑰。** 值宣告在 `.asgard-pipeline.yaml`、填在 Platform 上;密鑰走
  cluster Secret。這裡的東西會**原封不動**進到 volume 裡,任何寫在檔案裡的憑證等於發佈。
- **不是暫存區。** 進到這裡的東西會被同步出去,所以「先放這裡再說」沒有這個選項。

## `skills/` — runtime 技能

一個技能一個目錄,裡面一份 `SKILL.md`。詳細寫法見
[`skills/README.md`](skills/README.md)。

**它是空的,而且是刻意先立在這裡的。** runtime 技能這個需求太常見,所以先把位置定下來,
免得每個 engagement 各自發明一個路徑 —— 而路徑一旦不同,`SkillSet` 的 `searchPaths`
就每個 repo 長得不一樣。

## 怎麼生效

```
commit → Syncer 把這個 repo 同步進 SourceSet 的 volume
       → SkillSet.searchPaths 從裡面挑
       → Agent 或 SandboxBlueprint 綁定 SkillSet
```

`searchPaths` 指的是 **volume 裡的路徑**,而 Syncer 把整個 repo clone 到 `git/` 底下,
所以路徑帶著這個 repo 的目錄名:

```yaml
spec:
  sourceSetName: ss-sk-<name>
  searchPaths:
    - git/assets/skills/<skill>      # 對:一個技能一條
  # - git/assets/skills              # 錯:指父目錄會解析成零個技能
```

> **`searchPaths` 要逐個技能目錄列。** 平台以「一個 searchPath = 一個技能目錄」解析,
> 指向父目錄會**靜靜地**解析出零個技能 —— 平台不會報錯,agent 只是沒有技能。
> 這也是為什麼改動這個目錄的名字時,chart 裡的那串字串要一起改。

每個 SkillSet 有**自己專屬的一組 SourceSet + Syncer**(三件一套寫在同一個
`templates/skill_set/<name>.yaml` 裡,以 `---` 分隔)。平台的
`POST /v1/skill-set/from-git` 一次建出三件並互綁,Platform UI 照這個結構顯示;
多個 SkillSet 共用一個 SourceSet 的話,UI 找不到「這個技能集的 git 設定」。

> Anthropic 公開技能(pdf / docx / pptx / xlsx)由各 project 自己的 `syn-sk-base`
> 從 `anthropics/skills` 另外同步,**不放在本 repo**。

## 部署一律走 CD

**不要從本機 `helm upgrade`。** Syncer 以 `revision: {{ .Chart.AppVersion }}` 同步本 repo,
而只有 CI 會把發布 tag 蓋進 `appVersion`。本機安裝會把 `Chart.yaml` 裡的佔位版本當成
git ref 寫進 Syncer,那個 ref 不存在 → Syncer 每次都同步失敗。

所以 `asgard-cli render <release>` **只渲染、沒有安裝路徑**。
