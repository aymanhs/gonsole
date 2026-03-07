package gonsole

const (
	ScreenWidth  = 320
	ScreenHeight = 240
	VRAMSize     = 0x20000 // 128KB

	// Memory Map
	AddrFramebuffer = 0x00000
	AddrOAM         = 0x12C00 // Sprite Attribute Table
	AddrPatternData = 0x13000 // Sprite Pattern Data
	AddrPalette     = 0x17000 // 256 * 3 bytes (RGB)
)

type Console struct {
	VRAM       []byte
	UpdateFunc func() error
}

func NewConsole() *Console {
	c := &Console{
		VRAM: make([]byte, VRAMSize),
	}
	return c
}

// Update handles logic (to be expanded with stack-based VM)
func (c *Console) Update() error {
	if c.UpdateFunc != nil {
		return c.UpdateFunc()
	}
	return nil
}

// SetPixel sets the color index at (x, y)
func (c *Console) SetPixel(x, y int, colorIdx byte) {
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return
	}
	c.VRAM[AddrFramebuffer+y*ScreenWidth+x] = colorIdx
}

// GetPixel returns the color index at (x, y)
func (c *Console) GetPixel(x, y int) byte {
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return 0
	}
	return c.VRAM[AddrFramebuffer+y*ScreenWidth+x]
}

// DrawLine draws a line using Bresenham's algorithm
func (c *Console) DrawLine(x1, y1, x2, y2 int, colorIdx byte) {
	dx := abs(x2 - x1)
	dy := -abs(y2 - y1)
	sx := 1
	if x1 >= x2 { sx = -1 }
	sy := 1
	if y1 >= y2 { sy = -1 }
	err := dx + dy

	for {
		c.SetPixel(x1, y1, colorIdx)
		if x1 == x2 && y1 == y2 { break }
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

// DrawRect draws a rectangle
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

// DrawCircle draws a circle using midpoint algorithm
func (c *Console) DrawCircle(xc, yc, r int, colorIdx byte, filled bool) {
	x := 0
	y := r
	d := 3 - 2 * r
	
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
			d = d + 4 * (x - y) + 10
		} else {
			d = d + 4 * x + 6
		}
	}
}

// SetSprite updates OAM data for a specific sprite index
func (c *Console) SetSprite(index int, x, y int, patternID byte, flags byte) {
	if index < 0 || index >= 256 {
		return
	}
	addr := AddrOAM + index*4
	c.VRAM[addr] = byte(x)
	c.VRAM[addr+1] = byte(y)
	c.VRAM[addr+2] = patternID
	c.VRAM[addr+3] = flags
}

// GetSprite retrieves OAM data for a specific sprite index
func (c *Console) GetSprite(index int) (x, y int, patternID byte, flags byte) {
	if index < 0 || index >= 256 {
		return 0, 0, 0, 0
	}
	addr := AddrOAM + index*4
	return int(c.VRAM[addr]), int(c.VRAM[addr+1]), c.VRAM[addr+2], c.VRAM[addr+3]
}

// SetPalette updates a single palette entry
func (c *Console) SetPalette(index int, r, g, b byte) {
	if index < 0 || index >= 256 {
		return
	}
	addr := AddrPalette + index*3
	c.VRAM[addr] = r
	c.VRAM[addr+1] = g
	c.VRAM[addr+2] = b
}

// SetPattern updates a single sprite pattern (8x8)
func (c *Console) SetPattern(index int, data []byte) {
	if index < 0 || index >= 256 || len(data) != 64 {
		return
	}
	addr := AddrPatternData + index*64
	copy(c.VRAM[addr:addr+64], data)
}

func abs(x int) int {
	if x < 0 { return -x }
	return x
}
