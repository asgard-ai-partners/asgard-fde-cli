# The deck's design language

**What a deck looks like, so that this skill does not depend on a typesetting
skill being installed.** The rules below are the ones a proposal or discovery
deck uses; the content decisions are `SKILL.md`'s and are not repeated here -
with one exception, the `.co` callout, whose content rule sits on the class
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
   Sans only for labels, eyebrows and meta, in both
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
    --stone        #6b6a64    eyebrows and meta

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
    .eyebrow    9.5pt mono, tracking 2pt, --stone      the stable section label
    .lead       12pt, --olive                          one line under the title
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
    .co  11pt   the bottom callout, pinned 12mm from the foot

**`.co` is the one class in this file whose misuse produces bad content rather
than bad layout**, so its content rule travels with it. It carries why the slide
matters - a trade-off, a boundary, a next step - and it is **never the
narration**: not where the material came from, not what the slide is about, not
what you are about to say. One test, and it settles every line: **would a reader
copy this into their notes?** A slide that has nothing left to say does not owe
one, and an empty-looking slide is not a reason to find something.

**Which kind of deck that absence is evidence about:** the worked deck is a
discovery deck, and `SKILL.md` step 5 says a discovery deck's callout is often
nothing at all. A proposal has no worked example here, so nothing says how often
one earns the line there. The test above is the same either way.

**A caption on a slide is not a caption in a document.** 9pt is readable on
paper and is not from the back of a room: a slide caption is roughly 2.67x the
print size.

**Sizes land on the scale, never between its steps.** A reader cannot tell 13.5
from 14, so the difference registers as noise rather than as hierarchy.

Spacing is a 4pt base unit: 4-5pt inside a tag, 8-10pt inside a component,
16-20pt between components, 24-32pt around a section title.

## The content rules the layout enforces

| rule | detail |
|---|---|
| no divider slides | number sections with `.eyebrow` instead; it saves one slide per section |
| no CJK parentheses | `（...）` becomes `·` or a comma |
| the ghost deck test | read only the titles, in order: they must tell the argument. **This is a proposal rule and inverts for a discovery deck** - `SKILL.md` step 5 |
| a title names, never counts | a quantity spends the line without naming a subject - `SKILL.md` step 1, and a discovery deck keeps the customer's heading even where it counts |
| one evidence shape | one proof form per slide - a chart, a table, a screenshot, a quote. Split a slide with two |
| one line per bullet | trim until it fits. Never let one wrap. **Also a proposal rule** |
| empty space over 50% | a draft defect: merge with a neighbour, cut the slide, or add evidence that earns the space. Shrinking the page is a last resort and applies to the whole deck |
| empty space 25-50% | fine as it stands. Add a supporting line if there is one to add; never pad with prose, and **never with a `.co`** - a callout is a line you already had and had no room for, so an empty slot is the one place it cannot come from |

**The last three are the ones that damage a discovery deck**, and a layout
skill's own density checks will enforce them anyway. `SKILL.md` decides which
apply; this file only says what they are.

## The content contract

`slides.json` beside this file is the schema, **and it is the proposal's**:
`layout` is one of `cover`, `chapter`, `content`, `quote`, `metrics`, `close`;
a title is a declarative claim rather than a topic label; `items` is three to
five, each one line; `cap` is the `.co` above and carries the rule stated
there. The middle two are the rules the table above marks as the proposal's, so
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
