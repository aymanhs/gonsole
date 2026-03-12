package gonsole

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 320
	ScreenHeight = 240

	TransparentColor = 0 // palette index reserved for transparency

	// PaintSlot constants for PaintFunc — called between compositing steps.
	PaintSlotBegin   = 0 // before any rendering
	PaintSlotAfterL0 = 1 // after stamp layer 0 + tile layer 0
	PaintSlotAfterL1 = 2 // after stamp layer 1 + tile layer 1
	PaintSlotMid     = 3 // midpoint of the pipeline (between layer 1 and layer 2)
	PaintSlotAfterL2 = 4 // after stamp layer 2 + tile layer 2
	PaintSlotAfterL3 = 5 // after stamp layer 3 + tile layer 3
	PaintSlotEnd     = 6 // after HUD stamps (layers 4–7), before GPU upload
)

// defaultPalette is a PICO-8 inspired 16-color palette.
// Index 0 is reserved as transparent/black.
var defaultPalette = [16][3]byte{
	0:  {0, 0, 0},       // 0  black      (transparent)
	1:  {29, 43, 83},    // 1  dark blue
	2:  {126, 37, 83},   // 2  dark purple
	3:  {0, 135, 81},    // 3  dark green
	4:  {171, 82, 54},   // 4  brown
	5:  {95, 87, 79},    // 5  dark grey
	6:  {194, 195, 199}, // 6  light grey
	7:  {255, 241, 232}, // 7  white
	8:  {255, 0, 77},    // 8  red
	9:  {255, 163, 0},   // 9  orange
	10: {255, 236, 39},  // 10 yellow
	11: {0, 228, 54},    // 11 green
	12: {41, 173, 255},  // 12 blue
	13: {131, 118, 156}, // 13 lavender
	14: {255, 119, 168}, // 14 pink
	15: {255, 204, 170}, // 15 peach
}

// SetPalette updates a direct palette entry and syncs it to PaletteBank slot 0.
func (c *Console) SetPalette(index int, r, g, b byte) {
	if index < 0 || index >= 16 {
		return
	}
	c.Palette[index] = [3]byte{r, g, b}
	a := byte(255)
	if index == 0 {
		a = 0
	}
	c.PaletteBank.Colors[0][index] = [4]byte{r, g, b, a}
}

// SetPixel writes a palette color index at screen-space (x, y) directly into
// the scratch buffer. Call from inside PaintFunc; the slot determines where in
// the compositing order the pixel lands.
func (c *Console) SetPixel(x, y int, colorIdx byte) {
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return
	}
	rgba := &c.PaletteBank.Colors[c.FramePaletteID][colorIdx&0xF]
	if rgba[3] == 0 {
		return
	}
	dst := (y*ScreenWidth + x) * 4
	c.scratch[dst] = rgba[0]
	c.scratch[dst+1] = rgba[1]
	c.scratch[dst+2] = rgba[2]
	c.scratch[dst+3] = rgba[3]
}

