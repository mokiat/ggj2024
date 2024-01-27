package widget

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/std"
)

var Loading = co.Define(&loadingComponent{})

type loadingComponent struct {
	co.BaseComponent

	backImage  *ui.Image
	frontImage *ui.Image
	angle      sprec.Angle
}

func (c *loadingComponent) OnCreate() {
	c.backImage = co.OpenImage(c.Scope(), "ui/images/loading-back.png")
	c.frontImage = co.OpenImage(c.Scope(), "ui/images/loading-front.png")
}

func (c *loadingComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(std.ElementData{
			Essence:   c,
			IdealSize: opt.V(ui.NewSize(256, 256)),
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *loadingComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	elapsedTime := canvas.ElapsedTime()
	c.angle += sprec.Degrees(float32(elapsedTime.Seconds()) * 360.0 * 2.0)

	drawBounds := canvas.DrawBounds(element, false)

	canvas.Translate(drawBounds.Position)

	canvas.Reset()
	canvas.Rectangle(
		sprec.ZeroVec2(),
		drawBounds.Size,
	)
	canvas.Fill(ui.Fill{
		Rule:        ui.FillRuleSimple,
		Color:       ui.White(),
		Image:       c.backImage,
		ImageOffset: sprec.ZeroVec2(),
		ImageSize:   drawBounds.Size,
	})

	canvas.Push()
	canvas.Translate(sprec.NewVec2(128.0, 128.0))
	canvas.Rotate(c.angle)
	canvas.Reset()
	canvas.Rectangle(
		sprec.NewVec2(-128.0, -128.0),
		sprec.NewVec2(256.0, 256.0),
	)
	canvas.Fill(ui.Fill{
		Rule:        ui.FillRuleSimple,
		Color:       ui.White(),
		Image:       c.frontImage,
		ImageOffset: sprec.NewVec2(-128.0, -128.0),
		ImageSize:   sprec.NewVec2(256.0, 256.0),
	})
	canvas.Pop()

	element.Invalidate() // force redraw
}
