# SkillSet, SourceSet, Syncer

Getting skills to a deployed agent. **A SkillSet is a trio, not a CR** - and the
two repos disagree about how the trio is wired, in a way that is dated.

**Seen in:** two deployments that wire it differently - one file holding three
CRs per skill set, versus several skill sets sharing one store.

**Checked:** 2026-09-02 against both shapes: 1:1:1 in the newer deployment, one SourceSet shared across five SkillSets in the older one, and against the CRD.

**Unchecked:** nothing outstanding on the structure. The search-path rule is stated by the deployments themselves rather than enforced anywhere.

**Read the platform side first:** `asgard-cli wiki tools` -
how MCP Server, Skillset and Plugin differ. This page assumes you have.

## When this shape, and when not

Decide design time or runtime first - they are not interchangeable.

| | `.agents/skills/<skill>/SKILL.md` | `common/skills/<skill>/SKILL.md` |
|---|---|---|
| read by | the coding agent working in the repo | the deployed agent, at runtime |
| when | while authoring CRs | while answering a user |
| delivery | straight from the working tree | commit -> SourceSet -> SkillSet -> bound by Agent or SandboxBlueprint |
| reaches the cluster | never | yes, that is the point |

Everything below is about the second kind. A skill in the first kind needs no CR
at all.

## The shape

    SourceSet  ss-<name>     the store. members are directories in it.
      <- Syncer  syn-<name>  fills one member. syncerClass git / database / web.
    SkillSet   sk-<name>     sourceSetName + searchPaths, one path per skill
      <- Agent.managed.skillSetNames, or SandboxBlueprint.skillSetNames

## One SourceSet per SkillSet

`platform-api`'s `POST /v1/skill-set/from-git` creates a SkillSet **plus its own
SourceSet plus the git Syncer**, bound 1:1:1, and the Platform UI reads that
structure. Keep one file per skill set holding all three CRs separated by `---`.

The Syncer writes to **`destinationPath: "git/"`**, trailing slash included, and
a searchPath is then `git/skills/pdf/`. **The SourceSet declares no members** -
the paths its Syncers write to are the whole truth about what is in it.

> **Two older shapes you will find in charts that have not been touched
> recently.** Neither is worth copying, and the first one will not even apply:
>
> - **`members:` on the SourceSet, with `destinationMemberKey` / `stateMemberKey`
>   on the Syncer.** The member registry is gone; the fields are now
>   `destinationPath` and `statePath`.
> - **One SourceSet shared across several skill sets**, sliced apart with
>   searchPaths, *for a skill set that is its own unit*. Changed away from on
>   2026-08-28: it leaves the Platform UI
>   unable to find a skill set's git config, so it renders as a skill set with no
>   source - a UI failure, not a runtime one, which is why it survives unnoticed.

### The one place a shared store is right

A **Plugin bundle** is the exception, and it is deliberate rather than a chart
that was never updated. A deployment carrying 28 Plugins keeps one
`ss-skill-repos` and lets each bundle's SkillSet slice it with `searchPaths`,
because the skills all live in one repository and 28 SourceSets over the same
repository would be 28 clones of it.

That shape accepts the UI cost knowingly: those SkillSets carry no
`skill-set-name` annotation and no `managed-by` label, so they are not presented
as first-class skill sets in the UI at all - they are implementation detail of a
Plugin. See `asgard-cli usecase plugin`.

**The rule stands for anything a person picks in the UI.** It is a shared skills
monorepo behind a bundle that earns the exception, not convenience.

## Generate it

    asgard-cli add skillset <name> --repo <git url>

That writes the structure below with the fields that fail silently already in
place - the display annotation, the labels the UI needs, the current field names.
**Copying the skeleton by hand is where those get lost**, because nothing tells
you they are missing: not helm lint, not CRD validation, not a server dry-run.

The generated file marks the judgement calls TODO. Those are what the rest of
this page is about.

## The skeleton

One file per skill set, `templates/skill_set/sk-<name>.yaml`, holding all three
CRs separated by `---`.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: SourceSet
metadata:
  name: ss-sk-<name>
  annotations:
    asgard-ai.com/source-set-name: "<display name>"
  labels:
    # Hides the Syncer from the generic list and is what the platform cascades
    # on when the SkillSet is deleted. Needed on both CRs.
    asgard-ai.com/managed-by: skill-set
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  apiKey:
    valueFrom:
      secretKeyRef:
        key: asgard_resource_api_key
        name: {{ include "<chart>.appSecretName" . }}
---
apiVersion: asgard-ai.com/v1alpha1
kind: Syncer
metadata:
  name: syn-sk-<name>
  annotations:
    asgard-ai.com/syncer-name: "<display name>"
  labels:
    asgard-ai.com/managed-by: skill-set
    # Scheduler off; CD triggers it once after each helm upgrade.
    asgard-ai.com/syncer-suspend: "true"
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  sourceSetName: ss-sk-<name>
  # A relative path inside the volume. The trailing slash is load bearing:
  # every syncer class decides "this is a directory" from it, and without one
  # the CRD rejects the resource outright.
  destinationPath: "git/"
  syncerClass: git
  schedule: "0 */6 * * *"
  timeZone: Asia/Taipei
  git:
    repoUrl: "https://github.com/<org>/<repo>.git"
    # A public repo: pin main. This repo, private: pin the release tag, which
    # means the tag has to exist before the Syncer can find its commit.
    revision: "main"
    # auth only for a private repo
    auth:
      type: http
      username:
        value: "git"
      password:
        valueFrom:
          secretKeyRef:
            name: {{ include "<chart>.appSecretName" . }}
            key: asgard-github-pat-password
