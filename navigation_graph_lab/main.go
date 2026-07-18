package main

import (
	"fmt"
	"math"
	"path/filepath"

	. "github.com/gekko3d/gekko"
	"github.com/gekko3d/gekko/content"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const (
	labChunkSize       = 16
	labVoxelResolution = float32(1)
	labProfileID       = "walker"
	labSpanDebugLimit  = 256
)

var labProfile = content.NavAgentProfileDef{ID: labProfileID, Radius: 0.4, Height: 1.8, StepHeight: 1.1, MaxSlopeDegrees: 50}

type DemoModule struct{}

type DemoState struct {
	Summary content.NavGraphBakeSummary
	Route   content.NavRouteResult
	Failure content.NavRouteResult
	Error   string
}

type LabData struct {
	Sources []content.NavSourceTileDef
	Graphs  []content.NavGraphTileDef
}

type labWorld struct {
	voxels map[content.TerrainChunkCoordDef]map[[3]int]content.ImportedWorldVoxelDef
}

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{}, AssetServerModule{}, InputModule{},
		VoxelRtModule{WindowWidth: 1400, WindowHeight: 900, WindowTitle: "Navigation Graph Lab"},
		FlyingCameraModule{}, DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})
	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(drawHUD).InStage(PostUpdate).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func bakeLab(world *content.ImportedWorldDef, chunks []content.ImportedWorldChunkDef) error {
	bake, err := content.BakeNavGraphWorld(world, chunks, []content.NavAgentProfileDef{labProfile})
	if err != nil {
		return err
	}
	manifestPath := filepath.Join("generated", "navigation_graph_lab"+content.NavGraphManifestExtension)
	return content.SaveNavGraphBake(manifestPath, &bake)
}

func loadLab() (*DemoState, *LabData, error) {
	manifestPath := filepath.Join("generated", "navigation_graph_lab"+content.NavGraphManifestExtension)
	loaded, err := content.LoadNavGraphBake(manifestPath)
	if err != nil {
		return nil, nil, err
	}
	graphs := graphsForProfile(loaded.GraphTiles, labProfileID)
	start, goal := content.Vec3{1.5, 1, 1.5}, content.Vec3{30.5, 1, 14.5}
	state := &DemoState{}
	state.Route, err = content.FindNavGraphRoute(loaded.SourceTiles, graphs, labChunkSize, labVoxelResolution, start, goal)
	if err != nil {
		return nil, nil, err
	}
	state.Failure, err = content.FindNavGraphRoute(loaded.SourceTiles, graphs, labChunkSize, labVoxelResolution, content.Vec3{-4, 1, -4}, goal)
	if err != nil {
		return nil, nil, err
	}
	state.Summary = content.DiagnoseNavGraphBake(loaded)
	fmt.Printf("navigation graph lab: manifest=%s spans=%d regions=%d route_found=%t failure=%q\n", manifestPath, state.Summary.Spans, state.Summary.Regions, state.Route.Found, state.Failure.FailureReason)
	return state, &LabData{Sources: loaded.SourceTiles, Graphs: graphs}, nil
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	spawnCamera(cmd)
	spawnLighting(cmd)
	world, chunks := buildLabWorld()
	if err := bakeLab(world, chunks); err != nil {
		state.Error = err.Error()
		return
	}
	prepared, data, err := loadLab()
	if err != nil {
		state.Error = err.Error()
		return
	}
	state.Summary, state.Route, state.Failure = prepared.Summary, prepared.Route, prepared.Failure
	spawnVoxelWorld(cmd, assets, world, chunks)
	spawnGraphDebug(cmd, data.Sources, data.Graphs, state.Route)
}

