// Package particle implements genre-configurable particle emitter types.
package particle

import (
	"image/color"
	"math"
)

// EmitterConfig defines parameters for a specific emitter type.
type EmitterConfig struct {
	Color          color.RGBA
	ParticleCount  int
	Speed          float64
	Spread         float64
	Life           float64
	Size           float64
	VelocityScaleX float64
	VelocityScaleY float64
}

// MuzzleFlashEmitter creates weapon muzzle flash effects with genre-specific visuals.
type MuzzleFlashEmitter struct {
	system  *ParticleSystem
	genreID string
}

// NewMuzzleFlashEmitter creates a muzzle flash emitter.
func NewMuzzleFlashEmitter(system *ParticleSystem, genreID string) *MuzzleFlashEmitter {
	return &MuzzleFlashEmitter{
		system:  system,
		genreID: genreID,
	}
}

// Emit spawns a muzzle flash at the given position and angle.
func (e *MuzzleFlashEmitter) Emit(x, y, angle float64) {
	config := e.getConfig()

	// Convert angle to direction vector
	dirX := math.Cos(angle)
	dirY := math.Sin(angle)

	for i := 0; i < config.ParticleCount; i++ {
		// Randomize velocity with directional bias
		randAngle := angle + (e.system.rng.Float64()*2-1)*config.Spread
		velocity := config.Speed * (0.5 + e.system.rng.Float64()*0.5)

		vx := math.Cos(randAngle) * velocity * config.VelocityScaleX
		vy := math.Sin(randAngle) * velocity * config.VelocityScaleY
		vz := (e.system.rng.Float64()*2 - 1) * config.Spread

		offsetX := x + dirX*0.5
		offsetY := y + dirY*0.5

		e.system.Spawn(offsetX, offsetY, 0, vx, vy, vz, config.Life, config.Size, config.Color)
	}
}

// getConfig returns genre-specific muzzle flash configuration.
func (e *MuzzleFlashEmitter) getConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		// Smoky, powder-burn flash
		return EmitterConfig{
			Color:          color.RGBA{R: 220, G: 180, B: 100, A: 255},
			ParticleCount:  12,
			Speed:          15.0,
			Spread:         0.6,
			Life:           0.15,
			Size:           1.2,
			VelocityScaleX: 1.0,
			VelocityScaleY: 1.0,
		}
	case "scifi":
		// Bright energy discharge
		return EmitterConfig{
			Color:          color.RGBA{R: 100, G: 200, B: 255, A: 255},
			ParticleCount:  8,
			Speed:          25.0,
			Spread:         0.3,
			Life:           0.1,
			Size:           0.8,
			VelocityScaleX: 1.5,
			VelocityScaleY: 1.5,
		}
	case "horror":
		// Dim, weak flash
		return EmitterConfig{
			Color:          color.RGBA{R: 180, G: 160, B: 120, A: 200},
			ParticleCount:  6,
			Speed:          12.0,
			Spread:         0.5,
			Life:           0.12,
			Size:           1.0,
			VelocityScaleX: 0.8,
			VelocityScaleY: 0.8,
		}
	case "cyberpunk":
		// Neon-tinged flash
		return EmitterConfig{
			Color:          color.RGBA{R: 255, G: 50, B: 200, A: 255},
			ParticleCount:  10,
			Speed:          20.0,
			Spread:         0.4,
			Life:           0.13,
			Size:           1.1,
			VelocityScaleX: 1.2,
			VelocityScaleY: 1.2,
		}
	case "postapoc":
		// Dirty, uneven flash
		return EmitterConfig{
			Color:          color.RGBA{R: 200, G: 150, B: 80, A: 220},
			ParticleCount:  10,
			Speed:          14.0,
			Spread:         0.7,
			Life:           0.16,
			Size:           1.3,
			VelocityScaleX: 0.9,
			VelocityScaleY: 0.9,
		}
	default:
		return EmitterConfig{
			Color:          color.RGBA{R: 255, G: 220, B: 100, A: 255},
			ParticleCount:  8,
			Speed:          20.0,
			Spread:         0.5,
			Life:           0.1,
			Size:           1.0,
			VelocityScaleX: 1.0,
			VelocityScaleY: 1.0,
		}
	}
}

