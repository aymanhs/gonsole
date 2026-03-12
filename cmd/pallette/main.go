package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// 2-bit RGB palette: 4×4×4 = 64 colors in an 8×8 grid.
//
// Each axis uses a 3-bit reflected Gray code so every horizontal and vertical
// neighbor differs by exactly ONE bit in the 6-bit RGB value.
//
// X (3-bit Gray code, 8 cols) → even bit positions  0,2,4
// Y (3-bit Gray code, 8 rows) → odd  bit positions  1,3,5
//
// Final 6-bit index → R = bits[5:4], G = bits[3:2], B = bits[1:0]

const (
	cols    = 8  // swatches per row
	rows    = 8  // rows
	swatchW = 80 // pixels wide per swatch
	swatchH = 80 // pixels tall per swatch
	screenW = cols * swatchW
	screenH = rows * swatchH
)

func gray(n int) int      { return n ^ (n >> 1) }
func channel(v int) uint8 { return uint8(v * 85) } // 0→0, 1→85, 2→170, 3→255

func swatchColor(col, row int) color.RGBA {
	gx := gray(col) // 3-bit Gray code
	gy := gray(row) // 3-bit Gray code

	// Interleave: x-bits → even positions, y-bits → odd positions
	idx := 0
	for bit := 0; bit < 3; bit++ {
		idx |= ((gx >> bit) & 1) << (bit * 2)
		idx |= ((gy >> bit) & 1) << (bit*2 + 1)
	}

	r := (idx >> 4) & 3
	g := (idx >> 2) & 3
	b := idx & 3
	return color.RGBA{R: channel(r), G: channel(g), B: channel(b), A: 255}
}

type Game struct{ img *ebiten.Image }

func NewGame() *Game {
	img := ebiten.NewImage(screenW, screenH)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			swatch := ebiten.NewImage(swatchW, swatchH)
			swatch.Fill(swatchColor(col, row))
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*swatchW), float64(row*swatchH))
			img.DrawImage(swatch, op)
		}
	}
	return &Game{img: img}
}

func (g *Game) Update() error              { return nil }
func (g *Game) Draw(screen *ebiten.Image)  { screen.DrawImage(g.img, nil) }
func (g *Game) Layout(_, _ int) (int, int) { return screenW, screenH }

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("2-bit RGB – Gray code layout (1-bit neighbors)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
