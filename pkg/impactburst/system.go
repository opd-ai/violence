package impactburst

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages impact burst effects with visually realistic rendering.
type System struct {
	genreID string
	logger  *logrus.Entry
	rng     *rand.Rand

	// Global impact list for non-entity-attached impacts
	globalImpacts []Impact

	// Maximum concurrent impacts for performance
	maxGlobalImpacts int

	// Profile cache for fast lookup
	profiles map[profileKey]ImpactProfile
}

type profileKey struct {
	impactType ImpactType
	material   MaterialType
	genreID    string
}

// NewSystem creates a new impact burst rendering system.
func NewSystem(genreID string, seed int64) *System {
	s := &System{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "impactburst",
			"package":     "impactburst",
		}),
		rng:              rand.New(rand.NewSource(seed)),
		globalImpacts:    make([]Impact, 0, 64),
		maxGlobalImpacts: 64,
		profiles:         make(map[profileKey]ImpactProfile),
	}

	s.initializeProfiles()
	return s
}

// SetGenre updates the genre-specific visual style.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.initializeProfiles()
	s.logger = s.logger.WithField("genre", genreID)
}

// Update processes all active impacts, advancing their animation state.
func (s *System) Update(w *engine.World) {
	const deltaTime = 1.0 / 60.0

	// Update global impacts
	s.updateImpacts(s.globalImpacts, deltaTime)

	// Remove expired global impacts
	active := s.globalImpacts[:0]
	for i := range s.globalImpacts {
		if s.globalImpacts[i].Age < s.globalImpacts[i].MaxAge {
			active = append(active, s.globalImpacts[i])
		}
	}
	s.globalImpacts = active
}

// updateImpacts advances the animation state of a slice of impacts.
func (s *System) updateImpacts(impacts []Impact, deltaTime float64) {
	for i := range impacts {
		imp := &impacts[i]
		imp.Age += deltaTime

		// Calculate progress (0-1)
		progress := imp.Age / imp.MaxAge
		if progress > 1.0 {
			progress = 1.0
		}

		profile := s.getProfile(imp.Type, imp.Material)

		// Update shockwave
		if profile.HasShockwave {
			shockProgress := imp.Age / profile.ShockwaveDuration
			if shockProgress <= 1.0 {
				// Ease-out expansion (fast start, slow end)
				eased := 1.0 - math.Pow(1.0-shockProgress, 3)
				imp.ShockwaveRadius = profile.ShockwaveMaxRadius * eased * imp.Intensity

				// Inverse-square falloff for alpha
				imp.ShockwaveAlpha = 1.0 - shockProgress*shockProgress
			} else {
				imp.ShockwaveAlpha = 0
			}
		}

		// Update glow
		if profile.HasGlow {
			glowProgress := imp.Age / profile.GlowDuration
			if glowProgress <= 1.0 {
				// Smooth fade with slight pulse
				imp.GlowIntensity = (1.0 - glowProgress) * (1.0 + 0.2*math.Sin(glowProgress*math.Pi*4))
			} else {
				imp.GlowIntensity = 0
			}
		}

		// Update flash (quick flash at start)
		flashProgress := imp.Age / profile.FlashDuration
		if flashProgress <= 1.0 {
			imp.FlashAlpha = 1.0 - flashProgress*flashProgress
		} else {
			imp.FlashAlpha = 0
		}

		// Update debris particles
		for j := range imp.Debris {
			debris := &imp.Debris[j]
			debris.Age += deltaTime

			// Apply velocity
			debris.X += debris.VX * deltaTime
			debris.Y += debris.VY * deltaTime

			// Apply gravity
			debris.VY += debris.GravityScale * 100.0 * deltaTime

			// Apply rotation
			debris.Rotation += debris.RotationVel * deltaTime

			// Slow down over time (air resistance)
			debris.VX *= 0.98
			debris.VY *= 0.98
		}
	}
}

