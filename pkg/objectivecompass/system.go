package objectivecompass

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sirupsen/logrus"
)

// System manages objective compass indicators for navigation.
type System struct {
	genreID string
	style   GenreStyle
	styles  map[string]GenreStyle
	logger  *logrus.Entry

	// Objectives tracked by the compass
	objectives map[string]*Component

	// Screen dimensions
	screenWidth  int
	screenHeight int

	// Player state
	playerX     float64
	playerY     float64
	playerAngle float64 // View angle in radians

	// Global time for animations
	time float64

	// On-screen threshold (radial distance from center to consider "on screen")
	onScreenThreshold float64

	// Minimum indicator spacing to prevent overlap
	minIndicatorSpacing float32

	// Cached 1x1 white image for triangle drawing
	whiteImg *ebiten.Image
}

// NewSystem creates an objective compass system.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:             genreID,
		styles:              DefaultGenreStyles(),
		objectives:          make(map[string]*Component),
		screenWidth:         320,
		screenHeight:        200,
		onScreenThreshold:   0.7, // 70% of half-screen is "on screen"
		minIndicatorSpacing: 16.0,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "objectivecompass",
			"package": "objectivecompass",
		}),
	}
	s.applyGenreStyle()
	return s
}

// SetGenre updates genre-specific visual style.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenreStyle()
	s.logger.WithField("genre", genreID).Debug("Genre updated")
}

// applyGenreStyle loads the style preset for the current genre.
func (s *System) applyGenreStyle() {
	if style, ok := s.styles[s.genreID]; ok {
		s.style = style
	} else {
		s.style = s.styles["fantasy"]
	}
}

// SetScreenSize updates screen dimensions for indicator positioning.
func (s *System) SetScreenSize(width, height int) {
	s.screenWidth = width
	s.screenHeight = height
}

// SetPlayerPosition updates player position and view angle.
// Angle is in radians (0 = facing right/east, π/2 = north, etc.).
func (s *System) SetPlayerPosition(x, y, angle float64) {
	s.playerX = x
	s.playerY = y
	s.playerAngle = angle
}

// AddObjective registers an objective for compass tracking.
func (s *System) AddObjective(id string, objType ObjectiveType, x, y float64) {
	comp := NewComponent(id, objType, x, y)
	s.objectives[id] = comp
	s.logger.WithFields(logrus.Fields{
		"objective": id,
		"type":      objType,
		"x":         x,
		"y":         y,
	}).Debug("Objective registered")
}

// RemoveObjective unregisters an objective from compass tracking.
func (s *System) RemoveObjective(id string) {
	delete(s.objectives, id)
}

// CompleteObjective marks an objective as completed (will fade out).
func (s *System) CompleteObjective(id string) {
	if obj, ok := s.objectives[id]; ok {
		obj.Completed = true
	}
}

// UpdateObjectivePosition updates the world position of an objective.
func (s *System) UpdateObjectivePosition(id string, x, y float64) {
	if obj, ok := s.objectives[id]; ok {
		obj.WorldX = x
		obj.WorldY = y
	}
}

// ClearObjectives removes all tracked objectives.
func (s *System) ClearObjectives() {
	s.objectives = make(map[string]*Component)
}

