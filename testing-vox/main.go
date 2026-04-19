package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Playing
	Quit
)

const (
	benchWarmupEnv       = "GEKKO_BENCH_WARMUP_SECONDS"
	benchDurationEnv     = "GEKKO_BENCH_SECONDS"
	benchLabelEnv        = "GEKKO_BENCH_LABEL"
	forceHashLookupEnv   = "GEKKO_XBM_FORCE_HASH_LOOKUP"
	defaultBenchWarmupS  = 5.0
	defaultBenchCaptureS = 10.0
)

type SetupModule struct{}

type OrbiterModule struct{}

type SpinnerModule struct{}

type SpinnerComponent struct {
	AngularSpeed mgl32.Vec3
}

type DemoState struct {
	CameraTiles       map[ChunkCoord]EntityId
	SphereEntities    map[ChunkCoord]EntityId
	SpherePalettes    map[ChunkCoord]AssetId
	CameraChunkSize   float32
	SphereChunkSize   float32
	CameraTileModel   AssetId
	SphereModel       AssetId
	ParticleAtlas     AssetId
	GrassAtlas        AssetId
	BluePalette       AssetId
	RedPalette        AssetId
	GreyPalette       AssetId
	CollisionCounts   map[CollisionEventType]uint64
	CollisionLog      []string
	DestructionRadius float32

	ManipulatorMode bool
	GrabbedEntity   EntityId
	GrabOffset      mgl32.Vec3
	GrabDistance    float32
	ModePanel       EntityId
}

type BenchmarkState struct {
	Enabled             bool
	Label               string
	WarmupSeconds       float64
	CaptureSeconds      float64
	CaptureStarted      bool
	CaptureStartElapsed float64
	CaptureStartFrame   uint64
}

type OrbiterComponent struct {
	Radius       float32
	AngularSpeed float32
	Height       float32
	Center       mgl32.Vec3
	Phase        float32
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
			WindowHeight: 800,
			WindowTitle:  "Voxel Scene Demo",
			DebugMode:    true,
		},
		PhysicsModule{
			Synchronous: true,
		},
		VoxPhysicsModule{},
		WaterEffectsModule{},
		HierarchyModule{},
		ChunkObserverModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		OrbiterModule{},
		SpinnerModule{},
		SetupModule{},
		DestructionModule{},
		UiModule{},
	)

	app.Run()
}

func (SetupModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})

	if bench := loadBenchmarkState(); bench != nil {
		cmd.AddResources(bench)
		app.UseSystem(System(benchmarkSystem).InStage(Update).RunAlways())
	}

	app.UseSystem(System(setupScene).InStage(Prelude))
	app.UseSystem(
		System(quitSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
	app.UseSystem(System(collisionEventSystem))
	app.UseSystem(System(debugRaycastSystem))
	app.UseSystem(System(manipulatorSystem))
	app.UseSystem(System(spawnSphereAtClickSystem))
	app.UseSystem(System(chunkObserverHudSystem).InStage(Update).RunAlways())
}

func (OrbiterModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(orbiterMotionSystem).
			InStage(Update).
			RunAlways(),
	)
}

func (SpinnerModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(spinnerSystem).
			InStage(Update).
			RunAlways(),
	)
}

var setupDone bool

func loadBenchmarkState() *BenchmarkState {
	captureSeconds := envFloat64(benchDurationEnv, 0)
	if captureSeconds <= 0 {
		return nil
	}

	label := strings.TrimSpace(os.Getenv(benchLabelEnv))
	if label == "" {
		if os.Getenv(forceHashLookupEnv) == "1" {
			label = "hash-only"
		} else {
			label = "hybrid"
		}
	}

	return &BenchmarkState{
		Enabled:        true,
		Label:          label,
		WarmupSeconds:  envFloat64(benchWarmupEnv, defaultBenchWarmupS),
		CaptureSeconds: captureSeconds,
	}
}

func envFloat64(name string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		fmt.Printf("benchmark: ignoring invalid %s=%q: %v\n", name, raw, err)
		return fallback
	}
	return v
}

