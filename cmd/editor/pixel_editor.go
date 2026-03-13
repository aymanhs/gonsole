package main

import (
	"aymanhs/gonsole"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PixelEditor struct {
	targetType    int // 0 for sprite, 1 for tile, 2 for font
	targetBank    int // For tiles, 0-3
	targetID      int
	zoom          int
	cursorX       int
	cursorY       int
	selectedColor byte
	scratchIDs    [9]int
	scratchBanks  [9]int
	scratchTypes  [9]int
}

func (e *PixelEditor) Update(c *gonsole.Console, x, y int, clipboard *manip8x8) {
	mx, my := c.MousePos()
	cellSize := 16

	// Generic Manipulation Keys (H, V, R, Arrows)
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		e.manip(c, func(m *manip8x8) { m.flipH() })
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		e.manip(c, func(m *manip8x8) { m.flipV() })
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		e.manip(c, func(m *manip8x8) { m.rotate() })
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		e.manip(c, func(m *manip8x8) { m.shift(0, -1) })
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		e.manip(c, func(m *manip8x8) { m.shift(0, 1) })
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		e.manip(c, func(m *manip8x8) { m.shift(-1, 0) })
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		e.manip(c, func(m *manip8x8) { m.shift(1, 0) })
	}

	// Clipboard (Ctrl+C / Ctrl+V)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyC) {
		*clipboard = e.unpack(c)
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyV) {
		e.pack(c, *clipboard)
	}

	// Handle drawing on the grid (Sprite/Tile)
	if e.targetType != 2 && c.MousePressed(gonsole.MouseButtonLeft) {
		gx := (mx - x) / cellSize
		gy := (my - y) / cellSize
		if gx >= 0 && gx < 8 && gy >= 0 && gy < 8 {
			switch e.targetType {
			case 0: // Sprite
				setSpritePixel(c, e.targetID, gx, gy, e.selectedColor)
			case 1: // Tile
				c.TileBanks[e.targetBank].TileSetPixel(e.targetID, gx, gy, e.selectedColor)
			}
		}
	}

	// Handle Font bit-toggling (JustPressed)
	if e.targetType == 2 && c.JustPressedMouse(gonsole.MouseButtonLeft) {
		gx := (mx - x) / cellSize
		gy := (my - y) / cellSize
		if gx >= 0 && gx < 8 && gy >= 0 && gy < 8 {
			current := getFontPixel(c, e.targetID, gx, gy)
			if current == 0 {
				setFontPixel(c, e.targetID, gx, gy, 1)
			} else {
				setFontPixel(c, e.targetID, gx, gy, 0)
			}
		}
	}

	// Handle color selection from palette
	px, py := x, y+8*cellSize+10
	if e.targetType != 2 && c.JustPressedMouse(gonsole.MouseButtonLeft) {
		for i := 0; i < 16; i++ {
			cx := px + (i%8)*20
			cy := py + (i/8)*20
			if mx >= cx && mx < cx+16 && my >= cy && my < cy+16 {
				e.selectedColor = byte(i)
			}
		}
	}

	// Handle Bank Grid, Selectors, and Scratchpad if not Font
	if e.targetType != 2 && c.JustPressedMouse(gonsole.MouseButtonLeft) {
		// Target Type / Bank Selectors
		if my >= y && my < y+20 {
			if mx >= x+200 && mx < x+260 {
				e.targetType = 0
			} // Sprite
			if mx >= x+270 && mx < x+310 {
				e.targetType = 1
			} // Tile
			if e.targetType == 1 {
				for b := 0; b < 4; b++ {
					if mx >= x+320+b*20 && mx < x+335+b*20 {
						e.targetBank = b
					}
				}
			}
		}

		// Bank Grid
		gridX, gridY := x+200, y+30
		if mx >= gridX && mx < gridX+128 && my >= gridY && my < gridY+128 {
			col := (mx - gridX) / 8
			row := (my - gridY) / 8
			e.targetID = row*16 + col
		}

		// Scratchpad
		sx, sy := x+350, y+30
		if mx >= sx && mx < sx+72 && my >= sy && my < sy+72 {
			cx, cy := (mx-sx)/24, (my-sy)/24
			idx := cy*3 + cx
			e.scratchIDs[idx] = e.targetID
			e.scratchBanks[idx] = e.targetBank
			e.scratchTypes[idx] = e.targetType
		}
	}
}

