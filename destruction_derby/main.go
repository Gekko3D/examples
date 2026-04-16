package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	demoVoxelResolution   float32 = 0.15
	projectileVoxelSize   float32 = 0.12
	defaultDestroyRadius  float32 = 0.35
	minDestroyRadius      float32 = 0.10
	maxDestroyRadius      float32 = 1.00
	destroyRadiusStep     float32 = 0.05
	projectileSpeed       float32 = 22.0
	projectileImpactForce float32 = 13.0
)

const (
	Startup State = iota
	Quit
)

type DemoModule struct{}

type DemoSceneEntity struct{}

type ProjectileComponent struct {
	DestructionRadius float32
	ImpactImpulse     float32
}

type DemoState struct {
	FloorModel      AssetId
	CubeModel       AssetId
	SphereModel     AssetId
	PillarModel     AssetId
	ProjectileModel AssetId

	FloorPalette      AssetId
	WallPalette       AssetId
	TowerPalette      AssetId
	PillarPalette     AssetId
	ProjectilePalette AssetId

	DestructionRadius float32
	DebrisAlive       int
	HudEntity         EntityId
	ResetRequested    bool
	RespawnPending    bool
}

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{},
		AssetServerModule{},
		InputModule{},
		VoxelRtModule{
			WindowWidth:  1280,
			WindowHeight: 720,
			WindowTitle:  "Destruction Derby",
		},
		PhysicsModule{Synchronous: true},
		VoxPhysicsModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		DestructionModule{},
		UiModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{
		DestructionRadius: defaultDestroyRadius,
	})

	app.UseSystem(
		System(setupSystem).
			InStage(Prelude).
			InState(OnEnter(Startup)),
	)
	app.UseSystem(
		System(resetRequestSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
	app.UseSystem(
		System(radiusControlSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
	app.UseSystem(
		System(resetSystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(respawnSystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(clickDestroySystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(projectileImpactSystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(updateHudSystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(quitSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
}

func setupSystem(cmd *Commands, assets *AssetServer, state *DemoState) {
	ensureAssets(assets, state)
	if state.HudEntity == 0 {
		state.HudEntity = cmd.AddEntity(buildHudPanel(state))
	}
	spawnArena(cmd, state)
	printControls()
}

func ensureAssets(assets *AssetServer, state *DemoState) {
	if state.FloorModel != (AssetId{}) {
		return
	}

	state.FloorModel = assets.CreateCubeModel(160, 4, 160, 1.0)
	state.CubeModel = assets.CreateCubeModel(8, 8, 8, 1.0)
	state.SphereModel = assets.CreateSphereModel(5, 1.0)
	state.PillarModel = assets.CreateCubeModel(6, 20, 6, 1.0)
	state.ProjectileModel = assets.CreateSphereModel(3, 1.0)

	state.FloorPalette = assets.CreateSimplePalette([4]uint8{82, 86, 94, 255})
	state.WallPalette = assets.CreateSimplePalette([4]uint8{211, 117, 77, 255})
	state.TowerPalette = assets.CreateSimplePalette([4]uint8{104, 176, 233, 255})
	state.PillarPalette = assets.CreateSimplePalette([4]uint8{240, 212, 96, 255})
	state.ProjectilePalette = assets.CreateSimplePalette([4]uint8{255, 74, 74, 255})
}

func spawnArena(cmd *Commands, state *DemoState) {
	spawnSceneEntity(
		cmd,
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.18,
			Color:     [3]float32{1, 1, 1},
		},
	)

	spawnSceneEntity(
		cmd,
		&SkyboxLayerComponent{
			LayerType:  SkyboxLayerGradient,
			Resolution: [2]int{1024, 512},
			ColorA:     mgl32.Vec3{0.92, 0.58, 0.34},
			ColorB:     mgl32.Vec3{0.08, 0.10, 0.18},
			Opacity:    1.0,
			Priority:   0,
			Smooth:     true,
			BlendMode:  SkyboxBlendAlpha,
		},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: mgl32.Vec3{8, 12, 6},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-50), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(25), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    1.05,
			Color:        [3]float32{1.0, 0.96, 0.9},
			Range:        500,
			CastsShadows: true,
		},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: mgl32.Vec3{4.5, 3.2, -1.5},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:      LightTypePoint,
			Intensity: 5.5,
			Color:     [3]float32{1.0, 0.78, 0.48},
			Range:     18,
		},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: mgl32.Vec3{0, 0, 0},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 3.8, 10.5},
			LookAt:   mgl32.Vec3{0, 1.6, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -12,
			Fov:      60,
			Aspect:   1280.0 / 720.0,
			Near:     0.1,
			Far:      300,
		},
		&FlyingCameraComponent{Speed: 12, Sensitivity: 0.1},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: mgl32.Vec3{0, -0.2, 0},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      state.FloorModel,
			VoxelPalette:    state.FloorPalette,
			VoxelResolution: 0.1,
		},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.65, Restitution: 0.15},
	)

	for y := 0; y < 4; y++ {
		for x := 0; x < 5; x++ {
			spawnSceneEntity(
				cmd,
				&TransformComponent{
					Position: mgl32.Vec3{-3.0 + float32(x)*1.25, 0.6 + float32(y)*1.24, -3.3},
					Rotation: mgl32.QuatIdent(),
					Scale:    mgl32.Vec3{1, 1, 1},
				},
				&VoxelModelComponent{
					VoxelModel:      state.CubeModel,
					VoxelPalette:    state.WallPalette,
					VoxelResolution: demoVoxelResolution,
				},
				&RigidBodyComponent{Mass: 1, GravityScale: 1},
				&ColliderComponent{Friction: 0.4, Restitution: 0.3},
			)
		}
	}

	for i := 0; i < 5; i++ {
		spawnSceneEntity(
			cmd,
			&TransformComponent{
				Position: mgl32.Vec3{-5.6, 0.8 + float32(i)*1.48, 2.4},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{
				VoxelModel:      state.SphereModel,
				VoxelPalette:    state.TowerPalette,
				VoxelResolution: demoVoxelResolution,
			},
			&RigidBodyComponent{Mass: 1.2, GravityScale: 1},
			&ColliderComponent{Friction: 0.3, Restitution: 0.45},
		)
	}

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: mgl32.Vec3{4.8, 1.5, -0.8},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      state.PillarModel,
			VoxelPalette:    state.PillarPalette,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{Mass: 6.5, GravityScale: 1},
		&ColliderComponent{Friction: 0.45, Restitution: 0.15},
	)

	state.DebrisAlive = 0
}

func spawnSceneEntity(cmd *Commands, components ...any) EntityId {
	components = append(components, &DemoSceneEntity{})
	return cmd.AddEntity(components...)
}

func resetRequestSystem(input *Input, state *DemoState) {
	if input.JustPressed[KeyR] {
		state.ResetRequested = true
	}
}

func radiusControlSystem(input *Input, state *DemoState) {
	if input == nil || state == nil {
		return
	}

	changed := false
	if input.JustPressed[KeyEqual] || input.JustPressed[KeyKPPlus] || (!input.GuiCaptured && input.MouseScrollY > 0) {
		state.DestructionRadius += destroyRadiusStep
		changed = true
	}
	if input.JustPressed[KeyMinus] || input.JustPressed[KeyKPMinus] || (!input.GuiCaptured && input.MouseScrollY < 0) {
		state.DestructionRadius -= destroyRadiusStep
		changed = true
	}

	if changed {
		state.DestructionRadius = clamp(state.DestructionRadius, minDestroyRadius, maxDestroyRadius)
	}
}

func resetSystem(cmd *Commands, state *DemoState) {
	if state == nil || !state.ResetRequested {
		return
	}

	toRemove := map[EntityId]struct{}{}

	MakeQuery1[DemoSceneEntity](cmd).Map(func(eid EntityId, _ *DemoSceneEntity) bool {
		toRemove[eid] = struct{}{}
		return true
	})
	MakeQuery1[DebrisComponent](cmd).Map(func(eid EntityId, _ *DebrisComponent) bool {
		toRemove[eid] = struct{}{}
		return true
	})
	MakeQuery1[ProjectileComponent](cmd).Map(func(eid EntityId, _ *ProjectileComponent) bool {
		toRemove[eid] = struct{}{}
		return true
	})

	for eid := range toRemove {
		if eid == state.HudEntity {
			continue
		}
		cmd.RemoveEntity(eid)
	}

	state.ResetRequested = false
	state.RespawnPending = true
}

func respawnSystem(cmd *Commands, state *DemoState) {
	if state == nil || !state.RespawnPending {
		return
	}

	sceneCount := 0
	MakeQuery1[DemoSceneEntity](cmd).Map(func(eid EntityId, _ *DemoSceneEntity) bool {
		if eid != state.HudEntity {
			sceneCount++
		}
		return true
	})
	if sceneCount > 0 {
		return
	}

	debrisCount := 0
	MakeQuery1[DebrisComponent](cmd).Map(func(eid EntityId, _ *DebrisComponent) bool {
		debrisCount++
		return true
	})
	if debrisCount > 0 {
		return
	}

	projectileCount := 0
	MakeQuery1[ProjectileComponent](cmd).Map(func(eid EntityId, _ *ProjectileComponent) bool {
		projectileCount++
		return true
	})
	if projectileCount > 0 {
		return
	}

	spawnArena(cmd, state)
	state.RespawnPending = false
}

func clickDestroySystem(vox *VoxelRtState, input *Input, cmd *Commands, queue *DestructionQueue, state *DemoState) {
	if vox == nil || input == nil || queue == nil || state == nil || input.GuiCaptured {
		return
	}

	leftClick := input.JustPressed[MouseButtonLeft]
	rightClick := input.JustPressed[MouseButtonRight]
	if !leftClick && !rightClick {
		return
	}

	origin, dir, ok := aimRay(cmd, vox, input)
	if !ok {
		return
	}

	if rightClick {
		spawnProjectile(cmd, state, origin, dir)
		return
	}

	hit := vox.Raycast(origin, dir, 250)
	if !hit.Hit {
		return
	}

	hitPos := origin.Add(dir.Mul(hit.T))
	queue.Events = append(queue.Events, DestructionEvent{
		Entity: hit.Entity,
		Center: hitPos,
		Radius: state.DestructionRadius,
	})
	applyImpulseToEntity(cmd, hit.Entity, dir.Normalize().Mul(10.0))
}

func projectileImpactSystem(cmd *Commands, proxy *PhysicsProxy, queue *DestructionQueue) {
	if proxy == nil || queue == nil {
		return
	}

	events := proxy.DrainCollisionEvents()
	if len(events) == 0 {
		return
	}

	processed := map[EntityId]struct{}{}
	for _, event := range events {
		if event.Type != CollisionEventEnter {
			continue
		}

		if _, seen := processed[event.A]; !seen {
			if projectile, ok := projectileForEntity(cmd, event.A); ok {
				handleProjectileImpact(cmd, queue, event.A, event.B, event.Point, projectile)
				processed[event.A] = struct{}{}
			}
		}

		if _, seen := processed[event.B]; !seen {
			if projectile, ok := projectileForEntity(cmd, event.B); ok {
				handleProjectileImpact(cmd, queue, event.B, event.A, event.Point, projectile)
				processed[event.B] = struct{}{}
			}
		}
	}
}

func handleProjectileImpact(cmd *Commands, queue *DestructionQueue, projectileEid, targetEid EntityId, impactPoint mgl32.Vec3, projectile ProjectileComponent) {
	cmd.RemoveEntity(projectileEid)

	if targetEid == 0 {
		return
	}
	if _, targetIsProjectile := projectileForEntity(cmd, targetEid); targetIsProjectile {
		return
	}
	if !entityHasVoxelModel(cmd, targetEid) {
		return
	}

	queue.Events = append(queue.Events, DestructionEvent{
		Entity: targetEid,
		Center: impactPoint,
		Radius: projectile.DestructionRadius,
	})

	if velocity, ok := velocityForEntity(cmd, projectileEid); ok && velocity.Len() > 0 {
		applyImpulseToEntity(cmd, targetEid, velocity.Normalize().Mul(projectile.ImpactImpulse))
	}
}

func updateHudSystem(cmd *Commands, state *DemoState) {
	if state == nil || state.HudEntity == 0 {
		return
	}

	debrisAlive := 0
	MakeQuery1[DebrisComponent](cmd).Map(func(eid EntityId, _ *DebrisComponent) bool {
		debrisAlive++
		return true
	})
	state.DebrisAlive = debrisAlive

	cmd.AddComponents(state.HudEntity, buildHudPanel(state))
}

func spawnProjectile(cmd *Commands, state *DemoState, origin, dir mgl32.Vec3) {
	spawnPos := origin.Add(dir.Normalize().Mul(1.0))

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: spawnPos,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      state.ProjectileModel,
			VoxelPalette:    state.ProjectilePalette,
			VoxelResolution: projectileVoxelSize,
		},
		&RigidBodyComponent{
			Mass:         0.75,
			GravityScale: 1.0,
			Velocity:     dir.Normalize().Mul(projectileSpeed),
		},
		&ColliderComponent{Friction: 0.2, Restitution: 0.15},
		&LifetimeComponent{TimeLeft: 15},
		&ProjectileComponent{
			DestructionRadius: state.DestructionRadius,
			ImpactImpulse:     projectileImpactForce,
		},
	)
}

func applyImpulseToEntity(cmd *Commands, target EntityId, impulse mgl32.Vec3) {
	if impulse.Len() == 0 {
		return
	}

	MakeQuery1[RigidBodyComponent](cmd).Map(func(eid EntityId, rb *RigidBodyComponent) bool {
		if eid != target {
			return true
		}
		if rb != nil && !rb.IsStatic {
			rb.ApplyImpulse(impulse)
		}
		return false
	})
}

func velocityForEntity(cmd *Commands, eid EntityId) (mgl32.Vec3, bool) {
	velocity := mgl32.Vec3{}
	found := false
	MakeQuery1[RigidBodyComponent](cmd).Map(func(candidate EntityId, rb *RigidBodyComponent) bool {
		if candidate != eid {
			return true
		}
		if rb != nil {
			velocity = rb.Velocity
			found = true
		}
		return false
	})
	return velocity, found
}

func projectileForEntity(cmd *Commands, eid EntityId) (ProjectileComponent, bool) {
	for _, component := range cmd.GetAllComponents(eid) {
		switch typed := component.(type) {
		case *ProjectileComponent:
			return *typed, true
		case ProjectileComponent:
			return typed, true
		}
	}
	return ProjectileComponent{}, false
}

func entityHasVoxelModel(cmd *Commands, eid EntityId) bool {
	for _, component := range cmd.GetAllComponents(eid) {
		switch component.(type) {
		case *VoxelModelComponent, VoxelModelComponent:
			return true
		}
	}
	return false
}

func aimRay(cmd *Commands, vox *VoxelRtState, input *Input) (mgl32.Vec3, mgl32.Vec3, bool) {
	var cam *CameraComponent
	MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, camera *CameraComponent) bool {
		cam = camera
		return false
	})
	if cam == nil {
		return mgl32.Vec3{}, mgl32.Vec3{}, false
	}

	mouseX, mouseY := input.MouseX, input.MouseY
	if input.MouseCaptured && input.WindowWidth > 0 && input.WindowHeight > 0 {
		mouseX = float64(input.WindowWidth) * 0.5
		mouseY = float64(input.WindowHeight) * 0.5
	}

	origin, dir := vox.ScreenToWorldRay(mouseX, mouseY, cam)
	if dir.Len() == 0 {
		return mgl32.Vec3{}, mgl32.Vec3{}, false
	}
	return origin, dir.Normalize(), true
}

