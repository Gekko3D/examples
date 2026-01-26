package main

import (
	"fmt"
	"math"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Playing
	Quit
)

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
			WindowTitle:  "Agnostic Physics Demo",
			AmbientLight: mgl32.Vec3{0.2, 0.2, 0.2},
		},
		PhysicsModule{},
		VoxPhysicsModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		SetupModule{},
	)

	app.Run()
}

type SetupModule struct{}

func (SetupModule) Install(app *App, cmd *Commands) {
	app.UseSystem(System(setupScene).InStage(Prelude))

	app.UseSystem(
		System(quitSystem).
			InStage(PreUpdate).
			RunAlways(),
	)

	app.UseSystem(System(debugRaycastSystem))
}

var setupDone = false

func setupScene(cmd *Commands, assets *AssetServer, physicsWorld *PhysicsWorld) {
	if setupDone {
		return
	}
	//physicsWorld.CollisionMode = CollisionModeVoxel

	// Assets
	palette := assets.CreateSimplePalette([4]uint8{100, 100, 100, 255})
	bluePalette := assets.CreateSimplePalette([4]uint8{50, 100, 255, 255})
	redPalette := assets.CreateSimplePalette([4]uint8{255, 50, 50, 255})

	// Playa Model
	var playaModel AssetId
	var playaPalette AssetId
	hasPlaya := false
	playaFile, err := LoadVoxFile("assets/playa.vox")
	if err == nil && len(playaFile.Models) > 0 {
		playaModel = assets.CreateVoxelModel(playaFile.Models[0], 1.0)
		playaPalette = assets.CreateVoxelPalette(playaFile.Palette, playaFile.VoxMaterials)
		hasPlaya = true
	}

	// Use smaller scale for demo: 1 unit = 0.1m by default in renderer
	// So model size (100, 2, 100) voxels = (10, 0.2, 10) meters.
	floorModel := assets.CreateCubeModel(300, 2, 300, 1.0)
	cubeModel := assets.CreateCubeModel(10, 10, 10, 1.0)
	sphereModel := assets.CreateSphereModel(10, 1.0)
	coneModel := assets.CreateConeModel(10, 10, 1.0)
	pyramidModel := assets.CreatePyramidModel(10, 10, 1.0)

	// Camera
	cmd.AddEntity(
		&CameraComponent{
			Position: mgl32.Vec3{0, 5, 20},
			LookAt:   mgl32.Vec3{0, -2, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0, // Facing towards origin from (0, 10, 20)
			Fov:      45,
			Aspect:   1280.0 / 720.0,
			Near:     0.1,
			Far:      1000,
		},
		&FlyingCameraComponent{
			Speed:       10.0,
			Sensitivity: 0.1,
		},
	)

	// Lights
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{10, 30, 10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 8, Color: [3]float32{1, 1, 1}, Range: 40},
	)

	// Floor
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-10, -1, -10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: floorModel, VoxelPalette: palette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.5, Restitution: 0.2},
	)

	// Cube 1 (Lower) - starts at y=5, moved to X=5, Z=5 to be clearly on floor
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{5, 5, 5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: bluePalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{1, 2, 0.5}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	// Cube 2 (Upper) - starts at y=15, moved to X=5, Z=5
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{5, 15, 5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: redPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	// Sphere
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{5, 25, 5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: redPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	// Cone
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{5, 35, 5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: coneModel, VoxelPalette: redPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	// Pyramid
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{5, 45, 5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: pyramidModel, VoxelPalette: redPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	// Extra Cubes to see more interaction
	for i := 0; i < 3; i++ {
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{float32(i*5) - 5, 20 + float32(i*5), 0},
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: redPalette},
			&RigidBodyComponent{
				Mass:            1,
				GravityScale:    1,
				Velocity:        mgl32.Vec3{0, -5, 0},
				AngularVelocity: mgl32.Vec3{float32(i), 1, float32(i)},
			},
			&ColliderComponent{Friction: 0.3, Restitution: 0.5},
		)
	}

	if hasPlaya {
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{14, 30, 4}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&VoxelModelComponent{VoxelModel: playaModel, VoxelPalette: playaPalette},
			&RigidBodyComponent{Mass: 5, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 2, 1}},
			&ColliderComponent{Friction: 0.3, Restitution: 0.5},
		)
	}

	humanFile, err := LoadVoxFile("/Users/ddevidch/code/go/gekko3d/actiongame/assets/human.vox")
	humanAsset := assets.CreateVoxelFile(humanFile)
	assets.SpawnHierarchicalVoxelModel(cmd, humanAsset, TransformComponent{
		Position: mgl32.Vec3{5, 5, 5},
		Rotation: mgl32.QuatIdent(),
		Scale:    mgl32.Vec3{1, 1, 1},
	}, 1.0)

	castleFile, err := LoadVoxFile("/Users/ddevidch/code/MagicaVoxel/sponza.vox")
	castleAsset := assets.CreateVoxelFile(castleFile)
	assets.SpawnHierarchicalVoxelModel(cmd, castleAsset, TransformComponent{
		Position: mgl32.Vec3{50, 0, 50},
		Rotation: mgl32.QuatIdent(),
		Scale:    mgl32.Vec3{1, 1, 1},
	}, 1.0)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-5, 5, -5},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CellularVolumeComponent{
			Resolution:        [3]int{32, 48, 32},
			Type:              CellularFire,
			TickRate:          15.0,
			Diffusion:         0.25,
			Buoyancy:          0.6,
			Cooling:           0.0,
			Dissipation:       0.02,
			BridgeToParticles: false,
			BridgeToVoxels:    true,
			VoxelThreshold:    0.10,
			VoxelStride:       1,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-5, 5, 5},
			Rotation: mgl32.QuatRotate(45, mgl32.Vec3{0, 0, 1}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&ParticleEmitterComponent{
			Enabled:          true,
			MaxParticles:     5000,
			SpawnRate:        1600,
			LifetimeRange:    [2]float32{0.9, 1.8},
			StartSpeedRange:  [2]float32{6, 14},
			StartSizeRange:   [2]float32{0.05, 0.1},
			StartColorMin:    [4]float32{1.0, 0.6, 0.2, 0.18},
			StartColorMax:    [4]float32{1.0, 0.2, 0.0, 0.55},
			Gravity:          9.8,
			Drag:             0.12,
			ConeAngleDegrees: 24.0,
		},
	)

	setupDone = true
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}