// SpawnImpact creates a new impact burst at the specified world position.
func (s *System) SpawnImpact(x, y, angle float64, impactType ImpactType, material MaterialType, intensity float64) {
	if len(s.globalImpacts) >= s.maxGlobalImpacts {
		// Remove oldest impact to make room
		s.globalImpacts = s.globalImpacts[1:]
	}

	profile := s.getProfile(impactType, material)
	impact := s.createImpact(x, y, angle, impactType, material, intensity, profile)
	s.globalImpacts = append(s.globalImpacts, impact)

	s.logger.WithFields(logrus.Fields{
		"impact_type": impactType,
		"material":    material,
		"intensity":   intensity,
		"x":           x,
		"y":           y,
	}).Debug("spawned impact burst")
}

// createImpact initializes a new Impact with debris particles.
func (s *System) createImpact(x, y, angle float64, impactType ImpactType, material MaterialType, intensity float64, profile ImpactProfile) Impact {
	imp := Impact{
		X:         x,
		Y:         y,
		Angle:     angle,
		Type:      impactType,
		Material:  material,
		Intensity: clamp(intensity, 0.1, 2.0),
		Age:       0,
		MaxAge:    profile.Duration,
		Debris:    make([]DebrisParticle, 0, profile.ParticleCount+profile.ChunkCount),
	}

	// Spawn debris particles
	particleCount := int(float64(profile.ParticleCount) * intensity)
	for i := 0; i < particleCount; i++ {
		imp.Debris = append(imp.Debris, s.createDebrisParticle(angle, profile, false))
	}

	// Spawn debris chunks
	chunkCount := int(float64(profile.ChunkCount) * intensity)
	for i := 0; i < chunkCount; i++ {
		imp.Debris = append(imp.Debris, s.createDebrisParticle(angle, profile, true))
	}

	return imp
}

// createDebrisParticle generates a single debris particle with randomized properties.
func (s *System) createDebrisParticle(impactAngle float64, profile ImpactProfile, isChunk bool) DebrisParticle {
	var angle float64
	if profile.HasDirectionalDebris {
		// Debris flies away from impact direction (opposite to incoming)
		baseAngle := impactAngle + math.Pi
		spread := profile.ParticleSpread
		angle = baseAngle + (s.rng.Float64()*2-1)*spread
	} else {
		// Omnidirectional debris
		angle = s.rng.Float64() * 2 * math.Pi
	}

	speed := profile.ParticleSpeed * (0.5 + s.rng.Float64())
	if isChunk {
		speed *= 0.6 // Chunks move slower
	}

	var size float64
	var col color.RGBA
	var gravity float64
	if isChunk {
		size = profile.ChunkSize * (0.7 + s.rng.Float64()*0.6)
		col = profile.ChunkColor
		gravity = profile.ChunkGravity
	} else {
		size = profile.BaseParticleSize * (0.6 + s.rng.Float64()*0.8)
		// Randomly choose between debris and secondary color
		if s.rng.Float64() < 0.3 {
			col = profile.SecondaryColor
		} else {
			col = profile.DebrisColor
		}
		gravity = profile.ParticleGravity
	}

	return DebrisParticle{
		X:            0,
		Y:            0,
		VX:           math.Cos(angle) * speed,
		VY:           math.Sin(angle) * speed,
		Size:         size,
		Color:        col,
		Age:          0,
		MaxAge:       profile.Duration * (0.7 + s.rng.Float64()*0.6),
		Rotation:     s.rng.Float64() * 2 * math.Pi,
		RotationVel:  (s.rng.Float64()*2 - 1) * 10.0,
		IsChunk:      isChunk,
		GravityScale: gravity,
	}
}

// GetGlobalImpacts returns the list of active global impacts for rendering.
func (s *System) GetGlobalImpacts() []Impact {
	return s.globalImpacts
}

// getProfile retrieves or creates a cached profile for the given combination.
func (s *System) getProfile(impactType ImpactType, material MaterialType) ImpactProfile {
	key := profileKey{impactType, material, s.genreID}
	if profile, ok := s.profiles[key]; ok {
		return profile
	}

	profile := s.buildProfile(impactType, material)
	s.profiles[key] = profile
	return profile
}

