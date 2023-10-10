package main

import (
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/oliveira-a/gochip/pkg/chip8"
)

var (
	c8     *chip8.VM
	square *ebiten.Image
	game   *Game
)

func init() {
	square = ebiten.NewImage(10, 10)
	square.Fill(color.RGBA{R: 255, G: 140, B: 0, A: 1})
	c8 = chip8.New()
	game = &Game{}
	ebiten.SetWindowSize(640, 320)
}

func main() {
	rom, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	stat, err := rom.Stat()
	if err != nil {
		panic(err)
	}

	b := make([]byte, stat.Size())
	_, err = rom.Read(b)
	if err != nil {
		panic(err)
	}
	rom.Close()

	if err = c8.LoadRom(b); err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
}

func (g *Game) Update() error {
	if !c8.Running {
		go c8.Run()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for x := 0; x < chip8.Cols; x++ {
		for y := 0; y < chip8.Rows; y++ {
			if c8.Vram[x][y] == 1 {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(x*10), float64(y*10))
				screen.DrawImage(square, opts)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
