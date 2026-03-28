package heatdistort

import (
	"math"
	"testing"
)

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name       string
		sourceType HeatSourceType
		wantRadius float64
		wantIntMin float64
		wantIntMax float64
	}{
		{"torch", HeatTorch, 1.5, 0.3, 0.5},
		{"brazier", HeatBrazier, 2.5, 0.5, 0.7},
		{"lava", HeatLava, 4.0, 0.7, 0.9},
		{"explosion", HeatExplosion, 5.0, 0.9, 1.1},
		{"plasma", HeatPlasma, 2.0, 0.4, 0.6},
		{"radiation", HeatRadiation, 3.5, 0.5, 0.7},
		{"magic", HeatMagic, 2.0, 0.4, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComponent(tt.sourceType)

			if c.SourceType != tt.sourceType {
				t.Errorf("SourceType = %v, want %v", c.SourceType, tt.sourceType)
			}

			if c.Radius != tt.wantRadius {
				t.Errorf("Radius = %f, want %f", c.Radius, tt.wantRadius)
			}

			if c.Intensity < tt.wantIntMin || c.Intensity > tt.wantIntMax {
				t.Errorf("Intensity = %f, want in [%f, %f]", c.Intensity, tt.wantIntMin, tt.wantIntMax)
			}

			if c.Type() != "heatdistort.Component" {
				t.Errorf("Type() = %s, want heatdistort.Component", c.Type())
			}
		})
	}
}

func TestComponentLifetime(t *testing.T) {
	// Permanent source
	c := NewComponent(HeatTorch)
	if c.UpdateLifetime(1.0) {
		t.Error("permanent source should not expire")
	}

	// Temporary source (explosion)
	c = NewComponent(HeatExplosion)
	initialLifetime := c.Lifetime
	if initialLifetime <= 0 {
		t.Error("explosion should have positive initial lifetime")
	}

	// Should not expire after partial time
	if c.UpdateLifetime(initialLifetime / 2) {
		t.Error("should not expire after half lifetime")
	}

	// Should expire after remaining time
	if !c.UpdateLifetime(initialLifetime) {
		t.Error("should expire after full lifetime")
	}
}

func TestComponentIsActive(t *testing.T) {
	c := NewComponent(HeatTorch)

	// Initially active
	if !c.IsActive() {
		t.Error("new component should be active")
	}

	// Inactive when not visible
	c.SetVisible(false)
	if c.IsActive() {
		t.Error("invisible component should not be active")
	}

	// Inactive when intensity is zero
	c.SetVisible(true)
	c.Intensity = 0
	if c.IsActive() {
		t.Error("zero-intensity component should not be active")
	}
}

func TestComponentScreenPosition(t *testing.T) {
	c := NewComponent(HeatTorch)
	c.SetScreenPosition(100.5, 200.5)

	if c.ScreenX != 100.5 || c.ScreenY != 200.5 {
		t.Errorf("screen position = (%f, %f), want (100.5, 200.5)", c.ScreenX, c.ScreenY)
	}
}

func TestComponentPhaseUpdate(t *testing.T) {
	c := NewComponent(HeatTorch)
	initialPhase := c.WavePhase

	c.UpdatePhase(0.1, 5.0)

	expectedPhase := initialPhase + 0.1*5.0
	if math.Abs(c.WavePhase-expectedPhase) > 0.001 {
		t.Errorf("phase = %f, want %f", c.WavePhase, expectedPhase)
	}
}

func TestNewSystem(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)

	if s == nil {
		t.Fatal("NewSystem returned nil")
	}

	if !s.IsEnabled() {
		t.Error("system should be enabled by default")
	}

	if s.genre != "fantasy" {
		t.Errorf("genre = %s, want fantasy", s.genre)
	}

	preset := s.GetPreset()
	if preset.BaseIntensity <= 0 {
		t.Error("base intensity should be positive")
	}
}

func TestSystemGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			s := NewSystem(genre, 320, 200)
			preset := s.GetPreset()

			if preset.BaseIntensity <= 0 || preset.BaseIntensity > 2.0 {
				t.Errorf("BaseIntensity = %f, want in (0, 2]", preset.BaseIntensity)
			}
			if preset.WaveFrequency <= 0 || preset.WaveFrequency > 10.0 {
				t.Errorf("WaveFrequency = %f, want in (0, 10]", preset.WaveFrequency)
			}
			if preset.WaveAmplitude <= 0 || preset.WaveAmplitude > 20.0 {
				t.Errorf("WaveAmplitude = %f, want in (0, 20]", preset.WaveAmplitude)
			}
			if preset.VerticalBias < 0 || preset.VerticalBias > 1.0 {
				t.Errorf("VerticalBias = %f, want in [0, 1]", preset.VerticalBias)
			}
		})
	}
}

func TestSystemSetGenre(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	fantasyPreset := s.GetPreset()

	s.SetGenre("scifi")
	scifiPreset := s.GetPreset()

	// Presets should differ
	if fantasyPreset.WaveFrequency == scifiPreset.WaveFrequency &&
		fantasyPreset.VerticalBias == scifiPreset.VerticalBias {
		t.Error("genre presets should differ between fantasy and scifi")
	}
}

func TestSystemUnknownGenre(t *testing.T) {
	s := NewSystem("unknown_genre", 320, 200)

	// Should fall back to fantasy preset
	preset := s.GetPreset()
	fantasyPreset := genrePresets["fantasy"]

	if preset.BaseIntensity != fantasyPreset.BaseIntensity {
		t.Error("unknown genre should use fantasy defaults")
	}
}

func TestSystemAddSource(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	if s.GetSourceCount() != 0 {
		t.Error("source count should be 0 after clear")
	}

	s.AddSource(100, 100, 2.0, 0.5, 0, 1.0, 1.0, 1.0)

	if s.GetSourceCount() != 1 {
		t.Errorf("source count = %d, want 1", s.GetSourceCount())
	}
}

