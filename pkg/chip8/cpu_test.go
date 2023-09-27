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
	const ins = 0x00E0
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
	const ins = 0x00EE
	const addr = 1
	c8 := New(nil)
	c8.sp = addr

	c8.exec(ins)

	if c8.pc != addr && c8.sp != addr-1 {
		t.Fail()
	}
}

func TestSkipsNextInsIfNNIsEqualToRegisterX(t *testing.T) {
	const ins = 0x3f01
	const reg = (ins & 0x0f00) >> 8
	c8 := New(nil)
	initialPc := c8.pc
	c8.registers[reg] = (ins & 0x00ff)

	c8.exec(ins)

	if !hasSkipped(int(initialPc), int(c8.pc)) {
		t.Fail()
	}
}

func TestSkipsNextInsIfRegXNotEqualsNN(t *testing.T) {
	const ins = 0x463f
	const reg = (ins & 0x0f00) >> 8
	c8 := New(nil)
	initialPc := c8.pc
	c8.registers[reg] = 0x0012 // random value

	c8.exec(ins)

	if !hasSkipped(int(initialPc), int(c8.pc)) {
		t.Fail()
	}
}

func hasSkipped(initial int, current int) bool {
	return (current - initial) == 4
}
