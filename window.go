package main

import (
	"image/color"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type OverlayWindow struct {
	Window *app.Window
	Theme  *material.Theme
	Text   string
}

func NewOverlayWindow() *OverlayWindow {
	w := &app.Window{}
	w.Option(
		app.Title("Prototypage"),
		app.Size(unit.Dp(400), unit.Dp(100)),
		app.Decorated(true), // false si tu veux borderless
	)

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(nil))

	return &OverlayWindow{
		Window: w,
		Theme:  th,
		Text:   "Waiting...",
	}
}

func (o *OverlayWindow) Run() error {
	var ops op.Ops

	for {
		switch e := o.Window.Event().(type) {
		case app.DestroyEvent:
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Background noir
			paint.Fill(gtx.Ops, color.NRGBA{R: 30, G: 30, B: 30, A: 255})

			// Texte centré
			layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Label(o.Theme, unit.Sp(14), o.Text)
				label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				return label.Layout(gtx)
			})

			e.Frame(gtx.Ops)
		}
	}
}

// SetText met à jour le texte affiché (thread-safe via invalidate)
func (o *OverlayWindow) SetText(s string) {
	o.Text = s
	o.Window.Invalidate() // force un redraw
}