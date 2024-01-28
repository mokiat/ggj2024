package view

import (
	"github.com/mokiat/ggj2024/internal/game/data"
	"github.com/mokiat/ggj2024/internal/ui/global"
	"github.com/mokiat/ggj2024/internal/ui/model"
	"github.com/mokiat/ggj2024/internal/ui/widget"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

var DefeatScreen = co.Define(&defeatScreenComponent{})

type DefeatScreenData struct {
	AppModel     *model.Application
	LoadingModel *model.Loading
	PlayModel    *model.Play
}

var _ ui.ElementMouseHandler = (*defeatScreenComponent)(nil)
var _ ui.ElementKeyboardHandler = (*defeatScreenComponent)(nil)

type defeatScreenComponent struct {
	co.BaseComponent

	appModel     *model.Application
	loadingModel *model.Loading
	playModel    *model.Play
}

func (c *defeatScreenComponent) OnCreate() {
	data := co.GetData[DefeatScreenData](c.Properties())
	c.appModel = data.AppModel
	c.loadingModel = data.LoadingModel
	c.playModel = data.PlayModel
}

func (c *defeatScreenComponent) Render() co.Instance {
	return co.New(widget.Modal, func() {
		co.WithLayoutData(layout.Data{
			Width:            opt.V(520),
			Height:           opt.V(273),
			HorizontalCenter: opt.V(0),
			VerticalCenter:   opt.V(0),
		})
		co.WithChild("frame", co.New(std.Element, func() {
			co.WithData(std.ElementData{
				Essence:   c,
				Focusable: opt.V(true),
				Focused:   opt.V(true),
				Layout:    layout.Fill(),
			})

			co.WithChild("image", co.New(std.Picture, func() {
				co.WithData(std.PictureData{
					Image:      co.OpenImage(c.Scope(), "ui/images/defeat.png"),
					ImageColor: opt.V(ui.White()),
					Mode:       std.ImageModeStretch,
				})
			}))
		}))
	})
}

func (c *defeatScreenComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	if event.Action == ui.MouseActionDown && event.Button == ui.MouseButtonLeft {
		c.onContinue()
		return true
	}
	return false
}

func (c *defeatScreenComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	scope := c.Scope()
	if scope == nil {
		return false // TODO: Figure out why this case is at all possible.
	}
	if event.Action == ui.KeyboardActionDown {
		switch event.Code {
		case ui.KeyCodeEscape:
			co.Window(scope).Close()
			return true
		case ui.KeyCodeSpace, ui.KeyCodeEnter:
			c.onContinue()
			return true
		default:
			return false
		}
	}
	return false
}

func (c *defeatScreenComponent) onContinue() {
	co.CloseOverlay(c.Scope())

	context := co.TypedValue[global.Context](c.Scope())
	audioAPI := context.AudioAPI
	engine := context.Engine
	resourceSet := context.ResourceSet

	c.playModel.SetDataPromise(data.LoadPlayData(audioAPI, engine, resourceSet))

	promise := c.playModel.DataPromise()
	c.loadingModel.SetPromise(model.ToLoadingPromise(promise))
	c.loadingModel.SetNextViewName(model.ViewNamePlay)
	c.appModel.SetActiveView(model.ViewNameLoading)
}
