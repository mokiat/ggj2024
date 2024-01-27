package controller

import (
	"runtime"
	"time"

	"github.com/mokiat/ggj2024/internal/game/data"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/acceleration"
	"github.com/mokiat/lacking/game/preset"
	"github.com/mokiat/lacking/game/timestep"
	"github.com/mokiat/lacking/ui"
)

const (
	anchorDistance = 6.0
	cameraDistance = 11.0 * 3
)

func NewPlayController(window app.Window, audioAPI audio.API, engine *game.Engine, playData *data.PlayData) *PlayController {
	return &PlayController{
		window:   window,
		audioAPI: audioAPI,
		engine:   engine,
		playData: playData,
	}
}

type PlayController struct {
	window   app.Window
	audioAPI audio.API
	engine   *game.Engine
	playData *data.PlayData

	preUpdateSubscription  *timestep.UpdateSubscription
	postUpdateSubscription *timestep.UpdateSubscription

	scene        *game.Scene
	gfxScene     *graphics.Scene
	physicsScene *physics.Scene
	ecsScene     *ecs.Scene

	followCameraSystem        *preset.FollowCameraSystem
	airplaneGamepadController *AirplaneGamepadController
	// TODO: airplaneMouseController *AirplaneMouseController

	airplane *Airplane

	binNode *hierarchy.Node
	camera  *graphics.Camera

	soundtrackPlayback audio.Playback
}

func (c *PlayController) Start() {
	c.scene = c.engine.CreateScene()
	c.scene.Initialize(c.playData.Scene)

	c.binNode = hierarchy.NewNode()
	c.binNode.SetAbsoluteMatrix(dprec.TranslationMat4(
		0.0, 0.0, 100.0,
	))
	c.scene.Root().AppendChild(c.binNode)

	c.preUpdateSubscription = c.scene.SubscribePreUpdate(c.onPreUpdate)
	c.postUpdateSubscription = c.scene.SubscribePostUpdate(c.onPostUpdate)

	c.gfxScene = c.scene.Graphics()
	c.physicsScene = c.scene.Physics()
	c.ecsScene = c.scene.ECS()

	sceneModel := c.scene.FindModel("Content")
	c.scene.Root().AppendChild(sceneModel.Root())

	c.followCameraSystem = preset.NewFollowCameraSystem(c.ecsScene, c.window)
	c.followCameraSystem.UseDefaults()

	c.physicsScene.CreateGlobalAccelerator(acceleration.NewGravityDirection())

	airplanePosition := dprec.NewVec3(0.0, 100.0, 0.0)
	airplaneModel := c.scene.CreateModel(game.ModelInfo{
		Definition:        c.playData.Airplane,
		Name:              "Airplane",
		Position:          airplanePosition,
		Rotation:          dprec.IdentityQuat(),
		Scale:             dprec.NewVec3(1.0, 1.0, 1.0),
		IsDynamic:         true,
		PrepareAnimations: true,
	})
	c.airplane = NewAirplane(c.physicsScene, c.ecsScene, airplaneModel, airplanePosition)
	airplaneNode := c.airplane.Node

	gamepad := c.window.Gamepads()[0]
	if gamepad.Connected() && gamepad.Supported() {
		c.airplaneGamepadController = NewAirplaneGamepadController(c.airplane, gamepad)
	}

	c.camera = c.gfxScene.CreateCamera()
	c.camera.SetFoVMode(graphics.FoVModeHorizontalPlus)
	c.camera.SetFoV(sprec.Degrees(60))
	c.camera.SetAutoExposure(false)
	c.camera.SetExposure(2.0)
	c.camera.SetAutoFocus(false)
	c.gfxScene.SetActiveCamera(c.camera)

	cameraNode := c.scene.Root().FindNode("Camera")
	cameraNode.SetTarget(game.CameraNodeTarget{
		Camera: c.camera,
	})

	cameraEntity := c.ecsScene.CreateEntity()
	ecs.AttachComponent(cameraEntity, &preset.NodeComponent{
		Node: cameraNode,
	})
	ecs.AttachComponent(cameraEntity, &preset.FollowCameraComponent{
		Target:         airplaneNode,
		AnchorPosition: dprec.Vec3Sum(airplaneNode.Position(), dprec.NewVec3(0.0, 2.0, -cameraDistance)),
		AnchorDistance: anchorDistance,
		CameraDistance: cameraDistance,
		PitchAngle:     dprec.Degrees(-25),
		YawAngle:       dprec.Degrees(0),
		Zoom:           1.0,
	})

	lightNode := c.scene.Root().FindNode("Light")
	lightNode.UseTransformation(func(node *hierarchy.Node) dprec.Mat4 {
		base := node.BaseAbsoluteMatrix()
		// Remove parent's rotation
		base.M11 = 1.0
		base.M12 = 0.0
		base.M13 = 0.0
		base.M21 = 0.0
		base.M22 = 1.0
		base.M23 = 0.0
		base.M31 = 0.0
		base.M32 = 0.0
		base.M33 = 1.0
		return dprec.Mat4Prod(base, node.Matrix())
	})
	airplaneNode.AppendChild(lightNode)

	runtime.GC()
	c.engine.ResetDeltaTime()
	c.engine.SetActiveScene(c.scene)

	// c.soundtrackPlayback = c.audioAPI.Play(c.playData.Soundtrack, audio.PlayInfo{
	// 	Loop: true,
	// })
}

func (c *PlayController) Stop() {
	// c.soundtrackPlayback.Stop()
	c.engine.SetActiveScene(nil)
	c.preUpdateSubscription.Delete()
	c.postUpdateSubscription.Delete()
	c.scene.Delete()
}

func (c *PlayController) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	return false
}

func (c *PlayController) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	return false
}

func (c *PlayController) onPreUpdate(elapsedTime time.Duration) {
	if c.airplaneGamepadController != nil {
		c.airplaneGamepadController.Update(elapsedTime.Seconds())
	}
	c.airplane.UpdatePhysics(elapsedTime.Seconds())
}

func (c *PlayController) onPostUpdate(elapsedTime time.Duration) {
	c.followCameraSystem.Update(elapsedTime.Seconds())
}
