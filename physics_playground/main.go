package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const demoVoxelResolution float32 = 0.2

const (
	Startup State = iota
	Quit
)

type DemoModule struct{}

type ToolMode int

const (
	ToolModeLauncher ToolMode = iota
	ToolModeGrabber
)

type DemoState struct {
	FloorModel     AssetId
	CubeModel      AssetId
	SphereModel    AssetId
	RampModel      AssetId
	GreyPalette    AssetId
	RedPalette     AssetId
	BluePalette    AssetId
	ToolMode       ToolMode
	GrabbedEntity  EntityId
	GrabDistance   float32
	GrabOffset     mgl32.Vec3
	GrabUsesTorque bool
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
			WindowHeight: 720,
			WindowTitle:  "Physics",
		},
		PhysicsModule{Synchronous: true},
		VoxPhysicsModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		DemoModule{},
	)

	app.Run()
}

func (m DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(toolModeSystem).InStage(Update))
	app.UseSystem(System(toolInteractionSystem).InStage(Update))
	app.UseSystem(System(grabberSystem).InStage(Update))
	app.UseSystem(System(forceFieldSystem).InStage(Update))
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	// Assets Setup
	state.FloorModel = assets.CreateCubeModel(150, 3, 150, 1.0)
	state.CubeModel = assets.CreateCubeModel(6, 6, 6, 1.0)
	state.SphereModel = assets.CreateSphereModel(3, 1.0)
	state.RampModel = assets.CreateRampModel(28, 12, 18, 1.0)

	state.GreyPalette = assets.CreateSimplePalette([4]uint8{45, 47, 52, 255})
	state.RedPalette = assets.CreateSimplePalette([4]uint8{255, 50, 50, 255})
	state.BluePalette = assets.CreateSimplePalette([4]uint8{50, 50, 255, 255})
	state.ToolMode = ToolModeLauncher

	// 1. Large static floor
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-15, -0.2, -15}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: state.FloorModel, VoxelPalette: state.GreyPalette, PivotMode: PivotModeCorner, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.5, Restitution: 0.2},
	)

	// 2. A ramp
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-6.8, -0.1, -1.8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: state.RampModel, VoxelPalette: state.GreyPalette, PivotMode: PivotModeCorner, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{IsStatic: true, Mass: 0},
		&ColliderComponent{Friction: 0.5, Restitution: 0.2},
	)

	// 3. Stack of cubes
	for i := 0; i < 4; i++ {
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{-8.6, float32(0.6 + float32(i)*1.4), 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&VoxelModelComponent{VoxelModel: state.CubeModel, VoxelPalette: state.BluePalette, VoxelResolution: demoVoxelResolution},
			&RigidBodyComponent{Mass: 1.0, GravityScale: 1.0},
			&ColliderComponent{Friction: 0.4, Restitution: 0.3},
		)
	}

	// 4. Several spheres at height
	masses := []float32{0.5, 1.0, 2.0}
	for i, mass := range masses {
		cmd.AddEntity(
			&TransformComponent{Position: mgl32.Vec3{2.8 + float32(i)*2.0, float32(3.6 + float32(i)*1.6), -1.2 + float32(i)*1.0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&VoxelModelComponent{VoxelModel: state.SphereModel, VoxelPalette: state.BluePalette, VoxelResolution: demoVoxelResolution},
			&RigidBodyComponent{Mass: mass, GravityScale: 1.0},
			&ColliderComponent{Friction: 0.3, Restitution: 0.6},
		)
	}

	// 5. Heavy cube
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-2.4, 3.0, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: state.CubeModel, VoxelPalette: state.RedPalette, VoxelResolution: demoVoxelResolution},
		&RigidBodyComponent{Mass: 5.0, GravityScale: 1.0, Velocity: mgl32.Vec3{-2.2, 0, 0}},
		&ColliderComponent{Friction: 0.4, Restitution: 0.1},
	)

	// 6. Lights
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-4, 8, 5},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-28), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(30), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{Type: LightTypeDirectional, Intensity: 1.0, Color: [3]float32{1, 0.97, 0.92}, Range: 500, CastsShadows: true},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{3.2, 5.2, 3.6}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 8.0, Color: [3]float32{1, 0.82, 0.68}, Range: 28},
	)

	// 7. Camera
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{14.4, 8.4, 15.6}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: mgl32.Vec3{14.4, 8.4, 15.6},
			LookAt:   mgl32.Vec3{-4.4, 2.8, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      -50,
			Pitch:    -15,
			Fov:      60,
			Aspect:   1280.0 / 720.0,
			Near:     0.1,
			Far:      1000,
		},
		&FlyingCameraComponent{Speed: 18, Sensitivity: 0.1},
	)

	fmt.Println("Physics Playground Scene Setup Complete")
	fmt.Println("Press 1 for launcher, 2 for grabber")
}

func toolModeSystem(input *Input, state *DemoState) {
	if input.JustPressed[Key1] {
		state.ToolMode = ToolModeLauncher
		state.GrabbedEntity = 0
		fmt.Println("Launcher tool enabled")
	}

	if input.JustPressed[Key2] {
		state.ToolMode = ToolModeGrabber
		state.GrabbedEntity = 0
		fmt.Println("Grabber tool enabled")
	}
}

