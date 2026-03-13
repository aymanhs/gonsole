package main

import (
	"aymanhs/gonsole"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PaletteEditor struct {
	selectedBank  int
	selectedIndex int
}

func (e *PaletteEditor) Update(c *gonsole.Console, colorClipboard *[4]byte, bankClipboard *[16][4]byte) {
	mx, my := c.MousePos()

	// Mouse Selection: 4 banks in one column
	for b := 0; b < 4; b++ {
		bx := 20
		by := 50 + b*80
		for i := 0; i < 16; i++ {
			x := bx + (i%8)*24
			y := by + (i/8)*24
			if c.JustPressedMouse(gonsole.MouseButtonLeft) && mx >= x && mx < x+20 && my >= y && my < y+20 {
				e.selectedBank = b
				e.selectedIndex = i
			}
		}
	}

	// Clipboard (Ctrl+C / Ctrl+V)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyC) {
		if shift {
			// Copy entire bank
			*bankClipboard = c.PaletteBank.Colors[e.selectedBank]
		} else {
			// Copy single color
			*colorClipboard = c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
		}
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyV) {
		if shift {
			// Paste entire bank (skip target bank 0 since it's locked/system palette)
			if e.selectedBank > 0 {
				for i := 1; i < 16; i++ { // Skip index 0 (transparent)
					rgba := (*bankClipboard)[i]
					r2, g2, b2 := rgba[0]/85, rgba[1]/85, rgba[2]/85
					c.SetBankPalette(e.selectedBank, i, r2, g2, b2)
				}
			}
		} else {
			// Only paste single color if target is not Bank 0 and not Index 0
			if e.selectedBank > 0 && e.selectedIndex > 0 {
				rgba := *colorClipboard
				r2, g2, b2 := rgba[0]/85, rgba[1]/85, rgba[2]/85
				c.SetBankPalette(e.selectedBank, e.selectedIndex, r2, g2, b2)
			}
		}
	}

	// Adjustment Buttons: To the right (x=300 instead of 180 to give more breathing room)
	if e.selectedBank > 0 || (e.selectedBank == 0 && e.selectedIndex > 0) { // Allow editing non-zero index in Bank 0
		bx := 300
		by := 50
		if c.JustPressedMouse(gonsole.MouseButtonLeft) {
			if e.selectedBank > 0 {
				for i := 0; i < 3; i++ {
					yy := by + 80 + i*35
					// Minus [-]
					if mx >= bx+40 && mx < bx+60 && my >= yy && my < yy+20 {
						e.adjust(c, i, -1)
					}
					// Plus [+]
					if mx >= bx+100 && mx < bx+120 && my >= yy && my < yy+20 {
						e.adjust(c, i, 1)
					}
				}
			}

			// Blend Mod Toggle (Click on flags text box)
			if mx >= bx-5 && mx < bx+155 && my >= by+35 && my < by+55 {
				e.toggleBlend(c)
			}
		}
	}

	// Keyboard Shortcuts
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		e.selectedBank--
		if e.selectedBank < 0 {
			e.selectedBank = 3
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		e.selectedBank = (e.selectedBank + 1) % 4
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		e.selectedIndex--
		if e.selectedIndex < 0 {
			e.selectedIndex = 15
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		e.selectedIndex = (e.selectedIndex + 1) % 16
	}

	canEditBlend := e.selectedBank > 0 || (e.selectedBank == 0 && e.selectedIndex > 0)
	if canEditBlend && inpututil.IsKeyJustPressed(ebiten.KeyF) && !ctrl {
		e.toggleBlend(c)
	}

	if e.selectedBank > 0 && !ctrl {
		delta := 1
		if shift {
			delta = -1
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			e.adjust(c, 0, delta)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyG) {
			e.adjust(c, 1, delta)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			e.adjust(c, 2, delta)
		}
	}
}

func (e *PaletteEditor) toggleBlend(c *gonsole.Console) {
	rgba := c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
	blendMode := rgba[3]
	blendMode = (blendMode + 1) % 4
	c.PaletteBank.Colors[e.selectedBank][e.selectedIndex][3] = blendMode
}

func (e *PaletteEditor) adjust(c *gonsole.Console, channel int, delta int) {
	rgba := c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
	vals := [3]byte{rgba[0] / 85, rgba[1] / 85, rgba[2] / 85}

	val := int(vals[channel]) + delta
	if val < 0 {
		val = 3
	}
	if val > 3 {
		val = 0
	}
	vals[channel] = byte(val)

	c.SetBankPalette(e.selectedBank, e.selectedIndex, vals[0], vals[1], vals[2])
}

func (e *PaletteEditor) Draw(c *gonsole.Console) {
	// Draw palettes in a column on the left
	for b := 0; b < 4; b++ {
		bx := 20
		by := 50 + b*80

		label := fmt.Sprintf("B%d", b)
		if b == 0 {
			label += " (L)"
		}
		c.DrawText(bx, by-20, label)

		// Surround entire bank grid with grey
		// c.DrawRect(bx-1, by-1, 8*24+1, 2*24+1, 5, false)

		for i := 0; i < 16; i++ {
			x := bx + (i%8)*24
			y := by + (i/8)*24

			// Draw color square
			e.drawAbsoluteRect(c, x, y, 20, 20, c.PaletteBank.Colors[b][i])

			// Individual grey outline for each color square (helps see black)
			c.DrawRect(x, y, 20, 20, 5, false)

			// Draw blend flag indicator
			drawBlendIndicator(c, x, y, 20, 20, c.PaletteBank.Colors[b][i][3])

			if e.selectedBank == b && e.selectedIndex == i {
				// Selection highlight (thick white border)
				c.DrawRect(x-2, y-2, 24, 24, 7, false)
			}
		}
	}

	// Draw controls to the right
	bx := 300
	by := 50
	rgba := c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
	r2, g2, b2 := rgba[0]/85, rgba[1]/85, rgba[2]/85

	c.DrawText(bx, by, fmt.Sprintf("Selected: Bank %d Index %d", e.selectedBank, e.selectedIndex))

	// Values are always shown, but editing/buttons are conditional
	c.DrawText(bx, by+20, fmt.Sprintf("Current RGB: %d, %d, %d", r2, g2, b2))

	blendStrs := []string{"00: NORMAL", "01: SUBTRACT", "10: ADD", "11: TRANSPARENT"}
	bMode := rgba[3]
	if bMode > 3 {
		bMode = 0
	}

	if e.selectedBank == 0 && e.selectedIndex == 0 {
		c.DrawText(bx, by+40, "Flags: LOCKED")
		msg := "LOCKED (Fixed Black)"
		c.DrawText(bx, by+60, msg)
	} else {
		c.DrawRect(bx-5, by+35, 160, 20, 5, false) // Box around flag toggle
		c.DrawText(bx, by+40, fmt.Sprintf("Flags: %s", blendStrs[bMode]))

		if e.selectedBank == 0 {
			c.DrawText(bx, by+80, "RGB LOCKED (System Palette)")
		} else {
			channels := []string{"R", "G", "B"}
			vals := []byte{byte(r2), byte(g2), byte(b2)}
			for i, name := range channels {
				yy := by + 80 + i*35
				c.DrawText(bx, yy+5, name+":")

				// Minus button
				c.DrawRect(bx+40, yy, 20, 20, 5, true)
				c.DrawText(bx+45, yy+3, "-")

				// Value display
				c.DrawText(bx+75, yy+5, fmt.Sprintf("%d", vals[i]))

				// Plus button
				c.DrawRect(bx+100, yy, 20, 20, 5, true)
				c.DrawText(bx+105, yy+3, "+")
			}
		}

		c.DrawText(bx, by+190, "CTRL+C/V Copy/Paste Color")
		c.DrawText(bx, by+210, "CTRL+SHIFT+C/V Copy Bank")
	}
}

func (e *PaletteEditor) drawAbsoluteRect(c *gonsole.Console, x, y, w, h int, rgba [4]byte) {
	for dy := 0; dy < h; dy++ {
		yy := y + dy
		if yy < 0 || yy >= gonsole.ScreenHeight {
			continue
		}
		for dx := 0; dx < w; dx++ {
			xx := x + dx
			if xx < 0 || xx >= gonsole.ScreenWidth {
				continue
			}
			dst := (yy*gonsole.ScreenWidth + xx) * 4
			c.Scratch[dst] = rgba[0]
			c.Scratch[dst+1] = rgba[1]
			c.Scratch[dst+2] = rgba[2]
			c.Scratch[dst+3] = 255
		}
	}
}

func (e *PaletteEditor) DrawOverlay(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "PALETTE EDITOR", 10, 2)
}
