// Package status manages status effects applied to entities.
package status

import (
	"reflect"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// EffectType categorizes status effects for stacking rules.
type EffectType int

const (
	EffectDamage EffectType = iota // Damage over time
	EffectHeal                     // Healing over time
	EffectSlow                     // Movement speed reduction
	EffectStun                     // Cannot move or act
	EffectBuff                     // Stat increase
	EffectDebuff                   // Stat decrease
)

// Effect represents a status effect template.
type Effect struct {
	Name          string
	Type          EffectType
	Duration      time.Duration
	DamagePerTick float64 // Positive = damage, negative = healing
	TickInterval  time.Duration
	SpeedMul      float64 // Movement speed multiplier (1.0 = normal, 0.5 = half speed)
	Stackable     bool    // Can multiple instances exist on same entity
	VisualColor   uint32  // RGBA color for visual effects
}

// ActiveEffect represents an effect instance on an entity.
type ActiveEffect struct {
	EffectName    string
	TimeRemaining time.Duration
	LastTick      time.Time
	TickInterval  time.Duration
	DamagePerTick float64
	SpeedMul      float64
	VisualColor   uint32
}

// StatusComponent is an ECS component tracking active effects on an entity.
type StatusComponent struct {
	ActiveEffects []ActiveEffect
}

// Registry holds all known status effect templates.
type Registry struct {
	effects map[string]Effect
	logger  *logrus.Entry
}

// NewRegistry creates a new status effect registry.
func NewRegistry() *Registry {
	r := &Registry{
		effects: make(map[string]Effect),
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "status",
		}),
	}
	r.loadDefaultEffects("fantasy")
	return r
}

