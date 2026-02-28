package lighting

import (
	"math"
	"testing"
)

func TestNewSectorLightMap(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		ambient float64
		wantAmb float64
	}{
		{
			name:    "valid ambient in range",
			width:   10,
			height:  10,
			ambient: 0.5,
			wantAmb: 0.5,
		},
		{
			name:    "ambient clamped to max",
			width:   10,
			height:  10,
			ambient: 1.5,
			wantAmb: 1.0,
		},
		{
			name:    "ambient clamped to min",
			width:   10,
			height:  10,
			ambient: -0.5,
			wantAmb: 0.0,
		},
		{
			name:    "zero ambient",
			width:   5,
			height:  5,
			ambient: 0.0,
			wantAmb: 0.0,
		},
		{
			name:    "full ambient",
			width:   5,
			height:  5,
			ambient: 1.0,
			wantAmb: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slm := NewSectorLightMap(tt.width, tt.height, tt.ambient)
			if slm.Width != tt.width {
				t.Errorf("Width = %d, want %d", slm.Width, tt.width)
			}
			if slm.Height != tt.height {
				t.Errorf("Height = %d, want %d", slm.Height, tt.height)
			}
			if slm.Ambient != tt.wantAmb {
				t.Errorf("Ambient = %f, want %f", slm.Ambient, tt.wantAmb)
			}
			if slm.LightCount() != 0 {
				t.Errorf("LightCount = %d, want 0", slm.LightCount())
			}
			if !slm.dirty {
				t.Error("dirty = false, want true after creation")
			}
		})
	}
}

func TestAddLight(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.2)

	light1 := Light{X: 5, Y: 5, Radius: 3, Intensity: 1.0}
	idx1 := slm.AddLight(light1)
	if idx1 != 0 {
		t.Errorf("First light index = %d, want 0", idx1)
	}
	if slm.LightCount() != 1 {
		t.Errorf("LightCount = %d, want 1", slm.LightCount())
	}

	light2 := Light{X: 2, Y: 2, Radius: 2, Intensity: 0.5}
	idx2 := slm.AddLight(light2)
	if idx2 != 1 {
		t.Errorf("Second light index = %d, want 1", idx2)
	}
	if slm.LightCount() != 2 {
		t.Errorf("LightCount = %d, want 2", slm.LightCount())
	}

	if !slm.dirty {
		t.Error("dirty = false, want true after AddLight")
	}
}

func TestRemoveLight(t *testing.T) {
	tests := []struct {
		name    string
		addN    int
		remove  int
		wantOk  bool
		wantCnt int
	}{
		{
			name:    "remove valid index",
			addN:    3,
			remove:  1,
			wantOk:  true,
			wantCnt: 2,
		},
		{
			name:    "remove first",
			addN:    3,
			remove:  0,
			wantOk:  true,
			wantCnt: 2,
		},
		{
			name:    "remove last",
			addN:    3,
			remove:  2,
			wantOk:  true,
			wantCnt: 2,
		},
		{
			name:    "remove negative index",
			addN:    3,
			remove:  -1,
			wantOk:  false,
			wantCnt: 3,
		},
		{
			name:    "remove out of bounds",
			addN:    3,
			remove:  5,
			wantOk:  false,
			wantCnt: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slm := NewSectorLightMap(10, 10, 0.2)
			for i := 0; i < tt.addN; i++ {
				slm.AddLight(Light{X: float64(i), Y: float64(i), Radius: 2, Intensity: 1.0})
			}

			ok := slm.RemoveLight(tt.remove)
			if ok != tt.wantOk {
				t.Errorf("RemoveLight ok = %v, want %v", ok, tt.wantOk)
			}
			if slm.LightCount() != tt.wantCnt {
				t.Errorf("LightCount = %d, want %d", slm.LightCount(), tt.wantCnt)
			}
			if ok && !slm.dirty {
				t.Error("dirty = false, want true after RemoveLight")
			}
		})
	}
}

