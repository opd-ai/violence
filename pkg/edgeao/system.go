package edgeao

import (
	"math"
	"math/rand"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines edge AO intensity and falloff for a genre.
type GenrePreset struct {
	// BaseIntensity is the maximum AO darkness at contact points [0.0-1.0].
	BaseIntensity float64
	// FalloffDistance is how far (in tiles) AO fades from edges.
	FalloffDistance float64
	// CornerMultiplier scales AO intensity at inside corners.
	CornerMultiplier float64
	// NarrowMultiplier scales AO for narrow passages.
	NarrowMultiplier float64
	// NoiseAmount adds subtle variation to break up uniform shadows.
	NoiseAmount float64
	// UseSoftFalloff enables smooth quadratic falloff (vs linear).
	UseSoftFalloff bool
}

// System computes and stores edge ambient occlusion for level geometry.
type System struct {
	genreID string
	preset  GenrePreset
	aoMap   [][]float64 // Per-tile AO values [y][x]
	edgeMap [][]EdgeType
	width   int
	height  int
	logger  *logrus.Entry
	rng     *rand.Rand
}

// NewSystem creates an edge AO system with genre-appropriate settings.
func NewSystem(genreID string, seed int64) *System {
	s := &System{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "edgeao",
		}),
		rng: rand.New(rand.NewSource(seed)),
	}
	s.applyGenrePreset(genreID)
	return s
}

// applyGenrePreset sets AO parameters based on genre.
func (s *System) applyGenrePreset(genreID string) {
	switch genreID {
	case "fantasy":
		// Strong AO in dungeon corners - emphasizes cramped stone corridors
		s.preset = GenrePreset{
			BaseIntensity:    0.35,
			FalloffDistance:  0.8,
			CornerMultiplier: 1.4,
			NarrowMultiplier: 1.2,
			NoiseAmount:      0.05,
			UseSoftFalloff:   true,
		}
	case "scifi":
		// Subtle, clean edge definition on metal surfaces
		s.preset = GenrePreset{
			BaseIntensity:    0.20,
			FalloffDistance:  0.5,
			CornerMultiplier: 1.1,
			NarrowMultiplier: 1.0,
			NoiseAmount:      0.02,
			UseSoftFalloff:   true,
		}
	case "horror":
		// Deep shadows in corners and tight spaces
		s.preset = GenrePreset{
			BaseIntensity:    0.45,
			FalloffDistance:  1.0,
			CornerMultiplier: 1.6,
			NarrowMultiplier: 1.5,
			NoiseAmount:      0.08,
			UseSoftFalloff:   true,
		}
	case "cyberpunk":
		// Moderate AO with grungy variation
		s.preset = GenrePreset{
			BaseIntensity:    0.30,
			FalloffDistance:  0.6,
			CornerMultiplier: 1.3,
			NarrowMultiplier: 1.1,
			NoiseAmount:      0.06,
			UseSoftFalloff:   true,
		}
	case "postapoc":
		// Dusty accumulation at edges - debris and dirt collect at walls
		s.preset = GenrePreset{
			BaseIntensity:    0.38,
			FalloffDistance:  0.9,
			CornerMultiplier: 1.5,
			NarrowMultiplier: 1.3,
			NoiseAmount:      0.10,
			UseSoftFalloff:   false, // Linear falloff for dusty look
		}
	default:
		s.preset = GenrePreset{
			BaseIntensity:    0.25,
			FalloffDistance:  0.6,
			CornerMultiplier: 1.2,
			NarrowMultiplier: 1.1,
			NoiseAmount:      0.04,
			UseSoftFalloff:   true,
		}
	}
}

// SetGenre updates the system for a new genre.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenrePreset(genreID)
	// Rebuild AO map if level exists
	if s.width > 0 && s.height > 0 {
		s.rebuildAOValues()
	}
}

// BuildAOMap computes edge AO for an entire level tile map.
// tiles[y][x] where non-zero values are walls.
func (s *System) BuildAOMap(tiles [][]int) {
	if len(tiles) == 0 || len(tiles[0]) == 0 {
		return
	}

	s.height = len(tiles)
	s.width = len(tiles[0])

	// Initialize maps
	s.aoMap = make([][]float64, s.height)
	s.edgeMap = make([][]EdgeType, s.height)
	for y := 0; y < s.height; y++ {
		s.aoMap[y] = make([]float64, s.width)
		s.edgeMap[y] = make([]EdgeType, s.width)
	}

	// First pass: classify edges
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			if s.isWall(tiles, x, y) {
				continue // AO only applies to floor tiles
			}
			s.edgeMap[y][x] = s.classifyEdge(tiles, x, y)
		}
	}

	// Second pass: compute AO values with distance falloff
	s.rebuildAOValues()

	s.logger.WithFields(logrus.Fields{
		"width":  s.width,
		"height": s.height,
		"genre":  s.genreID,
	}).Debug("Edge AO map built")
}

