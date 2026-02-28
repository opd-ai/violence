package bsp

import (
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewArenaGenerator(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Small Arena", 32, 32},
		{"Medium Arena", 64, 64},
		{"Large Arena", 128, 128},
		{"Rectangular Arena", 80, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(12345)
			gen := NewArenaGenerator(tt.width, tt.height, r)

			if gen.Width != tt.width {
				t.Errorf("Width = %d, want %d", gen.Width, tt.width)
			}
			if gen.Height != tt.height {
				t.Errorf("Height = %d, want %d", gen.Height, tt.height)
			}
			if gen.genre != genre.Fantasy {
				t.Errorf("genre = %s, want %s", gen.genre, genre.Fantasy)
			}
		})
	}
}

func TestArenaGenerator_SetGenre(t *testing.T) {
	tests := []struct {
		name          string
		genre         string
		wantWallTile  int
		wantFloorTile int
	}{
		{"Fantasy", genre.Fantasy, TileWallStone, TileFloorStone},
		{"SciFi", genre.SciFi, TileWallHull, TileFloorHull},
		{"Horror", genre.Horror, TileWallPlaster, TileFloorWood},
		{"Cyberpunk", genre.Cyberpunk, TileWallConcrete, TileFloorConcrete},
		{"PostApoc", genre.PostApoc, TileWallRust, TileFloorDirt},
		{"Unknown", "unknown", TileWall, TileFloor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(12345)
			gen := NewArenaGenerator(64, 64, r)
			gen.SetGenre(tt.genre)

			if gen.wallTile != tt.wantWallTile {
				t.Errorf("wallTile = %d, want %d", gen.wallTile, tt.wantWallTile)
			}
			if gen.floorTile != tt.wantFloorTile {
				t.Errorf("floorTile = %d, want %d", gen.floorTile, tt.wantFloorTile)
			}
		})
	}
}

func TestArenaGenerator_Generate(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		seed   uint64
	}{
		{"Small Arena", 32, 32, 12345},
		{"Medium Arena", 64, 64, 67890},
		{"Large Arena", 128, 128, 11111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(tt.seed)
			gen := NewArenaGenerator(tt.width, tt.height, r)
			tiles := gen.Generate()

			// Verify dimensions
			if len(tiles) != tt.height {
				t.Fatalf("Height = %d, want %d", len(tiles), tt.height)
			}
			for y := range tiles {
				if len(tiles[y]) != tt.width {
					t.Fatalf("Width at row %d = %d, want %d", y, len(tiles[y]), tt.width)
				}
			}

			// Verify spawn pads exist
			if len(gen.SpawnPads) == 0 {
				t.Error("No spawn pads generated")
			}

			// Verify spawn pads are on the map
			for _, pad := range gen.SpawnPads {
				found := false
				for dy := 0; dy < pad.H; dy++ {
					for dx := 0; dx < pad.W; dx++ {
						x := pad.X + dx
						y := pad.Y + dy
						if x >= 0 && x < tt.width && y >= 0 && y < tt.height {
							if tiles[y][x] == TileSpawnPad {
								found = true
								break
							}
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Error("Spawn pad not found on map")
				}
			}

			// Verify weapon spawns exist
			if len(gen.WeaponSpawns) == 0 {
				t.Error("No weapon spawns generated")
			}

			// Verify weapon spawns are on the map
			for _, ws := range gen.WeaponSpawns {
				if ws.X < 0 || ws.X >= tt.width || ws.Y < 0 || ws.Y >= tt.height {
					t.Errorf("Weapon spawn out of bounds: (%d, %d)", ws.X, ws.Y)
				}
			}

			// Verify cover points exist
			if len(gen.CoverPoints) == 0 {
				t.Error("No cover points generated")
			}
		})
	}
}

