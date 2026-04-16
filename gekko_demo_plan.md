# Gekko3D Engine — Demo Library Plan

> **Purpose**: Provide a library of focused, self-contained demos in `examples/` that teach AI agents (and humans) how to use every major Gekko engine module. Each demo is small enough to implement in one session and targets specific subsystems.

## Engine Modules Covered

| Module | Key Files | Demo(s) |
|--------|-----------|---------|
| App bootstrap, Module, ECS, Time | `app_builder.go`, `mod.go`, `schedule.go`, `ecs.go`, `commands.go`, `mod_time.go` | #1 |
| VoxelRT renderer, Procedural primitives, Palettes, Lights | `mod_voxelrt_client.go`, `asset_procedural_primitives.go`, `light.go`, `material_presets.go` | #2 |
| Physics (sync), RigidBody, Collider, VoxPhysics | `mod_physics_module.go`, `mod_physics.go`, `mod_vox_physics.go` | #3 |
| Input, Flying Camera | `mod_input.go`, `mod_flying_camera.go` | #1, #2 |
| Hierarchy, Parent/Child transforms | `mod_hierarchy.go` | #4 |
| Skybox layers, Sun | `skybox_ecs.go` | #5 |
| Particles, Sprites | `particles_ecs.go`, `sprite_ecs.go` | #6 |
| Retained UI (panels, buttons, fields) | `mod_ui.go`, `mod_ui_retained.go` | #7 |
| Water surface, Water effects/splash | `water_surface_ecs.go`, `mod_water_effects.go`, `water_interaction_ecs.go` | #8 |
| Cellular Automata (fire, smoke) | `ca_ecs.go` | #9 |
| Destruction, Debris lifecycle | `mod_destruction.go`, `mod_lifecycle.go` | #10 |
| Vox file loading, Hierarchical models | `vox_loader.go`, `asset_vox_spawner.go` | #11 |
| Collision events, Raycasting | `mod_physics_collision.go`, `mod_voxelrt_client.go` (Raycast) | #12 |

---

## Tier 1 — Foundational (start here)

---

### Demo #1: `hello_voxel` — Minimal App Bootstrap

**What it teaches**: How to create a Gekko app, install core modules, spawn a single voxel entity with a camera, and run the game loop.

**Modules exercised**: `NewApp`, `UseStates`, `UseModules`, `TimeModule`, `AssetServerModule`, `InputModule`, `VoxelRtModule`, `FlyingCameraModule`, `Commands.AddEntity`, `TransformComponent`, `CameraComponent`, `FlyingCameraComponent`, `VoxelModelComponent`, procedural primitives (`CreateCubeModel`, `CreateSimplePalette`).

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a minimal Gekko3D engine demo called "hello_voxel".

## Project Setup
- Create directory: examples/hello_voxel/
- Create files: examples/hello_voxel/main.go, examples/hello_voxel/go.mod
- The go.mod should declare module "github.com/gekko3d/examples/hello_voxel" and
  require "github.com/gekko3d/gekko" with a replace directive:
  replace github.com/gekko3d/gekko => ../../gekko
- Copy go.sum from examples/testing-vox/go.sum as a starting point.
- Add to the root go.work file: ./examples/hello_voxel

## What to Build
A single-file main.go that:

1. Defines two states: `Startup` and `Quit` (State = iota pattern).
2. Creates a `DemoModule` struct implementing `gekko.Module`.
3. In main(), builds the app:
   ```go
   app := gekko.NewApp()
   app.UseStates(Startup, Quit)
   app.UseModules(
       gekko.TimeModule{},
       gekko.AssetServerModule{},
       gekko.InputModule{},
       gekko.VoxelRtModule{
           WindowWidth: 1280, WindowHeight: 720,
           WindowTitle: "Hello Voxel",
       },
       gekko.FlyingCameraModule{},
       DemoModule{},
   )
   app.Run()
   ```
4. DemoModule.Install registers:
   - A `setupSystem` in stage `Prelude`, state `OnEnter(Startup)` — runs once.
   - A `quitSystem` in stage `PreUpdate`, `RunAlways()` — checks KeyEscape.
5. setupSystem creates:
   - A camera entity with TransformComponent, CameraComponent (position [0,5,10],
     lookAt [0,0,0], fov 52, near 0.1, far 500), and FlyingCameraComponent{Speed: 8}.
   - A floor: CreateCubeModel(40, 1, 40, 1.0), CreateSimplePalette grey [80,82,88,255],
     TransformComponent at origin, VoxelModelComponent with PivotModeCorner.
   - A cube: CreateCubeModel(8,8,8, 1.0), blue palette [60,120,255,255],
     positioned at [0, 2, 0].
   - A sphere: CreateSphereModel(6, 1.0), red palette [255,80,80,255],
     positioned at [5, 3, -3].
   - A directional light: LightComponent{Type: LightTypeDirectional,
     Intensity: 0.8, Color: [1,0.95,0.9], Range: 500},
     with rotation via QuatRotate for angled sun.
6. Use a `setupDone bool` guard so setup only runs once.

## Key Patterns to Demonstrate
- Dot-import: `. "github.com/gekko3d/gekko"` for clean API.
- All components passed as pointers to AddEntity.
- Camera must have Yaw/Pitch fields (not just LookAt) for FlyingCamera to work.
- Tab key toggles mouse capture (handled by FlyingCameraModule).
- PivotModeCorner vs default center pivot for floor placement.

## Reference Files
- Read examples/testing-vox/main.go lines 61-90 for module installation pattern.
- Read gekko/mod_flying_camera.go for FlyingCameraComponent fields.
- Read gekko/asset_procedural_primitives.go for available procedural shapes.
- Read gekko/light.go for LightComponent and LightType constants.

## Verification
- go build should succeed from the demo directory.
- Running should open a window showing a grey floor, blue cube, red sphere,
  with directional lighting. WASD + mouse to fly around.
