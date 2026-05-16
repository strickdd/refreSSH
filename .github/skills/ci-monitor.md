# CI Monitoring Skill

## Purpose
Monitor GitHub Actions CI/CD pipelines locally using the `gh` CLI. Provide proactive status checks and failure diagnosis without requiring developers to manually check the GitHub web UI.

## Prerequisites
- `gh` CLI installed and authenticated (`gh auth status`)
- Repository accessible via `gh`
- Target PR number or branch identified

## Procedures

### Quick Status Check
```bash
gh run list --repo strickdd/refreSSH --branch <branch-or-pr> --json status,conclusion,createdAt,title --limit 5
```

### Detailed Run View
```bash
# List runs
gh run list --repo strickdd/refreSSH --branch <branch> --json status,conclusion,createdAt,id --limit 3

# View a specific run
gh run view <run_id> --repo strickdd/refreSSH --json status,conclusion,jobs --jq '.jobs[] | {name: .name, status: .status, conclusion: .conclusion}'
```

### Failed Job Logs
```bash
gh run view <run_id> --repo strickdd/refreSSH --log-failed
```

### Watch a Run to Completion
```bash
gh run watch <run_id> --repo strickdd/refreSSH --timeout 300
```

### Rerun Failed Jobs
```bash
gh run rerun <run_id> --failed --repo strickdd/refreSSH
```

### PR Status Summary
```bash
# Get PR number from branch
PR_NUM=$(gh pr list --repo strickdd/refreSSH --head <branch> --json number --jq '.[0].number')

# Get latest run for the PR
RUN_ID=$(gh run list --repo strickdd/refreSSH --pr $PR_NUM --json id --jq '.[0].id')

# Check status
gh run view $RUN_ID --repo strickdd/refreSSH
```

## Failure Diagnosis Workflow

1. **Identify the failure**: Run `gh run view <id> --log-failed`
2. **Categorize**:
   - Compile error → Check syntax, missing imports, type mismatches
   - Test failure → Identify the specific test, run it locally with `go test -v -run <TestName>`
   - Lint failure → Run `golangci-lint run ./...` locally
   - Timeout → Check for goroutine leaks or blocking operations
3. **Compare**: `git diff origin/main...HEAD` to see if the change could cause the failure
4. **Recommend**: Suggest specific fix or whether it's a pre-existing flake

## Report Template
```
## CI Status: [PASSED|FAILED|IN PROGRESS]
Branch: <branch>
Run ID: <id>
Started: <timestamp>

## Results
- Job 1: [passed/failed/skipped]
- Job 2: [passed/failed/skipped]

## Failure Details (if applicable)
[Key error message from failing job]

## Recommendation
[Specific next step: fix X, rerun Y, investigate Z]
```
