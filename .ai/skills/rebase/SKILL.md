---
name: rebase-docs
description: Rebase feature context documentation deltas into main project docs after a PR is merged. Reads the feature context file, applies each delta to its target doc (ARCHITECTURE.md, DEPENDENCIES.md, DEPLOYMENT.md), updates CHANGELOG, and marks the feature context as REBASED.
---

# Skill: Rebase Feature Context

## Description
Apply documentation deltas from a completed feature context file into the main project docs. Use after a PR is merged.

## When to Invoke
- Engineer says "rebase", "rebase feature", "apply deltas", "update main docs"
- Engineer says "merge docs for FEAT-XXXX"

---

## Instructions

1. Find the feature context file:
   - Look in `.ai/features/` for the matching filename or ticket number
   - If not specified, list active feature contexts and ask which one

2. Read the feature context file completely

3. For each populated delta section, apply to the target doc:
   - **Architecture Delta** → `.ai/docs/ARCHITECTURE.md`
   - **Dependency Delta** → `.ai/docs/DEPENDENCIES.md`
   - **API Delta** → `.ai/docs/ARCHITECTURE.md` (data flow section)
   - **Schema Delta** → `.ai/docs/ARCHITECTURE.md` (data layer)
   - **Configuration Delta** → `.ai/docs/DEPENDENCIES.md` (environment config)
   - **Deployment Delta** → `.ai/docs/DEPLOYMENT.md`
   - **Security Delta** → `.ai/docs/ARCHITECTURE.md` (security section)

4. Add an entry to `CHANGELOG.md` under `## [Unreleased]`

5. Update the feature context file:
   - Set `Status` to `REBASED`
   - Set `Rebased On` to today's date

6. Summarize all changes made across all files

## Rules
- Apply changes surgically — do NOT delete or overwrite unrelated content
- Maintain formatting consistency with each target document
- If unsure how to apply a delta, add `<!-- REBASE-TODO: [question] -->` and flag it

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
