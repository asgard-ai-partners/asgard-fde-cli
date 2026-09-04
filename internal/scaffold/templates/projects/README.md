# projects

每個 project 一個目錄:一份 Helm chart。

```
projects/<slug>/
  chart/
    app/templates/<kind>/        # 每個 CR 種類一個目錄,一個 CR 一個檔
    app/values.yaml              # 只放預設值。實際的值在 Platform 上
```

**這裡沒有部署目標。** 一份 chart 部署到哪裡,是根目錄 `.asgard-pipeline.yaml` 裡的
Release 決定的:一個 Release 綁一個 Platform Project(= 一個 namespace),由一條 tag 或
branch 規則觸發,一份 chart 可以有好幾個 Release。namespace 從 `.Values.asgard.namespace`
來,不寫在 chart 裡。

用 `asgard-cli project add <slug>` 登記,再跑 `asgard-cli scaffold`
產生上面的骨架。**不要手動建目錄** —— 根 README 的 project 表由 scaffold 依
`.asgard-config.json` 維護,而 `asgard-cli check` 會比對那張表與這裡實際的
資料夾,兩邊對不上就是紅的。

## 怎麼切

切分依**受眾**,不依整齊。內部可驗證的呼叫端與匿名訪客需要不同的入口形狀與讀取路徑,
那兩者無法共用,所以他們不能是同一個 project。詳見 `asgard-cli guide projects`
與根目錄 `AGENTS.md` 的「The three decisions that get answered wrong」。
