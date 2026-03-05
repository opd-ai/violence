// Package damagestate provides visual damage state rendering for entities.
package damagestate

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages damage state visualization for entities.
type System struct {
	genre        string
	overlayCache map[cacheKey]*ebiten.Image
	cacheMu      sync.RWMutex
	logger       *logrus.Entry
}

type cacheKey struct {
	level   int
	pattern int64
	genre   string
	dirX    float64
	dirY    float64
	size    int
}

// NewSystem creates a damage state visualization system.
func NewSystem(genre string) *System {
	return &System{
		genre:        genre,
		overlayCache: make(map[cacheKey]*ebiten.Image),
		logger: logrus.WithFields(logrus.Fields{
			"system": "damagestate",
		}),
	}
}

// Update processes damage state components and updates damage levels.
func (sys *System) Update(w *engine.World) {
	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, entity := range entities {
		compRaw, ok := w.GetComponent(entity, compType)
		if !ok {
			continue
		}
		comp, ok := compRaw.(*Component)
		if !ok {
			continue
		}

		comp.UpdateDamage()
	}
}

// RenderDamageOverlay generates a damage state overlay for a sprite.
func (sys *System) RenderDamageOverlay(comp *Component, baseSprite *ebiten.Image) *ebiten.Image {
	if comp == nil || baseSprite == nil || comp.DamageLevel == 0 {
		return baseSprite
	}

	bounds := baseSprite.Bounds()
	size := bounds.Dx()

	key := cacheKey{
		level:   comp.DamageLevel,
		pattern: comp.WoundPattern,
		genre:   sys.genre,
		dirX:    comp.LastDamageX,
		dirY:    comp.LastDamageY,
		size:    size,
	}

	sys.cacheMu.RLock()
	cached := sys.overlayCache[key]
	sys.cacheMu.RUnlock()

	if cached != nil && !comp.DirtyCache {
		result := ebiten.NewImageFromImage(baseSprite)
		op := &ebiten.DrawImageOptions{}
		result.DrawImage(cached, op)
		return result
	}

	overlay := sys.generateDamageOverlay(comp, size)

	sys.cacheMu.Lock()
	sys.overlayCache[key] = overlay
	sys.cacheMu.Unlock()

	comp.DirtyCache = false

	result := ebiten.NewImageFromImage(baseSprite)
	op := &ebiten.DrawImageOptions{}
	result.DrawImage(overlay, op)
	return result
}

// generateDamageOverlay creates the visual damage representation.
func (sys *System) generateDamageOverlay(comp *Component, size int) *ebiten.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(comp.WoundPattern))

	switch sys.genre {
	case "scifi":
		sys.renderSciFiDamage(rgba, comp, size, rng)
	case "horror":
		sys.renderHorrorDamage(rgba, comp, size, rng)
	case "cyberpunk":
		sys.renderCyberpunkDamage(rgba, comp, size, rng)
	case "postapoc":
		sys.renderPostApocDamage(rgba, comp, size, rng)
	default:
		sys.renderFantasyDamage(rgba, comp, size, rng)
	}

	return ebiten.NewImageFromImage(rgba)
}

// renderFantasyDamage renders blood and wounds for fantasy genre.
func (sys *System) renderFantasyDamage(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	bloodColor := color.RGBA{R: 120, G: 0, B: 0, A: 180}
	darkBloodColor := color.RGBA{R: 80, G: 0, B: 0, A: 200}

	woundCount := comp.DamageLevel * 2
	for i := 0; i < woundCount; i++ {
		cx := rng.Intn(size)
		cy := rng.Intn(size)
		radius := 2 + rng.Intn(comp.DamageLevel+1)

		// Bias wounds toward damage direction
		if comp.LastDamageX != 0 || comp.LastDamageY != 0 {
			cx = size/2 + int(comp.LastDamageX*float64(size)/4)
			cy = size/2 + int(comp.LastDamageY*float64(size)/4)
			cx += rng.Intn(size/3) - size/6
			cy += rng.Intn(size/3) - size/6
			cx = clamp(cx, 0, size-1)
			cy = clamp(cy, 0, size-1)
		}

		sys.fillCircle(rgba, cx, cy, radius, bloodColor)

		// Add darker center for depth
		if radius > 2 {
			sys.fillCircle(rgba, cx, cy, radius/2, darkBloodColor)
		}
	}

	// Critical damage: add blood drips
	if comp.DamageLevel >= 3 {
		dripCount := 2 + rng.Intn(3)
		for i := 0; i < dripCount; i++ {
			startX := rng.Intn(size)
			startY := rng.Intn(size / 2)
			dripLen := 3 + rng.Intn(size/4)

			for j := 0; j < dripLen; j++ {
				y := startY + j
				if y >= size {
					break
				}
				alpha := uint8(180 - j*15)
				if alpha < 60 {
					alpha = 60
				}
				rgba.Set(startX, y, color.RGBA{R: 100, G: 0, B: 0, A: alpha})
			}
		}
	}
}

