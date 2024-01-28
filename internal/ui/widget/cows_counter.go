package widget

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

type CowProvider interface {
	CowsRemaining() int
}

var CowsCounter = co.Define(&cowsCounterComponent{})

type CowsCounterData struct {
	Provider CowProvider
}

type cowsCounterComponent struct {
	co.BaseComponent

	provider CowProvider

	bgImage *ui.Image
	font    *ui.Font
}

func (c *cowsCounterComponent) OnCreate() {
	data := co.GetData[CowsCounterData](c.Properties())
	c.provider = data.Provider

	c.bgImage = co.OpenImage(c.Scope(), "ui/images/cows-counter.png")
	c.font = co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf")
}

func (c *cowsCounterComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(std.ElementData{
			Essence:   c,
			Layout:    layout.Anchor(),
			IdealSize: opt.V(ui.NewSize(128, 184)),
		})
	})
}

func (c *cowsCounterComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)
	canvas.Reset()
	canvas.Rectangle(
		drawBounds.Position,
		drawBounds.Size,
	)
	canvas.Fill(ui.Fill{
		Color:       ui.White(),
		Image:       c.bgImage,
		ImageOffset: drawBounds.Position,
		ImageSize:   drawBounds.Size,
	})

	text := []rune(fmt.Sprintf("%d", c.provider.CowsRemaining()))
	fontSize := float32(24.0)

	canvas.Reset()
	canvas.FillTextLine([]rune(text), sprec.Vec2{
		X: (128 - c.font.LineWidth(text, fontSize)) / 2,
		Y: 150.0,
	}, ui.Typography{
		Font:  c.font,
		Size:  fontSize,
		Color: ui.RGB(0x8B, 0x63, 0x28),
	})

	element.Invalidate()
}
