package gonsole

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Console struct {
	Framebuffer  [ScreenWidth * ScreenHeight]byte
	Palette      [16][3]byte
	SpriteData   [256][64]byte
	Sprites      [256]Sprite
	Buttons      byte
	MouseX       int
	MouseY       int
	MouseButtons byte
	OffsetX      int
	OffsetY      int
	Frame        uint64
	TimeMs       uint64
	UpdateFunc   func(frame, ms uint64) error
	DrawFunc     func(frame, ms uint64)
	screen       *ebiten.Image
	startTime    time.Time
}

func NewConsole() *Console {
	c := &Console{startTime: time.Now()}
	c.Palette = defaultPalette
	return c
}

func (c *Console) Update() error {
	c.Frame++
	c.TimeMs = uint64(time.Since(c.startTime).Milliseconds())
	c.pollInputs()
	c.pollMouse()
	if c.UpdateFunc != nil {
		return c.UpdateFunc(c.Frame, c.TimeMs)
	}
	return nil
}

func (c *Console) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func Run(c *Console) error {
	ebiten.SetWindowTitle("8-Bit Virtual Console")
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(c)
}
