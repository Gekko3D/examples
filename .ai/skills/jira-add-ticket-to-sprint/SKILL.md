---
name: jira-add-ticket-to-sprint
description: Add a Jira issue to an open sprint by discovering scrum boards and open sprints for a project, then prompting for selection when more than one sprint is available.
---

# Jira Add Ticket To Sprint

Use this skill when a workflow or user wants to put a Jira issue into a sprint
without already knowing the sprint id.

This skill is intentionally **write-capable**.
It exists so sprint assignment logic lives in one narrow reusable skill instead
of being reimplemented inside multiple workflows.

This skill owns:
- validating the Jira issue key
- discovering scrum boards for the project
- collecting open sprint candidates
- prompting for a choice when there is more than one candidate
- adding the issue to the chosen sprint

This skill does not own:
- creating or renaming sprints
- posting Jira comments
- changing issue status
- deciding broader workflow sequencing

## 1. Validate inputs

Required input:
- `jira_key`

Optional input:
- `project_key`

Validation rules:
- `jira_key` must match the basic Jira key pattern:
  `^[A-Z][A-Z0-9]+-[0-9]+$`
- If `project_key` is not supplied, derive it from the prefix before `-` in
  `jira_key`
- If the issue key is missing or invalid, ask the user for a valid Jira key
  before continuing

## 2. Discover candidate sprints

Use Jira MCP tools in this order:

1. `jira_get_agile_boards`
   - filter by `project_key`
   - prefer `board_type = scrum`
2. For each returned scrum board, call `jira_get_sprints_from_board`
   - once with `state = active`
   - once with `state = future`

Treat both `active` and `future` sprints as open sprint candidates.

Rules:
- Ignore closed sprints
- Merge candidates across boards
- Deduplicate by sprint id
- Keep board context when presenting a sprint so duplicate sprint names are still
  easy to distinguish

If no scrum boards are found:
- tell the user no scrum boards were found for the project
- do not invent a sprint id

If scrum boards exist but no active or future sprints are found:
- tell the user there are no open sprints for the project
- do not modify Jira

## 3. Choose the sprint

If exactly one open sprint candidate exists:
- use it automatically
- tell the user which sprint will be used before adding the issue

If more than one open sprint candidate exists:
- present a concise numbered list
- include at least:
  - sprint name
  - sprint state
  - board name
  - sprint id
  - dates when available
- ask the user to choose one sprint
- do not add the issue until the user replies with a selection

Presentation rules:
- prefer active sprints before future sprints
- within the same state, prefer the nearest end date when available
- if dates are missing, keep the original Jira order

## 4. Add the issue to the sprint

Once a sprint is selected, use:
- `jira_add_issues_to_sprint`

Call it with:
- `sprint_id = <selected sprint id>`
- `issue_keys = <jira_key>`

Rules:
- add only the validated issue key unless the caller explicitly supplied more
  than one validated issue
- if Jira rejects the update, report the failure clearly
- do not perform unrelated Jira writes

## 5. Output contract

On success, return a short confirmation in chat:

```text
Added RMS-1234 to sprint Sprint 24 (active, board: RMS Delivery, id: 10432).
```

On selection-needed flow, return a short prompt like:

```text
I found multiple open sprints for RMS. Choose one:
1. Sprint 24 — active — RMS Delivery — id 10432
2. Sprint 25 — future — RMS Delivery — id 10487
```

On failure, return a short explanation that says whether the problem was:
- invalid or missing issue key
- no scrum boards found
- no open sprints found
- Jira update failure
