# Water Footprint Demo

- Date: 2026-06-08
- Area: `examples/water_world`

## Scope

Update the water world example to visibly exercise grouped footprint water.

## Design

- Keep the existing fitted main/decorative pools as ordinary water-body examples.
- Add a segmented channel made from multiple explicit rectangular water bodies
  sharing one `ContinuityGroup`.
- Add stone slabs under the channel so the footprint water is easy to inspect
  from side and grazing views.
- Add one skimming body near the channel to exercise ripples on the new path.

## Validation

- Run `cd examples/water_world && env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go test ./...`.
- Manual smoke run remains recommended because the target is a visual renderer
  behavior.
