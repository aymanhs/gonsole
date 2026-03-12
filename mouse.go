package gonsole

import "github.com/hajimehoshi/ebiten/v2"

const (
	MouseButtonLeft   byte = 1 << 0
	MouseButtonRight  byte = 1 << 1
	MouseButtonMiddle byte = 1 << 2
)

func (c *Console) pollMouse() {
	c.prevMouseButtons = c.MouseButtons
	x, y := ebiten.CursorPosition()
	c.MouseX = x
	c.MouseY = y

	var b byte
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		b |= MouseButtonLeft
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		b |= MouseButtonRight
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		b |= MouseButtonMiddle
	}
	c.MouseButtons = b
}

// MousePos returns the current cursor position in screen coordinates.
func (c *Console) MousePos() (x, y int) {
	return c.MouseX, c.MouseY
}

// MousePressed returns true if the given mouse button is currently held.
func (c *Console) MousePressed(btn byte) bool {
	return c.MouseButtons&btn != 0
}

// JustPressedMouse returns true only on the frame the mouse button was first pressed.
func (c *Console) JustPressedMouse(btn byte) bool {
	return (c.MouseButtons&btn != 0) && (c.prevMouseButtons&btn == 0)
}
