package projectile

import (
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
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

// DamageVisualProvider interface for damage-type visual effects.
type DamageVisualProvider interface {
	ApplyDamageVisual(w *engine.World, entity engine.Entity, damageTypeName string, damage, x, y float64)
}

// System handles projectile movement, collision, and damage application.
type System struct {
	spatialGrid          SpatialGrid
	particleSpawner      ParticleSpawner
	feedbackProvider     FeedbackProvider
	damageVisualProvider DamageVisualProvider
	logger               *logrus.Entry
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

// SetDamageVisualProvider connects the damage visual effects system.
func (s *System) SetDamageVisualProvider(provider DamageVisualProvider) {
	s.damageVisualProvider = provider
}

// Update processes all projectile entities.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime

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

	queryRadius := calculateQueryRadius(proj)
	nearby := s.spatialGrid.QueryRadius(x, y, queryRadius*2.0)
	positionType := reflect.TypeOf((*engine.Position)(nil))

	for _, target := range nearby {
		if shouldSkipCollisionTarget(target, entity, proj) {
			continue
		}

		targetPos := s.getTargetPosition(w, target, positionType)
		if targetPos == nil {
			continue
		}

		if !s.checkShapeCollision(x, y, targetPos.X, targetPos.Y, proj) {
			continue
		}

		s.processTargetHit(w, target, entity, proj, targetPos.X, targetPos.Y, x, y, toRemove)
		if proj.PierceCount < 0 {
			return
		}
	}
}

// calculateQueryRadius determines the spatial query radius based on projectile shape.
func calculateQueryRadius(proj *ProjectileComponent) float64 {
	if proj.Shape == ShapeBeam {
		return proj.BeamWidth * 2.0
	}
	return proj.Radius
}

// shouldSkipCollisionTarget checks if a target should be excluded from collision processing.
func shouldSkipCollisionTarget(target, entity engine.Entity, proj *ProjectileComponent) bool {
	if target == entity {
		return true
	}
	targetID := int(target)
	if targetID == proj.OwnerID {
		return true
	}
	return proj.HitEntities[targetID]
}

// processTargetHit handles damage application and pierce mechanics for a hit target.
func (s *System) processTargetHit(w *engine.World, target, entity engine.Entity, proj *ProjectileComponent, targetX, targetY, projX, projY float64, toRemove *[]engine.Entity) {
	s.applyDamage(w, target, proj, targetX, targetY)

	targetID := int(target)
	proj.HitEntities[targetID] = true

	if proj.PierceCount >= 0 {
		proj.PierceCount--
		if proj.PierceCount < 0 {
			s.handleProjectileDeath(w, entity, proj, projX, projY)
			*toRemove = append(*toRemove, entity)
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

	// Apply damage-type visual effects
	if s.damageVisualProvider != nil {
		s.damageVisualProvider.ApplyDamageVisual(w, target, DamageTypeNames[proj.DamageType], finalDamage, targetX, targetY)
	} else {
		// Fallback: basic visual feedback if damage visual system not connected
		if s.particleSpawner != nil {
			s.particleSpawner.SpawnBurst(targetX, targetY, 0, 8, 3.0, 0.5, 0.3, 0.5, proj.Color)
		}
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

	s.spawnExplosionEffects(x, y, proj.Color)
	nearby := s.spatialGrid.QueryRadius(x, y, proj.ExplosionRadius)
	s.applyExplosionDamageToTargets(w, entity, proj, x, y, nearby)
}

// spawnExplosionEffects creates visual and feedback effects for an explosion.
func (s *System) spawnExplosionEffects(x, y float64, color color.RGBA) {
	if s.particleSpawner != nil {
		s.particleSpawner.SpawnBurst(x, y, 0, 30, 6.0, 1.0, 0.5, 0.2, color)
	}
	if s.feedbackProvider != nil {
		s.feedbackProvider.AddScreenShake(3.0)
	}
}

// applyExplosionDamageToTargets processes damage for all entities in the explosion radius.
func (s *System) applyExplosionDamageToTargets(w *engine.World, entity engine.Entity, proj *ProjectileComponent, x, y float64, nearby []engine.Entity) {
	positionType := reflect.TypeOf((*engine.Position)(nil))
	resistanceType := reflect.TypeOf((*ResistanceComponent)(nil))

	for _, target := range nearby {
		if s.shouldSkipExplosionTarget(target, entity, proj.OwnerID) {
			continue
		}

		targetPos := s.getTargetPosition(w, target, positionType)
		if targetPos == nil {
			continue
		}

		if s.applyExplosionDamageToTarget(w, target, targetPos, proj, x, y, resistanceType) {
			s.spawnTargetImpactEffects(targetPos.X, targetPos.Y, proj.Color)
		}
	}
}

// shouldSkipExplosionTarget checks if a target should be excluded from explosion damage.
func (s *System) shouldSkipExplosionTarget(target, projectileEntity engine.Entity, ownerID int) bool {
	return target == projectileEntity || int(target) == ownerID
}

// getTargetPosition retrieves the position component for a target entity.
func (s *System) getTargetPosition(w *engine.World, target engine.Entity, positionType reflect.Type) *engine.Position {
	targetPosComp, ok := w.GetComponent(target, positionType)
	if !ok {
		return nil
	}
	targetPos, ok := targetPosComp.(*engine.Position)
	if !ok {
		return nil
	}
	return targetPos
}

// applyExplosionDamageToTarget calculates and applies explosion damage to a single target.
func (s *System) applyExplosionDamageToTarget(w *engine.World, target engine.Entity, targetPos *engine.Position, proj *ProjectileComponent, x, y float64, resistanceType reflect.Type) bool {
	dist := calculateDistance(targetPos.X, targetPos.Y, x, y)
	if dist > proj.ExplosionRadius {
		return false
	}

	falloff := 1.0 - (dist / proj.ExplosionRadius)
	explosionDamage := proj.Damage * falloff
	finalDamage := s.calculateFinalDamage(w, target, explosionDamage, proj.DamageType, resistanceType)

	s.logger.WithFields(logrus.Fields{
		"target_id":   int(target),
		"damage":      finalDamage,
		"damage_type": DamageTypeNames[proj.DamageType],
		"explosion":   true,
	}).Debug("Explosion hit target")

	return true
}

// calculateFinalDamage applies resistance modifiers to explosion damage.
func (s *System) calculateFinalDamage(w *engine.World, target engine.Entity, damage float64, damageType DamageType, resistanceType reflect.Type) float64 {
	comp, ok := w.GetComponent(target, resistanceType)
	if !ok {
		return damage
	}
	resistance, ok := comp.(*ResistanceComponent)
	if !ok {
		return damage
	}
	return CalculateDamage(damage, damageType, resistance.Resistances)
}

// spawnTargetImpactEffects creates impact particles on a damaged target.
func (s *System) spawnTargetImpactEffects(x, y float64, color color.RGBA) {
	if s.particleSpawner != nil {
		s.particleSpawner.SpawnBurst(x, y, 0, 5, 2.5, 0.4, 0.2, 0.3, color)
	}
}

// calculateDistance computes the Euclidean distance between two points.
func calculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

// Type returns the system type for registration.
func (s *System) Type() string {
	return "ProjectileSystem"
}
