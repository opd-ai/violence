package damagestate

import (
	"testing"
)

func TestComponentUpdateDamage(t *testing.T) {
	tests := []struct {
		name          string
		currentHP     float64
		maxHP         float64
		expectedLevel int
	}{
		{"Pristine", 100, 100, 0},
		{"HighHP", 80, 100, 0},
		{"LightDamage", 60, 100, 1},
		{"ModerateDamage", 40, 100, 2},
		{"CriticalDamage", 20, 100, 3},
		{"VeryLowHP", 1, 100, 3},
		{"ZeroHP", 0, 100, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Component{
				CurrentHP: tt.currentHP,
				MaxHP:     tt.maxHP,
			}

			comp.UpdateDamage()

			if comp.DamageLevel != tt.expectedLevel {
				t.Errorf("UpdateDamage() level = %d, want %d", comp.DamageLevel, tt.expectedLevel)
			}
		})
	}
}

func TestComponentDirtyCacheOnLevelChange(t *testing.T) {
	comp := &Component{
		CurrentHP:   100,
		MaxHP:       100,
		DamageLevel: 0,
		DirtyCache:  false,
	}

	comp.UpdateDamage()
	if comp.DirtyCache {
		t.Error("DirtyCache should not be set when level doesn't change")
	}

	comp.CurrentHP = 40
	comp.UpdateDamage()
	if !comp.DirtyCache {
		t.Error("DirtyCache should be set when level changes")
	}
}

func TestComponentType(t *testing.T) {
	comp := &Component{}
	if comp.Type() != "damagestate" {
		t.Errorf("Type() = %s, want damagestate", comp.Type())
	}
}

func TestSystemCreation(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)
			if sys == nil {
				t.Fatal("NewSystem() returned nil")
			}
			if sys.genre != genre {
				t.Errorf("genre = %s, want %s", sys.genre, genre)
			}
			if sys.overlayCache == nil {
				t.Error("overlayCache not initialized")
			}
			if sys.logger == nil {
				t.Error("logger not initialized")
			}
		})
	}
}

func TestRenderDamageOverlay(t *testing.T) {
	// Skip tests that require graphics initialization
	t.Skip("Skipping graphics-dependent test in headless environment")
}

func TestRenderDamageOverlayCaching(t *testing.T) {
	// Skip tests that require graphics initialization
	t.Skip("Skipping graphics-dependent test in headless environment")
}

func TestGenreSpecificRendering(t *testing.T) {
	// Skip tests that require graphics initialization
	t.Skip("Skipping graphics-dependent test in headless environment")
}

func TestAllDamageLevels(t *testing.T) {
	// Skip tests that require graphics initialization
	t.Skip("Skipping graphics-dependent test in headless environment")
}

func TestDamageDirection(t *testing.T) {
	// Skip tests that require graphics initialization
	t.Skip("Skipping graphics-dependent test in headless environment")
}

func BenchmarkUpdateDamage(b *testing.B) {
	comp := &Component{
		CurrentHP: 50,
		MaxHP:     100,
	}

	for i := 0; i < b.N; i++ {
		comp.UpdateDamage()
	}
}

func BenchmarkRenderDamageOverlay(b *testing.B) {
	b.Skip("Skipping graphics-dependent benchmark in headless environment")
}

func BenchmarkGenerateDamageOverlay(b *testing.B) {
	b.Skip("Skipping graphics-dependent benchmark in headless environment")
}
