# Skill: Adversarial Code Review

## Purpose
Standard procedure for Vera (Senior Adversarial Code Reviewer) to systematically review code changes before PR submission.

## Prerequisites
- The branch/diff to review must be checked out or accessible via `git diff`
- The parent bead/task ID for tracking findings
- Knowledge of which agent to invoke (Vera for general review, Penelope for security review)

## Step 1: Diff Analysis

Walk through every changed file systematically.

1. Get the full diff: `git diff main...HEAD` or `git diff origin/main...HEAD`
2. For each changed file:
   a. Read the **original version** — understand what existed before
   b. Read the **diff** — understand what changed
   c. Read the **new version** — understand the resulting state
   d. Ask: "What was the intent of this change? Does the implementation match the intent?"

For each file, document:
- Purpose of change (based on commit message and diff)
- Files affected and their roles
- Any files that changed but weren't mentioned in the commit message

## Step 2: Data Flow Tracing

Trace user-supplied data through the entire codebase.

1. **Identify entry points** — CLI arguments, API endpoints, WebSocket messages, file inputs, environment variables
2. **Trace the path** — follow the data through:
   - Parsing / deserialization
   - Validation (or lack thereof)
   - Business logic transformation
   - Storage / persistence
   - Output / serialization / rendering
3. **Verify trust boundaries** — at each boundary crossing, verify:
   - Input is validated and sanitized
   - Types are enforced
   - Length limits are applied
   - Encoding/decoding is handled correctly

Look for:
- Data that crosses a trust boundary without validation
- Deserialization of untrusted data
- SQL injection, command injection, template injection vectors
- Path traversal, XXE, SSRF opportunities
- Integer overflow, buffer overflow risks

## Step 3: User Flow Validation

Map every user-facing interaction and verify correctness.

For each user interaction:
1. Identify the happy path — trace from input to expected output
2. Identify failure modes — what happens when input is wrong, missing, or malicious?
3. Verify error messages — are they informative but not leaking sensitive data?
4. Verify state transitions — are all state changes atomic? Are there race conditions?
5. Verify UI/UX — does the interface prevent common user errors?

## Step 4: Design Review

Evaluate the architecture and code quality.

Check:
- **Separation of concerns** — are responsibilities clearly divided?
- **Single responsibility** — does each function/type have one clear purpose?
- **Abstraction boundaries** — are interfaces clean? Are implementation details hidden?
- **Dependency management** — are dependencies minimal and well-chosen?
- **Testability** — can the code be tested in isolation? Are mocks/stubs needed?
- **Configuration** — is the code configurable? Are defaults reasonable?

## Step 5: Edge Case Discovery

Systematically enumerate edge cases.

For every function/method, check:
- Empty input (nil, empty string, empty slice, zero value)
- Large input (huge strings, massive payloads, deep nesting)
- Malformed input (invalid types, corrupted data, unexpected formats)
- Concurrent access (goroutine safety, mutex usage, channel operations)
- Timeout/deadline handling (context cancellation, timeout propagation)
- Resource cleanup (file handles, connections, goroutine leaks)

## Step 5.5: Linting

Run the project linter and treat violations as findings.

```bash
# Run golangci-lint with the project config
golangci-lint run ./...
```

Check:
- All lint errors are genuine and fixed, not suppressed
- `gosec` flagged no new security-related lint violations
- No new `staticcheck` or `revive` warnings introduced
- Linter configuration (`.golangci.yml`) was not weakened to suppress real issues
- If a lint violation is intentional, verify there is a `nolint` comment with explanation

Any lint failure is a **finding** at minimum severity LOW (or HIGHER if it flags a real issue like unchecked errors, unused variables, or unsafe patterns).

## Step 6: Test Gap Analysis

Review existing tests and identify gaps.

Check:
- Are all changed functions covered by tests?
- Do tests cover happy paths AND failure modes?
- Are there integration tests for complex interactions?
- Are mocks/stubs appropriate and realistic?
- Are there race condition tests for concurrent code?
- Are there fuzz tests for parsing/validation code?

## Step 7: Output Reporting

Produce structured findings.

Format each finding as:

```
FINDING #[number]
Severity: CRITICAL | HIGH | MEDIUM | LOW
Location: file_path:line_number
Category: design | data-flow | user-flow | edge-case | test-gap | correctness
Description: <clear description of the issue>
Impact: <what happens if this is not fixed>
Suggested Fix: <actionable recommendation>
```

Sort findings by severity (CRITICAL first).

## Invocation
- Call this skill when reviewing pre-PR changes
- Use for Vera persona (general adversarial review)
- If security impact is suspected, also invoke Penelope's security review after Vera's findings