```

</details>

---

### Demo #2: `pbr_gallery` — Materials & Lighting

**What it teaches**: PBR material system (roughness, metallic, emission, transparency), multiple light types (point, directional, spot, ambient), procedural shape variety.

**Modules exercised**: `CreatePBRPalette`, `CreateSimplePalette`, all procedural primitives (`CreateSphereModel`, `CreateCubeModel`, `CreateConeModel`, `CreatePyramidModel`, `CreateCylinderModel`, `CreateCapsuleModel`, `CreateRampModel`), `LightComponent` with all `LightType` variants, custom `SpinnerComponent` with per-frame rotation system.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "pbr_gallery" that showcases the
PBR material system and all procedural primitives.

## Project Setup
- Create examples/pbr_gallery/ with main.go and go.mod.
- Same module/replace pattern as hello_voxel. Add to go.work.

## What to Build

1. Same app bootstrap as hello_voxel (states, core modules, VoxelRtModule,
   FlyingCameraModule, quit system).

2. Create a custom SpinnerComponent struct:
   ```go
   type SpinnerComponent struct {
       AngularSpeed mgl32.Vec3 // radians/sec per axis
   }
   ```

3. Create a SpinnerModule with a system in Update/RunAlways that queries
   TransformComponent + SpinnerComponent and applies rotation each frame:
   ```go
   delta := mgl32.QuatRotate(spin.AngularSpeed.Y()*dt, mgl32.Vec3{0,1,0})
   tr.Rotation = tr.Rotation.Mul(delta).Normalize()
   ```

4. Scene setup (in OnEnter(Startup)):
   a) Grey floor (same as hello_voxel).
   b) 4 rows of 6 spheres (CreateSphereModel(8, 1.0)), each row demonstrating
      a different PBR axis. Use CreatePBRPalette(color, roughness, metallic,
      emission, ior):
      - Row 1 (Y=12): Gold metallic, roughness varies 0.0→1.0
      - Row 2 (Y=9): Blue dielectric, roughness varies 0.0→1.0
      - Row 3 (Y=6): Green emissive, emission varies 0.0→4.0
      - Row 4 (Y=3): Glass transparent, alpha varies 255→25
      All spheres get SpinnerComponent{AngularSpeed: {0, 0.5, 0}}.
   c) One instance of each procedural shape on a separate platform area,
      each with a distinct colored palette:
      - Cube, Sphere, Cone, Pyramid, Cylinder, Capsule, Ramp
   d) Lighting:
      - 1 directional light (sun) with shadows
      - 2-3 point lights near the gallery at different warm/cool colors
      - 1 spot light aimed at the shape showcase area

5. Camera starts at position giving a good overview of the gallery.

## Key Patterns
- CreatePBRPalette(baseColor [4]uint8, roughness, metallic, emission, ior float32)
- Loop with index to parametrically vary material properties.
- LightComponent fields: Type, Color [3]float32, Intensity, Range, ConeAngle,
  CastsShadows.
- Spot light needs a rotation on its TransformComponent to aim it.

## Reference Files
- Read examples/testing-vox/main.go lines 460-504 for PBR gallery pattern.
- Read gekko/asset_procedural_primitives.go for all Create*Model signatures.
- Read gekko/light.go for LightComponent struct.
- Read gekko/mod_voxelrt_client_materials.go for material pipeline details.

## Verification
- All 7 procedural shape types visible.
- Clear visual difference across roughness/metallic/emission/transparency rows.
- Multiple light types casting proper illumination and shadows.
```

</details>

---

### Demo #3: `physics_playground` — Rigid Body Physics

**What it teaches**: How to set up the synchronous physics module, create static and dynamic rigid bodies, configure mass/friction/restitution, apply impulses, and observe gravity + collision response.

**Modules exercised**: `PhysicsModule{Synchronous: true}`, `VoxPhysicsModule`, `RigidBodyComponent`, `ColliderComponent`, `PhysicsModel`, `AccumulatedImpulse`, input-driven force application.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "physics_playground" that showcases
the rigid body physics system.

## Project Setup
- Create examples/physics_playground/ with main.go and go.mod.
- Same module/replace pattern. Add to go.work.

## What to Build

1. App bootstrap with these modules (order matters):
   ```go
   gekko.TimeModule{},
   gekko.AssetServerModule{},
   gekko.InputModule{},
   gekko.VoxelRtModule{WindowWidth: 1280, WindowHeight: 720, WindowTitle: "Physics"},
   gekko.PhysicsModule{Synchronous: true},
   gekko.VoxPhysicsModule{},
   gekko.FlyingCameraModule{},
   gekko.LifecycleModule{},
   DemoModule{},
   ```

2. Scene setup:
   a) Large static floor:
      ```go
      &TransformComponent{Position: origin, Rotation: QuatIdent(), Scale: {1,1,1}},
      &VoxelModelComponent{VoxelModel: floorModel, VoxelPalette: greyPal, PivotMode: PivotModeCorner},
      &RigidBodyComponent{IsStatic: true, Mass: 0},
      &ColliderComponent{Friction: 0.5, Restitution: 0.2},
      ```
   b) A ramp (CreateRampModel) as a static body for objects to slide down.
   c) Stack of 3-4 cubes (dynamic, Mass: 1.0) positioned to topple.
   d) Several spheres at height with different masses (0.5, 1.0, 2.0).
   e) A "heavy" cube (Mass: 5.0) that can knock the stack over.

3. Interaction system — "spawnOnClick":
   - On left mouse click: get camera, compute ScreenToWorldRay, spawn a
     sphere at camera position with velocity along the ray direction * 20.
   - Sphere gets: TransformComponent, VoxelModelComponent (sphere, red palette),
     RigidBodyComponent{Mass: 1, GravityScale: 1, Velocity: dir.Mul(20)},
     ColliderComponent{Friction: 0.3, Restitution: 0.6},
     LifetimeComponent{TimeLeft: 15} (auto-cleanup).

4. Impulse system — "forceField":
   - When pressing F key, query all non-static RigidBodyComponents and apply
     an upward impulse: rb.AccumulatedImpulse = rb.AccumulatedImpulse.Add(Vec3{0, 5, 0})

5. Directional light + a point light for visual clarity.

## Key Patterns
- Static body: IsStatic: true, Mass: 0, always needs ColliderComponent.
- Dynamic body: needs Mass > 0, GravityScale: 1 for gravity.
- AccumulatedImpulse is consumed each physics tick — it's a one-frame force.
- ScreenToWorldRay(mouseX, mouseY, cam) returns (origin, direction).
- LifetimeComponent auto-removes entities after TimeLeft seconds.
- VoxPhysicsModule is required for voxel-shaped collision (AABB from voxel geometry).

## Reference Files
- Read gekko/mod_physics_module.go for PhysicsModule install and component types.
- Read gekko/mod_physics.go for RigidBodyComponent, ColliderComponent type aliases.
- Read gekko/mod_voxelrt_client.go lines 347-360 for ScreenToWorldRay.
- Read gekko/mod_lifecycle.go for LifetimeComponent.
- Read examples/testing-vox/main.go lines 288-328 for physics entity examples.