func resolveDemoAsset(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	alt := "examples/testing-vox/" + path
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return path
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	if setupDone {
		return
	}

	if state.CameraTiles == nil {
		state.CameraTiles = make(map[ChunkCoord]EntityId)
	}
	if state.SphereEntities == nil {
		state.SphereEntities = make(map[ChunkCoord]EntityId)
	}
	if state.SpherePalettes == nil {
		state.SpherePalettes = make(map[ChunkCoord]AssetId)
	}
	if state.CollisionCounts == nil {
		state.CollisionCounts = make(map[CollisionEventType]uint64)
	}

	state.CameraChunkSize = 16
	state.SphereChunkSize = 16
	if state.DestructionRadius <= 0 {
		state.DestructionRadius = 0.2
	}

	if state.CameraTileModel == (AssetId{}) {
		state.CameraTileModel = assets.CreateCubeModel(state.CameraChunkSize, 1, state.CameraChunkSize, 1.0)
	}
	if state.SphereModel == (AssetId{}) {
		state.SphereModel = assets.CreateSphereModel(6, 1.0)
	}
	if state.ParticleAtlas == (AssetId{}) {
		state.ParticleAtlas = assets.CreateTexture(resolveDemoAsset("assets/particle_atlas.png"))
	}
	if state.GrassAtlas == (AssetId{}) {
		state.GrassAtlas = assets.CreateTexture(resolveDemoAsset("assets/grass.png"))
	}

	state.GreyPalette = assets.CreateSimplePalette([4]uint8{60, 62, 68, 255})
	state.BluePalette = assets.CreateSimplePalette([4]uint8{50, 100, 255, 255})
	state.RedPalette = assets.CreateSimplePalette([4]uint8{255, 50, 50, 255})
	transparentPalette := assets.CreatePBRPalette([4]uint8{180, 220, 255, 210}, 0.15, 0.0, 0.0, 1.5)
	amberPalette := assets.CreateSimplePalette([4]uint8{255, 196, 82, 255})
	cyanPalette := assets.CreateSimplePalette([4]uint8{82, 214, 255, 255})
	limePalette := assets.CreateSimplePalette([4]uint8{154, 232, 96, 255})
	slatePalette := assets.CreateSimplePalette([4]uint8{24, 30, 38, 255})

	floorModel := assets.CreateCubeModel(500, 2, 500, 1.0)
	cubeModel := assets.CreateCubeModel(10, 10, 10, 1.0)
	sphereModel := assets.CreateSphereModel(10, 1.0)
	coneModel := assets.CreateConeModel(10, 10, 1.0)
	pyramidModel := assets.CreatePyramidModel(10, 10, 1.0)
	thinRodModel := assets.CreateCubeModel(28, 2, 2, 1.0)
	thinPlateModel := assets.CreateCubeModel(18, 2, 18, 1.0)
	thinPanelModel := assets.CreateCubeModel(2, 20, 14, 1.0)
	basinFloorModel := assets.CreateCubeModel(154, 2, 130, 1.0)
	basinWallLongModel := assets.CreateCubeModel(158, 20, 4, 1.0)
	basinWallShortModel := assets.CreateCubeModel(4, 20, 134, 1.0)
	basinPlinthModel := assets.CreateCubeModel(154, 16, 130, 1.0)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 8, 10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: mgl32.Vec3{0, 8, 10},
			LookAt:   mgl32.Vec3{0, 8, -15},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Fov:      52,
			Aspect:   1280.0 / 720.0,
			Near:     0.1,
			Far:      1000,
		},
		&FlyingCameraComponent{Speed: 10.0, Sensitivity: 0.1},
	)

	orbiterRadius := float32(160)
	orbiterHeight := float32(10)
	orbiterCenter := mgl32.Vec3{0, orbiterHeight, 0}

	orbiter := cmd.AddEntity(
		&TransformComponent{Position: orbiterCenter.Add(mgl32.Vec3{orbiterRadius, 0, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&OrbiterComponent{Radius: orbiterRadius, AngularSpeed: 0.005, Height: orbiterHeight, Center: orbiterCenter},
	)

	cmd.AddComponents(
		orbiter,
		&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: state.BluePalette},
		NewChunkObserver(3, state.SphereChunkSize).
			WithFilter(func(coord ChunkCoord) bool { return coord.Y >= -1 && coord.Y <= 1 }).
			WithCallbacks(
				func(c *Commands, observer EntityId, coord ChunkCoord) {
					seed := chunkSeed(coord)
					r := rand.New(rand.NewSource(seed))
					paletteID, ok := state.SpherePalettes[coord]
					if !ok {
						paletteID = assets.CreateSimplePalette([4]uint8{
							uint8(100 + r.Intn(150)),
							uint8(100 + r.Intn(150)),
							uint8(140 + r.Intn(110)),
							255,
						})
						state.SpherePalettes[coord] = paletteID
					}

					center := coord.ToCenter(state.SphereChunkSize)
					center[1] = orbiterHeight + (r.Float32()*20 - 10)
					scale := r.Float32()*0.6 + 0.4

					entity := c.AddEntity(
						&TransformComponent{Position: center, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{scale, scale, scale}},
						&VoxelModelComponent{VoxelModel: state.SphereModel, VoxelPalette: paletteID},
					)
					state.SphereEntities[coord] = entity
				},
				func(c *Commands, observer EntityId, coord ChunkCoord) {
					if entity, ok := state.SphereEntities[coord]; ok {
						c.RemoveEntity(entity)
						delete(state.SphereEntities, coord)
					}
				},
			),
	)

	// cmd.AddEntity(
	// 	&LightComponent{Type: LightTypeAmbient, Intensity: 0.035, Color: [3]float32{1, 1, 1}, Range: 40},
	// )
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{10, 30, 10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 2.4, Color: [3]float32{1, 1, 1}, Range: 36, CastsShadows: true},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-10, 18, 8},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-12), mgl32.Vec3{1, 0, 0}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{Type: LightTypeDirectional, Intensity: 0.82, Color: [3]float32{1.0, 0.95, 0.9}, Range: 500, CastsShadows: true},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-7.5, 4.0, 1.5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.5, Color: [3]float32{1.0, 0.65, 0.28}, Range: 12},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{3.0, 5.0, 0.0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 0.85, Color: [3]float32{0.84, 0.88, 1.0}, Range: 10},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{8.5, 2.7, 0.5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.1, Color: [3]float32{0.62, 0.8, 1.0}, Range: 12},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-20, 0, -20}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: floorModel, VoxelPalette: state.GreyPalette, PivotMode: PivotModeCorner},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.5, Restitution: 0.2},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{1.5, 5, 1.5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: state.BluePalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{1, 2, 0.5}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{2, 15, 3.5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: state.RedPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{1, 25, 2.5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: transparentPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{2, 35, 2.5}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: coneModel, VoxelPalette: state.RedPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{2, 45, 3}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: pyramidModel, VoxelPalette: state.RedPalette},
		&RigidBodyComponent{Mass: 1, GravityScale: 1, AngularVelocity: mgl32.Vec3{0, 1, 3}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.5},
	)

	// --- Blocky Water Showcase ---
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23, 0.8, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: basinPlinthModel, VoxelPalette: cyanPalette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.6, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23, 1.7, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: basinFloorModel, VoxelPalette: slatePalette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.6, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23, 2.8, 1.7}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: basinWallLongModel, VoxelPalette: amberPalette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.6, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23, 2.8, 14.3}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: basinWallLongModel, VoxelPalette: amberPalette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.6, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{15.3, 2.8, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: basinWallShortModel, VoxelPalette: amberPalette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.6, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{30.3, 2.8, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: basinWallShortModel, VoxelPalette: amberPalette},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.6, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23, 2.8, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&WaterSurfaceComponent{
			HalfExtents:     [2]float32{7.5, 6.4},
			Depth:           2.00,
			Color:           [3]float32{0.22, 0.48, 0.78},
			AbsorptionColor: [3]float32{0.06, 0.12, 0.2},
			Opacity:         0.3,
			Roughness:       0.1,
			Refraction:      0.22,
			FlowDirection:   [2]float32{1, 0.35},
			FlowSpeed:       0.95,
			WaveAmplitude:   0.02,
		},
		&WaterSplashEffectComponent{
			Texture:        state.ParticleAtlas,
			AtlasCols:      4,
			AtlasRows:      4,
			SplashSprite:   5,
			SpraySprite:    9,
			MinImpactSpeed: 2.0,
			StrengthScale:  1.0,
		},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23, 6.2, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.35, Color: [3]float32{0.48, 0.78, 1.0}, Range: 18},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{28, 8.5, 12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.1, Color: [3]float32{0.38, 0.68, 1.0}, Range: 12},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{20.2, 14.5, 4.6}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.95, 0.95, 0.95}},
		&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: transparentPalette},
		&RigidBodyComponent{Mass: 1.1, Velocity: mgl32.Vec3{1.2, -0.8, 0.6}, GravityScale: 1, AngularVelocity: mgl32.Vec3{1.3, 0.5, 1.9}},
		&ColliderComponent{Friction: 0.25, Restitution: 0.08},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{23.0, 17.0, 8.0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.9, 0.9, 0.9}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: amberPalette},
		&RigidBodyComponent{Mass: 1.2, Velocity: mgl32.Vec3{-0.5, -0.4, 0.3}, GravityScale: 1, AngularVelocity: mgl32.Vec3{0.8, 1.6, 0.9}},
		&ColliderComponent{Friction: 0.3, Restitution: 0.05},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{26.6, 12.8, 10.9}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.85, 0.85, 0.85}},
		&VoxelModelComponent{VoxelModel: pyramidModel, VoxelPalette: limePalette},
		&RigidBodyComponent{Mass: 1.0, Velocity: mgl32.Vec3{-1.0, -0.6, -0.8}, GravityScale: 1, AngularVelocity: mgl32.Vec3{1.1, 2.0, 0.4}},
		&ColliderComponent{Friction: 0.28, Restitution: 0.04},
	)

	// --- Thin-Shape Inertia Lane ---
	// These bodies make it easier to visually verify that rods and plates keep a
	// strong axis preference instead of falling back to generic isotropic spin.
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-14, 14, 4},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(18), mgl32.Vec3{0, 0, 1}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(10), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{VoxelModel: thinRodModel, VoxelPalette: amberPalette},
		&RigidBodyComponent{Mass: 1.2, GravityScale: 1, AngularVelocity: mgl32.Vec3{0.25, 0.8, 4.6}},
		&ColliderComponent{Friction: 0.45, Restitution: 0.15},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-8, 22, 2},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(24), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(-16), mgl32.Vec3{0, 0, 1}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{VoxelModel: thinPlateModel, VoxelPalette: cyanPalette},
		&RigidBodyComponent{Mass: 1.4, GravityScale: 1, AngularVelocity: mgl32.Vec3{2.4, 0.35, 0.7}},
		&ColliderComponent{Friction: 0.4, Restitution: 0.1},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-4, 18, -2},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-12), mgl32.Vec3{0, 1, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(8), mgl32.Vec3{1, 0, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{VoxelModel: thinPanelModel, VoxelPalette: limePalette},
		&RigidBodyComponent{Mass: 1.3, GravityScale: 1, AngularVelocity: mgl32.Vec3{0.55, 3.2, 0.2}},
		&ColliderComponent{Friction: 0.45, Restitution: 0.12},
	)

	// --- PBR Material Gallery ---
	// Rows of spheres showcasing different PBR properties
	for i := 0; i < 6; i++ {
		t := float32(i) / 5.0
		xPos := float32(i)*4.5 - 11.0

		// Row 1: Metallic (Gold-ish) with increasing Roughness
		// Demonstrates: Normalized specular highlights and Metallic Ambient fix
		goldPalette := assets.CreatePBRPalette([4]uint8{255, 215, 120, 255}, t, 1.0, 0.0, 1.5)
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{xPos, 14, -15}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1.5, 1.5, 1.5}},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: goldPalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)

		// Row 2: Dielectric (Plastic Blue) with increasing Roughness
		// Demonstrates: Diffuse/Specular balance and Fresnel at grazing angles
		plasticPalette := assets.CreatePBRPalette([4]uint8{50, 120, 255, 255}, t, 0.0, 0.0, 1.5)
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{xPos, 10, -15}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1.5, 1.5, 1.5}},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: plasticPalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)

		// Row 3: Emissive (Neon Green) with increasing Intensity
		// Demonstrates: Emission pipeline
		emission := t * 4.0
		emissivePalette := assets.CreatePBRPalette([4]uint8{100, 255, 100, 255}, 0.5, 0.0, emission, 1.0)
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{xPos, 6, -15}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1.5, 1.5, 1.5}},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: emissivePalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)

		// Row 4: Transparent (Glass) with increasing Transparency
		// Demonstrates: WBOIT shading and Resolve improvements
		opacity := 0.8 - (t * 0.7) // Range from 0.8 down to 0.1
		alpha := uint8(opacity * 255)
		glassPalette := assets.CreatePBRPalette([4]uint8{200, 220, 255, alpha}, 0.1, 0.0, 0.0, 1.5)
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{xPos, 2, -15}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1.5, 1.5, 1.5}},
			&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: glassPalette},
			&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.5, 0}},
		)
	}

	// --- Purely Transform-based Rotating Primitives (No Physics) ---
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-28, 5, 18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: state.BluePalette},
		&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 2, 0}},
	)

	// --- GPU Cellular Volumes ---
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-7.095, 6.61, 0.805},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1.55, 3.1, 1.55},
		},
		&CellularVolumeComponent{
			Type:        CellularFire,
			Preset:      CAVolumePresetTorch,
			Resolution:  [3]int{22, 42, 22},
			TickRate:    28,
			Diffusion:   0.08,
			Buoyancy:    0.98,
			Cooling:     0.025,
			Dissipation: 0.01,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0.205, 2.45, 1.205},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1.85, 2.35, 1.85},
		},
		&CellularVolumeComponent{
			Type:        CellularFire,
			Preset:      CAVolumePresetCampfire,
			Resolution:  [3]int{26, 20, 26},
			TickRate:    22,
			Diffusion:   0.18,
			Buoyancy:    0.52,
			Cooling:     0.06,
			Dissipation: 0.032,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-0.4, 2.82, 0.32},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{2.0, 1.6, 1.9},
		},
		&CellularVolumeComponent{
			Type:                  CellularSmoke,
			Preset:                CAVolumePresetCampfire,
			Resolution:            [3]int{18, 14, 16},
			TickRate:              18,
			Diffusion:             0.18,
			Buoyancy:              0.12,
			Dissipation:           0.042,
			UseAppearanceOverride: true,
			ScatterColor:          [3]float32{0.58, 0.52, 0.44},
			UseShadowTintOverride: true,
			ShadowTint:            [3]float32{0.28, 0.22, 0.18},
			UseAbsorptionOverride: true,
			AbsorptionColor:       [3]float32{0.12, 0.09, 0.07},
			Extinction:            0.88,
			Emission:              0.0,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{4.045, 4.155, -0.055},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{0.9, 2.1, 0.9},
		},
		&CellularVolumeComponent{
			Type:        CellularFire,
			Preset:      CAVolumePresetJetFlame,
			Resolution:  [3]int{21, 31, 21},
			TickRate:    30,
			Diffusion:   0.06,
			Buoyancy:    0.08,
			Cooling:     0.04,
			Dissipation: 0.02,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{12.9, 6.01, 4.1},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CellularVolumeComponent{
			Type:                  CellularFire,
			Preset:                CAVolumePresetExplosion,
			Resolution:            [3]int{48, 64, 48},
			TickRate:              30,
			Diffusion:             0.06,
			Buoyancy:              1.35,
			Cooling:               0.14,
			Dissipation:           0.02,
			UseAppearanceOverride: true,
			ScatterColor:          [3]float32{0.56, 0.5, 0.46},
			UseShadowTintOverride: true,
			ShadowTint:            [3]float32{0.22, 0.18, 0.15},
			UseAbsorptionOverride: true,
			AbsorptionColor:       [3]float32{0.1, 0.08, 0.06},
			Extinction:            1.4,
			Emission:              3.6,
		},
	)

	// --- Procedural Skybox Layers ---

	// 1. Base Sky (Atmospheric Gradient)
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

	// 2. Large Clouds
	cmd.AddEntity(&SkyboxLayerComponent{
		NoiseType:   SkyboxNoisePerlin,
		Seed:        42,
		Scale:       4.0,
		Octaves:     4,
		Persistence: 0.5,
		Lacunarity:  2.0,
		Resolution:  [2]int{1024, 512},
		ColorA:      mgl32.Vec3{1, 1, 1},       // Clouds
		ColorB:      mgl32.Vec3{0.8, 0.8, 0.9}, // Cloud shading
		Threshold:   0.5,
		Opacity:     0.8,
		Priority:    1,
		Smooth:      true,
		BlendMode:   SkyboxBlendAlpha,
		WindSpeed:   mgl32.Vec3{0.02, 0.01, 0}, // Moving clouds!
	})

	// 3. Small Detail Clouds / Haze
	cmd.AddEntity(&SkyboxLayerComponent{
		NoiseType:  SkyboxNoisePerlin,
		Seed:       999,
		Scale:      12.0,
		Octaves:    2,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.9, 0.9, 1.0},
		ColorB:     mgl32.Vec3{0.7, 0.7, 0.8},
		Threshold:  0.6,
		Opacity:    0.4,
		Priority:   2,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
		WindSpeed:  mgl32.Vec3{-0.03, 0.02, 0.01},
	})

	// --- Grass Patches (Demo Cylindrical Billboarding) ---
	for x := float32(-15); x <= 15; x += 1.5 {
		for z := float32(-15); z <= 15; z += 1.5 {
			// Randomize slightly
			rx := x + (rand.Float32()*0.5 - 0.25)
			rz := z + (rand.Float32()*0.5 - 0.25)
			if rand.Float32() > 0.2 {
				cmd.AddEntity(
					&SpriteComponent{
						Enabled:       true,
						Position:      mgl32.Vec3{rx, 0.75, rz}, // Half height up
						Size:          [2]float32{1.5, 1.5},
						Color:         [4]float32{1, 1, 1, 1},
						BillboardMode: BillboardCylindrical,
						Texture:       state.GrassAtlas,
						AtlasCols:     1,
						AtlasRows:     1,
					},
				)
			}
		}
	}

	// --- Particle Emitters ---

	// 1. Fire fountain
	// cmd.AddEntity(
	// 	&TransformComponent{Position: mgl32.Vec3{2, 0.1, 3}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
	// 	&ParticleEmitterComponent{
	// 		Enabled:          true,
	// 		MaxParticles:     10000,
	// 		SpawnRate:        500,
	// 		LifetimeRange:    [2]float32{1.5, 2.5},
	// 		StartSpeedRange:  [2]float32{10, 15},
	// 		StartSizeRange:   [2]float32{0.1, 0.3},
	// 		StartColorMin:    [4]float32{1.0, 0.5, 0.0, 1.0},
	// 		StartColorMax:    [4]float32{1.0, 0.2, 0.0, 1.0},
	// 		Gravity:          9.8,
	// 		Drag:             0.1,
	// 		ConeAngleDegrees: 15,
	// 		SpriteIndex:      4,
	// 		AtlasCols:        4,
	// 		AtlasRows:        4,
	// 		Texture:          state.ParticleAtlas,
	// 		AlphaMode:        SpriteAlphaLuminance,
	// 	},
	// )

	state.ModePanel = cmd.AddEntity(buildDemoModePanel(state))

	setupDone = true
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}

