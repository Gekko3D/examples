package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
	"github.com/gekko3d/gekko/voxelrt/rt/core"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const (
	windowWidth          = 1440
	windowHeight         = 900
	terrainGroupID       = uint32(4101)
	terrainChunkSize     = 32
	terrainChunkHeight   = 5
	terrainChunkDepth    = 32
	terrainTileRadius    = 3
	terrainTileCount     = terrainTileRadius*2 + 1
	terrainFieldSpan     = terrainTileCount * terrainChunkSize
	terrainSeamEpsilon   = float32(1.0)
	highlightShadowID    = uint32(4102)
	terrainVoxelSize     = float32(1.0)
	showcaseVoxelSize    = float32(1.0)
	transparentVoxelSize = float32(1.0)
)

type DemoModule struct{}

type DemoState struct {
	TerrainModel    AssetId
	MarkerXModel    AssetId
	MarkerZModel    AssetId
	ThinPlateModel  AssetId
	NeedleModel     AssetId
	GlassPanelModel AssetId
	ColumnModel     AssetId
	FloorModel      AssetId
	FloorPalette    AssetId
	TerrainPalette  AssetId
	MarkerPalette   AssetId
	ThinPalette     AssetId
	NeedlePalette   AssetId
	GlassPalette    AssetId
	ColumnPalette   AssetId
}

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{},
		AssetServerModule{},
		InputModule{},
		VoxelRtModule{
			WindowWidth:  windowWidth,
			WindowHeight: windowHeight,
			WindowTitle:  "Baked Normals",
		},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer, demo *DemoState, rt *VoxelRtState) {
	rt.RtApp.OcclusionMode = core.OcclusionOff

	demo.FloorModel = assets.CreateCubeModel(terrainFieldSpan+48, 1, terrainFieldSpan+48, 1.0)
	demo.TerrainModel = assets.CreateCubeModel(terrainChunkSize, terrainChunkHeight, terrainChunkDepth, 1.0)
	demo.MarkerXModel = assets.CreateCubeModel(terrainFieldSpan+2, 1, 1, 1.0)
	demo.MarkerZModel = assets.CreateCubeModel(1, 1, terrainFieldSpan+2, 1.0)
	demo.ThinPlateModel = assets.CreateCubeModel(28, 1, 7, 1.0)
	demo.NeedleModel = assets.CreateCubeModel(1, 18, 1, 1.0)
	demo.GlassPanelModel = assets.CreateCubeModel(1, 14, 18, 1.0)
	demo.ColumnModel = assets.CreateCubeModel(8, 12, 8, 1.0)

	demo.FloorPalette = assets.CreatePBRPalette([4]uint8{58, 60, 64, 255}, 0.82, 0.0, 0.0, 1.45)
	demo.TerrainPalette = assets.CreatePBRPalette([4]uint8{86, 124, 96, 255}, 0.72, 0.0, 0.0, 1.45)
	demo.MarkerPalette = assets.CreatePBRPalette([4]uint8{255, 216, 96, 255}, 0.45, 0.0, 0.0, 1.45)
	demo.ThinPalette = assets.CreatePBRPalette([4]uint8{86, 188, 218, 255}, 0.38, 0.0, 0.0, 1.45)
	demo.NeedlePalette = assets.CreatePBRPalette([4]uint8{236, 112, 88, 255}, 0.36, 0.0, 0.0, 1.45)
	demo.GlassPalette = assets.CreatePBRPaletteWithTransparency([4]uint8{124, 204, 236, 150}, 0.08, 0.0, 0.0, 1.35, 0.48)
	demo.ColumnPalette = assets.CreatePBRPalette([4]uint8{174, 156, 118, 255}, 0.52, 0.0, 0.0, 1.45)

	addCamera(cmd)
	addLighting(cmd)
	addReferenceFloor(cmd, demo)
	addTerrainChunks(cmd, demo)
	addNormalStressGeometry(cmd, demo)

	fmt.Println("Baked Normals demo ready")
	fmt.Printf("Inspect the highlighted cross and the %dx%d terrain tile field: adjacent chunks should shade as one continuous surface.\n", terrainTileCount, terrainTileCount)
	fmt.Println("WASD + mouse to fly, Escape to quit")
}

func addCamera(cmd *Commands) {
	cameraPos := mgl32.Vec3{84, 66, 132}
	cmd.AddEntity(
		&TransformComponent{
			Position: cameraPos,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&CameraComponent{
			Position: cameraPos,
			LookAt:   mgl32.Vec3{0, 4, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      -148,
			Pitch:    -27,
			Fov:      56,
			Aspect:   float32(windowWidth) / float32(windowHeight),
			Near:     0.1,
			Far:      1200,
		},
		&FlyingCameraComponent{Speed: 34.0, Sensitivity: 0.1},
	)
}

func addLighting(cmd *Commands) {
	cmd.AddEntity(&LightComponent{Type: LightTypeAmbient, Intensity: 0.16, Color: [3]float32{1, 1, 1}})
	cmd.AddEntity(&SkyAmbientComponent{SkyMix: 0.22})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.55, 0.70, 0.78},
		ColorB:     mgl32.Vec3{0.08, 0.14, 0.24},
		Opacity:    1.0,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-18, 52, 30},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-55), mgl32.Vec3{1, 0, 0}).Mul(
				mgl32.QuatRotate(mgl32.DegToRad(-30), mgl32.Vec3{0, 1, 0}),
			),
			Scale: mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{
			Type:         LightTypeDirectional,
			Intensity:    0.95,
			Color:        [3]float32{1.0, 0.96, 0.88},
			Range:        900,
			CastsShadows: true,
		},
	)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{2, 13, 11}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.1, Color: [3]float32{0.80, 0.92, 1.0}, Range: 38},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-22, 10, -17}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 0.65, Color: [3]float32{1.0, 0.72, 0.46}, Range: 34},
	)
}