func buildLabWorld() (*content.ImportedWorldDef, []content.ImportedWorldChunkDef) {
	w := labWorld{voxels: map[content.TerrainChunkCoordDef]map[[3]int]content.ImportedWorldVoxelDef{}}
	for x := 0; x < labChunkSize*2; x++ {
		for z := 0; z < labChunkSize; z++ {
			if x >= 5 && x <= 7 && z >= 5 && z <= 7 {
				continue
			}
			height := 0
			if x >= 10 && x <= 13 && z >= 2 && z <= 4 {
				height = x - 9
			}
			w.fillColumn(x, z, 0, height, 1)
		}
	}
	for x := 20; x <= 22; x++ {
		for z := 6; z <= 10; z++ {
			w.fillColumn(x, z, 1, 4, 2)
		}
	}
	for x := 2; x <= 6; x++ {
		for z := 10; z <= 14; z++ {
			w.setVoxel(x, 6, z, 3)
		}
	}
	for x := 24; x <= 26; x++ {
		for z := 2; z <= 4; z++ {
			w.setVoxel(x, 2, z, 3)
		}
	}
	chunks := w.chunks()
	entries := make([]content.ImportedWorldChunkEntryDef, 0, len(chunks))
	for _, chunk := range chunks {
		entries = append(entries, content.ImportedWorldChunkEntryDef{Coord: chunk.Coord, ChunkPath: fmt.Sprintf("chunk_%d_%d_%d.gkchunk", chunk.Coord.X, chunk.Coord.Y, chunk.Coord.Z), NonEmptyVoxelCount: len(chunk.Voxels)})
	}
	world := &content.ImportedWorldDef{
		WorldID: "navigation-graph-lab", SchemaVersion: content.CurrentImportedWorldSchemaVersion,
		Kind: content.ImportedWorldKindVoxelWorld, ChunkSize: labChunkSize, VoxelResolution: labVoxelResolution,
		Palette: []content.ImportedWorldPaletteColor{{}, {74, 86, 96, 255}, {124, 82, 66, 255}, {76, 116, 148, 255}}, Entries: entries,
	}
	content.EnsureImportedWorldSectors(world)
	return world, chunks
}

func (w labWorld) fillColumn(x, z, minY, maxY int, value uint8) {
	for y := minY; y <= maxY; y++ {
		w.setVoxel(x, y, z, value)
	}
}

func (w labWorld) setVoxel(x, y, z int, value uint8) {
	coord := content.TerrainChunkCoordDef{X: floorDiv(x, labChunkSize), Y: floorDiv(y, labChunkSize), Z: floorDiv(z, labChunkSize)}
	local := [3]int{positiveMod(x, labChunkSize), positiveMod(y, labChunkSize), positiveMod(z, labChunkSize)}
	if w.voxels[coord] == nil {
		w.voxels[coord] = map[[3]int]content.ImportedWorldVoxelDef{}
	}
	w.voxels[coord][local] = content.ImportedWorldVoxelDef{X: local[0], Y: local[1], Z: local[2], Value: value}
}

func (w labWorld) chunks() []content.ImportedWorldChunkDef {
	coords := []content.TerrainChunkCoordDef{{}, {X: 1}}
	chunks := make([]content.ImportedWorldChunkDef, 0, len(coords))
	for _, coord := range coords {
		voxels := make([]content.ImportedWorldVoxelDef, 0, len(w.voxels[coord]))
		for _, voxel := range w.voxels[coord] {
			voxels = append(voxels, voxel)
		}
		chunks = append(chunks, content.ImportedWorldChunkDef{WorldID: "navigation-graph-lab", Coord: coord, ChunkSize: labChunkSize, VoxelResolution: labVoxelResolution, Voxels: voxels, NonEmptyVoxelCount: len(voxels)})
	}
	return chunks
}

func spawnCamera(cmd *Commands) {
	position := mgl32.Vec3{16, 28, 42}
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{Position: position, LookAt: mgl32.Vec3{16, 2, 8}, Up: mgl32.Vec3{0, 1, 0}, Yaw: -90, Pitch: -34, Fov: 58, Aspect: 1400.0 / 900.0, Near: 0.1, Far: 500},
		&FlyingCameraComponent{Speed: 18, Sensitivity: 0.1},
	)
}

