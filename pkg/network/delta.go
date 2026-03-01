package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// EntitySnapshot represents the complete state of an entity at a specific tick.
type EntitySnapshot struct {
	EntityID   engine.Entity          `json:"entity_id"`
	Components map[string]interface{} `json:"components"`
	FieldMask  map[string]bool        `json:"field_mask"` // Presence bitmask for optional fields
}

// WorldSnapshot represents the complete state of the world at a specific tick.
type WorldSnapshot struct {
	TickNumber uint64                            `json:"tick_number"`
	Entities   map[engine.Entity]*EntitySnapshot `json:"entities"`
}

// DeltaPacket represents the difference between two world states.
type DeltaPacket struct {
	BaseTick   uint64                            `json:"base_tick"`   // Baseline tick this delta is relative to
	TargetTick uint64                            `json:"target_tick"` // Target tick this delta produces
	Added      map[engine.Entity]*EntitySnapshot `json:"added"`       // Newly created entities
	Modified   map[engine.Entity]*EntitySnapshot `json:"modified"`    // Modified entities (only changed fields)
	Removed    []engine.Entity                   `json:"removed"`     // Removed entities
}

// DeltaEncoder compresses world state into delta packets.
type DeltaEncoder struct {
	mu             sync.RWMutex
	baseline       *WorldSnapshot
	lastSnapshot   *WorldSnapshot
	snapshotBuffer []*WorldSnapshot // Ring buffer for lag compensation
	bufferSize     int
}

// NewDeltaEncoder creates a new delta encoder with a snapshot buffer.
func NewDeltaEncoder(bufferSize int) *DeltaEncoder {
	if bufferSize < 1 {
		bufferSize = 10 // Default: 500ms at 20 tick/s
	}
	return &DeltaEncoder{
		snapshotBuffer: make([]*WorldSnapshot, 0, bufferSize),
		bufferSize:     bufferSize,
	}
}

// CaptureSnapshot creates a complete snapshot of the world state.
func (e *DeltaEncoder) CaptureSnapshot(world *engine.World, tickNum uint64) *WorldSnapshot {
	e.mu.Lock()
	defer e.mu.Unlock()

	snapshot := &WorldSnapshot{
		TickNumber: tickNum,
		Entities:   make(map[engine.Entity]*EntitySnapshot),
	}

	// Query all entities (we don't filter by component types here)
	// In a real implementation, we'd iterate over all entities in the world
	// For now, we'll work with what the Query method can provide
	allEntities := world.Query() // Query with no component types returns all

	for _, entityID := range allEntities {
		entitySnap := e.captureEntity(world, entityID)
		if entitySnap != nil {
			snapshot.Entities[entityID] = entitySnap
		}
	}

	// Store in circular buffer
	if len(e.snapshotBuffer) >= e.bufferSize {
		// Remove oldest snapshot
		e.snapshotBuffer = e.snapshotBuffer[1:]
	}
	e.snapshotBuffer = append(e.snapshotBuffer, snapshot)

	// Update baseline if this is the first snapshot
	if e.baseline == nil {
		e.baseline = snapshot
	}

	e.lastSnapshot = snapshot
	return snapshot
}

// captureEntity creates a snapshot of a single entity.
func (e *DeltaEncoder) captureEntity(world *engine.World, entityID engine.Entity) *EntitySnapshot {
	snapshot := &EntitySnapshot{
		EntityID:   entityID,
		Components: make(map[string]interface{}),
		FieldMask:  make(map[string]bool),
	}

	// We need to iterate through all component types
	// Since the World doesn't expose all components directly, we'll use a workaround
	// In a production system, we'd extend World to expose component iteration
	// For now, we'll capture known component types via reflection

	// This is a simplified approach - in production, we'd have a registry
	// of all component types that can be serialized
	return snapshot
}

// EncodeDelta creates a delta packet from baseline to current state.
func (e *DeltaEncoder) EncodeDelta(world *engine.World, tickNum uint64) (*DeltaPacket, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	currentSnapshot := e.captureSnapshotInternal(world, tickNum)

	if e.baseline == nil {
		return e.createInitialDelta(currentSnapshot, tickNum), nil
	}

	delta := e.computeDelta(currentSnapshot, tickNum)
	e.updateEncoderState(currentSnapshot)
	e.logDelta(delta, tickNum)

	return delta, nil
}

