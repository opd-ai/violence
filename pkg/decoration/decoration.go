// Package decoration provides room decoration and environmental storytelling for procedural dungeons.
package decoration

import (
	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// RoomType defines the purpose and theme of a room.
type RoomType int

const (
	RoomGeneric RoomType = iota
	RoomArmory
	RoomLibrary
	RoomShrine
	RoomTreasure
	RoomPrison
	RoomBarracks
	RoomLaboratory
	RoomStorage
	RoomBoss
)

// DecoType defines decoration element types.
type DecoType int

const (
	DecoFurniture DecoType = iota
	DecoObstacle
	DecoDetail
	DecoLandmark
)

// Decoration represents a decorative element in a room.
type Decoration struct {
	X, Y     int
	Type     DecoType
	SpriteID int
	Blocking bool
	Seeded   bool
	RoomType RoomType
	GenreID  string
}

// RoomDecor holds all decorations for a room.
type RoomDecor struct {
	RoomType    RoomType
	Decorations []Decoration
}

// System manages room decoration and environmental storytelling.
type System struct {
	genre    string
	genreCfg *GenreConfig
}

// GenreConfig defines decoration parameters per genre.
type GenreConfig struct {
	FurnitureDensity float64
	ObstacleDensity  float64
	DetailDensity    float64
	LandmarkChance   float64
	RoomTypeWeights  map[RoomType]float64
}

var genreConfigs = map[string]*GenreConfig{
	genre.Fantasy: {
		FurnitureDensity: 0.15,
		ObstacleDensity:  0.10,
		DetailDensity:    0.25,
		LandmarkChance:   0.20,
		RoomTypeWeights: map[RoomType]float64{
			RoomArmory:   0.15,
			RoomLibrary:  0.10,
			RoomShrine:   0.15,
			RoomTreasure: 0.10,
			RoomPrison:   0.05,
			RoomBarracks: 0.10,
			RoomGeneric:  0.35,
		},
	},
	genre.SciFi: {
		FurnitureDensity: 0.20,
		ObstacleDensity:  0.15,
		DetailDensity:    0.30,
		LandmarkChance:   0.15,
		RoomTypeWeights: map[RoomType]float64{
			RoomArmory:     0.15,
			RoomLaboratory: 0.20,
			RoomStorage:    0.15,
			RoomBarracks:   0.10,
			RoomGeneric:    0.40,
		},
	},
	genre.Horror: {
		FurnitureDensity: 0.12,
		ObstacleDensity:  0.18,
		DetailDensity:    0.35,
		LandmarkChance:   0.25,
		RoomTypeWeights: map[RoomType]float64{
			RoomPrison:  0.15,
			RoomShrine:  0.10,
			RoomLibrary: 0.05,
			RoomStorage: 0.10,
			RoomGeneric: 0.60,
		},
	},
	genre.Cyberpunk: {
		FurnitureDensity: 0.18,
		ObstacleDensity:  0.20,
		DetailDensity:    0.40,
		LandmarkChance:   0.20,
		RoomTypeWeights: map[RoomType]float64{
			RoomLaboratory: 0.20,
			RoomStorage:    0.15,
			RoomArmory:     0.10,
			RoomGeneric:    0.55,
		},
	},
	genre.PostApoc: {
		FurnitureDensity: 0.10,
		ObstacleDensity:  0.25,
		DetailDensity:    0.30,
		LandmarkChance:   0.15,
		RoomTypeWeights: map[RoomType]float64{
			RoomStorage:  0.20,
			RoomArmory:   0.15,
			RoomBarracks: 0.10,
			RoomGeneric:  0.55,
		},
	},
}

// NewSystem creates a decoration system.
func NewSystem() *System {
	return &System{
		genre:    genre.Fantasy,
		genreCfg: genreConfigs[genre.Fantasy],
	}
}

// SetGenre configures the system for a specific genre.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	if cfg, ok := genreConfigs[genreID]; ok {
		s.genreCfg = cfg
	} else {
		s.genreCfg = genreConfigs[genre.Fantasy]
	}
	logrus.WithFields(logrus.Fields{
		"system": "decoration",
		"genre":  genreID,
	}).Debug("Genre set")
}

