package eyeglint

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

// System manages eye glint rendering for creature sprites.
type System struct {
	genreID    string
	preset     GenrePreset
	glintCache map[cacheKey]*ebiten.Image
	cacheMu    sync.RWMutex
	logger     *logrus.Entry
	time       float64
}

type cacheKey struct {
	eyeRadius int
	genreID   string
	frame     int
}

// NewSystem creates an eye glint rendering system for the given genre.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:    genreID,
		preset:     GetPreset(genreID),
		glintCache: make(map[cacheKey]*ebiten.Image),
		logger: logrus.WithFields(logrus.Fields{
			"system":  "eyeglint",
			"package": "eyeglint",
		}),
		time: 0.0,
	}
	s.logger.Debug("Eye glint system initialized")
	return s
}

// SetGenre updates the genre and reconfigures the preset.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = GetPreset(genreID)
	s.logger.WithField("genre", genreID).Debug("Genre updated")
}

// Update advances glint animation state for all entities with eye glint components.
func (s *System) Update(w *engine.World) {
	s.time += 0.0167

	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, entity := range entities {
		compRaw, ok := w.GetComponent(entity, compType)
		if !ok {
			continue
		}
		comp, ok := compRaw.(*Component)
		if !ok || !comp.Enabled {
			continue
		}

		comp.GlintPhase += s.preset.AnimationSpeed * 0.0167
		if comp.GlintPhase > 2*math.Pi {
			comp.GlintPhase -= 2 * math.Pi
		}
	}
}

// ApplyEyeGlints adds wet highlight reflections to detected eyes in a sprite.
func (s *System) ApplyEyeGlints(sprite *ebiten.Image, comp *Component, seed int64) *ebiten.Image {
	if sprite == nil || comp == nil || !comp.Enabled || comp.EyeCount() == 0 {
		return sprite
	}

	bounds := sprite.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	result := ebiten.NewImage(width, height)
	result.DrawImage(sprite, nil)

	rng := rand.New(rand.NewSource(seed))

	for i, eyePos := range comp.EyePositions {
		if i >= len(comp.EyeSizes) {
			break
		}
		eyeRadius := comp.EyeSizes[i]
		if eyeRadius < 1 {
			eyeRadius = 2
		}

		s.renderEyeGlint(result, eyePos[0], eyePos[1], eyeRadius, comp.GlintPhase, comp.GlintIntensity, rng)
	}

	return result
}

// renderEyeGlint draws the wet highlight reflections for a single eye.
func (s *System) renderEyeGlint(img *ebiten.Image, eyeX, eyeY, eyeRadius int, phase, intensity float64, rng *rand.Rand) {
	offsetX := -s.preset.PrimaryOffset * float64(eyeRadius)
	offsetY := -s.preset.PrimaryOffset * float64(eyeRadius)

	animOffsetX := math.Sin(phase) * s.preset.AnimationAmplitude * float64(eyeRadius)
	animOffsetY := math.Cos(phase*1.3) * s.preset.AnimationAmplitude * float64(eyeRadius)

	primaryX := float64(eyeX) + offsetX + animOffsetX
	primaryY := float64(eyeY) + offsetY + animOffsetY

	primaryRadius := s.preset.PrimarySize * float64(eyeRadius)
	if primaryRadius < 0.5 {
		primaryRadius = 0.5
	}

	if s.preset.GlowRadius > 0 {
		glowRadius := primaryRadius + s.preset.GlowRadius
		s.drawGlowCircle(img, int(primaryX), int(primaryY), glowRadius, s.preset.PrimaryColor, intensity*0.4)
	}

	s.drawGlintCircle(img, int(primaryX), int(primaryY), primaryRadius, s.preset.PrimaryColor, intensity)

	secondaryX := float64(eyeX) + offsetX*0.3 - animOffsetX*0.5
	secondaryY := float64(eyeY) + offsetY*0.3 - animOffsetY*0.5
	secondaryRadius := s.preset.SecondarySize * float64(eyeRadius)
	if secondaryRadius >= 0.3 {
		s.drawGlintCircle(img, int(secondaryX), int(secondaryY), secondaryRadius, s.preset.SecondaryColor, intensity*0.7)
	}

	_ = rng // Used for future random variation
}

