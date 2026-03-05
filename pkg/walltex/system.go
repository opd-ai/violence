// Package walltex provides wall texture variation system for enhanced environmental visuals.
package walltex

import (
	"image"
	"image/color"
	"sync"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// WallTextureComponent stores per-wall texture variation data.
type WallTextureComponent struct {
	GridX         int         // Wall grid position X
	GridY         int         // Wall grid position Y
	Material      Material    // Wall material type
	Weathering    float64     // Wear intensity 0.0-1.0
	Variant       int         // Texture variant index
	DetailSeed    uint64      // Seed for detail generation
	CachedTexture image.Image // Cached texture image
}

// Type returns the component type identifier.
func (w *WallTextureComponent) Type() string {
	return "WallTexture"
}

// WallTextureSystem manages procedural wall texture variation across the level.
type System struct {
	genre         string
	generator     *Generator
	textureCache  map[uint64]image.Image
	cacheMutex    sync.RWMutex
	maxCacheSize  int
	cacheHits     int
	cacheMisses   int
	logger        *logrus.Entry
	materialRules map[string]MaterialDistribution
}

// MaterialDistribution defines how materials vary by room type and depth.
type MaterialDistribution struct {
	PrimaryChance   float64
	SecondaryChance float64
	WeatheringBase  float64
	WeatheringRange float64
}

// NewSystem creates a new wall texture variation system.
func NewSystem(genre string, maxCacheSize int) *System {
	return &System{
		genre:        genre,
		generator:    NewGenerator(genre),
		textureCache: make(map[uint64]image.Image),
		maxCacheSize: maxCacheSize,
		logger: logrus.WithFields(logrus.Fields{
			"system": "walltex",
			"genre":  genre,
		}),
		materialRules: buildMaterialRules(genre),
	}
}

// Update processes wall texture component updates (no-op for this system - textures are cached).
func (s *System) Update(w *engine.World) {
	// This system generates textures on-demand rather than per-frame updates
}

// GenerateWallTexture creates or retrieves a cached texture for a wall position.
func (s *System) GenerateWallTexture(gridX, gridY int, roomType string, depth int, seed uint64) *WallTextureComponent {
	// Compute unique seed for this wall position
	wallSeed := hashPosition(gridX, gridY, seed)

	// Check cache first
	s.cacheMutex.RLock()
	if cached, ok := s.textureCache[wallSeed]; ok {
		s.cacheMutex.RUnlock()
		s.cacheHits++
		return &WallTextureComponent{
			GridX:         gridX,
			GridY:         gridY,
			Material:      s.selectMaterial(roomType, depth, wallSeed),
			Weathering:    s.calculateWeathering(roomType, depth, wallSeed),
			Variant:       int(wallSeed % 4),
			DetailSeed:    wallSeed,
			CachedTexture: cached,
		}
	}
	s.cacheMutex.RUnlock()
	s.cacheMisses++

	// Generate new texture
	material := s.selectMaterial(roomType, depth, wallSeed)
	weathering := s.calculateWeathering(roomType, depth, wallSeed)
	variant := int(wallSeed % 4)

	texture := s.generator.GenerateWithMaterial(64, material, variant, weathering, wallSeed)

	// Cache the texture (with size limit)
	s.cacheMutex.Lock()
	if len(s.textureCache) >= s.maxCacheSize {
		// Simple eviction: clear oldest half of cache
		s.evictHalfCache()
	}
	s.textureCache[wallSeed] = texture
	s.cacheMutex.Unlock()

	return &WallTextureComponent{
		GridX:         gridX,
		GridY:         gridY,
		Material:      material,
		Weathering:    weathering,
		Variant:       variant,
		DetailSeed:    wallSeed,
		CachedTexture: texture,
	}
}

// selectMaterial chooses wall material based on room type and depth.
func (s *System) selectMaterial(roomType string, depth int, seed uint64) Material {
	r := rng.NewRNG(seed)
	dist, ok := s.materialRules[roomType]
	if !ok {
		dist = s.materialRules["default"]
	}

	roll := r.Float64()
	if roll < dist.PrimaryChance {
		return s.generator.preset.PrimaryMaterial
	} else if roll < dist.PrimaryChance+dist.SecondaryChance {
		return s.generator.preset.SecondaryMaterial
	}

	// Depth-based material variation
	if depth > 5 && s.genre == "fantasy" {
		return MaterialOrganic // Deeper dungeons get more organic/corrupted walls
	}
	if depth > 8 && s.genre == "scifi" {
		return MaterialCrystal // Deep space stations get exotic materials
	}

	return s.generator.preset.PrimaryMaterial
}

// calculateWeathering determines wear intensity based on room type and depth.
func (s *System) calculateWeathering(roomType string, depth int, seed uint64) float64 {
	r := rng.NewRNG(seed)
	dist, ok := s.materialRules[roomType]
	if !ok {
		dist = s.materialRules["default"]
	}

	// Base weathering from room type
	base := dist.WeatheringBase

	// Add depth-based weathering (deeper = more wear)
	depthFactor := float64(depth) * 0.05

	// Add random variation
	variation := (r.Float64() - 0.5) * dist.WeatheringRange

	weathering := base + depthFactor + variation
	if weathering < 0.0 {
		weathering = 0.0
	}
	if weathering > 1.0 {
		weathering = 1.0
	}

	return weathering
}

// evictHalfCache removes half of cached textures (simple LRU-like eviction).
func (s *System) evictHalfCache() {
	targetSize := len(s.textureCache) / 2
	count := 0
	for key := range s.textureCache {
		delete(s.textureCache, key)
		count++
		if count >= targetSize {
			break
		}
	}
	s.logger.WithField("evicted", count).Debug("Evicted textures from cache")
}

// GetCacheStats returns cache performance statistics.
func (s *System) GetCacheStats() (hits, misses, size int) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	return s.cacheHits, s.cacheMisses, len(s.textureCache)
}

