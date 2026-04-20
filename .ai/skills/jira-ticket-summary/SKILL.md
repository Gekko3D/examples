---
name: jira-ticket-summary
description: Read a Jira ticket and generate a concise summary in chat, using recent comments to infer the latest status.
---


# Jira Ticket Summary

When asked to summarize a Jira ticket, follow this workflow.

## 1. Obtain and Validate the Jira Issue Key

If the user did not provide a Jira issue key, ask for one.

Validation rules:
- The value must be non-empty.
- It should match the basic Jira issue key pattern: `^[A-Z][A-Z0-9]+-[0-9]+$`
- If the value does not match, ask again until the user provides a valid key.

## 2. Fetch the Jira Issue with Comments

Use the Jira MCP tool to read the issue and include comments in the response.

Preferred workflow:
- Use `jira_get_issue`
- Request fields that are helpful for summarization, such as:
  - `summary`
  - `status`
  - `priority`
  - `assignee`
  - `description`
  - `updated`
  - `issuetype`
  - `reporter`
  - `created`
  - `labels`
- Set `comment_limit` high enough to inspect recent comments.
- Prefer the most recent comments when inferring current status.

If the issue cannot be fetched:
- Tell the user clearly that the Jira issue could not be read.
- Include the issue key in the response.
- Do not invent missing details.

## 3. Build the Summary from Fields and Comments

Use both the issue fields and comments, but treat them differently.

### Ticket fields describe the formal state
Use issue fields to summarize:
- what the ticket is about
- current Jira workflow status
- assignee / owner
- priority
- last updated time

### Comments describe the latest working status
Read the comments to determine:
- the newest implementation or investigation status
- blockers or dependencies
- next steps
- whether the work appears done, stalled, waiting on someone, or actively in progress

Guidance:
- Prefer more recent comments over older comments.
- Prefer comments over description text when answering "what is the latest status?"
- If comments conflict with the Jira status field, call that out explicitly.
- If there are no comments, say that latest status is based only on ticket fields.
- If comments are noisy or conversational, distill them into a short factual status update.
- Do not quote long comment text unless a short quote is necessary for clarity.

## 4. Summarize Conservatively

Write a concise summary that distinguishes between:
- the ticket intent
- the formal Jira state
- the latest status from comments

Rules:
- Do not invent facts not present in the ticket or comments.
- If the latest status is uncertain, say so.
- If a blocker, dependency, or ownership handoff is only implied, phrase it cautiously.
- Use plain language suitable for a teammate who wants a quick update.

## 5. Output Contract

When this skill is used, return the summary in chat only.
Do not modify the Jira issue.
Do not add a Jira comment.

Preferred format:

```text
RMS-1234: Improve OCI key rotation workflow

- Type: Story
- Status: In Progress
- Assignee: Jane Doe
- Priority: High
- Updated: 2026-03-10

Overview:
This ticket is about improving the OCI key rotation workflow for RMS admin tooling.

Latest status:
Recent comments indicate the implementation is mostly complete, but final validation and deployment wiring are still in progress.

Risks / blockers:
- Waiting on confirmation of release wiring
- Possible mismatch between Jira status and latest implementation comments
```

Output expectations:
- Start with `<ISSUE-KEY>: <summary>`
- Include a short metadata section
- Include `Overview:` based on the issue fields
- Include `Latest status:` based primarily on recent comments
- Include `Risks / blockers:` only when supported by the issue text or comments
- Keep the overall output concise and readable
