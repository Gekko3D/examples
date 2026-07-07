package main

import (
	"fmt"
	"math"
	"os"
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
	labChunkSize       = 32
	labVoxelResolution = 0.5
	labChunkMin        = -2
	labChunkMax        = 2
	labWorldID         = "navmesh-lab-world"
	labLevelID         = "navmesh-lab"
	labAgentProfileID  = "lab_agent"
)

const (
	matGround uint8 = iota + 1
	matTerrain
	matRamp
	matStair
	matPlatform
	matPillar
	matBuilding
)

var (
	navmeshLabDebugGroup = EntityGroupKey{Kind: "debug", ID: "navmesh_lab_nav"}
	navmeshLabNPCGroup   = EntityGroupKey{Kind: "debug", ID: "navmesh_lab_npc"}
)

type DemoModule struct{}

type DemoState struct {
	LevelPath       string
	NPCEntity       EntityId
	NavDebugSpawned bool
	NavDebugStats   navmeshLabDebugStats
	NPCDebugStatus  string
}

type labFixturePaths struct {
	Root      string
	LevelPath string
	WorldPath string
	ChunkDir  string
	NavPath   string
}

type labFixtureResult struct {
	Paths       labFixturePaths
	NavManifest *content.NavManifestDef
	Start       content.Vec3
	End         content.Vec3
}

type labVoxelWorld struct {
	voxels map[content.TerrainChunkCoordDef]map[[3]int]content.ImportedWorldVoxelDef
}

type NavmeshLabNPCComponent struct {
	Selected bool
	Target   content.Vec3
	Route    content.NavHierarchicalRouteResult
	Dirty    bool
	Status   string
}

type navmeshLabDebugGizmoItem struct {
	Type      GizmoType
	Color     [4]float32
	Position  mgl32.Vec3
	Rotation  mgl32.Quat
	Size      float32
	DepthMode GizmoDepthMode
}

type navmeshLabDebugStats struct {
	Available        bool
	Tiles            int
	Edges            int
	Portals          int
	RouteFound       bool
	RouteRefined     bool
	RouteSteps       int
	RouteWaypoints   int
	RefinementStatus string
	RefinementReason string
}

func main() {
	app := NewApp()
	app.UseStates(Startup, Quit)
	app.UseModules(
		TimeModule{},
		AssetServerModule{},
		InputModule{},
		VoxelRtModule{
			WindowWidth:  1400,
			WindowHeight: 900,
			WindowTitle:  "Navmesh Lab",
			DebugMode:    true,
		},
		PhysicsModule{Synchronous: true},
		VoxPhysicsModule{},
		StreamedLevelRuntimeModule{},
		FlyingCameraModule{},
		LifecycleModule{},
		DemoModule{},
	)
	app.Run()
}

func (DemoModule) Install(app *App, cmd *Commands) {
	cmd.AddResources(&DemoState{})

	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(Startup)))
	app.UseSystem(System(navmeshDebugSystem).InStage(Update).RunAlways())
	app.UseSystem(System(navmeshLabNPCInputSystem).InStage(Update).RunAlways())
	app.UseSystem(System(navmeshLabNPCRouteSystem).InStage(Update).RunAlways())
	app.UseSystem(System(navmeshLabNPCGizmoSystem).InStage(PostUpdate).RunAlways())
	app.UseSystem(System(navmeshDebugHUDSystem).InStage(PostUpdate).RunAlways())
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer, state *DemoState) {
	result, err := generateNavmeshLabFixture(defaultGeneratedRoot())
	if err != nil {
		fmt.Printf("navmesh lab setup failed: %v\n", err)
		return
	}
	state.LevelPath = result.Paths.LevelPath

	if err := StartStreamedLevelRuntime(cmd, assets, StreamedLevelRuntimeConfig{
		LevelPath:                  result.Paths.LevelPath,
		StreamingRadius:            4,
		StreamingKeepRadius:        5,
		StreamingPrefetchRadius:    5,
		StreamingCollisionRadius:   3,
		StreamingDestructionRadius: 2,
		MaxChunkCommitsPerFrame:    8,
		MaxStreamingCommitMillis:   6,
	}); err != nil {
		fmt.Printf("navmesh lab runtime failed: %v\n", err)
		return
	}

	spawnCamera(cmd)
	spawnLighting(cmd)
	spawnPointMarker(cmd, result.Start, [4]float32{0.1, 0.9, 0.25, 1}, 0.8)
	spawnPointMarker(cmd, result.End, [4]float32{1, 0.2, 0.16, 1}, 0.8)
	state.NPCEntity = spawnNavmeshLabNPC(cmd, result.Start, result.End)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{8, 0.62, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&GizmoComponent{Type: GizmoGrid, Size: 80, Steps: 20, Color: [4]float32{0.18, 0.35, 0.5, 0.35}},
	)

	fmt.Printf("navmesh lab generated: level=%s nav_tiles=%d sectors=%d\n", result.Paths.LevelPath, len(result.NavManifest.Tiles), len(result.NavManifest.Sectors))
}

func spawnNavmeshLabNPC(cmd *Commands, start content.Vec3, target content.Vec3) EntityId {
	return cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{start[0], start[1], start[2]},
			Rotation: mgl32.QuatIdent(),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&NavmeshLabNPCComponent{
			Target: target,
			Dirty:  true,
			Status: "idle",
		},
	)
}

