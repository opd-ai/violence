package hazard

import (
	"math"
	"testing"
)

func TestNewLegacySystem(t *testing.T) {
	s := NewLegacySystem(12345)
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if s.genre != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got %s", s.genre)
	}
	if len(s.hazards) != 0 {
		t.Errorf("Expected 0 initial hazards, got %d", len(s.hazards))
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name  string
		genre string
	}{
		{"Fantasy", "fantasy"},
		{"Scifi", "scifi"},
		{"Horror", "horror"},
		{"Cyberpunk", "cyberpunk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLegacySystem(12345)
			s.SetGenre(tt.genre)
			if s.genre != tt.genre {
				t.Errorf("Expected genre %s, got %s", tt.genre, s.genre)
			}
		})
	}
}

func TestGenerateHazards(t *testing.T) {
	testMap := make([][]int, 20)
	for i := range testMap {
		testMap[i] = make([]int, 20)
		for j := range testMap[i] {
			if i == 0 || i == 19 || j == 0 || j == 19 {
				testMap[i][j] = 1 // Walls
			} else {
				testMap[i][j] = 0 // Floor
			}
		}
	}

	s := NewLegacySystem(12345)
	s.GenerateHazards(testMap, 54321)

	if len(s.hazards) == 0 {
		t.Error("GenerateHazards created no hazards")
	}

	// Verify hazard placement
	for _, h := range s.hazards {
		x, y := int(h.X), int(h.Y)
		if x < 0 || x >= 20 || y < 0 || y >= 20 {
			t.Errorf("Hazard placed out of bounds: (%d, %d)", x, y)
		}
		if testMap[y][x] != 0 {
			t.Errorf("Hazard placed on non-floor tile at (%d, %d)", x, y)
		}
	}
}

func TestGenerateHazardsEmptyMap(t *testing.T) {
	s := NewLegacySystem(12345)
	s.GenerateHazards([][]int{}, 54321)
	if len(s.hazards) != 0 {
		t.Errorf("Expected 0 hazards for empty map, got %d", len(s.hazards))
	}
}

func TestGetGenreHazards(t *testing.T) {
	tests := []struct {
		name          string
		genre         string
		expectedTypes map[Type]bool
	}{
		{
			name:  "Fantasy",
			genre: "fantasy",
			expectedTypes: map[Type]bool{
				TypeSpikeTrap:    true,
				TypeFireGrate:    true,
				TypePoisonVent:   true,
				TypeFallingRocks: true,
				TypeAcidPool:     true,
			},
		},
		{
			name:  "Scifi",
			genre: "scifi",
			expectedTypes: map[Type]bool{
				TypeElectricFloor: true,
				TypeLaserGrid:     true,
				TypeCryoField:     true,
				TypePlasmaJet:     true,
				TypeGravityWell:   true,
			},
		},
		{
			name:  "Horror",
			genre: "horror",
			expectedTypes: map[Type]bool{
				TypeSpikeTrap:    true,
				TypePoisonVent:   true,
				TypeAcidPool:     true,
				TypeFallingRocks: true,
			},
		},
		{
			name:  "Cyberpunk",
			genre: "cyberpunk",
			expectedTypes: map[Type]bool{
				TypeElectricFloor: true,
				TypeLaserGrid:     true,
				TypePlasmaJet:     true,
				TypeGravityWell:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLegacySystem(12345)
			s.SetGenre(tt.genre)
			hazardTypes := s.getGenreHazards()

			if len(hazardTypes) == 0 {
				t.Error("No hazard types returned")
			}

			// Verify all types are expected
			for _, hType := range hazardTypes {
				if !tt.expectedTypes[hType] {
					t.Errorf("Unexpected hazard type %v for genre %s", hType, tt.genre)
				}
			}
		})
	}
}

