package main

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type myTheme struct {}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		return color.White
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func main() {
	myApp := app.New()

	// Set the theme to the custom theme
	myApp.Settings().SetTheme(&myTheme{})

	myWindow := myApp.NewWindow("GemPaint")

	brushButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		log.Println("Brush tool selected")
	})
	eraserButton := widget.NewButtonWithIcon("Erase", theme.MediaRecordIcon(), func() {
		log.Println("Eraser tool selected")
	})
	colorButton := widget.NewButtonWithIcon("Color", nil, nil)
	tools := container.NewVBox(brushButton, eraserButton, colorButton)

	rect := canvas.NewRectangle(color.White)

	content := container.NewBorder(nil, nil, tools, nil, rect)

	myWindow.SetContent(content)

	myWindow.ShowAndRun()
}
