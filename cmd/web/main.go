package main

import (
	_ "fmt"
	"time"
	"syscall/js"

	"github.com/oliveira-a/gochip/pkg/chip8"
)

var (
	c8       *chip8.VM
	beepChan chan int

	posX, posY    float64

	canvasWidth  = 640
	canvasHeight = 320

	canvas js.Value
)

func init() {
	// Setup the virtual machine
	beepChan = make(chan int)
	c8 = chip8.New(beepChan)

	doc := js.Global().Get("document")
	doc.Call(
		"addEventListener",
		"keydown",
		js.FuncOf(keyListener),
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
			
			c8 = chip8.New(beepChan)

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
	tick := time.Tick(16 * time.Millisecond)
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

func update(key string) {
	// TODO: set the positing in chip8
	switch key {
	case "A":
		posY -= 10
	case "B":
		posY += 10
	case "ArrowLeft":
		posX -= 10
	case "ArrowRight":
		posX += 10
	}
}

func keyListener(this js.Value, p []js.Value) interface{} {
	event := p[0]
	key := event.Get("key").String()

	update(key)

	return nil
}
