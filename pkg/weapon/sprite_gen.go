// Package weapon implements the weapon and firing system.
package weapon

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// FrameType represents the type of weapon sprite frame to generate.
type FrameType int

const (
	// FrameIdle is the default weapon appearance.
	FrameIdle FrameType = iota
	// FrameFire is the weapon during firing (with muzzle flash).
	FrameFire
	// FrameReload is the weapon during reload animation.
	FrameReload
)

// GenerateWeaponSprite creates a procedural weapon sprite using geometric primitives.
// Returns a 128x128 RGBA image with the weapon centered.
// The sprite is deterministic based on the seed.
func GenerateWeaponSprite(seed int64, weaponType WeaponType, frame FrameType) *image.RGBA {
	const size = 128
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(seed))

	// Generate base weapon geometry
	switch weaponType {
	case TypeMelee:
		generateMeleeWeapon(img, rng, frame)
	case TypeHitscan:
		generateHitscanWeapon(img, rng, frame)
	case TypeProjectile:
		generateProjectileWeapon(img, rng, frame)
	}

	return img
}

// generateMeleeWeapon draws a melee weapon (knife, sword, fist).
func generateMeleeWeapon(img *image.RGBA, rng *rand.Rand, frame FrameType) {
	// Vary colors slightly based on RNG
	metalVariation := uint8(rng.Intn(20))
	bladeColor := color.RGBA{R: 180 + metalVariation, G: 180 + metalVariation, B: 200, A: 255}
	handleColor := color.RGBA{R: 100, G: 60, B: 30 + uint8(rng.Intn(30)), A: 255}

	// Handle (bottom portion)
	fillRect(img, 50, 90, 78, 120, handleColor)

	// Guard (cross piece)
	fillRect(img, 40, 88, 88, 94, color.RGBA{R: 120, G: 80, B: 40, A: 255})

	// Blade (tapered)
	for y := 30; y < 90; y++ {
		width := 14 - (90-y)/8
		if width < 2 {
			width = 2
		}
		x1 := 64 - width/2
		x2 := 64 + width/2
		for x := x1; x < x2; x++ {
			img.Set(x, y, bladeColor)
		}
	}

	// Blade edge highlight
	for y := 30; y < 90; y++ {
		img.Set(64-1, y, color.RGBA{R: 220, G: 220, B: 240, A: 255})
	}

	// Pommel
	fillCircle(img, 64, 122, 6, color.RGBA{R: 140, G: 90, B: 50, A: 255})
}

// generateHitscanWeapon draws a gun (pistol, rifle, shotgun).
func generateHitscanWeapon(img *image.RGBA, rng *rand.Rand, frame FrameType) {
	baseColor := color.RGBA{R: 40, G: 40, B: 45, A: 255}
	highlightColor := color.RGBA{R: 80, G: 80, B: 90, A: 255}

	// Barrel (horizontal rectangle)
	fillRect(img, 64, 50, 110, 58, baseColor)
	fillRect(img, 64, 50, 110, 52, highlightColor) // Barrel top highlight

	// Receiver (main body)
	fillRect(img, 45, 54, 75, 75, baseColor)
	fillRect(img, 45, 54, 75, 58, highlightColor)

	// Grip (angled handle)
	fillRect(img, 50, 75, 60, 95, color.RGBA{R: 60, G: 40, B: 30, A: 255})

	// Trigger guard
	drawLine(img, 56, 72, 56, 78, color.RGBA{R: 100, G: 100, B: 110, A: 255})
	drawLine(img, 56, 78, 62, 78, color.RGBA{R: 100, G: 100, B: 110, A: 255})

	// Muzzle flash (only during fire frame)
	if frame == FrameFire {
		flashColor := color.RGBA{R: 255, G: 240, B: 100, A: 255}
		fillCircle(img, 110, 54, 8, flashColor)
		// Flash rays
		for i := 0; i < 6; i++ {
			angle := float64(i) * math.Pi / 3
			x := 110 + int(math.Cos(angle)*15)
			y := 54 + int(math.Sin(angle)*15)
			drawLine(img, 110, 54, x, y, flashColor)
		}
	}

	// Sight (front post)
	fillRect(img, 100, 46, 102, 50, color.RGBA{R: 80, G: 80, B: 90, A: 255})
}

// generateProjectileWeapon draws a launcher (rocket, plasma).
func generateProjectileWeapon(img *image.RGBA, rng *rand.Rand, frame FrameType) {
	bodyColor := color.RGBA{R: 50, G: 50, B: 55, A: 255}
	accentColor := color.RGBA{R: 120, G: 40, B: 40, A: 255}

	// Large barrel (tube)
	fillRect(img, 50, 45, 115, 70, bodyColor)
	fillRect(img, 50, 45, 115, 50, color.RGBA{R: 90, G: 90, B: 100, A: 255}) // Top highlight

	// Barrel opening (circle at end)
	fillCircle(img, 115, 57, 12, color.RGBA{R: 20, G: 20, B: 25, A: 255})
	fillCircle(img, 115, 57, 10, bodyColor)

	// Grip underneath
	fillRect(img, 55, 70, 70, 90, color.RGBA{R: 60, G: 40, B: 30, A: 255})

	// Energy coil accent (for plasma weapons)
	for i := 0; i < 5; i++ {
		x := 60 + i*10
		fillCircle(img, x, 57, 3, accentColor)
	}

	// Muzzle glow (during fire)
	if frame == FrameFire {
		glowColor := color.RGBA{R: 255, G: 200, B: 100, A: 255}
		fillCircle(img, 115, 57, 15, color.RGBA{R: 200, G: 150, B: 80, A: 128})
		fillCircle(img, 115, 57, 8, glowColor)
	}
}

// fillRect fills a rectangle with the given color.
func fillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

// fillCircle fills a circle with the given color.
func fillCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// drawLine draws a line from (x1, y1) to (x2, y2).
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
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
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
