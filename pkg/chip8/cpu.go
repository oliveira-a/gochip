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
	vY := registerY(ins)

	_ = n(ins)
	nn := nn(ins)
	nnn := nnn(ins)

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
		// Jump to to the location defined as nnn
		c.pc = nnn
		break
	case 0x2000:
		// Call a subroutine.
		c.sp += 1
		c.stack[c.sp] = c.pc
		c.pc = nnn
		break
	case 0x3000:
		if uint16(c.registers[vX]) == nn {
			c.pc += 2
		}
		break
	case 0x4000:
		if uint16(c.registers[vX]) != nn {
			c.pc += 2
		}
		break
	case 0x5000:
		if uint16(c.registers[vX]) == uint16(c.registers[vY]) {
			c.pc += 2
		}
		break
	case 0x6000:
		c.registers[vX] = uint8(nn)
		break
	case 0x7000:
		c.registers[vX] += uint8(nn)
		break
	case 0x8000:
		switch ins & 0x000f {
		case 0x0:
			c.registers[vX] = c.registers[vY]
			break
		case 0x1:
			c.registers[vX] |= c.registers[vY]
			break
		case 0x2:
			c.registers[vX] &= c.registers[vY]
			break
		case 0x3:
			c.registers[vX] ^= c.registers[vY]
			break
		case 0x4:
			var r uint16 = uint16(c.registers[vX]) + uint16(c.registers[vY])
			if r > 0xff {
				c.registers[0xf] = 1
			} else {
				c.registers[0xf] = 0
			}
			c.registers[vX] = uint8(r & 0x00ff)
			break
		case 0x5:
			if c.registers[vX] > c.registers[vY] {
				c.registers[0xf] = 1
			} else {
				c.registers[0xf] = 0
			}
			c.registers[vX] -= c.registers[vY]
			break
		case 0x6:
			c.registers[0xf] = c.registers[vX] & 1
			c.registers[vX] /= 2
			break
		case 0x7:
			if c.registers[vY] > c.registers[vX] {
				c.registers[0xf] = 1
			} else {
				c.registers[0xf] = 0
			}
			c.registers[vX] = c.registers[vY] - c.registers[vX]
			break
		case 0xe:
			c.registers[0xf] = c.registers[vX] >> 7
			c.registers[vX] *= 2
			break
		}
	case 0x9000:
		if c.registers[vX] != c.registers[vY] {
			c.pc += 2
		}
		break
	case 0xa000:
		c.ir = nnn
		break
	case 0xb000:
		c.pc = uint16(c.registers[0]) + nnn
		c.pc -= 2
		break
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

func registerY(ins uint16) uint16 {
	return uint16((ins & 0x00F0) >> 4)
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
