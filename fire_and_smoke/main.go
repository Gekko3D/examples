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

type DemoModule struct{}

type VolumeToggleComponent struct {
	Group int
}

type DemoState struct {
	GroupEnabled     [5]bool
	GroupNames       [5]string
	IntensityEnabled bool
	IntensityEntity  EntityId
	StressBurst      EntityId
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
			WindowTitle:  "Fire and Smoke",
		},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{
		GroupEnabled:     [5]bool{true, true, true, true, true},
		GroupNames:       [5]string{"torch", "campfire", "jet flame", "explosion", "phase5 stress"},
		IntensityEnabled: true,
	})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(toggleVolumesSystem).InStage(Update).RunAlways())
	app.UseSystem(System(stressPulseSystem).InStage(Update).RunAlways())
	app.UseSystem(System(demoHudSystem).InStage(PreRender).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	floorPalette := assets.CreateSimplePalette([4]uint8{32, 34, 40, 255})
	plinthPalette := assets.CreateSimplePalette([4]uint8{54, 52, 50, 255})
	emberPalette := assets.CreateSimplePalette([4]uint8{92, 72, 56, 255})

	floorModel := assets.CreateCubeModel(120, 2, 80, 1.0)
	plinthModel := assets.CreateCubeModel(6, 1.5, 6, 1.0)
	widePlinthModel := assets.CreateCubeModel(8, 1.5, 8, 1.0)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 10, 28},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 10, 28},
			LookAt:   mgl32.Vec3{1, 4.8, 6},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -13,
			Fov:      60,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      800,
		},
		&FlyingCameraComponent{Speed: 12.0, Sensitivity: 0.1},
	)

	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.05,
			Color:     [3]float32{0.7, 0.78, 0.95},
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{12, 22, 14},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-55), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(20), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    0.42,
			Color:        [3]float32{0.72, 0.74, 0.82},
			Range:        800,
			CastsShadows: true,
		},
	)
	cmd.AddEntity(&SkyAmbientComponent{SkyMix: 0.18})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.08, 0.08, 0.11},
		ColorB:     mgl32.Vec3{0.01, 0.01, 0.02},
		Opacity:    1.0,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerStars,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{1.0, 0.96, 0.88},
		ColorB:     mgl32.Vec3{0.55, 0.65, 0.9},
		Opacity:    0.42,
		Priority:   1,
		BlendMode:  SkyboxBlendAdd,
	})

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-60, -2, -40},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      floorModel,
			VoxelPalette:    floorPalette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: 1.0,
		},
	)

	addPlinth(cmd, plinthModel, plinthPalette, mgl32.Vec3{-16, 0, 0}, 6, 6)
	addPlinth(cmd, widePlinthModel, emberPalette, mgl32.Vec3{-4, 0, 0}, 8, 8)
	addPlinth(cmd, plinthModel, plinthPalette, mgl32.Vec3{8, 0, -1}, 6, 6)
	addPlinth(cmd, widePlinthModel, plinthPalette, mgl32.Vec3{18, 0, -4}, 8, 8)
	addPlinth(cmd, plinthModel, emberPalette, mgl32.Vec3{0, 0, 10}, 6, 6)
	addPlinth(cmd, widePlinthModel, emberPalette, mgl32.Vec3{10, 0, 10}, 8, 8)

	addWarmPointLight(cmd, mgl32.Vec3{-16, 6.5, 1}, [3]float32{1.0, 0.68, 0.34}, 1.85, 16)
	addWarmPointLight(cmd, mgl32.Vec3{-4, 4.6, 1.2}, [3]float32{1.0, 0.56, 0.22}, 2.1, 18)
	addWarmPointLight(cmd, mgl32.Vec3{8.8, 3.6, -1}, [3]float32{0.95, 0.7, 0.4}, 1.55, 14)
	addWarmPointLight(cmd, mgl32.Vec3{18, 8, -2.5}, [3]float32{1.0, 0.5, 0.24}, 2.5, 24)
	addWarmPointLight(cmd, mgl32.Vec3{0, 6.2, 11}, [3]float32{1.0, 0.62, 0.28}, 1.8, 15)
	addWarmPointLight(cmd, mgl32.Vec3{10, 6.6, 11.2}, [3]float32{1.0, 0.58, 0.3}, 2.2, 20)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-16, 5, 0},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1.5, 3, 1.5},
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
		&VolumeToggleComponent{Group: 0},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-4, 3.4, 0},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{2.4, 2.2, 2.4},
		},
		&CellularVolumeComponent{
			Type:        CellularFire,
			Preset:      CAVolumePresetCampfire,
			Resolution:  [3]int{28, 22, 28},
			TickRate:    22,
			Diffusion:   0.18,
			Buoyancy:    0.52,
			Cooling:     0.06,
			Dissipation: 0.032,
		},
		&VolumeToggleComponent{Group: 1},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-4, 5.3, 0.1},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{2.6, 1.9, 2.4},
		},
		&CellularVolumeComponent{
			Type:                  CellularSmoke,
			Preset:                CAVolumePresetCampfire,
			Resolution:            [3]int{20, 16, 18},
			TickRate:              18,
			Diffusion:             0.2,
			Buoyancy:              0.2,
			Dissipation:           0.04,
			UseAppearanceOverride: true,
			ScatterColor:          [3]float32{0.55, 0.5, 0.42},
			Extinction:            0.85,
			Emission:              0.0,
			UseShadowTintOverride: true,
			ShadowTint:            [3]float32{0.25, 0.2, 0.16},
		},
		&VolumeToggleComponent{Group: 1},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{8, 3.8, -1},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-90), mgl32.Vec3{0, 0, 1}),
			Scale:    mgl32.Vec3{3.2, 1.2, 1.2},
		},
		&CellularVolumeComponent{
			Type:        CellularFire,
			Preset:      CAVolumePresetJetFlame,
			Resolution:  [3]int{34, 18, 18},
			TickRate:    30,
			Diffusion:   0.05,
			Buoyancy:    0.04,
			Cooling:     0.035,
			Dissipation: 0.015,
		},
		&VolumeToggleComponent{Group: 2},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{18, 6.5, -4},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{3.2, 4.2, 3.2},
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
			ScatterColor:          [3]float32{0.62, 0.52, 0.42},
			Extinction:            1.25,
			Emission:              4.2,
			UseShadowTintOverride: true,
			ShadowTint:            [3]float32{0.22, 0.17, 0.13},
		},
		&VolumeToggleComponent{Group: 3},
	)

	state.IntensityEntity = cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 5, 10},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1.7, 3, 1.7},
		},
		&CellularVolumeComponent{
			Type:         CellularFire,
			Preset:       CAVolumePresetTorch,
			Resolution:   [3]int{24, 40, 24},
			TickRate:     24,
			Diffusion:    0.08,
			Buoyancy:     0.92,
			Cooling:      0.03,
			Dissipation:  0.012,
			UseIntensity: true,
			Intensity:    1.0,
			FadeInRate:   0.5,
			FadeOutRate:  0.3,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{9.2, 4.2, 10.2},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{0.9, 2.2, 0.9},
		},
		&CellularVolumeComponent{
			Type:        CellularFire,
			Preset:      CAVolumePresetTorch,
			Resolution:  [3]int{16, 32, 16},
			TickRate:    24,
			Diffusion:   0.08,
			Buoyancy:    0.95,
			Cooling:     0.03,
			Dissipation: 0.012,
		},
		&VolumeToggleComponent{Group: 4},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{10.1, 6.6, 10.6},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{4.6, 5.0, 4.6},
		},
		&CellularVolumeComponent{
			Type:                  CellularFire,
			Preset:                CAVolumePresetExplosion,
			Resolution:            [3]int{44, 60, 44},
			TickRate:              30,
			Diffusion:             0.06,
			Buoyancy:              1.28,
			Cooling:               0.14,
			Dissipation:           0.022,
			UseAppearanceOverride: true,
			ScatterColor:          [3]float32{0.6, 0.48, 0.4},
			Extinction:            1.2,
			Emission:              3.8,
			UseShadowTintOverride: true,
			ShadowTint:            [3]float32{0.22, 0.16, 0.12},
		},
		&VolumeToggleComponent{Group: 4},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{10.3, 8.4, 10.6},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{4.2, 2.8, 4.2},
		},
		&CellularVolumeComponent{
			Type:                  CellularSmoke,
			Preset:                CAVolumePresetCampfire,
			Resolution:            [3]int{28, 18, 28},
			TickRate:              18,
			Diffusion:             0.22,
			Buoyancy:              0.22,
			Dissipation:           0.04,
			UseAppearanceOverride: true,
			ScatterColor:          [3]float32{0.52, 0.48, 0.42},
			Extinction:            0.88,
			Emission:              0.0,
			UseShadowTintOverride: true,
			ShadowTint:            [3]float32{0.24, 0.2, 0.17},
		},
		&VolumeToggleComponent{Group: 4},
	)
	state.StressBurst = cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{11.8, 5.8, 9.4},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{2.6, 3.2, 2.6},
		},
		&CellularVolumeComponent{
			Type:         CellularFire,
			Preset:       CAVolumePresetExplosion,
			Resolution:   [3]int{28, 36, 28},
			TickRate:     30,
			Diffusion:    0.06,
			Buoyancy:     1.15,
			Cooling:      0.16,
			Dissipation:  0.024,
			UseIntensity: true,
			Intensity:    0,
			FadeInRate:   4.8,
			FadeOutRate:  6.4,
		},
		&VolumeToggleComponent{Group: 4},
	)

	fmt.Println("Fire and smoke demo ready")
	fmt.Println("Controls: WASD + mouse to fly, F toggles intensity fade, 1-5 toggle torch/campfire/jet/explosion/stress lane, ESC quits")
}

