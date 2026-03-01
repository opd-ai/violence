package progression

import (
	"fmt"
	"testing"
)

func TestNewProgression(t *testing.T) {
	p := NewProgression()
	if p == nil {
		t.Fatal("NewProgression returned nil")
	}
	if p.GetLevel() != 1 {
		t.Errorf("New progression should start at level 1, got %d", p.GetLevel())
	}
	if p.GetXP() != 0 {
		t.Errorf("New progression should start with 0 XP, got %d", p.GetXP())
	}
	if p.GetGenre() != "fantasy" {
		t.Errorf("New progression should default to fantasy genre, got %s", p.GetGenre())
	}
}

func TestAddXP(t *testing.T) {
	tests := []struct {
		name        string
		amounts     []int
		expectedXP  int
		expectedLvl int
	}{
		{"single_gain_no_levelup", []int{50}, 50, 1},
		{"reach_level_2", []int{100}, 0, 2},         // 100 XP = level up, 0 remaining
		{"exceed_level_2", []int{150}, 50, 2},       // 100 for level 2, 50 remaining
		{"multiple_gains", []int{50, 30, 20}, 0, 2}, // Total 100, auto-levels to 2
		{"zero_gain", []int{0}, 0, 1},
		{"multiple_levelups", []int{100, 200, 300}, 0, 4}, // 100->L2, 200->L3, 300->L4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgression()
			for _, amt := range tt.amounts {
				if err := p.AddXP(amt); err != nil {
					t.Fatalf("AddXP(%d) failed: %v", amt, err)
				}
			}
			if p.GetXP() != tt.expectedXP {
				t.Errorf("Expected %d XP, got %d", tt.expectedXP, p.GetXP())
			}
			if p.GetLevel() != tt.expectedLvl {
				t.Errorf("Expected level %d, got %d", tt.expectedLvl, p.GetLevel())
			}
		})
	}
}

func TestAutoLevelUp(t *testing.T) {
	p := NewProgression()

	// Level 1->2 requires 100 XP
	if err := p.AddXP(100); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	if p.GetLevel() != 2 {
		t.Errorf("Should auto-level to 2, got %d", p.GetLevel())
	}
	if p.GetXP() != 0 {
		t.Errorf("XP should reset after levelup, got %d", p.GetXP())
	}

	// Level 2->3 requires 200 XP
	if err := p.AddXP(200); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	if p.GetLevel() != 3 {
		t.Errorf("Should auto-level to 3, got %d", p.GetLevel())
	}
}

func TestProgressionCombined(t *testing.T) {
	p := NewProgression()

	// Kill some enemies (100 XP auto-levels to 2)
	if err := p.AddXP(100); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	if p.GetLevel() != 2 {
		t.Errorf("Expected level 2, got %d", p.GetLevel())
	}
	if p.GetXP() != 0 {
		t.Errorf("XP should reset to 0 after levelup, got %d", p.GetXP())
	}

	// More kills (200 XP auto-levels to 3)
	if err := p.AddXP(200); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	if p.GetLevel() != 3 {
		t.Errorf("Expected level 3, got %d", p.GetLevel())
	}
	if p.GetXP() != 0 {
		t.Errorf("XP should reset to 0 after levelup, got %d", p.GetXP())
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		genre     string
		expectErr bool
	}{
		{"fantasy", false},
		{"scifi", false},
		{"horror", false},
		{"cyberpunk", false},
		{"postapoc", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			p := NewProgression()
			err := p.SetGenre(tt.genre)
			if tt.expectErr && err == nil {
				t.Errorf("SetGenre(%q) should return error", tt.genre)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("SetGenre(%q) unexpected error: %v", tt.genre, err)
			}
			if !tt.expectErr && p.GetGenre() != tt.genre {
				t.Errorf("Genre not set correctly: expected %s, got %s", tt.genre, p.GetGenre())
			}
		})
	}
}

func TestXPAccumulation(t *testing.T) {
	p := NewProgression()

	// Add XP incrementally (10 * 10 = 100 = level up)
	for i := 1; i <= 10; i++ {
		if err := p.AddXP(10); err != nil {
			t.Fatalf("AddXP failed: %v", err)
		}
	}

	// Should have leveled up once at 100 XP
	if p.GetLevel() != 2 {
		t.Errorf("Should be level 2, got %d", p.GetLevel())
	}
	if p.GetXP() != 0 {
		t.Errorf("XP should reset after levelup, got %d", p.GetXP())
	}
}

