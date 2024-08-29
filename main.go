package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
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
	"gioui.org/x/explorer"
)

var debug = false

type GemPaintState struct {
	theme *material.Theme

	brushButton  widget.Clickable
	eraserButton widget.Clickable
	selectedTool SelectedTool

	increaseButton widget.Clickable
	decreaseButton widget.Clickable
	cursorRadius   int

	clearButton widget.Clickable
	saveButton  widget.Clickable

	colorButtons       []ColorButtonStyle
	selectedColorIndex int

	canvas                *image.RGBA
	mousePositionOnCanvas f32.Point
	previousPaintPosition f32.Point

	expl      *explorer.Explorer
	saveImage image.RGBA
	saveErr   error
	saveChan  chan error
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
	state := new(GemPaintState) // store the state on the heap
	*state = GemPaintState{     // TODO: move this to run()
		theme:        material.NewTheme(),
		selectedTool: Brush,
		cursorRadius: defaultCursorRadius,
		colorButtons: []ColorButtonStyle{
			{Color: red, Label: "Red", Clickable: &widget.Clickable{}, isSelected: true}, //TODO: what needs to be initialized?
			{Color: orange, Label: "Orange", Clickable: &widget.Clickable{}},
			{Color: green, Label: "Green", Clickable: &widget.Clickable{}},
			{Color: blue, Label: "Blue", Clickable: &widget.Clickable{}},
			{Color: yellow, Label: "Yellow", Clickable: &widget.Clickable{}},
			{Color: purple, Label: "Purple", Clickable: &widget.Clickable{}},
			{Color: darkGray, Label: "Gray", Clickable: &widget.Clickable{}},
		},
		selectedColorIndex:    0,
		canvas:                image.NewRGBA(defaultCanvasDimensions),
		mousePositionOnCanvas: mouseIsOutsideCanvas,
	}

	fillImageWithColor(state.canvas, defaultCanvasColor)

	go func() {
		window := new(app.Window)
		window.Option(app.Title("GemBoard"))
		window.Option(app.Size(unit.Dp(1000), unit.Dp(800)))
		state.expl = explorer.NewExplorer(window)

		// Run the program
		err := run(window, state)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func run(window *app.Window, state *GemPaintState) error {
	theme := material.NewTheme()
	state.saveChan = make(chan error)

	var ops op.Ops

	for {
		e := window.Event()

		state.expl.ListenEvents(e)
		switch e := e.(type) {
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

		// Handle save errors
		select {
		case state.saveErr = <-state.saveChan:
			if debug {
				fmt.Println("Error: ", state.saveErr)
			}
				
			window.Invalidate()
		default:
			// No save errors to process
		}

	}
}

func layoutSidebar(gtx layout.Context, state *GemPaintState, theme *material.Theme) layout.Dimensions {

	// Handle tool button clicks
	if state.brushButton.Clicked(gtx) {
		state.selectedTool = Brush
		state.previousPaintPosition = mouseIsOutsideCanvas
		if debug {
			fmt.Println("Current tool: ", state.selectedTool)
		}
	}

	if state.eraserButton.Clicked(gtx) {
		state.selectedTool = Eraser
		state.previousPaintPosition = mouseIsOutsideCanvas
		if debug {
			fmt.Println("Current tool: ", state.selectedTool)
		}
	}

	if state.increaseButton.Clicked(gtx) {
		if state.cursorRadius < maximumCursorRadius {
			state.cursorRadius += cursorRadiusChangeStep
		}
		if debug {
			fmt.Println("Cursor radius: ", state.cursorRadius)
		}
	}

	if state.decreaseButton.Clicked(gtx) {
		if state.cursorRadius > minimumCursorRadius {
			state.cursorRadius -= cursorRadiusChangeStep
		}
		if debug {
			fmt.Println("Cursor radius: ", state.cursorRadius)
		}
	}

	if state.clearButton.Clicked(gtx) {
		state.canvas = image.NewRGBA(defaultCanvasDimensions)
		fillImageWithColor(state.canvas, defaultCanvasColor)
		if debug {
			fmt.Println("Canvas cleared")
		}
	}

	if state.saveButton.Clicked(gtx) {
		go func(img image.RGBA) {
			if state.canvas == nil {
				state.saveChan <- fmt.Errorf("no image to save")
				return
			}

			extension := "png"
			fileName := "gem." + extension
			file, err := state.expl.CreateFile(fileName)
			if err != nil {
				state.saveChan <- fmt.Errorf("failed exporting image file: %w", err)
				return
			}
			defer func() {
				state.saveChan <- file.Close()
			}()

			if err := png.Encode(file, state.canvas); err != nil {
				state.saveChan <- fmt.Errorf("failed encoding PNG file: %w", err)
				return
			}
		}(state.saveImage)
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
			btn.isSelected = true

			if debug {
				fmt.Println("Selected color: ", btn.Label)
			}
		}
	}

	// Tool buttons
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.brushButton, BrushIcon, state.selectedTool == Brush, golangBlue, lightGray, "Brush").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.eraserButton, EraserIcon, state.selectedTool == Eraser, golangBlue, lightGray, "Brush").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.increaseButton, AddIcon, false, golangBlue, lightGray, "Increase").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.decreaseButton, MinusIcon, false, golangBlue, lightGray, "Decrease").Layout(gtx)
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
			return ToolButton(theme, &state.clearButton, ClearIcon, false, golangBlue, lightGray, "Clear").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.saveButton, SaveIcon, false, golangBlue, lightGray, "Save").Layout(gtx)
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

