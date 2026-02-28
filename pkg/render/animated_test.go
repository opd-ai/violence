package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/raycaster"
)

// mockAnimatedAtlas implements TextureAtlas with animated texture support.
type mockAnimatedAtlas struct {
	textures map[string]image.Image
	animated map[string][]image.Image
	genre    string
}

func newMockAnimatedAtlas() *mockAnimatedAtlas {
	return &mockAnimatedAtlas{
		textures: make(map[string]image.Image),
		animated: make(map[string][]image.Image),
		genre:    "fantasy",
	}
}

func (m *mockAnimatedAtlas) Get(name string) (image.Image, bool) {
	img, ok := m.textures[name]
	return img, ok
}

func (m *mockAnimatedAtlas) GetAnimatedFrame(name string, tick int) (image.Image, bool) {
	frames, ok := m.animated[name]
	if !ok || len(frames) == 0 {
		return nil, false
	}
	frameIndex := (tick / 30) % len(frames) // 30 fps
	return frames[frameIndex], true
}

func (m *mockAnimatedAtlas) SetGenre(genreID string) {
	m.genre = genreID
}

func (m *mockAnimatedAtlas) addAnimated(name string, frames []image.Image) {
	m.animated[name] = frames
}

func TestRenderer_Tick(t *testing.T) {
	rc := raycaster.NewRaycaster(60.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	if r.tick != 0 {
		t.Errorf("initial tick = %d, want 0", r.tick)
	}

	r.Tick()
	if r.tick != 1 {
		t.Errorf("after Tick(), tick = %d, want 1", r.tick)
	}

	for i := 0; i < 100; i++ {
		r.Tick()
	}
	if r.tick != 101 {
		t.Errorf("after 101 Tick() calls, tick = %d, want 101", r.tick)
	}
}

func TestRenderer_AnimatedTexturePlayback(t *testing.T) {
	// Create mock atlas with animated texture
	atlas := newMockAnimatedAtlas()

	// Create 4 frames with different colors
	frames := make([]image.Image, 4)
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},   // Red
		{R: 0, G: 255, B: 0, A: 255},   // Green
		{R: 0, G: 0, B: 255, A: 255},   // Blue
		{R: 255, G: 255, B: 0, A: 255}, // Yellow
	}

	for i := 0; i < 4; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				img.Set(x, y, colors[i])
			}
		}
		frames[i] = img
	}

	atlas.addAnimated("fantasy_anim", frames)

	// Test frame progression
	for tick := 0; tick < 120; tick++ {
		frame, ok := atlas.GetAnimatedFrame("fantasy_anim", tick)
		if !ok {
			t.Fatalf("GetAnimatedFrame failed at tick %d", tick)
		}

		// Verify correct frame is returned
		expectedFrameIndex := (tick / 30) % 4
		expectedColor := colors[expectedFrameIndex]

		r, g, b, _ := frame.At(32, 32).RGBA()
		actualColor := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: 255,
		}

		if actualColor != expectedColor {
			t.Errorf("tick %d: frame color = %v, want %v", tick, actualColor, expectedColor)
		}
	}
}

func TestRenderer_AnimatedWallRendering(t *testing.T) {
	rc := raycaster.NewRaycaster(60.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	atlas := newMockAnimatedAtlas()
	atlas.SetGenre("fantasy")

	// Create animated frames
	frames := make([]image.Image, 2)
	for i := 0; i < 2; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		brightness := uint8(100 + i*100)
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				img.Set(x, y, color.RGBA{R: brightness, G: brightness, B: 0, A: 255})
			}
		}
		frames[i] = img
	}
	atlas.addAnimated("fantasy_anim", frames)

	r.SetTextureAtlas(atlas)

	// Test at different ticks
	hit := raycaster.RayHit{
		Distance: 2.0,
		WallType: 5, // Animated wall
		TextureX: 0.5,
		Side:     0,
		HitX:     3.5,
		HitY:     1.5,
	}

	// Tick 0 should use frame 0
	r.tick = 0
	color1 := r.renderWall(160, 100, hit)
	if color1.R < 50 {
		t.Error("frame 0 should be brighter")
	}

	// Tick 30 should use frame 1
	r.tick = 30
	color2 := r.renderWall(160, 100, hit)
	if color2.R < 100 {
		t.Error("frame 1 should be bright")
	}
}

