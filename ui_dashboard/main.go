package main

import (
	"fmt"
	"math"
	"strings"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const (
	panelSettings = iota
	panelStats
	panelActions
	panelScroll
	panelCount
)

var renderModeOptions = []string{"Lit", "Albedo", "Normals"}

type DemoModule struct{}

type DemoConfig struct {
	PlayerName string
	Speed      float32
	Gravity    float32
	RenderMode int
	ShowDebug  bool
}

type DashboardState struct {
	Panels       [panelCount]EntityId
	PanelVisible [panelCount]bool
	EntityCount  int
	ActionCount  int
	LastAction   string
}

type SelectedEntity struct {
	ID EntityId
}

type SelectableComponent struct {
	Name string
}

type SpinComponent struct {
	Axis         mgl32.Vec3
	AngularSpeed float32
	OrbitRate    float32
	OrbitRadius  float32
	BobRate      float32
	BobScale     float32
	BasePosition mgl32.Vec3
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
			WindowTitle:  "UI Dashboard",
		},
		FlyingCameraModule{},
		UiModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(defaultDemoConfig())
	cmd.AddResources(&DashboardState{
		PanelVisible: [panelCount]bool{true, true, true, true},
		LastAction:   "Dashboard ready",
	})
	cmd.AddResources(&SelectedEntity{})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(togglePanelsSystem).InStage(PreUpdate).RunAlways())
	app.UseSystem(System(syncDemoConfigSystem).InStage(Update).RunAlways())
	app.UseSystem(System(pickSelectableSystem).InStage(Update).RunAlways())
	app.UseSystem(System(animateSceneSystem).InStage(Update).RunAlways())
	app.UseSystem(System(selectedMarkerSystem).InStage(Update).RunAlways())
	app.UseSystem(System(updateDashboardPanelsSystem).InStage(Update).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func defaultDemoConfig() *DemoConfig {
	return &DemoConfig{
		PlayerName: "Operator Nova",
		Speed:      12.0,
		Gravity:    9.81,
		RenderMode: 0,
		ShowDebug:  false,
	}
}

func resetDefaults(cfg *DemoConfig) {
	*cfg = *defaultDemoConfig()
}

func trackedEntity(cmd *Commands, state *DashboardState, components ...any) EntityId {
	state.EntityCount++
	return cmd.AddEntity(components...)
}

func setupScene(cmd *Commands, assets *AssetServer, cfg *DemoConfig, dash *DashboardState) {
	groundPalette := assets.CreateSimplePalette([4]uint8{66, 72, 78, 255})
	plinthPalette := assets.CreateSimplePalette([4]uint8{118, 126, 144, 255})
	warmPalette := assets.CreatePBRPalette([4]uint8{232, 178, 110, 255}, 0.35, 0.0, 0.0, 1.0)
	coolPalette := assets.CreatePBRPalette([4]uint8{86, 156, 255, 255}, 0.18, 0.04, 0.0, 1.0)
	accentPalette := assets.CreateSimplePalette([4]uint8{104, 226, 186, 255})
	rosePalette := assets.CreateSimplePalette([4]uint8{246, 110, 164, 255})

	floorModel := assets.CreateCubeModel(140, 2, 140, 1.0)
	columnModel := assets.CreateCubeModel(12, 26, 12, 1.0)
	plinthModel := assets.CreateCubeModel(18, 4, 18, 1.0)
	sphereModel := assets.CreateSphereModel(7, 1.0)
	cubeModel := assets.CreateCubeModel(12, 12, 12, 1.0)
	coneModel := assets.CreateConeModel(6, 16, 1.0)
	pyramidModel := assets.CreatePyramidModel(12, 14, 1.0)

	trackedEntity(
		cmd,
		dash,
		&TransformComponent{
			Position: mgl32.Vec3{24, 14, 30},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{24, 14, 30},
			LookAt:   mgl32.Vec3{0, 6, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      -38.7,
			Pitch:    -11.7,
			Fov:      58,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      1000,
		},
		&FlyingCameraComponent{Speed: cfg.Speed, Sensitivity: 0.1},
	)

	trackedEntity(
		cmd,
		dash,
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.09,
			Color:     [3]float32{0.82, 0.86, 1.0},
		},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{
			Position: mgl32.Vec3{8, 18, 8},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-38), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(28), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    1.35,
			Color:        [3]float32{1.0, 0.96, 0.9},
			Range:        900,
			CastsShadows: true,
		},
	)
	trackedEntity(cmd, dash, &SkyAmbientComponent{SkyMix: 0.25})
	trackedEntity(
		cmd,
		dash,
		&SkyboxLayerComponent{
			LayerType:  SkyboxLayerGradient,
			Resolution: [2]int{1024, 512},
			ColorA:     mgl32.Vec3{0.18, 0.34, 0.58},
			ColorB:     mgl32.Vec3{0.03, 0.06, 0.14},
			Opacity:    1.0,
			Priority:   0,
			Smooth:     true,
			BlendMode:  SkyboxBlendAlpha,
		},
	)

	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{0, -1, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: floorModel, VoxelPalette: groundPalette},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{-16, 2, -10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: plinthModel, VoxelPalette: plinthPalette},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{18, 2, -12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: plinthModel, VoxelPalette: plinthPalette},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{-2, 13, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: columnModel, VoxelPalette: accentPalette},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{-16, 10, -10}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: coolPalette},
		&SelectableComponent{Name: "Orbit Sphere"},
		&SpinComponent{
			Axis:         mgl32.Vec3{0, 1, 0},
			AngularSpeed: 1.2,
			OrbitRate:    0.8,
			OrbitRadius:  2.0,
			BobRate:      1.4,
			BobScale:     1.0,
			BasePosition: mgl32.Vec3{-16, 10, -10},
		},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{18, 10, -12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: cubeModel, VoxelPalette: warmPalette},
		&SelectableComponent{Name: "Data Cube"},
		&SpinComponent{
			Axis:         mgl32.Vec3{0.2, 1.0, 0.1},
			AngularSpeed: 0.95,
			OrbitRate:    -0.55,
			OrbitRadius:  3.5,
			BobRate:      1.1,
			BobScale:     0.75,
			BasePosition: mgl32.Vec3{18, 10, -12},
		},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{6, 8, 12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: coneModel, VoxelPalette: accentPalette},
		&SelectableComponent{Name: "Signal Cone"},
		&SpinComponent{
			Axis:         mgl32.Vec3{0, 1, 0.6},
			AngularSpeed: 1.55,
			OrbitRate:    0.35,
			OrbitRadius:  2.8,
			BobRate:      1.9,
			BobScale:     0.5,
			BasePosition: mgl32.Vec3{6, 8, 12},
		},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{-12, 7, 14}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: pyramidModel, VoxelPalette: rosePalette},
		&SelectableComponent{Name: "Pyramid Node"},
		&SpinComponent{
			Axis:         mgl32.Vec3{0.5, 1.0, 0},
			AngularSpeed: 1.35,
			OrbitRate:    -0.65,
			OrbitRadius:  2.2,
			BobRate:      1.6,
			BobScale:     0.9,
			BasePosition: mgl32.Vec3{-12, 7, 14},
		},
	)
	trackedEntity(
		cmd,
		dash,
		&TransformComponent{Position: mgl32.Vec3{28, 26, -30}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: sphereModel, VoxelPalette: warmPalette},
		&SelectableComponent{Name: "Sun Beacon"},
		&SpinComponent{
			Axis:         mgl32.Vec3{0.1, 1.0, 0.2},
			AngularSpeed: 0.65,
			OrbitRate:    0.15,
			OrbitRadius:  1.5,
			BobRate:      0.8,
			BobScale:     0.4,
			BasePosition: mgl32.Vec3{28, 26, -30},
		},
	)

	dash.Panels[panelSettings] = trackedEntity(cmd, dash, &UiPanel{Key: "settings", Visible: true})
	dash.Panels[panelStats] = trackedEntity(cmd, dash, &UiPanel{Key: "stats", Visible: true})
	dash.Panels[panelActions] = trackedEntity(cmd, dash, &UiPanel{Key: "actions", Visible: true})
	dash.Panels[panelScroll] = trackedEntity(cmd, dash, &UiPanel{Key: "scroll", Visible: true})
	dash.LastAction = "Panels created. Use F1-F4 to hide or restore them."
}