// drawGlintCircle renders a small circular highlight with soft edges.
func (s *System) drawGlintCircle(img *ebiten.Image, cx, cy int, radius float64, col color.RGBA, intensity float64) {
	bounds := img.Bounds()
	r := int(math.Ceil(radius)) + 1

	size := r*2 + 2
	highlight := image.NewRGBA(image.Rect(0, 0, size, size))

	centerX := size / 2
	centerY := size / 2

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > radius {
				continue
			}

			falloff := 1.0 - (dist / radius)
			falloff = falloff * falloff

			alpha := uint8(float64(col.A) * falloff * intensity)
			if alpha < 1 {
				continue
			}

			px := centerX + dx
			py := centerY + dy
			if px >= 0 && px < size && py >= 0 && py < size {
				highlight.Set(px, py, color.RGBA{R: col.R, G: col.G, B: col.B, A: alpha})
			}
		}
	}

	destX := cx - centerX
	destY := cy - centerY
	if destX < bounds.Min.X-size || destX > bounds.Max.X ||
		destY < bounds.Min.Y-size || destY > bounds.Max.Y {
		return
	}

	ebitenHighlight := ebiten.NewImageFromImage(highlight)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(destX), float64(destY))
	op.CompositeMode = ebiten.CompositeModeLighter
	img.DrawImage(ebitenHighlight, op)
}

// drawGlowCircle renders a soft glow around the highlight.
func (s *System) drawGlowCircle(img *ebiten.Image, cx, cy int, radius float64, col color.RGBA, intensity float64) {
	bounds := img.Bounds()
	r := int(math.Ceil(radius)) + 1

	size := r*2 + 2
	glow := image.NewRGBA(image.Rect(0, 0, size, size))

	centerX := size / 2
	centerY := size / 2

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > radius {
				continue
			}

			falloff := 1.0 - (dist / radius)
			falloff = falloff * falloff * falloff

			alpha := uint8(float64(col.A) * falloff * intensity * 0.3)
			if alpha < 1 {
				continue
			}

			px := centerX + dx
			py := centerY + dy
			if px >= 0 && px < size && py >= 0 && py < size {
				glow.Set(px, py, color.RGBA{R: col.R, G: col.G, B: col.B, A: alpha})
			}
		}
	}

	destX := cx - centerX
	destY := cy - centerY
	if destX < bounds.Min.X-size || destX > bounds.Max.X ||
		destY < bounds.Min.Y-size || destY > bounds.Max.Y {
		return
	}

	ebitenGlow := ebiten.NewImageFromImage(glow)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(destX), float64(destY))
	op.CompositeMode = ebiten.CompositeModeLighter
	img.DrawImage(ebitenGlow, op)
}

type eyeCandidate struct {
	x      int
	y      int
	radius int
}

// DetectEyes analyzes a sprite to find eye regions based on color and position.
// Note: This method calls sprite.At() which requires the Ebiten game loop to be running.
// For testing or pre-game detection, use DetectEyesFromRGBA instead.
func (s *System) DetectEyes(sprite *ebiten.Image, creatureType string) *Component {
	if sprite == nil {
		return NewComponent()
	}

	comp := NewComponent()
	comp.CreatureType = creatureType

	bounds := sprite.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	rgba := image.NewRGBA(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, sprite.At(x, y))
		}
	}

	s.detectEyesInImage(rgba, comp, creatureType)
	return comp
}

// DetectEyesFromRGBA analyzes an RGBA image to find eye regions.
// This is safe to call outside the game loop (e.g., during sprite generation or testing).
func (s *System) DetectEyesFromRGBA(img *image.RGBA, creatureType string) *Component {
	if img == nil {
		return NewComponent()
	}

	comp := NewComponent()
	comp.CreatureType = creatureType

	s.detectEyesInImage(img, comp, creatureType)
	return comp
}

