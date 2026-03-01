package raycaster

import (
	"math"
	"testing"
)

func TestNewRaycaster(t *testing.T) {
	tests := []struct {
		name   string
		fov    float64
		width  int
		height int
	}{
		{"standard FOV", 66.0, 320, 200},
		{"wide FOV", 90.0, 640, 480},
		{"narrow FOV", 45.0, 160, 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(tt.fov, tt.width, tt.height)
			if r.FOV != tt.fov {
				t.Errorf("FOV = %v, want %v", r.FOV, tt.fov)
			}
			if r.Width != tt.width {
				t.Errorf("Width = %v, want %v", r.Width, tt.width)
			}
			if r.Height != tt.height {
				t.Errorf("Height = %v, want %v", r.Height, tt.height)
			}
		})
	}
}

func TestRaycaster_SetMap(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	testMap := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}

	r.SetMap(testMap)

	if r.Map == nil {
		t.Error("Map not set")
	}
	if len(r.Map) != 3 {
		t.Errorf("Map height = %d, want 3", len(r.Map))
	}
	if len(r.Map[0]) != 3 {
		t.Errorf("Map width = %d, want 3", len(r.Map[0]))
	}
}

func TestRaycaster_CastRay_SingleWall(t *testing.T) {
	// Simple 5x5 room with walls on edges
	testMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	tests := []struct {
		name     string
		posX     float64
		posY     float64
		dirX     float64
		dirY     float64
		wantDist float64
		wantWall int
		wantSide int
		epsilon  float64
	}{
		{
			name:     "facing north",
			posX:     2.5,
			posY:     2.5,
			dirX:     0.0,
			dirY:     -1.0,
			wantDist: 1.5,
			wantWall: 1,
			wantSide: 1,
			epsilon:  0.01,
		},
		{
			name:     "facing south",
			posX:     2.5,
			posY:     2.5,
			dirX:     0.0,
			dirY:     1.0,
			wantDist: 1.5,
			wantWall: 1,
			wantSide: 1,
			epsilon:  0.01,
		},
		{
			name:     "facing east",
			posX:     2.5,
			posY:     2.5,
			dirX:     1.0,
			dirY:     0.0,
			wantDist: 1.5,
			wantWall: 1,
			wantSide: 0,
			epsilon:  0.01,
		},
		{
			name:     "facing west",
			posX:     2.5,
			posY:     2.5,
			dirX:     -1.0,
			dirY:     0.0,
			wantDist: 1.5,
			wantWall: 1,
			wantSide: 0,
			epsilon:  0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, 1, 1)
			r.SetMap(testMap)

			hit := r.castRay(tt.posX, tt.posY, tt.dirX, tt.dirY)

			if math.Abs(hit.Distance-tt.wantDist) > tt.epsilon {
				t.Errorf("Distance = %v, want %v (±%v)", hit.Distance, tt.wantDist, tt.epsilon)
			}
			if hit.WallType != tt.wantWall {
				t.Errorf("WallType = %v, want %v", hit.WallType, tt.wantWall)
			}
			if hit.Side != tt.wantSide {
				t.Errorf("Side = %v, want %v", hit.Side, tt.wantSide)
			}
		})
	}
}

func TestRaycaster_CastRay_Diagonal(t *testing.T) {
	testMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	r := NewRaycaster(66.0, 1, 1)
	r.SetMap(testMap)

	// 45-degree angle northeast
	dirX := math.Cos(math.Pi / 4)
	dirY := -math.Sin(math.Pi / 4)

	hit := r.castRay(2.5, 2.5, dirX, dirY)

	if hit.Distance <= 0 || hit.Distance > 3.0 {
		t.Errorf("Distance = %v, expected reasonable positive value < 3.0", hit.Distance)
	}
	if hit.WallType != 1 {
		t.Errorf("WallType = %v, want 1", hit.WallType)
	}
}

func TestRaycaster_CastRay_DifferentWallTypes(t *testing.T) {
	// Use only actual wall tile values (not floor=2 which is non-solid)
	testMap := [][]int{
		{10, 10, 10, 10, 10},
		{3, 0, 0, 0, 4},
		{3, 0, 0, 0, 4},
		{3, 0, 0, 0, 4},
		{5, 5, 5, 5, 5},
	}

	tests := []struct {
		name     string
		dirX     float64
		dirY     float64
		wantWall int
	}{
		{"north wall type 10", 0.0, -1.0, 10},
		{"south wall type 5", 0.0, 1.0, 5},
		{"east wall type 4", 1.0, 0.0, 4},
		{"west wall type 3", -1.0, 0.0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, 1, 1)
			r.SetMap(testMap)

			hit := r.castRay(2.5, 2.5, tt.dirX, tt.dirY)

			if hit.WallType != tt.wantWall {
				t.Errorf("WallType = %v, want %v", hit.WallType, tt.wantWall)
			}
		})
	}
}