func (e *PixelEditor) Draw(c *gonsole.Console, x, y int) {
	cellSize := 16

	// Draw Background outline
	c.DrawRect(x-1, y-1, 8*cellSize+2, 8*cellSize+2, 5, false)

	// Draw Grid
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			var colorIdx byte
			switch e.targetType {
			case 0:
				colorIdx = getSpritePixel(c, e.targetID, col, row)
			case 1:
				colorIdx = c.TileBanks[e.targetBank].TileGetPixel(e.targetID, col, row)
			case 2:
				colorIdx = getFontPixel(c, e.targetID, col, row)
				if colorIdx != 0 {
					colorIdx = 7 // White for "on"
				}
			}

			if colorIdx != 0 {
				c.DrawRect(x+col*cellSize, y+row*cellSize, cellSize, cellSize, colorIdx, true)
			}
			// Grid lines
			c.DrawRect(x+col*cellSize, y+row*cellSize, cellSize, cellSize, 5, false)
		}
	}

	// Draw Palette (only for Sprite/Tile)
	px, py := x, y+8*cellSize+10
	if e.targetType != 2 {
		for i := 0; i < 16; i++ {
			cx := px + (i%8)*20
			cy := py + (i/8)*20
			c.DrawRect(cx, cy, 16, 16, byte(i), true)

			// Draw blend flag indicator
			bMode := c.PaletteBank.Colors[0][i][3] // Using bank 0 palette for indicator here since that's what we draw with
			drawBlendIndicator(c, cx, cy, 16, 16, bMode)

			if e.selectedColor == byte(i) {
				c.DrawRect(cx-2, cy-2, 20, 20, 7, false)
			}
		}
	}

	// Draw Real-time Preview (1:1)
	c.DrawText(px+180, py-15, "Preview (1:1):")
	c.DrawRect(px+180, py, 10, 10, 5, false) // border

	if e.targetType != 2 {
		// Draw Mode / Bank Selectors
		c.DrawText(x+200, y, "MODE:")
		c.DrawText(x+240, y, "SPRITE")
		if e.targetType == 0 {
			c.DrawRect(x+238, y-2, 40, 12, 7, false)
		}

		c.DrawText(x+290, y, "TILE")
		if e.targetType == 1 {
			c.DrawRect(x+288, y-2, 30, 12, 7, false)
		}

		if e.targetType == 1 {
			c.DrawText(x+330, y, "BANK:")
			for b := 0; b < 4; b++ {
				c.DrawText(x+380+b*20, y, string(byte('0'+b)))
				if e.targetBank == b {
					c.DrawRect(x+378+b*20, y-2, 10, 12, 7, false)
				}
			}
		}

		// Draw Bank Grid 16x16
		gridX, gridY := x+200, y+30
		c.DrawRect(gridX-1, gridY-1, 130, 130, 5, false)
		pal := &c.PaletteBank.Colors[0]
		for i := 0; i < 256; i++ {
			gx := gridX + (i%16)*8
			gy := gridY + (i/16)*8
			if e.targetType == 0 {
				gonsole.BlitSprite(c.SpriteData[i][:], gx, gy, 0, pal, &c.Scratch)
			} else {
				gonsole.BlitTile(&c.TileBanks[e.targetBank], i, gx, gy, pal, &c.Scratch)
			}
			if i == e.targetID {
				c.DrawRect(gx-1, gy-1, 10, 10, 7, false)
			}
		}

		// Draw Scratchpad Preview (3x3 tiles, scaled 3x for visibility)
		sx, sy := x+400, y+30
		scale := 3
		c.DrawText(sx, sy-15, "Scratchpad (3x3):")
		c.DrawRect(sx-1, sy-1, 24*scale+2, 24*scale+2, 5, false)
		for cy := 0; cy < 3; cy++ {
			for cx := 0; cx < 3; cx++ {
				idx := cy*3 + cx
				sID := e.scratchIDs[idx]
				sBank := e.scratchBanks[idx]
				sType := e.scratchTypes[idx]

				for r := 0; r < 8; r++ {
					for cIdx := 0; cIdx < 8; cIdx++ {
						var colIdx byte
						if sType == 0 {
							colIdx = getSpritePixel(c, sID, cIdx, r)
						} else {
							colIdx = c.TileBanks[sBank].TileGetPixel(sID, cIdx, r)
						}
						if colIdx != 0 {
							c.DrawRect(sx+(cx*8+cIdx)*scale, sy+(cy*8+r)*scale, scale, scale, colIdx, true)
						}
					}
				}
			}
		}
	}

	switch e.targetType {
	case 0: // Sprite
		pal := &c.PaletteBank.Colors[0]
		gonsole.BlitSprite(c.SpriteData[e.targetID][:], px+181, py+1, 0, pal, &c.Scratch)
	case 1: // Tile
		pal := &c.PaletteBank.Colors[0]
		gonsole.BlitTile(&c.TileBanks[e.targetBank], e.targetID, px+181, py+1, pal, &c.Scratch)
	case 2: // Font
		// Draw 1-bit font preview
		for row := 0; row < 8; row++ {
			b := c.FontData[e.targetID][row]
			for col := 0; col < 8; col++ {
				if (b>>(7-col))&1 != 0 {
					c.SetPixel(px+181+col, py+1+row, 7)
				}
			}
		}
	}
}

