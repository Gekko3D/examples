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

type DemoModule struct{}

type DemoState struct {
	EmitterEntities [4]EntityId
	EmitterNames    [4]string
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
			WindowTitle:  "Particles and Sprites",
		},
		FlyingCameraModule{},
		LifecycleModule{},
		DemoModule{},
	)
	app.Run()
}

func (m DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{
		EmitterNames: [4]string{
			"fire fountain",
			"snow drift",
			"jet stream",
			"smoke puff",
		},
	})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(toggleEmittersSystem).InStage(Update).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func resolveDemoAsset(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}

	alt := "examples/particles_and_sprites/" + path
	if _, err := os.Stat(alt); err == nil {
		return alt
	}

	return path
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	atlasID := assets.CreateTexture(resolveDemoAsset("assets/particle_atlas.png"))

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 8, 24},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 8, 24},
			LookAt:   mgl32.Vec3{0, 4, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -9.5,
			Fov:      58,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      600,
		},
		&FlyingCameraComponent{Speed: 12.0, Sensitivity: 0.1},
	)

	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.16,
			Color:     [3]float32{0.9, 0.93, 1.0},
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{10, 30, 10},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-48), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(28), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    1.0,
			Color:        [3]float32{1.0, 0.96, 0.9},
			Range:        800,
			CastsShadows: true,
		},
	)

	floorPalette := assets.CreateSimplePalette([4]uint8{78, 84, 92, 255})
	plinthPalette := assets.CreateSimplePalette([4]uint8{112, 120, 132, 255})
	floorModel := assets.CreateCubeModel(80, 2, 80, 1.0)
	plinthModel := assets.CreateCubeModel(6, 2, 6, 1.0)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-40, -1, -40},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      floorModel,
			VoxelPalette:    floorPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: 1.0,
		},
	)

	plinthPositions := []mgl32.Vec3{
		{-12, 0, 0},
		{0, 8, 0},
		{0, 2, 10},
		{12, 0, 0},
	}
	for _, pos := range plinthPositions {
		cmd.AddEntity(
			&TransformComponent{
				Position: pos.Sub(mgl32.Vec3{3, 0, 3}),
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{
				VoxelModel:      plinthModel,
				VoxelPalette:    plinthPalette,
				PivotMode:       PivotModeCorner,
				VoxelResolution: 1.0,
			},
		)
	}

	state.EmitterEntities[0] = addEmitter(cmd, mgl32.Vec3{-12, 2.35, 0}, mgl32.QuatIdent(), &ParticleEmitterComponent{
		Enabled:          true,
		MaxParticles:     8192,
		SpawnRate:        400,
		LifetimeRange:    [2]float32{1.0, 2.5},
		StartSpeedRange:  [2]float32{8, 14},
		StartSizeRange:   [2]float32{0.1, 0.25},
		StartColorMin:    [4]float32{1.0, 0.5, 0.0, 1.0},
		StartColorMax:    [4]float32{1.0, 0.2, 0.0, 0.8},
		Gravity:          9.8,
		Drag:             0.1,
		ConeAngleDegrees: 15,
		SpriteIndex:      4,
		AtlasCols:        4,
		AtlasRows:        4,
		Texture:          atlasID,
		AlphaMode:        SpriteAlphaLuminance,
	})

	state.EmitterEntities[1] = addEmitter(cmd, mgl32.Vec3{0, 10.35, 0}, mgl32.QuatIdent(), &ParticleEmitterComponent{
		Enabled:          true,
		MaxParticles:     8192,
		SpawnRate:        260,
		LifetimeRange:    [2]float32{3.5, 6.5},
		StartSpeedRange:  [2]float32{0.5, 2.0},
		StartSizeRange:   [2]float32{0.12, 0.28},
		StartColorMin:    [4]float32{0.8, 0.88, 1.0, 0.8},
		StartColorMax:    [4]float32{1.0, 1.0, 1.0, 1.0},
		Gravity:          1.5,
		Drag:             0.15,
		ConeAngleDegrees: 180,
		SpriteIndex:      1,
		AtlasCols:        4,
		AtlasRows:        4,
		Texture:          atlasID,
		AlphaMode:        SpriteAlphaTexture,
	})

	jetRotation := mgl32.QuatRotate(mgl32.DegToRad(-90), mgl32.Vec3{0, 0, 1})
	state.EmitterEntities[2] = addEmitter(cmd, mgl32.Vec3{0, 4.35, 10}, jetRotation, &ParticleEmitterComponent{
		Enabled:          true,
		MaxParticles:     8192,
		SpawnRate:        320,
		LifetimeRange:    [2]float32{0.45, 0.9},
		StartSpeedRange:  [2]float32{20, 30},
		StartSizeRange:   [2]float32{0.08, 0.18},
		StartColorMin:    [4]float32{0.7, 0.95, 1.0, 0.75},
		StartColorMax:    [4]float32{1.0, 1.0, 1.0, 1.0},
		Gravity:          0,
		Drag:             0.02,
		ConeAngleDegrees: 3,
		SpriteIndex:      3,
		AtlasCols:        4,
		AtlasRows:        4,
		Texture:          atlasID,
		AlphaMode:        SpriteAlphaTexture,
	})

	state.EmitterEntities[3] = addEmitter(cmd, mgl32.Vec3{12, 2.35, 0}, mgl32.QuatIdent(), &ParticleEmitterComponent{
		Enabled:          true,
		MaxParticles:     8192,
		SpawnRate:        180,
		LifetimeRange:    [2]float32{2.5, 4.0},
		StartSpeedRange:  [2]float32{1.0, 4.0},
		StartSizeRange:   [2]float32{0.3, 0.6},
		StartColorMin:    [4]float32{0.2, 0.2, 0.2, 0.35},
		StartColorMax:    [4]float32{0.55, 0.55, 0.55, 0.75},
		Gravity:          -1,
		Drag:             3.0,
		ConeAngleDegrees: 40,
		SpriteIndex:      6,
		AtlasCols:        4,
		AtlasRows:        4,
		Texture:          atlasID,
		AlphaMode:        SpriteAlphaLuminance,
	})

	addWorldSprites(cmd, atlasID)
	addUISprites(cmd, atlasID)

	cmd.AddEntity(
		&SpriteComponent{
			Enabled:       true,
			Position:      mgl32.Vec3{-12, 5.5, 0},
			Size:          [2]float32{1.2, 1.2},
			Color:         [4]float32{1.0, 0.9, 0.7, 0.85},
			SpriteIndex:   7,
			AtlasCols:     4,
			AtlasRows:     4,
			Texture:       atlasID,
			BillboardMode: BillboardSpherical,
			AlphaMode:     SpriteAlphaLuminance,
			Unlit:         true,
		},
		&LifetimeComponent{TimeLeft: 12},
	)

	fmt.Println("Particles and sprites demo ready")
	fmt.Println("Controls: WASD + mouse to fly, ESC to quit, 1-4 toggle fire/snow/jet/smoke emitters")
}

