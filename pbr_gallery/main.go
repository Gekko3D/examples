package main

import (
	"math"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

// 2. SpinnerComponent
type SpinnerComponent struct {
	AngularSpeed mgl32.Vec3 // radians/sec per axis
}

// 3. SpinnerModule
type SpinnerModule struct{}

func (m SpinnerModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(spinnerSystem).
			InStage(Update).
			RunAlways(),
	)
}

func spinnerSystem(cmd *Commands, time *Time) {
	dt := float32(time.Dt)
	MakeQuery2[TransformComponent, SpinnerComponent](cmd).Map(func(eid EntityId, tr *TransformComponent, spin *SpinnerComponent) bool {
		// Apply rotation around each axis based on angular speed
		rotationX := mgl32.QuatRotate(spin.AngularSpeed.X()*dt, mgl32.Vec3{1, 0, 0})
		rotationY := mgl32.QuatRotate(spin.AngularSpeed.Y()*dt, mgl32.Vec3{0, 1, 0})
		rotationZ := mgl32.QuatRotate(spin.AngularSpeed.Z()*dt, mgl32.Vec3{0, 0, 1})

		delta := rotationZ.Mul(rotationY).Mul(rotationX)
		tr.Rotation = tr.Rotation.Mul(delta).Normalize()
		return true
	})
}

// 4. OrbiterComponent
type OrbiterComponent struct {
	Radius float32
	Speed  float32
	Center mgl32.Vec3
}

// 5. OrbiterModule
type OrbiterModule struct{}

func (m OrbiterModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(orbiterSystem).
			InStage(Update).
			RunAlways(),
	)
}

func orbiterSystem(cmd *Commands, time *Time) {
	t := float32(time.Elapsed)
	MakeQuery2[TransformComponent, OrbiterComponent](cmd).Map(func(eid EntityId, tr *TransformComponent, orb *OrbiterComponent) bool {
		tr.Position[0] = orb.Center[0] + float32(math.Cos(float64(t*orb.Speed)))*orb.Radius
		tr.Position[2] = orb.Center[2] + float32(math.Sin(float64(t*orb.Speed)))*orb.Radius
		return true
	})
}

type DemoModule struct{}

func (m DemoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(setupScene).
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
			WindowWidth:     1600,
			WindowHeight:    900,
			WindowTitle:     "PBR Gallery & Procedural Primitives",
			DebugMode:       true,
		},
		FlyingCameraModule{},
		SpinnerModule{},
		OrbiterModule{},
		DemoModule{},
	)
	app.Run()
}

