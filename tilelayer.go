package gonsole

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
