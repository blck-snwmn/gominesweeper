package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/blck-snwmn/gominesweeper"
)

func main() {
	const (
		height     = 9
		width      = 9
		cellSize   = 50
		maxBombNum = 15
	)
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	canvas.Set("height", js.ValueOf(height*cellSize))
	canvas.Set("width", js.ValueOf(width*cellSize))

	canvasCtx := canvas.Call("getContext", "2d")
	// font
	canvasCtx.Call("scale", 1, 1)
	fontFormat := "%dpx "
	for _, f := range strings.Split(canvasCtx.Get("font").String(), " ")[1:] {
		fontFormat += f
		fontFormat += " "
	}
	minesiper := gominesweeper.New(height, width, maxBombNum)

	for h := 0; h < minesiper.Height; h++ {
		for w := 0; w < minesiper.Width; w++ {
			var color string
			if (h+w)%2 == 0 {
				color = "#E6FFE9"
			} else {
				color = "#AEFFBD"
			}
			canvasCtx.Set("fillStyle", color)
			canvasCtx.Call("fillRect", w*cellSize, h*cellSize, cellSize, cellSize)
		}
	}

	go func() {
		sendCh := minesiper.GetNotify()
		for {
			select {
			case changes := <-sendCh:
				for ci := range changes {
					var color string
					if ci.State == gominesweeper.Bomb {
						color = "#FF0000"
					} else if (ci.X+ci.Y)%2 == 0 {
						color = "#FFFFEE"
					} else {
						color = "#FFFF99"
					}
					canvasCtx.Set("fillStyle", color)
					canvasCtx.Call("fillRect", ci.X*cellSize, ci.Y*cellSize, cellSize, cellSize)
					canvasCtx.Set("fillStyle", "black")
					canvasCtx.Set("font", fmt.Sprintf(fontFormat, cellSize/2))
					if ci.State == gominesweeper.Bomb {
						canvasCtx.Call("fillText", "x", ci.X*cellSize, (ci.Y+1)*cellSize)
					} else {
						canvasCtx.Call("fillText", ci.NumOfNearbyBomb, ci.X*cellSize, (ci.Y+1)*cellSize)
					}
				}
			}
		}
	}()

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
		x := (ox / cellSize)
		y := (oy / cellSize)
		fmt.Println(ox, oy, (ox/cellSize)+1, (oy/cellSize)+1)
		minesiper.PressCell(y, x)
		fmt.Println("send")
		return nil
	}))
	select {}
}
