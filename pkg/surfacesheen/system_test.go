package surfacesheen

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"unknown defaults to fantasy", "unknowngenre"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.logger == nil {
				t.Error("logger is nil")
			}
			// Check light direction is normalized
			length := math.Sqrt(sys.lightDirX*sys.lightDirX + sys.lightDirY*sys.lightDirY + sys.lightDirZ*sys.lightDirZ)
			if math.Abs(length-1.0) > 0.001 {
				t.Errorf("light direction not normalized: length = %v", length)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	genres := []string{"scifi", "horror", "cyberpunk", "postapoc", "fantasy"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.GetGenre() != genre {
			t.Errorf("SetGenre(%s): GetGenre() = %s", genre, sys.GetGenre())
		}
	}
}

func TestGenrePresets(t *testing.T) {
	// Verify all expected genres have presets
	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		preset, ok := genrePresets[genre]
		if !ok {
			t.Errorf("missing preset for genre: %s", genre)
			continue
		}

		// Validate preset values are within reasonable ranges
		if preset.BaseIntensity < 0 || preset.BaseIntensity > 2 {
			t.Errorf("%s: BaseIntensity out of range: %v", genre, preset.BaseIntensity)
		}
		if preset.WarmthShift < -1 || preset.WarmthShift > 1 {
			t.Errorf("%s: WarmthShift out of range: %v", genre, preset.WarmthShift)
		}
		if preset.SpecularTightness < 0.1 || preset.SpecularTightness > 10 {
			t.Errorf("%s: SpecularTightness out of range: %v", genre, preset.SpecularTightness)
		}
		if preset.WetSheenBoost < 0 || preset.WetSheenBoost > 5 {
			t.Errorf("%s: WetSheenBoost out of range: %v", genre, preset.WetSheenBoost)
		}
		if preset.MetalSaturation < 0 || preset.MetalSaturation > 1 {
			t.Errorf("%s: MetalSaturation out of range: %v", genre, preset.MetalSaturation)
		}
	}
}

func TestCalculateSheenForEntity_NoLights(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})

	sheenColor, intensity := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, nil)

	if intensity != 0 {
		t.Errorf("intensity with no lights = %v, want 0", intensity)
	}
	if sheenColor != (color.RGBA{}) {
		t.Errorf("sheenColor with no lights = %v, want zero", sheenColor)
	}
}

func TestCalculateSheenForEntity_NilComponent(t *testing.T) {
	sys := NewSystem("fantasy")
	lights := []LightSource{
		{X: 100, Y: 100, Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, Intensity: 1.0, Radius: 100},
	}

	_, intensity := sys.CalculateSheenForEntity(nil, 100, 100, 1.0, lights)

	if intensity != 0 {
		t.Errorf("intensity with nil component = %v, want 0", intensity)
	}
}

func TestCalculateSheenForEntity_ZeroIntensity(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})
	comp.Intensity = 0

	lights := []LightSource{
		{X: 100, Y: 100, Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, Intensity: 1.0, Radius: 100},
	}

	_, intensity := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, lights)

	if intensity != 0 {
		t.Errorf("intensity with zero component intensity = %v, want 0", intensity)
	}
}

func TestCalculateSheenForEntity_WithLight(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})

	lights := []LightSource{
		{X: 105, Y: 100, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 50},
	}

	_, intensity := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, lights)

	if intensity <= 0 {
		t.Errorf("expected positive intensity with nearby light, got %v", intensity)
	}
}

func TestCalculateSheenForEntity_DistantLight(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})

	lights := []LightSource{
		{X: 1000, Y: 1000, Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, Intensity: 1.0, Radius: 50},
	}

	_, intensity := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, lights)

	// Very distant light should produce minimal sheen
	if intensity > 0.1 {
		t.Errorf("distant light produced too much sheen: %v", intensity)
	}
}

func TestCalculateSheenForEntity_WetSurface(t *testing.T) {
	sys := NewSystem("fantasy")

	compDry := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})
	compDry.Wetness = 0

	compWet := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})
	compWet.Wetness = 1.0

	lights := []LightSource{
		{X: 110, Y: 100, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 50},
	}

	_, intensityDry := sys.CalculateSheenForEntity(compDry, 100, 100, 1.0, lights)
	_, intensityWet := sys.CalculateSheenForEntity(compWet, 100, 100, 1.0, lights)

	// Wet surface should have higher intensity
	if intensityWet <= intensityDry {
		t.Errorf("wet surface intensity (%v) should be greater than dry (%v)", intensityWet, intensityDry)
	}
}