func spawnLighting(cmd *Commands) {
	cmd.AddEntity(&LightComponent{Type: LightTypeAmbient, Intensity: 0.25, Color: [3]float32{0.9, 0.95, 1}})
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-20, 35, 25}, Rotation: mgl32.QuatRotate(mgl32.DegToRad(-50), mgl32.Vec3{1, 0, 0}), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypeDirectional, Intensity: 1.1, Color: [3]float32{1, 0.96, 0.88}, CastsShadows: true},
	)
	cmd.AddEntity(&SkyboxLayerComponent{LayerType: SkyboxLayerGradient, Resolution: [2]int{1024, 512}, ColorA: mgl32.Vec3{0.32, 0.48, 0.62}, ColorB: mgl32.Vec3{0.03, 0.06, 0.1}, Opacity: 1})
}

func spawnVoxelWorld(cmd *Commands, assets *AssetServer, world *content.ImportedWorldDef, chunks []content.ImportedWorldChunkDef) {
	palette := ImportedWorldPaletteAsset(assets, world)
	for i := range chunks {
		chunk := &chunks[i]
		geometry := assets.RegisterSharedVoxelGeometry(ImportedWorldChunkToXBrickMap(chunk), "")
		chunkWorldSize := float32(chunk.ChunkSize) * chunk.VoxelResolution
		cmd.AddEntity(
			&TransformComponent{
				Position: mgl32.Vec3{float32(chunk.Coord.X) * chunkWorldSize, float32(chunk.Coord.Y) * chunkWorldSize, float32(chunk.Coord.Z) * chunkWorldSize},
				Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{chunk.VoxelResolution / VoxelSize, chunk.VoxelResolution / VoxelSize, chunk.VoxelResolution / VoxelSize},
			},
			&VoxelModelComponent{OverrideGeometry: geometry, VoxelPalette: palette, PivotMode: PivotModeCorner},
		)
	}
}

func spawnGraphDebug(cmd *Commands, sources []content.NavSourceTileDef, graphs []content.NavGraphTileDef, route content.NavRouteResult) {
	graphByCoord := make(map[content.TerrainChunkCoordDef]content.NavGraphTileDef, len(graphs))
	for _, graph := range graphs {
		graphByCoord[graph.Coord] = graph
		base := tileBase(graph.Coord)
		for _, region := range graph.Regions {
			boundsMin := mgl32.Vec3{base.X() + region.BoundsMin[0], region.BoundsMin[1], base.Z() + region.BoundsMin[2]}
			boundsMax := mgl32.Vec3{base.X() + region.BoundsMax[0], region.BoundsMax[1], base.Z() + region.BoundsMax[2]}
			center := boundsMin.Add(boundsMax).Mul(0.5).Add(mgl32.Vec3{0, 0.12, 0})
			scale := boundsMax.Sub(boundsMin)
			scale[0], scale[1], scale[2] = max(scale[0], 0.1), max(scale[1], 0.25), max(scale[2], 0.1)
			addGizmo(cmd, GizmoCube, center, mgl32.QuatIdent(), scale, 1, [4]float32{0.2, 0.55, 1, 0.42})
			addGizmo(cmd, GizmoSphere, vec3(region.Center).Add(base).Add(mgl32.Vec3{0, 0.18, 0}), mgl32.QuatIdent(), mgl32.Vec3{1, 1, 1}, 0.11, [4]float32{0.25, 0.7, 1, 0.95})
		}
		for _, transition := range graph.Transitions {
			start, end := transition.CrossingStart, transition.CrossingEnd
			if transition.ToTile == graph.Coord {
				start[0], start[2] = start[0]+base.X(), start[2]+base.Z()
				end[0], end[2] = end[0]+base.X(), end[2]+base.Z()
			}
			addLine(cmd, vec3(start).Add(mgl32.Vec3{0, 0.24, 0}), vec3(end).Add(mgl32.Vec3{0, 0.24, 0}), [4]float32{1, 0.62, 0.12, 1})
		}
	}
	totalAccepted := 0
	for _, graph := range graphs {
		totalAccepted += len(graph.SpanIDs)
	}
	stride := max(1, (totalAccepted+labSpanDebugLimit-1)/labSpanDebugLimit)
	acceptedIndex := 0
	for _, source := range sources {
		graph := graphByCoord[source.Coord]
		accepted := map[uint32]bool{}
		for _, id := range graph.SpanIDs {
			accepted[id] = true
		}
		base := tileBase(source.Coord)
		for _, span := range source.Spans {
			color, size := [4]float32{1, 0.2, 0.15, 0.95}, float32(0.12)
			if accepted[span.ID] {
				acceptedIndex++
				if acceptedIndex%stride != 0 {
					continue
				}
				color, size = [4]float32{0.12, 1, 0.35, 0.9}, 0.065
			}
			position := mgl32.Vec3{base.X() + (float32(span.X)+0.5)*labVoxelResolution, span.SupportHeight + 0.08, base.Z() + (float32(span.Z)+0.5)*labVoxelResolution}
			addGizmo(cmd, GizmoSphere, position, mgl32.QuatIdent(), mgl32.Vec3{1, 1, 1}, size, color)
		}
	}
	for i := 1; i < len(route.Waypoints); i++ {
		addLine(cmd, vec3(route.Waypoints[i-1]).Add(mgl32.Vec3{0, 0.35, 0}), vec3(route.Waypoints[i]).Add(mgl32.Vec3{0, 0.35, 0}), [4]float32{0.08, 0.95, 1, 1})
	}
	for _, point := range route.Waypoints {
		addGizmo(cmd, GizmoSphere, vec3(point).Add(mgl32.Vec3{0, 0.35, 0}), mgl32.QuatIdent(), mgl32.Vec3{1, 1, 1}, 0.12, [4]float32{0.08, 0.95, 1, 1})
	}
}

