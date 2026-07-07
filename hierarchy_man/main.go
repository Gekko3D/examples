package main

import (
	"math"

	. "github.com/gekko3d/gekko"
	"github.com/gekko3d/gekko/content"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	Startup State = iota
	Quit
)

const demoVoxelResolution float32 = 0.06
const jointMarkerForward float32 = 0.36

type ManPartRole int

const (
	RoleTorso ManPartRole = iota
	RoleHead
	RoleUpperArm
	RoleForearm
	RoleHand
	RoleThigh
	RoleShin
	RoleFoot
	RoleChain
)

type ManAnimComponent struct {
	Phase  float32
	Mirror float32
	Role   ManPartRole
}

type DemoModule struct{}

func (m DemoModule) Install(app *App, cmd *Commands) {
	app.UseSystem(
		System(setupScene).
			InStage(Prelude).
			InState(OnEnter(Startup)),
	)
	app.UseSystem(
		System(manAnimationSystem).
			InStage(Update).
			RunAlways(),
	)
	app.UseSystem(
		System(quitSystem).
			InStage(PreUpdate).
			RunAlways(),
	)
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
			WindowWidth:  1440,
			WindowHeight: 900,
			WindowTitle:  "Hierarchy Man Debug",
		},
		HierarchyModule{},
		AnimationModule{},
		FlyingCameraModule{},
		DemoModule{},
	)
	app.Run()
}

func setupScene(cmd *Commands, assets *AssetServer) {
	if setupDone {
		return
	}

	models := manModels{
		Torso:    assets.CreateCubeModel(8, 16, 4, 1),
		Head:     assets.CreateCubeModel(6, 6, 6, 1),
		UpperArm: assets.CreateCubeModel(4, 10, 4, 1),
		Forearm:  assets.CreateCubeModel(4, 9, 4, 1),
		Hand:     assets.CreateCubeModel(4, 3, 4, 1),
		Thigh:    assets.CreateCubeModel(4, 10, 4, 1),
		Shin:     assets.CreateCubeModel(4, 10, 4, 1),
		Foot:     assets.CreateCubeModel(7, 3, 10, 1),
		Link:     assets.CreateCubeModel(4, 4, 4, 1),
		Joint:    assets.CreateCubeModel(2, 2, 2, 1),
		Pole:     assets.CreateCubeModel(1, 44, 1, 1),
	}
	palettes := manPalettes{
		Torso:  assets.CreateSimplePalette([4]uint8{80, 150, 240, 255}),
		Head:   assets.CreateSimplePalette([4]uint8{238, 202, 150, 255}),
		Left:   assets.CreateSimplePalette([4]uint8{230, 90, 90, 255}),
		Right:  assets.CreateSimplePalette([4]uint8{80, 210, 140, 255}),
		Legs:   assets.CreateSimplePalette([4]uint8{110, 95, 220, 255}),
		Feet:   assets.CreateSimplePalette([4]uint8{245, 185, 70, 255}),
		ChainA: assets.CreateSimplePalette([4]uint8{255, 255, 255, 255}),
		ChainB: assets.CreateSimplePalette([4]uint8{35, 35, 35, 255}),
		Joint:  assets.CreateSimplePalette([4]uint8{255, 255, 255, 255}),
		Pole:   assets.CreateSimplePalette([4]uint8{20, 20, 24, 255}),
	}

	addCamera(cmd)
	addLighting(cmd)
	addFloor(cmd, assets)

	spawnReferencePole(cmd, models, palettes, mgl32.Vec3{-4.5, 1.32, -0.42})
	spawnReferencePole(cmd, models, palettes, mgl32.Vec3{0, 1.32, -0.42})
	spawnReferencePole(cmd, models, palettes, mgl32.Vec3{4.5, 1.32, -0.42})
	spawnRawMan(cmd, models, palettes, mgl32.Vec3{-4.5, 0, 0}, false)
	spawnRawMan(cmd, models, palettes, mgl32.Vec3{0, 0, 0}, true)
	spawnAuthoredMan(cmd, assets, models, palettes, mgl32.Vec3{4.5, 0, 0})
	spawnDeepChain(cmd, models, palettes, mgl32.Vec3{8.2, 0.4, -2.4})

	setupDone = true
}

