package gonsole

import (
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Con16 struct {
	Frame  uint64
	TimeMs uint64

	tileBank      [256]Tile // A tile bank with 256 tiles (16x16 pixels each)
	screenWidth   int
	screenHeight  int
	colorMap      [64]uint32    // A precomputed color map for the 64 possible colors in the 6-bit palette
	screen        *ebiten.Image // An off-screen image used for drawing text and other elements
	frameBuffer   []byte        // A byte slice representing the RGBA pixel data for the screen
	frameBuffer32 []uint32      // A uint32 view of the frameBuffer for easier pixel manipulation
	updateFunc    func(frame, ms uint64) error
	drawFun       func(slot int, frame uint64)
}

type Tile [256]byte // 16x16 tile with 1 byte per pixel (color index)

func NewCon16(screenWidth, screenHeight int) *Con16 {
	c := &Con16{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		screen:       ebiten.NewImage(screenWidth, screenHeight),
		frameBuffer:  make([]byte, screenWidth*screenHeight*4),
	}
	c.frameBuffer32 = unsafe.Slice((*uint32)(unsafe.Pointer(&c.frameBuffer[0])), screenWidth*screenHeight)
	c.initColorMap()
	return c
}

func (c *Con16) initColorMap() {
	for i := 0; i < 64; i++ {
		r2 := (i >> 4) & 0b11
		g2 := (i >> 2) & 0b11
		b2 := (i >> 0) & 0b11
		r8 := uint32(r2 * 85) // 0, 85, 170, 255
		g8 := uint32(g2 * 85)
		b8 := uint32(b2 * 85)
		// the uint32 is little-endian RGBA, so the byte order is reversed (ABGR)
		c.colorMap[i] = (b8 << 16) | (g8 << 8) | r8
	}
}

// Implements ebiten.Game interface
func (c *Con16) Layout(outsideWidth, outsideHeight int) (int, int) {
	return c.screenWidth, c.screenHeight
}

// Implements ebiten.Game interface, called every frame to update game state
func (c *Con16) Update() error {
	if c.updateFunc != nil {
		return c.updateFunc(c.Frame, c.TimeMs)
	}
	c.TimeMs += 16 // Simulate ~60 FPS
	return nil
}

// Implements ebiten.Game interface, called every frame to render the screen
func (c *Con16) Draw(screen *ebiten.Image) {
	// c.screen.Fill(color.Black) // Clear the screen with black
	if c.drawFun != nil {
		c.drawFun(0, c.Frame)
	}
	c.Frame++
	screen.WritePixels(c.frameBuffer)
	screen.DrawImage(c.screen, nil)
}

func (c *Con16) SetUpdateFunc(f func(frame, ms uint64) error) {
	c.updateFunc = f
}

func (c *Con16) SetDrawFunc(f func(slot int, frame uint64)) {
	c.drawFun = f
}

func (c *Con16) Run() error {
	ebiten.SetWindowTitle("8-Bit Virtual Console")
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(c)
}

func (c *Con16) DisplayText(x, y int, text string, color byte) {
	ebitenutil.DebugPrintAt(c.screen, text, x, y)
}

// SetPixel sets a pixel at (x, y) with the specified color index and blend mode.
func (c *Con16) SetPixel(x, y int, color byte) {
	if x < 0 || x >= c.screenWidth || y < 0 || y >= c.screenHeight {
		return
	}
	index := (y*c.screenWidth + x)
	blendMode := ((color & 0xC0) >> 6) & 0x03
	colorIndex := color & 0x3F
	switch blendMode {
	case 0: // Normal
		c.frameBuffer32[index] = c.colorMap[colorIndex]
	case 1: // Additive blending
		existingColor := c.frameBuffer32[index]
		newColor := c.colorMap[colorIndex]
		r := min(((existingColor & 0xFF) + (newColor & 0xFF)), 255)
		g := min((((existingColor >> 8) & 0xFF) + ((newColor >> 8) & 0xFF)), 255)
		b := min((((existingColor >> 16) & 0xFF) + ((newColor >> 16) & 0xFF)), 255)
		c.frameBuffer32[index] = (b << 16) | (g << 8) | r
	case 2: // Subtractive blending
		existingColor := c.frameBuffer32[index]
		newColor := c.colorMap[colorIndex]
		r := max(((existingColor & 0xFF) - (newColor & 0xFF)), 0)
		g := max((((existingColor >> 8) & 0xFF) - ((newColor >> 8) & 0xFF)), 0)
		b := max((((existingColor >> 16) & 0xFF) - ((newColor >> 16) & 0xFF)), 0)
		c.frameBuffer32[index] = (b << 16) | (g << 8) | r
	case 3: // transparent, skip update
		return
	}
}

func (c *Con16) ClearScreen(color byte) {
	colorIndex := color & 0x3F
	for i := range c.frameBuffer32 {
		c.frameBuffer32[i] = c.colorMap[colorIndex]
	}
}

func (c *Con16) DrawTile(x, y int, tileIndex byte) {
	if x < 0 || x >= c.screenWidth || y < 0 || y >= c.screenHeight {
		return
	}
	tile := c.tileBank[tileIndex]
	for j := 0; j < 16; j++ {
		for i := 0; i < 16; i++ {
			colorIndex := tile[j*16+i]
			c.SetPixel(x+i, y+j, colorIndex)
		}
	}
}

func (c *Con16) GetTile(tileIndex byte) *Tile {
	return &c.tileBank[tileIndex]
}

func (t *Tile) Fill(color byte) {
	for i := range t {
		t[i] = color
	}
}

func (t *Tile) SetPixel(x, y int, color byte) {
	if x < 0 || x >= 16 || y < 0 || y >= 16 {
		return
	}
	t[y*16+x] = color
}

func (t *Tile) GetPixel(x, y int) byte {
	if x < 0 || x >= 16 || y < 0 || y >= 16 {
		return 0
	}
	return t[y*16+x]
}

func (t *Tile) FillRect(x, y, w, h int, color byte) {
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			t.SetPixel(x+i, y+j, color)
		}
	}
}
