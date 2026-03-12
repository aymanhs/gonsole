package gonsole

// ── PaletteColor ──────────────────────────────────────────────────────────────

// PaletteColor packs one 6-bit RGB color and 2 flag bits into a single byte:
//
//	bit 7    = transparent — this index is drawn fully transparent
//	bit 6    = alt flag    — reserved (e.g. priority, blending effect)
//	bits 5:4 = R  (0–3 → expands to 0, 85, 170, 255)
//	bits 3:2 = G  (0–3 → expands to 0, 85, 170, 255)
//	bits 1:0 = B  (0–3 → expands to 0, 85, 170, 255)
const (
	PaletteTransparent byte = 1 << 7 // draw this color as fully transparent
	PaletteAlt         byte = 1 << 6 // reserved alternate flag
)

type PaletteColor byte

// NewPaletteColor creates a PaletteColor from 2-bit R, G, B values (each 0–3).
func NewPaletteColor(r, g, b byte) PaletteColor {
	return PaletteColor((r&3)<<4 | (g&3)<<2 | (b & 3))
}

func (p PaletteColor) R() byte             { return byte((p >> 4) & 3) }
func (p PaletteColor) G() byte             { return byte((p >> 2) & 3) }
func (p PaletteColor) B() byte             { return byte(p & 3) }
func (p PaletteColor) IsTransparent() bool { return byte(p)&PaletteTransparent != 0 }

// ToRGB8 expands each 2-bit channel to 8-bit (0→0, 1→85, 2→170, 3→255).
func (p PaletteColor) ToRGB8() (r, g, b byte) {
	return p.R() * 85, p.G() * 85, p.B() * 85
}

// ── Palette ───────────────────────────────────────────────────────────────────

// Palette holds 16 PaletteColor entries for a sprite or tile layer.
type Palette struct {
	Colors [16]PaletteColor
}

// ── PaletteBank ───────────────────────────────────────────────────────────────

// PaletteBank holds 256 palettes × 16 colors, each pre-expanded to RGBA.
// Render pipeline lookup is a single array index: bank.Colors[bankID][colorIdx].
// Total size: 256 × 16 × 4 = 16 KB.
type PaletteBank struct {
	Colors [256][16][4]byte // [bankID][colorIdx] → {R, G, B, A}
}

// Set stores a raw RGBA entry directly.
func (pb *PaletteBank) Set(bankID, idx int, r, g, b, a byte) {
	pb.Colors[bankID][idx] = [4]byte{r, g, b, a}
}

// SetFrom stores a PaletteColor, expanding its 2-bit channels to 8-bit RGBA.
// Transparent colors get A=0; all others get A=255.
func (pb *PaletteBank) SetFrom(bankID, idx int, pc PaletteColor) {
	r, g, b := pc.ToRGB8()
	a := byte(255)
	if pc.IsTransparent() {
		a = 0
	}
	pb.Colors[bankID][idx] = [4]byte{r, g, b, a}
}

// RGBA returns a pointer to the pre-expanded [R,G,B,A] entry — zero allocation,
// suitable for direct use in the render pipeline.
func (pb *PaletteBank) RGBA(bankID, idx byte) *[4]byte {
	return &pb.Colors[bankID][idx]
}

// ── TileBank ──────────────────────────────────────────────────────────────────

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
