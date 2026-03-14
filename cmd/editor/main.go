package main

import (
	"log"

	"aymanhs/gonsole"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type EditorState struct {
	mode           int // 0: Palette, 1: Pixel, 2: Tilemap, 3: Font
	paletteEditor  PaletteEditor
	pixelEditor    PixelEditor
	tilemapEditor  TilemapEditor
	fontEditor     FontEditor
	clipboard      manip16x16
	colorClipboard [4]byte
	bankClipboard  [16][4]byte
}

func main() {
	c := gonsole.NewConsole()

	state := &EditorState{
		mode:        0,
		pixelEditor: PixelEditor{zoom: 16, targetType: 0}, // default to sprite
		fontEditor:  FontEditor{selectedChar: 'A'},
	}

	var statusMsg string
	var statusTimer int

	c.UpdateFunc = func(frame, ms uint64) error {
		// Switch modes with number keys
		if ebiten.IsKeyPressed(ebiten.Key1) {
			state.mode = 0
		}
		if ebiten.IsKeyPressed(ebiten.Key2) {
			state.mode = 1
		}
		if ebiten.IsKeyPressed(ebiten.Key3) {
			state.mode = 2
		}
		if ebiten.IsKeyPressed(ebiten.Key4) {
			state.mode = 3
		}

		// Save/Load
		if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyS) {
			if err := c.SaveJSON("cart.gon"); err == nil {
				statusMsg = "Saved to cart.gon"
				statusTimer = 120
			} else {
				statusMsg = "ERROR: " + err.Error()
				statusTimer = 120
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyL) {
			if err := c.LoadJSON("cart.gon"); err == nil {
				statusMsg = "Loaded from cart.gon"
				statusTimer = 120
			} else {
				statusMsg = "ERROR: " + err.Error()
				statusTimer = 120
			}
		}

		if statusTimer > 0 {
			statusTimer--
		}

		switch state.mode {
		case 0:
			state.paletteEditor.Update(c, &state.colorClipboard, &state.bankClipboard)
		case 1:
			state.pixelEditor.Update(c, 20, 40, &state.clipboard)
		case 2:
			state.tilemapEditor.Update(c)
		case 3:
			state.fontEditor.Update(c, 20, 40, &state.clipboard)
		}

		return nil
	}

	c.PaintFunc = func(slot int, frame uint64) {
		if slot == gonsole.PaintSlotEnd {
			switch state.mode {
			case 0:
				state.paletteEditor.Draw(c)
			case 1:
				state.pixelEditor.Draw(c, 20, 40)
			case 2:
				state.tilemapEditor.Draw(c)
			case 3:
				state.fontEditor.Draw(c, 20, 40)
			}
		}
	}

	c.OverlayFunc = func(screen *ebiten.Image) {
		// HUD (on high-res screen)
		ebitenutil.DebugPrintAt(screen, "[1] PAL [2] TILE [3] MAP [4] FONT | ESC Exit", 150, 3)
		ebitenutil.DebugPrintAt(screen, "CTRL+S Save  CTRL+L Load", 470, 3)

		if statusTimer > 0 {
			ebitenutil.DebugPrintAt(screen, "STATUS: "+statusMsg, 160, 220)
		}

		switch state.mode {
		case 0:
			state.paletteEditor.DrawOverlay(screen)
		}
	}

	if err := gonsole.Run(c); err != nil {
		log.Fatal(err)
	}
}