func addLine(cmd *Commands, start, end mgl32.Vec3, color [4]float32) {
	delta := end.Sub(start)
	length := delta.Len()
	if length <= 0.0001 || math.IsNaN(float64(length)) || math.IsInf(float64(length), 0) {
		return
	}
	addGizmo(cmd, GizmoLine, start, mgl32.QuatBetweenVectors(mgl32.Vec3{0, 0, 1}, delta.Normalize()).Normalize(), mgl32.Vec3{1, 1, 1}, length, color)
}

func addGizmo(cmd *Commands, kind GizmoType, position mgl32.Vec3, rotation mgl32.Quat, scale mgl32.Vec3, size float32, color [4]float32) {
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: rotation, Scale: scale},
		&GizmoComponent{Type: kind, Color: color, Size: size, DepthMode: GizmoDepthModeAlwaysVisible},
	)
}

func tileBase(coord content.TerrainChunkCoordDef) mgl32.Vec3 {
	return mgl32.Vec3{float32(coord.X*labChunkSize) * labVoxelResolution, 0, float32(coord.Z*labChunkSize) * labVoxelResolution}
}

func vec3(value content.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{value[0], value[1], value[2]}
}

func drawHUD(vox *VoxelRtState, state *DemoState) {
	if state.Error != "" {
		vox.DrawText("NAV GRAPH LAB ERROR: "+state.Error, 16, 24, 0.55, [4]float32{1, 0.3, 0.25, 1})
		return
	}
	text := fmt.Sprintf("NAV GRAPH LAB spans=%d accepted=%d regions=%d transitions=%d route=%t steps=%d waypoints=%d hard_errors=%d\nGreen=accepted spans  Red=rejected  Blue=regions  Orange=transitions  Cyan=route\nExpected failure: %s",
		state.Summary.Spans, state.Summary.AcceptedSpans, state.Summary.Regions, state.Summary.RegionTransitions, state.Route.Found, len(state.Route.Steps), len(state.Route.Waypoints), state.Summary.Validation.HardErrorCount, state.Failure.FailureReason)
	vox.DrawText(text, 16, 24, 0.5, [4]float32{0.82, 0.96, 1, 1})
}

func graphsForProfile(graphs []content.NavGraphTileDef, profile string) []content.NavGraphTileDef {
	result := make([]content.NavGraphTileDef, 0, len(graphs))
	for _, graph := range graphs {
		if graph.AgentProfileID == profile {
			result = append(result, graph)
		}
	}
	return result
}

func floorDiv(value, divisor int) int {
	result := value / divisor
	if value < 0 && value%divisor != 0 {
		result--
	}
	return result
}

func positiveMod(value, divisor int) int {
	result := value % divisor
	if result < 0 {
		result += divisor
	}
	return result
}

func quitSystem(cmd *Commands, input *Input) {
	if input != nil && input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
