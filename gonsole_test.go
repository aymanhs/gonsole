package gonsole

import (
	"testing"
)

func TestInputEdgeDetection(t *testing.T) {
	c := NewConsole()
	
	// Simulate press
	c.Buttons = ButtonA
	if !c.IsPressed(ButtonA) {
		t.Error("IsPressed(ButtonA) should be true")
	}
	if !c.JustPressed(ButtonA) {
		t.Error("JustPressed(ButtonA) should be true on first frame")
	}
	
	// Simulate next frame holding
	c.pollInputs() // c.prevButtons becomes ButtonA
	c.Buttons = ButtonA
	if !c.IsPressed(ButtonA) {
		t.Error("IsPressed(ButtonA) should be true while held")
	}
	if c.JustPressed(ButtonA) {
		t.Error("JustPressed(ButtonA) should be false while held")
	}
	
	// Simulate release
	c.pollInputs() // c.prevButtons becomes ButtonA
	c.Buttons = 0
	if c.IsPressed(ButtonA) {
		t.Error("IsPressed(ButtonA) should be false after release")
	}
	if c.JustPressed(ButtonA) {
		t.Error("JustPressed(ButtonA) should be false after release")
	}
}

func TestSpriteData(t *testing.T) {
	c := NewConsole()
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i % 16)
	}
	
	c.SetSpriteData(0, data)
	
	// Check storage (16x16 = 256 pixels = 256 bytes)
	for i := 0; i < 256; i++ {
		if c.SpriteData[0][i] != byte(i%16) {
			t.Errorf("Byte %d mismatch: got %d, want %d", i, c.SpriteData[0][i], i%16)
		}
	}
}
