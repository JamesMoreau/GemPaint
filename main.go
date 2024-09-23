package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"runtime"

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
	BucketButton widget.Clickable
	selectedTool SelectedTool

	increaseButton widget.Clickable
	decreaseButton widget.Clickable
	cursorRadius   int

	clearButton widget.Clickable
	saveButton  widget.Clickable

	colorButtons       []ColorButtonStyle
	selectedColorIndex int

	sidebarButtons layout.List

	canvas                *image.RGBA
	canvasInputTag        bool
	mousePositionOnCanvas f32.Point
	previousPaintPosition f32.Point

	expl *explorer.Explorer

	debug bool
}

type SelectedTool string

const (
	Brush  SelectedTool = "Brush"
	Eraser SelectedTool = "Eraser"
	Bucket SelectedTool = "Bucket"
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

	go func() {
		window := new(app.Window)
		window.Option(app.Title("GemPaint"))
		window.Option(app.Size(unit.Dp(1920), unit.Dp(1080)))

		// Run the program
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func run(window *app.Window) error {

	// Initialize the application state
	state := GemPaintState{
		theme:        material.NewTheme(),
		selectedTool: Brush,
		cursorRadius: defaultCursorRadius,
		colorButtons: []ColorButtonStyle{
			{Color: red, Label: "Red", Clickable: &widget.Clickable{}},
			{Color: orange, Label: "Orange", Clickable: &widget.Clickable{}},
			{Color: green, Label: "Green", Clickable: &widget.Clickable{}},
			{Color: blue, Label: "Blue", Clickable: &widget.Clickable{}},
			{Color: yellow, Label: "Yellow", Clickable: &widget.Clickable{}},
			{Color: purple, Label: "Purple", Clickable: &widget.Clickable{}},
			{Color: darkGray, Label: "Gray", Clickable: &widget.Clickable{}},
		},
		selectedColorIndex:    0,
		sidebarButtons:        layout.List{Axis: layout.Vertical},
		canvas:                image.NewRGBA(defaultCanvasDimensions),
		mousePositionOnCanvas: mouseIsOutsideCanvas,
		expl:                  explorer.NewExplorer(window),
	}

	fillImageWithColor(state.canvas, defaultCanvasColor)
	theme := material.NewTheme()

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
									return layoutSidebar(gtx, &state, theme)
								},
							),
							layout.Rigid(
								func(gtx layout.Context) layout.Dimensions {
									return layoutCanvas(gtx, &state)
								},
							),
						)
					},
				),
				layout.Stacked(
					func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(32).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							if debug {
								return material.Body1(theme, fmt.Sprintf("🐭: %.2f, %.2f", state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y)).Layout(gtx)
							}
							return layout.Dimensions{Size: gtx.Constraints.Min}
						})
					},
				),
			)

			e.Frame(gtx.Ops)
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

	if state.BucketButton.Clicked(gtx) {
		state.selectedTool = Bucket
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
		go func() { // Do not block the ui thread

			if state.canvas == nil {
				if debug {
					fmt.Println("Error: No image to save")
				}
				return
			}

			extension := "png"
			fileName := "gem." + extension

			// Depending on the platform, how we access the file system will differ
			platform := runtime.GOOS
			if debug {
				fmt.Println("Saving on GOOS: ", platform)
			}

			saveOnPlatform(state, fileName)

		}()
	}

	// Handle color button clicks
	for i := range state.colorButtons {
		btn := &state.colorButtons[i]
		wasClicked := btn.Clickable.Clicked(gtx)

		// Dynamically set isSelected based on selectedColorIndex
		btn.isSelected = (i == state.selectedColorIndex)

		if wasClicked {
			state.selectedColorIndex = i

			if debug {
				fmt.Println("Selected color: ", btn.Label)
			}
		}
	}

	// Tool buttons
	children := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.brushButton, BrushIcon, state.selectedTool == Brush, golangBlue, lightGray, "Brush").Layout(gtx)
		},
		layout.Spacer{Height: unit.Dp(8)}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.eraserButton, EraserIcon, state.selectedTool == Eraser, golangBlue, lightGray, "Brush").Layout(gtx)
		},
		layout.Spacer{Height: unit.Dp(8)}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.BucketButton, BucketIcon, state.selectedTool == Bucket, golangBlue, lightGray, "Bucket").Layout(gtx)
		},
		layout.Spacer{Height: unit.Dp(8)}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.increaseButton, AddIcon, false, golangBlue, lightGray, "Increase").Layout(gtx)
		},
		layout.Spacer{Height: unit.Dp(8)}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.decreaseButton, MinusIcon, false, golangBlue, lightGray, "Decrease").Layout(gtx)
		},
		layout.Spacer{Height: unit.Dp(16)}.Layout,
	}

	// Color buttons
	for i := range state.colorButtons {
		btn := &state.colorButtons[i]
		children = append(children,
			func(gtx layout.Context) layout.Dimensions {
				return btn.Layout(gtx, theme)
			},

			layout.Spacer{Height: unit.Dp(8)}.Layout,
		)
	}

	// Other buttons
	children = append(children,
		layout.Spacer{Height: unit.Dp(16)}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.clearButton, ClearIcon, false, golangBlue, lightGray, "Clear").Layout(gtx)
		},
		layout.Spacer{Height: unit.Dp(8)}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ToolButton(theme, &state.saveButton, SaveIcon, false, golangBlue, lightGray, "Save").Layout(gtx)
		},
	)

	return layout.Background{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, softBlue)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return state.sidebarButtons.Layout(gtx, len(children), func(gtx layout.Context, i int) layout.Dimensions {
				return children[i](gtx)
			})
		})
	})
}

