package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var debug = false

type ApplicationState struct {
	theme *material.Theme

	brushButton  widget.Clickable
	eraserButton widget.Clickable
	clearButton  widget.Clickable
	selectedTool SelectedTool

	cursorRadius int

	colorButtons       []ColorButtonStyle
	selectedColorIndex int

	canvas                *image.RGBA
	mousePositionOnCanvas f32.Point
}

type SelectedTool string

const (
	Brush  SelectedTool = "Brush"
	Eraser SelectedTool = "Eraser"
)

func main() {
	// Get arguments
	args := os.Args
	for _, arg := range args {
		if arg == "-debug" {
			debug = true
			fmt.Println("Debug mode enabled")
			continue
		}
	}

	// Initialize the application state
	state := new(ApplicationState) // store the state on the heap
	*state = ApplicationState{
		theme:        material.NewTheme(),
		selectedTool: Brush,
		cursorRadius: defaultCursorRadius,
		colorButtons: []ColorButtonStyle{
			{Color: red, Label: "Red", Clickable: &widget.Clickable{}, isSelected: true},
			{Color: orange, Label: "Orange", Clickable: &widget.Clickable{}},
			{Color: green, Label: "Green", Clickable: &widget.Clickable{}},
			{Color: blue, Label: "Blue", Clickable: &widget.Clickable{}},
			{Color: yellow, Label: "Yellow", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 128, G: 0, B: 128, A: 255}, Label: "Purple", Clickable: &widget.Clickable{}},
		},
		selectedColorIndex:    0,
		canvas:                image.NewRGBA(defaultCanvasDimensions),
		mousePositionOnCanvas: mouseIsOutsideCanvas,
	}

	fillImageWithColor(state.canvas, golangBlue)

	go func() {
		window := new(app.Window)
		window.Option(app.Title("GemBoard"))
		window.Option(app.Size(unit.Dp(1000), unit.Dp(800)))

		// Run the program
		err := run(window, state)
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
			gtx := app.NewContext(&ops, e)

			layout.Stack{Alignment: layout.NE}.Layout(gtx,
				layout.Expanded(
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
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
					},
				),
				layout.Stacked(
					func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(32).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Body1(theme, fmt.Sprintf("🐭: %.2f, %.2f", state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y)).Layout(gtx)
						})
					},
				),
			)

			e.Frame(gtx.Ops)
		}
	}
}

func layoutSidebar(gtx layout.Context, state *ApplicationState, theme *material.Theme) layout.Dimensions {

	// Handle tool button clicks
	if state.brushButton.Clicked(gtx) {
		state.selectedTool = Brush
		if debug {
			fmt.Println("Current tool: ", state.selectedTool)
		}
	}

	if state.eraserButton.Clicked(gtx) {
		state.selectedTool = Eraser
		if debug {
			fmt.Println("Current tool: ", state.selectedTool)
		}
	}

	if state.clearButton.Clicked(gtx) {
		state.canvas = image.NewRGBA(defaultCanvasDimensions)
		fillImageWithColor(state.canvas, defaultCanvasBackground)
		if debug {
			fmt.Println("Canvas cleared")
		}
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

			if debug {
				fmt.Println("Selected color: ", btn.Label)
			}
		}
	}

	// Tool buttons
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return toolButton(gtx, theme, &state.brushButton, BrushIcon, "Brush", state.selectedTool == Brush)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return toolButton(gtx, theme, &state.eraserButton, EraserIcon, "Eraser", state.selectedTool == Eraser)
		}),
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

	// Other buttons
	children = append(children,
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return toolButton(gtx, theme, &state.clearButton, ClearIcon, "Clear", false)
		}),
	)

	return layout.Background{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, softBlue)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		})
	})
}

func toolButton(gtx layout.Context, th *material.Theme, btn *widget.Clickable, icon *widget.Icon, label string, selected bool) layout.Dimensions {
	iconButton := material.IconButton(th, btn, icon, label)
	iconButton.Background = lightGray

	if selected {
		iconButton.Background = golangBlue
	}

	return iconButton.Layout(gtx)
}

