package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	"io"
	"io/fs"
	"log"
	"strings"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/go-mp3"
	"github.com/oliveira-a/gochip/chip8"
)

const (
	winWidth     = 640
	winHeight    = 320
	tileSize     = 10
	romListWidth = 150
)

var (
	square *ebiten.Image
	game   *Game

	//go:embed static/roms/*.ch8
	roms embed.FS

	//go:embed static/beep.mp3
	beepMp3 []byte

	beepChan chan int

	backgroundColor color.Color = color.Black
	tileColor       color.Color = color.White
)

// The single global game state structure that is created
// once and used throughout.
type Game struct {
	// The ebiten UI.
	ui *ebitenui.UI

	// The chip8 virtual machine that we load the ROM into.
	c8 *chip8.VM

	// This represents each square in the game screen.
	tile *ebiten.Image
}

func (g *Game) Update() error {
	g.c8.Cycle()

	g.c8.Keys[0x1] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key1)))
	g.c8.Keys[0x2] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key2)))
	g.c8.Keys[0x3] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key3)))
	g.c8.Keys[0xc] = uint8(btoi(ebiten.IsKeyPressed(ebiten.Key4)))

	g.c8.Keys[0x4] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyQ)))
	g.c8.Keys[0x5] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyW)))
	g.c8.Keys[0x6] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyE)))
	g.c8.Keys[0xd] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyR)))

	g.c8.Keys[0x7] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyA)))
	g.c8.Keys[0x8] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyS)))
	g.c8.Keys[0x9] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyD)))
	g.c8.Keys[0xe] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyF)))

	g.c8.Keys[0xa] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyZ)))
	g.c8.Keys[0x0] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyX)))
	g.c8.Keys[0xb] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyC)))
	g.c8.Keys[0xd] = uint8(btoi(ebiten.IsKeyPressed(ebiten.KeyV)))

	g.ui.Update()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %f", ebiten.ActualFPS()), 0, 0)

	screen.Fill(backgroundColor)
	g.tile.Fill(tileColor)

	for x := 0; x < chip8.Cols; x++ {
		for y := 0; y < chip8.Rows; y++ {
			if g.c8.Vram[x][y] == 1 {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(x*tileSize)+romListWidth, float64(y*tileSize))
				screen.DrawImage(g.tile, opts)
			}
		}
	}

	g.ui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {

	// UI setup
	//
	// Scan all of the available ROMs in the static/roms
	// directory and extract their name to create list items
	// for the romList rom selection.
	entries, err := fs.ReadDir(roms, "static/roms")
	if err != nil {
		log.Fatal(err)
	}

	var listItems []any
	for _, ent := range entries {
		// avoids reading files without the '.ch8' extension.
		if dn, ok := strings.CutSuffix(ent.Name(), ".ch8"); ok {
			li := listItem{
				name: dn,
				path: fmt.Sprintf("%s/%s", "static/roms", ent.Name()),
			}

			listItems = append(listItems, li)
		}
	}

	romList := newRomList(
		listItems,
		// Define how to handle the rom selection
		func(args *widget.ListEntrySelectedEventArgs) {
			rp := args.Entry.(listItem).path

			rom, err := roms.ReadFile(rp)
			if err != nil {
				log.Fatal(err)
			}

			if err = game.c8.LoadRom(rom); err != nil {
				log.Fatal(err)
			}
		},
		romListWidth,
		winHeight,
	)
	tickRateContextMenu := newTickRateContextMenu()

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.ContextMenu(tickRateContextMenu)),
	)
	root.AddChild(romList)

	beepChan = make(chan int)
	game = &Game{
		ui: &ebitenui.UI{Container: root},

		// todo: create a program flag for debug mode
		c8: chip8.New(beepChan, false),

		tile: ebiten.NewImage(tileSize, tileSize),
	}

	ebiten.SetWindowSize(winWidth+romListWidth, winHeight)

	// A go routine that listens for audio evenst through
	// the beep channel. Plays the sound from the 'beep.mp3'
	// sound.
	go listenForAudio()

	if err = ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func listenForAudio() {
	fileBytesReader := bytes.NewReader(beepMp3)
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		log.Printf("Error decoding mp3: %s\n", err)
		return
	}

	op := &oto.NewContextOptions{}
	op.SampleRate = 44100
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		log.Printf("Error creating new audio context: %s\n", err)
	}
	<-readyChan

	player := otoCtx.NewPlayer(decodedMp3)
	defer player.Close()

	for {
		<-beepChan
		player.Play()

		for player.IsPlaying() {
			time.Sleep(time.Millisecond)
		}

		_, err := player.Seek(0, io.SeekStart)
		if err != nil {
			panic("player.Seek failed: " + err.Error())
		}
	}
}

func btoi(b bool) uint8 {
	if b {
		return 1
	} else {
		return 0
	}
}
