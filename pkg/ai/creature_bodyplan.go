// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/violence/pkg/common"
)

// BodyPlan defines the anatomical structure of a creature type.
type BodyPlan int

const (
	BodyPlanHumanoid  BodyPlan = iota // Two legs, two arms, upright stance
	BodyPlanQuadruped                 // Four legs, horizontal body (wolves, bears, cats)
	BodyPlanInsect                    // Six legs, segmented body (spiders, beetles)
	BodyPlanSerpent                   // No legs, elongated body (snakes, worms)
	BodyPlanFlying                    // Wings, aerial pose (bats, birds, drakes)
	BodyPlanAmorphous                 // No fixed form (slimes, oozes, elementals)
)

// CreatureType defines specific creature variants within body plans.
type CreatureType string

const (
	// Quadruped variants
	CreatureWolf   CreatureType = "wolf"
	CreatureBear   CreatureType = "bear"
	CreatureLion   CreatureType = "lion"
	CreatureHound  CreatureType = "hound"
	CreatureRaptor CreatureType = "raptor"

	// Insect variants
	CreatureSpider   CreatureType = "spider"
	CreatureBeetle   CreatureType = "beetle"
	CreatureMantis   CreatureType = "mantis"
	CreatureScorpion CreatureType = "scorpion"
	CreatureAnt      CreatureType = "ant"

	// Serpent variants
	CreatureSnake   CreatureType = "snake"
	CreatureWorm    CreatureType = "worm"
	CreatureSerpent CreatureType = "serpent"
	CreatureLamia   CreatureType = "lamia"

	// Flying variants
	CreatureBat   CreatureType = "bat"
	CreatureDrake CreatureType = "drake"
	CreatureHarpy CreatureType = "harpy"
	CreatureWasp  CreatureType = "wasp"

	// Amorphous variants
	CreatureSlime     CreatureType = "slime"
	CreatureOoze      CreatureType = "ooze"
	CreatureElemental CreatureType = "elemental"
	CreatureWraith    CreatureType = "wraith"
)

// GetBodyPlan returns the body plan for a creature type.
func GetBodyPlan(ctype CreatureType) BodyPlan {
	switch ctype {
	case CreatureWolf, CreatureBear, CreatureLion, CreatureHound, CreatureRaptor:
		return BodyPlanQuadruped
	case CreatureSpider, CreatureBeetle, CreatureMantis, CreatureScorpion, CreatureAnt:
		return BodyPlanInsect
	case CreatureSnake, CreatureWorm, CreatureSerpent, CreatureLamia:
		return BodyPlanSerpent
	case CreatureBat, CreatureDrake, CreatureHarpy, CreatureWasp:
		return BodyPlanFlying
	case CreatureSlime, CreatureOoze, CreatureElemental, CreatureWraith:
		return BodyPlanAmorphous
	default:
		return BodyPlanHumanoid
	}
}

// GenerateCreatureSprite creates a sprite using the appropriate body plan.
func GenerateCreatureSprite(seed int64, ctype CreatureType, frame AnimFrame) *image.RGBA {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(seed))

	bodyPlan := GetBodyPlan(ctype)

	switch bodyPlan {
	case BodyPlanQuadruped:
		generateQuadrupedCreature(img, rng, ctype, frame)
	case BodyPlanInsect:
		generateInsectCreature(img, rng, ctype, frame)
	case BodyPlanSerpent:
		generateSerpentCreature(img, rng, ctype, frame)
	case BodyPlanFlying:
		generateFlyingCreature(img, rng, ctype, frame)
	case BodyPlanAmorphous:
		generateAmorphousCreature(img, rng, ctype, frame)
	default:
		// Fallback to simple humanoid if unknown
		generateSimpleHumanoid(img, rng)
	}

	return img
}

