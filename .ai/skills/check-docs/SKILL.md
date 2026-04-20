---
name: check-doc-tracking
description: Verify that documentation tracking is complete for the active feature. Compares git diff against the feature context file, checks that all dependency, API, config, schema, deployment, and security changes are logged as deltas. Flags gaps and suggests missing entries.
---

# Skill: Check Documentation Compliance

## Description
Verify that the active feature context file matches the actual code changes. Catches untracked documentation deltas.

## When to Invoke
- Engineer says "check docs", "verify tracking", "am I missing any doc changes"
- Before raising a PR
- After a series of code changes

---

## Instructions

1. Find the active feature context in `.ai/features/`:
   - If specified, use that file
   - Otherwise, look for any file with `Status: ACTIVE`

2. Determine what files changed:
   - Run `git diff --name-only` (or read the list of files you've modified this session)

3. For each changed file, check if doc tracking is needed:
   - Package manifest changed → Dependency Delta needed?
   - New route/controller/endpoint files → API Delta needed?
   - Migration files created → Schema Delta needed?
   - Config files changed → Configuration Delta needed?
   - Dockerfile, Helm, Terraform, CI/CD changed → Deployment Delta needed?
   - Auth/permission files changed → Security Delta needed?

4. Compare against what's already in the feature context

5. Report:
   - ✅ What's properly tracked
   - ⚠️ What's missing and needs to be added
   - Suggested delta entries for anything missing

6. Offer to update the feature context with the missing entries

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
