// Package hazard provides environmental hazards such as spike traps, fire grates,
// poison vents, and other genre-specific dangers that damage players on collision.
package hazard

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// Type represents the kind of environmental hazard.
type Type int

const (
	TypeSpikeTrap Type = iota
	TypeFireGrate
	TypePoisonVent
	TypeElectricFloor
	TypeFallingRocks
	TypeAcidPool
	TypeLaserGrid
	TypeCryoField
	TypePlasmaJet
	TypeGravityWell
)

// State represents the activation state of a hazard.
type State int

const (
	StateInactive State = iota
	StateCharging
	StateActive
	StateCooldown
)

// Hazard represents an environmental hazard in the game world.
type Hazard struct {
	Type             Type
	X, Y             float64
	Width, Height    float64
	State            State
	Timer            float64
	CycleDuration    float64
	ActiveDuration   float64
	ChargeDuration   float64
	CooldownDuration float64
	Damage           int
	StatusEffect     string
	Persistent       bool
	Triggered        bool
	Color            uint32
}

// System manages all environmental hazards.
type System struct {
	hazards []*Hazard
	rng     *rand.Rand
	genre   string
}

// NewSystem creates a new hazard system.
func NewSystem(seed int64) *System {
	return &System{
		hazards: make([]*Hazard, 0, 64),
		rng:     rand.New(rand.NewSource(seed)),
		genre:   "fantasy",
	}
}

// SetGenre changes the hazard system's genre.
func (s *System) SetGenre(genre string) {
	s.genre = genre
}

// GenerateHazards procedurally places hazards in a map.
func (s *System) GenerateHazards(worldMap [][]int, seed int64) {
	s.hazards = make([]*Hazard, 0, 64)
	localRNG := rand.New(rand.NewSource(seed))

	if len(worldMap) == 0 || len(worldMap[0]) == 0 {
		return
	}

	width := len(worldMap[0])
	height := len(worldMap)

	// Select hazard types based on genre
	hazardTypes := s.getGenreHazards()

	// Place hazards in valid locations
	attempts := 0
	maxAttempts := 100
	targetCount := 5 + localRNG.Intn(10)

	for len(s.hazards) < targetCount && attempts < maxAttempts {
		attempts++

		x := 1 + localRNG.Intn(width-2)
		y := 1 + localRNG.Intn(height-2)

		// Check if location is valid (floor, not near edges)
		if !s.isValidLocation(worldMap, x, y) {
			continue
		}

		// Create hazard
		hType := hazardTypes[localRNG.Intn(len(hazardTypes))]
		hazard := s.createHazard(hType, float64(x)+0.5, float64(y)+0.5, localRNG)
		s.hazards = append(s.hazards, hazard)
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "hazard",
		"count":       len(s.hazards),
		"genre":       s.genre,
	}).Debug("Generated environmental hazards")
}

// getGenreHazards returns hazard types appropriate for the current genre.
func (s *System) getGenreHazards() []Type {
	switch s.genre {
	case "fantasy":
		return []Type{TypeSpikeTrap, TypeFireGrate, TypePoisonVent, TypeFallingRocks, TypeAcidPool}
	case "scifi":
		return []Type{TypeElectricFloor, TypeLaserGrid, TypeCryoField, TypePlasmaJet, TypeGravityWell}
	case "horror":
		return []Type{TypeSpikeTrap, TypePoisonVent, TypeAcidPool, TypeFallingRocks}
	case "cyberpunk":
		return []Type{TypeElectricFloor, TypeLaserGrid, TypePlasmaJet, TypeGravityWell}
	default:
		return []Type{TypeSpikeTrap, TypeFireGrate, TypeElectricFloor, TypePoisonVent}
	}
}

// isValidLocation checks if a map location can contain a hazard.
func (s *System) isValidLocation(worldMap [][]int, x, y int) bool {
	width := len(worldMap[0])
	height := len(worldMap)

	// Must be a floor tile
	if worldMap[y][x] != 0 {
		return false
	}

	// Check neighbors for walls (prefer corridors and small rooms)
	wallCount := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dy == 0 && dx == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= width || ny < 0 || ny >= height {
				continue
			}
			if worldMap[ny][nx] != 0 {
				wallCount++
			}
		}
	}

	// Prefer locations with 2-5 adjacent walls (corridors, corners)
	return wallCount >= 2 && wallCount <= 5
}

