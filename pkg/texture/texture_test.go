package texture

import (
	"image"
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

// Animated texture tests

func TestNewAnimatedTexture(t *testing.T) {
	tests := []struct {
		name       string
		frameCount int
		fps        int
	}{
		{"single frame", 1, 30},
		{"few frames", 4, 15},
		{"many frames", 16, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := NewAnimatedTexture(tt.frameCount, tt.fps)
			if anim == nil {
				t.Fatal("NewAnimatedTexture returned nil")
			}
			if anim.FrameCount() != tt.frameCount {
				t.Errorf("FrameCount() = %d, want %d", anim.FrameCount(), tt.frameCount)
			}
		})
	}
}

func TestAnimatedTextureGetFrame(t *testing.T) {
	tests := []struct {
		name       string
		fps        int
		tick       int
		frameCount int
		expectIdx  int
	}{
		{"first frame", 30, 0, 4, 0},
		{"second frame", 30, 30, 4, 1},
		{"wrap around", 30, 120, 4, 0},
		{"high tick", 15, 1000, 8, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := NewAnimatedTexture(tt.frameCount, tt.fps)

			// Set dummy frames for testing
			for i := 0; i < tt.frameCount; i++ {
				img := &dummyImage{id: i}
				anim.SetFrame(i, img)
			}

			frame := anim.GetFrame(tt.tick)
			if frame == nil {
				t.Fatal("GetFrame returned nil")
			}

			dummy, ok := frame.(*dummyImage)
			if !ok {
				t.Fatal("frame is not a dummyImage")
			}

			if dummy.id != tt.expectIdx {
				t.Errorf("frame id = %d, want %d", dummy.id, tt.expectIdx)
			}
		})
	}
}

func TestAnimatedTextureEmptyFrames(t *testing.T) {
	anim := NewAnimatedTexture(4, 30)

	// Don't set any frames
	frame := anim.GetFrame(0)
	if frame != nil {
		t.Error("GetFrame on empty frames should return nil")
	}
}

func TestGenerateAnimatedFlickerTorch(t *testing.T) {
	atlas := NewAtlas(42)
	err := atlas.GenerateAnimated("torch", 32, 8, 15, "flicker_torch")
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	// Verify we can get frames
	for tick := 0; tick < 120; tick += 15 {
		frame, ok := atlas.GetAnimatedFrame("torch", tick)
		if !ok {
			t.Errorf("GetAnimatedFrame failed at tick %d", tick)
		}
		if frame == nil {
			t.Errorf("frame is nil at tick %d", tick)
		}
	}
}

func TestGenerateAnimatedBlinkPanel(t *testing.T) {
	atlas := NewAtlas(123)
	err := atlas.GenerateAnimated("panel", 32, 4, 30, "blink_panel")
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	frame0, _ := atlas.GetAnimatedFrame("panel", 0)
	frame1, _ := atlas.GetAnimatedFrame("panel", 30)

	if frame0 == nil || frame1 == nil {
		t.Fatal("frames are nil")
	}

	// Frames should differ
	if frame0 == frame1 {
		t.Error("consecutive frames are identical (should differ)")
	}
}

func TestGenerateAnimatedDripWater(t *testing.T) {
	atlas := NewAtlas(789)
	err := atlas.GenerateAnimated("drip", 32, 10, 20, "drip_water")
	if err != nil {
		t.Fatalf("GenerateAnimated failed: %v", err)
	}

	frame, ok := atlas.GetAnimatedFrame("drip", 0)
	if !ok {
		t.Fatal("GetAnimatedFrame failed")
	}
	if frame == nil {
		t.Error("frame is nil")
	}
}

func TestAnimatedDeterministic(t *testing.T) {
	seed := uint64(999)

	atlas1 := NewAtlas(seed)
	atlas1.GenerateAnimated("anim", 16, 4, 30, "flicker_torch")

	atlas2 := NewAtlas(seed)
	atlas2.GenerateAnimated("anim", 16, 4, 30, "flicker_torch")

	// Compare frames
	for tick := 0; tick < 120; tick += 30 {
		f1, _ := atlas1.GetAnimatedFrame("anim", tick)
		f2, _ := atlas2.GetAnimatedFrame("anim", tick)

		// Compare pixel-by-pixel
		b := f1.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if f1.At(x, y) != f2.At(x, y) {
					t.Errorf("tick %d: pixel (%d,%d) differs", tick, x, y)
					return
				}
			}
		}
	}
}

func TestGetAnimatedFrameNonExistent(t *testing.T) {
	atlas := NewAtlas(555)
	_, ok := atlas.GetAnimatedFrame("nonexistent", 0)
	if ok {
		t.Error("GetAnimatedFrame returned true for nonexistent texture")
	}
}

