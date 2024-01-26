package internal

import (
	"github.com/mokiat/ggj2024/internal/ui/global"
	"github.com/mokiat/ggj2024/internal/ui/model"
	"github.com/mokiat/ggj2024/internal/ui/view"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mvc"
)

func BootstrapApplication(window *ui.Window, gameController *game.Controller) {
	engine := gameController.Engine()
	eventBus := mvc.NewEventBus()

	scope := co.RootScope(window)
	scope = co.TypedValueScope(scope, eventBus)
	scope = co.TypedValueScope(scope, global.Context{
		AudioAPI:    window.AudioAPI(),
		Engine:      engine,
		ResourceSet: engine.CreateResourceSet(),
	})
	co.Initialize(scope, co.New(Bootstrap, nil))
}

var Bootstrap = co.Define(&bootstrapComponent{})

type bootstrapComponent struct {
	co.BaseComponent

	appModel     *model.Application
	loadingModel *model.Loading
}

func (c *bootstrapComponent) OnCreate() {
	eventBus := co.TypedValue[*mvc.EventBus](c.Scope())
	c.appModel = model.NewApplication(eventBus)
	c.loadingModel = model.NewLoading(eventBus)
}

func (c *bootstrapComponent) Render() co.Instance {
	return co.New(view.Application, func() {
		co.WithData(view.ApplicationData{
			AppModel:     c.appModel,
			LoadingModel: c.loadingModel,
		})
	})
}