func TestRaycaster_CastRay_EmptySpace(t *testing.T) {
	// Map with no walls (all zeros)
	testMap := [][]int{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	r := NewRaycaster(66.0, 1, 1)
	r.SetMap(testMap)

	hit := r.castRay(1.5, 1.5, 1.0, 0.0)

	// Should hit out-of-bounds after maxDepth iterations
	if hit.Distance < 1e10 {
		t.Errorf("Distance = %v, expected very large value (out of bounds)", hit.Distance)
	}
}

func TestRaycaster_CastRay_OutOfBounds(t *testing.T) {
	testMap := [][]int{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	r := NewRaycaster(66.0, 1, 1)
	r.SetMap(testMap)

	// Start at edge and look outward
	hit := r.castRay(0.1, 0.1, -1.0, -1.0)

	if hit.Distance < 1e10 {
		t.Errorf("Distance = %v, expected very large value (immediate out of bounds)", hit.Distance)
	}
}

func TestRaycaster_CastRays_FullFrame(t *testing.T) {
	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}

	tests := []struct {
		name  string
		width int
		posX  float64
		posY  float64
		dirX  float64
		dirY  float64
	}{
		{"320 columns facing north", 320, 3.5, 3.5, 0.0, -1.0},
		{"640 columns facing east", 640, 3.5, 3.5, 1.0, 0.0},
		{"160 columns facing south", 160, 3.5, 3.5, 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, tt.width, 200)
			r.SetMap(testMap)

			hits := r.CastRays(tt.posX, tt.posY, tt.dirX, tt.dirY)

			if len(hits) != tt.width {
				t.Errorf("len(hits) = %d, want %d", len(hits), tt.width)
			}

			// All rays should hit walls
			for i, hit := range hits {
				if hit.WallType == 0 {
					t.Errorf("Column %d: WallType = 0, expected wall hit", i)
				}
				if hit.Distance <= 0 || hit.Distance > 100 {
					t.Errorf("Column %d: Distance = %v, expected reasonable positive value", i, hit.Distance)
				}
			}

			// Center column should hit straight ahead
			centerHit := hits[tt.width/2]
			if math.Abs(centerHit.Distance-2.5) > 0.1 {
				t.Errorf("Center column distance = %v, want ~2.5", centerHit.Distance)
			}
		})
	}
}

func TestRaycaster_CastRays_SymmetricFOV(t *testing.T) {
	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}

	r := NewRaycaster(90.0, 320, 200)
	r.SetMap(testMap)

	hits := r.CastRays(3.5, 3.5, 0.0, -1.0)

	// With symmetric FOV and position, left and right edge columns
	// should have similar distances (accounting for perspective)
	leftDist := hits[0].Distance
	rightDist := hits[319].Distance

	if math.Abs(leftDist-rightDist) > 1.0 {
		t.Errorf("FOV asymmetry: left = %v, right = %v, diff = %v",
			leftDist, rightDist, math.Abs(leftDist-rightDist))
	}
}

func TestRaycaster_CastRay_ParallelToGrid(t *testing.T) {
	testMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	r := NewRaycaster(66.0, 1, 1)
	r.SetMap(testMap)

	// Ray parallel to Y-axis (dirY = 0)
	hit := r.castRay(2.5, 2.5, 1.0, 0.0)
	if hit.WallType == 0 {
		t.Error("Expected wall hit when ray parallel to Y-axis")
	}

	// Ray parallel to X-axis (dirX = 0)
	hit = r.castRay(2.5, 2.5, 0.0, 1.0)
	if hit.WallType == 0 {
		t.Error("Expected wall hit when ray parallel to X-axis")
	}
}

func TestRaycaster_CastRay_Corner(t *testing.T) {
	// Test ray hitting exactly at corner
	testMap := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}

	r := NewRaycaster(66.0, 1, 1)
	r.SetMap(testMap)

	// Ray aiming at corner (1,1) from center
	dirX := math.Cos(math.Pi / 4)
	dirY := math.Sin(math.Pi / 4)

	hit := r.castRay(1.5, 1.5, dirX, dirY)

	// Should hit a wall
	if hit.WallType == 0 {
		t.Error("Expected wall hit at corner")
	}
}

