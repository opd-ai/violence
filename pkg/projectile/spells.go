package projectile

import (
	"image/color"
	"math/rand"
)

// SpellTemplate defines a reusable spell configuration.
type SpellTemplate struct {
	Name            string
	DamageType      DamageType
	BaseDamage      float64
	Speed           float64
	Shape           ProjectileShape
	Radius          float64
	BeamWidth       float64
	PierceCount     int
	ExplodeOnDeath  bool
	ExplosionRadius float64
	TrailParticles  bool
	Lifetime        float64
}

// GenreSpellTemplates provides genre-specific spell varieties.
var GenreSpellTemplates = map[string][]SpellTemplate{
	"fantasy": {
		{Name: "Fireball", DamageType: DamageFire, BaseDamage: 35, Speed: 8.0, Shape: ShapeCircle, Radius: 0.3, ExplodeOnDeath: true, ExplosionRadius: 1.5, TrailParticles: true, Lifetime: 3.0},
		{Name: "Ice Shard", DamageType: DamageIce, BaseDamage: 25, Speed: 12.0, Shape: ShapeCircle, Radius: 0.2, PierceCount: 2, TrailParticles: true, Lifetime: 2.5},
		{Name: "Lightning Bolt", DamageType: DamageLightning, BaseDamage: 40, Speed: 20.0, Shape: ShapeBeam, BeamWidth: 0.15, PierceCount: -1, TrailParticles: true, Lifetime: 0.5},
		{Name: "Poison Cloud", DamageType: DamagePoison, BaseDamage: 15, Speed: 4.0, Shape: ShapeCircle, Radius: 0.4, ExplodeOnDeath: true, ExplosionRadius: 2.0, TrailParticles: true, Lifetime: 4.0},
		{Name: "Holy Smite", DamageType: DamageHoly, BaseDamage: 50, Speed: 15.0, Shape: ShapeCircle, Radius: 0.25, ExplodeOnDeath: true, ExplosionRadius: 1.0, TrailParticles: true, Lifetime: 2.0},
		{Name: "Shadow Bolt", DamageType: DamageShadow, BaseDamage: 30, Speed: 10.0, Shape: ShapeCircle, Radius: 0.2, PierceCount: 1, TrailParticles: true, Lifetime: 3.0},
		{Name: "Arcane Missile", DamageType: DamageArcane, BaseDamage: 20, Speed: 14.0, Shape: ShapeCircle, Radius: 0.15, PierceCount: 0, TrailParticles: true, Lifetime: 2.5},
	},
	"scifi": {
		{Name: "Plasma Bolt", DamageType: DamageFire, BaseDamage: 30, Speed: 15.0, Shape: ShapeCircle, Radius: 0.2, ExplodeOnDeath: true, ExplosionRadius: 1.0, TrailParticles: true, Lifetime: 3.0},
		{Name: "Cryo Beam", DamageType: DamageIce, BaseDamage: 25, Speed: 18.0, Shape: ShapeBeam, BeamWidth: 0.2, PierceCount: 3, TrailParticles: true, Lifetime: 1.5},
		{Name: "Tesla Arc", DamageType: DamageLightning, BaseDamage: 35, Speed: 22.0, Shape: ShapeBeam, BeamWidth: 0.1, PierceCount: -1, TrailParticles: true, Lifetime: 0.3},
		{Name: "Toxin Grenade", DamageType: DamagePoison, BaseDamage: 20, Speed: 6.0, Shape: ShapeCircle, Radius: 0.3, ExplodeOnDeath: true, ExplosionRadius: 2.5, TrailParticles: false, Lifetime: 5.0},
		{Name: "Laser Pulse", DamageType: DamagePhysical, BaseDamage: 28, Speed: 25.0, Shape: ShapeCircle, Radius: 0.15, PierceCount: 5, TrailParticles: true, Lifetime: 2.0},
	},
	"cyberpunk": {
		{Name: "Nanite Swarm", DamageType: DamagePoison, BaseDamage: 18, Speed: 7.0, Shape: ShapeCircle, Radius: 0.35, ExplodeOnDeath: true, ExplosionRadius: 1.8, TrailParticles: true, Lifetime: 4.0},
		{Name: "EMP Pulse", DamageType: DamageLightning, BaseDamage: 32, Speed: 12.0, Shape: ShapeCircle, Radius: 0.4, ExplodeOnDeath: true, ExplosionRadius: 2.2, TrailParticles: true, Lifetime: 3.0},
		{Name: "Railgun Shot", DamageType: DamagePhysical, BaseDamage: 60, Speed: 30.0, Shape: ShapeBeam, BeamWidth: 0.1, PierceCount: -1, TrailParticles: true, Lifetime: 0.8},
		{Name: "Incendiary Round", DamageType: DamageFire, BaseDamage: 28, Speed: 16.0, Shape: ShapeCircle, Radius: 0.2, ExplodeOnDeath: true, ExplosionRadius: 1.2, TrailParticles: true, Lifetime: 2.5},
	},
	"horror": {
		{Name: "Cursed Bolt", DamageType: DamageShadow, BaseDamage: 35, Speed: 6.0, Shape: ShapeCircle, Radius: 0.25, PierceCount: 1, TrailParticles: true, Lifetime: 4.0},
		{Name: "Blood Orb", DamageType: DamagePoison, BaseDamage: 22, Speed: 8.0, Shape: ShapeCircle, Radius: 0.3, ExplodeOnDeath: true, ExplosionRadius: 1.5, TrailParticles: true, Lifetime: 3.5},
		{Name: "Eldritch Beam", DamageType: DamageArcane, BaseDamage: 40, Speed: 10.0, Shape: ShapeBeam, BeamWidth: 0.25, PierceCount: -1, TrailParticles: true, Lifetime: 1.0},
		{Name: "Necrotic Bolt", DamageType: DamageShadow, BaseDamage: 30, Speed: 9.0, Shape: ShapeCircle, Radius: 0.2, PierceCount: 2, TrailParticles: true, Lifetime: 3.0},
	},
}

