package main

import (
	"os"

	"github.com/oliveira-a/gochip/pkg/chip8"
)

func main() {
	// load rom
	rom, err := os.Open("./pong.rom")
	if err != nil {
		panic(err)
	}

	defer rom.Close()

	stat, err := rom.Stat()
	if err != nil {
		panic(err)
	}

	b := make([]byte, stat.Size())
	_, err = rom.Read(b)
	if err != nil {
		panic(err)
	}

	c8 := chip8.New()
	if err = c8.LoadRom(b); err != nil {
		panic(err)
	}

	c8.Run()
}