func benchmarkSystem(cmd *Commands, bench *BenchmarkState, time *Time, state *VoxelRtState) {
	if bench == nil || !bench.Enabled || state == nil || state.RtApp == nil || time == nil {
		return
	}

	if !bench.CaptureStarted {
		if time.Elapsed < bench.WarmupSeconds {
			return
		}
		bench.CaptureStarted = true
		bench.CaptureStartElapsed = time.Elapsed
		bench.CaptureStartFrame = state.RtApp.RenderFrameIndex
		fmt.Printf(
			"BENCH start label=%s warmup_s=%.2f capture_s=%.2f force_hash=%t\n",
			bench.Label,
			bench.WarmupSeconds,
			bench.CaptureSeconds,
			os.Getenv(forceHashLookupEnv) == "1",
		)
		return
	}

	captureElapsed := time.Elapsed - bench.CaptureStartElapsed
	if captureElapsed < bench.CaptureSeconds {
		return
	}

	frameDelta := state.RtApp.RenderFrameIndex - bench.CaptureStartFrame
	avgFPS := 0.0
	avgFrameMS := 0.0
	if captureElapsed > 0 && frameDelta > 0 {
		avgFPS = float64(frameDelta) / captureElapsed
		avgFrameMS = (captureElapsed * 1000.0) / float64(frameDelta)
	}

	fmt.Printf(
		"BENCH result label=%s warmup_s=%.2f capture_s=%.2f elapsed_s=%.3f frames=%d avg_fps=%.2f avg_frame_ms=%.2f last_fps=%.2f objects=%d visible=%d terrain_chunks=%d voxel_sec_up=%d voxel_brk_up=%d\n",
		bench.Label,
		bench.WarmupSeconds,
		bench.CaptureSeconds,
		captureElapsed,
		frameDelta,
		avgFPS,
		avgFrameMS,
		state.FPS(),
		state.Counter("Objects"),
		state.Counter("Visible"),
		state.Counter("TerrainChunks"),
		state.Counter("VoxelSecUp"),
		state.Counter("VoxelBrkUp"),
	)
	if stats := strings.TrimSpace(state.RtApp.PreviousProfilerStats); stats != "" {
		fmt.Printf("BENCH profiler label=%s\n%s\n", bench.Label, stats)
	}

	bench.Enabled = false
	cmd.ChangeState(Quit)
}