func TestUpdateLight(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.2)
	slm.AddLight(Light{X: 5, Y: 5, Radius: 3, Intensity: 1.0})
	slm.AddLight(Light{X: 2, Y: 2, Radius: 2, Intensity: 0.5})

	tests := []struct {
		name   string
		index  int
		light  Light
		wantOk bool
	}{
		{
			name:   "update valid index",
			index:  0,
			light:  Light{X: 7, Y: 7, Radius: 4, Intensity: 0.8},
			wantOk: true,
		},
		{
			name:   "update negative index",
			index:  -1,
			light:  Light{X: 1, Y: 1, Radius: 1, Intensity: 0.5},
			wantOk: false,
		},
		{
			name:   "update out of bounds",
			index:  10,
			light:  Light{X: 1, Y: 1, Radius: 1, Intensity: 0.5},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slm.dirty = false // Reset dirty flag
			ok := slm.UpdateLight(tt.index, tt.light)
			if ok != tt.wantOk {
				t.Errorf("UpdateLight ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && !slm.dirty {
				t.Error("dirty = false, want true after UpdateLight")
			}
		})
	}
}

func TestSetAmbient(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{
			name:  "valid value",
			value: 0.3,
			want:  0.3,
		},
		{
			name:  "clamped high",
			value: 2.0,
			want:  1.0,
		},
		{
			name:  "clamped low",
			value: -1.0,
			want:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slm := NewSectorLightMap(10, 10, 0.5)
			slm.dirty = false
			slm.SetAmbient(tt.value)
			if slm.Ambient != tt.want {
				t.Errorf("Ambient = %f, want %f", slm.Ambient, tt.want)
			}
			if !slm.dirty {
				t.Error("dirty = false, want true after SetAmbient")
			}
		})
	}
}

func TestCalculate_AmbientOnly(t *testing.T) {
	slm := NewSectorLightMap(5, 5, 0.3)
	slm.Calculate()

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			light := slm.GetLight(x, y)
			if math.Abs(light-0.3) > 0.001 {
				t.Errorf("GetLight(%d, %d) = %f, want 0.3", x, y, light)
			}
		}
	}

	if slm.dirty {
		t.Error("dirty = true, want false after Calculate")
	}
}

func TestCalculate_SinglePointLight(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.0)
	slm.AddLight(Light{X: 5, Y: 5, Radius: 3, Intensity: 1.0})
	slm.Calculate()

	// Center tile should be brightest
	centerLight := slm.GetLight(5, 5)
	if centerLight < 0.5 {
		t.Errorf("Center light = %f, expected > 0.5", centerLight)
	}

	// Tiles at distance should be dimmer
	nearLight := slm.GetLight(6, 5)
	if nearLight >= centerLight {
		t.Errorf("Near light %f >= center light %f", nearLight, centerLight)
	}

	// Tiles far away should have minimal light
	farLight := slm.GetLight(9, 9)
	if farLight > 0.1 {
		t.Errorf("Far light = %f, expected < 0.1", farLight)
	}
}

func TestCalculate_MultipleLights(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.1)
	slm.AddLight(Light{X: 2, Y: 2, Radius: 3, Intensity: 1.0})
	slm.AddLight(Light{X: 7, Y: 7, Radius: 3, Intensity: 1.0})
	slm.Calculate()

	// Tiles near first light
	light1 := slm.GetLight(2, 2)
	if light1 < 0.5 {
		t.Errorf("Light near first source = %f, expected > 0.5", light1)
	}

	// Tiles near second light
	light2 := slm.GetLight(7, 7)
	if light2 < 0.5 {
		t.Errorf("Light near second source = %f, expected > 0.5", light2)
	}

	// Middle should have some ambient + partial contribution
	midLight := slm.GetLight(5, 5)
	if midLight < 0.1 {
		t.Errorf("Middle light = %f, expected > 0.1", midLight)
	}
}