// renderSciFiDamage renders sparks and cracks for sci-fi genre.
func (sys *System) renderSciFiDamage(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	sparkColor := color.RGBA{R: 255, G: 200, B: 50, A: 200}
	crackColor := color.RGBA{R: 40, G: 40, B: 40, A: 180}
	glitchColor := color.RGBA{R: 0, G: 255, B: 255, A: 150}

	sys.renderSciFiCracks(rgba, comp.DamageLevel, size, crackColor, rng)
	sys.renderSciFiSparks(rgba, comp.DamageLevel, size, sparkColor, rng)
	sys.renderSciFiGlitches(rgba, comp.DamageLevel, size, glitchColor, rng)
}

// renderSciFiCracks draws branching crack lines across the surface.
func (sys *System) renderSciFiCracks(rgba *image.RGBA, damageLevel, size int, crackColor color.RGBA, rng *rand.Rand) {
	for i := 0; i < damageLevel; i++ {
		startX := rng.Intn(size)
		startY := rng.Intn(size)
		angle := rng.Float64() * 2 * math.Pi
		length := 4 + rng.Intn(size/3)

		for j := 0; j < length; j++ {
			x := startX + int(float64(j)*math.Cos(angle))
			y := startY + int(float64(j)*math.Sin(angle))
			if x >= 0 && x < size && y >= 0 && y < size {
				rgba.Set(x, y, crackColor)
			}
			if j > 2 && rng.Float64() < 0.15 {
				sys.renderCrackBranch(rgba, x, y, size, angle, crackColor, rng)
			}
		}
	}
}

// renderCrackBranch draws a small branching crack from a main crack.
func (sys *System) renderCrackBranch(rgba *image.RGBA, x, y, size int, baseAngle float64, crackColor color.RGBA, rng *rand.Rand) {
	branchAngle := baseAngle + (rng.Float64()-0.5)*math.Pi/3
	branchLen := 2 + rng.Intn(4)
	for k := 0; k < branchLen; k++ {
		bx := x + int(float64(k)*math.Cos(branchAngle))
		by := y + int(float64(k)*math.Sin(branchAngle))
		if bx >= 0 && bx < size && by >= 0 && by < size {
			rgba.Set(bx, by, crackColor)
		}
	}
}

// renderSciFiSparks draws electric spark points with radiating rays for moderate damage.
func (sys *System) renderSciFiSparks(rgba *image.RGBA, damageLevel, size int, sparkColor color.RGBA, rng *rand.Rand) {
	if damageLevel < 2 {
		return
	}
	sparkCount := damageLevel - 1
	for i := 0; i < sparkCount; i++ {
		cx := rng.Intn(size)
		cy := rng.Intn(size)
		sys.fillCircle(rgba, cx, cy, 1, sparkColor)
		sys.renderSparkRays(rgba, cx, cy, size, rng)
	}
}

