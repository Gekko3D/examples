---
name: init-ai-codeassist
description: Initialize AI code assist structure from scratch. Create .ai/ directory, all templates, all skills, and auto-onboard the project. Use when a project has no .ai/ setup at all.
---

# Skill: Initialize AI Code Assist Structure

## Description
Create the complete `.ai/` directory structure and all template files from scratch in a project that has no AI code assist setup yet. After scaffolding, automatically triggers the onboard skill to populate everything from the codebase.

## When to Invoke
- Engineer says "init AI code assist", "initialize AI", "set up AI code assist from scratch"
- Engineer says "create .ai/ structure"
- The project has no `.ai/` directory at all

---

## Instructions

You are setting up the AI code assist framework from scratch in this project. Create every directory and file listed below, then onboard the project.

### Phase 1: Create Directory Structure

Create these directories:
```
.ai/
.ai/VERSION
.ai/docs/
.ai/features/
.ai/adr/
.ai/skills/
.ai/skills/onboard/
.ai/skills/start-feature/
.ai/skills/complete-task/
.ai/skills/resume/
.ai/skills/pr-review/
.ai/skills/rebase/
.ai/skills/check-docs/
.ai/skills/init/
.ai/skills/upgrade/
.ai/skills/apex-security-scan/
.ai/tools/
.github/
```

### Phase 2: Create Template Files

Create each file below with the EXACT content specified. Do not skip any file.

---

#### File: `.ai/RULES.md`

Create with the full behavioral rules content. This is the source of truth for agent behavior. Include all of the following sections:

- **STEP 0: Orient** — Read context.md, conventions.md, check features/, known-issues.md. If given a ticket ID, fetch via MCP first.
- **STEP 1: Ensure Feature Context Exists** — Create from template, populate from ticket tracker MCP if available.
- **Git Workflow** — Branch naming (`[TICKET_ID]-short-description`), commit format (`[[TICKET_ID]] description`), never commit to main, ask before push.
- **Testing** — Placeholders: `[TO BE FILLED DURING ONBOARDING]` for framework, test directory, test command.
- **STEP 2: Write Code** — Pattern matching, hard rules (no secrets, no PII in logs, no raw SQL, pin versions), quality rules (timeouts on external calls, tests required).
- **STEP 3: Track Documentation Changes** — Table mapping change types to delta sections. Never edit .ai/docs/ directly. Log in feature context.
- **STEP 4: Self-Check / Definition of Done** — Checklist: branch pushed, code follows patterns, no secrets, tests written, tests passing, feature context current, deltas logged, CHANGELOG updated, ADR if needed, ticket transitioned if MCP available.
- **STEP 5: Rebase** — Apply deltas surgically, don't delete unrelated content, mark REBASED.
- **Common Tasks** — NEW FEATURE (8 steps with plan approval gate), BUG FIX (with regression test), TASK COMPLETE (tests → changelog → push → ticket transition), RESUME WORK (read feature file → Next Steps → checkout branch → run tests → resume).
- **Task Complexity Guide** — Simple / Medium / Complex.
- **Commit Messages** — Format: `[[TICKET_ID]] Brief description`.
- **CHANGELOG Format** — Entries under `## [Unreleased]` with ticket IDs.
- **Key References** — Table of all .ai/ files and when to read them.

Use the content from the canonical RULES.md template. Include the `[TO BE FILLED DURING ONBOARDING]` placeholders — the onboard skill fills these.

---

#### File: `.ai/VERSION`

Create a plain text version file for the framework, starting at:
```
1.0.0
```

---

#### File: `.ai/context.md`

Create with template structure:
- Project Identity (name, purpose, tech stack, repo type, status — all as `[TO BE FILLED]` placeholders)
- Architecture Summary placeholder
- Key Entry Points table (empty)
- Critical Rules quick reference
- Current State section
- Deep References table pointing to all .ai/ files

---

#### File: `.ai/conventions.md`

Create with template structure:
- Language & Tooling (placeholders)
- Naming Conventions table (placeholders)
- Project Patterns sections: API Layer, Service Layer, Data Access, Error Handling, Logging, Testing, Feature Flags
- "Things AI agents should AVOID" section

---

#### File: `.ai/known-issues.md`

Create with template structure:
- Active Known Issues section with example format (severity, affected area, description, workaround, ticket, AI impact)
- Tech Debt Register table
- Gotchas & Warnings list

---

#### File: `.ai/docs/PROJECT.md`

Create with template: Overview, Business Context, Key Users & Stakeholders table, Tech Stack Summary table, Key Metrics, External Documentation links.

#### File: `.ai/docs/ARCHITECTURE.md`

Create with template: System Overview, Mermaid Architecture Diagram placeholder, Core Components, Data Flow, Security Model, Scalability & Resilience, Key Design Decisions table.

