package gonsole

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// Cartridge represents the serialized state of a gonsole project.
type Cartridge struct {
	Version    int          `json:"version"`
	Palettes   []string     `json:"palettes"`    // Hex strings of [16][4]byte (pre-expanded)
	TileBanks  []string     `json:"tile_banks"`  // Hex strings of [256][32]byte
	SpriteData string       `json:"sprite_data"` // Hex string of [256][32]byte
	TileLayers []TileLayer  `json:"tile_layers"`
	FontData   string       `json:"font_data"` // Hex string of [128][8]byte
}

// SaveJSON saves the console state to a JSON file with hex-encoded buffers.
func (c *Console) SaveJSON(path string) error {
	cart := Cartridge{
		Version: 1,
	}

	// Encode PaletteBanks
	for i := 0; i < 4; i++ {
		data := flattenPalette(c.PaletteBank.Colors[i])
		cart.Palettes = append(cart.Palettes, hex.EncodeToString(data))
	}

	// Encode TileBanks
	for i := 0; i < 4; i++ {
		data := flattenTileBank(c.TileBanks[i].Tiles)
		cart.TileBanks = append(cart.TileBanks, hex.EncodeToString(data))
	}

	// Encode SpriteData
	spriteData := flattenSpriteData(c.SpriteData)
	cart.SpriteData = hex.EncodeToString(spriteData)

	// Encode TileLayers
	for i := 0; i < TileLayerCount; i++ {
		cart.TileLayers = append(cart.TileLayers, c.TileLayers[i])
	}

	// Encode FontData
	fontData := flattenFontData(c.FontData)
	cart.FontData = hex.EncodeToString(fontData)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cart)
}

// LoadJSON loads the console state from a JSON file.
func (c *Console) LoadJSON(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var cart Cartridge
	if err := json.NewDecoder(f).Decode(&cart); err != nil {
		return err
	}

	// Decode Palettes
	for i, h := range cart.Palettes {
		if i >= 4 {
			break
		}
		data, err := hex.DecodeString(h)
		if err != nil {
			return fmt.Errorf("failed to decode palette %d: %v", i, err)
		}
		copyPalette(&c.PaletteBank.Colors[i], data)
	}

	// Decode TileBanks
	for i, h := range cart.TileBanks {
		if i >= 4 {
			break
		}
		data, err := hex.DecodeString(h)
		if err != nil {
			return fmt.Errorf("failed to decode tile bank %d: %v", i, err)
		}
		copyTileBank(&c.TileBanks[i].Tiles, data)
	}

	// Decode SpriteData
	data, err := hex.DecodeString(cart.SpriteData)
	if err != nil {
		return fmt.Errorf("failed to decode sprite data: %v", err)
	}
	copySpriteData(&c.SpriteData, data)

	// Decode TileLayers
	for i, l := range cart.TileLayers {
		if i >= TileLayerCount {
			break
		}
		c.TileLayers[i] = l
	}

	// Decode FontData
	if cart.FontData != "" {
		data, err := hex.DecodeString(cart.FontData)
		if err != nil {
			return fmt.Errorf("failed to decode font data: %v", err)
		}
		copyFontData(&c.FontData, data)
	}

	return nil
}

func flattenPalette(p [16][4]byte) []byte {
	out := make([]byte, 16*4)
	for i := 0; i < 16; i++ {
		copy(out[i*4:], p[i][:])
	}
	return out
}

func copyPalette(dst *[16][4]byte, src []byte) {
	for i := 0; i < 16 && i*4+4 <= len(src); i++ {
		copy(dst[i][:], src[i*4:i*4+4])
	}
}

func flattenTileBank(t [256][32]byte) []byte {
	out := make([]byte, 256*32)
	for i := 0; i < 256; i++ {
		copy(out[i*32:], t[i][:])
	}
	return out
}

func copyTileBank(dst *[256][32]byte, src []byte) {
	for i := 0; i < 256 && i*32+32 <= len(src); i++ {
		copy(dst[i][:], src[i*32:i*32+32])
	}
}

func flattenSpriteData(s [256][32]byte) []byte {
	out := make([]byte, 256*32)
	for i := 0; i < 256; i++ {
		copy(out[i*32:], s[i][:])
	}
	return out
}

func copySpriteData(dst *[256][32]byte, src []byte) {
	for i := 0; i < 256 && i*32+32 <= len(src); i++ {
		copy(dst[i][:], src[i*32:i*32+32])
	}
}

func flattenFontData(f [128][8]byte) []byte {
	out := make([]byte, 128*8)
	for i := 0; i < 128; i++ {
		copy(out[i*8:], f[i][:])
	}
	return out
}

func copyFontData(dst *[128][8]byte, src []byte) {
	for i := 0; i < 128 && i*8+8 <= len(src); i++ {
		copy(dst[i][:], src[i*8:i*8+8])
	}
}