// SparkEmitter creates spark effects for bullet impacts on metal surfaces.
type SparkEmitter struct {
	system  *ParticleSystem
	genreID string
}

// NewSparkEmitter creates a spark emitter.
func NewSparkEmitter(system *ParticleSystem, genreID string) *SparkEmitter {
	return &SparkEmitter{
		system:  system,
		genreID: genreID,
	}
}

// Emit spawns sparks at the given position with normal vector.
func (e *SparkEmitter) Emit(x, y, normalX, normalY float64) {
	config := e.getConfig()

	// Sparks reflect off surface normal
	baseAngle := math.Atan2(normalY, normalX)

	for i := 0; i < config.ParticleCount; i++ {
		angle := baseAngle + (e.system.rng.Float64()*2-1)*config.Spread
		velocity := config.Speed * (0.5 + e.system.rng.Float64()*0.5)

		vx := math.Cos(angle) * velocity
		vy := math.Sin(angle) * velocity
		vz := e.system.rng.Float64() * config.Spread * 5

		e.system.Spawn(x, y, 0, vx, vy, vz, config.Life, config.Size, config.Color)
	}
}

// getConfig returns genre-specific spark configuration.
func (e *SparkEmitter) getConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		// Metal-on-metal sparks (swords)
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 180, B: 50, A: 255},
			ParticleCount: 8,
			Speed:         18.0,
			Spread:        1.2,
			Life:          0.6,
			Size:          0.5,
		}
	case "scifi":
		// Energy sparks
		return EmitterConfig{
			Color:         color.RGBA{R: 100, G: 220, B: 255, A: 255},
			ParticleCount: 12,
			Speed:         22.0,
			Spread:        0.8,
			Life:          0.5,
			Size:          0.6,
		}
	case "horror":
		// Dim sparks
		return EmitterConfig{
			Color:         color.RGBA{R: 200, G: 150, B: 80, A: 200},
			ParticleCount: 6,
			Speed:         15.0,
			Spread:        1.0,
			Life:          0.4,
			Size:          0.4,
		}
	case "cyberpunk":
		// Neon sparks
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 0, B: 200, A: 255},
			ParticleCount: 10,
			Speed:         20.0,
			Spread:        0.9,
			Life:          0.55,
			Size:          0.7,
		}
	case "postapoc":
		// Rusty sparks
		return EmitterConfig{
			Color:         color.RGBA{R: 220, G: 120, B: 30, A: 220},
			ParticleCount: 9,
			Speed:         16.0,
			Spread:        1.1,
			Life:          0.5,
			Size:          0.5,
		}
	default:
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 200, B: 50, A: 255},
			ParticleCount: 10,
			Speed:         20.0,
			Spread:        1.0,
			Life:          0.5,
			Size:          0.5,
		}
	}
}

// BloodSplatterEmitter creates blood splatter effects for damage.
type BloodSplatterEmitter struct {
	system  *ParticleSystem
	genreID string
}

// NewBloodSplatterEmitter creates a blood splatter emitter.
func NewBloodSplatterEmitter(system *ParticleSystem, genreID string) *BloodSplatterEmitter {
	return &BloodSplatterEmitter{
		system:  system,
		genreID: genreID,
	}
}

// Emit spawns blood splatter at the given position with impact velocity.
func (e *BloodSplatterEmitter) Emit(x, y, impactVX, impactVY float64) {
	config := e.getConfig()

	for i := 0; i < config.ParticleCount; i++ {
		angle := e.system.rng.Float64() * 2 * math.Pi
		velocity := config.Speed * (0.3 + e.system.rng.Float64()*0.7)

		vx := math.Cos(angle)*velocity + impactVX*0.5
		vy := math.Sin(angle)*velocity + impactVY*0.5
		vz := (e.system.rng.Float64()*2 - 1) * config.Spread

		e.system.Spawn(x, y, 0, vx, vy, vz, config.Life, config.Size, config.Color)
	}
}

