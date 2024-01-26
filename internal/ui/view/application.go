package view

import (
	"github.com/mokiat/ggj2024/internal/ui/model"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mvc"
	"github.com/mokiat/lacking/ui/std"
)

var Application = mvc.EventListener(co.Define(&applicationComponent{}))

type ApplicationData struct {
	AppModel     *model.Application
	LoadingModel *model.Loading
	PlayModel    *model.Play
}

type applicationComponent struct {
	co.BaseComponent

	appModel     *model.Application
	loadingModel *model.Loading
	playModel    *model.Play
}

func (c *applicationComponent) OnUpsert() {
	appData := co.GetData[ApplicationData](c.Properties())
	c.appModel = appData.AppModel
	c.loadingModel = appData.LoadingModel
	c.playModel = appData.PlayModel
}

func (c *applicationComponent) Render() co.Instance {
	return co.New(std.Switch, func() {
		co.WithData(std.SwitchData{
			ChildKey: c.appModel.ActiveView(),
		})

		co.WithChild(model.ViewNameIntro, co.New(IntroScreen, func() {
			co.WithData(IntroScreenData{
				AppModel:     c.appModel,
				LoadingModel: c.loadingModel,
				PlayModel:    c.playModel,
			})
		}))

		co.WithChild(model.ViewNameLoading, co.New(LoadingScreen, func() {
			co.WithData(LoadingScreenData{
				AppModel:     c.appModel,
				LoadingModel: c.loadingModel,
			})
		}))

		co.WithChild(model.ViewNamePlay, co.New(PlayScreen, func() {
			co.WithData(PlayScreenData{
				AppModel:  c.appModel,
				PlayModel: c.playModel,
			})
		}))
	})
}

func (c *applicationComponent) OnEvent(event mvc.Event) {
	switch event.(type) {
	case *model.AppActiveViewSetEvent:
		c.Invalidate()
	}
}
