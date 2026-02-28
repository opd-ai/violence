package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/raycaster"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Standard 320x200", 320, 200},
		{"HD 640x400", 640, 400},
		{"Minimal 160x100", 160, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := raycaster.NewRaycaster(66.0, tt.width, tt.height)
			r := NewRenderer(tt.width, tt.height, rc)

			if r.Width != tt.width {
				t.Errorf("Width = %d, want %d", r.Width, tt.width)
			}
			if r.Height != tt.height {
				t.Errorf("Height = %d, want %d", r.Height, tt.height)
			}
			expectedSize := tt.width * tt.height * 4
			if len(r.framebuffer) != expectedSize {
				t.Errorf("Framebuffer size = %d, want %d", len(r.framebuffer), expectedSize)
			}
			if r.raycaster == nil {
				t.Error("Raycaster is nil")
			}
			if r.palette == nil {
				t.Error("Palette is nil")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"Fantasy", "fantasy"},
		{"SciFi", "scifi"},
		{"Horror", "horror"},
		{"Cyberpunk", "cyberpunk"},
		{"PostApoc", "postapoc"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := raycaster.NewRaycaster(66.0, 320, 200)
			r := NewRenderer(320, 200, rc)

			r.SetGenre(tt.genreID)

			if r.genreID != tt.genreID {
				t.Errorf("Genre ID = %s, want %s", r.genreID, tt.genreID)
			}

			if r.palette == nil {
				t.Error("Palette is nil after SetGenre")
			}

			if len(r.palette) == 0 {
				t.Error("Palette is empty after SetGenre")
			}
		})
	}
}

func TestGetPaletteForGenre(t *testing.T) {
	tests := []struct {
		name          string
		genreID       string
		expectedWalls int
	}{
		{"Fantasy palette", "fantasy", 5},
		{"SciFi palette", "scifi", 5},
		{"Horror palette", "horror", 5},
		{"Cyberpunk palette", "cyberpunk", 5},
		{"PostApoc palette", "postapoc", 5},
		{"Default palette", "unknown", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette := getPaletteForGenre(tt.genreID)

			if len(palette) != tt.expectedWalls {
				t.Errorf("Palette size = %d, want %d", len(palette), tt.expectedWalls)
			}

			for i := 0; i < tt.expectedWalls; i++ {
				c, ok := palette[i]
				if !ok {
					t.Errorf("Palette missing entry for key %d", i)
				}
				if c.A != 255 {
					t.Errorf("Color alpha = %d, want 255", c.A)
				}
			}
		})
	}
}

func TestPaletteDifference(t *testing.T) {
	fantasyPalette := getPaletteForGenre("fantasy")
	scifiPalette := getPaletteForGenre("scifi")

	if fantasyPalette[1] == scifiPalette[1] {
		t.Error("Fantasy and SciFi palettes should have different wall colors")
	}
}

func TestRenderWall(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	tests := []struct {
		name     string
		x        int
		y        int
		hit      raycaster.RayHit
		wantZero bool
	}{
		{
			name:     "No wall hit",
			x:        160,
			y:        100,
			hit:      raycaster.RayHit{Distance: 1e30, WallType: 0, Side: 0},
			wantZero: true,
		},
		{
			name:     "Empty tile",
			x:        160,
			y:        100,
			hit:      raycaster.RayHit{Distance: 5.0, WallType: 0, Side: 0},
			wantZero: true,
		},
		{
			name:     "Close wall",
			x:        160,
			y:        100,
			hit:      raycaster.RayHit{Distance: 2.0, WallType: 1, Side: 0},
			wantZero: false,
		},
		{
			name:     "Far wall",
			x:        160,
			y:        100,
			hit:      raycaster.RayHit{Distance: 10.0, WallType: 1, Side: 0},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := r.renderWall(tt.x, tt.y, tt.hit)
			isZero := c.A == 0
			if isZero != tt.wantZero {
				t.Errorf("renderWall alpha zero = %v, want %v", isZero, tt.wantZero)
			}
		})
	}
}