// DrawLine draws a line using Bresenham's algorithm.
func (c *Console) DrawLine(x1, y1, x2, y2 int, colorIdx byte) {
	dx := abs(x2 - x1)
	dy := -abs(y2 - y1)
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx + dy

	for {
		c.SetPixel(x1, y1, colorIdx)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

// DrawRect draws a filled or outlined rectangle.
func (c *Console) DrawRect(x, y, w, h int, colorIdx byte, filled bool) {
	if filled {
		for i := 0; i < h; i++ {
			for j := 0; j < w; j++ {
				c.SetPixel(x+j, y+i, colorIdx)
			}
		}
	} else {
		c.DrawLine(x, y, x+w-1, y, colorIdx)
		c.DrawLine(x, y+h-1, x+w-1, y+h-1, colorIdx)
		c.DrawLine(x, y, x, y+h-1, colorIdx)
		c.DrawLine(x+w-1, y, x+w-1, y+h-1, colorIdx)
	}
}

// DrawCircle draws a filled or outlined circle using the midpoint algorithm.
func (c *Console) DrawCircle(xc, yc, r int, colorIdx byte, filled bool) {
	x := 0
	y := r
	d := 3 - 2*r

	drawPoints := func(xc, yc, x, y int) {
		if filled {
			c.DrawLine(xc-x, yc+y, xc+x, yc+y, colorIdx)
			c.DrawLine(xc-x, yc-y, xc+x, yc-y, colorIdx)
			c.DrawLine(xc-y, yc+x, xc+y, yc+x, colorIdx)
			c.DrawLine(xc-y, yc-x, xc+y, yc-x, colorIdx)
		} else {
			c.SetPixel(xc+x, yc+y, colorIdx)
			c.SetPixel(xc-x, yc+y, colorIdx)
			c.SetPixel(xc+x, yc-y, colorIdx)
			c.SetPixel(xc-x, yc-y, colorIdx)
			c.SetPixel(xc+y, yc+x, colorIdx)
			c.SetPixel(xc-y, yc+x, colorIdx)
			c.SetPixel(xc+y, yc-x, colorIdx)
			c.SetPixel(xc-y, yc-x, colorIdx)
		}
	}

	for y >= x {
		drawPoints(xc, yc, x, y)
		x++
		if d > 0 {
			y--
			d = d + 4*(x-y) + 10
		} else {
			d = d + 4*x + 6
		}
	}
}

// callPaint invokes PaintFunc at the given slot if one is registered.
func (c *Console) callPaint(slot int) {
	if c.PaintFunc != nil {
		c.PaintFunc(slot, c.Frame)
	}
}

// Draw implements ebiten.Game. All compositing is done in software into a
// single scratch buffer (the framebuffer); one WritePixels call pushes it to
// the GPU each frame.
//
// Draw order:
//
//	stamps layer 0 → tile layer 0 → stamps layer 1 → tile layer 1
//	→ stamps layer 2 → tile layer 2
//	→ stamps layer 3 → tile layer 3 → stamps layers 4–7
func (c *Console) Draw(screen *ebiten.Image) {
	// ── 0. Pre-pass: bucket visible stamps by DrawLayer ───────────────────────
	var buckets [8][256]byte
	var bucketLen [8]byte
	for i := 0; i < 256; i++ {
		s := c.Stamps[i]
		if s.Props&StampPropVisible == 0 {
			continue
		}
		l := s.DrawLayer
		if l > 7 {
			l = 7
		}
		buckets[l][bucketLen[l]] = byte(i)
		bucketLen[l]++
	}

	// ── 1. Clear scratch to fully transparent black ───────────────────────────
	for i := range c.scratch {
		c.scratch[i] = 0
	}

	// ── 2. Composite in order ─────────────────────────────────────────────────
	c.callPaint(PaintSlotBegin)
	c.drawStampBucket(buckets[0][:bucketLen[0]])
	c.drawTileLayer(0)
	c.callPaint(PaintSlotAfterL0)
	c.drawStampBucket(buckets[1][:bucketLen[1]])
	c.drawTileLayer(1)
	c.callPaint(PaintSlotAfterL1)
	c.callPaint(PaintSlotMid)
	c.drawStampBucket(buckets[2][:bucketLen[2]])
	c.drawTileLayer(2)
	c.callPaint(PaintSlotAfterL2)
	c.drawStampBucket(buckets[3][:bucketLen[3]])
	c.drawTileLayer(3)
	c.callPaint(PaintSlotAfterL3)
	for l := 4; l < 8; l++ {
		c.drawStampBucket(buckets[l][:bucketLen[l]])
	}
	c.callPaint(PaintSlotEnd)

	// ── 3. One WritePixels → GPU ──────────────────────────────────────────────
	c.screenImg.WritePixels(c.scratch[:])
	screen.DrawImage(c.screenImg, nil)
}

// drawTileLayer composites one tile layer onto the scratch buffer.
func (c *Console) drawTileLayer(layerIdx int) {
	layer := &c.TileLayers[layerIdx]
	if layer.Count == 0 {
		return
	}
	div := layer.ParallaxDiv
	if div == 0 {
		div = 1
	}
	camX := int(c.CameraX) * layer.ParallaxMul / div
	camY := int(c.CameraY) * layer.ParallaxMul / div

	for s := 0; s < layer.Count; s++ {
		slot := &layer.Slots[s]
		sx := int(slot.WorldX) - camX
		sy := int(slot.WorldY) - camY

		// Cull tiles fully outside the screen
		if sx+8 <= 0 || sx >= ScreenWidth || sy+8 <= 0 || sy >= ScreenHeight {
			continue
		}

		bank := &c.TileBanks[slot.BankID&3]
		pal := &c.PaletteBank.Colors[slot.PaletteID]
		blitTile(bank, int(slot.TileID), sx, sy, pal, &c.scratch)
	}
}

// drawStampBucket composites a pre-bucketed list of stamp indices.
func (c *Console) drawStampBucket(indices []byte) {
	for _, idx := range indices {
		s := &c.Stamps[idx]
		var sx, sy int
		if s.Props&StampPropScreenSpace != 0 {
			sx = int(s.X)
			sy = int(s.Y)
		} else {
			sx = int(s.X) - int(c.CameraX)
			sy = int(s.Y) - int(c.CameraY)
		}

		// Cull stamps fully outside the screen
		if sx+8 <= 0 || sx >= ScreenWidth || sy+8 <= 0 || sy >= ScreenHeight {
			continue
		}

		pal := &c.PaletteBank.Colors[s.PalletID]
		blitSprite(c.SpriteData[idx][:], sx, sy, s.Props, pal, &c.scratch)
	}
}

// blitTile copies an 8×8 tile from a TileBank onto the scratch buffer.
// Pixels with alpha=0 in the palette are skipped (transparent).
func blitTile(bank *TileBank, tileID, sx, sy int, pal *[16][4]byte, scratch *[ScreenWidth * ScreenHeight * 4]byte) {
	tile := &bank.Tiles[tileID]
	for row := 0; row < 8; row++ {
		dy := sy + row
		if dy < 0 || dy >= ScreenHeight {
			continue
		}
		dstRow := dy * ScreenWidth
		for col := 0; col < 8; col++ {
			dx := sx + col
			if dx < 0 || dx >= ScreenWidth {
				continue
			}
			// Unpack nibble: high = even col, low = odd col
			b := tile[row*4+col/2]
			var nibble byte
			if col&1 == 0 {
				nibble = (b >> 4) & 0xF
			} else {
				nibble = b & 0xF
			}
			rgba := &pal[nibble]
			if rgba[3] == 0 {
				continue
			}
			dst := (dstRow + dx) * 4
			scratch[dst] = rgba[0]
			scratch[dst+1] = rgba[1]
			scratch[dst+2] = rgba[2]
			scratch[dst+3] = rgba[3]
		}
	}
}

// blitSprite copies an 8×8 sprite from legacy SpriteData onto the scratch buffer.
// Pixels with alpha=0 in the palette are skipped.
func blitSprite(data []byte, sx, sy int, props byte, pal *[16][4]byte, scratch *[ScreenWidth * ScreenHeight * 4]byte) {
	for row := 0; row < 8; row++ {
		srcRow := row
		if props&StampPropFlipV != 0 {
			srcRow = 7 - row
		}
		dy := sy + row
		if dy < 0 || dy >= ScreenHeight {
			continue
		}
		dstRow := dy * ScreenWidth
		for col := 0; col < 8; col++ {
			srcCol := col
			if props&StampPropFlipH != 0 {
				srcCol = 7 - col
			}
			dx := sx + col
			if dx < 0 || dx >= ScreenWidth {
				continue
			}
			colorIdx := data[srcRow*8+srcCol] & 0xF
			rgba := &pal[colorIdx]
			if rgba[3] == 0 {
				continue
			}
			dst := (dstRow + dx) * 4
			scratch[dst] = rgba[0]
			scratch[dst+1] = rgba[1]
			scratch[dst+2] = rgba[2]
			scratch[dst+3] = rgba[3]
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
