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
	app.UseSystem(System(system1).InStage(Update).InState(OnEnter(Startup)))
	app.UseSystem(System(system2).InStage(Update).InState(OnExecute(Startup)))
	app.UseSystem(System(system3).InStage(Update).InState(OnExit(Startup)))
	app.UseSystem(System(updateCamera).InStage(Update).RunAlways())

	// Custom stage
	testStage := Stage{
		Name:       "Very Last Stage",
		UpdateType: FixedUpdate,
	}

	app.UseStage(testStage, AfterStage(Finale))
	app.UseSystem(System(system4).InStage(testStage).InState(OnEnter(Startup)))

	app.UseSystem(System(startup).InStage(PreUpdate).InState(OnEnter(Startup)))
}

type Position struct {
	x float32
	y float32
	z float32
}

type Velocity struct {
	dx float32
	dy float32
}

type Name struct {
	name string
}

type X struct {
	seconds int
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

var cubeIndices = [...]int{
	0, 1, 2, 2, 3, 0, // top
	4, 5, 6, 6, 7, 4, // bottom
	8, 9, 10, 10, 11, 8, // right
	12, 13, 14, 14, 15, 12, // left
	16, 17, 18, 18, 19, 16, // front
	20, 21, 22, 22, 23, 20, // back
}

func startup(cmd *Commands, assets *AssetServer, state *WindowState) {
	fmt.Println("startup system")

	cmd.AddEntity(
		assets.LoadMesh(cubeVertices[:], cubeIndices[:]),
		assets.LoadMaterial("assets/shader.wgsl"),
		CameraComponent{
			Position:  mgl32.Vec3{0, 0, 0},
			Direction: mgl32.Vec3{0, 0, 0},
			Up:        mgl32.Vec3{0, 0, 0},
			Fov:       math.Pi / 4,
			Aspect:    float32(state.WindowWidth) / float32(state.WindowHeight),
		},
		TransformComponent{
			Position: mgl32.Vec3{0, 0, 0},
			Rotation: 0.0,
			Scale:    mgl32.Vec3{0, 0, 0},
		},
	)
}

func system1() {
	fmt.Println("system1")
}
func system2(cmd *Commands) {
	fmt.Println("system2")
}
func system3(cmd *Commands, time *Time) {
	fmt.Println("system3")

	MakeQuery2[Position, Velocity](cmd).Map2(func(entityId EntityId, pos *Position, vel *Velocity) bool {
		fmt.Printf("pos %v vel %v\n", *pos, *vel)

		return true
	})

	MakeQuery1[Name](cmd).Map1(func(entityId EntityId, name *Name) bool {
		fmt.Printf("name %v\n", *name)

		return true
	})
}

func updateCamera(cmd *Commands, input *Input) {
	fmt.Println("trying to query camera")
	MakeQuery1[CameraComponent](cmd).Map1(
		func(entityId EntityId, camera *CameraComponent) bool {
			var x float32 = 0.0
			var y float32 = 0.0
			var z float32 = 0.0
			if input.Pressed[KeyRight] {
				fmt.Println("ПРАВО РУЛЯ!")
				x = x + 0.1
			}
			if input.Pressed[KeyLeft] {
				fmt.Println("ЛЕВО РУЛЯ!")
				x = x - 0.1
			}
			if input.Pressed[KeyDown] {
				fmt.Println("ТАБАНЬ!")
				y = y - 0.1
			}
			if input.Pressed[KeyUp] {
				fmt.Println("ПОЛНЫЙ ВПЕРЁД!")
				y = y - 0.1
			}
			if input.Pressed[KeySpace] {
				fmt.Println("ВВЕРХ!")
				z = z + 0.1
			}
			if input.Pressed[KeyEnter] {
				fmt.Println("ВНИЗ!")
				z = z - 0.1
			}
			update := mgl32.Vec3{x, y, z}
			camera.Position = camera.Position.Add(update)
			return true
		})
}

func system4() {
	fmt.Println("system4")
}

func main() {
	app := NewApp().
		UseStates(Startup, Quit).
		UseModules(TimeModule{}).
		UseModules(AssetServerModule{}).
		UseModules(InputModule{}).
		UseModules(ClientModule{
			WindowWidth:  640,
			WindowHeight: 480,
			WindowTitle:  "Testing App",
		}).
		UseModules(TestModule{})

	cmd := app.Commands()
	cmd.AddEntity(&Position{x: 1, y: 2, z: 3}, &Velocity{dx: 69, dy: 420})
	cmd.AddEntity(&Name{name: "hello"})

	cmd.AddResources(&X{seconds: 69})

	app.Run()
}
