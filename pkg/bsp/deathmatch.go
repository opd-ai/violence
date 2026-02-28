// Package bsp provides deathmatch-specific arena generation.
package bsp

import (
	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

const (
	TileSpawnPad    = 5  // Player spawn location
	TileWeaponSpawn = 6  // Weapon pickup location
	TileCover       = 15 // Low cover (half-height wall)
)

// WeaponSpawn represents a weapon pickup location in the arena.
type WeaponSpawn struct {
	X, Y       int
	WeaponType string // e.g., "shotgun", "rifle", "rocket"
}

// ArenaGenerator produces balanced deathmatch arenas using BSP.
type ArenaGenerator struct {
	Width        int
	Height       int
	rng          *rng.RNG
	genre        string
	wallTile     int
	floorTile    int
	SpawnPads    []Room        // Symmetrical spawn locations
	WeaponSpawns []WeaponSpawn // Weapon pickup locations
	CoverPoints  []Room        // Cover positions for tactical gameplay
	sightlineMap [][]int       // Tracks open sightlines for balancing
}

// NewArenaGenerator creates a deathmatch arena generator.
func NewArenaGenerator(width, height int, r *rng.RNG) *ArenaGenerator {
	return &ArenaGenerator{
		Width:        width,
		Height:       height,
		rng:          r,
		genre:        genre.Fantasy,
		wallTile:     TileWall,
		floorTile:    TileFloor,
		SpawnPads:    make([]Room, 0),
		WeaponSpawns: make([]WeaponSpawn, 0),
		CoverPoints:  make([]Room, 0),
		sightlineMap: nil,
	}
}

// SetGenre configures arena generation for a genre.
func (g *ArenaGenerator) SetGenre(genreID string) {
	g.genre = genreID

	switch genreID {
	case genre.Fantasy:
		g.wallTile = TileWallStone
		g.floorTile = TileFloorStone
	case genre.SciFi:
		g.wallTile = TileWallHull
		g.floorTile = TileFloorHull
	case genre.Horror:
		g.wallTile = TileWallPlaster
		g.floorTile = TileFloorWood
	case genre.Cyberpunk:
		g.wallTile = TileWallConcrete
		g.floorTile = TileFloorConcrete
	case genre.PostApoc:
		g.wallTile = TileWallRust
		g.floorTile = TileFloorDirt
	default:
		g.wallTile = TileWall
		g.floorTile = TileFloor
	}
}

// Generate creates a symmetrical arena layout optimized for deathmatch.
func (g *ArenaGenerator) Generate() [][]int {
	tiles := make([][]int, g.Height)
	for y := range tiles {
		tiles[y] = make([]int, g.Width)
		for x := range tiles[y] {
			tiles[y][x] = g.wallTile
		}
	}

	// Create main arena floor (80% of space is open)
	centerX := g.Width / 2
	centerY := g.Height / 2
	arenaW := (g.Width * 8) / 10
	arenaH := (g.Height * 8) / 10

	// Carve rectangular arena with rounded corners
	cornerRadius := min(arenaW, arenaH) / 8
	g.carveArena(tiles, centerX-arenaW/2, centerY-arenaH/2, arenaW, arenaH, cornerRadius)

	// Place symmetrical spawn pads (4-way symmetry)
	g.placeSymmetricalSpawns(tiles, centerX, centerY, arenaW, arenaH)

	// Place weapon spawns at strategic locations
	g.placeWeaponSpawns(tiles, centerX, centerY, arenaW, arenaH)

	// Add cover points for tactical gameplay
	g.placeCoverPoints(tiles, centerX, centerY, arenaW, arenaH)

	// Balance sightlines
	g.analyzeSightlines(tiles)
	g.balanceSightlines(tiles)

	return tiles
}

// carveArena creates the main arena floor with optional rounded corners.
func (g *ArenaGenerator) carveArena(tiles [][]int, x, y, w, h, cornerRadius int) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			px := x + dx
			py := y + dy

			if px < 0 || px >= g.Width || py < 0 || py >= g.Height {
				continue
			}

			// Check if in corner region
			inCorner := false
			cornerDist := 0

			// Top-left corner
			if dx < cornerRadius && dy < cornerRadius {
				inCorner = true
				cornerDist = (cornerRadius-dx)*(cornerRadius-dx) + (cornerRadius-dy)*(cornerRadius-dy)
			}
			// Top-right corner
			if dx >= w-cornerRadius && dy < cornerRadius {
				inCorner = true
				cornerDist = (dx-(w-cornerRadius))*(dx-(w-cornerRadius)) + (cornerRadius-dy)*(cornerRadius-dy)
			}
			// Bottom-left corner
			if dx < cornerRadius && dy >= h-cornerRadius {
				inCorner = true
				cornerDist = (cornerRadius-dx)*(cornerRadius-dx) + (dy-(h-cornerRadius))*(dy-(h-cornerRadius))
			}
			// Bottom-right corner
			if dx >= w-cornerRadius && dy >= h-cornerRadius {
				inCorner = true
				cornerDist = (dx-(w-cornerRadius))*(dx-(w-cornerRadius)) + (dy-(h-cornerRadius))*(dy-(h-cornerRadius))
			}

			// Only carve if not in corner or within corner radius
			if !inCorner || cornerDist <= cornerRadius*cornerRadius {
				tiles[py][px] = g.floorTile
			}
		}
	}
}

