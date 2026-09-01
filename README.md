# asgard-fde-cli

Command line tool for Asgard FDE (`asgard-cli`).

## Development

```bash
go build -o asgard-cli ./cmd/asgard-cli   # build
go test ./...                             # test
./asgard-cli version                      # run
```

Layout:

```
cmd/asgard-cli/     main; signal handling and exit codes only
internal/cli/       cobra command tree, one file per subcommand
internal/config/    reads and writes .asgard-config.json
internal/work/      reads and writes the customer repo's own records of its work
                    (requests, task specs, open questions, decision records)
internal/deploy/    reads projects/<project>/deploy.yaml, the deployment SoT
internal/render/    renders a chart via helm, the way CD does
internal/gate/      the invariant checks on a rendered chart (xref, agent split)
internal/tool/      resolves helm/kubectl/python3, and how to install one
internal/version/   build information (injected by GoReleaser via ldflags)
```

To add a subcommand, write a `newXxxCmd()` in `internal/cli/` and register it in
the `cmd.AddCommand(...)` call in `root.go`.

See [AGENTS.md](AGENTS.md) for the conventions this repo follows.

## Commands

### `init`

Create `.asgard-config.json` in the current directory, binding it to a workspace.
The workspace is the customer:

```bash
asgard-cli init --workspace-id 7ab7f523-3cd9-7e87-a873-6f1fa6028104
```

```
Created /path/to/unitech-e-asgard-kube/.asgard-config.json
  workspace.id    7ab7f523-3cd9-7e87-a873-6f1fa6028104
  workspace.slug  unitech-e
  workspace.name  unitech-e
  repository      unitech-e-asgard-kube

No projects yet. Add one with `asgard-cli project add <slug>`.
```

`--workspace-id` is issued by the Asgard platform and is required. Its format is
not validated yet, pending an API to verify it.

`--workspace-slug` defaults to the directory name with a trailing `-asgard-kube`
removed, so running inside `unitech-e-asgard-kube` yields `unitech-e`.
`--workspace-name` defaults to the slug. `--project` is a repeatable shortcut
that creates projects up front, each with the `dev` environment only.

Rerunning is safe: when the config exists nothing is changed and the current
settings are printed. `--force` rebinds, and still requires `--workspace-id`.

### `project add`

A project is the unit of deployment: one Helm chart, one namespace per
environment. Projects are added as the engagement discovers them.

```bash
asgard-cli project add internal --env dev --env prod
```

```
Added project "internal" to /path/to/.asgard-config.json
  dev   asgard-unitech-e-internal-dev
  prod  asgard-unitech-e-internal-prod

Before the first deploy of each environment:
  1. tf-asgard must create the namespace and its app-secret first. Declaring an
     environment before they exist makes the next tag fail at helm upgrade.
  2. the project needs at least one Syncer. CD waits for a CronJob labelled
     asgard-ai.com/syncer-name and exits 1 after 180s if it finds none, even
     when helm upgrade succeeded.
```

`--env` may be repeated and defaults to `dev`; `dev` and `prod` are the only
valid symbols and a project may declare either or both. `--name` defaults to the
slug.

Both notes in the output are ordering traps: each one fails during CD rather than
here, which is far too late to find out.

### `scaffold`

Write the repository skeleton next to `.asgard-config.json`:

```bash
asgard-cli scaffold
```

It writes the part of a customer repo that is the same for every engagement:

| | |
|---|---|
| `AGENTS.md` | the platform contract, with the customer-specific sections marked TODO |
| `docs/` | the four-layer model (meeting-notes / decisions / living spec) and the SDD rules |
| `requirements/` | the task and request indexes |
| `scripts/check_*.py` | the four acceptance gates |
| `scripts/db/` | the query and introspection tool-chain, with an empty target registry |
| `.agents/skills/` | the three design-time skills |
| `common/` | `asgard-cli render`, the per-env overlay points, the runtime-skill directory |
| `.github/workflows/main.yaml` | tag-driven CD |
| `projects/<slug>/` | one chart skeleton per project, with `values-<env>.yaml` only for the environments that project declares |

