# The deck's design language

**What a deck looks like, so that this skill does not depend on a typesetting
skill being installed.** The rules below are the ones a proposal or discovery
deck uses; the content decisions are `SKILL.md`'s and are not repeated here.

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
    h2          24pt, weight 500, margin-bottom 14pt   the page title
    h3          15pt, weight 500, in --brand            a section heading
    .eyebrow    9.5pt mono, tracking 2pt, --stone       the stable section label
    .lead       12pt, --olive                           one line under the title

    .mt  16pt   module title, paired with .ml
    .ml  24pt   the large letter prefix, in --brand
    .ms  7.5pt  mono module sub-label, with a bottom rule
    .mb  11pt   module body
    .mi  11pt   module line item, 8pt vertical padding
    .mc  9.5pt  cadence note, with a top rule
    .co  11pt   the bottom callout, pinned 12mm from the foot

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
| one evidence shape | one proof form per slide - a chart, a table, a screenshot, a quote. Split a slide with two |
| one line per bullet | trim until it fits. Never let one wrap. **Also a proposal rule** |
| empty space over 50% | a draft defect: merge with a neighbour, pin a `.co` callout, or add a chart that earns the space. Shrinking the page is a last resort and applies to the whole deck |
| empty space 25-50% | fine with a pinned callout, otherwise add one supporting line. Never pad with prose |

**The last three are the ones that damage a discovery deck**, and a layout
skill's own density checks will enforce them anyway. `SKILL.md` decides which
apply; this file only says what they are.

## The content contract

`slides.json` beside this file is the schema: `layout` is one of `cover`,
`chapter`, `content`, `quote`, `metrics`, `close`; a title is a declarative
claim rather than a topic label; `items` is three to five, each one line; `cap`
says why the slide matters and never restates the title.

## Rendering it

**HTML to PDF, not a presentation format.** A pptx passed through a converter
loses CJK weight, tracking and glyph spacing; an HTML renderer embeds the fonts
exactly. `discovery-deck.html` beside this file is a working eight-page deck in
plain HTML with no separate content file - it is the shortest way to see all of
the above applied.

**PDF is what gets handed over.** Produce an editable file only when the
customer has said they want to edit it.