func togglePanelsSystem(input *Input, dash *DashboardState) {
	if input == nil {
		return
	}
	if input.JustPressed[KeyF1] {
		togglePanel(dash, panelSettings, "Settings")
	}
	if input.JustPressed[KeyF2] {
		togglePanel(dash, panelStats, "Stats")
	}
	if input.JustPressed[KeyF3] {
		togglePanel(dash, panelActions, "Actions")
	}
	if input.JustPressed[KeyF4] {
		togglePanel(dash, panelScroll, "Scroll")
	}
}

func togglePanel(dash *DashboardState, idx int, label string) {
	dash.PanelVisible[idx] = !dash.PanelVisible[idx]
	if dash.PanelVisible[idx] {
		dash.LastAction = fmt.Sprintf("%s panel shown", label)
		return
	}
	dash.LastAction = fmt.Sprintf("%s panel hidden", label)
}

func pickSelectableSystem(input *Input, state *VoxelRtState, dash *DashboardState, selected *SelectedEntity, cmd *Commands) {
	if input == nil || state == nil || selected == nil {
		return
	}
	if input.MouseCaptured || input.GuiCaptured || !input.JustPressed[MouseButtonLeft] {
		return
	}

	var cam *CameraComponent
	MakeQuery1[CameraComponent](cmd).Map(func(_ EntityId, c *CameraComponent) bool {
		cam = c
		return false
	})
	if cam == nil {
		return
	}

	origin, dir := state.ScreenToWorldRay(input.MouseX, input.MouseY, cam)
	hit := state.Raycast(origin, dir, 4000)
	if hit.Entity != 0 {
		if info, ok := selectableForEntity(cmd, hit.Entity); ok {
			selected.ID = hit.Entity
			dash.LastAction = fmt.Sprintf("Selected %s", info.Name)
			return
		}
	}

	if selected.ID != 0 {
		dash.LastAction = "Selection cleared"
	}
	selected.ID = 0
}