## Verification
- Objects fall under gravity, bounce off the floor.
- Cube stack topples when hit by projectile spheres.
- F key launches all dynamic objects upward briefly.
- Spawned projectiles auto-despawn after 15 seconds.
```

</details>

---

## Tier 2 — Intermediate

---

### Demo #4: `hierarchy_orrery` — Parent/Child Transforms

**What it teaches**: Entity hierarchy with `Parent` component and `LocalTransformComponent`, hierarchical rotation (solar system orrery), dynamic reparenting.

**Modules exercised**: `HierarchyModule`, `Parent`, `LocalTransformComponent`, `TransformComponent`, nested entity trees.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "hierarchy_orrery" — a miniature
solar system using parent/child transform hierarchies.

## Project Setup
- Create examples/hierarchy_orrery/ with main.go and go.mod. Add to go.work.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   HierarchyModule, FlyingCameraModule, DemoModule.

2. Create an OrbiterComponent:
   ```go
   type OrbiterComponent struct {
       AngularSpeed float32 // radians per second around parent
       OrbitRadius  float32
       Phase        float32 // current angle accumulator
   }
   ```

3. OrbiterModule with a system in Update/RunAlways that advances Phase and
   updates LocalTransformComponent.Position to orbit around the parent's origin:
   ```go
   orb.Phase += orb.AngularSpeed * dt
   local.Position = mgl32.Vec3{
       orb.OrbitRadius * float32(math.Cos(float64(orb.Phase))),
       0,
       orb.OrbitRadius * float32(math.Sin(float64(orb.Phase))),
   }
   ```
   Also apply a self-spin to LocalTransformComponent.Rotation.

4. Scene setup:
   a) Sun: large yellow emissive sphere at origin. No Parent component.
      Also a point light co-located with it.
      ```go
      sunEntity := cmd.AddEntity(
          &TransformComponent{Position: {0,0,0}, ...},
          &LocalTransformComponent{Position: {0,0,0}, ...},
          &VoxelModelComponent{...yellow emissive palette...},
      )
      ```
   b) Planet 1: medium blue sphere orbiting sun.
      ```go
      planet1 := cmd.AddEntity(
          &LocalTransformComponent{Position: {8,0,0}, Rotation: QuatIdent(), Scale: {1,1,1}},
          &TransformComponent{...},
          &Parent{Entity: sunEntity},
          &VoxelModelComponent{...},
          &OrbiterComponent{AngularSpeed: 0.5, OrbitRadius: 8},
      )
      ```
   c) Moon of Planet 1: small grey sphere orbiting planet1.
      ```go
      cmd.AddEntity(
          &LocalTransformComponent{Position: {2,0,0}, ...},
          &TransformComponent{...},
          &Parent{Entity: planet1},
          &VoxelModelComponent{...},
          &OrbiterComponent{AngularSpeed: 2.0, OrbitRadius: 2},
      )
      ```
   d) Planet 2 with 2 moons, Planet 3 with rings (multiple small cubes orbiting
      at same radius but different phases), etc.

5. Camera positioned at [0, 20, 25] looking down at the system.

## Key Patterns
- Every child needs BOTH LocalTransformComponent AND TransformComponent.
- The Parent component links to the parent's EntityId.
- HierarchyModule's TransformHierarchySystem computes world transforms from
  local transforms + parent chain (up to 8 depth passes).
- Root entities (no Parent) have their LocalTransform synced FROM TransformComponent.
- Children have their TransformComponent computed FROM parent's world + local.
- VoxelSize (0.1) affects pivot scaling in hierarchycd — use consistent scales.

## Reference Files
- Read gekko/mod_hierarchy.go for TransformHierarchySystem logic.
- Read gekko/mod_client.go lines 97-128 for TransformComponent, LocalTransformComponent, Parent.
- Read examples/testing-vox/main.go lines 211-257 for orbiter pattern.

## Verification
- Sun stays at center. Planets orbit around it.
- Moons orbit around their parent planet (compound motion).
- All transforms update correctly — no jittering or drifting.
```

</details>

---

### Demo #5: `skybox_showcase` — Procedural Skybox & Sun

**What it teaches**: Layered procedural skybox system with noise-based clouds, gradient backgrounds, stars, animated wind, and configurable sun.

**Modules exercised**: `SkyboxLayerComponent` (all layer types: Gradient, Noise, Stars, Nebula), `SkyboxSunComponent`, `SkyAmbientComponent`, blend modes, wind animation.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "skybox_showcase" that demonstrates
the procedural skybox layer system.

## Project Setup
- Create examples/skybox_showcase/ with main.go and go.mod. Add to go.work.

## What to Build

1. App bootstrap with core modules + VoxelRtModule + FlyingCameraModule.

2. Scene: just a small ground plane and a few objects for reference scale.

3. Skybox layers (each is a standalone entity with SkyboxLayerComponent):
   a) Layer 0 — Base gradient (Priority 0):
      ```go
      &SkyboxLayerComponent{
          LayerType:  SkyboxLayerGradient,
          Resolution: [2]int{1024, 512},
          ColorA:     mgl32.Vec3{0.15, 0.3, 0.6},  // horizon
          ColorB:     mgl32.Vec3{0.02, 0.05, 0.15}, // zenith (dark blue)
          Opacity:    1.0,
          Priority:   0,
          Smooth:     true,
          BlendMode:  SkyboxBlendAlpha,
      }
      ```
   b) Layer 1 — Stars (Priority 1):
      ```go
      &SkyboxLayerComponent{
          LayerType:  SkyboxLayerStars,
          Resolution: [2]int{2048, 1024},
          Seed:       12345,
          Scale:      20.0,
          ColorA:     mgl32.Vec3{1, 1, 1},
          Threshold:  0.92,
          Opacity:    0.6,
          Priority:   1,
          BlendMode:  SkyboxBlendAdd,
      }
      ```
   c) Layer 2 — Nebula (Priority 2):
      ```go
      &SkyboxLayerComponent{
          LayerType:   SkyboxLayerNebula,
          NoiseType:   SkyboxNoiseSimplex,
          Seed:        77,
          Scale:       3.0,
          Octaves:     3,
          Persistence: 0.5,
          Lacunarity:  2.0,
          Resolution:  [2]int{1024, 512},
          ColorA:      mgl32.Vec3{0.6, 0.1, 0.8}, // purple
          ColorB:      mgl32.Vec3{0.1, 0.2, 0.9}, // deep blue
          Threshold:   0.4,
          Opacity:     0.35,
          Priority:    2,
          BlendMode:   SkyboxBlendAdd,
      }
      ```
   d) Layer 3 — Animated clouds (Priority 3):
      ```go
      &SkyboxLayerComponent{
          LayerType:   SkyboxLayerNoise,
          NoiseType:   SkyboxNoisePerlin,
          Seed:        42,
          Scale:       4.0,
          Octaves:     4,
          Persistence: 0.5,
          Lacunarity:  2.0,
          Resolution:  [2]int{1024, 512},
          ColorA:      mgl32.Vec3{1, 1, 1},
          ColorB:      mgl32.Vec3{0.85, 0.85, 0.9},
          Threshold:   0.5,
          Opacity:     0.7,
          Priority:    3,
          Smooth:      true,
          BlendMode:   SkyboxBlendAlpha,
          WindSpeed:   mgl32.Vec3{0.02, 0.005, 0},
      }
      ```

