package gonsole

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// Cartridge represents the serialized state of a gonsole project.
type Cartridge struct {
	Version    int                  `json:"version"`
	Palettes   map[string]string    `json:"palettes,omitempty"`    // Hex strings of [16][4]byte (pre-expanded)
	TileBanks  map[string]string    `json:"tile_banks,omitempty"`  // Hex strings of [256][32]byte
	SpriteData string               `json:"sprite_data,omitempty"` // Hex string of [256][32]byte
	TileLayers map[string]TileLayer `json:"tile_layers,omitempty"`
	FontData   string               `json:"font_data,omitempty"` // Hex string of [128][8]byte
}

func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// SaveJSON saves the console state to a JSON file with hex-encoded buffers.
func (c *Console) SaveJSON(path string) error {
	cart := Cartridge{
		Version:    1,
		Palettes:   make(map[string]string),
		TileBanks:  make(map[string]string),
		TileLayers: make(map[string]TileLayer),
	}

	// Encode PaletteBanks
	for i := 0; i < 4; i++ {
		data := flattenPalette(c.PaletteBank.Colors[i])
		if !isZero(data) {
			cart.Palettes[fmt.Sprintf("%d", i)] = hex.EncodeToString(data)
		}
	}

	// Encode TileBanks
	for i := 0; i < 4; i++ {
		data := flattenTileBank(c.TileBanks[i].Tiles)
		if !isZero(data) {
			cart.TileBanks[fmt.Sprintf("%d", i)] = hex.EncodeToString(data)
		}
	}

	// Encode SpriteData
	spriteData := flattenSpriteData(c.SpriteData)
	if !isZero(spriteData) {
		cart.SpriteData = hex.EncodeToString(spriteData)
	}

	// Encode TileLayers
	emptyLayer := TileLayer{}
	for i := 0; i < TileLayerCount; i++ {
		if c.TileLayers[i] != emptyLayer {
			cart.TileLayers[fmt.Sprintf("%d", i)] = c.TileLayers[i]
		}
	}

	// Encode FontData
	fontData := flattenFontData(c.FontData)
	if !isZero(fontData) {
		cart.FontData = hex.EncodeToString(fontData)
	}

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

	// Use raw messages to support backwards compatibility with slices
	var raw struct {
		Version    int             `json:"version"`
		Palettes   json.RawMessage `json:"palettes"`
		TileBanks  json.RawMessage `json:"tile_banks"`
		SpriteData string          `json:"sprite_data"`
		TileLayers json.RawMessage `json:"tile_layers"`
		FontData   string          `json:"font_data"`
	}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return err
	}

	decodeStringSliceOrMap := func(raw json.RawMessage, maxCount int) (map[int]string, error) {
		res := make(map[int]string)
		if len(raw) == 0 || string(raw) == "null" {
			return res, nil
		}
		if raw[0] == '[' {
			var slice []string
			if err := json.Unmarshal(raw, &slice); err != nil {
				return nil, err
			}
			for i, v := range slice {
				res[i] = v
			}
		} else if raw[0] == '{' {
			var m map[string]string
			if err := json.Unmarshal(raw, &m); err != nil {
				return nil, err
			}
			for k, v := range m {
				var i int
				if _, err := fmt.Sscanf(k, "%d", &i); err == nil {
					res[i] = v
				}
			}
		}
		return res, nil
	}

	// Decode Palettes
	palettes, err := decodeStringSliceOrMap(raw.Palettes, 4)
	if err != nil {
		return fmt.Errorf("failed to decode palettes: %v", err)
	}
	for i, h := range palettes {
		if i >= 4 {
			continue
		}
		data, err := hex.DecodeString(h)
		if err != nil {
			return fmt.Errorf("failed to decode palette %d: %v", i, err)
		}
		copyPalette(&c.PaletteBank.Colors[i], data)
	}

	// Decode TileBanks
	tileBanks, err := decodeStringSliceOrMap(raw.TileBanks, 4)
	if err != nil {
		return fmt.Errorf("failed to decode tile banks: %v", err)
	}
	for i, h := range tileBanks {
		if i >= 4 {
			continue
		}
		data, err := hex.DecodeString(h)
		if err != nil {
			return fmt.Errorf("failed to decode tile bank %d: %v", i, err)
		}
		copyTileBank(&c.TileBanks[i].Tiles, data)
	}

	// Decode SpriteData
	if raw.SpriteData != "" {
		data, err := hex.DecodeString(raw.SpriteData)
		if err != nil {
			return fmt.Errorf("failed to decode sprite data: %v", err)
		}
		copySpriteData(&c.SpriteData, data)
	}

	// Decode TileLayers
	if len(raw.TileLayers) > 0 && string(raw.TileLayers) != "null" {
		if raw.TileLayers[0] == '[' {
			var layers []TileLayer
			if err := json.Unmarshal(raw.TileLayers, &layers); err != nil {
				return fmt.Errorf("failed to decode tile layers: %v", err)
			}
			for i, l := range layers {
				if i >= TileLayerCount {
					break
				}
				c.TileLayers[i] = l
			}
		} else if raw.TileLayers[0] == '{' {
			var m map[string]TileLayer
			if err := json.Unmarshal(raw.TileLayers, &m); err != nil {
				return fmt.Errorf("failed to decode tile layers: %v", err)
			}
			for k, l := range m {
				var i int
				if _, err := fmt.Sscanf(k, "%d", &i); err == nil && i < TileLayerCount {
					c.TileLayers[i] = l
				}
			}
		}
	}

	// Decode FontData
	if raw.FontData != "" {
		data, err := hex.DecodeString(raw.FontData)
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
