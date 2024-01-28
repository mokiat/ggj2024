package view

import (
	"fmt"
	"time"

	"github.com/mokiat/ggj2024/internal/game/data"
	"github.com/mokiat/ggj2024/internal/ui/controller"
	"github.com/mokiat/ggj2024/internal/ui/global"
	"github.com/mokiat/ggj2024/internal/ui/model"
	"github.com/mokiat/ggj2024/internal/ui/widget"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/debug/metric/metricui"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

var PlayScreen = co.Define(&playScreenComponent{})

type PlayScreenData struct {
	AppModel     *model.Application
	LoadingModel *model.Loading
	PlayModel    *model.Play
}

type playScreenComponent struct {
	co.BaseComponent

	appModel     *model.Application
	loadingModel *model.Loading
	playModel    *model.Play

	controller *controller.PlayController

	debugVisible bool
}

var _ ui.ElementKeyboardHandler = (*playScreenComponent)(nil)
var _ ui.ElementMouseHandler = (*playScreenComponent)(nil)

func (c *playScreenComponent) OnCreate() {
	context := co.TypedValue[global.Context](c.Scope())
	screenData := co.GetData[PlayScreenData](c.Properties())
	c.appModel = screenData.AppModel
	c.loadingModel = screenData.LoadingModel
	c.playModel = screenData.PlayModel

	// FIXME: This may actually panic if there is a third party
	// waiting / reading on this and it happens to match the Get call.
	promise := screenData.PlayModel.DataPromise()
	playData, err := promise.Wait()
	if err != nil {
		panic(fmt.Errorf("failed to get data: %w", err))
	}
	c.controller = controller.NewPlayController(co.Window(c.Scope()).Window, context.AudioAPI, context.Engine, playData)
	c.controller.Start(c.onVictory, c.onDefeat)
}

func (c *playScreenComponent) OnDelete() {
	defer c.controller.Stop()
}

func (c *playScreenComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	return c.controller.OnMouseEvent(element, event)
}

func (c *playScreenComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Code {
	case ui.KeyCodeEscape:
		c.onExit()
		return true
	case ui.KeyCodeTab:
		if event.Action == ui.KeyboardActionDown {
			c.debugVisible = !c.debugVisible
			c.Invalidate()
		}
		return true
	default:
		return c.controller.OnKeyboardEvent(event)
	}
}

func (c *playScreenComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithData(std.ElementData{
			Essence:   c,
			Focusable: opt.V(true),
			Focused:   opt.V(true),
			Layout:    layout.Anchor(),
		})

		if c.debugVisible {
			co.WithChild("flamegraph", co.New(metricui.FlameGraph, func() {
				co.WithData(metricui.FlameGraphData{
					UpdateInterval: time.Second,
				})
				co.WithLayoutData(layout.Data{
					Top:   opt.V(0),
					Left:  opt.V(0),
					Right: opt.V(0),
				})
			}))
		}

		co.WithChild("lower-border", co.New(std.Container, func() {
			co.WithLayoutData(layout.Data{
				Bottom: opt.V(0),
				Left:   opt.V(0),
				Right:  opt.V(0),
				Height: opt.V(15),
			})
			co.WithData(std.ContainerData{
				BackgroundColor: opt.V(ui.RGB(0xD9, 0xAD, 0x6C)),
			})
		}))

		co.WithChild("lower-left", co.New(std.Picture, func() {
			co.WithLayoutData(layout.Data{
				Width:  opt.V(304),
				Height: opt.V(99),
				Bottom: opt.V(0),
				Left:   opt.V(0),
			})
			co.WithData(std.PictureData{
				Image:      co.OpenImage(c.Scope(), "ui/images/lower-left.png"),
				ImageColor: opt.V(ui.White()),
				Mode:       std.ImageModeStretch,
			})
		}))

		co.WithChild("timer", co.New(widget.Timer, func() {
			co.WithLayoutData(layout.Data{
				Width:  opt.V(183),
				Height: opt.V(184),
				Bottom: opt.V(0),
				Right:  opt.V(0),
			})
			co.WithData(widget.TimerData{
				Provider: c.controller,
			})
		}))

		co.WithChild("counter", co.New(widget.CowsCounter, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(10),
				Left:   opt.V(10),
				Width:  opt.V(128),
				Height: opt.V(184),
			})
			co.WithData(widget.CowsCounterData{
				Provider: c.controller,
			})
		}))

		co.WithChild("reset", co.New(widget.ResetButton, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(10),
				Right:  opt.V(10),
				Width:  opt.V(115),
				Height: opt.V(58),
			})
			co.WithCallbackData(widget.ResetButtonCallbackData{
				OnClick: c.onReset,
			})
		}))
	})
}

func (c *playScreenComponent) onExit() {
	scope := c.Scope()
	if scope == nil {
		return // TODO: Figure out why this case is at all possible.
	}
	co.Window(scope).Close()
}

func (c *playScreenComponent) onVictory(gameTime time.Duration) {
	c.controller.Freeze()

	co.OpenOverlay(c.Scope(), co.New(VictoryScreen, func() {
		co.WithData(VictoryScreenData{
			AppModel:     c.appModel,
			LoadingModel: c.loadingModel,
			PlayModel:    c.playModel,
		})
	}))
}

func (c *playScreenComponent) onDefeat(remainingCows int) {
	c.controller.Freeze()

	co.OpenOverlay(c.Scope(), co.New(DefeatScreen, func() {
		co.WithData(DefeatScreenData{
			AppModel:     c.appModel,
			LoadingModel: c.loadingModel,
			PlayModel:    c.playModel,
		})
	}))
}

func (c *playScreenComponent) onReset() {
	c.controller.Freeze()

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