func TestCalculateSheenForEntity_AllMaterials(t *testing.T) {
	sys := NewSystem("fantasy")

	lights := []LightSource{
		{X: 110, Y: 100, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 50},
	}

	materials := []MaterialType{
		MaterialMetal,
		MaterialWet,
		MaterialPolished,
		MaterialOrganic,
		MaterialCloth,
		MaterialCrystal,
		MaterialDefault,
	}

	for _, mat := range materials {
		t.Run(mat.String(), func(t *testing.T) {
			comp := NewSheenComponent(mat, color.RGBA{R: 150, G: 150, B: 180, A: 255})
			sheenColor, intensity := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, lights)

			// All materials should produce some sheen with a nearby light
			if intensity <= 0 {
				t.Errorf("material %s produced no sheen", mat.String())
			}

			// Color should be valid
			if sheenColor.A != 255 {
				t.Errorf("material %s produced invalid alpha: %d", mat.String(), sheenColor.A)
			}
		})
	}
}

func TestCalculateSpecular(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		name      string
		material  MaterialType
		roughness float64
	}{
		{"metal low roughness", MaterialMetal, 0.1},
		{"metal high roughness", MaterialMetal, 0.9},
		{"polished", MaterialPolished, 0.05},
		{"cloth", MaterialCloth, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specular := sys.calculateSpecular(tt.material, 0.5, -0.5, tt.roughness, sys.preset)

			// Specular should be in valid range
			if specular < 0 || specular > 10 {
				t.Errorf("specular out of range: %v", specular)
			}
		})
	}
}

func TestCalculateMaterialColor_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})
	light := LightSource{
		X: 100, Y: 100,
		Color:     color.RGBA{R: 255, G: 240, B: 200, A: 255},
		Intensity: 1.0,
		Radius:    50,
	}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)
			sheenColor := sys.calculateMaterialColor(comp, light, sys.preset)

			// Color should be valid
			if sheenColor.A != 255 {
				t.Errorf("invalid alpha for %s: %d", genre, sheenColor.A)
			}
		})
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t  float64
		expected float64
	}{
		{0, 100, 0, 0},
		{0, 100, 1, 100},
		{0, 100, 0.5, 50},
		{100, 200, 0.25, 125},
	}

	for _, tt := range tests {
		got := lerp(tt.a, tt.b, tt.t)
		if math.Abs(got-tt.expected) > 0.001 {
			t.Errorf("lerp(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.expected)
		}
	}
}

func TestClampUint8(t *testing.T) {
	tests := []struct {
		input    float64
		expected uint8
	}{
		{0, 0},
		{255, 255},
		{-10, 0},
		{300, 255},
		{127.5, 127},
	}

	for _, tt := range tests {
		got := clampUint8(tt.input)
		if got != tt.expected {
			t.Errorf("clampUint8(%v) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestEnsureOverlay(t *testing.T) {
	sys := NewSystem("fantasy")

	// First call should create overlay
	sys.ensureOverlay(320, 200)
	if sys.overlay == nil {
		t.Fatal("overlay not created")
	}
	if sys.overlayW != 320 || sys.overlayH != 200 {
		t.Errorf("overlay dimensions = %dx%d, want 320x200", sys.overlayW, sys.overlayH)
	}

	// Same dimensions should not recreate
	oldOverlay := sys.overlay
	sys.ensureOverlay(320, 200)
	if sys.overlay != oldOverlay {
		t.Error("overlay recreated for same dimensions")
	}

	// Different dimensions should recreate
	sys.ensureOverlay(640, 400)
	if sys.overlayW != 640 || sys.overlayH != 400 {
		t.Errorf("overlay dimensions after resize = %dx%d, want 640x400", sys.overlayW, sys.overlayH)
	}
}

func TestRenderSheenOverlay_EmptyInputs(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(320, 200)
	defer screen.Dispose()

	// Empty components - should not panic
	sys.RenderSheenOverlay(screen, nil, nil, nil, nil, nil, nil, nil)
	sys.RenderSheenOverlay(screen, []*SheenComponent{}, []float64{}, []float64{}, []float64{}, []float64{}, []float64{}, nil)

	// Empty lights - should not panic
	comps := []*SheenComponent{NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})}
	sys.RenderSheenOverlay(screen, comps, []float64{100}, []float64{100}, []float64{16}, []float64{100}, []float64{100}, nil)
}

func TestRenderSheenOverlay_WithData(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(320, 200)
	defer screen.Dispose()

	comps := []*SheenComponent{
		NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255}),
		NewSheenComponent(MaterialWet, color.RGBA{R: 100, G: 100, B: 150, A: 255}),
	}
	screenX := []float64{100, 200}
	screenY := []float64{100, 100}
	radiusPx := []float64{16, 16}
	worldX := []float64{3.0, 6.0}
	worldY := []float64{3.0, 3.0}
	lights := []LightSource{
		{X: 4.0, Y: 3.0, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 5.0},
	}

	// Should not panic
	sys.RenderSheenOverlay(screen, comps, screenX, screenY, radiusPx, worldX, worldY, lights)
}

