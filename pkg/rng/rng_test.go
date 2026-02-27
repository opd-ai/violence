package rng

import "testing"

// TestNewRNG verifies RNG initialization.
func TestNewRNG(t *testing.T) {
	rng := NewRNG(12345)
	if rng == nil {
		t.Fatal("NewRNG() returned nil")
	}
	if rng.r == nil {
		t.Fatal("RNG.r is nil")
	}
}

// TestIntn verifies random integer generation.
func TestIntn(t *testing.T) {
	tests := []struct {
		name  string
		seed  uint64
		n     int
		count int
	}{
		{"small range", 42, 10, 100},
		{"medium range", 12345, 100, 100},
		{"large range", 99999, 1000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := NewRNG(tt.seed)

			for i := 0; i < tt.count; i++ {
				val := rng.Intn(tt.n)
				if val < 0 || val >= tt.n {
					t.Errorf("Intn(%d) returned %d, want [0, %d)", tt.n, val, tt.n)
				}
			}
		})
	}
}

// TestFloat64 verifies random float generation.
func TestFloat64(t *testing.T) {
	tests := []struct {
		name  string
		seed  uint64
		count int
	}{
		{"seed 42", 42, 100},
		{"seed 12345", 12345, 100},
		{"seed 99999", 99999, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := NewRNG(tt.seed)

			for i := 0; i < tt.count; i++ {
				val := rng.Float64()
				if val < 0.0 || val >= 1.0 {
					t.Errorf("Float64() returned %f, want [0.0, 1.0)", val)
				}
			}
		})
	}
}

// TestSeed verifies RNG re-seeding.
func TestSeed(t *testing.T) {
	rng := NewRNG(12345)

	// Generate first sequence
	first := make([]int, 10)
	for i := range first {
		first[i] = rng.Intn(100)
	}

	// Re-seed with same seed
	rng.Seed(12345)

	// Generate second sequence
	second := make([]int, 10)
	for i := range second {
		second[i] = rng.Intn(100)
	}

	// Sequences should match
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("Position %d: first=%d, second=%d", i, first[i], second[i])
		}
	}
}

// TestDeterminism verifies same seed produces same sequence.
func TestDeterminism(t *testing.T) {
	const seed = 42

	// Create two RNGs with same seed
	rng1 := NewRNG(seed)
	rng2 := NewRNG(seed)

	// Generate values
	for i := 0; i < 100; i++ {
		v1 := rng1.Intn(1000)
		v2 := rng2.Intn(1000)
		if v1 != v2 {
			t.Errorf("Position %d: rng1=%d, rng2=%d", i, v1, v2)
		}

		f1 := rng1.Float64()
		f2 := rng2.Float64()
		if f1 != f2 {
			t.Errorf("Position %d: rng1=%f, rng2=%f", i, f1, f2)
		}
	}
}

// BenchmarkIntn benchmarks random integer generation.
func BenchmarkIntn(b *testing.B) {
	rng := NewRNG(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Intn(100)
	}
}

// BenchmarkFloat64 benchmarks random float generation.
func BenchmarkFloat64(b *testing.B) {
	rng := NewRNG(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Float64()
	}
}