func setupScene(cmd *Commands, assets *AssetServer) {
	if setupDone {
		return
	}

	// 1. Camera
	// Position giving a good overview of the gallery
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 80, 200},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 80, 200},
			LookAt:   mgl32.Vec3{0, 40, 40},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -24,
			Fov:      65,
			Aspect:   1600.0 / 900.0,
			Near:     1.0,
			Far:      2000,
		},
		&FlyingCameraComponent{Speed: 40.0, Sensitivity: 0.1},
	)

	// 2. Ambient Light & Skybox
	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.1,
			Color:     [3]float32{1, 1, 1},
		},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.05, 0.1, 0.25}, // Deeper blue
		ColorB:     mgl32.Vec3{0.0, 0.0, 0.05},  // Near black zenith
		Opacity:    1.0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})

	// 3. Directional Light (Sun)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{30, 80, 40},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-50), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(35), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    1.0,
			Color:        [3]float32{1, 0.98, 0.85},
			Range:        500,
			CastsShadows: true,
		},
	)

	// 4. Floor
	greyPalette := assets.CreateSimplePalette([4]uint8{50, 52, 58, 255})
	floorModel := assets.CreateCubeModel(400, 1, 400, 1.0)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-200, -1, -50},
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

	// 5. Sphere Gallery
	sphereModel := assets.CreateSphereModel(16.0, 1.0)
	spacing := float32(24.0)
	centerX := float32(0.0)
	startX := centerX - (spacing * 2.5)

	for i := 0; i < 6; i++ {
		t := float32(i) / 5.0
		xPos := startX + float32(i)*spacing

		// Row 1 (Y=12): Gold metallic, roughness varies 0.0->1.0
		goldPalette := assets.CreatePBRPalette([4]uint8{255, 210, 100, 255}, t, 1.0, 0.0, 1.5)
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{xPos, 70, 0},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: goldPalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)

		// Row 2 (Y=9): Blue dielectric, roughness varies 0.0->1.0
		bluePalette := assets.CreatePBRPalette([4]uint8{30, 120, 255, 255}, t, 0.0, 0.0, 1.5)
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{xPos, 50, 0},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: bluePalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)

		// Row 3 (Y=6): Green emissive, emission varies 0.0->4.0
		emission := t * 4.0
		greenPalette := assets.CreatePBRPalette([4]uint8{100, 255, 100, 255}, 0.5, 0.0, emission, 1.0)
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{xPos, 30, 0},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: greenPalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)

		// Row 4 (Y=3): Glass transparent, alpha varies 255->25
		alpha := uint8(255 - (t * 225))
		glassPalette := assets.CreatePBRPalette([4]uint8{180, 220, 255, alpha}, 0.1, 0.0, 0.0, 1.5)
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{xPos, 10, 0},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: glassPalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)
	}

	// 6. Procedural Primitives Showcase
	shapes := []struct {
		name    string
		model   AssetId
		palette AssetId
	}{
		{"Cube", assets.CreateCubeModel(20, 20, 20, 1.0), assets.CreateSimplePalette([4]uint8{255, 80, 80, 255})},
		{"Sphere", assets.CreateSphereModel(10, 1.0), assets.CreateSimplePalette([4]uint8{80, 255, 80, 255})},
		{"Cone", assets.CreateConeModel(10, 20, 1.0), assets.CreateSimplePalette([4]uint8{80, 80, 255, 255})},
		{"Pyramid", assets.CreatePyramidModel(20, 20, 1.0), assets.CreateSimplePalette([4]uint8{255, 255, 80, 255})},
		{"Cylinder", assets.CreateCylinderModel(10, 20, 1.0), assets.CreateSimplePalette([4]uint8{255, 80, 255, 255})},
		{"Capsule", assets.CreateCapsuleModel(10, 30, 1.0), assets.CreateSimplePalette([4]uint8{80, 255, 255, 255})},
		{"Ramp", assets.CreateRampModel(20, 15, 20, 1.0), assets.CreateSimplePalette([4]uint8{255, 128, 0, 255})},
	}

	shapeStartX := centerX - (spacing * 3)
	for i, shape := range shapes {
		xPos := shapeStartX + float32(i)*spacing
		zPos := float32(80.0)
		// Separate platform area
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{xPos - 12, 0, zPos - 12},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{
				VoxelModel:      assets.CreateCubeModel(24, 0.5, 24, 1.0),
				VoxelPalette:    assets.CreateSimplePalette([4]uint8{80, 80, 85, 255}),
				PivotMode:       PivotModeCorner,
				VoxelResolution: 1.0,
			},
		)

		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{xPos, 12, zPos},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{
				VoxelModel:      shape.model,
				VoxelPalette:    shape.palette,
				VoxelResolution: 1.0,
			},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 1.0, 0.3}},
		)
	}

	// 7. Lighting
	// 2-3 point lights near the gallery at different warm/cool colors
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-60, 30, 10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 40.0, Color: [3]float32{1.0, 0.5, 0.2}, Range: 200},
		&OrbiterComponent{Radius: 100, Speed: 0.5, Center: mgl32.Vec3{0, 50, 40}},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{60, 30, 10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 40.0, Color: [3]float32{0.2, 0.5, 1.0}, Range: 200},
		&OrbiterComponent{Radius: 100, Speed: -0.4, Center: mgl32.Vec3{0, 50, 40}},
	)

	// Additional shadowed point lights for drama
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-40, 50, 40}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{
			Type:         LightTypePoint,
			Intensity:    50.0,
			Color:        [3]float32{1, 1, 0.9},
			Range:        200,
			CastsShadows: true,
		},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{40, 50, 40}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{
			Type:         LightTypePoint,
			Intensity:    50.0,
			Color:        [3]float32{0.9, 1, 1},
			Range:        200,
			CastsShadows: true,
		},
	)

	// 1 spot light aimed at the shape showcase area
	// From (0, 30, 100) aiming at (0, 0, 80)
	spotRotation := mgl32.QuatRotate(mgl32.DegToRad(-25), mgl32.Vec3{1, 0, 0})
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 40, 130},
			Rotation: spotRotation,
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeSpot,
			Intensity:    80.0,
			Color:        [3]float32{1, 1, 1},
			Range:        150,
			ConeAngle:    50,
			CastsShadows: true,
		},
	)

	setupDone = true
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
