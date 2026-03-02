package projectile

import (
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// SpatialGrid interface for collision broadphase.
type SpatialGrid interface {
	QueryRadius(x, y, radius float64) []engine.Entity
}

// ParticleSpawner interface for visual effects.
type ParticleSpawner interface {
	SpawnBurst(x, y, z float64, count int, speed, lifetime, fadeTime, gravity float64, col color.RGBA)
}

// FeedbackProvider interface for screen shake and flash.
type FeedbackProvider interface {
	AddScreenShake(intensity float64)
	AddHitFlash(intensity float64)
}

// System handles projectile movement, collision, and damage application.
type System struct {
	spatialGrid      SpatialGrid
	particleSpawner  ParticleSpawner
	feedbackProvider FeedbackProvider
	logger           *logrus.Entry
}

// NewSystem creates a new projectile system.
func NewSystem() *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system": "projectile",
		}),
	}
}

// SetSpatialGrid connects the spatial partitioning system.
func (s *System) SetSpatialGrid(grid SpatialGrid) {
	s.spatialGrid = grid
}

// SetParticleSpawner connects the particle system for trail and impact effects.
func (s *System) SetParticleSpawner(spawner ParticleSpawner) {
	s.particleSpawner = spawner
}

// SetFeedbackProvider connects the feedback system for impact effects.
func (s *System) SetFeedbackProvider(provider FeedbackProvider) {
	s.feedbackProvider = provider
}

// Update processes all projectile entities.
func (s *System) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0 // Assume 60 FPS

	projectileType := reflect.TypeOf((*ProjectileComponent)(nil))
	positionType := reflect.TypeOf((*engine.Position)(nil))

	entities := w.Query(projectileType, positionType)
	var toRemove []engine.Entity

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, projectileType)
		if !ok {
			continue
		}

		proj, ok := comp.(*ProjectileComponent)
		if !ok {
			continue
		}

		posComp, ok := w.GetComponent(entity, positionType)
		if !ok {
			continue
		}

		pos, ok := posComp.(*engine.Position)
		if !ok {
			continue
		}

		// Update lifetime
		proj.Lifetime -= deltaTime
		if proj.Lifetime <= 0 {
			s.handleProjectileDeath(w, entity, proj, pos.X, pos.Y)
			toRemove = append(toRemove, entity)
			continue
		}

		// Spawn trail particles
		if proj.TrailParticles && s.particleSpawner != nil {
			s.spawnTrailParticles(pos.X, pos.Y, proj)
		}

		// Update position
		pos.X += proj.VelX * deltaTime
		pos.Y += proj.VelY * deltaTime

		// Check collisions
		s.checkCollisions(w, entity, proj, pos.X, pos.Y, &toRemove)
	}

	// Remove expired projectiles
	for _, entity := range toRemove {
		w.RemoveEntity(entity)
	}
}

// spawnTrailParticles creates visual trail behind projectile.
func (s *System) spawnTrailParticles(x, y float64, proj *ProjectileComponent) {
	// Spawn fewer particles for older projectiles (fading trail)
	fadeRatio := proj.Lifetime / proj.MaxLifetime
	if fadeRatio < 0.3 {
		return
	}

	s.particleSpawner.SpawnBurst(x, y, 0, 1, 0.5, 0.3, 0.2, 0.0, proj.Color)
}

// checkCollisions performs collision detection and damage application.
func (s *System) checkCollisions(w *engine.World, entity engine.Entity, proj *ProjectileComponent, x, y float64, toRemove *[]engine.Entity) {
	if s.spatialGrid == nil {
		return
	}

	// Query nearby entities using spatial partitioning
	queryRadius := proj.Radius
	if proj.Shape == ShapeBeam {
		queryRadius = proj.BeamWidth * 2.0
	}

	nearby := s.spatialGrid.QueryRadius(x, y, queryRadius*2.0)

	positionType := reflect.TypeOf((*engine.Position)(nil))

	for _, target := range nearby {
		// Skip self
		if target == entity {
			continue
		}

		// Skip owner (friendly fire prevention)
		targetID := int(target)
		if targetID == proj.OwnerID {
			continue
		}

		// Skip already hit entities (for pierce mechanics)
		if proj.HitEntities[targetID] {
			continue
		}

		// Get target position
		targetPosComp, ok := w.GetComponent(target, positionType)
		if !ok {
			continue
		}
		targetPos, ok := targetPosComp.(*engine.Position)
		if !ok {
			continue
		}

		// Check shape-specific collision
		if !s.checkShapeCollision(x, y, targetPos.X, targetPos.Y, proj) {
			continue
		}

		// Apply damage
		s.applyDamage(w, target, proj, targetPos.X, targetPos.Y)

		// Track hit for pierce mechanics
		proj.HitEntities[targetID] = true

		// Decrement pierce count
		if proj.PierceCount >= 0 {
			proj.PierceCount--
			if proj.PierceCount < 0 {
				// Projectile is consumed
				s.handleProjectileDeath(w, entity, proj, x, y)
				*toRemove = append(*toRemove, entity)
				return
			}
		}
	}
}

