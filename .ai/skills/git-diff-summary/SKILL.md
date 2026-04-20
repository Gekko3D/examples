---
name: git-diff-summary
description: Generate a concise working-tree git summary and changed-files list from the current diff. Read-only; use during commit workflows after tests pass and before running `git commit`.
---


# Git Commit Summary

When asked to create a git commit summary, follow this workflow.

This skill is **read-only**.
It summarizes the current working tree and index.
It must **not** create Jira comments or perform other write-side effects.

## 1. Collect Working-Tree Diff Context (Deterministic)

Use git commands to understand the scope of current unstaged, staged, and
untracked changes:

```bash
# Unstaged high-level file summary
git diff --stat

# Staged high-level file summary
git diff --cached --stat

# Unstaged file status list
git diff --name-status

# Staged file status list
git diff --cached --name-status

# Untracked files
git status --porcelain

# Unstaged full patch
git -P diff

# Staged full patch
git -P diff --cached
```

From this, identify:
- The primary area(s) of change (e.g., auth, persistence, tests, frontend)
- The main intent (e.g., fix bug, add feature, refactor, improve tests)

Scope rules:
- Prefer the working tree and index only.
- Do not switch to branch-vs-base comparison in this skill.
- If there are no staged, unstaged, or untracked changes, say so plainly.

## 2. Draft a Summary

Using the diff context, draft a short commit summary that:
- Describes the main change, not every minor detail
- Is written in **imperative mood** ("Add …", "Fix …", "Update …", "Refactor …")
- Avoids Jira keys, issue numbers, or long technical digressions

Notes:
- The current skill behavior may output **multiple summary lines** (e.g., two lines) when multiple top-level intents exist.
- Prefer to keep the *first* summary line concise; do not add a trailing period.

Examples of good summaries:
- `Fix summarize_finding parsing edge cases`
- `Add retries to A2A task store operations`
- `Refactor security-central metrics layout`

Examples to avoid (and correct):
- `Fixed bug in summarize_finding` → `Fix bug in summarize_finding`
- `Adding retries to task store` → `Add retries to task store`
- `This commit changes how we handle X` → `Change handling of X`

## 3. Normalize Changed Files

Build a categorized file list from these sources in priority order:

1. `git diff --name-status`
2. `git diff --cached --name-status`
3. `git status --porcelain` only for untracked `??` files

Normalization rules:
- `D` => `deleted`
- `A` and `??` => `added`
- `M` => `modified`
- `R...` => `modified`
- `C...` => `modified`

If the same file appears multiple times, prefer the most significant status in
this order:

1. `deleted`
2. `added`
3. `modified`

## 4. Output Contract

When this skill is used:
- Output **only** the following, with no extra commentary:
  1) one or more summary lines (imperative mood), then
  2) a blank line, then
  3) a `Changed files:` header line, then
  4) a blank line, then
  5) one changed file per line.

Summary line rules:
- Up to **5 lines**
- Imperative mood ("Add …", "Fix …", "Update …", "Refactor …")
- No trailing period on any line

File list rules:
- Use the same categorization as in the Jira comment, combining:
  - `git diff --name-status`
  - `git diff --cached --name-status`
  - `git status --porcelain` (use only `??` and treat as `added`)
- Use one file per line, prefixed with `- <status>:` where status is `added`, `deleted`, or `modified`.
- Do not prompt for a Jira key.
- Do not create Jira comments.

Output example:

```
Change OCI key rotation design to ADW-only
Update working-tree summary rules

Changed files:

- modified: .agents/skills/git-diff-summary/SKILL.md
- modified: common/doc/oci_api_signing_key_rotation_framework.md
```