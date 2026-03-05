package playersprite

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/common"
)

// ClassTemplate defines visual characteristics for a player class.
type ClassTemplate struct {
	BaseColor       color.RGBA
	AccentColor     color.RGBA
	SkinTone        color.RGBA
	HairColor       color.RGBA
	BodyProportions BodyProps
}

// BodyProps defines humanoid body proportions (0.0-1.0 normalized).
type BodyProps struct {
	HeadSize   float64
	TorsoWidth float64
	ArmLength  float64
	LegLength  float64
}

// Generator creates player and NPC sprites with equipment rendering.
type Generator struct {
	genreID   string
	templates map[string]ClassTemplate
}

// NewGenerator creates a sprite generator for players and NPCs.
func NewGenerator(genreID string) *Generator {
	g := &Generator{
		genreID:   genreID,
		templates: make(map[string]ClassTemplate),
	}
	g.initTemplates()
	return g
}

// SetGenre updates the generator's genre configuration.
func (g *Generator) SetGenre(genreID string) {
	g.genreID = genreID
	g.initTemplates()
}

// initTemplates creates class-specific visual templates based on genre.
func (g *Generator) initTemplates() {
	switch g.genreID {
	case "fantasy":
		g.templates["warrior"] = ClassTemplate{
			BaseColor:   color.RGBA{120, 100, 80, 255},
			AccentColor: color.RGBA{180, 150, 100, 255},
			SkinTone:    color.RGBA{220, 180, 150, 255},
			HairColor:   color.RGBA{80, 60, 40, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.22,
				TorsoWidth: 0.35,
				ArmLength:  0.38,
				LegLength:  0.40,
			},
		}
		g.templates["mage"] = ClassTemplate{
			BaseColor:   color.RGBA{60, 40, 120, 255},
			AccentColor: color.RGBA{140, 100, 200, 255},
			SkinTone:    color.RGBA{230, 200, 170, 255},
			HairColor:   color.RGBA{200, 200, 220, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.24,
				TorsoWidth: 0.30,
				ArmLength:  0.35,
				LegLength:  0.40,
			},
		}
		g.templates["rogue"] = ClassTemplate{
			BaseColor:   color.RGBA{40, 40, 50, 255},
			AccentColor: color.RGBA{100, 80, 60, 255},
			SkinTone:    color.RGBA{200, 160, 130, 255},
			HairColor:   color.RGBA{50, 40, 35, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.20,
				TorsoWidth: 0.28,
				ArmLength:  0.40,
				LegLength:  0.42,
			},
		}
	case "scifi", "cyberpunk":
		g.templates["soldier"] = ClassTemplate{
			BaseColor:   color.RGBA{80, 90, 100, 255},
			AccentColor: color.RGBA{0, 180, 220, 255},
			SkinTone:    color.RGBA{190, 160, 140, 255},
			HairColor:   color.RGBA{100, 100, 110, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.20,
				TorsoWidth: 0.38,
				ArmLength:  0.36,
				LegLength:  0.40,
			},
		}
		g.templates["hacker"] = ClassTemplate{
			BaseColor:   color.RGBA{30, 30, 35, 255},
			AccentColor: color.RGBA{0, 255, 150, 255},
			SkinTone:    color.RGBA{210, 180, 160, 255},
			HairColor:   color.RGBA{150, 80, 200, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.22,
				TorsoWidth: 0.30,
				ArmLength:  0.38,
				LegLength:  0.40,
			},
		}
		g.templates["cyborg"] = ClassTemplate{
			BaseColor:   color.RGBA{100, 100, 110, 255},
			AccentColor: color.RGBA{255, 50, 50, 255},
			SkinTone:    color.RGBA{160, 160, 165, 255},
			HairColor:   color.RGBA{120, 120, 130, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.21,
				TorsoWidth: 0.36,
				ArmLength:  0.37,
				LegLength:  0.41,
			},
		}
	case "horror":
		g.templates["survivor"] = ClassTemplate{
			BaseColor:   color.RGBA{80, 70, 60, 255},
			AccentColor: color.RGBA{120, 80, 60, 255},
			SkinTone:    color.RGBA{200, 170, 150, 255},
			HairColor:   color.RGBA{90, 70, 60, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.22,
				TorsoWidth: 0.32,
				ArmLength:  0.38,
				LegLength:  0.40,
			},
		}
		g.templates["occultist"] = ClassTemplate{
			BaseColor:   color.RGBA{50, 30, 40, 255},
			AccentColor: color.RGBA{120, 40, 60, 255},
			SkinTone:    color.RGBA{180, 160, 150, 255},
			HairColor:   color.RGBA{40, 30, 35, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.23,
				TorsoWidth: 0.30,
				ArmLength:  0.36,
				LegLength:  0.39,
			},
		}
	default: // postapoc
		g.templates["scavenger"] = ClassTemplate{
			BaseColor:   color.RGBA{100, 90, 70, 255},
			AccentColor: color.RGBA{140, 100, 60, 255},
			SkinTone:    color.RGBA{190, 150, 120, 255},
			HairColor:   color.RGBA{80, 70, 60, 255},
			BodyProportions: BodyProps{
				HeadSize:   0.21,
				TorsoWidth: 0.34,
				ArmLength:  0.38,
				LegLength:  0.41,
			},
		}
	}
}

