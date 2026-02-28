package secret

import (
	"math"
	"testing"
)

func TestNewSecretWall(t *testing.T) {
	tests := []struct {
		name string
		x, y int
		dir  Direction
	}{
		{"north wall", 5, 10, DirNorth},
		{"south wall", 3, 7, DirSouth},
		{"east wall", 8, 2, DirEast},
		{"west wall", 1, 1, DirWest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := NewSecretWall(tt.x, tt.y, tt.dir)
			if sw == nil {
				t.Fatal("NewSecretWall returned nil")
			}
			if sw.X != tt.x || sw.Y != tt.y {
				t.Errorf("position = (%d, %d), want (%d, %d)", sw.X, sw.Y, tt.x, tt.y)
			}
			if sw.Direction != tt.dir {
				t.Errorf("direction = %d, want %d", sw.Direction, tt.dir)
			}
			if sw.State != StateIdle {
				t.Errorf("state = %d, want %d (StateIdle)", sw.State, StateIdle)
			}
			if sw.Progress != 0.0 {
				t.Errorf("progress = %f, want 0.0", sw.Progress)
			}
		})
	}
}

func TestSecretWall_Trigger(t *testing.T) {
	tests := []struct {
		name          string
		initialState  int
		entityID      string
		wantTriggered bool
		wantState     int
	}{
		{"trigger idle wall", StateIdle, "player1", true, StateAnimating},
		{"trigger animating wall", StateAnimating, "player1", false, StateAnimating},
		{"trigger open wall", StateOpen, "player1", false, StateOpen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := NewSecretWall(5, 5, DirNorth)
			sw.State = tt.initialState

			triggered := sw.Trigger(tt.entityID)
			if triggered != tt.wantTriggered {
				t.Errorf("Trigger() = %v, want %v", triggered, tt.wantTriggered)
			}
			if sw.State != tt.wantState {
				t.Errorf("state after trigger = %d, want %d", sw.State, tt.wantState)
			}
			if triggered && sw.DiscoveredBy != tt.entityID {
				t.Errorf("DiscoveredBy = %q, want %q", sw.DiscoveredBy, tt.entityID)
			}
		})
	}
}

func TestSecretWall_Update(t *testing.T) {
	tests := []struct {
		name            string
		initialState    int
		initialProgress float64
		deltaTime       float64
		wantCompleted   bool
		wantProgress    float64
		wantState       int
	}{
		{"idle wall no update", StateIdle, 0.0, 0.1, false, 0.0, StateIdle},
		{"animate partial", StateAnimating, 0.0, 0.5, false, 0.5, StateAnimating},
		{"animate complete", StateAnimating, 0.9, 0.15, true, 1.0, StateOpen},
		{"animate exact 1.0", StateAnimating, 0.0, 1.0, true, 1.0, StateOpen},
		{"open wall no update", StateOpen, 1.0, 0.1, false, 1.0, StateOpen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := NewSecretWall(5, 5, DirNorth)
			sw.State = tt.initialState
			sw.Progress = tt.initialProgress

			completed := sw.Update(tt.deltaTime)
			if completed != tt.wantCompleted {
				t.Errorf("Update() = %v, want %v", completed, tt.wantCompleted)
			}
			if math.Abs(sw.Progress-tt.wantProgress) > 0.01 {
				t.Errorf("progress = %f, want %f", sw.Progress, tt.wantProgress)
			}
			if sw.State != tt.wantState {
				t.Errorf("state = %d, want %d", sw.State, tt.wantState)
			}
		})
	}
}

func TestSecretWall_UpdateMultipleFrames(t *testing.T) {
	sw := NewSecretWall(5, 5, DirNorth)
	sw.Trigger("player1")

	// Simulate 16 frames at 60fps (approx 1 second)
	deltaTime := 1.0 / 16.0
	completedCount := 0

	for i := 0; i < 16; i++ {
		if sw.Update(deltaTime) {
			completedCount++
		}
	}

	if completedCount != 1 {
		t.Errorf("completed %d times, want 1", completedCount)
	}
	if !sw.IsOpen() {
		t.Error("wall should be open after 16 frames")
	}
	if sw.Progress != 1.0 {
		t.Errorf("progress = %f, want 1.0", sw.Progress)
	}
}

