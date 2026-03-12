package main

import (
	"log"

	"aymanhs/gonsole"
)

// cursorArrow is the default mouse cursor — a top-left pointing arrow.
// Colors: 7=white (body), 5=dark grey (inner shadow). 0=transparent.
var cursorArrow = [64]byte{
	7, 0, 0, 0, 0, 0, 0, 0,
	7, 7, 0, 0, 0, 0, 0, 0,
	7, 5, 7, 0, 0, 0, 0, 0,
	7, 5, 5, 7, 0, 0, 0, 0,
	7, 5, 5, 5, 7, 0, 0, 0,
	7, 7, 5, 0, 0, 0, 0, 0,
	0, 7, 5, 7, 0, 0, 0, 0,
	0, 0, 7, 0, 0, 0, 0, 0,
}

// cursorClick is shown while any mouse button is held — a diamond cross.
// Colors: 9=orange (body), 5=dark grey (detail). 0=transparent.
var cursorClick = [64]byte{
	0, 0, 9, 0, 0, 0, 0, 0,
	0, 9, 5, 9, 0, 0, 0, 0,
	9, 5, 9, 5, 9, 0, 0, 0,
	0, 9, 5, 9, 0, 0, 0, 0,
	0, 0, 9, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

func main() {
	c := gonsole.NewConsole()

	// Demo Movement Logic hooked into Console.Update
	c.UpdateFunc = func(frame, ms uint64) error {
		// Player-controlled sprite (index 0)
		s := c.GetStamp(0)
		if c.IsPressed(gonsole.ButtonUp) && s.Y > 0 {
			s.Y--
		}
		if c.IsPressed(gonsole.ButtonDown) {
			s.Y++
		}
		if c.IsPressed(gonsole.ButtonLeft) && s.X > 0 {
			s.X--
		}
		if c.IsPressed(gonsole.ButtonRight) {
			s.X++
		}

		// Clamp to screen
		if int(s.X) > gonsole.ScreenWidth-8 {
			s.X = uint16(gonsole.ScreenWidth - 8)
		}
		if int(s.Y) > gonsole.ScreenHeight-8 {
			s.Y = uint16(gonsole.ScreenHeight - 8)
		}
		c.SetStamp(0, s)

		// Other sprites (1-9) move automatically
		for i := 1; i < 10; i++ {
			s := c.GetStamp(i)
			s.X = (s.X + 1) % uint16(gonsole.ScreenWidth)
			s.Y = (s.Y + 1) % uint16(gonsole.ScreenHeight)
			c.SetStamp(i, s)
		}
		// Mouse cursor sprite (slot 10) — follows mouse, changes shape on click
		mx, my := c.MousePos()
		cursor := c.GetStamp(10)
		cursor.X = uint16(mx)
		cursor.Y = uint16(my)
		c.SetStamp(10, cursor)
		if c.MousePressed(gonsole.MouseButtonLeft) || c.MousePressed(gonsole.MouseButtonRight) {
			c.SetSpriteData(10, cursorClick[:])
		} else {
			c.SetSpriteData(10, cursorArrow[:])
		}

		return nil
	}

	c.PaintFunc = func(slot int, frame uint64) {
		if slot == gonsole.PaintSlotBegin {
			// Static Background — drawn every frame at the start
			c.DrawRect(0, 0, gonsole.ScreenWidth, gonsole.ScreenHeight, 1, true)  // dark blue fill
			c.DrawRect(0, 0, gonsole.ScreenWidth, gonsole.ScreenHeight, 7, false) // white border
			for b := 0; b < 4; b++ {
				c.DrawRect(20+b*80, 200, 20, 20, byte(3+b), true) // green/brown/grey/light-grey boxes
			}
			centerX, centerY := gonsole.ScreenWidth/2, gonsole.ScreenHeight/2
			c.DrawRect(centerX-20, centerY-20, 40, 40, 6, false)
			c.DrawLine(centerX-10, centerY, centerX+10, centerY, 6)
			c.DrawRect(centerX+10, centerY, 10, 20, 6, false)
		}
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

	// Sprites — each slot gets its own copy of pixel data, cycling through 3 patterns
	// Colors: 10=yellow, 8=red, 12=blue (from default palette)
	spriteColors := []byte{10, 8, 12}
	patternList := [][]byte{patterns[0], patterns[1], patterns[2]}
	for i := 0; i < 10; i++ {
		src := patternList[i%3]
		spriteData := make([]byte, 64)
		for j, val := range src {
			if val > 0 {
				spriteData[j] = spriteColors[i%3]
			}
		}
		c.SetSpriteData(i, spriteData)
		c.SetStamp(i, gonsole.Stamp{
			X:         uint16(50 + i*20),
			Y:         uint16(50 + i*15),
			Props:     gonsole.StampPropVisible,
			PaletteID: 0,
		})
	}

	// Mouse cursor sprite (slot 10) — screen-space so camera doesn't affect it
	c.SetSpriteData(10, cursorArrow[:])
	c.SetStamp(10, gonsole.Stamp{
		Props:     gonsole.StampPropVisible | gonsole.StampPropScreenSpace,
		PaletteID: 0,
	})
}
