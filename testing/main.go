package main

import (
	"fmt"
	"math"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

type TestModule struct{}

func (TestModule) Install(app *App, cmd *Commands) {
	// Add systems
	app.UseSystem(
		System(startup).
			InStage(PreUpdate).
			InState(OnEnter(Startup)),
	)
	app.UseSystem(
		System(updateCamera).
			InStage(Update).
			InAnyState(),
	)

	// Custom stage
	testStage := Stage{
		Name:       "Very Last Stage",
		UpdateType: FixedUpdate,
	}

	app.UseStage(testStage, AfterStage(Finale))
}

var cubeVertices = [...]mgl32.Vec3{
	// top (0, 0, 1)
	{-1, -1, 1},
	{1, -1, 1},
	{1, 1, 1},
	{-1, 1, 1},
	// bottom (0, 0, -1)
	{-1, 1, -1},
	{1, 1, -1},
	{1, -1, -1},
	{-1, -1, -1},
	// right (1, 0, 0)
	{1, -1, -1},
	{1, 1, -1},
	{1, 1, 1},
	{1, -1, 1},
	// left (-1, 0, 0)
	{-1, -1, 1},
	{-1, 1, 1},
	{-1, 1, -1},
	{-1, -1, -1},
	// front (0, 1, 0)
	{1, 1, -1},
	{-1, 1, -1},
	{-1, 1, 1},
	{1, 1, 1},
	// back (0, -1, 0)
	{1, -1, 1},
	{-1, -1, 1},
	{-1, -1, -1},
	{1, -1, -1},
}

var cubeIndices = [...]uint16{
	0, 1, 2, 2, 3, 0, // top
	4, 5, 6, 6, 7, 4, // bottom
	8, 9, 10, 10, 11, 8, // right
	12, 13, 14, 14, 15, 12, // left
	16, 17, 18, 18, 19, 16, // front
	20, 21, 22, 22, 23, 20, // back
}

func startup(cmd *Commands, assets *AssetServer, state *WindowState) {
	cmd.AddEntity(
		assets.LoadMesh(cubeVertices[:], cubeIndices[:]),
		assets.LoadMaterial("assets/shader.wgsl"),
		CameraComponent{
			Position: mgl32.Vec3{1.5, 4, 5},
			LookAt:   mgl32.Vec3{0, 0, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Fov:      math.Pi / 4,
			Aspect:   float32(state.WindowWidth) / float32(state.WindowHeight),
			Near:     1,
			Far:      100,
		},
		TransformComponent{
			Position: mgl32.Vec3{0, 0, 0},
			Rotation: 0.0,
			Scale:    mgl32.Vec3{1, 1, 1},
		},
	)
}

func updateCamera(cmd *Commands, input *Input) {
	fmt.Println("trying to query camera")
	MakeQuery1[CameraComponent](cmd).Map1(
		func(entityId EntityId, camera *CameraComponent) bool {
			var x float32 = 0.0
			var y float32 = 0.0
			var z float32 = 0.0
			if input.Pressed[KeyRight] {
				x = x + 0.1
			}
			if input.Pressed[KeyLeft] {
				x = x - 0.1
			}
			if input.Pressed[KeyDown] {
				y = y - 0.1
			}
			if input.Pressed[KeyUp] {
				y = y + 0.1
			}
			if input.Pressed[KeySpace] {
				z = z + 0.1
			}
			if input.Pressed[KeyEnter] {
				z = z - 0.1
			}
			update := mgl32.Vec3{x, y, z}
			camera.Position = camera.Position.Add(update)
			return true
		})
}

func main() {
	app := NewApp().
		UseStates(Startup, Quit).
		UseModules(TimeModule{}).
		UseModules(AssetServerModule{}).
		UseModules(InputModule{}).
		UseModules(ClientModule{
			WindowWidth:  1024,
			WindowHeight: 768,
			WindowTitle:  "Testing App",
		}).
		UseModules(TestModule{})

	app.Run()
}
