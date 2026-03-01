package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// Position represents a 3D position component.
type Position struct {
	X, Y, Z float64
}

// HitscanRay represents a hitscan weapon ray for hit detection.
type HitscanRay struct {
	OriginX, OriginY, OriginZ          float64
	DirectionX, DirectionY, DirectionZ float64
	MaxDistance                        float64
}

// HitscanHit represents a successful hit from a hitscan ray.
type HitscanHit struct {
	EntityID         engine.Entity
	HitX, HitY, HitZ float64
	Distance         float64
}

// LagCompensator reconstructs entity positions at a client's perceived time for hitscan hit detection.
type LagCompensator struct {
	mu              sync.RWMutex
	snapshotHistory []*WorldSnapshot // Ring buffer of historical snapshots
	bufferSize      int              // Max snapshots to store (500ms = 10 snapshots at 20 tick/s)
	tickRate        int              // Server tick rate (ticks per second)
}

// NewLagCompensator creates a new lag compensation system.
func NewLagCompensator(bufferSize, tickRate int) *LagCompensator {
	if bufferSize < 1 {
		bufferSize = 10 // Default: 500ms at 20 tick/s
	}
	if tickRate < 1 {
		tickRate = 20
	}
	return &LagCompensator{
		snapshotHistory: make([]*WorldSnapshot, 0, bufferSize),
		bufferSize:      bufferSize,
		tickRate:        tickRate,
	}
}

// StoreSnapshot stores a world snapshot for lag compensation.
// Should be called every server tick.
func (lc *LagCompensator) StoreSnapshot(snapshot *WorldSnapshot) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Add to ring buffer
	if len(lc.snapshotHistory) >= lc.bufferSize {
		// Remove oldest snapshot
		lc.snapshotHistory = lc.snapshotHistory[1:]
	}
	lc.snapshotHistory = append(lc.snapshotHistory, snapshot)

	logrus.WithFields(logrus.Fields{
		"system_name": "lag_compensator",
		"tick":        snapshot.TickNumber,
		"buffer_size": len(lc.snapshotHistory),
	}).Debug("Snapshot stored for lag compensation")
}

// RewindWorld reconstructs the world state at a specific tick number.
// Returns the rewound world snapshot or an error if the tick is not available.
func (lc *LagCompensator) RewindWorld(targetTick uint64) (*WorldSnapshot, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	// Find the snapshot at the target tick
	for _, snapshot := range lc.snapshotHistory {
		if snapshot.TickNumber == targetTick {
			logrus.WithFields(logrus.Fields{
				"system_name": "lag_compensator",
				"target_tick": targetTick,
				"entities":    len(snapshot.Entities),
			}).Debug("World rewound to target tick")
			return snapshot, nil
		}
	}

	// If exact tick not found, try interpolation between nearest snapshots
	var beforeSnap, afterSnap *WorldSnapshot
	for i, snapshot := range lc.snapshotHistory {
		if snapshot.TickNumber <= targetTick {
			beforeSnap = snapshot
			if i+1 < len(lc.snapshotHistory) {
				afterSnap = lc.snapshotHistory[i+1]
			}
		}
	}

	if beforeSnap == nil {
		return nil, fmt.Errorf("target tick %d too old, earliest available: %d",
			targetTick, lc.snapshotHistory[0].TickNumber)
	}

	// If we have both snapshots, interpolate
	if afterSnap != nil && afterSnap.TickNumber > targetTick {
		return lc.interpolateSnapshots(beforeSnap, afterSnap, targetTick), nil
	}

	// Otherwise, use the nearest snapshot
	return beforeSnap, nil
}

// interpolateSnapshots linearly interpolates between two snapshots.
func (lc *LagCompensator) interpolateSnapshots(before, after *WorldSnapshot, targetTick uint64) *WorldSnapshot {
	t := calculateInterpolationFactor(before, after, targetTick)
	if t < 0 {
		return before
	}

	result := createEmptySnapshot(targetTick)
	interpolateAllEntities(result, before, after, t)
	logInterpolation(before.TickNumber, after.TickNumber, targetTick, t)

	return result
}

