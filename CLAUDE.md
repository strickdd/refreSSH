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

## Agent Personas

This project uses specialized AI personas defined in `.github/agents/`. Invoke a persona by naming it explicitly at the start of your request (e.g., "Vera, review this diff" or "Riley, monitor the CI build for PR #12").

| Persona | Role | When to Invoke |
|---------|------|----------------|
| **Marcus** | TPM — planning, triaging, beads management | Breaking down features, managing backlog, release planning |
| **Naomi** | Go Expert — systems programming, PTY, TUI, config | Implementing features, fixing bugs, writing Go code |
| **Jarnathan** | DevSecOps — security, CI/CD, deployment implementer | Setting up CI/CD, writing deployment scripts, security features |
| **Penelope** | DevSecOps reviewer — adversarial security assessment | Reviewing security-critical code, auth, crypto, attack surface analysis |
| **Vera** | Adversarial code reviewer — design, data flow, correctness | Pre-PR code review, architecture review, data flow analysis, edge case discovery |
| **Riley** | DevOps monitor — local CI/CD build watcher | Monitoring PR builds, diagnosing CI failures, checking Copilot comments |

**Invocation rules:**
- Name the persona explicitly: "Penelope, review this PR" triggers the security reviewer
- Use **Vera** for general code review (design, correctness, data flow)
- Use **Penelope** when security is a concern (auth, crypto, injection, privilege escalation)
- Use **Vera + Penelope** together for PRs with security impact
- Use **Riley** to watch CI builds after pushing changes
- Use **Marcus** for task planning and beads management

## Build & Test

```bash
# Install dependencies
go mod tidy

# Run linter
golangci-lint run --timeout=5m ./...

# Run all tests (parallel, with race detector, across all platforms)
go test -v -race -p 4 ./...

# Build
go build ./cmd/refressh

# Cross-platform build
GOOS=linux GOARCH=amd64 go build -o dist/refressh_linux_amd64 ./cmd/refressh
GOOS=windows GOARCH=amd64 go build -o dist/refressh_windows_amd64.exe ./cmd/refressh
GOOS=darwin GOARCH=arm64 go build -o dist/refressh_darwin_arm64 ./cmd/refressh
```

## Architecture Overview

**refreSSH** is a CLI tool and daemon for managing remote SSH sessions with a local-first, security-focused design.

**Core components:**
- **CLI** (`cmd/refressh/`, `internal/cli/`) — Cobra-based CLI with subcommands for attach, create, daemon, list, sandbox, stop, tui, and ui
- **Daemon** (`internal/daemon/`) — Background process that manages session lifecycles, PTY allocation, and output broadcasting to multiple clients
- **REST API + WebSocket** (`internal/api/`) — Local-only (127.0.0.1) server with auth middleware, REST endpoints for session management, and WebSocket for PTY data streaming
- **TUI** (`internal/tui/`) — Bubble Tea-based terminal UI with tab navigation, command mode, scrollback buffer, and primary/VIEW-only control
- **Config** (`internal/config/`) — OS-standard config locations (AppData\.refreSSH on Windows, ~/.refreSSH on Unix)
- **Sandbox** (`cmd/refressh/sandbox.go`, `deploy/sandbox/`) — Docker-based containerized sandbox deployments

**Key flows:**
1. CLI `refressh attach` connects to a running daemon via the local REST API
2. Daemon allocates a PTY, spawns a shell, and broadcasts output to all connected clients
3. WebSocket provides real-time bidirectional PTY data for TUI sessions
4. Primary controller model ensures only one client has input access; others are VIEW-only

## Conventions & Patterns

- **Shell Commands:** 
  - **PowerShell (Windows):** Use `;` instead of `&&` for chaining commands (e.g., `cmd1; cmd2`).
  - **Unix (Linux/macOS):** Use `&&` for conditional chaining.
  - **Guidance:** Always determine the current OS before executing chained commands. Prefer single commands or multiple tool calls when possible to avoid shell-specific syntax issues.
- **Go Patterns:** Use standard `gofmt` and idiomatic Go practices. Standard library first, minimal external dependencies.
- **Platform-Specific Code:** Use `_windows.go` and `_unix.go` suffixes only when necessary, hidden behind clean interfaces.
- **Configuration:** Windows uses `AppData\.refreSSH`, Unix uses `~/.refreSSH`.
- **Error Handling:** Explicit error handling — no `panic` unless truly unrecoverable.
- **Naming:** Use "refreSSH" (proper casing) in all documentation and UI text.
- **Security:** Bind APIs to `127.0.0.1` by default. No sensitive data in logs.