4. Sun:
   ```go
   cmd.AddEntity(&SkyboxSunComponent{
       Direction:              mgl32.Vec3{0.5, -0.7, 0.3}.Normalize(),
       Intensity:              2.5,
       HaloColor:              mgl32.Vec3{1.0, 0.9, 0.7},
       CoreGlowStrength:       3.0,
       CoreGlowExponent:       128,
       AtmosphereExponent:     4.0,
       AtmosphereGlowStrength: 0.4,
       DiskColor:              mgl32.Vec3{1.0, 0.95, 0.8},
       DiskStrength:           10.0,
       DiskStart:              0.9997,
       DiskEnd:                0.99995,
   })
   ```

5. Sky ambient: `&SkyAmbientComponent{SkyMix: 0.3}`

6. Interaction: pressing 1-4 keys toggles individual layer visibility by
   modifying the Opacity field (0 = hidden, restore = visible).

## Key Patterns
- Each SkyboxLayerComponent is its own entity (no TransformComponent needed).
- Priority controls draw order (lower = behind).
- BlendMode: Alpha (standard), Add (glow/stars), Multiply (darken).
- WindSpeed animates the noise offset over time automatically.
- SkyboxSunComponent creates a visible sun disk in the skybox.
- Changes to layer fields are picked up automatically each frame.

## Reference Files
- Read gekko/skybox_ecs.go for all component fields.
- Read examples/testing-vox/main.go lines 619-667 for skybox layer examples.

## Verification
- Beautiful layered sky with gradient, stars, nebula, and animated clouds.
- Sun visible as bright disk with halo.
- Pressing 1-4 toggles layers on/off.
```

</details>

---

### Demo #6: `particles_and_sprites` — Particle Emitters & Billboards

**What it teaches**: GPU particle emitters with configurable spawn rate, lifetime, velocity, gravity, drag, cone angle, atlas sprites. World-space and UI-space billboards.

**Modules exercised**: `ParticleEmitterComponent`, `SpriteComponent`, `BillboardMode`, `SpriteAlphaMode`, texture atlas creation, `LifetimeComponent`.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "particles_and_sprites" showcasing
the GPU particle system and sprite billboard rendering.

## Project Setup
- Create examples/particles_and_sprites/ with main.go and go.mod. Add to go.work.
- Copy assets/particle_atlas.png from examples/testing-vox/assets/ into the
  new demo's assets/ directory.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   FlyingCameraModule, LifecycleModule, DemoModule.

2. Scene setup:
   a) Floor + directional light (reuse hello_voxel pattern).
   b) Load particle atlas texture:
      `atlasId := assets.CreateTexture("assets/particle_atlas.png")`

3. Particle emitters (each its own entity with Transform + ParticleEmitterComponent):
   a) Fire fountain:
      ```go
      &ParticleEmitterComponent{
          Enabled: true, MaxParticles: 8192,
          SpawnRate: 400, LifetimeRange: [2]float32{1.0, 2.5},
          StartSpeedRange: [2]float32{8, 14},
          StartSizeRange: [2]float32{0.1, 0.25},
          StartColorMin: [4]float32{1.0, 0.5, 0.0, 1.0},
          StartColorMax: [4]float32{1.0, 0.2, 0.0, 0.8},
          Gravity: 9.8, Drag: 0.1, ConeAngleDegrees: 15,
          SpriteIndex: 4, AtlasCols: 4, AtlasRows: 4,
          Texture: atlasId, AlphaMode: SpriteAlphaLuminance,
      }
      ```
   b) Snow/confetti (wide cone, slow, many particles):
      ConeAngleDegrees: 180, StartSpeedRange: [2]float32{0.5, 2},
      Gravity: 1.5, white/blue colors.
   c) Jet stream (narrow cone, fast, no gravity):
      ConeAngleDegrees: 3, Gravity: 0, StartSpeedRange: [2]float32{20, 30},
      cyan/white colors.
   d) Smoke puff (wide cone, drag-heavy, large particles):
      ConeAngleDegrees: 40, Gravity: -1 (rises), Drag: 3.0,
      StartSizeRange: [2]float32{0.3, 0.6}, grey colors.

4. Sprite billboards:
   a) World-space spherical billboards (like grass tufts or markers):
      ```go
      &SpriteComponent{
          Enabled: true,
          Position: mgl32.Vec3{x, 1.5, z},
          Size: [2]float32{1.5, 1.5},
          Color: [4]float32{1,1,1,1},
          BillboardMode: BillboardSpherical,
          Texture: atlasId, AtlasCols: 4, AtlasRows: 4, SpriteIndex: 0,
      }
      ```
   b) Cylindrical billboards (Y-axis aligned, like grass):
      BillboardMode: BillboardCylindrical
   c) UI-space sprite (fixed on screen):
      ```go
      &SpriteComponent{
          Enabled: true, IsUI: true,
          Position: mgl32.Vec3{50, 50, 0}, // screen pixels
          Size: [2]float32{64, 64},
          Color: [4]float32{1,1,1,1},
          SpriteIndex: 2, AtlasCols: 4, AtlasRows: 4,
          Texture: atlasId,
      }
      ```

5. Interaction: pressing 1-4 toggles emitters on/off (Enabled field).

## Key Patterns
- ParticleEmitterComponent needs a co-located TransformComponent for position.
- Particles are GPU-simulated — the component only configures spawn parameters.
- SpriteAlphaLuminance uses brightness as alpha (good for fire/smoke atlases).
- SpriteAlphaTexture uses the texture's alpha channel directly.
- BillboardSpherical always faces camera. Cylindrical only rotates around Y.
- IsUI: true makes position/size in screen pixels instead of world units.

## Reference Files
- Read gekko/particles_ecs.go for ParticleEmitterComponent fields.
- Read gekko/sprite_ecs.go for SpriteComponent and BillboardMode.
- Read examples/testing-vox/main.go lines 670-715 for emitter/sprite examples.

## Verification
- Fire fountain shoots upward with warm colors, particles arc under gravity.
- Snow drifts gently downward everywhere.
- Jet stream shoots horizontally with no sag.
- Sprites face camera correctly (spherical vs cylindrical).
- UI sprite stays fixed at screen corner regardless of camera movement.
```

</details>

---

