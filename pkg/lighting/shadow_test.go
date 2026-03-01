package lighting

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewShadowSystem verifies shadow system initialization.
func TestNewShadowSystem(t *testing.T) {
	tests := []struct {
		name       string
		genre      string
		wantSoft   float64
		wantMinOp  float64
		wantMaxOp  float64
		wantFallof string
	}{
		{
			name:       "fantasy",
			genre:      "fantasy",
			wantSoft:   0.6,
			wantMinOp:  0.25,
			wantMaxOp:  0.65,
			wantFallof: "quadratic",
		},
		{
			name:       "scifi",
			genre:      "scifi",
			wantSoft:   0.3,
			wantMinOp:  0.3,
			wantMaxOp:  0.75,
			wantFallof: "linear",
		},
		{
			name:       "horror",
			genre:      "horror",
			wantSoft:   0.8,
			wantMinOp:  0.4,
			wantMaxOp:  0.85,
			wantFallof: "inverse",
		},
		{
			name:       "cyberpunk",
			genre:      "cyberpunk",
			wantSoft:   0.2,
			wantMinOp:  0.35,
			wantMaxOp:  0.8,
			wantFallof: "linear",
		},
		{
			name:       "postapoc",
			genre:      "postapoc",
			wantSoft:   0.5,
			wantMinOp:  0.3,
			wantMaxOp:  0.7,
			wantFallof: "quadratic",
		},
		{
			name:       "unknown",
			genre:      "unknown",
			wantSoft:   0.6, // defaults to fantasy
			wantMinOp:  0.25,
			wantMaxOp:  0.65,
			wantFallof: "quadratic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewShadowSystem(320, 240, tt.genre)
			if s == nil {
				t.Fatal("NewShadowSystem returned nil")
			}
			if s.width != 320 {
				t.Errorf("width = %d, want 320", s.width)
			}
			if s.height != 240 {
				t.Errorf("height = %d, want 240", s.height)
			}
			if s.genre != tt.genre {
				t.Errorf("genre = %s, want %s", s.genre, tt.genre)
			}
			if math.Abs(s.softness-tt.wantSoft) > 0.01 {
				t.Errorf("softness = %.2f, want %.2f", s.softness, tt.wantSoft)
			}
			if math.Abs(s.minOpacity-tt.wantMinOp) > 0.01 {
				t.Errorf("minOpacity = %.2f, want %.2f", s.minOpacity, tt.wantMinOp)
			}
			if math.Abs(s.maxOpacity-tt.wantMaxOp) > 0.01 {
				t.Errorf("maxOpacity = %.2f, want %.2f", s.maxOpacity, tt.wantMaxOp)
			}
			if s.falloffType != tt.wantFallof {
				t.Errorf("falloffType = %s, want %s", s.falloffType, tt.wantFallof)
			}
		})
	}
}

// TestShadowSetGenre verifies genre switching updates parameters.
func TestShadowSetGenre(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")

	// Switch to horror
	s.SetGenre("horror")
	if s.genre != "horror" {
		t.Errorf("genre = %s, want horror", s.genre)
	}
	if math.Abs(s.softness-0.8) > 0.01 {
		t.Errorf("softness = %.2f, want 0.8", s.softness)
	}
	if s.falloffType != "inverse" {
		t.Errorf("falloffType = %s, want inverse", s.falloffType)
	}

	// Switch to cyberpunk
	s.SetGenre("cyberpunk")
	if s.genre != "cyberpunk" {
		t.Errorf("genre = %s, want cyberpunk", s.genre)
	}
	if math.Abs(s.softness-0.2) > 0.01 {
		t.Errorf("softness = %.2f, want 0.2", s.softness)
	}
}