func (e *PixelEditor) manip(c *gonsole.Console, f func(*manip8x8)) {
	m := e.unpack(c)
	f(&m)
	e.pack(c, m)
}

func (e *PixelEditor) unpack(c *gonsole.Console) manip8x8 {
	switch e.targetType {
	case 0: // Sprite
		return unpackNibble(c.SpriteData[e.targetID])
	case 1: // Tile
		return unpackNibble(c.TileBanks[e.targetBank].Tiles[e.targetID])
	case 2: // Font
		return unpackBit(c.FontData[e.targetID])
	}
	return manip8x8{}
}

func (e *PixelEditor) pack(c *gonsole.Console, m manip8x8) {
	switch e.targetType {
	case 0: // Sprite
		c.SpriteData[e.targetID] = packNibble(m)
	case 1: // Tile
		c.TileBanks[e.targetBank].Tiles[e.targetID] = packNibble(m)
	case 2: // Font
		c.FontData[e.targetID] = packBit(m)
	}
}

// Helpers for SpriteData since gonsole only has SetSpriteData (bulk)
func getSpritePixel(c *gonsole.Console, id, x, y int) byte {
	b := c.SpriteData[id][y*4+x/2]
	if x&1 == 0 {
		return (b >> 4) & 0xF
	}
	return b & 0xF
}

func setSpritePixel(c *gonsole.Console, id, x, y int, colorIdx byte) {
	i := y*4 + x/2
	if x&1 == 0 {
		c.SpriteData[id][i] = (c.SpriteData[id][i] & 0x0F) | (colorIdx << 4)
	} else {
		c.SpriteData[id][i] = (c.SpriteData[id][i] & 0xF0) | (colorIdx & 0xF)
	}
}

func getFontPixel(c *gonsole.Console, id, x, y int) byte {
	b := c.FontData[id][y]
	if (b>>(7-x))&1 != 0 {
		return 1
	}
	return 0
}

func setFontPixel(c *gonsole.Console, id, x, y int, colorIdx byte) {
	if colorIdx != 0 {
		c.FontData[id][y] |= (1 << (7 - x))
	} else {
		c.FontData[id][y] &= ^(1 << (7 - x))
	}
}
