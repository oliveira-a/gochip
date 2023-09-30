package chip8

import (
	"os"
	"testing"
)

var vm *VM

func setup() {
	vm = New()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func TestEmptyRomReturnsError(t *testing.T) {

	err := vm.LoadRom(make([]byte, 4097))

	if err == nil {
		t.Fail()
	}
}

func TestProgramCounterStartsAt0x200(t *testing.T) {
	if vm.pc != 0x200 {
		t.Fail()
	}
}

func TestFontIsLoadedToCorrectMemorySpace(t *testing.T) {
	for i := 0; i < len(font); i++ {
		if vm.memory[i] != font[i] {
			t.Fail()
		}
	}
}

func TestClearsDisplay(t *testing.T) {
	var ins uint16 = 0x00E0

	vm.vram[Rows/2][Rows/2] = 1

	vm.exec(ins)

	for y := 0; y < Rows; y++ {
		for x := 0; x < Cols; x++ {
			if vm.vram[x][y] != 0 {
				t.Fail()
			}
		}
	}
}

func TestReturnsFromASubroutine(t *testing.T) {
	var ins uint16 = 0x00EE
	var addr uint16 = 1

	vm.sp = addr

	vm.exec(ins)

	if vm.pc != addr && vm.sp != addr-1 {
		t.Fail()
	}
}

func TestSkipsNextInsIfNNIsEqualToRegisterX(t *testing.T) {
	var ins uint16 = 0x3f01
	var reg uint16 = (ins & 0x0f00) >> 8

	initialPc := vm.pc
	vm.registers[reg] = uint8(ins & 0x00ff)

	vm.exec(ins)

	if !hasSkipped(initialPc, vm.pc) {
		t.Fail()
	}
}

func TestSkipsNextInsIfRegXNotEqualsNN(t *testing.T) {
	var ins uint16 = 0x463f
	var reg uint16 = (ins & 0x0f00) >> 8

	initialPc := vm.pc
	vm.registers[reg] = 0x0012 // random value

	vm.exec(ins)

	if !hasSkipped(initialPc, vm.pc) {
		t.Fail()
	}
}

func TestSkipsNextInsIfRegXEqualsRegY(t *testing.T) {
	var ins uint16 = 0x5630
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4

	vm.registers[rX] = 0x1
	vm.registers[rY] = 0x1
	initialPc := vm.pc

	vm.exec(ins)

	if !hasSkipped(initialPc, vm.pc) {
		t.Fail()
	}
}

func TestValueNNIsSetToRegisterX(t *testing.T) {
	var ins uint16 = 0x6a02

	vm.exec(ins)

	if uint16(vm.registers[registerX(ins)]) != nn(ins) {
		t.Fail()
	}
}

func TestAddsNNValueToRegisterX(t *testing.T) {
	var ins uint16 = 0x7b02
	var rX uint16 = (ins & 0x0f00) >> 8

	vm.registers[rX] = 1
	expected := uint8(1 + (ins & 0x00ff))

	vm.exec(ins)

	if vm.registers[rX] != expected {
		t.Fail()
	}
}

func TestStoresValueOfRegXInRegY(t *testing.T) {
	var ins uint16 = 0x8070
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4

	vm.registers[rY] = 1

	vm.exec(ins)

	if vm.registers[rX] != vm.registers[rY] {
		t.Fail()
	}
}

func TestStoresBitwiseOrOnRegXAndRegYAndStoresInRegX(t *testing.T) {
	var ins uint16 = 0x8121
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4

	vm.registers[rX] = 0x0
	vm.registers[rY] = 0xf

	vm.exec(ins)

	if vm.registers[rX] != 0xf {
		t.Fail()
	}
}

func TestStoresBitwiseAndOnRegXAndRegYAndStoresInRegX(t *testing.T) {
	var ins uint16 = 0x8122
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4

	vm.registers[rX] = 0x0
	vm.registers[rY] = 0xf

	vm.exec(ins)

	if vm.registers[rX] == 0xf {
		t.Fail()
	}
}

func TestStoresBitwiseXorOnRegXAndRegYAndStoresInRegX(t *testing.T) {
	var ins uint16 = 0x8123
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4

	vm.registers[rX] = 0x0
	vm.registers[rY] = 0x0

	vm.exec(ins)

	if vm.registers[rX] == 0xf {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfAdditionResultIsGreaterThan8Bits(t *testing.T) {
	var ins uint16 = 0x8124
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4

	vm.registers[rX] = 250
	vm.registers[rY] = 10

	vm.exec(ins)

	if vm.registers[0xf] == 0 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfRegXGreaterThanRegY(t *testing.T) {
	var ins uint16 = 0x8125
	rX, rY := registersXAndYFromIns(ins)

	vm.registers[rX] = 3
	vm.registers[rY] = 2

	vm.exec(ins)

	if vm.registers[0xf] == 0 {
		t.Fail()
	}
}

func TestRegYIsSubstractedFromRegX(t *testing.T) {
	var ins uint16 = 0x8125
	rX, rY := registersXAndYFromIns(ins)

	vm.registers[rX] = 3
	vm.registers[rY] = 2

	vm.exec(ins)

	if vm.registers[rX] != 1 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfLeastSignifcantBitIs1(t *testing.T) {
	var ins uint16 = 0x80b6
	rX, _ := registersXAndYFromIns(ins)

	vm.registers[rX] = 0x0f

	vm.exec(ins)

	if vm.registers[0xf] != 1 {
		t.Fail()
	}

	vm.registers[rX] = 0x0

	vm.exec(ins)

	if vm.registers[0xf] != 0 {
		t.Fail()
	}

}

func TestRegXGetsDividedBy2(t *testing.T) {
	var ins uint16 = 0x80b6
	var val uint8 = 6
	rX, _ := registersXAndYFromIns(ins)

	vm.registers[rX] = val

	vm.exec(ins)

	if vm.registers[rX] != 3 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfRegYGreaterThatRegX(t *testing.T) {
	var ins uint16 = 0x8127
	rX, rY := registersXAndYFromIns(ins)

	vm.registers[rX] = 2
	vm.registers[rY] = 3

	vm.exec(ins)

	if vm.registers[0xf] == 0 {
		t.Fail()
	}
}

func TestRegXIsSubtractedFromRegYAndStoredInRegX(t *testing.T) {
	var ins uint16 = 0x8127
	rX, rY := registersXAndYFromIns(ins)

	vm.registers[rX] = 2
	vm.registers[rY] = 3

	vm.exec(ins)

	if vm.registers[rX] != 1 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfMostSignificantBitOfRegXIs1(t *testing.T) {
	var ins uint16 = 0x812e
	rX, _ := registersXAndYFromIns(ins)

	vm.registers[rX] = 128 // 10000000

	vm.exec(ins)

	if vm.registers[0xf] != 1 {
		t.Fail()
	}
}

func TestRegXGetsMultipliedByTwo(t *testing.T) {
	var ins uint16 = 0x812e
	rX, _ := registersXAndYFromIns(ins)

	vm.registers[rX] = 2

	vm.exec(ins)

	if vm.registers[rX] != 4 {
		t.Fail()
	}
}

func TestSkipsNextInsIfRegXAndRegYAreEqual(t *testing.T) {
	var ins uint16 = 0x9120
	rX, rY := registersXAndYFromIns(ins)

	initialPc := vm.pc
	vm.registers[rX] = 1
	vm.registers[rY] = 0

	vm.exec(ins)

	if !hasSkipped(initialPc, vm.pc) {
		t.Fail()
	}
}

func TestSetsRegIToNNN(t *testing.T) {
	var ins uint16 = 0xa2f0
	var nnn uint16 = ins & 0x0fff

	vm.exec(ins)

	if vm.ir != nnn {
		t.Fail()
	}
}

func TestJumpsToLocationNNNAndAddsRegister0(t *testing.T) {
	var ins uint16 = 0xb2f0
	var nnn uint16 = ins & 0x0fff

	vm.registers[0] = 1

	vm.exec(ins)

	if vm.pc != (uint16(vm.registers[0]) + nnn) {
		t.Fail()
	}
}

func TestSetsRegXToRandomgByte(t *testing.T) {
	var ins uint16 = 0xc717
	rX, _ := registersXAndYFromIns(ins)

	vm.registers[rX] = 1

	vm.exec(ins)

	if vm.registers[rX] == 1 {
		t.Fail()
	}
}

func registersXAndYFromIns(ins uint16) (uint16, uint16) {
	return ((ins & 0x0f00) >> 8), ((ins & 0x00f0) >> 4)
}

func hasSkipped(initial, current uint16) bool {
	return (current - initial) == 4
}
