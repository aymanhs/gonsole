package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"aymanhs/gonsole"
)

type PaletteEditor struct {
	selectedBank int
	selectedIndex int
}

func (e *PaletteEditor) Update(c *gonsole.Console, colorClipboard *[4]byte) {
	mx, my := c.MousePos()

	// Mouse Selection: 4 banks in one column
	for b := 0; b < 4; b++ {
		bx := 10
		by := 30 + b*50
		for i := 0; i < 16; i++ {
			x := bx + (i%8)*20
			y := by + (i/8)*20
			if c.JustPressedMouse(gonsole.MouseButtonLeft) && mx >= x && mx < x+16 && my >= y && my < y+16 {
				e.selectedBank = b
				e.selectedIndex = i
			}
		}
	}

	// Clipboard (Ctrl+C / Ctrl+V)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyC) {
		*colorClipboard = c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyV) {
		// Only paste if target is not Bank 0 and not Index 0
		if e.selectedBank > 0 && e.selectedIndex > 0 {
			rgba := *colorClipboard
			r2, g2, b2 := rgba[0]/85, rgba[1]/85, rgba[2]/85
			c.SetBankPalette(e.selectedBank, e.selectedIndex, r2, g2, b2)
		}
	}

	// Adjustment Buttons: To the right (x=180)
	if e.selectedBank > 0 && e.selectedIndex > 0 {
		bx := 180
		by := 30
		if c.JustPressedMouse(gonsole.MouseButtonLeft) {
			for i := 0; i < 3; i++ {
				yy := by + 30 + i*25
				// Minus [-]
				if mx >= bx+40 && mx < bx+55 && my >= yy && my < yy+15 {
					e.adjust(c, i, -1)
				}
				// Plus [+]
				if mx >= bx+85 && mx < bx+100 && my >= yy && my < yy+15 {
					e.adjust(c, i, 1)
				}
			}
		}
	}
}

func (e *PaletteEditor) adjust(c *gonsole.Console, channel int, delta int) {
	rgba := c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
	vals := [3]byte{rgba[0]/85, rgba[1]/85, rgba[2]/85}
	
	val := int(vals[channel]) + delta
	if val < 0 { val = 0 }
	if val > 3 { val = 3 }
	vals[channel] = byte(val)
	
	c.SetBankPalette(e.selectedBank, e.selectedIndex, vals[0], vals[1], vals[2])
}

func (e *PaletteEditor) Draw(c *gonsole.Console) {
	// Draw palettes in a column on the left
	for b := 0; b < 4; b++ {
		bx := 10
		by := 30 + b*50
		
		label := fmt.Sprintf("B%d", b)
		if b == 0 { label += " (L)" }
		c.DrawText(bx, by-10, label)
		
		// Surround entire bank grid with grey
		c.DrawRect(bx-1, by-1, 8*20+1, 2*20+1, 5, false)

		for i := 0; i < 16; i++ {
			x := bx + (i%8)*20
			y := by + (i/8)*20
			
			// Draw color square
			e.drawAbsoluteRect(c, x, y, 16, 16, c.PaletteBank.Colors[b][i])
			
			// Individual grey outline for each color square (helps see black)
			c.DrawRect(x, y, 16, 16, 5, false)

			if e.selectedBank == b && e.selectedIndex == i {
				// Selection highlight (thick white border)
				c.DrawRect(x-2, y-2, 20, 20, 7, false)
			}
		}
	}

	// Draw controls to the right
	bx := 180
	by := 30
	rgba := c.PaletteBank.Colors[e.selectedBank][e.selectedIndex]
	r2, g2, b2 := rgba[0]/85, rgba[1]/85, rgba[2]/85
	
	c.DrawText(bx, by, fmt.Sprintf("Selected: Bank %d Index %d", e.selectedBank, e.selectedIndex))

	// Values are always shown, but editing/buttons are conditional
	c.DrawText(bx, by+15, fmt.Sprintf("Current RGB: %d, %d, %d", r2, g2, b2))

	if e.selectedBank == 0 || e.selectedIndex == 0 {
		msg := "LOCKED"
		if e.selectedBank == 0 { msg = "Bank 0 Is System Palette" }
		if e.selectedIndex == 0 { msg = "Color 0 Is Transparent" }
		c.DrawText(bx, by+35, msg)
	} else {
		channels := []string{"R", "G", "B"}
		vals := []byte{r2, g2, b2}
		for i, name := range channels {
			yy := by + 30 + i*25
			c.DrawText(bx, yy+5, name + ":")
			
			// Minus button
			c.DrawRect(bx+40, yy, 15, 15, 5, true)
			c.DrawText(bx+45, yy+3, "-")
			
			// Value display
			c.DrawText(bx+65, yy+5, fmt.Sprintf("%d", vals[i]))
			
			// Plus button
			c.DrawRect(bx+85, yy, 15, 15, 5, true)
			c.DrawText(bx+90, yy+3, "+")
		}
		c.DrawText(bx, by+110, "CTRL+C/V Copy/Paste")
	}
}

func (e *PaletteEditor) drawAbsoluteRect(c *gonsole.Console, x, y, w, h int, rgba [4]byte) {
	for dy := 0; dy < h; dy++ {
		yy := y + dy
		if yy < 0 || yy >= gonsole.ScreenHeight { continue }
		for dx := 0; dx < w; dx++ {
			xx := x + dx
			if xx < 0 || xx >= gonsole.ScreenWidth { continue }
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
