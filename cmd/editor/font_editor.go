package main

import (
	"aymanhs/gonsole"
)

type FontEditor struct {
	selectedChar rune
	pixelEditor  PixelEditor
}

func (e *FontEditor) Update(c *gonsole.Console, x, y int, clipboard *manip8x8) {
	mx, my := c.MousePos()
	
	// Character Selection Grid (16x8 for 128 characters)
	// Positioned to the right of the main editor
	selX, selY := x + 150, y
	for i := 0; i < 128; i++ {
		cx := selX + (i%16)*8
		cy := selY + (i/16)*8
		if c.JustPressedMouse(gonsole.MouseButtonLeft) && mx >= cx && mx < cx+8 && my >= cy && my < cy+8 {
			e.selectedChar = rune(i)
		}
	}

	// Bridge PixelEditor to FontData
	e.pixelEditor.targetType = 2 // Font
	e.pixelEditor.targetID = int(e.selectedChar)
	e.pixelEditor.Update(c, x, y, clipboard)
}

func (e *FontEditor) Draw(c *gonsole.Console, x, y int) {
	e.pixelEditor.Draw(c, x, y)
	
	// Draw Character Selection Grid
	selX, selY := x + 150, y
	c.DrawText(selX, selY-15, "Select Char:")
	for i := 0; i < 128; i++ {
		cx := selX + (i%16)*8
		cy := selY + (i/16)*8
		
		if rune(i) == e.selectedChar {
			c.DrawRect(cx-1, cy-1, 10, 10, 7, false)
		}
		
		// Direct blit of small char (1-bit)
		charData := c.FontData[i]
		for row := 0; row < 8; row++ {
			b := charData[row]
			for col := 0; col < 8; col++ {
				if (b >> (7 - col)) & 1 != 0 {
					c.SetPixel(cx+col, cy+row, 7)
				}
			}
		}
	}

	c.DrawText(x, y-15, "Font: '" + string(e.selectedChar) + "'")
}
