package gonsole

import "github.com/hajimehoshi/ebiten/v2"

const (
	ButtonUp     = 1 << 0
	ButtonDown   = 1 << 1
	ButtonLeft   = 1 << 2
	ButtonRight  = 1 << 3
	ButtonA      = 1 << 4
	ButtonB      = 1 << 5
	ButtonStart  = 1 << 6
	ButtonSelect = 1 << 7
)

func (c *Console) pollInputs() {
	var b byte
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		b |= ButtonUp
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		b |= ButtonDown
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		b |= ButtonLeft
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		b |= ButtonRight
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyJ) {
		b |= ButtonA
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeyK) {
		b |= ButtonB
	}
	c.Buttons = b
}

// IsPressed returns true if the given button is currently held.
func (c *Console) IsPressed(btn byte) bool {
	return c.Buttons&btn != 0
}
