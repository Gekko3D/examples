# Project: Gekko3D Examples

## Overview

> **One-liner:** A runnable demo workspace for developing, testing, and showcasing features of the Gekko3D engine.

This workspace exists to make engine capabilities concrete. Instead of describing systems like entity groups, collisions, UI, water effects, or particle rendering abstractly, the repo provides small programs that developers can run locally and inspect directly. That makes it useful both as a manual verification surface for engine changes and as living reference code for new demos.

The primary users are engine contributors and collaborators working in the larger Gekko3D repository. The examples are not a production application; they are developer-facing artifacts that help prove behavior, demonstrate APIs, and expose issues that only appear in a real scene with rendering/input enabled.

## Business Context

- **Product Area:** Engine developer enablement / example content
- **Business Owner:** Gekko3D maintainers
- **Engineering Owner:** Gekko3D contributors working in this mono-worktree
- **Status:** Active development
- **Criticality:** P3 experimental / internal enablement

## Key Users & Stakeholders

| Role | Who | How They Use It |
|---|---|---|
| Engine contributors | Developers working in `../gekko` | Validate engine behavior against runnable demos |
| Example authors | Contributors adding new demos | Follow existing structure to build focused showcases |
| Reviewers | Maintainers reviewing changes | Use examples as evidence that a feature works in a scene |

## Tech Stack Summary

| Layer | Technology |
|---|---|
| Language | Go 1.24/1.25 |
| Engine | `github.com/gekko3d/gekko` from `../gekko` |
| Rendering runtime | WebGPU + GLFW |
| Math | `github.com/go-gl/mathgl` |
| Assets | Local textures, vox models, shaders under example folders |
| Workspace management | `../go.work` plus per-example `go.mod` files |

## Key Metrics

- Number of example modules: 17 at onboarding time
- Automated example-local test coverage is intentionally light
- Manual smoke runs are an expected part of validation for visual changes

## External Documentation

- Root examples overview: `./README.md`
- Parent workspace guidance: `../.github/copilot-instructions.md`
- AI onboarding context: `.ai/context.md`
