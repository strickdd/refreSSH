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

refreSSH uses specialized AI personas for different review and implementation tasks. Invoke agents by name prefix (e.g., `"Vera, review this diff"`).

| Agent | Role | Files |
|-------|------|-------|
| **Marcus** (TPM) | Planning, triaging, beads management | `.github/agents/tpm.md` |
| **Naomi** (Go Expert) | Systems programming and PTY management | `.github/agents/go-engineer.md` |
| **Jarnathan** (DevSecOps) | Security, CI/CD, deployment implementation | `.github/agents/devsecops.md` |
| **Vera** (Adversarial Review) | General code correctness, data flow, design flaws | `.github/agents/vera.md`, `.github/skills/code-review.md` |
| **Penelope** (Security Deep-Dive) | Security analysis: auth, crypto, attack surface | `.github/agents/penelope.md` |
| **Riley** (CI Monitor) | Local CI/CD monitoring via `gh` CLI | `.github/agents/riley.md`, `.github/skills/ci-monitor.md` |

### Invocation Rules
- Call agents by name: `"Vera, review this diff"` or `"Penelope, audit auth flow"`
- Vera + Penelope can be called together for security-impact changes
- Riley operates via `.github/skills/ci-monitor.md` skill using local `gh` CLI
- Skills are invoked via `bd skills list` and selected by name

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
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
