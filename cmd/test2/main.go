package main

import (
	"aymanhs/gonsole"
	"math"
)

func main() {
	c := gonsole.NewCon16(320, 240)

	t0 := c.GetTile(0)

	t0.Fill(0b01000001)

	c.SetDrawFunc(func(slot int, frame int) {
		x := c.CameraX + 120
		y := 120 + int(100*math.Sin(float64(x)/320.0*math.Pi*2))
		c.SetPixel(x, y, 0b111111)
		c.DrawTile(frame, y+10, 0)
		c.CameraX += 1
		c.DrawLine(0, 0, 319, 239, byte(frame%64))
	})

	c.Run()
}
