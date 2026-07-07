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

	footprintChannelSurfaceY float32 = 2.15
	footprintChannelGroup            = "water_world:footprint_channel"
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
	AmberLightPalette AssetId
	CyanLightPalette  AssetId
	PinkLightPalette  AssetId
	GreenLightPalette AssetId
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
	state.AmberLightPalette = assets.CreatePBRPalette([4]uint8{255, 190, 96, 255}, 0.2, 0, 3.8, 1.45)
	state.CyanLightPalette = assets.CreatePBRPalette([4]uint8{88, 210, 255, 255}, 0.18, 0, 3.4, 1.45)
	state.PinkLightPalette = assets.CreatePBRPalette([4]uint8{255, 112, 220, 255}, 0.18, 0, 3.2, 1.45)
	state.GreenLightPalette = assets.CreatePBRPalette([4]uint8{110, 255, 180, 255}, 0.2, 0, 3.1, 1.45)

	addWorldFloor(cmd, state)
	addMainPool(cmd, state)
	addDecorativePool(cmd, state)
	addFootprintChannel(cmd, state)
	addFallingBodies(cmd, state)
	addWakeSkimmers(cmd, state)
	addLights(cmd, state)
	addCamera(cmd)

	fmt.Println("Water World scene ready")
	fmt.Println("Right-side segmented channel uses continuity-group footprint water")
	fmt.Println("Left click to launch a camera-aimed projectile")
	fmt.Println("WASD + mouse to fly, Escape to quit")
}

