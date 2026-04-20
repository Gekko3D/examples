# Known Issues & Tech Debt — Gekko3D Examples

> Review this before changing runtime behavior or project structure.

## Active Known Issues

### [ISSUE-001] — Sparse Example-Level Automated Tests

- **Severity:** Medium
- **Affected Area:** `./*/main.go`, `./collision_events/main_test.go`
- **Description:** Most examples are validated primarily by manual runs rather than broad automated test suites.
- **Workaround:** Add a small regression test when the logic can be isolated; otherwise document manual smoke steps clearly.
- **Ticket:** —
- **AI Impact:** Do not overstate confidence from tests alone when a change affects rendering, input, or physics behavior.

### [ISSUE-002] — GPU and Windowing Dependency for Real Validation

- **Severity:** Medium
- **Affected Area:** Any example using `VoxelRtModule`, physics, particles, or UI
- **Description:** Many meaningful failures only appear during an actual desktop run with graphics/windowing support.
- **Workaround:** Pair code changes with a local `go run .` smoke test when possible.
- **Ticket:** —
- **AI Impact:** If the environment cannot open a window, record that limitation explicitly instead of claiming full validation.

### [ISSUE-003] — Asset Path Handling Is Not Fully Uniform

- **Severity:** Medium
- **Affected Area:** Examples with `assets/` directories
- **Description:** Some examples already resolve assets from both the example directory and the repo root, but the pattern is not applied everywhere.
- **Workaround:** Reuse or add a small resolver helper when touching file-backed assets.
- **Ticket:** —
- **AI Impact:** Avoid introducing new hard assumptions about the current working directory.

## Tech Debt Register

| ID | Description | Priority | Effort | Ticket |
|---|---|---|---|---|
| TD-001 | Module paths are inconsistent across example `go.mod` files | Medium | Medium | — |
| TD-002 | Several example folders leave behind local built binaries during development | Low | Low | — |

## Gotchas & Warnings

- The parent `../go.work` influences local development, but many example commands intentionally use `GOWORK=off`.
- `../gekko` is the real engine implementation; an example-only change may still require engine inspection to understand behavior.
- Untracked binaries and IDE files may already exist in the workspace; do not clean them up unless asked.
