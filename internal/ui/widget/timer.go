package widget

import (
	"fmt"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

type TimeProvider interface {
	RemainingTime() time.Duration
}

var Timer = co.Define(&timerComponent{})

type TimerData struct {
	Provider TimeProvider
}

type timerComponent struct {
	co.BaseComponent

	provider TimeProvider

	bgImage *ui.Image
	font    *ui.Font
}

func (c *timerComponent) OnCreate() {
	data := co.GetData[TimerData](c.Properties())
	c.provider = data.Provider

	c.bgImage = co.OpenImage(c.Scope(), "ui/images/lower-right.png")
	c.font = co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf")
}

func (c *timerComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(std.ElementData{
			Essence:   c,
			Layout:    layout.Anchor(),
			IdealSize: opt.V(ui.NewSize(128, 184)),
		})
	})
}

func (c *timerComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
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

	remainingTime := c.provider.RemainingTime().Truncate(time.Second)
	minutes := int(remainingTime.Seconds()) / 60
	seconds := int(remainingTime.Seconds()) % 60

	text := []rune(fmt.Sprintf("%02d:%02d", minutes, seconds))
	fontSize := float32(32.0)

	canvas.Reset()
	canvas.FillTextLine([]rune(text), sprec.Vec2{
		X: (178 - c.font.LineWidth(text, fontSize)) / 2,
		Y: 78.0,
	}, ui.Typography{
		Font:  c.font,
		Size:  fontSize,
		Color: ui.RGB(0x8B, 0x63, 0x28),
	})

	element.Invalidate()
}
