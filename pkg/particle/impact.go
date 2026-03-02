// Package particle implements genre-configurable impact effect emitters.
package particle

import (
	"image/color"
	"math"
)

// ImpactType categorizes the type of impact for visual differentiation.
type ImpactType int

const (
	ImpactMelee ImpactType = iota
	ImpactProjectile
	ImpactExplosion
	ImpactMagic
	ImpactCritical
	ImpactBlock
	ImpactDeath
)

// MaterialType defines the surface material being hit.
type MaterialType int

const (
	MaterialFlesh MaterialType = iota
	MaterialMetal
	MaterialStone
	MaterialWood
	MaterialEnergy
	MaterialEthereal
)

// ImpactEffectEmitter spawns genre-appropriate particles for combat impacts.
type ImpactEffectEmitter struct {
	system  *ParticleSystem
	genreID string
}

// NewImpactEffectEmitter creates an impact effect emitter.
func NewImpactEffectEmitter(system *ParticleSystem, genreID string) *ImpactEffectEmitter {
	return &ImpactEffectEmitter{
		system:  system,
		genreID: genreID,
	}
}

// EmitImpact spawns an impact effect burst at the given location.
func (e *ImpactEffectEmitter) EmitImpact(x, y float64, impactType ImpactType, material MaterialType, angle float64) {
	config := e.getImpactConfig(impactType, material)

	// Main impact particles
	e.spawnMainBurst(x, y, angle, config)

	// Secondary effects based on impact type
	switch impactType {
	case ImpactCritical:
		e.spawnCriticalFlare(x, y, config)
	case ImpactExplosion:
		e.spawnExplosionRing(x, y, config)
	case ImpactMagic:
		e.spawnMagicSparkles(x, y, config)
	case ImpactDeath:
		e.spawnDeathBurst(x, y, material, config)
	case ImpactBlock:
		e.spawnBlockSparks(x, y, angle, config)
	}

	// Material-specific debris
	e.spawnDebris(x, y, material, config)
}

// ImpactEffectConfig defines visual parameters for an impact effect.
type ImpactEffectConfig struct {
	PrimaryColor   color.RGBA
	SecondaryColor color.RGBA
	ParticleCount  int
	Speed          float64
	Spread         float64
	Life           float64
	Size           float64
	DebrisCount    int
	DebrisSize     float64
}

// getImpactConfig returns genre-and-type-specific configuration.
func (e *ImpactEffectEmitter) getImpactConfig(impactType ImpactType, material MaterialType) ImpactEffectConfig {
	base := e.getBaseConfig(material)

	// Modify based on impact type
	switch impactType {
	case ImpactCritical:
		base.ParticleCount = int(float64(base.ParticleCount) * 2.5)
		base.Speed *= 1.8
		base.Life *= 1.5
		base.Size *= 1.3
		// Add brightness to critical hits
		base.PrimaryColor = brightenColor(base.PrimaryColor, 1.4)
	case ImpactExplosion:
		base.ParticleCount = int(float64(base.ParticleCount) * 3.0)
		base.Speed *= 2.5
		base.Spread = math.Pi * 2 // Full circle
		base.Life *= 1.2
		base.Size *= 1.5
	case ImpactMagic:
		base.ParticleCount = int(float64(base.ParticleCount) * 1.8)
		base.Speed *= 1.2
		base.Life *= 2.0
		base.Size *= 0.8
	case ImpactBlock:
		base.ParticleCount = int(float64(base.ParticleCount) * 0.5)
		base.Speed *= 0.6
		base.Life *= 0.7
	case ImpactDeath:
		base.ParticleCount = int(float64(base.ParticleCount) * 4.0)
		base.Speed *= 1.5
		base.Spread = math.Pi * 2
		base.Life *= 2.5
		base.Size *= 1.2
	}

	return base
}

// getBaseConfig returns material-specific base configuration for the current genre.
func (e *ImpactEffectEmitter) getBaseConfig(material MaterialType) ImpactEffectConfig {
	switch e.genreID {
	case "fantasy":
		return e.getFantasyConfig(material)
	case "scifi":
		return e.getSciFiConfig(material)
	case "horror":
		return e.getHorrorConfig(material)
	case "cyberpunk":
		return e.getCyberpunkConfig(material)
	default:
		return e.getFantasyConfig(material)
	}
}