// generateQuadrupedCreature draws four-legged creatures.
func generateQuadrupedCreature(img *image.RGBA, rng *rand.Rand, ctype CreatureType, frame AnimFrame) {
	var bodyColor, accentColor, eyeColor color.RGBA
	var size float64

	// Type-specific characteristics
	switch ctype {
	case CreatureWolf:
		bodyColor = color.RGBA{R: 80, G: 80, B: 90, A: 255}
		accentColor = color.RGBA{R: 100, G: 100, B: 110, A: 255}
		eyeColor = color.RGBA{R: 255, G: 200, B: 0, A: 255}
		size = 1.0
	case CreatureBear:
		bodyColor = color.RGBA{R: 60, G: 40, B: 30, A: 255}
		accentColor = color.RGBA{R: 80, G: 60, B: 50, A: 255}
		eyeColor = color.RGBA{R: 40, G: 20, B: 10, A: 255}
		size = 1.3
	case CreatureLion:
		bodyColor = color.RGBA{R: 180, G: 140, B: 80, A: 255}
		accentColor = color.RGBA{R: 200, G: 160, B: 100, A: 255}
		eyeColor = color.RGBA{R: 255, G: 180, B: 0, A: 255}
		size = 1.1
	case CreatureHound:
		bodyColor = color.RGBA{R: 120, G: 100, B: 80, A: 255}
		accentColor = color.RGBA{R: 140, G: 120, B: 100, A: 255}
		eyeColor = color.RGBA{R: 200, G: 100, B: 50, A: 255}
		size = 0.9
	default: // Raptor
		bodyColor = color.RGBA{R: 100, G: 120, B: 80, A: 255}
		accentColor = color.RGBA{R: 120, G: 140, B: 100, A: 255}
		eyeColor = color.RGBA{R: 255, G: 50, B: 50, A: 255}
		size = 0.95
	}

	// Calculate animation offsets
	frontLeftY := int(30 * size)
	frontRightY := int(30 * size)
	backLeftY := int(30 * size)
	backRightY := int(30 * size)

	if frame == AnimFrameWalk1 {
		frontLeftY += 2
		backRightY += 2
		frontRightY -= 2
		backLeftY -= 2
	} else if frame == AnimFrameWalk2 {
		frontLeftY -= 2
		backRightY -= 2
		frontRightY += 2
		backLeftY += 2
	} else if frame == AnimFrameAttack {
		frontLeftY -= 3
		frontRightY -= 3
	}

	// Back legs (draw first for layering)
	legW := int(3 * size)
	common.FillRect(img, 18, backLeftY, 18+legW, int(48*size), bodyColor)
	common.FillRect(img, 43, backRightY, 43+legW, int(48*size), bodyColor)

	// Body (horizontal ellipse)
	bodyX := 32
	bodyY := int(28 * size)
	bodyW := int(16 * size)
	bodyH := int(10 * size)
	common.FillEllipse(img, bodyX, bodyY, bodyW, bodyH, bodyColor)

	// Accent stripe/pattern
	common.FillEllipse(img, bodyX, bodyY-1, bodyW-2, bodyH-2, accentColor)

	// Front legs (in front of body)
	common.FillRect(img, 22, frontLeftY, 22+legW, int(48*size), bodyColor)
	common.FillRect(img, 39, frontRightY, 39+legW, int(48*size), bodyColor)

	// Head
	headX := 32
	headY := int(18 * size)
	if frame == AnimFrameAttack {
		headY -= 2 // Lunge forward
	}
	headRadius := int(7 * size)
	common.FillCircle(img, headX, headY, headRadius, bodyColor)

	// Snout/muzzle
	common.FillEllipse(img, headX+4, headY+2, 4, 3, accentColor)

	// Eyes
	common.FillCircle(img, headX-2, headY-2, 2, eyeColor)
	common.FillCircle(img, headX+2, headY-2, 2, eyeColor)

	// Ears (pointed for wolf/raptor, rounded for bear)
	if ctype == CreatureBear {
		common.FillCircle(img, headX-5, headY-5, 3, bodyColor)
		common.FillCircle(img, headX+5, headY-5, 3, bodyColor)
	} else {
		common.FillTriangle(img, headX-5, headY-5, headX-3, headY-8, headX-7, headY-8, bodyColor)
		common.FillTriangle(img, headX+5, headY-5, headX+3, headY-8, headX+7, headY-8, bodyColor)
	}

	// Tail
	tailX := 32
	tailY := int(32 * size)
	if frame == AnimFrameWalk1 {
		tailX -= 2
	} else if frame == AnimFrameWalk2 {
		tailX += 2
	}
	common.FillRect(img, tailX-1, tailY, tailX+1, tailY+int(12*size), bodyColor)
}

