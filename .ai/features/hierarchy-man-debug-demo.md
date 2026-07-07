# Hierarchy Man Debug Demo

- Date: 2026-06-09
- Area: `examples/hierarchy_man`

## Scope

Add a procedural visual diagnostic for transform hierarchy and authored asset
animation. The demo is intended to isolate issues seen in imported rigid-body
NPCs where deeper body parts appear increasingly shifted.

## Design

- Spawn three color-coded humanoid rigs:
  - a static raw ECS hierarchy,
  - an animated raw ECS hierarchy,
  - an animated authored asset hierarchy spawned through `SpawnAuthoredAsset`.
- Add dark root reference poles and white joint markers so visual offsets can
  be compared against explicit transform origins instead of inferred from mesh
  silhouettes.
- Add a deep calibration chain whose links alternate colors and animate at the
  root. This makes hierarchy depth lag or composition errors easy to see.
- Keep the scene generated entirely from procedural voxel cuboids.

## Debug Contract

- Child transforms are expected to compose from parent entity origins.
- Voxel renderer pivots must not move child entities.
- Authored animation clips are expected to write local rigid transforms.
- If all three humanoids align but an imported MDL does not, the bug is likely
  in the importer-generated asset data or voxel part pivot semantics rather
  than the core hierarchy system.

## Verification

- `cd examples/hierarchy_man && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- Manual smoke run: `cd examples/hierarchy_man && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`
