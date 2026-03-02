package lighting

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// AOSystem calculates ambient occlusion for entities based on nearby geometry.
// Uses raycasting to sample occlusion in multiple directions.
type AOSystem struct {
	genre           string
	maxDistance     float64 // Maximum distance to check for occluders
	sampleCount     int     // Number of rays per direction
	baseIntensity   float64 // Genre-specific base occlusion intensity
	wallOcclusion   float64 // Occlusion strength from walls
	entityOcclusion float64 // Occlusion strength from entities
	logger          *logrus.Entry
	updateInterval  int // Update every N frames (performance)
	tick            int
	spatialGrid     SpatialGrid // Optional spatial partitioning for fast queries
}

// SpatialGrid interface for entity proximity queries.
type SpatialGrid interface {
	QueryRadius(x, y, radius float64) []engine.Entity
}

// NewAOSystem creates an ambient occlusion system.
func NewAOSystem(genreID string) *AOSystem {
	s := &AOSystem{
		genre:          genreID,
		maxDistance:    3.0, // Sample within 3 units
		sampleCount:    4,   // 4 samples per octant direction
		updateInterval: 4,   // Update every 4 frames (15 Hz at 60 FPS)
		tick:           0,
		logger: logrus.WithFields(logrus.Fields{
			"system": "ambient_occlusion",
			"genre":  genreID,
		}),
	}
	s.applyGenreSettings(genreID)
	return s
}

// applyGenreSettings configures AO intensity for different genres.
func (s *AOSystem) applyGenreSettings(genreID string) {
	switch genreID {
	case "fantasy":
		// Strong AO in dungeons — emphasizes cramped corridors and alcoves
		s.baseIntensity = 0.7
		s.wallOcclusion = 0.9
		s.entityOcclusion = 0.3
	case "scifi":
		// Moderate AO — clean facilities with defined edges
		s.baseIntensity = 0.5
		s.wallOcclusion = 0.7
		s.entityOcclusion = 0.2
	case "horror":
		// Very strong AO — deep shadows in crevices
		s.baseIntensity = 0.85
		s.wallOcclusion = 1.0
		s.entityOcclusion = 0.5
	case "cyberpunk":
		// Strong local AO — cluttered urban environments
		s.baseIntensity = 0.65
		s.wallOcclusion = 0.75
		s.entityOcclusion = 0.4
	case "postapoc":
		// Moderate AO with dusty ambient
		s.baseIntensity = 0.6
		s.wallOcclusion = 0.8
		s.entityOcclusion = 0.3
	default:
		s.baseIntensity = 0.5
		s.wallOcclusion = 0.7
		s.entityOcclusion = 0.2
	}
}

// SetGenre updates the system for a new genre.
func (s *AOSystem) SetGenre(genreID string) {
	s.genre = genreID
	s.applyGenreSettings(genreID)
}

// SetSpatialGrid assigns a spatial partitioning structure for fast proximity queries.
func (s *AOSystem) SetSpatialGrid(grid SpatialGrid) {
	s.spatialGrid = grid
}

