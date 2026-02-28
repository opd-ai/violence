package progression

import "testing"

func TestNewProgression(t *testing.T) {
	p := NewProgression()
	if p == nil {
		t.Fatal("NewProgression returned nil")
	}
	if p.Level != 1 {
		t.Errorf("New progression should start at level 1, got %d", p.Level)
	}
	if p.XP != 0 {
		t.Errorf("New progression should start with 0 XP, got %d", p.XP)
	}
}

func TestAddXP(t *testing.T) {
	tests := []struct {
		name     string
		amounts  []int
		expected int
	}{
		{"single_gain", []int{100}, 100},
		{"multiple_gains", []int{50, 30, 20}, 100},
		{"large_gain", []int{1000}, 1000},
		{"zero_gain", []int{0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgression()
			for _, amt := range tt.amounts {
				p.AddXP(amt)
			}
			if p.XP != tt.expected {
				t.Errorf("Expected %d XP, got %d", tt.expected, p.XP)
			}
		})
	}
}

func TestLevelUp(t *testing.T) {
	p := NewProgression()

	initialLevel := p.Level
	p.LevelUp()

	if p.Level != initialLevel+1 {
		t.Errorf("LevelUp should increase level by 1, got %d", p.Level)
	}

	// Multiple level ups
	p.LevelUp()
	p.LevelUp()

	if p.Level != initialLevel+3 {
		t.Errorf("After 3 level ups, expected level %d, got %d", initialLevel+3, p.Level)
	}
}

func TestProgressionCombined(t *testing.T) {
	p := NewProgression()

	// Simulate gameplay progression
	p.AddXP(100) // Kill some enemies
	p.LevelUp()  // Level 2

	if p.Level != 2 {
		t.Errorf("Expected level 2, got %d", p.Level)
	}
	if p.XP != 100 {
		t.Errorf("XP should persist at 100, got %d", p.XP)
	}

	p.AddXP(200) // More kills
	p.LevelUp()  // Level 3

	if p.Level != 3 {
		t.Errorf("Expected level 3, got %d", p.Level)
	}
	if p.XP != 300 {
		t.Errorf("Expected 300 total XP, got %d", p.XP)
	}
}

func TestSetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetGenre(%q) panicked: %v", genre, r)
				}
			}()
			SetGenre(genre)
		})
	}
}

func TestXPAccumulation(t *testing.T) {
	p := NewProgression()

	// Add XP incrementally
	for i := 1; i <= 10; i++ {
		p.AddXP(10)
		if p.XP != i*10 {
			t.Errorf("After %d additions, expected %d XP, got %d", i, i*10, p.XP)
		}
	}
}

func TestLevelProgression(t *testing.T) {
	p := NewProgression()

	// Level up multiple times
	for i := 1; i <= 10; i++ {
		p.LevelUp()
		expectedLevel := i + 1 // Started at 1
		if p.Level != expectedLevel {
			t.Errorf("After %d level ups, expected level %d, got %d", i, expectedLevel, p.Level)
		}
	}
}

func TestNegativeXP(t *testing.T) {
	p := NewProgression()
	p.AddXP(100)

	// Adding negative XP (edge case - could represent XP penalty)
	p.AddXP(-20)

	if p.XP != 80 {
		t.Errorf("After negative XP, expected 80, got %d", p.XP)
	}
}

func TestLargeXPValues(t *testing.T) {
	p := NewProgression()

	largeValue := 1000000
	p.AddXP(largeValue)

	if p.XP != largeValue {
		t.Errorf("Large XP value: expected %d, got %d", largeValue, p.XP)
	}
}

func TestMultipleProgressions(t *testing.T) {
	// Test multiple independent progression trackers
	p1 := NewProgression()
	p2 := NewProgression()

	p1.AddXP(100)
	p1.LevelUp()

	p2.AddXP(200)
	p2.LevelUp()
	p2.LevelUp()

	if p1.XP != 100 {
		t.Error("p1 XP affected by p2")
	}
	if p1.Level != 2 {
		t.Error("p1 level affected by p2")
	}
	if p2.XP != 200 {
		t.Error("p2 XP affected by p1")
	}
	if p2.Level != 3 {
		t.Error("p2 level affected by p1")
	}
}
