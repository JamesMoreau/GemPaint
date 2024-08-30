// +build js,wasm

package main

import (
	"syscall/js"
	"fmt"
	"bytes"
	"image/png"
)

func saveOnWeb(state *GemPaintState, fileName string) {
	if debug {
		fmt.Println("Saving image on wasm/js")
	}
	
	// Convert the image.RGBA to a JavaScript Uint8Array
	buf := bytes.Buffer{}
	if err := png.Encode(&buf, state.canvas); err != nil {
		if debug {
			fmt.Println("Error: ", err)
		}
		return
	}

	jsData := js.Global().Get("Uint8Array").New(buf.Len())
	js.CopyBytesToJS(jsData, buf.Bytes())

	// Create a Blob from the data
	blob := js.Global().Get("Blob").New([]interface{}{jsData})

	// Create a link element for the download
	a := js.Global().Get("document").Call("createElement", "a")
	a.Set("href", js.Global().Get("URL").Call("createObjectURL", blob))
	a.Set("download", fileName)

	// Simulate a click on the link
	a.Call("click")
}