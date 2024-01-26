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
	"github.com/mokiat/lacking/game/timestep"
	"github.com/mokiat/lacking/ui"
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

	sunLight := c.scene.Graphics().CreateDirectionalLight(graphics.DirectionalLightInfo{
		EmitColor: dprec.NewVec3(3, 3, 3),
		EmitRange: 16000, // FIXME
	})

	lightNode := hierarchy.NewNode()
	lightRotation := dprec.QuatProd(
		dprec.RotationQuat(dprec.Degrees(-55), dprec.BasisXVec3()),
		dprec.RotationQuat(dprec.Degrees(180), dprec.BasisZVec3()),
	)
	lightPosition := dprec.QuatVec3Rotation(lightRotation, dprec.NewVec3(0.0, 0.0, 100.0))
	lightNode.SetPosition(lightPosition)
	lightNode.SetRotation(lightRotation)
	lightNode.SetTarget(game.DirectionalLightNodeTarget{
		Light:                 sunLight,
		UseOnlyParentPosition: true,
	})
	c.scene.Root().AppendChild(lightNode)

	c.camera = c.gfxScene.CreateCamera()
	c.camera.SetFoVMode(graphics.FoVModeHorizontalPlus)
	c.camera.SetFoV(sprec.Degrees(60))
	c.camera.SetAutoExposure(false)
	c.camera.SetExposure(2.0)
	c.camera.SetAutoFocus(false)
	c.gfxScene.SetActiveCamera(c.camera)

	sceneModel := c.scene.FindModel("Content")
	c.scene.Root().AppendChild(sceneModel.Root())

	cameraNode := c.scene.Root().FindNode("Camera")
	cameraNode.SetTarget(game.CameraNodeTarget{
		Camera: c.camera,
	})

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

}

func (c *PlayController) onPostUpdate(elapsedTime time.Duration) {
}
