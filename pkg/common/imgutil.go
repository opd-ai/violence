// Package common provides shared utility functions used across multiple packages
// to eliminate code duplication.
package common

import (
	"image"
	"image/color"
)

// Abs returns the absolute value of an integer.
// This function consolidates duplicate abs implementations across multiple packages.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// FillRect fills a rectangle with the given color.
// The rectangle is defined by top-left (x1, y1) and bottom-right (x2, y2).
// This function consolidates duplicate fillRect implementations from:
// - pkg/sprite/sprite.go
// - pkg/ai/sprite_gen.go
// - pkg/weapon/sprite_gen.go
// - pkg/itemicon/system.go
func FillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	bounds := img.Bounds()
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < bounds.Dx() && y >= 0 && y < bounds.Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

// FillCircle fills a circle with the given color.
// The circle is centered at (cx, cy) with the given radius.
// This function consolidates duplicate fillCircle implementations from:
// - pkg/sprite/sprite.go
// - pkg/ai/sprite_gen.go
// - pkg/weapon/sprite_gen.go
// - pkg/itemicon/system.go
func FillCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	bounds := img.Bounds()
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				if x >= 0 && x < bounds.Dx() && y >= 0 && y < bounds.Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// DrawLine draws a line from (x1, y1) to (x2, y2) using Bresenham's algorithm.
// The thickness parameter specifies line width in pixels.
// This function consolidates duplicate drawLine implementations from:
// - pkg/ai/sprite_gen.go
// - pkg/weapon/sprite_gen.go
// - pkg/itemicon/system.go
func DrawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA, thickness int) {
	if thickness <= 1 {
		drawLineThin(img, x1, y1, x2, y2, c)
	} else {
		DrawThickLine(img, x1, y1, x2, y2, thickness, c)
	}
}

// drawLineThin draws a 1-pixel line using Bresenham's algorithm.
func drawLineThin(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	bounds := img.Bounds()
	dx := Abs(x2 - x1)
	dy := Abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < bounds.Dx() && y1 >= 0 && y1 < bounds.Dy() {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// DrawThickLine draws a thick line from (x1, y1) to (x2, y2).
// The line has the given thickness in pixels.
// This function consolidates the drawThickLine implementation from pkg/sprite/sprite.go.
func DrawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.RGBA) {
	bounds := img.Bounds()
	dx := Abs(x2 - x1)
	dy := Abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		for dt := -thickness / 2; dt <= thickness/2; dt++ {
			for dp := -thickness / 2; dp <= thickness/2; dp++ {
				px, py := x1+dt, y1+dp
				if px >= 0 && px < bounds.Dx() && py >= 0 && py < bounds.Dy() {
					img.Set(px, py, c)
				}
			}
		}

		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// DrawRect draws a rectangle outline with the given color.
// The rectangle is defined by top-left (x1, y1) and bottom-right (x2, y2).
// This function consolidates the drawRect implementation from pkg/itemicon/system.go.
func DrawRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	bounds := img.Bounds()
	for x := x1; x < x2; x++ {
		if x >= 0 && x < bounds.Dx() {
			if y1 >= 0 && y1 < bounds.Dy() {
				img.Set(x, y1, c)
			}
			if y2-1 >= 0 && y2-1 < bounds.Dy() {
				img.Set(x, y2-1, c)
			}
		}
	}
	for y := y1; y < y2; y++ {
		if y >= 0 && y < bounds.Dy() {
			if x1 >= 0 && x1 < bounds.Dx() {
				img.Set(x1, y, c)
			}
			if x2-1 >= 0 && x2-1 < bounds.Dx() {
				img.Set(x2-1, y, c)
			}
		}
	}
}

// FillEllipse fills an ellipse with the given color.
// The ellipse is centered at (cx, cy) with horizontal radius rx and vertical radius ry.
func FillEllipse(img *image.RGBA, cx, cy, rx, ry int, c color.RGBA) {
	bounds := img.Bounds()
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			if (dx*dx)/(float64(rx*rx))+(dy*dy)/(float64(ry*ry)) <= 1.0 {
				if x >= 0 && x < bounds.Dx() && y >= 0 && y < bounds.Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// FillTriangle fills a triangle with the given color.
// The triangle is defined by three vertices (x1, y1), (x2, y2), (x3, y3).
func FillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, c color.RGBA) {
	bounds := img.Bounds()

	// Find bounding box
	minX := min(x1, min(x2, x3))
	maxX := max(x1, max(x2, x3))
	minY := min(y1, min(y2, y3))
	maxY := max(y1, max(y2, y3))

	// Scan and fill
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if isPointInTriangle(x, y, x1, y1, x2, y2, x3, y3) {
				if x >= 0 && x < bounds.Dx() && y >= 0 && y < bounds.Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// isPointInTriangle checks if point (px, py) is inside triangle (x1,y1), (x2,y2), (x3,y3).
func isPointInTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	// Use barycentric coordinates
	denominator := ((y2-y3)*(x1-x3) + (x3-x2)*(y1-y3))
	if denominator == 0 {
		return false
	}

	a := ((y2-y3)*(px-x3) + (x3-x2)*(py-y3)) / denominator
	b := ((y3-y1)*(px-x3) + (x1-x3)*(py-y3)) / denominator
	c := 1 - a - b

	return a >= 0 && b >= 0 && c >= 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
