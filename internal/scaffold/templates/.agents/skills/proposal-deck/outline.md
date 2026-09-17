# The discovery deck - outline

> The worked example beside `discovery-deck.html`, and the worked example of an
> outline. Every deck gets one of these, written before a slide is edited.
> `SKILL.md` step 5 says what a discovery deck is for; this says what each page
> of this one does and what it rests on.

## The argument, in two beats

    一  每個情境要先拿到什麼才做得起來 - 一個子題一頁,左欄是問句本身
    二  問完之後那一頁的產出是什麼,包括查不到的時候它會怎麼回答

**A discovery deck's beats are the customer's, not ours.** The two above are
read off their own document and their own wording; a proposal's beats would be
conclusions, and `SKILL.md` step 5 is where that inverts.

## The pages

| # | page | what it does | what it rests on |
|---|---|---|---|
| 1 | 這份骨架怎麼用 | how to read the rest, and the one rule about the 產出 column | this skill |
| 2 | cover | who and when | - |
| 3 | 情境 1 - AI 客服與工單系統整合 | the customer's own flow, and their own list of what to verify | the customer's document |
| 4 | 工單系統資料查詢 | what to obtain before it can be built, and what it answers | their question, verbatim |
| 5 | 辨識客戶身分 | the binding question, and how much self-service already exists | their document, item 2 |
| 6 | 自動建立報修案件 | the write, and who reviews it | their flow |
| 7 | 設備操作 | the second scenario's write, and what must not be touched | their operator's words |
| 8 | 這一輪會交付什麼 | what lands regardless of how many scenarios are adopted | our scope |

## What each page rests on

**3, 4, 5, 6 carry the customer's own naming**, including 「工單系統」 and
「聊天管道」 rather than the CR kinds behind them. That is the discovery deck's
rule and not a stylistic choice - `SKILL.md` step 5 has why a proposal's
implementation nouns are wrong here too.

**4, 5 and 7 open with `.said`**, which takes somebody's actual words and
nothing else. Three more lines on these pages were sentences about the slide
rather than quotes; they are plain body text now, and question 0 of
`## Before you send it` is what found them.

**Every 產出 column ends with a failure behaviour** - 查不到就說查不到, and page
1 says why: in a customer-service scenario that line persuades more than any
capability does. A page whose 產出 column has no failure line is not finished.

**The questions are the left column and they are the page.** A sub-topic page is
one question per `h3`, blue, with the grey supporting line under it only where
there is one. The heading is the question itself rather than a label for it.

## Screenshots, and when to retake them

This deck has none, which is the discovery shape working: it asks rather than
shows. A deck that does carry them records, per screenshot, the exact path it
was taken from and what is visible on it - **a platform release ages every
screen, and an out-of-date screen looks exactly like a current one.**

## Not done

- **No proposal is worked here.** `references/design.md` says the same about its
  own class table: the split between a working set and a specialised one is
  evidence about a discovery deck, and nothing states how a proposal differs
  page by page.
- **The 產出 wording is one engagement's.** It reads well and no second
  engagement has been held against it.