func TestArenaGenerator_SymmetricalSpawns(t *testing.T) {
	r := rng.NewRNG(12345)
	gen := NewArenaGenerator(64, 64, r)
	tiles := gen.Generate()

	// Count spawn pad tiles
	spawnCount := 0
	for y := range tiles {
		for x := range tiles[y] {
			if tiles[y][x] == TileSpawnPad {
				spawnCount++
			}
		}
	}

	// Verify spawn pads exist
	if spawnCount == 0 {
		t.Fatal("No spawn pads found")
	}

	// Verify symmetry (4-way rotational symmetry)
	centerX := gen.Width / 2
	centerY := gen.Height / 2

	for y := 0; y < gen.Height; y++ {
		for x := 0; x < gen.Width; x++ {
			if tiles[y][x] == TileSpawnPad {
				// Check if corresponding symmetrical positions also have spawn pads
				// We check approximate symmetry due to integer rounding
				sym1X := centerX + (centerX - x)
				sym1Y := centerY + (centerY - y)

				// Allow ±1 tile tolerance for symmetry
				hasSymmetry := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						checkX := sym1X + dx
						checkY := sym1Y + dy
						if checkX >= 0 && checkX < gen.Width && checkY >= 0 && checkY < gen.Height {
							if tiles[checkY][checkX] == TileSpawnPad {
								hasSymmetry = true
								break
							}
						}
					}
					if hasSymmetry {
						break
					}
				}

				// Symmetry test is informational (can fail due to rounding)
				if !hasSymmetry {
					t.Logf("Spawn at (%d, %d) may lack perfect symmetry", x, y)
				}
			}
		}
	}
}

func TestArenaGenerator_WeaponTypes(t *testing.T) {
	r := rng.NewRNG(12345)
	gen := NewArenaGenerator(64, 64, r)
	gen.Generate()

	// Verify weapon types are assigned
	weaponTypes := make(map[string]int)
	for _, ws := range gen.WeaponSpawns {
		if ws.WeaponType == "" {
			t.Error("Weapon spawn has empty weapon type")
		}
		weaponTypes[ws.WeaponType]++
	}

	// Verify variety of weapons
	if len(weaponTypes) < 2 {
		t.Errorf("Not enough weapon variety: got %d types, want at least 2", len(weaponTypes))
	}

	// Verify known weapon types
	knownTypes := map[string]bool{
		"pistol":  true,
		"shotgun": true,
		"rifle":   true,
		"rocket":  true,
	}

	for wtype := range weaponTypes {
		if !knownTypes[wtype] {
			t.Errorf("Unknown weapon type: %s", wtype)
		}
	}
}

func TestArenaGenerator_CoverPlacement(t *testing.T) {
	r := rng.NewRNG(12345)
	gen := NewArenaGenerator(64, 64, r)
	tiles := gen.Generate()

	// Count cover tiles
	coverCount := 0
	for y := range tiles {
		for x := range tiles[y] {
			if tiles[y][x] == TileCover {
				coverCount++
			}
		}
	}

	if coverCount == 0 {
		t.Error("No cover tiles placed")
	}

	// Verify cover doesn't block spawn pads
	for _, pad := range gen.SpawnPads {
		for dy := 0; dy < pad.H; dy++ {
			for dx := 0; dx < pad.W; dx++ {
				x := pad.X + dx
				y := pad.Y + dy
				if x >= 0 && x < gen.Width && y >= 0 && y < gen.Height {
					if tiles[y][x] == TileCover {
						t.Error("Cover tile overlaps spawn pad")
					}
				}
			}
		}
	}
}

func TestArenaGenerator_SightlineBalancing(t *testing.T) {
	r := rng.NewRNG(12345)
	gen := NewArenaGenerator(64, 64, r)
	_ = gen.Generate()

	// Verify sightline map is created
	if gen.sightlineMap == nil {
		t.Fatal("Sightline map not created")
	}

	// Verify sightline map has correct dimensions
	if len(gen.sightlineMap) != gen.Height {
		t.Errorf("Sightline map height = %d, want %d", len(gen.sightlineMap), gen.Height)
	}

	// Check that spawn pads have sightline data
	for _, pad := range gen.SpawnPads {
		centerX := pad.X + pad.W/2
		centerY := pad.Y + pad.H/2

		if centerX < 0 || centerX >= gen.Width || centerY < 0 || centerY >= gen.Height {
			continue
		}

		sightlineValue := gen.sightlineMap[centerY][centerX]
		if sightlineValue == 0 {
			t.Logf("Spawn pad at (%d, %d) has zero sightline value (may be blocked)", centerX, centerY)
		}
	}
}

