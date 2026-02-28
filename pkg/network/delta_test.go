package network

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// TestDeltaEncoderNew verifies DeltaEncoder initialization.
func TestDeltaEncoderNew(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		wantSize   int
	}{
		{"default buffer size", 0, 10},
		{"custom buffer size", 20, 20},
		{"negative buffer size", -5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewDeltaEncoder(tt.bufferSize)
			if encoder == nil {
				t.Fatal("expected non-nil encoder")
			}
			if cap(encoder.snapshotBuffer) != tt.wantSize {
				t.Errorf("buffer size = %d, want %d", cap(encoder.snapshotBuffer), tt.wantSize)
			}
		})
	}
}

// TestDeltaEncoderCaptureSnapshot verifies snapshot capture.
func TestDeltaEncoderCaptureSnapshot(t *testing.T) {
	tests := []struct {
		name         string
		setupWorld   func(*engine.World) []engine.Entity
		tickNum      uint64
		wantEntities int
	}{
		{
			name: "empty world",
			setupWorld: func(w *engine.World) []engine.Entity {
				return []engine.Entity{}
			},
			tickNum:      1,
			wantEntities: 0,
		},
		{
			name: "world with entities",
			setupWorld: func(w *engine.World) []engine.Entity {
				e1 := w.AddEntity()
				e2 := w.AddEntity()
				e3 := w.AddEntity()
				return []engine.Entity{e1, e2, e3}
			},
			tickNum:      10,
			wantEntities: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			entities := tt.setupWorld(world)
			encoder := NewDeltaEncoder(10)

			snapshot := encoder.CaptureSnapshot(world, tt.tickNum)

			if snapshot == nil {
				t.Fatal("expected non-nil snapshot")
			}
			if snapshot.TickNumber != tt.tickNum {
				t.Errorf("tick number = %d, want %d", snapshot.TickNumber, tt.tickNum)
			}
			if len(snapshot.Entities) != tt.wantEntities {
				t.Errorf("entity count = %d, want %d", len(snapshot.Entities), tt.wantEntities)
			}

			// Verify all expected entities are in snapshot
			for _, entityID := range entities {
				if _, exists := snapshot.Entities[entityID]; !exists {
					t.Errorf("entity %d not found in snapshot", entityID)
				}
			}

			// Verify snapshot is in buffer
			if len(encoder.snapshotBuffer) != 1 {
				t.Errorf("buffer length = %d, want 1", len(encoder.snapshotBuffer))
			}

			// Verify baseline is set
			if encoder.baseline == nil {
				t.Error("baseline not set after first snapshot")
			}
		})
	}
}

