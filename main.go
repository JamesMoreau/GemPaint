package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"time"

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

	cursorSize float32

	colorButtons       []ColorButtonStyle
	selectedColorIndex int

	canvas                *Canvas
	mousePositionOnCanvas f32.Point
}

type Canvas struct {
	width  int
	height int
	paint  []Circle
}

type Circle struct {
	X, Y   float32
	Radius float32
	Color  color.NRGBA
}

type SelectedTool string

const (
	Brush  SelectedTool = "Brush"
	Eraser SelectedTool = "Eraser"
)

var defaultMargin = unit.Dp(10)

func main() {
	// Get arguments
	debugVariable := os.Getenv("DEBUG")
	debug = debugVariable == "true"

	state := new(ApplicationState) // store the state on the heap
	*state = ApplicationState{
		theme:        material.NewTheme(),
		selectedTool: Brush,
		cursorSize:   10,
		colorButtons: []ColorButtonStyle{
			{Color: color.NRGBA{R: 255, G: 0, B: 0, A: 255}, Label: "Red", Clickable: &widget.Clickable{}, isSelected: true},
			{Color: color.NRGBA{R: 255, G: 165, B: 0, A: 255}, Label: "Orange", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 0, G: 255, B: 0, A: 255}, Label: "Green", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 0, G: 0, B: 255, A: 255}, Label: "Blue", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 255, G: 255, B: 0, A: 255}, Label: "Yellow", Clickable: &widget.Clickable{}},
			{Color: color.NRGBA{R: 128, G: 0, B: 128, A: 255}, Label: "Purple", Clickable: &widget.Clickable{}},
		},
		selectedColorIndex:    0,
		canvas:                &Canvas{width: 1000, height: 800, paint: make([]Circle, 0)},
		mousePositionOnCanvas: mouseIsOutsideCanvas,
	}

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

	var lastFrame time.Time
	var fps float64

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:

			// Calculate FPS
			now := time.Now()
			if !lastFrame.IsZero() {
				elapsed := now.Sub(lastFrame).Seconds()
				fps = 1.0 / elapsed
			}
			lastFrame = now
			fmt.Println("FPS: ", fps)

			gtx := app.NewContext(&ops, e)

			layout.Stack{Alignment: layout.NE}.Layout(gtx,
				layout.Expanded(
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
		state.canvas.paint = make([]Circle, 0)
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
		defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, softBlue)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(defaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
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

var tag = new(bool)

func layoutCanvas(gtx layout.Context, state *ApplicationState) layout.Dimensions {

	// Handle user input
	r := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	defer r.Pop()

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

		case pointer.Enter:

		case pointer.Press, pointer.Drag:
			handlePaint(state, p)

		case pointer.Move:
			state.mousePositionOnCanvas = p.Position

		default:
			if debug {
				fmt.Printf("Error: UNKNOWN: %+v\n", ev)
			}
		}

		// fmt.Printf("Pointer Event: %+v\n", ev)
	}

	// Render the canvas
	return layout.Background{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, defaultCanvasBackground)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				// Draw the paint
				for _, circle := range state.canvas.paint {
					drawCircle(gtx, circle.X, circle.Y, circle.Radius, circle.Color)
				}

				// Draw the cursor
				doDrawCursor := state.mousePositionOnCanvas != mouseIsOutsideCanvas
				if doDrawCursor {
					var cursorColor color.NRGBA

					switch state.selectedTool {
					case Brush:
						cursorColor = state.colorButtons[state.selectedColorIndex].Color

					case Eraser:
						cursorColor = defaultCanvasBackground

					default:
						if debug {
							fmt.Println("Error: Using unknown tool")
						}
					}

					drawCircle(gtx, state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y, state.cursorSize, cursorColor)
				}

				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
		)
	})
}

func handlePaint(state *ApplicationState, p pointer.Event) {
	isPaintEvent := p.Kind == pointer.Press || p.Kind == pointer.Drag
	if !isPaintEvent {
		return
	}

	switch state.selectedTool {
	case Brush: // Draw a circle on the canvas
		color := state.colorButtons[state.selectedColorIndex].Color
		position := p.Position
		circle := Circle{X: position.X, Y: position.Y, Radius: state.cursorSize, Color: color}
		state.canvas.paint = append(state.canvas.paint, circle)

		// If the mouse is moved quickly, we need to add interpolated paint between the last and current mouse positions.
		if p.Kind != pointer.Drag {
			return
		}

		paintLength := len(state.canvas.paint)
		if paintLength == 0 {
			return
		}

		previousPosition := state.canvas.paint[paintLength-2]
		// fmt.Println("previousPosition: ", previousPosition)
		// fmt.Println("position: ", position)
		dx := position.X - previousPosition.X
		dy := position.Y - previousPosition.Y
		fmt.Println("dx: ", dx, "dy: ", dy)
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		// fmt.Println("distance: ", distance)

		// Check if the distance is great enough to warrant interpolation.
		threshould := state.cursorSize / 2
		if distance < threshould {
			return
		}

		// Add interpolated circles between the last and current mouse position.
		numInterpolatedCircles := int(distance / threshould)
		for i := 0; i < numInterpolatedCircles; i++ {
			interpolatedX := previousPosition.X + dx*float32(i)/float32(numInterpolatedCircles)
			interpolatedY := previousPosition.Y + dy*float32(i)/float32(numInterpolatedCircles)
			interpolatedCircle := Circle{X: interpolatedX, Y: interpolatedY, Radius: state.cursorSize, Color: color}
			state.canvas.paint = append(state.canvas.paint, interpolatedCircle)
			fmt.Println("Interpolated circle added")
		}

	case Eraser: // "Erase" the paint by drawing a circle the same color as the canvas background.
		position := p.Position
		circle := Circle{X: position.X, Y: position.Y, Radius: state.cursorSize, Color: defaultCanvasBackground}
		state.canvas.paint = append(state.canvas.paint, circle)

	default:
		if debug {
			fmt.Println("Error: Using unknown tool")
		}
		return
	}

}

func drawCircle(gtx layout.Context, x, y, radius float32, fillcolor color.NRGBA) {
	path := new(clip.Path)
	const k = 0.551915024494 // http://spencermortensen.com/articles/bezier-circle/

	path.Begin(gtx.Ops)
	path.Move(f32.Point{X: x + radius, Y: y})
	path.Cube(f32.Point{X: 0, Y: radius * k}, f32.Point{X: -radius + radius*k, Y: radius}, f32.Point{X: -radius, Y: radius})    // SE
	path.Cube(f32.Point{X: -radius * k, Y: 0}, f32.Point{X: -radius, Y: -radius + radius*k}, f32.Point{X: -radius, Y: -radius}) // SW
	path.Cube(f32.Point{X: 0, Y: -radius * k}, f32.Point{X: radius - radius*k, Y: -radius}, f32.Point{X: radius, Y: -radius})   // NW
	path.Cube(f32.Point{X: radius * k, Y: 0}, f32.Point{X: radius, Y: radius - radius*k}, f32.Point{X: radius, Y: radius})      // NE
	path.Close()
	stack := clip.Outline{Path: path.End()}.Op().Push(gtx.Ops)
	paint.ColorOp{Color: fillcolor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Pop()
}