// loadDefaultEffects populates genre-specific effect templates.
func (r *Registry) loadDefaultEffects(genreID string) {
	switch genreID {
	case "fantasy":
		r.effects = map[string]Effect{
			"poisoned":     {Name: "poisoned", Type: EffectDamage, Duration: 10 * time.Second, DamagePerTick: 2.0, TickInterval: time.Second, SpeedMul: 0.9, Stackable: false, VisualColor: 0x88FF0088},
			"burning":      {Name: "burning", Type: EffectDamage, Duration: 5 * time.Second, DamagePerTick: 5.0, TickInterval: 500 * time.Millisecond, SpeedMul: 1.0, Stackable: false, VisualColor: 0xFF880088},
			"bleeding":     {Name: "bleeding", Type: EffectDamage, Duration: 15 * time.Second, DamagePerTick: 1.0, TickInterval: time.Second, SpeedMul: 0.95, Stackable: true, VisualColor: 0xAA000088},
			"stunned":      {Name: "stunned", Type: EffectStun, Duration: 2 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.0, Stackable: false, VisualColor: 0xFFFF0088},
			"regeneration": {Name: "regeneration", Type: EffectHeal, Duration: 10 * time.Second, DamagePerTick: -2.0, TickInterval: time.Second, SpeedMul: 1.0, Stackable: false, VisualColor: 0x00FF0088},
			"blessed":      {Name: "blessed", Type: EffectBuff, Duration: 30 * time.Second, DamagePerTick: -1.0, TickInterval: 2 * time.Second, SpeedMul: 1.1, Stackable: false, VisualColor: 0xFFFFAA88},
			"cursed":       {Name: "cursed", Type: EffectDebuff, Duration: 20 * time.Second, DamagePerTick: 0.5, TickInterval: 2 * time.Second, SpeedMul: 0.85, Stackable: false, VisualColor: 0x660066AA},
			"slowed":       {Name: "slowed", Type: EffectSlow, Duration: 5 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.5, Stackable: false, VisualColor: 0x4444FFAA},
		}
	case "scifi":
		r.effects = map[string]Effect{
			"irradiated":  {Name: "irradiated", Type: EffectDamage, Duration: 15 * time.Second, DamagePerTick: 1.5, TickInterval: time.Second, SpeedMul: 0.9, Stackable: true, VisualColor: 0x00FF0088},
			"burning":     {Name: "burning", Type: EffectDamage, Duration: 4 * time.Second, DamagePerTick: 6.0, TickInterval: 500 * time.Millisecond, SpeedMul: 1.0, Stackable: false, VisualColor: 0xFF440088},
			"emp_stunned": {Name: "emp_stunned", Type: EffectStun, Duration: 3 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.0, Stackable: false, VisualColor: 0x00FFFF88},
			"nanoheal":    {Name: "nanoheal", Type: EffectHeal, Duration: 8 * time.Second, DamagePerTick: -3.0, TickInterval: time.Second, SpeedMul: 1.0, Stackable: false, VisualColor: 0x00AAFF88},
			"overcharged": {Name: "overcharged", Type: EffectBuff, Duration: 15 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 1.2, Stackable: false, VisualColor: 0xFFAA0088},
			"corroded":    {Name: "corroded", Type: EffectDebuff, Duration: 12 * time.Second, DamagePerTick: 1.0, TickInterval: 2 * time.Second, SpeedMul: 0.9, Stackable: true, VisualColor: 0x88880088},
			"slowed":      {Name: "slowed", Type: EffectSlow, Duration: 6 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.6, Stackable: false, VisualColor: 0x0088FF88},
		}
	case "horror":
		r.effects = map[string]Effect{
			"poisoned":     {Name: "poisoned", Type: EffectDamage, Duration: 12 * time.Second, DamagePerTick: 2.5, TickInterval: time.Second, SpeedMul: 0.85, Stackable: false, VisualColor: 0x44AA4488},
			"bleeding":     {Name: "bleeding", Type: EffectDamage, Duration: 20 * time.Second, DamagePerTick: 1.5, TickInterval: time.Second, SpeedMul: 0.9, Stackable: true, VisualColor: 0x880000AA},
			"terrified":    {Name: "terrified", Type: EffectDebuff, Duration: 8 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.7, Stackable: false, VisualColor: 0xAA00AAAA},
			"infected":     {Name: "infected", Type: EffectDamage, Duration: 25 * time.Second, DamagePerTick: 0.8, TickInterval: 2 * time.Second, SpeedMul: 0.95, Stackable: true, VisualColor: 0x00882288},
			"stunned":      {Name: "stunned", Type: EffectStun, Duration: 2500 * time.Millisecond, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.0, Stackable: false, VisualColor: 0xFFFFFF88},
			"regeneration": {Name: "regeneration", Type: EffectHeal, Duration: 10 * time.Second, DamagePerTick: -1.5, TickInterval: time.Second, SpeedMul: 1.0, Stackable: false, VisualColor: 0x44FF4488},
		}
	case "cyberpunk":
		r.effects = map[string]Effect{
			"burning":      {Name: "burning", Type: EffectDamage, Duration: 4 * time.Second, DamagePerTick: 7.0, TickInterval: 500 * time.Millisecond, SpeedMul: 1.0, Stackable: false, VisualColor: 0xFF00FF88},
			"hacked":       {Name: "hacked", Type: EffectDebuff, Duration: 10 * time.Second, DamagePerTick: 0.5, TickInterval: time.Second, SpeedMul: 0.8, Stackable: false, VisualColor: 0x00FF8888},
			"emp_stunned":  {Name: "emp_stunned", Type: EffectStun, Duration: 3 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.0, Stackable: false, VisualColor: 0x00FFFFAA},
			"stim_boosted": {Name: "stim_boosted", Type: EffectBuff, Duration: 12 * time.Second, DamagePerTick: -1.0, TickInterval: 2 * time.Second, SpeedMul: 1.3, Stackable: false, VisualColor: 0xFF0088AA},
			"nanoheal":     {Name: "nanoheal", Type: EffectHeal, Duration: 6 * time.Second, DamagePerTick: -4.0, TickInterval: time.Second, SpeedMul: 1.0, Stackable: false, VisualColor: 0x00DDFF88},
			"glitched":     {Name: "glitched", Type: EffectDebuff, Duration: 5 * time.Second, DamagePerTick: 1.0, TickInterval: time.Second, SpeedMul: 0.9, Stackable: false, VisualColor: 0xFF88FF88},
			"slowed":       {Name: "slowed", Type: EffectSlow, Duration: 5 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.5, Stackable: false, VisualColor: 0x8800FF88},
		}
	case "postapoc":
		r.effects = map[string]Effect{
			"irradiated": {Name: "irradiated", Type: EffectDamage, Duration: 20 * time.Second, DamagePerTick: 1.2, TickInterval: time.Second, SpeedMul: 0.9, Stackable: true, VisualColor: 0x88FF0088},
			"poisoned":   {Name: "poisoned", Type: EffectDamage, Duration: 15 * time.Second, DamagePerTick: 2.0, TickInterval: time.Second, SpeedMul: 0.85, Stackable: false, VisualColor: 0x668800AA},
			"bleeding":   {Name: "bleeding", Type: EffectDamage, Duration: 18 * time.Second, DamagePerTick: 1.3, TickInterval: time.Second, SpeedMul: 0.92, Stackable: true, VisualColor: 0xAA000088},
			"stunned":    {Name: "stunned", Type: EffectStun, Duration: 2 * time.Second, DamagePerTick: 0, TickInterval: time.Second, SpeedMul: 0.0, Stackable: false, VisualColor: 0xCCCCCC88},
			"stimmed":    {Name: "stimmed", Type: EffectBuff, Duration: 15 * time.Second, DamagePerTick: -0.5, TickInterval: 2 * time.Second, SpeedMul: 1.15, Stackable: false, VisualColor: 0xFFAA4488},
			"infected":   {Name: "infected", Type: EffectDamage, Duration: 30 * time.Second, DamagePerTick: 0.7, TickInterval: 2 * time.Second, SpeedMul: 0.95, Stackable: true, VisualColor: 0x448800AA},
			"corroded":   {Name: "corroded", Type: EffectDebuff, Duration: 10 * time.Second, DamagePerTick: 1.5, TickInterval: 2 * time.Second, SpeedMul: 0.9, Stackable: false, VisualColor: 0x886600AA},
		}
	default:
		r.loadDefaultEffects("fantasy")
	}
}

