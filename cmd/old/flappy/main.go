package main

import (
	"fmt"
	"log"
	"math/rand"

	"aymanhs/gonsole"
)

// ── constants ────────────────────────────────────────────────────────────────

const (
	gravity     = 0.1
	flapForce   = -2
	pipeSpeed   = 1.5
	pipeSpacing = 90 // horizontal distance between pipe pairs
	pipeGap     = 52 // vertical opening height
	pipeWidth   = 14
	groundY     = gonsole.ScreenHeight - 16
	birdX       = 48
	maxPipes    = 4
)

// ── sprite slot assignments ───────────────────────────────────────────────────
// slot 0  : bird
// slots 1–4 : top pipes   (one per pair)
// slots 5–8 : bottom pipes (one per pair)

// ── pixel art ────────────────────────────────────────────────────────────────

// bird: 8×8, colors 10=yellow body, 9=orange beak, 8=red wing accent, 7=white eye
var birdNormal = [64]byte{
	0, 0, 10, 10, 10, 10, 0, 0,
	0, 10, 10, 7, 10, 10, 10, 0,
	10, 10, 10, 7, 10, 9, 9, 0,
	10, 10, 10, 10, 10, 9, 9, 0,
	8, 10, 10, 10, 10, 10, 0, 0,
	8, 8, 10, 10, 10, 0, 0, 0,
	0, 8, 10, 10, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// bird flapping (wing up)
var birdFlap = [64]byte{
	0, 8, 8, 10, 10, 10, 0, 0,
	0, 10, 10, 7, 10, 10, 10, 0,
	10, 10, 10, 7, 10, 9, 9, 0,
	10, 10, 10, 10, 10, 9, 9, 0,
	0, 10, 10, 10, 10, 10, 0, 0,
	0, 0, 10, 10, 10, 0, 0, 0,
	0, 0, 10, 10, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// dead bird (x eyes)
var birdDead = [64]byte{
	0, 0, 10, 10, 10, 10, 0, 0,
	0, 10, 5, 7, 5, 10, 10, 0,
	10, 10, 7, 5, 7, 9, 9, 0,
	10, 10, 10, 10, 10, 9, 9, 0,
	10, 10, 10, 10, 10, 10, 0, 0,
	0, 10, 10, 10, 10, 0, 0, 0,
	0, 0, 10, 10, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// ── parallax tile art ───────────────────────────────────────────────────────

// cloudTileA: large cloud puff (tile 0, TileBank[0]) — 7=white, 6=light grey, 0=transparent
var cloudTileA = [8][8]byte{
	{0, 0, 6, 7, 7, 6, 0, 0},
	{0, 6, 7, 7, 7, 7, 6, 0},
	{6, 7, 7, 7, 7, 7, 7, 6},
	{6, 7, 7, 7, 7, 7, 7, 6},
	{0, 6, 7, 7, 7, 7, 6, 0},
	{0, 0, 6, 6, 6, 6, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

// cloudTileB: small cloud puff (tile 1, TileBank[0])
var cloudTileB = [8][8]byte{
	{0, 0, 0, 6, 6, 0, 0, 0},
	{0, 0, 6, 7, 7, 7, 0, 0},
	{0, 6, 7, 7, 7, 7, 7, 0},
	{0, 6, 7, 7, 7, 7, 7, 0},
	{0, 0, 6, 7, 7, 7, 0, 0},
	{0, 0, 0, 6, 6, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

// shrubTileA: round bush (tile 2, TileBank[0]) — 11=green, 3=dark green, 4=brown
var shrubTileA = [8][8]byte{
	{0, 0, 0, 3, 3, 0, 0, 0},
	{0, 0, 3, 11, 11, 3, 0, 0},
	{0, 3, 11, 11, 11, 11, 3, 0},
	{3, 11, 11, 11, 11, 11, 11, 3},
	{3, 11, 11, 11, 11, 11, 11, 3},
	{0, 3, 4, 4, 4, 4, 3, 0},
	{0, 0, 4, 4, 4, 4, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

// shrubTileB: spiky bush (tile 3, TileBank[0])
var shrubTileB = [8][8]byte{
	{0, 0, 3, 0, 0, 3, 0, 0},
	{0, 3, 11, 3, 3, 11, 3, 0},
	{3, 11, 11, 11, 11, 11, 11, 3},
	{3, 11, 11, 11, 11, 11, 11, 3},
	{0, 3, 11, 11, 11, 11, 3, 0},
	{0, 3, 4, 4, 4, 4, 3, 0},
	{0, 0, 0, 4, 4, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

// pipe segment tile: 8×8, color 11=green, 3=dark green edge
var pipeTile = [64]byte{
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
	3, 11, 11, 11, 11, 11, 11, 3,
}

// ── game state ────────────────────────────────────────────────────────────────

type state int

const (
	stateWait state = iota
	statePlaying
	stateDead
)

type pipe struct {
	x    float64
	gapY int // top of the gap
}

type game struct {
	c       *gonsole.Console
	st      state
	birdY   float64
	velY    float64
	pipes   [maxPipes]pipe
	score   int
	flapAge int // frames since last flap (for sprite swap)

	worldDist float64 // accumulated world pixels scrolled; drives CameraX

	// input edge detection
	// (removed prevFlap manually, now using c.JustPressed/JustPressedMouse)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (g *game) flapping() bool {
	return g.st == statePlaying &&
		g.c.IsPressed(gonsole.ButtonA) ||
		g.c.MousePressed(gonsole.MouseButtonLeft)
}

func (g *game) justPressed() bool {
	return g.c.JustPressed(gonsole.ButtonA) ||
		g.c.JustPressed(gonsole.ButtonUp) ||
		g.c.JustPressedMouse(gonsole.MouseButtonLeft)
}

// spawnPipe places pipe i at the right edge with a random gap.
func (g *game) spawnPipe(i int, x float64) {
	minGap := 24
	maxGap := groundY - pipeGap - 24
	g.pipes[i] = pipe{
		x:    x,
		gapY: minGap + rand.Intn(maxGap-minGap),
	}
}

// setupParallax loads tile graphics into TileBank[0] and configures parallax.
// Call once from main; tile pixel data is idempotent.
func (g *game) setupParallax() {
	c := g.c
	// Write nibble-packed pixel data for all four tile shapes
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			c.TileBanks[0].TileSetPixel(0, x, y, cloudTileA[y][x])
			c.TileBanks[0].TileSetPixel(1, x, y, cloudTileB[y][x])
			c.TileBanks[0].TileSetPixel(2, x, y, shrubTileA[y][x])
			c.TileBanks[0].TileSetPixel(3, x, y, shrubTileB[y][x])
		}
	}
	// TileLayer[0] = distant clouds at 1/4 camera speed
	c.TileLayers[0].ParallaxMul = 1
	c.TileLayers[0].ParallaxDiv = 4
	// TileLayer[1] = foreground shrubs at 3/2 camera speed
	c.TileLayers[1].ParallaxMul = 3
	c.TileLayers[1].ParallaxDiv = 2
}

// seedParallaxTiles clears both parallax layers and populates initial tile slots.
// Called from reset() so a retry gets a fresh tile layout.
func (g *game) seedParallaxTiles() {
	c := g.c
	c.TileLayers[0].Clear()
	c.TileLayers[1].Clear()

	// 12 cloud tiles spread across initial world view (worldX = 0..880)
	cloudYs := [12]uint16{20, 40, 16, 50, 28, 42, 18, 36, 22, 52, 34, 14}
	for i := range cloudYs {
		c.TileLayers[0].AddTile(uint16(i*80), cloudYs[i], byte(i%2), 0, 0)
	}

	// 16 shrub tiles along the ground: worldY = groundY-5 so rows 5-7 (roots)
	// are hidden under the ground strip drawn at PaintSlotAfterL1.
	shrubXs := [16]uint16{0, 22, 50, 80, 112, 142, 175, 210, 242, 275, 305, 338, 368, 398, 425, 454}
	for _, x := range shrubXs {
		tileID := byte(2 + (x/22)%2) // alternate tile 2 and 3
		c.TileLayers[1].AddTile(x, groundY-5, tileID, 0, 0)
	}
}

// recycleTiles moves tiles that have scrolled off the left edge back to the right.
func (g *game) recycleTiles() {
	c := g.c

	tl0 := &c.TileLayers[0]
	eff0 := int(c.CameraX) * tl0.ParallaxMul / tl0.ParallaxDiv
	for i := 0; i < tl0.Count; i++ {
		s := &tl0.Slots[i]
		if int(s.WorldX)-eff0 < -8 {
			s.WorldX = uint16(eff0 + gonsole.ScreenWidth + 40 + rand.Intn(200))
			s.WorldY = uint16(14 + rand.Intn(54))
			s.TileID = byte(rand.Intn(2))
		}
	}

	tl1 := &c.TileLayers[1]
	eff1 := int(c.CameraX) * tl1.ParallaxMul / tl1.ParallaxDiv
	for i := 0; i < tl1.Count; i++ {
		s := &tl1.Slots[i]
		if int(s.WorldX)-eff1 < -8 {
			s.WorldX = uint16(eff1 + gonsole.ScreenWidth + 16 + rand.Intn(56))
			s.TileID = byte(2 + rand.Intn(2))
		}
	}
}

func (g *game) reset() {
	g.birdY = float64(gonsole.ScreenHeight/2 - 4)
	g.velY = 0
	g.score = 0
	g.flapAge = 0
	g.worldDist = 0
	g.c.CameraX = 0
	g.seedParallaxTiles()
	for i := range g.pipes {
		g.spawnPipe(i, float64(gonsole.ScreenWidth+20+i*pipeSpacing))
	}
}

// ── digit rendering (3×5 pixel font) ─────────────────────────────────────────

// ── pipe drawing ──────────────────────────────────────────────────────────────

func (g *game) drawPipe(p pipe) {
	topH := p.gapY
	botY := p.gapY + pipeGap
	botH := groundY - botY
	px := int(p.x)

	// top pipe
	for row := 0; row < topH; row += 8 {
		h := 8
		if row+h > topH {
			h = topH - row
		}
		for r := 0; r < h; r++ {
			for col := 0; col < pipeWidth; col++ {
				tileIdx := (r * 8) + (col % 8)
				g.c.SetPixel(px+col, row+r, pipeTile[tileIdx])
			}
		}
	}

	// bottom pipe
	for row := 0; row < botH; row += 8 {
		h := 8
		if row+h > botH {
			h = botH - row
		}
		for r := 0; r < h; r++ {
			for col := 0; col < pipeWidth; col++ {
				tileIdx := (r * 8) + (col % 8)
				g.c.SetPixel(px+col, botY+row+r, pipeTile[tileIdx])
			}
		}
	}
}

// ── collision ─────────────────────────────────────────────────────────────────

func (g *game) collides() bool {
	by := int(g.birdY)

	// ground / ceiling
	if by+7 >= groundY || by < 0 {
		return true
	}

	// pipes (AABB, bird is 8×8 at birdX,birdY)
	for _, p := range g.pipes {
		px := int(p.x)
		if birdX+7 < px || birdX > px+pipeWidth-1 {
			continue
		}
		if by < p.gapY || by+7 > p.gapY+pipeGap {
			return true
		}
	}
	return false
}

// ── update ────────────────────────────────────────────────────────────────────

func (g *game) update(frame, ms uint64) error {

	switch g.st {
	case stateWait:
		if g.justPressed() {
			g.st = statePlaying
			g.velY = flapForce
		}

	case statePlaying:
		// flap
		if g.justPressed() {
			g.velY = flapForce
			g.flapAge = 0
		}
		g.flapAge++

		// physics
		g.velY += gravity
		g.velY = clamp(g.velY, -6, 7)
		g.birdY += g.velY

		// scroll pipes, score
		for i := range g.pipes {
			g.pipes[i].x -= pipeSpeed
			// passed the bird?
			if g.pipes[i].x+pipeWidth < birdX && g.pipes[i].x+pipeWidth >= birdX-pipeSpeed {
				g.score++
			}
			// off screen — recycle
			if g.pipes[i].x+pipeWidth < 0 {
				// find rightmost pipe and place after it
				rightmost := g.pipes[0].x
				for _, p := range g.pipes {
					if p.x > rightmost {
						rightmost = p.x
					}
				}
				g.spawnPipe(i, rightmost+float64(pipeSpacing))
			}
		}

		if g.collides() {
			g.st = stateDead
		}

	case stateDead:
		if g.justPressed() {
			g.reset()
			g.st = statePlaying
		}
	}

	// advance world camera while playing and recycle off-screen tiles
	if g.st == statePlaying {
		g.worldDist += pipeSpeed
		g.c.CameraX = uint16(g.worldDist)
		g.recycleTiles()
	}

	// update bird stamp (sprite data + position) every frame
	switch {
	case g.st == stateDead:
		g.c.SetSpriteData(0, birdDead[:])
	case g.flapAge < 8:
		g.c.SetSpriteData(0, birdFlap[:])
	default:
		g.c.SetSpriteData(0, birdNormal[:])
	}
	bird := g.c.GetStamp(0)
	bird.X = birdX
	bird.Y = uint16(g.birdY)
	g.c.SetStamp(0, bird)

	return nil
}

func (g *game) paint(slot int, frame uint64) {
	c := g.c
	switch slot {
	case gonsole.PaintSlotBegin:
		// sky — drawn first, clouds (TileLayer[0]) composite on top next
		c.DrawRect(0, 0, gonsole.ScreenWidth, groundY, 12, true)

	case gonsole.PaintSlotAfterL0:
		// pipes — drawn after distant clouds, before foreground shrubs
		for _, p := range g.pipes {
			g.drawPipe(p)
		}

	case gonsole.PaintSlotAfterL1:
		// ground strip — drawn after shrubs (TileLayer[1]) to hide their roots
		c.DrawRect(0, groundY, gonsole.ScreenWidth, 16, 4, true)
		c.DrawRect(0, groundY, gonsole.ScreenWidth, 2, 3, true)
		// bird stamp (DrawLayer=2) renders next, appearing above the ground

	case gonsole.PaintSlotEnd:
		// HUD: score and overlay text on top of everything
		c.DrawText(gonsole.ScreenWidth/2-12, 4, fmt.Sprintf("%d", g.score))
		switch g.st {
		case stateWait:
			c.DrawText(gonsole.ScreenWidth/2-47, gonsole.ScreenHeight/2-8, "PRESS A TO START")
		case stateDead:
			c.DrawText(gonsole.ScreenWidth/2-26, gonsole.ScreenHeight/2-10, "GAME OVER")
			c.DrawText(gonsole.ScreenWidth/2-47, gonsole.ScreenHeight/2+6, "PRESS A TO RETRY")
		}
	}
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	c := gonsole.NewConsole()

	g := &game{c: c}
	g.setupParallax() // load tile gfx + set parallax params (before reset seeds tiles)
	g.reset()

	// bird sprite setup — DrawLayer=2 so bird renders after shrubs (TileLayer[1])
	c.SetSpriteData(0, birdNormal[:])
	c.SetStamp(0, gonsole.Stamp{
		X:         birdX,
		Y:         uint16(g.birdY),
		Props:     gonsole.StampPropVisible | gonsole.StampPropScreenSpace,
		PaletteID: 0,
		DrawLayer: 2,
	})

	c.UpdateFunc = g.update
	c.PaintFunc = g.paint

	if err := gonsole.Run(c); err != nil {
		log.Fatal(err)
	}
}
