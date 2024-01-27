package controller

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/constraint"
)

func NewBall(physicsScene *physics.Scene, airplane *Airplane, model *game.Model) *Ball {
	hingeUpperNode := model.FindNode("UpperNode")
	hingeLowerNode := model.FindNode("LowerNode")
	ballNode := model.FindNode("BallNode")

	hingeBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   1.0,
		MomentOfInertia:        physics.SymmetricMomentOfInertia(1.0),
		FrictionCoefficient:    0.0,
		RestitutionCoefficient: 0.0,
		CollisionGroup:         airplane.CollisionGroup,
		DragFactor:             0.0,                          // TODO
		AngularDragFactor:      0.0,                          // TODO
		AerodynamicShapes:      []physics.AerodynamicShape{}, // TODO
	})

	ballBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   10.0,
		MomentOfInertia:        physics.SymmetricMomentOfInertia(10.0 / 2.0),
		FrictionCoefficient:    0.0,
		RestitutionCoefficient: 0.0,
		CollisionGroup:         airplane.CollisionGroup,
		DragFactor:             1.0,
		AngularDragFactor:      0.0,
		AerodynamicShapes: []physics.AerodynamicShape{
			physics.NewAerodynamicShape(
				physics.NewTransform(
					dprec.NewVec3(0.0, 0.0, 0.0),
					dprec.RotationQuat(dprec.Degrees(20), dprec.BasisXVec3()),
				),
				physics.NewSurfaceAerodynamicShape(1.0, 0.1, 0.5),
			),
		},
		CollisionSpheres: []collision.Sphere{
			collision.NewSphere(dprec.ZeroVec3(), 2.75),
		},
	})

	hingeBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       "Hinge",
		Definition: hingeBodyDef,
		Position:   airplane.Body.Position(),
		Rotation:   airplane.Body.Rotation(),
	})

	ballRelativePosition := dprec.Vec3Diff(
		hingeLowerNode.AbsoluteMatrix().Translation(),
		hingeUpperNode.AbsoluteMatrix().Translation(),
	)
	ballBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       "Ball",
		Definition: ballBodyDef,
		Position:   dprec.Vec3Sum(airplane.Body.Position(), ballRelativePosition),
		Rotation:   airplane.Body.Rotation(),
	})

	physicsScene.CreateDoubleBodyConstraint(hingeBody, airplane.Body, constraint.NewPairCombined(
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(dprec.ZeroVec3()).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisXVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(dprec.ZeroVec3()).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisZVec3()).
			SetOffset(0.0),
		constraint.NewMatchDirectionOffset().
			SetPrimaryRadius(dprec.ZeroVec3()).
			SetSecondaryRadius(dprec.ZeroVec3()).
			SetDirection(dprec.BasisYVec3()).
			SetOffset(0.0),
		constraint.NewCopyRotation(),
	))

	physicsScene.CreateDoubleBodyConstraint(ballBody, airplane.Body, constraint.NewPairCombined(
		constraint.NewCopyRotation(),
	))
	physicsScene.CreateDoubleBodyConstraint(airplane.Body, ballBody, constraint.NewPairCombined(
		constraint.NewClampDirectionOffset().
			SetDirection(dprec.BasisYVec3()).
			SetMax(-ballRelativePosition.Length()+1.2).
			SetMin(-ballRelativePosition.Length()),
	))

	physicsScene.CreateDoubleBodyConstraint(hingeBody, ballBody, constraint.NewPairCombined(
		constraint.NewHingedRod().SetLength(ballRelativePosition.Length()),
	))

	hingeUpperNode.SetSource(game.BodyNodeSource{
		Body: hingeBody,
	})
	hingeLowerNode.SetSource(game.BodyNodeSource{
		Body: ballBody,
	})
	ballNode.SetSource(game.BodyNodeSource{
		Body: ballBody,
	})

	return &Ball{
		Body: ballBody,
		Node: ballNode,
	}
}

type Ball struct {
	Body physics.Body
	Node *hierarchy.Node
}