func debugRaycastSystem(state *VoxelRtState, input *Input, cmd *Commands, assets *AssetServer, queue *DestructionQueue, demoState *DemoState) {
	// Radius control
	if input.JustPressed[KeyRightBracket] || input.MouseScrollY > 0 {
		demoState.DestructionRadius += 0.05
	}
	if input.JustPressed[KeyLeftBracket] || input.MouseScrollY < 0 {
		demoState.DestructionRadius -= 0.05
		if demoState.DestructionRadius < 0.05 {
			demoState.DestructionRadius = 0.05
		}
	}

	if input.JustPressed[MouseButtonLeft] || input.JustPressed[MouseButtonRight] || input.JustReleased[MouseButtonLeft] {
		var cam *CameraComponent
		MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, c *CameraComponent) bool {
			cam = c
			return false
		})

		if cam != nil {
			if demoState.ManipulatorMode {
				if input.JustPressed[MouseButtonLeft] {
					origin, dir := state.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
					result := state.Raycast(origin, dir, 1000)
					if result.Hit {
						// Check if it's a rigid body using a query for pointers
						found := false
						MakeQuery2[TransformComponent, RigidBodyComponent](cmd).Map(func(eid EntityId, tc *TransformComponent, rbc *RigidBodyComponent) bool {
							if eid == result.Entity {
								if !rbc.IsStatic {
									demoState.GrabbedEntity = result.Entity
									demoState.GrabDistance = result.T

									hitPos := origin.Add(dir.Mul(result.T))
									demoState.GrabOffset = tc.Rotation.Conjugate().Rotate(hitPos.Sub(tc.Position))

									fmt.Printf("Grabbed entity %v, distance: %.2f, offset: %v\n", result.Entity, result.T, demoState.GrabOffset)
									found = true
								}
								return false
							}
							return true
						})

						if !found {
							fmt.Printf("Clicked object %v is not a dynamic rigid body\n", result.Entity)
						}
					}
				} else if input.JustReleased[MouseButtonLeft] {
					demoState.GrabbedEntity = EntityId(0)
					fmt.Printf("Released entity\n")
				}
				return
			}

			if input.JustPressed[MouseButtonLeft] || input.JustPressed[MouseButtonRight] {
				origin, dir := state.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
				result := state.Raycast(origin, dir, 1000)
				if result.Hit {
					hitPos := origin.Add(dir.Mul(result.T))

					if input.JustPressed[MouseButtonLeft] {
						// Queue standard destruction
						queue.Events = append(queue.Events, DestructionEvent{
							Entity: result.Entity,
							Center: hitPos,
							Radius: demoState.DestructionRadius,
						})

						// Also apply impulse if it's a physics body
						var rb *RigidBodyComponent
						for _, c := range cmd.GetAllComponents(result.Entity) {
							if rbc, ok := c.(*RigidBodyComponent); ok {
								rb = rbc
							} else if rbc, ok := c.(RigidBodyComponent); ok {
								rb = &rbc
							}
						}

						if rb != nil {
							// Apply impulse outward from camera
							impulse := dir.Normalize().Mul(15.0)
							rb.Velocity = rb.Velocity.Add(impulse)
							fmt.Printf("Voxel Hit on Physics Body! Destruction + Impulse applied to entity %v\n", result.Entity)
						} else {
							fmt.Printf("Voxel Hit on Static Body! Destruction queued at: %v\n", result.Pos)
						}
					} else if input.JustPressed[MouseButtonRight] {
						// Demolish: Large-radius destruction
						queue.Events = append(queue.Events, DestructionEvent{
							Entity: result.Entity,
							Center: hitPos,
							Radius: demoState.DestructionRadius * 10.0,
						})
						fmt.Printf("Demolishing entity %v\n", result.Entity)
					}
					return
				}
				fmt.Printf("CPU Raycast MISS\n")
			}
		}
	}
}

