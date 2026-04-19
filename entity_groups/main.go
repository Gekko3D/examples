package main

import (
	"fmt"
	"reflect"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

var (
	alphaSystemGroup = EntityGroupKey{Kind: "system", ID: "alpha"}
	betaSystemGroup  = EntityGroupKey{Kind: "system", ID: "beta"}
	onlineBubble     = EntityGroupKey{Kind: "bubble", ID: "online"}
	typeOfText       = reflect.TypeOf(TextComponent{})
)

type DemoModule struct{}

type DemoState struct {
	HudEntity   EntityId
	LastAction  string
	SpawnCycles int
}

func (m DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{
		LastAction: "startup",
	})

	app.UseSystem(
		System(setupSystem).
			InStage(Prelude).
			InState(OnEnter(Startup)),
	)
	app.UseSystem(
		System(groupControlSystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(updateHudSystem).
			InStage(PostUpdate).
			RunAlways(),
	)
	app.UseSystem(
		System(quitSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
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
			WindowTitle:  "Entity Groups",
		},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func setupSystem(cmd *Commands, assets *AssetServer, state *DemoState) {
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 8, 22},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 8, 22},
			LookAt:   mgl32.Vec3{0, 2, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Fov:      52,
			Aspect:   1280.0 / 720.0,
			Near:     0.1,
			Far:      500,
		},
		&FlyingCameraComponent{Speed: 10, Sensitivity: 0.1},
	)

	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.3,
			Color:     [3]float32{1, 1, 1},
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{12, 40, 18},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:      LightTypeDirectional,
			Intensity: 0.8,
			Color:     [3]float32{1, 0.95, 0.9},
			Range:     500,
		},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.15, 0.24, 0.34},
		ColorB:     mgl32.Vec3{0.03, 0.05, 0.1},
		Opacity:    1,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})

	floorPalette := assets.CreateSimplePalette([4]uint8{90, 96, 108, 255})
	floorModel := assets.CreateCubeModel(48, 1, 48, 1)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-24, -1, -24},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      floorModel,
			VoxelPalette:    floorPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: 1,
		},
	)

	state.HudEntity = cmd.AddEntity(&TextComponent{
		Text:     "Entity Groups Demo",
		Position: [2]float32{20, 20},
		Scale:    0.85,
		Color:    [4]float32{1, 1, 1, 1},
	})

	spawnSystemGroup(cmd, assets, state, alphaSystemGroup, mgl32.Vec3{-8, 0, 0}, [4]uint8{80, 160, 255, 255})
	spawnSystemGroup(cmd, assets, state, betaSystemGroup, mgl32.Vec3{8, 0, 0}, [4]uint8{255, 140, 90, 255})
	printGroupInstructions()
}

func groupControlSystem(cmd *Commands, input *Input, assets *AssetServer, state *DemoState) {
	if input == nil || assets == nil || state == nil {
		return
	}

	if input.JustPressed[Key1] {
		spawnIfMissing(cmd, assets, state, alphaSystemGroup, mgl32.Vec3{-8, 0, 0}, [4]uint8{80, 160, 255, 255})
	}
	if input.JustPressed[Key2] {
		removeGroup(cmd, state, alphaSystemGroup)
	}
	if input.JustPressed[Key3] {
		spawnIfMissing(cmd, assets, state, betaSystemGroup, mgl32.Vec3{8, 0, 0}, [4]uint8{255, 140, 90, 255})
	}
	if input.JustPressed[Key4] {
		removeGroup(cmd, state, betaSystemGroup)
	}
	if input.JustPressed[KeyR] {
		removeGroup(cmd, state, alphaSystemGroup)
		removeGroup(cmd, state, betaSystemGroup)
		spawnSystemGroup(cmd, assets, state, alphaSystemGroup, mgl32.Vec3{-8, 0, 0}, [4]uint8{80, 160, 255, 255})
		spawnSystemGroup(cmd, assets, state, betaSystemGroup, mgl32.Vec3{8, 0, 0}, [4]uint8{255, 140, 90, 255})
		state.LastAction = "queued reset for alpha and beta"
	}
}

func updateHudSystem(cmd *Commands, state *DemoState) {
	if state == nil || state.HudEntity == 0 {
		return
	}

	textAny := cmd.GetComponent(state.HudEntity, typeOfText)
	text, ok := textAny.(*TextComponent)
	if !ok || text == nil {
		return
	}

	text.Text = fmt.Sprintf(
		"Entity Groups Demo\n\n1 spawn alpha system\n2 despawn alpha system\n3 spawn beta system\n4 despawn beta system\nR reset both groups\nESC quit\n\nlive groups:\n  alpha:  %d\n  beta:   %d\n  bubble: %d\n\nlast action:\n%s",
		len(cmd.GetEntitiesInGroup(alphaSystemGroup)),
		len(cmd.GetEntitiesInGroup(betaSystemGroup)),
		len(cmd.GetEntitiesInGroup(onlineBubble)),
		state.LastAction,
	)
}

func quitSystem(cmd *Commands, input *Input) {
	if input != nil && input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}

func spawnIfMissing(cmd *Commands, assets *AssetServer, state *DemoState, systemGroup EntityGroupKey, anchor mgl32.Vec3, color [4]uint8) {
	if len(cmd.GetEntitiesInGroup(systemGroup)) > 0 {
		state.LastAction = "spawn skipped for " + systemGroup.ID + " (already live)"
		return
	}
	spawnSystemGroup(cmd, assets, state, systemGroup, anchor, color)
}

func spawnSystemGroup(cmd *Commands, assets *AssetServer, state *DemoState, systemGroup EntityGroupKey, anchor mgl32.Vec3, color [4]uint8) {
	palette := assets.CreateSimplePalette(color)
	cubeModel := assets.CreateCubeModel(4, 4, 4, 1)
	columns := []mgl32.Vec3{
		anchor.Add(mgl32.Vec3{-4, 2, 0}),
		anchor.Add(mgl32.Vec3{0, 4, 0}),
		anchor.Add(mgl32.Vec3{4, 6, 0}),
	}

	for _, position := range columns {
		cmd.AddEntityInGroups(
			[]EntityGroupKey{systemGroup, onlineBubble},
			&TransformComponent{
				Position: position,
				Rotation: mgl32.QuatIdent(),
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&VoxelModelComponent{
				VoxelModel:      cubeModel,
				VoxelPalette:    palette,
				VoxelResolution: 1,
			},
		)
	}

	state.SpawnCycles++
	state.LastAction = fmt.Sprintf("queued spawn for %s (cycle %d)", systemGroup.ID, state.SpawnCycles)
}

func removeGroup(cmd *Commands, state *DemoState, group EntityGroupKey) {
	removed := cmd.RemoveEntitiesInGroup(group)
	state.LastAction = fmt.Sprintf("queued removal for %s (%d entities)", group.ID, len(removed))
}

func printGroupInstructions() {
	fmt.Println("Entity Groups example")
	fmt.Println("  1 spawn alpha system group")
	fmt.Println("  2 despawn alpha system group")
	fmt.Println("  3 spawn beta system group")
	fmt.Println("  4 despawn beta system group")
	fmt.Println("  R reset both groups")
	fmt.Println("  ESC quit")
}
