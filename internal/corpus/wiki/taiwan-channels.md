# The commerce channels a customer will name, and what we have

Written because the same names come up in every engagement here and the answer to
"do you already integrate with X" was being worked out from scratch each time.

## SHOPLINE: yes, and deeply

**A commerce middleware deployment integrates SHOPLINE through two skills**, and
they are the most developed channel material in existence here. Anyone asked
about SHOPLINE should read them before answering:

    asgard-freyr-skills/shopline/              the Open API
    asgard-freyr-skills/shopline-backoffice/   the back office

They are separate because the two surfaces are separate, and the split carries a
constraint worth knowing before promising anything:

**The Open API cannot write the merchant's own fields.** Store name, phone,
email are readable and not writable - there is no merchant write endpoint. What
the API can write is Merchant Metafields, the store-level custom fields.
Changing the store's own details means the back office.

**The back office is mapped, not browsed.** 88 page entry points, all 88
declared for whether anything deeper sits under them; 160 rows of operations
covering in-page tabs, dialogs, editor panels and apps inside an iframe; 14
API domains recorded. The discipline is **API first** - where a contract was
observed the skill calls the back office's own API rather than opening a
browser, and browser operation is the fallback rather than the method.

That 88-page map is the one `../usecase/browser-operation.md` refers to when
it says a capability was "a skill describing 88 pages plus everything the menu
cannot see". This is that skill.

**The token mechanism is decided and landed.** It is written up in the skill's
`access.md` rather than here, because it belongs with the calls it authorises.

## Everything else: no, and saying so is the answer

Seven reference deployments were searched for PChome, momo, 蝦皮 / Shopee and
Coupang. The only occurrence in the whole set is a customer's own document asking
for them. So:

    "Do you have a Shopee integration?"     "No. We have SHOPLINE, in depth.
                                             Here is what Shopee would take."

**That is a real answer and a better one than a hedge**, and it is stronger for
having SHOPLINE behind it: it says the work is understood rather than unfamiliar.

## What a new channel's answer depends on

No table of which platform offers an open API, because that changes and this page
would be stale before it was useful. What does not change is the ladder:

    an open API with a test environment   `../usecase/external-api.md`
    an open API, production only          the same shape - and **ask whether
                                          they permit testing against it** - in
                                          the meeting, not assumed here,
                                          rather than assuming
    a data export only            a Syncer over files, not a live integration
    only a web back office        browser operation - and SHOPLINE is what that
                                  costs: 88 pages before the first useful call

**"Sandbox" means the platform's own here.** A customer's test environment is
called that, everywhere, because the agent runs in a sandbox the platform starts
and the two collide in the same paragraph otherwise.

**Ask about each channel separately, in the interview.** This page is read by whoever is building an integration; that question is not theirs to answer, it is one to have asked before they got here. They differ, and one back-office-only
channel among four sets the cost of the whole item. A customer answering "yes we
have API access" usually means the one they use most.

**A channel is usually two surfaces, not one.** SHOPLINE's split - an API that
reads and a back office that writes the rest - is not a SHOPLINE peculiarity. Ask
what the API cannot do before pricing the API.

## What this means for an estimate

**A multi-channel integration is not N copies of one integration.** Before
pricing four, ask the question that most often collapses them:

    "Is there already something that pulls these together for you?"

An OMS, a middleware layer, a warehouse that consolidates the channels - which is
what the Freyr deployment is, from the other side. If one exists, four
integrations become one database and the estimate changes by an order of
magnitude. `asgard-cli guide requirements` puts this at question 3 for the
same reason.

If nobody knows, that is an open question, not an assumption to price on.

## Corresponding extracts

`../usecase/external-api.md` for an API integration's shape,
`browser-operation` for a channel offering no API,
`skill-layers` for how the SHOPLINE material is organised - it is the reference
for a channel skill at full size.

## Sources

- `asgard-freyr-skills`, read 2026-09-02: the two SHOPLINE skills, their
  frontmatter and the repository's own skill table
- Searched the same day across the seven reference deployment charts for PChome,
  momo, 蝔皮/Shopee and Coupang: no chart, skill or document mentions one

**Unchecked:** which of the other channels offers an open API today. Deliberately
not recorded - it is the vendor's to answer, it changes, and a stale answer here
would be worse than none. Also unchecked: whether the SHOPLINE back-office map is
still accurate against the current product, which is the standing risk with any
mapped UI.