// Update advances animations and computes indicator positions.
func (s *System) Update(deltaTime float64) {
	s.time += deltaTime

	halfW := float64(s.screenWidth) / 2
	halfH := float64(s.screenHeight) / 2
	padding := float64(s.style.EdgePadding)

	// Calculate screen-relative info for each objective
	for _, obj := range s.objectives {
		// Fade completed objectives
		if obj.Completed {
			obj.Alpha = lerpf(obj.Alpha, 0, deltaTime*3.0)
			continue
		}

		// Calculate direction from player to objective
		dx := obj.WorldX - s.playerX
		dy := obj.WorldY - s.playerY
		obj.Distance = math.Sqrt(dx*dx + dy*dy)

		// Angle from player position to objective in world space
		worldAngle := math.Atan2(-dy, dx) // Negative Y because screen Y is inverted

		// Relative angle considering player view direction
		relativeAngle := normalizeAngle(worldAngle - s.playerAngle)
		obj.ScreenAngle = relativeAngle

		// Check if objective is roughly on-screen (within FOV)
		// Assume ~90 degree FOV, so ±π/4 is on-screen
		fovHalf := math.Pi / 4
		obj.OnScreen = math.Abs(relativeAngle) < fovHalf && obj.Distance < 15

		// Update pulse animation
		obj.PulsePhase += s.style.PulseSpeed * deltaTime
		if obj.PulsePhase > 2*math.Pi {
			obj.PulsePhase -= 2 * math.Pi
		}

		// Calculate target alpha based on distance
		distanceFactor := 1.0 - math.Min(1.0, obj.Distance/s.style.MaxDistance)
		targetAlpha := s.style.MinAlpha + (1.0-s.style.MinAlpha)*distanceFactor

		// Hide on-screen objectives
		if obj.OnScreen {
			targetAlpha = 0
		}

		obj.Alpha = lerpf(obj.Alpha, targetAlpha, deltaTime*4.0)

		// Calculate scale based on distance (closer = larger)
		obj.Scale = 0.7 + 0.5*distanceFactor

		// Calculate screen-edge position
		obj.EdgeX, obj.EdgeY = s.calculateEdgePosition(relativeAngle, halfW, halfH, padding)
	}

	// Resolve overlapping indicators
	s.resolveOverlaps()
}

// calculateEdgePosition finds where an indicator should be placed on the screen edge.
func (s *System) calculateEdgePosition(angle, halfW, halfH, padding float64) (float32, float32) {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	// Effective screen half-dimensions accounting for padding
	effectiveHalfW := halfW - padding
	effectiveHalfH := halfH - padding

	var edgeX, edgeY float64

	// Find intersection with screen edge
	if math.Abs(cos) > 0.001 {
		// Check horizontal edges first
		t := effectiveHalfW / math.Abs(cos)
		testY := t * math.Abs(sin)

		if testY <= effectiveHalfH {
			// Hits left or right edge
			if cos > 0 {
				edgeX = halfW + effectiveHalfW
			} else {
				edgeX = halfW - effectiveHalfW
			}
			if sin > 0 {
				edgeY = halfH - t*sin
			} else {
				edgeY = halfH - t*sin
			}
		} else {
			// Hits top or bottom edge
			t = effectiveHalfH / math.Abs(sin)
			edgeX = halfW + t*cos
			if sin > 0 {
				edgeY = halfH - effectiveHalfH
			} else {
				edgeY = halfH + effectiveHalfH
			}
		}
	} else {
		// Nearly vertical direction
		edgeX = halfW
		if sin > 0 {
			edgeY = halfH - effectiveHalfH
		} else {
			edgeY = halfH + effectiveHalfH
		}
	}

	return float32(edgeX), float32(edgeY)
}

// resolveOverlaps adjusts indicators that are too close together.
func (s *System) resolveOverlaps() {
	// Simple approach: push overlapping indicators apart along the edge
	objectives := make([]*Component, 0, len(s.objectives))
	for _, obj := range s.objectives {
		if obj.Alpha > 0.05 && !obj.Completed {
			objectives = append(objectives, obj)
		}
	}

	for i := 0; i < len(objectives); i++ {
		for j := i + 1; j < len(objectives); j++ {
			obj1, obj2 := objectives[i], objectives[j]

			dx := obj2.EdgeX - obj1.EdgeX
			dy := obj2.EdgeY - obj1.EdgeY
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

			if dist < s.minIndicatorSpacing && dist > 0.001 {
				// Push apart
				overlap := s.minIndicatorSpacing - dist
				pushX := dx / dist * overlap * 0.5
				pushY := dy / dist * overlap * 0.5

				// Lower priority objective moves
				if obj1.ObjType > obj2.ObjType {
					obj1.EdgeX -= pushX
					obj1.EdgeY -= pushY
				} else {
					obj2.EdgeX += pushX
					obj2.EdgeY += pushY
				}
			}
		}
	}
}

