package main

import (
	"syscall/js"
	"time"

	"github.com/oliveira-a/gochip/chip8"
)

var (
	c8 *chip8.VM

	canvasWidth  = 640
	canvasHeight = 320

	canvas js.Value
)

func init() {
	// Setup the virtual machine
	c8 = chip8.New(make(chan int))

	doc := js.Global().Get("document")
	doc.Call(
		"addEventListener",
		"keydown",
		js.FuncOf(keyDownListener),
	)
	doc.Call(
		"addEventListener",
		"keyup",
		js.FuncOf(keyUpListener),
	)

	// Setup the canvas
	canvas = doc.Call("getElementById", "gameCanvas")
	canvas.Set("width", canvasWidth)
	canvas.Set("height", canvasHeight)

	// File loading setup
	fileInput := doc.Call("getElementById", "fileInput")
	fileInput.Set("oninput", js.FuncOf(func(v js.Value, x []js.Value) any {
		fileInput.Get("files").Call("item", 0).Call("arrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
			data := js.Global().Get("Uint8Array").New(x[0])
			dst := make([]byte, data.Get("length").Int())
			js.CopyBytesToGo(dst, data)

			c8 = chip8.New(make(chan int))

			if err := c8.LoadRom(dst); err != nil {
				panic(err)
			}

			return nil
		}))

		return nil
	}))
}

func main() {
	loop()

	select {}
}

func loop() {
	tick := time.Tick(8 * time.Millisecond)
	for {
		select {
		case <-tick:
			c8.Cycle()
			render()
		}
	}
}

func render() {
	ctx := canvas.Call("getContext", "2d")
	ctx.Call("clearRect", 0, 0, canvasWidth, canvasHeight)

	// background
	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", 0, 0, canvasWidth, canvasHeight)

	for x := 0; x < chip8.Cols; x++ {
		for y := 0; y < chip8.Rows; y++ {
			if c8.Vram[x][y] == 1 {
				x, y := x*10, y*10
				ctx.Call("beginPath")
				ctx.Set("fillStyle", "orange")
				ctx.Call("fillRect", x, y, 10, 10)
			}
		}
	}
}

func update(key string, pressed uint8) {
	switch key {
	case "1":
		c8.Keys[0x1] = pressed
	case "2":
		c8.Keys[0x2] = pressed
	case "3":
		c8.Keys[0x3] = pressed
	case "4":
		c8.Keys[0xc] = pressed

	case "q":
		c8.Keys[0x4] = pressed
	case "w":
		c8.Keys[0x5] = pressed
	case "e":
		c8.Keys[0x6] = pressed
	case "r":
		c8.Keys[0xd] = pressed

	case "a":
		c8.Keys[0x7] = pressed
	case "s":
		c8.Keys[0x8] = pressed
	case "d":
		c8.Keys[0x9] = pressed
	case "f":
		c8.Keys[0xe] = pressed

	case "z":
		c8.Keys[0xa] = pressed
	case "x":
		c8.Keys[0x0] = pressed
	case "c":
		c8.Keys[0xb] = pressed
	case "v":
		c8.Keys[0xd] = pressed
	}
}

func keyDownListener(this js.Value, p []js.Value) interface{} {
	event := p[0]
	key := event.Get("key").String()

	update(key, uint8(1))

	return nil
}

func keyUpListener(this js.Value, p []js.Value) interface{} {
	event := p[0]
	key := event.Get("key").String()

	update(key, uint8(0))

	return nil
}
