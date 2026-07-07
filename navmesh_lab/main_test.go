package main

import (
	"os"
	"testing"

	. "github.com/gekko3d/gekko"
	"github.com/gekko3d/gekko/content"
)

func TestGenerateNavmeshLabFixtureBuildsRoute(t *testing.T) {
	result, err := generateNavmeshLabFixture(t.TempDir())
	if err != nil {
		t.Fatalf("generateNavmeshLabFixture failed: %v", err)
	}
	for _, path := range []string{result.Paths.LevelPath, result.Paths.WorldPath, result.Paths.NavPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
	if result.NavManifest == nil {
		t.Fatal("expected nav manifest")
	}
	if len(result.NavManifest.Tiles) < 25 {
		t.Fatalf("expected nav tiles across the large level, got %d", len(result.NavManifest.Tiles))
	}
	if len(result.NavManifest.Sectors) < 4 {
		t.Fatalf("expected multiple nav sectors, got %d", len(result.NavManifest.Sectors))
	}

	route, err := content.FindHierarchicalNavRoute(
		result.NavManifest,
		result.Paths.NavPath,
		nil,
		"",
		result.Start,
		result.End,
		content.NavHierarchicalRouteOptions{
			SectorPath: content.NavSectorPathOptions{MaxSectorSearch: 64},
			LocalPath: content.NavPathOptions{
				AgentProfileID:      labAgentProfileID,
				MaxTileSearchRadius: 12,
				MaxTileLoads:        512,
			},
		},
	)
	if err != nil {
		t.Fatalf("FindHierarchicalNavRoute failed: %v", err)
	}
	if !route.Found || !route.Refined || !route.LocalPath.Found {
		t.Fatalf("expected refined route, got found=%t refined=%t local=%t status=%s reason=%s tile=%+v", route.Found, route.Refined, route.LocalPath.Found, route.RefinementStatus, route.RefinementReason, route.RefinementTile)
	}
	if len(route.LocalPath.Steps) < 8 || len(route.LocalPath.Waypoints) < 8 {
		t.Fatalf("expected a substantial path, steps=%d waypoints=%d", len(route.LocalPath.Steps), len(route.LocalPath.Waypoints))
	}
}

func TestNavmeshLabFixtureContainsRequestedGeometryClasses(t *testing.T) {
	result, err := generateNavmeshLabFixture(t.TempDir())
	if err != nil {
		t.Fatalf("generateNavmeshLabFixture failed: %v", err)
	}
	world, err := content.LoadImportedWorld(result.Paths.WorldPath)
	if err != nil {
		t.Fatalf("LoadImportedWorld failed: %v", err)
	}
	seen := map[uint8]bool{}
	for _, entry := range world.Entries {
		chunk, err := content.LoadImportedWorldChunk(content.ResolveImportedWorldChunkPath(entry, result.Paths.WorldPath))
		if err != nil {
			t.Fatalf("LoadImportedWorldChunk failed: %v", err)
		}
		for _, voxel := range chunk.Voxels {
			seen[voxel.Value] = true
		}
	}
	for name, material := range map[string]uint8{
		"terrain":  matTerrain,
		"ramp":     matRamp,
		"stairway": matStair,
		"platform": matPlatform,
		"pillar":   matPillar,
		"building": matBuilding,
	} {
		if !seen[material] {
			t.Fatalf("expected generated %s material %d", name, material)
		}
	}
}

func TestNavmeshLabDebugEntityItemsIncludeNavmeshAndRoute(t *testing.T) {
	result, err := generateNavmeshLabFixture(t.TempDir())
	if err != nil {
		t.Fatalf("generateNavmeshLabFixture failed: %v", err)
	}
	items, stats := navmeshLabDebugEntityItems(
		NewRuntimeNavigationService(result.NavManifest, result.Paths.NavPath, nil, ""),
		navmeshLabDebugRouteRequest(),
		4000,
		512,
	)
	if !stats.Available || stats.Tiles == 0 || stats.Edges == 0 {
		t.Fatalf("expected navmesh debug stats, got %+v", stats)
	}
	if !stats.RouteFound || !stats.RouteRefined || stats.RouteSteps == 0 || stats.RouteWaypoints == 0 {
		t.Fatalf("expected refined route debug stats, got %+v", stats)
	}
	var lines, spheres int
	for _, item := range items {
		switch item.Type {
		case GizmoLine:
			lines++
		case GizmoSphere:
			spheres++
		}
	}
	if lines == 0 || spheres == 0 {
		t.Fatalf("expected line and sphere debug items, lines=%d spheres=%d items=%d", lines, spheres, len(items))
	}
}

func TestNavmeshLabNavmeshPointAtFindsPolygonForTeleport(t *testing.T) {
	result, err := generateNavmeshLabFixture(t.TempDir())
	if err != nil {
		t.Fatalf("generateNavmeshLabFixture failed: %v", err)
	}
	runtime := &StreamedLevelRuntimeState{
		BaseNavManifestPath: result.Paths.NavPath,
		BaseNavManifest:     result.NavManifest,
		NavigationRevision:  1,
	}
	point, polygonID, ok := navmeshLabNavmeshPointAt(runtime, navmeshLabVec3(result.Start))
	if !ok {
		t.Fatal("expected navmesh point at route start")
	}
	if polygonID == "" {
		t.Fatal("expected polygon id")
	}
	if point.Sub(navmeshLabVec3(result.Start)).Len() > 1.0 {
		t.Fatalf("expected projected point near start, got %v start=%v", point, result.Start)
	}
}
