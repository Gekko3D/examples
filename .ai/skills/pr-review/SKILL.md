---
name: pr-review
description: Perform a comprehensive pull request code review as a Senior Staff Engineer. Reviews for correctness, security, performance, architecture, maintainability, error handling, and test coverage. Auto-detects PR context from git or fetches from Jira MCP. Outputs structured review with severity ratings and merge recommendation.
---

# Skill: PR Review

## Description
Perform a comprehensive pull request code review as a Senior Staff Engineer. Identifies correctness issues, security risks, performance concerns, architectural problems, and missing tests. Can auto-detect PR context from the current repo or fetch ticket details from a tracker MCP.

## When to Invoke
- Engineer says "review PR", "review this PR", "code review"
- Engineer says "review PR for PROJ-1234"
- Engineer says "review my changes before I push"
- Engineer says "check my code"

---

## Instructions

### Step 1: Gather PR Context

**Option A — Auto-detect from current repo (preferred):**
If you are already inside the project directory:
1. Run `git branch --show-current` → this is the PR branch
2. Run `git log --oneline -1` → latest commit summary
3. Determine the base branch:
   - Check if `main` exists: `git branch -r | grep origin/main`
   - If not, check for `develop`: `git branch -r | grep origin/develop`
   - If neither, ask the engineer
4. Read `.ai/context.md` for project context
5. Read `.ai/conventions.md` for coding standards to review against

**Option B — From ticket ID (if MCP available):**
If the engineer provides a ticket ID (e.g. "review PR for PROJ-1234"):
1. Fetch ticket via MCP → get summary, description, acceptance criteria
2. Look for the feature context file: `.ai/features/PROJ-1234-*.md`
3. Read the feature context to understand: objective, scope, what files should have changed, what tests were expected
4. Find the branch from feature context metadata
5. Checkout the branch if not already on it

**Option C — Manual input (fallback):**
If neither auto-detect nor MCP works, ask the engineer for:
- PR branch name
- Base branch (default: main)
- Short description of the change

### Step 2: Compute the Diff

```
git diff origin/[base-branch]...HEAD
git diff --name-only origin/[base-branch]...HEAD
git diff --stat origin/[base-branch]...HEAD
```

Note the scope: number of files changed, lines added/removed.

### Step 3: PR Intent Summary

Before reviewing line-by-line, write a high-level summary:

```
### PR Intent Summary
- **Purpose:** [what this change does and why]
- **Affected Modules:** [list of subsystems/directories touched]
- **Scope:** [number of files, lines changed]
- **Risk Areas:** [what could go wrong]
```

If a feature context file exists, validate that the changes match the stated scope. Flag any files changed that are NOT in scope.

### Step 4: Detailed Code Review

Review the diff across these categories. For each finding, note the file, line (if identifiable), severity, and a suggested fix.

**4.1 Correctness**
- Logic bugs, off-by-one errors
- Null/undefined handling
- Boundary conditions
- Concurrency / race conditions
- Incorrect assumptions about data types or formats

**4.2 Security**
- SQL injection (are bind variables used? — check against `.ai/conventions.md`)
- Command injection
- Hardcoded secrets, tokens, API keys
- Authentication/authorization gaps
- Insecure file operations
- Missing input validation/sanitization

**If the diff includes any files containing `application/pages/` or `application/shared_components/` (APEX export files):**
Auto-discover the export directory and run apex-security-scan:
```bash
find . -maxdepth 3 -name "create_application.sql" | sed 's|/application/create_application.sql||' | sed 's|^\./||'
python .ai/tools/apex_scan.py <discovered_dir> --format text
```
Include all findings in this Security section, mapped as: CRITICAL/HIGH → Critical/High severity, MEDIUM → Medium severity.

**4.3 Performance**
- Inefficient loops or N+1 queries
- Blocking I/O in async contexts
- Large memory allocations
- Unnecessary network calls
- Missing connection cleanup (database, HTTP clients)

**4.4 Maintainability**
- Readability and naming clarity
- Code duplication
- Function complexity (too long, too many branches)
- File size (new files over 300 lines should be flagged)

**4.5 Architecture**
- Layering violations (e.g. DB calls in controllers)
- Tight coupling between modules
- Breaking domain boundaries
- Inconsistency with patterns in `.ai/conventions.md`
- Changes that should have an ADR but don't

**4.6 Error Handling**
- Swallowed exceptions (catch blocks that do nothing)
- Missing error logging
- Silent failures
- Database connections not released in finally blocks (check project patterns)

### Step 5: Test Coverage Review

Check:
- Are new features/endpoints tested?
- Are edge cases covered?
- Are failure/error scenarios tested?
- If the feature context has a Testing section, is everything checked off?
- Are regression tests included for bug fixes?

Suggest specific tests that are missing.

### Step 6: Documentation Compliance

If a feature context file exists for this work, verify:
- All dependency changes logged in Dependency Delta
- All API changes logged in API Delta
- All config changes logged in Configuration Delta
- CHANGELOG.md updated under `[Unreleased]`

Flag any documentation gaps.

### Step 7: Suggested Patches

For Critical and High severity findings, provide example fixes:

```
### Suggested Patch
**File:** [path]
**Issue:** [what's wrong]

Before:
[original code]

After:
[fixed code]

**Why:** [explanation]
```

### Step 8: Structured Review Output

Present the complete review in this format:

```
---

## PR Review: [TICKET_ID] — [Short Description]

### Context
- **Repo:** [current directory]
- **Branch:** [pr-branch] → [base-branch]
- **Files Changed:** [count]
- **Lines:** +[added] / -[removed]

---

### Risk Overview

| Category | Risk Level |
|---|---|
| Architecture | [None / Low / Medium / High] |
| Security | [None / Low / Medium / High] |
| Performance | [None / Low / Medium / High] |
| Correctness | [None / Low / Medium / High] |

---

### Findings

#### [Finding 1 Title]
- **Severity:** [Critical / High / Medium / Low]
- **File:** [path]
- **Line:** [if identifiable]
- **Issue:** [description]
- **Fix:** [suggestion or patch]

#### [Finding 2 Title]
...

---

### Missing Tests
1. [specific test to add]
2. [specific test to add]

---

### Documentation Gaps
- [any missing deltas or changelog entries]

---

### Suggested Refactors
1. [refactoring opportunity]
2. [refactoring opportunity]

---

### Overall Assessment

**Risk Level:** [Low / Medium / High]

**Merge Recommendation:**
- [ ] Approve
- [ ] Approve with minor fixes
- [ ] Request changes — [reason]

**Summary:** [2-3 sentence overall assessment]
```

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