// generateInsectCreature draws multi-legged arthropod creatures.
func generateInsectCreature(img *image.RGBA, rng *rand.Rand, ctype CreatureType, frame AnimFrame) {
	colors := selectInsectColors(ctype)
	legCount := selectLegCount(ctype)

	centerX := 32
	centerY := 32

	drawInsectLegsAnimated(img, centerX, centerY, legCount, frame, colors.body)
	drawInsectBodyByType(img, centerX, centerY, ctype, frame, colors.body, colors.shell)
	drawInsectEyes(img, centerX, centerY, ctype, colors.eye)
}

// insectColors holds the color scheme for an insect creature.
type insectColors struct {
	body  color.RGBA
	shell color.RGBA
	eye   color.RGBA
}

// selectInsectColors returns the color scheme for a specific insect type.
func selectInsectColors(ctype CreatureType) insectColors {
	switch ctype {
	case CreatureSpider:
		return insectColors{
			body:  color.RGBA{R: 40, G: 30, B: 35, A: 255},
			shell: color.RGBA{R: 60, G: 50, B: 55, A: 255},
			eye:   color.RGBA{R: 255, G: 0, B: 0, A: 255},
		}
	case CreatureBeetle:
		return insectColors{
			body:  color.RGBA{R: 20, G: 60, B: 40, A: 255},
			shell: color.RGBA{R: 30, G: 100, B: 60, A: 255},
			eye:   color.RGBA{R: 200, G: 200, B: 50, A: 255},
		}
	case CreatureMantis:
		return insectColors{
			body:  color.RGBA{R: 80, G: 140, B: 60, A: 255},
			shell: color.RGBA{R: 100, G: 160, B: 80, A: 255},
			eye:   color.RGBA{R: 255, G: 255, B: 100, A: 255},
		}
	case CreatureScorpion:
		return insectColors{
			body:  color.RGBA{R: 140, G: 100, B: 60, A: 255},
			shell: color.RGBA{R: 160, G: 120, B: 80, A: 255},
			eye:   color.RGBA{R: 100, G: 50, B: 0, A: 255},
		}
	default: // Ant
		return insectColors{
			body:  color.RGBA{R: 100, G: 40, B: 40, A: 255},
			shell: color.RGBA{R: 120, G: 60, B: 60, A: 255},
			eye:   color.RGBA{R: 50, G: 50, B: 50, A: 255},
		}
	}
}

// selectLegCount returns the number of legs for a specific insect type.
func selectLegCount(ctype CreatureType) int {
	if ctype == CreatureSpider || ctype == CreatureScorpion {
		return 8
	}
	return 6
}

// drawInsectLegsAnimated renders animated legs radiating from the center.
func drawInsectLegsAnimated(img *image.RGBA, centerX, centerY, legCount int, frame AnimFrame, bodyColor color.RGBA) {
	legAngleOffset := calculateLegAngleOffset(frame)

	for i := 0; i < legCount; i++ {
		angle := (float64(i) / float64(legCount)) * 2 * math.Pi
		side := calculateLegSide(i, legCount)
		angle += legAngleOffset * float64(side)

		drawInsectLeg(img, centerX, centerY, angle, bodyColor)
	}
}

// calculateLegAngleOffset returns the leg animation offset for the current frame.
func calculateLegAngleOffset(frame AnimFrame) float64 {
	if frame == AnimFrameWalk1 {
		return 0.2
	}
	if frame == AnimFrameWalk2 {
		return -0.2
	}
	return 0.0
}

// calculateLegSide determines which side of the body a leg is on.
func calculateLegSide(legIndex, legCount int) int {
	if legIndex >= legCount/2 {
		return -1
	}
	return 1
}

// drawInsectLeg renders a single two-segment insect leg.
func drawInsectLeg(img *image.RGBA, centerX, centerY int, angle float64, bodyColor color.RGBA) {
	segment1X := centerX + int(8*math.Cos(angle))
	segment1Y := centerY + int(8*math.Sin(angle))
	segment2X := centerX + int(16*math.Cos(angle))
	segment2Y := centerY + int(16*math.Sin(angle))

	common.DrawLine(img, centerX, centerY, segment1X, segment1Y, bodyColor, 2)
	common.DrawLine(img, segment1X, segment1Y, segment2X, segment2Y, bodyColor, 1)
}

