package gonsole

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Console struct {
	// FramePaletteID selects which PaletteBank entry SetPixel/DrawLine etc. expand colors from.
	FramePaletteID byte

	// Palette and tile data
	Palette     [16][3]byte // legacy direct palette (kept for SetPalette compat)
	PaletteBank PaletteBank
	SpriteData  [256][256]byte // 1-byte per pixel sprite data (16×16)
	TileBanks   [4]TileBank    // up to 4 tile banks (each 256 tiles × 16×16)
	TileLayers  [TileLayerCount]TileLayer
	FontData    [128][8]byte // 128 characters, each 8x8 (1 bit per pixel, 8 bytes total)

	// Stamps (sprites/tiles positioned in the world or screen space)
	Stamps [256]Stamp

	// Camera (world-space pixel position of the top-left screen corner)
	CameraX uint16
	CameraY uint16

	// Input state
	Buttons          byte
	prevButtons      byte
	MouseX           int
	MouseY           int
	MouseButtons     byte
	prevMouseButtons byte

	// Timing
	Frame     uint64
	TimeMs    uint64
	startTime time.Time

	// Callbacks
	UpdateFunc func(frame, ms uint64) error
	PaintFunc  func(slot int, frame uint64)

	// OverlayFunc is called after the internal low-res buffer is drawn to screen,
	// allowing for high-res overlays like debug info or editor UI.
	OverlayFunc func(screen *ebiten.Image)

	// Internal: persistent GPU image and RGBA scratch buffer
	screenImg *ebiten.Image
	Scratch   [ScreenWidth * ScreenHeight * 4]byte

	pendingTexts []textDraw
}

type textDraw struct {
	x, y int
	text string
}

func NewConsole() *Console {
	c := &Console{startTime: time.Now()}
	c.Palette = defaultPalette
	// Pre-populate bank 0 from the default palette.
	for i, rgb := range defaultPalette {
		blend := BlendNormal
		if i == 0 {
			blend = BlendTransparent // Backwards compat: preserve index 0 as transparent by default in Bank 0
		}
		c.PaletteBank.Colors[0][i] = [4]byte{rgb[0], rgb[1], rgb[2], blend}
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

	// ── 1. Clear Scratch to fully transparent black ───────────────────────────
	for i := range c.Scratch {
		c.Scratch[i] = 0
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
	c.screenImg.WritePixels(c.Scratch[:])
	screen.DrawImage(c.screenImg, nil)

	for _, t := range c.pendingTexts {
		ebitenutil.DebugPrintAt(screen, t.text, t.x, t.y)
	}
	c.pendingTexts = c.pendingTexts[:0]

	if c.OverlayFunc != nil {
		c.OverlayFunc(screen)
	}
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
		if sx+TileSize <= 0 || sx >= ScreenWidth || sy+TileSize <= 0 || sy >= ScreenHeight {
			continue
		}

		bank := &c.TileBanks[slot.BankID&3]
		pal := &c.PaletteBank.Colors[slot.PaletteID]
		BlitTile(bank, int(slot.TileID), sx, sy, pal, &c.Scratch)
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
		if sx+TileSize <= 0 || sx >= ScreenWidth || sy+TileSize <= 0 || sy >= ScreenHeight {
			continue
		}

		pal := &c.PaletteBank.Colors[s.PaletteID]
		BlitSprite(c.SpriteData[idx][:], sx, sy, s.Props, pal, &c.Scratch)
	}
}
