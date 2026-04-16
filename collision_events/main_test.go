package main

import "testing"

func TestSupportFloorSitsBelowBowlingLane(t *testing.T) {
	floorMinY := float32(-0.9)
	floorHeight := float32(4) * demoVoxelResolution
	floorTopY := floorMinY + floorHeight

	laneMinY := float32(0.1)
	if floorTopY >= laneMinY {
		t.Fatalf("support floor top %.2f must stay below lane base %.2f", floorTopY, laneMinY)
	}
}
