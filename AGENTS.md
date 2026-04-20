# Agent Guide: Gekko3D Examples

This repo uses a collaboration-first workflow. For ambiguous, risky, or broad changes, align early, keep updates frequent, and leave a short status artifact that a reviewer can scan quickly.

## Start Here

Before editing code:

1. Read `.ai/context.md`.
2. Read `.ai/conventions.md`.
3. Check `.ai/features/` for active work.
4. Read `.ai/known-issues.md` if the task touches rendering, assets, or runtime behavior.
5. Read the relevant example's `main.go` and `go.mod`.
6. If engine behavior matters, inspect the relevant files in `../gekko`.

## Feature Context

For non-trivial feature work and bug fixes, create or update a file under `.ai/features/` before or during implementation.

- Use a ticket ID if one exists.
- If there is no ticket, use a descriptive slug.
- Record scope, design summary, validation, and doc deltas as work progresses.

## Collaboration Workflow

- If the change is unclear, risky, or larger than a small focused edit, run the confidence gate first.
- If confidence is not high, consult an SME dev early and write a status report under `reports/`.
- Keep reviewer asks explicit: decision points, risks, evidence, and what still needs verification.

## Repo-Specific Guardrails

- Each example is its own Go module. Preserve the local `replace github.com/gekko3d/gekko => ../../gekko` pattern.
- Prefer additive, example-local changes over broad cross-example refactors unless requested.
- Do not commit built binaries, `.DS_Store`, or IDE metadata.
- Avoid new asset path assumptions; many examples need to work from more than one working directory.
- Treat manual smoke runs as part of validation for visual or interactive changes.

## Validation

- Automated example-local test:
  - `cd <example> && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- Manual smoke run:
  - `cd <example> && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`
- If a change depends on engine behavior, run the relevant tests from `../gekko`.

## Skill Routing

When applicable, use these repo-preferred skills:

- `run-confidence-gate` for uncertain or broad work
- `consult-sme-triggers` when confidence is not high
- `write-status-report` for async review/handoff artifacts
- `apply-workflow-norms` to keep updates scannable and decision-focused

For onboarding and future refreshes of the AI layer, use `ai-codeassist-onboarding`.
