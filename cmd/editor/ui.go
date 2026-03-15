package main

import (
	"aymanhs/gonsole"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ToggleButton is a generic true/false boolean toggle component
type ToggleButton struct {
	X, Y          int
	Width, Height int
	Text          string
	Value         bool
}

func (t *ToggleButton) HandleInput(con *gonsole.Con16) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if mx >= t.X && mx < t.X+t.Width && my >= t.Y && my < t.Y+t.Height {
			t.Value = !t.Value
		}
	}
}

func (t *ToggleButton) Draw(con *gonsole.Con16) {
	// Draw the toggle box
	con.DrawRect(t.X, t.Y+2, t.X+12, t.Y+2+12, 63) // 63 is white in the default colorMap
	// Draw the filled indicator if true
	if t.Value {
		con.FillRect(t.X+3, t.Y+5, t.X+3+6, t.Y+5+6, 63)
	}
	// Draw the label
	con.DisplayText(t.X+16, t.Y, t.Text, 63)
}

// RadioGroup is a generic single-selection component from a list of strings
type RadioGroup struct {
	X, Y          int
	Width, Height int // Size of each item's click area
	Options       []string
	SelectedIndex int
	IsVertical    bool
}

func (r *RadioGroup) HandleInput(con *gonsole.Con16) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		for i := range r.Options {
			itemX, itemY := r.X, r.Y
			if r.IsVertical {
				itemY += i * r.Height
			} else {
				itemX += i * r.Width
			}

			// Check if click is inside this option's bounding box
			if mx >= itemX && mx < itemX+r.Width && my >= itemY && my < itemY+r.Height {
				r.SelectedIndex = i
				break
			}
		}
	}
}

func (r *RadioGroup) Draw(con *gonsole.Con16) {
	for i, opt := range r.Options {
		itemX, itemY := r.X, r.Y
		if r.IsVertical {
			itemY += i * r.Height
		} else {
			itemX += i * r.Width
		}

		// Draw the radio box
		con.DrawRect(itemX, itemY+2, itemX+12, itemY+2+12, 63)
		// Draw the filled indicator if this option is selected
		if r.SelectedIndex == i {
			con.FillRect(itemX+3, itemY+5, itemX+3+6, itemY+5+6, 63)
		}

		// Draw the option text
		con.DisplayText(itemX+16, itemY, opt, 63)
	}
}
