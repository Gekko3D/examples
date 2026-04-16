package main

import (
	"fmt"
	"os"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const demoVoxelResolution float32 = 0.2

const (
	Startup State = iota
	Quit
)

var (
	mainPoolCenter                       = mgl32.Vec3{0, 2.6, -2}
	mainPoolDepth                float32 = 4.6
	mainPoolFitBoundsHalfExtents         = mgl32.Vec3{9.4, 4.0, 6.9}

	decorativePoolCenter               = mgl32.Vec3{-13, 1.75, 8}
	decorativePoolFitBoundsHalfExtents = mgl32.Vec3{4.8, 2.4, 3.7}
)

type DemoModule struct{}

type DemoState struct {
	ParticleAtlas AssetId

	GroundModel    AssetId
	PlinthModel    AssetId
	PoolFloorModel AssetId
	PoolFrameModel AssetId
	SphereModel    AssetId
	CubeModel      AssetId
	TallCubeModel  AssetId

	StonePalette      AssetId
	WallPalette       AssetId
	SpherePalette     AssetId
	CubePalette       AssetId
	HeavyCubePalette  AssetId
	ProjectilePalette AssetId
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
			WindowTitle:  "Water World",
		},
		PhysicsModule{Synchronous: true},
		VoxPhysicsModule{},
		WaterEffectsModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		DemoModule{},
	)

	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(spawnProjectileSystem).InStage(Update))
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func resolveDemoAsset(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	alt := "examples/water_world/" + path
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return path
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	state.ParticleAtlas = assets.CreateTexture(resolveDemoAsset("assets/particle_atlas.png"))

	state.GroundModel = assets.CreateCubeModel(260, 2, 220, 1.0)
	state.PlinthModel = assets.CreateCubeModel(88, 8, 68, 1.0)
	state.PoolFloorModel = assets.CreateCubeModel(80, 4, 60, 1.0)
	state.PoolFrameModel = assets.CreateFrameModel(86, 14, 62, 4, 1.0)
	state.SphereModel = assets.CreateSphereModel(6, 1.0)
	state.CubeModel = assets.CreateCubeModel(6, 6, 6, 1.0)
	state.TallCubeModel = assets.CreateCubeModel(6, 10, 6, 1.0)

	state.StonePalette = assets.CreateSimplePalette([4]uint8{68, 74, 82, 255})
	state.WallPalette = assets.CreateSimplePalette([4]uint8{186, 148, 100, 255})
	state.SpherePalette = assets.CreateSimplePalette([4]uint8{130, 192, 255, 255})
	state.CubePalette = assets.CreateSimplePalette([4]uint8{255, 186, 96, 255})
	state.HeavyCubePalette = assets.CreateSimplePalette([4]uint8{214, 92, 92, 255})
	state.ProjectilePalette = assets.CreateSimplePalette([4]uint8{255, 240, 180, 255})

	addWorldFloor(cmd, state)
	addMainPool(cmd, state)
	addDecorativePool(cmd, state)
	addFallingBodies(cmd, state)
	addLights(cmd)
	addCamera(cmd)

	fmt.Println("Water World scene ready")
	fmt.Println("Left click to launch a camera-aimed projectile")
	fmt.Println("WASD + mouse to fly, Escape to quit")
}

func addWorldFloor(cmd *Commands, state *DemoState) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, -1.2, 2}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: state.GroundModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.9, Restitution: 0.05},
	)
}

func addMainPool(cmd *Commands, state *DemoState) {
	addPoolStructure(cmd, state, mainPoolCenter, state.PlinthModel, state.PoolFloorModel, state.PoolFrameModel)

	cmd.AddEntity(
		&TransformComponent{Position: mainPoolCenter, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&WaterBodyComponent{
			Mode:              WaterBodyModeFitBounds,
			SurfaceY:          mainPoolCenter.Y(),
			Depth:             mainPoolDepth,
			BoundsCenter:      mainPoolCenter,
			BoundsHalfExtents: mainPoolFitBoundsHalfExtents,
			MinCellSize:       demoVoxelResolution,
			Inset:             0.0,
			DebugName:         "main_pool",
			Color:             [3]float32{0.16, 0.36, 0.64},
			AbsorptionColor:   [3]float32{0.18, 0.28, 0.48},
			Opacity:           0.5,
			Roughness:         0.14,
			Refraction:        0.16,
			FlowDirection:     [2]float32{1, 0.3},
			FlowSpeed:         0.8,
			WaveAmplitude:     0.025,
		},
		&WaterSplashEffectComponent{
			Texture:        state.ParticleAtlas,
			AtlasCols:      4,
			AtlasRows:      4,
			SplashSprite:   5,
			SpraySprite:    9,
			FlashSprite:    10,
			MinImpactSpeed: 2.0,
			StrengthScale:  1.0,
		},
	)
}

func addDecorativePool(cmd *Commands, state *DemoState) {
	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter.Add(mgl32.Vec3{0, -1.95, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.55, 0.6, 0.55}},
		&VoxelModelComponent{VoxelModel: state.PlinthModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.8, Restitution: 0.0},
	)

	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter.Add(mgl32.Vec3{0, -1.6, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.55, 1, 0.55}},
		&VoxelModelComponent{VoxelModel: state.PoolFloorModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.65, Restitution: 0.0},
	)

	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter.Add(mgl32.Vec3{0, -1.4, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.55, 1, 0.55}},
		&VoxelModelComponent{VoxelModel: state.PoolFrameModel, VoxelPalette: state.WallPalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.7, Restitution: 0.0},
	)

	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&WaterBodyComponent{
			Mode:              WaterBodyModeFitBounds,
			SurfaceY:          decorativePoolCenter.Y(),
			Depth:             1.4,
			BoundsCenter:      decorativePoolCenter,
			BoundsHalfExtents: decorativePoolFitBoundsHalfExtents,
			MinCellSize:       demoVoxelResolution,
			Inset:             0.0,
			DebugName:         "decorative_pool",
			Color:             [3]float32{0.12, 0.64, 0.54},
			AbsorptionColor:   [3]float32{0.04, 0.14, 0.1},
			Opacity:           0.42,
			Roughness:         0.1,
			Refraction:        0.18,
			FlowDirection:     [2]float32{0.5, 0.2},
			FlowSpeed:         0.22,
			WaveAmplitude:     0.01,
		},
		&WaterSplashEffectComponent{
			Texture:        state.ParticleAtlas,
			AtlasCols:      4,
			AtlasRows:      4,
			SplashSprite:   5,
			SpraySprite:    9,
			FlashSprite:    10,
			MinImpactSpeed: 2.0,
			StrengthScale:  0.8,
		},
	)
}