func buildHudPanel(state *DemoState) *UiPanel {
	return &UiPanel{
		Key:      "destruction_hud",
		Anchor:   UiAnchorTopLeft,
		Position: [2]float32{16, 16},
		Width:    300,
		Padding:  10,
		Spacing:  6,
		Scale:    0.82,
		Title:    "Destruction Derby",
		Visible:  true,
		Children: []UiNode{
			UiLabel{Text: fmt.Sprintf("Radius: %.2f", state.DestructionRadius)},
			UiLabel{Text: fmt.Sprintf("Debris alive: %d", state.DebrisAlive)},
			UiSpacer{Height: 4},
			UiLabel{Text: "LMB: Destroy | RMB: Shoot | Scroll: Radius", Scale: 0.72, Dim: true},
			UiLabel{Text: "R or button: Reset arena | Tab: Mouse capture", Scale: 0.72, Dim: true},
			UiSpacer{Height: 6},
			UiButtonControl{
				Key:   "reset_arena",
				Label: "Reset [R]",
				Width: 128,
				Scale: 0.8,
				OnClick: func() {
					state.ResetRequested = true
				},
			},
		},
	}
}

func clamp(value, minValue, maxValue float32) float32 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func printControls() {
	fmt.Println("Destruction Derby")
	fmt.Println("LMB destroys the voxel surface under the crosshair")
	fmt.Println("RMB fires a projectile that destroys on impact")
	fmt.Println("Scroll or +/- adjusts destruction radius, R resets the arena")
	fmt.Println("Tab toggles mouse capture, Esc quits")
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