### Demo #7: `ui_dashboard` — Retained UI System

**What it teaches**: Building interactive UI panels with the retained UI system — labels, buttons, text fields, number fields, select cycles, rows, columns, anchoring, and scrolling.

**Modules exercised**: `UiModule`, `UiPanel`, `UiLabel`, `UiButtonControl`, `UiTextField`, `UiNumberField`, `UiSelectCycle`, `UiRow`, `UiColumn`, `UiSpacer`, `UiAnchor` variants.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "ui_dashboard" showcasing the
retained UI panel system.

## Project Setup
- Create examples/ui_dashboard/ with main.go and go.mod. Add to go.work.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   FlyingCameraModule, UiModule, DemoModule.

2. Create a DemoConfig resource struct:
   ```go
   type DemoConfig struct {
       PlayerName    string
       Speed         float32
       Gravity       float32
       RenderMode    int
       ShowDebug     bool
   }
   ```

3. Scene: minimal floor + a few objects to have something behind the UI.

4. UI system — register a system in Update/RunAlways that creates UI panels
   as ECS entities with UiPanel components. The UI is "retained" — you create
   the panel entity once and update the UiPanel component each frame.

   a) Settings panel (top-left):
      ```go
      cmd.AddComponents(settingsEntity, &UiPanel{
          Key: "settings", Anchor: UiAnchorTopLeft,
          Position: [2]float32{10, 10}, Width: 280,
          Title: "Settings", Visible: true,
          Children: []UiNode{
              UiRow{Children: []UiNode{
                  UiLabel{Text: "Name"}, UiTextField{Key: "name",
                      Value: config.PlayerName,
                      OnCommit: func(s string) { config.PlayerName = s }},
              }},
              UiRow{Children: []UiNode{
                  UiLabel{Text: "Speed"}, UiNumberField{Key: "speed",
                      Value: config.Speed, Precision: 1,
                      OnCommit: func(v float32) { config.Speed = v }},
              }},
              UiRow{Children: []UiNode{
                  UiLabel{Text: "Gravity"}, UiNumberField{Key: "grav",
                      Value: config.Gravity, Precision: 2,
                      OnCommit: func(v float32) { config.Gravity = v }},
              }},
              UiRow{Children: []UiNode{
                  UiLabel{Text: "Render"}, UiSelectCycle{Key: "render",
                      Options: []string{"Lit", "Albedo", "Normals"},
                      Selected: config.RenderMode,
                      OnChange: func(i int) { config.RenderMode = i }},
              }},
              UiSpacer{Height: 10},
              UiButtonControl{Key: "reset", Label: "Reset Defaults",
                  OnClick: func() { resetDefaults(config) }},
          },
      })
      ```

   b) Stats panel (top-right): shows FPS and profiler info.
      ```go
      &UiPanel{
          Key: "stats", Anchor: UiAnchorTopRight,
          Position: [2]float32{10, 10}, Width: 200,
          Title: "Stats", Visible: true,
          Children: []UiNode{
              UiLabel{Text: fmt.Sprintf("FPS: %.0f", voxState.FPS())},
              UiLabel{Text: fmt.Sprintf("Entities: %d", entityCount)},
          },
      }
      ```

   c) Actions panel (bottom-left) with multiple buttons.

   d) Scrollable panel (bottom-right) with MaxHeight set and many labels
      to demonstrate scroll behavior.

5. Toggle panels with keyboard shortcuts (F1-F4).

## Key Patterns
- UiPanel is an ECS component on an entity. Create the entity once in setup,
  then update the UiPanel component each frame with current values.
