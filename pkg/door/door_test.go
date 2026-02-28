package door

import "testing"

// TestTryOpenUnlocked verifies opening an unlocked door.
func TestTryOpenUnlocked(t *testing.T) {
	door := &Door{
		ID:              "door1",
		Locked:          false,
		RequiredKeycard: "",
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
				ID:              "door1",
				Locked:          true,
				RequiredKeycard: tt.required,
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
				ID:              "door1",
				Locked:          true,
				RequiredKeycard: tt.required,
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
		ID:              "test_door",
		Locked:          true,
		RequiredKeycard: "red",
	}

	if door.ID != "test_door" {
		t.Errorf("Expected ID 'test_door', got %s", door.ID)
	}
	if !door.Locked {
		t.Error("Expected Locked = true")
	}
	if door.RequiredKeycard != "red" {
		t.Errorf("Expected RequiredKeycard 'red', got %s", door.RequiredKeycard)
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
		{ID: "door1", Locked: true, RequiredKeycard: "red"},
		{ID: "door2", Locked: true, RequiredKeycard: "blue"},
		{ID: "door3", Locked: false, RequiredKeycard: ""},
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
	door := &Door{ID: "door1", Locked: true, RequiredKeycard: "red"}
	keycard := &Keycard{Color: "red"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TryOpen(door, keycard)
	}
}

// TestKeycardInventory verifies KeycardInventory functionality.
func TestKeycardInventory(t *testing.T) {
	inv := NewKeycardInventory()

	// Initially empty
	if inv.HasKeycard("red") {
		t.Error("Expected empty inventory")
	}

	// Add red keycard
	inv.AddKeycard("red")
	if !inv.HasKeycard("red") {
		t.Error("Expected red keycard in inventory")
	}

	// Add blue and yellow
	inv.AddKeycard("blue")
	inv.AddKeycard("yellow")

	if !inv.HasKeycard("blue") {
		t.Error("Expected blue keycard in inventory")
	}
	if !inv.HasKeycard("yellow") {
		t.Error("Expected yellow keycard in inventory")
	}
	if inv.HasKeycard("green") {
		t.Error("Expected green keycard NOT in inventory")
	}

	// Get all keycards
	all := inv.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 keycards, got %d", len(all))
	}
}

// TestDoorSystem verifies DoorSystem functionality.
func TestDoorSystem(t *testing.T) {
	ds := NewDoorSystem()
	inv := NewKeycardInventory()

	// Create doors
	door1 := NewDoor("door1", 10.0, 20.0, TypeSwing, true, "red")
	door2 := NewDoor("door2", 30.0, 40.0, TypeSliding, false, "")

	ds.AddDoor(door1)
	ds.AddDoor(door2)

	// Try to open locked door without keycard
	success, msg := ds.TryOpen(door1, inv)
	if success {
		t.Error("Expected locked door to reject opening without keycard")
	}
	if msg != "Need red keycard" {
		t.Errorf("Expected 'Need red keycard', got '%s'", msg)
	}

	// Add red keycard and try again
	inv.AddKeycard("red")
	success, msg = ds.TryOpen(door1, inv)
	if !success {
		t.Errorf("Expected door to open with red keycard, got: %s", msg)
	}
	if door1.State != StateOpening {
		t.Errorf("Expected door state StateOpening, got %d", door1.State)
	}

	// Animate door opening
	for door1.State == StateOpening {
		ds.Update()
	}

	if door1.State != StateOpen {
		t.Errorf("Expected door state StateOpen after animation, got %d", door1.State)
	}

	// Try to open already open door
	success, _ = ds.TryOpen(door1, inv)
	if success {
		t.Error("Expected already open door to not re-open")
	}

	// Close door
	if !ds.Close(door1) {
		t.Error("Expected door to close")
	}
	if door1.State != StateClosing {
		t.Errorf("Expected door state StateClosing, got %d", door1.State)
	}

	// Animate door closing
	for door1.State == StateClosing {
		ds.Update()
	}

	if door1.State != StateClosed {
		t.Errorf("Expected door state StateClosed after animation, got %d", door1.State)
	}
}

// TestDoorStates verifies door state transitions.
func TestDoorStates(t *testing.T) {
	door := NewDoor("test", 0, 0, TypeSwing, false, "")

	if door.State != StateClosed {
		t.Error("Expected new door to be StateClosed")
	}

	door.State = StateOpening
	door.AnimationFrame = 0

	ds := NewDoorSystem()
	ds.AddDoor(door)

	// Simulate animation
	for i := 0; i < door.MaxFrames; i++ {
		ds.Update()
	}

	if door.State != StateOpen {
		t.Errorf("Expected StateOpen after animation, got %d", door.State)
	}
	if door.AnimationFrame != door.MaxFrames {
		t.Errorf("Expected frame %d, got %d", door.MaxFrames, door.AnimationFrame)
	}
}

// TestDoorTypes verifies all door types can be created.
func TestDoorTypes(t *testing.T) {
	types := []DoorType{TypeSwing, TypeSliding, TypePortcullis, TypeShutter, TypeLaserBarrier}

	for _, dt := range types {
		door := NewDoor("test", 0, 0, dt, false, "")
		if door.Type != dt {
			t.Errorf("Expected door type %d, got %d", dt, door.Type)
		}
	}
}

// TestSetGenreKeycardNames verifies genre-specific keycard names.
func TestSetGenreKeycardNames(t *testing.T) {
	tests := []struct {
		genre    string
		color    string
		expected string
	}{
		{"fantasy", "red", "Crimson Rune"},
		{"scifi", "blue", "Blue Clearance"},
		{"horror", "yellow", "Rusty Key"},
		{"cyberpunk", "red", "Red Biometric"},
		{"postapoc", "blue", "Blue Badge"},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.color, func(t *testing.T) {
			SetGenre(tt.genre)
			name := GetKeycardDisplayName(tt.color)
			if name != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, name)
			}
		})
	}
}

// TestGenreDoorTypes verifies genre-specific door types.
func TestGenreDoorTypes(t *testing.T) {
	tests := []struct {
		genre    string
		expected DoorType
	}{
		{"fantasy", TypePortcullis},
		{"scifi", TypeSliding},
		{"horror", TypeSwing},
		{"cyberpunk", TypeShutter},
		{"postapoc", TypeSwing},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			SetGenre(tt.genre)
			doorType := GetGenreDoorType()
			if doorType != tt.expected {
				t.Errorf("Expected door type %d, got %d", tt.expected, doorType)
			}
		})
	}
}

// TestLockedDoorRejection verifies locked door behavior in DoorSystem.
func TestLockedDoorRejection(t *testing.T) {
	ds := NewDoorSystem()
	inv := NewKeycardInventory()
	door := NewDoor("locked", 0, 0, TypeSwing, true, "blue")
	ds.AddDoor(door)

	// Try without keycard
	success, msg := ds.TryOpen(door, inv)
	if success {
		t.Error("Expected failure without keycard")
	}
	if msg != "Need blue keycard" {
		t.Errorf("Expected 'Need blue keycard', got '%s'", msg)
	}

	// Try with wrong keycard
	inv.AddKeycard("red")
	success, msg = ds.TryOpen(door, inv)
	if success {
		t.Error("Expected failure with wrong keycard")
	}
	if msg != "Need blue keycard" {
		t.Errorf("Expected 'Need blue keycard', got '%s'", msg)
	}

	// Try with correct keycard
	inv.AddKeycard("blue")
	success, _ = ds.TryOpen(door, inv)
	if !success {
		t.Error("Expected success with correct keycard")
	}
}