// drawInsectBodyByType renders the body shape specific to each insect type.
func drawInsectBodyByType(img *image.RGBA, centerX, centerY int, ctype CreatureType, frame AnimFrame, bodyColor, shellColor color.RGBA) {
	switch ctype {
	case CreatureSpider:
		drawSpiderBody(img, centerX, centerY, bodyColor, shellColor)
	case CreatureMantis:
		drawMantisBody(img, centerX, centerY, frame, bodyColor, shellColor)
	default:
		drawSegmentedBody(img, centerX, centerY, ctype, bodyColor, shellColor)
	}
}

// drawSpiderBody renders a spider's round abdomen and cephalothorax.
func drawSpiderBody(img *image.RGBA, centerX, centerY int, bodyColor, shellColor color.RGBA) {
	common.FillCircle(img, centerX, centerY+4, 10, bodyColor)
	common.FillCircle(img, centerX, centerY+4, 8, shellColor)
	common.FillCircle(img, centerX, centerY-6, 6, bodyColor)
}

// drawMantisBody renders a mantis with elongated thorax and triangular head.
func drawMantisBody(img *image.RGBA, centerX, centerY int, frame AnimFrame, bodyColor, shellColor color.RGBA) {
	common.FillEllipse(img, centerX, centerY, 6, 12, bodyColor)
	common.FillEllipse(img, centerX, centerY, 5, 10, shellColor)
	common.FillTriangle(img, centerX, centerY-12, centerX-4, centerY-6, centerX+4, centerY-6, bodyColor)

	if frame == AnimFrameAttack {
		common.DrawLine(img, centerX-6, centerY-6, centerX-12, centerY-16, shellColor, 2)
		common.DrawLine(img, centerX+6, centerY-6, centerX+12, centerY-16, shellColor, 2)
	}
}

// drawSegmentedBody renders a segmented body for beetles, ants, and scorpions.
func drawSegmentedBody(img *image.RGBA, centerX, centerY int, ctype CreatureType, bodyColor, shellColor color.RGBA) {
	common.FillCircle(img, centerX, centerY-8, 5, bodyColor)
	common.FillEllipse(img, centerX, centerY, 7, 8, shellColor)
	common.FillEllipse(img, centerX, centerY+10, 8, 10, bodyColor)

	if ctype == CreatureScorpion {
		drawScorpionTail(img, centerX, centerY)
	}
}

// drawScorpionTail renders a scorpion's tail with stinger.
func drawScorpionTail(img *image.RGBA, centerX, centerY int) {
	tailSegments := 5
	bodyColor := color.RGBA{R: 140, G: 100, B: 60, A: 255}

	for i := 0; i < tailSegments; i++ {
		segY := centerY + 20 + i*3
		segSize := 4 - i/2
		common.FillCircle(img, centerX, segY, segSize, bodyColor)
	}

	stingerY := centerY + 20 + tailSegments*3
	common.FillTriangle(img, centerX, stingerY, centerX-3, stingerY+4, centerX+3, stingerY+4, color.RGBA{R: 200, G: 200, B: 50, A: 255})
}

// drawInsectEyes renders eyes appropriate for the insect type.
func drawInsectEyes(img *image.RGBA, centerX, centerY int, ctype CreatureType, eyeColor color.RGBA) {
	if ctype == CreatureSpider {
		common.FillCircle(img, centerX-3, centerY-8, 1, eyeColor)
		common.FillCircle(img, centerX-1, centerY-8, 1, eyeColor)
		common.FillCircle(img, centerX+1, centerY-8, 1, eyeColor)
		common.FillCircle(img, centerX+3, centerY-8, 1, eyeColor)
	} else {
		common.FillCircle(img, centerX-2, centerY-8, 2, eyeColor)
		common.FillCircle(img, centerX+2, centerY-8, 2, eyeColor)
	}
}