func addPlinth(cmd *Commands, model AssetId, palette AssetId, center mgl32.Vec3, width, depth float32) {
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{center.X() - width/2, 0, center.Z() - depth/2},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      model,
			VoxelPalette:    palette,
			PivotMode:       PivotModeCorner,
			VoxelResolution: 1.0,
		},
	)
}

func addWarmPointLight(cmd *Commands, pos mgl32.Vec3, color [3]float32, intensity, rng float32) {
	cmd.AddEntity(
		&TransformComponent{
			Position: pos,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:      LightTypePoint,
			Color:     color,
			Intensity: intensity,
			Range:     rng,
		},
	)
}

func toggleVolumesSystem(cmd *Commands, input *Input, state *DemoState) {
	keys := [5]int{Key1, Key2, Key3, Key4, Key5}

	for group, key := range keys {
		if !input.JustPressed[key] {
			continue
		}

		state.GroupEnabled[group] = !state.GroupEnabled[group]
		setVolumeGroupDisabled(cmd, group, !state.GroupEnabled[group])

		if state.GroupEnabled[group] {
			fmt.Printf("%s enabled\n", state.GroupNames[group])
		} else {
			fmt.Printf("%s disabled\n", state.GroupNames[group])
		}
	}

	if input.JustPressed[KeyF] {
		state.IntensityEnabled = !state.IntensityEnabled
		target := float32(0.0)
		if state.IntensityEnabled {
			target = 1.0
		}

		MakeQuery1[CellularVolumeComponent](cmd).Map(func(eid EntityId, volume *CellularVolumeComponent) bool {
			if eid != state.IntensityEntity {
				return true
			}
			volume.Intensity = target
			return false
		})

		fmt.Printf("intensity demo target set to %.0f\n", target)
	}
}

