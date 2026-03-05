package common

import (
	"image"
	"image/color"
	"testing"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: 5, want: 5},
		{input: -5, want: 5},
		{input: 0, want: 0},
		{input: 100, want: 100},
		{input: -100, want: 100},
	}

	for _, tt := range tests {
		got := Abs(tt.input)
		if got != tt.want {
			t.Errorf("Abs(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFillRect(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	FillRect(img, 10, 10, 20, 20, testColor)

	// Check that rectangle is filled
	for y := 10; y < 20; y++ {
		for x := 10; x < 20; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
				t.Errorf("pixel at (%d,%d) not filled correctly", x, y)
			}
		}
	}

	// Check that pixels outside rectangle are not filled
	c := img.At(5, 5)
	r, g, b, a := c.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("pixel outside rectangle should be transparent")
	}
}

func TestFillCircle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	FillCircle(img, 50, 50, 10, testColor)

	// Check center pixel is filled
	c := img.At(50, 50)
	r, g, b, a := c.RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Error("circle center pixel not filled correctly")
	}

	// Check a pixel inside radius (should be filled)
	c = img.At(55, 50)
	r, g, b, a = c.RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Error("circle inner pixel not filled correctly")
	}

	// Check a pixel outside radius (should not be filled)
	c = img.At(65, 50)
	r, g, b, a = c.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("pixel outside circle should be transparent")
	}
}

func TestDrawLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	DrawLine(img, 10, 10, 30, 30, testColor, 1)

	// Check that line endpoints are drawn
	c1 := img.At(10, 10)
	r1, g1, b1, a1 := c1.RGBA()
	if r1>>8 != 0 || g1>>8 != 0 || b1>>8 != 255 || a1>>8 != 255 {
		t.Error("line start pixel not drawn correctly")
	}

	c2 := img.At(30, 30)
	r2, g2, b2, a2 := c2.RGBA()
	if r2>>8 != 0 || g2>>8 != 0 || b2>>8 != 255 || a2>>8 != 255 {
		t.Error("line end pixel not drawn correctly")
	}

	// Check a midpoint pixel (diagonal line, so 20,20 should be on it)
	c3 := img.At(20, 20)
	r3, g3, b3, a3 := c3.RGBA()
	if r3>>8 != 0 || g3>>8 != 0 || b3>>8 != 255 || a3>>8 != 255 {
		t.Error("line midpoint pixel not drawn correctly")
	}
}

func TestDrawThickLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	DrawThickLine(img, 20, 20, 40, 40, 3, testColor)

	// Check that line start area has filled pixels
	c := img.At(20, 20)
	r, g, b, a := c.RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Error("thick line start pixel not drawn correctly")
	}

	// Check adjacent pixels due to thickness
	c = img.At(21, 20)
	r, g, b, a = c.RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Error("thick line adjacent pixel not drawn correctly")
	}
}

func TestDrawRect(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	DrawRect(img, 10, 10, 30, 30, testColor)

	// Check that corners are drawn
	c := img.At(10, 10)
	r, g, b, a := c.RGBA()
	if r>>8 != 128 || g>>8 != 128 || b>>8 != 128 || a>>8 != 255 {
		t.Error("rectangle top-left corner not drawn correctly")
	}

	c = img.At(29, 10)
	r, g, b, a = c.RGBA()
	if r>>8 != 128 || g>>8 != 128 || b>>8 != 128 || a>>8 != 255 {
		t.Error("rectangle top-right corner not drawn correctly")
	}

	// Check that center is empty (it's an outline)
	c = img.At(20, 20)
	r, g, b, a = c.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("rectangle center should be empty (outline only)")
	}
}

func TestBoundsChecking(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	testColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// These should not panic even with out-of-bounds coordinates
	FillRect(img, -10, -10, 60, 60, testColor)
	FillCircle(img, 0, 0, 100, testColor)
	DrawLine(img, -10, -10, 60, 60, testColor, 1)
	DrawThickLine(img, -10, -10, 60, 60, 5, testColor)
	DrawRect(img, -10, -10, 60, 60, testColor)
	FillEllipse(img, 25, 25, 100, 100, testColor)
	FillTriangle(img, -10, -10, 60, 60, 25, 25, testColor)

	// If we get here without panicking, the bounds checking works
}

func TestFillEllipse(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 255, G: 128, B: 0, A: 255}

	FillEllipse(img, 50, 50, 20, 10, testColor)

	// Check center pixel is filled
	c := img.At(50, 50)
	r, g, b, a := c.RGBA()
	if r>>8 != 255 || g>>8 != 128 || b>>8 != 0 || a>>8 != 255 {
		t.Error("ellipse center pixel not filled correctly")
	}

	// Check a pixel on the horizontal axis (should be filled)
	c = img.At(60, 50)
	r, g, b, a = c.RGBA()
	if r>>8 != 255 || g>>8 != 128 || b>>8 != 0 || a>>8 != 255 {
		t.Error("ellipse horizontal pixel not filled correctly")
	}

	// Check a pixel outside ellipse (should not be filled)
	c = img.At(75, 50)
	r, g, b, a = c.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("pixel outside ellipse should be transparent")
	}
}

func TestFillTriangle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColor := color.RGBA{R: 128, G: 255, B: 128, A: 255}

	FillTriangle(img, 50, 20, 30, 60, 70, 60, testColor)

	// Check a pixel inside the triangle (centroid)
	centroidX := (50 + 30 + 70) / 3
	centroidY := (20 + 60 + 60) / 3
	c := img.At(centroidX, centroidY)
	r, g, b, a := c.RGBA()
	if r>>8 != 128 || g>>8 != 255 || b>>8 != 128 || a>>8 != 255 {
		t.Error("triangle centroid pixel not filled correctly")
	}

	// Check a pixel outside the triangle
	c = img.At(10, 10)
	r, g, b, a = c.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("pixel outside triangle should be transparent")
	}
}