// getFantasyConfig returns fantasy-themed impact configurations.
func (e *ImpactEffectEmitter) getFantasyConfig(material MaterialType) ImpactEffectConfig {
	switch material {
	case MaterialFlesh:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 180, G: 20, B: 20, A: 255}, // Blood red
			SecondaryColor: color.RGBA{R: 100, G: 10, B: 10, A: 200}, // Dark red
			ParticleCount:  18,
			Speed:          12.0,
			Spread:         math.Pi * 0.6,
			Life:           0.4,
			Size:           1.5,
			DebrisCount:    3,
			DebrisSize:     0.8,
		}
	case MaterialMetal:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 255, G: 220, B: 100, A: 255}, // Spark yellow
			SecondaryColor: color.RGBA{R: 200, G: 150, B: 80, A: 220},  // Orange
			ParticleCount:  25,
			Speed:          18.0,
			Spread:         math.Pi * 0.8,
			Life:           0.25,
			Size:           1.2,
			DebrisCount:    8,
			DebrisSize:     0.6,
		}
	case MaterialStone:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 140, G: 130, B: 120, A: 255}, // Stone gray
			SecondaryColor: color.RGBA{R: 80, G: 75, B: 70, A: 200},    // Dark gray
			ParticleCount:  20,
			Speed:          10.0,
			Spread:         math.Pi * 0.5,
			Life:           0.5,
			Size:           1.3,
			DebrisCount:    6,
			DebrisSize:     1.0,
		}
	case MaterialWood:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 160, G: 120, B: 80, A: 255}, // Wood brown
			SecondaryColor: color.RGBA{R: 100, G: 70, B: 40, A: 200},  // Dark brown
			ParticleCount:  15,
			Speed:          8.0,
			Spread:         math.Pi * 0.4,
			Life:           0.6,
			Size:           1.1,
			DebrisCount:    5,
			DebrisSize:     0.9,
		}
	case MaterialEnergy:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 120, G: 180, B: 255, A: 255}, // Arcane blue
			SecondaryColor: color.RGBA{R: 180, G: 120, B: 255, A: 220}, // Purple
			ParticleCount:  30,
			Speed:          15.0,
			Spread:         math.Pi,
			Life:           0.35,
			Size:           1.0,
			DebrisCount:    0,
			DebrisSize:     0.0,
		}
	case MaterialEthereal:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 200, G: 200, B: 255, A: 180}, // Ghostly white
			SecondaryColor: color.RGBA{R: 150, G: 150, B: 200, A: 140}, // Faint blue
			ParticleCount:  12,
			Speed:          6.0,
			Spread:         math.Pi * 1.2,
			Life:           0.8,
			Size:           1.4,
			DebrisCount:    0,
			DebrisSize:     0.0,
		}
	default:
		return e.getFantasyConfig(MaterialFlesh)
	}
}

// getSciFiConfig returns sci-fi themed impact configurations.
func (e *ImpactEffectEmitter) getSciFiConfig(material MaterialType) ImpactEffectConfig {
	switch material {
	case MaterialFlesh:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 220, G: 40, B: 60, A: 255},
			SecondaryColor: color.RGBA{R: 140, G: 20, B: 30, A: 200},
			ParticleCount:  22,
			Speed:          14.0,
			Spread:         math.Pi * 0.5,
			Life:           0.3,
			Size:           1.3,
			DebrisCount:    4,
			DebrisSize:     0.7,
		}
	case MaterialMetal:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 100, G: 200, B: 255, A: 255}, // Energy blue
			SecondaryColor: color.RGBA{R: 255, G: 255, B: 255, A: 240}, // White flash
			ParticleCount:  30,
			Speed:          22.0,
			Spread:         math.Pi * 0.7,
			Life:           0.2,
			Size:           1.0,
			DebrisCount:    10,
			DebrisSize:     0.5,
		}
	case MaterialEnergy:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 0, G: 255, B: 200, A: 255},   // Plasma cyan
			SecondaryColor: color.RGBA{R: 100, G: 255, B: 255, A: 220}, // Bright cyan
			ParticleCount:  35,
			Speed:          18.0,
			Spread:         math.Pi * 1.5,
			Life:           0.25,
			Size:           0.9,
			DebrisCount:    0,
			DebrisSize:     0.0,
		}
	default:
		return e.getSciFiConfig(MaterialMetal)
	}
}

// getHorrorConfig returns horror-themed impact configurations.
func (e *ImpactEffectEmitter) getHorrorConfig(material MaterialType) ImpactEffectConfig {
	switch material {
	case MaterialFlesh:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 120, G: 20, B: 20, A: 200}, // Dark blood
			SecondaryColor: color.RGBA{R: 60, G: 10, B: 10, A: 160},  // Very dark
			ParticleCount:  15,
			Speed:          9.0,
			Spread:         math.Pi * 0.4,
			Life:           0.7,
			Size:           1.6,
			DebrisCount:    4,
			DebrisSize:     1.0,
		}
	case MaterialEthereal:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 100, G: 120, B: 100, A: 140}, // Sickly green
			SecondaryColor: color.RGBA{R: 60, G: 80, B: 60, A: 100},    // Dark green
			ParticleCount:  10,
			Speed:          5.0,
			Spread:         math.Pi,
			Life:           1.2,
			Size:           1.5,
			DebrisCount:    0,
			DebrisSize:     0.0,
		}
	default:
		return e.getHorrorConfig(MaterialFlesh)
	}
}

