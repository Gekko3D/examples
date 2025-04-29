package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
)

const (
	Startup State = iota
	Quit
)

type StartupModule struct{}

func (this StartupModule) Install(app *App, commands *Commands) {
	//
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

func main() {
	app := NewAppBuilder().
		UseStates(Startup, Quit).
		UseModule(TimeModule{}).
		UseModule(ClientModule{
			WindowWidth:  1024,
			WindowHeight: 768,
			WindowTitle:  "Testing App",
		}).
		UseModule(StartupModule{}).
		Build()

	cmd := app.Commands()
	cmd.AddEntity(&Position{x: 1, y: 2, z: 3}, &Velocity{dx: 69, dy: 420})
	cmd.AddEntity(&Name{name: "hello"})

	cmd.AddResources(&X{seconds: 69})
	cmd.UseSystem(system1, OnExecute(Startup))
	cmd.UseSystem(system2, OnExecute(Startup))
	cmd.UseSystem(system3, OnExecute(Startup))

	app.Run()
}
