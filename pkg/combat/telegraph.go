// Package combat - Telegraph component and rendering
package combat

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/rng"
)

// TelegraphShape defines the visual shape of an attack telegraph.
type TelegraphShape int

const (
	ShapeCone   TelegraphShape = iota // Directional cone (melee swipe, flamethrower)
	ShapeCircle                       // Radial AoE (slam, explosion)
	ShapeLine                         // Beam/charge attack
	ShapeRing                         // Donut-shaped zone
)

// TelegraphPhase tracks the attack state machine.
type TelegraphPhase int

const (
	PhaseInactive TelegraphPhase = iota // Idle, waiting to attack
	PhaseWindup                         // Charging/warning
	PhaseActive                         // Damage window
	PhaseCooldown                       // Recovery before next attack
)

// AttackPattern defines a repeatable enemy attack behavior.
type AttackPattern struct {
	Name         string
	Shape        TelegraphShape
	Range        float64 // Max distance
	Angle        float64 // Cone angle in radians (for cone shape)
	Width        float64 // Line width or ring thickness
	WindupTime   float64 // Seconds of warning
	ActiveTime   float64 // Seconds damage window is open
	CooldownTime float64 // Seconds before can attack again
	Damage       float64 // Base damage
	DamageType   DamageType
	KnockbackMul float64 // Knockback strength multiplier
}

// TelegraphComponent stores attack state for an entity.
type TelegraphComponent struct {
	Pattern    AttackPattern
	Phase      TelegraphPhase
	PhaseTimer float64 // Counts down to phase transition
	TargetX    float64 // Attack target position
	TargetY    float64
	DirectionX float64 // Attack direction vector (normalized)
	DirectionY float64
	HasHit     bool  // Prevents multi-hit in active phase
	Seed       int64 // For procedural pattern variation
}

// TelegraphColors returns genre-specific color scheme.
func TelegraphColors(genreID string, phase TelegraphPhase) (warning, active color.Color) {
	alpha := uint8(128)
	if phase == PhaseActive {
		alpha = 200
	}

	switch genreID {
	case "fantasy":
		warning = color.RGBA{255, 140, 0, alpha} // Orange warning
		active = color.RGBA{255, 0, 0, alpha}    // Red damage
	case "scifi":
		warning = color.RGBA{0, 200, 255, alpha} // Cyan warning
		active = color.RGBA{255, 0, 255, alpha}  // Magenta damage
	case "horror":
		warning = color.RGBA{160, 0, 160, alpha} // Purple warning
		active = color.RGBA{140, 0, 0, alpha}    // Dark red damage
	case "cyberpunk":
		warning = color.RGBA{0, 255, 255, alpha} // Cyan warning
		active = color.RGBA{255, 0, 128, alpha}  // Hot pink damage
	default:
		warning = color.RGBA{255, 140, 0, alpha}
		active = color.RGBA{255, 0, 0, alpha}
	}
	return warning, active
}

// RenderTelegraph draws the attack telegraph on screen.
func RenderTelegraph(screen *ebiten.Image, x, y, camX, camY float64, comp *TelegraphComponent, genreID string) {
	if comp.Phase == PhaseInactive || comp.Phase == PhaseCooldown {
		return
	}

	screenX := float32(x - camX + 320)
	screenY := float32(y - camY + 180)

	_, activeColor := TelegraphColors(genreID, comp.Phase)
	col := activeColor
	if comp.Phase == PhaseWindup {
		// Pulsing opacity during windup
		pulseFactor := 0.5 + 0.5*math.Sin(comp.PhaseTimer*8)
		r, g, b, a := col.RGBA()
		col = color.RGBA{
			uint8(r >> 8),
			uint8(g >> 8),
			uint8(b >> 8),
			uint8(float64(a>>8) * pulseFactor),
		}
	}

	switch comp.Pattern.Shape {
	case ShapeCone:
		renderCone(screen, screenX, screenY, comp, col)
	case ShapeCircle:
		renderCircle(screen, screenX, screenY, comp, col)
	case ShapeLine:
		renderLine(screen, screenX, screenY, comp, col)
	case ShapeRing:
		renderRing(screen, screenX, screenY, comp, col)
	}
}

func renderCone(screen *ebiten.Image, x, y float32, comp *TelegraphComponent, col color.Color) {
	// Direction angle
	angle := math.Atan2(comp.DirectionY, comp.DirectionX)
	halfAngle := comp.Pattern.Angle / 2

	// Draw cone as filled triangle fan
	segments := 16
	for i := 0; i < segments; i++ {
		t1 := -halfAngle + (float64(i)/float64(segments))*comp.Pattern.Angle
		t2 := -halfAngle + (float64(i+1)/float64(segments))*comp.Pattern.Angle

		x1 := x + float32(math.Cos(angle+t1)*comp.Pattern.Range)
		y1 := y + float32(math.Sin(angle+t1)*comp.Pattern.Range)
		x2 := x + float32(math.Cos(angle+t2)*comp.Pattern.Range)
		y2 := y + float32(math.Sin(angle+t2)*comp.Pattern.Range)

		vector.StrokeLine(screen, x, y, x1, y1, 2, col, false)
		vector.DrawFilledRect(screen, x, y, x1-x, y1-y, col, false)
		vector.StrokeLine(screen, x1, y1, x2, y2, 1, col, false)
	}
	// Outline
	vector.StrokeLine(screen, x, y,
		x+float32(math.Cos(angle-halfAngle)*comp.Pattern.Range),
		y+float32(math.Sin(angle-halfAngle)*comp.Pattern.Range), 2, col, false)
	vector.StrokeLine(screen, x, y,
		x+float32(math.Cos(angle+halfAngle)*comp.Pattern.Range),
		y+float32(math.Sin(angle+halfAngle)*comp.Pattern.Range), 2, col, false)
}

