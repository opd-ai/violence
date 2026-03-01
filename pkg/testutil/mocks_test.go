package testutil

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestMockScreen_NewMockScreen(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small screen", 64, 48},
		{"medium screen", 320, 240},
		{"large screen", 1280, 720},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := NewMockScreen(tt.width, tt.height)

			if screen.Width != tt.width {
				t.Errorf("Width: got %d, want %d", screen.Width, tt.width)
			}
			if screen.Height != tt.height {
				t.Errorf("Height: got %d, want %d", screen.Height, tt.height)
			}
			expectedPixels := tt.width * tt.height * 4
			if len(screen.Pixels) != expectedPixels {
				t.Errorf("Pixels length: got %d, want %d", len(screen.Pixels), expectedPixels)
			}
		})
	}
}

func TestMockScreen_Fill(t *testing.T) {
	tests := []struct {
		name  string
		color color.Color
	}{
		{"red", color.RGBA{255, 0, 0, 255}},
		{"green", color.RGBA{0, 255, 0, 255}},
		{"blue", color.RGBA{0, 0, 255, 255}},
		{"white", color.RGBA{255, 255, 255, 255}},
		{"black", color.RGBA{0, 0, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := NewMockScreen(10, 10)
			screen.Fill(tt.color)

			r, g, b, a := tt.color.RGBA()
			// Convert to 8-bit
			r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

			// Check first pixel
			if screen.Pixels[0] != r8 || screen.Pixels[1] != g8 ||
				screen.Pixels[2] != b8 || screen.Pixels[3] != a8 {
				t.Errorf("first pixel: got RGBA(%d,%d,%d,%d), want RGBA(%d,%d,%d,%d)",
					screen.Pixels[0], screen.Pixels[1], screen.Pixels[2], screen.Pixels[3],
					r8, g8, b8, a8)
			}

			// Check last pixel
			lastIdx := len(screen.Pixels) - 4
			if screen.Pixels[lastIdx] != r8 || screen.Pixels[lastIdx+1] != g8 ||
				screen.Pixels[lastIdx+2] != b8 || screen.Pixels[lastIdx+3] != a8 {
				t.Errorf("last pixel: got RGBA(%d,%d,%d,%d), want RGBA(%d,%d,%d,%d)",
					screen.Pixels[lastIdx], screen.Pixels[lastIdx+1],
					screen.Pixels[lastIdx+2], screen.Pixels[lastIdx+3],
					r8, g8, b8, a8)
			}
		})
	}
}

func TestMockScreen_Bounds(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"square", 100, 100},
		{"wide", 200, 100},
		{"tall", 100, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := NewMockScreen(tt.width, tt.height)
			bounds := screen.Bounds()

			if bounds.Min.X != 0 || bounds.Min.Y != 0 {
				t.Errorf("bounds min: got (%d,%d), want (0,0)", bounds.Min.X, bounds.Min.Y)
			}
			if bounds.Max.X != tt.width || bounds.Max.Y != tt.height {
				t.Errorf("bounds max: got (%d,%d), want (%d,%d)",
					bounds.Max.X, bounds.Max.Y, tt.width, tt.height)
			}
		})
	}
}

func TestMockInput_Keys(t *testing.T) {
	input := NewMockInput()

	// Initially no keys pressed
	if input.IsKeyPressed(ebiten.KeyW) {
		t.Error("KeyW should not be pressed initially")
	}

	// Press key
	input.PressKey(ebiten.KeyW)
	if !input.IsKeyPressed(ebiten.KeyW) {
		t.Error("KeyW should be pressed after PressKey")
	}

	// Release key
	input.ReleaseKey(ebiten.KeyW)
	if input.IsKeyPressed(ebiten.KeyW) {
		t.Error("KeyW should not be pressed after ReleaseKey")
	}
}

func TestMockInput_MultipleKeys(t *testing.T) {
	input := NewMockInput()

	keys := []ebiten.Key{ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD}

	// Press all keys
	for _, k := range keys {
		input.PressKey(k)
	}

	// Verify all pressed
	for _, k := range keys {
		if !input.IsKeyPressed(k) {
			t.Errorf("key %v should be pressed", k)
		}
	}

	// Release one key
	input.ReleaseKey(ebiten.KeyA)

	// Verify state
	if !input.IsKeyPressed(ebiten.KeyW) {
		t.Error("KeyW should still be pressed")
	}
	if input.IsKeyPressed(ebiten.KeyA) {
		t.Error("KeyA should be released")
	}
	if !input.IsKeyPressed(ebiten.KeyS) {
		t.Error("KeyS should still be pressed")
	}
}

func TestMockInput_GamepadButtons(t *testing.T) {
	input := NewMockInput()

	// Initially no buttons pressed
	if input.IsButtonPressed(ebiten.GamepadButton0) {
		t.Error("Button0 should not be pressed initially")
	}

	// Press button
	input.PressButton(ebiten.GamepadButton0)
	if !input.IsButtonPressed(ebiten.GamepadButton0) {
		t.Error("Button0 should be pressed after PressButton")
	}

	// Release button
	input.ReleaseButton(ebiten.GamepadButton0)
	if input.IsButtonPressed(ebiten.GamepadButton0) {
		t.Error("Button0 should not be pressed after ReleaseButton")
	}
}

