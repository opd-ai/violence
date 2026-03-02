// Package floor provides procedural floor tile variation and environmental detail.
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
