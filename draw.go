package gonsole

// SetPixel sets a pixel at (x, y) as World coordinates, applying the camera offset and blending mode from the color byte
func (c *Con16) SetPixel(x, y int, color byte) {
	x -= int(c.CameraX)
	y -= int(c.CameraY)
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
		newColor := c.colorMap[colorIndex]
		index = index << 2                                                                            // Convert pixel index to byte index for RGBA
		c.frameBuffer[index+0] = byte(min(int(c.frameBuffer[index+0])+int((newColor>>0)&0xFF), 255))  // R
		c.frameBuffer[index+1] = byte(min(int(c.frameBuffer[index+1])+int((newColor>>8)&0xFF), 255))  // G
		c.frameBuffer[index+2] = byte(min(int(c.frameBuffer[index+2])+int((newColor>>16)&0xFF), 255)) // B
	case 2: // Subtractive blending
		newColor := c.colorMap[colorIndex]
		index = index << 2                                                                          // Convert pixel index to byte index for RGBA
		c.frameBuffer[index+0] = byte(max(int(c.frameBuffer[index+0])-int((newColor>>0)&0xFF), 0))  // R
		c.frameBuffer[index+1] = byte(max(int(c.frameBuffer[index+1])-int((newColor>>8)&0xFF), 0))  // G
		c.frameBuffer[index+2] = byte(max(int(c.frameBuffer[index+2])-int((newColor>>16)&0xFF), 0)) // B
	case 3: // transparent, skip update
		return
	}
}

// DrawLine draws a line from (x1, y1) to (x2, y2) in World coordinates using Bresenham's algorithm,
// It is optimized to only call SetPixel for pixels that are actually on-screen
func (c *Con16) DrawLine(x1, y1, x2, y2 int, color byte) {
	// Cohen-Sutherland clipping region codes
	const (
		INSIDE = 0
		LEFT   = 1
		RIGHT  = 2
		BOTTOM = 4
		TOP    = 8
	)

	xmin, ymin := c.CameraX, c.CameraY
	xmax, ymax := c.CameraX+c.screenWidth-1, c.CameraY+c.screenHeight-1

	computeOutCode := func(x, y int) int {
		code := INSIDE
		if x < xmin {
			code |= LEFT
		} else if x > xmax {
			code |= RIGHT
		}
		if y < ymin {
			code |= TOP
		} else if y > ymax {
			code |= BOTTOM
		}
		return code
	}

	outcode0 := computeOutCode(x1, y1)
	outcode1 := computeOutCode(x2, y2)
	accept := false

	for {
		if outcode0|outcode1 == 0 {
			// both points inside window; trivially accept
			accept = true
			break
		} else if outcode0&outcode1 != 0 {
			// both points share an outside zone (trivially reject)
			break
		} else {
			var x, y int
			outcodeOut := outcode1
			if outcode0 > outcode1 {
				outcodeOut = outcode0
			}

			// Find intersection point. Use float64 to prevent integer division truncation errors.
			if outcodeOut&TOP != 0 {
				x = x1 + int(float64(x2-x1)*float64(ymin-y1)/float64(y2-y1))
				y = ymin
			} else if outcodeOut&BOTTOM != 0 {
				x = x1 + int(float64(x2-x1)*float64(ymax-y1)/float64(y2-y1))
				y = ymax
			} else if outcodeOut&RIGHT != 0 {
				y = y1 + int(float64(y2-y1)*float64(xmax-x1)/float64(x2-x1))
				x = xmax
			} else if outcodeOut&LEFT != 0 {
				y = y1 + int(float64(y2-y1)*float64(xmin-x1)/float64(x2-x1))
				x = xmin
			}

			if outcodeOut == outcode0 {
				x1, y1 = x, y
				outcode0 = computeOutCode(x1, y1)
			} else {
				x2, y2 = x, y
				outcode1 = computeOutCode(x2, y2)
			}
		}
	}

	if !accept {
		return // Line is completely off-screen
	}

	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		c.SetPixel(x1, y1, color)
		if x1 == x2 && y1 == y2 {
			break
		}
		err2 := err * 2
		if err2 > -dy {
			err -= dy
			x1 += sx
		}
		if err2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (c *Con16) DrawRect(x1, y1, x2, y2 int, color byte) {
	// DrawLine is our most optimized drawing function, so we can use it to draw rectangles by drawing 4 lines
	c.DrawLine(x1, y1, x2, y1, color)
	c.DrawLine(x1, y2, x2, y2, color)
	c.DrawLine(x1, y1, x1, y2, color)
	c.DrawLine(x2, y1, x2, y2, color)
}

func (c *Con16) FillRect(x1, y1, x2, y2 int, color byte) {
	for y := y1; y <= y2; y++ {
		c.DrawLine(x1, y, x2, y, color)
	}
}

func (c *Con16) ClearScreen(color byte) {
	colorIndex := color & 0x3F
	for i := range c.frameBuffer32 {
		c.frameBuffer32[i] = c.colorMap[colorIndex]
	}
}

func (t *Tile) FillRect(x, y, w, h int, color byte) {
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			t.SetPixel(x+i, y+j, color)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
