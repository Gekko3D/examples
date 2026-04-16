package main

import (
	"fmt"
	"os"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

type DemoModule struct{}

type SpinnerModule struct{}

type SpinnerComponent struct {
	AngularSpeed mgl32.Vec3
}

var setupDone bool

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{},
		AssetServerModule{},
		InputModule{},
		VoxelRtModule{
			WindowWidth:  1600,
			WindowHeight: 900,
			WindowTitle:  "VOX Models",
		},
		HierarchyModule{},
		FlyingCameraModule{},
		SpinnerModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func (SpinnerModule) Install(app *App, cmd *Commands) {
	app.UseSystem(System(spinnerSystem).InStage(Update).RunAlways())
}

func spinnerSystem(cmd *Commands, time *Time) {
	dt := float32(time.Dt)
	MakeQuery2[TransformComponent, SpinnerComponent](cmd).Map(func(_ EntityId, tr *TransformComponent, spin *SpinnerComponent) bool {
		rotX := mgl32.QuatRotate(spin.AngularSpeed.X()*dt, mgl32.Vec3{1, 0, 0})
		rotY := mgl32.QuatRotate(spin.AngularSpeed.Y()*dt, mgl32.Vec3{0, 1, 0})
		rotZ := mgl32.QuatRotate(spin.AngularSpeed.Z()*dt, mgl32.Vec3{0, 0, 1})
		tr.Rotation = tr.Rotation.Mul(rotZ.Mul(rotY).Mul(rotX)).Normalize()
		return true
	})
}

func resolveAsset(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	alt := "examples/vox_models/" + path
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return path
}

func setupScene(cmd *Commands, assets *AssetServer) {
	if setupDone {
		return
	}

	fmt.Println("vox_models: WASD + mouse to fly, Esc to quit.")

	spawnCamera(cmd)
	spawnLighting(cmd)
	spawnFloor(cmd, assets)

	if !spawnSingleModelShowcase(cmd, assets) {
		spawnFallbackSingleModelShowcase(cmd, assets)
	}
	if !spawnHierarchicalShowcase(cmd, assets) {
		spawnFallbackHierarchicalShowcase(cmd, assets)
	}

	setupDone = true
}

func spawnCamera(cmd *Commands) {
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{0, 16, 42},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: mgl32.Vec3{0, 16, 42},
			LookAt:   mgl32.Vec3{2, 6, 4},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -14,
			Fov:      56,
			Aspect:   1600.0 / 900.0,
			Near:     0.1,
			Far:      600,
		},
		&FlyingCameraComponent{Speed: 12, Sensitivity: 0.1},
	)
}

func spawnLighting(cmd *Commands) {
	cmd.AddEntity(&LightComponent{
		Type:      LightTypeAmbient,
		Intensity: 0.16,
		Color:     [3]float32{1, 1, 1},
	})
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{28, 40, 18},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-50), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(28), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    0.85,
			Color:        [3]float32{1.0, 0.96, 0.9},
			Range:        600,
			CastsShadows: true,
		},
	)
	cmd.AddEntity(&SkyAmbientComponent{SkyMix: 0.18})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.74, 0.82, 0.93},
		ColorB:     mgl32.Vec3{0.21, 0.38, 0.68},
		Opacity:    1,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})
}

func spawnFloor(cmd *Commands, assets *AssetServer) {
	floorModel := assets.CreateCubeModel(140, 1, 80, 1.0)
	floorPalette := assets.CreateSimplePalette([4]uint8{72, 76, 84, 255})
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-70, -1, -40},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:      floorModel,
			VoxelPalette:    floorPalette,
			VoxelResolution: 1.0,
			PivotMode:       PivotModeCorner,
		},
	)
}

func spawnSingleModelShowcase(cmd *Commands, assets *AssetServer) bool {
	voxPath := resolveAsset("assets/model.vox")
	voxFile, err := LoadVoxFile(voxPath)
	if err != nil {
		fmt.Printf("vox_models: %s not available, using procedural single-model fallback: %v\n", voxPath, err)
		return false
	}
	if len(voxFile.Models) == 0 {
		fmt.Printf("vox_models: %s contains no models, using procedural single-model fallback.\n", voxPath)
		return false
	}

	modelID := assets.CreateVoxelModel(voxFile.Models[0], 1.0)
	paletteID := assets.CreateVoxelPalette(voxFile.Palette, voxFile.VoxMaterials)
	fmt.Printf("vox_models: loaded single VOX %s (models=%d).\n", voxPath, len(voxFile.Models))

	spawnSingleModelInstances(cmd, modelID, paletteID)
	return true
}

func spawnSingleModelInstances(cmd *Commands, modelID, paletteID AssetId) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-22, 2.8, -12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-12, 1.4, -12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{0.5, 0.5, 0.5}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{2, 5.6, -12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{2, 2, 2}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{16, 2.8, -12},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(35), mgl32.Vec3{0, 1, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(-10), mgl32.Vec3{1, 0, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-22, 2.8, 4}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-8, 1.4, 4}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID, VoxelResolution: 0.05},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{10, 5.6, 4}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID, VoxelResolution: 0.2},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-12, 7, 18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID, PivotMode: PivotModeCenter},
		&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.85, 0}},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{8, 7, 18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: modelID, VoxelPalette: paletteID, PivotMode: PivotModeCorner},
		&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.85, 0}},
	)
}

