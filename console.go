package gonsole

import (
	"github.com/hajimehoshi/ebiten/v2"
)

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

const (
	ButtonUp    = 1 << 0
	ButtonDown  = 1 << 1
	ButtonLeft  = 1 << 2
	ButtonRight = 1 << 3
	ButtonA     = 1 << 4
	ButtonB     = 1 << 5
	ButtonStart = 1 << 6
	ButtonSelect = 1 << 7
)

type Console struct {
	VRAM       []byte
	Buttons    byte
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
	c.pollInputs()
	if c.UpdateFunc != nil {
		return c.UpdateFunc()
	}
	return nil
}

func (c *Console) pollInputs() {
	var b byte
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		b |= ButtonUp
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		b |= ButtonDown
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		b |= ButtonLeft
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		b |= ButtonRight
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyJ) {
		b |= ButtonA
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeyK) {
		b |= ButtonB
	}
	c.Buttons = b
}

// IsPressed returns true if the given button is currently pressed
func (c *Console) IsPressed(btn byte) bool {
	return (c.Buttons & btn) != 0
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
	c.VRAM[addr] = byte(x & 0xFF)
	c.VRAM[addr+1] = byte(y & 0xFF)
	c.VRAM[addr+2] = patternID
	// Use bit 0 of flags for the 9th bit of X
	if x > 255 {
		flags |= 0x01
	} else {
		flags &= 0xFE
	}
	c.VRAM[addr+3] = flags
}

// GetSprite retrieves OAM data for a specific sprite index
func (c *Console) GetSprite(index int) (x, y int, patternID byte, flags byte) {
	if index < 0 || index >= 256 {
		return 0, 0, 0, 0
	}
	addr := AddrOAM + index*4
	x = int(c.VRAM[addr])
	y = int(c.VRAM[addr+1])
	patternID = c.VRAM[addr+2]
	flags = c.VRAM[addr+3]
	
	// Check 9th bit of X in flags
	if (flags & 0x01) != 0 {
		x += 256
	}
	return x, y, patternID, flags
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

// Draw implements the ebiten.Game interface
func (c *Console) Draw(screen *ebiten.Image) {
	// 1. Render Framebuffer (Background)
	fb := c.VRAM[AddrFramebuffer : AddrFramebuffer+ScreenWidth*ScreenHeight]
	img := ebiten.NewImage(ScreenWidth, ScreenHeight)
	pixels := make([]byte, ScreenWidth*ScreenHeight*4)
	
	palette := c.VRAM[AddrPalette : AddrPalette+256*3]

	for i, idx := range fb {
		pAddr := int(idx) * 3
		pixels[i*4] = palette[pAddr]
		pixels[i*4+1] = palette[pAddr+1]
		pixels[i*4+2] = palette[pAddr+2]
		pixels[i*4+3] = 255
	}
	img.WritePixels(pixels)
	screen.DrawImage(img, nil)

	// 2. Render Sprites from OAM
	for i := 0; i < 256; i++ {
		x, y, patternID, _ := c.GetSprite(i)

		if x == 0 && y == 0 && patternID == 0 && i > 0 {
			continue // Skip unused sprites
		}

		// Draw 8x8 sprite
		patternAddr := AddrPatternData + int(patternID)*64
		spriteImg := ebiten.NewImage(8, 8)
		spritePixels := make([]byte, 8*8*4)
		for p := 0; p < 64; p++ {
			idx := c.VRAM[patternAddr+p]
			if idx == 0 { continue } // Transparency for index 0
			
			pAddr := int(idx) * 3
			spritePixels[p*4] = palette[pAddr]
			spritePixels[p*4+1] = palette[pAddr+1]
			spritePixels[p*4+2] = palette[pAddr+2]
			spritePixels[p*4+3] = 255
		}
		spriteImg.WritePixels(spritePixels)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(spriteImg, op)
	}
}

// Layout implements the ebiten.Game interface
func (c *Console) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// Run starts the virtual console with the given console state
func Run(c *Console) error {
	ebiten.SetWindowTitle("8-Bit Virtual Console")
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	return ebiten.RunGame(c)
}
