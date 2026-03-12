package gonsole

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Console struct {
	// FramePaletteID selects which PaletteBank entry SetPixel/DrawLine etc. expand colors from.
	FramePaletteID byte

	// Palette and tile data
	Palette     [16][3]byte // legacy direct palette (kept for SetPalette compat)
	PaletteBank PaletteBank
	SpriteData  [256][64]byte // legacy 8-bit indexed sprite pixel data
	TileBanks   [4]TileBank   // up to 4 tile banks (each 256 tiles × 8×8)
	TileLayers  [TileLayerCount]TileLayer

	// Stamps (sprites/tiles positioned in the world or screen space)
	Stamps [256]Stamp

	// Camera (world-space pixel position of the top-left screen corner)
	CameraX uint16
	CameraY uint16

	// Input state
	Buttons      byte
	MouseX       int
	MouseY       int
	MouseButtons byte

	// Timing
	Frame     uint64
	TimeMs    uint64
	startTime time.Time

	// Callbacks
	UpdateFunc func(frame, ms uint64) error
	// PaintFunc is called at 7 fixed slots during compositing (use PaintSlot* constants).
	// All drawing primitives (SetPixel, DrawLine, DrawRect, DrawCircle) write directly into
	// the scratch buffer at that point in the pipeline — scratch IS the framebuffer.
	// The slot IS the draw queue; there is no separate persistent buffer.
	PaintFunc func(slot int, frame uint64)

	// Internal: persistent GPU image and RGBA scratch buffer
	screenImg *ebiten.Image
	scratch   [ScreenWidth * ScreenHeight * 4]byte
}

func NewConsole() *Console {
	c := &Console{startTime: time.Now()}
	c.Palette = defaultPalette
	// Pre-populate bank 0 from the default palette.
	// Index 0 is transparent; all others are opaque.
	for i, rgb := range defaultPalette {
		a := byte(255)
		if i == 0 {
			a = 0
		}
		c.PaletteBank.Colors[0][i] = [4]byte{rgb[0], rgb[1], rgb[2], a}
	}
	// Default tile layers: full-speed parallax (1:1)
	for i := range c.TileLayers {
		c.TileLayers[i] = NewTileLayer()
	}
	// Persistent GPU image — written once per frame via WritePixels
	c.screenImg = ebiten.NewImage(ScreenWidth, ScreenHeight)
	return c
}

func (c *Console) Update() error {
	c.Frame++
	c.TimeMs = uint64(time.Since(c.startTime).Milliseconds())
	c.pollInputs()
	c.pollMouse()
	if c.UpdateFunc != nil {
		return c.UpdateFunc(c.Frame, c.TimeMs)
	}
	return nil
}

func (c *Console) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func Run(c *Console) error {
	ebiten.SetWindowTitle("8-Bit Virtual Console")
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(c)
}
