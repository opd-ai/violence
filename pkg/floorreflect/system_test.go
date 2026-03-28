package floorreflect

import (
	"image"
	"image/color"
	"testing"
)

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	if comp.Type() != "floorreflect" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "floorreflect")
	}
}

func TestNewComponent(t *testing.T) {
	comp := NewComponent()

	if comp.IntensityOverride != -1.0 {
		t.Errorf("IntensityOverride = %f, want -1.0", comp.IntensityOverride)
	}
	if !comp.Enabled {
		t.Error("Enabled should be true by default")
	}
	if comp.cacheValid {
		t.Error("cacheValid should be false initially")
	}
}

func TestSetSourceImage(t *testing.T) {
	comp := NewComponent()
	comp.cacheValid = true

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	comp.SetSourceImage(img)

	if comp.SourceImage != img {
		t.Error("SourceImage not set correctly")
	}
	if comp.cacheValid {
		t.Error("Cache should be invalidated after SetSourceImage")
	}
}

func TestSetPosition(t *testing.T) {
	comp := NewComponent()
	comp.SetPosition(100.5, 200.75)

	if comp.X != 100.5 {
		t.Errorf("X = %f, want 100.5", comp.X)
	}
	if comp.Y != 200.75 {
		t.Errorf("Y = %f, want 200.75", comp.Y)
	}
}

func TestSetIntensity(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 0.5, 0.5},
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
		{"negative", -0.5, -1.0},
		{"above one", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent()
			comp.SetIntensity(tt.input)
			if comp.IntensityOverride != tt.expected {
				t.Errorf("IntensityOverride = %f, want %f", comp.IntensityOverride, tt.expected)
			}
		})
	}
}

func TestNewSystem(t *testing.T) {
	tests := []struct {
		genreID          string
		wantReflectivity float64
	}{
		{"fantasy", 0.25},
		{"scifi", 0.5},
		{"horror", 0.35},
		{"cyberpunk", 0.6},
		{"postapoc", 0.3},
		{"unknown", 0.25}, // Falls back to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			sys := NewSystem(tt.genreID)
			if sys.preset.DefaultReflectivity != tt.wantReflectivity {
				t.Errorf("DefaultReflectivity = %f, want %f",
					sys.preset.DefaultReflectivity, tt.wantReflectivity)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "cyberpunk")
	}
	if sys.preset.DefaultReflectivity != 0.6 {
		t.Errorf("DefaultReflectivity = %f, want 0.6", sys.preset.DefaultReflectivity)
	}

	// Test fallback
	sys.SetGenre("invalid")
	if sys.preset.DefaultReflectivity != 0.25 {
		t.Errorf("Should fall back to fantasy preset")
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(640, 480)

	if sys.screenWidth != 640 {
		t.Errorf("screenWidth = %d, want 640", sys.screenWidth)
	}
	if sys.screenHeight != 480 {
		t.Errorf("screenHeight = %d, want 480", sys.screenHeight)
	}
	if sys.reflectionBuffer == nil {
		t.Error("reflectionBuffer should be allocated")
	}
}

func TestAddReflectiveFloor(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.AddReflectiveFloor(5, 10, ReflectWater, 0.8)

	key := 5*10000 + 10
	tile, ok := sys.reflectiveFloors[key]
	if !ok {
		t.Fatal("Reflective floor not added")
	}
	if tile.TileX != 5 || tile.TileY != 10 {
		t.Errorf("Tile position = (%d, %d), want (5, 10)", tile.TileX, tile.TileY)
	}
	if tile.LightLevel != 0.8 {
		t.Errorf("LightLevel = %f, want 0.8", tile.LightLevel)
	}
}

func TestClearReflectiveFloors(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.AddReflectiveFloor(1, 1, ReflectMetal, 1.0)
	sys.AddReflectiveFloor(2, 2, ReflectWater, 0.5)

	if len(sys.reflectiveFloors) != 2 {
		t.Errorf("Should have 2 floors, got %d", len(sys.reflectiveFloors))
	}

	sys.ClearReflectiveFloors()

	if len(sys.reflectiveFloors) != 0 {
		t.Errorf("Should have 0 floors after clear, got %d", len(sys.reflectiveFloors))
	}
}

func TestIsFloorReflective(t *testing.T) {
	sys := NewSystem("fantasy")
	tileSize := 32

	// Add a reflective water tile at (3, 4)
	sys.AddReflectiveFloor(3, 4, ReflectWater, 0.7)

	tests := []struct {
		name       string
		worldX     float64
		worldY     float64
		wantRefl   bool
		wantRefVal float64
	}{
		{"on water tile", 100, 140, true, ReflectWater.Reflectivity},
		{"off reflective area", 0, 0, false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reflective, material := sys.IsFloorReflective(tt.worldX, tt.worldY, tileSize)
			if reflective != tt.wantRefl {
				t.Errorf("IsFloorReflective = %v, want %v", reflective, tt.wantRefl)
			}
			if tt.wantRefl && material.Reflectivity != tt.wantRefVal {
				t.Errorf("Reflectivity = %f, want %f", material.Reflectivity, tt.wantRefVal)
			}
		})
	}
}