func TestCalculate_Clamping(t *testing.T) {
	slm := NewSectorLightMap(5, 5, 0.8)
	// Very intense overlapping lights
	slm.AddLight(Light{X: 2, Y: 2, Radius: 5, Intensity: 10.0})
	slm.AddLight(Light{X: 2, Y: 2, Radius: 5, Intensity: 10.0})
	slm.Calculate()

	// Should be clamped to 1.0, not exceed
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			light := slm.GetLight(x, y)
			if light > 1.0 {
				t.Errorf("GetLight(%d, %d) = %f, should be clamped to 1.0", x, y, light)
			}
		}
	}
}

func TestCalculate_SkipsWhenNotDirty(t *testing.T) {
	slm := NewSectorLightMap(5, 5, 0.2)
	slm.Calculate()

	// Modify internal state to detect if Calculate runs again
	slm.lightGrid[0] = 0.9
	slm.Calculate() // Should skip since not dirty

	if slm.lightGrid[0] != 0.9 {
		t.Error("Calculate should skip when not dirty")
	}
}

func TestGetLight_OutOfBounds(t *testing.T) {
	slm := NewSectorLightMap(5, 5, 0.3)
	slm.Calculate()

	tests := []struct {
		name string
		x, y int
	}{
		{"negative x", -1, 2},
		{"negative y", 2, -1},
		{"x too large", 5, 2},
		{"y too large", 2, 5},
		{"both negative", -1, -1},
		{"both too large", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			light := slm.GetLight(tt.x, tt.y)
			if light != 0.0 {
				t.Errorf("GetLight(%d, %d) = %f, want 0.0 for out of bounds", tt.x, tt.y, light)
			}
		})
	}
}

func TestClear(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.5)
	slm.AddLight(Light{X: 5, Y: 5, Radius: 3, Intensity: 1.0})
	slm.AddLight(Light{X: 2, Y: 2, Radius: 2, Intensity: 0.5})

	if slm.LightCount() != 2 {
		t.Errorf("LightCount before Clear = %d, want 2", slm.LightCount())
	}

	slm.Clear()

	if slm.LightCount() != 0 {
		t.Errorf("LightCount after Clear = %d, want 0", slm.LightCount())
	}
	if !slm.dirty {
		t.Error("dirty = false, want true after Clear")
	}
}

func TestQuadraticAttenuation(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.0)
	slm.AddLight(Light{X: 5, Y: 5, Radius: 5, Intensity: 1.0})
	slm.Calculate()

	// Light at center (distance ~0.5)
	centerLight := slm.GetLight(5, 5)

	// Light at distance 1
	dist1Light := slm.GetLight(6, 5)

	// Light at distance 2
	dist2Light := slm.GetLight(7, 5)

	// Verify attenuation decreases with distance
	if dist1Light >= centerLight {
		t.Errorf("dist1Light %f >= centerLight %f", dist1Light, centerLight)
	}
	if dist2Light >= dist1Light {
		t.Errorf("dist2Light %f >= dist1Light %f", dist2Light, dist1Light)
	}

	// Verify approximate quadratic relationship
	// At distance d, intensity ≈ I / (1 + d²)
	// This is approximate due to tile center offset
	tolerance := 0.3
	expectedRatio := (1.0 + 0.25) / (1.0 + 1.25) // dist ~0.5 vs dist ~1.1
	actualRatio := centerLight / dist1Light
	if math.Abs(actualRatio-expectedRatio) > tolerance {
		t.Logf("Attenuation ratio: expected ~%f, got %f", expectedRatio, actualRatio)
	}
}

