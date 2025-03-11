package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	"io/fs"
	"log"
	"time"

	"github.com/ebitenui/ebitenui"

	_ "github.com/hajimehoshi/ebiten/v2/text/v2"
	_ "golang.org/x/image/font/gofont/goregular"

	// todo: move this to the new library: github.com/faiface/beep/mp3
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/oliveira-a/gochip/chip8"
)

var (
	c8       *chip8.VM
	square   *ebiten.Image
	game     *Game
	beepChan chan int

	//go:embed static/roms/*.ch8
	roms embed.FS

	//go:embed static/beep.mp3
	beepMp3 []byte

	backgroundColor color.Color = color.Black
	tileColor       color.Color = color.White
)

type Game struct {
	ui *ebitenui.UI
}

func (g *Game) Update() error {
	c8.Cycle()

	c8.Keys[0x1] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key1)))
	c8.Keys[0x2] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key2)))
	c8.Keys[0x3] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key3)))
	c8.Keys[0xc] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key4)))

	c8.Keys[0x4] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyQ)))
	c8.Keys[0x5] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyW)))
	c8.Keys[0x6] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyE)))
	c8.Keys[0xd] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyR)))

	c8.Keys[0x7] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyA)))
	c8.Keys[0x8] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyS)))
	c8.Keys[0x9] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyD)))
	c8.Keys[0xe] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyF)))

	c8.Keys[0xa] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyZ)))
	c8.Keys[0x0] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyX)))
	c8.Keys[0xb] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyC)))
	c8.Keys[0xd] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyV)))

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %f", ebiten.ActualFPS()), 0, 0)

	screen.Fill(backgroundColor)

	for x := 0; x < chip8.Cols; x++ {
		for y := 0; y < chip8.Rows; y++ {
			if c8.Vram[x][y] == 1 {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(x*20), float64(y*20))
				screen.DrawImage(square, opts)
			}
		}
	}

	g.ui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640 * 2, 320 + 100*2
}

func init() {
	square = ebiten.NewImage(20, 20)
	square.Fill(tileColor)

	beepChan = make(chan int)

	c8 = chip8.New(beepChan)

	entries, err := fs.ReadDir(roms, "static/roms")
	if err != nil {
		log.Fatal(err)
	}
	var romNames []any
	for _, entry := range entries {
		// todo: remove '.ch8' from entry name
		romNames = append(romNames, entry.Name())
	}

	game = &Game{
		ui: createUI(&uiOptions{
			RomsListEntries: romNames,
		}),
	}

	ebiten.SetWindowSize((640 * 2), (320 * 2))
	ebiten.SetMaxTPS(120)
}

func main() {
	rom, err := roms.ReadFile("static/roms/pong.ch8")
	if err != nil {
		log.Fatal(err)
	}

	if err = c8.LoadRom(rom); err != nil {
		log.Fatal(err)
	}

	go listenForAudio()

	if err = ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Used for the embedded mp3 beep audio.
// 'mp3' requires an implementation of
// `io.ReadCloser` and a `Close()` method
// is needed.
type BytesReadCloser struct {
	*bytes.Reader
}

func (b *BytesReadCloser) Close() error {
	return nil
}

func listenForAudio() {
	b := &BytesReadCloser{Reader: bytes.NewReader(beepMp3)}
	s, format, err := mp3.Decode(b)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	speaker.Init(
		format.SampleRate,
		format.SampleRate.N(time.Second/10),
	)

	for {
		<-beepChan
		speaker.Play(s)
	}
}

func btoi(b bool) uint8 {
	if b {
		return 1
	} else {
		return 0
	}
}