// generateSerpentCreature draws snake-like elongated creatures.
func generateSerpentCreature(img *image.RGBA, rng *rand.Rand, ctype CreatureType, frame AnimFrame) {
	var bodyColor, bellyColor, eyeColor color.RGBA
	var thickness int

	switch ctype {
	case CreatureSnake:
		bodyColor = color.RGBA{R: 60, G: 100, B: 40, A: 255}
		bellyColor = color.RGBA{R: 140, G: 160, B: 120, A: 255}
		eyeColor = color.RGBA{R: 255, G: 200, B: 0, A: 255}
		thickness = 4
	case CreatureWorm:
		bodyColor = color.RGBA{R: 140, G: 100, B: 80, A: 255}
		bellyColor = color.RGBA{R: 160, G: 120, B: 100, A: 255}
		eyeColor = color.RGBA{R: 80, G: 60, B: 40, A: 255}
		thickness = 6
	case CreatureLamia:
		bodyColor = color.RGBA{R: 100, G: 80, B: 140, A: 255}
		bellyColor = color.RGBA{R: 180, G: 160, B: 200, A: 255}
		eyeColor = color.RGBA{R: 150, G: 50, B: 200, A: 255}
		thickness = 5
	default: // Serpent
		bodyColor = color.RGBA{R: 80, G: 60, B: 100, A: 255}
		bellyColor = color.RGBA{R: 140, G: 120, B: 160, A: 255}
		eyeColor = color.RGBA{R: 200, G: 50, B: 50, A: 255}
		thickness = 5
	}

	// Serpentine motion path (S-curve)
	segments := 12
	waveOffset := 0.0
	if frame == AnimFrameWalk1 {
		waveOffset = 0.5
	} else if frame == AnimFrameWalk2 {
		waveOffset = -0.5
	}

	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments-1)
		y := 10 + int(t*44)
		x := 32 + int(math.Sin(t*math.Pi*2+waveOffset)*12)

		segThick := thickness
		if i == 0 {
			segThick = thickness + 2 // Thicker head
		} else if i == segments-1 {
			segThick = thickness - 2 // Thinner tail
		}

		common.FillCircle(img, x, y, segThick, bodyColor)

		// Belly stripe
		if i > 0 && i < segments-1 {
			common.FillCircle(img, x, y, segThick-1, bellyColor)
		}
	}

	// Head details
	headY := 10
	headX := 32 + int(math.Sin(waveOffset)*12)

	// Eyes
	common.FillCircle(img, headX-2, headY, 1, eyeColor)
	common.FillCircle(img, headX+2, headY, 1, eyeColor)

	// Tongue (attack frame)
	if frame == AnimFrameAttack {
		common.DrawLine(img, headX, headY-2, headX, headY-8, color.RGBA{R: 200, G: 50, B: 50, A: 255}, 1)
	}

	// Scale pattern
	for i := 1; i < segments-1; i++ {
		t := float64(i) / float64(segments-1)
		y := 10 + int(t*44)
		x := 32 + int(math.Sin(t*math.Pi*2+waveOffset)*12)

		if i%2 == 0 {
			darker := color.RGBA{
				R: bodyColor.R - 20,
				G: bodyColor.G - 20,
				B: bodyColor.B - 20,
				A: 255,
			}
			common.FillCircle(img, x-2, y, 1, darker)
			common.FillCircle(img, x+2, y, 1, darker)
		}
	}
}

