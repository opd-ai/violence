// Package spatial provides ECS integration for spatial indexing.
package spatial

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System maintains spatial indices and provides fast proximity queries.
type System struct {
	grid   *Grid
	logger *logrus.Entry
}

// NewSystem creates a spatial indexing system with the given cell size.
// Recommended cell size: 2-4x your typical query radius.
// For a game with 10-unit attack ranges, use cellSize=32 or 64.
func NewSystem(cellSize float64) *System {
	return &System{
		grid: NewGrid(cellSize),
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "spatial",
		}),
	}
}

// Update rebuilds the spatial index from all entities with Position components.
// This runs each frame to keep the index synchronized with entity movement.
func (s *System) Update(w *engine.World) {
	s.grid.Clear()

	posType := reflect.TypeOf(&engine.Position{})
	entities := w.Query(posType)

	for _, e := range entities {
		comp, ok := w.GetComponent(e, posType)
		if !ok {
			continue
		}

		pos, ok := comp.(*engine.Position)
		if !ok {
			continue
		}

		s.grid.Insert(e, pos.X, pos.Y)
	}
}

// QueryRadius returns all entities within radius of (x, y).
// Returns entities in cells overlapping the query circle (broadphase).
func (s *System) QueryRadius(x, y, radius float64) []engine.Entity {
	return s.grid.QueryRadius(x, y, radius)
}

// QueryRadiusExact returns entities within radius, with exact distance filtering.
// Slower than QueryRadius but provides circular precision.
func (s *System) QueryRadiusExact(w *engine.World, x, y, radius float64) []engine.Entity {
	// Build position map for distance checks
	posType := reflect.TypeOf(&engine.Position{})
	entities := s.grid.QueryRadius(x, y, radius)

	positions := make(map[engine.Entity]*engine.Position)
	for _, e := range entities {
		comp, ok := w.GetComponent(e, posType)
		if !ok {
			continue
		}
		pos, ok := comp.(*engine.Position)
		if ok {
			positions[e] = pos
		}
	}

	return s.grid.QueryRadiusFiltered(x, y, radius, positions)
}

// QueryBounds returns all entities within the axis-aligned bounding box.
func (s *System) QueryBounds(minX, minY, maxX, maxY float64) []engine.Entity {
	return s.grid.QueryBounds(minX, minY, maxX, maxY)
}

// GetGrid returns the underlying spatial grid for advanced usage.
func (s *System) GetGrid() *Grid {
	return s.grid
}

// Count returns the number of indexed entities.
func (s *System) Count() int {
	return s.grid.Count()
}

// CellCount returns the number of occupied cells.
func (s *System) CellCount() int {
	return s.grid.CellCount()
}