What it does **not** write is the customer's own knowledge: which systems exist,
how the projects split, what the CRs look like. That is what the onboarding
produces, and no template can generate it.

Re-running is safe. Existing files are left alone and counted as already
present, so it can be run again after adding a project, or when a file was
deleted by hand. `--force` overwrites, which discards local edits.

The generated skeleton passes its own gate on the first run:

```bash
asgard-cli check                              # structure is consistent
helm lint projects/<slug>/chart/app           # 0 charts failed
```

### `request`, `task`, `question`, `decision`

Work arrives as a **request**: one thing the customer wants that the agent cannot
do today. Everything else hangs off it.

```bash
asgard-cli request add "warehouse staff want to ask about stock levels in chat"
asgard-cli request target REQ-001 erp
asgard-cli request ready REQ-001

asgard-cli task add "expose stock levels" --request REQ-001 --project erp --complexity M
asgard-cli task ready TASK-001
asgard-cli task start TASK-001
asgard-cli task done TASK-001

asgard-cli question add "which stock figure is authoritative" --blocks REQ-001 --ask "the warehouse lead"
asgard-cli question answered 1 "location 608 only" --decision 2026-09-04-safety-stock.md

asgard-cli decision add "the website reads through fixed query tools" --module architecture.md
```

**None of this state lives in the CLI.** Each command writes a file in the
customer's repository, because that repository is what the next agent opens:

| record | file | carries |
|---|---|---|
| request | `requirements/requests/REQ-xxx-<name>.md` + its registry row | date raised, status, target project, the customer's own wording |
| task spec | `requirements/tasks/TASK-xxx-<name>.md` + its queue row | date created, status, the SDD sections, a dated log of every transition |
| open question | a row in `docs/open-questions.md` | date raised, what it blocks, who can answer |
| decision | `docs/decisions/YYYY-MM-DD-<topic>.md` + a traceability row | the date in the file name, the module it changed |

The reason these are commands rather than instructions to write a file is that
each record lives in more than one place. A task's status is in the queue table,
in the spec's own `Meta`, and in the spec's execution log; a request's target
project is in the registry's Spec column and in its `Meta`. Moving one by hand
means three or four edits, and a repository where two of them disagree gives the
next reader no way to tell which is current. Every command here moves all of
them, and stamps the date rather than asking for it.

`asgard-cli next` reads all four back. Open questions print first, on every run,
before anything else it has to say.

### `render`, `verify`, `doctor`

These three are why the acceptance gate now runs on Windows.

```bash
asgard-cli render erp dev              # manifests to stdout, summary to stderr
asgard-cli verify                      # render every declared env, check invariants
asgard-cli verify --rendered file.yaml # check a stream that is already rendered
asgard-cli doctor                      # which external tools are here, and how to get them
```

`render` replaces the generated repo's `common/render.sh`, and `verify` replaces
its `check_chart_xref.py` and `check_agent_split.py`. The old chain was:

```
bash render.sh  ->  yq (read deploy.yaml)  ->  helm template  ->  python3 + PyYAML
```

Four external dependencies, of which **three do not work on Windows without WSL
or Git Bash** - while helm and kubectl both have native Windows builds. The
prerequisite is now `helm` alone, and one binary:

```
asgard-cli verify  ->  helm template  ->  internal/gate
```

`verify` renders in process, so there is no pipeline and no temporary file, and
`render` keeps the manifests on stdout with everything else on stderr so the pipe
forms still work identically in cmd, PowerShell and bash.

The gate rules were ported one for one, and the port was checked by running both
implementations over the same rendered chart: same findings, same counts. Step 4
of the gate still needs `kubectl` and `check_crd_fidelity.py`, because it talks
to a cluster.

### helm and kubectl are prerequisites, not dependencies

Nothing about how asgard-cli is distributed can install them. A tar.gz, a zip and
`go install` carry no dependency metadata and never can, and a dependency
declared on a Homebrew tap or a Scoop bucket would only cover people who install
that way - which is nobody today, since neither repository exists. Declaring one
anyway would read as a guarantee that does not hold.

