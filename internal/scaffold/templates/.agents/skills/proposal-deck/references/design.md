# The deck's design language

**This file is read while laying out, when the sentences do not exist yet.** The
check that catches what gets written into the layout afterwards is question 0 of
`## Before you send it` in `../SKILL.md`, and it is run per sentence rather than
per slide.

**What a deck looks like, so that this skill does not depend on a typesetting
skill being installed.** The rules below are the ones a proposal or discovery
deck uses; the content decisions are `SKILL.md`'s and are not repeated here -
with one exception, `.said`, whose content rule sits on the class
below because that is the file open while the line is being written.

**Taken from the `kami` skill's design system, read 2026-09-14.** Only the
deck-relevant half: the invariants, the palette, the type scale, the slide
classes and the print rules. Its landing-page, resume and portfolio material is
not here, and neither is its authoring process - `SKILL.md` says why taking that
process damages a discovery deck.

## The invariants

Each has a cost. Think before overriding one.

1. Page background is parchment `#f5f4ed`, **never pure white**
2. One accent, ink-blue `#1B365D`, and **no second chromatic colour**
3. Every grey is warm-toned. No cool blue-greys
4. English: serif for headlines and body. Chinese: serif headlines, sans body.
   Sans only for labels, scene lines and meta, in both
5. Serif weight is locked at 500. **No bold**
6. Line-height: tight headlines 1.1-1.3, dense body 1.4-1.45, reading body
   1.5-1.55
7. Letter-spacing: Chinese body 0.3pt, English body 0. Tracking only on short
   labels and overlines
8. A tag background is a solid hex, **never `rgba`** - the renderer draws a
   double rectangle otherwise
9. Depth is a ring or a whisper shadow. **No hard drop shadows**
10. **No italic.** Print templates carry none

**Ink-blue covers at most 5% of the surface.** More than that is ornament rather
than restraint, which is the whole argument of the palette.

## The palette

    --parchment    #f5f4ed    the page
    --ivory        #faf9f5    a lifted surface - a card is its fill, not an outline
    --border       #e8e6dc
    --border-soft  #e5e3d8
    --brand        #1B365D    the only chromatic colour
    --brand-tint   #EEF2F7
    --tag-bg       #E4ECF5
    --near-black   #141413    body text
    --dark-warm    #3d3d3a
    --charcoal     #4d4c48
    --olive        #504e49    a lead line under a title
    --stone        #6b6a64    scene lines and meta

## The page and the type on it

Default page is **280mm x 158mm**, set in `@page` and on `.slide` together.
297x167 for a little more room, 338x190 for heavy content. Changing it is a
decision for the whole deck, never for one slide.

    body        13pt / 1.55, letter-spacing 0.3pt   (CJK needs the tracking)

### The working set

`../discovery-deck.html` is this system applied, and it is built from these and
nothing else:

    h2          24pt, weight 500, margin-bottom 14pt   the page title
    h3          15pt, weight 500, in --brand           a section heading, and on a
                                                       question page the question
    .scene      9.5pt mono, tracking 2pt, --stone      the scenario this slide
                                                       belongs to, in the
                                                       customer's own naming.
                                                       Identical on every slide
                                                       of one scenario
    .said       12pt, --olive                          what somebody SAID or
                                                       WROTE, verbatim and in
                                                       quotation marks - the
                                                       customer's question, a
                                                       line from their document
    .mi         11pt, 8pt vertical padding             a line item, and the workhorse
    .c2         two columns, 22pt gap                  what we need beside what it
                                                       unlocks, on one page
    table.data  11pt rows, first cell in --brand       a term and its explanation

### The specialised set

Every class below is in the stylesheet and **used nowhere in that deck**, so
reaching for one is a decision rather than a default. The split is the deck's to
state and not this file's - recount it rather than trust the heading:

    grep -o 'class="[^"]*"' ../discovery-deck.html | tr ' ' '\n' | sort | uniq -c

    .mt  16pt   module title, paired with .ml
    .ml  24pt   the large letter prefix, in --brand
    .ms  7.5pt  mono module sub-label, with a bottom rule
    .mb  11pt   module body
    .mc  9.5pt  cadence note, with a top rule
    .t2x2       a 2x2 grid as a table - CSS grid misaligns the rows

