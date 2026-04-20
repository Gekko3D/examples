---
name: resume-work
description: Resume work on a ticket from a previous session. Reads the feature context file as external memory, finds Next Steps from the last session, checks out the existing branch, runs tests, and picks up where the previous session left off.
---

# Skill: Resume Work

## Description
Resume work on a ticket from a previous session. Uses the feature context file as external memory to restore context and pick up where the last session left off.

## When to Invoke
- Engineer says "continue work on", "resume", "pick up where I left off"
- Engineer says "continue PROJ-1234"
- Engineer starts a new session and references an existing ticket

---

## Instructions

### Step 1: Find the Feature Context
1. Look in `.ai/features/` for a file matching the ticket ID
2. If multiple matches, list them and ask the engineer which one
3. If no match found, tell the engineer — they may need to start fresh

### Step 2: Read Full Context
1. Read the feature context file **completely** — not just the top section
2. Pay attention to:
   - **Objective and Scope** — what's being built and what's excluded
   - **Change Log** — what was already done in previous sessions
   - **Next Steps** — what the last session said to do next (this is your starting point)
   - **Documentation Deltas** — what's already tracked
   - **Testing** — what tests exist

### Step 3: Restore Branch
1. Run `git checkout [TICKET_ID]-short-description`
2. If the branch doesn't exist locally: `git fetch origin && git checkout [TICKET_ID]-short-description`
3. If the branch doesn't exist at all, tell the engineer

### Step 4: Verify Current State
1. Run `git status` to see uncommitted changes
2. Run the test command to confirm the codebase is in a good state
3. If tests fail, report to the engineer before proceeding

### Step 5: Resume
1. Start from the **Next Steps** listed in the feature context
2. Do NOT re-do work already logged in the Change Log
3. Continue updating the Change Log and deltas as you work
4. At the end of this session, update Next Steps with what remains

### Step 6: Update Next Steps
Before ending the session, always update the feature context Change Log with:

```
| [today's date] | [what was done this session] | [files touched] |
```

And add a clear Next Steps note:
```
**Next Steps:**
- [ ] Exactly what to do next session (be specific)
- [ ] e.g. "Write tests for POST /api/v1/users/bulk"
- [ ] e.g. "Handle error case when LDAP is unreachable"
```

This is the contract between sessions. Without it, the next resume will lose context.

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
