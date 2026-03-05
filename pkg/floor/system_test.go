package floor

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}

	if sys.genre != "fantasy" {
		t.Errorf("genre = %s, want fantasy", sys.genre)
	}

	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}

	if sys.detailCache == nil {
		t.Error("detailCache not initialized")
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 32)
	sys.SetGenre("scifi")

	if sys.genre != "scifi" {
		t.Errorf("genre = %s, want scifi after SetGenre", sys.genre)
	}
}

func TestGenerateFloorDetails(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}

	details := sys.GenerateFloorDetails(tiles, 12345)

	if details == nil {
		t.Fatal("GenerateFloorDetails returned nil")
	}

	// Should have generated some details for floor tiles
	// With density 0.15 for fantasy, expect some but not all floor tiles
	floorTileCount := 0
	for y := range tiles {
		for x := range tiles[y] {
			if tiles[y][x] >= 2 && tiles[y][x] < 10 {
				floorTileCount++
			}
		}
	}

	t.Logf("Generated %d details for %d floor tiles", len(details), floorTileCount)

	// Verify details are on floor tiles only
	for _, detail := range details {
		x, y := detail.X, detail.Y
		if y >= len(tiles) || x >= len(tiles[0]) {
			t.Errorf("Detail at (%d, %d) is out of bounds", x, y)
			continue
		}

		tile := tiles[y][x]
		if tile < 2 || tile >= 10 {
			t.Errorf("Detail at (%d, %d) is on non-floor tile %d", x, y, tile)
		}
	}
}

func TestGenerateFloorDetailsGenreVariety(t *testing.T) {
	tiles := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 2, 2, 2, 2, 2, 1},
		{1, 2, 2, 2, 2, 2, 1},
		{1, 2, 2, 2, 2, 2, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 32)
			details := sys.GenerateFloorDetails(tiles, 54321)

			if details == nil {
				t.Fatal("GenerateFloorDetails returned nil")
			}

			t.Logf("%s: generated %d details", genre, len(details))

			// Verify all details have correct genre
			for _, detail := range details {
				if detail.GenreID != genre {
					t.Errorf("Detail has genre %s, want %s", detail.GenreID, genre)
				}
			}
		})
	}
}