func TestRenderWallSideShading(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	hitSide0 := raycaster.RayHit{Distance: 5.0, WallType: 1, Side: 0}
	hitSide1 := raycaster.RayHit{Distance: 5.0, WallType: 1, Side: 1}

	c0 := r.renderWall(160, 100, hitSide0)
	c1 := r.renderWall(160, 100, hitSide1)

	if c0.R <= c1.R {
		t.Error("Side 0 should be brighter than side 1")
	}
}

func TestRenderFloorAndCeiling(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})
	r := NewRenderer(320, 200, rc)

	floorColor := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)
	if floorColor.A != 255 {
		t.Error("Floor should have full alpha")
	}

	ceilingColor := r.renderCeiling(160, 50, 2.5, 2.5, 1.0, 0.0, 0.0)
	if ceilingColor.A != 255 {
		t.Error("Ceiling should have full alpha")
	}
}

func TestRender(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	simpleMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}
	rc.SetMap(simpleMap)

	r := NewRenderer(320, 200, rc)
	screen := ebiten.NewImage(320, 200)

	r.Render(screen, 2.5, 2.5, 1.0, 0.0, 0.0)

	if r.framebuffer == nil {
		t.Error("Framebuffer is nil after render")
	}

	allBlack := true
	for i := 0; i < len(r.framebuffer); i += 4 {
		if r.framebuffer[i] != 0 || r.framebuffer[i+1] != 0 || r.framebuffer[i+2] != 0 {
			allBlack = false
			break
		}
	}

	if allBlack {
		t.Error("Framebuffer should not be all black after render")
	}
}

func TestRenderResolutionMatches(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Standard", 320, 200},
		{"Wide", 640, 200},
		{"Tall", 320, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := raycaster.NewRaycaster(66.0, tt.width, tt.height)
			r := NewRenderer(tt.width, tt.height, rc)

			simpleMap := [][]int{
				{1, 1, 1},
				{1, 0, 1},
				{1, 1, 1},
			}
			rc.SetMap(simpleMap)

			screen := ebiten.NewImage(tt.width, tt.height)
			r.Render(screen, 1.5, 1.5, 1.0, 0.0, 0.0)

			expectedSize := tt.width * tt.height * 4
			if len(r.framebuffer) != expectedSize {
				t.Errorf("Framebuffer size = %d, want %d", len(r.framebuffer), expectedSize)
			}
		})
	}
}

func TestFrameImage(t *testing.T) {
	data := make([]byte, 10*10*4)
	for i := 0; i < len(data); i += 4 {
		data[i] = 255   // R
		data[i+1] = 128 // G
		data[i+2] = 64  // B
		data[i+3] = 255 // A
	}

	img := &frameImage{
		data:   data,
		width:  10,
		height: 10,
	}

	bounds := img.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("Bounds = %v, want 10x10", bounds)
	}

	c := img.At(5, 5)
	rgba, ok := c.(color.RGBA)
	if !ok {
		t.Fatal("Color is not RGBA")
	}
	if rgba.R != 255 || rgba.G != 128 || rgba.B != 64 || rgba.A != 255 {
		t.Errorf("Color at (5,5) = %v, want RGBA{255,128,64,255}", rgba)
	}

	outOfBounds := img.At(-1, -1)
	expected := color.RGBA{0, 0, 0, 255}
	if outOfBounds != expected {
		t.Errorf("Out of bounds color = %v, want %v", outOfBounds, expected)
	}

	if img.ColorModel() != color.RGBAModel {
		t.Error("ColorModel should be RGBAModel")
	}
}

func BenchmarkRender(b *testing.B) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	simpleMap := make([][]int, 20)
	for i := range simpleMap {
		simpleMap[i] = make([]int, 20)
		for j := range simpleMap[i] {
			if i == 0 || i == 19 || j == 0 || j == 19 {
				simpleMap[i][j] = 1
			}
		}
	}
	rc.SetMap(simpleMap)

	r := NewRenderer(320, 200, rc)
	screen := ebiten.NewImage(320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Render(screen, 10.5, 10.5, 1.0, 0.0, 0.0)
	}
}