func syncDemoConfigSystem(cfg *DemoConfig, vox *VoxelRtState, cmd *Commands) {
	cfg.Speed = clampf(cfg.Speed, 2.0, 40.0)
	cfg.Gravity = clampf(cfg.Gravity, 0.0, 32.0)
	cfg.RenderMode = clampi(cfg.RenderMode, 0, len(renderModeOptions)-1)

	MakeQuery1[FlyingCameraComponent](cmd).Map(func(_ EntityId, fly *FlyingCameraComponent) bool {
		fly.Speed = cfg.Speed
		return true
	})

	if vox == nil || vox.RtApp == nil {
		return
	}
	vox.RtApp.RenderMode = uint32(cfg.RenderMode)
	if cfg.ShowDebug {
		vox.SetDebugOverlayMode(VoxelRtDebugModeScene)
	} else {
		vox.SetDebugOverlayMode(VoxelRtDebugModeOff)
	}
}

func animateSceneSystem(time *Time, cfg *DemoConfig, cmd *Commands) {
	if time == nil {
		return
	}
	elapsed := float32(time.Elapsed)
	speedFactor := cfg.Speed * 0.035
	gravityFactor := 0.2 + cfg.Gravity*0.035

	MakeQuery2[TransformComponent, SpinComponent](cmd).Map(func(_ EntityId, tr *TransformComponent, spin *SpinComponent) bool {
		axis := spin.Axis
		if axis.Len() == 0 {
			axis = mgl32.Vec3{0, 1, 0}
		} else {
			axis = axis.Normalize()
		}

		angle := elapsed * spin.AngularSpeed
		orbitAngle := elapsed * spin.OrbitRate
		tr.Rotation = mgl32.QuatRotate(angle, axis)
		tr.Position = spin.BasePosition.Add(mgl32.Vec3{
			float32(math.Cos(float64(orbitAngle))) * spin.OrbitRadius * speedFactor,
			float32(math.Sin(float64(elapsed*spin.BobRate))) * spin.BobScale * gravityFactor,
			float32(math.Sin(float64(orbitAngle))) * spin.OrbitRadius * speedFactor,
		})
		return true
	})
}