func spawnCamera(cmd *Commands) {
	position := mgl32.Vec3{12, 28, 72}
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: position,
			LookAt:   mgl32.Vec3{8, 3, 8},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      -94,
			Pitch:    -24,
			Fov:      62,
			Aspect:   1400.0 / 900.0,
			Near:     0.1,
			Far:      1200,
		},
		&FlyingCameraComponent{Speed: 24, Sensitivity: 0.1},
		&StreamedLevelObserverComponent{Radius: 4, KeepRadius: 5, PrefetchRadius: 5, CollisionRadius: 3, DestructionRadius: 2},
	)
}

func spawnLighting(cmd *Commands) {
	cmd.AddEntity(
		&LightComponent{
			Type:      LightTypeAmbient,
			Intensity: 0.24,
			Color:     [3]float32{0.86, 0.92, 1},
		},
	)
	cmd.AddEntity(
		&TransformComponent{
			Position: mgl32.Vec3{-32, 70, 42},
			Rotation: mgl32.QuatRotate(mgl32.DegToRad(-55), mgl32.Vec3{1, 0, 0}).Mul(mgl32.QuatRotate(mgl32.DegToRad(-35), mgl32.Vec3{0, 1, 0})),
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&LightComponent{Type: LightTypeDirectional, Intensity: 1.1, Color: [3]float32{1, 0.96, 0.9}, Range: 800, CastsShadows: true},
	)
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.35, 0.48, 0.62},
		ColorB:     mgl32.Vec3{0.04, 0.07, 0.11},
		Opacity:    1,
	})
}

func navmeshDebugSystem(cmd *Commands, runtime *StreamedLevelRuntimeState, state *DemoState) {
	if state.NavDebugSpawned {
		return
	}
	service := RuntimeNavigationServiceFromStreamedLevelState(runtime)
	if !service.Available() {
		state.NavDebugStats = navmeshLabDebugStats{}
		return
	}
	cmd.RemoveEntitiesInGroup(navmeshLabDebugGroup)
	items, stats := navmeshLabDebugEntityItems(service, navmeshLabDebugRouteRequest(), 4000, 512)
	for _, item := range items {
		cmd.AddEntityInGroup(navmeshLabDebugGroup,
			&TransformComponent{
				Position: item.Position,
				Rotation: item.Rotation,
				Scale:    mgl32.Vec3{1, 1, 1},
			},
			&GizmoComponent{
				Type:      item.Type,
				Color:     item.Color,
				Size:      item.Size,
				DepthMode: item.DepthMode,
			},
		)
	}
	state.NavDebugStats = stats
	state.NavDebugSpawned = true
	fmt.Printf("navmesh lab debug ready: tiles=%d edges=%d portals=%d route_refined=%t route_steps=%d waypoints=%d\n", stats.Tiles, stats.Edges, stats.Portals, stats.RouteRefined, stats.RouteSteps, stats.RouteWaypoints)
}

func navmeshDebugHUDSystem(vox *VoxelRtState, state *DemoState) {
	if vox == nil || state == nil {
		return
	}
	stats := state.NavDebugStats
	color := [4]float32{0.45, 1, 0.65, 1}
	if !stats.Available || !stats.RouteRefined {
		color = [4]float32{1, 0.55, 0.25, 1}
	}
	line := fmt.Sprintf("NAVMESH LAB tiles=%d edges=%d portals=%d route_found=%t refined=%t steps=%d waypoints=%d status=%s",
		stats.Tiles,
		stats.Edges,
		stats.Portals,
		stats.RouteFound,
		stats.RouteRefined,
		stats.RouteSteps,
		stats.RouteWaypoints,
		stats.RefinementStatus,
	)
	if stats.RefinementReason != "" {
		line += " reason=" + stats.RefinementReason
	}
	vox.DrawText(line, 16, 76, 0.44, color)
	status := state.NPCDebugStatus
	if status == "" {
		status = "F6 select NPC | F7 clear | F8 teleport selected NPC to aimed navmesh"
	}
	vox.DrawText("NPC DEBUG "+status, 16, 58, 0.44, [4]float32{0.75, 0.92, 1, 1})
}

func navmeshLabDebugRouteRequest() RuntimeNavigationRouteRequest {
	start, end := labRouteEndpoints()
	return RuntimeNavigationRouteRequest{
		Start: start,
		End:   end,
		Options: content.NavHierarchicalRouteOptions{
			SectorPath: content.NavSectorPathOptions{MaxSectorSearch: 64},
			LocalPath: content.NavPathOptions{
				AgentProfileID:      labAgentProfileID,
				MaxTileSearchRadius: 12,
				MaxTileLoads:        512,
			},
		},
	}
}

func navmeshLabNPCInputSystem(cmd *Commands, input *Input, vox *VoxelRtState, runtime *StreamedLevelRuntimeState, state *DemoState) {
	if cmd == nil || input == nil || state == nil {
		return
	}
	if input.JustPressed[KeyF7] {
		navmeshLabClearNPCSelection(cmd)
		state.NPCDebugStatus = "selection cleared"
		return
	}
	if input.JustPressed[KeyF6] {
		eid, ok := navmeshLabNPCUnderAim(cmd, vox, input)
		navmeshLabClearNPCSelection(cmd)
		if ok {
			if npc := navmeshLabNPCComponent(cmd, eid); npc != nil {
				npc.Selected = true
			}
			state.NPCDebugStatus = fmt.Sprintf("selected npc=%d", eid)
		} else {
			state.NPCDebugStatus = "no NPC under aim"
		}
		return
	}
	if input.JustPressed[KeyF8] {
		eid, npc, tr, ok := navmeshLabSelectedNPC(cmd)
		if !ok {
			state.NPCDebugStatus = "no selected NPC"
			return
		}
		point, polygonID, ok := navmeshLabAimNavmeshPoint(cmd, vox, input, runtime)
		if !ok {
			state.NPCDebugStatus = "no aimed navmesh point"
			return
		}
		tr.Position = point
		npc.Dirty = true
		npc.Status = "teleported"
		state.NPCDebugStatus = fmt.Sprintf("teleported npc=%d polygon=%s", eid, polygonID)
	}
}