// createInitialDelta creates the first delta packet when no baseline exists.
func (e *DeltaEncoder) createInitialDelta(snapshot *WorldSnapshot, tickNum uint64) *DeltaPacket {
	e.baseline = snapshot
	e.lastSnapshot = snapshot

	return &DeltaPacket{
		BaseTick:   0,
		TargetTick: tickNum,
		Added:      snapshot.Entities,
		Modified:   make(map[engine.Entity]*EntitySnapshot),
		Removed:    []engine.Entity{},
	}
}

// computeDelta calculates the difference between baseline and current snapshot.
func (e *DeltaEncoder) computeDelta(currentSnapshot *WorldSnapshot, tickNum uint64) *DeltaPacket {
	delta := &DeltaPacket{
		BaseTick:   e.baseline.TickNumber,
		TargetTick: tickNum,
		Added:      make(map[engine.Entity]*EntitySnapshot),
		Modified:   make(map[engine.Entity]*EntitySnapshot),
		Removed:    []engine.Entity{},
	}

	e.findAddedAndModifiedEntities(delta, currentSnapshot)
	e.findRemovedEntities(delta, currentSnapshot)

	return delta
}

// findAddedAndModifiedEntities identifies new and changed entities.
func (e *DeltaEncoder) findAddedAndModifiedEntities(delta *DeltaPacket, currentSnapshot *WorldSnapshot) {
	for entityID, currentEntity := range currentSnapshot.Entities {
		baselineEntity, existed := e.baseline.Entities[entityID]

		if !existed {
			delta.Added[entityID] = currentEntity
		} else {
			modifiedEntity := e.computeEntityDiff(baselineEntity, currentEntity)
			if modifiedEntity != nil {
				delta.Modified[entityID] = modifiedEntity
			}
		}
	}
}

// findRemovedEntities identifies entities removed since baseline.
func (e *DeltaEncoder) findRemovedEntities(delta *DeltaPacket, currentSnapshot *WorldSnapshot) {
	for entityID := range e.baseline.Entities {
		if _, exists := currentSnapshot.Entities[entityID]; !exists {
			delta.Removed = append(delta.Removed, entityID)
		}
	}
}

// updateEncoderState updates snapshot buffer and last snapshot.
func (e *DeltaEncoder) updateEncoderState(currentSnapshot *WorldSnapshot) {
	e.lastSnapshot = currentSnapshot

	if len(e.snapshotBuffer) >= e.bufferSize {
		e.snapshotBuffer = e.snapshotBuffer[1:]
	}
	e.snapshotBuffer = append(e.snapshotBuffer, currentSnapshot)
}

// logDelta logs delta encoding statistics.
func (e *DeltaEncoder) logDelta(delta *DeltaPacket, tickNum uint64) {
	logrus.WithFields(logrus.Fields{
		"system_name": "delta_encoder",
		"tick":        tickNum,
		"added":       len(delta.Added),
		"modified":    len(delta.Modified),
		"removed":     len(delta.Removed),
	}).Debug("Delta encoded")
}

// captureSnapshotInternal creates a snapshot without locking (caller must hold lock).
func (e *DeltaEncoder) captureSnapshotInternal(world *engine.World, tickNum uint64) *WorldSnapshot {
	snapshot := &WorldSnapshot{
		TickNumber: tickNum,
		Entities:   make(map[engine.Entity]*EntitySnapshot),
	}

	// For simplicity, we'll use an empty query to get all entities
	allEntities := world.Query()

	for _, entityID := range allEntities {
		snapshot.Entities[entityID] = &EntitySnapshot{
			EntityID:   entityID,
			Components: make(map[string]interface{}),
			FieldMask:  make(map[string]bool),
		}
	}

	return snapshot
}

// computeEntityDiff computes the difference between two entity snapshots.
// Returns nil if entities are identical, otherwise returns a snapshot with only changed fields.
func (e *DeltaEncoder) computeEntityDiff(baseline, current *EntitySnapshot) *EntitySnapshot {
	diff := &EntitySnapshot{
		EntityID:   current.EntityID,
		Components: make(map[string]interface{}),
		FieldMask:  make(map[string]bool),
	}

	hasChanges := false

	// Compare components
	for compName, currentComp := range current.Components {
		baselineComp, existed := baseline.Components[compName]

		if !existed {
			// New component added
			diff.Components[compName] = currentComp
			diff.FieldMask[compName] = true
			hasChanges = true
			continue
		}

		// Component existed, check if modified
		if !reflect.DeepEqual(baselineComp, currentComp) {
			diff.Components[compName] = currentComp
			diff.FieldMask[compName] = true
			hasChanges = true
		}
	}

	// Check for removed components
	for compName := range baseline.Components {
		if _, exists := current.Components[compName]; !exists {
			diff.FieldMask[compName] = false // Mark as removed
			hasChanges = true
		}
	}

	if !hasChanges {
		return nil
	}

	return diff
}