func debugRaycastSystem(state *VoxelRtState, input *Input, cmd *Commands, assets *AssetServer) {
	if input.JustPressed[MouseButtonLeft] {
		var cam *CameraComponent
		MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, c *CameraComponent) bool {
			cam = c
			return false
		})

		if cam == nil {
			return
		}

		yawRad := mgl32.DegToRad(cam.Yaw)
		pitchRad := mgl32.DegToRad(cam.Pitch)
		forward := mgl32.Vec3{
			float32(math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad))),
			float32(math.Sin(float64(pitchRad))),
			float32(-math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad))),
		}.Normalize()

		// Offset ray slightly to avoid self-intersection with near plane logic if any
		start := cam.Position.Add(forward.Mul(0.5))
		res := state.Raycast(start, forward, 100.0)
		if res.Hit {
			fmt.Printf("CPU Raycast HID: Entity=%d T=%f Pos=%v Normal=%v\n", res.Entity, res.T, res.Pos, res.Normal)

			// Visual Hitmarker (RED CUBE)
			markerModel := assets.CreateCubeModel(1, 1, 1, 1.0)
			markerPalette := assets.CreateSimplePalette([4]uint8{255, 0, 0, 255})

			hitPos := start.Add(forward.Mul(res.T))
			cmd.AddEntity(
				&TransformComponent{Position: hitPos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.2, 0.2, 0.2}},
				&VoxelModelComponent{VoxelModel: markerModel, VoxelPalette: markerPalette},
				&LifetimeComponent{TimeLeft: 2.0},
			)

		} else {
			fmt.Printf("CPU Raycast MISS\n")
		}
	}
}