// renderSparkRays draws radiating rays emanating from a spark point.
func (sys *System) renderSparkRays(rgba *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	rayCount := 3 + rng.Intn(3)
	for j := 0; j < rayCount; j++ {
		angle := rng.Float64() * 2 * math.Pi
		rayLen := 2 + rng.Intn(3)
		for k := 1; k <= rayLen; k++ {
			rx := cx + int(float64(k)*math.Cos(angle))
			ry := cy + int(float64(k)*math.Sin(angle))
			if rx >= 0 && rx < size && ry >= 0 && ry < size {
				alpha := uint8(200 - k*40)
				rgba.Set(rx, ry, color.RGBA{R: 255, G: 200, B: 50, A: alpha})
			}
		}
	}
}

// renderSciFiGlitches draws horizontal glitch lines for critical damage.
func (sys *System) renderSciFiGlitches(rgba *image.RGBA, damageLevel, size int, glitchColor color.RGBA, rng *rand.Rand) {
	if damageLevel < 3 {
		return
	}
	glitchCount := 2 + rng.Intn(2)
	for i := 0; i < glitchCount; i++ {
		y := rng.Intn(size)
		startX := rng.Intn(size / 2)
		endX := startX + 3 + rng.Intn(size/3)
		for x := startX; x < endX && x < size; x++ {
			rgba.Set(x, y, glitchColor)
		}
	}
}

// renderHorrorDamage renders gore and decay for horror genre.
func (sys *System) renderHorrorDamage(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	goreColor := color.RGBA{R: 140, G: 0, B: 20, A: 220}
	decayColor := color.RGBA{R: 60, G: 80, B: 40, A: 180}

	// Heavy blood splatter
	splatterCount := comp.DamageLevel * 3
	for i := 0; i < splatterCount; i++ {
		cx := rng.Intn(size)
		cy := rng.Intn(size)
		radius := 1 + rng.Intn(comp.DamageLevel+2)

		// Irregular splatter shape
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if rng.Float64() < 0.6 {
					x := cx + dx
					y := cy + dy
					if x >= 0 && x < size && y >= 0 && y < size {
						rgba.Set(x, y, goreColor)
					}
				}
			}
		}
	}

	// Add decay patches for critical damage
	if comp.DamageLevel >= 2 {
		decayCount := comp.DamageLevel - 1
		for i := 0; i < decayCount; i++ {
			cx := rng.Intn(size)
			cy := rng.Intn(size)
			radius := 3 + rng.Intn(3)
			sys.fillCircle(rgba, cx, cy, radius, decayColor)
		}
	}
}

// renderCyberpunkDamage renders neon glitches and circuit damage.
func (sys *System) renderCyberpunkDamage(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	sys.renderCircuitCracks(rgba, comp, size, rng)
	if comp.DamageLevel >= 3 {
		sys.renderGlitchBlocks(rgba, size, rng)
	}
}

// renderCircuitCracks draws neon circuit cracks on the damaged surface.
func (sys *System) renderCircuitCracks(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	neonPink := color.RGBA{R: 255, G: 0, B: 128, A: 200}
	neonCyan := color.RGBA{R: 0, G: 255, B: 255, A: 200}
	darkGray := color.RGBA{R: 30, G: 30, B: 35, A: 180}

	crackCount := comp.DamageLevel * 2
	for i := 0; i < crackCount; i++ {
		neonColor := selectNeonColor(neonPink, neonCyan, rng)
		drawNeonCrack(rgba, size, neonColor, darkGray, rng)
	}
}

// selectNeonColor randomly chooses between pink and cyan neon colors.
func selectNeonColor(pink, cyan color.RGBA, rng *rand.Rand) color.RGBA {
	if rng.Float64() < 0.5 {
		return cyan
	}
	return pink
}

// drawNeonCrack renders a single neon crack line with alternating colors.
func drawNeonCrack(rgba *image.RGBA, size int, neonColor, darkColor color.RGBA, rng *rand.Rand) {
	startX := rng.Intn(size)
	startY := rng.Intn(size)
	angle := rng.Float64() * 2 * math.Pi
	length := 3 + rng.Intn(size/4)

	for j := 0; j < length; j++ {
		x := startX + int(float64(j)*math.Cos(angle))
		y := startY + int(float64(j)*math.Sin(angle))
		if x >= 0 && x < size && y >= 0 && y < size {
			if j%2 == 0 {
				rgba.Set(x, y, neonColor)
			} else {
				rgba.Set(x, y, darkColor)
			}
		}
	}
}

