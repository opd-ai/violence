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
// This function consolidates duplicate drawLine implementations from:
// - pkg/ai/sprite_gen.go
// - pkg/weapon/sprite_gen.go
// - pkg/itemicon/system.go
func DrawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
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
