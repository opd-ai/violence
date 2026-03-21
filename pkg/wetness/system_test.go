package wetness

import (
	"image"
	"image/color"
	"math/rand"
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
		{"unknown_falls_back", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID, 320, 200)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.screenW != 320 || sys.screenH != 200 {
				t.Errorf("Screen size = (%d, %d), want (320, 200)", sys.screenW, sys.screenH)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Add something to cache
	sys.cache[cacheKey{moisture: 0.5}] = nil

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "horror")
	}

	// Cache should be cleared
	if len(sys.cache) != 0 {
		t.Errorf("Cache should be cleared on genre change, got %d entries", len(sys.cache))
	}
}

func TestGenerateWetnessPattern(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Test with valid tiles
	tiles := [][]int{
		{0, 0, 0, 10, 10},
		{0, 0, 0, 10, 10},
		{0, 0, 0, 0, 0},
		{10, 10, 0, 0, 0},
		{10, 10, 0, 0, 0},
	}

	pattern := sys.GenerateWetnessPattern(tiles, 12345)

	if pattern == nil {
		t.Fatal("GenerateWetnessPattern returned nil")
	}

	if pattern.Width != 5 || pattern.Height != 5 {
		t.Errorf("Pattern size = (%d, %d), want (5, 5)", pattern.Width, pattern.Height)
	}

	if pattern.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", pattern.Seed)
	}

	// Check that wall tiles don't have wetness
	for y := 0; y < 2; y++ {
		for x := 3; x < 5; x++ {
			if pattern.Cells[y][x] != nil {
				t.Errorf("Wall tile at (%d, %d) should not have wetness", x, y)
			}
		}
	}
}

func TestGenerateWetnessPatternDeterminism(t *testing.T) {
	sys := NewSystem("cyberpunk", 320, 200)

	tiles := [][]int{
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
	}

	pattern1 := sys.GenerateWetnessPattern(tiles, 54321)
	pattern2 := sys.GenerateWetnessPattern(tiles, 54321)

	// Same seed should produce same pattern
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			m1 := pattern1.GetMoistureAt(x, y)
			m2 := pattern2.GetMoistureAt(x, y)
			if m1 != m2 {
				t.Errorf("Non-deterministic at (%d, %d): %f vs %f", x, y, m1, m2)
			}
		}
	}
}

func TestGenerateWetnessPatternEmpty(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Empty tiles
	pattern := sys.GenerateWetnessPattern([][]int{}, 12345)
	if pattern != nil {
		t.Error("Empty tiles should return nil pattern")
	}

	// Tiles with empty row
	pattern = sys.GenerateWetnessPattern([][]int{{}}, 12345)
	if pattern != nil {
		t.Error("Tiles with empty row should return nil pattern")
	}
}

func TestIsFloorTile(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	tests := []struct {
		tileValue int
		isFloor   bool
	}{
		{0, true},
		{1, true},
		{9, true},
		{10, false},
		{100, false},
		{-1, false},
	}

	for _, tt := range tests {
		got := sys.isFloorTile(tt.tileValue)
		if got != tt.isFloor {
			t.Errorf("isFloorTile(%d) = %v, want %v", tt.tileValue, got, tt.isFloor)
		}
	}
}

func TestCountAdjacentWalls(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	tiles := [][]int{
		{10, 10, 10},
		{10, 0, 10},
		{10, 10, 10},
	}

	// Center tile surrounded by walls
	count := sys.countAdjacentWalls(tiles, 1, 1)
	if count != 4 {
		t.Errorf("countAdjacentWalls for surrounded tile = %d, want 4", count)
	}

	tiles2 := [][]int{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	// No adjacent walls
	count = sys.countAdjacentWalls(tiles2, 1, 1)
	if count != 0 {
		t.Errorf("countAdjacentWalls for open tile = %d, want 0", count)
	}
}

func TestIsCorner(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// L-shaped corner
	tiles := [][]int{
		{10, 10, 0, 0},
		{10, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	// Position (1,1) is in a corner (wall above and to the left)
	if !sys.isCorner(tiles, 1, 1) {
		t.Error("(1,1) should be a corner")
	}

	// Position (2,2) is not a corner (surrounded by floor)
	if sys.isCorner(tiles, 2, 2) {
		t.Error("(2,2) should not be a corner")
	}
}

func TestGenerateNoiseMap(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	noise := sys.generateNoiseMap(10, 10, 12345)

	if len(noise) != 10 {
		t.Errorf("Noise map height = %d, want 10", len(noise))
	}

	if len(noise[0]) != 10 {
		t.Errorf("Noise map width = %d, want 10", len(noise[0]))
	}

	// Check values are in valid range
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if noise[y][x] < 0 || noise[y][x] > 1 {
				t.Errorf("Noise value at (%d, %d) = %f, want in [0, 1]", x, y, noise[y][x])
			}
		}
	}

	// Test determinism
	noise2 := sys.generateNoiseMap(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if noise[y][x] != noise2[y][x] {
				t.Errorf("Noise not deterministic at (%d, %d)", x, y)
			}
		}
	}
}

