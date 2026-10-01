---
name: plain-chinese
description: Use when writing or reviewing any 繁體中文 that a customer will read - a proposal slide, a reply to their question, a CR display name, a cube description, an agent's prompt. Removes the sentence shapes that mark a document as machine-written even when every word is defensible, and the opposite failure that no register pass catches - a sentence whose subject is only in the writer's head. 適用於「這段中文讀起來像 AI 寫的」「潤稿」「改得像人寫的」「這個在說什麼」「給主詞」。
version: 1.0.0-template
alwaysApply: false
---

# Plain Chinese

A customer who reads vendor documents recognises machine-written Chinese in
about four seconds, and concludes that nobody spent time on them.

The register rule - prefer a checkable sentence to an adjective - catches empty
words. It does not catch shapes: sentence patterns and section endings that
give a document away while every individual word survives inspection. This skill
covers those shapes. The register rule itself is stated in
`.agents/skills/proposal-deck/SKILL.md` under "The register", and holds for
everything on the table below, not only a deck.

Adapted from `writing-humanizer` (github.com/shyuan/writing-humanizer), which is
written for zh-TW essays. Most of it transfers to this work. One rule inverts;
read "Where this applies" below before applying anything else, because applying
that rule to a deck damages it.

## The two that are absolute

    never    不僅……更是……   /   這不僅僅是……,而是……
    never    a page, section or paragraph that ends by announcing its own
             importance

`不僅……更是……` is the most recognisable AI sentence in Chinese. Do not use it
anywhere. State the thing directly.

The second is 意義蓋章 - closing with 具有重要意義、奠定了基礎、
發揮關鍵作用、為……提供了重要框架. Test it by deleting the sentence. If the
text loses no information, delete it:

    no    導入語意層後,客服可直接查詢庫存,為後續智慧化轉型奠定重要基礎。
    yes   導入語意層後,客服不用再開 ERP 查庫存。下一階段才會碰到出貨單。

One 意義蓋章 reads as enthusiasm. When every section ends with one, the model is
filling the position where a conclusion goes because it has no conclusion.

## The sentence with a hole in it

Every other rule here prohibits a shape that appears in the text. This one is
about a noun that is missing, so none of the other rules catches it, and it
usually takes a second reader to notice:

    記下是哪一個              哪一個什麼?
    這個檔案裡只有這兩個編號    什麼編號?
    那段描述是模型唯一的判斷依據  哪段描述?
    失敗有四種,只有一種是「查無資料」  四種什麼?

The test, which you can run on your own text:

> 每個句子的主詞,在同一頁上找得到嗎?
> 「這個」「那段」「它們」、一個代名詞、一個沒有名詞的數字 - 它的名詞必須在
> 同一頁出現。只在你腦子裡,或只在上一頁,就把名詞寫出來。

It does not read as machine-written, so a register pass does not catch it. It
reads as somebody talking to themselves: the noun was in your head an hour ago,
so the sentence reads fluently to you, but the reader does not have the noun.

It matters more on a slide than in prose, because a slide has no preceding
paragraph to carry the antecedent. Every line on a slide is read without context.

Naming the nouns does not remove the count. Told that
「12 個 CR,14 次部署,4 個坑」was vague, the fix offered was
「一支查詢 API、552 份手冊、兩份 FAQ」- which names the nouns and still counts
them. Both rules apply, and satisfying one does not satisfy the other.

## The shapes

**四字標籤 + 冒號的排比清單.** `**流程斷點**:客服要開四個系統`. The bold
four-character label carries no information; it exists so the bullets look
parallel. The sentence after the colon is the content. Delete the label. If
three bullets then say the same thing, they were one bullet.

**句內關鍵詞排比粗體.** Bolding two to four matching nouns inside one running
sentence - `**經濟制裁**、**貿易優惠**、**投資協議**`. Emphasis comes from
position and structure. Use bold only for the one thing you want read first;
four bold phrases on a page mark nothing.

**元論述 / 導讀宣告.** 本節將說明、接下來我們來看、了解了這點就能明白. Delete
the announcement and say the thing. A slide has a title; a document has a
heading. Neither needs to introduce itself.

**升華結尾.** 共同邁向智慧化的未來、期待與貴公司攜手、讓我們一起. Finish on what
happens next week and who does it.

**排比金句.** 每一次……都是……,每一個……都是……. Do not write slogans; state the
finding.

**三段式反射.** The model writes three parallel items whether reality has three
or not. When you have written three, check whether reality has two, or four,
and write that number. If it really is three, keep it.

**破折號.** The em dash is a machine habit in Chinese. Use a comma or a full
stop instead.

**繫動詞迴避.** 「這項功能扮演著關鍵的角色」when the sentence is 「這個功能會
自動對帳」. Say what it does; the elaborate construction is the model avoiding
a plain 是 or a plain verb.

## The vocabulary