// placeSymmetricalSpawns places spawn pads with 4-way rotational symmetry.
func (g *ArenaGenerator) placeSymmetricalSpawns(tiles [][]int, centerX, centerY, arenaW, arenaH int) {
	// Place spawns at 4 corners of a smaller inner rectangle (70% of arena size)
	spawnOffsetX := (arenaW * 7) / 20
	spawnOffsetY := (arenaH * 7) / 20

	spawns := []struct{ x, y int }{
		{centerX - spawnOffsetX, centerY - spawnOffsetY}, // Top-left
		{centerX + spawnOffsetX, centerY - spawnOffsetY}, // Top-right
		{centerX - spawnOffsetX, centerY + spawnOffsetY}, // Bottom-left
		{centerX + spawnOffsetX, centerY + spawnOffsetY}, // Bottom-right
	}

	padSize := 3
	for _, spawn := range spawns {
		g.placeSpawnPad(tiles, spawn.x-padSize/2, spawn.y-padSize/2, padSize, padSize)
	}
}

// placeSpawnPad marks a spawn pad area on the map.
func (g *ArenaGenerator) placeSpawnPad(tiles [][]int, x, y, w, h int) {
	room := Room{X: x, Y: y, W: w, H: h}
	g.SpawnPads = append(g.SpawnPads, room)

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < g.Width && py >= 0 && py < g.Height {
				tiles[py][px] = TileSpawnPad
			}
		}
	}
}

// placeWeaponSpawns positions weapon pickups at strategic points.
func (g *ArenaGenerator) placeWeaponSpawns(tiles [][]int, centerX, centerY, arenaW, arenaH int) {
	// Weapon spawn positions (symmetrical placement)
	weapons := []struct {
		x, y       int
		weaponType string
	}{
		// Center weapon (power weapon)
		{centerX, centerY, "rocket"},

		// Cardinal directions (mid-tier weapons)
		{centerX, centerY - arenaH/3, "shotgun"},
		{centerX, centerY + arenaH/3, "shotgun"},
		{centerX - arenaW/3, centerY, "rifle"},
		{centerX + arenaW/3, centerY, "rifle"},

		// Diagonal positions (basic weapons)
		{centerX - arenaW/4, centerY - arenaH/4, "pistol"},
		{centerX + arenaW/4, centerY - arenaH/4, "pistol"},
		{centerX - arenaW/4, centerY + arenaH/4, "pistol"},
		{centerX + arenaW/4, centerY + arenaH/4, "pistol"},
	}

	for _, wp := range weapons {
		if wp.x >= 0 && wp.x < g.Width && wp.y >= 0 && wp.y < g.Height {
			if tiles[wp.y][wp.x] == g.floorTile {
				tiles[wp.y][wp.x] = TileWeaponSpawn
				g.WeaponSpawns = append(g.WeaponSpawns, WeaponSpawn{
					X:          wp.x,
					Y:          wp.y,
					WeaponType: wp.weaponType,
				})
			}
		}
	}
}

// placeCoverPoints adds tactical cover positions.
func (g *ArenaGenerator) placeCoverPoints(tiles [][]int, centerX, centerY, arenaW, arenaH int) {
	// Place cover in a ring around the center
	coverRadius := min(arenaW, arenaH) / 5
	coverCount := 8

	for i := 0; i < coverCount; i++ {
		angle := float64(i) * 6.283185 / float64(coverCount)          // 2*pi / count
		cx := centerX + int(float64(coverRadius)*g.rng.Float64()*0.8) // Randomize radius slightly
		cy := centerY + int(float64(coverRadius)*g.rng.Float64()*0.8)

		// Use angle to position
		offsetX := int(float64(coverRadius) * 1.2 * cosApprox(angle))
		offsetY := int(float64(coverRadius) * 1.2 * sinApprox(angle))

		cx += offsetX
		cy += offsetY

		coverW := 2 + g.rng.Intn(2)
		coverH := 2 + g.rng.Intn(2)

		g.placeCover(tiles, cx-coverW/2, cy-coverH/2, coverW, coverH)
	}
}