// buildProfile constructs an ImpactProfile for the given parameters.
func (s *System) buildProfile(impactType ImpactType, material MaterialType) ImpactProfile {
	// Get base profile for material
	base := s.getMaterialProfile(material)

	// Modify based on impact type
	switch impactType {
	case ImpactCritical:
		base.ParticleCount = int(float64(base.ParticleCount) * 2.0)
		base.ChunkCount = int(float64(base.ChunkCount) * 1.5)
		base.ParticleSpeed *= 1.5
		base.ShockwaveMaxRadius *= 1.3
		base.Duration *= 1.2
		base.HasGlow = true
		base.GlowColor = brightenColor(base.PrimaryColor, 1.5)

	case ImpactExplosion:
		base.ParticleCount = int(float64(base.ParticleCount) * 3.0)
		base.ChunkCount = int(float64(base.ChunkCount) * 2.0)
		base.ParticleSpeed *= 2.0
		base.ShockwaveMaxRadius *= 2.0
		base.ShockwaveRings = 3
		base.Duration *= 1.5
		base.HasShockwave = true
		base.HasGlow = true
		base.HasDirectionalDebris = false
		base.ParticleSpread = math.Pi * 2

	case ImpactMagic:
		base.ParticleCount = int(float64(base.ParticleCount) * 1.5)
		base.ChunkCount = 0 // Magic doesn't create physical chunks
		base.ParticleSpeed *= 0.8
		base.Duration *= 1.8
		base.HasGlow = true
		base.HasDirectionalDebris = false
		base.ParticleGravity = -0.5 // Magic particles float upward

	case ImpactBlock:
		base.ParticleCount = int(float64(base.ParticleCount) * 0.4)
		base.ChunkCount = 0
		base.ParticleSpeed *= 0.7
		base.Duration *= 0.6
		base.ShockwaveMaxRadius *= 0.5
		base.HasGlow = false

	case ImpactDeath:
		base.ParticleCount = int(float64(base.ParticleCount) * 4.0)
		base.ChunkCount = int(float64(base.ChunkCount) * 3.0)
		base.ParticleSpeed *= 1.2
		base.ShockwaveMaxRadius *= 1.5
		base.Duration *= 2.0
		base.HasShockwave = true
		base.HasGlow = true
		base.HasDirectionalDebris = false
		base.ParticleSpread = math.Pi * 2
	}

	return base
}

// getMaterialProfile returns the base profile for a given material.
func (s *System) getMaterialProfile(material MaterialType) ImpactProfile {
	switch s.genreID {
	case "fantasy":
		return s.getFantasyMaterialProfile(material)
	case "scifi":
		return s.getSciFiMaterialProfile(material)
	case "horror":
		return s.getHorrorMaterialProfile(material)
	case "cyberpunk":
		return s.getCyberpunkMaterialProfile(material)
	default:
		return s.getFantasyMaterialProfile(material)
	}
}