func addCamera(cmd *Commands) {
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{0, 4.2, 12}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&CameraComponent{
			Position: mgl32.Vec3{0, 4.2, 12},
			LookAt:   mgl32.Vec3{0, 1.4, 0},
			Up:       mgl32.Vec3{0, 1, 0},
			Yaw:      0,
			Pitch:    -12,
			Fov:      52,
			Aspect:   1440.0 / 900.0,
			Near:     0.1,
			Far:      500,
		},
		&FlyingCameraComponent{Speed: 8, Sensitivity: 0.1},
	)
}

func addLighting(cmd *Commands) {
	cmd.AddEntity(&LightComponent{Type: LightTypeAmbient, Intensity: 0.18, Color: [3]float32{0.8, 0.86, 1}})
	cmd.AddEntity(&LightComponent{
		Type:         LightTypeDirectional,
		Color:        [3]float32{1, 0.94, 0.82},
		Intensity:    3.2,
		CastsShadows: true,
	})
	cmd.AddEntity(&SkyboxLayerComponent{
		LayerType:  SkyboxLayerGradient,
		Resolution: [2]int{1024, 512},
		ColorA:     mgl32.Vec3{0.5, 0.65, 0.85},
		ColorB:     mgl32.Vec3{0.08, 0.1, 0.16},
		Opacity:    1,
		Priority:   0,
		Smooth:     true,
		BlendMode:  SkyboxBlendAlpha,
	})
}

func addFloor(cmd *Commands, assets *AssetServer) {
	palette := assets.CreateSimplePalette([4]uint8{74, 78, 84, 255})
	model := assets.CreateCubeModel(260, 2, 160, 1)
	cmd.AddEntity(
		&TransformComponent{Position: mgl32.Vec3{-7.8, -0.12, -4.8}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: model, VoxelPalette: palette, VoxelResolution: demoVoxelResolution, PivotMode: PivotModeCorner},
	)
}

type manModels struct {
	Torso    AssetId
	Head     AssetId
	UpperArm AssetId
	Forearm  AssetId
	Hand     AssetId
	Thigh    AssetId
	Shin     AssetId
	Foot     AssetId
	Link     AssetId
	Joint    AssetId
	Pole     AssetId
}

type manPalettes struct {
	Torso  AssetId
	Head   AssetId
	Left   AssetId
	Right  AssetId
	Legs   AssetId
	Feet   AssetId
	ChainA AssetId
	ChainB AssetId
	Joint  AssetId
	Pole   AssetId
}

type partDef struct {
	Name     string
	Role     ManPartRole
	Model    AssetId
	Palette  AssetId
	Position mgl32.Vec3
	Scale    mgl32.Vec3
	Mirror   float32
}

