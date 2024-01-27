package view

import (
	"time"

	"github.com/mokiat/ggj2024/internal/ui/model"
	"github.com/mokiat/ggj2024/internal/ui/widget"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

type LoadingScreenData struct {
	AppModel     *model.Application
	LoadingModel *model.Loading
}

var LoadingScreen = co.Define(&loadingScreenComponent{})

type loadingScreenComponent struct {
	co.BaseComponent
}

func (c *loadingScreenComponent) OnCreate() {
	screenData := co.GetData[LoadingScreenData](c.Properties())
	appModel := screenData.AppModel
	loadingModel := screenData.LoadingModel
	co.After(c.Scope(), 2*time.Second, func() {
		loadingModel.Promise().OnReady(func() {
			co.Schedule(c.Scope(), func() {
				appModel.SetActiveView(loadingModel.NextViewName())
			})
		})
	})
}

func (c *loadingScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("loading", co.New(widget.Loading, func() {
			co.WithLayoutData(layout.Data{
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
		}))

		co.WithChild("loading-text", co.New(std.Label, func() {
			co.WithLayoutData(layout.Data{
				Right:  opt.V(10),
				Bottom: opt.V(10),
			})
			co.WithData(std.LabelData{
				Font:      co.OpenFont(c.Scope(), "ui:///roboto-regular.ttf"),
				Text:      "Loading...",
				FontSize:  opt.V(float32(32)),
				FontColor: opt.V(ui.White()),
			})
		}))
	})
}
