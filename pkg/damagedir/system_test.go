package damagedir

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name      string
		genreID   string
		wantColor color.RGBA
	}{
		{"fantasy", "fantasy", color.RGBA{R: 200, G: 30, B: 30, A: 255}},
		{"scifi", "scifi", color.RGBA{R: 255, G: 100, B: 80, A: 255}},
		{"horror", "horror", color.RGBA{R: 150, G: 0, B: 0, A: 255}},
		{"cyberpunk", "cyberpunk", color.RGBA{R: 255, G: 50, B: 100, A: 255}},
		{"postapoc", "postapoc", color.RGBA{R: 180, G: 60, B: 40, A: 255}},
		{"unknown defaults to fantasy", "unknown", color.RGBA{R: 200, G: 30, B: 30, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.preset.BaseColor != tt.wantColor {
				t.Errorf("BaseColor = %v, want %v", sys.preset.BaseColor, tt.wantColor)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	originalColor := sys.preset.BaseColor

	sys.SetGenre("cyberpunk")
	if sys.preset.BaseColor == originalColor {
		t.Error("SetGenre did not change preset")
	}
	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.SetScreenSize(640, 480)
	if sys.screenW != 640 || sys.screenH != 480 {
		t.Errorf("screen size = %dx%d, want 640x480", sys.screenW, sys.screenH)
	}
}

func TestTriggerDamage(t *testing.T) {
	sys := NewSystem("fantasy")

	// Trigger damage from the right
	sys.TriggerDamage(150, 100, 100, 100, 50, 0)

	if len(sys.indicators) != 1 {
		t.Fatalf("indicator count = %d, want 1", len(sys.indicators))
	}

	ind := sys.indicators[0]
	if ind.Intensity <= 0 {
		t.Error("indicator intensity should be > 0")
	}
	if ind.Lifetime <= 0 {
		t.Error("indicator lifetime should be > 0")
	}
}

func TestTriggerDamage_DirectionCalculation(t *testing.T) {
	tests := []struct {
		name             string
		sourceX, sourceY float64
		playerX, playerY float64
		playerAngle      float64
		expectedDir      float64 // approximate
		tolerance        float64
	}{
		{
			name:    "damage from right, player facing east",
			sourceX: 150, sourceY: 100,
			playerX: 100, playerY: 100,
			playerAngle: 0,
			expectedDir: 0, // right of screen
			tolerance:   0.1,
		},
		{
			name:    "damage from left, player facing east",
			sourceX: 50, sourceY: 100,
			playerX: 100, playerY: 100,
			playerAngle: 0,
			expectedDir: math.Pi, // left of screen (or -math.Pi)
			tolerance:   0.1,
		},
		{
			name:    "damage from behind, player facing east",
			sourceX: 50, sourceY: 100,
			playerX: 100, playerY: 100,
			playerAngle: 0, // facing right, damage from left = behind
			expectedDir: math.Pi,
			tolerance:   0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem("fantasy")
			sys.TriggerDamage(tt.sourceX, tt.sourceY, tt.playerX, tt.playerY, 50, tt.playerAngle)

			if len(sys.indicators) != 1 {
				t.Fatalf("indicator count = %d, want 1", len(sys.indicators))
			}

			dir := sys.indicators[0].Direction
			// Check if direction is approximately correct (handle wraparound)
			diff := math.Abs(dir - tt.expectedDir)
			if diff > math.Pi {
				diff = 2*math.Pi - diff
			}
			if diff > tt.tolerance {
				t.Errorf("direction = %f, want approximately %f (diff=%f)", dir, tt.expectedDir, diff)
			}
		})
	}
}

func TestTriggerDamage_IgnoresZeroDistance(t *testing.T) {
	sys := NewSystem("fantasy")

	// Source at same position as player
	sys.TriggerDamage(100, 100, 100, 100, 50, 0)

	if len(sys.indicators) != 0 {
		t.Error("should not create indicator when source is at player position")
	}
}

func TestTriggerDamage_MaxIndicators(t *testing.T) {
	sys := NewSystem("fantasy")

	// Trigger more than max indicators
	for i := 0; i < 12; i++ {
		sys.TriggerDamage(float64(100+i*10), 100, 100, 100, 25, 0)
	}

	if len(sys.indicators) > sys.maxIndicators {
		t.Errorf("indicator count %d exceeds max %d", len(sys.indicators), sys.maxIndicators)
	}
}

func TestTriggerDamage_IntensityScaling(t *testing.T) {
	sys := NewSystem("fantasy")

	// Low damage
	sys.TriggerDamage(150, 100, 100, 100, 10, 0)
	lowIntensity := sys.indicators[0].Intensity

	// High damage
	sys.TriggerDamage(150, 100, 100, 100, 90, 0)
	highIntensity := sys.indicators[1].Intensity

	if highIntensity <= lowIntensity {
		t.Errorf("high damage intensity (%f) should be > low damage intensity (%f)",
			highIntensity, lowIntensity)
	}
}

func TestComponent_GetAlpha(t *testing.T) {
	comp := &Component{
		Intensity:   0.8,
		Lifetime:    1.5,
		MaxLifetime: 1.5,
	}

	// Full lifetime = full alpha
	alpha := comp.GetAlpha()
	if alpha < 0.7 {
		t.Errorf("alpha at full lifetime = %f, want ~0.8", alpha)
	}

	// Half lifetime = reduced alpha
	comp.Lifetime = 0.75
	alpha = comp.GetAlpha()
	if alpha >= 0.8 {
		t.Errorf("alpha at half lifetime = %f, should be < 0.8", alpha)
	}

	// Zero lifetime = zero alpha
	comp.Lifetime = 0
	alpha = comp.GetAlpha()
	if alpha != 0 {
		t.Errorf("alpha at zero lifetime = %f, want 0", alpha)
	}
}

func TestComponent_IsExpired(t *testing.T) {
	comp := &Component{Lifetime: 1.0}
	if comp.IsExpired() {
		t.Error("should not be expired with lifetime > 0")
	}

	comp.Lifetime = 0
	if !comp.IsExpired() {
		t.Error("should be expired with lifetime = 0")
	}

	comp.Lifetime = -0.1
	if !comp.IsExpired() {
		t.Error("should be expired with negative lifetime")
	}
}

func TestComponent_Type(t *testing.T) {
	comp := &Component{}
	expected := "damagedir.Component"
	if comp.Type() != expected {
		t.Errorf("Type() = %s, want %s", comp.Type(), expected)
	}
}

func TestGetActiveCount(t *testing.T) {
	sys := NewSystem("fantasy")

	if sys.GetActiveCount() != 0 {
		t.Error("initial active count should be 0")
	}

	sys.TriggerDamage(150, 100, 100, 100, 50, 0)
	if sys.GetActiveCount() != 1 {
		t.Errorf("active count = %d, want 1", sys.GetActiveCount())
	}
}

func TestClear(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.TriggerDamage(150, 100, 100, 100, 50, 0)
	sys.TriggerDamage(50, 100, 100, 100, 50, 0)

	sys.Clear()

	if sys.GetActiveCount() != 0 {
		t.Errorf("active count after Clear = %d, want 0", sys.GetActiveCount())
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		got := clamp(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestRender_NoIndicators(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(320, 200)

	// Should not panic with no indicators
	sys.Render(screen)
}

func TestRender_WithIndicators(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	screen := ebiten.NewImage(320, 200)

	sys.TriggerDamage(150, 100, 100, 100, 50, 0)

	// Should not panic
	sys.Render(screen)

	// Screen should have been drawn to (non-zero pixels)
	// Note: Due to additive blending and small arc, checking is tricky
	// Just verify no panic occurred
}

func BenchmarkTriggerDamage(b *testing.B) {
	sys := NewSystem("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.TriggerDamage(150, 100, 100, 100, 50, 0)
		if len(sys.indicators) > 4 {
			sys.indicators = sys.indicators[:0]
		}
	}
}

func BenchmarkRender(b *testing.B) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	screen := ebiten.NewImage(320, 200)

	// Add several indicators
	for i := 0; i < 4; i++ {
		sys.TriggerDamage(float64(100+i*25), 100, 100, 100, 50, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Render(screen)
	}
}
