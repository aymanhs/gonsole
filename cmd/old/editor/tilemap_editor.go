package main

import (
	"fmt"

	"aymanhs/gonsole"
)

type TilemapEditor struct {
	layerIdx     int
	selectedTile int
	selectedBank int
	showTileBank bool
}

func (e *TilemapEditor) Update(c *gonsole.Console) {
	mx, my := c.MousePos()

	// Toggle tile bank with Tab or B
	if c.Buttons&0x80 != 0 { // Select/Tab dummy? (let's use raw keys in main if needed)
		// e.showTileBank = !e.showTileBank
	}

	// For now, let's use right-click or a specific area for the tile bank
	if mx > 240 {
		e.showTileBank = true
	} else {
		// e.showTileBank = false
	}

	if e.showTileBank {
		// Tile Selection from bank (16x16 tiles)
		bx, by := 245, 20
		for i := 0; i < 256; i++ {
			tx := bx + (i%16)*4
			ty := by + (i/16)*4
			if c.JustPressedMouse(gonsole.MouseButtonLeft) && mx >= tx && mx < tx+4 && my >= ty && my < ty+4 {
				e.selectedTile = i
			}
		}
	}

	// World coordinates placement
	// Grid is 8x8
	if !e.showTileBank || mx < 240 {
		if c.MousePressed(gonsole.MouseButtonLeft) {
			// Calculate world position
			// Screen pos -> World pos
			wx := (uint16(mx) + c.CameraX) / 8 * 8
			wy := (uint16(my) + c.CameraY) / 8 * 8
			
			// Simple "SetTile" - remove existing at this pos first?
			// For now, just add. We might need a better "Set" method in gonsole.
			e.setTileAt(c, wx, wy, byte(e.selectedTile))
		}
	}

	// Scroll with arrows
	speed := uint16(2)
	if c.Buttons&0x01 != 0 { c.CameraY -= speed }
	if c.Buttons&0x02 != 0 { c.CameraY += speed }
	if c.Buttons&0x04 != 0 { c.CameraX -= speed }
	if c.Buttons&0x08 != 0 { c.CameraX += speed }
}

func (e *TilemapEditor) setTileAt(c *gonsole.Console, wx, wy uint16, tileID byte) {
	l := &c.TileLayers[e.layerIdx]
	// Search for existing slot at this exact coordinate
	for i := 0; i < l.Count; i++ {
		if l.Slots[i].WorldX == wx && l.Slots[i].WorldY == wy {
			l.Slots[i].TileID = tileID
			l.Slots[i].BankID = byte(e.selectedBank)
			return
		}
	}
	// Not found, add new
	l.AddTile(wx, wy, tileID, byte(e.selectedBank), 0)
}

func (e *TilemapEditor) Draw(c *gonsole.Console) {
	mx, my := c.MousePos()

	// Draw Tile Bank
	if e.showTileBank {
		c.DrawRect(240, 0, 80, 240, 5, true) // sidebar
		c.DrawText(245, 5, "Tiles:")
		bx, by := 245, 20
		for i := 0; i < 256; i++ {
			tx := bx + (i%16)*4
			ty := by + (i/16)*4
			
			// Draw mini tile
			color := byte(6)
			if i == e.selectedTile {
				color = 7
				c.DrawRect(tx-1, ty-1, 5, 5, 7, false)
			}
			c.DrawRect(tx, ty, 3, 3, color, true)
		}
	}

	// Draw cursor/brush in world space
	if mx < 240 {
		wx := (uint16(mx) + c.CameraX) / 8 * 8
		wy := (uint16(my) + c.CameraY) / 8 * 8
		sx := int(wx) - int(c.CameraX)
		sy := int(wy) - int(c.CameraY)
		c.DrawRect(sx, sy, 8, 8, 7, false)
	}

	c.DrawText(5, 230, fmt.Sprintf("Cam: %d,%d | Tile: %d", c.CameraX, c.CameraY, e.selectedTile))
}
