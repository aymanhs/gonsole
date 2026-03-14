package main

import (
	"fmt"
	"math/rand"
	"time"

	"aymanhs/gonsole"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	X, Y      float64
	DX, DY    float64
	TileID    int
	PaletteID byte
}

func main() {
	c := gonsole.NewConsole()

	// Disable vsync for benchmarking
	ebiten.SetVsyncEnabled(false)

	// Set up a palette with varied blend modes
	// We'll populate 4 different palettes with random bright colors and 3 different blend modes at index 1.
	blendModes := []byte{gonsole.BlendNormal, gonsole.BlendAdd, gonsole.BlendSubtract}
	for p := 0; p < 4; p++ {
		c.PaletteBank.Colors[p][1] = [4]byte{
			byte(rand.Intn(200) + 55), // R
			byte(rand.Intn(200) + 55), // G
			byte(rand.Intn(200) + 55), // B
			blendModes[rand.Intn(len(blendModes))],
		}
	}

	// Create a simple tile in Bank 0 (a small filled square)
	for y := 1; y < 15; y++ {
		for x := 1; x < 15; x++ {
			c.TileBanks[0].TileSetPixel(0, x, y, 1) // We'll draw this tile using different palettes
		}
	}

	var sprites []Sprite

	addSprite := func(count int) {
		for i := 0; i < count; i++ {
			sprites = append(sprites, Sprite{
				X:         rand.Float64() * float64(gonsole.ScreenWidth-gonsole.TileSize),
				Y:         rand.Float64() * float64(gonsole.ScreenHeight-gonsole.TileSize),
				DX:        (rand.Float64() - 0.5) * 4.0,
				DY:        (rand.Float64() - 0.5) * 4.0,
				TileID:    0,
				PaletteID: byte(rand.Intn(4)),
			})
		}
	}

	addSprite(1000) // Start with 1000 sprites

	c.UpdateFunc = func(frame, ms uint64) error {
		// Handle input to add/remove sprites
		if c.JustPressed(gonsole.ButtonUp) {
			addSprite(1000)
		}
		if c.JustPressed(gonsole.ButtonDown) {
			if len(sprites) > 1000 {
				sprites = sprites[:len(sprites)-1000]
			} else {
				sprites = sprites[:0]
			}
		}

		// Move sprites
		for i := range sprites {
			s := &sprites[i]
			s.X += s.DX
			s.Y += s.DY

			// Bounce off edges
			if s.X < 0 || s.X > float64(gonsole.ScreenWidth-gonsole.TileSize) {
				s.DX = -s.DX
				s.X += s.DX
			}
			if s.Y < 0 || s.Y > float64(gonsole.ScreenHeight-gonsole.TileSize) {
				s.DY = -s.DY
				s.Y += s.DY
			}
		}
		return nil
	}

	var lastFrameTime time.Duration
	var frameStartTime time.Time

	c.PaintFunc = func(slot int, frame uint64) {
		if slot == gonsole.PaintSlotBegin {
			frameStartTime = time.Now()

			// Draw all sprites manually to test raw blit performance
			for _, s := range sprites {
				gonsole.BlitTile(&c.TileBanks[0], s.TileID, int(s.X), int(s.Y), &c.PaletteBank.Colors[s.PaletteID], &c.Scratch)
			}

			// Record time taken right after blitting
			lastFrameTime = time.Since(frameStartTime)

		} else if slot == gonsole.PaintSlotEnd {
			// Draw stats
			c.DrawText(4, 4, fmt.Sprintf("Sprites: %d", len(sprites)))
			c.DrawText(4, 20, fmt.Sprintf("Draw Time: %v", lastFrameTime))
			c.DrawText(4, 36, fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
			c.DrawText(4, 52, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
			c.DrawText(4, 68, "Up/Down: +/- 1000 sprites")
		}
	}

	if err := gonsole.Run(c); err != nil {
		fmt.Println("Error:", err)
	}
}
