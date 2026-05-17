# Skill: CI/CD Monitor

## Purpose
Standard procedure for Riley (DevOps Engineer) to monitor and diagnose CI/CD builds locally using a sleep-and-check pattern.

## Prerequisites
- GitHub CLI (`gh`) installed and authenticated
- The PR number or branch name being monitored
- The workflow name to monitor (e.g., "CI" or "Release")

## Phase 1: Sleep-and-Check Monitoring Loop

### Start Monitoring

1. **Get the branch or PR context:**
   ```bash
   # For PR
   gh pr view --json number,headRefName
   
   # For current branch
   git branch --show-current
   ```

2. **Enter the monitoring loop** (pseudo-code for the agent):
   ```
   branch = <branch name or PR head ref>
   timeout = 20 minutes
   interval = 15 seconds
   start_time = now()
   
   while (now() - start_time) < timeout:
       runs = gh run list --branch <branch> --status conclusion --limit 3 --json conclusion,status,created_at,name,id
       
       for each run in runs:
           if run.status == 'completed':
               if run.conclusion == 'success':
                   output STATUS: all_clear
                   output STATUS: monitoring_complete
                   return SUCCESS
               elif run.conclusion == 'failure':
                   output STATUS: failure_detected
                   run_id = run.id
                   goto Phase 2
               elif run.conclusion == 'cancelled':
                   output STATUS: monitoring_complete
                   output ROOT_CAUSE: Workflow was cancelled
                   return CANCELLED
               elif run.conclusion == 'timed_out':
                   output STATUS: failure_detected
                   output FAILURE_TYPE: timeout
                   run_id = run.id
                   goto Phase 2
       
       sleep <interval>
   
   output STATUS: monitoring_complete
   output FAILURE_TYPE: timeout
   output ROOT_CAUSE: Monitoring timed out after 20 minutes
   return TIMEOUT
   ```

### Key Commands for the Loop

```bash
# Check latest runs for a branch - returns completed runs with conclusion
gh run list --repo <owner>/<repo> --branch <branch> --status conclusion --limit 3 --json conclusion,status,created_at,name,id

# For PR-specific monitoring
gh run list --repo <owner>/<repo> --pr <number> --status conclusion --limit 3 --json conclusion,status,created_at,name,id
```

## Phase 2: Diagnose Failures

Once a failure is detected, diagnose it.

### 1. Get Log Output

```bash
# Download and stream failed job logs
gh run view <run_id> --repo <owner>/<repo> --log-failed > /tmp/gh-logs/log.txt
```

### 2. Analyze Logs

Search for failure patterns:

```bash
# Test failures
rg "FAIL" /tmp/gh-logs/log.txt -n

# Panics
rg "panic:" /tmp/gh-logs/log.txt -n -i

# Compilation errors
rg "cannot find package|undefined|syntax" /tmp/gh-logs/log.txt -n -i

# Lint errors
rg "golangci-lint" /tmp/gh-logs/log.txt -n -A 5

# Timeout errors
rg "timeout|deadline exceeded" /tmp/gh-logs/log.txt -n -i

# Race detector output
rg "race|DATA RACE" /tmp/gh-logs/log.txt -n -i

# Build errors
rg "build|link" /tmp/gh-logs/log.txt -n -i
```

### 3. Identify Root Cause

Determine:
- **Failure type:** test_failure | lint_error | compilation_error | timeout | race_condition | platform_specific | other
- **Exact location:** file:line from error output
- **Root cause:** what went wrong and why
- **Scope:** one-off flaky failure or consistent issue

### 4. Suggest a Fix

- **Compilation error:** Missing import, wrong type, syntax issue. Suggest the exact correction.
- **Test failure:** Assertion that failed. Suggest whether the test is wrong or the code is wrong.
- **Lint error:** Rule violated. Suggest the code change or linter config adjustment.
- **Timeout:** Operation timing out. Suggest optimization or timeout adjustment.
- **Race condition:** Shared state issue. Suggest synchronization or refactoring.
- **Platform-specific:** Platform assumption. Suggest a platform-agnostic solution.

## Phase 3: Copilot Comment Detection

```bash
# Check PR for Copilot comments
gh pr reviews <pr_number> --repo <owner>/<repo> --json body,author,submittedAt
gh pr comment <pr_number> --repo <owner>/<repo> --json body,author,createdAt
```

Filter for comments where `author.login == 'copilot'` or body contains "copilot".

Include in output:
- Comment text
- Author
- Timestamp

## Output Report

Produce the final structured report to stdout:

```
FAILURE_TYPE: <test_failure|lint_error|compilation_error|timeout|race_condition|platform_specific|none>
ROOT_CAUSE: <concise description of what failed and why>
LOCATION: <file:line or workflow job name>
SUGGESTED_FIX: <actionable recommendation>
COPILOT_COMMENTS: <none | list of copilot comments with body, author, timestamp>
STATUS: <monitoring_complete|failure_detected|all_clear>
```

## Invocation
- Call this skill when a PR is created/updated and CI runs need monitoring
- Call this skill after code changes are pushed to verify the build passes
- Call this skill when a release tag push triggers workflow runs
- Runs as a subagent within opencode — output goes to calling agent's stdout
- Uses `gh` CLI locally — no GitHub Actions workflows involved