func renderCircle(screen *ebiten.Image, x, y float32, comp *TelegraphComponent, col color.Color) {
	// Pulsing circle that grows during windup
	growFactor := 1.0
	if comp.Phase == PhaseWindup {
		growFactor = 1.0 - (comp.PhaseTimer/comp.Pattern.WindupTime)*0.3
	}
	radius := float32(comp.Pattern.Range * growFactor)

	vector.StrokeCircle(screen, x, y, radius, 3, col, false)
	// Inner fill with lower opacity
	r, g, b, a := col.RGBA()
	fillCol := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a>>8) / 2}
	vector.DrawFilledCircle(screen, x, y, radius, fillCol, false)
}

func renderLine(screen *ebiten.Image, x, y float32, comp *TelegraphComponent, col color.Color) {
	// Beam telegraph - line in direction of attack
	endX := x + float32(comp.DirectionX*comp.Pattern.Range)
	endY := y + float32(comp.DirectionY*comp.Pattern.Range)

	width := float32(comp.Pattern.Width)
	if comp.Phase == PhaseWindup {
		// Narrow during windup, full width during active
		width *= float32(1.0 - comp.PhaseTimer/comp.Pattern.WindupTime)
	}

	vector.StrokeLine(screen, x, y, endX, endY, width, col, false)
}

func renderRing(screen *ebiten.Image, x, y float32, comp *TelegraphComponent, col color.Color) {
	// Ring (donut) - outer circle minus inner circle
	outerRadius := float32(comp.Pattern.Range)
	innerRadius := float32(comp.Pattern.Range - comp.Pattern.Width)

	vector.StrokeCircle(screen, x, y, outerRadius, 2, col, false)
	vector.StrokeCircle(screen, x, y, innerRadius, 2, col, false)
}

// DefaultPatterns returns genre-aware attack patterns for common enemy types.
func DefaultPatterns(genreID string, rng *rng.RNG) []AttackPattern {
	patterns := []AttackPattern{
		{
			Name:         "Melee Swipe",
			Shape:        ShapeCone,
			Range:        40,
			Angle:        math.Pi / 3,
			WindupTime:   0.4,
			ActiveTime:   0.15,
			CooldownTime: 1.0,
			Damage:       15,
			DamageType:   DamagePhysical,
			KnockbackMul: 1.5,
		},
		{
			Name:         "Ground Slam",
			Shape:        ShapeCircle,
			Range:        60,
			WindupTime:   0.8,
			ActiveTime:   0.2,
			CooldownTime: 2.5,
			Damage:       30,
			DamageType:   DamagePhysical,
			KnockbackMul: 3.0,
		},
		{
			Name:         "Charge Attack",
			Shape:        ShapeLine,
			Range:        120,
			Width:        20,
			WindupTime:   0.6,
			ActiveTime:   0.3,
			CooldownTime: 2.0,
			Damage:       25,
			DamageType:   DamagePhysical,
			KnockbackMul: 4.0,
		},
		{
			Name:         "Shockwave",
			Shape:        ShapeRing,
			Range:        80,
			Width:        20,
			WindupTime:   0.5,
			ActiveTime:   0.25,
			CooldownTime: 3.0,
			Damage:       20,
			DamageType:   DamageEnergy,
			KnockbackMul: 2.0,
		},
	}

	// Genre-specific adjustments
	switch genreID {
	case "scifi":
		patterns[0].DamageType = DamageEnergy
		patterns[2].Name = "Laser Beam"
		patterns[2].DamageType = DamagePlasma
		patterns[3].DamageType = DamagePlasma
	case "horror":
		patterns[0].Damage *= 1.3
		patterns[1].Name = "Tentacle Slam"
		patterns[1].Range = 70
	case "cyberpunk":
		patterns[2].Name = "Cyber Dash"
		patterns[2].DamageType = DamageEnergy
		patterns[3].Name = "EMP Pulse"
	}

	return patterns
}

// SelectPattern chooses an attack pattern based on distance and context.
func SelectPattern(patterns []AttackPattern, distToTarget float64, rng *rng.RNG) AttackPattern {
	// Filter to patterns that can reach the target
	valid := make([]AttackPattern, 0, len(patterns))
	for _, p := range patterns {
		if p.Range >= distToTarget {
			valid = append(valid, p)
		}
	}

	if len(valid) == 0 {
		// Default to longest range pattern if nothing in range
		return patterns[len(patterns)-1]
	}

	// Weighted random selection favoring closer-range attacks when in range
	idx := rng.Intn(len(valid))
	return valid[idx]
}