func spawnRawMan(cmd *Commands, models manModels, palettes manPalettes, origin mgl32.Vec3, animated bool) EntityId {
	root := cmd.AddEntity(
		&TransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LocalTransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
	)
	spawnJointMarker(cmd, root, models, palettes, mgl32.Vec3{0, 1.35, 0})

	torso := spawnPart(cmd, root, partDef{"torso", RoleTorso, models.Torso, palettes.Torso, mgl32.Vec3{0, 1.35, 0}, mgl32.Vec3{1, 1, 1}, 0}, animated)
	spawnJointMarker(cmd, torso, models, palettes, mgl32.Vec3{0, 0.78, 0})
	spawnPart(cmd, torso, partDef{"head", RoleHead, models.Head, palettes.Head, mgl32.Vec3{0, 0.78, 0}, mgl32.Vec3{1, 1, 1}, 0}, animated)

	leftShoulder := spawnJoint(cmd, torso, mgl32.Vec3{-0.42, 0.42, 0}, RoleUpperArm, -1, animated)
	spawnJointMarker(cmd, leftShoulder, models, palettes, mgl32.Vec3{0, 0, 0})
	leftUpper := spawnPart(cmd, leftShoulder, partDef{"left_upper_arm", RoleUpperArm, models.UpperArm, palettes.Left, mgl32.Vec3{0, -0.33, 0}, mgl32.Vec3{1, 1, 1}, -1}, animated)
	spawnJointMarker(cmd, leftUpper, models, palettes, mgl32.Vec3{0, -0.58, 0})
	leftForearm := spawnPart(cmd, leftUpper, partDef{"left_forearm", RoleForearm, models.Forearm, palettes.Left, mgl32.Vec3{0, -0.58, 0}, mgl32.Vec3{1, 1, 1}, -1}, animated)
	spawnJointMarker(cmd, leftForearm, models, palettes, mgl32.Vec3{0, -0.36, 0})
	spawnPart(cmd, leftForearm, partDef{"left_hand", RoleHand, models.Hand, palettes.Left, mgl32.Vec3{0, -0.36, 0}, mgl32.Vec3{1, 1, 1}, -1}, animated)

	rightShoulder := spawnJoint(cmd, torso, mgl32.Vec3{0.42, 0.42, 0}, RoleUpperArm, 1, animated)
	spawnJointMarker(cmd, rightShoulder, models, palettes, mgl32.Vec3{0, 0, 0})
	rightUpper := spawnPart(cmd, rightShoulder, partDef{"right_upper_arm", RoleUpperArm, models.UpperArm, palettes.Right, mgl32.Vec3{0, -0.33, 0}, mgl32.Vec3{1, 1, 1}, 1}, animated)
	spawnJointMarker(cmd, rightUpper, models, palettes, mgl32.Vec3{0, -0.58, 0})
	rightForearm := spawnPart(cmd, rightUpper, partDef{"right_forearm", RoleForearm, models.Forearm, palettes.Right, mgl32.Vec3{0, -0.58, 0}, mgl32.Vec3{1, 1, 1}, 1}, animated)
	spawnJointMarker(cmd, rightForearm, models, palettes, mgl32.Vec3{0, -0.36, 0})
	spawnPart(cmd, rightForearm, partDef{"right_hand", RoleHand, models.Hand, palettes.Right, mgl32.Vec3{0, -0.36, 0}, mgl32.Vec3{1, 1, 1}, 1}, animated)

	leftHip := spawnJoint(cmd, torso, mgl32.Vec3{-0.2, -0.62, 0}, RoleThigh, -1, animated)
	spawnJointMarker(cmd, leftHip, models, palettes, mgl32.Vec3{0, 0, 0})
	leftThigh := spawnPart(cmd, leftHip, partDef{"left_thigh", RoleThigh, models.Thigh, palettes.Legs, mgl32.Vec3{0, -0.38, 0}, mgl32.Vec3{1, 1, 1}, -1}, animated)
	spawnJointMarker(cmd, leftThigh, models, palettes, mgl32.Vec3{0, -0.64, 0})
	leftShin := spawnPart(cmd, leftThigh, partDef{"left_shin", RoleShin, models.Shin, palettes.Legs, mgl32.Vec3{0, -0.64, 0}, mgl32.Vec3{1, 1, 1}, -1}, animated)
	spawnJointMarker(cmd, leftShin, models, palettes, mgl32.Vec3{0, -0.42, 0.12})
	spawnPart(cmd, leftShin, partDef{"left_foot", RoleFoot, models.Foot, palettes.Feet, mgl32.Vec3{0, -0.42, 0.12}, mgl32.Vec3{1, 1, 1}, -1}, animated)

	rightHip := spawnJoint(cmd, torso, mgl32.Vec3{0.2, -0.62, 0}, RoleThigh, 1, animated)
	spawnJointMarker(cmd, rightHip, models, palettes, mgl32.Vec3{0, 0, 0})
	rightThigh := spawnPart(cmd, rightHip, partDef{"right_thigh", RoleThigh, models.Thigh, palettes.Legs, mgl32.Vec3{0, -0.38, 0}, mgl32.Vec3{1, 1, 1}, 1}, animated)
	spawnJointMarker(cmd, rightThigh, models, palettes, mgl32.Vec3{0, -0.64, 0})
	rightShin := spawnPart(cmd, rightThigh, partDef{"right_shin", RoleShin, models.Shin, palettes.Legs, mgl32.Vec3{0, -0.64, 0}, mgl32.Vec3{1, 1, 1}, 1}, animated)
	spawnJointMarker(cmd, rightShin, models, palettes, mgl32.Vec3{0, -0.42, 0.12})
	spawnPart(cmd, rightShin, partDef{"right_foot", RoleFoot, models.Foot, palettes.Feet, mgl32.Vec3{0, -0.42, 0.12}, mgl32.Vec3{1, 1, 1}, 1}, animated)

	return root
}

