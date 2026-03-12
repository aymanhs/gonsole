package gonsole

// TileBank stores up to 256 tiles, each 8×8 pixels with 4-bit (0–15) color indices.
// Two pixels are packed per byte: high nibble = left pixel (even x), low nibble = right pixel (odd x).
// Each tile is therefore 32 bytes (8 rows × 4 bytes/row).
type TileBank struct {
	Tiles [256][32]byte
}

// TileGetPixel returns the 4-bit color index at pixel (x, y) within a tile.
func (tb *TileBank) TileGetPixel(tile, x, y int) byte {
	b := tb.Tiles[tile][y*4+x/2]
	if x&1 == 0 {
		return (b >> 4) & 0xF
	}
	return b & 0xF
}

// TileSetPixel sets the 4-bit color index at pixel (x, y) within a tile.
func (tb *TileBank) TileSetPixel(tile, x, y int, idx byte) {
	i := y*4 + x/2
	if x&1 == 0 {
		tb.Tiles[tile][i] = (tb.Tiles[tile][i] & 0x0F) | (idx << 4)
	} else {
		tb.Tiles[tile][i] = (tb.Tiles[tile][i] & 0xF0) | (idx & 0xF)
	}
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

// BlitTile copies an 8×8 tile from a TileBank onto the scratch buffer.
// Pixels with alpha=0 in the palette are skipped (transparent).
func BlitTile(bank *TileBank, tileID, sx, sy int, pal *[16][4]byte, scratch *[ScreenWidth * ScreenHeight * 4]byte) {
	tile := &bank.Tiles[tileID]
	for row := 0; row < 8; row++ {
		dy := sy + row
		if dy < 0 || dy >= ScreenHeight {
			continue
		}
		dstRow := dy * ScreenWidth
		for col := 0; col < 8; col++ {
			dx := sx + col
			if dx < 0 || dx >= ScreenWidth {
				continue
			}
			// Unpack nibble: high = even col, low = odd col
			b := tile[row*4+col/2]
			var nibble byte
			if col&1 == 0 {
				nibble = (b >> 4) & 0xF
			} else {
				nibble = b & 0xF
			}
			rgba := &pal[nibble]
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