func navmeshLabNPCRouteSystem(cmd *Commands, runtime *StreamedLevelRuntimeState) {
	if cmd == nil || runtime == nil {
		return
	}
	service := RuntimeNavigationServiceFromStreamedLevelState(runtime)
	if !service.Available() {
		return
	}
	MakeQuery2[TransformComponent, NavmeshLabNPCComponent](cmd).Map(func(_ EntityId, tr *TransformComponent, npc *NavmeshLabNPCComponent) bool {
		if tr == nil || npc == nil || !npc.Dirty {
			return true
		}
		route, err := service.FindRoute(RuntimeNavigationRouteRequest{
			Start: content.Vec3{tr.Position.X(), tr.Position.Y(), tr.Position.Z()},
			End:   npc.Target,
			Options: content.NavHierarchicalRouteOptions{
				SectorPath: content.NavSectorPathOptions{MaxSectorSearch: 64},
				LocalPath: content.NavPathOptions{
					AgentProfileID:      labAgentProfileID,
					MaxTileSearchRadius: 12,
					MaxTileLoads:        512,
				},
			},
		})
		npc.Dirty = false
		if err != nil {
			npc.Status = "error: " + err.Error()
			npc.Route = content.NavHierarchicalRouteResult{}
			return true
		}
		npc.Route = route
		switch {
		case route.Refined:
			npc.Status = "route_ready"
		case route.Found:
			npc.Status = "coarse_route"
		default:
			npc.Status = "no_route"
		}
		return true
	})
}

func navmeshLabNPCGizmoSystem(cmd *Commands) {
	if cmd == nil {
		return
	}
	cmd.RemoveEntitiesInGroup(navmeshLabNPCGroup)
	MakeQuery2[TransformComponent, NavmeshLabNPCComponent](cmd).Map(func(_ EntityId, tr *TransformComponent, npc *NavmeshLabNPCComponent) bool {
		if tr == nil || npc == nil {
			return true
		}
		color := [4]float32{0.1, 0.9, 1, 0.95}
		size := float32(0.42)
		if npc.Selected {
			color = [4]float32{1, 0.92, 0.12, 1}
			size = 0.56
		}
		navmeshLabAddGizmoItem(cmd, navmeshLabNPCGroup, navmeshLabSphereItem(tr.Position.Add(mgl32.Vec3{0, 0.9, 0}), size, color))
		if npc.Route.Refined {
			for _, item := range navmeshLabRouteItems(content.Vec3{tr.Position.X(), tr.Position.Y(), tr.Position.Z()}, npc.Target, npc.Route) {
				navmeshLabAddGizmoItem(cmd, navmeshLabNPCGroup, item)
			}
		}
		return true
	})
}