func TestRenderDetail(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	detailTypes := []DetailType{
		DetailCrack,
		DetailStain,
		DetailDebris,
		DetailScorch,
		DetailWear,
		DetailGraffiti,
		DetailBlood,
		DetailRust,
		DetailCorrode,
	}

	for _, dtype := range detailTypes {
		t.Run(dtype.String(), func(t *testing.T) {
			detail := &FloorDetailComponent{
				X:          5,
				Y:          5,
				DetailType: dtype,
				Intensity:  0.7,
				Seed:       12345,
				GenreID:    "fantasy",
			}

			img := sys.RenderDetail(detail)

			if img == nil {
				t.Fatal("RenderDetail returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != 32 || bounds.Dy() != 32 {
				t.Errorf("sprite size = %dx%d, want 32x32", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestRenderDetailCaching(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	detail := &FloorDetailComponent{
		X:          5,
		Y:          5,
		DetailType: DetailCrack,
		Intensity:  0.7,
		Seed:       12345,
		GenreID:    "fantasy",
	}

	img1 := sys.RenderDetail(detail)
	img2 := sys.RenderDetail(detail)

	if img1 != img2 {
		t.Error("Second call to RenderDetail should return cached image")
	}

	// Different detail should generate different image
	detail2 := &FloorDetailComponent{
		X:          6,
		Y:          6,
		DetailType: DetailStain,
		Intensity:  0.5,
		Seed:       54321,
		GenreID:    "fantasy",
	}

	img3 := sys.RenderDetail(detail2)
	if img3 == img1 {
		t.Error("Different detail should generate different image")
	}
}

func TestGetDetailDensity(t *testing.T) {
	tests := []struct {
		genre   string
		wantMin float64
		wantMax float64
	}{
		{"fantasy", 0.1, 0.2},
		{"scifi", 0.15, 0.25},
		{"horror", 0.2, 0.3},
		{"cyberpunk", 0.17, 0.27},
		{"postapoc", 0.25, 0.35},
		{"unknown", 0.1, 0.2}, // Should fallback to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewSystem(tt.genre, 32)
			density := sys.getDetailDensity()

			if density < tt.wantMin || density > tt.wantMax {
				t.Errorf("density = %v, want between %v and %v", density, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGetGenreDetailTypes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 32)
			types := sys.getGenreDetailTypes()

			if len(types) == 0 {
				t.Error("getGenreDetailTypes returned empty slice")
			}

			// Verify no duplicate types
			seen := make(map[DetailType]bool)
			for _, dtype := range types {
				if seen[dtype] {
					t.Errorf("Duplicate detail type: %v", dtype)
				}
				seen[dtype] = true
			}
		})
	}
}

func TestIsFloorTile(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	tests := []struct {
		tile int
		want bool
	}{
		{0, false},  // Empty
		{1, false},  // Wall
		{2, true},   // Floor
		{5, true},   // Floor variant
		{9, true},   // Floor variant
		{10, false}, // Non-floor
		{20, false}, // Non-floor
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := sys.isFloorTile(tt.tile)
			if got != tt.want {
				t.Errorf("isFloorTile(%d) = %v, want %v", tt.tile, got, tt.want)
			}
		})
	}
}

func TestIsNearWall(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		{"center", 2, 2, false},
		{"near_wall", 1, 1, true},
		{"edge", 3, 1, true},
		{"corner", 1, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.isNearWall(tt.x, tt.y, tiles)
			if got != tt.want {
				t.Errorf("isNearWall(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestDetailTypeString(t *testing.T) {
	// Ensure DetailType can be used in string formatting
	dtype := DetailCrack
	s := dtype.String()

	// String() should return something meaningful
	if s == "" {
		t.Error("DetailType.String() returned empty string")
	}
}

func (d DetailType) String() string {
	names := []string{
		"None",
		"Crack",
		"Stain",
		"Debris",
		"Scorch",
		"Wear",
		"Graffiti",
		"Blood",
		"Rust",
		"Corrode",
	}
	if d >= 0 && int(d) < len(names) {
		return names[d]
	}
	return "Unknown"
}

func TestFloorDetailComponent(t *testing.T) {
	comp := &FloorDetailComponent{
		X:          10,
		Y:          20,
		DetailType: DetailCrack,
		Intensity:  0.5,
		Seed:       12345,
		GenreID:    "fantasy",
	}

	if comp.Type() != "floor_detail" {
		t.Errorf("Type() = %s, want floor_detail", comp.Type())
	}
}

func TestEmptyTiles(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	// Empty tiles
	details := sys.GenerateFloorDetails([][]int{}, 12345)
	if details != nil {
		t.Error("GenerateFloorDetails on empty tiles should return nil")
	}

	// Single empty row
	details = sys.GenerateFloorDetails([][]int{{}}, 12345)
	if details != nil {
		t.Error("GenerateFloorDetails on tiles with empty row should return nil")
	}
}

func BenchmarkGenerateFloorDetails(b *testing.B) {
	sys := NewSystem("fantasy", 32)

	tiles := make([][]int, 50)
	for y := range tiles {
		tiles[y] = make([]int, 50)
		for x := range tiles[y] {
			if x == 0 || y == 0 || x == 49 || y == 49 {
				tiles[y][x] = 1 // Wall
			} else {
				tiles[y][x] = 2 // Floor
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GenerateFloorDetails(tiles, int64(i))
	}
}

func BenchmarkRenderDetail(b *testing.B) {
	sys := NewSystem("fantasy", 32)

	detail := &FloorDetailComponent{
		X:          5,
		Y:          5,
		DetailType: DetailCrack,
		Intensity:  0.7,
		Seed:       12345,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.RenderDetail(detail)
	}
}

func BenchmarkRenderDetailCached(b *testing.B) {
	sys := NewSystem("fantasy", 32)

	detail := &FloorDetailComponent{
		X:          5,
		Y:          5,
		DetailType: DetailCrack,
		Intensity:  0.7,
		Seed:       12345,
		GenreID:    "fantasy",
	}

	// Warm up cache
	sys.RenderDetail(detail)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.RenderDetail(detail)
	}
}

// Tests for floor tile texture generation

func TestGenerateFloorTiles(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}

	floorTiles := sys.GenerateFloorTiles(tiles, 12345)

	if floorTiles == nil {
		t.Fatal("GenerateFloorTiles returned nil")
	}

	// Should have generated one tile per floor tile
	expectedCount := 9 // 3x3 floor area
	if len(floorTiles) != expectedCount {
		t.Errorf("Generated %d floor tiles, want %d", len(floorTiles), expectedCount)
	}

	// Verify each tile has valid data
	for _, tile := range floorTiles {
		if tile.GenreID != "fantasy" {
			t.Errorf("Tile has genre %s, want fantasy", tile.GenreID)
		}
		if tile.Material < 0 || tile.Material > MaterialCrystal {
			t.Errorf("Invalid material type: %v", tile.Material)
		}
		if tile.Variant < 0 || tile.Variant >= 4 {
			t.Errorf("Invalid variant: %d", tile.Variant)
		}
	}
}

func TestGenerateFloorTilesGenres(t *testing.T) {
	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 32)
			floorTiles := sys.GenerateFloorTiles(tiles, 54321)

			if len(floorTiles) == 0 {
				t.Fatal("No floor tiles generated")
			}

			// Count materials to verify genre-appropriate distribution
			materialCounts := make(map[MaterialType]int)
			for _, tile := range floorTiles {
				materialCounts[tile.Material]++
			}

			t.Logf("%s material distribution: %v", genre, materialCounts)

			// Verify at least one material type is used
			if len(materialCounts) == 0 {
				t.Error("No materials used")
			}
		})
	}
}

func TestRenderTile(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	materials := []MaterialType{
		MaterialStone,
		MaterialMetal,
		MaterialWood,
		MaterialConcrete,
		MaterialTile,
		MaterialDirt,
		MaterialGrass,
		MaterialFlesh,
		MaterialCrystal,
	}

	for _, material := range materials {
		t.Run(material.String(), func(t *testing.T) {
			tile := &FloorTileComponent{
				X:        5,
				Y:        5,
				Material: material,
				Variant:  0,
				Seed:     12345,
				GenreID:  "fantasy",
			}

			img := sys.RenderTile(tile)

			if img == nil {
				t.Fatal("RenderTile returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != 32 || bounds.Dy() != 32 {
				t.Errorf("tile size = %dx%d, want 32x32", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestRenderTileCaching(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	tile := &FloorTileComponent{
		X:        5,
		Y:        5,
		Material: MaterialStone,
		Variant:  0,
		Seed:     12345,
		GenreID:  "fantasy",
	}

	img1 := sys.RenderTile(tile)
	img2 := sys.RenderTile(tile)

	// Second call should return cached texture from component
	if img1 != img2 {
		t.Error("Second call to RenderTile should return cached texture")
	}
}

func TestTextureGeneratorCaching(t *testing.T) {
	gen := NewTextureGenerator(100, "fantasy")

	img1 := gen.GetTile(MaterialStone, 0, 12345, 32)
	img2 := gen.GetTile(MaterialStone, 0, 12345, 32)

	if img1 != img2 {
		t.Error("Generator should return cached tile for same parameters")
	}

	// Different seed should generate different tile
	img3 := gen.GetTile(MaterialStone, 0, 54321, 32)
	if img3 == img1 {
		t.Error("Different seed should generate different tile")
	}
}

func TestTextureGeneratorLRU(t *testing.T) {
	gen := NewTextureGenerator(3, "fantasy") // Small cache for testing

	// Fill cache
	img1 := gen.GetTile(MaterialStone, 0, 1, 32)
	_ = gen.GetTile(MaterialMetal, 0, 2, 32) // Will be evicted
	img3 := gen.GetTile(MaterialWood, 0, 3, 32)

	// Access img1 to move it to front
	gen.GetTile(MaterialStone, 0, 1, 32)

	// Add new item, should evict img2 (least recently used)
	gen.GetTile(MaterialConcrete, 0, 4, 32)

	// img1 and img3 should still be cached
	cached1 := gen.GetTile(MaterialStone, 0, 1, 32)
	if cached1 != img1 {
		t.Error("img1 should still be cached")
	}

	cached3 := gen.GetTile(MaterialWood, 0, 3, 32)
	if cached3 != img3 {
		t.Error("img3 should still be cached")
	}
}

func TestFloorTileComponent(t *testing.T) {
	comp := &FloorTileComponent{
		X:        10,
		Y:        20,
		Material: MaterialStone,
		Variant:  2,
		Seed:     12345,
		GenreID:  "scifi",
	}

	if comp.Type() != "floor_tile" {
		t.Errorf("Type() = %s, want floor_tile", comp.Type())
	}
}

func TestMaterialDistribution(t *testing.T) {
	sys := NewSystem("fantasy", 32)

	tiles := make([][]int, 20)
	for y := range tiles {
		tiles[y] = make([]int, 20)
		for x := range tiles[y] {
			tiles[y][x] = 2 // All floor
		}
	}

	floorTiles := sys.GenerateFloorTiles(tiles, 12345)

	// Count materials
	materialCounts := make(map[MaterialType]int)
	for _, tile := range floorTiles {
		materialCounts[tile.Material]++
	}

	t.Logf("Material distribution: %v", materialCounts)

	// Fantasy should favor stone
	if materialCounts[MaterialStone] == 0 {
		t.Error("Fantasy genre should include stone tiles")
	}

	// Should have some variety
	if len(materialCounts) < 2 {
		t.Error("Expected multiple material types")
	}
}

func (m MaterialType) String() string {
	names := []string{
		"Stone",
		"Metal",
		"Wood",
		"Concrete",
		"Tile",
		"Dirt",
		"Grass",
		"Flesh",
		"Crystal",
	}
	if m >= 0 && int(m) < len(names) {
		return names[m]
	}
	return "Unknown"
}

func BenchmarkGenerateFloorTiles(b *testing.B) {
	sys := NewSystem("fantasy", 32)

	tiles := make([][]int, 50)
	for y := range tiles {
		tiles[y] = make([]int, 50)
		for x := range tiles[y] {
			if x == 0 || y == 0 || x == 49 || y == 49 {
				tiles[y][x] = 1 // Wall
			} else {
				tiles[y][x] = 2 // Floor
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GenerateFloorTiles(tiles, int64(i))
	}
}

func BenchmarkRenderTile(b *testing.B) {
	sys := NewSystem("fantasy", 32)

	tile := &FloorTileComponent{
		X:        5,
		Y:        5,
		Material: MaterialStone,
		Variant:  0,
		Seed:     12345,
		GenreID:  "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.RenderTile(tile)
	}
}

func BenchmarkTextureGeneration(b *testing.B) {
	gen := NewTextureGenerator(100, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GetTile(MaterialStone, i%4, int64(i), 32)
	}
}
