package lighting

// LightComponent represents a dynamic light source attached to an entity.
// Pure data component following ECS architecture.
type LightComponent struct {
	PointLight               // Embedded point light
	AttachedToEntity bool    // Whether light follows entity position
	OffsetX, OffsetY float64 // Offset from entity position
	Pulsing          bool    // Whether light pulses with entity actions
	PulsePhase       float64 // Current pulse phase [0.0-1.0]
	PulseSpeed       float64 // Pulse speed multiplier
	FadeInDuration   float64 // Time to fade in (seconds)
	FadeOutDuration  float64 // Time to fade out (seconds)
	CurrentAge       float64 // Time since creation (seconds)
	Lifetime         float64 // Total lifetime (0 = infinite)
	Enabled          bool    // Whether light is currently active
}

// Type returns the component type identifier.
func (l *LightComponent) Type() string {
	return "Light"
}

// NewLightComponent creates a light component from a preset.
func NewLightComponent(preset LightPreset, seed int64) *LightComponent {
	return &LightComponent{
		PointLight:       NewPointLight(0, 0, preset, seed),
		AttachedToEntity: false,
		Enabled:          true,
		Lifetime:         0, // Infinite by default
		PulseSpeed:       1.0,
	}
}

// NewAttachedLight creates a light that follows an entity.
func NewAttachedLight(preset LightPreset, seed int64, offsetX, offsetY float64) *LightComponent {
	lc := NewLightComponent(preset, seed)
	lc.AttachedToEntity = true
	lc.OffsetX = offsetX
	lc.OffsetY = offsetY
	return lc
}

// NewTemporaryLight creates a light with a limited lifetime.
func NewTemporaryLight(preset LightPreset, seed int64, lifetime float64) *LightComponent {
	lc := NewLightComponent(preset, seed)
	lc.Lifetime = lifetime
	lc.FadeInDuration = lifetime * 0.1  // Quick fade in
	lc.FadeOutDuration = lifetime * 0.2 // Slower fade out
	return lc
}

// NewPulsingLight creates a light that pulses (e.g., for damage effects).
func NewPulsingLight(preset LightPreset, seed int64, pulseSpeed float64) *LightComponent {
	lc := NewLightComponent(preset, seed)
	lc.Pulsing = true
	lc.PulseSpeed = pulseSpeed
	return lc
}