func navmeshLabAddGizmoItem(cmd *Commands, group EntityGroupKey, item navmeshLabDebugGizmoItem) {
	cmd.AddEntityInGroup(group,
		&TransformComponent{
			Position: item.Position,
			Rotation: item.Rotation,
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		&GizmoComponent{
			Type:      item.Type,
			Color:     item.Color,
			Size:      item.Size,
			DepthMode: item.DepthMode,
		},
	)
}

func navmeshLabClearNPCSelection(cmd *Commands) {
	if cmd == nil {
		return
	}
	MakeQuery1[NavmeshLabNPCComponent](cmd).Map(func(_ EntityId, npc *NavmeshLabNPCComponent) bool {
		if npc != nil {
			npc.Selected = false
		}
		return true
	})
}

func navmeshLabSelectedNPC(cmd *Commands) (EntityId, *NavmeshLabNPCComponent, *TransformComponent, bool) {
	if cmd == nil {
		return 0, nil, nil, false
	}
	var selected EntityId
	var selectedNPC *NavmeshLabNPCComponent
	var selectedTransform *TransformComponent
	MakeQuery2[TransformComponent, NavmeshLabNPCComponent](cmd).Map(func(eid EntityId, tr *TransformComponent, npc *NavmeshLabNPCComponent) bool {
		if npc == nil || tr == nil || !npc.Selected {
			return true
		}
		selected = eid
		selectedNPC = npc
		selectedTransform = tr
		return false
	})
	return selected, selectedNPC, selectedTransform, selected != 0
}

func navmeshLabNPCComponent(cmd *Commands, eid EntityId) *NavmeshLabNPCComponent {
	var out *NavmeshLabNPCComponent
	MakeQuery1[NavmeshLabNPCComponent](cmd).Map(func(got EntityId, npc *NavmeshLabNPCComponent) bool {
		if got == eid {
			out = npc
			return false
		}
		return true
	})
	return out
}

func navmeshLabNPCUnderAim(cmd *Commands, vox *VoxelRtState, input *Input) (EntityId, bool) {
	origin, dir, ok := navmeshLabAimRay(cmd, vox, input)
	if !ok {
		return 0, false
	}
	return navmeshLabNearestNPCToRay(cmd, origin, dir, 120, 1.0)
}

func navmeshLabNearestNPCToRay(cmd *Commands, origin, dir mgl32.Vec3, maxDistance float32, selectRadius float32) (EntityId, bool) {
	if cmd == nil || dir.LenSqr() <= 1e-8 || maxDistance <= 0 {
		return 0, false
	}
	dir = dir.Normalize()
	bestEntity := EntityId(0)
	bestScore := float32(0)
	found := false
	radiusSq := selectRadius * selectRadius
	MakeQuery2[TransformComponent, NavmeshLabNPCComponent](cmd).Map(func(eid EntityId, tr *TransformComponent, npc *NavmeshLabNPCComponent) bool {
		if tr == nil || npc == nil {
			return true
		}
		aimPoint := tr.Position.Add(mgl32.Vec3{0, 0.9, 0})
		toNPC := aimPoint.Sub(origin)
		along := toNPC.Dot(dir)
		if along < 0 || along > maxDistance {
			return true
		}
		closest := origin.Add(dir.Mul(along))
		perpSq := aimPoint.Sub(closest).LenSqr()
		if perpSq > radiusSq {
			return true
		}
		score := perpSq + along*0.0001
		if !found || score < bestScore {
			found = true
			bestScore = score
			bestEntity = eid
		}
		return true
	})
	return bestEntity, found
}

func navmeshLabAimNavmeshPoint(cmd *Commands, vox *VoxelRtState, input *Input, runtime *StreamedLevelRuntimeState) (mgl32.Vec3, string, bool) {
	if vox == nil || runtime == nil {
		return mgl32.Vec3{}, "", false
	}
	origin, dir, ok := navmeshLabAimRay(cmd, vox, input)
	if !ok {
		return mgl32.Vec3{}, "", false
	}
	hit := vox.Raycast(origin, dir, 160)
	if !hit.Hit {
		return mgl32.Vec3{}, "", false
	}
	return navmeshLabNavmeshPointAt(runtime, origin.Add(dir.Mul(hit.T)))
}

func navmeshLabNavmeshPointAt(runtime *StreamedLevelRuntimeState, hit mgl32.Vec3) (mgl32.Vec3, string, bool) {
	service := RuntimeNavigationServiceFromStreamedLevelState(runtime)
	if !service.Available() || service.BaseNavManifest == nil {
		return mgl32.Vec3{}, "", false
	}
	point := content.Vec3{hit.X(), hit.Y(), hit.Z()}
	for _, entry := range service.BaseNavManifest.Tiles {
		if entry.AgentProfileID != labAgentProfileID || !navmeshLabTileEntryContainsXZ(entry, point) {
			continue
		}
		lookup, err := content.LoadEffectiveNavTile(service.BaseNavManifest, service.BaseNavManifestPath, service.WorldDelta, service.WorldDeltaPath, entry.Coord, labAgentProfileID)
		if err != nil || !lookup.Found || lookup.Empty || lookup.Tile == nil {
			continue
		}
		query, err := content.NewNavTileQuery(lookup.Tile)
		if err != nil {
			continue
		}
		polygon, ok := query.FindPolygonAt(point)
		if !ok {
			continue
		}
		center, ok := query.PolygonCenter(polygon.ID)
		if !ok {
			continue
		}
		return mgl32.Vec3{hit.X(), center[1], hit.Z()}, polygon.ID, true
	}
	return mgl32.Vec3{}, "", false
}

func navmeshLabTileEntryContainsXZ(entry content.NavTileEntryDef, point content.Vec3) bool {
	return point[0] >= entry.BoundsMin[0] && point[0] <= entry.BoundsMax[0] &&
		point[2] >= entry.BoundsMin[2] && point[2] <= entry.BoundsMax[2]
}

func navmeshLabAimRay(cmd *Commands, vox *VoxelRtState, input *Input) (mgl32.Vec3, mgl32.Vec3, bool) {
	camera := navmeshLabCamera(cmd)
	if camera == nil {
		return mgl32.Vec3{}, mgl32.Vec3{}, false
	}
	if input != nil && !input.MouseCaptured && vox != nil {
		origin, dir := vox.ScreenToWorldRay(input.MouseX, input.MouseY, camera)
		if dir.LenSqr() > 1e-8 {
			return origin, dir.Normalize(), true
		}
	}
	dir := navmeshLabForwardFromYawPitch(camera.Yaw, camera.Pitch)
	if dir.LenSqr() <= 1e-8 {
		return mgl32.Vec3{}, mgl32.Vec3{}, false
	}
	return camera.Position, dir.Normalize(), true
}

func navmeshLabCamera(cmd *Commands) *CameraComponent {
	if cmd == nil {
		return nil
	}
	var out *CameraComponent
	MakeQuery1[CameraComponent](cmd).Map(func(_ EntityId, camera *CameraComponent) bool {
		if camera == nil {
			return true
		}
		out = camera
		return false
	})
	return out
}

func navmeshLabForwardFromYawPitch(yawDeg, pitchDeg float32) mgl32.Vec3 {
	yawRad := mgl32.DegToRad(yawDeg)
	pitchRad := mgl32.DegToRad(pitchDeg)
	return mgl32.Vec3{
		float32(math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad))),
		float32(math.Sin(float64(pitchRad))),
		float32(-math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad))),
	}.Normalize()
}

