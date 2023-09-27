package chip8

import (
	"errors"
	"fmt"
)

const (
	Cols  = 64
	Rows  = 32
	Scale = 10
	Fps   = 60
)

type CPU struct {
	drawer  Drawer
	display [Cols * Rows]uint8

	memory    [4096]uint8
	registers [16]uint8
	stack     [16]uint16
	keys      [8]uint8

	// The program counter.
	pc uint16

	// Our index register.
	ir uint16

	// The stack pointer.
	sp uint16

	// The delay timer.
	dt uint16

	// The sound timer.
	st uint16
}

func New(d Drawer) *CPU {
	cpu := &CPU{
		drawer: d,
		pc:     0x200,
	}

	// Load font into memory - it must start at 0x50
	// as ROMs will be looking for sprites starting there.
	for i := 0; i < len(font); i++ {
		cpu.memory[0x50+i] = font[i]
	}

	return cpu
}

func (c *CPU) LoadRom(b []byte) error {
	if len(b) > len(c.memory)-512 {
		return errors.New("Rom buffer has exceeded the maximum size.")
	}

	for i := 0; i < len(b); i++ {
		c.memory[c.pc+uint16(i)] = b[i]
	}

	return nil
}

func (c *CPU) Run() error {
	ins := (uint16(c.memory[c.pc]) << 8) | uint16(c.memory[c.pc+1])

	if err := c.exec(ins); err != nil {
		return err
	}

	return nil
}

func (c *CPU) exec(ins uint16) error {
	opcode := opcode(ins)

	vX := registerX(ins)

	n := n(ins)
	nn := nn(ins)
	nnn := nnn(ins)

	printHex(opcode)
	printHex(n)
	printHex(nn)
	printHex(nnn)

	switch opcode {
	case 0x0000:
		switch ins {
		case 0x00E0:
			// Clear the display.
			for i := 0; i < len(c.display); i++ {
				c.display[i] = 0
			}
			break
		case 0x00EE:
			// Return from a subroutine.
			c.pc = c.stack[c.sp]
			c.sp -= 1
			break
		}
	case 0x1000:
		// Jump to nnn location.
		c.pc = nnn
		break
	case 0x2000:
		// Call a subroutine.
		c.sp += 1
		c.stack[c.sp] = c.pc
		c.pc = nnn
	case 0x3000:
		// Skip the next instruction if value of register 'x' is the same as nn.
		if vX == nn {
			c.pc += 2
		}
		c.pc += 2

	default:
		return errors.New("Unknown instrunction encountered.")
	}

	c.pc += 2

	return nil
}

func opcode(ins uint16) uint16 {
	return ins & 0xF000
}

func registerX(ins uint16) uint16 {
	return uint16((ins & 0x0F00) >> 8)
}

func n(ins uint16) uint16 {
	return ins & 0x000F
}

func nn(ins uint16) uint16 {
	return ins & 0x00FF
}

func nnn(ins uint16) uint16 {
	return ins & 0x0FFF
}

func printHex(v uint16) {
	fmt.Printf("%s\n", fmt.Sprintf("%x", v))
}