---
apiVersion: asgard-ai.com/v1alpha1
kind: SkillSet
metadata:
  name: sk-<name>
  annotations:
    asgard-ai.com/skill-set-name: "<display name>"
    asgard-ai.com/skill-set-description: "<what these skills cover>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  sourceSetName: ss-sk-<name>
  searchPaths:
    # One path per skill directory. A parent directory resolves to nothing.
    - git/skills/pdf/
    - git/skills/docx/
```

The skill files themselves live in `common/skills/<skill>/SKILL.md` when they
come from this repo, with `name` in the frontmatter matching the directory.

## Designing a skill - the part the generator leaves TODO

### What belongs in a runtime skill

Domain knowledge the CR vocabulary cannot carry: what a status code actually
means in this business, how two systems' entities correspond, which of two
similar figures people mean, the operating procedure for a task.

**Not** anything expressible as data or as a tool. A skill that says "call the
API and read the third field" should have been a tool.

### The description is the loading decision

An agent decides whether to load a skill from its `description` alone, so write
it as a **trigger condition**, not a summary:

    ✗ 關於庫存管理的知識
    ✓ Use when the user asks about stock levels, reorder points, or why a
      figure differs between two systems.

### Do not ship an empty skeleton

A skill directory with a heading and no content produces an agent that has been
told a capability exists and then finds nothing. One deployment ended up
declaring in the description that the skill was incomplete and must not be
used - which is a good recovery from a mistake worth not making. **Create the
skill when there is knowledge to put in it.**

### Dependencies are not resolved for you

If one skill only works when another is also bound, **the platform will not
figure that out**. Say it in the frontmatter, and make sure every consumer binds
both - a SkillSet lists paths, nothing more.

## Fields that are not obvious

**A searchPath is one skill directory.** The platform resolves one searchPath as
one skill; naming a parent directory silently resolves to **nothing**. `.../pdf/`,
not `.../skills/`. Not checkable from the CR - it shows up only as an agent with
fewer skills than expected.

**`asgard-ai.com/managed-by: skill-set`** goes on the SourceSet *and* the Syncer
in the 1:1:1 shape. It hides the Syncer from the generic list, and it is what the
platform cascades on when a SkillSet is deleted.

**`destinationPath` and `statePath` are relative paths inside the volume**, and
the CRD enforces the shape: no leading `/`, no `.` or `..` segments, no `//`.
`destinationPath` **must** end with `/`; `statePath` must **not**.

**No `statePath` on a git Syncer.** It clones the whole tree every run; the
incremental cursor belongs to a database syncer's `isMaxValueColumn`.

**`revision` decides what a private repo syncs:**

```yaml
    revision: {{ .Chart.AppVersion | quote }}   # this repo, private: needs the release tag
    revision: "main"                            # a public repo, or one not tied to releases
```

Pinning `AppVersion` means **the tag must be pushed before the Syncer can find
its commit**, because only CI stamps the release tag into `appVersion`. It is
also why a local `helm upgrade` breaks the Syncer: the placeholder version
renders as a git ref that does not exist.

**Auth only for a private repo.** A public one carries no `auth` block and needs
no `asgard-github-pat-password` in `app-secret`. Adding a private source adds
that key.

```yaml
    auth:
      type: http
      username: {value: "git"}
      password:
        valueFrom:
          secretKeyRef: {name: ..., key: asgard-github-pat-password}
```

## Two suspend switches, and they are not the same

    asgard-ai.com/syncer-suspend: "true"      the platform's. Stops the scheduler.
                                              Does NOT stop CD.
    asgard-ai.com/syncer-cd-trigger: "false"  this repo's own opt-out, read only
                                              by CD.

Both shapes set `syncer-suspend: "true"` and let CD trigger each Syncer once
after `helm upgrade` - sync timing is tied to deploys on purpose.

The CD step is `kubectl create job --from=cronjob`, which works fine against a
suspended CronJob. **Never "fix" CD by skipping suspended CronJobs** - every git
skill Syncer is suspended, and skipping them stops skills syncing everywhere.

That opt-out label lives **on the Syncer CR, and CD has to go and fetch it**. The
reconciler puts only `syncer-name` on the derived CronJob and copies nothing
else, so reading the opt-out off the CronJob silently finds nothing and triggers
anyway. That shipped once as a bug; the CronJob-vs-CR distinction is the whole
of it.

## And one that fails the deploy

**Whether CD requires at least one Syncer per deployed project is one `if` in
your own workflow.** The trigger-and-wait step polls for CronJobs labelled
`asgard-ai.com/syncer-name` and exits 1 after 180 seconds if it finds none, even
when `helm upgrade` succeeded - but some workflows count what the chart declares
first and skip the whole step at zero, and a production chart runs today with
none. Check before the first tag:

    grep -n syncer-name -A15 .github/workflows/*.y*ml

If it waits unconditionally, a project whose only skills are design-time still
needs one. If it skips, nothing in the pipeline is checking that project at all,
and a green deploy means helm returned.

## Verify

```bash
asgard-cli check     # frontmatter name matches the directory
asgard-cli verify <project>
```

The xref check resolves `sourceSetName`, confirms every `searchPath` falls under
a declared member, that a SourceSet backs at most one SkillSet, and that
`managed-by` is on both CRs.

**Neither catches the "one skill directory per searchPath" rule** - that is not
visible in the CRs. It shows up only as an agent with fewer skills than expected,
so check the resolved skills after the first sync.
