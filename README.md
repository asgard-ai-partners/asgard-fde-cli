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
internal/version/   build information (injected by GoReleaser via ldflags)
```

To add a subcommand, write a `newXxxCmd()` in `internal/cli/` and register it in
the `cmd.AddCommand(...)` call in `root.go`.

See [AGENTS.md](AGENTS.md) for the conventions this repo follows.

## Commands

### `init`

Create `.asgard-config.json` in the current directory, binding it to a workspace:

```bash
asgard-cli init --workspace-id ws_prod
asgard-cli init --workspace-id ws_prod --workspace-name production
```

`--workspace-id` is required. `--workspace-name` defaults to the directory name.

```json
{
  "workspace": {
    "id": "ws_prod",
    "name": "production"
  }
}
```

Rerunning is safe: when the config already exists nothing is changed and the
current settings are printed.

```
$ asgard-cli init
.asgard-config.json already exists
  workspace.id   ws_prod
  workspace.name production
```

To rebind to a different workspace, pass `--force` (still with `--workspace-id`):

```bash
asgard-cli init --force --workspace-id ws_staging
```

It exits 1 in these cases:

| Case | Message |
| --- | --- |
| No config and no `--workspace-id` | `Error: creating .asgard-config.json requires --workspace-id` |
| Existing workspace is incomplete | `Error: ... is incomplete; run asgard-cli lint ..., or pass --force to reinitialise` |
| Existing config has broken JSON | `Error: parse ...` (the file is left untouched) |

`.asgard-config.json` is project configuration and belongs in version control. Read
and write it through `internal/config` (`config.Load` / `config.Save` /
`config.Find`) rather than assembling JSON inside a command.

### `lint`

Check that the config exists and is valid. Suitable for CI or a pre-commit hook:

```bash
asgard-cli lint
```

```
ok  .asgard-config.json
```

Every problem is listed and the command exits 1 in these cases:

| Case | Output |
| --- | --- |
| Config not found | `Error: no .asgard-config.json found; run asgard-cli init first` |
| Broken JSON | `Error: parse ...: invalid character ...` |
| Incomplete workspace | `- workspace.id must not be empty` (one line per problem) |

Like the other commands it uses `config.Find`, so it still locates the project
config when run from a subdirectory.

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
