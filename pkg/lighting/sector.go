package lighting

import (
	"math"
)

// SectorLightMap manages per-tile lighting for a level sector.
// It maintains an ambient light level and combines contributions from multiple point lights.
type SectorLightMap struct {
	Width      int         // Map width in tiles
	Height     int         // Map height in tiles
	Ambient    float64     // Base ambient light level [0.0-1.0]
	lights     []Light     // Active point light sources
	coneLights []ConeLight // Active cone light sources (flashlights)
	lightGrid  []float64   // Cached per-tile illumination [0.0-1.0]
	dirty      bool        // True when lights changed, requires recalculation
}

// NewSectorLightMap creates a lighting map for the given dimensions.
// ambient specifies the base illumination level (0.0 = pitch black, 1.0 = full bright).
func NewSectorLightMap(width, height int, ambient float64) *SectorLightMap {
	return &SectorLightMap{
		Width:      width,
		Height:     height,
		Ambient:    clamp(ambient, 0.0, 1.0),
		lights:     make([]Light, 0, 16),
		coneLights: make([]ConeLight, 0, 4),
		lightGrid:  make([]float64, width*height),
		dirty:      true,
	}
}

// AddLight registers a new point light source.
// Returns the index of the added light for later removal.
func (s *SectorLightMap) AddLight(light Light) int {
	s.lights = append(s.lights, light)
	s.dirty = true
	return len(s.lights) - 1
}

// AddFlashlight registers a cone light source (flashlight, torch, etc).
// x, y: position, dirX, dirY: direction vector, coneAngle: half-angle in radians,
// range: maximum reach, intensity: brightness [0.0-1.0]
// Returns the index of the added cone light for later removal.
func (s *SectorLightMap) AddFlashlight(x, y, dirX, dirY, coneAngle, radius, intensity float64) int {
	// Normalize direction
	length := math.Sqrt(dirX*dirX + dirY*dirY)
	if length > 0.001 {
		dirX /= length
		dirY /= length
	} else {
		dirX = 1.0
		dirY = 0.0
	}

	cone := ConeLight{
		X:         x,
		Y:         y,
		DirX:      dirX,
		DirY:      dirY,
		Radius:    radius,
		Angle:     coneAngle,
		Intensity: clamp(intensity, 0.0, 1.0),
		R:         1.0,
		G:         1.0,
		B:         1.0,
		IsActive:  true,
	}
	s.coneLights = append(s.coneLights, cone)
	s.dirty = true
	return len(s.coneLights) - 1
}

// RemoveLight removes a light source by index.
// Returns true if the light was found and removed.
func (s *SectorLightMap) RemoveLight(index int) bool {
	if index < 0 || index >= len(s.lights) {
		return false
	}
	s.lights = append(s.lights[:index], s.lights[index+1:]...)
	s.dirty = true
	return true
}

// UpdateLight modifies an existing light source.
// Returns true if the light was found and updated.
func (s *SectorLightMap) UpdateLight(index int, light Light) bool {
	if index < 0 || index >= len(s.lights) {
		return false
	}
	s.lights[index] = light
	s.dirty = true
	return true
}

// SetAmbient updates the base ambient light level.
func (s *SectorLightMap) SetAmbient(ambient float64) {
	s.Ambient = clamp(ambient, 0.0, 1.0)
	s.dirty = true
}

// Calculate computes combined illumination for all tiles.
// Each tile receives ambient light plus contributions from all point lights and cone lights.
// Light intensity falls off as 1 / (1 + distance²) for quadratic attenuation.
func (s *SectorLightMap) Calculate() {
	if !s.dirty {
		return
	}

	// Reset grid to ambient level
	for i := range s.lightGrid {
		s.lightGrid[i] = s.Ambient
	}

	// Add point light contributions
	for _, light := range s.lights {
		s.addLightContribution(light)
	}

	// Add cone light contributions
	for _, cone := range s.coneLights {
		s.addConeLightContribution(cone)
	}

	s.dirty = false
}

// GetLight returns the computed illumination value at the given tile.
// Returns 0.0 for out-of-bounds coordinates.
// Call Calculate() before querying to ensure values are up-to-date.
func (s *SectorLightMap) GetLight(x, y int) float64 {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return 0.0
	}
	return s.lightGrid[y*s.Width+x]
}

