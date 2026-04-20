# AI Code Assist — Behavioral Rules for Gekko3D Examples

This file is the project-level source of truth for agent behavior in this workspace.

## Override Precedence

Project-local overrides live under `.ai.local/`:

| Override File | What It Overrides | Upgrade Behavior |
|---|---|---|
| `.ai.local/RULES.local.md` | Sections in this `RULES.md` | Never touched by upgrade |
| `.ai.local/conventions.local.md` | Sections in `.ai/conventions.md` | Never touched by upgrade |
| `.ai.local/skills/[name]/SKILL.local.md` | Extends `.ai/skills/[name]/SKILL.md` | Never touched by upgrade |
| `.ai.local/skills/[name]/SKILL.lock` | Prevents upgrade from replacing that skill | Never touched by upgrade |

When generating tool-specific rules, merge `.ai/RULES.md` with `.ai.local/RULES.local.md` if it exists.

## STEP 0: Orient (Do This Every Time)

Before writing code:

1. Read `.ai/context.md`.
2. Read `.ai/conventions.md`.
3. Scan `.ai/features/` for active work that overlaps the request.
4. Read `.ai/known-issues.md` if the task touches runtime assets, rendering, or GPU-sensitive logic.
5. Read the affected example's `main.go`, `go.mod`, and nearby source files.
6. If the task depends on engine APIs or behavior, inspect the relevant files in `../gekko`.

## STEP 1: Create or Update Feature Context

Feature context is required for non-trivial feature work and bug fixes.

- If the user gives a ticket ID, name the file `.ai/features/<TICKET>-short-description.md`.
- If there is no ticket, use a descriptive slug such as `.ai/features/water-world-splash-tuning.md`.
- Fill in scope, design notes, test plan, and documentation deltas before or during implementation.
- Tiny formatting fixes, comment-only edits, and test-only edits can skip a feature file.

## Git Workflow

- Prefer a new branch for substantive work. Use a ticket-based branch if one exists; otherwise use a descriptive slug.
- Do not revert unrelated user changes in this workspace.
- Ask before pushing.

## Testing

- Example-local automated validation:
  - `cd <example> && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- Example smoke run:
  - `cd <example> && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`
- Engine validation when example work depends on engine APIs:
  - Run targeted tests from `../gekko`.
- If a change cannot be validated automatically because it needs a windowed GPU run, record the manual validation steps in the feature file and final report.

## STEP 2: Write Code (Follow These Standards)

### Pattern Matching

- Match the surrounding example's structure and naming.
- Reuse existing engine modules and helper patterns before inventing new abstractions.

### Hard Rules

- No secrets, tokens, or machine-specific values in code or docs.
- Preserve local module `replace` directives unless the task is explicitly about dependency layout.
- Avoid hidden cwd assumptions for examples that load files from `assets/`.
- Do not add generated binaries or editor artifacts to version control.

### Quality Rules

- New demo logic should have a small automated regression test when feasible.
- User-facing controls or runtime instructions should stay explicit.
- Keep changes narrow to the touched example unless cross-example cleanup is requested.

## STEP 3: Track Documentation Changes

After each substantive code change, update the active feature file if any of these changed:

- dependencies
- asset requirements
- runtime controls or configuration
- architecture or engine integration assumptions
- manual validation steps

Do not directly rewrite `.ai/docs/` during feature execution unless the user explicitly asks for a docs refresh or onboarding update.

## STEP 4: Self-Check Before Completion

Before reporting completion:

- Code follows the local example pattern.
- Tests or manual validation steps are recorded.
- No generated binaries or editor artifacts were introduced intentionally.
- Any doc-impacting changes are logged in the feature file.
- If behavior changed visually, manual smoke steps are included.

## STEP 5: Rebase Docs (Only When Requested)

When asked to refresh or rebase docs:

1. Read the relevant feature file deltas.
2. Update only the matching sections in `.ai/docs/`.
3. Keep the docs specific to this examples workspace.
4. Mark the feature file as rebased and summarize what changed.
