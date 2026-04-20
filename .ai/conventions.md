# Coding Conventions — Gekko3D Examples

## Language & Tooling

- **Language:** Go 1.24/1.25
- **Framework:** Gekko ECS/game framework from `../gekko`
- **Style Guide:** Standard Go formatting and naming
- **Formatter:** `gofmt`
- **Primary validation:** targeted `go test` and `go run` inside the affected example module

## Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Types | PascalCase | `DemoModule`, `DemoState` |
| Functions / methods | camelCase | `setupScene`, `resolveDemoAsset` |
| Constants | Go-style CamelCase or lower camel for unexported values | `Startup`, `demoVoxelResolution` |
| Packages in examples | `main` | `package main` |
| Example folders | snake_case / descriptive demo names | `water_world`, `entity_groups` |
| Asset helpers | Verb-first utility names | `resolveAsset`, `resolveDemoAsset` |

## Project Patterns

### Example Structure

- One example per folder.
- Each example normally has `main.go` and its own `go.mod`.
- Each example should stay runnable on its own and document controls via printed output or README notes.

### App Assembly

- Build applications with `NewApp()`, `UseStates(...)`, `UseModules(...)`, then `app.Run()`.
- Prefer the existing module ordering pattern: timing/assets/input/rendering/physics/lifecycle/custom demo module.
- Keep window title and dimensions explicit inside the example's render module configuration.

### Demo Modules and Systems

- Demo-specific wiring usually lives in a `DemoModule` with an `Install(*App, *Commands)` method.
- Register resources early with `cmd.AddResources(...)`.
- Attach systems to explicit stages/states such as `Prelude`, `Update`, `PostUpdate`, and `OnEnter(...)`.
- Follow the existing ECS style for components, entity creation, and system signatures instead of inventing a parallel abstraction.

### Imports

- Many examples use a dot import for `github.com/gekko3d/gekko`. Match the file you are editing rather than forcing a style change.
- Keep external imports minimal. New dependencies should be rare and should be documented in the feature file.

### Assets and Paths

- If an example reads files from `assets/`, prefer a helper that tolerates both example-local and repo-root working directories.
- Keep asset names deterministic and colocated with the example that owns them.

### Testing

- For touched example modules, run `go test ./...` from that example directory when possible.
- If runtime behavior is visual or interactive, pair code changes with a lightweight regression test where feasible and a manual smoke test note.
- If a change depends on engine behavior in `../gekko`, run the relevant engine tests there instead of assuming the example is enough coverage.

## Things AI Agents Should Avoid

- Do not change sibling engine code in `../gekko` unless the task clearly requires it.
- Do not commit built binaries such as `./<example>/<example>`.
- Do not hardcode absolute filesystem paths.
- Do not break local `replace` directives or silently switch examples to remote engine dependencies.
- Do not add broad refactors across multiple examples when a focused change in one demo is enough.
