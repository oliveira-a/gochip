package chip8

import "testing"

func TestEmptyRomReturnsError(t *testing.T) {

	c8 := New(nil)
	err := c8.LoadRom(make([]byte, 4097))

	if err == nil {
		t.Fail()
	}
}
