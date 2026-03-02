// Package attacktrail provides visual attack trail rendering for weapons.
package attacktrail

import (
	"image/color"
)

// TrailComponent stores visual attack trail state for an entity.
type TrailComponent struct {
	// Active trails currently rendering
	Trails []*Trail

	// Maximum simultaneous trails
	MaxTrails int
}

// Type returns the component type identifier.
func (c *TrailComponent) Type() string {
	return "AttackTrail"
}

// Trail represents a single weapon attack visual trail.
type Trail struct {
	// Trail type determines rendering style
	Type TrailType

	// Spatial properties
	StartX, StartY float64 // Origin point
	EndX, EndY     float64 // End point (for slashes)
	Angle          float64 // Trail rotation
	Arc            float64 // Arc angle for sweeps (radians)
	Range          float64 // Length/radius

	// Visual properties
	Width     float64    // Trail thickness
	Color     color.RGBA // Base color
	Intensity float64    // Opacity multiplier (0-1)

	// Animation state
	Age       float64 // Time since creation (seconds)
	MaxAge    float64 // Duration before fade-out (seconds)
	FadeStart float64 // When fade begins (seconds)

	// Motion blur effect
	Segments    []TrailSegment // Historical positions for motion blur
	MaxSegments int            // Limit for performance
}

// TrailSegment stores a historical trail position for motion blur.
type TrailSegment struct {
	X, Y      float64
	Angle     float64
	Intensity float64 // Fades with age
}

// TrailType defines the visual style of an attack trail.
type TrailType int

const (
	// TrailSlash is a curved arc trail (sword swing, axe chop)
	TrailSlash TrailType = iota

	// TrailThrust is a linear piercing trail (spear, rapier)
	TrailThrust

	// TrailCleave is a wide sweeping arc (greatsword, scythe)
	TrailCleave

	// TrailSmash is a radial impact burst (hammer, mace)
	TrailSmash

	// TrailSpin is a full-circle rotation (staff, dual blades)
	TrailSpin

	// TrailProjectile is a streak following a projectile
	TrailProjectile
)

// NewTrailComponent creates a trail component.
func NewTrailComponent(maxTrails int) *TrailComponent {
	return &TrailComponent{
		Trails:    make([]*Trail, 0, maxTrails),
		MaxTrails: maxTrails,
	}
}

// AddTrail registers a new attack trail.
func (c *TrailComponent) AddTrail(trail *Trail) {
	// Remove oldest if at capacity
	if len(c.Trails) >= c.MaxTrails {
		c.Trails = c.Trails[1:]
	}
	c.Trails = append(c.Trails, trail)
}

// Update advances trail animations and removes expired trails.
func (c *TrailComponent) Update(deltaTime float64) {
	i := 0
	for _, trail := range c.Trails {
		trail.Age += deltaTime

		// Calculate fade
		if trail.Age >= trail.FadeStart {
			fadeProgress := (trail.Age - trail.FadeStart) / (trail.MaxAge - trail.FadeStart)
			trail.Intensity = 1.0 - fadeProgress
		}

		// Keep if not expired
		if trail.Age < trail.MaxAge {
			c.Trails[i] = trail
			i++
		}
	}
	c.Trails = c.Trails[:i]
}