// rebuildAOValues recalculates AO from edge classification.
func (s *System) rebuildAOValues() {
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			s.aoMap[y][x] = s.computeAOValue(x, y)
		}
	}
}

// classifyEdge determines the type of geometric edge at a floor tile.
func (s *System) classifyEdge(tiles [][]int, x, y int) EdgeType {
	// Count adjacent walls (4-connected)
	wallN := s.isWall(tiles, x, y-1)
	wallS := s.isWall(tiles, x, y+1)
	wallE := s.isWall(tiles, x+1, y)
	wallW := s.isWall(tiles, x-1, y)

	adjacentWalls := 0
	if wallN {
		adjacentWalls++
	}
	if wallS {
		adjacentWalls++
	}
	if wallE {
		adjacentWalls++
	}
	if wallW {
		adjacentWalls++
	}

	// Count diagonal walls (for corner detection)
	wallNE := s.isWall(tiles, x+1, y-1)
	wallNW := s.isWall(tiles, x-1, y-1)
	wallSE := s.isWall(tiles, x+1, y+1)
	wallSW := s.isWall(tiles, x-1, y+1)

	// Alcove: 3 adjacent walls (recessed area)
	if adjacentWalls >= 3 {
		return EdgeAlcove
	}

	// Inside corner: 2 adjacent walls forming an L
	if adjacentWalls == 2 {
		if (wallN && wallE) || (wallN && wallW) || (wallS && wallE) || (wallS && wallW) {
			return EdgeInsideCorner
		}
	}

	// Narrow passage: walls on opposite sides
	if (wallN && wallS) || (wallE && wallW) {
		return EdgeNarrowPassage
	}

	// Outside corner: single adjacent wall with diagonal gap
	if adjacentWalls == 1 {
		if wallN && !wallNE && !wallNW {
			return EdgeOutsideCorner
		}
		if wallS && !wallSE && !wallSW {
			return EdgeOutsideCorner
		}
		if wallE && !wallNE && !wallSE {
			return EdgeOutsideCorner
		}
		if wallW && !wallNW && !wallSW {
			return EdgeOutsideCorner
		}
	}

	// Simple wall junction: any adjacent wall
	if adjacentWalls >= 1 {
		return EdgeWallJunction
	}

	return EdgeNone
}

// computeAOValue calculates the AO darkness for a tile based on edge type and distance.
func (s *System) computeAOValue(x, y int) float64 {
	edgeType := s.edgeMap[y][x]
	if edgeType == EdgeNone {
		// Still apply distance-based AO from nearby edges
		return s.computeDistanceAO(x, y)
	}

	// Base intensity from edge type
	baseAO := s.preset.BaseIntensity
	switch edgeType {
	case EdgeInsideCorner:
		baseAO *= s.preset.CornerMultiplier
	case EdgeAlcove:
		baseAO *= s.preset.CornerMultiplier * 1.1
	case EdgeNarrowPassage:
		baseAO *= s.preset.NarrowMultiplier
	case EdgeOutsideCorner:
		baseAO *= 0.7 // Lighter on outside corners
	}

	// Add noise for variation
	if s.preset.NoiseAmount > 0 {
		noise := (s.rng.Float64() - 0.5) * 2 * s.preset.NoiseAmount
		baseAO += noise
	}

	// Clamp to valid range
	if baseAO < 0 {
		baseAO = 0
	}
	if baseAO > 1 {
		baseAO = 1
	}

	return baseAO
}