func manipulatorSystem(cmd *Commands, input *Input, rtState *VoxelRtState, demoState *DemoState, time *Time) {
	if !demoState.ManipulatorMode || demoState.GrabbedEntity == EntityId(0) {
		return
	}

	var cam *CameraComponent
	MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, c *CameraComponent) bool {
		cam = c
		return false
	})

	if cam == nil {
		return
	}

	// Find the grabbed entity's components using query for pointers
	found := false
	MakeQuery2[TransformComponent, RigidBodyComponent](cmd).Map(func(eid EntityId, transform *TransformComponent, rb *RigidBodyComponent) bool {
		if eid == demoState.GrabbedEntity {
			// Calculate target position in world space
			origin, dir := rtState.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
			targetPos := origin.Add(dir.Mul(demoState.GrabDistance))

			// Current grab point in world space
			rWorld := transform.Rotation.Rotate(demoState.GrabOffset)
			currentGrabPos := transform.Position.Add(rWorld)

			// PD Controller for pulling
			// Scale stiffness by mass to keep interaction feel similar
			stiffness := float32(300.0)
			damping := float32(20.0)

			if rb.Mass > 0 {
				stiffness *= rb.Mass
				damping *= rb.Mass
			}

			diff := targetPos.Sub(currentGrabPos)

			// Calculate velocity at the grab point
			velAtPoint := rb.Velocity.Add(rb.AngularVelocity.Cross(rWorld))

			// Total force to apply
			force := diff.Mul(stiffness).Sub(velAtPoint.Mul(damping))

			// Apply Linear Impulse
			rb.ApplyImpulse(force.Mul(float32(time.Dt)))

			// Apply Torque (r x F)
			torque := rWorld.Cross(force)
			rb.ApplyTorque(torque.Mul(float32(time.Dt)))

			// Extra damping for stability while grabbed
			rb.Velocity = rb.Velocity.Mul(0.98)
			rb.AngularVelocity = rb.AngularVelocity.Mul(0.95)

			found = true
			return false
		}
		return true
	})

	if !found {
		demoState.GrabbedEntity = EntityId(0)
	}
}

