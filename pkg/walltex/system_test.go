package walltex

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name         string
		genre        string
		maxCacheSize int
	}{
		{"fantasy", "fantasy", 100},
		{"scifi", "scifi", 50},
		{"horror", "horror", 200},
		{"cyberpunk", "cyberpunk", 100},
		{"postapoc", "postapoc", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genre, tt.maxCacheSize)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genre != tt.genre {
				t.Errorf("genre = %q, want %q", sys.genre, tt.genre)
			}
			if sys.maxCacheSize != tt.maxCacheSize {
				t.Errorf("maxCacheSize = %d, want %d", sys.maxCacheSize, tt.maxCacheSize)
			}
			if sys.generator == nil {
				t.Error("generator is nil")
			}
			if sys.textureCache == nil {
				t.Error("textureCache is nil")
			}
		})
	}
}

func TestGenerateWallTexture(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	seed := uint64(12345)

	tests := []struct {
		name     string
		gridX    int
		gridY    int
		roomType string
		depth    int
	}{
		{"corridor_shallow", 10, 20, "corridor", 1},
		{"room_mid", 5, 15, "room", 5},
		{"boss_deep", 30, 40, "boss", 10},
		{"treasure_mid", 15, 25, "treasure", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := sys.GenerateWallTexture(tt.gridX, tt.gridY, tt.roomType, tt.depth, seed)
			if comp == nil {
				t.Fatal("GenerateWallTexture returned nil")
			}
			if comp.GridX != tt.gridX {
				t.Errorf("GridX = %d, want %d", comp.GridX, tt.gridX)
			}
			if comp.GridY != tt.gridY {
				t.Errorf("GridY = %d, want %d", comp.GridY, tt.gridY)
			}
			if comp.CachedTexture == nil {
				t.Error("CachedTexture is nil")
			}
			if comp.Weathering < 0.0 || comp.Weathering > 1.0 {
				t.Errorf("Weathering out of range: %f", comp.Weathering)
			}

			// Verify texture dimensions
			bounds := comp.CachedTexture.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("texture size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestCaching(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	seed := uint64(12345)

	// Generate same texture twice
	comp1 := sys.GenerateWallTexture(10, 20, "corridor", 1, seed)
	comp2 := sys.GenerateWallTexture(10, 20, "corridor", 1, seed)

	// Second call should be cached
	hits, misses, size := sys.GetCacheStats()
	if hits < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", hits)
	}
	if misses < 1 {
		t.Errorf("Expected at least 1 cache miss, got %d", misses)
	}
	if size < 1 {
		t.Errorf("Expected cache size >= 1, got %d", size)
	}

	// Both components should reference the same cached texture
	if comp1.CachedTexture != comp2.CachedTexture {
		t.Error("Cache not reusing textures")
	}
}

func TestCacheEviction(t *testing.T) {
	sys := NewSystem("fantasy", 10) // Small cache
	seed := uint64(12345)

	// Generate more textures than cache can hold
	for i := 0; i < 20; i++ {
		sys.GenerateWallTexture(i, i, "corridor", 1, seed)
	}

	_, _, size := sys.GetCacheStats()
	if size > 10 {
		t.Errorf("Cache size %d exceeds max %d", size, 10)
	}
}

func TestSelectMaterial(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	seed := uint64(12345)

	tests := []struct {
		name     string
		roomType string
		depth    int
	}{
		{"corridor", "corridor", 1},
		{"room", "room", 3},
		{"boss", "boss", 10},
		{"treasure", "treasure", 5},
		{"unknown", "unknown_type", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			material := sys.selectMaterial(tt.roomType, tt.depth, seed)
			// Just verify it returns a valid material (no crash)
			if material < MaterialStone || material > MaterialTech {
				t.Errorf("invalid material value: %d", material)
			}
		})
	}
}

func TestCalculateWeathering(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	seed := uint64(12345)

	tests := []struct {
		name     string
		roomType string
		depth    int
	}{
		{"shallow", "corridor", 1},
		{"mid", "room", 5},
		{"deep", "boss", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weathering := sys.calculateWeathering(tt.roomType, tt.depth, seed)
			if weathering < 0.0 {
				t.Errorf("weathering < 0: %f", weathering)
			}
			if weathering > 1.0 {
				t.Errorf("weathering > 1: %f", weathering)
			}
		})
	}
}

