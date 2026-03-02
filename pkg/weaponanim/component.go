// Package weaponanim provides visual animation for weapon attacks.
package weaponanim

import (
	"image/color"
	"math"
)

// SwingType defines the type of weapon swing animation.
type SwingType int

const (
	SwingSlash    SwingType = iota // Horizontal arc
	SwingOverhead                  // Vertical downward
	SwingThrust                    // Linear stab
	SwingUppercut                  // Upward arc
	SwingWide                      // Large horizontal sweep
)

// WeaponAnimComponent stores the current weapon swing animation state.
type WeaponAnimComponent struct {
	Active      bool      // Whether an animation is playing
	SwingType   SwingType // Type of swing
	Progress    float64   // 0.0 to 1.0
	Duration    float64   // Total animation time in seconds
	StartAngle  float64   // Starting angle in radians
	EndAngle    float64   // Ending angle in radians
	ArcRadius   float64   // Distance from wielder to weapon tip
	TrailPoints []TrailPoint
	Color       color.RGBA // Trail color (weapon-dependent)
	Width       float64    // Trail width in pixels
	Seed        int64      // For deterministic visual variation
}

// TrailPoint represents a point in the weapon's motion trail.
type TrailPoint struct {
	X        float64
	Y        float64
	Age      float64 // 0.0 (fresh) to 1.0 (old)
	Rotation float64 // Angle at this point
}

// Type implements Component interface.
func (w *WeaponAnimComponent) Type() string {
	return "weaponanim"
}

// GetCurrentAngle returns the weapon's current angle based on animation progress.
func (w *WeaponAnimComponent) GetCurrentAngle() float64 {
	if !w.Active {
		return w.StartAngle
	}

	// Use easing for more natural motion
	t := easeInOut(w.Progress)
	return w.StartAngle + (w.EndAngle-w.StartAngle)*t
}

// GetTipPosition calculates the weapon tip position relative to wielder.
func (w *WeaponAnimComponent) GetTipPosition(wielderX, wielderY float64) (float64, float64) {
	angle := w.GetCurrentAngle()
	tipX := wielderX + math.Cos(angle)*w.ArcRadius
	tipY := wielderY + math.Sin(angle)*w.ArcRadius
	return tipX, tipY
}

// easeInOut applies cubic easing to create natural acceleration/deceleration.
func easeInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// GetSwingParameters returns animation parameters for a given swing type.
func GetSwingParameters(swingType SwingType, facing float64) (startAngle, endAngle, duration float64) {
	switch swingType {
	case SwingSlash:
		// Right to left horizontal slash
		startAngle = facing - math.Pi/3
		endAngle = facing + math.Pi/3
		duration = 0.3

	case SwingOverhead:
		// Overhead downward chop
		startAngle = facing - math.Pi/2
		endAngle = facing + math.Pi/4
		duration = 0.4

	case SwingThrust:
		// Quick forward stab
		startAngle = facing
		endAngle = facing
		duration = 0.2

	case SwingUppercut:
		// Upward rising slash
		startAngle = facing + math.Pi/4
		endAngle = facing - math.Pi/2
		duration = 0.35

	case SwingWide:
		// Full arc sweep
		startAngle = facing - math.Pi*2/3
		endAngle = facing + math.Pi*2/3
		duration = 0.5

	default:
		startAngle = facing
		endAngle = facing
		duration = 0.3
	}

	return startAngle, endAngle, duration
}