// getCyberpunkConfig returns cyberpunk-themed impact configurations.
func (e *ImpactEffectEmitter) getCyberpunkConfig(material MaterialType) ImpactEffectConfig {
	switch material {
	case MaterialFlesh:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 255, G: 0, B: 100, A: 255}, // Hot pink
			SecondaryColor: color.RGBA{R: 200, G: 0, B: 255, A: 220}, // Purple
			ParticleCount:  20,
			Speed:          13.0,
			Spread:         math.Pi * 0.5,
			Life:           0.35,
			Size:           1.2,
			DebrisCount:    3,
			DebrisSize:     0.7,
		}
	case MaterialMetal:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 0, G: 255, B: 255, A: 255}, // Neon cyan
			SecondaryColor: color.RGBA{R: 255, G: 255, B: 0, A: 240}, // Yellow
			ParticleCount:  28,
			Speed:          20.0,
			Spread:         math.Pi * 0.6,
			Life:           0.22,
			Size:           1.0,
			DebrisCount:    9,
			DebrisSize:     0.6,
		}
	case MaterialEnergy:
		return ImpactEffectConfig{
			PrimaryColor:   color.RGBA{R: 255, G: 0, B: 255, A: 255}, // Magenta
			SecondaryColor: color.RGBA{R: 0, G: 255, B: 255, A: 230}, // Cyan
			ParticleCount:  32,
			Speed:          17.0,
			Spread:         math.Pi * 1.2,
			Life:           0.28,
			Size:           0.9,
			DebrisCount:    0,
			DebrisSize:     0.0,
		}
	default:
		return e.getCyberpunkConfig(MaterialMetal)
	}
}

// spawnMainBurst creates the primary impact particle burst.
func (e *ImpactEffectEmitter) spawnMainBurst(x, y, angle float64, config ImpactEffectConfig) {
	// Impact direction opposite to incoming angle
	impactDir := angle + math.Pi

	for i := 0; i < config.ParticleCount; i++ {
		// Spread particles in a cone around the impact direction
		spreadAngle := (e.system.rng.Float64()*2 - 1) * config.Spread
		particleAngle := impactDir + spreadAngle

		// Randomize speed
		speed := config.Speed * (0.6 + e.system.rng.Float64()*0.8)

		vx := math.Cos(particleAngle) * speed
		vy := math.Sin(particleAngle) * speed
		vz := (e.system.rng.Float64()*2 - 1) * 3.0

		// Randomize color between primary and secondary
		c := config.PrimaryColor
		if e.system.rng.Float64() < 0.3 {
			c = config.SecondaryColor
		}

		// Randomize size slightly
		size := config.Size * (0.8 + e.system.rng.Float64()*0.4)

		e.system.Spawn(x, y, 0, vx, vy, vz, config.Life, size, c)
	}
}

// spawnCriticalFlare creates a bright flash for critical hits.
func (e *ImpactEffectEmitter) spawnCriticalFlare(x, y float64, config ImpactEffectConfig) {
	flareColor := brightenColor(config.PrimaryColor, 2.0)

	// Central bright flash
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi / 4
		speed := 5.0 + e.system.rng.Float64()*3.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed

		e.system.Spawn(x, y, 0, vx, vy, 0, 0.15, config.Size*1.5, flareColor)
	}

	// Outer ring expansion
	for i := 0; i < 12; i++ {
		angle := float64(i) * math.Pi / 6
		speed := 15.0 + e.system.rng.Float64()*5.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed

		c := config.PrimaryColor
		c.A = 180

		e.system.Spawn(x, y, 0, vx, vy, 0, 0.25, config.Size*0.8, c)
	}
}

// spawnExplosionRing creates an expanding ring effect for explosions.
func (e *ImpactEffectEmitter) spawnExplosionRing(x, y float64, config ImpactEffectConfig) {
	ringParticles := 24
	for i := 0; i < ringParticles; i++ {
		angle := float64(i) * 2 * math.Pi / float64(ringParticles)
		speed := 20.0 + e.system.rng.Float64()*8.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed

		// Alternating colors for visual interest
		c := config.PrimaryColor
		if i%2 == 0 {
			c = config.SecondaryColor
		}

		e.system.Spawn(x, y, 0, vx, vy, 0, 0.4, config.Size*1.2, c)
	}

	// Center flash
	flashColor := brightenColor(config.PrimaryColor, 1.8)
	e.system.SpawnBurst(x, y, 0, 15, 8.0, 2.0, 0.2, config.Size*2.0, flashColor)
}