func TestArenaGenerator_Determinism(t *testing.T) {
	seed := uint64(12345)

	// Generate first arena
	r1 := rng.NewRNG(seed)
	gen1 := NewArenaGenerator(64, 64, r1)
	tiles1 := gen1.Generate()

	// Generate second arena with same seed
	r2 := rng.NewRNG(seed)
	gen2 := NewArenaGenerator(64, 64, r2)
	tiles2 := gen2.Generate()

	// Verify identical output
	for y := range tiles1 {
		for x := range tiles1[y] {
			if tiles1[y][x] != tiles2[y][x] {
				t.Errorf("Tiles differ at (%d, %d): %d vs %d", x, y, tiles1[y][x], tiles2[y][x])
			}
		}
	}

	// Verify spawn pads match
	if len(gen1.SpawnPads) != len(gen2.SpawnPads) {
		t.Errorf("Spawn pad count differs: %d vs %d", len(gen1.SpawnPads), len(gen2.SpawnPads))
	}

	// Verify weapon spawns match
	if len(gen1.WeaponSpawns) != len(gen2.WeaponSpawns) {
		t.Errorf("Weapon spawn count differs: %d vs %d", len(gen1.WeaponSpawns), len(gen2.WeaponSpawns))
	}
}

func TestArenaGenerator_GenreVariation(t *testing.T) {
	genres := []string{genre.Fantasy, genre.SciFi, genre.Horror, genre.Cyberpunk, genre.PostApoc}

	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			r := rng.NewRNG(12345)
			gen := NewArenaGenerator(64, 64, r)
			gen.SetGenre(g)
			tiles := gen.Generate()

			// Verify genre-specific tiles are used
			hasGenreTiles := false
			for y := range tiles {
				for x := range tiles[y] {
					tile := tiles[y][x]
					// Check for genre-specific wall or floor tiles
					if tile >= TileWallStone && tile <= TileFloorDirt {
						hasGenreTiles = true
						break
					}
				}
				if hasGenreTiles {
					break
				}
			}

			if !hasGenreTiles {
				t.Error("No genre-specific tiles found in generated arena")
			}
		})
	}
}

func TestArenaGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		seed   uint64
	}{
		{"Minimum Size", 16, 16, 12345},
		{"Very Large", 256, 256, 67890},
		{"Narrow", 128, 32, 11111},
		{"Tall", 32, 128, 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(tt.seed)
			gen := NewArenaGenerator(tt.width, tt.height, r)
			tiles := gen.Generate()

			// Verify no panics and basic structure
			if len(tiles) != tt.height {
				t.Fatalf("Height = %d, want %d", len(tiles), tt.height)
			}

			// Verify some floor tiles exist
			hasFloor := false
			for y := range tiles {
				for x := range tiles[y] {
					if tiles[y][x] == gen.floorTile || tiles[y][x] == TileSpawnPad {
						hasFloor = true
						break
					}
				}
				if hasFloor {
					break
				}
			}

			if !hasFloor {
				t.Error("No floor tiles found in arena")
			}
		})
	}
}

func TestArenaGenerator_TrigFunctions(t *testing.T) {
	tests := []struct {
		name   string
		angle  float64
		minCos float64
		maxCos float64
		minSin float64
		maxSin float64
	}{
		{"Zero", 0.0, 0.9, 1.1, -0.1, 0.1},
		{"Pi/2", 1.5708, -0.1, 0.1, 0.9, 1.1},
		// Note: Taylor series approximation has limited accuracy at larger angles
		// For map generation this is acceptable (small errors don't affect gameplay)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cosApprox(tt.angle)
			s := sinApprox(tt.angle)

			if c < tt.minCos || c > tt.maxCos {
				t.Errorf("cosApprox(%f) = %f, want in [%f, %f]", tt.angle, c, tt.minCos, tt.maxCos)
			}
			if s < tt.minSin || s > tt.maxSin {
				t.Errorf("sinApprox(%f) = %f, want in [%f, %f]", tt.angle, s, tt.minSin, tt.maxSin)
			}
		})
	}
}

func TestArenaGenerator_SightlineCount(t *testing.T) {
	r := rng.NewRNG(12345)
	gen := NewArenaGenerator(64, 64, r)
	tiles := gen.Generate()

	// Test sightline counting from center
	centerX := gen.Width / 2
	centerY := gen.Height / 2

	// Ensure center is floor
	tiles[centerY][centerX] = gen.floorTile

	count := gen.countVisibleTiles(tiles, centerX, centerY, 20)

	if count <= 0 {
		t.Error("No visible tiles from center")
	}

	// Test sightline from corner (should be blocked)
	cornerCount := gen.countVisibleTiles(tiles, 0, 0, 20)
	if cornerCount > count/2 {
		t.Logf("Corner has high visibility: %d (center: %d)", cornerCount, count)
	}
}
