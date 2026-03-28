package caustics

import (
	"image/color"
	"testing"
)

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name       string
		worldX     float64
		worldY     float64
		intensity  float64
		radius     float64
		phase      float64
		sourceType SourceType
		col        color.RGBA
		seed       int64
	}{
		{
			name:       "puddle caustic",
			worldX:     10.5,
			worldY:     20.5,
			intensity:  0.8,
			radius:     2.0,
			phase:      1.5,
			sourceType: SourcePuddle,
			col:        color.RGBA{R: 200, G: 220, B: 255, A: 255},
			seed:       12345,
		},
		{
			name:       "pool caustic",
			worldX:     5.0,
			worldY:     5.0,
			intensity:  1.0,
			radius:     4.0,
			phase:      0,
			sourceType: SourcePool,
			col:        color.RGBA{R: 150, G: 200, B: 255, A: 255},
			seed:       67890,
		},
		{
			name:       "clamped intensity",
			worldX:     0,
			worldY:     0,
			intensity:  1.5, // Should be clamped to 1.0
			radius:     1.0,
			phase:      0,
			sourceType: SourceDrip,
			col:        color.RGBA{R: 255, G: 255, B: 255, A: 255},
			seed:       11111,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent(
				tt.worldX, tt.worldY,
				tt.intensity, tt.radius, tt.phase,
				tt.sourceType, tt.col, tt.seed,
			)

			if comp == nil {
				t.Fatal("NewComponent returned nil")
			}

			if comp.WorldX != tt.worldX {
				t.Errorf("WorldX = %f, want %f", comp.WorldX, tt.worldX)
			}
			if comp.WorldY != tt.worldY {
				t.Errorf("WorldY = %f, want %f", comp.WorldY, tt.worldY)
			}
			if comp.Radius != tt.radius {
				t.Errorf("Radius = %f, want %f", comp.Radius, tt.radius)
			}
			if comp.Phase != tt.phase {
				t.Errorf("Phase = %f, want %f", comp.Phase, tt.phase)
			}
			if comp.SourceType != tt.sourceType {
				t.Errorf("SourceType = %d, want %d", comp.SourceType, tt.sourceType)
			}
			if comp.Seed != tt.seed {
				t.Errorf("Seed = %d, want %d", comp.Seed, tt.seed)
			}

			// Check intensity clamping
			if comp.Intensity < 0 || comp.Intensity > 1 {
				t.Errorf("Intensity %f not clamped to [0,1]", comp.Intensity)
			}
		})
	}
}

func TestComponent_Type(t *testing.T) {
	comp := &Component{}
	if comp.Type() != "caustics" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "caustics")
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		min      float64
		max      float64
		expected float64
	}{
		{"within range", 0.5, 0, 1, 0.5},
		{"below min", -0.5, 0, 1, 0},
		{"above max", 1.5, 0, 1, 1},
		{"at min", 0, 0, 1, 0},
		{"at max", 1, 0, 1, 1},
		{"negative range", -0.5, -1, 0, -0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampFloat(tt.v, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clampFloat(%f, %f, %f) = %f, want %f",
					tt.v, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestSourceTypeConstants(t *testing.T) {
	// Verify source types are distinct
	types := []SourceType{SourcePuddle, SourcePool, SourceStream, SourceDrip}
	seen := make(map[SourceType]bool)

	for _, st := range types {
		if seen[st] {
			t.Errorf("Duplicate SourceType value: %d", st)
		}
		seen[st] = true
	}
}