func TestMultipleLevels(t *testing.T) {
	p := NewProgression()

	// Give enough XP to reach level 5
	// L1->L2: 100, L2->L3: 200, L3->L4: 300, L4->L5: 400 = 1000 total
	if err := p.AddXP(1000); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	if p.GetLevel() != 5 {
		t.Errorf("Should reach level 5, got %d", p.GetLevel())
	}
	if p.GetXP() != 0 {
		t.Errorf("Should have 0 remaining XP, got %d", p.GetXP())
	}
}

func TestNegativeXP(t *testing.T) {
	p := NewProgression()
	if err := p.AddXP(100); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	// Adding negative XP should return error
	err := p.AddXP(-200)
	if err == nil {
		t.Error("AddXP with negative amount should return error")
	}

	// XP should remain unchanged after error
	if p.GetXP() != 0 { // Started at 100 but leveled to 2, so 0 remaining
		t.Errorf("XP should be unchanged after error, got %d", p.GetXP())
	}

	// Valid negative that doesn't go below 0 should work
	if err := p.AddXP(50); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}
	if err := p.AddXP(-20); err != nil {
		t.Fatalf("AddXP with valid negative failed: %v", err)
	}
	if p.GetXP() != 30 {
		t.Errorf("Expected 30 XP after -20, got %d", p.GetXP())
	}
}

func TestLargeXPValues(t *testing.T) {
	p := NewProgression()

	largeValue := 1000000
	if err := p.AddXP(largeValue); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	// With large XP, should level to max (99)
	if p.GetLevel() != MaxLevel {
		t.Errorf("Large XP should reach max level %d, got %d", MaxLevel, p.GetLevel())
	}
}

func TestMultipleProgressions(t *testing.T) {
	// Test multiple independent progression trackers
	p1 := NewProgression()
	p2 := NewProgression()

	if err := p1.AddXP(100); err != nil {
		t.Fatalf("p1 AddXP failed: %v", err)
	}
	if err := p2.AddXP(300); err != nil {
		t.Fatalf("p2 AddXP failed: %v", err)
	}

	// p1: 100 XP = level 2, 0 remaining
	// p2: 300 XP = level 2 (100 consumed) + level 3 (200 consumed) = level 3, 0 remaining
	if p1.GetLevel() != 2 {
		t.Errorf("p1 should be level 2, got %d", p1.GetLevel())
	}
	if p1.GetXP() != 0 {
		t.Errorf("p1 should have 0 XP, got %d", p1.GetXP())
	}
	if p2.GetLevel() != 3 {
		t.Errorf("p2 should be level 3, got %d", p2.GetLevel())
	}
	if p2.GetXP() != 0 {
		t.Errorf("p2 should have 0 XP, got %d", p2.GetXP())
	}
}

func TestXPForNextLevel(t *testing.T) {
	tests := []struct {
		level    int
		addXP    int
		expected int
	}{
		{1, 0, 100},   // Level 1 needs 100 XP for level 2
		{1, 100, 200}, // At level 2, needs 200 XP for level 3
		{1, 300, 300}, // At level 3, needs 300 XP for level 4
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level_%d", tt.level), func(t *testing.T) {
			p := NewProgression()
			if tt.addXP > 0 {
				if err := p.AddXP(tt.addXP); err != nil {
					t.Fatalf("AddXP failed: %v", err)
				}
			}
			req := p.XPForNextLevel()
			if req != tt.expected {
				t.Errorf("At level %d, expected %d XP for next level, got %d", p.GetLevel(), tt.expected, req)
			}
		})
	}
}

func TestMaxLevelCap(t *testing.T) {
	p := NewProgression()

	// Give absurd amount of XP
	if err := p.AddXP(100000000); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}

	if p.GetLevel() > MaxLevel {
		t.Errorf("Level should cap at %d, got %d", MaxLevel, p.GetLevel())
	}

	// XPForNextLevel should return 0 at max level
	if p.XPForNextLevel() != 0 {
		t.Errorf("XPForNextLevel at max should be 0, got %d", p.XPForNextLevel())
	}
}

func TestConcurrentAccess(t *testing.T) {
	p := NewProgression()
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = p.AddXP(1)
			}
			done <- true
		}()
	}

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = p.GetXP()
				_ = p.GetLevel()
				_ = p.XPForNextLevel()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify state is consistent (no race condition crashes)
	level := p.GetLevel()
	xp := p.GetXP()
	if level < 1 || level > MaxLevel {
		t.Errorf("Invalid level after concurrent access: %d", level)
	}
	if xp < 0 {
		t.Errorf("Invalid XP after concurrent access: %d", xp)
	}
}