#### File: `.ai/docs/DEPENDENCIES.md`

Create with template: Mermaid Dependency Map placeholder, Upstream Dependencies table, Downstream Dependencies table, Infrastructure Dependencies table, Critical Libraries table, Shared Libraries table, API Contracts table, Environment Configuration table.

#### File: `.ai/docs/DEPLOYMENT.md`

Create with template: Overview, Mermaid Environment Promotion flow placeholder, Pipeline Stages, Kubernetes Resources table, Secrets & Configuration table, Deployment Commands, Rollback procedure, Health Checks table, Post-Deploy Checklist.

---

#### File: `.ai/features/_FEATURE_TEMPLATE.md`

Create the feature context template with:
- Metadata table (Ticket URL, Type, Status, Author, Created, Branch, PR)
- Objective, Scope (In/Out), Requirements (with MCP auto-populate hint)
- Design Summary
- Change Log table with **Next Steps** section (critical for resume workflow)
- Documentation Deltas: Architecture, Dependency, API, Schema, Configuration, Deployment, Security
- ADRs Created, Known Issues Resolved, New Tech Debt
- Testing checklist (unit, integration, regression, edge cases, all passing)
- Rebase Status

---

#### File: `.ai/adr/ADR-000-template.md`

Create with: Status, Date, Context, Decision, Alternatives Considered, Consequences (positive/negative), Related.

---

#### File: `.github/pull_request_template.md`

Create with: Description, Ticket link, Feature Context link, Type of Change checkboxes, Documentation Compliance section (deltas captured, ADR created, rebase plan), Testing checkboxes, Deployment Notes.

---

#### File: `CHANGELOG.md`

Create with Keep a Changelog format:
```
# Changelog

All notable changes documented here. Format: [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added
### Changed
### Fixed

---
```

---

#### Skill Files

Create each skill file in its directory. The content for each skill should match the canonical versions:

- `.ai/skills/init/SKILL.md` — This file (copy yourself into it)
- `.ai/skills/onboard/SKILL.md` — Full onboarding skill (codebase analysis, doc population, rules file generation)
- `.ai/skills/start-feature/SKILL.md` — Start feature with MCP ticket fetch, branch creation, plan writing
- `.ai/skills/complete-task/SKILL.md` — Tests, changelog, commit, push (with approval), ticket transition
- `.ai/skills/resume/SKILL.md` — Read feature file, find Next Steps, checkout branch, resume
- `.ai/skills/pr-review/SKILL.md` — Auto-detect or MCP-based PR review with structured output
- `.ai/skills/rebase/SKILL.md` — Apply feature context deltas to main docs
- `.ai/skills/check-docs/SKILL.md` — Compare git diff against feature context, flag gaps
- `.ai/skills/upgrade/SKILL.md` — Upgrade framework files from a newer template while preserving project content
- `.ai/skills/apex-security-scan/SKILL.md` — Static security scan of APEX split SQL export (SQLi, XSS, URL tampering, config)

For each skill, write the complete SKILL.md content with Description, When to Invoke, and full Instructions sections.

---

#### File: `.ai/tools/apex_scan.py`

Copy from the template source `.ai/tools/apex_scan.py`. This is the static analysis engine used by the `apex-security-scan` skill. It requires no database — reads APEX split SQL export files directly.

**If this is an APEX project** (detected by presence of any `application/create_application.sql` in the repo): copy this file.
**If this is not an APEX project**: skip this file.

---

#### File: `ONBOARD.md`

Create with instructions on how to trigger onboarding:
```
Open your AI tool and say:
"Onboard this project to AI code assist. Read .ai/skills/onboard/SKILL.md and follow every step."
```

---

### Phase 3: Immediately Trigger Onboarding

After all files are created, automatically execute the onboarding skill:
1. Read `.ai/skills/onboard/SKILL.md`
2. Follow every step — analyze codebase, populate docs, generate rules file
3. Report what was done

The engineer should see the scaffolding AND the populated docs in one session.

---

### Phase 4: Report

```
## AI Code Assist Initialized

### Structure Created:
- [x] .ai/ directory with all subdirectories
- [x] RULES.md (behavioral rules — source of truth)
- [x] Template files (context, conventions, known-issues, docs, features, adr)
- [x] 10 core skills (init, onboard, start-feature, complete-task, resume, pr-review, rebase, check-docs, upgrade, apex-security-scan)
- [x] `.ai/tools/apex_scan.py` (APEX static security scanner)
- [x] PR template, CHANGELOG, ONBOARD.md

### Onboarding Results:
[output from the onboard skill]

### What You Need to Do:
1. Review the populated .ai/docs/ files
2. Fill in known-issues.md with your team
3. Verify sections marked [NEEDS VERIFICATION]
4. Commit the .ai/ structure to the repo
```

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
