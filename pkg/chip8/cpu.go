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

type VM struct {
	// Display memory
	vram [Cols][Rows]uint8

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

func New() *VM {
	// TODO: Perhaps move this to the constructor
	// injected. Maybe the whole logger instance?
	log.SetPrefix("CHIP-8: ")
	log.SetFlags(log.Ltime)

	cpu := &VM{
		pc: 0x200,
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

	for i := 0; i < len(b); i++ {
		c.memory[c.pc+uint16(i)] = b[i]
	}

	return nil
}

func (vm *VM) Run() error {
	var ins uint16 = uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1])

	if err := vm.exec(ins); err != nil {
		return err
	}

	return nil
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
					vm.vram[x][y] = 0
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
		logInstruction(ins, "Skip the next instrunction if vX != vY.")
		if uint16(vm.registers[vX]) != nn {
			vm.pc += 2
		} else {
			vm.pc += 4
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
		logInstruction(ins, "Skip the next instrunction if vX = vY.")
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
		// Draw the sprite starting at memory location I.
		vm.registers[0xf] = 0
		height := n
		width := 8

		for i := 0; i < int(height); i++ {
			sprite := vm.memory[vm.ir+uint16(i)]
			for bit := 0; bit < width; bit++ {
				draw := (sprite >> bit) % 2
				x, y := (vm.registers[vX]+uint8(bit))%Cols, (vm.registers[vY]+uint8(i))%Rows

				isDraw := vm.vram[x][y] ^ draw
				vm.vram[x][y] = isDraw

				// If any bit got erased, then set vF to carry.
				if isDraw == 0 {
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
			// TODO
			if vm.keys[uint16(vm.registers[vX])] == Pressed {
				vm.pc += 4
			} else {
				vm.pc += 2
			}
			break
		case 0xa1:
			logInstruction(ins, "Skip next instrunction if key with value of vX is no pressed.")
			// TODO
			vm.pc += 2
			break
		}
		break
	case 0xf000:
		switch nn {
		case 0x07:
			logInstruction(ins, "Set vX = delay timer value.")
			vm.memory[vX] = uint8(vm.dt)
			vm.pc += 2
			break
		case 0x0a:
			logInstruction(ins, "Wait for a key press. Store the value of the key in vX.")
			vm.pc += 2
			break
		case 0x15:
			logInstruction(ins, "Set the delay timer to vX.")
			vm.pc += 2
			break
		case 0x18:
			logInstruction(ins, "Set sound timer = vX.")
			vm.pc += 2
			break
		case 0x1e:
			logInstruction(ins, "Set I = I + vX.")
			vm.pc += 2
			break
		case 0x29:
			logInstruction(ins, "Set I = location of sprite for digit vX.")
			vm.pc += 2
			break
		case 0x33:
			logInstruction(ins, "Store BCD representation of vX in memory location I, I+1, and I+2")
			vm.pc += 2
			break
		case 0x55:
			logInstruction(ins, "Store registers v0 through vX in memory locations I.")
			vm.pc += 2
			break
		case 0x65:
			logInstruction(ins, "Read registers v0 through vX from memory starting at location I.")
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
	log.Printf("| Executing '%04x': %s\n", ins, msg)
}
