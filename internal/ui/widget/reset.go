package widget

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

var ResetButton = co.Define(&resetButtonComponent{})

type ResetButtonCallbackData struct {
	OnClick std.OnActionFunc
}

type resetButtonComponent struct {
	co.BaseComponent
	std.BaseButtonComponent

	bgImage *ui.Image
}

func (c *resetButtonComponent) OnCreate() {
	callbackData := co.GetCallbackData[ResetButtonCallbackData](c.Properties())
	c.SetOnClickFunc(callbackData.OnClick)

	c.bgImage = co.OpenImage(c.Scope(), "ui/images/reset.png")
}

func (c *resetButtonComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(std.ElementData{
			Essence:   c,
			Layout:    layout.Anchor(),
			IdealSize: opt.V(ui.NewSize(115, 58)),
		})
	})
}

func (c *resetButtonComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
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
	element.Invalidate()
}
