// Package floor provides procedural floor tile variation and environmental detail.
//
// This package generates two types of floor visuals:
//
// 1. Base Floor Tiles (FloorTileComponent):
//   - Material-specific textures (stone, metal, wood, concrete, tile, dirt, grass, flesh, crystal)
//   - Procedural generation with grain patterns, wear, and shading
//   - Genre-aware material distribution (fantasy favors stone, scifi favors metal, etc.)
//   - LRU-cached texture generation for performance
//   - Seed-based consistency across sessions
//
// 2. Detail Overlays (FloorDetailComponent):
//   - Environmental detail sprites (cracks, stains, debris, scorch marks, etc.)
//   - Genre-specific detail types and density
//   - Intensity-based rendering with blend modes
//   - Positioned near walls for increased realism
//
// Usage:
//
//	sys := floor.NewSystem("fantasy", 64)
//
//	// Generate base tiles for the level
//	tiles := sys.GenerateFloorTiles(dungeonMap, seed)
//
//	// Generate detail overlays
//	details := sys.GenerateFloorDetails(dungeonMap, seed)
//
//	// Render in game loop
//	for _, tile := range tiles {
//	    img := sys.RenderTile(tile)
//	    screen.DrawImage(img, op)
//	}
//
//	for _, detail := range details {
//	    img := sys.RenderDetail(detail)
//	    screen.DrawImage(img, op)
//	}
package floor

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// DetailType categorizes floor detail overlays.
type DetailType int

const (
	DetailNone DetailType = iota
	DetailCrack
	DetailStain
	DetailDebris
	DetailScorch
	DetailWear
	DetailGraffiti
	DetailBlood
	DetailRust
	DetailCorrode
)

// FloorDetailComponent stores visual variation data for a floor tile.
type FloorDetailComponent struct {
	X, Y         int
	DetailType   DetailType
	Intensity    float64 // 0.0 to 1.0
	Seed         int64
	GenreID      string
	CachedSprite *ebiten.Image
}

// Type implements engine.Component interface.
func (f *FloorDetailComponent) Type() string {
	return "floor_detail"
}

// MaterialType defines floor material categories for texture generation.
type MaterialType int

const (
	MaterialStone MaterialType = iota
	MaterialMetal
	MaterialWood
	MaterialConcrete
	MaterialTile
	MaterialDirt
	MaterialGrass
	MaterialFlesh
	MaterialCrystal
)

// FloorTileComponent stores base material texture data for a floor tile.
type FloorTileComponent struct {
	X, Y          int
	Material      MaterialType
	Variant       int // Variant index for same material (0-3)
	Seed          int64
	GenreID       string
	CachedTexture *ebiten.Image
}

// Type implements engine.Component interface.
func (f *FloorTileComponent) Type() string {
	return "floor_tile"
}