func collisionEventSystem(cmd *Commands, proxy *PhysicsProxy, state *DemoState) {
	if proxy == nil || state == nil {
		return
	}

	events := proxy.DrainCollisionEvents()
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		state.CollisionCounts[event.Type]++
		if event.Type != CollisionEventEnter {
			continue
		}

		fmt.Printf(
			"Collision %s: %d <-> %d speed=%.2f impulse=%.2f point=%v\n",
			event.Type.String(),
			event.A,
			event.B,
			event.RelativeSpeed,
			event.NormalImpulse,
			event.Point,
		)

		state.CollisionLog = append(state.CollisionLog,
			fmt.Sprintf("%d/%d speed %.1f impulse %.1f", event.A, event.B, event.RelativeSpeed, event.NormalImpulse),
		)
		if len(state.CollisionLog) > 4 {
			state.CollisionLog = append([]string(nil), state.CollisionLog[len(state.CollisionLog)-4:]...)
		}

		cmd.AddEntity(
			&SpriteComponent{
				Enabled:       true,
				Position:      event.Point,
				Size:          [2]float32{1.2, 1.2},
				Color:         [4]float32{1.0, 0.5, 0.1, 0.95},
				SpriteIndex:   5,
				AtlasCols:     4,
				AtlasRows:     4,
				Texture:       state.ParticleAtlas,
				BillboardMode: BillboardSpherical,
				Unlit:         true,
				AlphaMode:     SpriteAlphaLuminance,
			},
			&LifetimeComponent{TimeLeft: 0.2},
		)
	}
}

