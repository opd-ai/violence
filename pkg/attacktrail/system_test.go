package attacktrail

import (
	"image/color"
	"math"
	"testing"
)

func TestTrailComponentCreation(t *testing.T) {
	comp := NewTrailComponent(5)
	if comp == nil {
		t.Fatal("NewTrailComponent returned nil")
	}
	if comp.MaxTrails != 5 {
		t.Errorf("Expected MaxTrails=5, got %d", comp.MaxTrails)
	}
	if comp.Type() != "AttackTrail" {
		t.Errorf("Expected Type()='AttackTrail', got %s", comp.Type())
	}
}

func TestTrailAddition(t *testing.T) {
	comp := NewTrailComponent(3)

	trail1 := CreateSlashTrail(10, 20, 0, 50, math.Pi/4, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	trail2 := CreateThrustTrail(15, 25, math.Pi/2, 60, 2, color.RGBA{R: 200, G: 200, B: 200, A: 255})
	trail3 := CreateCleaveTrail(20, 30, math.Pi, 70, math.Pi/3, 5, color.RGBA{R: 180, G: 180, B: 180, A: 255})
	trail4 := CreateSmashTrail(25, 35, 40, 4, color.RGBA{R: 160, G: 160, B: 160, A: 255})

	comp.AddTrail(trail1)
	comp.AddTrail(trail2)
	comp.AddTrail(trail3)

	if len(comp.Trails) != 3 {
		t.Errorf("Expected 3 trails, got %d", len(comp.Trails))
	}

	// Adding fourth should evict oldest
	comp.AddTrail(trail4)
	if len(comp.Trails) != 3 {
		t.Errorf("Expected 3 trails after eviction, got %d", len(comp.Trails))
	}
	if comp.Trails[0] != trail2 {
		t.Error("Oldest trail not evicted correctly")
	}
}

func TestTrailUpdate(t *testing.T) {
	comp := NewTrailComponent(5)

	trail := CreateSlashTrail(10, 20, 0, 50, math.Pi/4, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	comp.AddTrail(trail)

	// Update before fade start
	comp.Update(0.03)
	if trail.Age != 0.03 {
		t.Errorf("Expected Age=0.03, got %f", trail.Age)
	}
	if trail.Intensity != 1.0 {
		t.Errorf("Expected no fade before FadeStart, got Intensity=%f", trail.Intensity)
	}

	// Update into fade period
	comp.Update(0.05) // Total age = 0.08
	if trail.Intensity >= 1.0 {
		t.Error("Expected trail to start fading after FadeStart")
	}

	// Update past max age
	comp.Update(0.2) // Total age = 0.28, > MaxAge
	if len(comp.Trails) != 0 {
		t.Error("Expected trail to be removed after MaxAge")
	}
}

func TestTrailTypes(t *testing.T) {
	tests := []struct {
		name      string
		trailFunc func() *Trail
		wantType  TrailType
	}{
		{
			name: "Slash",
			trailFunc: func() *Trail {
				return CreateSlashTrail(0, 0, 0, 50, math.Pi/4, 3, color.RGBA{A: 255})
			},
			wantType: TrailSlash,
		},
		{
			name: "Thrust",
			trailFunc: func() *Trail {
				return CreateThrustTrail(0, 0, 0, 50, 2, color.RGBA{A: 255})
			},
			wantType: TrailThrust,
		},
		{
			name: "Cleave",
			trailFunc: func() *Trail {
				return CreateCleaveTrail(0, 0, 0, 50, math.Pi/3, 5, color.RGBA{A: 255})
			},
			wantType: TrailCleave,
		},
		{
			name: "Smash",
			trailFunc: func() *Trail {
				return CreateSmashTrail(0, 0, 40, 4, color.RGBA{A: 255})
			},
			wantType: TrailSmash,
		},
		{
			name: "Spin",
			trailFunc: func() *Trail {
				return CreateSpinTrail(0, 0, 50, 3, color.RGBA{A: 255})
			},
			wantType: TrailSpin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trail := tt.trailFunc()
			if trail.Type != tt.wantType {
				t.Errorf("Expected Type=%d, got %d", tt.wantType, trail.Type)
			}
			if trail.Intensity != 1.0 {
				t.Errorf("Expected initial Intensity=1.0, got %f", trail.Intensity)
			}
		})
	}
}

func TestSlashTrailProperties(t *testing.T) {
	x, y := 100.0, 200.0
	angle := math.Pi / 4
	range_ := 75.0
	arc := math.Pi / 3
	width := 4.0
	color := color.RGBA{R: 255, G: 200, B: 150, A: 200}

	trail := CreateSlashTrail(x, y, angle, range_, arc, width, color)

	if trail.StartX != x || trail.StartY != y {
		t.Errorf("Expected position (%f, %f), got (%f, %f)", x, y, trail.StartX, trail.StartY)
	}
	if trail.Angle != angle {
		t.Errorf("Expected angle %f, got %f", angle, trail.Angle)
	}
	if trail.Range != range_ {
		t.Errorf("Expected range %f, got %f", range_, trail.Range)
	}
	if trail.Arc != arc {
		t.Errorf("Expected arc %f, got %f", arc, trail.Arc)
	}
	if trail.Width != width {
		t.Errorf("Expected width %f, got %f", width, trail.Width)
	}
	if trail.Color != color {
		t.Errorf("Expected color %v, got %v", color, trail.Color)
	}
}

func TestSystemCreation(t *testing.T) {
	sys := NewSystem("fantasy")
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Expected genreID='fantasy', got '%s'", sys.genreID)
	}
}

func TestGenreWeaponColors(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "western"}

	for _, genre := range genres {
		sys := NewSystem(genre)
		color := sys.GetWeaponTrailColor("test_weapon", nil)

		if color.A == 0 {
			t.Errorf("Genre %s returned zero-alpha color", genre)
		}
	}
}

func BenchmarkTrailUpdate(b *testing.B) {
	comp := NewTrailComponent(10)

	// Fill with trails
	for i := 0; i < 10; i++ {
		trail := CreateSlashTrail(float64(i*10), float64(i*10), 0, 50, math.Pi/4, 3, color.RGBA{A: 255})
		comp.AddTrail(trail)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Update(0.016)
	}
}

func BenchmarkTrailAddition(b *testing.B) {
	comp := NewTrailComponent(100)
	trail := CreateSlashTrail(10, 20, 0, 50, math.Pi/4, 3, color.RGBA{A: 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.AddTrail(trail)
	}
}
