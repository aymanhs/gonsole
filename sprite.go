package gonsole

import "unsafe"

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

// SetSpriteData uploads 256 bytes of pixel data for sprite slot id (16×16, row-major).
func (c *Console) SetSpriteData(id int, data []byte) {
	if id < 0 || id >= 256 || len(data) != 256 {
		return
	}
	copy(c.SpriteData[id][:], data)
}

// BlitSprite copies an 16×16 sprite from legacy SpriteData onto the scratch buffer.
// Pixels with alpha=0 in the palette are skipped.
func BlitSprite(data []byte, sx, sy int, props byte, pal *[16][4]byte, scratch *[ScreenWidth * ScreenHeight * 4]byte) {
	palU32 := (*[16]uint32)(unsafe.Pointer(pal))

	for row := 0; row < TileSize; row++ {
		srcRow := row
		if props&StampPropFlipV != 0 {
			srcRow = (TileSize - 1) - row
		}
		dy := sy + row
		if dy < 0 || dy >= ScreenHeight {
			continue
		}
		dstRow := dy * ScreenWidth
		srcBase := srcRow * TileSize

		for col := 0; col < TileSize; col++ {
			srcCol := col
			if props&StampPropFlipH != 0 {
				srcCol = (TileSize - 1) - col
			}
			dx := sx + col
			if dx < 0 || dx >= ScreenWidth {
				continue
			}

			colorIdx := data[srcBase+srcCol]
			if colorIdx == 0 {
				continue
			}

			rgba := &pal[colorIdx]
			dst := (dstRow + dx) * 4

			if rgba[3] == BlendNormal {
				*(*uint32)(unsafe.Pointer(&scratch[dst])) = palU32[colorIdx]
			} else {
				ApplyBlend(scratch, dst, rgba)
			}
		}
	}
}