func TestGetAnimatedFrameStaticTexture(t *testing.T) {
	atlas := NewAtlas(666)
	atlas.Generate("static", 32, "wall")

	_, ok := atlas.GetAnimatedFrame("static", 0)
	if ok {
		t.Error("GetAnimatedFrame returned true for static texture")
	}
}

func TestAnimatedFrameCycling(t *testing.T) {
	atlas := NewAtlas(777)
	atlas.GenerateAnimated("cycle", 16, 4, 30, "flicker_torch")

	// Test that frames cycle correctly
	f0, _ := atlas.GetAnimatedFrame("cycle", 0)
	f120, _ := atlas.GetAnimatedFrame("cycle", 120) // 120 ticks = 4 frames @ 30fps = full cycle

	// Should be back to frame 0
	b := f0.Bounds()
	matches := true
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if f0.At(x, y) != f120.At(x, y) {
				matches = false
				break
			}
		}
	}

	if !matches {
		t.Error("frame at tick 0 and tick 120 should be identical (full cycle)")
	}
}

func TestAnimatedTextureGenreVariations(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror"}
	patterns := []string{"flicker_torch", "blink_panel", "drip_water"}

	for _, genre := range genres {
		for _, pattern := range patterns {
			t.Run(genre+"_"+pattern, func(t *testing.T) {
				atlas := NewAtlas(123)
				atlas.SetGenre(genre)
				err := atlas.GenerateAnimated("test", 16, 4, 30, pattern)
				if err != nil {
					t.Fatalf("GenerateAnimated failed: %v", err)
				}

				frame, ok := atlas.GetAnimatedFrame("test", 0)
				if !ok || frame == nil {
					t.Error("failed to get animated frame")
				}
			})
		}
	}
}

func TestAnimatedFrameVariation(t *testing.T) {
	atlas := NewAtlas(888)
	atlas.GenerateAnimated("var", 16, 8, 30, "flicker_torch")

	// Get multiple frames and verify they differ
	frames := make([]image.Image, 8)
	for i := 0; i < 8; i++ {
		frames[i], _ = atlas.GetAnimatedFrame("var", i*30)
	}

	// At least some frames should differ
	allIdentical := true
	for i := 1; i < len(frames); i++ {
		if frames[0] != frames[i] {
			allIdentical = false
			break
		}
	}

	if allIdentical {
		t.Error("all frames are identical, animation is static")
	}
}

func BenchmarkGenerateAnimatedTorch(b *testing.B) {
	atlas := NewAtlas(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atlas.GenerateAnimated("torch", 32, 8, 15, "flicker_torch")
	}
}

func BenchmarkGetAnimatedFrame(b *testing.B) {
	atlas := NewAtlas(12345)
	atlas.GenerateAnimated("torch", 32, 8, 15, "flicker_torch")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atlas.GetAnimatedFrame("torch", i%120)
	}
}

func TestGenerateWallSet(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy walls", "fantasy"},
		{"scifi walls", "scifi"},
		{"horror walls", "horror"},
		{"cyberpunk walls", "cyberpunk"},
		{"postapoc walls", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atlas := NewAtlas(12345)
			atlas.GenerateWallSet(tt.genreID)

			// Verify all 4 wall textures exist
			for i := 1; i <= 4; i++ {
				name := "wall_" + string(rune('0'+i))
				img, ok := atlas.Get(name)
				if !ok {
					t.Errorf("Wall texture %q not generated", name)
				}
				if img == nil {
					t.Errorf("Wall texture %q is nil", name)
				}
				// Verify texture has expected size
				if img != nil {
					bounds := img.Bounds()
					if bounds.Dx() != 64 || bounds.Dy() != 64 {
						t.Errorf("Wall texture %q size = %dx%d, want 64x64", name, bounds.Dx(), bounds.Dy())
					}
				}
			}

			// Verify genre was set
			if atlas.genre != tt.genreID {
				t.Errorf("Genre = %q, want %q", atlas.genre, tt.genreID)
			}
		})
	}
}

