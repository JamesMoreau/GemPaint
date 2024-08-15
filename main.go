package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

//TODO: put colors stuff into one array instead of 3 separate arrays

type AppState struct {
	theme *material.Theme

	BrushButton widget.Clickable
	BrushIcon   *widget.Icon

	EraserButton widget.Clickable
	EraserIcon   *widget.Icon

	SelectedTool SelectedTool

	colors             []color.NRGBA
	colorButtons       []widget.Clickable
	colorsLabels       []string
	selectedColorIndex int
	ColorIcon          *widget.Icon
}

type SelectedTool string

const (
	Brush  SelectedTool = "Brush"
	Eraser SelectedTool = "Eraser"
)

var blue = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var golangBlue = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
var lightGray = color.NRGBA{R: 200, G: 200, B: 200, A: 255}

var defaultMargin = unit.Dp(10)

func main() {
	state := AppState{
		theme:        material.NewTheme(),
		colorButtons: make([]widget.Clickable, 3),
		colors: []color.NRGBA{
			{R: 255, G: 0, B: 0, A: 255}, // Red
			{R: 0, G: 255, B: 0, A: 255}, // Green
			{R: 0, G: 0, B: 255, A: 255}, // Blue
		},
		colorsLabels: []string{"Red", "Green", "Blue"},
		SelectedTool: Brush,
	}

	state.BrushIcon  = loadIcon(icons.ImageBrush)
	state.EraserIcon = loadIcon(icons.ImageBrightness1)
	state.ColorIcon  = loadIcon(icons.ImageColorLens)

	go func() {
		window := new(app.Window)
		window.Option(app.Title("GemBoard"))

		// Run the program
		err := run(window, &state)
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

	// Handle color button clicks
	for i := range state.colorButtons {
		wasClicked := state.colorButtons[i].Clicked(gtx)
		if wasClicked {
			state.selectedColorIndex = i
			println("Selected color: ", state.colorsLabels[i])
		}
	}

	// Tool buttons
	children := []layout.FlexChild{
		layout.Rigid(toolButton(theme, &state.BrushButton, state.BrushIcon, "Brush", state.SelectedTool == Brush)),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(toolButton(theme, &state.EraserButton, state.EraserIcon, "Eraser", state.SelectedTool == Eraser)),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
	}

	// Color buttons
	for i, color := range state.colors {
		btn := &state.colorButtons[i]
		color_name := state.colorsLabels[i]
		children = append(children,
			layout.Rigid(colorButton(theme, btn, color, color_name, state.selectedColorIndex == i)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		)
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
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

func colorButton(th *material.Theme, btn *widget.Clickable, btn_color color.NRGBA, color_name string, selected bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		iconButton := material.IconButton(th, btn, nil, "color")
		iconButton.Background = btn_color
		iconButton.Color = btn_color

		if selected {
			iconButton.Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		}

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

func loadIcon(data []byte) *widget.Icon {
	icon, err := widget.NewIcon(data)
	if err != nil || icon == nil {
		log.Fatal(err)
	}
	
	return icon
}
