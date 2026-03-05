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

	width, height, valid := mapDimensions(worldMap)
	if !valid {
		return
	}

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
	return getGenreHazardTypes(s.genre)
}

// isValidLocation checks if a map location can contain a hazard.
func (s *ECSSystem) isValidLocation(worldMap [][]int, x, y int) bool {
	return isValidHazardLocation(worldMap, x, y)
}

// createHazardComponent creates a hazard component of the specified type using shared configuration.
func (s *ECSSystem) createHazardComponent(hType Type, rng *rand.Rand) *HazardComponent {
	cfg := getHazardConfig(hType)
	h := &HazardComponent{
		Type:             hType,
		State:            cfg.State,
		Width:            cfg.Width,
		Height:           cfg.Height,
		ChargeDuration:   cfg.ChargeDuration,
		ActiveDuration:   cfg.ActiveDuration,
		CooldownDuration: cfg.CooldownDuration,
		StatusEffect:     cfg.StatusEffect,
		Color:            cfg.Color,
		Persistent:       cfg.Persistent,
	}

	// Apply damage with random variance
	h.Damage = cfg.BaseDamage
	if cfg.DamageVariance > 0 {
		h.Damage += rng.Intn(cfg.DamageVariance)
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
