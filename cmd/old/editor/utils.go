package main

import "aymanhs/gonsole"

// manip16x16 is a helper for 16x16 pixel manipulation
type manip16x16 [16][16]byte

// manip8x8 is a helper for 8x8 pixel manipulation (Fonts)
type manip8x8 [8][8]byte

func (m *manip16x16) shift(dx, dy int) {
	var next manip16x16
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			nx := (x + dx + 16) % 16
			ny := (y + dy + 16) % 16
			next[ny][nx] = m[y][x]
		}
	}
	*m = next
}

func (m *manip16x16) flipH() {
	for y := 0; y < 16; y++ {
		for x := 0; x < 8; x++ {
			m[y][x], m[y][15-x] = m[y][15-x], m[y][x]
		}
	}
}

func (m *manip16x16) flipV() {
	for x := 0; x < 16; x++ {
		for y := 0; y < 8; y++ {
			m[y][x], m[15-y][x] = m[15-y][x], m[y][x]
		}
	}
}

func (m *manip16x16) rotate() {
	var next manip16x16
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			next[x][15-y] = m[y][x]
		}
	}
	*m = next
}

// Unpack Helpers

func unpackNibble(src [256]byte) (m manip16x16) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			m[y][x] = src[y*16+x]
		}
	}
	return
}

func packNibble(m manip16x16) (dst [256]byte) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			dst[y*16+x] = m[y][x]
		}
	}
	return
}

func unpackBit(src [8]byte) (m manip8x8) {
	for y := 0; y < 8; y++ {
		b := src[y]
		for x := 0; x < 8; x++ {
			if (b >> (7 - x)) & 1 != 0 {
				m[y][x] = 1
			} else {
				m[y][x] = 0
			}
		}
	}
	return
}

func packBit(m manip8x8) (dst [8]byte) {
	for y := 0; y < 8; y++ {
		var b byte
		for x := 0; x < 8; x++ {
			if m[y][x] != 0 {
				b |= (1 << (7 - x))
			}
		}
		dst[y] = b
	}
	return
}

func drawBlendIndicator(c *gonsole.Console, x, y, w, h int, blend byte) {
switch blend {
case 1: // Subtract (lower right)
c.DrawRect(x+w-5, y+h-5, 5, 5, 0, true)
c.DrawRect(x+w-4, y+h-4, 3, 3, 7, true)
c.SetPixel(x+w-3, y+h-3, 0)
case 2: // Add (top right)
c.DrawRect(x+w-5, y, 5, 5, 0, true)
c.DrawRect(x+w-4, y+1, 3, 3, 7, true)
c.SetPixel(x+w-3, y+2, 0)
case 3: // Transparent (X)
c.DrawLine(x, y, x+w-1, y+h-1, 7)
c.DrawLine(x+w-1, y, x, y+h-1, 7)
}
}
