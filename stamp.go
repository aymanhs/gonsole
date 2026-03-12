package gonsole

// ── Stamp props ───────────────────────────────────────────────────────────────

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
	PalletID  byte
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

// SetSpriteData uploads 64 bytes of pixel data for sprite slot id (8×8, row-major).
func (c *Console) SetSpriteData(id int, data []byte) {
	if id < 0 || id >= 256 || len(data) != 64 {
		return
	}
	copy(c.SpriteData[id][:], data)
}