func TestSystemAddSourceZeroIntensity(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	// Zero intensity source should be ignored
	s.AddSource(100, 100, 2.0, 0, 0, 1.0, 1.0, 1.0)

	if s.GetSourceCount() != 0 {
		t.Error("zero-intensity source should not be added")
	}
}

func TestSystemAddSourceFromComponent(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	c := NewComponent(HeatTorch)
	c.SetScreenPosition(150, 100)

	s.AddSourceFromComponent(c)

	if s.GetSourceCount() != 1 {
		t.Errorf("source count = %d, want 1", s.GetSourceCount())
	}
}

func TestSystemDisabled(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	s.SetEnabled(false)

	// Source should not be added when disabled
	s.AddSource(100, 100, 2.0, 0.5, 0, 1.0, 1.0, 1.0)

	if s.GetSourceCount() != 0 {
		t.Error("disabled system should not accept sources")
	}
}

func TestSystemUpdate(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	initialTime := s.time

	s.Update(0.016) // ~60fps frame

	if s.time <= initialTime {
		t.Error("time should advance after Update")
	}
}

func TestSystemApplyNoSources(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	framebuffer := make([]byte, 320*200*4)
	for i := 0; i < len(framebuffer); i++ {
		framebuffer[i] = 128
	}

	original := make([]byte, len(framebuffer))
	copy(original, framebuffer)

	s.Apply(framebuffer)

	// With no sources, framebuffer should be unchanged
	for i := range framebuffer {
		if framebuffer[i] != original[i] {
			t.Error("framebuffer should be unchanged with no sources")
			break
		}
	}
}

func TestSystemApplyWithSource(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	// Add a source in the center
	s.AddSource(160, 100, 2.0, 1.0, 0, 1.1, 0.9, 0.8)

	// Create test framebuffer with gradient
	framebuffer := make([]byte, 320*200*4)
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			idx := (y*320 + x) * 4
			framebuffer[idx] = uint8(x % 256)
			framebuffer[idx+1] = uint8(y % 256)
			framebuffer[idx+2] = 128
			framebuffer[idx+3] = 255
		}
	}

	original := make([]byte, len(framebuffer))
	copy(original, framebuffer)

	s.Apply(framebuffer)

	// Some pixels near the heat source should be modified
	modified := false
	centerX, centerY := 160, 100
	radius := int(2.0 * s.pixelsPerUnit)

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			x, y := centerX+dx, centerY+dy
			if x < 0 || x >= 320 || y < 0 || y >= 200 {
				continue
			}
			idx := (y*320 + x) * 4
			if framebuffer[idx] != original[idx] ||
				framebuffer[idx+1] != original[idx+1] ||
				framebuffer[idx+2] != original[idx+2] {
				modified = true
				break
			}
		}
		if modified {
			break
		}
	}

	if !modified {
		t.Error("pixels near heat source should be modified")
	}
}

func TestFastSin(t *testing.T) {
	s := NewSystem("fantasy", 320, 200)

	// Test key angles
	tests := []struct {
		angle     float64
		expected  float64
		tolerance float64
	}{
		{0, 0, 0.05},
		{math.Pi / 2, 1, 0.05},
		{math.Pi, 0, 0.05},
		{3 * math.Pi / 2, -1, 0.05},
	}

	for _, tt := range tests {
		result := s.fastSin(tt.angle)
		if math.Abs(result-tt.expected) > tt.tolerance {
			t.Errorf("fastSin(%f) = %f, want %f (±%f)", tt.angle, result, tt.expected, tt.tolerance)
		}
	}
}

func TestRenderDistortionPreview(t *testing.T) {
	s := NewSystem("fantasy", 100, 100)
	s.ClearSources()
	s.AddSource(50, 50, 1.0, 0.8, 0, 1.1, 0.9, 0.8)

	img := s.RenderDistortionPreview(100, 100)

	if img == nil {
		t.Fatal("preview should not be nil")
	}

	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
		t.Errorf("preview size = (%d, %d), want (100, 100)",
			img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Check that center (near heat source) has modified colors
	centerPixel := img.RGBAAt(50, 50)
	edgePixel := img.RGBAAt(0, 0)

	if centerPixel.R <= edgePixel.R {
		t.Error("center pixel should be brighter (warmer) than edge")
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		v, min, max, want int
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		if got := clampInt(tt.v, tt.min, tt.max); got != tt.want {
			t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{0.5, 0, 1, 0.5},
		{-0.5, 0, 1, 0},
		{1.5, 0, 1, 1},
		{0, 0, 1, 0},
		{1, 0, 1, 1},
	}

	for _, tt := range tests {
		if got := clampFloat(tt.v, tt.min, tt.max); got != tt.want {
			t.Errorf("clampFloat(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func BenchmarkSystemApply(b *testing.B) {
	s := NewSystem("fantasy", 320, 200)
	s.ClearSources()

	// Add several heat sources
	s.AddSource(80, 50, 2.0, 0.7, 0, 1.1, 0.9, 0.8)
	s.AddSource(160, 100, 2.5, 0.8, 0.5, 1.15, 0.85, 0.75)
	s.AddSource(240, 150, 1.5, 0.6, 1.0, 1.05, 0.95, 0.9)

	framebuffer := make([]byte, 320*200*4)
	for i := 0; i < len(framebuffer); i++ {
		framebuffer[i] = 128
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Apply(framebuffer)
	}
}

func BenchmarkFastSin(b *testing.B) {
	s := NewSystem("fantasy", 320, 200)
	angle := 0.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.fastSin(angle)
		angle += 0.1
	}
}
