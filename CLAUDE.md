# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`stdidx` is a Go CLI that syncs a remote Git repository of Markdown coding standards into a dedicated, user-level directory (default `~/.stdidx/`) and builds a hierarchical `standards-tree.yaml` index alongside it from each file's YAML frontmatter. The generated tree is consumed by AI agents, which traverse it (root → leaves, matching on `scope` and `topics`) to pull only the standards relevant to the current task into context. The clone location is overridable per invocation via `--clone-path` / `-p`.

The published binary name is `stdidx`, but the Go module is `github.com/PSauerborn/stdidx` and the repo directory is `std-index` — these are intentional and should not be "fixed."

## Commands

```bash
make run-tests   # go test ./...
make coverage    # generates and opens coverage.html
make lint        # go fmt + go mod tidy + golangci-lint
make scan-secrets # regenerate .secrets.baseline via detect-secrets
```

Single test: `go test -run TestSync/success_-_existing_directory ./...` (subtest names use `_` for spaces).

Build: `go build -o stdidx`. Cross-platform release binaries live in `bin/` and are produced by the `release.yaml` GitHub workflow, which is `workflow_dispatch` only.

Requires Go ≥ 1.25.

## Architecture

The entire CLI is a single `package main` at the repo root — there is no `cmd/` or `internal/` layout. Files are split by responsibility:

- `main.go` — `urfave/cli/v3` command wiring (`sync`, `index`), the top-level `Sync` / `Index` functions, and two path helpers: `DefaultClonePath()` (returns `$HOME/.stdidx`, falling back to the relative `.stdidx` if the home dir can't be resolved) and `ExpandPath()` (expands a leading `~` or `~/` in user-supplied paths).
- `git.go` — `GitCloner` interface and the `ExecGitCloner` implementation that shells out to `git clone`. The interface exists so tests can swap in `MockGitCloner` (defined in `git_test.go`, which copies a fixture directory instead of cloning).
- `tree.go` — frontmatter parsing (`ExtractMDHeader`), filesystem walk (`ParseMDDocuments`), hierarchy construction (`BuildHierarchy`), and YAML serialization (`GenerateStandardsTree`). `GenerateStandardsTree` takes an explicit output path so callers (and tests) control where the tree is written.
- `types.go` — `StandardsHeader`, `Node`, `StandardsTree`, `GitRepository`. Note `Node.Parent` is tagged `yaml:"-"` because the YAML tree encodes the hierarchy via nested `Children`, not back-pointers.
- `constants.go` — `DefaultCloneDirName = ".stdidx"`, `TreeFileName = "standards-tree.yaml"`, and `SuggestedAgentInstructionsTemplate` (a printf template with one `%s` placeholder, substituted at print time with the absolute path to the generated tree).

### Frontmatter → tree flow

1. `ParseMDDocuments` walks the clone root and calls `ExtractMDHeader` on every `.md` file.
2. `ExtractMDHeader` runs `validator.Struct` on the parsed header. **Validation failure returns `(nil, nil)`, not an error** — the file is silently skipped and logged as a warning. This is intentional: arbitrary `.md` files (READMEs, etc.) in a standards repo must not break indexing.
3. A header's `parent` field is a path *relative to the clone root* in the source file, but the in-memory `nodes` map is keyed by the same root-prefixed path that `filepath.WalkDir` produces for `Path`. `ParseMDDocuments` rewrites each `Parent` to `path.Join(root, *parent)` so parent lookup in `BuildHierarchy` succeeds. If you change how paths are constructed in either place, both must stay in sync or the hierarchy collapses to all-roots.
4. `BuildHierarchy` does two passes: first builds the `nodes` map, then attaches each node to its parent's `Children`. Roots (no `Parent`) are collected separately. Missing parents are logged and the node is dropped.
5. `sortChildren` sorts every level alphabetically by `Title` — the expected output in `tests/fixtures/expected_tree.yaml` is byte-for-byte compared, so any change to sort order or YAML field order will break tests.

### Tests

- `tests/mock_repository/` is the input fixture (a small tree of valid standards files plus `INVALID.md` to exercise the skip path).
- `tests/fixtures/expected_tree.yaml` is the golden output. `TestSync` and `TestIndex` both assert byte-equality against it.
- Tests write `standards-tree.yaml` into `tests/tmp/` (the same temp directory used as the clone target), so a `defer os.RemoveAll("tests/tmp")` cleans up everything — no manual cleanup needed even if a test fails partway.
- `tests/tmp/` is created and cleaned up per test; don't commit it.

## Conventions

- Logging: `sirupsen/logrus` with `WithFields` for structured context, never `fmt.Println`.
- The `examples/` directory ships a real-world standards repo plus its generated `standards-tree.yaml` as a reference for users — keep them in sync if you change the tree format.

## Development

The global @~/.claude/CLAUDE.md contains detailed instructions on how to develop applications, coding standards that need to be followed and more. Make sure to review the @~/.claude/CLAUDE.md file before planning or making any code changes. These instructions __MUST__ be followed. Instructions in the local @CLAUDE.md file supersede any instructions provided globally.
