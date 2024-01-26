package view

import (
	"time"

	"github.com/mokiat/ggj2024/internal/game/data"
	"github.com/mokiat/ggj2024/internal/ui/global"
	"github.com/mokiat/ggj2024/internal/ui/model"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

type IntroScreenData struct {
	AppModel     *model.Application
	LoadingModel *model.Loading
	PlayModel    *model.Play
}

var IntroScreen = co.Define(&introScreenComponent{})

type introScreenComponent struct {
	co.BaseComponent
}

func (c *introScreenComponent) OnCreate() {
	co.Window(c.Scope()).SetCursorVisible(false)

	context := co.TypedValue[global.Context](c.Scope())
	audioAPI := context.AudioAPI
	engine := context.Engine
	resourceSet := context.ResourceSet

	introData := co.GetData[IntroScreenData](c.Properties())
	appModel := introData.AppModel
	loadingModel := introData.LoadingModel
	playModel := introData.PlayModel
	playModel.SetDataPromise(data.LoadPlayData(audioAPI, engine, resourceSet))

	co.After(c.Scope(), time.Second, func() {
		promise := playModel.DataPromise()
		if promise.Ready() {
			appModel.SetActiveView(model.ViewNamePlay)
		} else {
			loadingModel.SetPromise(model.ToLoadingPromise(promise))
			loadingModel.SetNextViewName(model.ViewNamePlay)
			appModel.SetActiveView(model.ViewNameLoading)
		}
	})
}

func (c *introScreenComponent) OnDelete() {
	co.Window(c.Scope()).SetCursorVisible(true)
}

func (c *introScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("logo-picture", co.New(std.Picture, func() {
			co.WithLayoutData(layout.Data{
				Width:            opt.V(512),
				Height:           opt.V(128),
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
			co.WithData(std.PictureData{
				BackgroundColor: opt.V(ui.Transparent()),
				Image:           co.OpenImage(c.Scope(), "ui/images/logo.png"),
				Mode:            std.ImageModeFit,
			})
		}))
	})
}
