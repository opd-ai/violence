package raycaster

import (
	"math"
	"testing"
)

func TestTrigLookupTables(t *testing.T) {
	tests := []struct {
		name      string
		angle     float64 // in radians
		tolerance float64 // acceptable error vs math library
	}{
		{"zero", 0.0, 0.001},
		{"pi/6", math.Pi / 6, 0.001},
		{"pi/4", math.Pi / 4, 0.001},
		{"pi/3", math.Pi / 3, 0.001},
		{"pi/2", math.Pi / 2, 0.001},
		{"2pi/3", 2 * math.Pi / 3, 0.001},
		{"3pi/4", 3 * math.Pi / 4, 0.001},
		{"5pi/6", 5 * math.Pi / 6, 0.001},
		{"pi", math.Pi, 0.001},
		{"3pi/2", 3 * math.Pi / 2, 0.001},
		{"2pi", 2 * math.Pi, 0.001},
		{"negative_pi/4", -math.Pi / 4, 0.001},
		{"negative_pi/2", -math.Pi / 2, 0.001},
		{"large_positive", 10 * math.Pi, 0.001},
		{"large_negative", -10 * math.Pi, 0.001},
	}

	for _, tt := range tests {
		t.Run("Sin_"+tt.name, func(t *testing.T) {
			got := Sin(tt.angle)
			want := math.Sin(tt.angle)
			if math.Abs(got-want) > tt.tolerance {
				t.Errorf("Sin(%f) = %f, want %f (diff: %f)", tt.angle, got, want, math.Abs(got-want))
			}
		})

		t.Run("Cos_"+tt.name, func(t *testing.T) {
			got := Cos(tt.angle)
			want := math.Cos(tt.angle)
			if math.Abs(got-want) > tt.tolerance {
				t.Errorf("Cos(%f) = %f, want %f (diff: %f)", tt.angle, got, want, math.Abs(got-want))
			}
		})
	}
}

func TestTan(t *testing.T) {
	tests := []struct {
		name      string
		angle     float64
		tolerance float64
	}{
		{"zero", 0.0, 0.001},
		{"pi/6", math.Pi / 6, 0.001},
		{"pi/4", math.Pi / 4, 0.001},
		{"pi/3", math.Pi / 3, 0.001},
		// Skip pi/2 - tangent is undefined
		{"2pi/3", 2 * math.Pi / 3, 0.01},
		{"3pi/4", 3 * math.Pi / 4, 0.01},
		{"5pi/6", 5 * math.Pi / 6, 0.01},
		{"pi", math.Pi, 0.001},
		// Skip 3pi/2 - tangent is undefined
		{"2pi", 2 * math.Pi, 0.001},
		{"negative_pi/4", -math.Pi / 4, 0.001},
		{"small_angle", 0.1, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tan(tt.angle)
			want := math.Tan(tt.angle)
			if math.Abs(got-want) > tt.tolerance {
				t.Errorf("Tan(%f) = %f, want %f (diff: %f)", tt.angle, got, want, math.Abs(got-want))
			}
		})
	}
}

func TestTrigTablesInitialized(t *testing.T) {
	// Verify tables are initialized
	if sinTable[0] != 0.0 {
		t.Errorf("sinTable[0] should be 0, got %f", sinTable[0])
	}

	// Check table size
	if len(sinTable) != tableSize {
		t.Errorf("sinTable size = %d, want %d", len(sinTable), tableSize)
	}
	if len(cosTable) != tableSize {
		t.Errorf("cosTable size = %d, want %d", len(cosTable), tableSize)
	}
	if len(tanTable) != tableSize {
		t.Errorf("tanTable size = %d, want %d", len(tanTable), tableSize)
	}

	// Verify some known values in tables
	// sin(π/2) should be 1.0 (at index 900 for tableSize 3600)
	idx90deg := tableSize / 4
	if math.Abs(sinTable[idx90deg]-1.0) > 0.001 {
		t.Errorf("sinTable[%d] (90°) = %f, want 1.0", idx90deg, sinTable[idx90deg])
	}

	// cos(0) should be 1.0
	if math.Abs(cosTable[0]-1.0) > 0.001 {
		t.Errorf("cosTable[0] (0°) = %f, want 1.0", cosTable[0])
	}

	// cos(π) should be -1.0 (at index 1800)
	idx180deg := tableSize / 2
	if math.Abs(cosTable[idx180deg]+1.0) > 0.001 {
		t.Errorf("cosTable[%d] (180°) = %f, want -1.0", idx180deg, cosTable[idx180deg])
	}
}

func TestTrigIdentities(t *testing.T) {
	// Test sin²(x) + cos²(x) = 1 for various angles
	angles := []float64{0, math.Pi / 6, math.Pi / 4, math.Pi / 3, math.Pi / 2, math.Pi}

	for _, angle := range angles {
		sin := Sin(angle)
		cos := Cos(angle)
		identity := sin*sin + cos*cos
		if math.Abs(identity-1.0) > 0.001 {
			t.Errorf("sin²(%f) + cos²(%f) = %f, want 1.0", angle, angle, identity)
		}
	}
}

func TestRaycasterUsesLookupTables(t *testing.T) {
	// Create a raycaster and verify it produces same results with lookup tables
	r := NewRaycaster(90.0, 320, 200)
	tileMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}
	r.SetMap(tileMap)

	// Cast rays and verify we get reasonable results
	hits := r.CastRays(2.5, 2.5, 1.0, 0.0)
	if len(hits) != 320 {
		t.Errorf("CastRays returned %d hits, want 320", len(hits))
	}

	// Center ray should hit the east wall at distance ~1.5
	centerHit := hits[160]
	if centerHit.Distance < 1.0 || centerHit.Distance > 2.0 {
		t.Errorf("Center ray distance = %f, expected ~1.5", centerHit.Distance)
	}
}

// Benchmark trig functions
func BenchmarkSinLookup(b *testing.B) {
	angle := math.Pi / 4
	for i := 0; i < b.N; i++ {
		_ = Sin(angle)
	}
}

func BenchmarkSinMath(b *testing.B) {
	angle := math.Pi / 4
	for i := 0; i < b.N; i++ {
		_ = math.Sin(angle)
	}
}

func BenchmarkCosLookup(b *testing.B) {
	angle := math.Pi / 4
	for i := 0; i < b.N; i++ {
		_ = Cos(angle)
	}
}

func BenchmarkCosMath(b *testing.B) {
	angle := math.Pi / 4
	for i := 0; i < b.N; i++ {
		_ = math.Cos(angle)
	}
}

func BenchmarkTanLookup(b *testing.B) {
	angle := math.Pi / 4
	for i := 0; i < b.N; i++ {
		_ = Tan(angle)
	}
}

func BenchmarkTanMath(b *testing.B) {
	angle := math.Pi / 4
	for i := 0; i < b.N; i++ {
		_ = math.Tan(angle)
	}
}

func BenchmarkRaycasterWithLookup(b *testing.B) {
	r := NewRaycaster(90.0, 320, 200)
	tileMap := make([][]int, 50)
	for i := range tileMap {
		tileMap[i] = make([]int, 50)
		for j := range tileMap[i] {
			if i == 0 || j == 0 || i == 49 || j == 49 {
				tileMap[i][j] = 1
			}
		}
	}
	r.SetMap(tileMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.CastRays(25.0, 25.0, 1.0, 0.0)
	}
}