func TestRaycaster_HitCoordinates(t *testing.T) {
	testMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	r := NewRaycaster(66.0, 1, 1)
	r.SetMap(testMap)

	// Face north from center
	hit := r.castRay(2.5, 2.5, 0.0, -1.0)

	// Hit should be at Y=1.0 (top wall of cell [1][*])
	if math.Abs(hit.HitY-1.0) > 0.01 {
		t.Errorf("HitY = %v, want ~1.0", hit.HitY)
	}
	// Hit X should be at ray origin X
	if math.Abs(hit.HitX-2.5) > 0.01 {
		t.Errorf("HitX = %v, want ~2.5", hit.HitX)
	}
}

func BenchmarkRaycaster_CastRays_320x200(b *testing.B) {
	testMap := make([][]int, 32)
	for i := range testMap {
		testMap[i] = make([]int, 32)
		for j := range testMap[i] {
			if i == 0 || i == 31 || j == 0 || j == 31 {
				testMap[i][j] = 1
			}
		}
	}

	r := NewRaycaster(66.0, 320, 200)
	r.SetMap(testMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.CastRays(16.5, 16.5, 1.0, 0.0)
	}
}

func BenchmarkRaycaster_CastRay_Single(b *testing.B) {
	testMap := make([][]int, 32)
	for i := range testMap {
		testMap[i] = make([]int, 32)
		for j := range testMap[i] {
			if i == 0 || i == 31 || j == 0 || j == 31 {
				testMap[i][j] = 1
			}
		}
	}

	r := NewRaycaster(66.0, 320, 200)
	r.SetMap(testMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.castRay(16.5, 16.5, 1.0, 0.0)
	}
}

func TestRaycaster_CastFloorCeiling_BasicRow(t *testing.T) {
	tests := []struct {
		name   string
		row    int
		width  int
		height int
		posX   float64
		posY   float64
		dirX   float64
		dirY   float64
		pitch  float64
	}{
		{"floor row bottom", 150, 320, 200, 5.0, 5.0, 1.0, 0.0, 0.0},
		{"ceiling row top", 50, 320, 200, 5.0, 5.0, 1.0, 0.0, 0.0},
		{"floor at horizon", 100, 320, 200, 5.0, 5.0, 0.0, 1.0, 0.0},
		{"pitched down", 150, 320, 200, 5.0, 5.0, 1.0, 0.0, 0.2},
		{"pitched up", 50, 320, 200, 5.0, 5.0, 1.0, 0.0, -0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, tt.width, tt.height)

			pixels := r.CastFloorCeiling(tt.row, tt.posX, tt.posY, tt.dirX, tt.dirY, tt.pitch)

			if len(pixels) != tt.width {
				t.Errorf("len(pixels) = %d, want %d", len(pixels), tt.width)
			}

			// Verify all pixels have valid data
			for i, px := range pixels {
				if math.IsNaN(px.WorldX) || math.IsNaN(px.WorldY) {
					t.Errorf("Pixel %d has NaN coordinates", i)
				}
				if px.Distance < 0 {
					t.Errorf("Pixel %d has negative distance: %v", i, px.Distance)
				}
				// Check floor/ceiling classification
				expectedFloor := tt.row > tt.height/2
				if px.IsFloor != expectedFloor {
					t.Errorf("Pixel %d: IsFloor = %v, want %v", i, px.IsFloor, expectedFloor)
				}
			}
		})
	}
}

func TestRaycaster_CastFloorCeiling_Distance(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	// Test that distance increases as we move closer to horizon
	rowNearHorizon := 110
	rowFarFromHorizon := 190

	pixelsNear := r.CastFloorCeiling(rowNearHorizon, 5.0, 5.0, 1.0, 0.0, 0.0)
	pixelsFar := r.CastFloorCeiling(rowFarFromHorizon, 5.0, 5.0, 1.0, 0.0, 0.0)

	// Center pixels for comparison
	centerIdx := 160

	// Rows closer to horizon have larger distance (farther away)
	if pixelsNear[centerIdx].Distance <= pixelsFar[centerIdx].Distance {
		t.Errorf("Distance at row %d (%v) should be greater than row %d (%v)",
			rowNearHorizon, pixelsNear[centerIdx].Distance,
			rowFarFromHorizon, pixelsFar[centerIdx].Distance)
	}
}

