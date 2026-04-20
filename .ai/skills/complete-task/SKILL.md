---
name: complete-task
description: Complete and close out a finished feature or bug fix. Runs tests, updates CHANGELOG, commits, pushes branch (with approval), transitions Jira ticket to In Review via MCP, and adds a summary comment.
---

# Skill: Complete Task

## Description
Close out a completed feature or bug fix. Runs final tests, updates CHANGELOG, commits, pushes, and transitions the ticket in the project tracker.

## When to Invoke
- Engineer says "complete task", "finish up", "close out", "wrap up", "mark done"
- Engineer says "task complete for PROJ-1234"
- After implementation and testing are done and the engineer confirms readiness

---

## Instructions

### Step 1: Verify Definition of Done

Read the active feature context file and check each item:

- [ ] Code follows conventions (spot-check against `.ai/conventions.md`)
- [ ] No hardcoded secrets or environment-specific values
- [ ] Tests written for new code
- [ ] All doc-impacting changes logged as deltas in the feature context

Report any gaps to the engineer before proceeding. Fix gaps first.

### Step 2: Run Tests
1. Run the project's test command (e.g. `npm test`, `pytest`, `mvn test`)
2. If tests fail → report failures and fix before continuing
3. If all pass → proceed

### Step 3: Update CHANGELOG
1. Open `CHANGELOG.md`
2. Under `## [Unreleased]`, add an entry:
   - For features: under `### Added` → `- [[TICKET_ID]] Description`
   - For bug fixes: under `### Fixed` → `- [[TICKET_ID]] Description`
   - For breaking changes: under `### Changed` → `- [[TICKET_ID]] Description`

### Step 4: Final Commit
1. Stage all changes: `git add -A`
2. Commit: `git commit -m "[[TICKET_ID]] Final changes — [brief summary]"`

### Step 5: Push (Ask First)
1. **Ask the engineer:** "Ready to push to remote? Branch: [TICKET_ID]-short-description"
2. If approved: `git push origin [TICKET_ID]-short-description`
3. If not approved: wait

### Step 6: Update Ticket Tracker (if MCP available)
1. Attempt to transition the ticket to "In Review" (or equivalent status) via MCP
2. Add a comment to the ticket with:
   - Branch name
   - Summary of changes made
   - List of files modified
   - Link to PR (if available)
3. If MCP is not available, tell the engineer to manually update the ticket

### Step 7: Report
Summarize:
```
## Task Complete: [TICKET_ID]

### Changes Made:
- [list of files and what changed]

### Tests: All passing ✅
### CHANGELOG: Updated ✅
### Branch: [branch-name] pushed ✅
### Ticket: Transitioned to In Review ✅ (or: needs manual update)

### Next: Create PR and request review
```

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