func orbiterMotionSystem(cmd *Commands, time *Time) {
	if time == nil {
		return
	}

	dt := float32(time.Dt)
	if dt <= 0 {
		return
	}

	MakeQuery2[TransformComponent, OrbiterComponent](cmd).Map(func(eid EntityId, transform *TransformComponent, orbiter *OrbiterComponent) bool {
		if orbiter == nil || transform == nil {
			return true
		}

		if orbiter.AngularSpeed == 0 {
			orbiter.AngularSpeed = 0.1
		}

		orbiter.Phase += orbiter.AngularSpeed * dt
		angle := float64(orbiter.Phase) * 2 * math.Pi

		transform.Position = mgl32.Vec3{
			orbiter.Center.X() + orbiter.Radius*float32(math.Cos(angle)),
			orbiter.Height,
			orbiter.Center.Z() + orbiter.Radius*float32(math.Sin(angle)),
		}
		return true
	})
}

func spinnerSystem(cmd *Commands, time *Time) {
	if time == nil {
		return
	}

	dt := float32(time.Dt)
	MakeQuery2[TransformComponent, SpinnerComponent](cmd).Map(func(eid EntityId, transform *TransformComponent, spinner *SpinnerComponent) bool {
		rotX := mgl32.QuatRotate(spinner.AngularSpeed.X()*dt, mgl32.Vec3{1, 0, 0})
		rotY := mgl32.QuatRotate(spinner.AngularSpeed.Y()*dt, mgl32.Vec3{0, 1, 0})
		rotZ := mgl32.QuatRotate(spinner.AngularSpeed.Z()*dt, mgl32.Vec3{0, 0, 1})

		transform.Rotation = transform.Rotation.Mul(rotX).Mul(rotY).Mul(rotZ).Normalize()
		return true
	})
}