var tag = new(bool) // tag is a unique identifier for the canvas

func layoutCanvas(gtx layout.Context, state *GemPaintState) layout.Dimensions {

	// Render the canvas
	return layout.Stack{Alignment: layout.NW}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, defaultCanvasBackground)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
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

				pointerEvent, ok := ev.(pointer.Event) // type assertion
				if !ok {
					continue
				}

				switch pointerEvent.Kind {
				case pointer.Leave:
					state.mousePositionOnCanvas = mouseIsOutsideCanvas
					gtx.Execute(op.InvalidateCmd{}) // Redraws the ui.

				case pointer.Enter:
					state.mousePositionOnCanvas = pointerEvent.Position

				case pointer.Press, pointer.Drag:
					state.mousePositionOnCanvas = pointerEvent.Position
					handlePaint(state, pointerEvent)

				case pointer.Move:
					state.mousePositionOnCanvas = pointerEvent.Position

				default:
					if debug {
						fmt.Printf("Error: UNKNOWN: %+v\n", ev)
					}
				}

				// fmt.Printf("Pointer Event: %+v\n", ev)
			}

			// Draw the canvas
			op := paint.NewImageOp(state.canvas)

			return widget.Image{
				Src: op,
				Fit: widget.Unscaled,

				Scale: 1.0,
			}.Layout(gtx)
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			doDrawCursor := state.mousePositionOnCanvas != mouseIsOutsideCanvas
			if !doDrawCursor {
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}

			var cursorColor color.NRGBA

			switch state.selectedTool {
			case Brush:
				cursorColor = state.colorButtons[state.selectedColorIndex].Color
				drawCircle(gtx, state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y, float32(state.cursorRadius), cursorColor)

			case Eraser:
				cursorColor = defaultCanvasColor
				drawCircle(gtx, state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y, float32(state.cursorRadius), lightGray)
				drawCircle(gtx, state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y, float32(state.cursorRadius-1), cursorColor)

			default:
				if debug {
					fmt.Println("Error: Using unknown tool")
				}
			}

			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
	)
}

// CURSOR STUFF