// renderGlitchBlocks draws glitchy rectangular blocks for critical damage.
func (sys *System) renderGlitchBlocks(rgba *image.RGBA, size int, rng *rand.Rand) {
	neonPink := color.RGBA{R: 255, G: 0, B: 128, A: 200}
	neonCyan := color.RGBA{R: 0, G: 255, B: 255, A: 200}

	glitchCount := 2 + rng.Intn(2)
	for i := 0; i < glitchCount; i++ {
		glitchColor := selectNeonColor(neonPink, neonCyan, rng)
		drawGlitchBlock(rgba, size, glitchColor, rng)
	}
}

// drawGlitchBlock renders a single rectangular glitch block.
func drawGlitchBlock(rgba *image.RGBA, size int, glitchColor color.RGBA, rng *rand.Rand) {
	x := rng.Intn(size - 4)
	y := rng.Intn(size - 4)
	w := 2 + rng.Intn(3)
	h := 2 + rng.Intn(3)

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			if x+dx < size && y+dy < size {
				rgba.Set(x+dx, y+dy, glitchColor)
			}
		}
	}
}

// renderPostApocDamage renders rust, dirt, and wear.
func (sys *System) renderPostApocDamage(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	sys.renderRustPatches(rgba, comp, size, rng)
	sys.renderDirtStreaks(rgba, comp, size, rng)
}

// renderRustPatches draws rust damage patches across the surface.
func (sys *System) renderRustPatches(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	rustColor := color.RGBA{R: 140, G: 70, B: 30, A: 180}
	rustCount := comp.DamageLevel * 2

	for i := 0; i < rustCount; i++ {
		cx := rng.Intn(size)
		cy := rng.Intn(size)
		radius := 2 + rng.Intn(comp.DamageLevel+1)
		sys.drawIrregularPatch(rgba, cx, cy, radius, size, rustColor, rng)
	}
}

// drawIrregularPatch renders an irregular circular patch with random gaps.
func (sys *System) drawIrregularPatch(rgba *image.RGBA, cx, cy, radius, size int, col color.RGBA, rng *rand.Rand) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			distSq := dx*dx + dy*dy
			if distSq <= radius*radius && rng.Float64() < 0.7 {
				x := cx + dx
				y := cy + dy
				if x >= 0 && x < size && y >= 0 && y < size {
					rgba.Set(x, y, col)
				}
			}
		}
	}
}

// renderDirtStreaks draws vertical dirt streaks for higher damage levels.
func (sys *System) renderDirtStreaks(rgba *image.RGBA, comp *Component, size int, rng *rand.Rand) {
	if comp.DamageLevel < 2 {
		return
	}

	dirtColor := color.RGBA{R: 80, G: 70, B: 60, A: 160}
	streakCount := comp.DamageLevel

	for i := 0; i < streakCount; i++ {
		sys.drawDirtStreak(rgba, size, dirtColor, rng)
	}
}

// drawDirtStreak renders a single dirt streak with mostly vertical orientation.
func (sys *System) drawDirtStreak(rgba *image.RGBA, size int, col color.RGBA, rng *rand.Rand) {
	startX := rng.Intn(size)
	startY := rng.Intn(size / 2)
	streakLen := 2 + rng.Intn(size/3)
	angle := math.Pi/2 + (rng.Float64()-0.5)*math.Pi/6

	for j := 0; j < streakLen; j++ {
		x := startX + int(float64(j)*math.Sin(angle))
		y := startY + int(float64(j)*math.Cos(angle))
		if x >= 0 && x < size && y >= 0 && y < size {
			rgba.Set(x, y, col)
		}
	}
}

// fillCircle fills a circle at (cx, cy) with the given radius and color.
func (sys *System) fillCircle(rgba *image.RGBA, cx, cy, radius int, col color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				x := cx + dx
				y := cy + dy
				if x >= 0 && x < rgba.Bounds().Dx() && y >= 0 && y < rgba.Bounds().Dy() {
					rgba.Set(x, y, col)
				}
			}
		}
	}
}

// clamp constrains a value between min and max.
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
