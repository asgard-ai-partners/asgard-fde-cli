<<if .Projects>>No request is open, no task is open, and every project has a read path and an
entry point:

**Before picking the next thing up, one minute on what was never read.**
`asgard-cli reading` lists the pages this engagement opened and the ones it did
not. Most of the second list is irrelevant and that is fine; the ones worth a
look are those you would expect to be relevant - a page about the thing that took
a day. Every expensive mistake found in this material so far was a page in that
column, present and never followed.

<<range .Projects>>  <<printf "%-12s" .Slug>><<.Summary>>
<<end>><<else>>No request is open and no task is open. There are no projects either, which at
this point means the split has not been decided rather than that work is
missing: the split follows a requirement, and none is recorded.
<<end>>
So there is one question, and it is the customer's to answer:

    What do they want the agent to do that it cannot do today?

    asgard-cli next --stage requirements    <- the interview, read it first
    asgard-cli request add "<what they asked for, in their words>"

That writes `requirements/requests/REQ-xxx-<name>.md` with today's date and
`draft` on it, registers it, and `asgard-cli next` walks that request from then
on instead of printing this page. Their words, not the design you already have in
mind: translating it into a project, a read path and an entry point is the
request's own job, and keeping the original is what lets the next reader check
that the translation was right. One request per thing they asked for.

If there is nothing new, the work that is left is to answer an open question
(anything below this line), tag what is already green (`asgard-cli next --stage
deploy`), or write down what the engagement learned - the parts of `AGENTS.md`
still marked TODO and the living spec under `docs/spec/<<.SpecSlug>>/`.