// DetermineRoomType assigns a room type based on size, position, and genre.
func (s *System) DetermineRoomType(width, height, roomIndex, totalRooms int, r *rng.RNG) RoomType {
	area := width * height

	// Boss room: last large room
	if roomIndex == totalRooms-1 && area > 80 {
		return RoomBoss
	}

	// Treasure: medium rooms, not too early
	if area >= 40 && area <= 70 && roomIndex > totalRooms/3 && r.Float64() < 0.15 {
		return RoomTreasure
	}

	// Use weighted random selection for other types
	totalWeight := 0.0
	for _, w := range s.genreCfg.RoomTypeWeights {
		totalWeight += w
	}

	roll := r.Float64() * totalWeight
	cumulative := 0.0
	for rt, w := range s.genreCfg.RoomTypeWeights {
		cumulative += w
		if roll <= cumulative {
			return rt
		}
	}

	return RoomGeneric
}

// DecorateRoom generates decorations for a room based on type and genre.
func (s *System) DecorateRoom(roomType RoomType, x, y, width, height int, tiles [][]int, r *rng.RNG) *RoomDecor {
	decor := &RoomDecor{
		RoomType:    roomType,
		Decorations: make([]Decoration, 0, 16),
	}

	// Place landmark decoration first (center focal point)
	if r.Float64() < s.genreCfg.LandmarkChance {
		s.placeLandmark(decor, roomType, x, y, width, height, tiles, r)
	}

	// Place furniture along walls
	s.placeFurniture(decor, roomType, x, y, width, height, tiles, r)

	// Scatter obstacles
	s.placeObstacles(decor, roomType, x, y, width, height, tiles, r)

	// Add detail elements
	s.placeDetails(decor, roomType, x, y, width, height, tiles, r)

	logrus.WithFields(logrus.Fields{
		"system":      "decoration",
		"room_type":   roomType,
		"decorations": len(decor.Decorations),
	}).Debug("Room decorated")

	return decor
}

// placeLandmark creates a central focal point decoration.
func (s *System) placeLandmark(decor *RoomDecor, roomType RoomType, x, y, width, height int, tiles [][]int, r *rng.RNG) {
	centerX := x + width/2
	centerY := y + height/2

	if !s.isWalkable(centerX, centerY, tiles) {
		return
	}

	spriteID := s.getLandmarkSprite(roomType, r)
	decor.Decorations = append(decor.Decorations, Decoration{
		X:        centerX,
		Y:        centerY,
		Type:     DecoLandmark,
		SpriteID: spriteID,
		Blocking: true,
		Seeded:   true,
		RoomType: roomType,
		GenreID:  s.genre,
	})
}

// placeFurniture adds furniture along walls.
func (s *System) placeFurniture(decor *RoomDecor, roomType RoomType, x, y, width, height int, tiles [][]int, r *rng.RNG) {
	count := int(float64(width+height) * s.genreCfg.FurnitureDensity)
	placed := 0

	for attempt := 0; attempt < count*3 && placed < count; attempt++ {
		fx := x + 1 + r.Intn(width-2)
		fy := y + 1 + r.Intn(height-2)

		if !s.isWalkable(fx, fy, tiles) {
			continue
		}

		// Prefer wall-adjacent positions
		if !s.isNearWall(fx, fy, tiles) && r.Float64() > 0.3 {
			continue
		}

		// Check spacing
		if s.tooClose(fx, fy, decor.Decorations, 2) {
			continue
		}

		spriteID := s.getFurnitureSprite(roomType, r)
		decor.Decorations = append(decor.Decorations, Decoration{
			X:        fx,
			Y:        fy,
			Type:     DecoFurniture,
			SpriteID: spriteID,
			Blocking: true,
			Seeded:   true,
			RoomType: roomType,
			GenreID:  s.genre,
		})
		placed++
	}
}

// placeObstacles scatters blocking obstacles.
func (s *System) placeObstacles(decor *RoomDecor, roomType RoomType, x, y, width, height int, tiles [][]int, r *rng.RNG) {
	count := int(float64(width*height) * s.genreCfg.ObstacleDensity / 10)
	placed := 0

	for attempt := 0; attempt < count*4 && placed < count; attempt++ {
		ox := x + 1 + r.Intn(width-2)
		oy := y + 1 + r.Intn(height-2)

		if !s.isWalkable(ox, oy, tiles) {
			continue
		}

		// Check spacing and avoid blocking paths
		if s.tooClose(ox, oy, decor.Decorations, 2) {
			continue
		}

		// Don't create choke points
		if s.blocksMovement(ox, oy, tiles) {
			continue
		}

		spriteID := s.getObstacleSprite(roomType, r)
		decor.Decorations = append(decor.Decorations, Decoration{
			X:        ox,
			Y:        oy,
			Type:     DecoObstacle,
			SpriteID: spriteID,
			Blocking: true,
			Seeded:   true,
			RoomType: roomType,
			GenreID:  s.genre,
		})
		placed++
	}
}

