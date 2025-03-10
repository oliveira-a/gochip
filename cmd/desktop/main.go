package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"os"
	"path"
	"runtime"
	"time"
	"embed"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

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

	backgroundColor color.Color = color.Black
	tileColor color.Color = color.White
)

type Game struct {
	ui *ebitenui.UI
}

func init() {
	square = ebiten.NewImage(20, 20)
	square.Fill(tileColor)

	beepChan = make(chan int)

	c8 = chip8.New(beepChan)

	res, err := loadResources()
	if err != nil {
		panic(err)
	}

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	ui := ebitenui.UI{
		Container: root,
	}

	game = &Game{
		ui: &ui,
	}

	// setup the toolbar
	toolbar := newToolbar(&ui, res)
	root.AddChild(toolbar.container)

	ebiten.SetWindowSize((640 * 2), (320 * 2))
	ebiten.SetMaxTPS(120)
}

type resources struct {
	font text.Face
}

func loadResources() (*resources, error) {
	fnt, err := loadFont(16)
	if err != nil {
		return nil, err
	}
	return &resources{
		font: fnt,
	}, nil
}

func loadFont(size float64) (text.Face, error) {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}, nil

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

func listenForAudio() {
	_, p, _, _ := runtime.Caller(0)
	p = path.Dir(p)

	f, err := os.Open(fmt.Sprintf("%s/beep.mp3", p))
	if err != nil {
		panic(err)
	}

	s, format, err := mp3.Decode(f)
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
	return outsideWidth, outsideHeight
}