func selectedMarkerSystem(cmd *Commands, assets *AssetServer, state *VoxelRtState, time *Time, selected *SelectedEntity) {
	if state == nil || time == nil || selected == nil || selected.ID == 0 {
		return
	}

	var cam *CameraComponent
	MakeQuery1[CameraComponent](cmd).Map(func(_ EntityId, c *CameraComponent) bool {
		cam = c
		return false
	})
	if cam == nil {
		return
	}

	info, tr, vox, geometry, ok := selectedTransform(cmd, assets, selected.ID)
	if !ok {
		return
	}

	x, y, screenRadius, onScreen := projectedSelection(state, cam, tr, vox, geometry)
	if !onScreen {
		return
	}

	pulse := float32(0.5 + 0.5*math.Sin(time.Elapsed*4.0))
	color := [4]float32{1.0, 0.9 + 0.1*pulse, 0.25, 1.0}
	padding := float32(4) + pulse*2
	drawSelectionMarker(state, x, y, screenRadius, info.Name, color, padding)
}

func updateDashboardPanelsSystem(cmd *Commands, cfg *DemoConfig, dash *DashboardState, selected *SelectedEntity, vox *VoxelRtState, input *Input) {
	if dash.Panels[panelSettings] == 0 {
		return
	}

	cmd.AddComponents(dash.Panels[panelSettings], buildSettingsPanel(cfg, dash))
	cmd.AddComponents(dash.Panels[panelStats], buildStatsPanel(cfg, dash, selected, vox, input, cmd))
	cmd.AddComponents(dash.Panels[panelActions], buildActionsPanel(cfg, dash, selected))
	cmd.AddComponents(dash.Panels[panelScroll], buildScrollPanel(cfg, dash, selected, vox, input, cmd))
}

func buildSettingsPanel(cfg *DemoConfig, dash *DashboardState) *UiPanel {
	return &UiPanel{
		Key:      "settings",
		Anchor:   UiAnchorTopLeft,
		Position: [2]float32{10, 10},
		Width:    280,
		Title:    "Settings",
		Visible:  dash.PanelVisible[panelSettings],
		Children: []UiNode{
			UiRow{
				LabelWidth: 72,
				Children: []UiNode{
					UiLabel{Text: "Name"},
					UiTextField{
						Key:   "name",
						Value: cfg.PlayerName,
						Width: 172,
						OnChange: func(s string) {
							dash.LastAction = fmt.Sprintf("Editing name: %s", s)
						},
						OnCommit: func(s string) {
							cfg.PlayerName = s
							dash.LastAction = fmt.Sprintf("Committed player name: %s", s)
						},
					},
				},
			},
			UiRow{
				LabelWidth: 72,
				Children: []UiNode{
					UiLabel{Text: "Speed"},
					UiNumberField{
						Key:       "speed",
						Value:     cfg.Speed,
						Width:     172,
						Precision: 1,
						OnChange: func(v float32) {
							cfg.Speed = clampf(v, 2.0, 40.0)
						},
						OnCommit: func(v float32) {
							cfg.Speed = clampf(v, 2.0, 40.0)
							dash.LastAction = fmt.Sprintf("Camera speed set to %.1f", cfg.Speed)
						},
					},
				},
			},
			UiRow{
				LabelWidth: 72,
				Children: []UiNode{
					UiLabel{Text: "Gravity"},
					UiNumberField{
						Key:       "gravity",
						Value:     cfg.Gravity,
						Width:     172,
						Precision: 2,
						OnChange: func(v float32) {
							cfg.Gravity = clampf(v, 0.0, 32.0)
						},
						OnCommit: func(v float32) {
							cfg.Gravity = clampf(v, 0.0, 32.0)
							dash.LastAction = fmt.Sprintf("Motion gravity set to %.2f", cfg.Gravity)
						},
					},
				},
			},
			UiRow{
				LabelWidth: 72,
				Children: []UiNode{
					UiLabel{Text: "Render"},
					UiSelectCycle{
						Key:      "render",
						Options:  renderModeOptions,
						Selected: cfg.RenderMode,
						Width:    172,
						OnChange: func(i int) {
							cfg.RenderMode = clampi(i, 0, len(renderModeOptions)-1)
							dash.LastAction = fmt.Sprintf("Render mode: %s", renderModeOptions[cfg.RenderMode])
						},
					},
				},
			},
			UiRow{
				LabelWidth: 72,
				Children: []UiNode{
					UiLabel{Text: "Debug"},
					UiSelectCycle{
						Key:      "debug",
						Options:  []string{"Off", "Scene"},
						Selected: boolIndex(cfg.ShowDebug),
						Width:    172,
						OnChange: func(i int) {
							cfg.ShowDebug = i == 1
							if cfg.ShowDebug {
								dash.LastAction = "Scene debug overlay enabled"
								return
							}
							dash.LastAction = "Scene debug overlay disabled"
						},
					},
				},
			},
			UiSpacer{Height: 10},
			UiButtonControl{
				Key:   "reset",
				Label: "Reset Defaults",
				Width: 244,
				OnClick: func() {
					resetDefaults(cfg)
					dash.ActionCount++
					dash.LastAction = "Defaults restored"
				},
			},
		},
	}
}

