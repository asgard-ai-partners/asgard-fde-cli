# Nothing in flight

**This stage is a repository with no request and no task open.** What moves the
work on is the customer, so the only question left is what they want next.

**Before picking the next thing up, one minute on what was never read.**
the material lists the pages this engagement opened and the ones it did
not. Most of the second list is irrelevant and that is fine; the ones worth a
look are those you would expect to be relevant - a page about the thing that took
a day. Every expensive mistake found in this material so far was a page in that
column, present and never followed.

    What do they want the agent to do that it cannot do today?

    ../guide/requirements.md    <- the interview, read it first
    asgard-cli request add "<what they asked for, in their words>"

That writes `requirements/requests/REQ-xxx-<name>.md` with today's date and
`draft` on it, registers it, and `asgard-cli request` reports it from then
on instead of printing this page. Their words, not the design you already have in
mind: translating it into a project, a read path and an entry point is the
request's own job, and keeping the original is what lets the next reader check
that the translation was right. One request per thing they asked for.

If there is nothing new, the work that is left is to answer an open question
(anything below this line), tag what is already green (`../guide/deploy.md`), or write down what the engagement learned - the parts of `AGENTS.md`
still marked TODO and the living spec under `docs/spec/<spec-slug>/`.

**Checked:** 2026-09-04 - **no platform claims to check.** This document reads
the repository's own records back and says what is left when none of them is
open.

**Unchecked:** that the state it describes is a normal one rather than a gap.
That reading comes from the engagement this was written in, where long stretches
waited on a meeting, an account or somebody's approval, and being told only that
nothing was in flight read as being behind. **Nothing can corroborate it**, and
a reader for whom it is not true should trust their own read of where the work
stands.
