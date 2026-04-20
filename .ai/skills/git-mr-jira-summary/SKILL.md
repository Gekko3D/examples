---
name: git-mr-jira-summary
description: Generate a concise merge-request summary and changed-files list for the current branch relative to base. Legacy name retained for compatibility; this skill is read-only and must not write to Jira.
---


# Git Merge Request Jira Summary

When asked to prepare a merge-request summary before creating a merge request,
follow this workflow.

Despite the legacy skill name, this skill is **read-only**.
It prepares branch summary content only.
Workflows own any Jira comment creation.

## 1. Collect Branch Change Context (Deterministic)

Use git commands to understand the scope of changes on the current branch
relative to its upstream merge base.

Preferred workflow:

```bash
# Current branch
git rev-parse --abbrev-ref HEAD

# Upstream branch for the current branch
git rev-parse --abbrev-ref --symbolic-full-name @{u}

# Merge base between HEAD and upstream
git merge-base HEAD @{u}

# High-level file summary for branch changes
git diff --stat $(git merge-base HEAD @{u})...HEAD

# Name/status list for branch changes
git diff --name-status $(git merge-base HEAD @{u})...HEAD

# Full patch for branch changes
git -P diff $(git merge-base HEAD @{u})...HEAD
```

Fallback guidance:
- If `@{u}` is not configured, try a reasonable default such as `origin/main` or
  `main` only if it exists locally.
- If no comparison target can be determined safely, ask the user which base
  branch should be used.

From this, identify:
- The primary area(s) of change
- The main intent of the branch
- Any notable cross-cutting work that should be reflected in the merge-request
  summary

## 2. Draft a Summary

Using the branch diff context, draft a short merge-request summary that:
- Describes the main change, not every minor detail
- Is written in **imperative mood** ("Add …", "Fix …", "Update …", "Refactor …")
- Avoids Jira keys, issue numbers, or long technical digressions

Notes:
- The skill may output **multiple summary lines** when multiple top-level
  intents exist.
- Prefer to keep the first summary line concise.
- Do not add a trailing period.

## 3. Normalize Changed Files

Build a list of all files changed on the current branch using:

```bash
git diff --name-status $(git merge-base HEAD @{u})...HEAD
```

Normalize statuses for reporting:
- `D` => `deleted`
- `A` => `added`
- `M` => `modified`
- `R...` => `modified`
- `C...` => `modified`

If a fallback base branch was required because `@{u}` was not configured,
use the same base consistently for the file list and summary.

## 4. Output Contract

When this skill is used, output the prepared content in this format:

1. One or more summary lines
2. A blank line
3. `Changed files:`
4. A blank line
5. One changed file per line

Summary line rules:
- Up to **5 lines**
- Total visible characters across all summary lines should be **<= 200**
- Imperative mood ("Add …", "Fix …", "Update …", "Refactor …")
- No trailing period on any line

File list rules:
- Use the branch comparison described above
- Use one file per line, prefixed with `- <status>:` where status is `added`,
  `deleted`, or `modified`
- For renames, preserve the git path form when available, for example:
  `old/path.sql -> new/path.sql`
- Do not prompt for a Jira key.
- Do not add Jira comments.

Output example:

```
Add merge-request Jira summary skill
Update workflow for branch-based diff reporting

Changed files:

- added: .agents/skills/git-mr-jira-summary/SKILL.md
- modified: .agents/skills/git-diff-summary/SKILL.md
```