func TestRaycaster_CastFloorCeiling_FloorVsCeiling(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	floorRow := 150  // Below horizon
	ceilingRow := 50 // Above horizon

	floorPixels := r.CastFloorCeiling(floorRow, 5.0, 5.0, 1.0, 0.0, 0.0)
	ceilingPixels := r.CastFloorCeiling(ceilingRow, 5.0, 5.0, 1.0, 0.0, 0.0)

	// All floor pixels should be marked as floor
	for i, px := range floorPixels {
		if !px.IsFloor {
			t.Errorf("Floor pixel %d marked as ceiling", i)
		}
	}

	// All ceiling pixels should be marked as ceiling
	for i, px := range ceilingPixels {
		if px.IsFloor {
			t.Errorf("Ceiling pixel %d marked as floor", i)
		}
	}
}

func TestRaycaster_CastFloorCeiling_PerspectiveCorrect(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	row := 150
	pixels := r.CastFloorCeiling(row, 5.0, 5.0, 1.0, 0.0, 0.0)

	// Center pixel should be directly in front
	centerIdx := 160
	centerPx := pixels[centerIdx]

	// With direction (1, 0), center should have X > posX
	if centerPx.WorldX <= 5.0 {
		t.Errorf("Center pixel WorldX = %v, expected > 5.0", centerPx.WorldX)
	}

	// Y should be close to posY for center pixel
	if math.Abs(centerPx.WorldY-5.0) > 1.0 {
		t.Errorf("Center pixel WorldY = %v, expected ~5.0", centerPx.WorldY)
	}
}

func TestRaycaster_CastFloorCeiling_EdgePixels(t *testing.T) {
	r := NewRaycaster(90.0, 320, 200) // Wide FOV

	row := 150
	pixels := r.CastFloorCeiling(row, 5.0, 5.0, 1.0, 0.0, 0.0)

	// Left and right edge pixels should differ in Y coordinate
	leftPx := pixels[0]
	rightPx := pixels[319]

	if math.Abs(leftPx.WorldY-rightPx.WorldY) < 0.1 {
		t.Errorf("Edge pixels have nearly identical Y: left=%v, right=%v",
			leftPx.WorldY, rightPx.WorldY)
	}
}

func TestRaycaster_CastFloorCeiling_WithPitch(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	row := 150
	pitchZero := r.CastFloorCeiling(row, 5.0, 5.0, 1.0, 0.0, 0.0)
	pitchDown := r.CastFloorCeiling(row, 5.0, 5.0, 1.0, 0.0, 0.3)
	pitchUp := r.CastFloorCeiling(row, 5.0, 5.0, 1.0, 0.0, -0.3)

	centerIdx := 160

	// Pitch down increases camera height, making floor appear farther
	if pitchDown[centerIdx].Distance <= pitchZero[centerIdx].Distance {
		t.Errorf("Pitch down distance (%v) should be greater than zero pitch (%v)",
			pitchDown[centerIdx].Distance, pitchZero[centerIdx].Distance)
	}

	// Pitch up decreases camera height, making floor appear closer
	if pitchUp[centerIdx].Distance >= pitchZero[centerIdx].Distance {
		t.Errorf("Pitch up distance (%v) should be less than zero pitch (%v)",
			pitchUp[centerIdx].Distance, pitchZero[centerIdx].Distance)
	}
}

func TestRaycaster_CastFloorCeiling_DifferentDirections(t *testing.T) {
	tests := []struct {
		name string
		dirX float64
		dirY float64
	}{
		{"north", 0.0, -1.0},
		{"south", 0.0, 1.0},
		{"east", 1.0, 0.0},
		{"west", -1.0, 0.0},
		{"northeast", math.Cos(math.Pi / 4), -math.Sin(math.Pi / 4)},
		{"southwest", -math.Cos(math.Pi / 4), math.Sin(math.Pi / 4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, 320, 200)
			pixels := r.CastFloorCeiling(150, 5.0, 5.0, tt.dirX, tt.dirY, 0.0)

			if len(pixels) != 320 {
				t.Errorf("len(pixels) = %d, want 320", len(pixels))
			}

			// All pixels should have valid coordinates
			for i, px := range pixels {
				if math.IsInf(px.Distance, 0) || math.IsNaN(px.Distance) {
					t.Errorf("Pixel %d has invalid distance", i)
				}
			}
		})
	}
}

func BenchmarkRaycaster_CastFloorCeiling(b *testing.B) {
	r := NewRaycaster(66.0, 320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.CastFloorCeiling(150, 5.0, 5.0, 1.0, 0.0, 0.0)
	}
}

