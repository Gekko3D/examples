package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gekko3d/gekko/content"
)

func TestLabBakeUsesCurrentBinaryBundle(t *testing.T) {
	world, chunks := buildLabWorld()
	bake, err := content.BakeNavGraphWorld(world, chunks, []content.NavAgentProfileDef{labProfile})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "lab"+content.NavGraphManifestExtension)
	if err := content.SaveNavGraphBake(path, &bake); err != nil {
		t.Fatal(err)
	}
	loaded, err := content.LoadNavGraphBake(path)
	if err != nil {
		t.Fatal(err)
	}
	legacy := func(path string) bool {
		return strings.HasSuffix(path, ".gknavsource") || strings.HasSuffix(path, ".gknavgraph")
	}
	for _, tile := range loaded.Manifest.SourceTiles {
		if legacy(tile.TilePath) {
			t.Fatalf("legacy source tile path %q", tile.TilePath)
		}
	}
	for _, tile := range loaded.Manifest.GraphTiles {
		if legacy(tile.TilePath) {
			t.Fatalf("legacy graph tile path %q", tile.TilePath)
		}
	}
	query, err := content.NewNavGraphQuery(loaded.SourceTiles, graphsForProfile(loaded.GraphTiles, labProfileID), labChunkSize, labVoxelResolution)
	if err != nil {
		t.Fatal(err)
	}
	route, err := query.FindRoute(content.Vec3{1.5, 1, 1.5}, content.Vec3{30.5, 1, 14.5})
	if err != nil || !route.Found || len(route.TileDependencies) == 0 {
		t.Fatalf("route=%+v err=%v", route, err)
	}
}
