package gonsole

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// KeyPressed returns true if the specified key is currently pressed.
func KeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

// KeyJustPressed returns true if the specified key was just pressed in the current frame.
func KeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

// KeyJustReleased returns true if the specified key was just released in the current frame.
func KeyJustReleased(key ebiten.Key) bool {
	return inpututil.IsKeyJustReleased(key)
}

// MouseButtonPressed returns true if the specified mouse button is currently pressed.
func MouseButtonPressed(button ebiten.MouseButton) bool {
	return ebiten.IsMouseButtonPressed(button)
}

// MouseButtonJustPressed returns true if the specified mouse button was just pressed in the current frame.
func MouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

// MouseButtonJustReleased returns true if the specified mouse button was just released in the current frame.
func MouseButtonJustReleased(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustReleased(button)
}

// CursorPosition returns the current cursor (mouse) position coordinates.
func CursorPosition() (int, int) {
	return ebiten.CursorPosition()
}

// Wheel returns the x and y offsets of the mouse wheel scrolling.
func Wheel() (float64, float64) {
	return ebiten.Wheel()
}