func spawnReferencePole(cmd *Commands, models manModels, palettes manPalettes, pos mgl32.Vec3) {
	cmd.AddEntity(
		&TransformComponent{Position: pos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&VoxelModelComponent{VoxelModel: models.Pole, VoxelPalette: palettes.Pole, VoxelResolution: demoVoxelResolution, DisableShadows: true},
	)
}

func spawnJointMarker(cmd *Commands, parent EntityId, models manModels, palettes manPalettes, pos mgl32.Vec3) {
	markerPos := pos.Add(mgl32.Vec3{0, 0, jointMarkerForward})
	cmd.AddEntity(
		&LocalTransformComponent{Position: markerPos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: markerPos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: parent},
		&VoxelModelComponent{VoxelModel: models.Joint, VoxelPalette: palettes.Joint, VoxelResolution: demoVoxelResolution, DisableShadows: true},
	)
}

func spawnJoint(cmd *Commands, parent EntityId, pos mgl32.Vec3, role ManPartRole, mirror float32, animated bool) EntityId {
	comps := []any{
		&LocalTransformComponent{Position: pos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&TransformComponent{Position: pos, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&Parent{Entity: parent},
	}
	if animated {
		comps = append(comps, &ManAnimComponent{Role: role, Mirror: mirror})
	}
	return cmd.AddEntity(comps...)
}

func spawnPart(cmd *Commands, parent EntityId, def partDef, animated bool) EntityId {
	comps := []any{
		&LocalTransformComponent{Position: def.Position, Rotation: mgl32.QuatIdent(), Scale: def.Scale},
		&TransformComponent{Position: def.Position, Rotation: mgl32.QuatIdent(), Scale: def.Scale},
		&Parent{Entity: parent},
		&VoxelModelComponent{VoxelModel: def.Model, VoxelPalette: def.Palette, VoxelResolution: demoVoxelResolution},
	}
	if animated {
		comps = append(comps, &ManAnimComponent{Role: def.Role, Mirror: def.Mirror})
	}
	return cmd.AddEntity(comps...)
}

func spawnAuthoredMan(cmd *Commands, assets *AssetServer, models manModels, palettes manPalettes, origin mgl32.Vec3) {
	def := authoredManAsset(models, palettes)
	_, _ = SpawnAuthoredAsset(cmd, assets, def, TransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}})
}

func authoredManAsset(models manModels, palettes manPalettes) *content.AssetDef {
	def := content.NewAssetDef("procedural_hierarchy_man")
	def.Runtime = &content.AssetRuntimeDef{CollapseVoxelParts: false}
	def.Materials = []content.AssetMaterialDef{
		{ID: "torso", Name: "torso", BaseColor: [4]uint8{80, 150, 240, 255}, Roughness: 0.55, IOR: 1.5},
		{ID: "head", Name: "head", BaseColor: [4]uint8{238, 202, 150, 255}, Roughness: 0.55, IOR: 1.5},
		{ID: "left", Name: "left", BaseColor: [4]uint8{230, 90, 90, 255}, Roughness: 0.55, IOR: 1.5},
		{ID: "right", Name: "right", BaseColor: [4]uint8{80, 210, 140, 255}, Roughness: 0.55, IOR: 1.5},
		{ID: "legs", Name: "legs", BaseColor: [4]uint8{110, 95, 220, 255}, Roughness: 0.55, IOR: 1.5},
		{ID: "feet", Name: "feet", BaseColor: [4]uint8{245, 185, 70, 255}, Roughness: 0.55, IOR: 1.5},
		{ID: "joint", Name: "joint", BaseColor: [4]uint8{255, 255, 255, 255}, Roughness: 0.5, IOR: 1.5},
	}
	addAssetPart := func(id, parent string, model AssetId, materialID string, pos content.Vec3, scale content.Vec3) {
		def.Parts = append(def.Parts, content.AssetPartDef{
			ID:       id,
			Name:     id,
			ParentID: parent,
			Transform: content.AssetTransformDef{
				Position: pos,
				Rotation: content.Quat{0, 0, 0, 1},
				Scale:    scale,
			},
			Source: content.AssetSourceDef{Kind: content.AssetSourceKindProceduralPrimitive, Primitive: "cube", Params: map[string]float32{
				"sx": modelSizeForAsset(model, models, "x"),
				"sy": modelSizeForAsset(model, models, "y"),
				"sz": modelSizeForAsset(model, models, "z"),
			}, MaterialID: materialID},
			VoxelResolution: demoVoxelResolution,
		})
	}

	addAssetPart("torso", "", models.Torso, "torso", content.Vec3{0, 1.35, 0}, content.Vec3{1, 1, 1})
	addAssetPart("head", "torso", models.Head, "head", content.Vec3{0, 0.78, 0}, content.Vec3{1, 1, 1})
	addAssetPart("left_upper_arm", "torso", models.UpperArm, "left", content.Vec3{-0.42, 0.09, 0}, content.Vec3{1, 1, 1})
	addAssetPart("left_forearm", "left_upper_arm", models.Forearm, "left", content.Vec3{0, -0.58, 0}, content.Vec3{1, 1, 1})
	addAssetPart("left_hand", "left_forearm", models.Hand, "left", content.Vec3{0, -0.36, 0}, content.Vec3{1, 1, 1})
	addAssetPart("right_upper_arm", "torso", models.UpperArm, "right", content.Vec3{0.42, 0.09, 0}, content.Vec3{1, 1, 1})
	addAssetPart("right_forearm", "right_upper_arm", models.Forearm, "right", content.Vec3{0, -0.58, 0}, content.Vec3{1, 1, 1})
	addAssetPart("right_hand", "right_forearm", models.Hand, "right", content.Vec3{0, -0.36, 0}, content.Vec3{1, 1, 1})
	addAssetPart("left_thigh", "torso", models.Thigh, "legs", content.Vec3{-0.2, -1.0, 0}, content.Vec3{1, 1, 1})
	addAssetPart("left_shin", "left_thigh", models.Shin, "legs", content.Vec3{0, -0.64, 0}, content.Vec3{1, 1, 1})
	addAssetPart("left_foot", "left_shin", models.Foot, "feet", content.Vec3{0, -0.42, 0.12}, content.Vec3{1, 1, 1})
	addAssetPart("right_thigh", "torso", models.Thigh, "legs", content.Vec3{0.2, -1.0, 0}, content.Vec3{1, 1, 1})
	addAssetPart("right_shin", "right_thigh", models.Shin, "legs", content.Vec3{0, -0.64, 0}, content.Vec3{1, 1, 1})
	addAssetPart("right_foot", "right_shin", models.Foot, "feet", content.Vec3{0, -0.42, 0.12}, content.Vec3{1, 1, 1})
	addAssetPart("joint_root", "", models.Joint, "joint", markerVec(0, 1.35, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_head", "torso", models.Joint, "joint", markerVec(0, 0.78, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_left_shoulder", "torso", models.Joint, "joint", markerVec(-0.42, 0.42, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_left_elbow", "left_upper_arm", models.Joint, "joint", markerVec(0, -0.58, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_left_wrist", "left_forearm", models.Joint, "joint", markerVec(0, -0.36, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_right_shoulder", "torso", models.Joint, "joint", markerVec(0.42, 0.42, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_right_elbow", "right_upper_arm", models.Joint, "joint", markerVec(0, -0.58, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_right_wrist", "right_forearm", models.Joint, "joint", markerVec(0, -0.36, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_left_hip", "torso", models.Joint, "joint", markerVec(-0.2, -0.62, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_left_knee", "left_thigh", models.Joint, "joint", markerVec(0, -0.64, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_left_ankle", "left_shin", models.Joint, "joint", markerVec(0, -0.42, 0.12), content.Vec3{1, 1, 1})
	addAssetPart("joint_right_hip", "torso", models.Joint, "joint", markerVec(0.2, -0.62, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_right_knee", "right_thigh", models.Joint, "joint", markerVec(0, -0.64, 0), content.Vec3{1, 1, 1})
	addAssetPart("joint_right_ankle", "right_shin", models.Joint, "joint", markerVec(0, -0.42, 0.12), content.Vec3{1, 1, 1})

	def.AnimationClips = []content.AssetAnimationClipDef{authoredIdleClip()}
	return def
}

func markerVec(x, y, z float32) content.Vec3 {
	return content.Vec3{x, y, z + jointMarkerForward}
}

func modelSizeForAsset(model AssetId, models manModels, axis string) float32 {
	switch model {
	case models.Torso:
		return axisValue(axis, 8, 16, 4)
	case models.Head:
		return axisValue(axis, 6, 6, 6)
	case models.UpperArm:
		return axisValue(axis, 4, 10, 4)
	case models.Forearm:
		return axisValue(axis, 4, 9, 4)
	case models.Hand:
		return axisValue(axis, 4, 3, 4)
	case models.Thigh:
		return axisValue(axis, 4, 10, 4)
	case models.Shin:
		return axisValue(axis, 4, 10, 4)
	case models.Foot:
		return axisValue(axis, 7, 3, 10)
	case models.Joint:
		return 2
	default:
		return 4
	}
}

func axisValue(axis string, x, y, z float32) float32 {
	switch axis {
	case "x":
		return x
	case "y":
		return y
	default:
		return z
	}
}

func authoredIdleClip() content.AssetAnimationClipDef {
	return content.AssetAnimationClipDef{
		ID:       "idle",
		Name:     "idle",
		FPS:      30,
		Duration: 2,
		Loop:     true,
		Tracks: []content.AssetAnimationTrackDef{
			rotationTrack("left_upper_arm", 0.22, 0.5, 1.0, -1),
			rotationTrack("left_forearm", 0.16, 0.25, 0.75, -1),
			rotationTrack("right_upper_arm", -0.22, -0.5, -1.0, 1),
			rotationTrack("right_forearm", -0.16, -0.25, -0.75, 1),
			rotationTrack("left_thigh", -0.12, -0.32, -0.8, -1),
			rotationTrack("left_shin", 0.18, 0.42, 0.6, -1),
			rotationTrack("right_thigh", 0.12, 0.32, 0.8, 1),
			rotationTrack("right_shin", -0.18, -0.42, -0.6, 1),
		},
	}
}

func rotationTrack(id string, a, b, c float32, mirror float32) content.AssetAnimationTrackDef {
	return content.AssetAnimationTrackDef{
		TargetID: id,
		RotationKeys: []content.AssetQuatKeyDef{
			{Time: 0, Value: quatContent(mgl32.QuatRotate(a*mirror, mgl32.Vec3{1, 0, 0}))},
			{Time: 1, Value: quatContent(mgl32.QuatRotate(b*mirror, mgl32.Vec3{1, 0, 0}).Mul(mgl32.QuatRotate(c*0.25, mgl32.Vec3{0, 0, 1})).Normalize())},
			{Time: 2, Value: quatContent(mgl32.QuatRotate(a*mirror, mgl32.Vec3{1, 0, 0}))},
		},
	}
}

func quatContent(q mgl32.Quat) content.Quat {
	q = q.Normalize()
	return content.Quat{q.V.X(), q.V.Y(), q.V.Z(), q.W}
}

func spawnDeepChain(cmd *Commands, models manModels, palettes manPalettes, origin mgl32.Vec3) {
	parent := cmd.AddEntity(
		&TransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&LocalTransformComponent{Position: origin, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
		&ManAnimComponent{Role: RoleChain},
	)
	for i := 0; i < 14; i++ {
		palette := palettes.ChainA
		if i%2 == 1 {
			palette = palettes.ChainB
		}
		parent = cmd.AddEntity(
			&LocalTransformComponent{Position: mgl32.Vec3{0, 0.3, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&TransformComponent{Position: mgl32.Vec3{0, 0.3, 0}, Rotation: mgl32.QuatIdent(), Scale: mgl32.Vec3{1, 1, 1}},
			&Parent{Entity: parent},
			&VoxelModelComponent{VoxelModel: models.Link, VoxelPalette: palette, VoxelResolution: demoVoxelResolution},
		)
	}
}

func manAnimationSystem(cmd *Commands, time *Time) {
	if time == nil {
		return
	}
	dt := float32(time.Dt)
	MakeQuery2[LocalTransformComponent, ManAnimComponent](cmd).Map(func(_ EntityId, local *LocalTransformComponent, anim *ManAnimComponent) bool {
		anim.Phase += dt
		phase := float32(math.Sin(float64(anim.Phase * 2)))
		switch anim.Role {
		case RoleUpperArm:
			local.Rotation = mgl32.QuatRotate((0.35+0.25*phase)*anim.Mirror, mgl32.Vec3{1, 0, 0}).Mul(mgl32.QuatRotate(0.22*anim.Mirror, mgl32.Vec3{0, 0, 1})).Normalize()
		case RoleForearm:
			local.Rotation = mgl32.QuatRotate((0.25+0.16*phase)*anim.Mirror, mgl32.Vec3{1, 0, 0})
		case RoleThigh:
			local.Rotation = mgl32.QuatRotate((0.2-0.18*phase)*anim.Mirror, mgl32.Vec3{1, 0, 0})
		case RoleShin:
			local.Rotation = mgl32.QuatRotate((-0.28+0.12*phase)*anim.Mirror, mgl32.Vec3{1, 0, 0})
		case RoleFoot:
			local.Rotation = mgl32.QuatRotate(0.08*phase*anim.Mirror, mgl32.Vec3{1, 0, 0})
		case RoleHead:
			local.Rotation = mgl32.QuatRotate(0.08*phase, mgl32.Vec3{0, 1, 0})
		case RoleChain:
			local.Rotation = mgl32.QuatRotate(0.45*phase, mgl32.Vec3{0, 0, 1})
		}
		return true
	})
}

func quitSystem(cmd *Commands, input *Input) {
	if input != nil && input.Pressed[KeyEscape] {
		cmd.ChangeState(Quit)
	}
}
