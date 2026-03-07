package gonsole

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Run starts the virtual console with the given console state
func Run(c *Console) error {
	ebiten.SetWindowTitle("8-Bit Virtual Console")
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	d := NewDisplay(c)
	return ebiten.RunGame(d)
}

type Display struct {
	console *Console
}

func NewDisplay(c *Console) *Display {
	return &Display{
		console: c,
	}
}

func (d *Display) Update() error {
	return d.console.Update()
}

func (d *Display) Draw(screen *ebiten.Image) {
	// 1. Render Framebuffer (Background)
	fb := d.console.VRAM[AddrFramebuffer : AddrFramebuffer+ScreenWidth*ScreenHeight]
	img := ebiten.NewImage(ScreenWidth, ScreenHeight)
	pixels := make([]byte, ScreenWidth*ScreenHeight*4)
	
	palette := d.console.VRAM[AddrPalette : AddrPalette+256*3]

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
		addr := AddrOAM + i*4
		x := int(d.console.VRAM[addr])
		y := int(d.console.VRAM[addr+1])
		patternID := int(d.console.VRAM[addr+2])
		// flags := d.console.VRAM[addr+3] // To be used later

		if x == 0 && y == 0 && patternID == 0 && i > 0 {
			continue // Skip unused sprites
		}

		// Draw 8x8 sprite
		patternAddr := AddrPatternData + patternID*64
		spriteImg := ebiten.NewImage(8, 8)
		spritePixels := make([]byte, 8*8*4)
		for p := 0; p < 64; p++ {
			idx := d.console.VRAM[patternAddr+p]
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

func (d *Display) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