func handlePaint(state *GemPaintState, p pointer.Event) {
	isPaintEvent := p.Kind == pointer.Press || p.Kind == pointer.Drag // sanity check
	if !isPaintEvent {
		return
	}

	switch state.selectedTool {
	case Brush:
		color := state.colorButtons[state.selectedColorIndex].Color
		positionOnCanvas := image.Point{X: int(p.Position.X), Y: int(p.Position.Y)}
		paintCircle(state.canvas, positionOnCanvas, state.cursorRadius, color)

		// Due to the way the ui frameworks returns pointer drag events, if the user drags the mouse too quickly, some pixels will be skipped.
		// To fix this, we need to fill in pixels between the previous and current mouse positions, that is, use interpolation.
		previousPaintPositionIsOutsideCanvas := state.previousPaintPosition == mouseIsOutsideCanvas
		if !previousPaintPositionIsOutsideCanvas && p.Kind == pointer.Drag {
			interpolatePaintBetweenPoints(state.previousPaintPosition, p.Position, state.canvas, state.cursorRadius, color)
		}

		// Update at the end of the paint operation
		state.previousPaintPosition = p.Position

	case Eraser:
		color := defaultCanvasColor
		positionOnCanvas := image.Point{X: int(p.Position.X), Y: int(p.Position.Y)}
		paintCircle(state.canvas, positionOnCanvas, state.cursorRadius, color)

		previousPaintPositionIsOutsideCanvas := state.previousPaintPosition == mouseIsOutsideCanvas
		if !previousPaintPositionIsOutsideCanvas && p.Kind == pointer.Drag {
			interpolatePaintBetweenPoints(state.previousPaintPosition, p.Position, state.canvas, state.cursorRadius, color)
		}

		state.previousPaintPosition = p.Position

	default:
		if debug {
			fmt.Println("Error: Using unknown tool")
		}
		return
	}

}

func interpolatePaintBetweenPoints(start, end f32.Point, canvas *image.RGBA, radius int, color color.Color) {
	dx := end.X - start.X
	dy := end.Y - start.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	step := float32(radius) / 4.0 // Step size based on brush radius
	if distance > step {
		// Interpolate points between prevPos and position
		for t := float32(0.0); t <= distance; t += step {
			interpX := int(start.X + t/distance*dx)
			interpY := int(start.Y + t/distance*dy)
			interpPosition := image.Point{X: interpX, Y: interpY}

			paintCircle(canvas, interpPosition, radius, color)
		}
	}
}

func paintCircle(canvas *image.RGBA, position image.Point, radius int, color color.Color) {
	rSquared := radius * radius
	for x := position.X - radius; x <= position.X+radius; x++ { // Loop through the bounding box of the circle, ie, the square
		for y := position.Y - radius; y <= position.Y+radius; y++ {

			dx, dy := x-position.X, y-position.Y
			pixelIsWithinCircle := dx*dx+dy*dy < rSquared
			if !pixelIsWithinCircle {
				continue
			}

			// Ensure the coordinates are within bounds before painting
			if x < 0 || x >= canvas.Bounds().Dx() || y < 0 || y >= canvas.Bounds().Dy() {
				continue
			}

			canvas.Set(x, y, color)
		}
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

func drawCircle(gtx layout.Context, x, y, radius float32, fillcolor color.NRGBA) {
	path := new(clip.Path)
	ops := gtx.Ops
	const k = 0.551915024494 // http://spencermortensen.com/articles/bezier-circle/

	path.Begin(ops)
	path.Move(f32.Point{X: x + radius, Y: y})
	path.Cube(f32.Point{X: 0, Y: radius * k}, f32.Point{X: -radius + radius*k, Y: radius}, f32.Point{X: -radius, Y: radius})    // SE
	path.Cube(f32.Point{X: -radius * k, Y: 0}, f32.Point{X: -radius, Y: -radius + radius*k}, f32.Point{X: -radius, Y: -radius}) // SW
	path.Cube(f32.Point{X: 0, Y: -radius * k}, f32.Point{X: radius - radius*k, Y: -radius}, f32.Point{X: radius, Y: -radius})   // NW
	path.Cube(f32.Point{X: radius * k, Y: 0}, f32.Point{X: radius, Y: radius - radius*k}, f32.Point{X: radius, Y: radius})      // NE
	path.Close()
	stack := clip.Outline{Path: path.End()}.Op().Push(ops)
	paint.ColorOp{Color: fillcolor}.Add(ops)
	paint.PaintOp{}.Add(ops)
	stack.Pop()
}