// LightCount returns the number of active light sources.
func (s *SectorLightMap) LightCount() int {
	return len(s.lights)
}

// ConeLightCount returns the number of active cone light sources.
func (s *SectorLightMap) ConeLightCount() int {
	return len(s.coneLights)
}

// Clear removes all light sources and resets to ambient.
func (s *SectorLightMap) Clear() {
	s.lights = s.lights[:0]
	s.coneLights = s.coneLights[:0]
	s.dirty = true
}

// addLightContribution adds a point light's contribution to the light grid.
// Uses quadratic attenuation: intensity = baseIntensity / (1 + distance²)
func (s *SectorLightMap) addLightContribution(light Light) {
	// Calculate bounding box to avoid processing entire grid
	radiusTiles := int(math.Ceil(light.Radius))
	minX := max(0, int(light.X)-radiusTiles)
	maxX := min(s.Width-1, int(light.X)+radiusTiles)
	minY := max(0, int(light.Y)-radiusTiles)
	maxY := min(s.Height-1, int(light.Y)+radiusTiles)

	radiusSq := light.Radius * light.Radius

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) + 0.5 - light.X
			dy := float64(y) + 0.5 - light.Y
			distSq := dx*dx + dy*dy

			// Skip tiles outside light radius
			if distSq > radiusSq {
				continue
			}

			// Quadratic attenuation: intensity / (1 + distance²)
			contribution := light.Intensity / (1.0 + distSq)
			idx := y*s.Width + x
			s.lightGrid[idx] = clamp(s.lightGrid[idx]+contribution, 0.0, 1.0)
		}
	}
}

// addConeLightContribution adds a cone light's contribution to the light grid.
// Uses dot product angle test and quadratic distance attenuation.
func (s *SectorLightMap) addConeLightContribution(cone ConeLight) {
	if !cone.IsActive {
		return
	}

	// Calculate bounding box
	radiusTiles := int(math.Ceil(cone.Radius))
	minX := max(0, int(cone.X)-radiusTiles)
	maxX := min(s.Width-1, int(cone.X)+radiusTiles)
	minY := max(0, int(cone.Y)-radiusTiles)
	maxY := min(s.Height-1, int(cone.Y)+radiusTiles)

	radiusSq := cone.Radius * cone.Radius
	cosHalfAngle := math.Cos(cone.Angle)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Vector from light to tile center
			tileX := float64(x) + 0.5
			tileY := float64(y) + 0.5
			dx := tileX - cone.X
			dy := tileY - cone.Y
			distSq := dx*dx + dy*dy

			// Skip tiles outside radius
			if distSq > radiusSq || distSq < 0.0001 {
				continue
			}

			// Normalize direction to tile
			dist := math.Sqrt(distSq)
			tileDirX := dx / dist
			tileDirY := dy / dist

			// Dot product angle test: dot(lightDir, tileDir) > cos(coneAngle)
			dotProduct := cone.DirX*tileDirX + cone.DirY*tileDirY
			if dotProduct < cosHalfAngle {
				continue // Outside cone
			}

			// Quadratic attenuation with angular falloff
			distAttenuation := cone.Intensity / (1.0 + distSq)
			// Angular attenuation: brighter in center
			angleAttenuation := (dotProduct - cosHalfAngle) / (1.0 - cosHalfAngle)
			contribution := distAttenuation * angleAttenuation

			idx := y*s.Width + x
			s.lightGrid[idx] = clamp(s.lightGrid[idx]+contribution, 0.0, 1.0)
		}
	}
}

// clamp restricts value to [min, max] range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetGenre configures ambient light level for a specific genre.
// Fantasy=0.3, Scifi=0.5, Horror=0.15, Cyberpunk=0.25, Postapoc=0.35
func (s *SectorLightMap) SetGenre(genreID string) {
	ambient := genreAmbientLevel(genreID)
	s.SetAmbient(ambient)
}

// genreAmbientLevel returns the base ambient light level for a genre.
func genreAmbientLevel(genreID string) float64 {
	switch genreID {
	case "fantasy":
		return 0.3
	case "scifi":
		return 0.5
	case "horror":
		return 0.15
	case "cyberpunk":
		return 0.25
	case "postapoc":
		return 0.35
	default:
		return 0.3 // Default to fantasy
	}
}