So the binary is the mechanism. Every command that needs helm resolves it through
`internal/tool` first and refuses with the install line for the machine it is on,
and `asgard-cli doctor` reports all of them at once:

```
$ asgard-cli doctor
platform  darwin/arm64

MISSING helm               render a chart (asgard-cli render) and lint it
ok    kubectl              v1.35.1
ok    python3 (optional)   Python 3.14.7

helm is not on PATH, and asgard-cli needs it to render a chart and lint it

Install it with:

    brew install helm
```

It works out the command for the machine it runs on, including which Linux
distribution - neither kubectl nor helm is in the Debian or Ubuntu default
repositories, so the honest answer there is not an `apt install`. It exits
non-zero when a required tool is missing, so it works as a CI preflight.

`init`, `scaffold`, `next`, `project`, `request`, `task`, `question`, `decision`
and `check` need none of these tools. `render` and `verify` need helm; step 4 of
the gate needs kubectl.

### The config file

```json
{
  "workspace": {
    "id": "7ab7f523-3cd9-7e87-a873-6f1fa6028104",
    "slug": "unitech-e",
    "name": "unitech-e"
  },
  "projects": [
    { "slug": "internal", "name": "internal", "environments": ["dev", "prod"] },
    { "slug": "website",  "name": "official site", "environments": ["dev"] }
  ]
}
```

The slug is load bearing. Two names are **derived** from it rather than stored,
so the config cannot drift from the layout on disk:

| derived | rule | example |
|---|---|---|
| repository | `<workspace.slug>-asgard-kube` | `unitech-e-asgard-kube` |
| namespace | `asgard-<workspace.slug>-<project.slug>-<env>` | `asgard-unitech-e-internal-dev` |

Both commands validate before writing, so an invalid config is never created:

| check | why |
|---|---|
| slugs are DNS-1123 labels | a slug becomes part of a namespace, so anything else is rejected later by the apiserver |
| environments are `dev` or `prod`, no duplicates | those are the only symbols the platform accepts |
| project slugs are unique | two projects with one slug would fight over a namespace |
| every derived namespace fits in 63 characters | names derived from a namespace inherit its length; without this the failure surfaces in `helm upgrade` during CD |

Every problem is reported at once rather than one per run.

`.asgard-config.json` is project configuration and belongs in version control.
Read and write it through `internal/config` (`config.Load` / `config.Save` /
`config.Find`) rather than assembling JSON inside a command. `config.Find` walks
up from the current directory, so a command still resolves the project root when
run from a subdirectory.

## Releasing

Releases are driven by [GoReleaser](https://goreleaser.com). Pushing a tag triggers
`.github/workflows/release.yml`:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

One run produces binaries for linux / darwin / windows x amd64 / arm64, `.deb` /
`.rpm` / `.apk` packages and checksums, all attached to the GitHub Release. The
changelog is grouped automatically from conventional commit messages (`feat:`,
`fix:`).

To verify locally without publishing anything (output lands in `dist/`):

```bash
goreleaser check
goreleaser release --snapshot --clean --skip=publish
```

### Channels not yet enabled

The bottom of `.goreleaser.yaml` has ready-made **Homebrew tap** and **Scoop
bucket** blocks. Create the corresponding repo and a PAT with write access to it,
then uncomment:

| Channel | Prerequisite |
| --- | --- |
| Homebrew | Create `asgard-ai-partners/homebrew-tap`, secret `HOMEBREW_TAP_TOKEN` |
| Scoop | Create `asgard-ai-partners/scoop-bucket`, secret `SCOOP_BUCKET_TOKEN` |

Other things worth knowing:

- **CGO**: builds run with `CGO_ENABLED=0` so cross-compilation fits on a single
  runner. Pulling in a cgo dependency (sqlite and friends) means switching to
  zig cc or per-platform runners.
- **macOS signing**: an unsigned binary downloaded through a browser is blocked by
  Gatekeeper (installing via Homebrew is not affected). Add `anchore/quill` when
  notarization becomes necessary.
- **LICENSE**: the archive config picks up `LICENSE*`. No such file exists yet, so
  GoReleaser prints one warning, which does not affect the release.
