package chip8

import (
	"errors"
)

type CPU struct {
	screen Drawer

	memory    [4096]uint8
	registers [16]uint8
	stack     [16]uint16
	keys      [8]uint8

	programCounter uint16
	indexRegister  uint16
	stackPointer   uint16
	delayTimer     uint16
	soundTimer     uint16
}

func New(d Drawer) *CPU {
	return &CPU{
		screen: d,
	}
}

func (c *CPU) LoadRom(b []byte) error {
	if len(b) > len(c.memory)-512 {
		return errors.New("Rom buffer has exceeded the maximum size.")
	}

	for i := 0; i < len(b); i++ {
		// Program memory space starts at addr 0x200 (512 in base10)
		c.memory[i+512] = b[i]
	}

	return nil
}

func (c *CPU) Run() error {
	return nil
}
