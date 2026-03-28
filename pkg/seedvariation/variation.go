// Package seedvariation provides seed-driven structural variation for procedural entities.
// It ensures that different seeds produce visually distinct individuals by modulating
// body proportions, feature placement, accessory presence, and asymmetric traits.
package seedvariation

import (
	"image/color"
	"math/rand"
)

// Variation holds seed-derived parameters for an individual entity.
// These parameters create structural diversity beyond simple color swaps.
type Variation struct {
	// Body proportions (multipliers from 0.7 to 1.3)
	LimbLengthRatio float64 // Arm/leg length relative to base
	BodyWidthRatio  float64 // Torso width relative to base
	HeadSizeRatio   float64 // Head size relative to base
	NeckLengthRatio float64 // Neck extension

	// Asymmetric features
	LeftArmOffset  int // Pixels offset for left arm position
	RightArmOffset int // Pixels offset for right arm position
	ShoulderTilt   int // Shoulder asymmetry in pixels

	// Creature-specific segment variation
	SegmentCount      int     // Number of body segments (for serpents, insects)
	SegmentSizeGrowth float64 // How segment size changes along body

	// Feature presence (seed-determined accessories)
	HasScar      bool // Visible scar on face/body
	HasHorns     bool // Horns/spikes present
	HornCount    int  // Number of horns (1-4)
	HasExtraEyes bool // Additional eyes
	EyeCount     int  // Total eye count
	HasSpotting  bool // Color spots/markings
	HasStripes   bool // Stripe pattern
	MarkingCount int  // Number of spots/stripes

	// Surface variation
	WearLevel     float64 // Damage/weathering amount [0.0-1.0]
	MissingPart   bool    // Missing limb/feature
	MissingPartID int     // Which part is missing (0=none, 1=left arm, etc.)

	// Color variation (within genre palette bounds)
	ColorShiftR int8    // Red channel shift [-30, +30]
	ColorShiftG int8    // Green channel shift [-30, +30]
	ColorShiftB int8    // Blue channel shift [-30, +30]
	Saturation  float64 // Saturation multiplier [0.8-1.2]
	Brightness  float64 // Brightness multiplier [0.85-1.15]

	// Pattern seed (for consistent marking placement)
	PatternSeed int64
}

// Generator creates seed-derived variations for entities.
type Generator struct {
	genre string
}

// NewGenerator creates a variation generator with genre context.
func NewGenerator(genreID string) *Generator {
	return &Generator{
		genre: genreID,
	}
}

// SetGenre updates the genre for variation generation.
func (g *Generator) SetGenre(genreID string) {
	g.genre = genreID
}

// GenerateHumanoidVariation creates variation parameters for humanoid entities.
func (g *Generator) GenerateHumanoidVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Body proportions vary ±15%
		LimbLengthRatio: 0.85 + rng.Float64()*0.30,
		BodyWidthRatio:  0.85 + rng.Float64()*0.30,
		HeadSizeRatio:   0.90 + rng.Float64()*0.20,
		NeckLengthRatio: 0.90 + rng.Float64()*0.20,

		// Subtle asymmetry
		LeftArmOffset:  rng.Intn(3) - 1,
		RightArmOffset: rng.Intn(3) - 1,
		ShoulderTilt:   rng.Intn(3) - 1,

		// Color variation
		ColorShiftR: int8(rng.Intn(61) - 30),
		ColorShiftG: int8(rng.Intn(61) - 30),
		ColorShiftB: int8(rng.Intn(61) - 30),
		Saturation:  0.8 + rng.Float64()*0.4,
		Brightness:  0.85 + rng.Float64()*0.30,

		// Surface wear
		WearLevel: rng.Float64() * 0.5,

		// Pattern seed for consistent markings
		PatternSeed: rng.Int63(),
	}

	// Scar chance: 20%
	v.HasScar = rng.Float64() < 0.20

	// Missing part chance: 5% (post-apocalyptic increases this)
	missingChance := 0.05
	if g.genre == "postapoc" {
		missingChance = 0.15
	} else if g.genre == "horror" {
		missingChance = 0.10
	}
	v.MissingPart = rng.Float64() < missingChance
	if v.MissingPart {
		v.MissingPartID = 1 + rng.Intn(4) // 1=left arm, 2=right arm, 3=ear, 4=eye
	}

	return v
}