// detectEyesInImage performs the actual eye detection on an RGBA image.
func (s *System) detectEyesInImage(rgba *image.RGBA, comp *Component, creatureType string) {
	switch creatureType {
	case "humanoid", "skeleton", "zombie":
		s.detectHumanoidEyes(rgba, comp)
	case "quadruped", "wolf", "bear":
		s.detectQuadrupedEyes(rgba, comp)
	case "insect", "spider", "scorpion":
		s.detectInsectEyes(rgba, comp)
	case "serpent", "snake", "dragon":
		s.detectSerpentEyes(rgba, comp)
	case "flying", "bat", "bird":
		s.detectFlyingEyes(rgba, comp)
	case "amorphous", "slime", "ooze":
		s.detectAmorphousEyes(rgba, comp)
	default:
		s.detectGenericEyes(rgba, comp)
	}
}

func (s *System) detectHumanoidEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	searchTop := height / 6
	searchBot := height * 2 / 5
	searchLeft := width / 4
	searchRight := width * 3 / 4

	eyes := s.findEyePixels(img, searchLeft, searchTop, searchRight, searchBot)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

func (s *System) detectQuadrupedEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	searchTop := height / 8
	searchBot := height / 3
	searchLeft := width / 3
	searchRight := width * 2 / 3

	eyes := s.findEyePixels(img, searchLeft, searchTop, searchRight, searchBot)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

func (s *System) detectInsectEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	searchTop := 0
	searchBot := height / 3
	searchLeft := width / 6
	searchRight := width * 5 / 6

	eyes := s.findEyePixels(img, searchLeft, searchTop, searchRight, searchBot)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

func (s *System) detectSerpentEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	searchTop := height / 4
	searchBot := height / 2
	searchLeft := width / 4
	searchRight := width * 3 / 4

	eyes := s.findEyePixels(img, searchLeft, searchTop, searchRight, searchBot)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

func (s *System) detectFlyingEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	searchTop := 0
	searchBot := height / 3
	searchLeft := width / 4
	searchRight := width * 3 / 4

	eyes := s.findEyePixels(img, searchLeft, searchTop, searchRight, searchBot)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

func (s *System) detectAmorphousEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	eyes := s.findEyePixels(img, 0, 0, width, height)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

func (s *System) detectGenericEyes(img *image.RGBA, comp *Component) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	searchTop := height / 6
	searchBot := height / 2
	searchLeft := width / 5
	searchRight := width * 4 / 5

	eyes := s.findEyePixels(img, searchLeft, searchTop, searchRight, searchBot)
	for _, eye := range eyes {
		comp.AddEye(eye.x, eye.y, eye.radius)
	}
}

// findEyePixels searches for eye-colored regions within the specified bounds.
func (s *System) findEyePixels(img *image.RGBA, left, top, right, bottom int) []eyeCandidate {
	eyes := make([]eyeCandidate, 0, 4)

	bounds := img.Bounds()
	if left < bounds.Min.X {
		left = bounds.Min.X
	}
	if top < bounds.Min.Y {
		top = bounds.Min.Y
	}
	if right > bounds.Max.X {
		right = bounds.Max.X
	}
	if bottom > bounds.Max.Y {
		bottom = bounds.Max.Y
	}

	visited := make(map[int]bool)

	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			idx := y*img.Stride + x*4
			if visited[idx] {
				continue
			}

			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

			if a8 < 128 {
				continue
			}

			if s.isEyeColor(r8, g8, b8) {
				cx, cy, radius := s.measureEyeRegion(img, x, y, left, top, right, bottom, visited)
				if radius >= 1 {
					eyes = append(eyes, eyeCandidate{x: cx, y: cy, radius: radius})
				}
			}
		}
	}

	return eyes
}

// isEyeColor checks if a color looks like an eye color.
func (s *System) isEyeColor(r, g, b uint8) bool {
	// White/bright (sclera or highlight)
	if r > 200 && g > 200 && b > 200 {
		return false // Skip sclera, we want pupil/iris
	}

	// Yellow/gold eyes (common for creatures)
	if r > 200 && g > 150 && b < 100 {
		return true
	}

	// Red eyes (hostile creatures)
	if r > 200 && g < 100 && b < 100 {
		return true
	}

	// Green eyes
	if g > 150 && r < 150 && b < 150 {
		return true
	}

	// Blue eyes
	if b > 150 && r < 150 && g < 200 {
		return true
	}

	// Orange eyes
	if r > 200 && g > 100 && g < 180 && b < 80 {
		return true
	}

	// Dark/black (pupil)
	if r < 80 && g < 80 && b < 80 {
		return true
	}

	return false
}