func BenchmarkSetGenre(b *testing.B) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.SetGenre(genres[i%len(genres)])
	}
}

func TestSetTextureAtlas(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	if r.atlas != nil {
		t.Error("Atlas should be nil initially")
	}

	// Create a mock atlas (we'll use the real one for integration)
	atlas := &mockAtlas{}
	r.SetTextureAtlas(atlas)

	if r.atlas == nil {
		t.Error("Atlas should be set after SetTextureAtlas")
	}
}

func TestRenderFloorWithTexture(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})
	r := NewRenderer(320, 200, rc)

	// Render without texture atlas (palette mode)
	colorNoTex := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)
	if colorNoTex.A != 255 {
		t.Error("Floor should have full alpha without texture")
	}

	// Now with atlas that doesn't have floor texture (should fallback to palette)
	atlas := &mockAtlas{hasFloor: false}
	r.SetTextureAtlas(atlas)
	colorNoFloor := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)
	if colorNoFloor.A != 255 {
		t.Error("Floor should have full alpha when texture missing")
	}

	// With atlas that has floor texture
	atlas2 := &mockAtlas{hasFloor: true}
	r.SetTextureAtlas(atlas2)
	colorWithTex := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)
	if colorWithTex.A != 255 {
		t.Error("Floor should have full alpha with texture")
	}
}

func TestRenderCeilingWithTexture(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})
	r := NewRenderer(320, 200, rc)

	// Without texture atlas
	colorNoTex := r.renderCeiling(160, 50, 2.5, 2.5, 1.0, 0.0, 0.0)
	if colorNoTex.A != 255 {
		t.Error("Ceiling should have full alpha without texture")
	}

	// With atlas that has ceiling texture
	atlas := &mockAtlas{hasCeiling: true}
	r.SetTextureAtlas(atlas)
	colorWithTex := r.renderCeiling(160, 50, 2.5, 2.5, 1.0, 0.0, 0.0)
	if colorWithTex.A != 255 {
		t.Error("Ceiling should have full alpha with texture")
	}
}

func TestSampleTexture(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	// Create a simple test texture
	tex := &mockTexture{
		w:     8,
		h:     8,
		color: color.RGBA{R: 100, G: 150, B: 200, A: 255},
	}

	tests := []struct {
		name   string
		worldX float64
		worldY float64
	}{
		{"origin", 0.0, 0.0},
		{"positive coords", 1.5, 2.3},
		{"negative coords", -0.5, -1.2},
		{"large coords", 10.7, 15.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := r.sampleTexture(tex, tt.worldX, tt.worldY)
			if c.A != 255 {
				t.Errorf("sampled color alpha = %d, want 255", c.A)
			}
			// Should return the test color
			if c.R != 100 || c.G != 150 || c.B != 200 {
				t.Errorf("sampled color = %v, want RGB(100,150,200)", c)
			}
		})
	}
}

func TestSampleTextureWrapping(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	tex := &mockTexture{
		w:     4,
		h:     4,
		color: color.RGBA{R: 50, G: 100, B: 150, A: 255},
	}

	// Test that coordinates wrap correctly
	c1 := r.sampleTexture(tex, 0.0, 0.0)
	c2 := r.sampleTexture(tex, 1.0, 1.0) // Should wrap to same texel
	c3 := r.sampleTexture(tex, 2.0, 2.0) // Should also wrap to same texel

	if c1 != c2 || c1 != c3 {
		t.Error("Texture wrapping should produce identical colors at integer boundaries")
	}
}

func BenchmarkSampleTexture(b *testing.B) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	tex := &mockTexture{
		w:     64,
		h:     64,
		color: color.RGBA{R: 128, G: 128, B: 128, A: 255},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.sampleTexture(tex, float64(i%100)/10.0, float64(i%100)/10.0)
	}
}

// mockAtlas is a minimal texture.Atlas for testing
type mockAtlas struct {
	hasFloor   bool
	hasCeiling bool
}

