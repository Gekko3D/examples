package main

import (
	"fmt"

	. "github.com/gekko3d/gekko"
	app_rt "github.com/gekko3d/gekko/voxelrt/rt/app"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	startup State = iota
	quit

	windowWidth  = 1280
	windowHeight = 720
	decalTile    = 32
	decalCols    = 2
	decalRows    = 2
)

type demoModule struct{}

func main() {
	NewApp().
		UseStates(startup, quit).
		UseModules(
			TimeModule{},
			AssetServerModule{},
			InputModule{},
			VoxelRtModule{
				WindowWidth:  windowWidth,
				WindowHeight: windowHeight,
				WindowTitle:  "Deferred Decals",
				RenderMode:   RenderModeLit,
				RenderFeatures: []VoxelRtRenderFeature{
					&app_rt.DecalFeature{},
				},
			},
			FlyingCameraModule{},
			demoModule{},
		).
		Run()
}

func (demoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(System(setupScene).InStage(Prelude).InState(OnEnter(startup)))
	app.UseSystem(System(quitSystem).InStage(PreUpdate).RunAlways())
}

func setupScene(cmd *Commands, assets *AssetServer, rt *VoxelRtState) {
	addCamera(cmd)
	addLights(cmd)
	addGeometry(cmd, assets)

	atlas := assets.CreateTextureFromTexels(
		decalAtlasTexels(), decalTile*decalCols, decalTile*decalRows, 1,
		TextureDimension2D, TextureFormatRGBA8Unorm,
	)
	rt.SetRuntimeDecals([]DecalInstance{
		frontDecal(mgl32.Vec3{-5.5, 8.8, 0.02}, [4]float32{0.72, 0.04, 0.03, 1}, atlas, 0, [3]float32{2.0, 1.35, 0.12}),
		frontDecal(mgl32.Vec3{0, 5.7, 0.02}, [4]float32{0.08, 0.08, 0.08, 0.92}, atlas, 1, [3]float32{2.4, 1.6, 0.12}),
		frontDecal(mgl32.Vec3{5.3, 9.7, 0.02}, [4]float32{0.78, 0.10, 0.04, 1}, atlas, 2, [3]float32{1.65, 2.15, 0.12}),
		floorDecal(mgl32.Vec3{-3.0, 0.02, 4.0}, [4]float32{0.12, 0.12, 0.12, 0.9}, atlas, 3, [3]float32{2.5, 1.7, 0.12}),
	})

	fmt.Println("Deferred decals demo ready")
	fmt.Println("WASD + mouse to fly; inspect the wall and floor marks. Escape quits.")
}

func addCamera(cmd *Commands) {
	position := mgl32.Vec3{0, 7, 23}
	cmd.AddEntity(
		&TransformComponent{Position: position, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: position, LookAt: mgl32.Vec3{0, 6, 0}, Up: mgl32.Vec3{0, 1, 0},
			Yaw: 0, Pitch: -3, Fov: 54, Aspect: float32(windowWidth) / windowHeight, Near: 0.1, Far: 300,
		},
		&FlyingCameraComponent{Speed: 12, Sensitivity: 0.1},
	)
}

func addLights(cmd *Commands) {
	cmd.AddEntity(&LightComponent{Type: LightTypeAmbient, Intensity: 0.18, Color: [3]float32{1, 1, 1}})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType: SkyboxLayerGradient, Resolution: [2]int{1024, 512},
		ColorA: mgl32.Vec3{0.42, 0.54, 0.66}, ColorB: mgl32.Vec3{0.04, 0.07, 0.13},
		Opacity: 1, Smooth: true, BlendMode: SkyboxBlendAlpha,
	})
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-12, 24, 14}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 1.25, Color: [3]float32{1.0, 0.83, 0.67}, Range: 38},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{10, 12, 8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LightComponent{Type: LightTypePoint, Intensity: 0.75, Color: [3]float32{0.52, 0.70, 1.0}, Range: 26},
	)
}

func addGeometry(cmd *Commands, assets *AssetServer) {
	wallPalette := assets.CreatePBRPalette([4]uint8{122, 114, 100, 255}, 0.78, 0, 0, 1.4)
	floorPalette := assets.CreatePBRPalette([4]uint8{68, 72, 78, 255}, 0.68, 0, 0, 1.3)
	trimPalette := assets.CreatePBRPalette([4]uint8{48, 54, 62, 255}, 0.48, 0, 0, 1.25)
	wall := assets.CreateCubeModel(20, 13, 1, 1)
	floor := assets.CreateCubeModel(28, 1, 22, 1)
	trim := assets.CreateCubeModel(22, 1, 1, 1)

	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-10, 0, -1}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: wall, VoxelPalette: wallPalette, PivotMode: PivotModeCorner, VoxelResolution: 1},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-14, -1, -8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: floor, VoxelPalette: floorPalette, PivotMode: PivotModeCorner, VoxelResolution: 1},
	)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-11, 13, -1}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: trim, VoxelPalette: trimPalette, PivotMode: PivotModeCorner, VoxelResolution: 1},
	)
}

func frontDecal(position mgl32.Vec3, color [4]float32, texture AssetId, tile uint32, extents [3]float32) DecalInstance {
	return DecalInstance{
		Position: position, Rotation: mgl32.QuatIdent(), HalfExtents: extents, Color: color,
		NormalCutoff: DefaultDecalNormalCutoff, Texture: texture, SpriteIndex: tile, AtlasCols: decalCols, AtlasRows: decalRows,
	}
}

func floorDecal(position mgl32.Vec3, color [4]float32, texture AssetId, tile uint32, extents [3]float32) DecalInstance {
	return DecalInstance{
		Position: position, Rotation: mgl32.QuatRotate(mgl32.DegToRad(-90), mgl32.Vec3{1, 0, 0}), HalfExtents: extents, Color: color,
		NormalCutoff: DefaultDecalNormalCutoff, Texture: texture, SpriteIndex: tile, AtlasCols: decalCols, AtlasRows: decalRows,
	}
}

func decalAtlasTexels() []uint8 {
	width := decalTile * decalCols
	height := decalTile * decalRows
	texels := make([]uint8, width*height*4)
	for y := range height {
		for x := range width {
			tileX, tileY := x/decalTile, y/decalTile
			localX, localY := x%decalTile-decalTile/2, y%decalTile-decalTile/2
			absX, absY := localX, localY
			if absX < 0 {
				absX = -absX
			}
			if absY < 0 {
				absY = -absY
			}
			alpha := decalAlpha(tileY*decalCols+tileX, localX, localY, absX, absY)
			offset := (y*width + x) * 4
			texels[offset], texels[offset+1], texels[offset+2], texels[offset+3] = 255, 255, 255, uint8(alpha)
		}
	}
	return texels
}

func decalAlpha(tile, x, y, absX, absY int) int {
	distance := x*x + y*y
	switch tile {
	case 0:
		if distance < 180 || (absX < 3 && absY < 15) {
			return 255
		}
	case 1:
		if distance < 210 && distance > 55 {
			return 220
		}
	case 2:
		if absX+absY < 17 || (absX < 3 && absY < 15) {
			return 245
		}
	case 3:
		if distance < 195 && (x+y)&3 != 0 {
			return 200
		}
	}
	return 0
}

func quitSystem(cmd *Commands, input *Input) {
	if input.Pressed[KeyEscape] {
		cmd.ChangeState(quit)
	}
}
