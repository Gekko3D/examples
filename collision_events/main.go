package main

import (
	"fmt"
	"os"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const (
	demoVoxelResolution    float32 = 0.2
	impactImpulseThreshold float32 = 1.6
	impactFlashSeconds     float32 = 0.22
	sensorFlashSeconds     float32 = 0.65
	highlightHoldSeconds   float32 = 0.5
	maxRecentEvents                = 10
)

type DemoModule struct{}

type DemoSceneEntity struct{}

type DemoTagComponent struct {
	Label string
}

type SensorComponent struct{}

type BowlingBallComponent struct{}

type VisualFeedbackComponent struct {
	BasePalette         AssetId
	CollisionPalette    AssetId
	HighlightPalette    AssetId
	SensorActivePalette AssetId
	CollisionFlash      float32
	SensorFlash         float32
}

type DemoState struct {
	FloorModel       AssetId
	LaneModel        AssetId
	RailModel        AssetId
	BackstopModel    AssetId
	PinModel         AssetId
	BallModel        AssetId
	SensorModel      AssetId
	ParticleAtlas    AssetId
	FloorPalette     AssetId
	LanePalette      AssetId
	RailPalette      AssetId
	BackstopPalette  AssetId
	PinPalette       AssetId
	BallPalette      AssetId
	FlashPalette     AssetId
	HighlightPalette AssetId
	SensorIdle       AssetId
	SensorActive     AssetId
	HudPanel         EntityId
	LogPanel         EntityId

	CollisionCounts map[CollisionEventType]uint64
	RecentEvents    []string
	HighlightEntity EntityId
	HighlightTimer  float32

	LastCollision    PhysicsCollisionEvent
	HasLastCollision bool
	LastRaycast      RaycastHit
	HasRaycastHit    bool
	LastRayLabel     string

	LaunchCount uint64
}

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{},
		AssetServerModule{},
		InputModule{},
		VoxelRtModule{
			WindowWidth:  1440,
			WindowHeight: 900,
			WindowTitle:  "Collision Events",
		},
		PhysicsModule{Synchronous: true},
		VoxPhysicsModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		UiModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{
		CollisionCounts: make(map[CollisionEventType]uint64),
	})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(launchBallSystem).InStage(Update).RunAlways())
	app.UseSystem(System(collisionEventSystem).InStage(Update).RunAlways())
	app.UseSystem(System(raycastSystem).InStage(Update).RunAlways())
	app.UseSystem(System(clickImpulseSystem).InStage(Update).RunAlways())
	app.UseSystem(System(visualFeedbackSystem).InStage(Update).RunAlways())
	app.UseSystem(System(updateHudPanelsSystem).InStage(Update).RunAlways())
	app.UseSystem(System(overlaySystem).InStage(PreRender).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func resolveDemoAsset(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}

	alt := "examples/collision_events/" + path
	if _, err := os.Stat(alt); err == nil {
		return alt
	}

	return path
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	if state == nil {
		return
	}

	state.ParticleAtlas = assets.CreateTexture(resolveDemoAsset("assets/particle_atlas.png"))
	state.FloorModel = assets.CreateCubeModel(220, 4, 220, 1.0)
	state.LaneModel = assets.CreateCubeModel(18, 1, 180, 1.0)
	state.RailModel = assets.CreateCubeModel(2, 7, 180, 1.0)
	state.BackstopModel = assets.CreateCubeModel(26, 14, 4, 1.0)
	state.PinModel = assets.CreateCubeModel(3, 8, 3, 1.0)
	state.BallModel = assets.CreateSphereModel(3.0, 1.0)
	state.SensorModel = assets.CreateCubeModel(4, 4, 4, 1.0)

	state.FloorPalette = assets.CreateSimplePalette([4]uint8{46, 50, 58, 255})
	state.LanePalette = assets.CreateSimplePalette([4]uint8{168, 128, 82, 255})
	state.RailPalette = assets.CreateSimplePalette([4]uint8{108, 82, 56, 255})
	state.BackstopPalette = assets.CreateSimplePalette([4]uint8{74, 82, 96, 255})
	state.PinPalette = assets.CreateSimplePalette([4]uint8{235, 232, 220, 255})
	state.BallPalette = assets.CreateSimplePalette([4]uint8{54, 98, 210, 255})
	state.FlashPalette = assets.CreateSimplePalette([4]uint8{255, 170, 78, 255})
	state.HighlightPalette = assets.CreateSimplePalette([4]uint8{132, 255, 140, 255})
	state.SensorIdle = assets.CreateSimplePalette([4]uint8{74, 210, 255, 255})
	state.SensorActive = assets.CreateSimplePalette([4]uint8{255, 86, 170, 255})

	spawnSceneEntity(
		cmd,
		&SkyboxLayerComponent{
			LayerType:  SkyboxLayerGradient,
			Resolution: [2]int{1024, 512},
			ColorA:     mgl32.Vec3{0.84, 0.62, 0.38},
			ColorB:     mgl32.Vec3{0.06, 0.08, 0.14},
			Opacity:    1.0,
			Priority:   0,
			Smooth:     true,
			BlendMode:  SkyboxBlendAlpha,
		},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{
			Position: mgl32.Vec3{0, 3.8, 10.5},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 3.8, 10.5},
			LookAt:   mgl32.Vec3{0, 1.2, -10},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -10,
			Fov:      60,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      400,
		},
		&FlyingCameraComponent{Speed: 12, Sensitivity: 0.1},
	)

	spawnSceneEntity(
		cmd,
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.18,
			Color:     [3]float32{0.92, 0.95, 1.0},
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
			Intensity:    1.0,
			Color:        [3]float32{1.0, 0.96, 0.9},
			Range:        500,
			CastsShadows: true,
		},
	)
	spawnSceneEntity(
		cmd,
		&TransformComponent{Position: mgl32.Vec3{0, 4.4, -10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 6.0, Color: [3]float32{1.0, 0.78, 0.62}, Range: 30},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{Position: mgl32.Vec3{-22, -0.9, -34}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{
			VoxelModel:      state.FloorModel,
			VoxelPalette:    state.FloorPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.7, Restitution: 0.1},
		&DemoTagComponent{Label: "floor"},
	)

	spawnSceneEntity(
		cmd,
		&TransformComponent{Position: mgl32.Vec3{-1.8, 0.1, -28}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{
			VoxelModel:      state.LaneModel,
			VoxelPalette:    state.LanePalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.14, Restitution: 0.05},
		&DemoTagComponent{Label: "lane"},
	)
	spawnSceneEntity(
		cmd,
		&TransformComponent{Position: mgl32.Vec3{-2.4, 0.1, -28}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{
			VoxelModel:      state.RailModel,
			VoxelPalette:    state.RailPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.3, Restitution: 0.05},
		&DemoTagComponent{Label: "rail-left"},
	)
	spawnSceneEntity(
		cmd,
		&TransformComponent{Position: mgl32.Vec3{2.0, 0.1, -28}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{
			VoxelModel:      state.RailModel,
			VoxelPalette:    state.RailPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.3, Restitution: 0.05},
		&DemoTagComponent{Label: "rail-right"},
	)
	spawnSceneEntity(
		cmd,
		&TransformComponent{Position: mgl32.Vec3{-2.4, 0.1, -28.8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{
			VoxelModel:      state.BackstopModel,
			VoxelPalette:    state.BackstopPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.4, Restitution: 0.08},
		&DemoTagComponent{Label: "backstop"},
	)

	state.HudPanel = cmd.AddEntity(buildHudPanel(state))
	state.LogPanel = cmd.AddEntity(buildEventLogPanel(state))

	spawnPinRack(cmd, state)
	spawnSensorArray(cmd, state)
	spawnBowlingBall(cmd, state, mgl32.Vec3{0, 0.9, 6.3}, mgl32.Vec3{0, 0, 0}, "ball-resting", 0)

	fmt.Println("Collision Events demo ready")
	fmt.Println("Press F to launch a heavy bowling ball from the camera center")
	fmt.Println("Aim the crosshair to raycast entities and inspect collision events")
}

func spawnPinRack(cmd *Commands, state *DemoState) {
	baseZ := float32(-22.8)
	rowSpacing := float32(1.05)
	columnSpacing := float32(0.74)
	pinIndex := 1

	for row := 0; row < 4; row++ {
		z := baseZ - float32(row)*rowSpacing
		startX := -0.5 * float32(row) * columnSpacing
		for col := 0; col <= row; col++ {
			x := startX + float32(col)*columnSpacing
			label := fmt.Sprintf("pin-%02d", pinIndex)
			pinIndex++
			spawnSceneEntity(
				cmd,
				&TransformComponent{Position: mgl32.Vec3{x, 1.1, z}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
				&VoxelModelComponent{
					VoxelModel:      state.PinModel,
					VoxelPalette:    state.PinPalette,
					VoxelResolution: demoVoxelResolution,
				},
				&RigidBodyComponent{Mass: 0.8, GravityScale: 1.0, LinearDamping: 0.01, AngularDamping: 0.04},
				&ColliderComponent{Friction: 0.45, Restitution: 0.12},
				&VisualFeedbackComponent{
					BasePalette:         state.PinPalette,
					CollisionPalette:    state.FlashPalette,
					HighlightPalette:    state.HighlightPalette,
					SensorActivePalette: state.SensorActive,
				},
				&DemoTagComponent{Label: label},
			)
		}
	}
}

func spawnSensorArray(cmd *Commands, state *DemoState) {
	sensors := []struct {
		label string
		pos   mgl32.Vec3
	}{
		{label: "sensor-left", pos: mgl32.Vec3{-1.1, 2.2, -18.0}},
		{label: "sensor-mid", pos: mgl32.Vec3{0.0, 2.8, -20.5}},
		{label: "sensor-right", pos: mgl32.Vec3{1.1, 2.2, -18.0}},
	}

	for _, sensor := range sensors {
		spawnSceneEntity(
			cmd,
			&TransformComponent{Position: sensor.pos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&VoxelModelComponent{
				VoxelModel:      state.SensorModel,
				VoxelPalette:    state.SensorIdle,
				VoxelResolution: demoVoxelResolution,
			},
			&RigidBodyComponent{IsStatic: true, Mass: 0, GravityScale: 0.0},
			&ColliderComponent{Friction: 0.2, Restitution: 0.6},
			&SensorComponent{},
			&VisualFeedbackComponent{
				BasePalette:         state.SensorIdle,
				CollisionPalette:    state.FlashPalette,
				HighlightPalette:    state.HighlightPalette,
				SensorActivePalette: state.SensorActive,
			},
			&DemoTagComponent{Label: sensor.label},
		)
	}
}

func spawnBowlingBall(cmd *Commands, state *DemoState, position, velocity mgl32.Vec3, label string, lifeSeconds float32) EntityId {
	comps := []any{
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{
			VoxelModel:      state.BallModel,
			VoxelPalette:    state.BallPalette,
			VoxelResolution: demoVoxelResolution,
		},
		&RigidBodyComponent{Mass: 1.6, GravityScale: 1.0, Velocity: velocity, LinearDamping: 0.002, AngularDamping: 0.01},
		&ColliderComponent{Friction: 0.05, Restitution: 0.12},
		&BowlingBallComponent{},
		&VisualFeedbackComponent{
			BasePalette:         state.BallPalette,
			CollisionPalette:    state.FlashPalette,
			HighlightPalette:    state.HighlightPalette,
			SensorActivePalette: state.SensorActive,
		},
		&DemoTagComponent{Label: label},
	}
	if lifeSeconds > 0 {
		comps = append(comps, &LifetimeComponent{TimeLeft: lifeSeconds})
	}
	return spawnSceneEntity(cmd, comps...)
}

func launchBallSystem(input *Input, vox *VoxelRtState, cmd *Commands, state *DemoState) {
	if input == nil || vox == nil || state == nil || !input.JustPressed[KeyF] {
		return
	}

	origin, dir, ok := centerAimRay(cmd, vox, input)
	if !ok {
		return
	}

	state.LaunchCount++
	spawnPos := origin.Add(dir.Mul(1.4))
	if spawnPos.Y() < 1.0 {
		spawnPos[1] = 1.0
	}
	velocity := dir.Mul(18.0)
	label := fmt.Sprintf("ball-launch-%02d", state.LaunchCount)
	spawnBowlingBall(cmd, state, spawnPos, velocity, label, 18)
	fmt.Printf("Launched %s from %v with velocity %v\n", label, spawnPos, velocity)
}

func collisionEventSystem(proxy *PhysicsProxy, state *DemoState, cmd *Commands) {
	if proxy == nil || state == nil {
		return
	}

	events := proxy.DrainCollisionEvents()
	if len(events) == 0 {
		return
	}

	for _, ev := range events {
		state.CollisionCounts[ev.Type]++
		state.LastCollision = ev
		state.HasLastCollision = true

		labelA := entityLabel(cmd, ev.A)
		labelB := entityLabel(cmd, ev.B)

		switch ev.Type {
		case CollisionEventEnter:
			state.RecentEvents = append(state.RecentEvents,
				fmt.Sprintf("ENTER: %s <-> %s imp=%.1f rel=%.1f", labelA, labelB, ev.NormalImpulse, ev.RelativeSpeed),
			)

			if ev.NormalImpulse >= impactImpulseThreshold {
				triggerCollisionFlash(cmd, ev.A, impactFlashSeconds)
				triggerCollisionFlash(cmd, ev.B, impactFlashSeconds)
				spawnSparkBurst(cmd, state, ev.Point, ev.Normal, ev.NormalImpulse)
			}
			triggerSensorFlash(cmd, ev.A)
			triggerSensorFlash(cmd, ev.B)

		case CollisionEventStay:
			if ev.NormalImpulse >= impactImpulseThreshold*0.7 {
				state.RecentEvents = append(state.RecentEvents,
					fmt.Sprintf("STAY:  %s <-> %s imp=%.1f rel=%.1f", labelA, labelB, ev.NormalImpulse, ev.RelativeSpeed),
				)
				triggerSensorFlash(cmd, ev.A)
				triggerSensorFlash(cmd, ev.B)
			}

		case CollisionEventExit:
			state.RecentEvents = append(state.RecentEvents,
				fmt.Sprintf("EXIT:  %s <-> %s", labelA, labelB),
			)
		}
	}

	if len(state.RecentEvents) > maxRecentEvents {
		state.RecentEvents = append([]string(nil), state.RecentEvents[len(state.RecentEvents)-maxRecentEvents:]...)
	}
}

func raycastSystem(vox *VoxelRtState, input *Input, cmd *Commands, state *DemoState) {
	if vox == nil || input == nil || state == nil {
		return
	}
	origin, dir, ok := aimRay(cmd, vox, input)
	if !ok {
		state.HasRaycastHit = false
		return
	}

	hit := vox.Raycast(origin, dir.Normalize(), 100)
	if !hit.Hit {
		state.HasRaycastHit = false
		return
	}

	state.LastRaycast = hit
	state.HasRaycastHit = true
	state.LastRayLabel = entityLabel(cmd, hit.Entity)
	state.HighlightEntity = hit.Entity
	state.HighlightTimer = highlightHoldSeconds
}

func clickImpulseSystem(input *Input, vox *VoxelRtState, cmd *Commands, state *DemoState) {
	if input == nil || vox == nil || state == nil || input.GuiCaptured || !input.JustPressed[MouseButtonLeft] {
		return
	}
	target := state.HighlightEntity
	if target == 0 || state.HighlightTimer <= 0 {
		return
	}
	if entityHasComponent[SensorComponent](cmd, target) {
		return
	}

	_, dir, ok := aimRay(cmd, vox, input)
	if !ok {
		return
	}

	push := dir.Mul(2.5)
	if isBowlingBall(cmd, target) {
		push = dir.Mul(5.5).Add(mgl32.Vec3{0, 0.15, 0})
	}
	if !applyImpulseToEntity(cmd, target, push) {
		return
	}
	fmt.Printf("Pushed %s with impulse %v\n", entityLabel(cmd, target), push)
}

func visualFeedbackSystem(cmd *Commands, time *Time, state *DemoState) {
	if time == nil || state == nil {
		return
	}

	dt := float32(time.Dt)
	if dt < 0 {
		dt = 0
	}

	if state.HighlightTimer > 0 {
		state.HighlightTimer = maxf(0, state.HighlightTimer-dt)
		if state.HighlightTimer == 0 {
			state.HighlightEntity = 0
		}
	}

	MakeQuery2[VoxelModelComponent, VisualFeedbackComponent](cmd).Map(func(eid EntityId, vox *VoxelModelComponent, visual *VisualFeedbackComponent) bool {
		if vox == nil || visual == nil {
			return true
		}

		visual.CollisionFlash = maxf(0, visual.CollisionFlash-dt)
		visual.SensorFlash = maxf(0, visual.SensorFlash-dt)

		palette := visual.BasePalette
		if visual.SensorFlash > 0 && visual.SensorActivePalette != (AssetId{}) {
			palette = visual.SensorActivePalette
		}
		if state.HighlightEntity == eid && state.HighlightTimer > 0 && visual.HighlightPalette != (AssetId{}) {
			palette = visual.HighlightPalette
		}
		if visual.CollisionFlash > 0 && visual.CollisionPalette != (AssetId{}) {
			palette = visual.CollisionPalette
		}
		vox.VoxelPalette = palette

		return true
	})
}

func updateHudPanelsSystem(cmd *Commands, state *DemoState) {
	if state == nil {
		return
	}
	if state.HudPanel != 0 {
		cmd.AddComponents(state.HudPanel, buildHudPanel(state))
	}
	if state.LogPanel != 0 {
		cmd.AddComponents(state.LogPanel, buildEventLogPanel(state))
	}
}

func overlaySystem(vox *VoxelRtState, input *Input, state *DemoState) {
	if vox == nil || input == nil || state == nil {
		return
	}

	cx := float32(input.WindowWidth) * 0.5
	cy := float32(input.WindowHeight) * 0.5
	vox.DrawText("+", cx-4, cy-10, 0.9, [4]float32{1.0, 0.95, 0.5, 1.0})

	if state.HasRaycastHit {
		label := fmt.Sprintf("%s  %.1fm", state.LastRayLabel, state.LastRaycast.T)
		vox.DrawText(label, cx+16, cy-26, 0.6, [4]float32{0.8, 1.0, 0.85, 1.0})
	}
}

func buildHudPanel(state *DemoState) *UiPanel {
	children := []UiNode{
		UiLabel{Text: fmt.Sprintf("Enter: %d", state.CollisionCounts[CollisionEventEnter])},
		UiLabel{Text: fmt.Sprintf("Stay: %d", state.CollisionCounts[CollisionEventStay])},
		UiLabel{Text: fmt.Sprintf("Exit: %d", state.CollisionCounts[CollisionEventExit])},
		UiSpacer{Height: 6},
		UiLabel{Text: "Raycast", Dim: true},
	}

	if state.HasRaycastHit {
		children = append(children,
			UiLabel{Text: fmt.Sprintf("Entity: %s (%d)", state.LastRayLabel, state.LastRaycast.Entity)},
			UiLabel{Text: fmt.Sprintf("Distance: %.2f", state.LastRaycast.T)},
			UiLabel{Text: fmt.Sprintf("Pos: [%d %d %d]", state.LastRaycast.Pos[0], state.LastRaycast.Pos[1], state.LastRaycast.Pos[2]), Scale: 0.78, Dim: true},
			UiLabel{Text: fmt.Sprintf("Normal: [%.2f %.2f %.2f]", state.LastRaycast.Normal.X(), state.LastRaycast.Normal.Y(), state.LastRaycast.Normal.Z()), Scale: 0.78, Dim: true},
		)
	} else {
		children = append(children, UiLabel{Text: "Entity: none", Dim: true})
	}

	children = append(children, UiSpacer{Height: 6}, UiLabel{Text: "Last Collision", Dim: true})
	if state.HasLastCollision {
		children = append(children,
			UiLabel{Text: fmt.Sprintf("Type: %s", state.LastCollision.Type.String())},
			UiLabel{Text: fmt.Sprintf("A/B: %d / %d", state.LastCollision.A, state.LastCollision.B)},
			UiLabel{Text: fmt.Sprintf("Point: [%.2f %.2f %.2f]", state.LastCollision.Point.X(), state.LastCollision.Point.Y(), state.LastCollision.Point.Z()), Scale: 0.78, Dim: true},
			UiLabel{Text: fmt.Sprintf("Normal: [%.2f %.2f %.2f]", state.LastCollision.Normal.X(), state.LastCollision.Normal.Y(), state.LastCollision.Normal.Z()), Scale: 0.78, Dim: true},
			UiLabel{Text: fmt.Sprintf("Penetration: %.3f", state.LastCollision.Penetration), Scale: 0.78, Dim: true},
			UiLabel{Text: fmt.Sprintf("Impulse: %.2f", state.LastCollision.NormalImpulse), Scale: 0.78, Dim: true},
			UiLabel{Text: fmt.Sprintf("Relative speed: %.2f", state.LastCollision.RelativeSpeed), Scale: 0.78, Dim: true},
		)
	} else {
		children = append(children, UiLabel{Text: "Waiting for contacts...", Dim: true})
	}

	children = append(children,
		UiSpacer{Height: 6},
		UiLabel{Text: "F: launch bowling ball", Scale: 0.76, Dim: true},
		UiLabel{Text: "LMB: push highlighted body", Scale: 0.76, Dim: true},
		UiLabel{Text: "Space/Ctrl: fly up/down, Tab: capture mouse", Scale: 0.76, Dim: true},
		UiLabel{Text: "Crosshair uses ScreenToWorldRay + Raycast", Scale: 0.76, Dim: true},
	)

	return &UiPanel{
		Key:      "collision_events_hud",
		Anchor:   UiAnchorTopLeft,
		Position: [2]float32{12, 12},
		Width:    320,
		Title:    "Collision Events",
		Visible:  true,
		Children: children,
	}
}

func buildEventLogPanel(state *DemoState) *UiPanel {
	nodes := []UiNode{
		UiLabel{Text: "Recent events are appended from DrainCollisionEvents().", Scale: 0.76, Dim: true},
		UiSpacer{Height: 6},
	}

	if len(state.RecentEvents) == 0 {
		if state.HasLastCollision {
			nodes = append(nodes,
				UiLabel{Text: fmt.Sprintf("Latest: %s", state.LastCollision.Type.String()), Dim: true},
				UiLabel{Text: fmt.Sprintf("%s <-> %s", entityLabelFromState(state, state.LastCollision.A), entityLabelFromState(state, state.LastCollision.B)), Scale: 0.78, Dim: true},
				UiLabel{Text: fmt.Sprintf("imp=%.2f rel=%.2f", state.LastCollision.NormalImpulse, state.LastCollision.RelativeSpeed), Scale: 0.78, Dim: true},
			)
		} else {
			nodes = append(nodes, UiLabel{Text: "No collision events yet.", Dim: true})
		}
	} else {
		for _, line := range state.RecentEvents {
			nodes = append(nodes, UiLabel{Text: line, Scale: 0.78})
		}
	}

	return &UiPanel{
		Key:       "collision_events_log",
		Anchor:    UiAnchorTopRight,
		Position:  [2]float32{12, 12},
		Width:     420,
		MaxHeight: 280,
		Title:     "Event Log",
		Visible:   true,
		Children: []UiNode{
			UiColumn{
				Spacing:  4,
				Children: nodes,
			},
		},
	}
}

func triggerCollisionFlash(cmd *Commands, eid EntityId, duration float32) {
	if eid == 0 {
		return
	}

	for _, comp := range cmd.GetAllComponents(eid) {
		if visual, ok := comp.(*VisualFeedbackComponent); ok && visual != nil {
			visual.CollisionFlash = maxf(visual.CollisionFlash, duration)
			return
		}
	}
}

func triggerSensorFlash(cmd *Commands, eid EntityId) {
	if eid == 0 {
		return
	}

	isSensor := false
	for _, comp := range cmd.GetAllComponents(eid) {
		if _, ok := comp.(*SensorComponent); ok {
			isSensor = true
			break
		}
	}
	if !isSensor {
		return
	}

	for _, comp := range cmd.GetAllComponents(eid) {
		if visual, ok := comp.(*VisualFeedbackComponent); ok && visual != nil {
			visual.SensorFlash = maxf(visual.SensorFlash, sensorFlashSeconds)
			return
		}
	}
}

func spawnSparkBurst(cmd *Commands, state *DemoState, point, normal mgl32.Vec3, impulse float32) {
	if state == nil || state.ParticleAtlas == (AssetId{}) {
		return
	}

	offset := mgl32.Vec3{0, 0.05, 0}
	if normal.Len() > 0 {
		offset = normal.Normalize().Mul(0.08)
	}
	spawnPos := point.Add(offset)
	rate := clampf(900+impulse*240, 900, 2600)
	speedMax := clampf(3.5+impulse*0.75, 4.0, 8.5)

	cmd.AddEntity(
		&TransformComponent{Position: spawnPos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&ParticleEmitterComponent{
			Enabled:          true,
			MaxParticles:     2048,
			SpawnRate:        rate,
			LifetimeRange:    [2]float32{0.18, 0.35},
			StartSpeedRange:  [2]float32{1.8, speedMax},
			StartSizeRange:   [2]float32{0.08, 0.18},
			StartColorMin:    [4]float32{1.0, 0.72, 0.25, 0.8},
			StartColorMax:    [4]float32{1.0, 1.0, 0.75, 1.0},
			Gravity:          2.0,
			Drag:             2.2,
			ConeAngleDegrees: 180,
			SpriteIndex:      5,
			AtlasCols:        4,
			AtlasRows:        4,
			Texture:          state.ParticleAtlas,
			AlphaMode:        SpriteAlphaLuminance,
		},
		&LifetimeComponent{TimeLeft: 0.12},
	)
	cmd.AddEntity(
		&SpriteComponent{
			Enabled:       true,
			Position:      spawnPos,
			Size:          [2]float32{0.65, 0.65},
			Color:         [4]float32{1.0, 0.8, 0.2, 0.95},
			SpriteIndex:   4,
			AtlasCols:     4,
			AtlasRows:     4,
			Texture:       state.ParticleAtlas,
			BillboardMode: BillboardSpherical,
			Unlit:         true,
			AlphaMode:     SpriteAlphaLuminance,
		},
		&LifetimeComponent{TimeLeft: 0.08},
	)
}

func centerAimRay(cmd *Commands, vox *VoxelRtState, input *Input) (mgl32.Vec3, mgl32.Vec3, bool) {
	cam := primaryCamera(cmd)
	if cam == nil || vox == nil || input == nil {
		return mgl32.Vec3{}, mgl32.Vec3{}, false
	}

	mouseX := float64(input.WindowWidth) * 0.5
	mouseY := float64(input.WindowHeight) * 0.5
	origin, dir := vox.ScreenToWorldRay(mouseX, mouseY, cam)
	if dir.Len() == 0 {
		return mgl32.Vec3{}, mgl32.Vec3{}, false
	}
	return origin, dir.Normalize(), true
}

func aimRay(cmd *Commands, vox *VoxelRtState, input *Input) (mgl32.Vec3, mgl32.Vec3, bool) {
	cam := primaryCamera(cmd)
	if cam == nil || vox == nil || input == nil {
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

func spawnSceneEntity(cmd *Commands, components ...any) EntityId {
	components = append(components, &DemoSceneEntity{})
	return cmd.AddEntity(components...)
}

func primaryCamera(cmd *Commands) *CameraComponent {
	var cam *CameraComponent
	MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, camera *CameraComponent) bool {
		cam = camera
		return false
	})
	return cam
}

func entityLabel(cmd *Commands, eid EntityId) string {
	if eid == 0 {
		return "none"
	}

	for _, comp := range cmd.GetAllComponents(eid) {
		if tag, ok := comp.(*DemoTagComponent); ok && tag != nil && tag.Label != "" {
			return tag.Label
		}
	}

	return fmt.Sprintf("entity-%d", eid)
}

func entityHasComponent[T any](cmd *Commands, eid EntityId) bool {
	for _, comp := range cmd.GetAllComponents(eid) {
		if _, ok := comp.(*T); ok {
			return true
		}
		if _, ok := comp.(T); ok {
			return true
		}
	}
	return false
}

func isBowlingBall(cmd *Commands, eid EntityId) bool {
	return entityHasComponent[BowlingBallComponent](cmd, eid)
}

func applyImpulseToEntity(cmd *Commands, target EntityId, impulse mgl32.Vec3) bool {
	applied := false
	MakeQuery1[RigidBodyComponent](cmd).Map(func(eid EntityId, rb *RigidBodyComponent) bool {
		if eid != target {
			return true
		}
		if rb != nil && !rb.IsStatic {
			rb.ApplyImpulse(impulse)
			applied = true
		}
		return false
	})
	return applied
}

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func entityLabelFromState(state *DemoState, eid EntityId) string {
	if state == nil || eid == 0 {
		return "none"
	}
	if state.HasLastCollision {
		// The demo keeps human-readable labels only in the recent event strings,
		// so fall back to entity IDs in the panel summary.
		return fmt.Sprintf("entity-%d", eid)
	}
	return fmt.Sprintf("entity-%d", eid)
}

func quitSystem(cmd *Commands, input *Input) {
	if input != nil && input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
