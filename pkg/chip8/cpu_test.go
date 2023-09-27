package chip8

import "testing"

func TestEmptyRomReturnsError(t *testing.T) {

	c8 := New(nil)
	err := c8.LoadRom(make([]byte, 4097))

	if err == nil {
		t.Fail()
	}
}

func TestProgramCounterStartsAt0x200(t *testing.T) {
	c8 := New(nil)

	if c8.pc != 0x200 {
		t.Fail()
	}
}

func TestFontIsLoadedToCorrectMemorySpace(t *testing.T) {
	c8 := New(nil)

	for i := 0; i < len(font); i++ {
		if c8.memory[0x50+i] != font[i] {
			t.Fail()
		}
	}
}

func TestClearsDisplay(t *testing.T) {
	var ins uint16 = 0x00E0
	c8 := New(nil)
	for i := 0; i < len(c8.display); i++ {
		c8.display[i] = 1
	}

	c8.exec(ins)

	for i := 0; i < len(c8.display); i++ {
		if c8.display[i] != 0 {
			t.Fail()
		}
	}
}

func TestReturnsFromASubroutine(t *testing.T) {
	var ins uint16 = 0x00EE
	var addr uint16 = 1
	c8 := New(nil)
	c8.sp = addr

	c8.exec(ins)

	if c8.pc != addr && c8.sp != addr-1 {
		t.Fail()
	}
}

func TestSkipsNextInsIfNNIsEqualToRegisterX(t *testing.T) {
	var ins uint16 = 0x3f01
	var reg uint16 = (ins & 0x0f00) >> 8
	c8 := New(nil)
	initialPc := c8.pc
	c8.registers[reg] = uint8(ins & 0x00ff)

	c8.exec(ins)

	if !hasSkipped(initialPc, c8.pc) {
		t.Fail()
	}
}

func TestSkipsNextInsIfRegXNotEqualsNN(t *testing.T) {
	var ins uint16 = 0x463f
	var reg uint16 = (ins & 0x0f00) >> 8
	c8 := New(nil)
	initialPc := c8.pc
	c8.registers[reg] = 0x0012 // random value

	c8.exec(ins)

	if !hasSkipped(initialPc, c8.pc) {
		t.Fail()
	}
}

func TestSkipsNextInsIfRegXEqualsRegY(t *testing.T) {
	var ins uint16 = 0x5630
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4
	c8 := New(nil)
	c8.registers[rX] = 0x1
	c8.registers[rY] = 0x1
	initialPc := c8.pc

	c8.exec(ins)

	if !hasSkipped(initialPc, c8.pc) {
		t.Fail()
	}
}

func TestValueNNIsSetToRegisterX(t *testing.T) {
	var ins uint16 = 0x6a02
	c8 := New(nil)

	c8.exec(ins)

	if uint16(c8.registers[registerX(ins)]) != nn(ins) {
		t.Fail()
	}
}

func TestAddsNNValueToRegisterX(t *testing.T) {
	var ins uint16 = 0x7b02
	var rX uint16 = (ins & 0x0f00) >> 8
	c8 := New(nil)
	c8.registers[rX] = 1
	expected := uint8(1 + (ins & 0x00ff))

	c8.exec(ins)

	if c8.registers[rX] != expected {
		t.Fail()
	}
}

func TestStoresValueOfRegXInRegY(t *testing.T) {
	var ins uint16 = 0x8070
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4
	c8 := New(nil)
	c8.registers[rY] = 1

	c8.exec(ins)

	if c8.registers[rX] != c8.registers[rY] {
		t.Fail()
	}
}

func TestStoresBitwiseOrOnRegXAndRegYAndStoresInRegX(t *testing.T) {
	var ins uint16 = 0x8121
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4
	c8 := New(nil)
	c8.registers[rX] = 0x0
	c8.registers[rY] = 0xf

	c8.exec(ins)

	if c8.registers[rX] != 0xf {
		t.Fail()
	}
}

func TestStoresBitwiseAndOnRegXAndRegYAndStoresInRegX(t *testing.T) {
	var ins uint16 = 0x8122
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4
	c8 := New(nil)
	c8.registers[rX] = 0x0
	c8.registers[rY] = 0xf

	c8.exec(ins)

	if c8.registers[rX] == 0xf {
		t.Fail()
	}
}

func TestStoresBitwiseXorOnRegXAndRegYAndStoresInRegX(t *testing.T) {
	var ins uint16 = 0x8123
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4
	c8 := New(nil)
	c8.registers[rX] = 0x0
	c8.registers[rY] = 0x0

	c8.exec(ins)

	if c8.registers[rX] == 0xf {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfAdditionResultIsGreaterThan8Bits(t *testing.T) {
	var ins uint16 = 0x8124
	var rX uint16 = (ins & 0x0f00) >> 8
	var rY uint16 = (ins & 0x00f0) >> 4
	c8 := New(nil)
	c8.registers[rX] = 250
	c8.registers[rY] = 10

	c8.exec(ins)

	if c8.registers[0xf] == 0 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfRegXGreaterThanRegY(t *testing.T) {
	var ins uint16 = 0x8125
	rX, rY := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 3
	c8.registers[rY] = 2

	c8.exec(ins)

	if c8.registers[0xf] == 0 {
		t.Fail()
	}
}

func TestRegYIsSubstractedFromRegX(t *testing.T) {
	var ins uint16 = 0x8125
	rX, rY := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 3
	c8.registers[rY] = 2

	c8.exec(ins)

	if c8.registers[rX] != 1 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfLeastSignifcantBitIs1(t *testing.T) {
	var ins uint16 = 0x80b6
	rX, _ := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 0x0f

	c8.exec(ins)

	if c8.registers[0xf] != 1 {
		t.Fail()
	}

	c8.registers[rX] = 0x0

	c8.exec(ins)

	if c8.registers[0xf] != 0 {
		t.Fail()
	}

}

func TestRegXGetsDividedBy2(t *testing.T) {
	var ins uint16 = 0x80b6
	var val uint8 = 6
	rX, _ := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = val

	c8.exec(ins)

	if c8.registers[rX] != 3 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfRegYGreaterThatRegX(t *testing.T) {
	var ins uint16 = 0x8127
	rX, rY := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 2
	c8.registers[rY] = 3

	c8.exec(ins)

	if c8.registers[0xf] == 0 {
		t.Fail()
	}
}

func TestRegXIsSubtractedFromRegYAndStoredInRegX(t *testing.T) {
	var ins uint16 = 0x8127
	rX, rY := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 2
	c8.registers[rY] = 3

	c8.exec(ins)

	if c8.registers[rX] != 1 {
		t.Fail()
	}
}

func TestSetsCarryFlagTo1IfMostSignificantBitOfRegXIs1(t *testing.T) {
	var ins uint16 = 0x812e
	rX, _ := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 128 // 10000000

	c8.exec(ins)

	if c8.registers[0xf] != 1 {
		t.Fail()
	}
}

func TestRegXGetsMultipliedByTwo(t *testing.T) {
	var ins uint16 = 0x812e
	rX, _ := registersXAndYFromIns(ins)
	c8 := New(nil)
	c8.registers[rX] = 2

	c8.exec(ins)

	if c8.registers[rX] != 4 {
		t.Fail()
	}
}

func registersXAndYFromIns(ins uint16) (uint16, uint16) {
	return ((ins & 0x0f00) >> 8), ((ins & 0x00f0) >> 4)
}

func hasSkipped(initial, current uint16) bool {
	return (current - initial) == 4
}