func TestRenderer_StaticTextureForNonAnimatedWalls(t *testing.T) {
	rc := raycaster.NewRaycaster(60.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	atlas := newMockAnimatedAtlas()

	// Add static texture
	staticImg := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			staticImg.Set(x, y, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		}
	}
	atlas.textures["wall_1"] = staticImg

	r.SetTextureAtlas(atlas)

	hit := raycaster.RayHit{
		Distance: 2.0,
		WallType: 1, // Static wall
		TextureX: 0.5,
		Side:     0,
		HitX:     0.5,
		HitY:     0.5,
	}

	// Should use static texture regardless of tick
	r.tick = 0
	color1 := r.renderWall(160, 100, hit)

	r.tick = 100
	color2 := r.renderWall(160, 100, hit)

	// Colors should be similar (with possible fog/lighting differences)
	if abs(int(color1.R)-int(color2.R)) > 5 {
		t.Error("static wall color should not change with tick")
	}
}

func TestAnimatedTextureLooping(t *testing.T) {
	atlas := newMockAnimatedAtlas()

	frames := make([]image.Image, 3)
	for i := 0; i < 3; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 8, 8))
		val := uint8(i * 100)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				img.Set(x, y, color.RGBA{R: val, A: 255})
			}
		}
		frames[i] = img
	}

	atlas.addAnimated("test_anim", frames)

	// Test loop behavior
	tests := []struct {
		tick          int
		expectedFrame int
	}{
		{0, 0},
		{29, 0},
		{30, 1},
		{59, 1},
		{60, 2},
		{89, 2},
		{90, 0}, // Loop back to frame 0
		{120, 1},
	}

	for _, tt := range tests {
		frame, ok := atlas.GetAnimatedFrame("test_anim", tt.tick)
		if !ok {
			t.Fatalf("GetAnimatedFrame failed at tick %d", tt.tick)
		}

		r, _, _, _ := frame.At(0, 0).RGBA()
		frameVal := uint8(r >> 8)
		expectedVal := uint8(tt.expectedFrame * 100)

		if frameVal != expectedVal {
			t.Errorf("tick %d: frame value = %d, want %d (frame %d)",
				tt.tick, frameVal, expectedVal, tt.expectedFrame)
		}
	}
}

func TestRenderer_GenreAnimatedTextureName(t *testing.T) {
	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "fantasy_anim"},
		{"scifi", "scifi_anim"},
		{"horror", "horror_anim"},
		{"cyberpunk", "cyberpunk_anim"},
		{"postapoc", "postapoc_anim"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			rc := raycaster.NewRaycaster(60.0, 320, 200)
			r := NewRenderer(320, 200, rc)
			r.genreID = tt.genre

			atlas := newMockAnimatedAtlas()
			frames := make([]image.Image, 1)
			frames[0] = image.NewRGBA(image.Rect(0, 0, 64, 64))
			atlas.addAnimated(tt.expected, frames)

			r.SetTextureAtlas(atlas)

			hit := raycaster.RayHit{
				Distance: 2.0,
				WallType: 5,
				TextureX: 0.5,
				Side:     0,
			}

			// Should not panic
			_ = r.renderWall(160, 100, hit)
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkRenderer_AnimatedTextureRendering(b *testing.B) {
	rc := raycaster.NewRaycaster(60.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	atlas := newMockAnimatedAtlas()
	frames := make([]image.Image, 8)
	for i := 0; i < 8; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				img.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
			}
		}
		frames[i] = img
	}
	atlas.addAnimated("fantasy_anim", frames)
	r.SetTextureAtlas(atlas)

	hit := raycaster.RayHit{
		Distance: 2.0,
		WallType: 5,
		TextureX: 0.5,
		Side:     0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.tick = i
		r.renderWall(160, 100, hit)
	}
}
