package gonsole

const (
	ScreenWidth  = 640
	ScreenHeight = 480
	TileSize     = 16

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

// ApplyBlend applies the calculated blend mode values into the pixel byte array.
func ApplyBlend(scratch *[ScreenWidth * ScreenHeight * 4]byte, dst int, rgba *[4]byte) {
	blend := rgba[3]
	if blend == BlendTransparent {
		return // Skip rendering
	}

	if blend == BlendNormal {
		scratch[dst] = rgba[0]
		scratch[dst+1] = rgba[1]
		scratch[dst+2] = rgba[2]
		scratch[dst+3] = 255
	} else if blend == BlendSubtract {
		r := int(scratch[dst]) - int(rgba[0])
		g := int(scratch[dst+1]) - int(rgba[1])
		b := int(scratch[dst+2]) - int(rgba[2])
		if r < 0 {
			r = 0
		}
		if g < 0 {
			g = 0
		}
		if b < 0 {
			b = 0
		}
		scratch[dst] = byte(r)
		scratch[dst+1] = byte(g)
		scratch[dst+2] = byte(b)
		scratch[dst+3] = 255
	} else if blend == BlendAdd {
		r := int(scratch[dst]) + int(rgba[0])
		g := int(scratch[dst+1]) + int(rgba[1])
		b := int(scratch[dst+2]) + int(rgba[2])
		if r > 255 {
			r = 255
		}
		if g > 255 {
			g = 255
		}
		if b > 255 {
			b = 255
		}
		scratch[dst] = byte(r)
		scratch[dst+1] = byte(g)
		scratch[dst+2] = byte(b)
		scratch[dst+3] = 255
	}
}

// SetPixel writes a palette color index at screen-space (x, y) directly into
// the scratch buffer. Call from inside PaintFunc; the slot determines where in
// the compositing order the pixel lands.
func (c *Console) SetPixel(x, y int, colorIdx byte) {
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return
	}
	rgba := &c.PaletteBank.Colors[c.FramePaletteID][colorIdx&0xF]
	ApplyBlend(&c.Scratch, (y*ScreenWidth+x)*4, rgba)
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

// DrawText draws a string at screen-space (x, y) using the built-in ebiten
// debug font (6×16px, white). Safe to call from PaintFunc at any slot.
// Text is drawn onto screenImg after scratch is uploaded to the GPU, so it
// always appears on top regardless of which slot PaintFunc calls it from.
func (c *Console) DrawText(x, y int, text string) {
	c.pendingTexts = append(c.pendingTexts, textDraw{x, y, text})
}

// DrawCustomText draws a string using the Console's internal FontData.
// Safe to call from PaintFunc; writes directly into the scratch buffer.
func (c *Console) DrawCustomText(x, y int, text string, paletteID byte) {
	for i, r := range text {
		if r > 127 {
			r = '?'
		}
		charData := c.FontData[r]
		for row := 0; row < 8; row++ {
			b := charData[row]
			for col := 0; col < 8; col++ {
				if (b>>(7-col))&1 != 0 {
					c.SetPixel(x+i*8+col, y+row, 1) // Default to color 1? Or from palette?
					// Actually, let's use the provided paletteID color at index 7 (usually white)
					c.SetPixel(x+i*8+col, y+row, 7)
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