func (s *System) getFantasyMaterialProfile(material MaterialType) ImpactProfile {
	switch material {
	case MaterialFlesh:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 200, G: 30, B: 30, A: 255},
			SecondaryColor:       color.RGBA{R: 140, G: 20, B: 20, A: 220},
			GlowColor:            color.RGBA{R: 255, G: 100, B: 100, A: 180},
			ShockwaveColor:       color.RGBA{R: 180, G: 40, B: 40, A: 150},
			DebrisColor:          color.RGBA{R: 160, G: 25, B: 25, A: 255},
			ChunkColor:           color.RGBA{R: 120, G: 15, B: 15, A: 255},
			Duration:             0.5,
			ShockwaveDuration:    0.2,
			FlashDuration:        0.08,
			GlowDuration:         0.3,
			ShockwaveMaxRadius:   25.0,
			ShockwaveWidth:       2.0,
			BaseParticleSize:     2.0,
			ChunkSize:            4.0,
			ParticleCount:        18,
			ChunkCount:           3,
			ShockwaveRings:       1,
			ParticleSpeed:        60.0,
			ParticleSpread:       math.Pi * 0.5,
			ParticleGravity:      1.0,
			ChunkGravity:         1.5,
			HasGlow:              false,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialMetal:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 255, G: 230, B: 120, A: 255},
			SecondaryColor:       color.RGBA{R: 255, G: 180, B: 80, A: 220},
			GlowColor:            color.RGBA{R: 255, G: 240, B: 200, A: 200},
			ShockwaveColor:       color.RGBA{R: 200, G: 200, B: 180, A: 120},
			DebrisColor:          color.RGBA{R: 180, G: 180, B: 160, A: 255},
			ChunkColor:           color.RGBA{R: 140, G: 140, B: 130, A: 255},
			Duration:             0.35,
			ShockwaveDuration:    0.15,
			FlashDuration:        0.05,
			GlowDuration:         0.2,
			ShockwaveMaxRadius:   20.0,
			ShockwaveWidth:       1.5,
			BaseParticleSize:     1.5,
			ChunkSize:            2.5,
			ParticleCount:        25,
			ChunkCount:           5,
			ShockwaveRings:       2,
			ParticleSpeed:        90.0,
			ParticleSpread:       math.Pi * 0.7,
			ParticleGravity:      0.8,
			ChunkGravity:         1.2,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialStone:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 150, G: 140, B: 130, A: 255},
			SecondaryColor:       color.RGBA{R: 100, G: 95, B: 90, A: 220},
			GlowColor:            color.RGBA{R: 180, G: 170, B: 160, A: 120},
			ShockwaveColor:       color.RGBA{R: 130, G: 125, B: 120, A: 100},
			DebrisColor:          color.RGBA{R: 120, G: 115, B: 110, A: 255},
			ChunkColor:           color.RGBA{R: 90, G: 85, B: 80, A: 255},
			Duration:             0.6,
			ShockwaveDuration:    0.25,
			FlashDuration:        0.06,
			GlowDuration:         0.15,
			ShockwaveMaxRadius:   22.0,
			ShockwaveWidth:       2.5,
			BaseParticleSize:     2.5,
			ChunkSize:            5.0,
			ParticleCount:        20,
			ChunkCount:           6,
			ShockwaveRings:       1,
			ParticleSpeed:        50.0,
			ParticleSpread:       math.Pi * 0.4,
			ParticleGravity:      1.5,
			ChunkGravity:         2.0,
			HasGlow:              false,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialWood:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 180, G: 140, B: 90, A: 255},
			SecondaryColor:       color.RGBA{R: 140, G: 100, B: 60, A: 220},
			GlowColor:            color.RGBA{R: 200, G: 160, B: 100, A: 100},
			ShockwaveColor:       color.RGBA{R: 160, G: 130, B: 90, A: 80},
			DebrisColor:          color.RGBA{R: 160, G: 120, B: 80, A: 255},
			ChunkColor:           color.RGBA{R: 120, G: 90, B: 50, A: 255},
			Duration:             0.5,
			ShockwaveDuration:    0.2,
			FlashDuration:        0.04,
			GlowDuration:         0.1,
			ShockwaveMaxRadius:   18.0,
			ShockwaveWidth:       1.5,
			BaseParticleSize:     1.8,
			ChunkSize:            3.5,
			ParticleCount:        15,
			ChunkCount:           4,
			ShockwaveRings:       1,
			ParticleSpeed:        45.0,
			ParticleSpread:       math.Pi * 0.35,
			ParticleGravity:      1.2,
			ChunkGravity:         1.8,
			HasGlow:              false,
			HasShockwave:         false,
			HasDirectionalDebris: true,
		}

	case MaterialEnergy:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 100, G: 180, B: 255, A: 255},
			SecondaryColor:       color.RGBA{R: 180, G: 120, B: 255, A: 220},
			GlowColor:            color.RGBA{R: 150, G: 200, B: 255, A: 255},
			ShockwaveColor:       color.RGBA{R: 120, G: 160, B: 255, A: 180},
			DebrisColor:          color.RGBA{R: 140, G: 180, B: 255, A: 200},
			ChunkColor:           color.RGBA{R: 100, G: 140, B: 220, A: 180},
			Duration:             0.4,
			ShockwaveDuration:    0.18,
			FlashDuration:        0.1,
			GlowDuration:         0.35,
			ShockwaveMaxRadius:   30.0,
			ShockwaveWidth:       2.0,
			BaseParticleSize:     1.8,
			ChunkSize:            0,
			ParticleCount:        28,
			ChunkCount:           0,
			ShockwaveRings:       2,
			ParticleSpeed:        70.0,
			ParticleSpread:       math.Pi,
			ParticleGravity:      -0.3, // Floats upward
			ChunkGravity:         0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: false,
		}

	case MaterialEthereal:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 200, G: 200, B: 255, A: 180},
			SecondaryColor:       color.RGBA{R: 160, G: 160, B: 220, A: 140},
			GlowColor:            color.RGBA{R: 220, G: 220, B: 255, A: 150},
			ShockwaveColor:       color.RGBA{R: 180, G: 180, B: 240, A: 100},
			DebrisColor:          color.RGBA{R: 180, G: 180, B: 230, A: 160},
			ChunkColor:           color.RGBA{R: 150, G: 150, B: 200, A: 120},
			Duration:             0.8,
			ShockwaveDuration:    0.4,
			FlashDuration:        0.15,
			GlowDuration:         0.6,
			ShockwaveMaxRadius:   35.0,
			ShockwaveWidth:       1.5,
			BaseParticleSize:     2.2,
			ChunkSize:            0,
			ParticleCount:        12,
			ChunkCount:           0,
			ShockwaveRings:       2,
			ParticleSpeed:        30.0,
			ParticleSpread:       math.Pi * 1.5,
			ParticleGravity:      -0.5,
			ChunkGravity:         0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: false,
		}

	default:
		return s.getFantasyMaterialProfile(MaterialFlesh)
	}
}

