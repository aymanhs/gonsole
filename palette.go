package gonsole

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

// Palette holds 16 PaletteColor entries for a sprite or tile layer.
type Palette struct {
	Colors [16]PaletteColor
}

// PaletteBank holds 4 palettes × 16 colors, each pre-expanded to RGBA.
// Render pipeline lookup is a single array index: bank.Colors[bankID][colorIdx].
// Total size: 4 × 16 × 4 = 256 bytes.
type PaletteBank struct {
	Colors [4][16][4]byte // [bankID][colorIdx] → {R, G, B, A}
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

// defaultPalette is a PICO-8 inspired 16-color palette.
// Index 0 is reserved as transparent/black.
var defaultPalette = [16][3]byte{
	0:  {0, 0, 0},       // 0  black      (transparent)
	1:  {29, 43, 83},    // 1  dark blue
	2:  {126, 37, 83},   // 2  dark purple
	3:  {0, 135, 81},    // 3  dark green
	4:  {171, 82, 54},   // 4  brown
	5:  {95, 87, 79},    // 5  dark grey
	6:  {194, 195, 199}, // 6  light grey
	7:  {255, 241, 232}, // 7  white
	8:  {255, 0, 77},    // 8  red
	9:  {255, 163, 0},   // 9  orange
	10: {255, 236, 39},  // 10 yellow
	11: {0, 228, 54},    // 11 green
	12: {41, 173, 255},  // 12 blue
	13: {131, 118, 156}, // 13 lavender
	14: {255, 119, 168}, // 14 pink
	15: {255, 204, 170}, // 15 peach
}

// SetPalette updates a direct palette entry and syncs it to PaletteBank slot 0.
func (c *Console) SetPalette(index int, r, g, b byte) {
	if index < 0 || index >= 16 {
		return
	}
	c.Palette[index] = [3]byte{r, g, b}
	a := byte(255)
	if index == 0 {
		a = 0
	}
	c.PaletteBank.Colors[0][index] = [4]byte{r, g, b, a}
}
