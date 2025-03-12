package chip8

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	Cols  = 64
	Rows  = 32
	Scale = 10
	Fps   = 60
)

var debug bool

type VM struct {
	Vram [Cols][Rows]uint8

	memory    [4096]uint8
	registers [16]uint8
	stack     [16]uint16

	Keys [16]uint8

	// To be provided by the client.
	// The vm will use this channel to notify when to beep.
	audio chan int

	// The program counter.
	pc uint16

	// Our index register.
	ir uint16

	// The stack pointer.
	sp uint8

	// The delay timer.
	dt uint8

	// The sound timer.
	st uint8
}

func init() {
	log.SetPrefix("CHIP-8: ")
	log.SetFlags(log.Ltime)
}

func New(audio chan int, debugMode bool) *VM {
	debug = debugMode

	cpu := &VM{
		pc:    0x200,
		audio: audio,
	}

	for i := 0; i < len(font); i++ {
		cpu.memory[i] = font[i]
	}

	return cpu
}

func (c *VM) LoadRom(b []byte) error {
	if len(b) > len(c.memory)-512 {
		return errors.New("Rom buffer has exceeded the maximum size.")
	}

	c.pc = 0x200
	c.ir = 0
	c.sp = 0
	c.dt = 0
	c.st = 0

	// ensure memory is cleared
	for i := c.pc; i < uint16(len(c.memory)); i++ {
		c.memory[uint16(i)] = 0
	}

	// ensure vram is cleared
	for y := 0; y < Rows; y++ {
		for x := 0; x < Cols; x++ {
			c.Vram[x][y] = 0
		}
	}

	// load the buffer into memory
	for i := 0; i < len(b); i++ {
		c.memory[c.pc+uint16(i)] = b[i]
	}

	return nil
}

func (vm *VM) Cycle() {
	vm.exec(vm.fetchInstruction())

	if vm.dt > 0 {
		vm.dt--
	}

	if vm.st > 0 {
		vm.audio <- 1
		vm.st--
	}
}

func (vm *VM) fetchInstruction() uint16 {
	return uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1])
}

