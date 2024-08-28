package main

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type ToolButtonStyle struct {
	Icon            *widget.Icon
	SelectedColor   color.NRGBA
	UnselectedColor color.NRGBA
	Label           string
	Clickable       *widget.Clickable
	Theme           *material.Theme
	IsSelected      bool
}

func ToolButton(theme *material.Theme, clickable *widget.Clickable, icon *widget.Icon, isSelected bool, selectedColor, unselectedColor color.NRGBA, label string) ToolButtonStyle {
	return ToolButtonStyle{
		Theme:           theme,
		Clickable:       clickable,
		Icon:            icon,
		SelectedColor:   selectedColor,
		UnselectedColor: unselectedColor,
		Label:           label,
		IsSelected:      isSelected,
	}
}

func (t ToolButtonStyle) Layout(gtx layout.Context) layout.Dimensions {
	iconButton := material.IconButton(t.Theme, t.Clickable, t.Icon, t.Label)
	iconButton.Background = t.UnselectedColor

	if t.IsSelected {
		iconButton.Background = t.SelectedColor
	}

	return iconButton.Layout(gtx)
}
