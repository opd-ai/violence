package lighting

// AOComponent represents ambient occlusion data for an entity.
// Stores occlusion factor for each cardinal direction and corner.
type AOComponent struct {
	// Directional occlusion [0.0 = fully occluded, 1.0 = no occlusion]
	North     float64
	South     float64
	East      float64
	West      float64
	NorthEast float64
	NorthWest float64
	SouthEast float64
	SouthWest float64

	// Overall occlusion factor (average of all directions)
	Overall float64

	// Radius for occlusion sampling (in world units)
	SampleRadius float64

	// Whether this entity casts occlusion on others
	CastsOcclusion bool

	// Genre-specific occlusion intensity multiplier
	IntensityMultiplier float64

	// Cache validation
	needsUpdate bool
}

// Type returns the component type identifier.
func (a *AOComponent) Type() string {
	return "AmbientOcclusion"
}

// NewAOComponent creates an ambient occlusion component with defaults.
func NewAOComponent(sampleRadius float64) *AOComponent {
	return &AOComponent{
		North:               1.0,
		South:               1.0,
		East:                1.0,
		West:                1.0,
		NorthEast:           1.0,
		NorthWest:           1.0,
		SouthEast:           1.0,
		SouthWest:           1.0,
		Overall:             1.0,
		SampleRadius:        sampleRadius,
		CastsOcclusion:      true,
		IntensityMultiplier: 1.0,
		needsUpdate:         true,
	}
}

// Invalidate marks the occlusion data as needing recalculation.
func (a *AOComponent) Invalidate() {
	a.needsUpdate = true
}

// IsValid returns whether the cached occlusion data is current.
func (a *AOComponent) IsValid() bool {
	return !a.needsUpdate
}

// GetOcclusionAt returns occlusion factor for a specific direction (radians).
// Interpolates between stored directional values.
func (a *AOComponent) GetOcclusionAt(angle float64) float64 {
	// Normalize angle to [0, 2π)
	for angle < 0 {
		angle += 6.283185307179586 // 2π
	}
	for angle >= 6.283185307179586 {
		angle -= 6.283185307179586
	}

	// Map angle to octant and interpolate
	octant := int(angle / 0.7853981633974483) // π/4
	t := (angle - float64(octant)*0.7853981633974483) / 0.7853981633974483

	var v1, v2 float64
	switch octant {
	case 0: // E to NE
		v1, v2 = a.East, a.NorthEast
	case 1: // NE to N
		v1, v2 = a.NorthEast, a.North
	case 2: // N to NW
		v1, v2 = a.North, a.NorthWest
	case 3: // NW to W
		v1, v2 = a.NorthWest, a.West
	case 4: // W to SW
		v1, v2 = a.West, a.SouthWest
	case 5: // SW to S
		v1, v2 = a.SouthWest, a.South
	case 6: // S to SE
		v1, v2 = a.South, a.SouthEast
	case 7: // SE to E
		v1, v2 = a.SouthEast, a.East
	default:
		return a.Overall
	}

	return v1*(1.0-t) + v2*t
}
