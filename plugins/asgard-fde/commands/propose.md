---
description: Build the deck that goes in front of the customer.
---

Before writing a single slide, load the deck skill - it holds the rules that were
learned the expensive way:

```
.agents/skills/proposal-deck/SKILL.md
```

The three that get broken most:

- **No CR kinds, no `ts-` / `sl-` / `dc-` prefixes, no internal tool names** -
  not in the text, and not inside a screenshot either
- **Say what it does, in their words.** Not what it is made of
- **A question page is not a list of everything you do not know.** It is the
  short list of things whose answer changes what gets built

What the deck is proposing has to be decided before it is written, and the
material for that is:

```bash
`.agents/skills/asgard-platform/guide/requirements.md`    # what the interview should have settled
asgard-cli size <shape>                 # how big the thing actually is
`.agents/skills/asgard-platform/wiki/screenshots.md`             # which picture answers which question
```

**Open any screenshot before it goes on a slide.** Nothing records when they were
captured, five of the 39 indexed have been opened, and two of those five needed
cropping. A screenshot of a console the customer will not recognise is worse than
no screenshot.