// GetSnapshot retrieves a snapshot from the buffer by tick number.
func (e *DeltaEncoder) GetSnapshot(tickNum uint64) *WorldSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, snapshot := range e.snapshotBuffer {
		if snapshot.TickNumber == tickNum {
			return snapshot
		}
	}
	return nil
}

// SetBaseline explicitly sets a baseline snapshot.
func (e *DeltaEncoder) SetBaseline(snapshot *WorldSnapshot) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.baseline = snapshot
}

// DeltaDecoder reconstructs world state from delta packets.
type DeltaDecoder struct {
	mu       sync.RWMutex
	baseline *WorldSnapshot
}

// NewDeltaDecoder creates a new delta decoder.
func NewDeltaDecoder() *DeltaDecoder {
	return &DeltaDecoder{}
}

// SetBaseline sets the baseline snapshot for delta decoding.
func (d *DeltaDecoder) SetBaseline(baseline *WorldSnapshot) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.baseline = baseline
}

// ApplyDelta reconstructs a full world snapshot by applying a delta to the baseline.
func (d *DeltaDecoder) ApplyDelta(delta *DeltaPacket) (*WorldSnapshot, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.baseline == nil {
		return nil, fmt.Errorf("no baseline snapshot available")
	}

	if delta.BaseTick != d.baseline.TickNumber {
		return nil, fmt.Errorf("delta base tick %d does not match baseline %d",
			delta.BaseTick, d.baseline.TickNumber)
	}

	// Create new snapshot starting from baseline
	result := &WorldSnapshot{
		TickNumber: delta.TargetTick,
		Entities:   make(map[engine.Entity]*EntitySnapshot),
	}

	// Copy baseline entities
	for entityID, entity := range d.baseline.Entities {
		result.Entities[entityID] = d.copyEntitySnapshot(entity)
	}

	// Apply removals
	for _, entityID := range delta.Removed {
		delete(result.Entities, entityID)
	}

	// Apply additions
	for entityID, entity := range delta.Added {
		result.Entities[entityID] = d.copyEntitySnapshot(entity)
	}

	// Apply modifications
	for entityID, modifiedEntity := range delta.Modified {
		existing, exists := result.Entities[entityID]
		if !exists {
			// Entity doesn't exist in baseline, treat as addition
			result.Entities[entityID] = d.copyEntitySnapshot(modifiedEntity)
			continue
		}

		// Merge modifications into existing entity
		d.mergeEntitySnapshot(existing, modifiedEntity)
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "delta_decoder",
		"base_tick":   delta.BaseTick,
		"target_tick": delta.TargetTick,
		"entities":    len(result.Entities),
	}).Debug("Delta applied")

	return result, nil
}

// copyEntitySnapshot creates a deep copy of an entity snapshot.
func (d *DeltaDecoder) copyEntitySnapshot(src *EntitySnapshot) *EntitySnapshot {
	dst := &EntitySnapshot{
		EntityID:   src.EntityID,
		Components: make(map[string]interface{}),
		FieldMask:  make(map[string]bool),
	}

	for k, v := range src.Components {
		dst.Components[k] = d.deepCopyValue(v)
	}

	for k, v := range src.FieldMask {
		dst.FieldMask[k] = v
	}

	return dst
}

// deepCopyValue creates a deep copy of a value using gob encoding.
func (d *DeltaDecoder) deepCopyValue(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(src); err != nil {
		return src // Fallback to shallow copy on error
	}

	var dst interface{}
	if err := dec.Decode(&dst); err != nil {
		return src // Fallback to shallow copy on error
	}

	return dst
}

// mergeEntitySnapshot merges modifications from src into dst.
func (d *DeltaDecoder) mergeEntitySnapshot(dst, src *EntitySnapshot) {
	// Apply component updates
	for compName, compValue := range src.Components {
		if present, exists := src.FieldMask[compName]; exists && present {
			dst.Components[compName] = d.deepCopyValue(compValue)
			dst.FieldMask[compName] = true
		} else if exists && !present {
			// Component marked for removal
			delete(dst.Components, compName)
			delete(dst.FieldMask, compName)
		}
	}
}