// createHazard creates a hazard of the specified type.
func (s *System) createHazard(hType Type, x, y float64, rng *rand.Rand) *Hazard {
	h := &Hazard{
		Type:   hType,
		X:      x,
		Y:      y,
		Width:  1.0,
		Height: 1.0,
		State:  StateInactive,
	}

	switch hType {
	case TypeSpikeTrap:
		h.ChargeDuration = 0.5
		h.ActiveDuration = 0.3
		h.CooldownDuration = 2.0
		h.Damage = 15 + rng.Intn(10)
		h.Color = 0x808080
		h.Persistent = false

	case TypeFireGrate:
		h.ChargeDuration = 1.0
		h.ActiveDuration = 2.0
		h.CooldownDuration = 3.0
		h.Damage = 5
		h.StatusEffect = "burning"
		h.Color = 0xFF4400
		h.Persistent = true

	case TypePoisonVent:
		h.ChargeDuration = 0.8
		h.ActiveDuration = 3.0
		h.CooldownDuration = 5.0
		h.Damage = 3
		h.StatusEffect = "poisoned"
		h.Color = 0x00AA00
		h.Persistent = true
		h.Width = 1.5
		h.Height = 1.5

	case TypeElectricFloor:
		h.ChargeDuration = 1.5
		h.ActiveDuration = 1.0
		h.CooldownDuration = 3.0
		h.Damage = 20 + rng.Intn(15)
		h.StatusEffect = "stunned"
		h.Color = 0x00CCFF
		h.Persistent = false

	case TypeFallingRocks:
		h.ChargeDuration = 1.0
		h.ActiveDuration = 0.5
		h.CooldownDuration = 8.0
		h.Damage = 25 + rng.Intn(20)
		h.Color = 0x885533
		h.Persistent = false
		h.Width = 2.0
		h.Height = 2.0

	case TypeAcidPool:
		h.ActiveDuration = 999999.0 // Permanent
		h.Damage = 8
		h.StatusEffect = "corroded"
		h.Color = 0xAAFF00
		h.Persistent = true
		h.State = StateActive
		h.Width = 1.2
		h.Height = 1.2

	case TypeLaserGrid:
		h.ChargeDuration = 0.3
		h.ActiveDuration = 1.5
		h.CooldownDuration = 2.0
		h.Damage = 30 + rng.Intn(15)
		h.Color = 0xFF0000
		h.Persistent = false
		h.Width = 0.8
		h.Height = 2.0

	case TypeCryoField:
		h.ChargeDuration = 1.2
		h.ActiveDuration = 2.5
		h.CooldownDuration = 4.0
		h.Damage = 10
		h.StatusEffect = "slowed"
		h.Color = 0x88CCFF
		h.Persistent = true
		h.Width = 2.0
		h.Height = 2.0

	case TypePlasmaJet:
		h.ChargeDuration = 1.5
		h.ActiveDuration = 0.8
		h.CooldownDuration = 4.0
		h.Damage = 35 + rng.Intn(20)
		h.StatusEffect = "burning"
		h.Color = 0xFF00FF
		h.Persistent = false

	case TypeGravityWell:
		h.ChargeDuration = 2.0
		h.ActiveDuration = 3.0
		h.CooldownDuration = 6.0
		h.Damage = 5
		h.StatusEffect = "pulled"
		h.Color = 0x4400AA
		h.Persistent = true
		h.Width = 2.5
		h.Height = 2.5
	}

	h.CycleDuration = h.ChargeDuration + h.ActiveDuration + h.CooldownDuration
	h.Timer = rng.Float64() * h.CycleDuration // Random starting phase

	return h
}

// Update advances hazard states and timers.
func (s *System) Update(deltaTime float64) {
	for _, h := range s.hazards {
		h.Timer += deltaTime

		// State machine for cycling hazards
		if h.CycleDuration > 0 && h.State != StateActive || !h.Persistent {
			cycleTime := math.Mod(h.Timer, h.CycleDuration)

			if cycleTime < h.ChargeDuration {
				h.State = StateCharging
				h.Triggered = false
			} else if cycleTime < h.ChargeDuration+h.ActiveDuration {
				h.State = StateActive
			} else {
				h.State = StateCooldown
			}
		}
	}
}

// CheckCollision tests if a position collides with any active hazard.
// Returns (hit, damage, statusEffect).
func (s *System) CheckCollision(x, y float64) (bool, int, string) {
	for _, h := range s.hazards {
		if h.State != StateActive {
			continue
		}

		// AABB collision
		dx := math.Abs(x - h.X)
		dy := math.Abs(y - h.Y)

		if dx < h.Width/2 && dy < h.Height/2 {
			// For one-shot hazards, only trigger once per activation
			if !h.Persistent && h.Triggered {
				continue
			}
			h.Triggered = true
			return true, h.Damage, h.StatusEffect
		}
	}
	return false, 0, ""
}

// GetHazards returns all hazards for rendering.
func (s *System) GetHazards() []*Hazard {
	return s.hazards
}

// Clear removes all hazards.
func (s *System) Clear() {
	s.hazards = make([]*Hazard, 0, 64)
}