// Apply adds a status effect to an entity.
func (r *Registry) Apply(name string) {
	// Deprecated - use ApplyToEntity instead
}

// ApplyToEntity applies a status effect to an entity in the ECS world.
func (r *Registry) ApplyToEntity(w *engine.World, entity engine.Entity, effectName string) {
	template, exists := r.effects[effectName]
	if !exists {
		r.logger.Warnf("Unknown status effect: %s", effectName)
		return
	}

	statusType := reflect.TypeOf(&StatusComponent{})
	comp, ok := w.GetComponent(entity, statusType)

	var statusComp *StatusComponent
	if !ok {
		// Create new status component
		statusComp = &StatusComponent{
			ActiveEffects: []ActiveEffect{},
		}
		w.AddComponent(entity, statusComp)
	} else {
		statusComp = comp.(*StatusComponent)
	}

	// Check if effect already exists and is not stackable
	if !template.Stackable {
		for i, active := range statusComp.ActiveEffects {
			if active.EffectName == effectName {
				// Refresh duration
				statusComp.ActiveEffects[i].TimeRemaining = template.Duration
				r.logger.Debugf("Refreshed %s on entity %d", effectName, entity)
				return
			}
		}
	}

	// Add new effect instance
	newEffect := ActiveEffect{
		EffectName:    effectName,
		TimeRemaining: template.Duration,
		LastTick:      time.Now(),
		TickInterval:  template.TickInterval,
		DamagePerTick: template.DamagePerTick,
		SpeedMul:      template.SpeedMul,
		VisualColor:   template.VisualColor,
	}

	statusComp.ActiveEffects = append(statusComp.ActiveEffects, newEffect)
	r.logger.Debugf("Applied %s to entity %d", effectName, entity)
}

// Tick advances all active effects by one tick.
func (r *Registry) Tick() {
	// Deprecated - use System.Update instead
}

// System processes all entities with status effects each frame.
type System struct {
	registry *Registry
	logger   *logrus.Entry
}

// NewSystem creates a new status effect system.
func NewSystem(registry *Registry) *System {
	return &System{
		registry: registry,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "status_system",
		}),
	}
}

// Update processes status effects for all entities.
func (s *System) Update(w *engine.World) {
	statusType := reflect.TypeOf(&StatusComponent{})
	healthType := reflect.TypeOf(&engine.Health{})

	entities := w.Query(statusType)
	now := time.Now()

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, statusType)
		if !ok {
			continue
		}

		statusComp := comp.(*StatusComponent)

		// Process each active effect
		remainingEffects := []ActiveEffect{}
		for _, effect := range statusComp.ActiveEffects {
			// Tick effect if interval elapsed
			if now.Sub(effect.LastTick) >= effect.TickInterval {
				effect.LastTick = now

				// Apply damage/healing if entity has health
				if effect.DamagePerTick != 0 {
					healthComp, hasHealth := w.GetComponent(entity, healthType)
					if hasHealth {
						health := healthComp.(*engine.Health)
						damage := int(effect.DamagePerTick)

						health.Current -= damage
						if health.Current > health.Max {
							health.Current = health.Max
						}
						if health.Current < 0 {
							health.Current = 0
						}

						if damage > 0 {
							s.logger.Debugf("Entity %d took %d damage from %s (%d/%d HP)", entity, damage, effect.EffectName, health.Current, health.Max)
						} else {
							s.logger.Debugf("Entity %d healed %d from %s (%d/%d HP)", entity, -damage, effect.EffectName, health.Current, health.Max)
						}
					}
				}
			}

			// Update duration
			effect.TimeRemaining -= 16 * time.Millisecond // ~60 FPS frame time

			if effect.TimeRemaining > 0 {
				remainingEffects = append(remainingEffects, effect)
			} else {
				s.logger.Debugf("Effect %s expired on entity %d", effect.EffectName, entity)
			}
		}

		statusComp.ActiveEffects = remainingEffects
	}
}

// GetSpeedMultiplier returns the combined speed multiplier from all active effects.
func GetSpeedMultiplier(w *engine.World, entity engine.Entity) float64 {
	statusType := reflect.TypeOf(&StatusComponent{})
	comp, ok := w.GetComponent(entity, statusType)
	if !ok {
		return 1.0
	}

	statusComp := comp.(*StatusComponent)
	multiplier := 1.0

	for _, effect := range statusComp.ActiveEffects {
		// Take the minimum speed multiplier (most restrictive)
		if effect.SpeedMul < multiplier {
			multiplier = effect.SpeedMul
		}
	}

	return multiplier
}

// IsStunned checks if an entity is currently stunned.
func IsStunned(w *engine.World, entity engine.Entity) bool {
	return GetSpeedMultiplier(w, entity) == 0.0
}

var currentGenre = "fantasy"

// SetGenre configures status effects for a genre.
func SetGenre(genreID string) {
	currentGenre = genreID
}

// GetCurrentGenre returns the current global genre setting.
func GetCurrentGenre() string {
	return currentGenre
}
