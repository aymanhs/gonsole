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

func TestSpritePacking(t *testing.T) {
	c := NewConsole()
	data := make([]byte, 64)
	for i := range data {
		data[i] = byte(i % 16)
	}
	
	c.SetSpriteData(0, data)
	
	// Check packed data
	for i := 0; i < 32; i++ {
		b := c.SpriteData[0][i]
		hi := (b >> 4) & 0xF
		lo := b & 0xF
		if hi != byte((i*2)%16) {
			t.Errorf("Byte %d high nibble mismatch: got %d, want %d", i, hi, (i*2)%16)
		}
		if lo != byte((i*2+1)%16) {
			t.Errorf("Byte %d low nibble mismatch: got %d, want %d", i, lo, (i*2+1)%16)
		}
	}
}