- Use AddComponents to replace the UiPanel each frame (it's a full rebuild).
- Children is a []UiNode slice — use concrete types (UiLabel, UiRow, etc.)
  which all implement the UiNode interface.
- OnCommit fires when Enter is pressed in a text/number field.
- OnChange fires on every keystroke (text) or click (select cycle).
- UiPanel.Visible controls whether the panel renders.
- MaxHeight enables scrolling when content exceeds it.
- Mouse is auto-captured by UI when hovering panels (input.GuiCaptured).
- UiRow with LabelWidth aligns label + control pairs nicely.

## Reference Files
- Read gekko/mod_ui_retained.go lines 94-210 for all UiNode types.
- Read gekko/mod_ui.go for UiAnchor constants and resolveUiPosition.
- Read examples/testing-vox/main.go for panel creation pattern.

## Verification
- Settings panel lets you edit values, commit with Enter.
- Select cycle clicks through options.
- Stats panel shows live FPS.
- Scrollable panel scrolls with mouse wheel.
- F1-F4 toggles panel visibility.
```

</details>

---

## Tier 3 — Advanced (multi-system integration)

---

### Demo #8: `water_world` — Water Surface & Splash Effects

**What it teaches**: Water surface rendering with flow, waves, refraction, and physics-triggered splash particle effects.

**Modules exercised**: `WaterSurfaceComponent`, `WaterSplashEffectComponent`, `WaterEffectsModule`, water-physics interaction, particle atlas for splashes.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "water_world" that showcases the
water surface rendering and splash particle effects.

## Project Setup
- Create examples/water_world/ with main.go and go.mod. Add to go.work.
- Copy particle_atlas.png from examples/testing-vox/assets/.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   PhysicsModule{Synchronous: true}, VoxPhysicsModule{},
   WaterEffectsModule{}, FlyingCameraModule, LifecycleModule, DemoModule.

2. Scene setup:
   a) Build a basin (pool) from static voxel walls and floor:
      - Floor slab under the pool
      - 4 walls around it (front, back, left, right)
      - All with RigidBodyComponent{IsStatic: true} + ColliderComponent
   b) Water surface inside the basin:
      ```go
      cmd.AddEntity(
          &TransformComponent{Position: poolCenter, ...},
          &WaterSurfaceComponent{
              HalfExtents:     [2]float32{7, 5},
              Depth:           2.0,
              Color:           [3]float32{0.2, 0.45, 0.75},
              AbsorptionColor: [3]float32{0.05, 0.1, 0.2},
              Opacity:         0.35,
              Roughness:       0.12,
              Refraction:      0.2,
              FlowDirection:   [2]float32{1, 0.3},
              FlowSpeed:       0.8,
              WaveAmplitude:   0.025,
          },
          &WaterSplashEffectComponent{
              Texture: particleAtlas, AtlasCols: 4, AtlasRows: 4,
              SplashSprite: 5, SpraySprite: 9, FlashSprite: 10,
              MinImpactSpeed: 2.0, StrengthScale: 1.0,
          },
      )
      ```
   c) Objects above the pool that will fall in:
      - Spheres, cubes at varying heights with RigidBody + Collider
      - Different masses and restitution values
   d) Click-to-spawn projectile (same as physics_playground) aimed at pool.

3. A second, smaller decorative pool with different water parameters:
   - Calmer water: lower FlowSpeed, WaveAmplitude
   - Different color (turquoise/emerald)

4. Lighting: directional sun + blue-tinted point lights near water for ambiance.

## Key Patterns
- WaterSurfaceComponent is on an entity with a TransformComponent.
- The transform's Position.Y defines the water surface height.
- HalfExtents define the XZ rectangle of the water surface.
- Depth is how deep the water volume extends below the surface.
- WaterSplashEffectComponent on the SAME entity as WaterSurfaceComponent
  configures splash particles when physics objects hit the water.
- FlowDirection + FlowSpeed create animated surface flow.
- WaveAmplitude adds vertical displacement to the surface.
- Splash effects require WaterEffectsModule and a particle atlas texture.
- Objects need RigidBody + Collider to trigger water interaction events.

## Reference Files
- Read gekko/water_surface_ecs.go for WaterSurfaceComponent fields.
- Read gekko/mod_water_effects.go for WaterSplashEffectComponent and splash spawning.
- Read gekko/water_interaction_ecs.go for water-physics interaction.
- Read examples/testing-vox/main.go lines 330-416 for basin + water example.

## Verification
- Water surface visible with flow animation and wave displacement.
- Objects falling into water create splash particles (droplets + spray + flash).
- Different pool color/parameters are visually distinct.
- Refraction effect visible on objects behind/under water.
```

</details>

---

### Demo #9: `fire_and_smoke` — Cellular Automata Volumes

**What it teaches**: GPU-simulated volumetric fire and smoke using the cellular automata system, with different presets (torch, campfire, jet, explosion), intensity fading, and appearance customization.

**Modules exercised**: `CellularVolumeComponent`, `CellularType` (Fire, Smoke), `CAVolumePreset`, appearance overrides, intensity/fade parameters.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "fire_and_smoke" that showcases
the GPU cellular automata volume system for fire and smoke effects.

## Project Setup
- Create examples/fire_and_smoke/ with main.go and go.mod. Add to go.work.

## What to Build

1. App bootstrap with core modules + VoxelRtModule + FlyingCameraModule + DemoModule.

2. Scene: dark floor, moody directional lighting (low intensity), warm point
   lights near fire sources.

3. Fire/smoke volumes — each is an entity with TransformComponent +
   CellularVolumeComponent:

   a) Torch (tall, narrow, bright):
      ```go
      &TransformComponent{Position: {-8, 5, 0}, Scale: {1.5, 3, 1.5}},
      &CellularVolumeComponent{
          Type: CellularFire, Preset: CAVolumePresetTorch,
          Resolution: [3]int{22, 42, 22}, TickRate: 28,
          Diffusion: 0.08, Buoyancy: 0.98,
          Cooling: 0.025, Dissipation: 0.01,
      },
      ```
   b) Campfire (wide, warm, with smoke above):
      Fire entity + separate smoke entity positioned slightly above.
      Smoke uses CellularSmoke with appearance overrides:
      ```go
      UseAppearanceOverride: true,
      ScatterColor: [3]float32{0.55, 0.5, 0.42},
      Extinction: 0.85, Emission: 0.0,
      UseShadowTintOverride: true,
      ShadowTint: [3]float32{0.25, 0.2, 0.16},
      ```
   c) Jet flame (directional, intense):
      Preset: CAVolumePresetJetFlame, narrow resolution, low buoyancy.
      Rotated transform to aim sideways.
   d) Explosion (large, expanding):
      Preset: CAVolumePresetExplosion, large resolution [48,64,48],
      high buoyancy, high emission override.

4. Intensity control demo:
   - One fire volume with UseIntensity: true, Intensity: 1.0,
     FadeInRate: 0.5, FadeOutRate: 0.3.
   - Pressing F toggles Intensity between 0 and 1, showing smooth fade.

5. Interactive: pressing 1-4 toggles Disabled field on each volume.

## Key Patterns
- CellularVolumeComponent needs a TransformComponent for world positioning.
- Transform.Scale controls the physical size of the volume in world space.
- Resolution controls the voxel grid density (higher = more detail, more cost).
- TickRate is simulation Hz (15-30 typical).
- Diffusion: how fast density spreads to neighbors (0-1).
- Buoyancy: upward bias for smoke/fire rise.
- Cooling: temperature decay per tick (fire only).
- Dissipation: density loss per tick.
- Preset provides default GPU rendering parameters; UseAppearanceOverride
  lets you customize ScatterColor, Extinction, Emission, etc.
- AnchorMode defaults to CAVolumeAnchorCenter — Transform.Position is the
  center of the volume.
- Fire and Smoke are GPU-rendered volumetrically when intensity > 0.

## Reference Files
- Read gekko/ca_ecs.go for CellularVolumeComponent fields and all presets.
- Read examples/testing-vox/main.go lines 514-617 for CA volume examples.

