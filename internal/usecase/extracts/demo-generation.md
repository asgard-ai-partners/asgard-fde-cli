# A demo, and the pipeline that generates one

Building something that looks like a prospect's business when you have none of
their systems, none of their data and no credentials - and having it be a real
deployment rather than a slide.

**Seen in:** a generator carrying twelve industries, ten of them with a complete
chart. It is the largest body of Asgard chart material in existence: 123
SkillSets, 80 Workflows, 64 SemanticLayers, 64 Agents and 51 Toolsets, all
produced from one source shape repeated per industry.

**Checked:** 2026-09-02 against that repository - its wiring diagram, the
frontmatter of a retail story, its consistency checker, and the CR kinds each
industry's chart declares.

**Unchecked:** whether a demo built this way has ever converted into an
engagement, and what the second one costs once the first industry exists. The
repository records neither.

**Read the platform side first:** `asgard-cli wiki product-suite` -
which product a request lands in. This page assumes you have.

## When this shape, and when not

**Use it when the other side is a prospect rather than a customer.** No access,
no credentials, no network path, and a date. Everything the interview asks for -
who is on the other end, which systems, who issues the account - has no answer
yet and asking produces a meeting rather than a demo.

**Do not use it for a customer who has systems.** The MVP rule holds: a first
delivery runs against their real channel with their real documents, and a demo on
invented data proves nothing they are measuring. This shape is what you build
*before* that conversation, to have one.

The distinction that matters: this deployment is **complete and fake**, where an
MVP is **partial and real**. Both are legitimate; confusing them is not.

## The pipeline

    stories/*.md          the story, and its frontmatter is the design
      + attachments/      letters, drawings, notices - the unstructured sources
        |
    systems/db/*.sql      the data, and the SQL is the source of truth
        |  applied to Postgres, one schema per system
    skills/<name>/SKILL.md
        |
        v   one transform
    <industry>/chart/     a Helm chart of asgard-ai.com CRs
        |
        v   helm install
    namespace asgard-demo-<industry>

**The story is the entry point and its frontmatter is the whole design.** Before
a CR exists, one file declares what the demo is made of:

    id, title, trigger
    agents:  [ag-store-ops, ag-allocation, ag-replenishment, ...]
    systems: [pos, wms, erp, eshop, supplier, crm]
    skills:  [stockout-detect, demand-forecast, transfer-optimize, ...]

That is the same decision the interview reaches by asking - which systems, what
capability, who acts - written down first because there is nobody to ask.

## Two scenarios per industry, and they are the two halves

Every industry carries a **flagship** and an **insight**, and the split is not
stylistic - it is the read/write divide made into a demo:

| | flagship | insight |
|---|---|---|
| retail | a stockout becomes a cross-store transfer | clearing slow-moving stock |
| semiconductor | an ECO hot lot pushed in | WAT drift, root-caused |
| finance | a large redemption's liquidity | portfolio concentration |

The flagship ends at an **approval gate** - a write, a person, a decision. The
insight ends at an **analysis** - no write, and it is the Mimir half of the same
data. A demo showing only the flagship looks like automation nobody controls; one
showing only the insight looks like a report.

`asgard-cli usecase write-path` is the flagship's mechanism and
`asgard-cli usecase mimir-dashboard` is the insight's.

## The rule the whole repository is built on

    read   -> SemanticLayer, and the agent decides for itself
    write  -> Toolset + Workflow, and a person approves

Stated once, at the top, and every industry follows it. **Anything SELECT can do
goes through the layer rather than becoming a tool.** A demo that wires a query
as a tool has spent effort producing something narrower than the layer already
offered, and the effort shows up as a tool list nobody can hold in their head.

## Nothing exists that no story uses

The consistency checker enforces referential integrity **in both directions**, and
the second direction is the one worth copying:

  - every skill, system and attachment a story names must exist
  - **every skill, system and attachment that exists must be named by a story**

An orphan is an error, not a warning. That is what keeps twelve industries from
accumulating half-built assets nobody can date, and it is the check a customer
repository does not have - `asgard-cli check` finds a chart that references
something missing, never something present that nothing references.

It runs across the whole repository and has to be clean, not clean for the
industry you touched.

## Fields that are not obvious

**One namespace per industry**, `asgard-demo-<industry>`, created by Terraform
along with `app-secret` before anything is deployed - the same order every
engagement follows, and the same project environment id that cannot be
known until it exists.

**helm has no `--app-version`.** Setting one means packaging first and upgrading
from the package:

    pkg=$(helm package <industry>/chart/app --app-version dev-0.1.0 -d /tmp | sed 's/.*: //')
    helm upgrade <name> "$pkg" --install -n asgard-demo-<industry> \

**Each industry declares its own `CompletionModel`.** Twenty-four of them across
the generator - see `asgard-cli wiki settings` for the class enum and the two
rules the CRD enforces that helm does not.

**The SQL is the source of truth, not the database.** Data is applied from files
per industry, so a demo can be rebuilt from the repository after somebody has
been clicking around in it - which they will have been.

## Read the platform side first

`asgard-cli wiki product-suite` for which product a scenario belongs to;
`asgard-cli usecase write-path` for the approval gate;
`asgard-cli usecase mimir-dashboard` for the insight half.
