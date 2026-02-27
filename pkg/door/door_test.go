package door

import "testing"

// TestTryOpenUnlocked verifies opening an unlocked door.
func TestTryOpenUnlocked(t *testing.T) {
	door := &Door{
		ID:     "door1",
		Locked: false,
	}
	keycard := &Keycard{Color: "red"}

	if !TryOpen(door, keycard) {
		t.Error("Expected unlocked door to open")
	}
}

// TestTryOpenLockedWithCorrectKeycard verifies opening a locked door with correct keycard.
func TestTryOpenLockedWithCorrectKeycard(t *testing.T) {
	tests := []struct {
		name     string
		required string
		color    string
		expected bool
	}{
		{"red keycard", "red", "red", true},
		{"blue keycard", "blue", "blue", true},
		{"yellow keycard", "yellow", "yellow", true},
		{"green keycard", "green", "green", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			door := &Door{
				ID:       "door1",
				Locked:   true,
				Required: tt.required,
			}
			keycard := &Keycard{Color: tt.color}

			result := TryOpen(door, keycard)
			if result != tt.expected {
				t.Errorf("TryOpen() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestTryOpenLockedWithWrongKeycard verifies locked door rejects wrong keycard.
func TestTryOpenLockedWithWrongKeycard(t *testing.T) {
	tests := []struct {
		name     string
		required string
		color    string
	}{
		{"red door, blue keycard", "red", "blue"},
		{"blue door, yellow keycard", "blue", "yellow"},
		{"yellow door, green keycard", "yellow", "green"},
		{"green door, red keycard", "green", "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			door := &Door{
				ID:       "door1",
				Locked:   true,
				Required: tt.required,
			}
			keycard := &Keycard{Color: tt.color}

			if TryOpen(door, keycard) {
				t.Error("Expected locked door with wrong keycard to stay closed")
			}
		})
	}
}

// TestDoorStruct verifies Door structure.
func TestDoorStruct(t *testing.T) {
	door := &Door{
		ID:       "test_door",
		Locked:   true,
		Required: "red",
	}

	if door.ID != "test_door" {
		t.Errorf("Expected ID 'test_door', got %s", door.ID)
	}
	if !door.Locked {
		t.Error("Expected Locked = true")
	}
	if door.Required != "red" {
		t.Errorf("Expected Required 'red', got %s", door.Required)
	}
}

// TestKeycardStruct verifies Keycard structure.
func TestKeycardStruct(t *testing.T) {
	keycard := &Keycard{Color: "blue"}

	if keycard.Color != "blue" {
		t.Errorf("Expected Color 'blue', got %s", keycard.Color)
	}
}

// TestSetGenre verifies SetGenre function exists and doesn't panic.
func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			SetGenre(tt.genreID)
		})
	}
}

// TestMultipleDoors verifies multiple doors with different keycards.
func TestMultipleDoors(t *testing.T) {
	doors := []*Door{
		{ID: "door1", Locked: true, Required: "red"},
		{ID: "door2", Locked: true, Required: "blue"},
		{ID: "door3", Locked: false, Required: ""},
	}

	redKey := &Keycard{Color: "red"}
	blueKey := &Keycard{Color: "blue"}

	// Test door1 with red key (should open)
	if !TryOpen(doors[0], redKey) {
		t.Error("door1 should open with red key")
	}

	// Test door1 with blue key (should not open)
	if TryOpen(doors[0], blueKey) {
		t.Error("door1 should not open with blue key")
	}

	// Test door2 with blue key (should open)
	if !TryOpen(doors[1], blueKey) {
		t.Error("door2 should open with blue key")
	}

	// Test door2 with red key (should not open)
	if TryOpen(doors[1], redKey) {
		t.Error("door2 should not open with red key")
	}

	// Test door3 (unlocked, should open with any key)
	if !TryOpen(doors[2], redKey) {
		t.Error("door3 should open (unlocked)")
	}
	if !TryOpen(doors[2], blueKey) {
		t.Error("door3 should open (unlocked)")
	}
}

// BenchmarkTryOpenUnlocked benchmarks opening unlocked doors.
func BenchmarkTryOpenUnlocked(b *testing.B) {
	door := &Door{ID: "door1", Locked: false}
	keycard := &Keycard{Color: "red"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TryOpen(door, keycard)
	}
}

// BenchmarkTryOpenLocked benchmarks opening locked doors.
func BenchmarkTryOpenLocked(b *testing.B) {
	door := &Door{ID: "door1", Locked: true, Required: "red"}
	keycard := &Keycard{Color: "red"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TryOpen(door, keycard)
	}
}