func addWorldFloor(cmd *Commands, state *DemoState) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, -1.2, 2}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: state.GroundModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
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
			Overlap:           0.35,
			DebugName:         "main_pool",
			Color:             [3]float32{0.16, 0.36, 0.64},
			AbsorptionColor:   [3]float32{0.18, 0.28, 0.48},
			Opacity:           0.5,
			Roughness:         0.14,
			Refraction:        0.16,
			FlowDirection:     [2]float32{1, 0.3},
			FlowSpeed:         0.8,
			WaveAmplitude:     0.025,
			VisualCellSize:    0.22,
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
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
		&ColliderComponent{Friction: 0.8, Restitution: 0.0},
	)

	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter.Add(mgl32.Vec3{0, -1.6, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.55, 1, 0.55}},
		&VoxelModelComponent{VoxelModel: state.PoolFloorModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
		&ColliderComponent{Friction: 0.65, Restitution: 0.0},
	)

	cmd.AddEntity(
		&TransformComponent{Position: decorativePoolCenter.Add(mgl32.Vec3{0, -1.4, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.55, 1, 0.55}},
		&VoxelModelComponent{VoxelModel: state.PoolFrameModel, VoxelPalette: state.WallPalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
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
			Overlap:           0.22,
			DebugName:         "decorative_pool",
			Color:             [3]float32{0.12, 0.64, 0.54},
			AbsorptionColor:   [3]float32{0.04, 0.14, 0.1},
			Opacity:           0.42,
			Roughness:         0.1,
			Refraction:        0.18,
			FlowDirection:     [2]float32{0.5, 0.2},
			FlowSpeed:         0.22,
			WaveAmplitude:     0.01,
			VisualCellSize:    0.16,
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

func addFootprintChannel(cmd *Commands, state *DemoState) {
	channel := []struct {
		center      mgl32.Vec3
		halfExtents [2]float32
	}{
		{center: mgl32.Vec3{10.0, footprintChannelSurfaceY, 9.0}, halfExtents: [2]float32{4.2, 1.1}},
		{center: mgl32.Vec3{14.2, footprintChannelSurfaceY, 11.8}, halfExtents: [2]float32{1.1, 3.9}},
		{center: mgl32.Vec3{17.1, footprintChannelSurfaceY, 15.7}, halfExtents: [2]float32{4.0, 1.1}},
	}

	for i, segment := range channel {
		addChannelStoneSlab(cmd, state, segment.center, segment.halfExtents)
		cmd.AddEntity(
			&TransformComponent{Position: segment.center, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&WaterBodyComponent{
				Mode:            WaterBodyModeExplicitRect,
				SurfaceY:        segment.center.Y(),
				Depth:           1.25,
				RectHalfExtents: segment.halfExtents,
				ContinuityGroup: footprintChannelGroup,
				DebugName:       fmt.Sprintf("footprint_channel_%d", i),
				Color:           [3]float32{0.1, 0.47, 0.7},
				AbsorptionColor: [3]float32{0.07, 0.26, 0.36},
				Opacity:         0.48,
				Roughness:       0.12,
				Refraction:      0.2,
				FlowDirection:   [2]float32{0.95, 0.32},
				FlowSpeed:       0.55,
				WaveAmplitude:   0.018,
				VisualCellSize:  0.18,
			},
			&WaterSplashEffectComponent{
				Texture:        state.ParticleAtlas,
				AtlasCols:      4,
				AtlasRows:      4,
				SplashSprite:   5,
				SpraySprite:    9,
				FlashSprite:    10,
				MinImpactSpeed: 1.8,
				StrengthScale:  0.75,
			},
		)
	}

	addSkimmerBody(cmd, state.SphereModel, state.CyanLightPalette, mgl32.Vec3{7.1, footprintChannelSurfaceY + 0.42, 9.0}, mgl32.Vec3{6.4, -0.08, 2.6}, 0.5)
}

func addChannelStoneSlab(cmd *Commands, state *DemoState, center mgl32.Vec3, halfExtents [2]float32) {
	cmd.AddEntity(
		&TransformComponent{
			Position: center.Add(mgl32.Vec3{0, -1.42, 0}),
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{halfExtents[0] * 2.15 / 1.2, 0.45, halfExtents[1] * 2.15 / 1.2},
		},
		&VoxelModelComponent{VoxelModel: state.CubeModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
		&ColliderComponent{Friction: 0.75, Restitution: 0.0},
	)
}

func addPoolStructure(cmd *Commands, state *DemoState, center mgl32.Vec3, plinthModel, floorModel, frameModel AssetId) {
	cmd.AddEntity(
		&TransformComponent{Position: center.Add(mgl32.Vec3{0, -4.35, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1.4, 1}},
		&VoxelModelComponent{VoxelModel: plinthModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
		&ColliderComponent{Friction: 0.8, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: center.Add(mgl32.Vec3{0, -5.1, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: floorModel, VoxelPalette: state.StonePalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
		&ColliderComponent{Friction: 0.65, Restitution: 0.0},
	)
	cmd.AddEntity(
		&TransformComponent{Position: center.Add(mgl32.Vec3{0, -1.95, 0}), Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1.55, 1}},
		&VoxelModelComponent{VoxelModel: frameModel, VoxelPalette: state.WallPalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{BodyMode: BodyModeStatic, Mass: 0},
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

func addWakeSkimmers(cmd *Commands, state *DemoState) {
	addSkimmerBody(cmd, state.SphereModel, state.ProjectilePalette, mainPoolCenter.Add(mgl32.Vec3{-7.4, 0.36, -4.6}), mgl32.Vec3{7.8, -0.15, 1.0}, 0.62)
	addSkimmerBody(cmd, state.SphereModel, state.SpherePalette, mainPoolCenter.Add(mgl32.Vec3{6.8, 0.42, 3.8}), mgl32.Vec3{-7.2, -0.12, -1.4}, 0.58)
	addSkimmerBody(cmd, state.CubeModel, state.CubePalette, mainPoolCenter.Add(mgl32.Vec3{-6.2, 0.5, 2.4}), mgl32.Vec3{6.2, -0.1, -2.4}, 0.54)
}

func addSkimmerBody(cmd *Commands, model AssetId, palette AssetId, position, velocity mgl32.Vec3, scale float32) {
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{scale, scale, scale}},
		&VoxelModelComponent{VoxelModel: model, VoxelPalette: palette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{
			Mass:            0.6,
			GravityScale:    0.05,
			Velocity:        velocity,
			AngularVelocity: mgl32.Vec3{0.2, 2.0, 0.4},
		},
		&ColliderComponent{Shape: ShapeSphere, Radius: 0.8, Friction: 0.12, Restitution: 0.02},
	)
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

func addLights(cmd *Commands, state *DemoState) {
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

	addFloatingLight(cmd, state, mainPoolCenter.Add(mgl32.Vec3{-5.6, 3.15, -4.3}), state.AmberLightPalette, [3]float32{1.0, 0.58, 0.22}, 3.8, 15.5, 0.46)
	addFloatingLight(cmd, state, mainPoolCenter.Add(mgl32.Vec3{0.4, 4.0, 4.5}), state.CyanLightPalette, [3]float32{0.28, 0.75, 1.0}, 4.1, 17.0, 0.52)
	addFloatingLight(cmd, state, mainPoolCenter.Add(mgl32.Vec3{5.9, 3.35, -1.7}), state.PinkLightPalette, [3]float32{1.0, 0.32, 0.86}, 3.4, 14.5, 0.44)
	addFloatingLight(cmd, state, decorativePoolCenter.Add(mgl32.Vec3{0.2, 2.65, 0.3}), state.GreenLightPalette, [3]float32{0.24, 1.0, 0.68}, 2.8, 10.5, 0.38)
}

func addFloatingLight(cmd *Commands, state *DemoState, position mgl32.Vec3, palette AssetId, color [3]float32, intensity, lightRange, bulbScale float32) {
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{bulbScale, bulbScale, bulbScale}},
		&VoxelModelComponent{VoxelModel: state.SphereModel, VoxelPalette: palette, VoxelResolution: demoVoxelResolution},
	)
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: intensity, Color: color, Range: lightRange},
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