// getConfig returns genre-specific blood splatter configuration.
func (e *BloodSplatterEmitter) getConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		// Bright red blood
		return EmitterConfig{
			Color:         color.RGBA{R: 200, G: 0, B: 0, A: 255},
			ParticleCount: 15,
			Speed:         10.0,
			Spread:        2.0,
			Life:          1.2,
			Size:          0.8,
		}
	case "scifi":
		// Synthetic/alien fluids
		return EmitterConfig{
			Color:         color.RGBA{R: 100, G: 255, B: 100, A: 255},
			ParticleCount: 12,
			Speed:         12.0,
			Spread:        1.5,
			Life:          1.0,
			Size:          0.7,
		}
	case "horror":
		// Dark, viscous blood
		return EmitterConfig{
			Color:         color.RGBA{R: 120, G: 0, B: 0, A: 255},
			ParticleCount: 18,
			Speed:         8.0,
			Spread:        2.5,
			Life:          1.5,
			Size:          1.0,
		}
	case "cyberpunk":
		// Neon-tinted blood
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 0, B: 100, A: 255},
			ParticleCount: 14,
			Speed:         11.0,
			Spread:        1.8,
			Life:          1.1,
			Size:          0.75,
		}
	case "postapoc":
		// Dirty blood
		return EmitterConfig{
			Color:         color.RGBA{R: 160, G: 20, B: 0, A: 230},
			ParticleCount: 16,
			Speed:         9.0,
			Spread:        2.2,
			Life:          1.3,
			Size:          0.85,
		}
	default:
		return EmitterConfig{
			Color:         color.RGBA{R: 180, G: 0, B: 0, A: 255},
			ParticleCount: 15,
			Speed:         10.0,
			Spread:        2.0,
			Life:          1.2,
			Size:          0.8,
		}
	}
}

// ExplosionEmitter creates explosion effects with genre-specific characteristics.
type ExplosionEmitter struct {
	system  *ParticleSystem
	genreID string
}

// NewExplosionEmitter creates an explosion emitter.
func NewExplosionEmitter(system *ParticleSystem, genreID string) *ExplosionEmitter {
	return &ExplosionEmitter{
		system:  system,
		genreID: genreID,
	}
}

// Emit spawns an explosion at the given position with intensity.
func (e *ExplosionEmitter) Emit(x, y, intensity float64) {
	config := e.getConfig()

	// Core flash
	coreCount := int(float64(config.ParticleCount) * 0.3 * intensity)
	e.system.SpawnBurst(x, y, 0, coreCount, config.Speed*1.5, config.Spread*0.5, config.Life*0.3, config.Size*1.5, config.Color)

	// Outer fire/energy
	fireConfig := e.getFireConfig()
	fireCount := int(float64(fireConfig.ParticleCount) * intensity)
	e.system.SpawnBurst(x, y, 0, fireCount, fireConfig.Speed, fireConfig.Spread, fireConfig.Life, fireConfig.Size, fireConfig.Color)

	// Smoke/debris
	smokeConfig := e.getSmokeConfig()
	smokeCount := int(float64(smokeConfig.ParticleCount) * intensity)
	e.system.SpawnBurst(x, y, 0, smokeCount, smokeConfig.Speed, smokeConfig.Spread, smokeConfig.Life*1.5, smokeConfig.Size*1.2, smokeConfig.Color)
}