// GenerateQuadrupedVariation creates variation parameters for quadruped entities.
func (g *Generator) GenerateQuadrupedVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Body proportions - quadrupeds have more variation
		LimbLengthRatio: 0.80 + rng.Float64()*0.40, // ±20%
		BodyWidthRatio:  0.80 + rng.Float64()*0.40,
		HeadSizeRatio:   0.85 + rng.Float64()*0.30,
		NeckLengthRatio: 0.70 + rng.Float64()*0.60, // Neck length varies a lot

		// Color variation
		ColorShiftR: int8(rng.Intn(61) - 30),
		ColorShiftG: int8(rng.Intn(61) - 30),
		ColorShiftB: int8(rng.Intn(61) - 30),
		Saturation:  0.75 + rng.Float64()*0.50,
		Brightness:  0.80 + rng.Float64()*0.40,

		// Surface pattern
		WearLevel:   rng.Float64() * 0.3,
		PatternSeed: rng.Int63(),
	}

	// Horns: 30% for fantasy/postapoc
	hornChance := 0.10
	if g.genre == "fantasy" || g.genre == "postapoc" {
		hornChance = 0.30
	}
	v.HasHorns = rng.Float64() < hornChance
	if v.HasHorns {
		v.HornCount = 1 + rng.Intn(3) // 1-3 horns
	}

	// Spotting: 25%
	v.HasSpotting = rng.Float64() < 0.25
	if v.HasSpotting {
		v.MarkingCount = 3 + rng.Intn(6)
	}

	// Stripes: 20% (mutually exclusive with spots)
	if !v.HasSpotting {
		v.HasStripes = rng.Float64() < 0.20
		if v.HasStripes {
			v.MarkingCount = 4 + rng.Intn(4)
		}
	}

	return v
}

// GenerateInsectVariation creates variation parameters for insect entities.
func (g *Generator) GenerateInsectVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Segment variation is key for insects
		SegmentCount:      3 + rng.Intn(3), // 3-5 segments
		SegmentSizeGrowth: 0.8 + rng.Float64()*0.4,

		// Body proportions
		LimbLengthRatio: 0.85 + rng.Float64()*0.30,
		BodyWidthRatio:  0.80 + rng.Float64()*0.40,
		HeadSizeRatio:   0.80 + rng.Float64()*0.40,

		// Color variation - insects can be quite varied
		ColorShiftR: int8(rng.Intn(61) - 30),
		ColorShiftG: int8(rng.Intn(61) - 30),
		ColorShiftB: int8(rng.Intn(61) - 30),
		Saturation:  0.70 + rng.Float64()*0.60,
		Brightness:  0.80 + rng.Float64()*0.40,

		PatternSeed: rng.Int63(),
	}

	// Extra eyes: 15%
	v.HasExtraEyes = rng.Float64() < 0.15
	v.EyeCount = 2
	if v.HasExtraEyes {
		v.EyeCount = 4 + rng.Intn(3) // 4-6 eyes
	}

	// Spotting on carapace: 20%
	v.HasSpotting = rng.Float64() < 0.20
	if v.HasSpotting {
		v.MarkingCount = 2 + rng.Intn(4)
	}

	return v
}

// GenerateSerpentVariation creates variation parameters for serpent entities.
func (g *Generator) GenerateSerpentVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Serpents vary primarily in segment count and body thickness
		SegmentCount:      6 + rng.Intn(5),           // 6-10 segments
		SegmentSizeGrowth: 0.85 + rng.Float64()*0.30, // How much segments shrink

		BodyWidthRatio: 0.80 + rng.Float64()*0.40, // Body thickness
		HeadSizeRatio:  0.85 + rng.Float64()*0.30,

		// Color variation
		ColorShiftR: int8(rng.Intn(61) - 30),
		ColorShiftG: int8(rng.Intn(61) - 30),
		ColorShiftB: int8(rng.Intn(61) - 30),
		Saturation:  0.80 + rng.Float64()*0.40,
		Brightness:  0.85 + rng.Float64()*0.30,

		PatternSeed: rng.Int63(),
	}

	// Horns/ridges: 25%
	v.HasHorns = rng.Float64() < 0.25
	if v.HasHorns {
		v.HornCount = 1 + rng.Intn(2) // Small horns/ridges
	}

	// Stripes/bands: 40% (common for serpents)
	v.HasStripes = rng.Float64() < 0.40
	if v.HasStripes {
		v.MarkingCount = 3 + rng.Intn(5)
	}

	// Spotting: 25% (mutually exclusive)
	if !v.HasStripes {
		v.HasSpotting = rng.Float64() < 0.35
		if v.HasSpotting {
			v.MarkingCount = 5 + rng.Intn(8)
		}
	}

	return v
}

// GenerateFlyingVariation creates variation parameters for flying entities.
func (g *Generator) GenerateFlyingVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Wings and body shape
		LimbLengthRatio: 0.90 + rng.Float64()*0.40, // Wing span varies a lot
		BodyWidthRatio:  0.75 + rng.Float64()*0.50,
		HeadSizeRatio:   0.85 + rng.Float64()*0.30,

		// Color variation
		ColorShiftR: int8(rng.Intn(61) - 30),
		ColorShiftG: int8(rng.Intn(61) - 30),
		ColorShiftB: int8(rng.Intn(61) - 30),
		Saturation:  0.80 + rng.Float64()*0.40,
		Brightness:  0.85 + rng.Float64()*0.30,

		PatternSeed: rng.Int63(),
	}

	// Extra eyes (bat-like): 10%
	v.HasExtraEyes = rng.Float64() < 0.10
	v.EyeCount = 2
	if v.HasExtraEyes {
		v.EyeCount = 4
	}

	// Wing patterns: 30%
	v.HasSpotting = rng.Float64() < 0.30
	if v.HasSpotting {
		v.MarkingCount = 2 + rng.Intn(4)
	}

	return v
}

