# Navmesh Lab Demo

## Scope

Add a standalone `examples/navmesh_lab` demo for testing navmesh generation and
pathfinding without reimporting HL1 maps.

## Design Summary

- Generate a large imported voxel world locally under `examples/navmesh_lab/generated`.
- Include flat plazas, stepped ramps, stairways, terraced terrain, pillars, and
  simple buildings.
- Bake `.gknav` and `.gknavtile` sidecars through the existing
  `content.SaveNavBakeForImportedWorldManifest` path.
- Load the generated level through `StreamedLevelRuntimeModule`.
- Draw actiongame-style nav debug UI locally in the demo: navmesh edge gizmos,
  portal markers, route segment/waypoint gizmos, and a HUD stats line from the
  public runtime navigation service.
- Spawn a gizmo-sphere NPC with actiongame-style debug controls:
  `F6` selects the NPC under the aim ray, `F7` clears selection, and `F8`
  teleports the selected NPC to the aimed navmesh polygon.

## Validation

- `cd examples/navmesh_lab && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`
- Manual GPU smoke: `cd examples/navmesh_lab && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`

## Notes

The current nav builder is voxel-occupancy based. Ramps in this fixture are
stepped voxel ramps, which exercise the same step-height connectivity as the
runtime builder.
