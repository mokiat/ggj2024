package controller

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/constraint"
	"github.com/mokiat/lacking/game/preset"
)

var (
	maxAileronAngle  = dprec.Degrees(30) // TODO
	maxElevatorAngle = dprec.Degrees(30) // TODO
	maxRudderAngle   = dprec.Degrees(30) // TODO
	maxThrust        = 9.8 * 1.5         // 1.5g
	thrustRampUp     = maxThrust / 2.0
)

func NewAirplane(physicsScene *physics.Scene, ecsScene *ecs.Scene, model *game.Model, position dprec.Vec3) *Airplane {
	collisionGroup := physics.NewCollisionGroup()

	airplaneBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   1500.0,
		MomentOfInertia:        physics.SymmetricMomentOfInertia(1500.0 / 2.0),
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
		Mass:                   25.0,
		MomentOfInertia:        physics.SolidSphereMomentOfInertia(50.0, 0.4),
		DragFactor:             0.0, // TODO
		AngularDragFactor:      1.0, // TODO
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
		Mass:                   50.0,
		MomentOfInertia:        physics.SolidSphereMomentOfInertia(100.0, 0.4),
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
		Mass:                   25.0,
		MomentOfInertia:        physics.SolidSphereMomentOfInertia(100.0, 0.4),
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
		Mass:                   300.0,
		MomentOfInertia:        physics.SolidSphereMomentOfInertia(10.0, 0.4),
		DragFactor:             0.0,
		AngularDragFactor:      0.0,
		RestitutionCoefficient: 0.0,
	})

	airplaneNode := model.Root().FindNode("Body")
	airplaneBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       airplaneNode.Name(),
		Definition: airplaneBodyDef,
		Position:   airplaneNode.AbsoluteMatrix().Translation(),
		Rotation:   dprec.IdentityQuat(),
	})
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
	leftAileronBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       leftAileronNode.Name(),
		Definition: aileronBodyDef,
		Position:   leftAileronNode.AbsoluteMatrix().Translation(),
		Rotation:   dprec.IdentityQuat(),
	})
	leftAileronNode.SetSource(game.BodyNodeSource{
		Body: leftAileronBody,
	})

	leftAileronRelativePosition := dprec.Vec3Diff(
		leftAileronNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
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
	rightAileronBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       rightAileronNode.Name(),
		Definition: aileronBodyDef,
		Position:   rightAileronNode.AbsoluteMatrix().Translation(),
		Rotation:   dprec.IdentityQuat(),
	})
	rightAileronNode.SetSource(game.BodyNodeSource{
		Body: rightAileronBody,
	})

	rightAileronRelativePosition := dprec.Vec3Diff(
		rightAileronNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
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
	elevatorBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       elevatorNode.Name(),
		Definition: elevatorBodyDef,
		Position:   elevatorNode.AbsoluteMatrix().Translation(),
		Rotation:   dprec.IdentityQuat(),
	})
	elevatorNode.SetSource(game.BodyNodeSource{
		Body: elevatorBody,
	})

	elevatorRelativePosition := dprec.Vec3Diff(
		elevatorNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)
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

	rudderRotation := constraint.NewMatchDirections().
		SetPrimaryDirection(dprec.BasisZVec3()).
		SetSecondaryDirection(dprec.BasisZVec3())

	rudderNode := model.FindNode("Rudder")
	leftRudderBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       "Rudder",
		Definition: rudderBodyDef,
		Position:   rudderNode.AbsoluteMatrix().Translation(),
		Rotation:   dprec.IdentityQuat(),
	})
	rudderNode.SetSource(game.BodyNodeSource{
		Body: leftRudderBody,
	})

	rudderRelativePosition := dprec.Vec3Diff(
		rudderNode.AbsoluteMatrix().Translation(),
		airplaneNode.AbsoluteMatrix().Translation(),
	)

	physicsScene.CreateDoubleBodyConstraint(airplaneBody, leftRudderBody, constraint.NewPairCombined(
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

	entity := ecsScene.CreateEntity()
	entity.SetComponent(preset.NodeComponentID, &preset.NodeComponent{
		Node: airplaneNode,
	})

	return &Airplane{
		Entity: entity,
		Node:   airplaneNode,

		Body:                      airplaneBody,
		LeftAileronRotConstraint:  leftAileronRotation,
		RightAileronRotConstraint: rightAileronRotation,
		ElevatorRotConstraint:     elevatorRotation,
		RudderRotConstraint:       rudderRotation,
	}
}

type Airplane struct {
	Entity *ecs.Entity
	Node   *hierarchy.Node

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

type AirplaneMouseController struct{} // TODO
