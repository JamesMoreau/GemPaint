package main

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/widget"

	"golang.org/x/exp/shiny/materialdesign/icons"
)

var golangBlue = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
var lightGray = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
var softBlue = color.NRGBA{R: 230, G: 240, B: 250, A: 255}
var red = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
var orange = color.NRGBA{R: 255, G: 165, B: 0, A: 255}
var green = color.NRGBA{R: 0, G: 128, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var yellow = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var purple = color.NRGBA{R: 128, G: 0, B: 128, A: 255}
var darkGray = color.NRGBA{R: 30, G: 30, B: 30, A: 255}

var defaultCanvasDimensions = image.Rect(0, 0, 1920, 1080)
var defaultCanvasBackground = lightGray
var defaultCanvasColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

var mouseIsOutsideCanvas = f32.Point{X: -1, Y: -1}

var defaultCursorRadius = 20
var minimumCursorRadius = 10
var maximumCursorRadius = 100
var cursorRadiusChangeStep = 10

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

var AddIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentAdd)
	return icon
}()

var MinusIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentRemove)
	return icon
}()

var SaveIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentSave)
	return icon
}()

