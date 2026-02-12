# `stdidx` - Standards Indexing For Agents

A lightweight, local-first CLI tool that

1. Syncs coding standards from a remote Git repository to your project directory.
2. Indexes Markdown files from Frontmatter declarations into a YAML tree that AI coding agents can traverse to find relevant coding standards.

## Quickstart

Run the `stdidx sync` command to sync a standards repository to your project directory and generate the index tree file at `standards-tree.yaml`.

```bash
stdidx sync -r https://github.com/example/coding-standards.git
```

Instruct your agent to use the `standards-tree.yaml` file to find relevant coding standards based on the type of task being worked on, the file type, and the frameworks/tools being used.

```txt
When working on a task, consult the standards tree in standards-tree.yaml
to find applicable coding standards.

1. Always start at the root nodes. Read any root node whose scope
   matches the files you're working with or whose scope is "*".

2. For each node you read, check its children. Descend into a child
   if its scope or tags match your current context.

3. Stop descending a branch when no children match your context.

4. Collect all matching nodes from root to leaf. Standards at every
   level in the path apply — a child does not replace its parent,
   it adds to it.

5. If a child standard contradicts a parent, the child takes precedence.

To determine if a node matches your context:
- description: compare the description of the node to the task you're working on
- scope: compare against the file extensions you're editing
  ("*.py", "*.ts", "*" matches everything)
- topics: compare against the project's detected frameworks/tools
  (e.g. if package.json has "react" as a dependency, the "react"
  topic matches)
```

## Overview

Coding standards are essential for ensuring that agents deliver code that is high quality and consistent. Maintaining a repository of standards is a must for anyone that wants to make the most of agentic engineering. However, this introduces three major problems:

1. Syncing - manually copying standards files from one repository to another is a chore.
2. Searchability - standards are often spread across multiple files, making it hard to find relevant standards when working on a specific task without.
3. Context Limits - Agents have small context windows. You don't want to fill it with irrelevant coding standards. If an agent is working on a task to build a Python REST API, Python standards about ETL pipelines do nothing but fill precious context.

`stdidx` is an attempt to tackle these problems. It syncs standards from a remote Git repository to your project directory and indexes files into a YAML tree that AI agents can traverse to find relevant coding standards based on the type of task being worked on, the file type, and the frameworks/tools being used. This ensures that your standards files are in sync, and agents can find relevant coding standards quickly and efficiently.

### How does `stdidx` sync standards?

`stdidx` clones the standards repository to a local directory using a simple `git clone` command. By default, it clones to `.stdidx` in the root of your project directory.

### How does `stdidx` index files?

`stdidx` iterates over all MD files in the standards repository and looks for YAML frontmatter defining the required metadata for the index in each file. If found, it parses the frontmatter and adds the file to the index tree.

For example, a standards file containing standards for writing REST APIs in Go might look like this:

```md
---
title: Golang REST API Standards
description: Standards for writing REST APIs in Go.
parent: golang/GENERAL.md
scope: '*.go'
topics:
- golang
- api
- rest
- gin-gonic
---

Use the following guidelines when writing REST APIs in Go:

- Use `gin-gonic` for REST APIs.
- Use `logrus` for structured logging.

```

The `parent` field references the relative path to a parent standard file. This is used to build the hierarchy of the index tree. If not specified, the file is added as a root node.

For a complete example of a working set of standards files and the generated index, see the [examples/](examples/) directory.

### What data can I include in my standards frontmatter?

The following fields can be included in the frontmatter of a standards file:

| Field | Required | Description |
| --- | --- | --- |
| `title` | Yes | Human-readable name of the standard |
| `description` | Yes | Brief summary of what the standard covers |
| `scope` | Yes | Glob pattern for file extensions the standard applies to (e.g. `*.go`, `*.py`, `*` for all) |
| `topics` | Yes | List of frameworks, tools, or domains the standard relates to (e.g. `golang`, `rest`, `gin-gonic`) |
| `parent` | No | Relative path to a parent standard file — used to build the hierarchy |

If a markdown file does not contain valid frontmatter, it will simply be ignored and wont be included in the index.

## Installation

### Pre-built Binaries

Pre-built binaries for macOS and Linux are available in the [`bin/`](bin/) directory of this repository and are updated with each release. Download the binary that matches your platform and architecture:

| Binary | OS | Architecture |
| --- | --- | --- |
| `stdidx-darwin-amd64` | macOS | Intel (x86_64) |
| `stdidx-darwin-arm64` | macOS | Apple Silicon (M-series) |
| `stdidx-linux-amd64` | Linux | x86_64 |
| `stdidx-linux-arm64` | Linux | ARM64 |

### Go Install

If you have Go ≥ 1.25 installed, you can install directly:

```bash
go install github.com/PSauerborn/stdidx@latest
```

### Build from Source

```bash
git clone https://github.com/psauerborn/std-index.git
cd std-index
go build -o stdidx
```

## Usage

### `stdidx sync` — Clone and Index

Clone a standards repository, parse all Markdown frontmatter, and generate `standards-tree.yaml` in one step:

```bash
stdidx sync --repository <git-url> [--branch <branch> | --tag <tag>]
```

**Flags:**

| Flag | Alias | Required | Description |
| --- | --- | --- | --- |
| `--repository` | `-r` | Yes | Git repository URL to clone |
| `--branch` | `-b` | No | Branch to checkout |
| `--tag` | `-t` | No | Tag to checkout |

> **Note:** `--branch` and `--tag` are mutually exclusive.

**Example:**

```bash
stdidx sync -r git@github.com:your-org/coding-standards.git -b main
```

The repository is cloned to a local `.stdidx/` directory (overwritten on each sync), the tree is built, and the output is written to `standards-tree.yaml`.

### `index` — Re-index an Existing Clone

If the standards repository has already been cloned (i.e. the `.stdidx/` directory exists), you can regenerate the tree without re-cloning:

```bash
stdidx index
```

### Integrating with Your Agent

After running `sync` or `index`, `std-index` prints suggested instructions that you can add to your AI agent's prompt or configuration. The instructions tell the agent how to walk the generated tree:

```
When working on a task, consult the standards tree in standards-tree.yaml
to find applicable coding standards.

1. Always start at the root nodes. Read any root node whose scope
   matches the files you're working with or whose scope is "*".

2. For each node you read, check its children. Descend into a child
   if its scope or tags match your current context.

3. Stop descending a branch when no children match your context.

4. Collect all matching nodes from root to leaf. Standards at every
   level in the path apply — a child does not replace its parent,
   it adds to it.

5. If a child standard contradicts a parent, the child takes precedence.
```

## Contributing

Contributions are welcome! Please follow these steps:

1. **Clone** the repository:
   ```bash
   git clone https://github.com/psauerborn/std-index.git
   ```
2. **Create a feature branch** off `master`:
   ```bash
   git checkout -b feature/my-feature
   ```
3. **Make your changes.** Ensure all tests pass and linting is clean:
   ```bash
   make run-tests   # run the test suite
   make lint         # format, tidy, and lint
   ```
4. **Commit** with a clear, descriptive commit message.
5. **Open a Pull Request** against `master`.

### Development Prerequisites

- Go ≥ 1.25
- [pre-commit](https://pre-commit.com/) (optional, but recommended — hooks run formatting, linting, secret scanning, and tests automatically)

Install the pre-commit hooks:

```bash
pre-commit install
```

### Running Tests

```bash
make run-tests
```

To generate an HTML coverage report:

```bash
make coverage
```

## License

This project is licensed under the [Apache License 2.0](LICENSE).