func (vm *VM) exec(ins uint16) error {
	opcode := opcode(ins)

	vX := registerX(ins)
	vY := registerY(ins)

	n := n(ins)
	nn := nn(ins)
	nnn := nnn(ins)

	switch opcode {
	case 0x0000:
		switch ins {
		case 0x00E0:
			logInstruction(ins, "Clear the display.")
			for y := 0; y < Rows; y++ {
				for x := 0; x < Cols; x++ {
					vm.Vram[x][y] = 0
				}
			}
			vm.pc += 2
			break
		case 0x00EE:
			logInstruction(ins, "Return from a subroutine.")
			vm.pc = vm.stack[vm.sp] + 2
			vm.sp--
			break
		}
	case 0x1000:
		logInstruction(ins, "Jump to the location.")
		vm.pc = nnn
		break
	case 0x2000:
		logInstruction(ins, "Call a subroutine.")
		vm.sp += 1
		vm.stack[vm.sp] = vm.pc
		vm.pc = nnn
		break
	case 0x3000:
		logInstruction(ins, "Skip the next instruction if vX = nn.")
		if uint16(vm.registers[vX]) == nn {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
		break
	case 0x4000:
		logInstruction(ins, "Skip the next instrunction if vX != nn.")
		if uint16(vm.registers[vX]) != nn {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
		break
	case 0x5000:
		logInstruction(ins, "Skip the next instrunction if vX != vY.")
		if uint16(vm.registers[vX]) == uint16(vm.registers[vY]) {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
		break
	case 0x6000:
		logInstruction(ins, "Load value nn into vX.")
		vm.registers[vX] = uint8(nn)
		vm.pc += 2
		break
	case 0x7000:
		logInstruction(ins, "Set vX = vX + nn.")
		vm.registers[vX] += uint8(nn)
		vm.pc += 2
		break
	case 0x8000:
		switch n {
		case 0x0:
			logInstruction(ins, "Set vX = vY.")
			vm.registers[vX] = vm.registers[vY]
			vm.pc += 2
			break
		case 0x1:
			logInstruction(ins, "Set vX |= vY.")
			vm.registers[vX] |= vm.registers[vY]
			vm.pc += 2
			break
		case 0x2:
			logInstruction(ins, "Set vX &= vY.")
			vm.registers[vX] &= vm.registers[vY]
			vm.pc += 2
			break
		case 0x3:
			logInstruction(ins, "Set vX ^= vY.")
			vm.registers[vX] ^= vm.registers[vY]
			vm.pc += 2
			break
		case 0x4:
			logInstruction(ins, "Set vX = vX + vY, set VF = carry.")
			var r uint16 = uint16(vm.registers[vX]) + uint16(vm.registers[vY])
			if r > 0xff {
				vm.registers[0xf] = 1
			} else {
				vm.registers[0xf] = 0
			}
			vm.registers[vX] = uint8(r & 0x00ff)
			vm.pc += 2
			break
		case 0x5:
			logInstruction(ins, "Set vX = vX - vY, set VF = NOT borrow.")
			if vm.registers[vX] > vm.registers[vY] {
				vm.registers[0xf] = 1
			} else {
				vm.registers[0xf] = 0
			}
			vm.registers[vX] -= vm.registers[vY]
			vm.pc += 2
			break
		case 0x6:
			logInstruction(ins, "Set vX = vX SHR 1.")
			vm.registers[0xf] = vm.registers[vX] & 1
			vm.registers[vX] /= 2
			vm.pc += 2
			break
		case 0x7:
			logInstruction(ins, "Set vX = vX - vY, set VF = NOT borrow.")
			if vm.registers[vY] > vm.registers[vX] {
				vm.registers[0xf] = 1
			} else {
				vm.registers[0xf] = 0
			}
			vm.registers[vX] = vm.registers[vY] - vm.registers[vX]
			vm.pc += 2
			break
		case 0xe:
			logInstruction(ins, "Set vX = vX SHL 1.")
			vm.registers[0xf] = vm.registers[vX] >> 7
			vm.registers[vX] *= 2
			vm.pc += 2
			break
		}
	case 0x9000:
		logInstruction(ins, "Skip next instrunction if vX != vY.")
		if vm.registers[vX] != vm.registers[vY] {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
		break
	case 0xa000:
		logInstruction(ins, "Set vI to nnn.")
		vm.ir = nnn
		vm.pc += 2
		break
	case 0xb000:
		logInstruction(ins, "Jump to location nnn + v0.")
		vm.pc = uint16(vm.registers[0]) + nnn
		break
	case 0xc000:
		logInstruction(ins, "Set vX = random byte AND nn.")
		for {
			s := rand.NewSource(time.Now().UnixMilli())
			r := rand.New(s)
			num := uint16(r.Intn(255))

			val := uint8(num & nn)
			if vm.registers[vX] != val {
				vm.registers[vX] = val
				break
			}
		}
		vm.pc += 2
		break
	case 0xd000:
		logInstruction(ins, "Draw.")

		vm.registers[0xf] = 0
		height := n

		for i := 0; i < int(height); i++ {
			sprite := vm.memory[vm.ir+uint16(i)]

			for bit := 0; bit < 8; bit++ {
				draw := (sprite >> (8 - (bit + 1))) % 2
				x, y := (vm.registers[vX]+uint8(bit))%Cols, (vm.registers[vY]+uint8(i))%Rows

				vm.Vram[x][y] ^= draw

				// If any bit got erased, then set vF to carry.
				if vm.Vram[x][y] == 0 {
					vm.registers[0xf] = 1
				}
			}
		}
		vm.pc += 2
		break
	case 0xe000:
		switch nn {
		case 0x9e:
			logInstruction(ins, "Skip next instrunction if key with value of vX is pressed.")
			if vm.Keys[vm.registers[vX]] == 1 {
				vm.pc += 4
			} else {
				vm.pc += 2
			}
			break
		case 0xa1:
			logInstruction(ins, "Skip next instrunction if key with value of vX is not pressed.")
			if vm.Keys[vm.registers[vX]] == 0 {
				vm.pc += 4
			} else {
				vm.pc += 2
			}
			break
		}
		break
	case 0xf000:
		switch nn {
		case 0x7:
			logInstruction(ins, "Set vX = delay timer value.")
			vm.registers[vX] = uint8(vm.dt)
			vm.pc += 2
			break
		case 0xa:
			logInstruction(ins, "Wait for a key press. Store the value of the key in vX.")
			for i, k := range vm.Keys {
				if k == 1 {
					vm.registers[vX] = uint8(i)
					vm.Keys[i] = 0
					vm.pc += 2
					break
				}
			}
			break
		case 0x15:
			logInstruction(ins, "Set the delay timer to vX.")
			vm.dt = vm.registers[vX]
			vm.pc += 2
			break
		case 0x18:
			logInstruction(ins, "Set sound timer = vX.")
			vm.st = vm.registers[vX]
			vm.pc += 2
			break
		case 0x1e:
			logInstruction(ins, "Set I = I + vX.")
			vm.ir += uint16(vm.registers[vX])
			if vm.ir > 0xfff {
				vm.registers[0xf] = 1
			}
			vm.pc += 2
			break
		case 0x29:
			// Find the character in the font map
			// Set the ir to point to the right
			// address memory which corresponds to
			// the character in register x. Each
			// character is at 5 apart.
			logInstruction(ins, "Set I = location of sprite for digit vX.")
			p := 0
			for i := 0; i < int(vm.registers[vX]); i++ {
				p += 5
			}
			vm.ir = uint16(p)
			vm.pc += 2
			break
		case 0x33:
			logInstruction(ins, "Store BCD representation of vX in memory location I, I+1, and I+2")
			// 128
			v := vm.registers[vX]

			//         1                  2                 8
			b, c, d := uint8((v/100)%10), uint8((v/10)%10), uint8((v/1)%10)

			vm.memory[vm.ir] = b
			vm.memory[vm.ir+1] = c
			vm.memory[vm.ir+2] = d

			vm.pc += 2
			break
		case 0x55:
			logInstruction(ins, "Store registers v0 through vX in memory locations I.")
			for r := 0; r <= int(vX); r++ {
				vm.memory[vm.ir+uint16(r)] = vm.registers[r]
			}
			vm.pc += 2
			break
		case 0x65:
			logInstruction(ins, "Read registers v0 through vX from memory starting at location I.")
			for i := 0; i <= int(vX); i++ {
				vm.registers[i] = vm.memory[vm.ir+uint16(i)]
			}
			vm.pc += 2
			break
		}
	default:
		return errors.New(fmt.Sprintf("Unsupported instruction: %04x", ins))
	}

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

func logInstruction(ins uint16, msg string) {
	if !debug {
		return
	}

	log.Printf("| Executing '%04x': %s\n", ins, msg)
}
