package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const (
	exampleWidth  = 1600
	exampleHeight = 900
	exampleFar    = float32(4200)
	exampleDepth  = VoxelRtDepthModeReverseZ
)

type DemoModule struct{}

func (m DemoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(System(setupSystem).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(hudSystem).InStage(PostUpdate).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
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
			WindowWidth:  exampleWidth,
			WindowHeight: exampleHeight,
			WindowTitle:  "Far Range Depth Harness",
			DepthMode:    exampleDepth,
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

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 42, 120}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position:  mgl32.Vec3{0, 42, 120},
			LookAt:    mgl32.Vec3{0, 40, -600},
			Up:        mgl32.Vec3{0, 1, 0},
			Fov:       48,
			Aspect:    float32(exampleWidth) / float32(exampleHeight),
			Near:      0.1,
			Far:       exampleFar,
			DepthMode: exampleDepth,
		},
		&FlyingCameraComponent{Speed: 40.0, Sensitivity: 0.1},
	)

	cmd.AddEntity(&LightComponent{
		Type:      LightTypeAmbient,
		Intensity: 0.1,
		Color:     [3]float32{0.85, 0.9, 1.0},
	})
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{60, 400, 60},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-42), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(30), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:      LightTypeDirectional,
			Intensity: 1.0,
			Color:     [3]float32{1.0, 0.96, 0.88},
			Range:     5000,
		},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerStars,
		Seed:       42,
		Scale:      1.0,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.7, 0.82, 1.0},
		ColorB:     mgl32.Vec3{1.0, 1.0, 1.0},
		Threshold:  0.985,
		Opacity:    1.0,
		Priority:   0,
		BlendMode:  SkyboxBlendAdd,
	})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.05, 0.08, 0.14},
		ColorB:     mgl32.Vec3{0.0, 0.0, 0.02},
		Opacity:    1.0,
		Smooth:     true,
		Priority:   1,
		BlendMode:  SkyboxBlendAlpha,
	})

	groundPalette := assets.CreateSimplePalette([4]uint8{34, 38, 48, 255})
	nearPalette := assets.CreateSimplePalette([4]uint8{244, 230, 180, 255})
	warmPalette := assets.CreateSimplePalette([4]uint8{255, 172, 96, 255})
	coolPalette := assets.CreateSimplePalette([4]uint8{110, 198, 255, 255})
	planetPalette := assets.CreateSimplePalette([4]uint8{128, 150, 255, 255})

	groundModel := assets.CreateCubeModel(600, 1, 4200, 1.0)
	anchorModel := assets.CreateCubeModel(24, 24, 24, 1.0)
	towerModel := assets.CreateCubeModel(28, 240, 28, 1.0)
	slabModel := assets.CreateCubeModel(260, 12, 120, 1.0)
	planetModel := assets.CreateSphereModel(72, 1.0)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-300, -1, -4000}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: groundModel, VoxelPalette: groundPalette, PivotMode: PivotModeCorner, AmbientOcclusionMode: VoxelAODisabled},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 12, -180}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: anchorModel, VoxelPalette: nearPalette, AmbientOcclusionMode: VoxelAODisabled},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{90, 120, -1500}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: towerModel, VoxelPalette: warmPalette, AmbientOcclusionMode: VoxelAODisabled},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-280, 140, -2700}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: planetModel, VoxelPalette: planetPalette, AmbientOcclusionMode: VoxelAODisabled},
	)

	precisionRotationA := mgl32.QuatRotate(mgl32.DegToRad(8), mgl32.Vec3{0, 1, 0})
	precisionRotationB := mgl32.QuatRotate(mgl32.DegToRad(-6), mgl32.Vec3{0, 1, 0})
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{40, 40, -3380}, Rotation: precisionRotationA, Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: slabModel, VoxelPalette: warmPalette, AmbientOcclusionMode: VoxelAODisabled},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{54, 34, -3388}, Rotation: precisionRotationB, Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: slabModel, VoxelPalette: coolPalette, AmbientOcclusionMode: VoxelAODisabled},
	)

	setupDone = true
}

func hudSystem(state *VoxelRtState) {
	if state == nil || state.RtApp == nil || state.RtApp.Camera == nil {
		return
	}
	text := fmt.Sprintf(
		"Depth mode: %s  Near/Far: %.1f / %.0f\nTargets: near anchor 180u, tower 1500u, planet 2700u, precision slabs 3380u",
		state.RtApp.Camera.DepthMode,
		state.RtApp.Camera.Near,
		state.RtApp.Camera.Far,
	)
	state.DrawText(text, 20, 20, 0.8, [4]float32{1.0, 0.96, 0.84, 1.0})
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