func TestSampleNoise(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Test determinism
	v1 := sys.sampleNoise(5, 5, 12345)
	v2 := sys.sampleNoise(5, 5, 12345)
	if v1 != v2 {
		t.Errorf("sampleNoise not deterministic: %f vs %f", v1, v2)
	}

	// Test range
	for i := 0; i < 100; i++ {
		v := sys.sampleNoise(i, i, 12345)
		if v < 0 || v > 1 {
			t.Errorf("sampleNoise(%d, %d) = %f, want in [0, 1]", i, i, v)
		}
	}

	// Different positions should give different values
	v3 := sys.sampleNoise(10, 10, 12345)
	if v1 == v3 {
		t.Log("Warning: Same noise value for different positions (possible but unlikely)")
	}
}

func TestCalculateSpecular(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Non-puddle should return 0
	dryComp := NewComponent(0, 0, 0.3, 12345)
	spec := sys.calculateSpecular(dryComp, 100, 100, []LightSource{})
	if spec != 0 {
		t.Errorf("Specular for non-puddle = %f, want 0", spec)
	}

	// Puddle with no lights should return 0
	puddleComp := NewComponent(0, 0, 0.8, 12345)
	spec = sys.calculateSpecular(puddleComp, 100, 100, nil)
	if spec != 0 {
		t.Errorf("Specular with no lights = %f, want 0", spec)
	}

	// Puddle with nearby light
	lights := []LightSource{
		{X: 100, Y: 100, Radius: 200, Intensity: 1.0, R: 1, G: 1, B: 1},
	}
	spec = sys.calculateSpecular(puddleComp, 100-float64(sys.tileSize)/2, 100-float64(sys.tileSize)/2, lights)
	if spec <= 0 {
		t.Errorf("Specular with nearby light = %f, want > 0", spec)
	}

	// Puddle with far light
	farLights := []LightSource{
		{X: 1000, Y: 1000, Radius: 100, Intensity: 1.0},
	}
	spec = sys.calculateSpecular(puddleComp, 0, 0, farLights)
	if spec != 0 {
		t.Errorf("Specular with far light = %f, want 0", spec)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, expected float64
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		got := clamp(tt.v, tt.min, tt.max)
		if got != tt.expected {
			t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.expected)
		}
	}
}

func TestGenerateWetSprite(t *testing.T) {
	sys := NewSystem("cyberpunk", 320, 200)

	// Test damp surface
	dampComp := NewComponent(5, 5, 0.4, 12345)
	sprite := sys.generateWetSprite(dampComp)

	if sprite == nil {
		t.Fatal("generateWetSprite returned nil for damp surface")
	}

	bounds := sprite.Bounds()
	if bounds.Dx() != sys.tileSize || bounds.Dy() != sys.tileSize {
		t.Errorf("Sprite size = (%d, %d), want (%d, %d)",
			bounds.Dx(), bounds.Dy(), sys.tileSize, sys.tileSize)
	}

	// Test puddle
	puddleComp := NewComponent(5, 5, 0.9, 12345)
	puddleSprite := sys.generateWetSprite(puddleComp)

	if puddleSprite == nil {
		t.Fatal("generateWetSprite returned nil for puddle")
	}
}

