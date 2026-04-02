# CLAUDE.md — hashtray-Go

## Project Overview

**hashtray** is an OSINT tool for Gravatar. It finds Gravatar accounts from email addresses and locates email addresses from Gravatar usernames or hashes. This repository is a Go rewrite of the original Python tool ([balestek/hashtray](https://github.com/balestek/hashtray)).

## Goal

Produce a single, statically-linked Go binary that runs on Linux, macOS, and Windows with no external runtime dependencies. See Issue #1.

## Repository Layout (target)

```
.
├── cmd/hashtray/       # main package — CLI entry point
├── internal/
│   ├── gravatar/       # Gravatar API client & HTML scraper
│   ├── enumerator/     # email enumeration logic
│   ├── permutator/     # combination/permutation generator
│   └── elements/       # element extraction from Gravatar profiles
├── data/               # embedded JSON domain lists
├── .github/workflows/  # CI (test) and Release (build + publish) workflows
├── go.mod / go.sum
├── Makefile
└── README.md
```

## Build & Test Commands

```bash
make build      # build binary
make test       # run all tests
make lint       # run golangci-lint
make clean      # remove build artifacts
```

## Conventions

- Use `//go:embed` to embed domain list JSON files into the binary.
- Use standard library where possible; minimise third-party dependencies.
- Acceptable dependencies: `github.com/spf13/cobra` (CLI), `github.com/fatih/color` (terminal colors), `golang.org/x/net/html` (HTML parsing).
- All packages under `internal/` — not intended for external import.
- Tests live alongside source files (`*_test.go`).
- CGO disabled (`CGO_ENABLED=0`) for all builds.
- Target platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.

## Workflow

See `AGENTS.md` for the Claude Code / Codex division-of-labor protocol.
