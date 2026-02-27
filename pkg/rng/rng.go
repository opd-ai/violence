// Package rng provides a seed-based random number generator.
package rng

import "math/rand/v2"

// RNG wraps a seeded random source.
type RNG struct {
	r *rand.Rand
}

// NewRNG creates a new RNG with the given seed.
func NewRNG(seed uint64) *RNG {
	return &RNG{r: rand.New(rand.NewPCG(seed, seed^0xda3e39cb94b95bdb))}
}

// Intn returns a non-negative random int in [0, n).
func (g *RNG) Intn(n int) int {
	return g.r.IntN(n)
}

// Float64 returns a random float64 in [0.0, 1.0).
func (g *RNG) Float64() float64 {
	return g.r.Float64()
}

// Seed resets the RNG with a new seed.
func (g *RNG) Seed(seed uint64) {
	g.r = rand.New(rand.NewPCG(seed, seed^0xda3e39cb94b95bdb))
}
