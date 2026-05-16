# Code Review Skill

## Purpose
Perform thorough, adversarial code review following the structured Vera methodology, including mandatory linting.

## Prerequisites
- `golangci-lint` installed and available in PATH
- Git repository with clean working tree
- Target branch or diff identified

## Procedure

### Step 1: Setup
```bash
# Ensure we're on the right branch
git fetch origin
git diff origin/main...HEAD
```

### Step 2: Lint (MANDATORY)
Run the linter before any review. If lint fails, report and stop — do not review code with known lint violations.

```bash
golangci-lint run ./...
```

### Step 3: Context Reading
Read the full diff with surrounding context:
- Understand the feature or fix being implemented
- Identify all modified files and their relationships
- Read the commit messages for intent

### Step 4: Data Flow Tracing
For each modified code path:
1. Find the entry point (API handler, CLI command, event handler)
2. Trace data transformation at each step
3. Identify exit points (DB write, HTTP response, file output)
4. Check for data corruption at each transformation

### Step 5: Edge Case Stress
For each code path, consider:
- Empty input vs. malformed input vs. very large input
- Network timeouts and partial responses
- Concurrent access to shared state
- File descriptor and connection limits
- Disk space exhaustion during writes

### Step 6: Error Path Verification
For every `if err != nil` block:
- Is the error returned or logged appropriately?
- Are resources cleaned up (defer Close, rollback)?
- Does the error handling match the function's contract?

### Step 7: Report
Summarize findings:
- BLOCKER: Must fix before merge
- MAJOR: Should fix before merge
- MINOR: Can merge with follow-up
- STYLE: Cosmetic or convention

## Output Template
```
## Review Summary
[1-2 sentence overview of what was reviewed]

## Findings

### BLOCKER: [title]
[file:line] [description] [recommendation]

### MAJOR: [title]
[file:line] [description] [recommendation]

### MINOR: [title]
[file:line] [description]

### Lint
[golangci-lint output or "All clean"]
```
