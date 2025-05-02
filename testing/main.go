package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
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

func startup(cmd *Commands, assets *AssetServer) {
	fmt.Println("startup system")

	cmd.AddEntity(
		assets.LoadMesh("hleb.3d"),
		assets.LoadMaterial("hleb.wgl", Vec3{X: 1.0, Y: 1.0, Z: 1.0}),
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

func system4() {
	fmt.Println("system4")
}

func main() {
	app := NewApp().
		UseStates(Startup, Quit).
		UseModules(TimeModule{}).
		UseModules(AssetServerModule{}).
		UseModules(ClientModule{
			WindowWidth:  1024,
			WindowHeight: 768,
			WindowTitle:  "Testing App",
		}).
		UseModules(TestModule{})

	cmd := app.Commands()
	cmd.AddEntity(&Position{x: 1, y: 2, z: 3}, &Velocity{dx: 69, dy: 420})
	cmd.AddEntity(&Name{name: "hello"})

	cmd.AddResources(&X{seconds: 69})

	app.Run()
}
