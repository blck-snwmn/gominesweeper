package main

import (
	"fmt"
	"syscall/js"

	"github.com/blck-snwmn/gominesweeper"
)

func main() {
	const (
		height   = 30
		width    = 38
		cellSize = 10
	)
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	canvas.Set("height", js.ValueOf(height*cellSize))
	canvas.Set("width", js.ValueOf(width*cellSize))

	canvasCtx := canvas.Call("getContext", "2d")
	canvasCtx.Call("scale", 1, 1)
	minesiper := gominesweeper.New(height, width)

	for h := 0; h < minesiper.Height; h++ {
		for w := 0; w < minesiper.Width; w++ {
			var color string
			if (h+w)%2 == 0 {
				color = "white"
			} else {
				color = "black"
			}
			canvasCtx.Set("fillStyle", color)
			canvasCtx.Call("fillRect", w*cellSize, h*cellSize, cellSize, cellSize)
		}
	}

	canvas.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		getInt := func(s string) int {
			jv := args[0].Get(s)
			if jv.Type() == js.TypeNumber {
				return jv.Int()
			}
			return 0
		}
		ox := getInt("offsetX")
		oy := getInt("offsetY")
		fmt.Println(ox, oy, (ox/cellSize)+1, (oy/cellSize)+1)
		return nil
	}))
	select {}
}