// computeDistanceAO calculates AO for tiles not directly adjacent to walls.
func (s *System) computeDistanceAO(x, y int) float64 {
	// Find minimum distance to any edge
	minDist := s.preset.FalloffDistance + 1.0
	maxEdgeAO := 0.0

	searchRadius := int(math.Ceil(s.preset.FalloffDistance)) + 1
	for dy := -searchRadius; dy <= searchRadius; dy++ {
		for dx := -searchRadius; dx <= searchRadius; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= s.width || ny < 0 || ny >= s.height {
				continue
			}
			if s.edgeMap[ny][nx] == EdgeNone {
				continue
			}

			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist < minDist {
				minDist = dist
				// Get base AO at the edge tile
				edgeAO := s.computeEdgeBaseAO(s.edgeMap[ny][nx])
				if edgeAO > maxEdgeAO {
					maxEdgeAO = edgeAO
				}
			}
		}
	}

	if minDist > s.preset.FalloffDistance {
		return 0
	}

	// Apply falloff
	t := minDist / s.preset.FalloffDistance
	var falloff float64
	if s.preset.UseSoftFalloff {
		// Quadratic falloff for smooth shadows
		falloff = 1 - t*t
	} else {
		// Linear falloff for dusty/grungy look
		falloff = 1 - t
	}

	return maxEdgeAO * falloff
}

// computeEdgeBaseAO gets base AO for an edge type without position-specific modifiers.
func (s *System) computeEdgeBaseAO(edgeType EdgeType) float64 {
	baseAO := s.preset.BaseIntensity
	switch edgeType {
	case EdgeInsideCorner:
		baseAO *= s.preset.CornerMultiplier
	case EdgeAlcove:
		baseAO *= s.preset.CornerMultiplier * 1.1
	case EdgeNarrowPassage:
		baseAO *= s.preset.NarrowMultiplier
	case EdgeOutsideCorner:
		baseAO *= 0.7
	}
	if baseAO > 1 {
		baseAO = 1
	}
	return baseAO
}

// isWall checks if a tile is a wall (solid, non-floor).
func (s *System) isWall(tiles [][]int, x, y int) bool {
	if x < 0 || x >= len(tiles[0]) || y < 0 || y >= len(tiles) {
		return true // Treat out-of-bounds as wall
	}
	tileID := tiles[y][x]
	// Wall tiles are typically 1, 10-19 (wall variants), etc.
	// Floor tiles are 0, 2, 20-29
	return tileID != 0 && tileID != 2 && (tileID < 20 || tileID > 29)
}

// GetAO returns the AO factor for a world position.
// worldX, worldY are in tile coordinates (can be fractional).
// Returns 0.0 (no occlusion) to 1.0 (maximum occlusion).
func (s *System) GetAO(worldX, worldY float64) float64 {
	if s.aoMap == nil {
		return 0
	}

	// Get integer tile coordinates
	tileX := int(math.Floor(worldX))
	tileY := int(math.Floor(worldY))

	if tileX < 0 || tileX >= s.width || tileY < 0 || tileY >= s.height {
		return 0
	}

	// For sub-tile precision, interpolate between adjacent tiles
	fracX := worldX - float64(tileX)
	fracY := worldY - float64(tileY)

	// Bilinear interpolation for smooth AO gradients
	ao00 := s.getAOSafe(tileX, tileY)
	ao10 := s.getAOSafe(tileX+1, tileY)
	ao01 := s.getAOSafe(tileX, tileY+1)
	ao11 := s.getAOSafe(tileX+1, tileY+1)

	// Interpolate
	ao0 := ao00*(1-fracX) + ao10*fracX
	ao1 := ao01*(1-fracX) + ao11*fracX
	return ao0*(1-fracY) + ao1*fracY
}

// getAOSafe returns AO with bounds checking.
func (s *System) getAOSafe(x, y int) float64 {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return 0
	}
	return s.aoMap[y][x]
}

// GetEdgeType returns the edge classification for a tile.
func (s *System) GetEdgeType(x, y int) EdgeType {
	if s.edgeMap == nil || x < 0 || x >= s.width || y < 0 || y >= s.height {
		return EdgeNone
	}
	return s.edgeMap[y][x]
}

// Update implements the ECS System interface.
// Edge AO is precomputed, so Update does nothing.
func (s *System) Update(w *engine.World) {
	// No per-frame updates needed - AO is static for level geometry
}

// GetPreset returns the current genre preset (for testing/debugging).
func (s *System) GetPreset() GenrePreset {
	return s.preset
}

// ApplyAO applies ambient occlusion to a color.
// Returns darkened color based on AO factor.
func ApplyAO(r, g, b uint8, aoFactor float64) (uint8, uint8, uint8) {
	mult := 1.0 - aoFactor
	return uint8(float64(r) * mult),
		uint8(float64(g) * mult),
		uint8(float64(b) * mult)
}
