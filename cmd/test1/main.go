package main

import (
	"log"

	"aymanhs/gonsole"
)

func main() {
	c := gonsole.NewConsole()

	// Demo Movement Logic hooked into Console.Update
	c.UpdateFunc = func() error {
		for i := 0; i < 10; i++ {
			x, y, patternID, flags := c.GetSprite(i)
			x = (x + 1) % gonsole.ScreenWidth
			y = (y + 1) % gonsole.ScreenHeight
			c.SetSprite(i, x, y, patternID, flags)
		}
		return nil
	}

	setupDemo(c)

	if err := gonsole.Run(c); err != nil {
		log.Fatal(err)
	}
}

func setupDemo(c *gonsole.Console) {
	// Create patterns
	patterns := map[int][]byte{
		0: { // Smiley
			0, 0, 1, 1, 1, 1, 0, 0,
			0, 1, 0, 0, 0, 0, 1, 0,
			1, 0, 1, 0, 0, 1, 0, 1,
			1, 0, 0, 0, 0, 0, 0, 1,
			1, 0, 1, 1, 1, 1, 0, 1,
			1, 0, 0, 0, 0, 0, 0, 1,
			0, 1, 0, 0, 0, 0, 1, 0,
			0, 0, 1, 1, 1, 1, 0, 0,
		},
		1: { // Heart
			0, 1, 1, 0, 0, 1, 1, 0,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			0, 1, 1, 1, 1, 1, 1, 0,
			0, 0, 1, 1, 1, 1, 0, 0,
			0, 0, 0, 1, 1, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		2: { // Ghost
			0, 0, 1, 1, 1, 1, 0, 0,
			0, 1, 1, 1, 1, 1, 1, 0,
			1, 1, 0, 1, 1, 0, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 0, 1, 0, 1, 0, 1, 0,
		},
	}

	for id, p := range patterns {
		patternData := make([]byte, 64)
		for i, val := range p {
			colorIdx := byte(0)
			if val > 0 {
				colorIdx = byte(id + 1)
			}
			patternData[i] = colorIdx
		}
		c.SetPattern(id, patternData)
	}

	// Palette
	c.SetPalette(1, 255, 255, 0)   // 1: Yellow
	c.SetPalette(2, 255, 0, 0)     // 2: Red
	c.SetPalette(3, 200, 200, 255) // 3: Light Blue

	// Fill gradient
	for i := 4; i < 256; i++ {
		r := byte((i * 7) % 256)
		g := byte((i * 11) % 256)
		b := byte((i * 17) % 256)
		c.SetPalette(i, r, g, b)
	}

	// Sprites
	for i := 0; i < 10; i++ {
		c.SetSprite(i, 50+i*20, 50+i*15, byte(i%3), 0)
	}

	// Static Background
	c.DrawRect(0, 0, gonsole.ScreenWidth, gonsole.ScreenHeight, 40, true)
	c.DrawRect(0, 0, gonsole.ScreenWidth, gonsole.ScreenHeight, 255, false)
	for b := 0; b < 4; b++ {
		c.DrawRect(20+b*80, 200, 20, 20, byte(100+b*40), true)
	}
	centerX, centerY := gonsole.ScreenWidth/2, gonsole.ScreenHeight/2
	c.DrawRect(centerX-20, centerY-20, 40, 40, 200, false)
	c.DrawLine(centerX-10, centerY, centerX+10, centerY, 200)
	c.DrawRect(centerX+10, centerY, 10, 20, 200, false)
}
