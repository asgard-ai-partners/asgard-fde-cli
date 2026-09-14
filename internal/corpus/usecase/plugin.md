# Plugin

A named bundle of capabilities that a blueprint loads by name - and can pick
**per request**.

**Seen in:** a content-generation deployment with 28 of them, split into domain
knowledge (`pg-med-*`, `pg-biz-*`, `pg-pr-*`) and writing style (`pg-style-*`).

**Checked:** 2026-09-02, recounted 2026-09-11 against that deployment at `edb0ad0`: **29 Plugins and 29 SkillSets**, every one naming `ss-skill-repos`, and the CRD. The count moves as the deployment grows; what does not is one store for all of them.

**Unchecked:** how to divide capability into bundles. The naming IS the taxonomy, and no deployment's taxonomy has been reviewed here. That deployment is also the only one of eight that declares a `Plugin` at all (`../wiki/coverage.md`), so there is no second arrangement to tell the shape from its choices.

**Read the platform side first:** `../wiki/tools.md` -
how MCP Server, Skillset and Plugin differ. This page assumes you have.

## When this shape, and when not

Use a Plugin when **the same agent needs different capabilities on different
turns** and the caller knows which. Writing an article about health for one
publication and about sport for another is the same agent with two different
bundles of domain knowledge and house style.

Do **not** use one when the capability set is fixed. Attaching skills straight to
the Agent, or to the blueprint, is simpler and there is nothing to choose at run
time.

The bundle is the point: a Plugin can carry skill sets, toolsets, source set
mounts and hooks together, so "everything needed to write in this style" is one
name rather than four lists that have to stay in step.

## The shape

    Plugin  pg-<name>
      skillSets[]          -> SkillSet
      toolsets[]           -> Toolset
      sourceSetMounts[]    -> SourceSet, mounted
      hooks[]
    SandboxBlueprint.pluginNames  -> the Plugins to load

## Generate it

    asgard-cli add plugin <name>

That writes the structure below with the fields that fail silently already in
place. **Copying a skeleton by hand is where those get lost**, because nothing
tells you they are missing: not helm lint, not CRD validation, not a server
dry-run. The generated file marks the judgement calls TODO - those are what the
rest of this page is about.

## The shared skill store

Every Plugin's SkillSet points at **one** `ss-skill-repos`, not at a SourceSet of
its own. That is the opposite of the 1:1:1 rule in
`../usecase/skill-set.md`, and it is deliberate: the skills live in one
repository, so a SourceSet per bundle would clone the same repository once per
bundle. The deployment with 29 has exactly one.

The cost is paid on the UI side and paid knowingly - these SkillSets carry no
`skill-set-name` annotation and no `managed-by: skill-set` label, so the platform
does not present them as skill sets a person picks. They are the Plugin's
implementation. A skill set someone is meant to choose from the UI still gets its
own SourceSet.

## The skeleton

`templates/plugin/<domain>.yaml`, with the Plugin and its SkillSet in one file so
the bundle and its contents stay together.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: Plugin
metadata:
  name: pg-<domain>
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  skillSets:
    - name: sk-<domain>
  # Written explicitly rather than omitted, so the bundle reads as
  # "this one carries skills only".
  toolsets: []
  sourceSetMounts: []
  hooks: []
---
apiVersion: asgard-ai.com/v1alpha1
kind: SkillSet
metadata:
  name: sk-<domain>
  # No `skill-set-name` and no `managed-by` - see above. The platform does not
  # present a bundled skill set, so there is no name for it to show, and
  # `asgard-cli verify` does not ask for one here.
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  sourceSetName: ss-skill-repos
  searchPaths:
    - <member>/<skill-dir>/
```

Load it from the blueprint:

```yaml
  pluginNames:
    value: "pg-<domain>,pg-<other>"
```

## Choosing plugins per request

`pluginNames` is a `ValueExprTemplate` - it takes either a static `value:` or an
`expression:` of JavaScript evaluated **per turn**, with the BotProvider's
payload available as `prevPayload`. So the blueprint can compute the list:

```yaml
  pluginNames:
    expression: |-
      return ['pg-writing-base', ...(prevPayload.additional_plugin_names || [])].join(",")
