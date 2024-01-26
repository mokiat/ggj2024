package model

import (
	"errors"

	"github.com/mokiat/ggj2024/internal/game/data"
	"github.com/mokiat/lacking/ui/mvc"
	"github.com/mokiat/lacking/util/async"
)

func NewPlay(eventBus *mvc.EventBus) *Play {
	return &Play{
		eventBus: eventBus,
		promise:  async.NewFailedPromise[*data.PlayData](errors.New("not scheduled")),
	}
}

type Play struct {
	eventBus *mvc.EventBus
	promise  async.Promise[*data.PlayData]
}

func (h *Play) DataPromise() async.Promise[*data.PlayData] {
	return h.promise
}

func (h *Play) SetDataPromise(promise async.Promise[*data.PlayData]) {
	h.promise = promise
}
