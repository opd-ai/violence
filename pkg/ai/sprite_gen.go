// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"image"
	"image/color"
	"math/rand"
)

// AnimFrame represents an enemy animation frame type.
type AnimFrame int

const (
	// AnimFrameIdle is the standing pose.
	AnimFrameIdle AnimFrame = iota
	// AnimFrameWalk1 is the first walk cycle frame.
	AnimFrameWalk1
	// AnimFrameWalk2 is the second walk cycle frame.
	AnimFrameWalk2
	// AnimFrameAttack is the attacking pose.
	AnimFrameAttack
	// AnimFrameDeath is the death pose.
	AnimFrameDeath
)

// EnemyArchetype defines enemy visual characteristics.
type EnemyArchetype string

const (
	// ArchetypeFantasyGuard is a medieval guard.
	ArchetypeFantasyGuard EnemyArchetype = "fantasy_guard"
	// ArchetypeSciFiSoldier is a futuristic soldier.
	ArchetypeSciFiSoldier EnemyArchetype = "scifi_soldier"
	// ArchetypeHorrorCultist is a horror cultist.
	ArchetypeHorrorCultist EnemyArchetype = "horror_cultist"
	// ArchetypeCyberpunkDrone is a cyberpunk drone.
	ArchetypeCyberpunkDrone EnemyArchetype = "cyberpunk_drone"
	// ArchetypePostapocScavenger is a post-apocalyptic scavenger.
	ArchetypePostapocScavenger EnemyArchetype = "postapoc_scavenger"
)

// GenerateEnemySprite creates a procedural enemy sprite using body part composition.
// Returns a 64x64 RGBA image with the enemy centered, facing forward.
// The sprite is deterministic based on the seed and archetype.
func GenerateEnemySprite(seed int64, archetype EnemyArchetype, frame AnimFrame) *image.RGBA {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(seed))

	// Generate sprite based on archetype
	switch archetype {
	case ArchetypeFantasyGuard:
		generateFantasyGuard(img, rng, frame)
	case ArchetypeSciFiSoldier:
		generateSciFiSoldier(img, rng, frame)
	case ArchetypeHorrorCultist:
		generateHorrorCultist(img, rng, frame)
	case ArchetypeCyberpunkDrone:
		generateCyberpunkDrone(img, rng, frame)
	case ArchetypePostapocScavenger:
		generatePostapocScavenger(img, rng, frame)
	default:
		generateFantasyGuard(img, rng, frame)
	}

	return img
}

// generateFantasyGuard draws a medieval guard with armor and sword.
func generateFantasyGuard(img *image.RGBA, rng *rand.Rand, frame AnimFrame) {
	armorColor := color.RGBA{R: 120, G: 120, B: 130, A: 255}
	skinColor := color.RGBA{R: 210, G: 180, B: 160, A: 255}
	helmetColor := color.RGBA{R: 140, G: 140, B: 150, A: 255}

	// Calculate animation offsets
	leftLegY := 35
	rightLegY := 35
	leftArmY := 22
	rightArmY := 22
	if frame == AnimFrameWalk1 {
		leftLegY += 2
		rightLegY -= 2
	} else if frame == AnimFrameWalk2 {
		leftLegY -= 2
		rightLegY += 2
	} else if frame == AnimFrameAttack {
		rightArmY -= 3
	}

	// Legs
	fillRect(img, 26, leftLegY, 30, 50, armorColor)  // Left leg
	fillRect(img, 34, rightLegY, 38, 50, armorColor) // Right leg

	// Body (torso)
	fillRect(img, 24, 18, 40, 35, armorColor)

	// Arms
	fillRect(img, 18, leftArmY, 24, 32, armorColor)  // Left arm
	fillRect(img, 40, rightArmY, 46, 32, armorColor) // Right arm

	// Head
	fillCircle(img, 32, 12, 6, skinColor)

	// Helmet
	fillRect(img, 28, 8, 36, 12, helmetColor)
	fillRect(img, 26, 8, 38, 10, helmetColor) // Helmet brim

	// Shield (left hand)
	fillRect(img, 14, 24, 20, 34, color.RGBA{R: 160, G: 140, B: 100, A: 255})

	// Sword (right hand)
	if frame == AnimFrameAttack {
		fillRect(img, 44, 10, 46, 28, color.RGBA{R: 200, G: 200, B: 220, A: 255})
	} else {
		fillRect(img, 44, 26, 46, 38, color.RGBA{R: 200, G: 200, B: 220, A: 255})
	}
}