// SampleTexture retrieves a color sample from a cached wall texture.
func (s *System) SampleTexture(comp *WallTextureComponent, u, v float64) color.RGBA {
	if comp.CachedTexture == nil {
		return color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}

	bounds := comp.CachedTexture.Bounds()
	x := int(u * float64(bounds.Dx()))
	y := int(v * float64(bounds.Dy()))

	if x < 0 {
		x = 0
	}
	if x >= bounds.Dx() {
		x = bounds.Dx() - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= bounds.Dy() {
		y = bounds.Dy() - 1
	}

	c := comp.CachedTexture.At(bounds.Min.X+x, bounds.Min.Y+y)
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

// SetGenre updates the genre for the system and regenerates material rules.
func (s *System) SetGenre(genre string) {
	s.genre = genre
	s.generator = NewGenerator(genre)
	s.materialRules = buildMaterialRules(genre)
	s.logger = s.logger.WithField("genre", genre)
}

// buildMaterialRules creates material distribution rules for different room types.
func buildMaterialRules(genre string) map[string]MaterialDistribution {
	rules := make(map[string]MaterialDistribution)

	switch genre {
	case "fantasy":
		rules["corridor"] = MaterialDistribution{
			PrimaryChance:   0.8,
			SecondaryChance: 0.15,
			WeatheringBase:  0.5,
			WeatheringRange: 0.3,
		}
		rules["room"] = MaterialDistribution{
			PrimaryChance:   0.7,
			SecondaryChance: 0.25,
			WeatheringBase:  0.4,
			WeatheringRange: 0.4,
		}
		rules["boss"] = MaterialDistribution{
			PrimaryChance:   0.9,
			SecondaryChance: 0.1,
			WeatheringBase:  0.2,
			WeatheringRange: 0.2,
		}
		rules["treasure"] = MaterialDistribution{
			PrimaryChance:   0.6,
			SecondaryChance: 0.3,
			WeatheringBase:  0.3,
			WeatheringRange: 0.3,
		}
	case "scifi":
		rules["corridor"] = MaterialDistribution{
			PrimaryChance:   0.9,
			SecondaryChance: 0.08,
			WeatheringBase:  0.2,
			WeatheringRange: 0.15,
		}
		rules["room"] = MaterialDistribution{
			PrimaryChance:   0.85,
			SecondaryChance: 0.12,
			WeatheringBase:  0.15,
			WeatheringRange: 0.2,
		}
		rules["boss"] = MaterialDistribution{
			PrimaryChance:   0.5,
			SecondaryChance: 0.4,
			WeatheringBase:  0.1,
			WeatheringRange: 0.1,
		}
		rules["treasure"] = MaterialDistribution{
			PrimaryChance:   0.3,
			SecondaryChance: 0.6,
			WeatheringBase:  0.05,
			WeatheringRange: 0.1,
		}
	case "horror":
		rules["corridor"] = MaterialDistribution{
			PrimaryChance:   0.7,
			SecondaryChance: 0.25,
			WeatheringBase:  0.7,
			WeatheringRange: 0.25,
		}
		rules["room"] = MaterialDistribution{
			PrimaryChance:   0.65,
			SecondaryChance: 0.3,
			WeatheringBase:  0.75,
			WeatheringRange: 0.2,
		}
		rules["boss"] = MaterialDistribution{
			PrimaryChance:   0.5,
			SecondaryChance: 0.45,
			WeatheringBase:  0.85,
			WeatheringRange: 0.15,
		}
		rules["treasure"] = MaterialDistribution{
			PrimaryChance:   0.6,
			SecondaryChance: 0.35,
			WeatheringBase:  0.65,
			WeatheringRange: 0.3,
		}
	case "cyberpunk":
		rules["corridor"] = MaterialDistribution{
			PrimaryChance:   0.85,
			SecondaryChance: 0.1,
			WeatheringBase:  0.5,
			WeatheringRange: 0.25,
		}
		rules["room"] = MaterialDistribution{
			PrimaryChance:   0.75,
			SecondaryChance: 0.2,
			WeatheringBase:  0.45,
			WeatheringRange: 0.3,
		}
		rules["boss"] = MaterialDistribution{
			PrimaryChance:   0.4,
			SecondaryChance: 0.55,
			WeatheringBase:  0.3,
			WeatheringRange: 0.2,
		}
		rules["treasure"] = MaterialDistribution{
			PrimaryChance:   0.3,
			SecondaryChance: 0.65,
			WeatheringBase:  0.25,
			WeatheringRange: 0.2,
		}
	default: // postapoc or unknown
		rules["corridor"] = MaterialDistribution{
			PrimaryChance:   0.75,
			SecondaryChance: 0.2,
			WeatheringBase:  0.8,
			WeatheringRange: 0.15,
		}
		rules["room"] = MaterialDistribution{
			PrimaryChance:   0.7,
			SecondaryChance: 0.25,
			WeatheringBase:  0.85,
			WeatheringRange: 0.1,
		}
		rules["boss"] = MaterialDistribution{
			PrimaryChance:   0.8,
			SecondaryChance: 0.15,
			WeatheringBase:  0.7,
			WeatheringRange: 0.2,
		}
		rules["treasure"] = MaterialDistribution{
			PrimaryChance:   0.6,
			SecondaryChance: 0.35,
			WeatheringBase:  0.75,
			WeatheringRange: 0.2,
		}
	}

	// Add default for unknown room types
	rules["default"] = MaterialDistribution{
		PrimaryChance:   0.8,
		SecondaryChance: 0.15,
		WeatheringBase:  0.5,
		WeatheringRange: 0.3,
	}

	return rules
}

// hashPosition creates a unique hash for a grid position combined with level seed.
func hashPosition(x, y int, seed uint64) uint64 {
	h := seed
	h ^= uint64(x) * 0x9e3779b97f4a7c15
	h ^= uint64(y) * 0x517cc1b727220a95
	h ^= h >> 30
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 27
	h *= 0x94d049bb133111eb
	h ^= h >> 31
	return h
}
