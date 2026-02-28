// Package ammo manages ammunition types and pools.
package ammo

const (
	Bullets = "bullets"
	Shells  = "shells"
	Rockets = "rockets"
	Cells   = "cells"
	Mana    = "mana"
)

// Pool tracks ammunition counts by type.
type Pool struct {
	counts map[string]int
}

// NewPool creates an empty ammo pool.
func NewPool() *Pool {
	return &Pool{counts: make(map[string]int)}
}

// Add increases ammo of the given type.
func (p *Pool) Add(ammoType string, amount int) {
	p.counts[ammoType] += amount
}

// Consume decreases ammo of the given type. Returns false if insufficient.
func (p *Pool) Consume(ammoType string, amount int) bool {
	if p.counts[ammoType] < amount {
		return false
	}
	p.counts[ammoType] -= amount
	return true
}

// Get returns the current amount of the given ammo type.
func (p *Pool) Get(ammoType string) int {
	return p.counts[ammoType]
}

// Set sets the amount of the given ammo type directly (used for save/load).
func (p *Pool) Set(ammoType string, amount int) {
	p.counts[ammoType] = amount
}

var currentGenre = "fantasy"

// SetGenre configures ammo types for a genre.
func SetGenre(genreID string) {
	currentGenre = genreID
}

// GetCurrentGenre returns the current global genre setting.
func GetCurrentGenre() string {
	return currentGenre
}
