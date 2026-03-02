package projectile

import (
	"image/color"
)

// ProjectileShape defines the collision geometry for different projectile types.
type ProjectileShape int

const (
	ShapeCircle ProjectileShape = iota // Standard circular projectile
	ShapeBeam                          // Line segment with width
	ShapeAOE                           // Area of effect explosion
)

// ProjectileComponent represents a projectile entity.
// Pure data component - no methods with logic.
type ProjectileComponent struct {
	VelX, VelY      float64    // Velocity vector
	Damage          float64    // Base damage amount
	DamageType      DamageType // Type of damage dealt
	Lifetime        float64    // Remaining lifetime in seconds
	MaxLifetime     float64    // Maximum lifetime for trail effects
	OwnerID         int        // Entity ID of the projectile's owner (for friendly fire prevention)
	Shape           ProjectileShape
	Radius          float64      // For circle and AoE
	BeamWidth       float64      // For beam projectiles
	PierceCount     int          // How many entities it can pierce through (-1 = infinite)
	Color           color.RGBA   // Visual color
	TrailParticles  bool         // Whether to spawn trail particles
	ExplodeOnDeath  bool         // Whether to create AoE explosion on impact
	ExplosionRadius float64      // Radius of explosion if ExplodeOnDeath is true
	HitEntities     map[int]bool // Track hit entities for pierce mechanics
}

// Type returns the component type identifier.
func (p *ProjectileComponent) Type() string {
	return "ProjectileComponent"
}

// NewProjectileComponent creates a standard projectile.
func NewProjectileComponent(velX, velY, damage float64, damageType DamageType, ownerID int) *ProjectileComponent {
	return &ProjectileComponent{
		VelX:            velX,
		VelY:            velY,
		Damage:          damage,
		DamageType:      damageType,
		Lifetime:        5.0, // Default 5 second lifetime
		MaxLifetime:     5.0,
		OwnerID:         ownerID,
		Shape:           ShapeCircle,
		Radius:          0.2,
		BeamWidth:       0.1,
		PierceCount:     0,
		Color:           getDamageTypeColor(damageType),
		TrailParticles:  true,
		ExplodeOnDeath:  false,
		ExplosionRadius: 0.0,
		HitEntities:     make(map[int]bool),
	}
}

// getDamageTypeColor returns a color for each damage type.
func getDamageTypeColor(dt DamageType) color.RGBA {
	switch dt {
	case DamagePhysical:
		return color.RGBA{R: 192, G: 192, B: 192, A: 255}
	case DamageFire:
		return color.RGBA{R: 255, G: 80, B: 20, A: 255}
	case DamageIce:
		return color.RGBA{R: 100, G: 200, B: 255, A: 255}
	case DamageLightning:
		return color.RGBA{R: 255, G: 255, B: 100, A: 255}
	case DamagePoison:
		return color.RGBA{R: 100, G: 255, B: 100, A: 255}
	case DamageHoly:
		return color.RGBA{R: 255, G: 255, B: 200, A: 255}
	case DamageShadow:
		return color.RGBA{R: 80, G: 60, B: 120, A: 255}
	case DamageArcane:
		return color.RGBA{R: 200, G: 100, B: 255, A: 255}
	default:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
}
