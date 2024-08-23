package main

import (
	"fmt"
	"image/color"
	"log"
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

type ApplicationState struct {
	theme *material.Theme

	brushButton  widget.Clickable
	eraserButton widget.Clickable
	clearButton  widget.Clickable
	selectedTool SelectedTool

	colorButtons       []ColorButtonStyle
	selectedColorIndex int

	canvas *Canvas
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
	state := new(ApplicationState) // store the state on the heap
	*state = ApplicationState{
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
		selectedTool: Brush,
		canvas: &Canvas{width: 1000, height: 800, paint: make([]Circle, 0)},
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

	// Handle tool button clicks
	if state.brushButton.Clicked(gtx) {
		state.selectedTool = Brush
		fmt.Println("Current tool: ", state.selectedTool)
	}

	if state.eraserButton.Clicked(gtx) {
		state.selectedTool = Eraser
		fmt.Println("Current tool: ", state.selectedTool)
	}

	if state.clearButton.Clicked(gtx) {
		state.canvas.paint = make([]Circle, 0)
		fmt.Println("Canvas cleared")
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

			fmt.Println("Selected color: ", btn.Label)
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
	r := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	defer r.Pop()

	for {	
		ev, ok := gtx.Event(
		  pointer.Filter{
			Target: tag,
			Kinds:  pointer.Press | pointer.Drag | pointer.Move,
		  },
		)

		if !ok {
		  break
		}

		p, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		
		switch p.Kind {
			case pointer.Press, pointer.Drag:
				fmt.Printf("PRESS | DRAG: %+v\n", ev)
				color := state.colorButtons[state.selectedColorIndex].Color
				circle := Circle{X: ev.(pointer.Event).Position.X, Y: ev.(pointer.Event).Position.Y, Radius: 10, Color: color}
				state.canvas.paint = append(state.canvas.paint, circle)

			case pointer.Move:
				fmt.Printf("MOVE: %+v\n", ev)
				state.mousePositionOnCanvas = p.Position

			default:
				fmt.Printf("UNKNOWN: %+v\n", ev)
		}
	}

	// Render the canvas
    return layout.Stack{Alignment: layout.Center}.Layout(gtx,
        layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// Draw the paint
			for _, circle := range state.canvas.paint {
				drawCircle(gtx, circle.X, circle.Y, circle.Radius, circle.Color)
			}

			// Draw the cursor
			cursorColor := state.colorButtons[state.selectedColorIndex].Color
			drawCircle(gtx, state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y, 10, cursorColor)

            return layout.Dimensions{Size: gtx.Constraints.Max}
        }),
    )
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