// var canvasInputTag bool // tag is a unique identifier for the canvas

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
			event.Op(gtx.Ops, state.canvasInputTag)

			for {
				ev, ok := gtx.Event(
					pointer.Filter{
						Target: state.canvasInputTag,
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
				Src:   op,
				Fit:   widget.Unscaled,
				Scale: 1.0 / gtx.Metric.PxPerDp,
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

			case Bucket:
				cursorColor = state.colorButtons[state.selectedColorIndex].Color
				drawCircle(gtx, state.mousePositionOnCanvas.X, state.mousePositionOnCanvas.Y, 5, cursorColor)
			default:
				if debug {
					fmt.Println("Error: Using unknown tool")
				}
			}

			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
	)
}

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

	case Bucket:
		if p.Kind != pointer.Press { // We only want to fill the bucket on the initial click
			return
		}

		positionOnCanvas := image.Point{X: int(p.Position.X), Y: int(p.Position.Y)}
		newColor := state.colorButtons[state.selectedColorIndex].Color

		// Find all pixels that need to be replaced with the new color that are connected to the clicked pixel
		err := floodFill(state.canvas, positionOnCanvas, newColor)
		if err != nil && debug {
			fmt.Println(err)
		}

	default:
		if debug {
			fmt.Println("Error: Using unknown tool")
		}
		return
	}

}

func floodFill(canvas *image.RGBA, start image.Point, newColor color.Color) error {
	if !start.In(canvas.Rect) {
		return fmt.Errorf("start point is outside canvas") // Nothing to be done!
	}

	oldColor := canvas.At(start.X, start.Y)

	if colorsAreEqual(oldColor, newColor) {
		return fmt.Errorf("old color is the same as new fill color")
	}

	queue := []image.Point{start}

	for len(queue) > 0 {
		// Dequeue a point
		currentPixel := queue[0]
		queue = queue[1:]

		if !currentPixel.In(canvas.Rect) {
			continue
		}

		currentPixelColor := canvas.At(currentPixel.X, currentPixel.Y)
		if !colorsAreEqual(currentPixelColor, oldColor) {
			continue
		}

		canvas.Set(currentPixel.X, currentPixel.Y, newColor)

		// Add the neighboring pixels to the queue
		queue = append(queue, image.Point{X: currentPixel.X + 1, Y: currentPixel.Y})
		queue = append(queue, image.Point{X: currentPixel.X - 1, Y: currentPixel.Y})
		queue = append(queue, image.Point{X: currentPixel.X, Y: currentPixel.Y + 1})
		queue = append(queue, image.Point{X: currentPixel.X, Y: currentPixel.Y - 1})
	}

	return nil
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

			isWithinBounds := x >= 0 && x < canvas.Bounds().Dx() && y >= 0 && y < canvas.Bounds().Dy()
			if !isWithinBounds {
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

func colorsAreEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