func TestGetFloorReflectivity(t *testing.T) {
	sys := NewSystem("scifi") // Default reflectivity 0.5
	tileSize := 32

	sys.AddReflectiveFloor(2, 2, ReflectMetal, 1.0)

	// On metal tile
	refl := sys.GetFloorReflectivity(70, 70, tileSize)
	if refl != ReflectMetal.Reflectivity {
		t.Errorf("Reflectivity on metal = %f, want %f", refl, ReflectMetal.Reflectivity)
	}

	// Off reflective area - should return default
	refl = sys.GetFloorReflectivity(0, 0, tileSize)
	if refl != sys.preset.DefaultReflectivity {
		t.Errorf("Default reflectivity = %f, want %f", refl, sys.preset.DefaultReflectivity)
	}
}

func TestGenerateReflection(t *testing.T) {
	sys := NewSystem("fantasy")

	// Create a simple test image
	source := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Fill with solid red
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			source.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	reflection := sys.GenerateReflection(source, ReflectWater, 0.8)

	if reflection == nil {
		t.Fatal("GenerateReflection returned nil")
	}

	bounds := reflection.Bounds()
	if bounds.Dx() != 16 {
		t.Errorf("Reflection width = %d, want 16", bounds.Dx())
	}

	// Height should be capped by MaxReflectionHeight if source is larger
	if bounds.Dy() > sys.preset.MaxReflectionHeight {
		t.Errorf("Reflection height %d exceeds max %d", bounds.Dy(), sys.preset.MaxReflectionHeight)
	}

	// Check that reflection has some pixels with alpha > 0
	hasVisiblePixels := false
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c := reflection.At(x, y).(color.RGBA)
			if c.A > 0 {
				hasVisiblePixels = true
				break
			}
		}
		if hasVisiblePixels {
			break
		}
	}
	if !hasVisiblePixels {
		t.Error("Reflection has no visible pixels")
	}
}

func TestGenerateReflectionNilSource(t *testing.T) {
	sys := NewSystem("fantasy")
	reflection := sys.GenerateReflection(nil, ReflectWater, 0.5)
	if reflection != nil {
		t.Error("GenerateReflection with nil source should return nil")
	}
}