// CreateSpellProjectile creates a projectile from a spell template.
func CreateSpellProjectile(template SpellTemplate, dirX, dirY float64, ownerID int, rng *rand.Rand) *ProjectileComponent {
	proj := NewProjectileComponent(
		dirX*template.Speed,
		dirY*template.Speed,
		template.BaseDamage,
		template.DamageType,
		ownerID,
	)

	proj.Shape = template.Shape
	proj.Radius = template.Radius
	proj.BeamWidth = template.BeamWidth
	proj.PierceCount = template.PierceCount
	proj.ExplodeOnDeath = template.ExplodeOnDeath
	proj.ExplosionRadius = template.ExplosionRadius
	proj.TrailParticles = template.TrailParticles
	proj.Lifetime = template.Lifetime
	proj.MaxLifetime = template.Lifetime

	// Add slight random variance to damage (±10%)
	if rng != nil {
		variance := 0.9 + rng.Float64()*0.2
		proj.Damage *= variance
	}

	return proj
}

// GetRandomSpellForGenre returns a random spell template for the given genre.
func GetRandomSpellForGenre(genre string, rng *rand.Rand) SpellTemplate {
	templates, exists := GenreSpellTemplates[genre]
	if !exists || len(templates) == 0 {
		// Fallback to fantasy
		templates = GenreSpellTemplates["fantasy"]
	}

	if rng == nil {
		return templates[0]
	}

	return templates[rng.Intn(len(templates))]
}

// CreateResistanceProfile creates a genre-appropriate resistance profile for an entity.
func CreateResistanceProfile(genre, entityType string, rng *rand.Rand) *ResistanceComponent {
	rc := NewResistanceComponent()

	switch genre {
	case "fantasy":
		switch entityType {
		case "fire_elemental":
			rc.Resistances[DamageFire] = 0.75    // 75% fire resistance
			rc.Resistances[DamageIce] = -0.5     // 50% ice weakness
			rc.Resistances[DamagePhysical] = 0.3 // 30% physical resistance
		case "ice_elemental":
			rc.Resistances[DamageIce] = 0.75
			rc.Resistances[DamageFire] = -0.5
			rc.Resistances[DamagePhysical] = 0.3
		case "undead":
			rc.Resistances[DamagePoison] = 1.0 // Immune to poison
			rc.Resistances[DamageHoly] = -0.75 // 75% holy weakness
			rc.Resistances[DamagePhysical] = 0.2
		case "demon":
			rc.Resistances[DamageFire] = 0.5
			rc.Resistances[DamageShadow] = 0.5
			rc.Resistances[DamageHoly] = -1.0 // Double holy damage
		case "construct":
			rc.Resistances[DamagePoison] = 1.0     // Immune to poison
			rc.Resistances[DamageLightning] = -0.3 // Lightning weakness
			rc.Resistances[DamagePhysical] = 0.4
		}

	case "scifi":
		switch entityType {
		case "robot":
			rc.Resistances[DamageLightning] = -0.5 // EMP weakness
			rc.Resistances[DamagePoison] = 1.0     // Immune to toxins
			rc.Resistances[DamagePhysical] = 0.3
		case "bio_mutant":
			rc.Resistances[DamagePoison] = 0.6
			rc.Resistances[DamageFire] = -0.3
			rc.Resistances[DamagePhysical] = 0.1
		case "shield_drone":
			rc.Resistances[DamagePhysical] = 0.5
			rc.Resistances[DamageFire] = 0.5
			rc.Resistances[DamageLightning] = 0.2
		}

	case "horror":
		switch entityType {
		case "ghost":
			rc.Resistances[DamagePhysical] = 0.8 // Hard to hit with physical
			rc.Resistances[DamageArcane] = -0.5
			rc.Resistances[DamageShadow] = 0.7
		case "cultist":
			rc.Resistances[DamageShadow] = 0.4
			rc.Resistances[DamageHoly] = -0.6
		case "flesh_horror":
			rc.Resistances[DamagePoison] = 0.5
			rc.Resistances[DamageFire] = -0.4
		}
	}

	return rc
}

// GetDamageTypeColor returns the visual color for a damage type.
func GetDamageTypeColor(dt DamageType) color.RGBA {
	return getDamageTypeColor(dt)
}
