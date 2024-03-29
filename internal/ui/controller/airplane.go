package controller

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/constraint"
	"github.com/mokiat/lacking/game/preset"
	"github.com/mokiat/lacking/ui"
)

var (
	maxAileronAngle  = dprec.Degrees(30) // TODO
	maxElevatorAngle = dprec.Degrees(30) // TODO
	maxRudderAngle   = dprec.Degrees(20) // TODO
	maxThrust        = 9.8 * 1.5         // 1.5g
	thrustRampUp     = maxThrust / 2.0
)

func NewAirplane(physicsScene *physics.Scene, ecsScene *ecs.Scene, model *game.Model, position dprec.Vec3) *Airplane {
	var (
		airplaneMass            = 1500.0
		airplaneMomentOfInertia = physics.SymmetricMomentOfInertia(1500.0 / 2.0)

		aileronMass            = 25.0
		aileronMomentOfInertia = physics.SolidSphereMomentOfInertia(50.0, 0.4)

		elevatorMass            = 50.0
		elevatorMomentOfInertia = physics.SolidSphereMomentOfInertia(100.0, 0.4)

		rudderMass            = 25.0
		rudderMomentOfInertia = physics.SolidSphereMomentOfInertia(100.0, 0.4)

		counterweigthMass            = 100.0
		counterweightMomentOfInertia = physics.SolidSphereMomentOfInertia(10.0, 0.4)
	)

	collisionGroup := physics.NewCollisionGroup()

	airplaneBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   airplaneMass,
		MomentOfInertia:        airplaneMomentOfInertia,
		DragFactor:             0.0, // TODO
		AngularDragFactor:      0.0, // TODO
		RestitutionCoefficient: 0.0,
		CollisionGroup:         collisionGroup,
		CollisionBoxes: []collision.Box{
			collision.NewBox( // body
				dprec.NewVec3(0.0, 0.0, -1.75),
				dprec.IdentityQuat(),
				dprec.NewVec3(1.6, 1.5, 16.0),
			),
			collision.NewBox( // left wing
				dprec.NewVec3(4.8, 0.0, -1.7),
				dprec.IdentityQuat(),
				dprec.NewVec3(8.0, 0.3, 2.3),
			),
			collision.NewBox( // right wing
				dprec.NewVec3(-4.8, 0.0, -1.7),
				dprec.IdentityQuat(),
				dprec.NewVec3(8.0, 0.3, 2.3),
			),
			collision.NewBox( // rear
				dprec.NewVec3(0.0, 0.75, -8.7),
				dprec.IdentityQuat(),
				dprec.NewVec3(6.6, 2.3, 2.0),
			),
		},
		AerodynamicShapes: []physics.AerodynamicShape{
			// wings
			physics.NewAerodynamicShape(
				physics.NewTransform(
					dprec.NewVec3(0.0, 0.0, -1.7),
					dprec.RotationQuat(dprec.Degrees(-5), dprec.BasisXVec3()),
				),
				physics.NewSurfaceAerodynamicShape(16.0, 0.1, 2.4),
			),
		},
	})

	aileronBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   aileronMass,
		MomentOfInertia:        aileronMomentOfInertia,
		DragFactor:             0.0, // TODO
		AngularDragFactor:      0.0, // TODO
		RestitutionCoefficient: 0.0,
		AerodynamicShapes: []physics.AerodynamicShape{
			physics.NewAerodynamicShape(
				physics.NewTransform(
					dprec.NewVec3(0.0, 0.0, -0.4),
					dprec.IdentityQuat(),
				),
				physics.NewSurfaceAerodynamicShape(3.0, 0.1, 1.1),
			),
		},
	})

	elevatorBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   elevatorMass,
		MomentOfInertia:        elevatorMomentOfInertia,
		DragFactor:             0.0, // TODO
		AngularDragFactor:      0.0, // TODO
		RestitutionCoefficient: 0.0,
		AerodynamicShapes: []physics.AerodynamicShape{
			physics.NewAerodynamicShape(
				physics.NewTransform(
					dprec.NewVec3(0.0, 0.0, -0.4),
					dprec.IdentityQuat(),
				),
				physics.NewSurfaceAerodynamicShape(4.4, 0.1, 0.8),
			),
		},
	})

	rudderBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   rudderMass,
		MomentOfInertia:        rudderMomentOfInertia,
		DragFactor:             0.0, // TODO
		AngularDragFactor:      0.0, // TODO
		RestitutionCoefficient: 0.0,
		AerodynamicShapes: []physics.AerodynamicShape{
			physics.NewAerodynamicShape(
				physics.NewTransform(
					dprec.NewVec3(0.0, 0.0, 0.0),
					dprec.RotationQuat(dprec.Degrees(90), dprec.BasisZVec3()),
				),
				physics.NewSurfaceAerodynamicShape(2.0, 0.1, 1.0),
			),
		},
	})

	counterweightBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   counterweigthMass,
		MomentOfInertia:        counterweightMomentOfInertia,
		DragFactor:             0.0,
		AngularDragFactor:      0.0,
		RestitutionCoefficient: 0.0,
	})

	airplaneNode := model.Root().FindNode("Body")
	airplaneBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       airplaneNode.Name(),
		Definition: airplaneBodyDef,
		Position:   dprec.Vec3Sum(position, airplaneNode.AbsoluteMatrix().Translation()),
		Rotation:   dprec.IdentityQuat(),
	})
	airplaneBody.SetVelocity(dprec.NewVec3(0.0, 0.0, maxThrust))
	airplaneNode.SetSource(game.BodyNodeSource{
		Body: airplaneBody,
	})

	counterweightRelativePosition := dprec.NewVec3(0.0, 0.0, 5.0)
	counterweightBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       "Counterweight",
		Definition: counterweightBodyDef,
		Position: dprec.Vec3Sum(
			airplaneBody.Position(),
			counterweightRelativePosition,
		),
		Rotation: dprec.IdentityQuat(),
	})
	physicsScene.CreateDoubleBodyConstraint(airplaneBody, counterweightBody, constraint.NewPairCombined(
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(counterweightRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisXVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(counterweightRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisZVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(counterweightRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisYVec3()).
			SetOffset(0.0),
	))

	leftAileronNode := model.Root().FindNode("LeftAileron")
	leftAileronRelativePosition := dprec.Vec3Diff(
		leftAileronNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
	leftAileronBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       leftAileronNode.Name(),
		Definition: aileronBodyDef,
		Position:   dprec.Vec3Sum(airplaneBody.Position(), leftAileronRelativePosition),
		Rotation:   dprec.IdentityQuat(),
	})
	leftAileronNode.SetSource(game.BodyNodeSource{
		Body: leftAileronBody,
	})
	leftAileronRotation := constraint.NewMatchDirections().
		SetPrimaryDirection(dprec.BasisZVec3()).
		SetSecondaryDirection(dprec.BasisZVec3())
	physicsScene.CreateDoubleBodyConstraint(airplaneBody, leftAileronBody, constraint.NewPairCombined(
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(leftAileronRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisXVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(leftAileronRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisZVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(leftAileronRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisYVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirections().
			SetPrimaryDirection(dprec.BasisXVec3()).
			SetSecondaryDirection(dprec.BasisXVec3()),
		leftAileronRotation,
	))

	rightAileronNode := model.Root().FindNode("RightAileron")
	rightAileronRelativePosition := dprec.Vec3Diff(
		rightAileronNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
	rightAileronBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       rightAileronNode.Name(),
		Definition: aileronBodyDef,
		Position:   dprec.Vec3Sum(airplaneBody.Position(), rightAileronRelativePosition),
		Rotation:   dprec.IdentityQuat(),
	})
	rightAileronNode.SetSource(game.BodyNodeSource{
		Body: rightAileronBody,
	})
	rightAileronRotation := constraint.NewMatchDirections().
		SetPrimaryDirection(dprec.BasisZVec3()).
		SetSecondaryDirection(dprec.BasisZVec3())
	physicsScene.CreateDoubleBodyConstraint(airplaneBody, rightAileronBody, constraint.NewPairCombined(
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(rightAileronRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisXVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(rightAileronRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisZVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(rightAileronRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisYVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirections().
			SetPrimaryDirection(dprec.BasisXVec3()).
			SetSecondaryDirection(dprec.BasisXVec3()),
		rightAileronRotation,
	))

	elevatorNode := model.Root().FindNode("Elevators")
	elevatorRelativePosition := dprec.Vec3Diff(
		elevatorNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
	elevatorBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       elevatorNode.Name(),
		Definition: elevatorBodyDef,
		Position:   dprec.Vec3Sum(airplaneBody.Position(), elevatorRelativePosition),
		Rotation:   dprec.IdentityQuat(),
	})
	elevatorNode.SetSource(game.BodyNodeSource{
		Body: elevatorBody,
	})
	elevatorRotation := constraint.NewMatchDirections().
		SetPrimaryDirection(dprec.BasisZVec3()).
		SetSecondaryDirection(dprec.BasisZVec3())
	physicsScene.CreateDoubleBodyConstraint(airplaneBody, elevatorBody, constraint.NewPairCombined(
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(elevatorRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisXVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(elevatorRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisZVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(elevatorRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisYVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirections().
			SetPrimaryDirection(dprec.BasisXVec3()).
			SetSecondaryDirection(dprec.BasisXVec3()),
		elevatorRotation,
	))

	rudderNode := model.FindNode("Rudder")
	rudderRelativePosition := dprec.Vec3Diff(
		rudderNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
	rudderBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       "Rudder",
		Definition: rudderBodyDef,
		Position:   dprec.Vec3Sum(airplaneBody.Position(), rudderRelativePosition),
		Rotation:   dprec.IdentityQuat(),
	})
	rudderNode.SetSource(game.BodyNodeSource{
		Body: rudderBody,
	})
	rudderRotation := constraint.NewMatchDirections().
		SetPrimaryDirection(dprec.BasisZVec3()).
		SetSecondaryDirection(dprec.BasisZVec3())
	physicsScene.CreateDoubleBodyConstraint(airplaneBody, rudderBody, constraint.NewPairCombined(
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(rudderRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisXVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(rudderRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisZVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(rudderRelativePosition).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisYVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirections().
			SetPrimaryDirection(dprec.BasisYVec3()).
			SetSecondaryDirection(dprec.BasisYVec3()),
		rudderRotation,
	))

	properllerNode := model.FindNode("Propeller")

	entity := ecsScene.CreateEntity()
	entity.SetComponent(preset.NodeComponentID, &preset.NodeComponent{
		Node: airplaneNode,
	})

	return &Airplane{
		Entity:        entity,
		Node:          airplaneNode,
		PropellerNode: properllerNode,

		CollisionGroup: collisionGroup,

		Body:                      airplaneBody,
		LeftAileronRotConstraint:  leftAileronRotation,
		RightAileronRotConstraint: rightAileronRotation,
		ElevatorRotConstraint:     elevatorRotation,
		RudderRotConstraint:       rudderRotation,

		TargetThrust: maxThrust / 1.5,
		Thrust:       maxThrust,
	}
}

type Airplane struct {
	Entity        *ecs.Entity
	Node          *hierarchy.Node
	PropellerNode *hierarchy.Node

	CollisionGroup int

	Body                      physics.Body
	LeftAileronRotConstraint  *constraint.MatchDirections
	RightAileronRotConstraint *constraint.MatchDirections
	ElevatorRotConstraint     *constraint.MatchDirections
	RudderRotConstraint       *constraint.MatchDirections
	PropellerRotConstraint    *constraint.MatchDirections

	TargetThrust  float64
	Thrust        float64
	AileronAngle  dprec.Angle
	ElevatorAngle dprec.Angle
	RudderAngle   dprec.Angle
}

func (a *Airplane) UpdatePhysics(elapsedSeconds float64) {
	if a.Thrust < a.TargetThrust {
		deltaThrust := dprec.Min(thrustRampUp*elapsedSeconds, a.TargetThrust-a.Thrust)
		a.Thrust += deltaThrust
	}
	if a.Thrust > a.TargetThrust {
		deltaThrust := dprec.Max(-thrustRampUp*elapsedSeconds, a.Thrust-a.TargetThrust)
		a.Thrust -= deltaThrust
	}

	velocity := a.Body.Velocity()
	deltaVelocity := dprec.Vec3Prod(a.Body.Rotation().OrientationZ(), a.Thrust*elapsedSeconds)
	a.Body.SetVelocity(dprec.Vec3Sum(velocity, deltaVelocity))

	leftAileronQuat := dprec.RotationQuat(-a.AileronAngle, dprec.BasisXVec3())
	rightAileronQuat := dprec.RotationQuat(a.AileronAngle, dprec.BasisXVec3())
	a.LeftAileronRotConstraint.SetPrimaryDirection(dprec.QuatVec3Rotation(leftAileronQuat, dprec.BasisZVec3()))
	a.RightAileronRotConstraint.SetPrimaryDirection(dprec.QuatVec3Rotation(rightAileronQuat, dprec.BasisZVec3()))

	elevatorQuat := dprec.RotationQuat(a.ElevatorAngle, dprec.BasisXVec3())
	a.ElevatorRotConstraint.SetPrimaryDirection(dprec.QuatVec3Rotation(elevatorQuat, dprec.BasisZVec3()))

	rudderQuat := dprec.RotationQuat(a.RudderAngle, dprec.BasisYVec3())
	a.RudderRotConstraint.SetPrimaryDirection(dprec.QuatVec3Rotation(rudderQuat, dprec.BasisZVec3()))

	rotationSpeed := 360 * (1.0 + a.Thrust) * elapsedSeconds
	rotation := dprec.RotationQuat(dprec.Degrees(rotationSpeed), dprec.BasisZVec3())
	a.PropellerNode.SetRotation(dprec.QuatProd(
		a.PropellerNode.Rotation(),
		rotation,
	))
}

func NewAirplaneGamepadController(airplane *Airplane, gamepad app.Gamepad) *AirplaneGamepadController {
	return &AirplaneGamepadController{
		airplane: airplane,
		gamepad:  gamepad,
	}
}

type AirplaneGamepadController struct {
	airplane *Airplane
	gamepad  app.Gamepad
}

func (c *AirplaneGamepadController) Update(elapsedSeconds float64) {
	// Thrust
	if c.gamepad.ActionDownButton() {
		c.airplane.TargetThrust += elapsedSeconds * maxThrust
	}
	if c.gamepad.ActionLeftButton() {
		c.airplane.TargetThrust -= elapsedSeconds * maxThrust
	}
	c.airplane.TargetThrust = dprec.Clamp(c.airplane.TargetThrust, 0.0, maxThrust)

	// Ailerons
	c.airplane.AileronAngle = dprec.Angle(c.gamepad.LeftStickX()) * maxAileronAngle

	// Elevators
	c.airplane.ElevatorAngle = dprec.Angle(c.gamepad.LeftStickY()) * maxElevatorAngle

	// Rudder
	c.airplane.RudderAngle = dprec.Angle(0)
	c.airplane.RudderAngle -= dprec.Angle(c.gamepad.LeftTrigger()) * maxRudderAngle
	c.airplane.RudderAngle += dprec.Angle(c.gamepad.RightTrigger()) * maxRudderAngle
}

func NewAirplaneKeyboardController(airplane *Airplane) *AirplaneKeyboardController {
	return &AirplaneKeyboardController{
		airplane: airplane,
	}
}

type AirplaneKeyboardController struct {
	airplane *Airplane

	targetRoll  dprec.Angle
	targetPitch dprec.Angle

	rollLeft     bool
	rollRight    bool
	pitchUp      bool
	pitchDown    bool
	throttleUp   bool
	throttleDown bool
	rudderLeft   bool
	rudderRight  bool
}

func (c *AirplaneKeyboardController) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	switch event.Code {
	case ui.KeyCodeArrowLeft:
		c.rollLeft = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeArrowRight:
		c.rollRight = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeArrowUp:
		c.pitchUp = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeArrowDown:
		c.pitchDown = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeW:
		c.throttleUp = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeS:
		c.throttleDown = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeA:
		c.rudderLeft = event.Action != ui.KeyboardActionUp
		return true
	case ui.KeyCodeD:
		c.rudderRight = event.Action != ui.KeyboardActionUp
		return true
	}
	return false
}

func (c *AirplaneKeyboardController) Update(elapsedSeconds float64) {
	var (
		rudderTurnSpeed    = 60.0
		rudderRestoreSpeed = 45.0

		pitchTurnSpeed = 60.0
		maxPitch       = dprec.Degrees(45)
		minPitch       = dprec.Degrees(-45)

		rollTurnSpeed = 90.0
		maxRoll       = dprec.Degrees(60)
	)

	if c.throttleUp {
		c.airplane.TargetThrust += elapsedSeconds * maxThrust
	}
	if c.throttleDown {
		c.airplane.TargetThrust -= elapsedSeconds * maxThrust
	}
	c.airplane.TargetThrust = dprec.Clamp(c.airplane.TargetThrust, 0.0, maxThrust)

	if c.rudderLeft {
		c.airplane.RudderAngle -= dprec.Degrees(rudderTurnSpeed * elapsedSeconds)
	}
	if c.rudderRight {
		c.airplane.RudderAngle += dprec.Degrees(rudderTurnSpeed * elapsedSeconds)
	}
	c.airplane.RudderAngle = dprec.Clamp(c.airplane.RudderAngle, -maxRudderAngle, maxRudderAngle)
	if !c.rudderLeft && !c.rudderRight {
		if c.airplane.RudderAngle > 0 {
			c.airplane.RudderAngle -= dprec.Degrees(rudderRestoreSpeed * elapsedSeconds)
			c.airplane.RudderAngle = max(c.airplane.RudderAngle, 0)
		}
		if c.airplane.RudderAngle < 0 {
			c.airplane.RudderAngle += dprec.Degrees(rudderRestoreSpeed * elapsedSeconds)
			c.airplane.RudderAngle = min(c.airplane.RudderAngle, 0)
		}
	}

	if c.pitchUp {
		c.targetPitch -= dprec.Degrees(pitchTurnSpeed * elapsedSeconds)
	}
	if c.pitchDown {
		c.targetPitch += dprec.Degrees(pitchTurnSpeed * elapsedSeconds)
	}
	c.targetPitch = dprec.Clamp(c.targetPitch, minPitch, maxPitch)

	if c.rollRight {
		c.targetRoll += dprec.Degrees(rollTurnSpeed * elapsedSeconds)
	}
	if c.rollLeft {
		c.targetRoll -= dprec.Degrees(rollTurnSpeed * elapsedSeconds)
	}
	c.targetRoll = dprec.Clamp(c.targetRoll, -maxRoll, maxRoll)

	directionRoll := dprec.Radians(math.Asin(dprec.Vec3Dot(
		dprec.UnitVec3(c.airplane.Body.Rotation().OrientationX()),
		dprec.BasisYVec3(),
	)))
	c.airplane.AileronAngle = c.targetRoll - directionRoll
	c.airplane.AileronAngle = dprec.Clamp(c.airplane.AileronAngle, -maxAileronAngle, maxAileronAngle)

	direction := c.airplane.Body.Velocity()
	if direction.Length() > 0.1 {
		lateralDirection := dprec.NewVec2(direction.X, direction.Z).Length()
		directionPitch := dprec.Radians(math.Atan2(direction.Y, lateralDirection))

		c.airplane.ElevatorAngle = c.targetPitch - directionPitch
		c.airplane.ElevatorAngle = dprec.Clamp(c.airplane.ElevatorAngle, -maxElevatorAngle, maxElevatorAngle)
	}

}