func TestGetWetSpriteCache(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	comp := NewComponent(5, 5, 0.7, 12345)

	// First call should generate and cache
	sprite1 := sys.getWetSprite(comp)
	if sprite1 == nil {
		t.Fatal("First getWetSprite returned nil")
	}

	cacheSize := len(sys.cache)
	if cacheSize != 1 {
		t.Errorf("Cache size after first call = %d, want 1", cacheSize)
	}

	// Second call with same params should return cached
	sprite2 := sys.getWetSprite(comp)
	if sprite2 != sprite1 {
		t.Error("Second call should return cached sprite")
	}

	if len(sys.cache) != 1 {
		t.Error("Cache should not grow for duplicate request")
	}
}

func TestGetWetSpriteCacheLimit(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)
	sys.cacheLimit = 5

	// Fill cache beyond limit
	for i := 0; i < 10; i++ {
		comp := NewComponent(i, i, 0.5+float64(i)*0.05, int64(i))
		sys.getWetSprite(comp)
	}

	if len(sys.cache) > sys.cacheLimit {
		t.Errorf("Cache size = %d, should not exceed limit %d", len(sys.cache), sys.cacheLimit)
	}
}

func TestRenderWetness(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	tiles := [][]int{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	pattern := sys.GenerateWetnessPattern(tiles, 12345)

	// Create a test screen
	screen := ebiten.NewImage(320, 200)

	// Should not panic with valid inputs
	sys.RenderWetness(screen, pattern, nil, 0, 0)

	// Should not panic with nil pattern
	sys.RenderWetness(screen, nil, nil, 0, 0)

	// Should not panic with lights
	lights := []LightSource{
		{X: 50, Y: 50, Radius: 100, Intensity: 1.0},
	}
	sys.RenderWetness(screen, pattern, lights, 0, 0)
}

func TestAddSpecularHighlights(t *testing.T) {
	sys := NewSystem("cyberpunk", 320, 200)

	img := image.NewRGBA(image.Rect(0, 0, sys.tileSize, sys.tileSize))

	// Fill with base wet color with some alpha
	for y := 0; y < sys.tileSize; y++ {
		for x := 0; x < sys.tileSize; x++ {
			img.Set(x, y, color.RGBA{R: 20, G: 30, B: 40, A: 100})
		}
	}

	comp := NewComponent(0, 0, 0.9, 12345)

	// Create an rng for the test
	rng := rand.New(rand.NewSource(12345))

	// Add specular highlights
	sys.addSpecularHighlights(img, comp, rng)

	// Should not panic, highlights are visual only
}

func TestGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset, ok := genrePresets[genre]
			if !ok {
				t.Fatalf("Missing preset for genre %q", genre)
			}

			// Validate preset values
			if preset.BaseMoisture < 0 || preset.BaseMoisture > 1 {
				t.Errorf("BaseMoisture = %f, want in [0, 1]", preset.BaseMoisture)
			}
			if preset.PuddleDensity < 0 || preset.PuddleDensity > 1 {
				t.Errorf("PuddleDensity = %f, want in [0, 1]", preset.PuddleDensity)
			}
			if preset.SpecularStrength < 0 || preset.SpecularStrength > 1 {
				t.Errorf("SpecularStrength = %f, want in [0, 1]", preset.SpecularStrength)
			}
		})
	}
}

func BenchmarkGenerateWetnessPattern(b *testing.B) {
	sys := NewSystem("fantasy", 320, 200)

	tiles := make([][]int, 50)
	for y := range tiles {
		tiles[y] = make([]int, 50)
		for x := range tiles[y] {
			if x == 0 || y == 0 || x == 49 || y == 49 {
				tiles[y][x] = 10 // Wall
			} else {
				tiles[y][x] = 0 // Floor
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GenerateWetnessPattern(tiles, int64(i))
	}
}

func BenchmarkGenerateWetSprite(b *testing.B) {
	sys := NewSystem("cyberpunk", 320, 200)
	comp := NewComponent(5, 5, 0.8, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.generateWetSprite(comp)
	}
}

func BenchmarkCalculateSpecular(b *testing.B) {
	sys := NewSystem("fantasy", 320, 200)
	comp := NewComponent(5, 5, 0.9, 12345)
	lights := []LightSource{
		{X: 50, Y: 50, Radius: 100, Intensity: 1.0},
		{X: 150, Y: 50, Radius: 80, Intensity: 0.8},
		{X: 100, Y: 100, Radius: 120, Intensity: 0.9},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.calculateSpecular(comp, 100, 100, lights)
	}
}
