package wetness

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name       string
		x, y       int
		moisture   float64
		seed       int64
		wantPuddle bool
	}{
		{"dry_surface", 5, 5, 0.2, 12345, false},
		{"damp_surface", 10, 10, 0.5, 12345, false},
		{"shallow_puddle", 3, 3, 0.7, 12345, true},
		{"deep_puddle", 8, 8, 1.0, 12345, true},
		{"threshold_exact", 0, 0, 0.6, 12345, false},
		{"just_above_threshold", 1, 1, 0.61, 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent(tt.x, tt.y, tt.moisture, tt.seed)

			if comp.X != tt.x || comp.Y != tt.y {
				t.Errorf("Position = (%d, %d), want (%d, %d)", comp.X, comp.Y, tt.x, tt.y)
			}

			if comp.Moisture != tt.moisture {
				t.Errorf("Moisture = %f, want %f", comp.Moisture, tt.moisture)
			}

			if comp.IsPuddle != tt.wantPuddle {
				t.Errorf("IsPuddle = %v, want %v", comp.IsPuddle, tt.wantPuddle)
			}

			if comp.Seed != tt.seed {
				t.Errorf("Seed = %d, want %d", comp.Seed, tt.seed)
			}

			// Specular should increase with moisture
			if comp.SpecularIntensity < 0.3 || comp.SpecularIntensity > 1.0 {
				t.Errorf("SpecularIntensity = %f, want in [0.3, 1.0]", comp.SpecularIntensity)
			}
		})
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent(0, 0, 0.5, 12345)
	expected := "wetness.Component"
	if comp.Type() != expected {
		t.Errorf("Type() = %q, want %q", comp.Type(), expected)
	}
}

func TestWetnessPatternGetMoistureAt(t *testing.T) {
	pattern := &WetnessPattern{
		Width:  3,
		Height: 3,
		Cells:  make([][]*Component, 3),
	}

	for y := range pattern.Cells {
		pattern.Cells[y] = make([]*Component, 3)
	}

	// Set one cell
	pattern.Cells[1][1] = NewComponent(1, 1, 0.8, 12345)

	tests := []struct {
		name     string
		x, y     int
		expected float64
	}{
		{"wet_cell", 1, 1, 0.8},
		{"dry_cell", 0, 0, 0},
		{"out_of_bounds_negative", -1, -1, 0},
		{"out_of_bounds_positive", 5, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pattern.GetMoistureAt(tt.x, tt.y)
			if got != tt.expected {
				t.Errorf("GetMoistureAt(%d, %d) = %f, want %f", tt.x, tt.y, got, tt.expected)
			}
		})
	}
}

func TestWetnessPatternGetComponentAt(t *testing.T) {
	pattern := &WetnessPattern{
		Width:  3,
		Height: 3,
		Cells:  make([][]*Component, 3),
	}

	for y := range pattern.Cells {
		pattern.Cells[y] = make([]*Component, 3)
	}

	comp := NewComponent(1, 1, 0.8, 12345)
	pattern.Cells[1][1] = comp

	// Valid position with component
	got := pattern.GetComponentAt(1, 1)
	if got != comp {
		t.Error("GetComponentAt(1, 1) should return the component")
	}

	// Valid position without component
	got = pattern.GetComponentAt(0, 0)
	if got != nil {
		t.Error("GetComponentAt(0, 0) should return nil for empty cell")
	}

	// Out of bounds
	got = pattern.GetComponentAt(-1, 0)
	if got != nil {
		t.Error("GetComponentAt(-1, 0) should return nil for out of bounds")
	}

	got = pattern.GetComponentAt(0, 10)
	if got != nil {
		t.Error("GetComponentAt(0, 10) should return nil for out of bounds")
	}
}
