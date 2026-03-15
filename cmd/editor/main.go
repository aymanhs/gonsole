package main

import (
	"aymanhs/gonsole"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// EditorComponent defines a simple interface for UI objects that manage their own input and rendering.
type EditorComponent interface {
	HandleInput(con *gonsole.Con16)
	Draw(con *gonsole.Con16)
}

// TileDisplay handles drawing and editing a magnified tile.
type TileDisplay struct {
	X         int
	Y         int
	TileIndex byte
}

func (t *TileDisplay) HandleInput(con *gonsole.Con16) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		if mouseX >= t.X && mouseX < t.X+(16*5) && mouseY >= t.Y && mouseY < t.Y+(16*5) {
			col := (mouseX - t.X) / 5
			row := (mouseY - t.Y) / 5
			con.GetTile(t.TileIndex).SetPixel(col, row, palette.selectedColor)
		}
	}
}

func (t *TileDisplay) Draw(con *gonsole.Con16) {
	sx := t.X
	sy := t.Y
	s := 4
	for r := 0; r < 16; r++ {
		for c := 0; c < 16; c++ {
			colorIndex := con.GetTile(t.TileIndex).GetPixel(c, r)
			con.FillRect(sx, sy, sx+s-1, sy+s-1, colorIndex)
			sx += s + 1
		}
		sx = t.X
		sy += s + 1
	}
}

// PaletteDisplay handles drawing the color grid and selecting colors.
type PaletteDisplay struct {
	X             int
	Y             int
	selectedColor byte
}

var palette = &PaletteDisplay{X: 20, Y: 120}

func (p *PaletteDisplay) HandleInput(con *gonsole.Con16) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		if mouseX >= p.X && mouseX < p.X+(16*5) && mouseY >= p.Y && mouseY < p.Y+(4*5) {
			col := (mouseX - p.X) / 5
			row := (mouseY - p.Y) / 5
			// save the selected color index in the lower 6 bits, and keep the blend mode in the upper 2 bits
			p.selectedColor = (p.selectedColor & 0xC0) | byte(row*16+col)
		}
		// check if one of the blend mode options was clicked
		if mouseX >= p.X && mouseX < p.X+(20*4) && mouseY >= p.Y+25 && mouseY < p.Y+25+10 {
			col := (mouseX - p.X) / 21 // 20 for the box + 1 for spacing
			if col < 4 {
				p.selectedColor = (p.selectedColor & 0x3F) | (byte(col) << 6)
			}
		}
	}
}

func (p *PaletteDisplay) Draw(con *gonsole.Con16) {
	sx := p.X
	sy := p.Y
	s := 4
	for i := 0; i < 64; i++ {
		con.FillRect(sx, sy, sx+s-1, sy+s-1, byte(i))
		if byte(i) == (p.selectedColor & 0x3F) {
			con.DrawRect(sx-1, sy-1, sx+s, sy+s, 63)
		}

		sx += s + 1
		if (i+1)%16 == 0 {
			sx = p.X
			sy += s + 1
		}
	}
	// draw the blend mode options below the palette
	sx = p.X
	sy = p.Y + 25
	mode := int(p.selectedColor>>6) & 0x03
	for i := 0; i < 4; i++ {
		con.FillRect(sx, sy, sx+15, sy+10, byte(64+i))
		if i == mode {
			con.DrawRect(sx-1, sy-1, sx+15, sy+10, 63)
		}
		sx += 20 + 1
	}
}

func main() {
	c := gonsole.NewCon16(640, 480)
	c.GetTile(0).Fill(0b01000001)

	components := []EditorComponent{
		&TileDisplay{X: 20, Y: 20, TileIndex: 0},
		palette,
		&ToggleButton{X: 250, Y: 120, Width: 100, Height: 20, Text: "Show Grid", Value: false},
		&RadioGroup{
			X: 250, Y: 160,
			Width: 100, Height: 20,
			Options:       []string{"Pencil tool", "Fill tool", "Select tool"},
			SelectedIndex: 0,
			IsVertical:    true,
		},
		&RadioGroup{
			X: 250, Y: 240,
			Width: 60, Height: 20,
			Options:       []string{"1x", "2x", "4x"},
			SelectedIndex: 0,
			IsVertical:    false,
		},
	}

	c.SetUpdateFunc(func(frame, ms int) error {
		for _, comp := range components {
			comp.HandleInput(c)
		}
		return nil
	})

	c.SetDrawFunc(func(slot int, frame int) {
		c.ClearScreen(0)
		for _, comp := range components {
			comp.Draw(c)
		}
		// put the tile in actual size for reference
		c.DrawTile(200, 20, 0)
	})

	c.Run()
}