func navmeshLabDebugEntityItems(service RuntimeNavigationService, routeRequest RuntimeNavigationRouteRequest, maxEdges int, maxPortals int) ([]navmeshLabDebugGizmoItem, navmeshLabDebugStats) {
	stats := navmeshLabDebugStats{Available: service.Available()}
	if !service.Available() || service.BaseNavManifest == nil {
		return nil, stats
	}
	profile := routeRequest.Options.LocalPath.AgentProfileID
	if profile == "" {
		profile = labAgentProfileID
	}
	items := make([]navmeshLabDebugGizmoItem, 0)
	remainingEdges := maxEdges
	remainingPortals := maxPortals
	for _, entry := range service.BaseNavManifest.Tiles {
		if entry.AgentProfileID != profile {
			continue
		}
		if remainingEdges <= 0 && remainingPortals <= 0 {
			break
		}
		lookup, err := content.LoadEffectiveNavTile(service.BaseNavManifest, service.BaseNavManifestPath, service.WorldDelta, service.WorldDeltaPath, entry.Coord, profile)
		if err != nil || !lookup.Found || lookup.Empty || lookup.Tile == nil {
			continue
		}
		stats.Tiles++
		edgeItems := navmeshLabTileEdgeItems(lookup.Tile, remainingEdges)
		remainingEdges -= len(edgeItems)
		stats.Edges += len(edgeItems)
		items = append(items, edgeItems...)
		portalItems := navmeshLabTilePortalItems(lookup.Tile, remainingPortals)
		remainingPortals -= len(portalItems) / 3
		stats.Portals += len(portalItems) / 3
		items = append(items, portalItems...)
	}
	route, err := service.FindRoute(routeRequest)
	if err != nil {
		stats.RefinementStatus = "error"
		stats.RefinementReason = err.Error()
		return items, stats
	}
	stats.RouteFound = route.Found
	stats.RouteRefined = route.Refined
	stats.RouteSteps = len(route.LocalPath.Steps)
	stats.RouteWaypoints = len(route.LocalPath.Waypoints)
	stats.RefinementStatus = route.RefinementStatus
	stats.RefinementReason = route.RefinementReason
	items = append(items, navmeshLabRouteItems(routeRequest.Start, routeRequest.End, route)...)
	return items, stats
}

func navmeshLabTileEdgeItems(tile *content.NavTileDef, maxEdges int) []navmeshLabDebugGizmoItem {
	if tile == nil || maxEdges == 0 {
		return nil
	}
	items := make([]navmeshLabDebugGizmoItem, 0)
	seen := make(map[[2]int]struct{})
	for _, polygon := range tile.Polygons {
		if len(polygon.Vertices) < 2 {
			continue
		}
		color := navmeshLabAreaColor(polygon.Area)
		if len(polygon.Neighbors) == 0 {
			color = [4]float32{1, 0.32, 0.22, 0.9}
		}
		for i := range polygon.Vertices {
			if len(items) >= maxEdges {
				return items
			}
			a := polygon.Vertices[i]
			b := polygon.Vertices[(i+1)%len(polygon.Vertices)]
			if a < 0 || b < 0 || a >= len(tile.Vertices) || b >= len(tile.Vertices) || a == b {
				continue
			}
			edge := navmeshLabEdgeKey(a, b)
			if _, ok := seen[edge]; ok {
				continue
			}
			seen[edge] = struct{}{}
			start := navmeshLabRaised(navmeshLabVec3(tile.Vertices[a]), 0.16)
			end := navmeshLabRaised(navmeshLabVec3(tile.Vertices[b]), 0.16)
			item, ok := navmeshLabLineItem(start, end, color)
			if ok {
				items = append(items, item)
			}
		}
	}
	return items
}

func navmeshLabTilePortalItems(tile *content.NavTileDef, maxPortals int) []navmeshLabDebugGizmoItem {
	if tile == nil || maxPortals == 0 {
		return nil
	}
	items := make([]navmeshLabDebugGizmoItem, 0)
	for _, portal := range tile.Portals {
		if maxPortals > 0 && len(items)/3 >= maxPortals {
			return items
		}
		start := navmeshLabRaised(navmeshLabVec3(portal.Start), 0.32)
		end := navmeshLabRaised(navmeshLabVec3(portal.End), 0.32)
		line, ok := navmeshLabLineItem(start, end, [4]float32{1, 0.85, 0.18, 0.96})
		if !ok {
			continue
		}
		items = append(items, line)
		items = append(items, navmeshLabSphereItem(start, 0.12, [4]float32{1, 0.95, 0.3, 0.96}))
		items = append(items, navmeshLabSphereItem(end, 0.12, [4]float32{1, 0.95, 0.3, 0.96}))
	}
	return items
}

func navmeshLabRouteItems(start content.Vec3, end content.Vec3, route content.NavHierarchicalRouteResult) []navmeshLabDebugGizmoItem {
	points := []mgl32.Vec3{navmeshLabRaised(navmeshLabVec3(start), 0.42)}
	if route.Refined && route.LocalPath.Found {
		for _, waypoint := range route.LocalPath.Waypoints {
			points = append(points, navmeshLabRaised(navmeshLabVec3(waypoint), 0.42))
		}
	}
	points = append(points, navmeshLabRaised(navmeshLabVec3(end), 0.42))
	color := [4]float32{0.15, 0.85, 1, 0.98}
	if !route.Refined {
		color = [4]float32{1, 0.45, 0.25, 0.96}
	}
	items := make([]navmeshLabDebugGizmoItem, 0, len(points)*2)
	for i := 0; i+1 < len(points); i++ {
		item, ok := navmeshLabLineItem(points[i], points[i+1], color)
		if ok {
			items = append(items, item)
		}
	}
	for i := 1; i+1 < len(points); i++ {
		items = append(items, navmeshLabSphereItem(points[i], 0.11, [4]float32{0.15, 0.85, 1, 0.98}))
	}
	items = append(items, navmeshLabSphereItem(points[0], 0.2, [4]float32{0.1, 0.9, 0.25, 1}))
	items = append(items, navmeshLabSphereItem(points[len(points)-1], 0.2, [4]float32{1, 0.2, 0.16, 1}))
	return items
}

