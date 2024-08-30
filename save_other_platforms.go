//go:build !js && !wasm

package main

import (
	"fmt"
	"image/png"
)

func saveOnPlatform(state *GemPaintState, fileName string) {
	file, err := state.expl.CreateFile(fileName)
	if err != nil {
		if debug {
			fmt.Println("Error: ", err)
		}
		return
	}

	if err := png.Encode(file, state.canvas); err != nil {
		if debug {
			fmt.Println("Error: ", err)
		}
		return
	}
}