func (m *mockAtlas) Get(name string) (image.Image, bool) {
	if name == "floor_main" && m.hasFloor {
		return &mockTexture{w: 8, h: 8, color: color.RGBA{R: 80, G: 70, B: 60, A: 255}}, true
	}
	if name == "ceiling_main" && m.hasCeiling {
		return &mockTexture{w: 8, h: 8, color: color.RGBA{R: 60, G: 50, B: 40, A: 255}}, true
	}
	return nil, false
}

func (m *mockAtlas) SetGenre(genreID string) {}

// mockTexture is a simple test texture that returns a constant color
type mockTexture struct {
	w     int
	h     int
	color color.RGBA
}

func (m *mockTexture) ColorModel() color.Model {
	return color.RGBAModel
}

func (m *mockTexture) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.w, m.h)
}

func (m *mockTexture) At(x, y int) color.Color {
	return m.color
}

// mockLightMap is a simple test light map
type mockLightMap struct {
	width   int
	height  int
	uniform float64
	custom  map[int]float64 // Key is y*width+x
}

func (m *mockLightMap) GetLight(x, y int) float64 {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return 0.0
	}
	if m.custom != nil {
		if val, ok := m.custom[y*m.width+x]; ok {
			return val
		}
	}
	return m.uniform
}

func (m *mockLightMap) Calculate() {}

func TestSetLightMap(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	if r.lightMap != nil {
		t.Error("LightMap should be nil initially")
	}

	lightMap := &mockLightMap{width: 10, height: 10, uniform: 0.5}
	r.SetLightMap(lightMap)

	if r.lightMap == nil {
		t.Error("LightMap should be set after SetLightMap")
	}
}

func TestGetLightMultiplier(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	r := NewRenderer(320, 200, rc)

	tests := []struct {
		name     string
		lightMap *mockLightMap
		worldX   float64
		worldY   float64
		expected float64
	}{
		{
			name:     "No light map - full brightness",
			lightMap: nil,
			worldX:   5.5,
			worldY:   5.5,
			expected: 1.0,
		},
		{
			name:     "Uniform light map - 50% brightness",
			lightMap: &mockLightMap{width: 20, height: 20, uniform: 0.5},
			worldX:   5.5,
			worldY:   5.5,
			expected: 0.5,
		},
		{
			name:     "Uniform light map - full brightness",
			lightMap: &mockLightMap{width: 20, height: 20, uniform: 1.0},
			worldX:   10.2,
			worldY:   15.7,
			expected: 1.0,
		},
		{
			name:     "Uniform light map - dark",
			lightMap: &mockLightMap{width: 20, height: 20, uniform: 0.1},
			worldX:   3.9,
			worldY:   7.1,
			expected: 0.1,
		},
		{
			name: "Custom light map - bright spot",
			lightMap: &mockLightMap{
				width:   20,
				height:  20,
				uniform: 0.3,
				custom: map[int]float64{
					5*20 + 5: 0.9, // tile (5,5) is bright
				},
			},
			worldX:   5.5,
			worldY:   5.5,
			expected: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lightMap != nil {
				r.SetLightMap(tt.lightMap)
			} else {
				r.SetLightMap(nil)
			}

			result := r.getLightMultiplier(tt.worldX, tt.worldY)
			if result != tt.expected {
				t.Errorf("getLightMultiplier(%v, %v) = %v, want %v",
					tt.worldX, tt.worldY, result, tt.expected)
			}
		})
	}
}

