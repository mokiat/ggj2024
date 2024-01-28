package controller

import (
	"math/rand"
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

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	anchorDistance = 6.0
	cameraDistance = 11.0 * 5
)

const (
	defeatAfter = 120 * time.Second
)

func NewPlayController(window app.Window, audioAPI audio.API, engine *game.Engine, playData *data.PlayData) *PlayController {
	return &PlayController{
		window:   window,
		audioAPI: audioAPI,
		engine:   engine,
		playData: playData,

		lastRubbingTime: time.Now().Add(-time.Minute),

		introAfter: 1 * time.Second,
		pilotAfter: 90 * time.Second,
		towerAfter: 45 * time.Second,
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

	followCameraSystem         *preset.FollowCameraSystem
	airplaneGamepadController  *AirplaneGamepadController
	airplaneKeyboardController *AirplaneKeyboardController

	airplane   *Airplane
	ball       *Ball
	cowSpawner *CowSpawner
	cows       []*Cow

	binNode *hierarchy.Node
	camera  *graphics.Camera

	soundtrackPlayback audio.Playback
	popSound           audio.Media
	rubbingSound       audio.Media
	lastRubbingTime    time.Time

	introSound audio.Media
	introAfter time.Duration
	pilotSound audio.Media
	pilotAfter time.Duration
	towerSound audio.Media
	towerAfter time.Duration

	gameTime time.Duration

	onVictory func(time.Duration)
	onDefeat  func(int)
}

func (c *PlayController) Start(onVictory func(time.Duration), onDefeat func(int)) {
	c.onVictory = onVictory
	c.onDefeat = onDefeat

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

	ballModel := c.scene.CreateModel(game.ModelInfo{
		Definition:        c.playData.Ball,
		Name:              "Ball",
		Position:          dprec.ZeroVec3(),
		Rotation:          dprec.IdentityQuat(),
		Scale:             dprec.NewVec3(1.0, 1.0, 1.0),
		IsDynamic:         true,
		PrepareAnimations: true,
	})
	c.ball = NewBall(c.physicsScene, c.airplane, ballModel)

	c.airplaneKeyboardController = NewAirplaneKeyboardController(c.airplane)

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
	cameraNode.SetPosition(dprec.Vec3Sum(
		c.airplane.Body.Position(),
		dprec.NewVec3(0.0, 50.0, -cameraDistance),
	))
	cameraNode.ApplyToTarget(false)

	cameraEntity := c.ecsScene.CreateEntity()
	ecs.AttachComponent(cameraEntity, &preset.NodeComponent{
		Node: cameraNode,
	})
	targetNode := c.airplane.Node
	ecs.AttachComponent(cameraEntity, &preset.FollowCameraComponent{
		Target:         targetNode,
		AnchorPosition: dprec.Vec3Sum(c.airplane.Body.Position(), dprec.NewVec3(0.0, 2.0, -cameraDistance)),
		AnchorDistance: anchorDistance,
		CameraDistance: cameraDistance,
		PitchAngle:     dprec.Degrees(-30),
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
	c.airplane.Node.AppendChild(lightNode)

	c.cowSpawner = NewCowSpawner(c.scene, c.playData.Cow, c.playData.Burst)

	for i := 0; i < 100; i++ {
		cow := c.cowSpawner.SpawnCow(dprec.NewVec3(
			(random.Float64()*2.0-1.0)*400.0,
			random.Float64()*200.0,
			(random.Float64()*2.0-1.0)*400.0,
		))
		c.cows = append(c.cows, cow)
	}

	runtime.GC()
	c.engine.ResetDeltaTime()
	c.engine.SetActiveScene(c.scene)

	c.soundtrackPlayback = c.audioAPI.Play(c.playData.Soundtrack, audio.PlayInfo{
		Gain: 1.0,
		Loop: true,
	})
	c.popSound = c.playData.Pop
	c.rubbingSound = c.playData.Rubbing
	c.introSound = c.playData.IntroSound
	c.pilotSound = c.playData.PilotSound
	c.towerSound = c.playData.TowerSound

	c.physicsScene.SubscribeDoubleBodyCollision(func(first physics.Body, second physics.Body, active bool) {
		var sourceBody physics.Body
		var targetBody physics.Body
		switch {
		case first == c.ball.Body || first == c.airplane.Body:
			sourceBody = first
			targetBody = second
		case second == c.ball.Body || second == c.airplane.Body:
			sourceBody = second
			targetBody = first
		default:
			return
		}

		for _, cow := range c.cows {
			if !cow.Active {
				continue
			}
			if cow.Body == targetBody {
				if sourceBody == c.ball.Body {
					c.audioAPI.Play(c.popSound, audio.PlayInfo{
						Gain: 1.0,
					})
					cow.Burst(c.scene)
				}
				if sourceBody == c.airplane.Body && time.Since(c.lastRubbingTime) > time.Second {
					c.audioAPI.Play(c.rubbingSound, audio.PlayInfo{
						Gain: 1.0,
					})
					c.lastRubbingTime = time.Now()
				}
			}
		}
	})
}

func (c *PlayController) Freeze() {
	c.scene.Freeze()
}

func (c *PlayController) Stop() {
	c.soundtrackPlayback.Stop()
	c.engine.SetActiveScene(nil)
	c.preUpdateSubscription.Delete()
	c.postUpdateSubscription.Delete()
	c.scene.Delete()
}

func (c *PlayController) CowsRemaining() int {
	return c.remainingCows()
}

func (c *PlayController) RemainingTime() time.Duration {
	return max(0, defeatAfter-c.gameTime)
}

func (c *PlayController) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	return false
}

func (c *PlayController) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	if c.airplaneKeyboardController == nil {
		return false
	}
	return c.airplaneKeyboardController.OnKeyboardEvent(event)
}

func (c *PlayController) onPreUpdate(elapsedTime time.Duration) {
	if c.airplaneGamepadController == nil {
		gamepad := c.window.Gamepads()[0]
		if gamepad.Connected() && gamepad.Supported() {
			c.airplaneGamepadController = NewAirplaneGamepadController(c.airplane, gamepad)
			c.airplaneKeyboardController = nil
		}
	}
	if c.airplaneGamepadController != nil {
		c.airplaneGamepadController.Update(elapsedTime.Seconds())
	}
	if c.airplaneKeyboardController != nil {
		c.airplaneKeyboardController.Update(elapsedTime.Seconds())
	}
	c.airplane.UpdatePhysics(elapsedTime.Seconds())
}

func (c *PlayController) onPostUpdate(elapsedTime time.Duration) {
	if c.onVictory == nil || c.onDefeat == nil {
		return
	}

	c.followCameraSystem.Update(elapsedTime.Seconds())
	for _, cow := range c.cows {
		cow.Update(elapsedTime)
	}
	c.introAfter -= elapsedTime
	if c.introAfter < 0 {
		c.audioAPI.Play(c.introSound, audio.PlayInfo{
			Gain: 1.0,
		})
		c.introAfter = 24 * time.Hour
	}
	c.pilotAfter -= elapsedTime
	if c.pilotAfter < 0 {
		c.audioAPI.Play(c.pilotSound, audio.PlayInfo{
			Gain: 1.0,
		})
		c.pilotAfter = 24 * time.Hour
	}
	c.towerAfter -= elapsedTime
	if c.towerAfter < 0 {
		c.audioAPI.Play(c.towerSound, audio.PlayInfo{
			Gain: 1.0,
		})
		c.towerAfter = 24 * time.Hour
	}

	countCows := c.remainingCows()

	if countCows == 0 {
		c.onVictory(c.gameTime)
		c.onVictory = nil
		return
	}

	c.gameTime += elapsedTime
	if c.gameTime > defeatAfter {
		c.onDefeat(countCows)
		c.onDefeat = nil
		return
	}
}

func (c *PlayController) remainingCows() int {
	var count int
	for _, cow := range c.cows {
		if cow.Active {
			count++
		}
	}
	return count
}