func addReferenceFloor(cmd *Commands, state *DemoState) {
	origin := -float32(terrainFieldSpan)/2 - 24
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{origin, -1, origin},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:           state.FloorModel,
			VoxelPalette:         state.FloorPalette,
			PivotMode:            PivotModeCorner,
			VoxelResolution:      terrainVoxelSize,
			AmbientOcclusionMode: VoxelAODisabled,
		},
	)
}

func addTerrainChunks(cmd *Commands, state *DemoState) {
	for z := -terrainTileRadius; z <= terrainTileRadius; z++ {
		for x := -terrainTileRadius; x <= terrainTileRadius; x++ {
			addTerrainChunk(
				cmd,
				state,
				mgl32.Vec3{
					float32(x * terrainChunkSize),
					0,
					float32(z * terrainChunkDepth),
				},
				[3]int{x, 0, z},
			)
		}
	}

	minCoord := -float32(terrainTileRadius * terrainChunkSize)
	lineY := float32(terrainChunkHeight) + 0.06

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{minCoord - 1, lineY, -0.5},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 0.12, 1},
		},
		&VoxelModelComponent{
			VoxelModel:           state.MarkerXModel,
			VoxelPalette:         state.MarkerPalette,
			PivotMode:            PivotModeCorner,
			VoxelResolution:      terrainVoxelSize,
			AmbientOcclusionMode: VoxelAODisabled,
			ShadowGroupID:        highlightShadowID,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-0.5, lineY, minCoord - 1},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 0.12, 1},
		},
		&VoxelModelComponent{
			VoxelModel:           state.MarkerZModel,
			VoxelPalette:         state.MarkerPalette,
			PivotMode:            PivotModeCorner,
			VoxelResolution:      terrainVoxelSize,
			AmbientOcclusionMode: VoxelAODisabled,
			ShadowGroupID:        highlightShadowID,
		},
	)
}

func addTerrainChunk(cmd *Commands, state *DemoState, position mgl32.Vec3, chunkCoord [3]int) {
	cmd.AddEntity(
		&TransformComponent{
			Position: position,
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:             state.TerrainModel,
			VoxelPalette:           state.TerrainPalette,
			PivotMode:              PivotModeCorner,
			VoxelResolution:        terrainVoxelSize,
			AmbientOcclusionMode:   VoxelAOEnabled,
			ShadowGroupID:          terrainGroupID,
			ShadowSeamWorldEpsilon: terrainSeamEpsilon,
			IsTerrainChunk:         true,
			TerrainGroupID:         terrainGroupID,
			TerrainChunkCoord:      chunkCoord,
			TerrainChunkSize:       terrainChunkSize,
			ShadowCasterGroupID:    uint64(terrainGroupID),
			ShadowCasterGroupLimit: 8,
			ShadowMaxDistance:      220,
			DisableShadows:         false,
		},
	)
}

func addNormalStressGeometry(cmd *Commands, state *DemoState) {
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-6, float32(terrainChunkHeight) + 3, 12},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-14), mgl32.Vec3{0, 1, 0}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:             state.ThinPlateModel,
			VoxelPalette:           state.ThinPalette,
			VoxelResolution:        showcaseVoxelSize,
			AmbientOcclusionMode:   VoxelAOEnabled,
			ShadowGroupID:          highlightShadowID,
			ShadowSeamWorldEpsilon: terrainSeamEpsilon,
			ShadowCasterGroupID:    uint64(highlightShadowID),
			ShadowCasterGroupLimit: 8,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-17, float32(terrainChunkHeight) + 9, 8},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:           state.NeedleModel,
			VoxelPalette:         state.NeedlePalette,
			VoxelResolution:      showcaseVoxelSize,
			AmbientOcclusionMode: VoxelAOEnabled,
			ShadowGroupID:        highlightShadowID,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{7, float32(terrainChunkHeight) + 6.5, 4},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(18), mgl32.Vec3{0, 1, 0}),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&VoxelModelComponent{
			VoxelModel:             state.GlassPanelModel,
			VoxelPalette:           state.GlassPalette,
			VoxelResolution:        transparentVoxelSize,
			AmbientOcclusionMode:   VoxelAOEnabled,
			ShadowGroupID:          highlightShadowID,
			ShadowSeamWorldEpsilon: terrainSeamEpsilon,
			ShadowCasterGroupID:    uint64(highlightShadowID),
			ShadowCasterGroupLimit: 8,
		},
	)

	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{19, float32(terrainChunkHeight) + 6, -6},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(28), mgl32.Vec3{0, 1, 0}),
			Scale:    mgl32.Vec3{0.7, 1.45, 1.1},
		},
		&VoxelModelComponent{
			VoxelModel:           state.ColumnModel,
			VoxelPalette:         state.ColumnPalette,
			VoxelResolution:      showcaseVoxelSize,
			AmbientOcclusionMode: VoxelAOEnabled,
			ShadowGroupID:        highlightShadowID,
		},
	)
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
