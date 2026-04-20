---
name: start-feature
description: Start work on a new feature or bug fix. Fetches ticket from Jira or project tracker via MCP, creates feature context file, checks out a git branch, writes an implementation plan, and awaits approval before coding.
---

# Skill: Start Feature

## Description
Start work on a new feature or bug fix. Fetches ticket details from the project tracker (if MCP available), creates a feature context file, checks out a branch, writes a plan, and awaits approval before coding.

## When to Invoke
- Engineer says "start feature", "start working on", "pick up ticket", "work on PROJ-1234"
- Engineer mentions a ticket ID and asks to begin work
- Engineer says "new feature" or "fix bug" with a ticket reference

---

## Instructions

### Step 1: Orient
1. Read `.ai/context.md` to understand the project
2. Read `.ai/conventions.md` to understand coding standards
3. Check `.ai/features/` for any overlapping active work
4. Read `.ai/known-issues.md` for related gotchas

### Step 2: Fetch Ticket (if tracker MCP is available)
1. Attempt to fetch the ticket via the project tracker MCP server (Jira, Linear, GitHub Issues, etc.)
2. If MCP is available and connected:
   - Extract: summary, description, acceptance criteria, type, labels, linked issues
   - Use this data to populate the feature context
3. If MCP is NOT available or fails:
   - Ask the engineer to paste the ticket summary and acceptance criteria
   - Proceed with what the engineer provides

### Step 3: Create Feature Context
1. Create `.ai/features/[TICKET_ID]-[short-description].md` from `.ai/features/_FEATURE_TEMPLATE.md`
2. Fill in:
   - **Metadata:** Ticket URL, status ACTIVE, date, branch name
   - **Objective:** From ticket summary
   - **Scope:** In scope = acceptance criteria. Out of scope = ask engineer if unclear.
   - **Requirements:** Populated from acceptance criteria (mark as `<!-- From ticket tracker via MCP -->` if auto-populated)

### Step 4: Create Branch
1. Run `git checkout -b [TICKET_ID]-short-description`
2. Confirm branch creation in the feature context metadata

### Step 5: Write Plan
1. Based on the ticket requirements and your codebase analysis, write a plan:
   - List the files likely to change
   - List subtasks in dependency order
   - Note any risks or unknowns
2. Add the plan to the feature context under Design Summary

### Step 6: Report and Await Approval
Report back to the engineer with:
- What you understand the task to be
- The plan (files to change, subtask order)
- Any questions or risks identified
- **Ask: "Shall I proceed with this plan?"**

**Do NOT write any implementation code until the engineer approves the plan.**

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