// Update processes all AO components and recalculates occlusion.
func (s *AOSystem) Update(w *engine.World) {
	s.tick++

	// Throttle updates for performance
	if s.tick%s.updateInterval != 0 {
		return
	}

	aoType := reflect.TypeOf(&AOComponent{})
	positionType := reflect.TypeOf(&PositionComponent{})
	colliderType := reflect.TypeOf(&ColliderComponent{})

	entities := w.Query(aoType, positionType)

	for _, entity := range entities {
		aoComp, ok := w.GetComponent(entity, aoType)
		if !ok {
			continue
		}
		ao, ok := aoComp.(*AOComponent)
		if !ok {
			continue
		}

		// Skip if cache is valid
		if ao.IsValid() {
			continue
		}

		posComp, ok := w.GetComponent(entity, positionType)
		if !ok {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate occlusion for all 8 directions
		s.calculateOcclusion(w, entity, pos, ao, positionType, colliderType)

		// Mark as valid
		ao.needsUpdate = false
	}
}

// calculateOcclusion samples occlusion in all 8 cardinal/diagonal directions.
func (s *AOSystem) calculateOcclusion(
	w *engine.World,
	entity engine.Entity,
	pos *PositionComponent,
	ao *AOComponent,
	positionType, colliderType reflect.Type,
) {
	// Sample 8 directions: N, NE, E, SE, S, SW, W, NW
	directions := []struct {
		name  string
		angle float64
		dst   *float64
	}{
		{"E", 0.0, &ao.East},
		{"NE", math.Pi / 4, &ao.NorthEast},
		{"N", math.Pi / 2, &ao.North},
		{"NW", 3 * math.Pi / 4, &ao.NorthWest},
		{"W", math.Pi, &ao.West},
		{"SW", 5 * math.Pi / 4, &ao.SouthWest},
		{"S", 3 * math.Pi / 2, &ao.South},
		{"SE", 7 * math.Pi / 4, &ao.SouthEast},
	}

	totalOcclusion := 0.0

	for _, dir := range directions {
		occlusion := s.sampleDirection(w, entity, pos, dir.angle, ao.SampleRadius, positionType, colliderType)
		*dir.dst = 1.0 - occlusion*s.baseIntensity*ao.IntensityMultiplier
		totalOcclusion += occlusion
	}

	// Calculate overall occlusion (average)
	ao.Overall = 1.0 - (totalOcclusion/8.0)*s.baseIntensity*ao.IntensityMultiplier

	// Clamp all values to [0, 1]
	ao.North = clampF(ao.North, 0.0, 1.0)
	ao.South = clampF(ao.South, 0.0, 1.0)
	ao.East = clampF(ao.East, 0.0, 1.0)
	ao.West = clampF(ao.West, 0.0, 1.0)
	ao.NorthEast = clampF(ao.NorthEast, 0.0, 1.0)
	ao.NorthWest = clampF(ao.NorthWest, 0.0, 1.0)
	ao.SouthEast = clampF(ao.SouthEast, 0.0, 1.0)
	ao.SouthWest = clampF(ao.SouthWest, 0.0, 1.0)
	ao.Overall = clampF(ao.Overall, 0.0, 1.0)
}

// sampleDirection casts rays in a direction and returns occlusion factor [0.0-1.0].
func (s *AOSystem) sampleDirection(
	w *engine.World,
	entity engine.Entity,
	pos *PositionComponent,
	angle float64,
	maxDist float64,
	positionType, colliderType reflect.Type,
) float64 {
	totalOcclusion := 0.0
	validSamples := 0

	// Cast multiple rays at slightly different angles for smoother results
	angleSpread := math.Pi / 16 // ±11.25 degrees
	for i := 0; i < s.sampleCount; i++ {
		t := float64(i)/float64(s.sampleCount-1) - 0.5 // [-0.5, 0.5]
		rayAngle := angle + t*angleSpread
		rayDX := math.Cos(rayAngle)
		rayDY := math.Sin(rayAngle)

		// Sample at multiple distances
		for dist := 0.5; dist <= maxDist; dist += 0.5 {
			sampleX := pos.X + rayDX*dist
			sampleY := pos.Y + rayDY*dist

			// Check for wall collision (assume grid-based world)
			if s.isWall(sampleX, sampleY) {
				// Distance-based falloff: closer walls occlude more
				falloff := 1.0 - (dist / maxDist)
				totalOcclusion += s.wallOcclusion * falloff
				validSamples++
				break
			}

			// Check for nearby entities
			if s.checkEntityOcclusion(w, entity, sampleX, sampleY, positionType, colliderType) {
				falloff := 1.0 - (dist / maxDist)
				totalOcclusion += s.entityOcclusion * falloff
				validSamples++
				break
			}
		}
		validSamples++
	}

	if validSamples == 0 {
		return 0.0
	}

	return totalOcclusion / float64(validSamples)
}

// isWall checks if a world position is solid (wall).
// Uses simple grid-based collision detection.
func (s *AOSystem) isWall(x, y float64) bool {
	// Convert world coordinates to grid coordinates
	gridX := int(x)
	gridY := int(y)

	// Assume walls at grid boundaries (stub — would integrate with actual map)
	// For now, check if position is near integer boundaries
	fracX := x - float64(gridX)
	fracY := y - float64(gridY)

	// Consider positions within 0.1 units of grid lines as walls
	if fracX < 0.1 || fracX > 0.9 || fracY < 0.1 || fracY > 0.9 {
		return true
	}

	return false
}

// checkEntityOcclusion checks if there's an occluding entity near a sample point.
func (s *AOSystem) checkEntityOcclusion(
	w *engine.World,
	selfEntity engine.Entity,
	x, y float64,
	positionType, colliderType reflect.Type,
) bool {
	// Use spatial grid if available for fast queries
	if s.spatialGrid != nil {
		nearby := s.spatialGrid.QueryRadius(x, y, 0.5)
		for _, other := range nearby {
			if other == selfEntity {
				continue
			}

			posComp, ok := w.GetComponent(other, positionType)
			if !ok {
				continue
			}
			otherPos, ok := posComp.(*PositionComponent)
			if !ok {
				continue
			}

			// Check if entity has AO component and casts occlusion
			aoType := reflect.TypeOf(&AOComponent{})
			if aoComp, found := w.GetComponent(other, aoType); found {
				if ao, ok := aoComp.(*AOComponent); ok && !ao.CastsOcclusion {
					continue
				}
			}

			// Simple distance check
			dx := otherPos.X - x
			dy := otherPos.Y - y
			distSq := dx*dx + dy*dy

			if distSq < 0.25 { // Within 0.5 unit radius
				return true
			}
		}
		return false
	}

	// Fallback: brute-force check all entities with position + collider
	entities := w.Query(positionType, colliderType)
	for _, other := range entities {
		if other == selfEntity {
			continue
		}

		posComp, ok := w.GetComponent(other, positionType)
		if !ok {
			continue
		}
		otherPos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Check if entity has AO component and casts occlusion
		aoType := reflect.TypeOf(&AOComponent{})
		if aoComp, found := w.GetComponent(other, aoType); found {
			if ao, ok := aoComp.(*AOComponent); ok && !ao.CastsOcclusion {
				continue
			}
		}

		dx := otherPos.X - x
		dy := otherPos.Y - y
		distSq := dx*dx + dy*dy

		if distSq < 0.25 {
			return true
		}
	}

	return false
}

// InvalidateAll marks all AO components as needing recalculation.
// Call this when the map changes significantly.
func (s *AOSystem) InvalidateAll(w *engine.World) {
	aoType := reflect.TypeOf(&AOComponent{})
	entities := w.Query(aoType)

	for _, entity := range entities {
		aoComp, ok := w.GetComponent(entity, aoType)
		if !ok {
			continue
		}
		if ao, ok := aoComp.(*AOComponent); ok {
			ao.Invalidate()
		}
	}

	s.logger.Debug("Invalidated all AO components")
}

// ColliderComponent represents a simple collider (defined for reference).
type ColliderComponent struct {
	Radius float64
	Solid  bool
}

// Type returns the component type identifier.
func (c *ColliderComponent) Type() string {
	return "Collider"
}
