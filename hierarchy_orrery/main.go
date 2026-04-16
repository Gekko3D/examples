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

const orbitVoxelResolution float32 = 0.1
const sunEmitterLinkID uint32 = 1

type OrbiterComponent struct {
	AngularSpeed float32
	OrbitRadius  float32
	Phase        float32
}

type OrbiterModule struct{}

func (m OrbiterModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(orbiterSystem).
			InStage(Update).
			RunAlways(),
	)
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
			WindowWidth:  1440,
			WindowHeight: 900,
			WindowTitle:  "Hierarchy Orrery",
		},
		HierarchyModule{},
		FlyingCameraModule{},
		OrbiterModule{},
		DemoModule{},
	)
	app.Run()
}

func setupScene(cmd *Commands, assets *AssetServer) {
	if setupDone {
		return
	}

	sunPalette := assets.CreatePBRPalette([4]uint8{255, 220, 96, 255}, 0.35, 0.0, 3.0, 1.0)
	planet1Palette := assets.CreateSimplePalette([4]uint8{70, 130, 255, 255})
	planet2Palette := assets.CreateSimplePalette([4]uint8{220, 130, 80, 255})
	planet3Palette := assets.CreateSimplePalette([4]uint8{120, 220, 140, 255})
	moonPalette := assets.CreateSimplePalette([4]uint8{170, 174, 184, 255})
	ringPalette := assets.CreateSimplePalette([4]uint8{214, 200, 150, 255})

	sunModel := assets.CreateSphereModel(10, 1.0)
	planetLargeModel := assets.CreateSphereModel(5, 1.0)
	planetMediumModel := assets.CreateSphereModel(4, 1.0)
	planetSmallModel := assets.CreateSphereModel(3, 1.0)
	moonModel := assets.CreateSphereModel(2, 1.0)
	ringCubeModel := assets.CreateCubeModel(2, 2, 2, 1.0)

	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.04,
			Color:     [3]float32{0.6, 0.65, 0.8},
		},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerStars,
		Seed:       1337,
		Scale:      1.0,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.8, 0.9, 1.0},
		ColorB:     mgl32.Vec3{1.0, 1.0, 1.0},
		Threshold:  0.98,
		Opacity:    1.0,
		Priority:   0,
		BlendMode:  SkyboxBlendAdd,
	})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerStars,
		Seed:       4242,
		Scale:      2.0,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{1.0, 0.8, 0.7},
		ColorB:     mgl32.Vec3{1.0, 1.0, 1.0},
		Threshold:  0.995,
		Opacity:    1.0,
		Priority:   1,
		BlendMode:  SkyboxBlendAdd,
	})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:   SkyboxLayerNebula,
		Seed:        987,
		Scale:       2.5,
		Octaves:     5,
		Persistence: 0.5,
		Lacunarity:  2.0,
		Resolution:  [2]int{1024, 512},
		ColorA:      mgl32.Vec3{0.4, 0.1, 0.6},
		ColorB:      mgl32.Vec3{0.1, 0.4, 0.8},
		Threshold:   0.6,
		Opacity:     0.3,
		Priority:    2,
		Smooth:      true,
		BlendMode:   SkyboxBlendAdd,
		WindSpeed:   mgl32.Vec3{0.005, 0.002, 0.003},
	})

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 20, 25}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: mgl32.Vec3{0, 20, 25},
			LookAt:   mgl32.Vec3{0, 0, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Fov:      55,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      500,
			Yaw:      0,
			Pitch:    -28,
		},
		&FlyingCameraComponent{Speed: 10.0, Sensitivity: 0.1},
	)

	sunEntity := cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 0, 0},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LocalTransformComponent{
			Position: mgl32.Vec3{0, 0, 0},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      sunModel,
			VoxelPalette:    sunPalette,
			VoxelResolution: orbitVoxelResolution,
			EmitterLinkID:   sunEmitterLinkID,
		},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 0, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{
			Type:          LightTypePoint,
			Intensity:     20,
			Color:         [3]float32{1.0, 0.92, 0.75},
			Range:         80,
			CastsShadows:  true,
			SourceRadius:  1.1,
			EmitterLinkID: sunEmitterLinkID,
		},
	)

	planet1 := spawnOrbitingBody(cmd, sunEntity, planetLargeModel, planet1Palette, 8, 0.5, 0.0)
	addMoon(cmd, planet1, moonModel, moonPalette, 2, 2.0, 0.0)

	planet2 := spawnOrbitingBody(cmd, sunEntity, planetMediumModel, planet2Palette, 16, 0.32, math.Pi*0.45)
	addMoon(cmd, planet2, moonModel, moonPalette, 2.5, 1.4, 0.4)
	addMoon(cmd, planet2, moonModel, moonPalette, 3.5, -0.95, 2.4)

	planet3 := spawnOrbitingBody(cmd, sunEntity, planetSmallModel, planet3Palette, 25, 0.2, math.Pi*0.9)
	for i := 0; i < 8; i++ {
		phase := float32(i) * (2 * math.Pi / 8)
		cmd.AddEntity(
			&LocalTransformComponent{
				Position: orbitPosition(4.5, phase),
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{0.45, 0.45, 0.45},
			},
			&TransformComponent{
				Position: orbitPosition(4.5, phase),
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{0.45, 0.45, 0.45},
			},
			&Parent{Entity: planet3},
			&VoxelModelComponent{
				VoxelModel:      ringCubeModel,
				VoxelPalette:    ringPalette,
				VoxelResolution: orbitVoxelResolution,
			},
			&OrbiterComponent{
				AngularSpeed: 0.9,
				OrbitRadius:  4.5,
				Phase:        phase,
			},
		)
	}

	setupDone = true
}