// getConfig returns core flash configuration.
func (e *ExplosionEmitter) getConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		// Magical explosion
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 220, B: 150, A: 255},
			ParticleCount: 10,
			Speed:         25.0,
			Spread:        2.0,
			Life:          0.3,
			Size:          2.5,
		}
	case "scifi":
		// Energy explosion
		return EmitterConfig{
			Color:         color.RGBA{R: 150, G: 200, B: 255, A: 255},
			ParticleCount: 12,
			Speed:         30.0,
			Spread:        1.5,
			Life:          0.25,
			Size:          2.0,
		}
	case "horror":
		// Muted explosion
		return EmitterConfig{
			Color:         color.RGBA{R: 200, G: 180, B: 150, A: 220},
			ParticleCount: 8,
			Speed:         20.0,
			Spread:        2.2,
			Life:          0.35,
			Size:          2.8,
		}
	case "cyberpunk":
		// Neon explosion
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 100, B: 255, A: 255},
			ParticleCount: 11,
			Speed:         28.0,
			Spread:        1.7,
			Life:          0.28,
			Size:          2.2,
		}
	case "postapoc":
		// Dirty explosion
		return EmitterConfig{
			Color:         color.RGBA{R: 220, G: 180, B: 100, A: 240},
			ParticleCount: 9,
			Speed:         22.0,
			Spread:        2.5,
			Life:          0.32,
			Size:          3.0,
		}
	default:
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 255, B: 200, A: 255},
			ParticleCount: 10,
			Speed:         25.0,
			Spread:        2.0,
			Life:          0.3,
			Size:          2.5,
		}
	}
}

// getFireConfig returns fire/energy configuration.
func (e *ExplosionEmitter) getFireConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 100, B: 0, A: 255},
			ParticleCount: 20,
			Speed:         15.0,
			Spread:        1.5,
			Life:          0.6,
			Size:          1.8,
		}
	case "scifi":
		return EmitterConfig{
			Color:         color.RGBA{R: 100, G: 150, B: 255, A: 255},
			ParticleCount: 18,
			Speed:         18.0,
			Spread:        1.2,
			Life:          0.5,
			Size:          1.5,
		}
	case "horror":
		return EmitterConfig{
			Color:         color.RGBA{R: 180, G: 80, B: 0, A: 200},
			ParticleCount: 16,
			Speed:         12.0,
			Spread:        1.8,
			Life:          0.7,
			Size:          2.0,
		}
	case "cyberpunk":
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 0, B: 200, A: 255},
			ParticleCount: 19,
			Speed:         17.0,
			Spread:        1.3,
			Life:          0.55,
			Size:          1.6,
		}
	case "postapoc":
		return EmitterConfig{
			Color:         color.RGBA{R: 200, G: 90, B: 0, A: 230},
			ParticleCount: 17,
			Speed:         13.0,
			Spread:        2.0,
			Life:          0.65,
			Size:          2.2,
		}
	default:
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 100, B: 0, A: 255},
			ParticleCount: 20,
			Speed:         15.0,
			Spread:        1.5,
			Life:          0.6,
			Size:          1.8,
		}
	}
}

// getSmokeConfig returns smoke/debris configuration.
func (e *ExplosionEmitter) getSmokeConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		return EmitterConfig{
			Color:         color.RGBA{R: 80, G: 70, B: 60, A: 200},
			ParticleCount: 15,
			Speed:         8.0,
			Spread:        1.0,
			Life:          2.0,
			Size:          2.0,
		}
	case "scifi":
		return EmitterConfig{
			Color:         color.RGBA{R: 100, G: 100, B: 120, A: 180},
			ParticleCount: 12,
			Speed:         10.0,
			Spread:        0.8,
			Life:          1.8,
			Size:          1.8,
		}
	case "horror":
		return EmitterConfig{
			Color:         color.RGBA{R: 60, G: 50, B: 50, A: 220},
			ParticleCount: 18,
			Speed:         6.0,
			Spread:        1.2,
			Life:          2.5,
			Size:          2.5,
		}
	case "cyberpunk":
		return EmitterConfig{
			Color:         color.RGBA{R: 120, G: 80, B: 140, A: 190},
			ParticleCount: 14,
			Speed:         9.0,
			Spread:        0.9,
			Life:          1.9,
			Size:          1.9,
		}
	case "postapoc":
		return EmitterConfig{
			Color:         color.RGBA{R: 70, G: 60, B: 50, A: 210},
			ParticleCount: 16,
			Speed:         7.0,
			Spread:        1.5,
			Life:          2.3,
			Size:          2.3,
		}
	default:
		return EmitterConfig{
			Color:         color.RGBA{R: 80, G: 80, B: 80, A: 200},
			ParticleCount: 15,
			Speed:         8.0,
			Spread:        1.0,
			Life:          2.0,
			Size:          2.0,
		}
	}
}