func buildStatsPanel(cfg *DemoConfig, dash *DashboardState, selected *SelectedEntity, vox *VoxelRtState, input *Input, cmd *Commands) *UiPanel {
	selectedLabel := "None"
	if info, ok := selectedInfo(cmd, selected); ok {
		selectedLabel = info.Name
	}

	children := []UiNode{
		UiLabel{Text: fmt.Sprintf("FPS: %.0f", fpsValue(vox))},
		UiLabel{Text: fmt.Sprintf("Entities: %d", dash.EntityCount)},
		UiLabel{Text: fmt.Sprintf("Objects: %d", counterValue(vox, "Objects"))},
		UiLabel{Text: fmt.Sprintf("Visible: %d", counterValue(vox, "Visible"))},
		UiLabel{Text: fmt.Sprintf("Panels: %d / %d", visiblePanelCount(dash), panelCount)},
		UiLabel{Text: fmt.Sprintf("Render: %s", renderModeOptions[cfg.RenderMode])},
		UiLabel{Text: fmt.Sprintf("Selected: %s", selectedLabel)},
		UiLabel{Text: fmt.Sprintf("GuiCaptured: %t", input != nil && input.GuiCaptured)},
		UiSpacer{Height: 6},
		UiLabel{Text: "Profiler", Dim: true},
	}

	for _, line := range profilerLines(vox, 4) {
		children = append(children, UiLabel{Text: line, Scale: 0.8, Dim: true})
	}

	return &UiPanel{
		Key:      "stats",
		Anchor:   UiAnchorTopRight,
		Position: [2]float32{10, 10},
		Width:    230,
		Title:    "Stats",
		Visible:  dash.PanelVisible[panelStats],
		Children: children,
	}
}