func TestRaycaster_CastSprites_Empty(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(2.5, 2.5, 0.0, -1.0)

	// No sprites
	sprites := []Sprite{}
	hits := r.CastSprites(sprites, 2.5, 2.5, 0.0, -1.0, wallHits)

	if len(hits) != 0 {
		t.Errorf("len(hits) = %d, want 0", len(hits))
	}
}

func TestRaycaster_CastSprites_SingleVisible(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 5.5, 0.0, -1.0)

	// Sprite in front of camera
	sprites := []Sprite{
		{X: 3.5, Y: 3.5, Type: 1, Width: 1.0, Height: 1.0},
	}

	hits := r.CastSprites(sprites, 3.5, 5.5, 0.0, -1.0, wallHits)

	if len(hits) != 1 {
		t.Errorf("len(hits) = %d, want 1", len(hits))
	}

	if len(hits) > 0 {
		hit := hits[0]
		if hit.Type != 1 {
			t.Errorf("Type = %d, want 1", hit.Type)
		}
		if hit.Distance <= 0 {
			t.Errorf("Distance = %v, expected positive", hit.Distance)
		}
		if hit.ScreenWidth <= 0 || hit.ScreenHeight <= 0 {
			t.Errorf("Invalid screen dimensions: %dx%d", hit.ScreenWidth, hit.ScreenHeight)
		}
	}
}

func TestRaycaster_CastSprites_BehindCamera(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 3.5, 0.0, -1.0)

	// Sprite behind camera
	sprites := []Sprite{
		{X: 3.5, Y: 5.5, Type: 1, Width: 1.0, Height: 1.0},
	}

	hits := r.CastSprites(sprites, 3.5, 3.5, 0.0, -1.0, wallHits)

	if len(hits) != 0 {
		t.Errorf("len(hits) = %d, want 0 (sprite behind camera)", len(hits))
	}
}

func TestRaycaster_CastSprites_Occluded(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 5.5, 0.0, -1.0)

	// Sprite behind wall
	sprites := []Sprite{
		{X: 3.5, Y: 1.5, Type: 1, Width: 1.0, Height: 1.0},
	}

	hits := r.CastSprites(sprites, 3.5, 5.5, 0.0, -1.0, wallHits)

	if len(hits) != 0 {
		t.Errorf("len(hits) = %d, want 0 (sprite behind wall)", len(hits))
	}
}

func TestRaycaster_CastSprites_DepthSort(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 5.5, 0.0, -1.0)

	// Multiple sprites at different distances
	sprites := []Sprite{
		{X: 3.5, Y: 4.5, Type: 1, Width: 0.5, Height: 0.5}, // Close
		{X: 3.5, Y: 2.5, Type: 2, Width: 0.5, Height: 0.5}, // Far
		{X: 3.5, Y: 3.5, Type: 3, Width: 0.5, Height: 0.5}, // Middle
	}

	hits := r.CastSprites(sprites, 3.5, 5.5, 0.0, -1.0, wallHits)

	if len(hits) != 3 {
		t.Errorf("len(hits) = %d, want 3", len(hits))
	}

	// Should be sorted farthest to nearest (painter's algorithm)
	if len(hits) == 3 {
		if hits[0].Distance <= hits[1].Distance {
			t.Errorf("Sprites not depth-sorted: %v <= %v", hits[0].Distance, hits[1].Distance)
		}
		if hits[1].Distance <= hits[2].Distance {
			t.Errorf("Sprites not depth-sorted: %v <= %v", hits[1].Distance, hits[2].Distance)
		}
	}
}

func TestRaycaster_CastSprites_OffScreen(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 3.5, 0.0, -1.0)

	// Sprite far to the left (off-screen)
	sprites := []Sprite{
		{X: 1.0, Y: 2.5, Type: 1, Width: 0.5, Height: 0.5},
	}

	hits := r.CastSprites(sprites, 3.5, 3.5, 0.0, -1.0, wallHits)

	// May or may not be visible depending on FOV, but shouldn't crash
	if len(hits) < 0 {
		t.Error("Negative hit count")
	}
}

func TestRaycaster_CastSprites_DifferentSizes(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 5.5, 0.0, -1.0)

	tests := []struct {
		name   string
		width  float64
		height float64
	}{
		{"small sprite", 0.5, 0.5},
		{"medium sprite", 1.0, 1.0},
		{"large sprite", 2.0, 2.0},
		{"tall sprite", 0.5, 2.0},
		{"wide sprite", 2.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprites := []Sprite{
				{X: 3.5, Y: 3.5, Type: 1, Width: tt.width, Height: tt.height},
			}

			hits := r.CastSprites(sprites, 3.5, 5.5, 0.0, -1.0, wallHits)

			if len(hits) != 1 {
				t.Errorf("len(hits) = %d, want 1", len(hits))
			}

			if len(hits) > 0 {
				if hits[0].ScreenWidth <= 0 || hits[0].ScreenHeight <= 0 {
					t.Errorf("Invalid dimensions: %dx%d", hits[0].ScreenWidth, hits[0].ScreenHeight)
				}
			}
		})
	}
}

