// Package raycaster implements the core raycasting engine.
package raycaster

import "math"

// Trig lookup table constants
const (
	// tableSize = 3600 entries for 0.1° resolution (360° * 10)
	tableSize = 3600
	// angleScale converts radians to table index
	angleScale = tableSize / (2 * math.Pi)
)

// Pre-computed trig tables initialized on package load
var (
	sinTable [tableSize]float64
	cosTable [tableSize]float64
	tanTable [tableSize]float64
)

func init() {
	// Pre-compute sin, cos, tan for all 0.1° increments
	for i := 0; i < tableSize; i++ {
		angle := float64(i) * 2.0 * math.Pi / tableSize
		sinTable[i] = math.Sin(angle)
		cosTable[i] = math.Cos(angle)
		tanTable[i] = math.Tan(angle)
	}
}

// Sin returns sine using lookup table with linear interpolation.
func Sin(radians float64) float64 {
	// Normalize angle to [0, 2π)
	for radians < 0 {
		radians += 2 * math.Pi
	}
	for radians >= 2*math.Pi {
		radians -= 2 * math.Pi
	}

	// Convert to table index with fractional part
	index := radians * angleScale
	i0 := int(index) % tableSize
	i1 := (i0 + 1) % tableSize

	// Linear interpolation between table entries
	frac := index - float64(i0)
	return sinTable[i0]*(1-frac) + sinTable[i1]*frac
}

// Cos returns cosine using lookup table with linear interpolation.
func Cos(radians float64) float64 {
	// Normalize angle to [0, 2π)
	for radians < 0 {
		radians += 2 * math.Pi
	}
	for radians >= 2*math.Pi {
		radians -= 2 * math.Pi
	}

	// Convert to table index with fractional part
	index := radians * angleScale
	i0 := int(index) % tableSize
	i1 := (i0 + 1) % tableSize

	// Linear interpolation between table entries
	frac := index - float64(i0)
	return cosTable[i0]*(1-frac) + cosTable[i1]*frac
}

// Tan returns tangent using lookup table with linear interpolation.
func Tan(radians float64) float64 {
	// Normalize angle to [0, 2π)
	for radians < 0 {
		radians += 2 * math.Pi
	}
	for radians >= 2*math.Pi {
		radians -= 2 * math.Pi
	}

	// Convert to table index with fractional part
	index := radians * angleScale
	i0 := int(index) % tableSize
	i1 := (i0 + 1) % tableSize

	// Handle discontinuities near π/2 and 3π/2 by falling back to math.Tan
	// Tangent approaches infinity at these points
	angle := float64(i0) * 2.0 * math.Pi / tableSize
	if math.Abs(math.Cos(angle)) < 0.01 {
		return math.Tan(radians)
	}

	// Linear interpolation between table entries
	frac := index - float64(i0)
	return tanTable[i0]*(1-frac) + tanTable[i1]*frac
}
