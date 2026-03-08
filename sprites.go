package gonsole

const (
	SpritePropVisible     byte = 1 << 0 // draw this sprite
	SpritePropFlipH       byte = 1 << 1 // flip horizontally
	SpritePropFlipV       byte = 1 << 2 // flip vertically
	SpritePropScreenSpace byte = 1 << 3 // ignore camera offset (HUD, cursor, etc.)
)

type Sprite struct {
	X     int16
	Y     int16
	Props byte
	Flags byte
}

// SetSprite sets the sprite entry at index.
func (c *Console) SetSprite(index int, s Sprite) {
	if index < 0 || index >= 256 {
		return
	}
	c.Sprites[index] = s
}

// GetSprite returns the sprite entry at index.
func (c *Console) GetSprite(index int) Sprite {
	if index < 0 || index >= 256 {
		return Sprite{}
	}
	return c.Sprites[index]
}

// SetSpriteData uploads 64 bytes of pixel data for sprite slot id (8x8, row-major).
func (c *Console) SetSpriteData(id int, data []byte) {
	if id < 0 || id >= 256 || len(data) != 64 {
		return
	}
	copy(c.SpriteData[id][:], data)
}
