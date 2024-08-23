package main

import (
	
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var BrushIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ImageBrush)
	return icon
}()

var EraserIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ImageBrightness1)
	return icon
}()

var ClearIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionDelete)
	return icon
}()