// Generate creates a player sprite with equipment and animation state.
func (g *Generator) Generate(class string, seed int64, animState AnimationState, frame int, weapon, armor string) *ebiten.Image {
	const size = 48
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(seed))

	// Get or default template
	template, ok := g.templates[class]
	if !ok {
		// Default to first available template
		for _, t := range g.templates {
			template = t
			break
		}
	}

	// Animation offsets
	legOffset := 0.0
	armOffset := 0.0
	bodyBob := 0.0
	weaponAngle := 0.0

	switch animState {
	case AnimWalk:
		t := float64(frame%8) / 8.0
		legOffset = math.Sin(t*math.Pi*2) * 2
		bodyBob = math.Abs(math.Sin(t*math.Pi*2)) * 1
	case AnimAttack:
		t := float64(frame) / 8.0
		weaponAngle = math.Sin(t*math.Pi) * 60
		armOffset = -math.Sin(t*math.Pi) * 4
	case AnimHurt:
		bodyBob = 2
		armOffset = 2
	case AnimDodge:
		t := float64(frame) / 8.0
		bodyBob = -math.Sin(t*math.Pi) * 3
	case AnimCast:
		armOffset = -3
		t := float64(frame) / 8.0
		glowPulse := math.Sin(t * math.Pi * 2)
		_ = glowPulse // Used for magic glow effect
	}

	centerX := size / 2
	baseY := int(float64(size)*0.7 - bodyBob)

	// Calculate body part positions based on proportions
	headSize := int(float64(size) * template.BodyProportions.HeadSize)
	torsoHeight := int(math.Round(float64(size) * 0.30))
	torsoWidth := int(float64(size) * template.BodyProportions.TorsoWidth)
	legLength := int(float64(size) * template.BodyProportions.LegLength)
	armLength := int(float64(size) * template.BodyProportions.ArmLength)

	// Draw character from bottom up (z-ordering)

	// Legs (back leg first)
	leftLegX := centerX - torsoWidth/3
	rightLegX := centerX + torsoWidth/3
	legY := baseY
	legEndY := baseY + legLength

	g.drawLeg(img, rightLegX, legY+int(legOffset), rightLegX, legEndY+int(legOffset), template.BaseColor, armor)
	g.drawLeg(img, leftLegX, legY-int(legOffset), leftLegX, legEndY-int(legOffset), template.BaseColor, armor)

	// Torso
	torsoY := baseY - torsoHeight
	g.drawTorso(img, centerX-torsoWidth/2, torsoY, torsoWidth, torsoHeight, template.BaseColor, template.AccentColor, armor)

	// Arms (back arm first)
	shoulderY := torsoY + 2
	rightArmX := centerX + torsoWidth/2 + 1
	leftArmX := centerX - torsoWidth/2 - 1

	g.drawArm(img, rightArmX, shoulderY+int(armOffset), armLength, weaponAngle, template.SkinTone, template.BaseColor, weapon, false, rng)
	g.drawArm(img, leftArmX, shoulderY-int(armOffset), armLength, -weaponAngle*0.3, template.SkinTone, template.BaseColor, "", true, rng)

	// Head
	headY := torsoY - headSize
	g.drawHead(img, centerX-headSize/2, headY, headSize, template.SkinTone, template.HairColor, class, rng)

	// Convert to Ebiten image
	ebitenImg := ebiten.NewImageFromImage(img)
	return ebitenImg
}