// generateSciFiSoldier draws a futuristic soldier with armor and gun.
func generateSciFiSoldier(img *image.RGBA, rng *rand.Rand, frame AnimFrame) {
	armorColor := color.RGBA{R: 40, G: 60, B: 80, A: 255}
	visorColor := color.RGBA{R: 100, G: 180, B: 255, A: 255}
	accentColor := color.RGBA{R: 80, G: 200, B: 240, A: 255}

	// Animation offsets
	leftLegY := 35
	rightLegY := 35
	if frame == AnimFrameWalk1 {
		leftLegY += 2
		rightLegY -= 2
	} else if frame == AnimFrameWalk2 {
		leftLegY -= 2
		rightLegY += 2
	}

	// Legs
	fillRect(img, 26, leftLegY, 30, 52, armorColor)
	fillRect(img, 34, rightLegY, 38, 52, armorColor)

	// Body
	fillRect(img, 24, 18, 40, 36, armorColor)
	// Chest accent
	fillRect(img, 28, 22, 36, 24, accentColor)

	// Arms
	fillRect(img, 18, 22, 24, 34, armorColor)
	fillRect(img, 40, 22, 46, 34, armorColor)

	// Helmet
	fillCircle(img, 32, 12, 7, armorColor)
	// Visor
	fillRect(img, 28, 10, 36, 14, visorColor)

	// Weapon (rifle across chest)
	weaponY := 24
	if frame == AnimFrameAttack {
		weaponY = 20
	}
	fillRect(img, 18, weaponY, 46, weaponY+2, color.RGBA{R: 60, G: 60, B: 70, A: 255})

	// Shoulder pads
	fillRect(img, 17, 18, 24, 22, color.RGBA{R: 60, G: 80, B: 100, A: 255})
	fillRect(img, 40, 18, 47, 22, color.RGBA{R: 60, G: 80, B: 100, A: 255})
}

// generateHorrorCultist draws a horror cultist with robes.
func generateHorrorCultist(img *image.RGBA, rng *rand.Rand, frame AnimFrame) {
	robeColor := color.RGBA{R: 60, G: 20, B: 20, A: 255}
	skinColor := color.RGBA{R: 180, G: 170, B: 160, A: 255}

	// Animation sway
	bodyX := 32
	if frame == AnimFrameWalk1 {
		bodyX += 1
	} else if frame == AnimFrameWalk2 {
		bodyX -= 1
	}

	// Robe (wide bottom, tapered top)
	for y := 50; y >= 20; y-- {
		width := 20 - (50-y)/3
		if width < 8 {
			width = 8
		}
		fillRect(img, bodyX-width/2, y, bodyX+width/2, y+1, robeColor)
	}

	// Arms (extended, menacing)
	armY := 26
	if frame == AnimFrameAttack {
		armY = 22
	}
	fillRect(img, 12, armY, 24, armY+2, robeColor) // Left arm
	fillRect(img, 40, armY, 52, armY+2, robeColor) // Right arm

	// Hands (pale)
	fillCircle(img, 12, armY, 2, skinColor)
	fillCircle(img, 52, armY, 2, skinColor)

	// Hood
	fillCircle(img, 32, 12, 8, robeColor)

	// Face (shadowed, only eyes visible)
	fillCircle(img, 30, 12, 1, color.RGBA{R: 255, G: 50, B: 50, A: 255}) // Left eye (red)
	fillCircle(img, 34, 12, 1, color.RGBA{R: 255, G: 50, B: 50, A: 255}) // Right eye (red)

	// Ritual dagger (in right hand, attack frame)
	if frame == AnimFrameAttack {
		fillRect(img, 50, armY, 56, armY+1, color.RGBA{R: 180, G: 180, B: 190, A: 255})
	}
}

