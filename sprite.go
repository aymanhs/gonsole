package gonsole

const (
	StampPropVisible     byte = 1 << 0 // draw this stamp
	StampPropFlipH       byte = 1 << 1 // flip horizontally
	StampPropFlipV       byte = 1 << 2 // flip vertically
	StampPropScreenSpace byte = 1 << 3 // ignore camera offset (HUD, cursor, etc.)
	StampPropTilemap     byte = 1 << 4 // use tile data as a tilemap
)

// Stamp is a positioned rendering instance used for both sprites and tilemap stamps.
// X, Y are world-space coordinates (uint16 — world origin is 0,0; scroll by moving camera).
// DrawLayer controls compositing order (0–7); see Console draw order docs.
type Stamp struct {
	X         uint16
	Y         uint16
	Props     byte
	Flags     byte
	BankID    byte
	TileID    byte
	PaletteID byte
	DrawLayer byte // 0–7: interleaved with tile layers
}

// SetStamp sets the stamp entry at index.
func (c *Console) SetStamp(index int, s Stamp) {
	if index < 0 || index >= 256 {
		return
	}
	c.Stamps[index] = s
}

// GetStamp returns the stamp entry at index.
func (c *Console) GetStamp(index int) Stamp {
	if index < 0 || index >= 256 {
		return Stamp{}
	}
	return c.Stamps[index]
}

// SetSpriteData uploads 64 bytes of pixel data for sprite slot id (8×8, row-major)
// and packs it into 32 bytes (4 bits per pixel).
func (c *Console) SetSpriteData(id int, data []byte) {
	if id < 0 || id >= 256 || len(data) != 64 {
		return
	}
	for i := 0; i < 32; i++ {
		// Pack two pixels into one byte
		hi := data[i*2] & 0xF
		lo := data[i*2+1] & 0xF
		c.SpriteData[id][i] = (hi << 4) | lo
	}
}

// BlitSprite copies an 8×8 sprite from legacy SpriteData onto the scratch buffer.
// Pixels with alpha=0 in the palette are skipped.
func BlitSprite(data []byte, sx, sy int, props byte, pal *[16][4]byte, scratch *[ScreenWidth * ScreenHeight * 4]byte) {
	for row := 0; row < 8; row++ {
		srcRow := row
		if props&StampPropFlipV != 0 {
			srcRow = 7 - row
		}
		dy := sy + row
		if dy < 0 || dy >= ScreenHeight {
			continue
		}
		dstRow := dy * ScreenWidth
		for col := 0; col < 8; col++ {
			srcCol := col
			if props&StampPropFlipH != 0 {
				srcCol = 7 - col
			}
			dx := sx + col
			if dx < 0 || dx >= ScreenWidth {
				continue
			}
			// Unpack nibble: high = even col, low = odd col
			b := data[srcRow*4+srcCol/2]
			var colorIdx byte
			if srcCol&1 == 0 {
				colorIdx = (b >> 4) & 0xF
			} else {
				colorIdx = b & 0xF
			}
			rgba := &pal[colorIdx]
			if rgba[3] == 0 {
				continue
			}
			dst := (dstRow + dx) * 4
			scratch[dst] = rgba[0]
			scratch[dst+1] = rgba[1]
			scratch[dst+2] = rgba[2]
			scratch[dst+3] = rgba[3]
		}
	}
}