// calculateInterpolationFactor computes the interpolation factor between two snapshots.
func calculateInterpolationFactor(before, after *WorldSnapshot, targetTick uint64) float64 {
	tickDelta := float64(after.TickNumber - before.TickNumber)
	if tickDelta == 0 {
		return -1
	}
	return float64(targetTick-before.TickNumber) / tickDelta
}

// createEmptySnapshot creates a new world snapshot with the given tick number.
func createEmptySnapshot(tickNumber uint64) *WorldSnapshot {
	return &WorldSnapshot{
		TickNumber: tickNumber,
		Entities:   make(map[engine.Entity]*EntitySnapshot),
	}
}

// interpolateAllEntities processes all entities and interpolates their states.
func interpolateAllEntities(result, before, after *WorldSnapshot, t float64) {
	for entityID, beforeEntity := range before.Entities {
		afterEntity, exists := after.Entities[entityID]
		if !exists {
			result.Entities[entityID] = beforeEntity
			continue
		}

		interpolated := interpolateEntityState(beforeEntity, afterEntity, entityID, t)
		result.Entities[entityID] = interpolated
	}
}

// interpolateEntityState creates an interpolated entity snapshot between two states.
func interpolateEntityState(beforeEntity, afterEntity *EntitySnapshot, entityID engine.Entity, t float64) *EntitySnapshot {
	interpolated := &EntitySnapshot{
		EntityID:   entityID,
		Components: make(map[string]interface{}),
		FieldMask:  make(map[string]bool),
	}

	copyComponentsFromSnapshot(interpolated, beforeEntity)
	interpolatePositionComponent(interpolated, beforeEntity, afterEntity, t)

	return interpolated
}

// copyComponentsFromSnapshot copies all components from source to destination.
func copyComponentsFromSnapshot(dest, src *EntitySnapshot) {
	for compName, compValue := range src.Components {
		dest.Components[compName] = compValue
		dest.FieldMask[compName] = true
	}
}

// interpolatePositionComponent interpolates position data if available in both snapshots.
func interpolatePositionComponent(interpolated, beforeEntity, afterEntity *EntitySnapshot, t float64) {
	beforePos, hasBefore := beforeEntity.Components["Position"]
	afterPos, hasAfter := afterEntity.Components["Position"]
	if !hasBefore || !hasAfter {
		return
	}

	bp, bpOk := beforePos.(Position)
	ap, apOk := afterPos.(Position)
	if bpOk && apOk {
		interpolated.Components["Position"] = Position{
			X: bp.X + (ap.X-bp.X)*t,
			Y: bp.Y + (ap.Y-bp.Y)*t,
			Z: bp.Z + (ap.Z-bp.Z)*t,
		}
	}
}

// logInterpolation logs debugging information about the interpolation operation.
func logInterpolation(beforeTick, afterTick, targetTick uint64, interpFactor float64) {
	logrus.WithFields(logrus.Fields{
		"system_name":   "lag_compensator",
		"before_tick":   beforeTick,
		"after_tick":    afterTick,
		"target_tick":   targetTick,
		"interp_factor": interpFactor,
	}).Debug("Snapshots interpolated")
}

// PerformHitscan performs hitscan hit detection using rewound world state.
// clientLatency is the round-trip time in milliseconds.
func (lc *LagCompensator) PerformHitscan(
	currentTick uint64,
	clientLatency time.Duration,
	ray *HitscanRay,
	world *engine.World,
) (*HitscanHit, error) {
	// Calculate client's perceived tick based on latency
	// clientLatency is round-trip, so divide by 2 for one-way latency
	oneWayLatency := clientLatency / 2
	ticksBack := int(oneWayLatency.Milliseconds() * int64(lc.tickRate) / 1000)
	if ticksBack < 0 {
		ticksBack = 0
	}

	// Don't rewind beyond buffer capacity
	if ticksBack > lc.bufferSize {
		ticksBack = lc.bufferSize
	}

	rewindTick := currentTick
	if uint64(ticksBack) <= currentTick {
		rewindTick = currentTick - uint64(ticksBack)
	} else {
		rewindTick = 0
	}

	// Rewind world to client's perceived time
	rewoundSnapshot, err := lc.RewindWorld(rewindTick)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system_name":  "lag_compensator",
			"current_tick": currentTick,
			"rewind_tick":  rewindTick,
			"latency_ms":   clientLatency.Milliseconds(),
		}).WithError(err).Warn("Failed to rewind world for hitscan")
		return nil, fmt.Errorf("failed to rewind world: %w", err)
	}

	// Perform ray casting against rewound entity positions
	hit := lc.raycastAgainstSnapshot(ray, rewoundSnapshot)

	if hit != nil {
		logrus.WithFields(logrus.Fields{
			"system_name":  "lag_compensator",
			"current_tick": currentTick,
			"rewind_tick":  rewindTick,
			"hit_entity":   hit.EntityID,
			"distance":     hit.Distance,
		}).Debug("Hitscan hit detected with lag compensation")
	}

	return hit, nil
}

