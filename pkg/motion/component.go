package motion

// Component holds organic motion state for visually realistic animation.
type Component struct {
	// Easing state for acceleration/deceleration
	EasedVelocityX  float64
	EasedVelocityY  float64
	TargetVelocityX float64
	TargetVelocityY float64
	EaseRate        float64 // Higher = faster acceleration (1.0-10.0)

	// Squash and stretch for impact/landing
	SquashX    float64 // Scale multiplier on X axis (1.0 = normal)
	SquashY    float64 // Scale multiplier on Y axis (1.0 = normal)
	ImpactTime float64 // Time since last impact

	// Secondary motion for trailing elements
	TrailOffsetX   []float64 // Historical X positions for tail/cloth
	TrailOffsetY   []float64 // Historical Y positions
	TrailLength    int       // Number of segments
	TrailStiffness float64   // 0.0-1.0: how much trail follows (0=loose, 1=rigid)

	// Breathing idle animation
	BreathPhase     float64 // 0.0-2π sine wave phase
	BreathAmplitude float64 // Pixel offset amplitude
	BreathFrequency float64 // Breaths per second

	// Weight-based movement properties
	Mass          float64 // Entity mass (affects acceleration)
	Grounded      bool    // On ground vs airborne
	LastPositionX float64
	LastPositionY float64
}

// Type implements Component interface.
func (c *Component) Type() string {
	return "motion"
}