func TestRaycaster_CastSprites_Clipping(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	r.SetMap(testMap)

	wallHits := r.CastRays(3.5, 5.5, 0.0, -1.0)

	// Sprite visible in view
	sprites := []Sprite{
		{X: 3.5, Y: 3.5, Type: 1, Width: 1.0, Height: 1.0},
	}

	hits := r.CastSprites(sprites, 3.5, 5.5, 0.0, -1.0, wallHits)

	if len(hits) > 0 {
		hit := hits[0]

		// Verify clipping bounds are within screen
		if hit.DrawStartX < 0 {
			t.Errorf("DrawStartX = %d, should be >= 0", hit.DrawStartX)
		}
		if hit.DrawEndX >= r.Width {
			t.Errorf("DrawEndX = %d, should be < %d", hit.DrawEndX, r.Width)
		}
		if hit.DrawStartY < 0 {
			t.Errorf("DrawStartY = %d, should be >= 0", hit.DrawStartY)
		}
		if hit.DrawEndY >= r.Height {
			t.Errorf("DrawEndY = %d, should be < %d", hit.DrawEndY, r.Height)
		}
	}
}

func BenchmarkRaycaster_CastSprites(b *testing.B) {
	testMap := make([][]int, 32)
	for i := range testMap {
		testMap[i] = make([]int, 32)
		for j := range testMap[i] {
			if i == 0 || i == 31 || j == 0 || j == 31 {
				testMap[i][j] = 1
			}
		}
	}

	r := NewRaycaster(66.0, 320, 200)
	r.SetMap(testMap)

	wallHits := r.CastRays(16.5, 16.5, 1.0, 0.0)

	sprites := make([]Sprite, 20)
	for i := range sprites {
		sprites[i] = Sprite{
			X:      float64(10 + i),
			Y:      float64(10 + i%5),
			Type:   i % 4,
			Width:  1.0,
			Height: 1.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.CastSprites(sprites, 16.5, 16.5, 1.0, 0.0, wallHits)
	}
}

func TestRaycaster_ApplyFog_NoDistance(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.FogColor = [3]float64{0.5, 0.5, 0.5}
	r.FogDensity = 0.1

	baseColor := [3]float64{1.0, 0.0, 0.0} // Red
	result := r.ApplyFog(baseColor, 0.0)   // Zero distance

	// At zero distance, should be mostly base color
	if result[0] < 0.9 {
		t.Errorf("Red channel = %v, expected close to 1.0 at zero distance", result[0])
	}
}

func TestRaycaster_ApplyFog_LargeDistance(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.FogColor = [3]float64{0.5, 0.5, 0.5}
	r.FogDensity = 0.1

	baseColor := [3]float64{1.0, 0.0, 0.0} // Red
	result := r.ApplyFog(baseColor, 50.0)  // Large distance

	// At large distance, should be mostly fog color
	epsilon := 0.1
	if math.Abs(result[0]-r.FogColor[0]) > epsilon {
		t.Errorf("Red channel = %v, expected close to fog color %v at large distance",
			result[0], r.FogColor[0])
	}
}

func TestRaycaster_ApplyFog_Gradient(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.FogColor = [3]float64{0.0, 0.0, 0.0} // Black fog
	r.FogDensity = 0.1

	baseColor := [3]float64{1.0, 1.0, 1.0} // White

	// Test that fog increases with distance
	result1 := r.ApplyFog(baseColor, 1.0)
	result10 := r.ApplyFog(baseColor, 10.0)
	result20 := r.ApplyFog(baseColor, 20.0)

	// Each should be darker than the previous
	if result1[0] <= result10[0] {
		t.Errorf("Distance 1 (%v) should be brighter than distance 10 (%v)",
			result1[0], result10[0])
	}
	if result10[0] <= result20[0] {
		t.Errorf("Distance 10 (%v) should be brighter than distance 20 (%v)",
			result10[0], result20[0])
	}
}

func TestRaycaster_ApplyFog_ColorChannels(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.FogColor = [3]float64{0.2, 0.4, 0.6} // Blue-ish fog
	r.FogDensity = 0.1

	baseColor := [3]float64{1.0, 0.0, 0.0} // Red
	result := r.ApplyFog(baseColor, 10.0)

	// All channels should be in valid range
	for i, ch := range result {
		if ch < 0.0 || ch > 1.0 {
			t.Errorf("Channel %d = %v, out of range [0, 1]", i, ch)
		}
	}

	// Green and blue should increase due to fog
	if result[1] < baseColor[1] {
		t.Errorf("Green channel should increase from %v to at least %v due to fog",
			baseColor[1], r.FogColor[1])
	}
	if result[2] < baseColor[2] {
		t.Errorf("Blue channel should increase from %v to at least %v due to fog",
			baseColor[2], r.FogColor[2])
	}
}

func TestRaycaster_SetGenre_Fantasy(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.SetGenre("fantasy")

	if r.FogColor[0] <= 0.0 || r.FogColor[2] <= 0.0 {
		t.Error("Fantasy fog should have purple-ish tint")
	}
	if r.FogDensity <= 0.0 {
		t.Error("Fog density should be positive")
	}
}

func TestRaycaster_SetGenre_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			r := NewRaycaster(66.0, 320, 200)
			r.SetGenre(genre)

			// All color channels should be in valid range
			for i, ch := range r.FogColor {
				if ch < 0.0 || ch > 1.0 {
					t.Errorf("Genre %s: FogColor[%d] = %v, out of range [0, 1]",
						genre, i, ch)
				}
			}

			// Density should be positive
			if r.FogDensity <= 0.0 {
				t.Errorf("Genre %s: FogDensity = %v, should be positive", genre, r.FogDensity)
			}
		})
	}
}