// GenerateAmorphousVariation creates variation parameters for amorphous entities.
func (g *Generator) GenerateAmorphousVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Amorphous have significant size/shape variation
		BodyWidthRatio: 0.70 + rng.Float64()*0.60, // Very variable
		HeadSizeRatio:  0.60 + rng.Float64()*0.80, // Some have larger "core"

		// Color variation - amorphous can be quite colorful
		ColorShiftR: int8(rng.Intn(61) - 30),
		ColorShiftG: int8(rng.Intn(61) - 30),
		ColorShiftB: int8(rng.Intn(61) - 30),
		Saturation:  0.60 + rng.Float64()*0.80,
		Brightness:  0.70 + rng.Float64()*0.60,

		PatternSeed: rng.Int63(),
	}

	// Multiple eyes: 30%
	v.HasExtraEyes = rng.Float64() < 0.30
	v.EyeCount = 1 + rng.Intn(3) // 1-3 eyes normally
	if v.HasExtraEyes {
		v.EyeCount = 4 + rng.Intn(5) // 4-8 eyes
	}

	// Internal bubbles/spots: 50%
	v.HasSpotting = rng.Float64() < 0.50
	if v.HasSpotting {
		v.MarkingCount = 3 + rng.Intn(6)
	}

	return v
}

// GeneratePropVariation creates variation for props (wear, damage, details).
func (g *Generator) GeneratePropVariation(seed int64) Variation {
	rng := rand.New(rand.NewSource(seed))

	v := Variation{
		// Size variation for props
		BodyWidthRatio: 0.90 + rng.Float64()*0.20,

		// Wear and damage
		WearLevel: rng.Float64() * 0.6,

		// Color variation (weathering)
		ColorShiftR: int8(rng.Intn(41) - 20), // Less extreme for props
		ColorShiftG: int8(rng.Intn(41) - 20),
		ColorShiftB: int8(rng.Intn(41) - 20),
		Saturation:  0.80 + rng.Float64()*0.30, // Weather fades color
		Brightness:  0.85 + rng.Float64()*0.20,

		PatternSeed: rng.Int63(),
	}

	// Post-apocalyptic props have more wear
	if g.genre == "postapoc" {
		v.WearLevel = 0.3 + rng.Float64()*0.5
		v.Saturation *= 0.9
	} else if g.genre == "horror" {
		v.WearLevel = 0.2 + rng.Float64()*0.4
		v.Brightness *= 0.9
	}

	// Attached details (moss, labels, dents): 40%
	v.HasSpotting = rng.Float64() < 0.40
	if v.HasSpotting {
		v.MarkingCount = 1 + rng.Intn(4)
	}

	return v
}

// ApplyColorVariation modifies a base color using the variation parameters.
func (v *Variation) ApplyColorVariation(base color.RGBA) color.RGBA {
	// Apply shift
	r := int(base.R) + int(v.ColorShiftR)
	g := int(base.G) + int(v.ColorShiftG)
	b := int(base.B) + int(v.ColorShiftB)

	// Apply brightness
	r = int(float64(r) * v.Brightness)
	g = int(float64(g) * v.Brightness)
	b = int(float64(b) * v.Brightness)

	// Clamp values
	if r < 0 {
		r = 0
	} else if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	} else if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	} else if b > 255 {
		b = 255
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: base.A}
}

// ScaleBodyPart returns a dimension scaled by the appropriate ratio.
func (v *Variation) ScaleBodyPart(baseSize int, partType string) int {
	var ratio float64
	switch partType {
	case "limb", "leg", "arm":
		ratio = v.LimbLengthRatio
	case "body", "torso":
		ratio = v.BodyWidthRatio
	case "head":
		ratio = v.HeadSizeRatio
	case "neck":
		ratio = v.NeckLengthRatio
	default:
		ratio = 1.0
	}
	return int(float64(baseSize) * ratio)
}

// GetSegmentRadius returns the radius for a body segment at given index.
// Used for serpents and insects with variable segment sizing.
func (v *Variation) GetSegmentRadius(baseRadius, segmentIndex int) int {
	// Each segment shrinks based on SegmentSizeGrowth
	shrinkFactor := 1.0
	for i := 0; i < segmentIndex; i++ {
		shrinkFactor *= v.SegmentSizeGrowth
	}
	radius := int(float64(baseRadius) * shrinkFactor)
	if radius < 2 {
		radius = 2
	}
	return radius
}