func TestRenderWallWithLighting(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})

	r := NewRenderer(320, 200, rc)

	tests := []struct {
		name      string
		lightMap  *mockLightMap
		hitX      float64
		hitY      float64
		expectDim bool
	}{
		{
			name:      "No lighting - full brightness",
			lightMap:  nil,
			hitX:      2.5,
			hitY:      2.5,
			expectDim: false,
		},
		{
			name:      "Dim lighting",
			lightMap:  &mockLightMap{width: 5, height: 5, uniform: 0.3},
			hitX:      2.5,
			hitY:      2.5,
			expectDim: true,
		},
		{
			name:      "Bright lighting",
			lightMap:  &mockLightMap{width: 5, height: 5, uniform: 1.0},
			hitX:      2.5,
			hitY:      2.5,
			expectDim: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lightMap != nil {
				r.SetLightMap(tt.lightMap)
			} else {
				r.SetLightMap(nil)
			}

			hit := raycaster.RayHit{
				Distance: 2.0,
				WallType: 1,
				Side:     0,
				HitX:     tt.hitX,
				HitY:     tt.hitY,
			}

			colorWithLight := r.renderWall(160, 100, hit)

			// Without light map for comparison
			r.SetLightMap(nil)
			colorNoLight := r.renderWall(160, 100, hit)

			if tt.expectDim {
				// With dim lighting, color should be darker
				if colorWithLight.R >= colorNoLight.R {
					t.Errorf("Expected dimmer color with lighting, got R=%d vs %d",
						colorWithLight.R, colorNoLight.R)
				}
			} else {
				// With full/no lighting, colors should be similar
				diff := int(colorNoLight.R) - int(colorWithLight.R)
				if diff < -5 || diff > 5 {
					t.Errorf("Expected similar brightness, got R=%d vs %d (diff %d)",
						colorWithLight.R, colorNoLight.R, diff)
				}
			}
		})
	}
}

func TestRenderFloorWithLighting(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})

	r := NewRenderer(320, 200, rc)

	// Test with dim lighting
	lightMap := &mockLightMap{width: 5, height: 5, uniform: 0.4}
	r.SetLightMap(lightMap)

	colorWithLight := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)

	// Without light map for comparison
	r.SetLightMap(nil)
	colorNoLight := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)

	// With dim lighting, floor should be darker
	if colorWithLight.R >= colorNoLight.R {
		t.Errorf("Expected dimmer floor with lighting, got R=%d vs %d",
			colorWithLight.R, colorNoLight.R)
	}
}

func TestRenderCeilingWithLighting(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})

	r := NewRenderer(320, 200, rc)

	// Test with bright lighting
	lightMap := &mockLightMap{width: 5, height: 5, uniform: 0.9}
	r.SetLightMap(lightMap)

	colorWithLight := r.renderCeiling(160, 50, 2.5, 2.5, 1.0, 0.0, 0.0)

	// Without light map for comparison
	r.SetLightMap(nil)
	colorNoLight := r.renderCeiling(160, 50, 2.5, 2.5, 1.0, 0.0, 0.0)

	// With bright lighting (0.9), ceiling should be only slightly darker
	diff := int(colorNoLight.R) - int(colorWithLight.R)
	if diff > 15 {
		t.Errorf("Expected similar brightness with 0.9 lighting, got R=%d vs %d (diff %d)",
			colorWithLight.R, colorNoLight.R, diff)
	}
}

func TestRenderWithTextureAndLighting(t *testing.T) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})

	r := NewRenderer(320, 200, rc)

	// Set up both texture atlas and light map
	atlas := &mockAtlas{hasFloor: true, hasCeiling: true}
	r.SetTextureAtlas(atlas)

	lightMap := &mockLightMap{width: 5, height: 5, uniform: 0.6}
	r.SetLightMap(lightMap)

	// Render floor with both texture and lighting
	colorWithBoth := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)

	// Render with only texture
	r.SetLightMap(nil)
	colorTextureOnly := r.renderFloor(160, 150, 2.5, 2.5, 1.0, 0.0, 0.0)

	// With lighting applied, textured floor should be darker
	if colorWithBoth.R >= colorTextureOnly.R {
		t.Errorf("Expected darker textured floor with lighting, got R=%d vs %d",
			colorWithBoth.R, colorTextureOnly.R)
	}

	// Verify the color is non-zero (texture was sampled)
	if colorWithBoth.R == 0 && colorWithBoth.G == 0 && colorWithBoth.B == 0 {
		t.Error("Expected non-zero color from textured floor")
	}
}

func BenchmarkRenderWithLighting(b *testing.B) {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	})

	r := NewRenderer(320, 200, rc)
	lightMap := &mockLightMap{width: 10, height: 10, uniform: 0.5}
	r.SetLightMap(lightMap)

	screen := ebiten.NewImage(320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Render(screen, 5.5, 5.5, 1.0, 0.0, 0.0)
	}
}
