package main

import (
	"image/color"

	"gioui.org/layout"
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

	var borderColor color.NRGBA
	var borderWidth unit.Dp

	border := widget.Border{
		Color:        borderColor,
		Width:        borderWidth,
		CornerRadius: unit.Dp(4),
	}

	if cb.isSelected {
		borderColor = lightGray // Black or any contrasting color
		borderWidth = unit.Dp(2) // Thicker border for selected state
	} else {
		borderColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0} // Transparent for unselected state
		borderWidth = unit.Dp(1) // Thinner or no border for unselected state
	}

	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return btn.Layout(gtx)
	})
}