# Baked Normal Seam Demo

## Context

The voxel renderer now bakes per-voxel normals so a hit does not need to sample
neighbor voxels while shading. Cross-object correctness still matters for terrain
chunks and planet tiles: when adjacent terrain objects belong to the same group,
surface normals at the boundary must be computed with neighbor occupancy in mind.

## Scope

`examples/baked_normals` is a visual smoke demo. It is not a profiler,
benchmark, or streaming-world implementation. The scene contains:

- A 7x7 endless-looking terrain tile field with shared terrain and shadow
  grouping.
- A highlighted center cross where chunk-boundary lighting discontinuities would
  be easiest to see while scanning across many internal tile edges.
- Thin one-voxel geometry to catch normal bias and edge cases.
- A transparent voxel panel to exercise the transparent material path.
- Directional and point lighting so baked-normal artifacts are visible from
  multiple directions.

## Ownership

The demo only owns example assets and scene composition. Renderer-side invalidation
and baked-normal recomputation remain owned by `gekko/voxelrt`.

## Verification

- `cd examples/baked_normals && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- Manual GPU check: run the example and inspect the highlighted center cross and
  surrounding internal tile edges. The field should shade as one continuous slab,
  without false vertical boundaries caused by missing cross-object neighbor
  normals.