func TestSecretWall_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		dir      Direction
		progress float64
		wantX    float64
		wantY    float64
	}{
		{"north half open", DirNorth, 0.5, 0.0, -0.5},
		{"north fully open", DirNorth, 1.0, 0.0, -1.0},
		{"south half open", DirSouth, 0.5, 0.0, 0.5},
		{"south fully open", DirSouth, 1.0, 0.0, 1.0},
		{"east half open", DirEast, 0.5, 0.5, 0.0},
		{"east fully open", DirEast, 1.0, 1.0, 0.0},
		{"west half open", DirWest, 0.5, -0.5, 0.0},
		{"west fully open", DirWest, 1.0, -1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := NewSecretWall(5, 5, tt.dir)
			sw.State = StateAnimating
			sw.Progress = tt.progress

			x, y := sw.GetOffset()
			if math.Abs(x-tt.wantX) > 0.01 || math.Abs(y-tt.wantY) > 0.01 {
				t.Errorf("GetOffset() = (%f, %f), want (%f, %f)", x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestSecretWall_GetOffsetIdle(t *testing.T) {
	sw := NewSecretWall(5, 5, DirNorth)
	x, y := sw.GetOffset()
	if x != 0.0 || y != 0.0 {
		t.Errorf("idle wall offset = (%f, %f), want (0.0, 0.0)", x, y)
	}
}

func TestSecretWall_IsOpen(t *testing.T) {
	tests := []struct {
		state    int
		wantOpen bool
	}{
		{StateIdle, false},
		{StateAnimating, false},
		{StateOpen, true},
	}

	for _, tt := range tests {
		sw := NewSecretWall(5, 5, DirNorth)
		sw.State = tt.state
		if sw.IsOpen() != tt.wantOpen {
			t.Errorf("IsOpen() with state %d = %v, want %v", tt.state, sw.IsOpen(), tt.wantOpen)
		}
	}
}

func TestSecretWall_IsAnimating(t *testing.T) {
	tests := []struct {
		state         int
		wantAnimating bool
	}{
		{StateIdle, false},
		{StateAnimating, true},
		{StateOpen, false},
	}

	for _, tt := range tests {
		sw := NewSecretWall(5, 5, DirNorth)
		sw.State = tt.state
		if sw.IsAnimating() != tt.wantAnimating {
			t.Errorf("IsAnimating() with state %d = %v, want %v", tt.state, sw.IsAnimating(), tt.wantAnimating)
		}
	}
}

func TestNewManager(t *testing.T) {
	m := NewManager(64)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.width != 64 {
		t.Errorf("width = %d, want 64", m.width)
	}
	if m.GetTotalCount() != 0 {
		t.Errorf("initial count = %d, want 0", m.GetTotalCount())
	}
}

func TestManager_AddAndGet(t *testing.T) {
	m := NewManager(64)
	m.Add(10, 20, DirNorth)
	m.Add(15, 25, DirSouth)

	sw1 := m.Get(10, 20)
	if sw1 == nil {
		t.Fatal("Get(10, 20) returned nil")
	}
	if sw1.X != 10 || sw1.Y != 20 || sw1.Direction != DirNorth {
		t.Errorf("secret at (10, 20) has wrong properties")
	}

	sw2 := m.Get(15, 25)
	if sw2 == nil {
		t.Fatal("Get(15, 25) returned nil")
	}
	if sw2.Direction != DirSouth {
		t.Errorf("secret at (15, 25) direction = %d, want %d", sw2.Direction, DirSouth)
	}

	sw3 := m.Get(99, 99)
	if sw3 != nil {
		t.Error("Get(99, 99) should return nil for non-existent secret")
	}
}

func TestManager_TriggerAt(t *testing.T) {
	m := NewManager(64)
	m.Add(10, 20, DirNorth)

	// Trigger existing secret
	if !m.TriggerAt(10, 20, "player1") {
		t.Error("TriggerAt(10, 20) should return true")
	}

	sw := m.Get(10, 20)
	if sw.State != StateAnimating {
		t.Errorf("triggered secret state = %d, want %d", sw.State, StateAnimating)
	}

	// Try to trigger non-existent secret
	if m.TriggerAt(99, 99, "player1") {
		t.Error("TriggerAt(99, 99) should return false for non-existent secret")
	}

	// Try to trigger already animating secret
	if m.TriggerAt(10, 20, "player2") {
		t.Error("TriggerAt on animating secret should return false")
	}
}

func TestManager_Update(t *testing.T) {
	m := NewManager(64)
	m.Add(10, 20, DirNorth)
	m.Add(15, 25, DirSouth)
	m.Add(20, 30, DirEast)

	// Trigger two secrets
	m.TriggerAt(10, 20, "player1")
	m.TriggerAt(15, 25, "player1")

	// Update partially
	completed := m.Update(0.5)
	if completed != 0 {
		t.Errorf("partial update completed = %d, want 0", completed)
	}

	// Update to completion
	completed = m.Update(0.6)
	if completed != 2 {
		t.Errorf("complete update completed = %d, want 2", completed)
	}

	// Verify secrets are open
	if !m.Get(10, 20).IsOpen() {
		t.Error("secret (10, 20) should be open")
	}
	if !m.Get(15, 25).IsOpen() {
		t.Error("secret (15, 25) should be open")
	}
	if m.Get(20, 30).IsOpen() {
		t.Error("secret (20, 30) should not be open")
	}
}

func TestManager_GetAll(t *testing.T) {
	m := NewManager(64)
	m.Add(10, 20, DirNorth)
	m.Add(15, 25, DirSouth)
	m.Add(20, 30, DirEast)

	all := m.GetAll()
	if len(all) != 3 {
		t.Errorf("GetAll() returned %d secrets, want 3", len(all))
	}

	// Verify all secrets are present (order doesn't matter)
	found := make(map[int]bool)
	for _, sw := range all {
		key := sw.Y*64 + sw.X
		found[key] = true
	}

	expectedKeys := []int{20*64 + 10, 25*64 + 15, 30*64 + 20}
	for _, key := range expectedKeys {
		if !found[key] {
			t.Errorf("secret with key %d not found in GetAll()", key)
		}
	}
}

func TestManager_GetDiscoveredCount(t *testing.T) {
	m := NewManager(64)
	m.Add(10, 20, DirNorth)
	m.Add(15, 25, DirSouth)
	m.Add(20, 30, DirEast)

	if m.GetDiscoveredCount() != 0 {
		t.Errorf("initial discovered count = %d, want 0", m.GetDiscoveredCount())
	}

	// Trigger one
	m.TriggerAt(10, 20, "player1")
	if m.GetDiscoveredCount() != 1 {
		t.Errorf("discovered count after 1 trigger = %d, want 1", m.GetDiscoveredCount())
	}

	// Trigger another
	m.TriggerAt(15, 25, "player1")
	if m.GetDiscoveredCount() != 2 {
		t.Errorf("discovered count after 2 triggers = %d, want 2", m.GetDiscoveredCount())
	}

	// Complete animations
	m.Update(1.0)
	if m.GetDiscoveredCount() != 2 {
		t.Errorf("discovered count after completion = %d, want 2", m.GetDiscoveredCount())
	}
}

func TestManager_GetTotalCount(t *testing.T) {
	m := NewManager(64)
	if m.GetTotalCount() != 0 {
		t.Errorf("empty manager total = %d, want 0", m.GetTotalCount())
	}

	m.Add(10, 20, DirNorth)
	if m.GetTotalCount() != 1 {
		t.Errorf("total after 1 add = %d, want 1", m.GetTotalCount())
	}

	m.Add(15, 25, DirSouth)
	m.Add(20, 30, DirEast)
	if m.GetTotalCount() != 3 {
		t.Errorf("total after 3 adds = %d, want 3", m.GetTotalCount())
	}
}

func TestEaseInOut(t *testing.T) {
	tests := []struct {
		t    float64
		want float64
	}{
		{0.0, 0.0},
		{0.25, 0.125},
		{0.5, 0.5},
		{0.75, 0.875},
		{1.0, 1.0},
	}

	for _, tt := range tests {
		got := EaseInOut(tt.t)
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("EaseInOut(%f) = %f, want %f", tt.t, got, tt.want)
		}
	}
}

func TestSmoothStep(t *testing.T) {
	tests := []struct {
		t    float64
		want float64
	}{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{-0.5, 0.0}, // Clamped to 0
		{1.5, 1.0},  // Clamped to 1
	}

	for _, tt := range tests {
		got := SmoothStep(tt.t)
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("SmoothStep(%f) = %f, want %f", tt.t, got, tt.want)
		}
	}
}

func TestManager_MultipleUpdates(t *testing.T) {
	m := NewManager(64)
	m.Add(10, 20, DirNorth)

	m.TriggerAt(10, 20, "player1")

	// Simulate many small updates
	totalTime := 0.0
	deltaTime := 1.0 / 60.0 // 60 FPS

	for totalTime < 1.5 {
		m.Update(deltaTime)
		totalTime += deltaTime
	}

	sw := m.Get(10, 20)
	if !sw.IsOpen() {
		t.Error("secret should be open after sufficient time")
	}
	if sw.Progress != 1.0 {
		t.Errorf("progress = %f, want 1.0", sw.Progress)
	}
}
