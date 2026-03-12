package gonsole

import "github.com/hajimehoshi/ebiten/v2/ebitenutil"

// DrawText draws a string at screen-space (x, y) using the built-in ebiten
// debug font (6×16px, white). Safe to call from PaintFunc at any slot.
// Text is drawn onto screenImg after scratch is uploaded to the GPU, so it
// always appears on top regardless of which slot PaintFunc calls it from.
func (c *Console) DrawText(x, y int, text string) {
	if c.screenImg != nil {
		ebitenutil.DebugPrintAt(c.screenImg, text, x, y)
	}
}
