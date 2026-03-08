package gonsole

import "github.com/hajimehoshi/ebiten/v2/ebitenutil"

// DrawText draws a string directly onto the screen at (x, y).
// Uses the built-in ebiten debug font (6×16px, white).
// Must be called from DrawFunc (or UpdateFunc before Draw runs).
func (c *Console) DrawText(x, y int, text string) {
	if c.screen != nil {
		ebitenutil.DebugPrintAt(c.screen, text, x, y)
	}
}