func TestRenderSheenOverlay_MismatchedLengths(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(320, 200)
	defer screen.Dispose()

	// Mismatched array lengths - should handle gracefully
	comps := []*SheenComponent{
		NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255}),
		NewSheenComponent(MaterialWet, color.RGBA{R: 100, G: 100, B: 150, A: 255}),
		NewSheenComponent(MaterialPolished, color.RGBA{R: 255, G: 255, B: 255, A: 255}),
	}
	screenX := []float64{100} // Only one position
	screenY := []float64{100}
	radiusPx := []float64{16, 16}
	worldX := []float64{3.0}
	worldY := []float64{3.0}
	lights := []LightSource{
		{X: 4.0, Y: 3.0, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 5.0},
	}

	// Should not panic with mismatched lengths
	sys.RenderSheenOverlay(screen, comps, screenX, screenY, radiusPx, worldX, worldY, lights)
}

func TestDrawSheenSpot(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.ensureOverlay(320, 200)

	// Should not panic
	sys.drawSheenSpot(100, 100, 16, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)

	// Edge cases
	sys.drawSheenSpot(0, 0, 16, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
	sys.drawSheenSpot(319, 199, 16, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
	sys.drawSheenSpot(-10, -10, 16, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
	sys.drawSheenSpot(500, 500, 16, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
}

func TestDrawSheenSpot_SmallRadius(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.ensureOverlay(320, 200)

	// Very small radius should still work
	sys.drawSheenSpot(100, 100, 0.5, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
	sys.drawSheenSpot(100, 100, 1.0, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
}

func TestNilOverlay(t *testing.T) {
	sys := NewSystem("fantasy")
	// overlay is nil before ensureOverlay is called

	// drawSheenSpot should handle nil overlay gracefully
	sys.drawSheenSpot(100, 100, 16, color.RGBA{R: 255, G: 240, B: 200, A: 255}, 0.5)
}

func TestMultipleLightSources(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})

	// Single light
	singleLight := []LightSource{
		{X: 105, Y: 100, Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, Intensity: 1.0, Radius: 50},
	}
	_, intensitySingle := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, singleLight)

	// Multiple lights
	multipleLights := []LightSource{
		{X: 105, Y: 100, Color: color.RGBA{R: 255, G: 200, B: 200, A: 255}, Intensity: 1.0, Radius: 50},
		{X: 95, Y: 105, Color: color.RGBA{R: 200, G: 200, B: 255, A: 255}, Intensity: 1.0, Radius: 50},
	}
	_, intensityMultiple := sys.CalculateSheenForEntity(comp, 100, 100, 1.0, multipleLights)

	// Multiple lights should produce more sheen (up to cap)
	if intensityMultiple < intensitySingle {
		t.Errorf("multiple lights (%v) should produce >= sheen than single (%v)", intensityMultiple, intensitySingle)
	}
}

func BenchmarkCalculateSheenForEntity(b *testing.B) {
	sys := NewSystem("fantasy")
	comp := NewSheenComponent(MaterialMetal, color.RGBA{R: 180, G: 180, B: 200, A: 255})
	lights := []LightSource{
		{X: 110, Y: 100, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 50},
		{X: 90, Y: 110, Color: color.RGBA{R: 200, G: 200, B: 255, A: 255}, Intensity: 0.8, Radius: 40},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CalculateSheenForEntity(comp, 100, 100, 1.0, lights)
	}
}

func BenchmarkRenderSheenOverlay(b *testing.B) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(320, 200)
	defer screen.Dispose()

	// Create test data
	comps := make([]*SheenComponent, 20)
	screenX := make([]float64, 20)
	screenY := make([]float64, 20)
	radiusPx := make([]float64, 20)
	worldX := make([]float64, 20)
	worldY := make([]float64, 20)

	for i := 0; i < 20; i++ {
		comps[i] = NewSheenComponent(MaterialType(i%7), color.RGBA{R: 180, G: 180, B: 200, A: 255})
		screenX[i] = float64(i * 16)
		screenY[i] = float64(100)
		radiusPx[i] = 16
		worldX[i] = float64(i)
		worldY[i] = 3.0
	}

	lights := []LightSource{
		{X: 5.0, Y: 3.0, Color: color.RGBA{R: 255, G: 240, B: 200, A: 255}, Intensity: 1.0, Radius: 10},
		{X: 15.0, Y: 3.0, Color: color.RGBA{R: 200, G: 220, B: 255, A: 255}, Intensity: 0.8, Radius: 8},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.RenderSheenOverlay(screen, comps, screenX, screenY, radiusPx, worldX, worldY, lights)
	}
}