func addPoolStructure(cmd *Commands, state *DemoState, center mgl32.Vec3, plinthModel, floorModel, frameModel AssetId) {
	cmd.AddEntity(
		&TransformComponent{Position: center.Add(mgl32.Vec3{0, -4.35, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1.4, 1}},
		&VoxelModelComponent{VoxelModel: plinthModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.8, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: center.Add(mgl32.Vec3{0, -5.1, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: floorModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.65, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: center.Add(mgl32.Vec3{0, -1.95, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1.55, 1}},
		&VoxelModelComponent{VoxelModel: frameModel, VoxelPalette: state.WallPalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.7, Restitution: 0.0},
	)
}

func addFallingBodies(cmd *Commands, state *DemoState) {
	addDynamicBody(cmd, state.SphereModel, state.SpherePalette, mgl32.Vec3{-4.8, 11.5, -4.8}, mgl32.Vec3{0.4, -0.2, 0.3}, mgl32.Vec3{0.95, 0.95, 0.95}, 0.8, 0.14)
	addDynamicBody(cmd, state.CubeModel, state.CubePalette, mgl32.Vec3{-1.2, 14.0, 0.8}, mgl32.Vec3{0.2, -0.3, -0.1}, mgl32.Vec3{1, 1, 1}, 1.3, 0.08)
	addDynamicBody(cmd, state.SphereModel, state.HeavyCubePalette, mgl32.Vec3{2.3, 18.0, -3.5}, mgl32.Vec3{-0.5, -0.1, 0.2}, mgl32.Vec3{1.25, 1.25, 1.25}, 2.4, 0.04)
	addDynamicBody(cmd, state.TallCubeModel, state.CubePalette, mgl32.Vec3{5.1, 13.0, 1.6}, mgl32.Vec3{-0.25, -0.4, -0.2}, mgl32.Vec3{0.9, 0.9, 0.9}, 1.1, 0.18)
	addDynamicBody(cmd, state.CubeModel, state.HeavyCubePalette, mgl32.Vec3{8.0, 16.4, -1.4}, mgl32.Vec3{-0.9, -0.2, 0.4}, mgl32.Vec3{1.3, 1.3, 1.3}, 3.1, 0.03)
}

func addDynamicBody(cmd *Commands, model AssetId, palette AssetId, position, velocity, scale mgl32.Vec3, mass, restitution float32) {
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: scale},
		&VoxelModelComponent{VoxelModel: model, VoxelPalette: palette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{
			Mass:            mass,
			GravityScale:    1.0,
			Velocity:        velocity,
			AngularVelocity: mgl32.Vec3{0.8 * mass, 1.2, 0.9},
		},
		&ColliderComponent{Friction: 0.35, Restitution: restitution},
	)
}

func addLights(cmd *Commands) {
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{14, 18, 16},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-42), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(28), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{Type: LightTypeDirectional, Intensity: 1.15, Color: [3]float32{1.0, 0.97, 0.93}, Range: 500, CastsShadows: true},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mainPoolCenter.Add(mgl32.Vec3{0, 4.8, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.65, Color: [3]float32{0.36, 0.66, 1.0}, Range: 18},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mainPoolCenter.Add(mgl32.Vec3{5, 3.6, 4}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.2, Color: [3]float32{0.2, 0.46, 0.92}, Range: 12},
	)
	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter.Add(mgl32.Vec3{0, 3.2, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.15, Color: [3]float32{0.18, 0.9, 0.78}, Range: 10},
	)
}

func addCamera(cmd *Commands) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{5.5, 8.8, 21}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: mgl32.Vec3{5.5, 8.8, 21},
			LookAt:   mgl32.Vec3{-2.5, 2.8, -0.5},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      -20.5,
			Pitch:    -14.5,
			Fov:      58,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      1000,
		},
		&FlyingCameraComponent{Speed: 12, Sensitivity: 0.1},
	)
}

func spawnProjectileSystem(cmd *Commands, input *Input, rtState *VoxelRtState, state *DemoState) {
	if !input.JustPressed[MouseButtonLeft] {
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

	origin, rayDir := rtState.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
	launchDir := rayDir
	if launchDir.Len() == 0 {
		launchDir = mgl32.Vec3{0, 0, -1}
	}
	launchDir = launchDir.Normalize()

	cmd.AddEntity(
		&TransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.9, 0.9, 0.9}},
		&VoxelModelComponent{VoxelModel: state.SphereModel, VoxelPalette: state.ProjectilePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{
			Mass:            1.0,
			GravityScale:    1.0,
			Velocity:        launchDir.Mul(18),
			AngularVelocity: mgl32.Vec3{0.5, 1.3, 0.8},
		},
		&ColliderComponent{Friction: 0.28, Restitution: 0.3},
		&LifetimeComponent{TimeLeft: 18},
	)
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
