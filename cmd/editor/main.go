package main

import (
	"aymanhs/gonsole"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var selectedColor byte = 0

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
			con.GetTile(t.TileIndex).SetPixel(col, row, selectedColor)
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
	X int
	Y int
}

func (p *PaletteDisplay) HandleInput(con *gonsole.Con16) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		if mouseX >= p.X && mouseX < p.X+(16*5) && mouseY >= p.Y && mouseY < p.Y+(4*5) {
			col := (mouseX - p.X) / 5
			row := (mouseY - p.Y) / 5
			selectedColor = byte(row*16 + col)
		}
	}
}

func (p *PaletteDisplay) Draw(con *gonsole.Con16) {
	sx := p.X
	sy := p.Y
	s := 4
	for i := 0; i < 64; i++ {
		con.FillRect(sx, sy, sx+s-1, sy+s-1, byte(i))
		if byte(i) == selectedColor {
			con.DrawRect(sx-1, sy-1, sx+s, sy+s, 63)
		}

		sx += s + 1
		if (i+1)%16 == 0 {
			sx = p.X
			sy += s + 1
		}
	}
}

func main() {
	c := gonsole.NewCon16(640, 480)
	c.GetTile(0).Fill(0b01000001)

	components := []EditorComponent{
		&TileDisplay{X: 20, Y: 20, TileIndex: 0},
		&PaletteDisplay{X: 20, Y: 120},
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
