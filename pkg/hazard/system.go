package hazard

import (
	"math"
	"math/rand"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// ECSSystem manages environmental hazards using the ECS architecture.
type ECSSystem struct {
	rng   *rand.Rand
	genre string
}

// NewECSSystem creates a new ECS-based hazard system.
func NewECSSystem(seed int64) *ECSSystem {
	return &ECSSystem{
		rng:   rand.New(rand.NewSource(seed)),
		genre: "fantasy",
	}
}

// SetGenre changes the hazard system's genre.
func (s *ECSSystem) SetGenre(genre string) {
	s.genre = genre
}

// Update advances hazard states and timers (implements System interface).
func (s *ECSSystem) Update(w *engine.World) {
	// Query all entities with HazardComponent
	hazardType := reflect.TypeOf((*HazardComponent)(nil))
	entities := w.Query(hazardType)

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, hazardType)
		if !ok {
			continue
		}

		hazard, ok := comp.(*HazardComponent)
		if !ok {
			continue
		}

		// Advance timer
		hazard.Timer += 1.0 / 60.0 // Assuming 60 FPS

		// State machine for cycling hazards
		if hazard.CycleDuration > 0 && (hazard.State != StateActive || !hazard.Persistent) {
			cycleTime := math.Mod(hazard.Timer, hazard.CycleDuration)

			if cycleTime < hazard.ChargeDuration {
				hazard.State = StateCharging
				hazard.Triggered = false
			} else if cycleTime < hazard.ChargeDuration+hazard.ActiveDuration {
				hazard.State = StateActive
			} else {
				hazard.State = StateCooldown
			}
		}
	}
}

// GenerateHazards procedurally places hazards as entities in the world.
func (s *ECSSystem) GenerateHazards(w *engine.World, worldMap [][]int, seed int64) {
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
	placedCount := 0

	for placedCount < targetCount && attempts < maxAttempts {
		attempts++

		x := 1 + localRNG.Intn(width-2)
		y := 1 + localRNG.Intn(height-2)

		// Check if location is valid (floor, not near edges)
		if !s.isValidLocation(worldMap, x, y) {
			continue
		}

		// Create hazard entity
		hType := hazardTypes[localRNG.Intn(len(hazardTypes))]
		entity := w.AddEntity()

		// Add position component
		w.AddComponent(entity, &PositionComponent{
			X: float64(x) + 0.5,
			Y: float64(y) + 0.5,
		})

		// Add hazard component
		hazardComp := s.createHazardComponent(hType, localRNG)
		w.AddComponent(entity, hazardComp)

		placedCount++
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "hazard_ecs",
		"count":       placedCount,
		"genre":       s.genre,
	}).Debug("Generated environmental hazards")
}

// getGenreHazards returns hazard types appropriate for the current genre.
func (s *ECSSystem) getGenreHazards() []Type {
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
func (s *ECSSystem) isValidLocation(worldMap [][]int, x, y int) bool {
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

// createHazardComponent creates a hazard component of the specified type.
func (s *ECSSystem) createHazardComponent(hType Type, rng *rand.Rand) *HazardComponent {
	h := &HazardComponent{
		Type:   hType,
		State:  StateInactive,
		Width:  1.0,
		Height: 1.0,
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

// CheckCollision tests if a position collides with any active hazard entity.
// Returns (hit, damage, statusEffect).
func (s *ECSSystem) CheckCollision(w *engine.World, x, y float64) (bool, int, string) {
	hazardType := reflect.TypeOf((*HazardComponent)(nil))
	posType := reflect.TypeOf((*PositionComponent)(nil))

	// Query all entities with both HazardComponent and PositionComponent
	entities := w.Query(hazardType, posType)

	for _, entity := range entities {
		hazardComp, ok := w.GetComponent(entity, hazardType)
		if !ok {
			continue
		}
		hazard, ok := hazardComp.(*HazardComponent)
		if !ok {
			continue
		}

		if hazard.State != StateActive {
			continue
		}

		posComp, ok := w.GetComponent(entity, posType)
		if !ok {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// AABB collision
		dx := math.Abs(x - pos.X)
		dy := math.Abs(y - pos.Y)

		if dx < hazard.Width/2 && dy < hazard.Height/2 {
			// For one-shot hazards, only trigger once per activation
			if !hazard.Persistent && hazard.Triggered {
				continue
			}
			hazard.Triggered = true
			return true, hazard.Damage, hazard.StatusEffect
		}
	}
	return false, 0, ""
}

// GetHazardsForRendering returns all hazard entities with their position and component data.
func (s *ECSSystem) GetHazardsForRendering(w *engine.World) []HazardRenderData {
	hazardType := reflect.TypeOf((*HazardComponent)(nil))
	posType := reflect.TypeOf((*PositionComponent)(nil))

	entities := w.Query(hazardType, posType)
	result := make([]HazardRenderData, 0, len(entities))

	for _, entity := range entities {
		hazardComp, ok := w.GetComponent(entity, hazardType)
		if !ok {
			continue
		}
		hazard, ok := hazardComp.(*HazardComponent)
		if !ok {
			continue
		}

		posComp, ok := w.GetComponent(entity, posType)
		if !ok {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		result = append(result, HazardRenderData{
			X:      pos.X,
			Y:      pos.Y,
			Width:  hazard.Width,
			Height: hazard.Height,
			Type:   hazard.Type,
			State:  hazard.State,
			Color:  hazard.Color,
		})
	}

	return result
}

// HazardRenderData contains information needed to render a hazard.
type HazardRenderData struct {
	X, Y          float64
	Width, Height float64
	Type          Type
	State         State
	Color         uint32
}
