package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"aymanhs/gonsole"
)

type PixelEditor struct {
	targetType int // 0 for sprite, 1 for tile, 2 for font
	targetID   int
	zoom       int
	cursorX    int
	cursorY    int
	selectedColor byte
}

func (e *PixelEditor) Update(c *gonsole.Console, x, y int, clipboard *manip8x8) {
	mx, my := c.MousePos()
	cellSize := 16

	// Generic Manipulation Keys (H, V, R, Arrows)
	if inpututil.IsKeyJustPressed(ebiten.KeyH) { e.manip(c, func(m *manip8x8) { m.flipH() }) }
	if inpututil.IsKeyJustPressed(ebiten.KeyV) { e.manip(c, func(m *manip8x8) { m.flipV() }) }
	if inpututil.IsKeyJustPressed(ebiten.KeyR) { e.manip(c, func(m *manip8x8) { m.rotate() }) }
	
	if inpututil.IsKeyJustPressed(ebiten.KeyUp)    { e.manip(c, func(m *manip8x8) { m.shift(0, -1) }) }
	if inpututil.IsKeyJustPressed(ebiten.KeyDown)  { e.manip(c, func(m *manip8x8) { m.shift(0, 1) }) }
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft)  { e.manip(c, func(m *manip8x8) { m.shift(-1, 0) }) }
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) { e.manip(c, func(m *manip8x8) { m.shift(1, 0) }) }

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
				c.TileBanks[0].TileSetPixel(e.targetID, gx, gy, e.selectedColor)
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
				colorIdx = c.TileBanks[0].TileGetPixel(e.targetID, col, row)
			case 2:
				colorIdx = getFontPixel(c, e.targetID, col, row)
				if colorIdx != 0 {
					colorIdx = 7 // White for "on"
				}
			}
			
			if colorIdx != 0 {
				c.DrawRect(x + col*cellSize, y + row*cellSize, cellSize, cellSize, colorIdx, true)
			}
			// Grid lines
			c.DrawRect(x + col*cellSize, y + row*cellSize, cellSize, cellSize, 5, false)
		}
	}
	
	// Draw Palette (only for Sprite/Tile)
	px, py := x, y+8*cellSize+10
	if e.targetType != 2 {
		c.DrawText(px, py-15, "Palette:")
		for i := 0; i < 16; i++ {
			cx := px + (i%8)*20
			cy := py + (i/8)*20
			c.DrawRect(cx, cy, 16, 16, byte(i), true)
			if e.selectedColor == byte(i) {
				c.DrawRect(cx-2, cy-2, 20, 20, 7, false)
			}
		}
	}

	// Draw Real-time Preview (1:1)
	c.DrawText(px + 180, py-15, "Preview (1:1):")
	c.DrawRect(px + 180, py, 10, 10, 5, false) // border
	switch e.targetType {
	case 0: // Sprite
		pal := &c.PaletteBank.Colors[0]
		gonsole.BlitSprite(c.SpriteData[e.targetID][:], px + 181, py + 1, 0, pal, &c.Scratch)
	case 1: // Tile
		pal := &c.PaletteBank.Colors[0]
		gonsole.BlitTile(&c.TileBanks[0], e.targetID, px + 181, py + 1, pal, &c.Scratch)
	case 2: // Font
		// Draw 1-bit font preview
		for row := 0; row < 8; row++ {
			b := c.FontData[e.targetID][row]
			for col := 0; col < 8; col++ {
				if (b >> (7 - col)) & 1 != 0 {
					c.SetPixel(px + 181 + col, py + 1 + row, 7)
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
		return unpackNibble(c.TileBanks[0].Tiles[e.targetID])
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
		c.TileBanks[0].Tiles[e.targetID] = packNibble(m)
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
	if (b >> (7 - x)) & 1 != 0 {
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