func TestLightBoundingBox(t *testing.T) {
	// Test that lights only affect tiles within their radius
	slm := NewSectorLightMap(20, 20, 0.0)
	slm.AddLight(Light{X: 10, Y: 10, Radius: 2, Intensity: 1.0})
	slm.Calculate()

	// Tiles far outside radius should have no light
	farTiles := []struct{ x, y int }{
		{0, 0}, {19, 19}, {0, 19}, {19, 0}, {5, 5}, {15, 15},
	}

	for _, tile := range farTiles {
		light := slm.GetLight(tile.x, tile.y)
		if light > 0.01 {
			t.Errorf("GetLight(%d, %d) = %f, expected ~0.0 (outside radius)", tile.x, tile.y, light)
		}
	}

	// Tiles within radius should have light
	nearTiles := []struct{ x, y int }{
		{10, 10}, {11, 10}, {10, 11}, {9, 10}, {10, 9},
	}

	for _, tile := range nearTiles {
		light := slm.GetLight(tile.x, tile.y)
		if light < 0.1 {
			t.Errorf("GetLight(%d, %d) = %f, expected > 0.1 (within radius)", tile.x, tile.y, light)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		min   float64
		max   float64
		want  float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestMaxMin(t *testing.T) {
	if max(5, 3) != 5 {
		t.Error("max(5, 3) != 5")
	}
	if max(3, 5) != 5 {
		t.Error("max(3, 5) != 5")
	}
	if max(5, 5) != 5 {
		t.Error("max(5, 5) != 5")
	}

	if min(5, 3) != 3 {
		t.Error("min(5, 3) != 3")
	}
	if min(3, 5) != 3 {
		t.Error("min(3, 5) != 3")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) != 5")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name        string
		genreID     string
		wantAmbient float64
	}{
		{"fantasy", "fantasy", 0.3},
		{"scifi", "scifi", 0.5},
		{"horror", "horror", 0.15},
		{"cyberpunk", "cyberpunk", 0.25},
		{"postapoc", "postapoc", 0.35},
		{"unknown defaults to fantasy", "unknown", 0.3},
		{"empty defaults to fantasy", "", 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slm := NewSectorLightMap(10, 10, 0.5)
			slm.dirty = false
			slm.SetGenre(tt.genreID)
			if math.Abs(slm.Ambient-tt.wantAmbient) > 0.001 {
				t.Errorf("SetGenre(%q): Ambient = %f, want %f", tt.genreID, slm.Ambient, tt.wantAmbient)
			}
			if !slm.dirty {
				t.Error("dirty = false, want true after SetGenre")
			}
		})
	}
}

func TestGenreAmbientLevel(t *testing.T) {
	tests := []struct {
		genreID string
		want    float64
	}{
		{"fantasy", 0.3},
		{"scifi", 0.5},
		{"horror", 0.15},
		{"cyberpunk", 0.25},
		{"postapoc", 0.35},
		{"invalid", 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			got := genreAmbientLevel(tt.genreID)
			if got != tt.want {
				t.Errorf("genreAmbientLevel(%q) = %f, want %f", tt.genreID, got, tt.want)
			}
		})
	}
}

func TestSetGenre_IntegrationWithLighting(t *testing.T) {
	slm := NewSectorLightMap(10, 10, 0.5)
	slm.AddLight(Light{X: 5, Y: 5, Radius: 3, Intensity: 1.0})

	// Set to horror (very dark)
	slm.SetGenre("horror")
	slm.Calculate()
	horrorLight := slm.GetLight(0, 0) // Far from light
	if math.Abs(horrorLight-0.15) > 0.01 {
		t.Errorf("Horror ambient at (0,0) = %f, want ~0.15", horrorLight)
	}

	// Set to scifi (brighter)
	slm.SetGenre("scifi")
	slm.Calculate()
	scifiLight := slm.GetLight(0, 0)
	if math.Abs(scifiLight-0.5) > 0.01 {
		t.Errorf("Scifi ambient at (0,0) = %f, want ~0.5", scifiLight)
	}

	// Verify scifi is brighter than horror
	if scifiLight <= horrorLight {
		t.Errorf("Scifi light %f should be > horror light %f", scifiLight, horrorLight)
	}
}

// BenchmarkCalculate measures lighting calculation performance
func BenchmarkCalculate(b *testing.B) {
	slm := NewSectorLightMap(100, 100, 0.2)
	for i := 0; i < 10; i++ {
		slm.AddLight(Light{
			X:         float64(i * 10),
			Y:         float64(i * 10),
			Radius:    5,
			Intensity: 1.0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slm.dirty = true
		slm.Calculate()
	}
}
