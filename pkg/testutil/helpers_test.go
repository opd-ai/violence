package testutil

import (
	"image/color"
	"testing"
)

func TestAssertFloatEqual(t *testing.T) {
	// This is a meta-test: testing the test helper
	// We'll use a mock *testing.T to capture errors

	tests := []struct {
		name      string
		got       float64
		want      float64
		epsilon   float64
		shouldErr bool
	}{
		{"exact match", 1.0, 1.0, 0.001, false},
		{"within epsilon", 1.0, 1.0001, 0.001, false},
		{"outside epsilon", 1.0, 1.1, 0.001, true},
		{"negative values", -5.0, -5.0001, 0.001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertFloatEqual(mockT, tt.got, tt.want, tt.epsilon)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertIntEqual(t *testing.T) {
	tests := []struct {
		name      string
		got       int
		want      int
		shouldErr bool
	}{
		{"equal", 42, 42, false},
		{"not equal", 42, 43, true},
		{"negative", -10, -10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertIntEqual(mockT, tt.got, tt.want)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertStringEqual(t *testing.T) {
	tests := []struct {
		name      string
		got       string
		want      string
		shouldErr bool
	}{
		{"equal", "hello", "hello", false},
		{"not equal", "hello", "world", true},
		{"empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertStringEqual(mockT, tt.got, tt.want)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertTrue(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		shouldErr bool
	}{
		{"true", true, false},
		{"false", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertTrue(mockT, tt.condition)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertFalse(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		shouldErr bool
	}{
		{"false", false, false},
		{"true", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertFalse(mockT, tt.condition)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertNil(t *testing.T) {
	tests := []struct {
		name      string
		val       interface{}
		shouldErr bool
	}{
		{"nil", nil, false},
		{"not nil", "string", true},
		{"nil pointer", (*int)(nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertNil(mockT, tt.val)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertNotNil(t *testing.T) {
	tests := []struct {
		name      string
		val       interface{}
		shouldErr bool
	}{
		{"not nil", "string", false},
		{"nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertNotNil(mockT, tt.val)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

func TestAssertPanic(t *testing.T) {
	mockT := &mockTestingT{}

	// Function that panics should not error
	AssertPanic(mockT, func() {
		panic("test panic")
	})
	if mockT.errored {
		t.Error("AssertPanic should not error when function panics")
	}

	// Function that doesn't panic should error
	mockT2 := &mockTestingT{}
	AssertPanic(mockT2, func() {
		// no panic
	})
	if !mockT2.errored {
		t.Error("AssertPanic should error when function doesn't panic")
	}
}

func TestAssertNoPanic(t *testing.T) {
	mockT := &mockTestingT{}

	// Function that doesn't panic should not error
	AssertNoPanic(mockT, func() {
		// no panic
	})
	if mockT.errored {
		t.Error("AssertNoPanic should not error when function doesn't panic")
	}

	// Function that panics should error
	mockT2 := &mockTestingT{}
	AssertNoPanic(mockT2, func() {
		panic("test panic")
	})
	if !mockT2.errored {
		t.Error("AssertNoPanic should error when function panics")
	}
}

func TestCreateSolidImage(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		color  color.Color
	}{
		{"red square", 10, 10, color.RGBA{255, 0, 0, 255}},
		{"green rect", 20, 10, color.RGBA{0, 255, 0, 255}},
		{"blue tall", 10, 20, color.RGBA{0, 0, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := CreateSolidImage(tt.width, tt.height, tt.color)

			// Check dimensions
			bounds := img.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("bounds: got %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), tt.width, tt.height)
			}

			// Check color at random points
			if img.At(0, 0) != tt.color {
				t.Errorf("color at (0,0): got %v, want %v", img.At(0, 0), tt.color)
			}
			if img.At(tt.width-1, tt.height-1) != tt.color {
				t.Errorf("color at (%d,%d): got %v, want %v",
					tt.width-1, tt.height-1, img.At(tt.width-1, tt.height-1), tt.color)
			}
		})
	}
}

func TestCreateCheckerboardImage(t *testing.T) {
	color1 := color.RGBA{255, 255, 255, 255}
	color2 := color.RGBA{0, 0, 0, 255}

	img := CreateCheckerboardImage(20, 20, 10, color1, color2)

	// Top-left tile should be color1
	if img.At(0, 0) != color1 {
		t.Errorf("(0,0): got %v, want %v", img.At(0, 0), color1)
	}

	// Top-right tile should be color2
	if img.At(10, 0) != color2 {
		t.Errorf("(10,0): got %v, want %v", img.At(10, 0), color2)
	}

	// Bottom-left tile should be color2
	if img.At(0, 10) != color2 {
		t.Errorf("(0,10): got %v, want %v", img.At(0, 10), color2)
	}

	// Bottom-right tile should be color1
	if img.At(10, 10) != color1 {
		t.Errorf("(10,10): got %v, want %v", img.At(10, 10), color1)
	}
}

func TestAssertColorEqual(t *testing.T) {
	tests := []struct {
		name      string
		got       color.Color
		want      color.Color
		tolerance uint8
		shouldErr bool
	}{
		{
			"exact match",
			color.RGBA{255, 0, 0, 255},
			color.RGBA{255, 0, 0, 255},
			0,
			false,
		},
		{
			"within tolerance",
			color.RGBA{255, 0, 0, 255},
			color.RGBA{250, 5, 5, 250},
			10,
			false,
		},
		{
			"outside tolerance",
			color.RGBA{255, 0, 0, 255},
			color.RGBA{200, 0, 0, 255},
			10,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			AssertColorEqual(mockT, tt.got, tt.want, tt.tolerance)
			if mockT.errored != tt.shouldErr {
				t.Errorf("errored=%v, want %v", mockT.errored, tt.shouldErr)
			}
		})
	}
}

// mockTestingT is a minimal mock of *testing.T for testing helpers
type mockTestingT struct {
	errored bool
}

func (m *mockTestingT) Helper() {}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errored = true
}

func (m *mockTestingT) Error(args ...interface{}) {
	m.errored = true
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.errored = true
	panic("fatal")
}

func (m *mockTestingT) Fatal(args ...interface{}) {
	m.errored = true
	panic("fatal")
}
