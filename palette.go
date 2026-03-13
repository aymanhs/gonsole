package gonsole

// PaletteColor packs one 6-bit RGB color and 2 flag bits into a single byte:
//
//	bit 7    = transparent — this index is drawn fully transparent
//	bit 6    = alt flag    — reserved (e.g. priority, blending effect)
//	bits 5:4 = R  (0–3 → expands to 0, 85, 170, 255)
//	bits 3:2 = G  (0–3 → expands to 0, 85, 170, 255)
//	bits 1:0 = B  (0–3 → expands to 0, 85, 170, 255)
const (
	PaletteTransparent byte = 1 << 7 // (Legacy flag) draw this color as fully transparent
	PaletteAlt         byte = 1 << 6 // reserved alternate flag
)

// Blend mode constants interpreted during software rendering.
// They are stored in the 4th (alpha) channel of the internal [4]byte palette.
const (
	BlendNormal      byte = 0 // Overwrite destination
	BlendSubtract    byte = 1 // Darken destination (clamp at 0)
	BlendAdd         byte = 2 // Lighten destination (clamp at 255)
	BlendTransparent byte = 3 // Skip rendering
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
// Uses the transparent flag to map to the new BlendTransparent (3) mode, otherwise normal (0).
func (pb *PaletteBank) SetFrom(bankID, idx int, pc PaletteColor) {
	r, g, b := pc.ToRGB8()
	a := BlendNormal
	if pc.IsTransparent() {
		a = BlendTransparent
	}
	pb.Colors[bankID][idx] = [4]byte{r, g, b, a}
}

// RGBA returns a pointer to the pre-expanded [R,G,B,A] entry — zero allocation,
// suitable for direct use in the render pipeline.
func (pb *PaletteBank) RGBA(bankID, idx byte) *[4]byte {
	return &pb.Colors[bankID][idx]
}

// defaultPalette is a PICO-8 inspired 16-color palette.
// Colors are normalized to the 64-color range (each channel must be 0, 85, 170, or 255).
var defaultPalette = [16][3]byte{
	0:  {0, 0, 0},       // 0  black      (now standard black, default transparent is blend mode 3)
	1:  {0, 85, 85},     // 1  dark blue
	2:  {85, 0, 85},     // 2  dark purple
	3:  {0, 170, 85},    // 3  dark green
	4:  {170, 85, 85},   // 4  brown
	5:  {85, 85, 85},    // 5  dark grey
	6:  {170, 170, 170}, // 6  light grey
	7:  {255, 255, 255}, // 7  white
	8:  {255, 0, 85},    // 8  red
	9:  {255, 170, 0},   // 9  orange
	10: {255, 255, 0},   // 10 yellow
	11: {0, 255, 85},    // 11 green
	12: {0, 170, 255},   // 12 blue
	13: {170, 85, 170},  // 13 lavender
	14: {255, 85, 170},  // 14 pink
	15: {255, 170, 170}, // 15 peach
}

// SetPalette updates a direct palette entry (Bank 0) using 8-bit RGB values.
// Note: It clears alpha/blend mode back to BlendNormal (0).
func (c *Console) SetPalette(index int, r, g, b byte) {
	if index < 0 || index >= 16 {
		return
	}
	c.Palette[index] = [3]byte{r, g, b}
	// Blend mode is 0 (BlendNormal) by default
	c.PaletteBank.Colors[0][index] = [4]byte{r, g, b, 0}
}

// SetBankPalette updates a palette entry in a specific bank using 2-bit (0-3) RGB values.
// By default, assigning via SetBankPalette resets the blend mode to 0 (BlendNormal).
func (c *Console) SetBankPalette(bank, index int, r, g, b byte) {
	if bank < 0 || bank >= 4 || index < 0 || index >= 16 {
		return
	}

	// Convert 2-bit to 8-bit (0-255)
	r8, g8, b8 := (r&3)*85, (g&3)*85, (b&3)*85

	// Default to BlendNormal (0)
	c.PaletteBank.Colors[bank][index] = [4]byte{r8, g8, b8, 0}

	// If bank 0, also update the main Palette array (as 8-bit)
	if bank == 0 {
		c.Palette[index] = [3]byte{r8, g8, b8}
	}
}