func (s *System) getSciFiMaterialProfile(material MaterialType) ImpactProfile {
	switch material {
	case MaterialFlesh:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 220, G: 50, B: 70, A: 255},
			SecondaryColor:       color.RGBA{R: 180, G: 30, B: 50, A: 220},
			GlowColor:            color.RGBA{R: 255, G: 120, B: 140, A: 180},
			ShockwaveColor:       color.RGBA{R: 200, G: 60, B: 80, A: 150},
			DebrisColor:          color.RGBA{R: 200, G: 40, B: 60, A: 255},
			ChunkColor:           color.RGBA{R: 150, G: 25, B: 40, A: 255},
			Duration:             0.4,
			ShockwaveDuration:    0.15,
			FlashDuration:        0.06,
			GlowDuration:         0.25,
			ShockwaveMaxRadius:   22.0,
			ShockwaveWidth:       1.5,
			BaseParticleSize:     1.8,
			ChunkSize:            3.0,
			ParticleCount:        20,
			ChunkCount:           4,
			ShockwaveRings:       1,
			ParticleSpeed:        70.0,
			ParticleSpread:       math.Pi * 0.45,
			ParticleGravity:      0.8,
			ChunkGravity:         1.2,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialMetal:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 100, G: 200, B: 255, A: 255},
			SecondaryColor:       color.RGBA{R: 255, G: 255, B: 255, A: 240},
			GlowColor:            color.RGBA{R: 150, G: 220, B: 255, A: 220},
			ShockwaveColor:       color.RGBA{R: 120, G: 180, B: 220, A: 160},
			DebrisColor:          color.RGBA{R: 200, G: 200, B: 210, A: 255},
			ChunkColor:           color.RGBA{R: 160, G: 160, B: 170, A: 255},
			Duration:             0.3,
			ShockwaveDuration:    0.12,
			FlashDuration:        0.04,
			GlowDuration:         0.2,
			ShockwaveMaxRadius:   25.0,
			ShockwaveWidth:       2.0,
			BaseParticleSize:     1.3,
			ChunkSize:            2.0,
			ParticleCount:        30,
			ChunkCount:           8,
			ShockwaveRings:       2,
			ParticleSpeed:        100.0,
			ParticleSpread:       math.Pi * 0.6,
			ParticleGravity:      0.5,
			ChunkGravity:         1.0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialEnergy:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 0, G: 255, B: 200, A: 255},
			SecondaryColor:       color.RGBA{R: 100, G: 255, B: 255, A: 230},
			GlowColor:            color.RGBA{R: 50, G: 255, B: 230, A: 255},
			ShockwaveColor:       color.RGBA{R: 0, G: 220, B: 180, A: 200},
			DebrisColor:          color.RGBA{R: 50, G: 230, B: 200, A: 220},
			ChunkColor:           color.RGBA{R: 0, G: 180, B: 150, A: 180},
			Duration:             0.35,
			ShockwaveDuration:    0.15,
			FlashDuration:        0.08,
			GlowDuration:         0.3,
			ShockwaveMaxRadius:   35.0,
			ShockwaveWidth:       2.5,
			BaseParticleSize:     1.5,
			ChunkSize:            0,
			ParticleCount:        35,
			ChunkCount:           0,
			ShockwaveRings:       3,
			ParticleSpeed:        80.0,
			ParticleSpread:       math.Pi * 1.5,
			ParticleGravity:      -0.2,
			ChunkGravity:         0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: false,
		}

	default:
		return s.getSciFiMaterialProfile(MaterialMetal)
	}
}

