package main

import (
	"math"

	. "github.com/gekko3d/gekko"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

type TestModule struct{}

func (TestModule) Install(app *App, cmd *Commands) {
	// Add systems
	app.UseSystem(
		System(startup).
			InStage(PreUpdate).
			InState(OnEnter(Startup)),
	)
	app.UseSystem(
		System(updateCamera).
			InStage(Update).
			InAnyState(),
	)
	app.UseSystem(
		System(updateUniforms).
			InStage(Update).
			InAnyState(),
	)

	// Custom stage
	testStage := Stage{
		Name:       "Very Last Stage",
		UpdateType: FixedUpdate,
	}

	app.UseStage(testStage, AfterStage(Finale))
}

type MyVertex struct {
	pos      [3]float32 `gekko:"layout" location:"0" format:"float3"`
	texCoord [2]float32 `gekko:"layout" location:"1" format:"float2"`
}

type MyMaterial struct {
	Uniforms MyUniforms `gekko:"buffer" usage:"uniform,copy_dst" group:"0" binding:"0"`
	Texture  AssetId    `gekko:"texture" group:"0"  binding:"1"`
}

type MyUniforms struct {
	Transform mgl32.Mat4
	Color     [4]float32
}

type Rotating struct{}

func vertex(pos1, pos2, pos3, tc1, tc2 float32) MyVertex {
	return MyVertex{
		pos:      [3]float32{pos1, pos2, pos3},
		texCoord: [2]float32{tc1, tc2},
	}
}

var cubeVertices = []MyVertex{
	// top (0, 0, 1)
	vertex(-1, -1, 1, 0, 0),
	vertex(1, -1, 1, 1, 0),
	vertex(1, 1, 1, 1, 1),
	vertex(-1, 1, 1, 0, 1),
	// bottom (0, 0, -1)
	vertex(-1, 1, -1, 1, 0),
	vertex(1, 1, -1, 0, 0),
	vertex(1, -1, -1, 0, 1),
	vertex(-1, -1, -1, 1, 1),
	// right (1, 0, 0)
	vertex(1, -1, -1, 0, 0),
	vertex(1, 1, -1, 1, 0),
	vertex(1, 1, 1, 1, 1),
	vertex(1, -1, 1, 0, 1),
	// left (-1, 0, 0)
	vertex(-1, -1, 1, 1, 0),
	vertex(-1, 1, 1, 0, 0),
	vertex(-1, 1, -1, 0, 1),
	vertex(-1, -1, -1, 1, 1),
	// front (0, 1, 0)
	vertex(1, 1, -1, 1, 0),
	vertex(-1, 1, -1, 0, 0),
	vertex(-1, 1, 1, 0, 1),
	vertex(1, 1, 1, 1, 1),
	// back (0, -1, 0)
	vertex(1, -1, 1, 0, 0),
	vertex(-1, -1, 1, 1, 0),
	vertex(-1, -1, -1, 1, 1),
	vertex(1, -1, -1, 0, 1),
}

var cubeIndices = []uint16{
	0, 1, 2, 2, 3, 0, // top
	4, 5, 6, 6, 7, 4, // bottom
	8, 9, 10, 10, 11, 8, // right
	12, 13, 14, 14, 15, 12, // left
	16, 17, 18, 18, 19, 16, // front
	20, 21, 22, 22, 23, 20, // back
}

const texelsSize = 256

func createMandelbrotTexels() (texels [texelsSize * texelsSize]uint8) {
	for id := 0; id < (texelsSize * texelsSize); id++ {
		cx := 3.0*float32(id%texelsSize)/float32(texelsSize-1) - 2.0
		cy := 2.0*float32(id/texelsSize)/float32(texelsSize-1) - 1.0
		x, y, count := float32(cx), float32(cy), uint8(0)
		for count < 0xFF && x*x+y*y < 4.0 {
			oldX := x
			x = x*x - y*y + cx
			y = 2.0*oldX*y + cy
			count += 1
		}
		texels[id] = count
	}

	return texels
}

func startup(cmd *Commands, assets *AssetServer, state *WindowState) {
	texels := createMandelbrotTexels()
	textureId := assets.CreateTextureFromTexels(texels[:], texelsSize, texelsSize, 1, TextureDimension2D, TextureFormatR8Uint)
	mesh := assets.CreateMesh(MakeAnySlice(cubeVertices), cubeIndices)
	material := assets.CreateMaterial("assets/shader.wgsl", MyVertex{})
	camera := CameraComponent{
		Position: mgl32.Vec3{1.5, 4, 5},
		LookAt:   mgl32.Vec3{0, 0, 0},
		Up:       mgl32.Vec3{0, 1, 0},
		Fov:      math.Pi / 4,
		Aspect:   float32(state.WindowWidth) / float32(state.WindowHeight),
		Near:     1,
		Far:      100,
	}

	cmd.AddEntity(
		mesh,
		material,
		MyMaterial{Texture: textureId, Uniforms: MyUniforms{mgl32.Mat4{}, [4]float32{1, .25, .25, 1}}},
		camera,
		TransformComponent{
			Position: mgl32.Vec3{2, 2, 2},
			Rotation: 0.0,
			Scale:    mgl32.Vec3{1, 1, 1},
		},
	)

	cmd.AddEntity(
		mesh,
		material,
		MyMaterial{Texture: textureId, Uniforms: MyUniforms{mgl32.Mat4{}, [4]float32{.25, 1, .25, 1}}},
		camera,
		TransformComponent{
			Position: mgl32.Vec3{0, 0, 0},
			Rotation: 0.0,
			Scale:    mgl32.Vec3{1, 1, 1},
		},
		Rotating{},
	)
}

func updateCamera(cmd *Commands, input *Input) {
	MakeQuery1[CameraComponent](cmd).Map(
		func(entityId EntityId, camera *CameraComponent) bool {
			var x float32 = 0.0
			var y float32 = 0.0
			var z float32 = 0.0
			if input.Pressed[KeyRight] {
				x = x + 0.1
			}
			if input.Pressed[KeyLeft] {
				x = x - 0.1
			}
			if input.Pressed[KeyDown] {
				y = y - 0.1
			}
			if input.Pressed[KeyUp] {
				y = y + 0.1
			}
			if input.Pressed[KeySpace] {
				z = z + 0.1
			}
			if input.Pressed[KeyEnter] {
				z = z - 0.1
			}
			update := mgl32.Vec3{x, y, z}
			camera.Position = camera.Position.Add(update)
			return true
		})
}

func updateUniforms(cmd *Commands, time *Time) {
	MakeQuery4[CameraComponent, TransformComponent, MyMaterial, Rotating](cmd).Map(
		func(entityId EntityId, camera *CameraComponent, transform *TransformComponent, mat *MyMaterial, rotating *Rotating) bool {
			if nil != rotating {
				transform.Rotation += float32(1.6 * time.Dt)
			}

			mat.Uniforms.Transform = buildMvpMatrix(camera, transform)
			return true
		}, Rotating{})
}

func buildMvpMatrix(c *CameraComponent, t *TransformComponent) mgl32.Mat4 {
	model := mgl32.Translate3D(t.Position.X(), t.Position.Y(), t.Position.Z()).
		Mul4(mgl32.HomogRotate3DZ(t.Rotation)).
		Mul4(mgl32.Scale3D(t.Scale.X(), t.Scale.Y(), t.Scale.Z()))
	view := mgl32.LookAtV(
		c.Position,
		c.LookAt,
		c.Up,
	)
	projection := mgl32.Perspective(
		c.Fov,
		c.Aspect,
		c.Near,
		c.Far,
	)
	return projection.Mul4(view).Mul4(model)
}

func main() {
	app := NewApp().
		UseStates(Startup, Quit).
		UseModules(TimeModule{}).
		UseModules(AssetServerModule{}).
		UseModules(InputModule{}).
		UseModules(ClientModule{
			WindowWidth:  1024,
			WindowHeight: 768,
			WindowTitle:  "Testing App",
		}).
		UseModules(TestModule{})

	app.Run()
}
