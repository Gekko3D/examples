package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const (
	exampleWidth  = 1600
	exampleHeight = 900
)

type DemoModule struct{}

type DemoProbeComponent struct {
	Label string
}

func resolveAsset(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	alt := "examples/entity_lod/" + path
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return path
}

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{},
		AssetServerModule{},
		InputModule{},
		VoxelRtModule{
			WindowWidth:  exampleWidth,
			WindowHeight: exampleHeight,
			WindowTitle:  "Entity LOD Render Paths",
			DebugMode:    true,
		},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(hudSystem).InStage(PostUpdate).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 48, 220}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: mgl32.Vec3{0, 48, 220},
			LookAt:   mgl32.Vec3{0, 24, -900},
			Up:       mgl32.Vec3{0, 1, 0},
			Fov:      50,
			Aspect:   float32(exampleWidth) / float32(exampleHeight),
			Near:     0.1,
			Far:      2600,
		},
		&FlyingCameraComponent{Speed: 65, Sensitivity: 0.1},
	)

	cmd.AddEntity(&LightComponent{
		Type:      LightTypeAmbient,
		Intensity: 0.18,
		Color:     [3]float32{0.9, 0.92, 1.0},
	})
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{120, 300, 60},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-52), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(22), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:      LightTypeDirectional,
			Intensity: 1.0,
			Color:     [3]float32{1.0, 0.96, 0.88},
			Range:     5000,
		},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.06, 0.1, 0.16},
		ColorB:     mgl32.Vec3{0.0, 0.0, 0.03},
		Opacity:    1.0,
		Smooth:     true,
		Priority:   0,
		BlendMode:  SkyboxBlendAlpha,
	})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerStars,
		Seed:       17,
		Scale:      1.0,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.72, 0.82, 1.0},
		ColorB:     mgl32.Vec3{1.0, 1.0, 1.0},
		Threshold:  0.985,
		Opacity:    0.9,
		Priority:   1,
		BlendMode:  SkyboxBlendAdd,
	})

	groundPalette := assets.CreateSimplePalette([4]uint8{36, 40, 54, 255})
	groundModel := assets.CreateCubeModel(900, 2, 2800, 1.0)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-450, -2, -1600}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: groundModel, VoxelPalette: groundPalette, PivotMode: PivotModeCorner, AmbientOcclusionMode: VoxelAODisabled},
	)

	type archetype struct {
		label   string
		model   AssetId
		palette AssetId
		scale   mgl32.Vec3
	}
	archetypes := []archetype{
		{label: "anchor", model: assets.CreateCubeModel(22, 16, 26, 1.0), palette: assets.CreateSimplePalette([4]uint8{255, 194, 96, 255}), scale: mgl32.Vec3{1, 1, 1}},
		{label: "tower", model: assets.CreateCylinderModel(12, 72, 1.0), palette: assets.CreateSimplePalette([4]uint8{126, 236, 170, 255}), scale: mgl32.Vec3{1, 1, 1}},
		{label: "hub", model: assets.CreateSphereModel(18, 1.0), palette: assets.CreateSimplePalette([4]uint8{112, 206, 255, 255}), scale: mgl32.Vec3{1, 1, 1}},
		{label: "frame", model: assets.CreateFrameModel(34, 18, 34, 3, 1.0), palette: assets.CreateSimplePalette([4]uint8{255, 140, 220, 255}), scale: mgl32.Vec3{1, 1, 1}},
		{label: "capsule", model: assets.CreateCapsuleModel(10, 56, 1.0), palette: assets.CreateSimplePalette([4]uint8{255, 238, 128, 255}), scale: mgl32.Vec3{1, 1, 1}},
		{label: "ramp", model: assets.CreateRampModel(34, 18, 24, 1.0), palette: assets.CreateSimplePalette([4]uint8{170, 180, 255, 255}), scale: mgl32.Vec3{1, 1, 1}},
	}
	if voxFile, err := LoadVoxFile(resolveAsset("../vox_models/assets/model.vox")); err == nil && len(voxFile.Models) > 0 {
		archetypes = append(archetypes, archetype{
			label:   "model-vox",
			model:   assets.CreateVoxelModel(voxFile.Models[0], 1.0),
			palette: assets.CreateVoxelPalette(voxFile.Palette, voxFile.VoxMaterials),
			scale:   mgl32.Vec3{0.8, 0.8, 0.8},
		})
	}

	zBands := []float32{-160, -260, -380, -540, -760, -1040, -1380, -1760}
	xOffsets := []float32{-220, -140, -60, 20, 100, 180, 260, 340}
	probeIndex := 0
	for zi, z := range zBands {
		for xi, x := range xOffsets {
			shape := archetypes[(zi+xi)%len(archetypes)]
			y := float32(12 + (zi%3)*10 + (xi%2)*6)
			scale := shape.scale
			if zi >= 5 {
				scale = scale.Mul(1.4)
			}
			label := fmt.Sprintf("%s-%02d", shape.label, probeIndex)
			spawnLODProbe(cmd, mgl32.Vec3{x, y, z}, shape.model, shape.palette, scale, label)
			probeIndex++
		}
	}

	for i, z := range []float32{-180, -340, -620, -980, -1460, -1980} {
		shapeIndex := -1
		for idx := range archetypes {
			if archetypes[idx].label == "model-vox" {
				shapeIndex = idx
				break
			}
		}
		if shapeIndex < 0 {
			break
		}
		shape := archetypes[shapeIndex]
		scale := shape.scale
		if i >= 3 {
			scale = scale.Mul(1.2)
		}
		spawnLODProbe(cmd, mgl32.Vec3{430, 18 + float32(i%2)*10, z}, shape.model, shape.palette, scale, fmt.Sprintf("complex-%02d", i))
	}
}

