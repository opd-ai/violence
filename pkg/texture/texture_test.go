package texture

import (
	"image/color"
	"testing"
)

func TestNewAtlas(t *testing.T) {
	tests := []struct {
		name string
		seed uint64
	}{
		{"zero seed", 0},
		{"positive seed", 12345},
		{"large seed", 0xFFFFFFFFFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atlas := NewAtlas(tt.seed)
			if atlas == nil {
				t.Fatal("NewAtlas returned nil")
			}
			if atlas.textures == nil {
				t.Error("textures map not initialized")
			}
			if atlas.genre != "fantasy" {
				t.Errorf("default genre = %q, want %q", atlas.genre, "fantasy")
			}
			if atlas.seed != tt.seed {
				t.Errorf("seed = %d, want %d", atlas.seed, tt.seed)
			}
		})
	}
}

func TestGenerateWall(t *testing.T) {
	tests := []struct {
		name string
		seed uint64
		size int
	}{
		{"small texture", 42, 64},
		{"medium texture", 123, 128},
		{"large texture", 999, 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atlas := NewAtlas(tt.seed)
			err := atlas.Generate("test_wall", tt.size, "wall")
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			img, ok := atlas.Get("test_wall")
			if !ok {
				t.Fatal("texture not found in atlas")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.size || bounds.Dy() != tt.size {
				t.Errorf("texture size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), tt.size, tt.size)
			}
		})
	}
}

func TestGenerateFloor(t *testing.T) {
	atlas := NewAtlas(789)
	err := atlas.Generate("test_floor", 128, "floor")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	img, ok := atlas.Get("test_floor")
	if !ok {
		t.Fatal("floor texture not found")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 128 || bounds.Dy() != 128 {
		t.Errorf("floor texture size = %dx%d, want 128x128", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateCeiling(t *testing.T) {
	atlas := NewAtlas(456)
	err := atlas.Generate("test_ceiling", 64, "ceiling")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	img, ok := atlas.Get("test_ceiling")
	if !ok {
		t.Fatal("ceiling texture not found")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("ceiling texture size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateDeterministic(t *testing.T) {
	seed := uint64(12345)
	size := 64

	atlas1 := NewAtlas(seed)
	atlas1.Generate("wall1", size, "wall")
	img1, _ := atlas1.Get("wall1")

	atlas2 := NewAtlas(seed)
	atlas2.Generate("wall1", size, "wall")
	img2, _ := atlas2.Get("wall1")

	// Compare pixel-by-pixel
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				t.Errorf("pixel at (%d, %d) differs: %v vs %v", x, y, c1, c2)
				return
			}
		}
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	size := 64

	atlas1 := NewAtlas(111)
	atlas1.Generate("wall", size, "wall")
	img1, _ := atlas1.Get("wall")

	atlas2 := NewAtlas(222)
	atlas2.Generate("wall", size, "wall")
	img2, _ := atlas2.Get("wall")

	// Images should differ
	bounds := img1.Bounds()
	diffCount := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img1.At(x, y) != img2.At(x, y) {
				diffCount++
			}
		}
	}

	// At least 50% of pixels should differ
	totalPixels := size * size
	if diffCount < totalPixels/2 {
		t.Errorf("only %d/%d pixels differ, textures too similar", diffCount, totalPixels)
	}
}

func TestGetNonExistent(t *testing.T) {
	atlas := NewAtlas(123)
	_, ok := atlas.Get("nonexistent")
	if ok {
		t.Error("Get returned true for nonexistent texture")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			atlas := NewAtlas(42)
			atlas.SetGenre(tt.genre)

			// Generate texture and verify it was created
			atlas.Generate("test", 64, "wall")
			img, ok := atlas.Get("test")
			if !ok {
				t.Fatal("texture not generated after SetGenre")
			}

			// Verify genre affects color
			c := img.At(32, 32)
			r, g, b, a := c.RGBA()
			if a == 0 {
				t.Error("texture is fully transparent")
			}

			// All genres should produce different base colors
			if r == 0 && g == 0 && b == 0 {
				t.Error("texture is completely black")
			}
		})
	}
}

func TestGenreColorDifferences(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	seed := uint64(987)

	colors := make(map[string]color.RGBA)

	for _, genre := range genres {
		atlas := NewAtlas(seed)
		atlas.SetGenre(genre)

		// Get base color
		baseColor := atlas.getGenreBaseColor()
		colors[genre] = baseColor
	}

	// Verify each genre has a distinct base color
	for i, g1 := range genres {
		for j, g2 := range genres {
			if i >= j {
				continue
			}
			c1 := colors[g1]
			c2 := colors[g2]

			// Colors should differ in at least one channel
			if c1.R == c2.R && c1.G == c2.G && c1.B == c2.B {
				t.Errorf("genres %s and %s have identical base colors", g1, g2)
			}
		}
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		same bool
	}{
		{"identical", "test", "test", true},
		{"different", "test1", "test2", false},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := hashString(tt.s1)
			h2 := hashString(tt.s2)

			if tt.same && h1 != h2 {
				t.Errorf("identical strings produced different hashes: %d vs %d", h1, h2)
			}
			if !tt.same && h1 == h2 {
				t.Errorf("different strings produced same hash: %d", h1)
			}
		})
	}
}