func (s *System) getHorrorMaterialProfile(material MaterialType) ImpactProfile {
	switch material {
	case MaterialFlesh:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 100, G: 20, B: 20, A: 255},
			SecondaryColor:       color.RGBA{R: 60, G: 10, B: 10, A: 200},
			GlowColor:            color.RGBA{R: 120, G: 40, B: 40, A: 120},
			ShockwaveColor:       color.RGBA{R: 80, G: 20, B: 20, A: 100},
			DebrisColor:          color.RGBA{R: 90, G: 15, B: 15, A: 255},
			ChunkColor:           color.RGBA{R: 60, G: 10, B: 10, A: 255},
			Duration:             0.7,
			ShockwaveDuration:    0.3,
			FlashDuration:        0.1,
			GlowDuration:         0.4,
			ShockwaveMaxRadius:   20.0,
			ShockwaveWidth:       2.0,
			BaseParticleSize:     2.5,
			ChunkSize:            5.0,
			ParticleCount:        15,
			ChunkCount:           4,
			ShockwaveRings:       1,
			ParticleSpeed:        40.0,
			ParticleSpread:       math.Pi * 0.4,
			ParticleGravity:      1.5,
			ChunkGravity:         2.0,
			HasGlow:              false,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialEthereal:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 80, G: 100, B: 80, A: 160},
			SecondaryColor:       color.RGBA{R: 50, G: 70, B: 50, A: 120},
			GlowColor:            color.RGBA{R: 100, G: 130, B: 100, A: 140},
			ShockwaveColor:       color.RGBA{R: 70, G: 90, B: 70, A: 80},
			DebrisColor:          color.RGBA{R: 70, G: 90, B: 70, A: 140},
			ChunkColor:           color.RGBA{R: 50, G: 70, B: 50, A: 100},
			Duration:             1.0,
			ShockwaveDuration:    0.5,
			FlashDuration:        0.15,
			GlowDuration:         0.7,
			ShockwaveMaxRadius:   40.0,
			ShockwaveWidth:       1.5,
			BaseParticleSize:     2.0,
			ChunkSize:            0,
			ParticleCount:        10,
			ChunkCount:           0,
			ShockwaveRings:       2,
			ParticleSpeed:        25.0,
			ParticleSpread:       math.Pi,
			ParticleGravity:      -0.3,
			ChunkGravity:         0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: false,
		}

	default:
		return s.getHorrorMaterialProfile(MaterialFlesh)
	}
}

