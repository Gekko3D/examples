# Gekko examples

Focused samples using [Gekko](../gekko/README.md). Each sample is a separate module;
[go.work](../go.work) lists workspace membership.

| Examples | Focus |
| --- | --- |
| [hello_voxel](hello_voxel/), [vox_models](vox_models/), [testing-vox](testing-vox/) | Voxel scenes and model loading |
| [baked_normals](baked_normals/), [pbr_gallery](pbr_gallery/), [far_range_depth](far_range_depth/) | Lighting and depth |
| [particles_and_sprites](particles_and_sprites/), [water_world](water_world/), [skybox_showcase](skybox_showcase/), [deferred_decals](deferred_decals/) | Rendering features |
| [physics_playground](physics_playground/), [collision_events](collision_events/), [destruction_derby](destruction_derby/) | Physics, events and edits |
| [entity_groups](entity_groups/), [entity_lod](entity_lod/), [hierarchy_man](hierarchy_man/), [hierarchy_orrery](hierarchy_orrery/) | Entity ownership, hierarchy and LOD |
| [navigation_graph_lab](navigation_graph_lab/), [ui_dashboard](ui_dashboard/) | Navigation and UI |

Run from the sample module with `env GOCACHE=/tmp/gekko3d-gocache go run .`.
For samples outside go.work, use `env GOWORK=off GOCACHE=/tmp/gekko3d-gocache go run .`.
Windowed samples require a desktop/GPU session. `testing/` was a known failing
sample at cleanup; inspect its current code before using it as verification.
Rendering preserves the [voxel style](../gekko/README.md#voxel-rendering-style).
Historical documentation: [workspace backup](../gekko/docs/archive/README.md).
