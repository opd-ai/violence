// Package testutil provides test helpers for common testing patterns.
package testutil

import (
	"image"
	"image/color"
	"math"
)

// TestingT is a minimal interface satisfied by *testing.T and *testing.B.
type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
}

// AssertFloatEqual checks if two float64 values are equal within epsilon.
func AssertFloatEqual(t TestingT, got, want, epsilon float64, msgAndArgs ...interface{}) {
	t.Helper()
	if math.Abs(got-want) > epsilon {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: got %f, want %f (epsilon %f)", msgAndArgs[0], got, want, epsilon)
		} else {
			t.Errorf("got %f, want %f (epsilon %f)", got, want, epsilon)
		}
	}
}

// AssertIntEqual checks if two int values are equal.
func AssertIntEqual(t TestingT, got, want int, msgAndArgs ...interface{}) {
	t.Helper()
	if got != want {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: got %d, want %d", msgAndArgs[0], got, want)
		} else {
			t.Errorf("got %d, want %d", got, want)
		}
	}
}

// AssertStringEqual checks if two string values are equal.
func AssertStringEqual(t TestingT, got, want string, msgAndArgs ...interface{}) {
	t.Helper()
	if got != want {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: got %q, want %q", msgAndArgs[0], got, want)
		} else {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// AssertTrue checks if a boolean is true.
func AssertTrue(t TestingT, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !condition {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected true, got false", msgAndArgs[0])
		} else {
			t.Error("expected true, got false")
		}
	}
}

// AssertFalse checks if a boolean is false.
func AssertFalse(t TestingT, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if condition {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected false, got true", msgAndArgs[0])
		} else {
			t.Error("expected false, got true")
		}
	}
}

// AssertNil checks if a value is nil.
func AssertNil(t TestingT, val interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if val != nil {
		// Check for typed nil (e.g., (*int)(nil))
		// Using reflection to handle typed nil pointers
		if isNil(val) {
			return
		}
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected nil, got %v", msgAndArgs[0], val)
		} else {
			t.Errorf("expected nil, got %v", val)
		}
	}
}

// isNil checks if a value is nil, including typed nil pointers
func isNil(val interface{}) bool {
	if val == nil {
		return true
	}
	// Use type assertion to check for common nil pointer types
	switch v := val.(type) {
	case *int:
		return v == nil
	case *string:
		return v == nil
	case *bool:
		return v == nil
	case *float64:
		return v == nil
	default:
		return false
	}
}

// AssertNotNil checks if a value is not nil.
func AssertNotNil(t TestingT, val interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if val == nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected non-nil value", msgAndArgs[0])
		} else {
			t.Error("expected non-nil value")
		}
	}
}

// AssertPanic checks that a function panics when called.
func AssertPanic(t TestingT, fn func(), msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("%v: expected panic but none occurred", msgAndArgs[0])
			} else {
				t.Error("expected panic but none occurred")
			}
		}
	}()
	fn()
}

// AssertNoPanic checks that a function does not panic when called.
func AssertNoPanic(t TestingT, fn func(), msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("%v: unexpected panic: %v", msgAndArgs[0], r)
			} else {
				t.Errorf("unexpected panic: %v", r)
			}
		}
	}()
	fn()
}

// CreateSolidImage creates a solid color image for testing.
func CreateSolidImage(width, height int, clr color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, clr)
		}
	}
	return img
}

// CreateCheckerboardImage creates a checkerboard pattern image for testing textures.
func CreateCheckerboardImage(width, height, tileSize int, color1, color2 color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if ((x/tileSize)+(y/tileSize))%2 == 0 {
				img.Set(x, y, color1)
			} else {
				img.Set(x, y, color2)
			}
		}
	}
	return img
}

// AssertColorEqual checks if two colors are equal within a tolerance.
func AssertColorEqual(t TestingT, got, want color.Color, tolerance uint8, msgAndArgs ...interface{}) {
	t.Helper()
	r1, g1, b1, a1 := got.RGBA()
	r2, g2, b2, a2 := want.RGBA()

	// Convert from 16-bit to 8-bit
	r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
	r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

	tol := uint32(tolerance)
	if absDiff(r1, r2) > tol || absDiff(g1, g2) > tol || absDiff(b1, b2) > tol || absDiff(a1, a2) > tol {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: colors differ: got RGBA(%d,%d,%d,%d), want RGBA(%d,%d,%d,%d), tolerance %d",
				msgAndArgs[0], r1, g1, b1, a1, r2, g2, b2, a2, tolerance)
		} else {
			t.Errorf("colors differ: got RGBA(%d,%d,%d,%d), want RGBA(%d,%d,%d,%d), tolerance %d",
				r1, g1, b1, a1, r2, g2, b2, a2, tolerance)
		}
	}
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
