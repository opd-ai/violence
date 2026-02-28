package ammo

import "testing"

func TestNewPool(t *testing.T) {
	pool := NewPool()
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
	if pool.counts == nil {
		t.Error("Pool counts map not initialized")
	}
}

func TestAddAmmo(t *testing.T) {
	tests := []struct {
		name     string
		ammoType string
		amounts  []int
		expected int
	}{
		{"bullets_single", Bullets, []int{50}, 50},
		{"bullets_multiple", Bullets, []int{50, 30, 20}, 100},
		{"shells_single", Shells, []int{8}, 8},
		{"rockets_accumulate", Rockets, []int{5, 10, 3}, 18},
		{"cells_zero", Cells, []int{0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool()
			for _, amt := range tt.amounts {
				pool.Add(tt.ammoType, amt)
			}
			if pool.counts[tt.ammoType] != tt.expected {
				t.Errorf("Expected %d %s, got %d", tt.expected, tt.ammoType, pool.counts[tt.ammoType])
			}
		})
	}
}

func TestConsumeAmmo(t *testing.T) {
	tests := []struct {
		name       string
		ammoType   string
		initial    int
		consume    int
		wantOk     bool
		wantRemain int
	}{
		{"sufficient_ammo", Bullets, 50, 10, true, 40},
		{"exact_ammo", Shells, 8, 8, true, 0},
		{"insufficient_ammo", Rockets, 5, 10, false, 5},
		{"zero_consume", Cells, 40, 0, true, 40},
		{"consume_from_zero", Bullets, 0, 1, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool()
			pool.Add(tt.ammoType, tt.initial)

			ok := pool.Consume(tt.ammoType, tt.consume)
			if ok != tt.wantOk {
				t.Errorf("Consume returned %v, want %v", ok, tt.wantOk)
			}

			remain := pool.counts[tt.ammoType]
			if remain != tt.wantRemain {
				t.Errorf("Expected %d remaining, got %d", tt.wantRemain, remain)
			}
		})
	}
}

func TestConsumeMultipleTimes(t *testing.T) {
	pool := NewPool()
	pool.Add(Bullets, 100)

	// First consumption
	if !pool.Consume(Bullets, 25) {
		t.Error("First consume failed")
	}
	if pool.counts[Bullets] != 75 {
		t.Errorf("After first consume: expected 75, got %d", pool.counts[Bullets])
	}

	// Second consumption
	if !pool.Consume(Bullets, 50) {
		t.Error("Second consume failed")
	}
	if pool.counts[Bullets] != 25 {
		t.Errorf("After second consume: expected 25, got %d", pool.counts[Bullets])
	}

	// Third consumption (should fail)
	if pool.Consume(Bullets, 30) {
		t.Error("Third consume should have failed but succeeded")
	}
	if pool.counts[Bullets] != 25 {
		t.Errorf("After failed consume: expected 25, got %d", pool.counts[Bullets])
	}
}

func TestMultipleAmmoTypes(t *testing.T) {
	pool := NewPool()

	// Add different ammo types
	pool.Add(Bullets, 100)
	pool.Add(Shells, 20)
	pool.Add(Rockets, 10)
	pool.Add(Cells, 50)

	// Verify all types stored correctly
	expected := map[string]int{
		Bullets: 100,
		Shells:  20,
		Rockets: 10,
		Cells:   50,
	}

	for ammoType, want := range expected {
		if pool.counts[ammoType] != want {
			t.Errorf("AmmoType %s: expected %d, got %d", ammoType, want, pool.counts[ammoType])
		}
	}

	// Consume from one type shouldn't affect others
	pool.Consume(Bullets, 50)
	if pool.counts[Shells] != 20 {
		t.Error("Consuming bullets affected shells")
	}
	if pool.counts[Rockets] != 10 {
		t.Error("Consuming bullets affected rockets")
	}
}

func TestAmmoConstants(t *testing.T) {
	// Verify ammo type constants are defined
	types := []string{Bullets, Shells, Rockets, Cells, Mana}
	expected := []string{"bullets", "shells", "rockets", "cells", "mana"}

	for i, typ := range types {
		if typ != expected[i] {
			t.Errorf("Ammo type constant %d: expected %q, got %q", i, expected[i], typ)
		}
	}
}

func TestNegativeAmounts(t *testing.T) {
	pool := NewPool()
	pool.Add(Bullets, 50)

	// Adding negative amount (edge case - should decrease)
	pool.Add(Bullets, -10)
	if pool.counts[Bullets] != 40 {
		t.Errorf("After adding -10: expected 40, got %d", pool.counts[Bullets])
	}

	// Consuming negative amount (edge case - should always succeed)
	ok := pool.Consume(Bullets, -5)
	if !ok {
		t.Error("Consuming negative amount should succeed")
	}
	if pool.counts[Bullets] != 45 {
		t.Errorf("After consuming -5: expected 45, got %d", pool.counts[Bullets])
	}
}

func TestSetGenre(t *testing.T) {
	// Test that SetGenre doesn't panic
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			SetGenre(genre) // Should not panic
		})
	}
}

func TestGetAmmo(t *testing.T) {
	tests := []struct {
		name     string
		setup    map[string]int
		ammoType string
		want     int
	}{
		{"empty_pool", map[string]int{}, Bullets, 0},
		{"has_bullets", map[string]int{Bullets: 50}, Bullets, 50},
		{"has_multiple", map[string]int{Bullets: 50, Shells: 20}, Shells, 20},
		{"zero_amount", map[string]int{Rockets: 0}, Rockets, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool()
			for ammoType, amount := range tt.setup {
				pool.Add(ammoType, amount)
			}
			got := pool.Get(tt.ammoType)
			if got != tt.want {
				t.Errorf("Get(%s) = %d, want %d", tt.ammoType, got, tt.want)
			}
		})
	}
}
