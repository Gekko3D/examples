# Navigation Graph Lab

## Scope

Visual acceptance fixture for compact, immutable voxel navigation snapshots.

## Design

- Build two synthetic imported-world chunks in memory.
- Bake, save, reload, validate, and route through `.gkns`/`.gkng` graph contracts.
- Render synthetic voxel geometry plus accepted/rejected spans, region bounds,
  transitions, and route waypoints in 3D. Keep summary and expected failure in
  compact HUD text.
- Show route tile dependencies so snapshot/epoch publication is inspectable.
- Keep runtime streaming and locomotion out; Actiongame owns those checks.

## Validation

- `env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- `env GOCACHE=/tmp/gekko3d-gocache go run . -bake-only`
- `env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`

The headless test rejects legacy tile extensions and requires route dependency
metadata. The voxel fixture and 3D graph overlays still require a GPU smoke run.