// drawLeg renders a character leg with optional armor plating.
func (g *Generator) drawLeg(img *image.RGBA, x1, y1, x2, y2 int, baseColor color.RGBA, armor string) {
	legWidth := 3

	// Thigh
	common.FillRect(img, x1-1, y1, x1+2, y1+legWidth*2, baseColor)

	// Knee joint (slight darkness)
	kneeY := (y1 + y2) / 2
	darker := darken(baseColor, 0.8)
	common.FillRect(img, x1-1, kneeY, x1+2, kneeY+1, darker)

	// Lower leg
	common.FillRect(img, x1-1, kneeY+1, x1+2, y2, baseColor)

	// Boot/foot
	bootColor := darken(baseColor, 0.6)
	common.FillRect(img, x1-2, y2-3, x1+3, y2, bootColor)

	// Armor plating highlight
	if armor == "heavy" || armor == "medium" {
		plateColor := lighten(baseColor, 1.2)
		common.FillRect(img, x1, y1+1, x1+1, y1+legWidth, plateColor)
		common.FillRect(img, x1, kneeY+2, x1+1, kneeY+legWidth, plateColor)
	}
}

// drawTorso renders the character torso with layered armor.
func (g *Generator) drawTorso(img *image.RGBA, x, y, width, height int, baseColor, accentColor color.RGBA, armor string) {
	// Base torso
	common.FillRect(img, x, y, x+width, y+height, baseColor)

	// Shading for depth
	shadowColor := darken(baseColor, 0.7)
	common.FillRect(img, x, y, x+2, y+height, shadowColor)
	common.FillRect(img, x+width-2, y, x+width, y+height, shadowColor)

	// Armor details
	switch armor {
	case "heavy":
		// Chest plate
		plateColor := lighten(baseColor, 1.3)
		common.FillRect(img, x+3, y+2, x+width-3, y+height-2, plateColor)
		// Rivets
		img.Set(x+4, y+4, accentColor)
		img.Set(x+width-5, y+4, accentColor)
	case "medium":
		// Leather straps
		strapColor := darken(accentColor, 0.8)
		common.FillRect(img, x+2, y+height/3, x+width-2, y+height/3+1, strapColor)
		common.FillRect(img, x+2, y+2*height/3, x+width-2, y+2*height/3+1, strapColor)
	case "light":
		// Cloth texture variation
		img.Set(x+width/2-1, y+3, accentColor)
		img.Set(x+width/2+1, y+3, accentColor)
	}

	// Belt
	beltColor := darken(baseColor, 0.5)
	common.FillRect(img, x, y+height-3, x+width, y+height-1, beltColor)
	// Belt buckle
	img.Set(x+width/2, y+height-2, lighten(accentColor, 1.4))
}

// drawArm renders an arm with optional weapon.
func (g *Generator) drawArm(img *image.RGBA, startX, startY, length int, angle float64, skinTone, sleeveColor color.RGBA, weapon string, isLeft bool, rng *rand.Rand) {
	// Calculate arm endpoint based on angle
	rad := angle * math.Pi / 180.0
	endX := startX + int(float64(length)*math.Cos(rad))
	endY := startY + int(float64(length)*math.Sin(rad))

	// Upper arm (with sleeve)
	midX := (startX + endX) / 2
	midY := (startY + endY) / 2

	common.DrawLine(img, startX, startY, midX, midY, sleeveColor, 3)

	// Forearm
	common.DrawLine(img, midX, midY, endX, endY, skinTone, 2)

	// Hand
	handColor := darken(skinTone, 0.9)
	common.FillRect(img, endX-1, endY-1, endX+2, endY+2, handColor)

	// Weapon rendering
	if weapon != "" && !isLeft {
		g.drawWeapon(img, endX, endY, weapon, angle, rng)
	}
}