func TestRaycaster_SetGenre_Distinctiveness(t *testing.T) {
	r1 := NewRaycaster(66.0, 320, 200)
	r1.SetGenre("fantasy")

	r2 := NewRaycaster(66.0, 320, 200)
	r2.SetGenre("scifi")

	// Different genres should have different fog
	if r1.FogColor == r2.FogColor {
		t.Error("Fantasy and scifi should have distinct fog colors")
	}
}

func BenchmarkRaycaster_ApplyFog(b *testing.B) {
	r := NewRaycaster(66.0, 320, 200)
	r.FogColor = [3]float64{0.1, 0.1, 0.2}
	r.FogDensity = 0.05

	baseColor := [3]float64{1.0, 0.8, 0.6}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ApplyFog(baseColor, 10.0)
	}
}

func TestRaycaster_CastRay_NilMap(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	// Map is nil by default

	hit := r.castRay(5.0, 5.0, 1.0, 0.0)

	// Should return safe default instead of panicking
	if hit.Distance < 1e10 {
		t.Errorf("Distance = %v, expected very large value for nil map", hit.Distance)
	}
	if hit.WallType != 1 {
		t.Errorf("WallType = %v, expected 1 for nil map boundary", hit.WallType)
	}
}

func TestRaycaster_CastRay_EmptyMap(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.SetMap([][]int{})

	hit := r.castRay(5.0, 5.0, 1.0, 0.0)

	// Should return safe default instead of panicking
	if hit.Distance < 1e10 {
		t.Errorf("Distance = %v, expected very large value for empty map", hit.Distance)
	}
	if hit.WallType != 1 {
		t.Errorf("WallType = %v, expected 1 for empty map boundary", hit.WallType)
	}
}

func TestRaycaster_CastRay_EmptyRow(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)
	r.SetMap([][]int{{}})

	hit := r.castRay(5.0, 5.0, 1.0, 0.0)

	// Should return safe default instead of panicking
	if hit.Distance < 1e10 {
		t.Errorf("Distance = %v, expected very large value for map with empty row", hit.Distance)
	}
	if hit.WallType != 1 {
		t.Errorf("WallType = %v, expected 1 for empty row boundary", hit.WallType)
	}
}

func TestRaycaster_CastFloorCeiling_HorizonLine(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	// Test row exactly at horizon (height/2 = 100)
	horizonRow := r.Height / 2
	pixels := r.CastFloorCeiling(horizonRow, 5.0, 5.0, 1.0, 0.0, 0.0)

	if len(pixels) != r.Width {
		t.Errorf("len(pixels) = %d, want %d", len(pixels), r.Width)
	}

	// All pixels at horizon should have infinite distance
	for i, px := range pixels {
		if px.Distance < 1e10 {
			t.Errorf("Pixel %d: Distance = %v, expected very large value at horizon", i, px.Distance)
		}
		if math.IsNaN(px.WorldX) || math.IsNaN(px.WorldY) {
			t.Errorf("Pixel %d has NaN coordinates at horizon", i)
		}
	}
}

