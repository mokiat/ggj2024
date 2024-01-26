package view

import (
	"github.com/mokiat/ggj2024/internal/ui/model"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mvc"
	"github.com/mokiat/lacking/ui/std"
)

var Application = mvc.EventListener(co.Define(&applicationComponent{}))

type ApplicationData struct {
	AppModel *model.Application
}

type applicationComponent struct {
	co.BaseComponent

	appModel *model.Application
}

func (c *applicationComponent) OnUpsert() {
	appData := co.GetData[ApplicationData](c.Properties())
	c.appModel = appData.AppModel
}

func (c *applicationComponent) Render() co.Instance {
	return co.New(std.Switch, func() {
		co.WithData(std.SwitchData{
			ChildKey: c.appModel.ActiveView(),
		})

		co.WithChild(model.ViewNameIntro, co.New(IntroScreen, func() {
			co.WithData(IntroScreenData{
				AppModel: c.appModel,
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