// checkShapeCollision performs shape-specific collision detection.
func (s *System) checkShapeCollision(projX, projY, targetX, targetY float64, proj *ProjectileComponent) bool {
	switch proj.Shape {
	case ShapeCircle:
		// Circle-circle collision
		dx := targetX - projX
		dy := targetY - projY
		distSq := dx*dx + dy*dy
		combinedRadius := proj.Radius + 0.3 // Assume target radius ~0.3
		return distSq < combinedRadius*combinedRadius

	case ShapeBeam:
		// Point-to-line distance collision
		// For now, simplify to circle collision
		dx := targetX - projX
		dy := targetY - projY
		distSq := dx*dx + dy*dy
		return distSq < (proj.BeamWidth+0.3)*(proj.BeamWidth+0.3)

	case ShapeAOE:
		// Already exploded - should not be colliding
		return false
	}

	return false
}

// applyDamage applies projectile damage to a target entity.
func (s *System) applyDamage(w *engine.World, target engine.Entity, proj *ProjectileComponent, targetX, targetY float64) {
	// Get target's resistance component
	resistanceType := reflect.TypeOf((*ResistanceComponent)(nil))
	finalDamage := proj.Damage

	if comp, ok := w.GetComponent(target, resistanceType); ok {
		if resistance, ok := comp.(*ResistanceComponent); ok {
			finalDamage = CalculateDamage(proj.Damage, proj.DamageType, resistance.Resistances)
		}
	}

	// Log the damage event
	s.logger.WithFields(logrus.Fields{
		"target_id":   int(target),
		"damage":      finalDamage,
		"damage_type": DamageTypeNames[proj.DamageType],
	}).Debug("Projectile hit target")

	// Visual feedback
	if s.particleSpawner != nil {
		s.particleSpawner.SpawnBurst(targetX, targetY, 0, 8, 3.0, 0.5, 0.3, 0.5, proj.Color)
	}
}

// handleProjectileDeath handles projectile expiration or destruction.
func (s *System) handleProjectileDeath(w *engine.World, entity engine.Entity, proj *ProjectileComponent, x, y float64) {
	// Handle explosion
	if proj.ExplodeOnDeath && proj.ExplosionRadius > 0 {
		s.createExplosion(w, entity, proj, x, y)
	}

	// Impact particles
	if s.particleSpawner != nil && !proj.ExplodeOnDeath {
		s.particleSpawner.SpawnBurst(x, y, 0, 5, 2.0, 0.4, 0.2, 0.3, proj.Color)
	}
}

// createExplosion creates an AoE damage zone.
func (s *System) createExplosion(w *engine.World, entity engine.Entity, proj *ProjectileComponent, x, y float64) {
	if s.spatialGrid == nil {
		return
	}

	// Visual explosion
	if s.particleSpawner != nil {
		s.particleSpawner.SpawnBurst(x, y, 0, 30, 6.0, 1.0, 0.5, 0.2, proj.Color)
	}

	// Screen shake for explosions
	if s.feedbackProvider != nil {
		s.feedbackProvider.AddScreenShake(3.0)
	}

	// Query all entities in explosion radius
	nearby := s.spatialGrid.QueryRadius(x, y, proj.ExplosionRadius)

	positionType := reflect.TypeOf((*engine.Position)(nil))
	resistanceType := reflect.TypeOf((*ResistanceComponent)(nil))

	for _, target := range nearby {
		// Skip the projectile itself
		if target == entity {
			continue
		}

		// Skip owner
		targetID := int(target)
		if targetID == proj.OwnerID {
			continue
		}

		// Get target position
		targetPosComp, ok := w.GetComponent(target, positionType)
		if !ok {
			continue
		}
		targetPos, ok := targetPosComp.(*engine.Position)
		if !ok {
			continue
		}

		dx := targetPos.X - x
		dy := targetPos.Y - y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > proj.ExplosionRadius {
			continue
		}

		// Falloff damage based on distance
		falloff := 1.0 - (dist / proj.ExplosionRadius)
		explosionDamage := proj.Damage * falloff

		// Apply damage with resistance
		finalDamage := explosionDamage

		if comp, ok := w.GetComponent(target, resistanceType); ok {
			if resistance, ok := comp.(*ResistanceComponent); ok {
				finalDamage = CalculateDamage(explosionDamage, proj.DamageType, resistance.Resistances)
			}
		}

		s.logger.WithFields(logrus.Fields{
			"target_id":   targetID,
			"damage":      finalDamage,
			"damage_type": DamageTypeNames[proj.DamageType],
			"explosion":   true,
		}).Debug("Explosion hit target")

		// Impact particles on each hit entity
		if s.particleSpawner != nil {
			s.particleSpawner.SpawnBurst(targetPos.X, targetPos.Y, 0, 5, 2.5, 0.4, 0.2, 0.3, proj.Color)
		}
	}
}

// Type returns the system type for registration.
func (s *System) Type() string {
	return "ProjectileSystem"
}
