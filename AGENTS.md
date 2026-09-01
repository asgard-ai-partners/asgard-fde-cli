# AGENTS.md

Rules for AI agents working in this repo. Human contributors follow the same ones.

## Write everything in English

Comments, user-facing strings, error messages, command help, documentation and
commit messages are all written in English.

```go
// Save writes cfg to path.
return fmt.Errorf("write %s: %w", path, err)
```

## Plain ASCII only, no emoji

Source code, comments, command output, documentation and commit messages stay in
printable ASCII. No emoji, and no decorative Unicode either (no check marks, arrows
or box-drawing characters). They render inconsistently across terminals, get
mangled in logs and CI output, and are awkward to grep for.

Mark status with words or ASCII punctuation instead:

```
ok  .asgard-config.json
- workspace.id must not be empty
```

## Put generated files in `.out/`

Anything produced by a command or a test that does not belong in version control
goes into `.out/` at the repo root. Do not scatter such files across the repo, and
do not write them to `/tmp`. This covers:

- files written by tests, fixture snapshots, actual output for golden files
- coverage and profiling reports (`coverage.out`, `*.pprof`)
- throwaway debug scripts and scratch programs
- command output, logs, sample data downloaded for inspection
- binaries built by hand

```bash
mkdir -p .out
go build -o .out/asgard-cli ./cmd/asgard-cli
go test -coverprofile=.out/coverage.out ./...
go tool cover -html=.out/coverage.out -o .out/coverage.html
```

`.out/` is in `.gitignore` and can be deleted at any time:

```bash
rm -rf .out
```

One exception: a Go test that only needs a scratch directory should use
`t.TempDir()`, which cleans up on its own and is more reliable than managing paths
under `.out/`. Reserve `.out/` for output a person still needs to look at **after**
the test finishes, such as an actual-vs-expected dump from a failure.
