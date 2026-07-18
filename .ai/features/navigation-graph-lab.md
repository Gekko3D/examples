# Navigation Graph Lab

## Scope

Phase 8 visual inspection fixture for pure-Go voxel navigation graphs.

## Design

- Build two synthetic imported-world chunks in memory.
- Bake, save, reload, validate, and route through new graph contracts.
- Render synthetic voxel geometry plus accepted/rejected spans, region bounds,
  transitions, and route waypoints in 3D. Keep summary and expected failure in
  compact HUD text.
- Keep runtime streaming and locomotion out; those belong to Phases 9 and 10.

## Validation

- `env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- `env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`

Compile and headless bake/load/route checks pass. User GPU smoke accepted the
voxel fixture and 3D graph overlays.
