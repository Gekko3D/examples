# STATUS_REPORT

## Metadata
- Owner: Codex
- Date: 2026-04-20
- Project/Area: Gekko3D examples onboarding
- Related Links: `INIT-AI.md`, `.ai/`, `AGENTS.md`, `.clinerules`
- Reviewers Requested: Gekko examples maintainer

## Executive Summary
- Goal: Prepare the `examples/` workspace for agent-driven development with a persistent `.ai/` knowledge layer and tool rules.
- Current state: The workspace had no `.ai/`, `AGENTS.md`, `.github/`, or changelog artifacts at this level.
- Progress: Template scaffolding was installed and the main docs/rules were rewritten for the actual Go examples workspace.
- Risk/Blocker: Runtime validation in this repo still depends heavily on local GPU/windowing support and manual smoke tests.
- Ask: Review the `[NEEDS VERIFICATION]` items, especially whether the ticket/branch workflow and module-path inconsistency should be normalized later.

## Background / Context
- Why now: The repo needs agent-readable context and guardrails so future edits follow the examples workspace structure instead of generic service templates.
- Scope: Add onboarding scaffolding, populate `.ai/docs`, create Codex/Cline rules, and record the onboarding result.
- Non-scope: Normalizing example module paths, cleaning untracked binaries, or changing engine/example code behavior.
- Constraints: Keep changes additive, preserve the existing workspace shape, and avoid destructive cleanup in a dirty worktree.

## Current Status
- Completed:
  - Installed onboarding scaffolding from the `ai-codeassist-onboarding` template (evidence: new `.ai/`, `.github/`, `INIT-AI.md`, `CHANGELOG.md`, `.clinerules`)
  - Replaced placeholder project docs with repo-specific content for architecture, dependencies, deployment, project context, conventions, and known issues
  - Added a Codex-focused `AGENTS.md` aligned to the new `.ai/` layer
  - Created `.ai.local/` override stubs for future local customization
- In progress:
  - None
- Next steps:
  - Confirm the verification items below
  - Use the new `.ai/features/` template on the next substantive task

## Proposed Change
- User-visible behavior: Agents now have persistent repo context, conventions, and task workflow guidance.
- API/Interface changes: None to runtime code; this is process/documentation scaffolding only.
- Data changes: Added new repo files under `.ai/`, `.ai.local/`, `.github/`, `reports/`, and root onboarding docs.
- Operational changes: Future agent tasks should start with `.ai/context.md` and create feature files for non-trivial work.

## Design Options

### Option A: Full template copy with repo-specific rewrites
- Approach: Copy the onboarding template, then rewrite the key docs/rules for this workspace.
- Pros:
  - Fastest path to a complete onboarding layer
  - Preserves upgradeable template structure
- Cons:
  - Some bundled skills remain generic because they were copied as-is
- Risk:
  - Reviewers may assume every copied skill is already tailored

### Option B: Hand-build only the minimal docs and rules
- Approach: Create only a few custom files without the full template layout.
- Pros:
  - Less surface area
- Cons:
  - Loses template consistency and future upgrade path
  - Misses expected onboarding artifacts from the named skill
- Risk:
  - Future agent behavior becomes less standardized

### Recommendation
- Chosen option: A
- Rationale: It satisfies the requested onboarding workflow while still grounding the most important artifacts in the real repo structure.
- Open questions:
  - [NEEDS VERIFICATION] Should future feature files require ticket IDs by policy, or are descriptive slugs the default here?
  - [NEEDS VERIFICATION] Do maintainers want to normalize inconsistent example module paths later, or keep them as-is?
  - [NEEDS VERIFICATION] Should `INIT-AI.md` remain in the repo for reruns, or be deleted after first-time onboarding?

## Testing Plan
- Unit tests: Not applicable to the onboarding docs themselves
- Integration tests: Not applicable
- Edge cases:
  - Dirty worktree with unrelated files present
  - No existing `.ai/` structure to preserve
  - Non-APEX repo, so the APEX scanner was removed
- How to validate locally:
  - Open `.ai/context.md` and `.ai/conventions.md`
  - Confirm `AGENTS.md` and `.clinerules` reflect the examples workspace
  - Verify the onboarding files exist at the repo root

## Rollout / Compatibility
- Backward compatibility: Yes; changes are additive and do not alter example runtime code.
- Feature flags: No
- Migration/Backfill: No runtime migration needed
- Rollback plan: Remove the added onboarding files if the team rejects this structure

## Risks & Mitigations
- Risk 1: Generic copied skills may imply stronger project tailoring than currently exists.  Mitigation: The main repo-specific guidance lives in `.ai/context.md`, `.ai/conventions.md`, `.ai/RULES.md`, `AGENTS.md`, and this report.
- Risk 2: Future agents may over-claim validation in a GPU-dependent repo.  Mitigation: Known issues and rules explicitly require manual smoke notes for visual changes.

## Specific Questions for Reviewer
1. Should ticketless descriptive feature files be the default for this repo?
2. Should `INIT-AI.md` stay committed as a rerun entry point?
3. Is module-path normalization a desired follow-up, or should it remain documented tech debt?

# END_STATUS_REPORT