func buildActionsPanel(cfg *DemoConfig, dash *DashboardState, selected *SelectedEntity) *UiPanel {
	debugLabel := "Enable Scene Debug"
	if cfg.ShowDebug {
		debugLabel = "Disable Scene Debug"
	}

	return &UiPanel{
		Key:      "actions",
		Anchor:   UiAnchorBottomLeft,
		Position: [2]float32{10, 10},
		Width:    270,
		Title:    "Actions",
		Visible:  dash.PanelVisible[panelActions],
		Children: []UiNode{
			UiColumn{
				Spacing: 6,
				Children: []UiNode{
					UiButtonControl{
						Key:   "toggle_debug",
						Label: debugLabel,
						Width: 238,
						OnClick: func() {
							cfg.ShowDebug = !cfg.ShowDebug
							dash.ActionCount++
							if cfg.ShowDebug {
								dash.LastAction = "Scene debug overlay enabled"
								return
							}
							dash.LastAction = "Scene debug overlay disabled"
						},
					},
					UiButtonControl{
						Key:   "cycle_render",
						Label: "Cycle Render Mode",
						Width: 238,
						OnClick: func() {
							cfg.RenderMode = (cfg.RenderMode + 1) % len(renderModeOptions)
							dash.ActionCount++
							dash.LastAction = fmt.Sprintf("Render mode: %s", renderModeOptions[cfg.RenderMode])
						},
					},
					UiButtonControl{
						Key:   "boost_speed",
						Label: "Boost Camera +2",
						Width: 238,
						OnClick: func() {
							cfg.Speed = clampf(cfg.Speed+2.0, 2.0, 40.0)
							dash.ActionCount++
							dash.LastAction = fmt.Sprintf("Camera speed boosted to %.1f", cfg.Speed)
						},
					},
					UiButtonControl{
						Key:   "rename_player",
						Label: "Rotate Player Alias",
						Width: 238,
						OnClick: func() {
							cfg.PlayerName = nextPlayerName(dash.ActionCount)
							dash.ActionCount++
							dash.LastAction = fmt.Sprintf("Player alias changed to %s", cfg.PlayerName)
						},
					},
					UiButtonControl{
						Key:   "clear_selection",
						Label: "Clear Selection",
						Width: 238,
						OnClick: func() {
							selected.ID = 0
							dash.ActionCount++
							dash.LastAction = "Selection cleared"
						},
					},
					UiSpacer{Height: 4},
					UiLabel{Text: fmt.Sprintf("Last action: %s", dash.LastAction), Scale: 0.8, Dim: true},
				},
			},
		},
	}
}

func buildScrollPanel(cfg *DemoConfig, dash *DashboardState, selected *SelectedEntity, vox *VoxelRtState, input *Input, cmd *Commands) *UiPanel {
	selectedLabel := "None"
	if info, ok := selectedInfo(cmd, selected); ok {
		selectedLabel = info.Name
	}

	lines := []string{
		"Scrollable panel: wheel over this panel to move.",
		fmt.Sprintf("Player profile: %s", cfg.PlayerName),
		fmt.Sprintf("Flying speed: %.1f", cfg.Speed),
		fmt.Sprintf("Motion gravity: %.2f", cfg.Gravity),
		fmt.Sprintf("Render mode: %s", renderModeOptions[cfg.RenderMode]),
		fmt.Sprintf("Scene debug: %t", cfg.ShowDebug),
		fmt.Sprintf("Dashboard entities: %d", dash.EntityCount),
		fmt.Sprintf("Action count: %d", dash.ActionCount),
		fmt.Sprintf("Selected object: %s", selectedLabel),
		fmt.Sprintf("FPS snapshot: %.0f", fpsValue(vox)),
		fmt.Sprintf("Objects / Visible: %d / %d", counterValue(vox, "Objects"), counterValue(vox, "Visible")),
		fmt.Sprintf("HiZ culled: %d", counterValue(vox, "HiZCulled")),
		fmt.Sprintf("Shadow casters: %d", counterValue(vox, "ShadowCasters")),
		fmt.Sprintf("Particles: %d", counterValue(vox, "Particles")),
		fmt.Sprintf("GuiCaptured: %t", input != nil && input.GuiCaptured),
		"F1: Settings panel",
		"F2: Stats panel",
		"F3: Actions panel",
		"F4: Scroll panel",
		"With free cursor, click scene objects to select them.",
		"Selected objects get the animated corner frame from spacegame_go.",
		"Text and number fields commit on Enter.",
		"Select cycles advance on click.",
		"Panels are ECS entities with retained keys.",
		"AddComponents rebuilds the UiPanel every frame.",
		"Rows use LabelWidth to keep controls aligned.",
		"Anchors place panels in each screen corner.",
		"UiColumn groups action buttons vertically.",
		"UiSpacer adds breathing room between clusters.",
		"Background scene continues animating behind UI.",
		"Use Tab to capture or release the flying camera.",
	}
	lines = append(lines, profilerLines(vox, 12)...)

	nodes := make([]UiNode, 0, len(lines))
	for i, line := range lines {
		nodes = append(nodes, UiLabel{
			Text:  line,
			Scale: 0.8,
			Dim:   i >= 10,
		})
	}

	return &UiPanel{
		Key:       "scroll",
		Anchor:    UiAnchorBottomRight,
		Position:  [2]float32{10, 10},
		Width:     320,
		MaxHeight: 260,
		Title:     "Activity Feed",
		Visible:   dash.PanelVisible[panelScroll],
		Children: []UiNode{
			UiColumn{
				Spacing:  4,
				Children: nodes,
			},
		},
	}
}