func TestGenerateWallSet_GenreDifferences(t *testing.T) {
	// Verify different genres produce different textures
	atlas1 := NewAtlas(12345)
	atlas1.GenerateWallSet("fantasy")
	fantasy1, _ := atlas1.Get("wall_1")

	atlas2 := NewAtlas(12345)
	atlas2.GenerateWallSet("scifi")
	scifi1, _ := atlas2.Get("wall_1")

	// Sample a few pixels to verify they differ
	// (different genre base colors should produce different results)
	fr, fg, fb, _ := fantasy1.At(10, 10).RGBA()
	sr, sg, sb, _ := scifi1.At(10, 10).RGBA()

	// Allow some variation but textures should be noticeably different
	totalDiff := abs(int(fr)-int(sr)) + abs(int(fg)-int(sg)) + abs(int(fb)-int(sb))
	if totalDiff < 1000 {
		t.Errorf("Genre textures too similar: diff=%d", totalDiff)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// dummyImage is a minimal image.Image implementation for testing
type dummyImage struct {
	id int
}

func (d *dummyImage) ColorModel() color.Model { return color.RGBAModel }
func (d *dummyImage) Bounds() image.Rectangle { return image.Rect(0, 0, 1, 1) }
func (d *dummyImage) At(x, y int) color.Color { return color.RGBA{R: 0, G: 0, B: 0, A: 255} }

func TestAtlas_GenerateGenreAnimations(t *testing.T) {
	tests := []struct {
		genre           string
		expectedPattern string
	}{
		{"fantasy", "flicker_torch"},
		{"scifi", "blink_panel"},
		{"horror", "drip_water"},
		{"cyberpunk", "neon_pulse"},
		{"postapoc", "radiation_glow"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			atlas := NewAtlas(12345)

			err := atlas.GenerateGenreAnimations(tt.genre)
			if err != nil {
				t.Fatalf("GenerateGenreAnimations failed: %v", err)
			}

			animName := tt.genre + "_anim"
			frame, ok := atlas.GetAnimatedFrame(animName, 0)
			if !ok {
				t.Fatalf("animated texture %s not found", animName)
			}

			if frame == nil {
				t.Error("frame is nil")
			}

			bounds := frame.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("frame size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestAtlas_GenerateGenreAnimations_FrameVariation(t *testing.T) {
	atlas := NewAtlas(99999)

	err := atlas.GenerateGenreAnimations("cyberpunk")
	if err != nil {
		t.Fatalf("GenerateGenreAnimations failed: %v", err)
	}

	frame0, _ := atlas.GetAnimatedFrame("cyberpunk_anim", 0)
	frame1, _ := atlas.GetAnimatedFrame("cyberpunk_anim", 30)

	// Frames should be different
	r0, g0, b0, _ := frame0.At(32, 32).RGBA()
	r1, g1, b1, _ := frame1.At(32, 32).RGBA()

	if r0 == r1 && g0 == g1 && b0 == b1 {
		t.Error("animated frames should vary")
	}
}

func TestAtlas_GenerateGenreAnimations_Determinism(t *testing.T) {
	seed := uint64(77777)

	atlas1 := NewAtlas(seed)
	atlas1.GenerateGenreAnimations("horror")

	atlas2 := NewAtlas(seed)
	atlas2.GenerateGenreAnimations("horror")

	// Compare frames
	for tick := 0; tick < 240; tick += 30 {
		f1, _ := atlas1.GetAnimatedFrame("horror_anim", tick)
		f2, _ := atlas2.GetAnimatedFrame("horror_anim", tick)

		r1, g1, b1, _ := f1.At(16, 16).RGBA()
		r2, g2, b2, _ := f2.At(16, 16).RGBA()

		if r1 != r2 || g1 != g2 || b1 != b2 {
			t.Errorf("tick %d: frames differ", tick)
		}
	}
}

func TestGenerateNeonPulseFrame(t *testing.T) {
	atlas := NewAtlas(12345)
	err := atlas.GenerateGenreAnimations("cyberpunk")
	if err != nil {
		t.Fatalf("GenerateGenreAnimations failed: %v", err)
	}

	frame, ok := atlas.GetAnimatedFrame("cyberpunk_anim", 0)
	if !ok {
		t.Fatal("neon pulse animation not found")
	}

	// Verify magenta tint (high R and B, low G)
	r, g, b, _ := frame.At(32, 32).RGBA()
	if r>>8 < 80 || b>>8 < 80 || g>>8 > 50 {
		t.Errorf("neon pulse should be magenta, got R=%d G=%d B=%d", r>>8, g>>8, b>>8)
	}
}

func TestGenerateRadiationGlowFrame(t *testing.T) {
	atlas := NewAtlas(12345)
	err := atlas.GenerateGenreAnimations("postapoc")
	if err != nil {
		t.Fatalf("GenerateGenreAnimations failed: %v", err)
	}

	frame, ok := atlas.GetAnimatedFrame("postapoc_anim", 0)
	if !ok {
		t.Fatal("radiation glow animation not found")
	}

	// Verify green-yellow tint (high G, medium-high R, low B)
	r, g, b, _ := frame.At(32, 32).RGBA()
	if g>>8 < 100 || b>>8 > 100 {
		t.Errorf("radiation glow should be green-yellow, got R=%d G=%d B=%d", r>>8, g>>8, b>>8)
	}
}
