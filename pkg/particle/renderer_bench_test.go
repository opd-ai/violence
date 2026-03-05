package particle

import (
	"testing"
)

func BenchmarkDetermineShape(b *testing.B) {
	rs := NewRenderSystem()
	p := Particle{
		VX: 50, VY: 40, VZ: 0,
		R: 255, G: 100, B: 0, A: 255,
		Size: 3, Active: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.DetermineShape(&p, "fantasy")
	}
}

func BenchmarkDetermineShape_AllTypes(b *testing.B) {
	rs := NewRenderSystem()
	particles := []Particle{
		{VX: 100, VY: 50, VZ: 10, R: 255, G: 200, B: 50, A: 255, Size: 2, Active: true},
		{VX: 5, VY: 3, VZ: -10, R: 100, G: 100, B: 100, A: 150, Size: 3, Active: true},
		{VX: 20, VY: 15, VZ: 0, R: 180, G: 20, B: 20, A: 255, Size: 2, Active: true},
		{VX: 10, VY: 10, VZ: 0, R: 255, G: 220, B: 100, A: 255, Size: 2, Active: true},
		{VX: 50, VY: 40, VZ: 0, R: 255, G: 100, B: 0, A: 255, Size: 3, Active: true},
		{VX: 70, VY: 60, VZ: 0, R: 100, G: 100, B: 100, A: 255, Size: 2, Active: true},
		{VX: 10, VY: 10, VZ: 0, R: 100, G: 100, B: 100, A: 255, Size: 2, Active: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range particles {
			_ = rs.DetermineShape(&particles[j], "fantasy")
		}
	}
}