func profilerLines(vox *VoxelRtState, limit int) []string {
	stats := ""
	if vox != nil {
		stats = strings.TrimSpace(vox.ProfilerStats())
	}
	if stats == "" {
		return []string{"Profiler warming up"}
	}
	parts := strings.Split(stats, "\n")
	lines := make([]string, 0, min(limit, len(parts)))
	for _, line := range parts {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= limit {
			break
		}
	}
	if len(lines) == 0 {
		return []string{"Profiler warming up"}
	}
	return lines
}

func visiblePanelCount(dash *DashboardState) int {
	count := 0
	for _, visible := range dash.PanelVisible {
		if visible {
			count++
		}
	}
	return count
}

func counterValue(vox *VoxelRtState, name string) int {
	if vox == nil {
		return 0
	}
	return vox.Counter(name)
}

func fpsValue(vox *VoxelRtState) float64 {
	if vox == nil {
		return 0
	}
	return vox.FPS()
}

func boolIndex(v bool) int {
	if v {
		return 1
	}
	return 0
}

func selectableForEntity(cmd *Commands, id EntityId) (SelectableComponent, bool) {
	var out SelectableComponent
	found := false
	MakeQuery1[SelectableComponent](cmd).Map(func(eid EntityId, info *SelectableComponent) bool {
		if eid != id {
			return true
		}
		out = *info
		found = true
		return false
	})
	return out, found
}

func selectedTransform(cmd *Commands, assets *AssetServer, id EntityId) (SelectableComponent, *TransformComponent, *VoxelModelComponent, *VoxelGeometryAsset, bool) {
	var out SelectableComponent
	var tr *TransformComponent
	var vmc VoxelModelComponent
	var geometry *VoxelGeometryAsset
	found := false
	MakeQuery3[SelectableComponent, TransformComponent, VoxelModelComponent](cmd).Map(func(eid EntityId, info *SelectableComponent, transform *TransformComponent, vox *VoxelModelComponent) bool {
		if eid != id {
			return true
		}
		out = *info
		tr = transform
		vmc = *vox
		if _, asset, ok := ResolveVoxelGeometry(assets, &vmc); ok {
			geometry = asset
		}
		found = true
		return false
	})
	if !found {
		return out, nil, nil, nil, false
	}
	return out, tr, &vmc, geometry, true
}

func selectedInfo(cmd *Commands, selected *SelectedEntity) (SelectableComponent, bool) {
	if selected == nil || selected.ID == 0 {
		return SelectableComponent{}, false
	}
	return selectableForEntity(cmd, selected.ID)
}