func (s *System) getCyberpunkMaterialProfile(material MaterialType) ImpactProfile {
	switch material {
	case MaterialFlesh:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 255, G: 0, B: 100, A: 255},
			SecondaryColor:       color.RGBA{R: 200, G: 0, B: 150, A: 220},
			GlowColor:            color.RGBA{R: 255, G: 50, B: 150, A: 200},
			ShockwaveColor:       color.RGBA{R: 220, G: 0, B: 120, A: 150},
			DebrisColor:          color.RGBA{R: 230, G: 20, B: 100, A: 255},
			ChunkColor:           color.RGBA{R: 180, G: 0, B: 80, A: 255},
			Duration:             0.4,
			ShockwaveDuration:    0.18,
			FlashDuration:        0.06,
			GlowDuration:         0.3,
			ShockwaveMaxRadius:   24.0,
			ShockwaveWidth:       2.0,
			BaseParticleSize:     1.8,
			ChunkSize:            3.0,
			ParticleCount:        20,
			ChunkCount:           3,
			ShockwaveRings:       2,
			ParticleSpeed:        65.0,
			ParticleSpread:       math.Pi * 0.5,
			ParticleGravity:      0.8,
			ChunkGravity:         1.2,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialMetal:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 0, G: 255, B: 255, A: 255},
			SecondaryColor:       color.RGBA{R: 255, G: 255, B: 0, A: 240},
			GlowColor:            color.RGBA{R: 100, G: 255, B: 255, A: 220},
			ShockwaveColor:       color.RGBA{R: 0, G: 200, B: 200, A: 180},
			DebrisColor:          color.RGBA{R: 200, G: 200, B: 180, A: 255},
			ChunkColor:           color.RGBA{R: 150, G: 150, B: 140, A: 255},
			Duration:             0.3,
			ShockwaveDuration:    0.12,
			FlashDuration:        0.04,
			GlowDuration:         0.2,
			ShockwaveMaxRadius:   28.0,
			ShockwaveWidth:       2.0,
			BaseParticleSize:     1.4,
			ChunkSize:            2.2,
			ParticleCount:        28,
			ChunkCount:           7,
			ShockwaveRings:       2,
			ParticleSpeed:        95.0,
			ParticleSpread:       math.Pi * 0.55,
			ParticleGravity:      0.6,
			ChunkGravity:         1.0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: true,
		}

	case MaterialEnergy:
		return ImpactProfile{
			PrimaryColor:         color.RGBA{R: 255, G: 0, B: 255, A: 255},
			SecondaryColor:       color.RGBA{R: 0, G: 255, B: 255, A: 230},
			GlowColor:            color.RGBA{R: 200, G: 100, B: 255, A: 255},
			ShockwaveColor:       color.RGBA{R: 180, G: 0, B: 180, A: 180},
			DebrisColor:          color.RGBA{R: 200, G: 50, B: 200, A: 220},
			ChunkColor:           color.RGBA{R: 150, G: 0, B: 150, A: 180},
			Duration:             0.35,
			ShockwaveDuration:    0.15,
			FlashDuration:        0.07,
			GlowDuration:         0.3,
			ShockwaveMaxRadius:   32.0,
			ShockwaveWidth:       2.5,
			BaseParticleSize:     1.6,
			ChunkSize:            0,
			ParticleCount:        32,
			ChunkCount:           0,
			ShockwaveRings:       3,
			ParticleSpeed:        75.0,
			ParticleSpread:       math.Pi * 1.2,
			ParticleGravity:      -0.2,
			ChunkGravity:         0,
			HasGlow:              true,
			HasShockwave:         true,
			HasDirectionalDebris: false,
		}

	default:
		return s.getCyberpunkMaterialProfile(MaterialMetal)
	}
}

// initializeProfiles pre-builds common profile combinations.
func (s *System) initializeProfiles() {
	s.profiles = make(map[profileKey]ImpactProfile)

	// Pre-cache common combinations for fast lookup
	materials := []MaterialType{MaterialFlesh, MaterialMetal, MaterialStone, MaterialWood, MaterialEnergy, MaterialEthereal}
	types := []ImpactType{ImpactMelee, ImpactProjectile, ImpactExplosion, ImpactMagic, ImpactCritical, ImpactBlock, ImpactDeath}

	for _, mat := range materials {
		for _, typ := range types {
			key := profileKey{typ, mat, s.genreID}
			s.profiles[key] = s.buildProfile(typ, mat)
		}
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func brightenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(clamp(float64(c.R)*factor, 0, 255)),
		G: uint8(clamp(float64(c.G)*factor, 0, 255)),
		B: uint8(clamp(float64(c.B)*factor, 0, 255)),
		A: c.A,
	}
}