// spawnMagicSparkles creates sparkly magic effect particles.
func (e *ImpactEffectEmitter) spawnMagicSparkles(x, y float64, config ImpactEffectConfig) {
	sparkleCount := 15
	for i := 0; i < sparkleCount; i++ {
		angle := e.system.rng.Float64() * 2 * math.Pi
		speed := 3.0 + e.system.rng.Float64()*8.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		vz := e.system.rng.Float64() * 5.0

		// Twinkling effect with varied colors
		c := config.PrimaryColor
		if e.system.rng.Float64() < 0.4 {
			c = config.SecondaryColor
		}
		// Brighten sparkles
		c = brightenColor(c, 1.3)

		life := config.Life * (0.8 + e.system.rng.Float64()*0.6)
		size := config.Size * (0.5 + e.system.rng.Float64()*0.8)

		e.system.Spawn(x, y, 0, vx, vy, vz, life, size, c)
	}
}

// spawnDeathBurst creates a large particle burst for entity death.
func (e *ImpactEffectEmitter) spawnDeathBurst(x, y float64, material MaterialType, config ImpactEffectConfig) {
	// Large omnidirectional burst
	burstCount := 40
	for i := 0; i < burstCount; i++ {
		angle := e.system.rng.Float64() * 2 * math.Pi
		speed := 5.0 + e.system.rng.Float64()*15.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		vz := (e.system.rng.Float64()*2 - 1) * 8.0

		c := config.PrimaryColor
		if e.system.rng.Float64() < 0.4 {
			c = config.SecondaryColor
		}

		life := config.Life * (0.7 + e.system.rng.Float64()*0.6)
		size := config.Size * (0.8 + e.system.rng.Float64()*0.8)

		e.system.Spawn(x, y, 0, vx, vy, vz, life, size, c)
	}

	// Additional ethereal wisps for some materials
	if material == MaterialEthereal || material == MaterialEnergy {
		for i := 0; i < 8; i++ {
			angle := float64(i) * math.Pi / 4
			speed := 3.0
			vx := math.Cos(angle) * speed
			vy := math.Sin(angle) * speed

			wispColor := config.SecondaryColor
			wispColor.A = 120

			e.system.Spawn(x, y, 0, vx, vy, 0, config.Life*2.0, config.Size*2.5, wispColor)
		}
	}
}

// spawnBlockSparks creates deflection sparks for blocked attacks.
func (e *ImpactEffectEmitter) spawnBlockSparks(x, y, angle float64, config ImpactEffectConfig) {
	// Deflect particles at angles away from block direction
	blockDir := angle + math.Pi

	sparkCount := 10
	for i := 0; i < sparkCount; i++ {
		// Sparks deflect to sides
		deflectAngle := blockDir + (e.system.rng.Float64()*2-1)*math.Pi*0.7
		speed := config.Speed * (0.5 + e.system.rng.Float64()*0.8)

		vx := math.Cos(deflectAngle) * speed
		vy := math.Sin(deflectAngle) * speed
		vz := e.system.rng.Float64() * 4.0

		// Metallic or energy sparks
		c := color.RGBA{R: 255, G: 230, B: 150, A: 255}
		if e.genreID == "scifi" || e.genreID == "cyberpunk" {
			c = color.RGBA{R: 150, G: 200, B: 255, A: 255}
		}

		e.system.Spawn(x, y, 0, vx, vy, vz, config.Life*0.8, config.Size*0.9, c)
	}
}

// spawnDebris creates material-specific debris particles.
func (e *ImpactEffectEmitter) spawnDebris(x, y float64, material MaterialType, config ImpactEffectConfig) {
	if config.DebrisCount == 0 {
		return
	}

	for i := 0; i < config.DebrisCount; i++ {
		angle := e.system.rng.Float64() * 2 * math.Pi
		speed := 4.0 + e.system.rng.Float64()*6.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		vz := e.system.rng.Float64() * 6.0

		// Debris uses darker secondary color
		c := config.SecondaryColor
		c = darkenColor(c, 0.7)

		life := config.Life * 1.5
		size := config.DebrisSize * (0.7 + e.system.rng.Float64()*0.6)

		e.system.Spawn(x, y, 0, vx, vy, vz, life, size, c)
	}
}

// brightenColor increases color brightness by a factor.
func brightenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(c.R)*factor)),
		G: uint8(math.Min(255, float64(c.G)*factor)),
		B: uint8(math.Min(255, float64(c.B)*factor)),
		A: c.A,
	}
}

// darkenColor decreases color brightness by a factor.
func darkenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}