// placeDetails adds non-blocking visual elements.
func (s *System) placeDetails(decor *RoomDecor, roomType RoomType, x, y, width, height int, tiles [][]int, r *rng.RNG) {
	count := int(float64(width*height) * s.genreCfg.DetailDensity / 10)
	placed := 0

	for attempt := 0; attempt < count*2 && placed < count; attempt++ {
		dx := x + 1 + r.Intn(width-2)
		dy := y + 1 + r.Intn(height-2)

		if !s.isWalkable(dx, dy, tiles) {
			continue
		}

		spriteID := s.getDetailSprite(roomType, r)
		decor.Decorations = append(decor.Decorations, Decoration{
			X:        dx,
			Y:        dy,
			Type:     DecoDetail,
			SpriteID: spriteID,
			Blocking: false,
			Seeded:   true,
			RoomType: roomType,
			GenreID:  s.genre,
		})
		placed++
	}
}

// getLandmarkSprite returns a sprite ID for a landmark based on room type.
func (s *System) getLandmarkSprite(roomType RoomType, r *rng.RNG) int {
	base := 1000 + int(roomType)*100
	return base + r.Intn(10)
}

// getFurnitureSprite returns a sprite ID for furniture.
func (s *System) getFurnitureSprite(roomType RoomType, r *rng.RNG) int {
	base := 2000 + int(roomType)*100
	return base + r.Intn(20)
}

// getObstacleSprite returns a sprite ID for obstacles.
func (s *System) getObstacleSprite(roomType RoomType, r *rng.RNG) int {
	base := 3000 + int(roomType)*100
	return base + r.Intn(15)
}

// getDetailSprite returns a sprite ID for details.
func (s *System) getDetailSprite(roomType RoomType, r *rng.RNG) int {
	base := 4000 + int(roomType)*100
	return base + r.Intn(30)
}

// isWalkable checks if a tile is walkable.
func (s *System) isWalkable(x, y int, tiles [][]int) bool {
	if y < 0 || y >= len(tiles) || x < 0 || x >= len(tiles[0]) {
		return false
	}
	tile := tiles[y][x]
	return tile >= 2 && tile < 10
}

// isNearWall checks if a position is adjacent to a wall.
func (s *System) isNearWall(x, y int, tiles [][]int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if ny >= 0 && ny < len(tiles) && nx >= 0 && nx < len(tiles[0]) {
				if tiles[ny][nx] == 1 || tiles[ny][nx] >= 10 && tiles[ny][nx] < 20 {
					return true
				}
			}
		}
	}
	return false
}

// tooClose checks if position is too close to existing decorations.
func (s *System) tooClose(x, y int, decorations []Decoration, minDist int) bool {
	for _, d := range decorations {
		dx := x - d.X
		dy := y - d.Y
		if dx*dx+dy*dy < minDist*minDist {
			return true
		}
	}
	return false
}

// blocksMovement checks if placing obstacle would create impassable area.
func (s *System) blocksMovement(x, y int, tiles [][]int) bool {
	walkable := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if s.isWalkable(nx, ny, tiles) {
				walkable++
			}
		}
	}
	return walkable < 4
}

// GetRoomTypeName returns human-readable room type name.
func GetRoomTypeName(rt RoomType) string {
	switch rt {
	case RoomArmory:
		return "Armory"
	case RoomLibrary:
		return "Library"
	case RoomShrine:
		return "Shrine"
	case RoomTreasure:
		return "Treasure Room"
	case RoomPrison:
		return "Prison"
	case RoomBarracks:
		return "Barracks"
	case RoomLaboratory:
		return "Laboratory"
	case RoomStorage:
		return "Storage"
	case RoomBoss:
		return "Boss Arena"
	default:
		return "Chamber"
	}
}
