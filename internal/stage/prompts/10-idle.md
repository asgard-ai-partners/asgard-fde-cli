---
description: the one question left when no request and no task are open, and why that state is not being behind
---
# Nothing in flight

This stage applies to a repository with no request and no task open. The next
step comes from the customer, so the question is what they want next.

<<if .Projects>>No request is open, no task is open, and every project has a read path and an
entry point:

Before picking the next thing up, spend a minute on the material you have not
read. The material under `.agents/skills/asgard-platform/` covers more than any
one engagement uses, and most of what went unopened is irrelevant. Look at the
pages you would expect to be relevant, such as a page about something that took
a day. Every expensive mistake found in this material so far was covered by a
page that existed and was not read.

<<range .Projects>>  <<printf "%-12s" .Slug>><<.Summary>>
<<end>><<else>>No request is open and no task is open. There are no projects either, which at
this point means the split has not been decided rather than that work is
missing: the split follows a requirement, and none is recorded.
<<end>>
The question for the customer is:

    What do they want the agent to do that it cannot do today?

    ../guide/requirements.md    <- the interview, read it first
    asgard-cli request add "<what they asked for, in their words>"

That writes `requirements/requests/REQ-xxx-<name>.md` with today's date and
`draft` on it, registers it, and `asgard-cli request` reports it from then
on instead of printing this page. Use their words, not the design you already
have in mind. Translating it into a project, a read path and an entry point is
done inside the request, and keeping the original lets the next reader check
that the translation was right. One request per thing they asked for.

If there is nothing new, the work that is left is to answer an open question
(anything below this line), tag what is already green (`../guide/deploy.md`), or write down what the engagement learned - the parts of `AGENTS.md`
still marked TODO and the living spec under `docs/spec/<<.SpecSlug>>/`.

**Checked:** no platform claims to check. This document reads
the repository's own records back and says what is left when none of them is
open.

**Unchecked:** that the state it describes is a normal one rather than a gap is
one engagement's reading, and a reader for whom it is not true should trust
their own read of where the work stands.
