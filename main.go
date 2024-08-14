package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	_ "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type AppState struct {
	BrushButton widget.Clickable
	BrushIcon   *widget.Icon

	EraserButton widget.Clickable
	EraserIcon   *widget.Icon

	ColorButton  widget.Clickable
	ColorIcon    *widget.Icon

	RedButton    widget.Clickable
	GreenButton  widget.Clickable
	BlueButton   widget.Clickable

	SelectedTool SelectedTool
	SelectedColor SelectedColor
}

type SelectedTool string

const (
	Brush       SelectedTool = "Brush"
	Eraser      SelectedTool = "Eraser"
)

type SelectedColor string

const (
	Red SelectedColor = "Red"
	Green SelectedColor = "Green"
	Blue SelectedColor = "Blue"
)

var blue = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var golangBlue = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
var lightGray = color.NRGBA{R: 200, G: 200, B: 200, A: 255}

var defaultMargin = unit.Dp(10)

func main() {
	go func() {
		var err error
		window := new(app.Window)
		window.Option(app.Title("GemPaint"))

		state := AppState{}

		// Load icons
		state.BrushIcon, err = widget.NewIcon(icons.ImageBrush)
		if err != nil {
			log.Fatal(err)
		}

		state.EraserIcon, err = widget.NewIcon(icons.ImageBrightness1)
		if err != nil {
			log.Fatal(err)
		}

		state.ColorIcon, err = widget.NewIcon(icons.ImageColorLens)
		if err != nil {
			log.Fatal(err)
		}

		state.SelectedTool = Brush

		// Run the program
		err = run(window, &state)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func run(window *app.Window, state *AppState) error {
	theme := material.NewTheme()
	var ops op.Ops

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			myLayout := layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}
			myLayout.Layout(gtx,
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						return layoutSidebar(gtx, state, theme)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						return layoutCanvas(gtx, state)
					},
				),
			)

			e.Frame(gtx.Ops)
		}
	}
}

func layoutSidebar(gtx layout.Context, state *AppState, theme *material.Theme) layout.Dimensions {
	inset := layout.UniformInset(defaultMargin)

	if state.BrushButton.Clicked(gtx) {
		state.SelectedTool = Brush
		println("Current tool: ", state.SelectedTool)
	}

	if state.EraserButton.Clicked(gtx) {
		state.SelectedTool = Eraser
		println("Current tool: ", state.SelectedTool)
	}

	if state.ColorButton.Clicked(gtx) {
		// state.SelectedTool = ColorPicker
		println("Current tool: ", state.SelectedTool)
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		rd_btn := material.Button(theme, &state.RedButton, "")
		rd_btn.Background = color.NRGBA{R: 255, G: 0, B: 0, A: 255}

		blue_btn := material.Button(theme, &state.BlueButton, "")
		blue_btn.Background = color.NRGBA{R: 0, G: 0, B: 255, A: 255}

		grn_btn := material.Button(theme, &state.GreenButton, "")
		grn_btn.Background = color.NRGBA{R: 0, G: 255, B: 0, A: 255}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(toolButton(theme, &state.BrushButton, state.BrushIcon, "Brush", state.SelectedTool == Brush)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(toolButton(theme, &state.EraserButton, state.EraserIcon, "Eraser", state.SelectedTool == Eraser)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			
			layout.Rigid(rd_btn.Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(blue_btn.Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(grn_btn.Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(colorButton(theme, &state.RedButton, color.NRGBA{R: 255, G: 0, B: 0, A: 255}, "Color", state.SelectedColor == Red)),
		)
	})
}

func toolButton(th *material.Theme, btn *widget.Clickable, icon *widget.Icon, label string, selected bool) layout.Widget {
    return func(gtx layout.Context) layout.Dimensions {
        iconButton := material.IconButton(th, btn, icon, label)
		iconButton.Background = lightGray
		
        if selected {
			iconButton.Background = golangBlue
        }

        return iconButton.Layout(gtx)
    }
}

func colorButton(th *material.Theme, btn *widget.Clickable, color color.NRGBA, label string, selected bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		iconButton := material.Button(th, btn, label)
		iconButton.Background = color
		iconButton.Color = color
		
		return iconButton.Layout(gtx)
	}
}

func layoutCanvas(gtx layout.Context, state *AppState) layout.Dimensions {
	// r := image.Rectangle{Max: image.Point{X: 800, Y: 600}}
	// area := clip.Rect(r).Push(ops)
	// event.Op{Tag: h}.Add(ops)
	// area.Pop()

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return widget.Border{Color: golangBlue, Width: unit.Dp(4)}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
		}),
	)
}