// drawWeapon renders weapon visuals attached to the hand.
func (g *Generator) drawWeapon(img *image.RGBA, handX, handY int, weaponType string, angle float64, rng *rand.Rand) {
	rad := angle * math.Pi / 180.0

	switch weaponType {
	case "sword", "blade":
		length := 12
		endX := handX + int(float64(length)*math.Cos(rad))
		endY := handY + int(float64(length)*math.Sin(rad))

		bladeColor := color.RGBA{200, 200, 220, 255}
		common.DrawLine(img, handX, handY, endX, endY, bladeColor, 2)

		// Blade highlight
		highlightX := handX + int(float64(length/2)*math.Cos(rad))
		highlightY := handY + int(float64(length/2)*math.Sin(rad))
		img.Set(highlightX, highlightY, color.RGBA{240, 240, 255, 255})

	case "staff", "wand":
		length := 14
		endX := handX + int(float64(length)*math.Cos(rad))
		endY := handY + int(float64(length)*math.Sin(rad))

		staffColor := color.RGBA{100, 70, 40, 255}
		common.DrawLine(img, handX, handY, endX, endY, staffColor, 2)

		// Magic orb at tip
		orbColor := color.RGBA{150, 100, 255, 255}
		common.FillRect(img, endX-1, endY-1, endX+2, endY+2, orbColor)
		img.Set(endX, endY, color.RGBA{200, 150, 255, 255})

	case "gun", "pistol":
		length := 8
		endX := handX + int(float64(length)*math.Cos(rad))
		endY := handY + int(float64(length)*math.Sin(rad))

		gunColor := color.RGBA{60, 60, 70, 255}
		common.DrawLine(img, handX, handY, endX, endY, gunColor, 3)

		// Barrel highlight
		img.Set(endX, endY, color.RGBA{100, 100, 120, 255})

	case "dagger":
		length := 6
		endX := handX + int(float64(length)*math.Cos(rad))
		endY := handY + int(float64(length)*math.Sin(rad))

		daggerColor := color.RGBA{180, 180, 200, 255}
		common.DrawLine(img, handX, handY, endX, endY, daggerColor, 1)
	}
}

// drawHead renders the character head with hair and facial features.
func (g *Generator) drawHead(img *image.RGBA, x, y, size int, skinTone, hairColor color.RGBA, class string, rng *rand.Rand) {
	// Face
	common.FillRect(img, x, y, x+size, y+size, skinTone)

	// Shading for roundness
	shadowColor := darken(skinTone, 0.85)
	common.FillRect(img, x, y, x+1, y+size, shadowColor)
	common.FillRect(img, x+size-1, y, x+size, y+size, shadowColor)
	common.FillRect(img, x, y, x+size, y+1, shadowColor)

	// Hair
	hairStyle := 0
	if class == "warrior" || class == "soldier" {
		hairStyle = 0 // Short
	} else if class == "mage" || class == "hacker" {
		hairStyle = 1 // Medium
	} else {
		hairStyle = rng.Intn(2)
	}

	if hairStyle == 0 {
		// Short hair
		common.FillRect(img, x, y, x+size, y+size/3, hairColor)
	} else {
		// Longer hair
		common.FillRect(img, x, y, x+size, y+size/2, hairColor)
		common.FillRect(img, x, y+size/3, x+2, y+size, hairColor)
		common.FillRect(img, x+size-2, y+size/3, x+size, y+size, hairColor)
	}

	// Eyes
	eyeColor := color.RGBA{50, 50, 60, 255}
	eyeY := y + size/2
	img.Set(x+size/3, eyeY, eyeColor)
	img.Set(x+2*size/3, eyeY, eyeColor)

	// Eye highlights
	img.Set(x+size/3, eyeY-1, color.RGBA{200, 200, 220, 255})
	img.Set(x+2*size/3, eyeY-1, color.RGBA{200, 200, 220, 255})
}

// darken reduces color brightness.
func darken(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

// lighten increases color brightness.
func lighten(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(c.R)*factor)),
		G: uint8(math.Min(255, float64(c.G)*factor)),
		B: uint8(math.Min(255, float64(c.B)*factor)),
		A: c.A,
	}
}
