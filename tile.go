package gonsole

import "unsafe"

// TileBank stores up to 256 tiles, each 16×16 pixels with 1-byte color indices.
// Each tile is therefore 256 bytes (16 rows × 16 bytes/row).
type TileBank struct {
	Tiles [256][256]byte
}

// TileGetPixel returns the 4-bit color index at pixel (x, y) within a tile.
func (tb *TileBank) TileGetPixel(tile, x, y int) byte {
	return tb.Tiles[tile][y*TileSize+x]
}

// TileSetPixel sets the 4-bit color index at pixel (x, y) within a tile.
func (tb *TileBank) TileSetPixel(tile, x, y int, idx byte) {
	tb.Tiles[tile][y*TileSize+x] = idx
}

const (
	TileLayerCount = 4    // fixed number of tile layers
	TileSlotCount  = 1024 // slots per tile layer
)

// TileSlot is one entry in a tile layer's sparse array.
// WorldX/WorldY are in world-space pixels (uint16).
// BankID selects which TileBank to pull the 8×8 tile data from.
// TileID selects the tile within that bank.
// PaletteID selects the palette from the Console's PaletteBank.
type TileSlot struct {
	WorldX    uint16
	WorldY    uint16
	TileID    byte
	BankID    byte
	PaletteID byte
	_         [3]byte // padding
}

// TileLayer is a sparse list of up to TileSlotCount tile stamps.
// ParallaxMul/Div control how much the camera moves this layer:
//
//	effectiveCameraX = cameraX * ParallaxMul / ParallaxDiv
//
// Use Mul=1, Div=1 for normal (full-speed) scrolling.
// Use Mul=1, Div=2 for half-speed (background parallax).
// Use Mul=0, Div=1 for static (fixed to screen).
type TileLayer struct {
	Slots       [TileSlotCount]TileSlot
	Count       int // number of active slots (slots[0..Count-1])
	ParallaxMul int
	ParallaxDiv int
}

// NewTileLayer returns a TileLayer with 1:1 parallax (full-speed scroll).
func NewTileLayer() TileLayer {
	return TileLayer{ParallaxMul: 1, ParallaxDiv: 1}
}

// AddTile appends a tile slot. Returns the slot index, or -1 if full.
func (tl *TileLayer) AddTile(worldX, worldY uint16, tileID, bankID, paletteID byte) int {
	if tl.Count >= TileSlotCount {
		return -1
	}
	i := tl.Count
	tl.Slots[i] = TileSlot{
		WorldX:    worldX,
		WorldY:    worldY,
		TileID:    tileID,
		BankID:    bankID,
		PaletteID: paletteID,
	}
	tl.Count++
	return i
}

// Clear resets the layer to zero active slots.
func (tl *TileLayer) Clear() { tl.Count = 0 }

// BlitTile copies an 16×16 tile from a TileBank onto the scratch buffer.
// Pixels with alpha=0 in the palette are skipped (transparent).
func BlitTile(bank *TileBank, tileID, sx, sy int, pal *[16][4]byte, scratch *[ScreenWidth * ScreenHeight * 4]byte) {
	tile := &bank.Tiles[tileID]
	// Speed up palette lookup by casting to uint32 pointers for BlendNormal case
	palU32 := (*[16]uint32)(unsafe.Pointer(pal))

	for row := 0; row < TileSize; row++ {
		dy := sy + row
		if dy < 0 || dy >= ScreenHeight {
			continue
		}
		dstRow := dy * ScreenWidth
		srcBase := row * TileSize

		for col := 0; col < TileSize; col++ {
			dx := sx + col
			if dx < 0 || dx >= ScreenWidth {
				continue
			}

			colorIdx := tile[srcBase+col]
			if colorIdx == 0 {
				continue // 0 is always transparent
			}

			rgba := &pal[colorIdx]
			dst := (dstRow + dx) * 4

			if rgba[3] == BlendNormal {
				// Fast path: direct uint32 copy
				*(*uint32)(unsafe.Pointer(&scratch[dst])) = palU32[colorIdx]
			} else {
				// Slow path: manual alpha blending
				ApplyBlend(scratch, dst, rgba)
			}
		}
	}
}
