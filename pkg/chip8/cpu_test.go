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

	if c8.programCounter != 0x200 {
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
