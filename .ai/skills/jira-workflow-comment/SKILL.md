---
name: jira-workflow-comment
description: Add a standardized Jira comment for workflow events such as feature-branch creation and merge-request creation.
---

# Jira Workflow Comment

Use this skill when a workflow needs to add a Jira comment as part of an
explicit automation step.

This skill is intentionally **write-capable**.
It exists so workflows can share one clear implementation for Jira comments
instead of duplicating formatting and posting rules.

The workflow still owns:
- when the comment should be posted
- what event occurred
- what validated values are available

This skill owns:
- comment structure
- use of the Jira MCP `jira_add_comment` tool
- consistent fallback wording

## General rules

- Require a validated Jira issue key before posting.
- Use the Jira MCP tool `jira_add_comment`.
- If the Jira update fails, report the failure clearly to the workflow.
- Do not invent branch names, MR URLs, or summary text.
- Do not use this skill for ticket summarization or arbitrary ad hoc comments.

## Supported modes

### 1. `branch_created`

Use when a feature-branch workflow has successfully created and checked out a
branch.

Required inputs:
- `jira_key`
- `branch_name`

Comment body:

```text
Created feature branch

Branch: <branch_name>
```

### 2. `merge_request_created`

Use when an MR workflow has completed the push/MR creation step and has the
final branch and MR context available.

Required inputs:
- `jira_key`
- `branch_name`
- `mr_summary`

Optional inputs:
- `mr_url`

Comment body:

```text
Merge request created

Branch: <branch_name>
Merge Request: <mr_url or URL not available from push output>

<mr_summary>
```

Rules:
- `mr_summary` should already contain the summary lines and `Changed files:`
  section.
- If `mr_url` is unavailable, use exactly:

```text
Merge Request: URL not available from push output
```

## Output contract

When this skill is used:
- post the standardized Jira comment
- return a short success/failure result to the caller
- do not perform unrelated git or Jira reads