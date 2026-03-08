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

	// input edge detection
	prevFlap bool
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
	cur := g.c.IsPressed(gonsole.ButtonA) ||
		g.c.IsPressed(gonsole.ButtonUp) ||
		g.c.MousePressed(gonsole.MouseButtonLeft)
	fired := cur && !g.prevFlap
	g.prevFlap = cur
	return fired
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

func (g *game) reset() {
	g.birdY = float64(gonsole.ScreenHeight/2 - 4)
	g.velY = 0
	g.score = 0
	g.flapAge = 0
	g.prevFlap = false
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

	return nil
}

func (g *game) draw(frame, ms uint64) {
	c := g.c

	// sky
	c.DrawRect(0, 0, gonsole.ScreenWidth, groundY, 12, true)
	// ground
	c.DrawRect(0, groundY, gonsole.ScreenWidth, 16, 4, true)
	c.DrawRect(0, groundY, gonsole.ScreenWidth, 2, 3, true)

	// pipes
	for _, p := range g.pipes {
		g.drawPipe(p)
	}

	// bird sprite data
	switch {
	case g.st == stateDead:
		c.SetSpriteData(0, birdDead[:])
	case g.flapAge < 8:
		c.SetSpriteData(0, birdFlap[:])
	default:
		c.SetSpriteData(0, birdNormal[:])
	}

	bird := c.GetSprite(0)
	bird.X = birdX
	bird.Y = int16(g.birdY)
	c.SetSprite(0, bird)

	// score
	c.DrawText(gonsole.ScreenWidth/2-12, 4, fmt.Sprintf("%d", g.score))

	// overlay messages
	switch g.st {
	case stateWait:
		c.DrawText(gonsole.ScreenWidth/2-47, gonsole.ScreenHeight/2-8, "PRESS A TO START")
	case stateDead:
		c.DrawText(gonsole.ScreenWidth/2-26, gonsole.ScreenHeight/2-10, "GAME OVER")
		c.DrawText(gonsole.ScreenWidth/2-47, gonsole.ScreenHeight/2+6, "PRESS A TO RETRY")
	}
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	c := gonsole.NewConsole()

	g := &game{c: c}
	g.reset()

	// bird sprite setup
	c.SetSpriteData(0, birdNormal[:])
	c.SetSprite(0, gonsole.Sprite{
		X:     birdX,
		Y:     int16(g.birdY),
		Props: gonsole.SpritePropVisible | gonsole.SpritePropScreenSpace,
	})

	c.UpdateFunc = g.update
	c.DrawFunc = g.draw

	if err := gonsole.Run(c); err != nil {
		log.Fatal(err)
	}
}
