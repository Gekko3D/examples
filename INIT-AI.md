# Initialize AI Code Assist

## One-Line Setup

Drop this file into any project root and tell your AI tool:

### For Cline:
```
Read INIT-AI.md and do what it says.
```

### For Codex:
```
Read INIT-AI.md and do what it says.
```

That's it. No manual copying, no shell scripts.

---

## What This Does

The AI will:
1. Create the entire `.ai/` directory structure (docs, features, adr, skills)
2. Create all template files with proper content
3. Create all 10 skills (init, onboard, start-feature, complete-task, resume, pr-review, rebase, check-docs, upgrade, apex-security-scan)
4. Analyze your codebase and populate the docs from what it finds
5. Generate the rules file for your tool (`.clinerules` for Cline, `AGENTS.md` for Codex, or both)
6. Report what was done and what needs human review

---

## Instructions for AI Agent

You are initializing AI code assist for this project. Follow these steps:

### Step 1: Check project state

Determine which situation applies:

**A) No .ai/ directory exists:**
→ This is a fresh project. Proceed to Step 2 (full initialization).

**B) .ai/ directory exists but files contain only template placeholders (unfilled [BRACKETS]):**
→ Structure exists but was never onboarded. Skip scaffolding, proceed to Step 3 (onboard).

**C) .ai/ directory exists and files have real project content:**
→ This project is already onboarded. Do NOT look for template sources in other directories or repos.

Ask the engineer:
"This project is already onboarded. What would you like to do?
1. Re-run onboarding to refresh docs from the codebase
2. Upgrade to a newer template version — tell me the path to the updated template
3. Nothing — I'll stop here"

**IMPORTANT:** Never scan the filesystem for other repos or directories to use as a template source. Only use a template path the engineer explicitly provides.

### Step 2: Create the structure

Read `.ai/skills/init/SKILL.md` if it exists and follow its instructions.

If it doesn't exist yet (first time), create the entire `.ai/` structure by following these steps:

1. Create all directories: `.ai/`, `.ai/docs/`, `.ai/features/`, `.ai/adr/`, `.ai/skills/` (with subdirectories for each skill), `.ai/tools/`, `.github/`
2. Create `.ai.local/` directory with: `RULES.local.md` (empty template), `conventions.local.md` (empty template), `skills/` (empty directory)
3. Create every template file with proper content (RULES.md, context.md, conventions.md, known-issues.md, all docs, feature template, ADR template, all skill files, PR template, CHANGELOG)
4. Write complete, functional skill files — not stubs
5. **Detect if this is an APEX project:**
   Run: `find . -maxdepth 3 -name "create_application.sql" 2>/dev/null | head -1`
   - If a result is found → this is an APEX project:
     - Copy `.ai/tools/apex_scan.py` from the template source into `.ai/tools/apex_scan.py`
     - Confirm: "APEX export detected — apex-security-scan skill and scanner tool installed."
   - If no result → skip `.ai/tools/` (non-APEX project, tool not needed)

### Step 3: Onboard the project

After scaffolding, immediately:
1. Analyze the codebase (package manifests, source structure, configs, CI/CD, tests, existing docs)
2. Populate all `.ai/docs/` files from what you find
3. Synthesize `context.md` and extract `conventions.md` from actual code
4. Detect test framework and MCP servers
5. Fill RULES.md placeholders with actual values
6. Generate the tool-specific rules file (`.clinerules` and/or `AGENTS.md`)

### Step 4: Clean up

- Delete this `INIT-AI.md` file (it's no longer needed after initialization)
- Or ask the engineer if they want to keep it

### Step 5: Report

Summarize what was created, what was populated, and what needs human review.

---

## After Initialization

Your project now has a complete AI code assist framework. Use these prompts:

| What You Want | What You Say |
|---|---|
| Start a ticket | `Work on PROJ-1234` |
| Check doc tracking | `Check if my doc tracking is complete` |
| Review a PR | `Review this PR` or `Review PR for PROJ-1234` |
| Finish and push | `Complete task` |
| Resume next day | `Continue work on PROJ-1234` |
| Merge docs after PR | `Rebase PROJ-1234 into main docs` |
| Scan APEX app for security issues | `scan my APEX app` or `security check` |
