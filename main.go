package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type ApplicationState struct {
	theme *material.Theme

	BrushButton  widget.Clickable
	EraserButton widget.Clickable
	SelectedTool SelectedTool

	colorButtons       []ColorButtonStyle
	selectedColorIndex int
}

type SelectedTool string

const (
	Brush  SelectedTool = "Brush"
	Eraser SelectedTool = "Eraser"
)

var defaultMargin = unit.Dp(10)

func main() {
	state := ApplicationState{
		theme: material.NewTheme(),
		colorButtons: []ColorButtonStyle{
			{Color: color.NRGBA{R: 255, G: 0,   B: 0,   A: 255}, Label: "Red",    Clickable: &widget.Clickable{}, isSelected: true},
			{Color: color.NRGBA{R: 255, G: 165, B: 0,   A: 255}, Label: "Orange", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 0,   G: 255, B: 0,   A: 255}, Label: "Green",  Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 0,   G: 0,   B: 255, A: 255}, Label: "Blue",   Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 255, G: 255, B: 0,   A: 255}, Label: "Yellow", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 128, G: 0,   B: 128, A: 255}, Label: "Purple", Clickable: &widget.Clickable{}},
		},
		selectedColorIndex: 0,
		SelectedTool:       Brush,
	}

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

func run(window *app.Window, state *ApplicationState) error {
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

func layoutSidebar(gtx layout.Context, state *ApplicationState, theme *material.Theme) layout.Dimensions {
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
		btn := &state.colorButtons[i]
		wasClicked := btn.Clickable.Clicked(gtx)
		if wasClicked {
			for j := range state.colorButtons {
				state.colorButtons[j].isSelected = false
			}

			state.selectedColorIndex = i
			state.colorButtons[i].isSelected = true

			println("Selected color: ", btn.Label)
		}
	}

	// Tool buttons
	children := []layout.FlexChild{
		layout.Rigid(toolButton(theme, &state.BrushButton, BrushIcon, "Brush", state.SelectedTool == Brush)),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(toolButton(theme, &state.EraserButton, EraserIcon, "Eraser", state.SelectedTool == Eraser)),
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
	}

	// Color buttons
	for i := range state.colorButtons {
		btn := &state.colorButtons[i]
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return btn.Layout(gtx, theme)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		)
	}

	return layout.Background{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, softBlue)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(defaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
		})
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

func layoutCanvas(gtx layout.Context, state *ApplicationState) layout.Dimensions {
	// r := image.Rectangle{Max: image.Point{X: 800, Y: 600}}
	// area := clip.Rect(r).Push(ops)
	// event.Op{Tag: h}.Add(ops)
	// area.Pop()

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
	)
}
