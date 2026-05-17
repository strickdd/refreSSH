# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd prime` for full workflow context.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work atomically
bd close <id>         # Complete work
bd dolt push          # Push beads data to remote
```

## Cross-Platform Shell Commands

Always determine the current OS before executing chained commands.

- **PowerShell (Windows):** Use `;` instead of `&&` for chaining commands (e.g., `cmd1; cmd2`).
- **Unix (Linux/macOS):** Use `&&` for conditional chaining.
- **Best Practice:** Prefer single commands or multiple tool calls when possible to avoid shell-specific syntax issues.

## Non-Interactive Shell Commands

**ALWAYS use non-interactive flags** with file operations to avoid hanging on confirmation prompts.

Shell commands like `cp`, `mv`, and `rm` may be aliased to include `-i` (interactive) mode on some systems, causing the agent to hang indefinitely waiting for y/n input.

**Use these forms instead:**
```bash
# Force overwrite without prompting
cp -f source dest           # NOT: cp source dest
mv -f source dest           # NOT: mv source dest
rm -f file                  # NOT: rm file

# For recursive operations
rm -rf directory            # NOT: rm -r directory
cp -rf source dest          # NOT: cp -r source dest
```

**Other commands that may prompt:**
- `scp` - use `-o BatchMode=yes` for non-interactive
- `ssh` - use `-o BatchMode=yes` to fail instead of prompting
- `apt-get` - use `-y` flag
- `brew` - use `HOMEBREW_NO_AUTO_UPDATE=1` env var

## Agent Personas

This project uses specialized AI personas defined in `.github/agents/`. Invoke a persona by naming it explicitly at the start of your request (e.g., "Vera, review this diff" or "Riley, monitor the CI build for PR #12"). Each persona has a specific role — choose the one (or combination) that matches the task.

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
- Use **Vera + Penelope** together for PRs with security impact — Vera reviews first, then passes to Penelope
- Use **Riley** to watch CI builds after pushing changes
- Use **Marcus** for task planning and beads management

**Agent files are located in:** `.github/agents/` — each persona has a dedicated `.md` file with full role definition.

**Skills are located in:** `.github/skills/` — each skill defines a step-by-step procedure for a specific agent task.

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