func chunkSeed(coord ChunkCoord) int64 {
	return int64(coord.X)*73856093 ^ int64(coord.Y)*19349663 ^ int64(coord.Z)*83492791
}

func spawnSphereAtClickSystem(cmd *Commands, input *Input, rtState *VoxelRtState, state *DemoState, assets *AssetServer) {
	if input.JustPressed[KeyF] {
		var cam *CameraComponent
		MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, c *CameraComponent) bool {
			cam = c
			return false
		})

		if cam != nil {
			origin, dir := rtState.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
			const spawnDistance = float32(5.0)
			dir = dir.Normalize()
			spawnPos := origin.Add(dir.Mul(spawnDistance))

			// Random model selection
			var model AssetId
			switch rand.Intn(4) {
			case 0:
				model = state.SphereModel
			default:
				model = state.CameraTileModel
			}

			palette := assets.CreateSimplePalette([4]uint8{
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				255,
			})

			fmt.Printf("Spawning physics body at %v with impulse\n", spawnPos)
			cmd.AddEntity(
				&TransformComponent{Position: spawnPos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.5, 0.5, 0.5}},
				&VoxelModelComponent{VoxelModel: model, VoxelPalette: palette},
				&RigidBodyComponent{
					Mass:     1,
					Velocity: dir.Mul(20.0), // Initial impulse
				},
				&ColliderComponent{Friction: 0.5, Restitution: 0.5},
			)
		}
	}
}

func chunkObserverHudSystem(state *VoxelRtState, demoState *DemoState, cmd *Commands) {
	if state == nil || demoState == nil {
		return
	}

	debrisCount := 0
	MakeQuery1[DebrisComponent](cmd).Map(func(eid EntityId, d *DebrisComponent) bool {
		debrisCount++
		return true
	})

	// Draw HUD status
	modeText := "EDIT MODE"
	if demoState.ManipulatorMode {
		modeText = "MANIPULATE MODE"
	}
	hudText := fmt.Sprintf("Mode: %s\nDestruction Radius: %.2f ( [ ] to adjust)\nDebris Count: %d\n'F' to Fire Ball",
		modeText, demoState.DestructionRadius, debrisCount)

	state.DrawText(hudText, 20, 20, 1.0, [4]float32{1, 1, 1, 1})

	// Collision log
	logX := float32(20)
	logY := float32(100)
	state.DrawText("Collision Log:", logX, logY, 0.8, [4]float32{1, 1, 0, 1})
	for i, entry := range demoState.CollisionLog {
		state.DrawText(entry, logX+20, logY+30+float32(i*25), 0.7, [4]float32{0.8, 0.8, 0.8, 1})
	}

	if demoState.ModePanel == 0 {
		demoState.ModePanel = cmd.AddEntity(buildDemoModePanel(demoState))
	} else {
		cmd.AddComponents(demoState.ModePanel, buildDemoModePanel(demoState))
	}
}

func buildDemoModePanel(state *DemoState) *UiPanel {
	label := "MODE: EDIT"
	if state.ManipulatorMode {
		label = "MODE: MANIPULATE"
	}

	return &UiPanel{
		Anchor:   UiAnchorTopLeft,
		Position: [2]float32{20, 180},
		Width:    220,
		Scale:    0.8,
		Visible:  true,
		Children: []UiNode{
			UiButtonControl{
				Key:   "toggle_mode",
				Label: label,
				Width: 180,
				Scale: 0.8,
				OnClick: func() {
					state.ManipulatorMode = !state.ManipulatorMode
				},
			},
		},
	}
}
