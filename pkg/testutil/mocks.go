// Package testutil provides mock interfaces and test helpers for testing Violence components.
package testutil

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// MockScreen implements a basic in-memory screen for testing rendering.
// It satisfies the subset of ebiten.Image methods used in tests.
type MockScreen struct {
	Width  int
	Height int
	Pixels []byte
}

// NewMockScreen creates a mock screen with the given dimensions.
func NewMockScreen(width, height int) *MockScreen {
	return &MockScreen{
		Width:  width,
		Height: height,
		Pixels: make([]byte, width*height*4),
	}
}

// DrawImage records that DrawImage was called (no-op for testing).
func (m *MockScreen) DrawImage(img *ebiten.Image, opts *ebiten.DrawImageOptions) {
	// No-op: tests can override if needed
}

// Fill sets all pixels to the specified color.
func (m *MockScreen) Fill(clr color.Color) {
	r, g, b, a := clr.RGBA()
	// Convert from 16-bit to 8-bit color components
	r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
	for i := 0; i < len(m.Pixels); i += 4 {
		m.Pixels[i] = r8
		m.Pixels[i+1] = g8
		m.Pixels[i+2] = b8
		m.Pixels[i+3] = a8
	}
}

// Bounds returns the image bounds.
func (m *MockScreen) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.Width, m.Height)
}

// MockInput provides a controllable input state for testing.
type MockInput struct {
	PressedKeys    map[ebiten.Key]bool
	PressedButtons map[ebiten.GamepadButton]bool
	MouseX         int
	MouseY         int
	GamepadAxes    map[int]float64
	GamepadID      ebiten.GamepadID
}

// NewMockInput creates a mock input state.
func NewMockInput() *MockInput {
	return &MockInput{
		PressedKeys:    make(map[ebiten.Key]bool),
		PressedButtons: make(map[ebiten.GamepadButton]bool),
		GamepadAxes:    make(map[int]float64),
		GamepadID:      0,
	}
}

// PressKey marks a key as pressed.
func (m *MockInput) PressKey(key ebiten.Key) {
	m.PressedKeys[key] = true
}

// ReleaseKey marks a key as released.
func (m *MockInput) ReleaseKey(key ebiten.Key) {
	delete(m.PressedKeys, key)
}

// IsKeyPressed returns whether a key is currently pressed.
func (m *MockInput) IsKeyPressed(key ebiten.Key) bool {
	return m.PressedKeys[key]
}

// PressButton marks a gamepad button as pressed.
func (m *MockInput) PressButton(btn ebiten.GamepadButton) {
	m.PressedButtons[btn] = true
}

// ReleaseButton marks a gamepad button as released.
func (m *MockInput) ReleaseButton(btn ebiten.GamepadButton) {
	delete(m.PressedButtons, btn)
}

// IsButtonPressed returns whether a gamepad button is currently pressed.
func (m *MockInput) IsButtonPressed(btn ebiten.GamepadButton) bool {
	return m.PressedButtons[btn]
}

// SetMousePosition sets the mock mouse position.
func (m *MockInput) SetMousePosition(x, y int) {
	m.MouseX = x
	m.MouseY = y
}

// SetAxisValue sets a gamepad axis value (-1.0 to 1.0).
func (m *MockInput) SetAxisValue(axisID int, value float64) {
	m.GamepadAxes[axisID] = value
}

// GetAxisValue returns a gamepad axis value.
func (m *MockInput) GetAxisValue(axisID int) float64 {
	return m.GamepadAxes[axisID]
}

// MockTextureAtlas provides a simple in-memory texture atlas for testing.
type MockTextureAtlas struct {
	Textures map[string]image.Image
	GenreID  string
}

// NewMockTextureAtlas creates a mock texture atlas.
func NewMockTextureAtlas() *MockTextureAtlas {
	return &MockTextureAtlas{
		Textures: make(map[string]image.Image),
		GenreID:  "fantasy",
	}
}

// Get retrieves a texture by name.
func (m *MockTextureAtlas) Get(name string) (image.Image, bool) {
	img, ok := m.Textures[name]
	return img, ok
}

// GetAnimatedFrame retrieves an animated texture frame.
func (m *MockTextureAtlas) GetAnimatedFrame(name string, tick int) (image.Image, bool) {
	// For mock, just return the base texture
	return m.Get(name)
}

// SetGenre sets the current genre.
func (m *MockTextureAtlas) SetGenre(genreID string) {
	m.GenreID = genreID
}

// AddTexture adds a texture to the mock atlas.
func (m *MockTextureAtlas) AddTexture(name string, img image.Image) {
	m.Textures[name] = img
}

// MockLightMap provides a simple in-memory light map for testing.
type MockLightMap struct {
	Width  int
	Height int
	Lights [][]float64
}

// NewMockLightMap creates a mock light map with uniform lighting.
func NewMockLightMap(width, height int) *MockLightMap {
	lights := make([][]float64, height)
	for y := range lights {
		lights[y] = make([]float64, width)
		for x := range lights[y] {
			lights[y][x] = 1.0 // Default: full brightness
		}
	}
	return &MockLightMap{
		Width:  width,
		Height: height,
		Lights: lights,
	}
}

// GetLight returns the light value at a tile position.
func (m *MockLightMap) GetLight(x, y int) float64 {
	if y < 0 || y >= m.Height || x < 0 || x >= m.Width {
		return 0.0
	}
	return m.Lights[y][x]
}

// SetLight sets the light value at a tile position.
func (m *MockLightMap) SetLight(x, y int, value float64) {
	if y >= 0 && y < m.Height && x >= 0 && x < m.Width {
		m.Lights[y][x] = value
	}
}

// Calculate is a no-op for the mock (lights are set directly).
func (m *MockLightMap) Calculate() {
	// No-op: tests set lights directly via SetLight
}