| avoid | say |
|---|---|
| 至關重要 | 重要,或說明為什麼 |
| 深入探討 | 討論、看 |
| 賦能 | 讓……可以 |
| 無縫銜接 | 順暢,或說明中間少了哪一步 |
| 提升 / 增強 | 改善,或給數字 |
| 打造 / 構建 | 做、建 |
| 全面 / 智慧化 | delete, or name the thing it now does |
| 大幅 | give the number, or delete |
| 複雜性 / 錯綜複雜 | 複雜、麻煩 |
| 展現 / 彰顯 | 讓人看到,或直接說 |
| 在當今……的時代 | delete, start with the subject |
| 隨著……的快速發展 | delete |
| 值得一提的是 | delete, say it |
| 眾所周知 | delete |
| 綜上所述 / 總而言之 | delete |
| ……的重要性不言而喻 | say why it matters, concretely |

## Where this applies, and the one place it inverts

| surface | applies | note |
|---|---|---|
| a proposal or discovery deck | yes, except one rule - see below | `.agents/skills/proposal-deck/SKILL.md` owns the rest of how a deck reads, including the register rule this page builds on |
| a reply to a customer's own question | in full | `docs/open-questions.md`, the "What the customer asked us" table |
| the covering mail, meeting notes they will see | in full | prose, so every rule holds |
| a CR's `<kind>-name` display annotation | the vocabulary, not the shapes | it is a noun phrase, not a sentence |
| a cube, dimension or measure `description` | in full, and it matters more here | a model reads these to choose what to query, so a word like 至關重要 there is noise the agent has to guess past |
| a deployed agent's prompt | in full | the agent writes in the register the prompt is written in |
| a decision record | in full | it is read years later by somebody working out what was agreed, and 為後續發展奠定基礎 helps them with none of it |
| a task spec, a request record | in full | read by whoever picks the work up, which is often not you |
| meeting notes | your half of them | what the customer said stays verbatim - that is the record. What you wrote around it is yours |

Everything an engagement writes is covered. There is no internal-only
exemption: a decision record outlives the engagement, and a task spec is read by
whoever picks the work up. The only thing never rewritten is what the customer
themselves said - in a meeting note, in section 1 of a request, in a quoted
question. Those are the record; do not edit them.

The rule that inverts: the source skill says that a Chinese document written as
headings and bullet lists is AI, and that論說文 should be prose. That holds for
prose, not for a deck. A slide is legitimately a list; an essay pasted onto a
slide is unusable. Do not apply that rule to a deck.

What does apply everywhere, decks included: each bullet must be a sentence with
content, not a decorative label. A list is fine; a list that says nothing is not.

## The second pass

Do not correct these while writing. You will produce some anyway, and hunting
them mid-draft takes attention away from the argument.

When the document reads correctly end to end, make one pass that asks two
separate questions. Ask both; asking only the first is the usual mistake:

    這一頁哪裡看得出來是機器寫的?
    這一頁的每個句子,主詞都在同一頁上找得到嗎?

The first catches a sentence that says too much in a recognisable way. The
second catches one that says too little. That kind does not look
machine-written, so the first question does not find it.

Read it as the customer, who has seen twenty vendor documents this year.
List what you find before changing anything. If you edit as you go, it is easy
to stop early and assume you were thorough.

### The thoughts that end the pass early

| the thought | the answer |
|---|---|
| 「這頁已經夠自然了」 | The first pass always misses some. Run the second pass. |
| 「這裡三項是合理的」 | Check whether reality has three. Usually it has two. |
| 「破折號在這裡讀起來比較有力」 | That cadence is what marks it as machine-written. Use a comma. |
| 「客戶自己的文件就是這樣寫的」 | Keep their words - 報修單 stays 報修單. Do not copy their sentence patterns. You are answering their document, not imitating it. |
| 「這句對仗工整,拿來當標題很好」 | Use a finding as the title, not a slogan. |
| 「加一句總結,讀者比較清楚」 | That is 意義蓋章. The page already said it. |
| 「改太多會偏離原意」 | The AI pattern was added by the model; removing it does not change the meaning. |

## What this skill will not do for you

It cannot make a thin page substantial. If a capability's page is short
because the customer's own document is short, that is accurate and useful: they
have not worked it out either, and the meeting can start there. Do not pad it
into fuller-sounding Chinese.

It cannot make an unchecked claim safe. 全面提升效率 and 改善對帳流程 are
both wrong if nobody has checked whether the agent can reach the ledger. Plain
language makes a false claim easier to catch, but it does not replace checking.

**Checked:** no platform claims to check. Every rule here is
about 繁體中文 prose, and the sentence shapes it bans are checkable by reading
what they produce.

**Unchecked:** that these are the shapes that mark a document as machine-written
to this audience. They were collected from one customer's reactions, in one
industry, in Taiwan, and a reader whose customer speaks differently should treat
the list as evidence rather than as rules.