func TestRaycaster_CastFloorCeiling_HorizonWithPitch(t *testing.T) {
	r := NewRaycaster(66.0, 320, 200)

	// Even with pitch, horizon line should handle division by zero
	horizonRow := r.Height / 2
	pixels := r.CastFloorCeiling(horizonRow, 5.0, 5.0, 1.0, 0.0, 0.5)

	for i, px := range pixels {
		if px.Distance < 1e10 {
			t.Errorf("Pixel %d: Distance = %v, expected very large value at horizon with pitch", i, px.Distance)
		}
		if math.IsInf(px.Distance, 0) && math.IsNaN(px.Distance) {
			t.Errorf("Pixel %d has invalid distance at horizon with pitch", i)
		}
	}
}

func TestRayHit_TextureX(t *testing.T) {
	tests := []struct {
		name     string
		testMap  [][]int
		posX     float64
		posY     float64
		dirX     float64
		dirY     float64
		wantSide int
		checkTex bool
	}{
		{
			name: "vertical wall hit",
			testMap: [][]int{
				{1, 1, 1, 1, 1},
				{1, 0, 0, 0, 1},
				{1, 0, 0, 0, 1},
				{1, 0, 0, 0, 1},
				{1, 1, 1, 1, 1},
			},
			posX:     2.5,
			posY:     2.5,
			dirX:     1.0,
			dirY:     0.0,
			wantSide: 0, // vertical
			checkTex: true,
		},
		{
			name: "horizontal wall hit",
			testMap: [][]int{
				{1, 1, 1, 1, 1},
				{1, 0, 0, 0, 1},
				{1, 0, 0, 0, 1},
				{1, 0, 0, 0, 1},
				{1, 1, 1, 1, 1},
			},
			posX:     2.5,
			posY:     2.5,
			dirX:     0.0,
			dirY:     1.0,
			wantSide: 1, // horizontal
			checkTex: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, 320, 200)
			r.SetMap(tt.testMap)

			hits := r.CastRays(tt.posX, tt.posY, tt.dirX, tt.dirY)
			if len(hits) == 0 {
				t.Fatal("No hits returned")
			}

			centerHit := hits[r.Width/2]

			// Verify TextureX is in valid range
			if centerHit.TextureX < 0.0 || centerHit.TextureX > 1.0 {
				t.Errorf("TextureX = %f, want in range [0.0, 1.0]", centerHit.TextureX)
			}

			// Verify side matches expectation
			if tt.checkTex && centerHit.Side != tt.wantSide {
				t.Errorf("Side = %d, want %d", centerHit.Side, tt.wantSide)
			}
		})
	}
}

func TestRayHit_TextureX_Consistency(t *testing.T) {
	// Test that adjacent rays have consistent texture coordinates
	testMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	r := NewRaycaster(66.0, 320, 200)
	r.SetMap(testMap)

	hits := r.CastRays(5.0, 2.5, 1.0, 0.0)

	// All rays hitting same wall should have valid TextureX
	for i, hit := range hits {
		if hit.TextureX < 0.0 || hit.TextureX > 1.0 {
			t.Errorf("Ray %d: TextureX = %f, want in range [0.0, 1.0]", i, hit.TextureX)
		}
	}
}

func TestRayHit_TextureX_DifferentPositions(t *testing.T) {
	testMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	tests := []struct {
		name string
		posX float64
		posY float64
		dirX float64
		dirY float64
	}{
		{"facing east", 2.5, 2.5, 1.0, 0.0},
		{"facing west", 2.5, 2.5, -1.0, 0.0},
		{"facing north", 2.5, 2.5, 0.0, -1.0},
		{"facing south", 2.5, 2.5, 0.0, 1.0},
		{"diagonal NE", 2.5, 2.5, 0.707, -0.707},
		{"diagonal SE", 2.5, 2.5, 0.707, 0.707},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRaycaster(66.0, 320, 200)
			r.SetMap(testMap)

			hits := r.CastRays(tt.posX, tt.posY, tt.dirX, tt.dirY)
			centerHit := hits[r.Width/2]

			if centerHit.TextureX < 0.0 || centerHit.TextureX > 1.0 {
				t.Errorf("TextureX = %f, want in range [0.0, 1.0]", centerHit.TextureX)
			}
		})
	}
}