func TestHashCoord(t *testing.T) {
	tests := []struct {
		name string
		x1   int
		y1   int
		x2   int
		y2   int
		same bool
	}{
		{"identical", 5, 10, 5, 10, true},
		{"different x", 5, 10, 6, 10, false},
		{"different y", 5, 10, 5, 11, false},
		{"swapped", 5, 10, 10, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := hashCoord(tt.x1, tt.y1)
			h2 := hashCoord(tt.x2, tt.y2)

			if tt.same && h1 != h2 {
				t.Errorf("identical coords produced different hashes")
			}
			if !tt.same && h1 == h2 {
				t.Errorf("different coords produced same hash")
			}
		})
	}
}

func TestFade(t *testing.T) {
	tests := []struct {
		t      float64
		expect float64
	}{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
	}

	for _, tt := range tests {
		result := fade(tt.t)
		if result < 0 || result > 1 {
			t.Errorf("fade(%f) = %f, out of range [0, 1]", tt.t, result)
		}
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a      float64
		b      float64
		t      float64
		expect float64
	}{
		{0, 10, 0.0, 0},
		{0, 10, 0.5, 5},
		{0, 10, 1.0, 10},
		{5, 15, 0.5, 10},
	}

	for _, tt := range tests {
		result := lerp(tt.a, tt.b, tt.t)
		if result != tt.expect {
			t.Errorf("lerp(%f, %f, %f) = %f, want %f",
				tt.a, tt.b, tt.t, result, tt.expect)
		}
	}
}

func TestDot2(t *testing.T) {
	tests := []struct {
		gx     float64
		gy     float64
		x      float64
		y      float64
		expect float64
	}{
		{1, 0, 2, 3, 2},
		{0, 1, 2, 3, 3},
		{1, 1, 2, 3, 5},
		{2, 3, 4, 5, 23},
	}

	for _, tt := range tests {
		result := dot2(tt.gx, tt.gy, tt.x, tt.y)
		if result != tt.expect {
			t.Errorf("dot2(%f, %f, %f, %f) = %f, want %f",
				tt.gx, tt.gy, tt.x, tt.y, result, tt.expect)
		}
	}
}

func TestClampUint8(t *testing.T) {
	tests := []struct {
		name   string
		value  float64
		expect uint8
	}{
		{"negative", -10, 0},
		{"zero", 0, 0},
		{"normal", 128, 128},
		{"max", 255, 255},
		{"overflow", 300, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampUint8(tt.value)
			if result != tt.expect {
				t.Errorf("clampUint8(%f) = %d, want %d", tt.value, result, tt.expect)
			}
		})
	}
}

func TestApplyNoise(t *testing.T) {
	atlas := NewAtlas(123)
	base := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	tests := []struct {
		name  string
		noise float64
	}{
		{"darken", -0.5},
		{"neutral", 0.0},
		{"brighten", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := atlas.applyNoise(base, tt.noise)

			if result.A != 255 {
				t.Errorf("alpha changed: got %d, want 255", result.A)
			}

			// Verify clamping works
			if result.R > 255 || result.G > 255 || result.B > 255 {
				t.Errorf("color values exceed 255: %+v", result)
			}
		})
	}
}

func TestPerlinNoise(t *testing.T) {
	atlas := NewAtlas(555)

	// Test that noise is deterministic
	n1 := atlas.perlinNoise(1.5, 2.5, atlas.getRNG())
	n2 := atlas.perlinNoise(1.5, 2.5, atlas.getRNG())

	if n1 != n2 {
		t.Error("perlinNoise is not deterministic")
	}

	// Test that different coordinates give different values
	n3 := atlas.perlinNoise(3.5, 4.5, atlas.getRNG())
	if n1 == n3 {
		t.Error("different coordinates produced identical noise")
	}
}

func TestGenerateUnknownType(t *testing.T) {
	atlas := NewAtlas(777)
	err := atlas.Generate("test", 64, "unknown")
	if err != nil {
		t.Fatalf("Generate failed on unknown type: %v", err)
	}

	// Should fallback to wall texture
	img, ok := atlas.Get("test")
	if !ok {
		t.Error("texture not generated for unknown type")
	}
	if img == nil {
		t.Error("generated image is nil")
	}
}

func TestMultipleTextures(t *testing.T) {
	atlas := NewAtlas(321)

	names := []string{"wall1", "wall2", "floor1", "ceiling1"}
	types := []string{"wall", "wall", "floor", "ceiling"}

	for i, name := range names {
		err := atlas.Generate(name, 64, types[i])
		if err != nil {
			t.Fatalf("Generate(%s) failed: %v", name, err)
		}
	}

	// Verify all textures exist
	for _, name := range names {
		_, ok := atlas.Get(name)
		if !ok {
			t.Errorf("texture %s not found", name)
		}
	}
}

// Benchmark texture generation
func BenchmarkGenerateWall64(b *testing.B) {
	atlas := NewAtlas(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atlas.Generate("wall", 64, "wall")
	}
}

func BenchmarkGenerateWall128(b *testing.B) {
	atlas := NewAtlas(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atlas.Generate("wall", 128, "wall")
	}
}

func BenchmarkGenerateWall256(b *testing.B) {
	atlas := NewAtlas(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atlas.Generate("wall", 256, "wall")
	}
}
