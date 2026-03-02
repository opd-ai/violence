package trap

import (
	"fmt"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// PositionComponent provides entity position.
type PositionComponent struct {
	PosX, PosY float64
}

// HealthComponent provides entity health.
type HealthComponent struct {
	Current, Max int
}

// System manages all traps in the game world using ECS architecture.
type System struct {
	traps  []*Trap
	rng    *rng.RNG
	genre  string
	logger *logrus.Entry
}

// NewSystem creates a new trap system.
func NewSystem(seed int64) *System {
	return &System{
		traps: make([]*Trap, 0, 32),
		rng:   rng.NewRNG(uint64(seed)),
		genre: "fantasy",
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "trap",
		}),
	}
}

// SetGenre sets the genre for trap generation.
func (s *System) SetGenre(genre string) {
	s.genre = genre
	s.logger.WithField("genre", genre).Debug("Set trap system genre")
}

// Update processes all traps and checks for entity triggers (implements engine.System).
func (s *System) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0 // Assuming 60 FPS

	// Update trap states
	for _, trap := range s.traps {
		trap.Update(deltaTime)
	}

	// Query all entities with position components
	posType := reflect.TypeOf((*PositionComponent)(nil))
	entities := w.Query(posType)

	// Check entity interactions with traps
	for _, entity := range entities {
		s.checkEntityTraps(w, entity, deltaTime)
	}
}

// checkEntityTraps tests an entity against all traps.
func (s *System) checkEntityTraps(w *engine.World, entity engine.Entity, deltaTime float64) {
	posType := reflect.TypeOf((*PositionComponent)(nil))
	posComp, ok := w.GetComponent(entity, posType)
	if !ok {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Get entity skills if available
	detectSkill := 0
	disarmSkill := 0

	info := &TriggerInfo{
		EntityID:    fmt.Sprintf("%d", entity),
		X:           pos.PosX,
		Y:           pos.PosY,
		IsPlayer:    false, // TODO: check player component
		DetectSkill: detectSkill,
		DisarmSkill: disarmSkill,
	}

	for _, trap := range s.traps {
		dx := pos.PosX - trap.X
		dy := pos.PosY - trap.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Only process nearby traps
		if dist > trap.EffectRadius+2.0 {
			continue
		}

		result := trap.CheckTrigger(info)

		if result.Detected {
			s.logger.WithFields(logrus.Fields{
				"entity":    entity,
				"trap_type": trap.Type,
				"position":  [2]float64{trap.X, trap.Y},
			}).Debug("Trap detected")
		}

		if result.Disarmed {
			s.logger.WithFields(logrus.Fields{
				"entity":    entity,
				"trap_type": trap.Type,
			}).Info("Trap disarmed")
		}

		if result.Triggered {
			s.applyTrapEffect(w, entity, result)
		}
	}
}

// applyTrapEffect applies a trap's effects to an entity.
func (s *System) applyTrapEffect(w *engine.World, entity engine.Entity, result *EffectResult) {
	if result.Damage > 0 {
		healthType := reflect.TypeOf((*HealthComponent)(nil))
		healthComp, ok := w.GetComponent(entity, healthType)
		if ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				health.Current -= result.Damage
				if health.Current < 0 {
					health.Current = 0
				}

				s.logger.WithFields(logrus.Fields{
					"entity": entity,
					"damage": result.Damage,
				}).Debug("Trap dealt damage")
			}
		}
	}

	if result.StatusEffect != "" {
		s.logger.WithFields(logrus.Fields{
			"entity": entity,
			"status": result.StatusEffect,
		}).Debug("Trap applied status effect")
	}

	if result.KnockbackX != 0 || result.KnockbackY != 0 {
		s.logger.WithFields(logrus.Fields{
			"entity":    entity,
			"knockback": [2]float64{result.KnockbackX, result.KnockbackY},
		}).Debug("Trap applied knockback")
	}

	if result.SpawnProjectile {
		s.logger.WithFields(logrus.Fields{
			"entity":          entity,
			"projectile_type": result.ProjectileType,
			"angle":           result.ProjectileAngle,
		}).Debug("Trap spawned projectile")
	}
}

// GenerateTraps procedurally places traps in a dungeon map.
func (s *System) GenerateTraps(worldMap [][]int, seed int64) {
	s.traps = make([]*Trap, 0, 32)

	if len(worldMap) == 0 || len(worldMap[0]) == 0 {
		return
	}

	width := len(worldMap[0])
	height := len(worldMap)

	localRNG := rng.NewRNG(uint64(seed))
	trapTypes := GetGenreTraps(s.genre)

	// Place traps in corridors and doorways
	targetCount := 3 + localRNG.Intn(8)
	attempts := 0
	maxAttempts := 200

	for len(s.traps) < targetCount && attempts < maxAttempts {
		attempts++

		x := 2 + localRNG.Intn(width-4)
		y := 2 + localRNG.Intn(height-4)

		// Check if location is valid (floor tile)
		if worldMap[y][x] != 2 && worldMap[y][x] != 20 && worldMap[y][x] != 21 &&
			worldMap[y][x] != 22 && worldMap[y][x] != 23 && worldMap[y][x] != 24 {
			continue
		}

		// Prefer corridors (narrow passages)
		neighborCount := 0
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dy == 0 && dx == 0 {
					continue
				}
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					if worldMap[ny][nx] >= 2 {
						neighborCount++
					}
				}
			}
		}

		// Prefer locations with 3-5 floor neighbors (corridors, doorways)
		if neighborCount < 3 || neighborCount > 5 {
			continue
		}

		// Create trap
		trapType := trapTypes[localRNG.Intn(len(trapTypes))]
		trap := NewTrap(trapType, float64(x)+0.5, float64(y)+0.5, seed+int64(x*1000+y))
		trap.Genre = s.genre
		s.traps = append(s.traps, trap)
	}

	s.logger.WithFields(logrus.Fields{
		"count": len(s.traps),
		"genre": s.genre,
	}).Info("Generated traps")
}

// GetTraps returns all traps in the system.
func (s *System) GetTraps() []*Trap {
	return s.traps
}

// AddTrap adds a trap to the system.
func (s *System) AddTrap(trap *Trap) {
	s.traps = append(s.traps, trap)
}

// RemoveTrap removes a trap from the system.
func (s *System) RemoveTrap(trap *Trap) {
	for i, t := range s.traps {
		if t == trap {
			s.traps = append(s.traps[:i], s.traps[i+1:]...)
			break
		}
	}
}

// ClearTraps removes all traps.
func (s *System) ClearTraps() {
	s.traps = make([]*Trap, 0, 32)
}
