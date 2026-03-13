package main

// manip8x8 is a helper for 8x8 pixel manipulation
type manip8x8 [8][8]byte

func (m *manip8x8) shift(dx, dy int) {
	var next manip8x8
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			nx := (x + dx + 8) % 8
			ny := (y + dy + 8) % 8
			next[ny][nx] = m[y][x]
		}
	}
	*m = next
}

func (m *manip8x8) flipH() {
	for y := 0; y < 8; y++ {
		for x := 0; x < 4; x++ {
			m[y][x], m[y][7-x] = m[y][7-x], m[y][x]
		}
	}
}

func (m *manip8x8) flipV() {
	for x := 0; x < 8; x++ {
		for y := 0; y < 4; y++ {
			m[y][x], m[7-y][x] = m[7-y][x], m[y][x]
		}
	}
}

func (m *manip8x8) rotate() {
	var next manip8x8
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			next[x][7-y] = m[y][x]
		}
	}
	*m = next
}

// Unpack Helpers

func unpackNibble(src [32]byte) (m manip8x8) {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			b := src[y*4+x/2]
			if x&1 == 0 {
				m[y][x] = (b >> 4) & 0xF
			} else {
				m[y][x] = b & 0xF
			}
		}
	}
	return
}

func packNibble(m manip8x8) (dst [32]byte) {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			i := y*4 + x/2
			if x&1 == 0 {
				dst[i] = (dst[i] & 0x0F) | (m[y][x] << 4)
			} else {
				dst[i] = (dst[i] & 0xF0) | (m[y][x] & 0xF)
			}
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