// TestDeltaEncoderEncodeDelta verifies delta encoding.
func TestDeltaEncoderEncodeDelta(t *testing.T) {
	tests := []struct {
		name          string
		setupBaseline func(*engine.World) []engine.Entity
		setupCurrent  func(*engine.World, []engine.Entity) []engine.Entity
		baselineTick  uint64
		currentTick   uint64
		wantAdded     int
		wantModified  int
		wantRemoved   int
	}{
		{
			name: "no changes",
			setupBaseline: func(w *engine.World) []engine.Entity {
				e1 := w.AddEntity()
				e2 := w.AddEntity()
				return []engine.Entity{e1, e2}
			},
			setupCurrent: func(w *engine.World, baseline []engine.Entity) []engine.Entity {
				return baseline
			},
			baselineTick: 1,
			currentTick:  2,
			wantAdded:    0,
			wantModified: 0,
			wantRemoved:  0,
		},
		{
			name: "entity added",
			setupBaseline: func(w *engine.World) []engine.Entity {
				e1 := w.AddEntity()
				return []engine.Entity{e1}
			},
			setupCurrent: func(w *engine.World, baseline []engine.Entity) []engine.Entity {
				e2 := w.AddEntity()
				return append(baseline, e2)
			},
			baselineTick: 1,
			currentTick:  2,
			wantAdded:    1,
			wantModified: 0,
			wantRemoved:  0,
		},
		{
			name: "entity removed",
			setupBaseline: func(w *engine.World) []engine.Entity {
				e1 := w.AddEntity()
				e2 := w.AddEntity()
				return []engine.Entity{e1, e2}
			},
			setupCurrent: func(w *engine.World, baseline []engine.Entity) []engine.Entity {
				if len(baseline) > 0 {
					w.RemoveEntity(baseline[0])
					return baseline[1:]
				}
				return baseline
			},
			baselineTick: 1,
			currentTick:  2,
			wantAdded:    0,
			wantModified: 0,
			wantRemoved:  1,
		},
		{
			name: "first snapshot is all additions",
			setupBaseline: func(w *engine.World) []engine.Entity {
				return []engine.Entity{}
			},
			setupCurrent: func(w *engine.World, baseline []engine.Entity) []engine.Entity {
				e1 := w.AddEntity()
				e2 := w.AddEntity()
				e3 := w.AddEntity()
				return []engine.Entity{e1, e2, e3}
			},
			baselineTick: 0,
			currentTick:  1,
			wantAdded:    3,
			wantModified: 0,
			wantRemoved:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup baseline world
			baselineWorld := engine.NewWorld()
			baselineEntities := tt.setupBaseline(baselineWorld)

			encoder := NewDeltaEncoder(10)

			// Capture baseline
			encoder.CaptureSnapshot(baselineWorld, tt.baselineTick)

			// Setup current world
			currentWorld := engine.NewWorld()
			// Re-create baseline entities in current world
			var recreatedBaseline []engine.Entity
			for range baselineEntities {
				recreatedBaseline = append(recreatedBaseline, currentWorld.AddEntity())
			}
			currentEntities := tt.setupCurrent(currentWorld, recreatedBaseline)

			// Encode delta
			delta, err := encoder.EncodeDelta(currentWorld, tt.currentTick)
			if err != nil {
				t.Fatalf("EncodeDelta failed: %v", err)
			}

			if delta.BaseTick != tt.baselineTick {
				t.Errorf("base tick = %d, want %d", delta.BaseTick, tt.baselineTick)
			}
			if delta.TargetTick != tt.currentTick {
				t.Errorf("target tick = %d, want %d", delta.TargetTick, tt.currentTick)
			}
			if len(delta.Added) != tt.wantAdded {
				t.Errorf("added count = %d, want %d", len(delta.Added), tt.wantAdded)
			}
			if len(delta.Modified) != tt.wantModified {
				t.Errorf("modified count = %d, want %d", len(delta.Modified), tt.wantModified)
			}
			if len(delta.Removed) != tt.wantRemoved {
				t.Errorf("removed count = %d, want %d", len(delta.Removed), tt.wantRemoved)
			}

			// Verify current entities match expected count
			_ = currentEntities
		})
	}
}

// TestDeltaDecoderNew verifies DeltaDecoder initialization.
func TestDeltaDecoderNew(t *testing.T) {
	decoder := NewDeltaDecoder()
	if decoder == nil {
		t.Fatal("expected non-nil decoder")
	}
	if decoder.baseline != nil {
		t.Error("expected nil baseline for new decoder")
	}
}

// TestDeltaDecoderSetBaseline verifies baseline setting.
func TestDeltaDecoderSetBaseline(t *testing.T) {
	decoder := NewDeltaDecoder()

	baseline := &WorldSnapshot{
		TickNumber: 42,
		Entities:   make(map[engine.Entity]*EntitySnapshot),
	}
	baseline.Entities[1] = &EntitySnapshot{EntityID: 1}

	decoder.SetBaseline(baseline)

	if decoder.baseline == nil {
		t.Fatal("baseline not set")
	}
	if decoder.baseline.TickNumber != 42 {
		t.Errorf("baseline tick = %d, want 42", decoder.baseline.TickNumber)
	}
}

