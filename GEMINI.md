# Project Instructions for AI Agents

This file provides instructions and context for AI coding agents working on this project.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:7510c1e2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
bd dolt push          # Push beads data to remote
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
    ```bash
    git pull --rebase
    bd dolt push
    git push
    git status  # MUST show "up to date with origin"
    ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->

## Tech Stack

- **Language:** Go (Golang) 1.24+
- **UI:** TUI (Terminal User Interface) via CLI, and a local-only Web UI embedded in the binary.
- **Frontend (Web):** xterm.js for terminal emulation, served via `go:embed`.

## Architecture Overview

**refreSSH** is a cross-platform (Linux, Windows, macOS) user-level background daemon designed to host and manage persistent terminal sessions, specifically optimized for long-running AI CLI agents.

1. **Daemon (Server):** Runs in the background as a user-level process. It manages Pseudo-Terminals (PTYs), spawns child processes, and maintains session state. It exposes a local API (REST + WebSocket on `127.0.0.1`) for clients to connect.
2. **CLI / TUI Client:** The primary interface for users to list, attach to, and manage sessions via Bubble Tea-based TUI.
3. **Web UI:** A local-only web interface served directly by the Daemon on `127.0.0.1`. It provides full interactive terminal emulation and session management, mirroring the capabilities of the CLI/TUI.

## Lifecycle Management

The daemon starts on-demand when the user runs the first `refressh` command. It persists across SSH disconnects, keeping all child processes alive until explicitly terminated or the host system restarts.

## Build & Test

```bash
# Install dependencies
go mod tidy

# Run linter
golangci-lint run --timeout=5m ./...

# Run all tests (parallel, with race detector)
go test -v -race -p 4 ./...

# Build
go build ./cmd/refressh

# Cross-platform build
GOOS=linux GOARCH=amd64 go build -o dist/refressh_linux_amd64 ./cmd/refressh
GOOS=windows GOARCH=amd64 go build -o dist/refressh_windows_amd64.exe ./cmd/refressh
GOOS=darwin GOARCH=arm64 go build -o dist/refressh_darwin_arm64 ./cmd/refressh
```

## Conventions & Patterns

- **Formatting:** Standard `gofmt` / `goimports`. Idiomatic Go, standard library first.
- **Security:** The Web UI and REST API must enforce strict local-only binding (`127.0.0.1`) to prevent external access. No sensitive data in logs.
- **State Management:** Robust handling of PTY state, especially when clients abruptly disconnect. Primary controller model ensures only one client has input access.
- **Shell Commands:** 
  - **PowerShell (Windows):** Use `;` instead of `&&` for chaining commands (e.g., `cmd1; cmd2`).
  - **Unix (Linux/macOS):** Use `&&` for conditional chaining.
  - **Guidance:** Always determine the current OS before executing chained commands. Prefer single commands or multiple tool calls when possible to avoid shell-specific syntax issues.