func TestMockInput_MousePosition(t *testing.T) {
	input := NewMockInput()

	// Initial position should be 0,0
	if input.MouseX != 0 || input.MouseY != 0 {
		t.Errorf("initial mouse pos: got (%d,%d), want (0,0)", input.MouseX, input.MouseY)
	}

	// Set position
	input.SetMousePosition(100, 200)
	if input.MouseX != 100 || input.MouseY != 200 {
		t.Errorf("mouse pos after set: got (%d,%d), want (100,200)", input.MouseX, input.MouseY)
	}
}

func TestMockInput_GamepadAxes(t *testing.T) {
	input := NewMockInput()

	// Initial axis value should be 0
	if input.GetAxisValue(0) != 0.0 {
		t.Errorf("initial axis 0: got %f, want 0.0", input.GetAxisValue(0))
	}

	// Set axis value
	input.SetAxisValue(0, 0.5)
	if input.GetAxisValue(0) != 0.5 {
		t.Errorf("axis 0 after set: got %f, want 0.5", input.GetAxisValue(0))
	}

	// Set negative value
	input.SetAxisValue(1, -0.75)
	if input.GetAxisValue(1) != -0.75 {
		t.Errorf("axis 1 after set: got %f, want -0.75", input.GetAxisValue(1))
	}
}

func TestMockTextureAtlas_GetSet(t *testing.T) {
	atlas := NewMockTextureAtlas()

	// Initially empty
	_, ok := atlas.Get("wall")
	if ok {
		t.Error("texture should not exist initially")
	}

	// Add texture
	img := CreateSolidImage(64, 64, color.RGBA{255, 0, 0, 255})
	atlas.AddTexture("wall", img)

	// Retrieve texture
	retrieved, ok := atlas.Get("wall")
	if !ok {
		t.Fatal("texture should exist after adding")
	}
	if retrieved != img {
		t.Error("retrieved texture should be the same as added")
	}
}

func TestMockTextureAtlas_Genre(t *testing.T) {
	atlas := NewMockTextureAtlas()

	// Default genre
	if atlas.GenreID != "fantasy" {
		t.Errorf("default genre: got %q, want %q", atlas.GenreID, "fantasy")
	}

	// Set genre
	atlas.SetGenre("scifi")
	if atlas.GenreID != "scifi" {
		t.Errorf("after SetGenre: got %q, want %q", atlas.GenreID, "scifi")
	}
}

func TestMockTextureAtlas_GetAnimatedFrame(t *testing.T) {
	atlas := NewMockTextureAtlas()
	img := CreateSolidImage(64, 64, color.RGBA{0, 255, 0, 255})
	atlas.AddTexture("animated", img)

	// Should return the same image regardless of tick
	frame1, ok1 := atlas.GetAnimatedFrame("animated", 0)
	frame2, ok2 := atlas.GetAnimatedFrame("animated", 100)

	if !ok1 || !ok2 {
		t.Fatal("animated frames should exist")
	}
	if frame1 != img || frame2 != img {
		t.Error("animated frames should return base texture")
	}
}

func TestMockLightMap_GetSet(t *testing.T) {
	lightMap := NewMockLightMap(10, 10)

	// Default brightness
	if lightMap.GetLight(5, 5) != 1.0 {
		t.Errorf("default light: got %f, want 1.0", lightMap.GetLight(5, 5))
	}

	// Set light
	lightMap.SetLight(5, 5, 0.5)
	if lightMap.GetLight(5, 5) != 0.5 {
		t.Errorf("after SetLight: got %f, want 0.5", lightMap.GetLight(5, 5))
	}
}

func TestMockLightMap_OutOfBounds(t *testing.T) {
	lightMap := NewMockLightMap(10, 10)

	tests := []struct {
		name string
		x, y int
	}{
		{"negative x", -1, 5},
		{"negative y", 5, -1},
		{"x too large", 10, 5},
		{"y too large", 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetLight should return 0 for out-of-bounds
			if lightMap.GetLight(tt.x, tt.y) != 0.0 {
				t.Errorf("out-of-bounds GetLight: got %f, want 0.0",
					lightMap.GetLight(tt.x, tt.y))
			}

			// SetLight should not panic for out-of-bounds
			lightMap.SetLight(tt.x, tt.y, 0.5)
		})
	}
}

func TestMockLightMap_Calculate(t *testing.T) {
	lightMap := NewMockLightMap(10, 10)

	// Set some lights
	lightMap.SetLight(0, 0, 0.2)
	lightMap.SetLight(5, 5, 0.8)

	// Calculate is a no-op for mock
	lightMap.Calculate()

	// Lights should remain unchanged
	if lightMap.GetLight(0, 0) != 0.2 {
		t.Errorf("light (0,0) changed after Calculate")
	}
	if lightMap.GetLight(5, 5) != 0.8 {
		t.Errorf("light (5,5) changed after Calculate")
	}
}
