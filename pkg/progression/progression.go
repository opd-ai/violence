// Package progression manages player experience and leveling.
package progression

// Progression tracks a player's XP and level.
type Progression struct {
	XP    int
	Level int
}

// NewProgression creates a new progression tracker at level 1.
func NewProgression() *Progression {
	return &Progression{Level: 1}
}

// AddXP grants experience points and triggers level-ups as needed.
func (p *Progression) AddXP(amount int) {
	p.XP += amount
}

// LevelUp advances the player to the next level.
func (p *Progression) LevelUp() {
	p.Level++
}

// SetGenre configures progression curves for a genre.
func SetGenre(genreID string) {}
