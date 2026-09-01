Data connectors exist but nothing reads through them yet. **This is the first of
the three decisions that get answered wrong.**

Needing a read path:
<<range .Projects>><<if not (.Has "SemanticLayer" "Toolset")>>  - <<.Slug>>
<<end>><<end>>
## The decision

| audience | answer |
|---|---|
| internal, authenticated | SemanticLayer, one per source system, one Agent each |
| public, anonymous | a Toolset of zero-parameter fixed queries, no semantic layer |

## Why a semantic layer is wrong for a public audience

Mounted without allowedCubes it lets the agent compose arbitrary SQL over every
cube in it - and **the exposed surface grows by itself every time a cube is
added**. Nobody goes back to narrow it. Excluding the sensitive tables is not the
fix, because the shape itself is the risk.

With fixed tools, what can be asked is decided by a few statements in version
control, no user input reaches SQL so the injection surface is zero, and widening
it takes a CR change and a review.

> This was answered wrong once already: a public site got a SemanticLayer first,
> and it was later removed and replaced with five zero-parameter query tools.

Read the shape before writing it:

    asgard-cli usecase semantic-layer
    asgard-cli usecase fixed-query-tools

A fixed query tool is a Workflow, so read how a chain passes values between its
processors before writing one - that is where the silent failures are:

    asgard-cli usecase workflow-chain

## If the answer is a semantic layer

One system, one layer, one Agent. An Agent mounting two layers has the search
space the split was meant to shrink.

  - Every cube, dimension and measure needs a description in 繁體中文. It is what
    the agent reads to decide which column answers a question; one without a
    description is invisible to the model.
  - A dimension's name is the column its sql selects, never a re-cased alias.
  - Common analysis views go in top-level sampleQueries, never as a cube-level
    sql: virtual cube. **Run every one against the live database before
    committing it** - a sampleQuery that errors actively misleads the agent.

## If the answer is fixed tools

Every tool takes zero parameters, and requestConsent is false because they are
read-only.

**Do not ship two tools that sit on the same FROM/JOIN and differ only in
projection.** Each one's description then has to name the other as the
alternative, and that is exactly where a model picks wrong. Merge them and make
the difference a column value instead of a tool choice.

Tool usage guidance goes in each tool's Workflow entries[].tooling.description.
Toolset.spec.instruction does not exist any more - adding it back passes
dry-run and then fails the real deploy.

Done when: every project reads through one shape or the other, and
asgard-cli verify plus asgard-cli verify are green.