func navmeshLabLineItem(start mgl32.Vec3, end mgl32.Vec3, color [4]float32) (navmeshLabDebugGizmoItem, bool) {
	delta := end.Sub(start)
	length := delta.Len()
	if length <= 0.0001 || math.IsNaN(float64(length)) || math.IsInf(float64(length), 0) {
		return navmeshLabDebugGizmoItem{}, false
	}
	return navmeshLabDebugGizmoItem{
		Type:      GizmoLine,
		Color:     color,
		Position:  start,
		Rotation:  mgl32.QuatBetweenVectors(mgl32.Vec3{0, 0, 1}, delta.Normalize()).Normalize(),
		Size:      length,
		DepthMode: GizmoDepthModeAlwaysVisible,
	}, true
}

func navmeshLabSphereItem(position mgl32.Vec3, radius float32, color [4]float32) navmeshLabDebugGizmoItem {
	return navmeshLabDebugGizmoItem{
		Type:      GizmoSphere,
		Color:     color,
		Position:  position,
		Rotation:  mgl32.QuatIdent(),
		Size:      radius,
		DepthMode: GizmoDepthModeAlwaysVisible,
	}
}

func navmeshLabAreaColor(area string) [4]float32 {
	switch area {
	case content.NavTraversalWalk, "":
		return [4]float32{0.18, 0.95, 0.42, 0.82}
	case content.NavTraversalJump:
		return [4]float32{1, 0.72, 0.25, 0.9}
	case content.NavTraversalDrop:
		return [4]float32{1, 0.48, 0.16, 0.9}
	default:
		return [4]float32{0.55, 0.78, 1, 0.86}
	}
}

func navmeshLabEdgeKey(a, b int) [2]int {
	if b < a {
		return [2]int{b, a}
	}
	return [2]int{a, b}
}

func navmeshLabVec3(value content.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{value[0], value[1], value[2]}
}

func navmeshLabRaised(value mgl32.Vec3, offset float32) mgl32.Vec3 {
	return value.Add(mgl32.Vec3{0, offset, 0})
}

func quitSystem(cmd *Commands, input *Input) {
	if input != nil && input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}

func spawnPointMarker(cmd *Commands, point content.Vec3, color [4]float32, size float32) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{point[0], point[1], point[2]}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&GizmoComponent{Type: GizmoSphere, Size: size, Color: color, DepthMode: GizmoDepthModeAlwaysVisible},
	)
}

func defaultGeneratedRoot() string {
	if _, err := os.Stat(filepath.Join("examples", "navmesh_lab", "go.mod")); err == nil {
		return filepath.Join("examples", "navmesh_lab", "generated")
	}
	return "generated"
}

func generateNavmeshLabFixture(root string) (labFixtureResult, error) {
	paths := labFixturePaths{
		Root:      root,
		LevelPath: filepath.Join(root, "navmesh_lab.gklevel"),
		WorldPath: filepath.Join(root, "worlds", "navmesh_lab.gkworld"),
		ChunkDir:  filepath.Join(root, "worlds", "chunks"),
		NavPath:   filepath.Join(root, "worlds", "navmesh_lab.gknav"),
	}
	if err := os.MkdirAll(paths.ChunkDir, 0755); err != nil {
		return labFixtureResult{}, err
	}

	world, chunks := buildNavmeshLabWorld()
	entries := make([]content.ImportedWorldChunkEntryDef, 0, len(chunks))
	for coord, chunk := range chunks {
		chunkPath := filepath.Join(paths.ChunkDir, fmt.Sprintf("chunk_%d_%d_%d.gkchunk", coord.X, coord.Y, coord.Z))
		if err := content.SaveImportedWorldChunk(chunkPath, chunk); err != nil {
			return labFixtureResult{}, err
		}
		entries = append(entries, content.ImportedWorldChunkEntryDef{
			Coord:              coord,
			ChunkPath:          content.AuthorDocumentPath(chunkPath, paths.WorldPath),
			NonEmptyVoxelCount: chunk.NonEmptyVoxelCount,
			PayloadKind:        content.ImportedWorldChunkPayloadSparseJSONV1,
			Tags:               []string{"navmesh_lab"},
		})
	}
	world.Entries = entries
	content.EnsureImportedWorldDefaults(world)
	if err := content.SaveImportedWorld(paths.WorldPath, world); err != nil {
		return labFixtureResult{}, err
	}

	start, end := labRouteEndpoints()
	level := content.NewLevelDef("Navmesh Lab")
	level.ID = labLevelID
	level.ChunkSize = labChunkSize
	level.VoxelResolution = labVoxelResolution
	level.Tags = []string{"navmesh_lab", "navigation_test"}
	level.BaseWorld = &content.LevelBaseWorldDef{
		Kind:             content.ImportedWorldKindVoxelWorld,
		ManifestPath:     content.AuthorDocumentPath(paths.WorldPath, paths.LevelPath),
		CollisionEnabled: true,
		Tags:             []string{"generated"},
	}
	level.Navigation = &content.LevelNavigationDef{
		ManifestPath: content.AuthorDocumentPath(paths.NavPath, paths.LevelPath),
		Tags:         []string{"generated"},
	}
	level.Markers = []content.LevelMarkerDef{
		labMarker("route_start", content.LevelMarkerKindPlayerSpawn, start),
		labMarker("route_end", "route_target", end),
	}
	level.Environment = &content.LevelEnvironmentDef{Preset: "clear_day"}

	bakeResult, err := content.SaveNavBakeForImportedWorldManifest(paths.WorldPath, paths.NavPath, content.NavBakeOptions{
		NavID:   labWorldID + "-nav",
		LevelID: level.ID,
		AgentProfiles: []content.NavAgentProfileDef{
			{
				ID:              labAgentProfileID,
				Name:            "Navmesh lab standing agent",
				Radius:          0.35,
				Height:          1.8,
				StepHeight:      0.55,
				NavCellSize:     0.5,
				MaxSlopeDegrees: 45,
			},
			{
				ID:              "wide_agent",
				Name:            "Wide clearance test agent",
				Radius:          0.75,
				Height:          2.0,
				StepHeight:      0.55,
				NavCellSize:     1.0,
				MaxSlopeDegrees: 45,
			},
		},
	})
	if err != nil {
		return labFixtureResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(paths.LevelPath), 0755); err != nil {
		return labFixtureResult{}, err
	}
	if err := content.SaveLevel(paths.LevelPath, level); err != nil {
		return labFixtureResult{}, err
	}
	if validation := content.ValidateLevel(level, content.LevelValidationOptions{DocumentPath: paths.LevelPath}); validation.HasErrors() {
		return labFixtureResult{}, fmt.Errorf("generated level validation failed: %s", validation.Error())
	}

	return labFixtureResult{
		Paths:       paths,
		NavManifest: bakeResult.Manifest,
		Start:       start,
		End:         end,
	}, nil
}