func projectedSelection(state *VoxelRtState, cam *CameraComponent, tr *TransformComponent, vox *VoxelModelComponent, geometry *VoxelGeometryAsset) (float32, float32, float32, bool) {
	if tr == nil || vox == nil {
		return 0, 0, 0, false
	}
	fbW, fbH := state.WindowSize()
	var winW, winH int
	if state.RtApp != nil && state.RtApp.Window != nil {
		winW, winH = state.RtApp.Window.GetSize()
	}
	if winW == 0 {
		winW = 1280
	}
	if winH == 0 {
		winH = 720
	}
	scaleX := float32(fbW) / float32(winW)
	scaleY := float32(fbH) / float32(winH)
	localMin, localMax := selectionLocalBounds(geometry)
	pivot := selectionPivot(vox, localMin, localMax)
	worldTr := TransformComponent{
		Position: tr.Position,
		Rotation: tr.Rotation,
		Scale:    EffectiveVoxelScale(vox, tr),
		Pivot:    pivot,
	}
	objToWorld := worldTr.ObjectToWorld()

	corners := [8]mgl32.Vec3{
		{localMin.X(), localMin.Y(), localMin.Z()},
		{localMax.X(), localMin.Y(), localMin.Z()},
		{localMin.X(), localMax.Y(), localMin.Z()},
		{localMax.X(), localMax.Y(), localMin.Z()},
		{localMin.X(), localMin.Y(), localMax.Z()},
		{localMax.X(), localMin.Y(), localMax.Z()},
		{localMin.X(), localMax.Y(), localMax.Z()},
		{localMax.X(), localMax.Y(), localMax.Z()},
	}

	minX, minY := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY := float32(-math.MaxFloat32), float32(-math.MaxFloat32)
	visible := false
	for _, corner := range corners {
		world := objToWorld.Mul4x1(corner.Vec4(1.0)).Vec3()
		x, y, onScreen := state.Project(world, cam)
		if !onScreen {
			continue
		}
		x *= scaleX
		y *= scaleY
		minX = minf(minX, x)
		minY = minf(minY, y)
		maxX = maxf(maxX, x)
		maxY = maxf(maxY, y)
		visible = true
	}
	if !visible {
		return 0, 0, 0, false
	}

	centerX := (minX + maxX) * 0.5
	centerY := (minY + maxY) * 0.5
	halfW := (maxX - minX) * 0.5
	halfH := (maxY - minY) * 0.5
	screenRadius := maxf(halfW, halfH)
	screenRadius = clampf(screenRadius, 4, 420)
	return centerX, centerY, screenRadius, true
}

func selectionLocalBounds(geometry *VoxelGeometryAsset) (mgl32.Vec3, mgl32.Vec3) {
	if geometry == nil {
		return mgl32.Vec3{}, mgl32.Vec3{1, 1, 1}
	}
	minB := geometry.LocalMin
	maxB := geometry.LocalMax
	if maxB.Sub(minB).Len() <= 1e-4 {
		maxB = mgl32.Vec3{
			float32(geometry.VoxModel.SizeX),
			float32(geometry.VoxModel.SizeY),
			float32(geometry.VoxModel.SizeZ),
		}
	}
	if maxB.Sub(minB).Len() <= 1e-4 {
		maxB = mgl32.Vec3{1, 1, 1}
	}
	return minB, maxB
}

func selectionPivot(vox *VoxelModelComponent, localMin, localMax mgl32.Vec3) mgl32.Vec3 {
	if vox == nil {
		return mgl32.Vec3{}
	}
	switch vox.PivotMode {
	case PivotModeCenter:
		return localMin.Add(localMax).Mul(0.5)
	case PivotModeCustom:
		return vox.CustomPivot
	case PivotModeCorner:
		fallthrough
	default:
		return mgl32.Vec3{}
	}
}

func drawSelectionMarker(state *VoxelRtState, x, y, screenRadius float32, label string, color [4]float32, padding float32) {
	drawCelestialMarker(state, x, y, screenRadius+padding, color)
	labelScale := float32(0.55)
	labelWidth, _ := state.MeasureText(label, labelScale)
	state.DrawText(label, x-labelWidth*0.5, y+screenRadius+padding+18, labelScale, color)
}

func drawCelestialMarker(state *VoxelRtState, x, y, r float32, color [4]float32) {
	d := float32(12)

	state.DrawText("+--", x-r, y-r, 0.5, color)
	state.DrawText("|", x-r, y-r+d, 0.5, color)

	tw, _ := state.MeasureText("--+", 0.5)
	state.DrawText("--+", x+r-tw, y-r, 0.5, color)
	state.DrawText("|", x+r, y-r+d, 0.5, color)

	state.DrawText("|", x-r, y+r-d, 0.5, color)
	state.DrawText("+--", x-r, y+r, 0.5, color)

	state.DrawText("|", x+r, y+r-d, 0.5, color)
	state.DrawText("--+", x+r-tw, y+r, 0.5, color)
}

func nextPlayerName(seed int) string {
	names := []string{
		"Operator Nova",
		"Panel Runner",
		"Retained Pilot",
		"Viewport Scout",
		"Scroll Captain",
	}
	return names[seed%len(names)]
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