## Verification
- Each volume type has distinct visual character (torch vs campfire vs jet).
- Smoke volume shows grey/brown volumetric scattering above fire.
- Intensity toggle smoothly fades fire in/out.
- Toggling volumes on/off works cleanly.
```

</details>

---

### Demo #10: `destruction_derby` — Voxel Destruction & Debris

**What it teaches**: Runtime voxel destruction with sphere carving, automatic fragment splitting, debris lifecycle, and physics-driven destruction chains.

**Modules exercised**: `DestructionModule`, `DestructionQueue`, `DestructionEvent`, `LifecycleModule`, `DebrisComponent`, `EnsureEditableVoxelGeometry`, raycasting for click-to-destroy.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "destruction_derby" showcasing
the voxel destruction and debris system.

## Project Setup
- Create examples/destruction_derby/ with main.go and go.mod. Add to go.work.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   PhysicsModule{Synchronous: true}, VoxPhysicsModule{},
   FlyingCameraModule, LifecycleModule, DestructionModule, UiModule, DemoModule.

2. Scene setup:
   a) Static floor with physics.
   b) A "wall" of cubes (5 wide × 4 tall × 1 deep) — each cube is a separate
      entity with:
      - TransformComponent, VoxelModelComponent (cube 8×8×8)
      - RigidBodyComponent{Mass: 1, GravityScale: 1}
      - ColliderComponent{Friction: 0.4, Restitution: 0.3}
   c) A tower of stacked spheres.
   d) A large "pillar" (single tall cube 6×20×6) as a dramatic target.

3. Destruction system — click to destroy:
   - Left click: raycast from camera, on hit queue a DestructionEvent:
     ```go
     queue.Events = append(queue.Events, DestructionEvent{
         Entity: hit.Entity,
         Center: hitWorldPos,
         Radius: demoState.DestructionRadius,
     })
     ```
   - Also apply an impulse to the hit entity to knock it back.
   - Right click: spawn a fast projectile sphere aimed at crosshair.

4. Destruction radius control:
   - Mouse scroll or +/- keys adjust DestructionRadius (0.1 to 1.0).
   - Display current radius in a small UI panel.

5. "Reset" button (R key) that removes all entities and reruns setup.

6. UI panel showing:
   - Current destruction radius
   - Number of debris entities alive
   - Instructions: "LMB: Destroy | RMB: Shoot | Scroll: Radius"

## Key Patterns
- DestructionModule processes DestructionQueue.Events each frame.
- Each DestructionEvent carves a sphere of voxels from the target entity.
- If the carve disconnects voxel groups, they split into new entities with:
  - Their own VoxelModelComponent (OverrideGeometry with carved geometry)
  - RigidBodyComponent inheriting parent velocity + angular cross product
  - DebrisComponent with MaxAge for auto-cleanup
- Largest fragment stays as the original entity; smaller ones become debris.
- Fragments < 8 voxels are discarded.
- DebrisComponent fades material transparency in last 2 seconds before removal.
- LifecycleModule handles LifetimeComponent (projectile cleanup) and
  DebrisComponent (debris cleanup).
- DestructionRadius controls the sphere carve size in world units.

## Reference Files
- Read gekko/mod_destruction.go for DestructionEvent, processDestructionEvent.
- Read gekko/mod_lifecycle.go for DebrisComponent and debrisCleanupSystem.
- Read gekko/mod_voxelrt_client.go lines 283-292 for VoxelSphereEdit.
- Read examples/testing-vox/main.go lines 728-800 for raycast+destruction.

## Verification
- Clicking on objects carves spherical holes in voxel geometry.
- Disconnected fragments break off as physics-enabled debris.
- Debris inherits velocity and tumbles realistically.
- Debris fades out and auto-removes after ~15-20 seconds.
- Projectiles knock objects and trigger destruction on impact.
```

</details>

---

### Demo #11: `vox_models` — Loading MagicaVoxel Files

**What it teaches**: Loading .vox files from disk, spawning single models and hierarchical multi-part models with proper transforms, palettes, and pivot handling.

**Modules exercised**: `AssetServer.LoadVoxFile`, `CreateVoxelModel`, `CreateVoxelPalette`, `SpawnHierarchicalVoxelModel`, `VoxelModelComponent` fields (`VoxelModel`, `VoxelPalette`, `VoxelResolution`, `PivotMode`), `HierarchyModule`.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "vox_models" that shows how to
load and display MagicaVoxel .vox files.

## Project Setup
- Create examples/vox_models/ with main.go and go.mod. Add to go.work.
- Create examples/vox_models/assets/ directory.
- You'll need .vox files — check if any exist in examples/testing-vox/assets/
  or elsewhere in the repo. If none available, the demo should create
  procedural models as fallback and document where to place .vox files.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   HierarchyModule, FlyingCameraModule, DemoModule.

2. Asset resolution helper (check both local and prefixed paths):
   ```go
   func resolveAsset(path string) string {
       if _, err := os.Stat(path); err == nil { return path }
       alt := "examples/vox_models/" + path
       if _, err := os.Stat(alt); err == nil { return alt }
       return path
   }
   ```

3. Loading a single-model .vox file:
   ```go
   voxFileId := assets.LoadVoxFile(resolveAsset("assets/model.vox"))
   // Extract first model and palette
   modelId := assets.CreateVoxelModel(voxFile.Models[0], 1.0)
   paletteId := assets.CreateVoxelPalette(voxFile.Palette, voxFile.VoxMaterials)

   cmd.AddEntity(
       &TransformComponent{Position: {0, 2, 0}, Rotation: QuatIdent(), Scale: {1,1,1}},
       &VoxelModelComponent{VoxelModel: modelId, VoxelPalette: paletteId},
   )
   ```

4. Loading a hierarchical (multi-part) .vox scene:
   ```go
   voxFileId := assets.LoadVoxFile(resolveAsset("assets/vehicle.vox"))
   rootEntity := assets.SpawnHierarchicalVoxelModel(
       cmd, voxFileId,
       TransformComponent{Position: {10, 2, 0}, Rotation: QuatIdent(), Scale: {1,1,1}},
       1.0, // voxelScale
   )
   ```

5. Display multiple instances at different scales and rotations:
   - Same model at Scale {0.5, 0.5, 0.5} and {2, 2, 2}.
   - Rotated instances using QuatRotate.

6. VoxelResolution showcase:
   - Same model with VoxelResolution: 0 (default 0.1) vs explicit 0.05 (finer)
     vs 0.2 (chunkier).

7. PivotMode showcase:
   - PivotModeCenter (default): model rotates around its center.
   - PivotModeCorner: model's origin is at corner (useful for terrain tiles).

8. If no .vox files available, create procedural stand-ins and add comments
   showing where to drop .vox files.

## Key Patterns
- LoadVoxFile returns an AssetId for the VoxFile (contains models + palette).
- VoxFile.Models is a map of VoxModel (geometry).
- VoxFile.Palette is [256][4]uint8 color palette.
- VoxFile.VoxMaterials contains PBR material data per palette index.
- CreateVoxelModel converts a VoxModel to GPU-ready geometry.
- SpawnHierarchicalVoxelModel creates the full node tree with Parent/Local
  transforms matching the MagicaVoxel scene graph.
- VoxelResolution on VoxelModelComponent overrides the default VoxelSize (0.1).
- MagicaVoxel uses a different coordinate system — the engine handles the
  basis swap in decodeVoxRotation.

## Reference Files
- Read gekko/vox_loader.go for LoadVoxFile and VoxFile structure.
- Read gekko/asset_vox_spawner.go for SpawnHierarchicalVoxelModel.
- Read gekko/mod_assets.go for AssetServer types and CreateVoxelModel.
- Read gekko/asset_vox_model.go for CreateVoxelModel, CreateVoxelPalette.

## Verification
- Single .vox model renders with correct colors from its palette.
- Hierarchical model has correct part relationships (parts move together).
- Scale variations work correctly (0.5x smaller, 2x larger).
- Different PivotModes change the rotation center visibly.
```

</details>

---

### Demo #12: `collision_events` — Physics Events & Raycasting

**What it teaches**: Consuming collision events from the physics system, world-space raycasting, entity identification from hits, and building gameplay responses to physics interactions.

**Modules exercised**: `PhysicsProxy.DrainCollisionEvents`, `PhysicsCollisionEvent`, `VoxelRtState.Raycast`, `VoxelRtState.ScreenToWorldRay`, collision enter/stay/exit lifecycle, event-driven entity manipulation.

<details>
<summary><strong>Implementation Prompt</strong></summary>

```
You are implementing a Gekko3D demo called "collision_events" that showcases
physics collision event handling and raycasting.

