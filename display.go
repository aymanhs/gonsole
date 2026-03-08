package gonsole

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 320
	ScreenHeight = 240

	TransparentColor = 0 // palette index reserved for transparency
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

// SetPalette updates a single palette entry (index 0–15).
func (c *Console) SetPalette(index int, r, g, b byte) {
	if index < 0 || index >= 16 {
		return
	}
	c.Palette[index] = [3]byte{r, g, b}
}

// SetPixel sets the color index at world position (x, y), adjusted by OffsetX/OffsetY.
func (c *Console) SetPixel(x, y int, colorIdx byte) {
	x -= c.OffsetX
	y -= c.OffsetY
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return
	}
	c.Framebuffer[y*ScreenWidth+x] = colorIdx
}

// GetPixel returns the color index at world position (x, y), adjusted by OffsetX/OffsetY.
func (c *Console) GetPixel(x, y int) byte {
	x -= c.OffsetX
	y -= c.OffsetY
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return 0
	}
	return c.Framebuffer[y*ScreenWidth+x]
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

// Draw implements the ebiten.Game interface.
func (c *Console) Draw(screen *ebiten.Image) {
	c.screen = screen
	defer func() { c.screen = nil }()

	// 1. Render Framebuffer
	img := ebiten.NewImage(ScreenWidth, ScreenHeight)
	pixels := make([]byte, ScreenWidth*ScreenHeight*4)
	for i, idx := range c.Framebuffer {
		rgb := c.Palette[idx&0xF]
		pixels[i*4] = rgb[0]
		pixels[i*4+1] = rgb[1]
		pixels[i*4+2] = rgb[2]
		pixels[i*4+3] = 255
	}
	img.WritePixels(pixels)
	screen.DrawImage(img, nil)

	// 2. Render Sprites
	for i := 0; i < 256; i++ {
		s := c.Sprites[i]
		if s.Props&SpritePropVisible == 0 {
			continue
		}

		spriteImg := ebiten.NewImage(8, 8)
		spritePixels := make([]byte, 8*8*4)
		for p, colorIdx := range c.SpriteData[i] {
			if colorIdx == TransparentColor {
				continue
			}
			rgb := c.Palette[colorIdx&0xF]
			spritePixels[p*4] = rgb[0]
			spritePixels[p*4+1] = rgb[1]
			spritePixels[p*4+2] = rgb[2]
			spritePixels[p*4+3] = 255
		}
		spriteImg.WritePixels(spritePixels)

		op := &ebiten.DrawImageOptions{}
		if s.Props&SpritePropFlipH != 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(8, 0)
		}
		if s.Props&SpritePropFlipV != 0 {
			op.GeoM.Scale(1, -1)
			op.GeoM.Translate(0, 8)
		}
		dx := float64(s.X)
		dy := float64(s.Y)
		if s.Props&SpritePropScreenSpace == 0 {
			dx -= float64(c.OffsetX)
			dy -= float64(c.OffsetY)
		}
		op.GeoM.Translate(dx, dy)
		screen.DrawImage(spriteImg, op)
	}

	// 3. User draw callback
	if c.DrawFunc != nil {
		c.DrawFunc(c.Frame, c.TimeMs)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