func buildNavmeshLabWorld() (*content.ImportedWorldDef, map[content.TerrainChunkCoordDef]*content.ImportedWorldChunkDef) {
	world := &content.ImportedWorldDef{
		WorldID:            labWorldID,
		SchemaVersion:      content.CurrentImportedWorldSchemaVersion,
		Kind:               content.ImportedWorldKindVoxelWorld,
		ChunkSize:          labChunkSize,
		VoxelResolution:    labVoxelResolution,
		ChunkPayloadKind:   content.ImportedWorldChunkPayloadSparseJSONV1,
		SourceBuildVersion: "navmesh_lab_v1",
		Tags:               []string{"generated", "navmesh_lab", "large_level"},
		Palette: []content.ImportedWorldPaletteColor{
			{0, 0, 0, 0},
			{88, 126, 78, 255},
			{116, 138, 120, 255},
			{160, 142, 82, 255},
			{188, 176, 128, 255},
			{110, 114, 130, 255},
			{150, 154, 164, 255},
			{108, 96, 88, 255},
		},
		Materials: []content.ImportedWorldMaterialDef{
			labMaterial(matGround, "ground"),
			labMaterial(matTerrain, "complex_terrain"),
			labMaterial(matRamp, "ramp"),
			labMaterial(matStair, "stairway"),
			labMaterial(matPlatform, "flat_platform"),
			labMaterial(matPillar, "pillar"),
			labMaterial(matBuilding, "building"),
		},
	}

	voxelWorld := labVoxelWorld{voxels: make(map[content.TerrainChunkCoordDef]map[[3]int]content.ImportedWorldVoxelDef)}
	minVoxel := labChunkMin * labChunkSize
	maxVoxel := (labChunkMax + 1) * labChunkSize
	for gx := minVoxel; gx < maxVoxel; gx++ {
		for gz := minVoxel; gz < maxVoxel; gz++ {
			h, material := labBaseGround(gx, gz)
			voxelWorld.fillColumn(gx, gz, 0, h, material)
		}
	}

	voxelWorld.addRaisedRect(50, -20, 82, 18, 6, matPlatform)
	voxelWorld.addStairway(36, -18, 50, 18, 6)
	voxelWorld.addRamp(54, 20, 76, 44, 6)
	voxelWorld.addBuilding(-48, -48, -24, -24, 12)
	voxelWorld.addBuilding(-8, -46, 20, -20, 10)
	voxelWorld.addBuilding(22, -18, 42, 8, 14)
	voxelWorld.addPillars(-44, 8, -18, 34, 4, 14)
	voxelWorld.addPillars(8, 18, 42, 42, 5, 18)
	voxelWorld.addPillars(54, -12, 76, 10, 4, 16)
	voxelWorld.addLowMaze(-52, 34, -20, 58)

	chunks := voxelWorld.chunks()
	return world, chunks
}

func labBaseGround(gx, gz int) (int, uint8) {
	if gx < -54 && gz > 16 {
		return 1 + int(math.Round((math.Sin(float64(gx)*0.16)+math.Cos(float64(gz)*0.18)+2)*0.75)), matTerrain
	}
	if gx > 18 && gz > 20 {
		return int(math.Round((math.Sin(float64(gx+gz)*0.11) + 1) * 0.5)), matTerrain
	}
	return 0, matGround
}

func labMaterial(id uint8, name string) content.ImportedWorldMaterialDef {
	return content.ImportedWorldMaterialDef{
		ID:            int(id),
		PaletteIndex:  id,
		Kind:          name,
		CollisionKind: "solid",
		Roughness:     0.9,
		Tags:          []string{"navmesh_lab", "material:" + name},
	}
}

func labMarker(id, kind string, point content.Vec3) content.LevelMarkerDef {
	return content.LevelMarkerDef{
		ID:   id,
		Name: id,
		Kind: kind,
		Transform: content.LevelTransformDef{
			Position: point,
			Rotation: content.Quat{0, 0, 0, 1},
			Scale:    content.Vec3{1, 1, 1},
		},
	}
}

func labRouteEndpoints() (content.Vec3, content.Vec3) {
	return content.Vec3{-28, 0.8, -28}, content.Vec3{42, 0.8, 42}
}

func (w labVoxelWorld) addRaisedRect(x0, z0, x1, z1, height int, material uint8) {
	for gx := x0; gx <= x1; gx++ {
		for gz := z0; gz <= z1; gz++ {
			w.fillColumn(gx, gz, 0, height, material)
		}
	}
}