func TestSampleTexture(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	seed := uint64(12345)
	comp := sys.GenerateWallTexture(10, 20, "corridor", 1, seed)

	tests := []struct {
		name string
		u    float64
		v    float64
	}{
		{"top_left", 0.0, 0.0},
		{"top_right", 1.0, 0.0},
		{"bottom_left", 0.0, 1.0},
		{"bottom_right", 1.0, 1.0},
		{"center", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := sys.SampleTexture(comp, tt.u, tt.v)
			// Just verify it returns a color without crashing
			if color.A == 0 {
				t.Error("sampled color has alpha = 0")
			}
		})
	}
}

func TestSampleTextureNil(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	comp := &WallTextureComponent{
		GridX:         10,
		GridY:         20,
		CachedTexture: nil, // Nil texture
	}

	color := sys.SampleTexture(comp, 0.5, 0.5)
	// Should return default gray without crashing
	if color.R != 128 || color.G != 128 || color.B != 128 {
		t.Errorf("expected gray (128,128,128), got (%d,%d,%d)", color.R, color.G, color.B)
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	genres := []string{"scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys.SetGenre(genre)
			if sys.genre != genre {
				t.Errorf("genre = %q, want %q", sys.genre, genre)
			}
			if sys.generator.genre != genre {
				t.Errorf("generator.genre = %q, want %q", sys.generator.genre, genre)
			}
		})
	}
}

func TestWallTextureComponentType(t *testing.T) {
	comp := &WallTextureComponent{}
	if comp.Type() != "WallTexture" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "WallTexture")
	}
}

func TestBuildMaterialRules(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			rules := buildMaterialRules(genre)
			if rules == nil {
				t.Fatal("buildMaterialRules returned nil")
			}

			// All rule sets should have default
			if _, ok := rules["default"]; !ok {
				t.Error("missing default rule")
			}

			// Check common room types
			for _, roomType := range []string{"corridor", "room", "boss", "treasure"} {
				if rule, ok := rules[roomType]; ok {
					// Validate rule ranges
					if rule.PrimaryChance < 0 || rule.PrimaryChance > 1 {
						t.Errorf("%s PrimaryChance out of range: %f", roomType, rule.PrimaryChance)
					}
					if rule.SecondaryChance < 0 || rule.SecondaryChance > 1 {
						t.Errorf("%s SecondaryChance out of range: %f", roomType, rule.SecondaryChance)
					}
					if rule.WeatheringBase < 0 || rule.WeatheringBase > 1 {
						t.Errorf("%s WeatheringBase out of range: %f", roomType, rule.WeatheringBase)
					}
					if rule.WeatheringRange < 0 || rule.WeatheringRange > 1 {
						t.Errorf("%s WeatheringRange out of range: %f", roomType, rule.WeatheringRange)
					}
				}
			}
		})
	}
}

func TestHashPosition(t *testing.T) {
	seed := uint64(12345)

	// Test that same position gives same hash
	h1 := hashPosition(10, 20, seed)
	h2 := hashPosition(10, 20, seed)
	if h1 != h2 {
		t.Error("same position should give same hash")
	}

	// Test that different positions give different hashes
	h3 := hashPosition(11, 20, seed)
	if h1 == h3 {
		t.Error("different positions should give different hashes")
	}

	h4 := hashPosition(10, 21, seed)
	if h1 == h4 {
		t.Error("different positions should give different hashes")
	}

	// Test that different seeds give different hashes
	h5 := hashPosition(10, 20, seed+1)
	if h1 == h5 {
		t.Error("different seeds should give different hashes")
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy", 100)
	// Update should not panic (it's a no-op for this system)
	sys.Update(nil)
}
