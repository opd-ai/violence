// Package biome provides biome material integration with ECS.
package biome

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// MaterialComponent marks an entity as a biome material item.
type MaterialComponent struct {
	MaterialID string
	Amount     int
	BiomeType  BiomeType
	SpawnTime  float64
}

// Type returns the component type identifier.
func (c *MaterialComponent) Type() string {
	return "MaterialComponent"
}

// PositionComponent stores entity position (matches engine definition).
type PositionComponent struct {
	X, Y float64
}

// Type returns the component type identifier.
func (c *PositionComponent) Type() string {
	return "PositionComponent"
}

// BiomeMaterialSystem manages biome-specific material spawning.
type BiomeMaterialSystem struct {
	gameTime            float64
	currentGenre        string
	onMaterialSpawned   func(materialID string, amount int, x, y float64)
	materialEntityCache map[engine.Entity]bool
	logger              *logrus.Entry
}

// NewBiomeMaterialSystem creates the biome material system.
func NewBiomeMaterialSystem(genreID string) *BiomeMaterialSystem {
	return &BiomeMaterialSystem{
		currentGenre:        genreID,
		materialEntityCache: make(map[engine.Entity]bool),
		logger: logrus.WithFields(logrus.Fields{
			"system": "biome_material",
		}),
	}
}

// SetGenre configures the system for a genre.
func (s *BiomeMaterialSystem) SetGenre(genreID string) {
	s.currentGenre = genreID
	s.logger.WithField("genre", genreID).Debug("Biome material system genre set")
}

// SetMaterialSpawnCallback sets the callback for material spawning.
func (s *BiomeMaterialSystem) SetMaterialSpawnCallback(fn func(materialID string, amount int, x, y float64)) {
	s.onMaterialSpawned = fn
}

// Update processes biome material entities.
func (s *BiomeMaterialSystem) Update(w *engine.World) {
	deltaTime := 0.016 // Assume 60 FPS
	s.gameTime += deltaTime

	// Query all material entities to keep cache updated
	matType := reflect.TypeOf((*MaterialComponent)(nil))
	entities := w.Query(matType)

	// Clear cache of removed entities
	if len(s.materialEntityCache) > len(entities)*2 {
		s.materialEntityCache = make(map[engine.Entity]bool)
	}
	for _, e := range entities {
		s.materialEntityCache[e] = true
	}
}

// SpawnMaterialsAtPosition spawns biome materials at a location.
func (s *BiomeMaterialSystem) SpawnMaterialsAtPosition(w *engine.World, biomeType BiomeType, tier int, x, y float64, seed uint64) {
	materials := RollMaterials(biomeType, tier, s.currentGenre, seed)

	for i, mat := range materials {
		offsetX := x + float64(i%3-1)*0.2
		offsetY := y + float64(i/3)*0.2

		matEntity := w.AddEntity()
		w.AddComponent(matEntity, &PositionComponent{X: offsetX, Y: offsetY})
		w.AddComponent(matEntity, &MaterialComponent{
			MaterialID: mat.MaterialID,
			Amount:     mat.MinAmount,
			BiomeType:  biomeType,
			SpawnTime:  s.gameTime,
		})

		s.logger.WithFields(logrus.Fields{
			"material_id": mat.MaterialID,
			"amount":      mat.MinAmount,
			"biome":       biomeType.String(),
			"x":           offsetX,
			"y":           offsetY,
		}).Debug("Material spawned")

		if s.onMaterialSpawned != nil {
			s.onMaterialSpawned(mat.MaterialID, mat.MinAmount, offsetX, offsetY)
		}
	}

	if len(materials) > 0 {
		s.logger.WithFields(logrus.Fields{
			"biome":          biomeType.String(),
			"tier":           tier,
			"material_count": len(materials),
		}).Info("Spawned biome materials")
	}
}

// GetMaterialCount returns count of all material entities.
func (s *BiomeMaterialSystem) GetMaterialCount() int {
	return len(s.materialEntityCache)
}