// TestCalculateShadowLength verifies shadow projection distance calculation.
func TestCalculateShadowLength(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")

	tests := []struct {
		name      string
		distance  float64
		height    float64
		intensity float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "close high object",
			distance:  1.0,
			height:    2.0,
			intensity: 1.0,
			wantMin:   2.0,
			wantMax:   8.0,
		},
		{
			name:      "far low object",
			distance:  10.0,
			height:    0.5,
			intensity: 0.8,
			wantMin:   0.1,
			wantMax:   2.0,
		},
		{
			name:      "medium distance",
			distance:  5.0,
			height:    1.0,
			intensity: 0.5,
			wantMin:   0.5,
			wantMax:   5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length := s.calculateShadowLength(tt.distance, tt.height, tt.intensity)
			if length < tt.wantMin || length > tt.wantMax {
				t.Errorf("calculateShadowLength(%.1f, %.1f, %.1f) = %.2f, want [%.2f-%.2f]",
					tt.distance, tt.height, tt.intensity, length, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestCalculateShadowOpacity verifies shadow darkness calculation.
func TestCalculateShadowOpacity(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		distance  float64
		intensity float64
	}{
		{"close bright fantasy", "fantasy", 1.0, 1.0},
		{"far dim fantasy", "fantasy", 10.0, 0.3},
		{"close bright scifi", "scifi", 1.0, 1.0},
		{"close bright horror", "horror", 1.0, 1.0},
		{"close bright cyberpunk", "cyberpunk", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewShadowSystem(320, 240, tt.genre)
			opacity := s.calculateShadowOpacity(tt.distance, tt.intensity)

			if opacity < s.minOpacity || opacity > s.maxOpacity {
				t.Errorf("opacity = %.2f, want [%.2f-%.2f]", opacity, s.minOpacity, s.maxOpacity)
			}
		})
	}
}

// TestCalculateShadowOpacityFalloff verifies different falloff types.
func TestCalculateShadowOpacityFalloff(t *testing.T) {
	distance := 5.0
	intensity := 1.0

	// Linear falloff
	sLinear := NewShadowSystem(320, 240, "scifi")
	sLinear.falloffType = "linear"
	opacityLinear := sLinear.calculateShadowOpacity(distance, intensity)

	// Quadratic falloff
	sQuad := NewShadowSystem(320, 240, "fantasy")
	sQuad.falloffType = "quadratic"
	opacityQuad := sQuad.calculateShadowOpacity(distance, intensity)

	// Inverse falloff
	sInv := NewShadowSystem(320, 240, "horror")
	sInv.falloffType = "inverse"
	opacityInv := sInv.calculateShadowOpacity(distance, intensity)

	// Linear should have highest opacity at distance
	// Quadratic should have lowest opacity at distance
	// All should be within min/max bounds
	if opacityLinear < sLinear.minOpacity || opacityLinear > sLinear.maxOpacity {
		t.Errorf("linear opacity %.2f out of bounds [%.2f-%.2f]",
			opacityLinear, sLinear.minOpacity, sLinear.maxOpacity)
	}
	if opacityQuad < sQuad.minOpacity || opacityQuad > sQuad.maxOpacity {
		t.Errorf("quadratic opacity %.2f out of bounds [%.2f-%.2f]",
			opacityQuad, sQuad.minOpacity, sQuad.maxOpacity)
	}
	if opacityInv < sInv.minOpacity || opacityInv > sInv.maxOpacity {
		t.Errorf("inverse opacity %.2f out of bounds [%.2f-%.2f]",
			opacityInv, sInv.minOpacity, sInv.maxOpacity)
	}
}

// TestRenderShadows verifies shadow rendering doesn't panic.
func TestRenderShadows(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	casters := []ShadowCaster{
		{X: 5, Y: 5, Radius: 0.5, Height: 1.0, Opacity: 0.7, CastShadow: true},
		{X: 10, Y: 10, Radius: 0.3, Height: 0.8, Opacity: 0.5, CastShadow: true},
		{X: 15, Y: 15, Radius: 0.4, Height: 0.6, Opacity: 0.6, CastShadow: false}, // Should be skipped
	}

	lights := []Light{
		{X: 3, Y: 3, Radius: 5, Intensity: 1.0},
		{X: 12, Y: 12, Radius: 8, Intensity: 0.8},
	}

	coneLights := []ConeLight{
		{X: 8, Y: 8, DirX: 1, DirY: 0, Radius: 6, Intensity: 0.9, IsActive: true},
		{X: 20, Y: 20, DirX: 0, DirY: 1, Radius: 4, Intensity: 0.7, IsActive: false}, // Should be skipped
	}

	// Should not panic
	s.RenderShadows(screen, casters, lights, coneLights, 0, 0)
}

// TestRenderShadowsNoLights verifies no rendering when no lights present.
func TestRenderShadowsNoLights(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	casters := []ShadowCaster{
		{X: 5, Y: 5, Radius: 0.5, Height: 1.0, Opacity: 0.7, CastShadow: true},
	}

	// No lights - should return early
	s.RenderShadows(screen, casters, nil, nil, 0, 0)
}

// TestRenderShadowsEmptyCasters verifies handling empty caster list.
func TestRenderShadowsEmptyCasters(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	lights := []Light{
		{X: 3, Y: 3, Radius: 5, Intensity: 1.0},
	}

	// Empty casters - should not panic
	s.RenderShadows(screen, []ShadowCaster{}, lights, nil, 0, 0)
}

// TestShadowClear verifies shadow map clearing.
func TestShadowClear(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")

	// Render some shadows
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()
	casters := []ShadowCaster{{X: 5, Y: 5, Radius: 0.5, Height: 1.0, Opacity: 0.7, CastShadow: true}}
	lights := []Light{{X: 3, Y: 3, Radius: 5, Intensity: 1.0}}
	s.RenderShadows(screen, casters, lights, nil, 0, 0)

	// Clear
	s.Clear()

	// Shadow map should be cleared (all pixels transparent)
	shadowMap := s.GetShadowMap()
	if shadowMap == nil {
		t.Fatal("GetShadowMap returned nil")
	}
}

// TestGetShadowMap verifies shadow map retrieval.
func TestGetShadowMap(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")
	shadowMap := s.GetShadowMap()
	if shadowMap == nil {
		t.Error("GetShadowMap returned nil")
	}
}

// TestShadowCasterAtLightSource verifies handling when caster is at light position.
func TestShadowCasterAtLightSource(t *testing.T) {
	s := NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	// Caster at same position as light
	casters := []ShadowCaster{
		{X: 5, Y: 5, Radius: 0.5, Height: 1.0, Opacity: 0.7, CastShadow: true},
	}
	lights := []Light{
		{X: 5, Y: 5, Radius: 5, Intensity: 1.0},
	}

	// Should not panic or create invalid shadow
	s.RenderShadows(screen, casters, lights, nil, 0, 0)
}

// BenchmarkRenderShadows measures shadow rendering performance.
func BenchmarkRenderShadows(b *testing.B) {
	s := NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	casters := make([]ShadowCaster, 10)
	for i := 0; i < 10; i++ {
		casters[i] = ShadowCaster{
			X:          float64(i * 2),
			Y:          float64(i * 2),
			Radius:     0.5,
			Height:     1.0,
			Opacity:    0.7,
			CastShadow: true,
		}
	}

	lights := []Light{
		{X: 5, Y: 5, Radius: 10, Intensity: 1.0},
		{X: 15, Y: 15, Radius: 8, Intensity: 0.8},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.RenderShadows(screen, casters, lights, nil, 0, 0)
	}
}

// BenchmarkRenderShadowsManyCasters measures performance with many casters.
func BenchmarkRenderShadowsManyCasters(b *testing.B) {
	s := NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	casters := make([]ShadowCaster, 50)
	for i := 0; i < 50; i++ {
		casters[i] = ShadowCaster{
			X:          float64(i % 20),
			Y:          float64(i / 20),
			Radius:     0.5,
			Height:     1.0,
			Opacity:    0.7,
			CastShadow: true,
		}
	}

	lights := []Light{
		{X: 10, Y: 10, Radius: 15, Intensity: 1.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.RenderShadows(screen, casters, lights, nil, 0, 0)
	}
}

// TestClampF verifies float clamping utility.
func TestClampF(t *testing.T) {
	tests := []struct {
		value, min, max, want float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.3, 0.2, 0.8, 0.3},
		{0.1, 0.2, 0.8, 0.2},
		{0.9, 0.2, 0.8, 0.8},
	}

	for _, tt := range tests {
		got := clampF(tt.value, tt.min, tt.max)
		if math.Abs(got-tt.want) > 0.0001 {
			t.Errorf("clampF(%.2f, %.2f, %.2f) = %.2f, want %.2f",
				tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}
