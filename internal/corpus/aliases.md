# The index: what a customer says, and what to search for

**This is an index, not a page about the platform.** It lives beside `pages/`
rather than in it, next to `index.md` and `log.md`, for the same reason those do
- it is the wiki's own bookkeeping. Keeping it inside the corpus made it compete
with what it points at: it lists every alias, so it was unusually likely to be
the one document carrying every term of a translated query, and a search for 電商
returned this table instead of `wiki/taiwan-channels.md`.

`asgard-cli find` reads both tables below and applies them to a query before
searching, printing what it actually searched for. Read it with
`asgard-cli wiki --aliases`.

## How a row gets here

**Every row is a term somebody searched for.** Add one when a search of yours
came back empty and the subject turned out to exist under another name; that is
the only test, and **a row nobody has needed is a guess.** `asgard-cli find`
records the queries that landed nowhere - `asgard-cli reading --misses` reads
them back, and that list is where new rows come from.

Point a row at a page where one page answers it: the pointer is checked by
`asgard-cli audit-material --links`, which a bare word would not be.

## What a customer says, in the words this material uses

The corpus is English and a customer conversation is not, so a query taken from
what somebody actually said lands on nothing and the reader concludes the
subject is missing rather than named differently.

**These replace the word.** A Chinese term appears nowhere in an English corpus,
so keeping it in the query would only add a term that lands nowhere.

| they said | search for |
|---|---|
| 原始碼 | source, repository |
| 程式碼 | source |
| 文件 | documentation, sources |
| 電商 | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |
| 庫存 | inventory, stock |
| 訂單 | order |
| 客服 | customer service, help desk - and `usecase/chat-channel.md` |
| 權限 | permission, scope, console - and `wiki/platform-unknowns.md` P1 |
| 稽核 | audit, logging - and `wiki/platform-unknowns.md` P2 |
| 報表 | dashboard, report, view - and `usecase/mimir-dashboard.md` |
| 儀表板 | dashboard - and `usecase/mimir-dashboard.md` |
| 知識庫 | knowledge, drive, context index |
| 白名單 | allowlist, outbound - and `wiki/operations.md` |
| 網路 | network, reachable, allowlist |
| 排程 | schedule, trigger, cron |
| 核准 | approval, consent, requestConsent - and `usecase/write-path.md` |
| 寫入 | write - and `usecase/write-path.md` |
| 投影片 | deck, slides - and the proposal-deck design-time skill |
| 簡報 | deck, slides - and the proposal-deck design-time skill |
| 截圖 | screenshot - and `wiki/screenshots.md` |
| 語意層 | semantic layer |
| 資料庫 | database, DataConnector |
| 技能 | skill, SkillSet |
| 價格 | cost, billing - and `wiki/fehu.md` |
| 計費 | billing - and `wiki/fehu.md` |

## Names the material covers

Somebody searched the reference deployments for each of these and wrote down what
came back, so the row routes to a page that has the answer.

**These are added to a query rather than replacing it**, which is the difference
from the table above: the name may be written verbatim in a page - SHOPLINE is -
and that page is the best answer there is. Replacing the name would throw it away.

| they said | also search for |
|---|---|
| shopline | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |
| shopee | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |
| 蝦皮 | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |
| momo | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |
| pchome | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |
| coupang | commerce, marketplace, channel - and `wiki/taiwan-channels.md` |

## Names it only routes

**Nothing here names these, and `asgard-cli find` says so.** The row exists
because somebody searched for it and the search was recorded; what it routes to
is **the shape the thing belongs to**, which is what this material actually has.

This is the distinction the table above does not carry on its own, and it is the
one that matters: a row that routes reads exactly like a row that answers, and a
reader who cannot tell them apart takes results about a shape as results about a
product. Moving a row up means somebody did the search and recorded the answer -
that is what `## Names the material covers` means, and nothing else.

| they said | also search for |
|---|---|
| 綠界 | payment gateway, requestConsent, approval, external api |
| ecpay | payment gateway, requestConsent, approval, external api |
| 藍新 | payment gateway, requestConsent, approval, external api |
| newebpay | payment gateway, requestConsent, approval, external api |
| 第三方支付 | payment gateway, requestConsent, approval, external api |
