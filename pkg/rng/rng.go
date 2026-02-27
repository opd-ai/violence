// Package rng provides a seed-based random number generator.
package rng

import "math/rand"

// RNG wraps a seeded random source.
type RNG struct {
	r *rand.Rand
}

// NewRNG creates a new RNG with the given seed.
func NewRNG(seed int64) *RNG {
	return &RNG{r: rand.New(rand.NewSource(seed))}
}

// Intn returns a non-negative random int in [0, n).
func (g *RNG) Intn(n int) int {
	return g.r.Intn(n)
}

// Float64 returns a random float64 in [0.0, 1.0).
func (g *RNG) Float64() float64 {
	return g.r.Float64()
}

// Seed resets the RNG with a new seed.
func (g *RNG) Seed(seed int64) {
	g.r = rand.New(rand.NewSource(seed))
}