func spawnHierarchicalShowcase(cmd *Commands, assets *AssetServer) bool {
	voxPath := resolveAsset("assets/vehicle.vox")
	voxFile, err := LoadVoxFile(voxPath)
	if err != nil {
		fmt.Printf("vox_models: %s not available, using procedural hierarchy fallback: %v\n", voxPath, err)
		return false
	}

	voxFileID := assets.CreateVoxelFile(voxFile)
	root := assets.SpawnHierarchicalVoxelModel(
		cmd,
		voxFileID,
		TransformComponent{
			Position: mgl32.Vec3{28, 7, 14},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(18), mgl32.Vec3{0, 1, 0}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		1.0,
	)
	cmd.AddComponents(root, &SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.45, 0}})
	fmt.Printf("vox_models: loaded hierarchical VOX %s (models=%d nodes=%d).\n", voxPath, len(voxFile.Models), len(voxFile.Nodes))
	return true
}

func spawnFallbackSingleModelShowcase(cmd *Commands, assets *AssetServer) {
	modelID, paletteID := buildProceduralSingleModel(assets)
	fmt.Println("vox_models: drop a MagicaVoxel file at examples/vox_models/assets/model.vox to replace the procedural stand-in.")
	spawnSingleModelInstances(cmd, modelID, paletteID)
}

func spawnFallbackHierarchicalShowcase(cmd *Commands, assets *AssetServer) {
	fmt.Println("vox_models: drop a MagicaVoxel scene at examples/vox_models/assets/vehicle.vox to replace the procedural hierarchy.")

	hullModel := assets.CreateCubeModel(14, 4, 5, 1.0)
	wingModel := assets.CreateCubeModel(5, 1, 14, 1.0)
	cabinModel := assets.CreateCubeModel(4, 3, 4, 1.0)
	thrusterModel := assets.CreateCubeModel(2, 2, 2, 1.0)

	hullPalette := assets.CreateSimplePalette([4]uint8{112, 146, 189, 255})
	wingPalette := assets.CreateSimplePalette([4]uint8{78, 99, 128, 255})
	cabinPalette := assets.CreateSimplePalette([4]uint8{216, 194, 112, 255})
	thrusterPalette := assets.CreatePBRPalette([4]uint8{255, 168, 92, 255}, 0.3, 0.0, 0.4, 1.4)

	root := cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{28, 7, 14}, Rotation: mgl32.QuatRotate(mgl32.DegToRad(18), mgl32.Vec3{0, 1, 0}), Scale: mgl32.Vec3{1, 1, 1}},
		&LocalTransformComponent{Position: mgl32.Vec3{28, 7, 14}, Rotation: mgl32.QuatRotate(mgl32.DegToRad(18), mgl32.Vec3{0, 1, 0}), Scale: mgl32.Vec3{1, 1, 1}},
		&SpinnerComponent{AngularSpeed: mgl32.Vec3{0, 0.45, 0}},
	)

	cmd.AddEntity(
		&LocalTransformComponent{Position: mgl32.Vec3{0, 0, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: mgl32.Vec3{0, 0, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: root},
		&VoxelModelComponent{VoxelModel: hullModel, VoxelPalette: hullPalette, VoxelResolution: 0.12},
	)
	cmd.AddEntity(
		&LocalTransformComponent{Position: mgl32.Vec3{-0.2, 0.35, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: mgl32.Vec3{-0.2, 0.35, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: root},
		&VoxelModelComponent{VoxelModel: cabinModel, VoxelPalette: cabinPalette, VoxelResolution: 0.12},
	)
	cmd.AddEntity(
		&LocalTransformComponent{Position: mgl32.Vec3{0, 0, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: mgl32.Vec3{0, 0, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: root},
		&VoxelModelComponent{VoxelModel: wingModel, VoxelPalette: wingPalette, VoxelResolution: 0.12},
	)
	cmd.AddEntity(
		&LocalTransformComponent{Position: mgl32.Vec3{-0.95, -0.15, -0.18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: mgl32.Vec3{-0.95, -0.15, -0.18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: root},
		&VoxelModelComponent{VoxelModel: thrusterModel, VoxelPalette: thrusterPalette, VoxelResolution: 0.12},
	)
	cmd.AddEntity(
		&LocalTransformComponent{Position: mgl32.Vec3{-0.95, -0.15, 0.18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: mgl32.Vec3{-0.95, -0.15, 0.18}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: root},
		&VoxelModelComponent{VoxelModel: thrusterModel, VoxelPalette: thrusterPalette, VoxelResolution: 0.12},
	)
}

func buildProceduralSingleModel(assets *AssetServer) (AssetId, AssetId) {
	var palette VoxPalette
	palette[1] = [4]uint8{78, 128, 212, 255}
	palette[2] = [4]uint8{242, 191, 73, 255}
	palette[3] = [4]uint8{82, 194, 142, 255}
	palette[4] = [4]uint8{232, 112, 92, 255}

	voxels := make([]Voxel, 0, 256)
	appendBox := func(minX, maxX, minY, maxY, minZ, maxZ uint32, color byte) {
		for x := minX; x < maxX; x++ {
			for y := minY; y < maxY; y++ {
				for z := minZ; z < maxZ; z++ {
					voxels = append(voxels, Voxel{X: x, Y: y, Z: z, ColorIndex: color})
				}
			}
		}
	}

	appendBox(1, 7, 0, 2, 1, 5, 1)
	appendBox(2, 6, 2, 8, 2, 4, 2)
	appendBox(4, 7, 5, 8, 0, 2, 3)
	appendBox(0, 3, 6, 10, 3, 6, 4)

	model := VoxModel{
		SizeX:  8,
		SizeY:  10,
		SizeZ:  6,
		Voxels: voxels,
	}

	return assets.CreateVoxelModel(model, 1.0), assets.CreateVoxelPalette(palette, nil)
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