func (w labVoxelWorld) addStairway(x0, z0, x1, z1, maxHeight int) {
	width := max(1, x1-x0)
	for gx := x0; gx <= x1; gx++ {
		step := int(math.Round(float64(gx-x0) / float64(width) * float64(maxHeight)))
		for gz := z0; gz <= z1; gz++ {
			w.fillColumn(gx, gz, 0, step, matStair)
		}
	}
}

func (w labVoxelWorld) addRamp(x0, z0, x1, z1, maxHeight int) {
	depth := max(1, z1-z0)
	for gx := x0; gx <= x1; gx++ {
		for gz := z0; gz <= z1; gz++ {
			t := float64(gz-z0) / float64(depth)
			height := int(math.Round(t * float64(maxHeight)))
			w.fillColumn(gx, gz, 0, height, matRamp)
		}
	}
}

func (w labVoxelWorld) addBuilding(x0, z0, x1, z1, wallHeight int) {
	for gx := x0; gx <= x1; gx++ {
		for gz := z0; gz <= z1; gz++ {
			onWall := gx <= x0+1 || gx >= x1-1 || gz <= z0+1 || gz >= z1-1
			if !onWall {
				continue
			}
			doorSouth := gz <= z0+1 && gx >= (x0+x1)/2-2 && gx <= (x0+x1)/2+2
			doorEast := gx >= x1-1 && gz >= (z0+z1)/2-2 && gz <= (z0+z1)/2+2
			if doorSouth || doorEast {
				continue
			}
			base, _ := labBaseGround(gx, gz)
			w.fillColumn(gx, gz, base+1, base+wallHeight, matBuilding)
		}
	}
	for gx := x0 + 2; gx <= x1-2; gx++ {
		for gz := z0 + 2; gz <= z1-2; gz++ {
			base, _ := labBaseGround(gx, gz)
			w.setVoxel(gx, base+wallHeight, gz, matBuilding)
		}
	}
}

func (w labVoxelWorld) addPillars(x0, z0, x1, z1, spacing, height int) {
	for gx := x0; gx <= x1; gx += spacing {
		for gz := z0; gz <= z1; gz += spacing {
			base, _ := labBaseGround(gx, gz)
			for dx := -1; dx <= 1; dx++ {
				for dz := -1; dz <= 1; dz++ {
					if dx*dx+dz*dz > 2 {
						continue
					}
					w.fillColumn(gx+dx, gz+dz, base+1, base+height, matPillar)
				}
			}
		}
	}
}

func (w labVoxelWorld) addLowMaze(x0, z0, x1, z1 int) {
	for gx := x0; gx <= x1; gx++ {
		for gz := z0; gz <= z1; gz++ {
			if (gx-x0)%8 == 0 || (gz-z0)%10 == 0 {
				if (gx+gz)%17 < 4 {
					continue
				}
				base, _ := labBaseGround(gx, gz)
				w.fillColumn(gx, gz, base+1, base+4, matBuilding)
			}
		}
	}
}

func (w labVoxelWorld) fillColumn(gx, gz, y0, y1 int, material uint8) {
	for gy := y0; gy <= y1; gy++ {
		w.setVoxel(gx, gy, gz, material)
	}
}

func (w labVoxelWorld) setVoxel(gx, gy, gz int, material uint8) {
	if gy < 0 || gy >= labChunkSize {
		return
	}
	coord := content.TerrainChunkCoordDef{
		X: floorDiv(gx, labChunkSize),
		Y: 0,
		Z: floorDiv(gz, labChunkSize),
	}
	if coord.X < labChunkMin || coord.X > labChunkMax || coord.Z < labChunkMin || coord.Z > labChunkMax {
		return
	}
	local := [3]int{positiveMod(gx, labChunkSize), gy, positiveMod(gz, labChunkSize)}
	chunk := w.voxels[coord]
	if chunk == nil {
		chunk = make(map[[3]int]content.ImportedWorldVoxelDef)
		w.voxels[coord] = chunk
	}
	chunk[local] = content.ImportedWorldVoxelDef{X: local[0], Y: local[1], Z: local[2], Value: material}
}

func (w labVoxelWorld) chunks() map[content.TerrainChunkCoordDef]*content.ImportedWorldChunkDef {
	out := make(map[content.TerrainChunkCoordDef]*content.ImportedWorldChunkDef)
	for cx := labChunkMin; cx <= labChunkMax; cx++ {
		for cz := labChunkMin; cz <= labChunkMax; cz++ {
			coord := content.TerrainChunkCoordDef{X: cx, Y: 0, Z: cz}
			voxelsByLocal := w.voxels[coord]
			voxels := make([]content.ImportedWorldVoxelDef, 0, len(voxelsByLocal))
			for _, voxel := range voxelsByLocal {
				voxels = append(voxels, voxel)
			}
			out[coord] = &content.ImportedWorldChunkDef{
				WorldID:            labWorldID,
				SchemaVersion:      content.CurrentImportedWorldChunkSchemaVersion,
				Coord:              coord,
				ChunkSize:          labChunkSize,
				VoxelResolution:    labVoxelResolution,
				PayloadKind:        content.ImportedWorldChunkPayloadSparseJSONV1,
				Voxels:             voxels,
				NonEmptyVoxelCount: len(voxels),
				Tags:               []string{"navmesh_lab"},
			}
		}
	}
	return out
}

func floorDiv(value, divisor int) int {
	quotient := value / divisor
	remainder := value % divisor
	if remainder != 0 && ((remainder < 0) != (divisor < 0)) {
		quotient--
	}
	return quotient
}

func positiveMod(value, divisor int) int {
	mod := value % divisor
	if mod < 0 {
		mod += divisor
	}
	return mod
}
