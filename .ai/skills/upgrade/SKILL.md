---
name: upgrade-ai-codeassist
description: Upgrade the .ai/ framework to the latest version. Updates skills, rules, and templates from a newer template source without overwriting project-specific docs, conventions, active features, locked skills, or local overrides.
---

# Skill: Upgrade AI Code Assist

## Description
Update the `.ai/` framework to the latest version while preserving project-specific content, local overrides, and locked skills. Replaces framework files, adds new skills, and regenerates tool rules files (`.clinerules` / `AGENTS.md`) without deleting project-specific skills by default.

## When to Invoke
- Engineer says "upgrade AI code assist", "update AI framework", "sync to latest template"
- Engineer says "we have new skills, update my project"
- A new version of the AI code assist template has been released

---

## Instructions

### Step 1: Check Current Version

1. Read `.ai/VERSION` for the current framework version.
2. If missing, treat current version as `0.0.0`.
3. Report the current version.

### Step 2: Locate the Updated Template Source

Ask the engineer for one source:
1. Local template directory path
2. Updated `INIT-AI.md` source
3. Cloned template repo path

Read the source `.ai/VERSION` and compare with current.

If versions match:
- Report "Already on latest version"
- Ask if they want a force re-sync.

### Step 3: Build a Change Plan

Categorize files before changing anything.

**REPLACE (framework files):**
- `.ai/RULES.md`
- `.ai/skills/*/SKILL.md` for template-managed skills
- `.ai/features/_FEATURE_TEMPLATE.md`
- `.ai/adr/ADR-000-template.md`
- `.github/pull_request_template.md`
- `INIT-AI.md`
- `BOOTSTRAP-GUIDE.md`
- `.ai/tools/apex_scan.py` (always replace — this is a framework tool, not project-specific)
- `.ai/VERSION`

**PRESERVE (project files):**
- `.ai/context.md`
- `.ai/conventions.md`
- `.ai/conventions.local.md` (if exists)
- `.ai/known-issues.md`
- `.ai/RULES.local.md` (if exists)
- `.ai/docs/*.md`
- `.ai/skills/` (project-specific) — custom skills not in the template, do not delete
- `.ai/features/FEAT-*.md`
- `.ai/features/FIX-*.md`
- `.ai/features/BUG-*.md`
- `.ai/features/SECURITY-*.md`
- `.ai/adr/ADR-001+`
- `CHANGELOG.md`

**REGENERATE (derived files):**
- `.clinerules`
- `AGENTS.md`

Present the plan and request approval before writing files.

### Step 4: Preserve Local-Only Skills (Mandatory Safety Rule)

1. Compare skills in template source vs project:
   - Template skills = source `.ai/skills/*`
   - Project skills = current `.ai/skills/*`
2. Any skill present in project but missing from template is **local-only**.
3. **Do NOT delete local-only skills by default.**
4. If engineer explicitly asks to remove them, list each one and require explicit confirmation before deletion.
5. If a skill has `SKILL.lock`, do not replace its `SKILL.md` even when that skill exists in the template.
6. If a skill has `SKILL.local.md`, always preserve that file.

### Step 5: Update Skills (Selective — Preserve Project-Specific Skills)

Compare skill directories between template and project.

**Categorize every skill:**

| Category | Condition | Action |
|---|---|---|
| Locked skill | `SKILL.lock` exists in project skill directory | SKIP — do not touch |
| Template skill (exists in both) | Template has it AND project has it AND no lock | REPLACE with latest version |
| New template skill | Template has it, project does NOT | ADD to project |
| Project-specific skill | Project has it, template does NOT | PRESERVE — do not touch |
| Local extension file | `SKILL.local.md` exists | ALWAYS PRESERVE |

Steps:
1. List all skill directories in the template: `ls [template]/.ai/skills/`
2. List all skill directories in the project: `ls .ai/skills/`
3. For each skill in the template:
   - If `SKILL.lock` exists in project skill directory, skip this skill entirely
   - Else if it exists in the project, replace `SKILL.md` with the template version
   - If it does not exist in the project, create the directory and `SKILL.md`
4. For each skill in the project that is NOT in the template:
   - **Do NOT delete or modify it** — this is a project-specific skill
   - Log it as preserved
5. For every project skill directory:
   - If `SKILL.local.md` exists, never modify or delete it

Report to the engineer:
```
Skills Update Plan:
- REPLACE (template updates): init, onboard, start-feature, complete-task, resume, pr-review, rebase, check-docs, upgrade
- ADD (new in template): [list any new skills]
- LOCKED (SKILL.lock present — untouched): [list any locked skills]
- PRESERVE (project-specific, untouched): apex-code-review, plsql-code-review, git-workflow, jira-integration
- LOCAL EXTENSIONS (SKILL.local.md — untouched): [list any local extensions]
```

Ask: **"Proceed? Any project-specific skills you want me to update or remove?"**

### Step 6: Update Other Framework Files

Use the REPLACE list from Step 3 for non-skill framework files:
- `.ai/RULES.md`
- `.ai/features/_FEATURE_TEMPLATE.md`
- `.ai/adr/ADR-000-template.md`
- `.github/pull_request_template.md`
- `INIT-AI.md`
- `BOOTSTRAP-GUIDE.md`

After file replacement, regenerate tool rules files with precedence:
1. Start from updated `.ai/RULES.md`
2. If `.ai/RULES.local.md` exists, apply local overrides (local wins on conflicts)
3. Preserve project-specific customizations not already represented in `.ai/RULES.local.md`
4. Regenerate `.clinerules` and/or `AGENTS.md`

### Step 7: Bump Version

Write the new version into `.ai/VERSION`.

### Step 8: Report

Report:
- Version change (`old -> new`)
- Updated files
- Added template skills
- Locked skills kept unchanged
- Preserved local-only skills
- Preserved `SKILL.local.md` extensions
- Any skills pending explicit removal confirmation

### Step 8.5: Refresh Skills Index

Regenerate `.ai/SKILLS.md` by scanning both directories:
1. Read every `.ai/skills/*/SKILL.md` — extract `name` and `description` from YAML frontmatter
2. Read every `.ai.local/skills/*/SKILL.md` — extract `name` and `description` from YAML frontmatter
3. Map triggers from Skill Routing table
4. Write updated `.ai/SKILLS.md`

This ensures the index reflects any new skills added or removed during the upgrade.

---

## Reference: Local Skill Control

### `SKILL.lock` (full ownership)
Use an empty `SKILL.lock` file in a skill directory to prevent replacement during upgrades.

### `SKILL.local.md` (extension layer)
Use `SKILL.local.md` beside `SKILL.md` to append project-specific steps after the base skill workflow.

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