// Render draws all active compass indicators.
func (s *System) Render(screen *ebiten.Image) {
	if s.whiteImg == nil {
		s.whiteImg = ebiten.NewImage(1, 1)
		s.whiteImg.Fill(color.White)
	}

	// Sort by type to ensure main objectives render on top
	// (Lower type = higher priority = render last)
	var mainObjs, otherObjs []*Component
	for _, obj := range s.objectives {
		if obj.Alpha < 0.02 {
			continue
		}
		if obj.ObjType == TypeMain || obj.ObjType == TypeExit {
			mainObjs = append(mainObjs, obj)
		} else {
			otherObjs = append(otherObjs, obj)
		}
	}

	// Render lower priority first
	for _, obj := range otherObjs {
		s.renderIndicator(screen, obj)
	}
	for _, obj := range mainObjs {
		s.renderIndicator(screen, obj)
	}
}

// renderIndicator draws a single compass indicator.
func (s *System) renderIndicator(screen *ebiten.Image, obj *Component) {
	if obj.Alpha < 0.02 {
		return
	}

	// Get color based on type
	col := s.getObjectiveColor(obj.ObjType)
	col.A = uint8(float64(col.A) * obj.Alpha)

	// Calculate pulse effect
	pulseScale := float32(1.0 + 0.12*math.Sin(obj.PulsePhase))
	size := s.style.ArrowSize * float32(obj.Scale) * pulseScale

	// Draw glow layer first (if genre supports it)
	if s.style.GlowIntensity > 0 {
		s.drawIndicatorGlow(screen, obj.EdgeX, obj.EdgeY, obj.ScreenAngle, size, col, obj.Alpha)
	}

	// Draw main indicator arrow
	s.drawIndicatorArrow(screen, obj.EdgeX, obj.EdgeY, obj.ScreenAngle, size, col)

	// Draw distance tick marks for main objectives
	if obj.ObjType == TypeMain || obj.ObjType == TypeExit {
		s.drawDistanceTicks(screen, obj.EdgeX, obj.EdgeY, obj.ScreenAngle, float64(size), obj.Distance, col)
	}
}

// getObjectiveColor returns the color for an objective type.
func (s *System) getObjectiveColor(objType ObjectiveType) color.RGBA {
	switch objType {
	case TypeMain:
		return s.style.MainColor
	case TypeBonus:
		return s.style.BonusColor
	case TypePOI:
		return s.style.POIColor
	case TypeExit:
		return s.style.ExitColor
	default:
		return s.style.MainColor
	}
}

// drawIndicatorGlow renders a soft glow behind the indicator.
func (s *System) drawIndicatorGlow(screen *ebiten.Image, x, y float32, angle float64, size float32, col color.RGBA, alpha float64) {
	glowSize := size * 1.8
	glowAlpha := uint8(float64(col.A) * s.style.GlowIntensity * alpha * 0.5)
	glowCol := color.RGBA{R: col.R, G: col.G, B: col.B, A: glowAlpha}

	// Draw as a simple filled circle
	s.drawCircleFilled(screen, x, y, glowSize*0.6, glowCol)
}

