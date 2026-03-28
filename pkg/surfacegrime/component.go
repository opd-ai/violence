package surfacegrime

import (
	"image"
	"image/color"
)

// Component holds grime accumulation data for a room or screen region.
type Component struct {
	// RoomID identifies which room this grime belongs to.
	RoomID string
	// Overlay is the pre-rendered grime texture (RGBA with alpha).
	Overlay *image.RGBA
	// Width and Height of the overlay in pixels.
	Width, Height int
	// Seed used for deterministic grime generation.
	Seed int64
	// Intensity scales overall grime visibility [0.0-1.0].
	Intensity float64
	// Dirty marks the overlay for regeneration.
	Dirty bool
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "surfacegrime"
}

// GrimeType identifies the kind of deposit.
type GrimeType int

const (
	// GrimeDirt is general brown/tan dirt and soil.
	GrimeDirt GrimeType = iota
	// GrimeDust is fine grey particulate matter.
	GrimeDust
	// GrimeMoss is green organic growth on stone/wood.
	GrimeMoss
	// GrimeMold is dark organic growth in damp areas.
	GrimeMold
	// GrimeOil is dark slick residue on metal/concrete.
	GrimeOil
	// GrimeSoot is black carbon scoring from fire/exhaust.
	GrimeSoot
	// GrimeRust is orange-brown oxidation on metal.
	GrimeRust
	// GrimeAsh is grey volcanic/fire ash.
	GrimeAsh
	// GrimeOoze is organic slimy residue.
	GrimeOoze
)

// GrimeColors returns the primary colors for each grime type.
var GrimeColors = map[GrimeType]color.RGBA{
	GrimeDirt: {R: 101, G: 67, B: 33, A: 180},   // Brown earth
	GrimeDust: {R: 128, G: 128, B: 120, A: 140}, // Grey dust
	GrimeMoss: {R: 60, G: 90, B: 40, A: 160},    // Green moss
	GrimeMold: {R: 40, G: 50, B: 45, A: 170},    // Dark mold
	GrimeOil:  {R: 30, G: 25, B: 20, A: 150},    // Dark oil
	GrimeSoot: {R: 25, G: 25, B: 25, A: 160},    // Black soot
	GrimeRust: {R: 139, G: 69, B: 19, A: 155},   // Rust orange
	GrimeAsh:  {R: 100, G: 100, B: 100, A: 130}, // Grey ash
	GrimeOoze: {R: 50, G: 70, B: 40, A: 165},    // Organic green
}

// GrimeSecondaryColors provides variation within each grime type.
var GrimeSecondaryColors = map[GrimeType][]color.RGBA{
	GrimeDirt: {
		{R: 90, G: 60, B: 30, A: 170},
		{R: 110, G: 75, B: 40, A: 175},
		{R: 80, G: 55, B: 25, A: 165},
	},
	GrimeDust: {
		{R: 140, G: 135, B: 125, A: 130},
		{R: 120, G: 120, B: 115, A: 145},
		{R: 135, G: 130, B: 125, A: 135},
	},
	GrimeMoss: {
		{R: 50, G: 80, B: 35, A: 155},
		{R: 70, G: 100, B: 50, A: 165},
		{R: 55, G: 85, B: 45, A: 150},
	},
	GrimeMold: {
		{R: 35, G: 45, B: 40, A: 165},
		{R: 45, G: 55, B: 50, A: 175},
		{R: 30, G: 40, B: 35, A: 160},
	},
	GrimeOil: {
		{R: 25, G: 20, B: 15, A: 145},
		{R: 35, G: 30, B: 25, A: 155},
		{R: 20, G: 18, B: 15, A: 140},
	},
	GrimeSoot: {
		{R: 20, G: 20, B: 20, A: 155},
		{R: 30, G: 30, B: 28, A: 165},
		{R: 18, G: 18, B: 18, A: 150},
	},
	GrimeRust: {
		{R: 150, G: 75, B: 25, A: 150},
		{R: 130, G: 60, B: 15, A: 160},
		{R: 145, G: 80, B: 30, A: 145},
	},
	GrimeAsh: {
		{R: 90, G: 90, B: 90, A: 125},
		{R: 110, G: 108, B: 105, A: 135},
		{R: 95, G: 95, B: 92, A: 120},
	},
	GrimeOoze: {
		{R: 45, G: 65, B: 35, A: 160},
		{R: 55, G: 75, B: 45, A: 170},
		{R: 40, G: 60, B: 30, A: 155},
	},
}

// GenreGrime defines which grime types appear in each genre with weights.
type GenreGrime struct {
	// Types lists the grime types present in this genre.
	Types []GrimeType
	// Weights controls relative frequency of each type (same order as Types).
	Weights []float64
	// CornerIntensity scales grime in inside corners.
	CornerIntensity float64
	// EdgeIntensity scales grime along wall bases.
	EdgeIntensity float64
	// CeilingIntensity scales grime at ceiling joints.
	CeilingIntensity float64
	// SpreadDistance is how far grime extends from edges (in pixels at internal res).
	SpreadDistance int
	// NoiseScale controls the size of noise features.
	NoiseScale float64
}

// genreGrimePresets defines grime characteristics per genre.
var genreGrimePresets = map[string]GenreGrime{
	"fantasy": {
		Types:            []GrimeType{GrimeDirt, GrimeDust, GrimeMoss, GrimeMold},
		Weights:          []float64{0.35, 0.25, 0.25, 0.15},
		CornerIntensity:  1.0,
		EdgeIntensity:    0.8,
		CeilingIntensity: 0.3,
		SpreadDistance:   12,
		NoiseScale:       0.15,
	},
	"scifi": {
		Types:            []GrimeType{GrimeDust, GrimeOil, GrimeSoot},
		Weights:          []float64{0.4, 0.35, 0.25},
		CornerIntensity:  0.5,
		EdgeIntensity:    0.4,
		CeilingIntensity: 0.2,
		SpreadDistance:   8,
		NoiseScale:       0.1,
	},
	"horror": {
		Types:            []GrimeType{GrimeMold, GrimeOoze, GrimeDirt, GrimeDust},
		Weights:          []float64{0.35, 0.25, 0.25, 0.15},
		CornerIntensity:  1.2,
		EdgeIntensity:    1.0,
		CeilingIntensity: 0.6,
		SpreadDistance:   16,
		NoiseScale:       0.2,
	},
	"cyberpunk": {
		Types:            []GrimeType{GrimeOil, GrimeSoot, GrimeDust, GrimeRust},
		Weights:          []float64{0.3, 0.3, 0.25, 0.15},
		CornerIntensity:  0.9,
		EdgeIntensity:    0.7,
		CeilingIntensity: 0.4,
		SpreadDistance:   10,
		NoiseScale:       0.12,
	},
	"postapoc": {
		Types:            []GrimeType{GrimeAsh, GrimeDust, GrimeRust, GrimeDirt},
		Weights:          []float64{0.3, 0.3, 0.25, 0.15},
		CornerIntensity:  1.3,
		EdgeIntensity:    1.1,
		CeilingIntensity: 0.8,
		SpreadDistance:   18,
		NoiseScale:       0.18,
	},
}

// GetGenreGrime returns the grime preset for a genre.
func GetGenreGrime(genreID string) GenreGrime {
	if preset, ok := genreGrimePresets[genreID]; ok {
		return preset
	}
	return genreGrimePresets["fantasy"]
}
