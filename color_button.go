package main

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type ColorButton struct {
	Color color.NRGBA
	Label string
	Size unit.Dp

	Clickable *widget.Clickable
	isSelected bool
	OnClick func()
}

func (cb *ColorButton) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if cb.Clickable.Clicked(gtx) {
		cb.isSelected = true
		if cb.OnClick != nil {
			cb.OnClick()
		}
	}

	btn := material.IconButton(th, cb.Clickable, nil, cb.Label)
	btn.Background = cb.Color

	borderColor := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	if cb.isSelected {
		borderColor = color.NRGBA{R: 255, G: 0, B: 0, A: 0}
	}

	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, 8).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, borderColor)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		},
		func(gtx layout.Context) layout.Dimensions {
			cb.Size = unit.Dp(10)
			return btn.Layout(gtx)
		},
	)
}