var tag = new(bool) // tag is a unique identifier for the canvas

func layoutCanvas(gtx layout.Context, state *ApplicationState) layout.Dimensions {

	// Render the canvas
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, defaultCanvasBackground)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		},
		func(gtx layout.Context) layout.Dimensions {
			return widget.Border{Color: color.NRGBA{R: 0, G: 0, B: 0, A: 0}, Width: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions { // TODO: remove this border

				return layout.Stack{Alignment: layout.NW}.Layout(gtx,
					layout.Expanded(func(gtx layout.Context) layout.Dimensions {

						// Handle user input. We only want to handle pointer events inside the canvas image.
						defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
						event.Op(gtx.Ops, tag)

						for {
							ev, ok := gtx.Event(
								pointer.Filter{
									Target: tag,
									Kinds:  pointer.Press | pointer.Drag | pointer.Move | pointer.Leave | pointer.Enter,
								},
							)

							if !ok {
								break
							}

							p, ok := ev.(pointer.Event) // type assertion
							if !ok {
								continue
							}

							switch p.Kind {
							case pointer.Leave:
								state.mousePositionOnCanvas = mouseIsOutsideCanvas
								gtx.Execute(op.InvalidateCmd{}) // Redraws the ui.

							case pointer.Enter:
								state.mousePositionOnCanvas = p.Position

							case pointer.Press, pointer.Drag:
								state.mousePositionOnCanvas = p.Position
								handlePaint(gtx, state, p)

							case pointer.Move:
								state.mousePositionOnCanvas = p.Position

							default:
								if debug {
									fmt.Printf("Error: UNKNOWN: %+v\n", ev)
								}
							}

							fmt.Printf("Pointer Event: %+v\n", ev)
						}

						// Draw the canvas
						op := paint.NewImageOp(state.canvas)

						return widget.Image{
							Src:   op,
							Fit:   widget.Unscaled,
							Scale: 1.0,
						}.Layout(gtx)
					}),
					layout.Expanded(func(gtx layout.Context) layout.Dimensions {
						doDrawCursor := state.mousePositionOnCanvas != mouseIsOutsideCanvas
						if doDrawCursor {
							// var cursorColor color.NRGBA

							switch state.selectedTool {
							case Brush:
								//TODO: Implement brush cursor

							case Eraser:
								// TODO: Implement eraser cursor

							default:
								if debug {
									fmt.Println("Error: Using unknown tool")
								}
							}

						}

						return layout.Dimensions{Size: gtx.Constraints.Max}
					}),
				)
			})
		},
	)
}

func handlePaint(gtx layout.Context, state *ApplicationState, p pointer.Event) {
	isPaintEvent := p.Kind == pointer.Press || p.Kind == pointer.Drag // sanity check
	if !isPaintEvent {
		return
	}

	switch state.selectedTool {
	case Brush:
		color := state.colorButtons[state.selectedColorIndex].Color
		position := image.Point{X: int(p.Position.X), Y: int(p.Position.Y)}

		rSquared := state.cursorRadius * state.cursorRadius
		for x := position.X - state.cursorRadius; x < position.X+state.cursorRadius; x++ { // Loop through the bounding box of the circle
			for y := position.Y - state.cursorRadius; y < position.Y+state.cursorRadius; y++ {

				// Check if the point is acutally within the circle
				pixelIsWithinCircle := math.Pow(float64(x-position.X), 2)+math.Pow(float64(y-position.Y), 2) < float64(rSquared)
				if !pixelIsWithinCircle {
					continue
				}

				// Paint the pixel
				state.canvas.Set(position.X, position.Y, color)
			}
		}

	case Eraser:
		// position := p.Position
		// circle := Circle{X: position.X, Y: position.Y, Radius: state.cursorRadius, Color: defaultCanvasBackground}
		// state.canvas.paint = append(state.canvas.paint, circle)

	default:
		if debug {
			fmt.Println("Error: Using unknown tool")
		}
		return
	}

}

func fillImageWithColor(img *image.RGBA, col color.Color) {
	if img == nil {
		return
	}

	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			img.Set(x, y, col)
		}
	}
}
