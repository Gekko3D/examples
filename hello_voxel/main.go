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

func (m DemoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(setupSystem).
			InStage(Prelude).
			InState(OnEnter(Startup)),
	)
	app.UseSystem(
		System(quitSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
}

var setupDone bool

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
			WindowTitle:  "Hello Voxel",
		},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func setupSystem(cmd *Commands, assets *AssetServer) {
	if setupDone {
		return
	}

	// 1. Camera
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 5, 12},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 5, 12},
			LookAt:   mgl32.Vec3{0, 2, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,   // Facing towards negative Z (the origin)
			Pitch:    -20, // Looking slightly down
			Fov:      52,
			Aspect:   1280.0 / 720.0,
			Near:     0.1,
			Far:      500,
		},
		&FlyingCameraComponent{Speed: 8.0, Sensitivity: 0.1},
	)

	// 2. Ambient Light & Skybox
	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.25,
			Color:     [3]float32{1, 1, 1},
		},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.2, 0.5, 0.8}, // Horizon
		ColorB:     mgl32.Vec3{0.1, 0.3, 0.7}, // Zenith
		Opacity:    1.0,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})

	// 3. Floor
	greyPalette := assets.CreateSimplePalette([4]uint8{80, 82, 88, 255})
	floorModel := assets.CreateCubeModel(40, 1, 40, 1.0)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-20, -1, -20}, // Center at origin, Y offset so top is at 0
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      floorModel,
			VoxelPalette:    greyPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: 1.0,
		},
	)

	// 4. Cube
	bluePalette := assets.CreateSimplePalette([4]uint8{60, 120, 255, 255})
	cubeModel := assets.CreateCubeModel(8, 8, 8, 1.0)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 4, 0}, // Offset by height/2 (8/2 = 4)
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      cubeModel,
			VoxelPalette:    bluePalette,
			VoxelResolution: 1.0,
		},
	)

	// 5. Sphere
	redPalette := assets.CreateSimplePalette([4]uint8{255, 80, 80, 255})
	sphereModel := assets.CreateSphereModel(6, 1.0)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{10, 6, -3}, // Offset so it's on/near floor and not inside cube
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      sphereModel,
			VoxelPalette:    redPalette,
			VoxelResolution: 1.0,
		},
	)

	// 5. Directional Light
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{10, 50, 10},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-45), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(30), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:      LightTypeDirectional,
			Intensity: 0.8,
			Color:     [3]float32{1, 0.95, 0.9},
			Range:     500,
		},
	)

	setupDone = true
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