func stressPulseSystem(cmd *Commands, t *Time, state *DemoState) {
	if state == nil || state.StressBurst == 0 {
		return
	}
	stressEnabled := state.GroupEnabled[4]
	cycle := math.Mod(t.Elapsed, 4.2)
	target := float32(0.0)
	if stressEnabled && (cycle < 0.9 || (cycle > 2.0 && cycle < 2.35)) {
		target = 1.0
	}

	MakeQuery1[CellularVolumeComponent](cmd).Map(func(eid EntityId, volume *CellularVolumeComponent) bool {
		if eid != state.StressBurst {
			return true
		}
		volume.Intensity = target
		return false
	})
}

func demoHudSystem(state *VoxelRtState) {
	if state == nil {
		return
	}
	text := "1-4 presets  5 stress lane  F fade demo\nStress lane: overlapping torch/explosion/smoke + pulsing burst"
	state.DrawText(text, 20, 20, 0.8, [4]float32{1, 0.95, 0.82, 1})
}

func setVolumeGroupDisabled(cmd *Commands, group int, disabled bool) {
	MakeQuery2[CellularVolumeComponent, VolumeToggleComponent](cmd).Map(func(_ EntityId, volume *CellularVolumeComponent, toggle *VolumeToggleComponent) bool {
		if toggle.Group == group {
			volume.Disabled = disabled
		}
		return true
	})
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