// measureEyeRegion flood-fills from an eye pixel to find the eye center and radius.
func (s *System) measureEyeRegion(img *image.RGBA, startX, startY, left, top, right, bottom int, visited map[int]bool) (int, int, int) {
	stack := [][2]int{{startX, startY}}
	pixels := make([][2]int, 0, 32)

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		x, y := current[0], current[1]
		if x < left || x >= right || y < top || y >= bottom {
			continue
		}

		idx := y*img.Stride + x*4
		if visited[idx] {
			continue
		}
		visited[idx] = true

		c := img.At(x, y)
		r, g, b, a := c.RGBA()
		r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

		if a8 < 128 {
			continue
		}

		if !s.isEyeColor(r8, g8, b8) && !s.isSimilarColor(r8, g8, b8, startX, startY, img) {
			continue
		}

		pixels = append(pixels, [2]int{x, y})

		// Flood fill neighbors (4-connectivity)
		stack = append(stack, [2]int{x + 1, y})
		stack = append(stack, [2]int{x - 1, y})
		stack = append(stack, [2]int{x, y + 1})
		stack = append(stack, [2]int{x, y - 1})

		// Limit search
		if len(pixels) > 100 {
			break
		}
	}

	if len(pixels) < 2 {
		return 0, 0, 0
	}

	// Calculate centroid
	sumX, sumY := 0, 0
	for _, p := range pixels {
		sumX += p[0]
		sumY += p[1]
	}
	cx := sumX / len(pixels)
	cy := sumY / len(pixels)

	// Calculate radius (distance to farthest pixel)
	maxDist := 0.0
	for _, p := range pixels {
		dx := float64(p[0] - cx)
		dy := float64(p[1] - cy)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > maxDist {
			maxDist = dist
		}
	}

	radius := int(math.Ceil(maxDist))
	if radius < 1 {
		radius = 1
	}
	if radius > 10 {
		radius = 10 // Cap eye size
	}

	return cx, cy, radius
}

// isSimilarColor checks if a color is similar to the starting eye color.
func (s *System) isSimilarColor(r, g, b uint8, startX, startY int, img *image.RGBA) bool {
	startC := img.At(startX, startY)
	sr, sg, sb, _ := startC.RGBA()
	sr8, sg8, sb8 := uint8(sr>>8), uint8(sg>>8), uint8(sb>>8)

	dr := int(r) - int(sr8)
	dg := int(g) - int(sg8)
	db := int(b) - int(sb8)

	dist := dr*dr + dg*dg + db*db
	return dist < 3000 // Allow some color variation within eye
}

// RenderEyeGlint draws a glint highlight at the specified position (for manual use).
func (s *System) RenderEyeGlint(screen *ebiten.Image, eyeX, eyeY, eyeRadius int, intensity float64) {
	rng := rand.New(rand.NewSource(int64(eyeX*1000 + eyeY)))
	s.renderEyeGlint(screen, eyeX, eyeY, eyeRadius, s.time, intensity, rng)
}

// Render draws eye glints for all visible entities with eye glint components.
// This is called during the render phase to apply glints to entity sprites.
func (s *System) Render(screen *ebiten.Image, w *engine.World) {
	if screen == nil || w == nil {
		return
	}

	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, entity := range entities {
		compRaw, ok := w.GetComponent(entity, compType)
		if !ok {
			continue
		}
		comp, ok := compRaw.(*Component)
		if !ok || !comp.Enabled {
			continue
		}

		// Skip if no detected eyes
		if comp.EyeCount() == 0 {
			continue
		}

		// Calculate intensity based on animation phase
		phaseIntensity := 0.5 + 0.5*math.Sin(comp.GlintPhase)
		finalIntensity := comp.GlintIntensity * phaseIntensity

		// Render glint for each detected eye at its screen position
		// Note: The actual screen position would need to come from sprite rendering
		// For now, the eye positions are stored relative to sprite
		// The caller should use ApplyEyeGlints during sprite compositing instead
		_ = entity
		_ = finalIntensity
	}
}

// GetPresetForGenre returns the current genre preset (for testing/debugging).
func (s *System) GetPresetForGenre() GenrePreset {
	return s.preset
}
