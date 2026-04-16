package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const demoVoxelResolution float32 = 0.25

const (
	Startup State = iota
	Quit
)

type DemoModule struct{}

type DemoState struct {
	LayerEntities  [4]EntityId
	LayerOpacities [4]float32
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
			WindowTitle:  "Skybox Showcase",
		},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func (m DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(toggleSkyboxLayersSystem).InStage(Update).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	groundPalette := assets.CreateSimplePalette([4]uint8{58, 66, 74, 255})
	columnPalette := assets.CreateSimplePalette([4]uint8{170, 182, 204, 255})
	warmPalette := assets.CreatePBRPalette([4]uint8{220, 180, 120, 255}, 0.35, 0.0, 0.0, 1.0)
	coolPalette := assets.CreatePBRPalette([4]uint8{120, 170, 240, 255}, 0.2, 0.05, 0.0, 1.0)

	groundModel := assets.CreateCubeModel(160, 4, 160, 1.0)
	columnModel := assets.CreateCubeModel(10, 40, 10, 1.0)
	plinthModel := assets.CreateCubeModel(20, 4, 20, 1.0)
	sphereModel := assets.CreateSphereModel(10, 1.0)
	cubeModel := assets.CreateCubeModel(14, 14, 14, 1.0)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{26, 14, 30},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{26, 14, 30},
			LookAt:   mgl32.Vec3{0, 6, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      -135,
			Pitch:    -18,
			Fov:      60,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      1000,
		},
		&FlyingCameraComponent{Speed: 14.0, Sensitivity: 0.1},
	)

	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.08,
			Color:     [3]float32{0.8, 0.86, 1.0},
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{6, 9, 6},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-40), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(35), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    1.3,
			Color:        [3]float32{1.0, 0.95, 0.88},
			Range:        800,
			CastsShadows: true,
		},
	)
	cmd.AddEntity(&SkyAmbientComponent{SkyMix: 0.3})

	state.LayerEntities[0] = cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.15, 0.3, 0.6},
		ColorB:     mgl32.Vec3{0.02, 0.05, 0.15},
		Opacity:    1.0,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})
	state.LayerOpacities[0] = 1.0

	state.LayerEntities[1] = cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerStars,
		Resolution: [2]int{2048, 1024},
		Seed:       12345,
		Scale:      20.0,
		ColorA:     mgl32.Vec3{1, 1, 1},
		Threshold:  0.92,
		Opacity:    0.6,
		Priority:   1,
		BlendMode:  SkyboxBlendAdd,
	})
	state.LayerOpacities[1] = 0.6

	state.LayerEntities[2] = cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:   SkyboxLayerNebula,
		NoiseType:   SkyboxNoiseSimplex,
		Seed:        77,
		Scale:       3.0,
		Octaves:     3,
		Persistence: 0.5,
		Lacunarity:  2.0,
		Resolution:  [2]int{1024, 512},
		ColorA:      mgl32.Vec3{0.6, 0.1, 0.8},
		ColorB:      mgl32.Vec3{0.1, 0.2, 0.9},
		Threshold:   0.4,
		Opacity:     0.35,
		Priority:    2,
		BlendMode:   SkyboxBlendAdd,
	})
	state.LayerOpacities[2] = 0.35

	state.LayerEntities[3] = cmd.AddEntity(&SkyboxLayerComponent{
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
	})
	state.LayerOpacities[3] = 0.7

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

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-20, -1, -20},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      groundModel,
			VoxelPalette:    groundPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: demoVoxelResolution,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-8, 0.5, -8},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      plinthModel,
			VoxelPalette:    groundPalette,
			VoxelResolution: demoVoxelResolution,
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-6, 5.0, -6},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      sphereModel,
			VoxelPalette:    warmPalette,
			VoxelResolution: demoVoxelResolution,
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{10, 0.5, 6},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(18), mgl32.Vec3{0, 1, 0}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      columnModel,
			VoxelPalette:    columnPalette,
			VoxelResolution: demoVoxelResolution,
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{4, 2.0, -2},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(32), mgl32.Vec3{0, 1, 0}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      cubeModel,
			VoxelPalette:    coolPalette,
			VoxelResolution: demoVoxelResolution,
		},
	)

	fmt.Println("Skybox showcase ready")
	fmt.Println("Controls: WASD + mouse to fly, ESC to quit, 1-4 toggle gradient/stars/nebula/clouds")
}

func toggleSkyboxLayersSystem(cmd *Commands, input *Input, state *DemoState) {
	keys := [4]int{Key1, Key2, Key3, Key4}
	names := [4]string{"gradient", "stars", "nebula", "clouds"}

	for i, key := range keys {
		if !input.JustPressed[key] {
			continue
		}
		toggleSkyboxLayer(cmd, state.LayerEntities[i], state.LayerOpacities[i], names[i])
	}
}

func toggleSkyboxLayer(cmd *Commands, target EntityId, visibleOpacity float32, name string) {
	MakeQuery1[SkyboxLayerComponent](cmd).Map(func(eid EntityId, layer *SkyboxLayerComponent) bool {
		if eid != target {
			return true
		}

		if layer.Opacity > 0 {
			layer.Opacity = 0
			fmt.Printf("Skybox layer %s hidden\n", name)
		} else {
			layer.Opacity = visibleOpacity
			fmt.Printf("Skybox layer %s visible\n", name)
		}
		layer.SetDirty()
		return false
	})
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
