package global

import (
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/game"
)

type Context struct {
	AudioAPI    audio.API
	Engine      *game.Engine
	ResourceSet *game.ResourceSet
}