func spawnOrbitingBody(cmd *Commands, parent EntityId, model AssetId, palette AssetId, radius, speed float32, phase float64) EntityId {
	initialPosition := orbitPosition(radius, float32(phase))
	return cmd.AddEntity(
		&LocalTransformComponent{
			Position: initialPosition,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&TransformComponent{
			Position: initialPosition,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&Parent{Entity: parent},
		&VoxelModelComponent{
			VoxelModel:      model,
			VoxelPalette:    palette,
			VoxelResolution: orbitVoxelResolution,
		},
		&OrbiterComponent{
			AngularSpeed: speed,
			OrbitRadius:  radius,
			Phase:        float32(phase),
		},
	)
}

func addMoon(cmd *Commands, parent EntityId, model AssetId, palette AssetId, radius, speed float32, phase float64) {
	initialPosition := orbitPosition(radius, float32(phase))
	cmd.AddEntity(
		&LocalTransformComponent{
			Position: initialPosition,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{0.75, 0.75, 0.75},
		},
		&TransformComponent{
			Position: initialPosition,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{0.75, 0.75, 0.75},
		},
		&Parent{Entity: parent},
		&VoxelModelComponent{
			VoxelModel:      model,
			VoxelPalette:    palette,
			VoxelResolution: orbitVoxelResolution,
		},
		&OrbiterComponent{
			AngularSpeed: speed,
			OrbitRadius:  radius,
			Phase:        float32(phase),
		},
	)
}

func orbitPosition(radius, phase float32) mgl32.Vec3 {
	return mgl32.Vec3{
		radius * float32(math.Cos(float64(phase))),
		0,
		radius * float32(math.Sin(float64(phase))),
	}
}

func orbiterSystem(cmd *Commands, time *Time) {
	if time == nil {
		return
	}

	dt := float32(time.Dt)
	if dt <= 0 {
		return
	}

	MakeQuery2[LocalTransformComponent, OrbiterComponent](cmd).Map(func(eid EntityId, local *LocalTransformComponent, orb *OrbiterComponent) bool {
		orb.Phase += orb.AngularSpeed * dt
		local.Position = mgl32.Vec3{
			orb.OrbitRadius * float32(math.Cos(float64(orb.Phase))),
			0,
			orb.OrbitRadius * float32(math.Sin(float64(orb.Phase))),
		}

		spin := mgl32.QuatRotate(orb.AngularSpeed*dt*1.6, mgl32.Vec3{0, 1, 0})
		local.Rotation = local.Rotation.Mul(spin).Normalize()
		return true
	})
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
