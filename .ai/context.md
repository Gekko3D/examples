# AI Agent Context — Gekko3D Examples

> Read this file first before performing any task in this workspace.

## Project Identity

- **Name:** Gekko3D examples
- **Purpose:** A collection of runnable example programs that exercise engine features in the sibling `../gekko` module.
- **Tech Stack:** Go 1.24/1.25, Gekko ECS/game framework, WebGPU/GLFW rendering, `mathgl`, local assets
- **Repo Type:** Multi-module examples workspace inside a larger Go worktree
- **Status:** Active development / experimental demos

## Architecture Summary

This workspace is not a standalone product service. Each example folder is its own small Go module with a `main.go` entry point and a local `go.mod` that replaces `github.com/gekko3d/gekko` with `../../gekko`. Most examples build a `NewApp()`, register engine modules such as `TimeModule`, `AssetServerModule`, `InputModule`, and `VoxelRtModule`, then install one demo-specific module that wires systems/resources for that scene. The real engine behavior lives in the sibling `../gekko` module, so example changes often depend on engine APIs, asset loading behavior, and desktop GPU/windowing support. The parent `../go.work` ties these modules together for local development.

## Key Entry Points

| What | Location |
|---|---|
| Example entry points | `./<example>/main.go` |
| Example module definitions | `./<example>/go.mod` |
| Workspace configuration | `../go.work` |
| Core engine dependency | `../gekko/` |
| Existing examples index | `./README.md` |
| Example-local tests | `./collision_events/main_test.go` |
| Engine tests often relevant to example work | `../gekko/*_test.go` |
| Runtime assets | `./<example>/assets/` |

## Critical Rules (Quick Reference)

1. Treat each example directory as a standalone Go module; preserve its local `replace github.com/gekko3d/gekko => ../../gekko`.
2. Match existing example structure: `main.go`, `DemoModule`, resource structs, staged systems, and printed runtime instructions.
3. If an example loads assets from disk, make it work both from the example directory and from the repo root when practical.
4. Do not commit generated binaries, `.DS_Store`, or IDE metadata changes unless explicitly requested.
5. Validate with targeted commands in the touched example module; avoid assuming a single root-level `go test ./...` is the right command for this workspace.

## Current State

- **Active Features:** Check `.ai/features/` for in-flight work before starting.
- **Known Issues:** See `.ai/known-issues.md` for GPU/runtime and asset-path caveats.
- **Recent Decisions:** See `.ai/adr/` for any recorded architecture decisions.

## Deep References

- Architecture and workspace shape: `.ai/docs/ARCHITECTURE.md`
- Dependencies and module relationships: `.ai/docs/DEPENDENCIES.md`
- Local execution and validation workflow: `.ai/docs/DEPLOYMENT.md`
- Project/stakeholder context: `.ai/docs/PROJECT.md`
- Coding patterns and validation expectations: `.ai/conventions.md`
