package main

import "aymanhs/gonsole"

func main() {
	c := gonsole.NewCon16(320, 240)

	t0 := c.GetTile(0)

	t0.Fill(0b10010100)

	c.SetDrawFunc(func(slot int, frame uint64) {
		// c.DisplayText(10, int(10+frame%240), "Hello, world!", 0b111111)
		x := int(frame % 320)
		y := int(frame % 240)
		c.SetPixel(x, y, 0b111111)
		c.DrawTile(x+10, y+10, 0)
	})

	c.Run()
}