// generateFlyingCreature draws winged aerial creatures.
func generateFlyingCreature(img *image.RGBA, rng *rand.Rand, ctype CreatureType, frame AnimFrame) {
	var bodyColor, wingColor, eyeColor color.RGBA
	var wingSpan int

	switch ctype {
	case CreatureBat:
		bodyColor = color.RGBA{R: 60, G: 40, B: 50, A: 255}
		wingColor = color.RGBA{R: 80, G: 60, B: 70, A: 255}
		eyeColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
		wingSpan = 20
	case CreatureDrake:
		bodyColor = color.RGBA{R: 100, G: 60, B: 60, A: 255}
		wingColor = color.RGBA{R: 140, G: 80, B: 80, A: 255}
		eyeColor = color.RGBA{R: 255, G: 200, B: 0, A: 255}
		wingSpan = 24
	case CreatureHarpy:
		bodyColor = color.RGBA{R: 180, G: 140, B: 120, A: 255}
		wingColor = color.RGBA{R: 200, G: 180, B: 160, A: 255}
		eyeColor = color.RGBA{R: 100, G: 150, B: 255, A: 255}
		wingSpan = 18
	default: // Wasp
		bodyColor = color.RGBA{R: 200, G: 180, B: 20, A: 255}
		wingColor = color.RGBA{R: 180, G: 220, B: 240, A: 150}
		eyeColor = color.RGBA{R: 50, G: 50, B: 50, A: 255}
		wingSpan = 16
	}

	centerX := 32
	centerY := 28

	// Wing flap animation
	wingAngle := 0.3
	if frame == AnimFrameWalk1 || frame == AnimFrameWalk2 {
		wingAngle = 0.6 // Wings up
	}

	// Wings (behind body)
	leftWingTipX := centerX - int(float64(wingSpan)*math.Cos(wingAngle))
	leftWingTipY := centerY - int(float64(wingSpan)*math.Sin(wingAngle))
	rightWingTipX := centerX + int(float64(wingSpan)*math.Cos(wingAngle))
	rightWingTipY := centerY - int(float64(wingSpan)*math.Sin(wingAngle))

	// Draw wing membranes
	common.FillTriangle(img, centerX, centerY, leftWingTipX, leftWingTipY, centerX-4, centerY+8, wingColor)
	common.FillTriangle(img, centerX, centerY, rightWingTipX, rightWingTipY, centerX+4, centerY+8, wingColor)

	// Body
	common.FillEllipse(img, centerX, centerY, 6, 10, bodyColor)

	// Head
	headY := centerY - 8
	common.FillCircle(img, centerX, headY, 5, bodyColor)

	// Eyes
	common.FillCircle(img, centerX-2, headY, 2, eyeColor)
	common.FillCircle(img, centerX+2, headY, 2, eyeColor)

	// Type-specific features
	if ctype == CreatureBat {
		// Large ears
		common.FillTriangle(img, centerX-4, headY-4, centerX-6, headY-10, centerX-2, headY-8, bodyColor)
		common.FillTriangle(img, centerX+4, headY-4, centerX+6, headY-10, centerX+2, headY-8, bodyColor)
	} else if ctype == CreatureDrake {
		// Horns
		common.DrawLine(img, centerX-4, headY-4, centerX-6, headY-10, bodyColor, 2)
		common.DrawLine(img, centerX+4, headY-4, centerX+6, headY-10, bodyColor, 2)
		// Tail
		common.FillRect(img, centerX-1, centerY+10, centerX+1, centerY+20, bodyColor)
	} else if ctype == CreatureWasp {
		// Striped abdomen
		for i := 0; i < 4; i++ {
			stripeY := centerY + 10 + i*3
			stripColor := bodyColor
			if i%2 == 0 {
				stripColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
			}
			common.FillCircle(img, centerX, stripeY, 4-i/2, stripColor)
		}
		// Stinger
		common.FillTriangle(img, centerX, centerY+22, centerX-2, centerY+26, centerX+2, centerY+26, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	}

	// Talons/feet (if not wasp)
	if ctype != CreatureWasp {
		common.FillCircle(img, centerX-3, centerY+12, 2, color.RGBA{R: 80, G: 80, B: 80, A: 255})
		common.FillCircle(img, centerX+3, centerY+12, 2, color.RGBA{R: 80, G: 80, B: 80, A: 255})
	}
}

// generateAmorphousCreature draws formless or semi-fluid creatures.
func generateAmorphousCreature(img *image.RGBA, rng *rand.Rand, ctype CreatureType, frame AnimFrame) {
	var coreColor, accentColor, glowColor color.RGBA
	var wobble float64

	switch ctype {
	case CreatureSlime:
		coreColor = color.RGBA{R: 80, G: 200, B: 100, A: 200}
		accentColor = color.RGBA{R: 120, G: 240, B: 140, A: 180}
		glowColor = color.RGBA{R: 160, G: 255, B: 180, A: 100}
		wobble = 1.5
	case CreatureOoze:
		coreColor = color.RGBA{R: 100, G: 80, B: 120, A: 220}
		accentColor = color.RGBA{R: 140, G: 120, B: 160, A: 200}
		glowColor = color.RGBA{R: 180, G: 160, B: 200, A: 120}
		wobble = 1.2
	case CreatureElemental:
		coreColor = color.RGBA{R: 200, G: 100, B: 60, A: 180}
		accentColor = color.RGBA{R: 255, G: 140, B: 80, A: 140}
		glowColor = color.RGBA{R: 255, G: 200, B: 100, A: 100}
		wobble = 2.0
	default: // Wraith
		coreColor = color.RGBA{R: 60, G: 60, B: 80, A: 150}
		accentColor = color.RGBA{R: 100, G: 100, B: 140, A: 120}
		glowColor = color.RGBA{R: 140, G: 140, B: 200, A: 80}
		wobble = 1.8
	}

	centerX := 32
	centerY := 32

	// Animation pulse
	pulsePhase := 0.0
	if frame == AnimFrameWalk1 {
		pulsePhase = 0.5
	} else if frame == AnimFrameWalk2 {
		pulsePhase = 1.0
	}

	// Generate blob shape using multiple overlapping circles
	numBlobs := 8
	for i := 0; i < numBlobs; i++ {
		angle := (float64(i) / float64(numBlobs)) * 2 * math.Pi
		radius := 12.0 + wobble*math.Sin(angle*3+pulsePhase*math.Pi)

		blobX := centerX + int(radius*0.6*math.Cos(angle))
		blobY := centerY + int(radius*0.6*math.Sin(angle))
		blobR := int(radius)

		common.FillCircle(img, blobX, blobY, blobR, coreColor)
	}

	// Accent layer (slightly smaller)
	for i := 0; i < numBlobs; i++ {
		angle := (float64(i) / float64(numBlobs)) * 2 * math.Pi
		radius := 10.0 + wobble*math.Sin(angle*3+pulsePhase*math.Pi)

		blobX := centerX + int(radius*0.5*math.Cos(angle))
		blobY := centerY + int(radius*0.5*math.Sin(angle))
		blobR := int(radius * 0.8)

		common.FillCircle(img, blobX, blobY, blobR, accentColor)
	}

	// Core glow
	common.FillCircle(img, centerX, centerY, 8, glowColor)

	// Type-specific features
	if ctype == CreatureElemental {
		// Flame-like tendrils
		for i := 0; i < 6; i++ {
			angle := (float64(i) / 6.0) * 2 * math.Pi
			length := 10 + int(wobble*4*math.Sin(pulsePhase*math.Pi))

			tendrilX := centerX + int(float64(length)*math.Cos(angle))
			tendrilY := centerY + int(float64(length)*math.Sin(angle))

			common.DrawLine(img, centerX, centerY, tendrilX, tendrilY, accentColor, 2)
		}
	} else if ctype == CreatureWraith {
		// Ghostly wispy trails
		for i := 0; i < 4; i++ {
			trailY := centerY + 12 + i*4
			trailWidth := 8 - i
			alpha := uint8(150 - i*30)

			trailColor := color.RGBA{R: coreColor.R, G: coreColor.G, B: coreColor.B, A: alpha}
			common.FillEllipse(img, centerX, trailY, trailWidth, 2, trailColor)
		}
	}

	// Eyes (glowing orbs)
	eyeY := centerY - 4
	eyeGlow := color.RGBA{R: 255, G: 255, B: 100, A: 200}
	if ctype == CreatureWraith {
		eyeGlow = color.RGBA{R: 100, G: 255, B: 255, A: 200}
	}

	common.FillCircle(img, centerX-4, eyeY, 3, eyeGlow)
	common.FillCircle(img, centerX+4, eyeY, 3, eyeGlow)
	common.FillCircle(img, centerX-4, eyeY, 2, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	common.FillCircle(img, centerX+4, eyeY, 2, color.RGBA{R: 50, G: 50, B: 50, A: 255})
}

// generateSimpleHumanoid is a fallback for unknown creature types.
func generateSimpleHumanoid(img *image.RGBA, rng *rand.Rand) {
	bodyColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	// Legs
	common.FillRect(img, 26, 35, 30, 50, bodyColor)
	common.FillRect(img, 34, 35, 38, 50, bodyColor)

	// Body
	common.FillRect(img, 24, 18, 40, 35, bodyColor)

	// Arms
	common.FillRect(img, 18, 22, 24, 32, bodyColor)
	common.FillRect(img, 40, 22, 46, 32, bodyColor)

	// Head
	common.FillCircle(img, 32, 12, 6, bodyColor)
}