// placeCover marks a cover area on the map.
func (g *ArenaGenerator) placeCover(tiles [][]int, x, y, w, h int) {
	room := Room{X: x, Y: y, W: w, H: h}
	g.CoverPoints = append(g.CoverPoints, room)

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < g.Width && py >= 0 && py < g.Height {
				// Don't overwrite spawn pads or weapon spawns
				if tiles[py][px] == g.floorTile {
					tiles[py][px] = TileCover
				}
			}
		}
	}
}

// analyzeSightlines computes sightline map for balancing.
func (g *ArenaGenerator) analyzeSightlines(tiles [][]int) {
	g.sightlineMap = make([][]int, g.Height)
	for y := range g.sightlineMap {
		g.sightlineMap[y] = make([]int, g.Width)
	}

	// For each floor tile, count visible tiles using raycasting
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			if tiles[y][x] == g.floorTile || tiles[y][x] == TileSpawnPad {
				visibleCount := g.countVisibleTiles(tiles, x, y, 20)
				g.sightlineMap[y][x] = visibleCount
			}
		}
	}
}

// countVisibleTiles counts how many tiles are visible from (x, y) within maxDist.
func (g *ArenaGenerator) countVisibleTiles(tiles [][]int, x, y, maxDist int) int {
	count := 0

	// Sample rays in 16 directions
	for angle := 0; angle < 16; angle++ {
		rad := float64(angle) * 6.283185 / 16.0
		dx := cosApprox(rad)
		dy := sinApprox(rad)

		for dist := 1; dist <= maxDist; dist++ {
			tx := x + int(dx*float64(dist))
			ty := y + int(dy*float64(dist))

			if tx < 0 || tx >= g.Width || ty < 0 || ty >= g.Height {
				break
			}

			if tiles[ty][tx] == g.wallTile || tiles[ty][tx] == TileCover {
				break
			}

			count++
		}
	}

	return count
}

// balanceSightlines adjusts cover placement to balance sightlines.
func (g *ArenaGenerator) balanceSightlines(tiles [][]int) {
	// Find spawn pads with excessive sightlines
	threshold := 150 // Max visible tiles from spawn

	for _, spawn := range g.SpawnPads {
		centerX := spawn.X + spawn.W/2
		centerY := spawn.Y + spawn.H/2

		if centerX < 0 || centerX >= g.Width || centerY < 0 || centerY >= g.Height {
			continue
		}

		if g.sightlineMap[centerY][centerX] > threshold {
			// Add blocking cover nearby
			offsetX := 3 + g.rng.Intn(2)
			offsetY := 3 + g.rng.Intn(2)

			// Place cover in random direction
			if g.rng.Intn(2) == 0 {
				offsetX = -offsetX
			}
			if g.rng.Intn(2) == 0 {
				offsetY = -offsetY
			}

			coverX := centerX + offsetX
			coverY := centerY + offsetY

			if coverX >= 0 && coverX < g.Width && coverY >= 0 && coverY < g.Height {
				if tiles[coverY][coverX] == g.floorTile {
					tiles[coverY][coverX] = TileCover
				}
			}
		}
	}
}

// cosApprox approximates cosine using Taylor series (good enough for map gen).
func cosApprox(x float64) float64 {
	// Normalize to [0, 2*pi)
	for x < 0 {
		x += 6.283185
	}
	for x >= 6.283185 {
		x -= 6.283185
	}

	// Taylor series: cos(x) ≈ 1 - x²/2 + x⁴/24 - x⁶/720
	x2 := x * x
	return 1.0 - x2/2.0 + x2*x2/24.0 - x2*x2*x2/720.0
}

// sinApprox approximates sine using Taylor series.
func sinApprox(x float64) float64 {
	// Normalize to [0, 2*pi)
	for x < 0 {
		x += 6.283185
	}
	for x >= 6.283185 {
		x -= 6.283185
	}

	// Taylor series: sin(x) ≈ x - x³/6 + x⁵/120 - x⁷/5040
	x2 := x * x
	return x - x*x2/6.0 + x*x2*x2/120.0 - x*x2*x2*x2/5040.0
}