```

A base bundle always loads; the caller adds more through the BotProvider payload.
Note the value is a **comma-separated string**, not a list - that is the shape of
every `*Names` field on a blueprint.

This is what makes 28 bundles tractable. Without it, either every agent carries
every skill, or there is one agent per combination.

### Which means the caller chooses the agent's capabilities

**Before writing this shape, settle who can set that payload field - it is a question about our own deployment, not one for the customer.** The chart pins
a floor - `pg-writing-base` here - and everything above the floor arrives in the
request. Whoever can call the BotProvider decides which bundles load, which is
correct when the caller is your own front end deciding "this article is health,
in this outlet's voice", and is a capability-selection primitive handed to a
stranger when the entry point is anonymous.

The same deployment does it twice over: `sourceSetMounts` is also an expression,
and it mounts `ss-article-workspace` at a `subPath` taken straight from
`prevPayload.article_id`, **`readOnly: false`**. So the payload chooses both what
the agent can do and which writable directory it does it in.

Neither is wrong - it is how the shape works. What is wrong is arriving at it
without noticing, which is easy, because the expression reads as chart
configuration and is a public parameter. Two questions settle it: is this
BotProvider's `authMode` `none`, and does anything validate the names before
they are joined?

In this deployment the first answer is no - `authMode: api-key`, with the key read
from the release's own Secret - so every caller is one the operator issued a key to, and handing
that caller its own bundle selection is a reasonable thing to do. **The second
answer is that nothing validates them**: the names go from the payload into
`join(",")` untouched. That is safe here because the caller is authenticated and
because an unknown name loads nothing rather than something else. It stops being
safe the moment the same expression sits behind an anonymous entry point, and
the expression will not have changed - only the BotProvider beside it.

### A Plugin does not have to carry a skill set

`pg-public-opinion` in the same chart is `skillSets: []` with one toolset, and
that is not a mistake or a stub. A Plugin is a named bundle of **capabilities**,
and a bundle of one toolset is a legitimate bundle: it exists so that the tool
can be selected per request by name, which a toolset on the blueprint cannot be.
Of the 28, twenty-six wrap exactly one skill set, one wraps two plus a toolset -
the base - and one wraps a toolset alone.

So the question when reaching for a Plugin is not "do I have skills to bundle"
but **"does this need to be selectable per request"**. If it does, wrap it, even
if the bundle has one member.

## Designing the bundles - the part the generator leaves TODO

### What makes one bundle

**Everything needed to do one job well**, from the caller's point of view -
which usually means a domain plus the skills, tools and mounts it implies.

Decide the axes before the fifth bundle, because the naming is the taxonomy and
it is expensive to change later. One that works: **what it knows** versus **how
it presents**. A bundle per subject area, and a bundle per house style, chosen
independently.

### Base plus additions

Load one bundle always - the shared conventions, the document handling - and let
the caller add to it. That keeps the common case a static list and makes the
expression carry only the variable part.

### The bundle is the unit of review

A capability set that changes together should be one Plugin, so that reviewing a
change means reading one file. Splitting things that always move together
produces four lists that drift.

## Fields that are not obvious

**The same capability can be declared in more than one place, and the order is
fixed: Plugin, then Agent, then Blueprint.** A skill set named by a loaded Plugin
takes priority over the same one named on the agent or on the blueprint itself.

> That ordering is recorded in a deployment's own notes, citing the controller
> that performs the aggregation. **What happens when two loaded Plugins name the
> same thing has not been verified** - do not rely on a particular winner there.
> If it matters, read the controller or test it, and write down what you find.

**Naming carries the taxonomy.** With this many bundles, the prefix is what makes
the set navigable: domain (`pg-med-*`, `pg-biz-*`) versus presentation
(`pg-style-*`). Decide the axes before the fifth one.

**The relationship runs Plugin -> Agent, not the other way.** An `Agent` CR has
no `pluginNames`; the blueprint names the plugins. A chart that has no
blueprint - the agent-hub shape - has nowhere to attach one, which is why plugins
appear only alongside a self-hosted entry point.

## Verify

```bash
asgard-cli check
asgard-cli verify <project>
```

The xref check parses `pluginNames` out of the blueprint - including the
comma-separated string an expression produces - and resolves each one to a real
Plugin, then each Plugin's `skillSets` and `toolsets` to real CRs.

**It cannot check an expression's logic.** A `pluginNames` expression that
returns a name nothing defines is caught; one that returns the wrong bundle for a
given payload is not. Exercise the paths that matter with a real request.