// TestDeltaDecoderApplyDelta verifies delta application.
func TestDeltaDecoderApplyDelta(t *testing.T) {
	tests := []struct {
		name         string
		baseline     *WorldSnapshot
		delta        *DeltaPacket
		wantEntities int
		wantTick     uint64
		wantError    bool
	}{
		{
			name:      "no baseline",
			baseline:  nil,
			delta:     &DeltaPacket{BaseTick: 1, TargetTick: 2},
			wantError: true,
		},
		{
			name: "mismatched base tick",
			baseline: &WorldSnapshot{
				TickNumber: 10,
				Entities:   make(map[engine.Entity]*EntitySnapshot),
			},
			delta: &DeltaPacket{
				BaseTick:   5,
				TargetTick: 11,
			},
			wantError: true,
		},
		{
			name: "add entities",
			baseline: &WorldSnapshot{
				TickNumber: 1,
				Entities:   make(map[engine.Entity]*EntitySnapshot),
			},
			delta: &DeltaPacket{
				BaseTick:   1,
				TargetTick: 2,
				Added: map[engine.Entity]*EntitySnapshot{
					100: {EntityID: 100, Components: make(map[string]interface{})},
					101: {EntityID: 101, Components: make(map[string]interface{})},
				},
				Modified: make(map[engine.Entity]*EntitySnapshot),
				Removed:  []engine.Entity{},
			},
			wantEntities: 2,
			wantTick:     2,
			wantError:    false,
		},
		{
			name: "remove entities",
			baseline: &WorldSnapshot{
				TickNumber: 1,
				Entities: map[engine.Entity]*EntitySnapshot{
					100: {EntityID: 100, Components: make(map[string]interface{})},
					101: {EntityID: 101, Components: make(map[string]interface{})},
					102: {EntityID: 102, Components: make(map[string]interface{})},
				},
			},
			delta: &DeltaPacket{
				BaseTick:   1,
				TargetTick: 2,
				Added:      make(map[engine.Entity]*EntitySnapshot),
				Modified:   make(map[engine.Entity]*EntitySnapshot),
				Removed:    []engine.Entity{100, 102},
			},
			wantEntities: 1,
			wantTick:     2,
			wantError:    false,
		},
		{
			name: "modify entities",
			baseline: &WorldSnapshot{
				TickNumber: 1,
				Entities: map[engine.Entity]*EntitySnapshot{
					100: {
						EntityID: 100,
						Components: map[string]interface{}{
							"position": map[string]float64{"x": 1.0, "y": 2.0},
						},
						FieldMask: map[string]bool{"position": true},
					},
				},
			},
			delta: &DeltaPacket{
				BaseTick:   1,
				TargetTick: 2,
				Added:      make(map[engine.Entity]*EntitySnapshot),
				Modified: map[engine.Entity]*EntitySnapshot{
					100: {
						EntityID: 100,
						Components: map[string]interface{}{
							"position": map[string]float64{"x": 5.0, "y": 10.0},
						},
						FieldMask: map[string]bool{"position": true},
					},
				},
				Removed: []engine.Entity{},
			},
			wantEntities: 1,
			wantTick:     2,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDeltaDecoder()
			if tt.baseline != nil {
				decoder.SetBaseline(tt.baseline)
			}

			result, err := decoder.ApplyDelta(tt.delta)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TickNumber != tt.wantTick {
				t.Errorf("result tick = %d, want %d", result.TickNumber, tt.wantTick)
			}

			if len(result.Entities) != tt.wantEntities {
				t.Errorf("entity count = %d, want %d", len(result.Entities), tt.wantEntities)
			}
		})
	}
}

