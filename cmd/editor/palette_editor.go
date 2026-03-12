package main

import (
	"fmt"
	"aymanhs/gonsole"
)

type PaletteEditor struct {
	bankID   int
	selected int
}

func (e *PaletteEditor) Update(c *gonsole.Console) {
	mx, my := c.MousePos()

	// Mouse Selection
	for i := 0; i < 16; i++ {
		x := 10 + (i % 8) * 20
		y := 40 + (i / 8) * 20
		if c.JustPressedMouse(gonsole.MouseButtonLeft) && mx >= x && mx < x+16 && my >= y && my < y+16 {
			e.selected = i
		}
	}

	// Keyboard selection
	if c.Buttons&0x01 != 0 { e.selected = (e.selected - 1 + 16) % 16 }
	if c.Buttons&0x02 != 0 { e.selected = (e.selected + 1) % 16 }

	// Color adjustment (Q/W/E/R for R/G/B/A step)
	// For now, let's just show info. Editing palette requires SetPalette call.
}

func (e *PaletteEditor) Draw(c *gonsole.Console) {
	c.DrawText(10, 25, "Color Palette:")

	// Draw the 16 colors as squares
	for i := 0; i < 16; i++ {
		x := 10 + (i % 8) * 20
		y := 40 + (i / 8) * 20
		
		colorIdx := byte(i)
		c.DrawRect(x, y, 16, 16, colorIdx, true)
		
		if e.selected == i {
			c.DrawRect(x-2, y-2, 20, 20, 7, false) // White border for selection
		}
	}

	// Draw RGB info for selected
	rgba := c.PaletteBank.Colors[e.bankID][e.selected]
	info := fmt.Sprintf("Index %d | R:%d G:%d B:%d A:%d", 
		e.selected, rgba[0], rgba[1], rgba[2], rgba[3])
	c.DrawText(10, 85, info)
}
