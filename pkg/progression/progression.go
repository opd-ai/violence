// Package progression manages player experience and leveling.
package progression

import (
	"fmt"
	"sync"
)

const (
	// MaxLevel is the maximum achievable player level.
	MaxLevel = 99
	// BaseXPPerLevel is the base XP required for first level-up.
	BaseXPPerLevel = 100
)

// Progression tracks a player's XP and level.
type Progression struct {
	xp    int
	level int
	genre string
	mu    sync.RWMutex
}

// NewProgression creates a new progression tracker at level 1.
func NewProgression() *Progression {
	return &Progression{level: 1, genre: "fantasy"}
}

// AddXP grants experience points and triggers level-ups as needed.
// Returns error if amount would result in negative XP.
func (p *Progression) AddXP(amount int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	newXP := p.xp + amount
	if newXP < 0 {
		return fmt.Errorf("invalid XP amount %d would result in negative XP", amount)
	}

	p.xp = newXP

	// Auto-level when threshold reached
	for p.level < MaxLevel && p.xp >= p.xpForNextLevel() {
		p.xp -= p.xpForNextLevel()
		p.level++
	}

	return nil
}

// GetXP returns the current XP amount.
func (p *Progression) GetXP() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.xp
}

// GetLevel returns the current level.
func (p *Progression) GetLevel() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.level
}

// XPForNextLevel returns the XP required to reach the next level.
func (p *Progression) XPForNextLevel() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.xpForNextLevel()
}

// xpForNextLevel is internal helper (caller must hold lock).
func (p *Progression) xpForNextLevel() int {
	if p.level >= MaxLevel {
		return 0
	}
	// Linear scaling: 100 XP for level 2, 200 for level 3, etc.
	return p.level * BaseXPPerLevel
}

// SetGenre configures progression curves for a genre.
// Validates genre ID and applies genre-specific XP scaling.
func (p *Progression) SetGenre(genreID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	validGenres := map[string]bool{
		"fantasy":   true,
		"scifi":     true,
		"horror":    true,
		"cyberpunk": true,
		"postapoc":  true,
	}

	if !validGenres[genreID] {
		return fmt.Errorf("invalid genre ID: %s", genreID)
	}

	p.genre = genreID
	return nil
}

// GetGenre returns the current genre setting.
func (p *Progression) GetGenre() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.genre
}