// generateCyberpunkDrone draws a hovering cybernetic drone.
func generateCyberpunkDrone(img *image.RGBA, rng *rand.Rand, frame AnimFrame) {
	bodyColor := color.RGBA{R: 30, G: 30, B: 35, A: 255}
	neonColor := color.RGBA{R: 255, G: 0, B: 128, A: 255}
	cyanColor := color.RGBA{R: 0, G: 200, B: 255, A: 255}

	// Hover offset (bobbing animation)
	hoverY := 0
	if frame == AnimFrameWalk1 {
		hoverY = -1
	} else if frame == AnimFrameWalk2 {
		hoverY = 1
	}

	// Main body (sphere)
	fillCircle(img, 32, 28+hoverY, 10, bodyColor)

	// Eye/sensor (central glowing circle)
	eyeRadius := 4
	if frame == AnimFrameAttack {
		eyeRadius = 6 // Expand when attacking
	}
	fillCircle(img, 32, 28+hoverY, eyeRadius, neonColor)

	// Weapon pods (left and right)
	fillRect(img, 16, 26+hoverY, 22, 30+hoverY, bodyColor)
	fillRect(img, 42, 26+hoverY, 48, 30+hoverY, bodyColor)

	// Weapon barrels
	fillCircle(img, 16, 28+hoverY, 2, color.RGBA{R: 80, G: 80, B: 90, A: 255})
	fillCircle(img, 48, 28+hoverY, 2, color.RGBA{R: 80, G: 80, B: 90, A: 255})

	// Cyan accent lights
	fillCircle(img, 32, 20+hoverY, 2, cyanColor)
	fillCircle(img, 32, 36+hoverY, 2, cyanColor)

	// Muzzle flash (attack frame)
	if frame == AnimFrameAttack {
		fillCircle(img, 16, 28+hoverY, 4, color.RGBA{R: 255, G: 200, B: 0, A: 255})
		fillCircle(img, 48, 28+hoverY, 4, color.RGBA{R: 255, G: 200, B: 0, A: 255})
	}

	// Propulsion glow (bottom)
	fillCircle(img, 32, 40+hoverY, 3, cyanColor)
}

// generatePostapocScavenger draws a post-apocalyptic scavenger with makeshift armor.
func generatePostapocScavenger(img *image.RGBA, rng *rand.Rand, frame AnimFrame) {
	scrapColor := color.RGBA{R: 100, G: 80, B: 60, A: 255}
	clothColor := color.RGBA{R: 80, G: 70, B: 60, A: 255}
	skinColor := color.RGBA{R: 190, G: 160, B: 140, A: 255}

	// Animation
	leftLegY := 35
	rightLegY := 35
	if frame == AnimFrameWalk1 {
		leftLegY += 2
		rightLegY -= 2
	} else if frame == AnimFrameWalk2 {
		leftLegY -= 2
		rightLegY += 2
	}

	// Legs (ragged pants)
	fillRect(img, 26, leftLegY, 30, 52, clothColor)
	fillRect(img, 34, rightLegY, 38, 52, clothColor)

	// Body (scrap armor plates)
	fillRect(img, 24, 20, 40, 36, scrapColor)

	// Arms
	fillRect(img, 18, 24, 24, 36, clothColor) // Left arm
	fillRect(img, 40, 24, 46, 36, scrapColor) // Right arm (armored)

	// Head (bandana/mask)
	fillCircle(img, 32, 12, 6, skinColor)
	fillRect(img, 28, 14, 36, 16, clothColor) // Bandana over mouth

	// Goggles
	fillCircle(img, 29, 10, 2, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	fillCircle(img, 35, 10, 2, color.RGBA{R: 50, G: 50, B: 50, A: 255})

	// Weapon (makeshift pipe gun)
	weaponY := 28
	if frame == AnimFrameAttack {
		weaponY = 24
	}
	fillRect(img, 40, weaponY, 54, weaponY+2, color.RGBA{R: 70, G: 60, B: 50, A: 255})
	fillCircle(img, 54, weaponY+1, 2, color.RGBA{R: 90, G: 80, B: 70, A: 255})

	// Scrap metal shoulder pad
	fillRect(img, 40, 20, 48, 24, scrapColor)
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