func TestGenerateReflectionFadeGradient(t *testing.T) {
	sys := NewSystem("fantasy")

	// Create test image
	source := image.NewRGBA(image.Rect(0, 0, 8, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 8; x++ {
			source.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	reflection := sys.GenerateReflection(source, MaterialReflectivity{
		Reflectivity: 1.0,
		TintR:        1.0, TintG: 1.0, TintB: 1.0,
		Distortion: 0.0,
		FadeRate:   1.0,
	}, 1.0)

	if reflection == nil {
		t.Fatal("Reflection is nil")
	}

	// Check that alpha decreases as Y increases (fade out)
	topAlpha := reflection.At(4, 0).(color.RGBA).A
	bottomAlpha := reflection.At(4, reflection.Bounds().Dy()-1).(color.RGBA).A

	if topAlpha <= bottomAlpha {
		t.Errorf("Top alpha (%d) should be > bottom alpha (%d) for fade effect",
			topAlpha, bottomAlpha)
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy")
	initialFrame := sys.frameCount

	sys.Update()
	sys.Update()
	sys.Update()

	if sys.frameCount != initialFrame+3 {
		t.Errorf("frameCount = %d, want %d", sys.frameCount, initialFrame+3)
	}
}

func TestSetSeed(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetSeed(12345)

	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng should not be nil after SetSeed")
	}
}

func TestMaterialReflectivityPresets(t *testing.T) {
	tests := []struct {
		name   string
		mat    MaterialReflectivity
		minRef float64
		maxRef float64
	}{
		{"metal", ReflectMetal, 0.5, 1.0},
		{"wet stone", ReflectWetStone, 0.3, 0.6},
		{"water", ReflectWater, 0.4, 0.7},
		{"oil slick", ReflectOilSlick, 0.5, 0.8},
		{"polished tile", ReflectPolishedTile, 0.4, 0.7},
		{"none", ReflectNone, 0.0, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mat.Reflectivity < tt.minRef || tt.mat.Reflectivity > tt.maxRef {
				t.Errorf("Reflectivity %f not in range [%f, %f]",
					tt.mat.Reflectivity, tt.minRef, tt.maxRef)
			}
		})
	}
}

func TestTintColor(t *testing.T) {
	tests := []struct {
		name                string
		input               color.RGBA
		tintR, tintG, tintB float64
		expected            color.RGBA
	}{
		{
			"no tint",
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
			1.0, 1.0, 1.0,
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
		},
		{
			"red tint",
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
			1.5, 1.0, 1.0,
			color.RGBA{R: 150, G: 100, B: 100, A: 255},
		},
		{
			"clamped overflow",
			color.RGBA{R: 200, G: 200, B: 200, A: 255},
			2.0, 2.0, 2.0,
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tintColor(tt.input, tt.tintR, tt.tintG, tt.tintB)
			if result.R != tt.expected.R || result.G != tt.expected.G ||
				result.B != tt.expected.B || result.A != tt.expected.A {
				t.Errorf("tintColor = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLerpColor(t *testing.T) {
	a := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	b := color.RGBA{R: 100, G: 200, B: 150, A: 255}

	mid := lerpColor(a, b, 0.5)
	if mid.R != 50 || mid.G != 100 || mid.B != 75 {
		t.Errorf("lerpColor(a, b, 0.5) = %v, want (50, 100, 75)", mid)
	}

	start := lerpColor(a, b, 0.0)
	if start.R != 0 || start.G != 0 || start.B != 0 {
		t.Errorf("lerpColor(a, b, 0.0) = %v, want (0, 0, 0)", start)
	}

	end := lerpColor(a, b, 1.0)
	if end.R != 100 || end.G != 200 || end.B != 150 {
		t.Errorf("lerpColor(a, b, 1.0) = %v, want (100, 200, 150)", end)
	}
}

func TestGetPreset(t *testing.T) {
	sys := NewSystem("horror")
	preset := sys.GetPreset()

	if preset.DefaultReflectivity != 0.35 {
		t.Errorf("Horror DefaultReflectivity = %f, want 0.35", preset.DefaultReflectivity)
	}
}

func TestGetGenreID(t *testing.T) {
	sys := NewSystem("cyberpunk")
	if sys.GetGenreID() != "cyberpunk" {
		t.Errorf("GetGenreID() = %q, want %q", sys.GetGenreID(), "cyberpunk")
	}
}

func TestGetReflectiveFloorCount(t *testing.T) {
	sys := NewSystem("fantasy")

	if sys.GetReflectiveFloorCount() != 0 {
		t.Error("Initial count should be 0")
	}

	sys.AddReflectiveFloor(1, 1, ReflectMetal, 1.0)
	sys.AddReflectiveFloor(2, 2, ReflectWater, 0.5)
	sys.AddReflectiveFloor(3, 3, ReflectOilSlick, 0.7)

	if sys.GetReflectiveFloorCount() != 3 {
		t.Errorf("Count = %d, want 3", sys.GetReflectiveFloorCount())
	}
}

func TestSetReflectiveFloorsFromWetness(t *testing.T) {
	sys := NewSystem("cyberpunk")

	wetTiles := map[int]float64{
		10001: 0.8,  // Heavy water
		20002: 0.5,  // Wet stone
		30003: 0.2,  // Light wetness
		40004: 0.05, // Too dry, should be skipped
	}

	sys.SetReflectiveFloorsFromWetness(wetTiles, 32)

	// Should have 3 reflective floors (0.05 is too low)
	if sys.GetReflectiveFloorCount() != 3 {
		t.Errorf("Count = %d, want 3 (one should be skipped)", sys.GetReflectiveFloorCount())
	}
}

func BenchmarkGenerateReflection(b *testing.B) {
	sys := NewSystem("fantasy")
	source := image.NewRGBA(image.Rect(0, 0, 32, 64))
	// Fill with test pattern
	for y := 0; y < 64; y++ {
		for x := 0; x < 32; x++ {
			source.Set(x, y, color.RGBA{R: uint8(x * 8), G: uint8(y * 4), B: 128, A: 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GenerateReflection(source, ReflectWater, 0.7)
	}
}

func BenchmarkIsFloorReflective(b *testing.B) {
	sys := NewSystem("fantasy")
	// Add many reflective floors
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			sys.AddReflectiveFloor(i, j, ReflectWater, 0.5)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.IsFloorReflective(float64(i%3200), float64(i%3200), 32)
	}
}