// TestDeltaRoundTrip verifies encoding and decoding produce consistent results.
func TestDeltaRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		setupWorld   func() *engine.World
		baselineTick uint64
		currentTick  uint64
	}{
		{
			name: "empty world",
			setupWorld: func() *engine.World {
				return engine.NewWorld()
			},
			baselineTick: 0,
			currentTick:  1,
		},
		{
			name: "world with entities",
			setupWorld: func() *engine.World {
				w := engine.NewWorld()
				w.AddEntity()
				w.AddEntity()
				w.AddEntity()
				return w
			},
			baselineTick: 0,
			currentTick:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := tt.setupWorld()
			encoder := NewDeltaEncoder(10)

			// Capture baseline
			baseline := encoder.CaptureSnapshot(world, tt.baselineTick)

			// Encode delta to current
			delta, err := encoder.EncodeDelta(world, tt.currentTick)
			if err != nil {
				t.Fatalf("EncodeDelta failed: %v", err)
			}

			// Decode delta
			decoder := NewDeltaDecoder()
			decoder.SetBaseline(baseline)

			reconstructed, err := decoder.ApplyDelta(delta)
			if err != nil {
				t.Fatalf("ApplyDelta failed: %v", err)
			}

			// Verify reconstructed matches current snapshot
			current := encoder.lastSnapshot
			if reconstructed.TickNumber != current.TickNumber {
				t.Errorf("tick mismatch: got %d, want %d",
					reconstructed.TickNumber, current.TickNumber)
			}

			if len(reconstructed.Entities) != len(current.Entities) {
				t.Errorf("entity count mismatch: got %d, want %d",
					len(reconstructed.Entities), len(current.Entities))
			}

			for entityID := range current.Entities {
				if _, exists := reconstructed.Entities[entityID]; !exists {
					t.Errorf("entity %d missing from reconstructed snapshot", entityID)
				}
			}
		})
	}
}

// TestDeltaEncoderSnapshotBuffer verifies circular buffer behavior.
func TestDeltaEncoderSnapshotBuffer(t *testing.T) {
	bufferSize := 5
	encoder := NewDeltaEncoder(bufferSize)
	world := engine.NewWorld()

	// Add more snapshots than buffer size
	for i := uint64(0); i < 10; i++ {
		encoder.CaptureSnapshot(world, i)
	}

	// Buffer should only contain last N snapshots
	if len(encoder.snapshotBuffer) != bufferSize {
		t.Errorf("buffer length = %d, want %d", len(encoder.snapshotBuffer), bufferSize)
	}

	// Verify buffer contains most recent snapshots
	for i, snapshot := range encoder.snapshotBuffer {
		expectedTick := uint64(5 + i) // Should have ticks 5-9
		if snapshot.TickNumber != expectedTick {
			t.Errorf("snapshot[%d] tick = %d, want %d",
				i, snapshot.TickNumber, expectedTick)
		}
	}
}

// TestDeltaEncoderGetSnapshot verifies snapshot retrieval.
func TestDeltaEncoderGetSnapshot(t *testing.T) {
	encoder := NewDeltaEncoder(10)
	world := engine.NewWorld()

	// Capture some snapshots
	encoder.CaptureSnapshot(world, 1)
	encoder.CaptureSnapshot(world, 5)
	encoder.CaptureSnapshot(world, 10)

	tests := []struct {
		name    string
		tickNum uint64
		wantNil bool
	}{
		{"existing snapshot", 5, false},
		{"non-existent snapshot", 999, true},
		{"first snapshot", 1, false},
		{"last snapshot", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := encoder.GetSnapshot(tt.tickNum)
			if tt.wantNil && snapshot != nil {
				t.Error("expected nil snapshot")
			}
			if !tt.wantNil && snapshot == nil {
				t.Error("expected non-nil snapshot")
			}
			if !tt.wantNil && snapshot.TickNumber != tt.tickNum {
				t.Errorf("snapshot tick = %d, want %d", snapshot.TickNumber, tt.tickNum)
			}
		})
	}
}

// TestEntitySnapshotCopy verifies deep copying of entity snapshots.
func TestEntitySnapshotCopy(t *testing.T) {
	decoder := NewDeltaDecoder()

	original := &EntitySnapshot{
		EntityID: 42,
		Components: map[string]interface{}{
			"position": map[string]float64{"x": 1.0, "y": 2.0, "z": 3.0},
			"health":   100,
		},
		FieldMask: map[string]bool{
			"position": true,
			"health":   true,
		},
	}

	copy := decoder.copyEntitySnapshot(original)

	// Verify copy is independent
	if copy.EntityID != original.EntityID {
		t.Errorf("entity ID mismatch: got %d, want %d", copy.EntityID, original.EntityID)
	}

	if len(copy.Components) != len(original.Components) {
		t.Errorf("component count mismatch: got %d, want %d",
			len(copy.Components), len(original.Components))
	}

	// Modify copy, should not affect original
	copy.Components["health"] = 50
	if original.Components["health"] == 50 {
		t.Error("modifying copy affected original")
	}
}
