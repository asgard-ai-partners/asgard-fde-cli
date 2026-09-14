---
name: plain-chinese
description: Use when writing or reviewing any 繁體中文 that a customer will read - a proposal slide, a reply to their question, a CR display name, a cube description, an agent's prompt. Removes the sentence shapes that mark a document as machine-written even when every word is defensible. 適用於「這段中文讀起來像 AI 寫的」「潤稿」「改得像人寫的」。
version: 1.0.0-template
alwaysApply: false
---

# Plain Chinese

A customer who reads vendor documents recognises machine-written Chinese in
about four seconds. What they conclude is not "this is AI". It is **"nobody here
spent time on us"**, and that costs more than any wording.

The register rule - prefer a checkable sentence to an adjective - catches empty
words. It does not catch **shapes**: sentence patterns and section endings that
give a document away while every individual word survives inspection. Those are
what this skill is for.

Adapted from `writing-humanizer` (github.com/shyuan/writing-humanizer), which is
written for zh-TW essays. Most of it transfers to this work. One rule inverts,
and the inversion is in "Where this applies" below - read that before applying
anything else, because getting it wrong damages a deck.

## The two that are absolute

    never    不僅……更是……   /   這不僅僅是……,而是……
    never    a page, section or paragraph that ends by announcing its own
             importance

`不僅……更是……` is the single most recognisable AI sentence in Chinese. There is
no context in which it survives. State the thing directly.

The second is **意義蓋章** - closing with 具有重要意義、奠定了基礎、
發揮關鍵作用、為……提供了重要框架. **Test it by deleting the sentence.** If what
remains loses no information, it never had any:

    no    導入語意層後,客服可直接查詢庫存,為後續智慧化轉型奠定重要基礎。
    yes   導入語意層後,客服不用再開 ERP 查庫存。下一階段才會碰到出貨單。

One 意義蓋章 reads as enthusiasm. Every section carrying one is the tell: the
model is filling the position where a conclusion goes, because it has no
conclusion.

## The shapes

**四字標籤 + 冒號的排比清單.** `**流程斷點**:客服要開四個系統`. The bold
four-character label carries no information; it exists so the bullets look
parallel. The sentence after the colon is the content. Delete the label. If
three bullets then say the same thing, they were one bullet.

**句內關鍵詞排比粗體.** Bolding two to four matching nouns inside one running
sentence - `**經濟制裁**、**貿易優惠**、**投資協議**`. Emphasis comes from
position and structure. Bold marks **the one thing you want read first**, and
four bold phrases on a page mark nothing.

**元論述 / 導讀宣告.** 本節將說明、接下來我們來看、了解了這點就能明白. Delete
the announcement and say the thing. A slide has a title; a document has a
heading. Neither needs to introduce itself.

**升華結尾.** 共同邁向智慧化的未來、期待與貴公司攜手、讓我們一起. Finish on what
happens next week and who does it.

**排比金句.** 每一次……都是……,每一個……都是……. A slogan you invented is not a
finding, and repeating it later does not make it one.

**三段式反射.** Three parallel items is a tell, because the model reaches for
three whether reality has three or not. **When you have written three, check
whether reality has two, or four**, and write that number. If it genuinely is
three, keep it - the rule is to stop producing three by reflex, not to ban the
number.

**破折號.** The em dash is a machine habit in Chinese. A comma or a full stop
does the same work without the salesroom cadence.

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
| a proposal or discovery deck | yes, **except one rule** - see below | `proposal-deck` owns the rest of how a deck reads |
| a reply to a customer's own question | in full | `docs/open-questions.md`, the "What the customer asked us" table |
| the covering mail, meeting notes they will see | in full | prose, so every rule holds |
| a CR's `<kind>-name` display annotation | the vocabulary, not the shapes | it is a noun phrase, not a sentence |
| a cube, dimension or measure `description` | in full, and it matters twice | a model reads these to choose what to query, so 至關重要 there is not just noise - it is noise the agent has to guess past |
| a deployed agent's prompt | in full | the agent writes in the register the prompt is written in |
| a decision record | in full | it is read years later by somebody working out what was agreed, and 為後續發展奠定基礎 helps them with none of it |
| a task spec, a request record | in full | read by whoever picks the work up, which is often not you |
| meeting notes | your half of them | what the customer said stays verbatim - that is the record. What you wrote around it is yours |

**Everything an engagement writes is covered.** There is no internal-only
exemption: a decision record outlives the engagement, and a task spec is read by
whoever picks the work up. The only thing that is never rewritten is **what the
customer themselves said** - in a meeting note, in section 1 of a request, in a
quoted question. Those are the record, and editing them destroys the thing they
are for.

**The rule that inverts.** The source skill's strongest claim is that a Chinese
document written as headings and bullet lists is AI, and that論說文 should be
prose. **That is right for prose and wrong for a deck.** A slide is legitimately
a list; an essay pasted onto a slide is unusable, and an agent applying that rule
to a deck will destroy it.

Keep the rule one level down, where it holds everywhere: **each bullet must be a
sentence with content**, not a decorative label. The failure is never that the
page is a list. It is that the list says nothing.

## The second pass

**Do not correct these while writing.** You will produce some regardless - they
are what the model reaches for - and hunting them mid-draft costs the argument,
which is the thing that actually matters.

When the document reads correctly end to end, make one pass that asks a single
question:

    這一頁哪裡看得出來是機器寫的?

Read it as the customer, who has seen twenty vendor documents this year.
**List what you find before changing anything.** Naming them is what makes the
pass honest; editing as you go lets you stop early and believe you were thorough.

### The thoughts that end the pass early

| the thought | the answer |
|---|---|
| 「這頁已經夠自然了」 | The first pass always misses some. That is what the second pass is. |
| 「這裡三項是合理的」 | Check whether reality has three. Usually it has two. |
| 「破折號在這裡讀起來比較有力」 | That force is the tell. Use a comma. |
| 「客戶自己的文件就是這樣寫的」 | Their **words** stay - 報修單 stays 報修單. Their **sentence patterns** do not. You are answering their document, not imitating it. |
| 「這句對仗工整,拿來當標題很好」 | A slogan you invented is not a finding. |
| 「加一句總結,讀者比較清楚」 | That is 意義蓋章. The page already said it. |
| 「改太多會偏離原意」 | The AI pattern is not the original meaning. It is what the model added. |

## What this skill will not do for you

**It cannot make a thin page substantial.** If a capability's page is short
because the customer's own document is short, the emptiness is true and says
something useful - they have not worked it out either, and the meeting can start
there. Rewriting it into fuller-sounding Chinese is the failure this skill is
supposed to prevent, arriving by a different route.

**It cannot make an unchecked claim safe.** 全面提升效率 and 改善對帳流程 are
both wrong if nobody has checked whether the agent can reach the ledger. Plain
language makes a false claim easier to catch, which is a reason to prefer it and
not a substitute for checking.

**Checked:** 2026-09-04 - **no platform claims to check.** Every rule here is
about 繁體中文 prose, and the sentence shapes it bans are checkable by reading
what they produce.

**Unchecked:** that these are the shapes that mark a document as machine-written
to **this** audience. They were collected from what a customer read and reacted
to in one engagement, in one industry, in Taiwan. Nothing corroborates them and
nothing makes them a style guide; a reader whose customer speaks differently
should treat the list as evidence rather than as rules.
