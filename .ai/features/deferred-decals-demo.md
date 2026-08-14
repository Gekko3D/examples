# Deferred Decals Demo

## Scope

Add `examples/deferred_decals`, a self-contained scene that opts into
`DecalFeature`, creates a procedural 2x2 alpha atlas, and submits wall and
floor `DecalInstance` values through `VoxelRtState`.

## Design

- Keep atlas art procedural so the example has no asset-path dependency.
- Use the shared renderer-facing decal API; the example owns no renderer code.
- Use opaque voxel receivers with front and upward-facing normals to show
  projection orientation and normal rejection.

## Validation

- `cd examples/deferred_decals && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- Manual GPU check: run the example and verify that the marks stay fixed to the
  wall and floor while flying the camera.