// EnergyDischargeEmitter creates energy weapon discharge effects.
type EnergyDischargeEmitter struct {
	system  *ParticleSystem
	genreID string
}

// NewEnergyDischargeEmitter creates an energy discharge emitter.
func NewEnergyDischargeEmitter(system *ParticleSystem, genreID string) *EnergyDischargeEmitter {
	return &EnergyDischargeEmitter{
		system:  system,
		genreID: genreID,
	}
}

// Emit spawns energy discharge at the given position with direction.
func (e *EnergyDischargeEmitter) Emit(x, y, dirX, dirY float64) {
	config := e.getConfig()

	baseAngle := math.Atan2(dirY, dirX)

	for i := 0; i < config.ParticleCount; i++ {
		angle := baseAngle + (e.system.rng.Float64()*2-1)*config.Spread
		velocity := config.Speed * (0.7 + e.system.rng.Float64()*0.3)

		vx := math.Cos(angle) * velocity
		vy := math.Sin(angle) * velocity
		vz := (e.system.rng.Float64()*2 - 1) * config.Spread * 2

		e.system.Spawn(x, y, 0, vx, vy, vz, config.Life, config.Size, config.Color)
	}
}

// getConfig returns genre-specific energy discharge configuration.
func (e *EnergyDischargeEmitter) getConfig() EmitterConfig {
	switch e.genreID {
	case "fantasy":
		// Magical energy (bright white/blue)
		return EmitterConfig{
			Color:         color.RGBA{R: 200, G: 220, B: 255, A: 255},
			ParticleCount: 10,
			Speed:         20.0,
			Spread:        0.5,
			Life:          0.4,
			Size:          1.0,
		}
	case "scifi":
		// Plasma discharge (blue/cyan)
		return EmitterConfig{
			Color:         color.RGBA{R: 100, G: 200, B: 255, A: 255},
			ParticleCount: 14,
			Speed:         28.0,
			Spread:        0.3,
			Life:          0.3,
			Size:          0.8,
		}
	case "horror":
		// Unnatural energy (green/sickly)
		return EmitterConfig{
			Color:         color.RGBA{R: 150, G: 255, B: 100, A: 220},
			ParticleCount: 8,
			Speed:         18.0,
			Spread:        0.6,
			Life:          0.5,
			Size:          1.2,
		}
	case "cyberpunk":
		// Neon discharge (magenta/cyan)
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 0, B: 255, A: 255},
			ParticleCount: 12,
			Speed:         24.0,
			Spread:        0.4,
			Life:          0.35,
			Size:          0.9,
		}
	case "postapoc":
		// Unstable energy (yellow/orange)
		return EmitterConfig{
			Color:         color.RGBA{R: 255, G: 180, B: 50, A: 240},
			ParticleCount: 11,
			Speed:         22.0,
			Spread:        0.7,
			Life:          0.45,
			Size:          1.1,
		}
	default:
		return EmitterConfig{
			Color:         color.RGBA{R: 150, G: 200, B: 255, A: 255},
			ParticleCount: 12,
			Speed:         24.0,
			Spread:        0.4,
			Life:          0.35,
			Size:          1.0,
		}
	}
}