func addEmitter(cmd *Commands, position mgl32.Vec3, rotation mgl32.Quat, emitter *ParticleEmitterComponent) EntityId {
	return cmd.AddEntity(
		&TransformComponent{
			Position: position,
			Rotation: rotation,
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		emitter,
	)
}

func addWorldSprites(cmd *Commands, atlasID AssetId) {
	sphericalMarkers := []mgl32.Vec3{
		{-18, 1.5, -8},
		{-10, 1.5, -12},
		{8, 1.5, -11},
		{16, 1.5, -7},
	}
	sphericalColors := [][4]float32{
		{1.0, 0.35, 0.35, 1.0},
		{0.35, 1.0, 0.45, 1.0},
		{0.35, 0.7, 1.0, 1.0},
		{1.0, 0.85, 0.3, 1.0},
	}
	for i, pos := range sphericalMarkers {
		cmd.AddEntity(&SpriteComponent{
			Enabled:       true,
			Position:      pos,
			Size:          [2]float32{1.5, 1.5},
			Color:         sphericalColors[i],
			BillboardMode: BillboardSpherical,
			Texture:       atlasID,
			AtlasCols:     4,
			AtlasRows:     4,
			SpriteIndex:   uint32(i % 4),
			AlphaMode:     SpriteAlphaTexture,
		})
	}

	cylindricalMarkers := []mgl32.Vec3{
		{-20, 1.2, 8},
		{-14, 1.2, 12},
		{-8, 1.2, 9},
		{8, 1.2, 11},
		{14, 1.2, 8},
		{20, 1.2, 12},
	}
	cylindricalColors := [][4]float32{
		{0.8, 1.0, 0.55, 1.0},
		{0.55, 1.0, 0.85, 1.0},
		{1.0, 0.65, 0.95, 1.0},
		{0.65, 0.85, 1.0, 1.0},
		{1.0, 0.78, 0.55, 1.0},
		{0.9, 0.9, 1.0, 1.0},
	}
	for i, pos := range cylindricalMarkers {
		cmd.AddEntity(&SpriteComponent{
			Enabled:       true,
			Position:      pos,
			Size:          [2]float32{2.0, 2.4},
			Color:         cylindricalColors[i],
			BillboardMode: BillboardCylindrical,
			Texture:       atlasID,
			AtlasCols:     4,
			AtlasRows:     4,
			SpriteIndex:   uint32((i % 2) + 8),
			AlphaMode:     SpriteAlphaTexture,
		})
	}
}

func addUISprites(cmd *Commands, atlasID AssetId) {
	cmd.AddEntity(&SpriteComponent{
		Enabled:     true,
		IsUI:        true,
		Position:    mgl32.Vec3{50, 50, 0},
		Size:        [2]float32{64, 64},
		Color:       [4]float32{1, 1, 1, 1},
		SpriteIndex: 2,
		AtlasCols:   4,
		AtlasRows:   4,
		Texture:     atlasID,
		AlphaMode:   SpriteAlphaTexture,
		Unlit:       true,
	})
}

func toggleEmittersSystem(cmd *Commands, input *Input, state *DemoState) {
	keys := [4]int{Key1, Key2, Key3, Key4}
	for i, key := range keys {
		if !input.JustPressed[key] {
			continue
		}
		toggleEmitter(cmd, state.EmitterEntities[i], state.EmitterNames[i])
	}
}

func toggleEmitter(cmd *Commands, target EntityId, name string) {
	MakeQuery1[ParticleEmitterComponent](cmd).Map(func(eid EntityId, emitter *ParticleEmitterComponent) bool {
		if eid != target {
			return true
		}

		emitter.Enabled = !emitter.Enabled
		if emitter.Enabled {
			fmt.Printf("Emitter %s enabled\n", name)
		} else {
			fmt.Printf("Emitter %s disabled\n", name)
		}
		return false
	})
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
