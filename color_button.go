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

type ColorButtonStyle struct {
	Color color.NRGBA
	Label string

	Clickable  *widget.Clickable
	isSelected bool
	OnClick    func()
}

func (cb *ColorButtonStyle) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	btn := material.IconButton(th, cb.Clickable, nil, cb.Label)
	btn.Background = cb.Color
	// btn.Size = unit.Dp(14)
	
	if !cb.isSelected {
		return btn.Layout(gtx)
	}

	border_width := unit.Dp(2)
	btn.Size = btn.Size - border_width

	return layout.Background{}.Layout(gtx, 
		func(gtx layout.Context) layout.Dimensions {
			rr := (gtx.Constraints.Min.X + gtx.Constraints.Min.Y) / 4
			defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, rr).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, lightGray)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(border_width).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return btn.Layout(gtx)
		})
	})
}
