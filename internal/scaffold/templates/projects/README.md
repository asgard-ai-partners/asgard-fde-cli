# projects

每個 project 一個目錄:一份 Helm chart,部署到一個 namespace。

```
projects/<slug>/
  deploy.yaml                    # 哪些 env 要部署 + 每個 env 的 namespace 與 values
  chart/
    app/templates/<kind>/        # 每個 CR 種類一個目錄,一個 CR 一個檔
    values-dev.yaml
    values-prod.yaml
```

用 `asgard-cli project add <slug> --env dev` 登記,再跑 `asgard-cli scaffold`
產生上面的骨架。**不要手動建目錄** —— 根 README 的 project 表由 scaffold 依
`.asgard-config.json` 維護,而 `asgard-cli check` 會比對那張表與這裡實際的
資料夾,兩邊對不上就是紅的。

## 怎麼切

切分依**受眾**,不依整齊。內部可驗證的呼叫端與匿名訪客需要不同的入口形狀與讀取路徑,
那兩者無法共用,所以他們不能是同一個 project。詳見 `asgard-cli guide projects`
與根目錄 `AGENTS.md` 的「The three decisions that get answered wrong」。