func spawnLODProbe(cmd *Commands, position mgl32.Vec3, model AssetId, palette AssetId, scale mgl32.Vec3, label string) {
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: scale},
		&VoxelModelComponent{
			VoxelModel:           model,
			VoxelPalette:         palette,
			PivotMode:            PivotModeCenter,
			AmbientOcclusionMode: VoxelAODisabled,
		},
		&EntityLODComponent{
			Bands: []EntityLODBand{
				{MaxDistance: 220, Representation: EntityLODRepresentationFullVoxel},
				{MaxDistance: 520, Representation: EntityLODRepresentationSimplifiedVoxel},
				{MaxDistance: 1100, Representation: EntityLODRepresentationImpostor},
				{MaxDistance: 0, Representation: EntityLODRepresentationDot},
			},
		},
		&DemoProbeComponent{Label: label},
	)
}

func hudSystem(cmd *Commands, state *VoxelRtState) {
	if state == nil || state.RtApp == nil || state.RtApp.Camera == nil {
		return
	}

	counts := map[EntityLODRepresentation]int{
		EntityLODRepresentationFullVoxel:       0,
		EntityLODRepresentationSimplifiedVoxel: 0,
		EntityLODRepresentationImpostor:        0,
		EntityLODRepresentationDot:             0,
	}
	probes := make([]string, 0, 12)
	MakeQuery2[EntityLODComponent, DemoProbeComponent](cmd).Map(func(entityId EntityId, lod *EntityLODComponent, probe *DemoProbeComponent) bool {
		if lod == nil || probe == nil || !lod.SelectionValid {
			return true
		}
		counts[lod.ActiveRepresentation]++
		probes = append(probes, fmt.Sprintf("%s  dist %.0f  %s", probe.Label, lod.ActiveDistance, lod.ActiveRepresentation))
		return true
	})
	sort.Strings(probes)
	if len(probes) > 10 {
		probes = probes[:10]
	}

	lines := []string{
		"WASD/Space/Ctrl move, Tab capture mouse, Esc quit",
		"Near to far path: full voxel -> simplified voxel -> impostor -> dot",
		fmt.Sprintf("Camera: %.0f %.0f %.0f", state.RtApp.Camera.Position.X(), state.RtApp.Camera.Position.Y(), state.RtApp.Camera.Position.Z()),
		fmt.Sprintf("Counts: full=%d simplified=%d impostor=%d dot=%d runtime-sprites=%d", counts[EntityLODRepresentationFullVoxel], counts[EntityLODRepresentationSimplifiedVoxel], counts[EntityLODRepresentationImpostor], counts[EntityLODRepresentationDot], state.RuntimeSpriteCount()),
	}
	lines = append(lines, probes...)

	state.DrawText(strings.Join(lines, "\n"), 20, 20, 0.8, [4]float32{1.0, 0.96, 0.86, 1.0})
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