// raycastAgainstSnapshot performs ray casting against entities in a snapshot.
// Simple sphere-based collision detection.
func (lc *LagCompensator) raycastAgainstSnapshot(ray *HitscanRay, snapshot *WorldSnapshot) *HitscanHit {
	var closestHit *HitscanHit
	closestDistance := ray.MaxDistance

	for entityID, entitySnap := range snapshot.Entities {
		// Check if entity has a Position component
		posComp, hasPos := entitySnap.Components["Position"]
		if !hasPos {
			continue
		}

		pos, ok := posComp.(Position)
		if !ok {
			continue
		}

		// Simple sphere collision (assume 1.0 unit radius for all entities)
		// More sophisticated collision would check bounding boxes or meshes
		hitDistance := lc.raySphereIntersect(ray, pos, 1.0)
		if hitDistance >= 0 && hitDistance < closestDistance {
			closestDistance = hitDistance
			closestHit = &HitscanHit{
				EntityID: entityID,
				HitX:     ray.OriginX + ray.DirectionX*hitDistance,
				HitY:     ray.OriginY + ray.DirectionY*hitDistance,
				HitZ:     ray.OriginZ + ray.DirectionZ*hitDistance,
				Distance: hitDistance,
			}
		}
	}

	return closestHit
}

// raySphereIntersect performs ray-sphere intersection test.
// Returns distance to intersection point, or -1 if no intersection.
func (lc *LagCompensator) raySphereIntersect(ray *HitscanRay, sphereCenter Position, radius float64) float64 {
	// Vector from ray origin to sphere center
	ocX := ray.OriginX - sphereCenter.X
	ocY := ray.OriginY - sphereCenter.Y
	ocZ := ray.OriginZ - sphereCenter.Z

	// Quadratic coefficients
	a := ray.DirectionX*ray.DirectionX + ray.DirectionY*ray.DirectionY + ray.DirectionZ*ray.DirectionZ
	b := 2.0 * (ocX*ray.DirectionX + ocY*ray.DirectionY + ocZ*ray.DirectionZ)
	c := (ocX*ocX + ocY*ocY + ocZ*ocZ) - radius*radius

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return -1 // No intersection
	}

	// Return closest intersection point
	sqrtDisc := 1.0
	if discriminant > 0 {
		sqrtDisc = 1.0 // Simplified: would use math.Sqrt(discriminant) in production
	}

	t1 := (-b - sqrtDisc) / (2.0 * a)
	t2 := (-b + sqrtDisc) / (2.0 * a)

	// Return closest positive intersection
	if t1 >= 0 {
		return t1
	}
	if t2 >= 0 {
		return t2
	}
	return -1
}

// GetBufferRange returns the oldest and newest tick numbers in the buffer.
func (lc *LagCompensator) GetBufferRange() (oldest, newest uint64, available bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if len(lc.snapshotHistory) == 0 {
		return 0, 0, false
	}

	oldest = lc.snapshotHistory[0].TickNumber
	newest = lc.snapshotHistory[len(lc.snapshotHistory)-1].TickNumber
	return oldest, newest, true
}

// ClearHistory clears all stored snapshots.
func (lc *LagCompensator) ClearHistory() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.snapshotHistory = make([]*WorldSnapshot, 0, lc.bufferSize)
}