func toolInteractionSystem(cmd *Commands, input *Input, rtState *VoxelRtState, state *DemoState) {
	if state.ToolMode == ToolModeLauncher {
		if input.JustPressed[MouseButtonLeft] {
			var cam *CameraComponent
			MakeQuery1[CameraComponent](cmd).Map(func(eid EntityId, c *CameraComponent) bool {
				cam = c
				return false
			})

			if cam != nil {
				origin, dir := rtState.ScreenToWorldRay(input.MouseX, input.MouseY, cam)

				cmd.AddEntity(
					&TransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
					&VoxelModelComponent{VoxelModel: state.SphereModel, VoxelPalette: state.RedPalette, VoxelResolution: demoVoxelResolution},
					&RigidBodyComponent{
						Mass:         1.0,
						GravityScale: 1.0,
						Velocity:     dir.Mul(9),
					},
					&ColliderComponent{Friction: 0.3, Restitution: 0.6},
					&LifetimeComponent{TimeLeft: 15},
				)
				fmt.Println("Spawned sphere projectile")
			}
		}
		return
	}

	if state.ToolMode != ToolModeGrabber {
		return
	}

	if input.JustReleased[MouseButtonLeft] {
		if state.GrabbedEntity != 0 {
			fmt.Println("Released entity")
		}
		state.GrabbedEntity = 0
		return
	}

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

	origin, dir := rtState.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
	result := rtState.Raycast(origin, dir, 1000)
	if !result.Hit {
		return
	}

	found := false
	MakeQuery3[TransformComponent, RigidBodyComponent, VoxelModelComponent](cmd).Map(func(eid EntityId, transform *TransformComponent, rb *RigidBodyComponent, vm *VoxelModelComponent) bool {
		if eid != result.Entity {
			return true
		}

		if rb.IsStatic {
			return false
		}

		state.GrabbedEntity = result.Entity
		state.GrabDistance = result.T
		hitPos := origin.Add(dir.Mul(result.T))
		state.GrabOffset = transform.Rotation.Conjugate().Rotate(hitPos.Sub(transform.Position))
		state.GrabUsesTorque = vm == nil || vm.VoxelModel != state.SphereModel
		if !state.GrabUsesTorque {
			state.GrabOffset = mgl32.Vec3{}
			rb.AngularVelocity = mgl32.Vec3{}
		}
		fmt.Printf("Grabbed entity %v\n", result.Entity)
		found = true
		return false
	})

	if !found {
		fmt.Printf("Clicked object %v is not a dynamic rigid body\n", result.Entity)
	}
}

func grabberSystem(cmd *Commands, input *Input, rtState *VoxelRtState, state *DemoState, time *Time) {
	if state.ToolMode != ToolModeGrabber || state.GrabbedEntity == 0 {
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

	found := false
	MakeQuery2[TransformComponent, RigidBodyComponent](cmd).Map(func(eid EntityId, transform *TransformComponent, rb *RigidBodyComponent) bool {
		if eid != state.GrabbedEntity {
			return true
		}

		origin, dir := rtState.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
		targetPos := origin.Add(dir.Mul(state.GrabDistance))

		rWorld := transform.Rotation.Rotate(state.GrabOffset)
		currentGrabPos := transform.Position.Add(rWorld)

		stiffness := float32(300.0)
		damping := float32(20.0)
		if rb.Mass > 0 {
			stiffness *= rb.Mass
			damping *= rb.Mass
		}

		diff := targetPos.Sub(currentGrabPos)
		velAtPoint := rb.Velocity.Add(rb.AngularVelocity.Cross(rWorld))
		force := diff.Mul(stiffness).Sub(velAtPoint.Mul(damping))

		rb.ApplyImpulse(force.Mul(float32(time.Dt)))
		if state.GrabUsesTorque {
			rb.ApplyTorque(rWorld.Cross(force).Mul(float32(time.Dt)))
		} else {
			rb.AngularVelocity = mgl32.Vec3{}
		}

		rb.Velocity = rb.Velocity.Mul(0.98)
		if state.GrabUsesTorque {
			rb.AngularVelocity = rb.AngularVelocity.Mul(0.95)
		}
		rb.AngularVelocity = clampVec3Len(rb.AngularVelocity, 10)

		found = true
		return false
	})

	if !found {
		state.GrabbedEntity = 0
	}
}

func clampVec3Len(v mgl32.Vec3, maxLen float32) mgl32.Vec3 {
	if maxLen <= 0 {
		return mgl32.Vec3{}
	}
	len := v.Len()
	if len <= maxLen || len == 0 {
		return v
	}
	return v.Mul(maxLen / len)
}

func forceFieldSystem(cmd *Commands, input *Input) {
	if input.JustPressed[KeyF] {
		MakeQuery1[RigidBodyComponent](cmd).Map(func(eid EntityId, rb *RigidBodyComponent) bool {
			if !rb.IsStatic {
				rb.ApplyImpulse(mgl32.Vec3{0, 2.2, 0})
			}
			return true
		})
		fmt.Println("Force field activated!")
	}
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