// drawIndicatorArrow draws the triangular directional indicator.
func (s *System) drawIndicatorArrow(screen *ebiten.Image, x, y float32, angle float64, size float32, col color.RGBA) {
	// Arrow points inward toward screen center
	inwardAngle := angle + math.Pi

	cos := float32(math.Cos(inwardAngle))
	sin := float32(math.Sin(inwardAngle))

	// Tip of arrow (pointing inward)
	tipX := x + cos*size*0.8
	tipY := y - sin*size*0.8

	// Base vertices
	perpCos := float32(math.Cos(inwardAngle + math.Pi/2))
	perpSin := float32(math.Sin(inwardAngle + math.Pi/2))
	baseOffset := size * 0.5

	base1X := x + perpCos*baseOffset
	base1Y := y - perpSin*baseOffset
	base2X := x - perpCos*baseOffset
	base2Y := y + perpSin*baseOffset

	// Draw filled triangle
	var path vector.Path
	path.MoveTo(tipX, tipY)
	path.LineTo(base1X, base1Y)
	path.LineTo(base2X, base2Y)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(col.R) / 255.0
		vs[i].ColorG = float32(col.G) / 255.0
		vs[i].ColorB = float32(col.B) / 255.0
		vs[i].ColorA = float32(col.A) / 255.0
	}
	screen.DrawTriangles(vs, is, s.whiteImg, &ebiten.DrawTrianglesOptions{})

	// Draw outline for better visibility
	outlineCol := color.RGBA{R: 0, G: 0, B: 0, A: col.A / 2}
	vector.StrokeLine(screen, tipX, tipY, base1X, base1Y, 1.0, outlineCol, false)
	vector.StrokeLine(screen, base1X, base1Y, base2X, base2Y, 1.0, outlineCol, false)
	vector.StrokeLine(screen, base2X, base2Y, tipX, tipY, 1.0, outlineCol, false)
}

// drawDistanceTicks renders small tick marks indicating distance.
func (s *System) drawDistanceTicks(screen *ebiten.Image, x, y float32, angle, size, distance float64, col color.RGBA) {
	// Number of ticks based on distance (1-3 ticks)
	numTicks := 3 - int(math.Min(2, distance/20.0))
	if numTicks < 1 {
		numTicks = 1
	}

	tickAlpha := col.A / 2
	tickCol := color.RGBA{R: col.R, G: col.G, B: col.B, A: tickAlpha}

	// Position ticks behind the arrow
	outwardAngle := angle
	cos := float32(math.Cos(outwardAngle))
	sin := float32(math.Sin(outwardAngle))

	perpCos := float32(math.Cos(outwardAngle + math.Pi/2))
	perpSin := float32(math.Sin(outwardAngle + math.Pi/2))

	tickLen := float32(size) * 0.3
	tickSpacing := float32(size) * 0.25

	for i := 0; i < numTicks; i++ {
		offset := float32(size)*0.4 + float32(i)*tickSpacing
		tickX := x - cos*offset
		tickY := y + sin*offset

		// Draw perpendicular tick
		t1X := tickX + perpCos*tickLen
		t1Y := tickY - perpSin*tickLen
		t2X := tickX - perpCos*tickLen
		t2Y := tickY + perpSin*tickLen

		vector.StrokeLine(screen, t1X, t1Y, t2X, t2Y, 1.5, tickCol, false)
	}
}

// drawCircleFilled draws a filled circle using triangles.
func (s *System) drawCircleFilled(screen *ebiten.Image, x, y, radius float32, col color.RGBA) {
	const segments = 12
	var path vector.Path

	for i := 0; i <= segments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(segments)
		px := x + radius*float32(math.Cos(angle))
		py := y + radius*float32(math.Sin(angle))
		if i == 0 {
			path.MoveTo(px, py)
		} else {
			path.LineTo(px, py)
		}
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(col.R) / 255.0
		vs[i].ColorG = float32(col.G) / 255.0
		vs[i].ColorB = float32(col.B) / 255.0
		vs[i].ColorA = float32(col.A) / 255.0
	}
	screen.DrawTriangles(vs, is, s.whiteImg, &ebiten.DrawTrianglesOptions{})
}

// GetObjectiveCount returns the number of tracked objectives.
func (s *System) GetObjectiveCount() int {
	return len(s.objectives)
}

// GetVisibleCount returns the number of currently visible indicators.
func (s *System) GetVisibleCount() int {
	count := 0
	for _, obj := range s.objectives {
		if obj.Alpha > 0.05 && !obj.OnScreen {
			count++
		}
	}
	return count
}

// GetStyle returns the current genre style (for testing).
func (s *System) GetStyle() GenreStyle {
	return s.style
}

// Helper functions

// normalizeAngle constrains an angle to [-π, π].
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// lerpf performs linear interpolation for float64.
func lerpf(a, b, t float64) float64 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return a + (b-a)*t
}