func TestIsValidLocation(t *testing.T) {
	testMap := make([][]int, 10)
	for i := range testMap {
		testMap[i] = make([]int, 10)
		for j := range testMap[i] {
			if i == 0 || i == 9 || j == 0 || j == 9 {
				testMap[i][j] = 1
			} else {
				testMap[i][j] = 0
			}
		}
	}

	// Add a wall in the middle
	testMap[5][5] = 1

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"Wall tile", 0, 0, false},
		{"Center floor (few walls)", 4, 4, false},
		{"Near corner (many walls)", 1, 1, true},
		{"On wall", 5, 5, false},
	}

	s := NewLegacySystem(12345)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.isValidLocation(testMap, tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("isValidLocation(%d, %d) = %v, expected %v",
					tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestCreateHazard(t *testing.T) {
	s := NewLegacySystem(12345)

	tests := []struct {
		name         string
		hType        Type
		checkDamage  bool
		checkStatus  bool
		checkPersist bool
	}{
		{"SpikeTrap", TypeSpikeTrap, true, false, false},
		{"FireGrate", TypeFireGrate, true, true, true},
		{"PoisonVent", TypePoisonVent, true, true, true},
		{"ElectricFloor", TypeElectricFloor, true, true, false},
		{"FallingRocks", TypeFallingRocks, true, false, false},
		{"AcidPool", TypeAcidPool, true, true, true},
		{"LaserGrid", TypeLaserGrid, true, false, false},
		{"CryoField", TypeCryoField, true, true, true},
		{"PlasmaJet", TypePlasmaJet, true, true, false},
		{"GravityWell", TypeGravityWell, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := s.createHazard(tt.hType, 5.0, 5.0, s.rng)

			if h == nil {
				t.Fatal("createHazard returned nil")
			}

			if h.Type != tt.hType {
				t.Errorf("Expected type %v, got %v", tt.hType, h.Type)
			}

			if h.X != 5.0 || h.Y != 5.0 {
				t.Errorf("Expected position (5.0, 5.0), got (%.1f, %.1f)", h.X, h.Y)
			}

			if tt.checkDamage && h.Damage <= 0 {
				t.Error("Expected damage > 0")
			}

			if tt.checkStatus && h.StatusEffect == "" {
				t.Error("Expected status effect to be set")
			}

			if h.Persistent != tt.checkPersist {
				t.Errorf("Expected persistent=%v, got %v", tt.checkPersist, h.Persistent)
			}

			if h.CycleDuration <= 0 && tt.hType != TypeAcidPool {
				t.Error("Expected CycleDuration > 0 for non-permanent hazard")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	s := NewLegacySystem(12345)
	h := s.createHazard(TypeSpikeTrap, 5.0, 5.0, s.rng)
	h.Timer = 0
	s.hazards = []*Hazard{h}

	// Update and verify state transitions
	deltaTime := 0.1

	// Should start in charging state
	s.Update(deltaTime)
	if h.Timer <= 0 {
		t.Error("Timer should advance")
	}

	// Advance to active state
	for h.State != StateActive {
		s.Update(deltaTime)
		if h.Timer > h.CycleDuration*2 {
			t.Fatal("Hazard never activated")
		}
	}

	// Verify it eventually returns to inactive
	initialTimer := h.Timer
	for i := 0; i < 100; i++ {
		s.Update(deltaTime)
	}
	if h.Timer <= initialTimer {
		t.Error("Timer should continue advancing")
	}
}

func TestCheckCollision(t *testing.T) {
	s := NewLegacySystem(12345)
	h := s.createHazard(TypeFireGrate, 5.0, 5.0, s.rng)
	h.State = StateActive
	s.hazards = []*Hazard{h}

	tests := []struct {
		name        string
		x, y        float64
		expectHit   bool
		expectDmg   bool
		expectEfect bool
	}{
		{"Direct hit", 5.0, 5.0, true, true, true},
		{"Near hit", 5.2, 5.2, true, true, true},
		{"Miss", 10.0, 10.0, false, false, false},
		{"Edge hit", 5.49, 5.0, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hit, dmg, effect := s.CheckCollision(tt.x, tt.y)

			if hit != tt.expectHit {
				t.Errorf("Expected hit=%v, got %v", tt.expectHit, hit)
			}

			if tt.expectDmg && dmg == 0 {
				t.Error("Expected damage > 0")
			}

			if tt.expectEfect && effect == "" {
				t.Error("Expected status effect")
			}
		})
	}
}

func TestCheckCollisionInactive(t *testing.T) {
	s := NewLegacySystem(12345)
	h := s.createHazard(TypeSpikeTrap, 5.0, 5.0, s.rng)
	h.State = StateInactive
	s.hazards = []*Hazard{h}

	hit, dmg, effect := s.CheckCollision(5.0, 5.0)
	if hit {
		t.Error("Inactive hazard should not trigger collision")
	}
	if dmg != 0 {
		t.Error("Inactive hazard should deal no damage")
	}
	if effect != "" {
		t.Error("Inactive hazard should have no effect")
	}
}

func TestCheckCollisionOneShot(t *testing.T) {
	s := NewLegacySystem(12345)
	h := s.createHazard(TypeSpikeTrap, 5.0, 5.0, s.rng)
	h.State = StateActive
	h.Persistent = false
	h.Triggered = false
	s.hazards = []*Hazard{h}

	// First hit should trigger
	hit1, dmg1, _ := s.CheckCollision(5.0, 5.0)
	if !hit1 || dmg1 == 0 {
		t.Error("First collision should hit and damage")
	}

	// Second hit on same activation should not trigger
	hit2, dmg2, _ := s.CheckCollision(5.0, 5.0)
	if hit2 || dmg2 != 0 {
		t.Error("Second collision on one-shot hazard should not trigger")
	}
}

func TestGetHazards(t *testing.T) {
	s := NewLegacySystem(12345)
	h1 := s.createHazard(TypeSpikeTrap, 5.0, 5.0, s.rng)
	h2 := s.createHazard(TypeFireGrate, 7.0, 7.0, s.rng)
	s.hazards = []*Hazard{h1, h2}

	hazards := s.GetHazards()
	if len(hazards) != 2 {
		t.Errorf("Expected 2 hazards, got %d", len(hazards))
	}
}

func TestClear(t *testing.T) {
	s := NewLegacySystem(12345)
	s.hazards = []*Hazard{
		s.createHazard(TypeSpikeTrap, 5.0, 5.0, s.rng),
		s.createHazard(TypeFireGrate, 7.0, 7.0, s.rng),
	}

	s.Clear()
	if len(s.hazards) != 0 {
		t.Errorf("Expected 0 hazards after Clear(), got %d", len(s.hazards))
	}
}

func TestHazardStateMachine(t *testing.T) {
	s := NewLegacySystem(12345)
	h := s.createHazard(TypeElectricFloor, 5.0, 5.0, s.rng)
	h.Timer = 0
	h.ChargeDuration = 1.0
	h.ActiveDuration = 1.0
	h.CooldownDuration = 1.0
	h.CycleDuration = 3.0
	s.hazards = []*Hazard{h}

	// Test state progression
	states := []State{}
	for i := 0; i < 40; i++ {
		s.Update(0.1)
		if len(states) == 0 || states[len(states)-1] != h.State {
			states = append(states, h.State)
		}
	}

	// Should cycle through states
	if len(states) < 3 {
		t.Errorf("Expected at least 3 state changes, got %d: %v", len(states), states)
	}
}

func TestHazardDimensions(t *testing.T) {
	s := NewLegacySystem(12345)

	tests := []struct {
		hType         Type
		expectedWidth float64
	}{
		{TypeFallingRocks, 2.0},
		{TypeCryoField, 2.0},
		{TypeGravityWell, 2.5},
		{TypeSpikeTrap, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.hType.String(), func(t *testing.T) {
			h := s.createHazard(tt.hType, 5.0, 5.0, s.rng)
			if math.Abs(h.Width-tt.expectedWidth) > 0.01 {
				t.Errorf("Expected width %.1f, got %.1f", tt.expectedWidth, h.Width)
			}
		})
	}
}

func BenchmarkUpdate(b *testing.B) {
	s := NewLegacySystem(12345)
	for i := 0; i < 64; i++ {
		h := s.createHazard(Type(i%10), float64(i), float64(i), s.rng)
		s.hazards = append(s.hazards, h)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(0.016)
	}
}

func BenchmarkCheckCollision(b *testing.B) {
	s := NewLegacySystem(12345)
	for i := 0; i < 64; i++ {
		h := s.createHazard(Type(i%10), float64(i), float64(i), s.rng)
		h.State = StateActive
		s.hazards = append(s.hazards, h)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CheckCollision(32.0, 32.0)
	}
}
