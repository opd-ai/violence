package impactburst

import "image/color"

// ImpactType categorizes the nature of the impact for visual differentiation.
type ImpactType int

const (
	// ImpactMelee is a melee weapon strike impact.
	ImpactMelee ImpactType = iota
	// ImpactProjectile is a ranged projectile impact.
	ImpactProjectile
	// ImpactExplosion is an area-of-effect explosion.
	ImpactExplosion
	// ImpactMagic is a magical spell impact.
	ImpactMagic
	// ImpactCritical is a critical hit with enhanced visuals.
	ImpactCritical
	// ImpactBlock is a blocked or deflected attack.
	ImpactBlock
	// ImpactDeath is an entity death burst.
	ImpactDeath
)

// MaterialType defines the surface material being impacted.
type MaterialType int

const (
	// MaterialFlesh is organic flesh material (blood, chunks).
	MaterialFlesh MaterialType = iota
	// MaterialMetal is metallic material (sparks, shards).
	MaterialMetal
	// MaterialStone is stone material (chips, dust).
	MaterialStone
	// MaterialWood is wooden material (splinters).
	MaterialWood
	// MaterialEnergy is energy-based material (plasma, electricity).
	MaterialEnergy
	// MaterialEthereal is ghostly/spectral material (wisps).
	MaterialEthereal
)

// Impact represents an active impact burst effect in the world.
type Impact struct {
	// World position
	X, Y float64

	// Impact angle (direction of incoming force)
	Angle float64

	// Classification
	Type     ImpactType
	Material MaterialType

	// Intensity multiplier (0.0-2.0, affects size and particle count)
	Intensity float64

	// Timing
	Age    float64 // Current age in seconds
	MaxAge float64 // Maximum lifetime

	// Visual state
	ShockwaveRadius float64 // Current radius of expanding shockwave
	ShockwaveAlpha  float64 // Shockwave opacity (fades with expansion)
	GlowIntensity   float64 // Glow bloom intensity
	FlashAlpha      float64 // Initial flash opacity

	// Debris particles (tracked separately from particle system for rendering)
	Debris []DebrisParticle
}

// DebrisParticle represents a single piece of debris from the impact.
type DebrisParticle struct {
	X, Y   float64 // Position offset from impact center
	VX, VY float64 // Velocity
	Size   float64 // Particle size
	Color  color.RGBA
	Age    float64
	MaxAge float64
	// Rotation for non-circular debris (chunks)
	Rotation     float64
	RotationVel  float64
	IsChunk      bool // True for larger debris pieces
	GravityScale float64
}

// Component stores impact burst state for an entity.
type Component struct {
	// ActiveImpacts is the list of currently-rendering impacts at this entity.
	ActiveImpacts []Impact
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "ImpactBurstComponent"
}

// ImpactProfile defines visual parameters for a specific impact type/material combination.
type ImpactProfile struct {
	// Colors
	PrimaryColor   color.RGBA // Main burst color
	SecondaryColor color.RGBA // Secondary/debris color
	GlowColor      color.RGBA // Glow bloom color
	ShockwaveColor color.RGBA // Shockwave ring color
	DebrisColor    color.RGBA // Debris particle color
	ChunkColor     color.RGBA // Large debris chunk color

	// Timing
	Duration          float64 // Total effect duration
	ShockwaveDuration float64 // How long shockwave expands
	FlashDuration     float64 // Initial flash duration
	GlowDuration      float64 // Glow fade duration

	// Sizing
	ShockwaveMaxRadius float64 // Maximum shockwave expansion
	ShockwaveWidth     float64 // Ring line width
	BaseParticleSize   float64 // Base particle size
	ChunkSize          float64 // Large debris chunk size

	// Counts
	ParticleCount  int // Number of small debris particles
	ChunkCount     int // Number of large debris chunks
	ShockwaveRings int // Number of concentric shockwave rings

	// Behavior
	ParticleSpeed        float64 // Base particle velocity
	ParticleSpread       float64 // Angular spread (radians)
	ParticleGravity      float64 // Gravity effect on particles
	ChunkGravity         float64 // Gravity effect on chunks
	HasGlow              bool    // Whether to render glow bloom
	HasShockwave         bool    // Whether to render shockwave rings
	HasDirectionalDebris bool    // Whether debris follows impact angle
}