## Project Setup
- Create examples/collision_events/ with main.go and go.mod. Add to go.work.
- Copy particle_atlas.png for collision spark effects.

## What to Build

1. App bootstrap with: TimeModule, AssetServerModule, InputModule, VoxelRtModule,
   PhysicsModule{Synchronous: true}, VoxPhysicsModule{},
   FlyingCameraModule, LifecycleModule, UiModule, DemoModule.

2. DemoState resource tracking collision stats:
   ```go
   type DemoState struct {
       CollisionCounts map[CollisionEventType]uint64
       RecentEvents    []string // last N event descriptions
       HighlightEntity EntityId
       HighlightTimer  float32
   }
   ```

3. Scene:
   a) Floor (static).
   b) "Bowling lane": 10 pins (small cylinders, dynamic) arranged in triangle.
   c) A bowling ball (sphere, heavy Mass: 5) that player can launch.
   d) Floating "sensor" cubes that change color on collision.

4. Collision event system (in Update):
   ```go
   func collisionEventSystem(proxy *PhysicsProxy, state *DemoState, cmd *Commands) {
       events := proxy.DrainCollisionEvents()
       for _, ev := range events {
           state.CollisionCounts[ev.Type]++
           switch ev.Type {
           case CollisionEventEnter:
               // Flash the colliding entities, spawn spark particles
               state.RecentEvents = append(state.RecentEvents,
                   fmt.Sprintf("ENTER: %v↔%v imp=%.1f", ev.A, ev.B, ev.NormalImpulse))
           case CollisionEventStay:
               // Track ongoing contacts
           case CollisionEventExit:
               // Entities separated
               state.RecentEvents = append(state.RecentEvents,
                   fmt.Sprintf("EXIT: %v↔%v", ev.A, ev.B))
           }
       }
       // Trim to last 10 events
       if len(state.RecentEvents) > 10 {
           state.RecentEvents = state.RecentEvents[len(state.RecentEvents)-10:]
       }
   }
   ```

5. Visual feedback on collision:
   - On CollisionEventEnter with NormalImpulse > threshold:
     swap the entity's palette to a "flash" color briefly,
     spawn a short-lived particle burst at ev.Point.

6. Raycast system:
   - Continuous raycast from camera center (crosshair):
     ```go
     origin, dir := voxState.ScreenToWorldRay(
         float64(input.WindowWidth/2), float64(input.WindowHeight/2), cam)
     hit := voxState.Raycast(origin, dir, 100)
     if hit.Hit {
         state.HighlightEntity = hit.Entity
         state.HighlightTimer = 0.5
     }
     ```
   - Display hit entity info in UI.

7. UI panel showing:
   - Collision counters (Enter/Stay/Exit counts).
   - Recent event log (scrollable).
   - Raycast info: entity under crosshair, distance, normal.
   - PhysicsCollisionEvent fields: Point, Normal, Penetration,
     NormalImpulse, RelativeSpeed.

8. Launch ball: press Space to spawn heavy sphere in camera direction.

## Key Patterns
- PhysicsProxy.DrainCollisionEvents() returns and clears buffered events.
- Events have Type (Enter/Stay/Exit), A/B entity IDs, contact Point/Normal.
- NormalImpulse indicates collision strength (useful for damage/sound triggers).
- RelativeSpeed is approach speed at contact.
- Call DrainCollisionEvents once per frame in a system.
- Raycast returns RaycastHit{Hit, T, Pos, Normal, Entity}.
- ScreenToWorldRay converts screen-space mouse coordinates to world ray.
- Collision events are generated by the physics engine each tick.

## Reference Files
- Read gekko/mod_physics_module.go lines 32-42 for PhysicsCollisionEvent.
- Read gekko/mod_physics_module.go lines 346-361 for DrainCollisionEvents.
- Read gekko/mod_voxelrt_client.go lines 376-417 for Raycast.
- Read gekko/mod_voxelrt_client.go lines 347-360 for ScreenToWorldRay.
- Read examples/testing-vox/main.go for collision event and raycast patterns.

## Verification
- Collision event counters increment as objects collide.
- Recent event log updates in real-time.
- High-impulse collisions trigger visual particle sparks.
- Crosshair raycast identifies entities and shows distance.
- Bowling ball knocks over pins generating many Enter events.
```

</details>

---

## Implementation Order

> [!TIP]
> Recommended sequence for an agent implementing these demos. Each demo builds on patterns introduced by previous ones.

| Phase | Demo | Dependencies |
|-------|------|-------------|
| 1 | #1 `hello_voxel` | None — establishes project scaffold pattern |
| 2 | #2 `pbr_gallery` | Reuses #1 scaffold, adds materials + lights |
| 3 | #3 `physics_playground` | Adds physics on top of #1 patterns |
| 4 | #4 `hierarchy_orrery` | Reuses #1, adds hierarchy module |
| 5 | #5 `skybox_showcase` | Standalone, minimal scene |
| 6 | #6 `particles_and_sprites` | Reuses #1, adds asset loading |
| 7 | #7 `ui_dashboard` | Reuses #1, adds UI module |
| 8 | #8 `water_world` | Combines #3 physics + new water module |
| 9 | #9 `fire_and_smoke` | Standalone CA volumes |
| 10 | #10 `destruction_derby` | Combines #3 physics + destruction |
| 11 | #11 `vox_models` | Standalone file loading + hierarchy |
| 12 | #12 `collision_events` | Combines #3 physics + raycast + UI |

## Common Scaffold

Every demo follows this pattern:

```go
package main

import (
    . "github.com/gekko3d/gekko"
    "github.com/go-gl/mathgl/mgl32"
)

const (
    Startup State = iota
    Quit
)

type DemoModule struct{}

func (DemoModule) Install(app *App, cmd *Commands) {
    app.UseSystem(System(setup).InStage(Prelude).InState(OnEnter(Startup)))
    app.UseSystem(System(quit).InStage(PreUpdate).RunAlways())
    // ... demo-specific systems ...
}

func main() {
    app := NewApp()
    app.UseStates(Startup, Quit)
    app.UseModules(
        TimeModule{},
        AssetServerModule{},
        InputModule{},
        VoxelRtModule{WindowWidth: 1280, WindowHeight: 720, WindowTitle: "Demo"},
        // ... demo-specific modules ...
        DemoModule{},
    )
    app.Run()
}
```
