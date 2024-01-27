package controller

import (
	"math/rand"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/constraint"
)

func NewCowSpawner(scene *game.Scene, modelDef *game.ModelDefinition) *CowSpawner {
	bodyDef := scene.Physics().Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   100.0,
		MomentOfInertia:        physics.SymmetricMomentOfInertia(100.0 / 2.0),
		FrictionCoefficient:    1.0,
		RestitutionCoefficient: 1.0,
		DragFactor:             0.0,
		AngularDragFactor:      0.0,
		CollisionSpheres: []collision.Sphere{
			collision.NewSphere(dprec.ZeroVec3(), 7.5),
		},
	})

	return &CowSpawner{
		scene:    scene,
		bodyDef:  bodyDef,
		modelDef: modelDef,
	}
}

type CowSpawner struct {
	scene *game.Scene

	bodyDef  *physics.BodyDefinition
	modelDef *game.ModelDefinition
}

func (s *CowSpawner) SpawnCow(location dprec.Vec3) {
	body := s.scene.Physics().CreateBody(physics.BodyInfo{
		Name:       "Cow",
		Definition: s.bodyDef,
		Position:   location,
		Rotation:   dprec.IdentityQuat(),
	})
	body.SetRotation(dprec.RotationQuat(dprec.Degrees(rand.Float64()*90), dprec.BasisYVec3()))
	s.scene.Physics().CreateSingleBodyConstraint(body,
		constraint.NewStaticPosition().SetPosition(location),
	)
	model := s.scene.CreateModel(game.ModelInfo{
		Definition:        s.modelDef,
		Name:              "Cow",
		Position:          location,
		Rotation:          dprec.IdentityQuat(),
		Scale:             dprec.NewVec3(1.0, 1.0, 1.0),
		IsDynamic:         true,
		PrepareAnimations: true,
	})
	model.Root().SetSource(game.BodyNodeSource{
		Body: body,
	})
}