**There is no slot for a line about the slide, and that is deliberate.** This
system used to carry three - a label above the title, a line under it, and a
callout pinned to the foot - each named for WHERE it sat rather than for what it
held, and a slot named for a position accepts any sentence. All three are gone.
What is left is named for its content: `.scene` takes the scenario's name,
`.said` takes somebody's actual words. A sentence that is neither has no class
to sit in and reads as the loose body text it is.

**Removing the slots is not the cure and was never claimed to be.** The same
sentence, once deleted from a callout, came back as a caption, and then as the
third column of a table with no class on it at all. The rule below is what
catches it, and the classes only stop the easiest three landings.

**Most slides need no caption at all, and empty beats composed.** A screenshot
that needs a sentence explaining what it shows is a screenshot on the wrong
slide. The commonest caption written here restates the table directly above it,
which is the title's argument arriving a third time.

**When one is earned, a caption on a slide is not a caption in a document.** 9pt
is readable on paper and is not from the back of a room: a slide caption is
roughly 2.67x the print size.

**Sizes land on the scale, never between its steps.** A reader cannot tell 13.5
from 14, so the difference registers as noise rather than as hierarchy.

Spacing is a 4pt base unit: 4-5pt inside a tag, 8-10pt inside a component,
16-20pt between components, 24-32pt around a section title.

## The content rules the layout enforces

| rule | detail |
|---|---|
| no divider slides | name the scenario with `.scene` instead; it saves one slide per section. **The name is the scenario's**, so it is identical on every slide of one - a `.scene` line that changes per slide has stopped being a name and become a caption |
| no CJK parentheses | `（...）` becomes `·` or a comma |
| the ghost deck test | read only the titles, in order: they must tell the argument. **This is a proposal rule and inverts for a discovery deck** - `SKILL.md` step 5 |
| a title names, never counts | a quantity spends the line without naming a subject - `SKILL.md` step 1, and a discovery deck keeps the customer's heading even where it counts |
| the title is the argument | a sentence that restates the title, or argues for it, is filler - **wherever it sits**. A table column, a closing paragraph, a caption and a parenthesis are not exemptions, and neither is the sentence reading to you as argument rather than as narration: that judgement is the one that fails, because a writer classifies their own reasoning as argument. The test needs no judgement - **cover the title and read the sentence. If it still tells you something the slide did not already show, it stays; otherwise it goes.** Given a correction about such a sentence, delete it; moving it to another container is the same sentence |
| one evidence shape | one proof form per slide - a chart, a table, a screenshot, a quote. Split a slide with two |
| one line per bullet | trim until it fits. Never let one wrap. **Also a proposal rule** |
| empty space over 50% | a draft defect: merge with a neighbour, cut the slide, or add evidence that earns the space. Shrinking the page is a last resort and applies to the whole deck |
| empty space 25-50% | fine as it stands. Add a supporting line only if there is a FACT to add; **never pad with prose**. There is no callout class to reach for any more, and an empty-looking slide is the one place a line can never come from - anything written to fill it is written about the slide rather than about the subject |

**The last three are the ones that damage a discovery deck**, and a layout
skill's own density checks will enforce them anyway. `SKILL.md` decides which
apply; this file only says what they are.

## The content contract

`slides.json` beside this file is the schema, **and it is the proposal's**:
`layout` is one of `cover`, `chapter`, `content`, `quote`, `metrics`, `close`;
a title is a declarative claim rather than a topic label; `items` is three to
five, each one line. **There is no caption or callout field**, because the
schema is what a writer fills in and a field is a slot. The middle two are the rules the table above marks as the proposal's, so
a discovery deck is not written against this schema at all - `SKILL.md` step 5
says why it keeps no structured copy of its content.

## Rendering it

**HTML to PDF, not a presentation format.** A pptx passed through a converter
loses CJK weight, tracking and glyph spacing; an HTML renderer embeds the fonts
exactly. `../discovery-deck.html` is a working deck in plain HTML with no separate
content file - it is the shortest way to see all of the above applied, and the
only thing that says how often each class is right.

**PDF is what gets handed over.** Produce an editable file only when the
customer has said they want to edit it.
