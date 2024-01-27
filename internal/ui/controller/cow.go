package controller

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game"
)

func NewCowSpawner(modelDef *game.ModelDefinition) *CowSpawner {
	return &CowSpawner{
		modelDef: modelDef,
	}
}

type CowSpawner struct {
	modelDef *game.ModelDefinition
}

func (s *CowSpawner) SpawnCow(scene *game.Scene, location dprec.Vec3) {
	_ = scene.CreateModel(game.ModelInfo{
		Definition:        s.modelDef,
		Name:              "Airplane",
		Position:          location,
		Rotation:          dprec.IdentityQuat(),
		Scale:             dprec.NewVec3(1.0, 1.0, 1.0),
		IsDynamic:         true,
		PrepareAnimations: true,
	})
}
