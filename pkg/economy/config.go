// Package economy provides configurable game economy and reward systems.
package economy

import (
	"sync"
)

// Config holds all economy configuration parameters.
type Config struct {
	// Base rewards for various actions
	BaseKillReward      int `toml:"base_kill_reward"`
	BaseMissionReward   int `toml:"base_mission_reward"`
	BaseObjectiveReward int `toml:"base_objective_reward"`
	BaseItemPrice       int `toml:"base_item_price"`
	BaseWeaponPrice     int `toml:"base_weapon_price"`
	BaseArmorPrice      int `toml:"base_armor_price"`

	// Genre-specific multipliers
	GenreMultipliers map[string]float64 `toml:"genre_multipliers"`

	// Difficulty scaling
	DifficultyMultipliers map[string]float64 `toml:"difficulty_multipliers"`

	// Level progression scaling
	LevelScaling []LevelScaleEntry `toml:"level_scaling"`

	mu sync.RWMutex
}

// LevelScaleEntry defines reward scaling for a level range.
type LevelScaleEntry struct {
	MinLevel   int     `toml:"min_level"`
	MaxLevel   int     `toml:"max_level"`
	Multiplier float64 `toml:"multiplier"`
}

// NewConfig creates a default economy configuration.
func NewConfig() *Config {
	return &Config{
		BaseKillReward:      100,
		BaseMissionReward:   500,
		BaseObjectiveReward: 250,
		BaseItemPrice:       50,
		BaseWeaponPrice:     300,
		BaseArmorPrice:      200,
		GenreMultipliers: map[string]float64{
			"horror":    1.2,
			"scifi":     0.9,
			"fantasy":   1.0,
			"cyberpunk": 1.1,
			"postapoc":  1.15,
		},
		DifficultyMultipliers: map[string]float64{
			"easy":      0.8,
			"normal":    1.0,
			"hard":      1.3,
			"nightmare": 1.5,
		},
		LevelScaling: []LevelScaleEntry{
			{MinLevel: 1, MaxLevel: 3, Multiplier: 1.0},
			{MinLevel: 4, MaxLevel: 6, Multiplier: 1.2},
			{MinLevel: 7, MaxLevel: 9, Multiplier: 1.4},
			{MinLevel: 10, MaxLevel: 999, Multiplier: 1.7},
		},
	}
}

// CalculateKillReward computes reward for killing an enemy.
func (c *Config) CalculateKillReward(genre, difficulty string, playerLevel int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	base := float64(c.BaseKillReward)
	genreMult := c.getGenreMultiplier(genre)
	diffMult := c.getDifficultyMultiplier(difficulty)
	levelMult := c.getLevelMultiplier(playerLevel)

	return int(base * genreMult * diffMult * levelMult)
}

// CalculateMissionReward computes reward for completing a mission.
func (c *Config) CalculateMissionReward(genre, difficulty string, playerLevel int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	base := float64(c.BaseMissionReward)
	genreMult := c.getGenreMultiplier(genre)
	diffMult := c.getDifficultyMultiplier(difficulty)
	levelMult := c.getLevelMultiplier(playerLevel)

	return int(base * genreMult * diffMult * levelMult)
}

// CalculateObjectiveReward computes reward for completing an objective.
func (c *Config) CalculateObjectiveReward(genre, difficulty string, playerLevel int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	base := float64(c.BaseObjectiveReward)
	genreMult := c.getGenreMultiplier(genre)
	diffMult := c.getDifficultyMultiplier(difficulty)
	levelMult := c.getLevelMultiplier(playerLevel)

	return int(base * genreMult * diffMult * levelMult)
}

// CalculateItemPrice computes the price for an item.
func (c *Config) CalculateItemPrice(genre string, playerLevel int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	base := float64(c.BaseItemPrice)
	genreMult := c.getGenreMultiplier(genre)
	levelMult := c.getLevelMultiplier(playerLevel)

	return int(base * genreMult * levelMult)
}

// CalculateWeaponPrice computes the price for a weapon.
func (c *Config) CalculateWeaponPrice(genre string, playerLevel int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	base := float64(c.BaseWeaponPrice)
	genreMult := c.getGenreMultiplier(genre)
	levelMult := c.getLevelMultiplier(playerLevel)

	return int(base * genreMult * levelMult)
}

// CalculateArmorPrice computes the price for armor.
func (c *Config) CalculateArmorPrice(genre string, playerLevel int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	base := float64(c.BaseArmorPrice)
	genreMult := c.getGenreMultiplier(genre)
	levelMult := c.getLevelMultiplier(playerLevel)

	return int(base * genreMult * levelMult)
}

// SetGenreMultiplier updates the multiplier for a genre.
func (c *Config) SetGenreMultiplier(genre string, multiplier float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GenreMultipliers[genre] = multiplier
}

// SetDifficultyMultiplier updates the multiplier for a difficulty.
func (c *Config) SetDifficultyMultiplier(difficulty string, multiplier float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DifficultyMultipliers[difficulty] = multiplier
}

// getGenreMultiplier returns the multiplier for a genre (internal, no lock).
func (c *Config) getGenreMultiplier(genre string) float64 {
	if mult, ok := c.GenreMultipliers[genre]; ok {
		return mult
	}
	return 1.0
}

// getDifficultyMultiplier returns the multiplier for a difficulty (internal, no lock).
func (c *Config) getDifficultyMultiplier(difficulty string) float64 {
	if mult, ok := c.DifficultyMultipliers[difficulty]; ok {
		return mult
	}
	return 1.0
}

// getLevelMultiplier returns the multiplier for a player level (internal, no lock).
func (c *Config) getLevelMultiplier(level int) float64 {
	for _, scale := range c.LevelScaling {
		if level >= scale.MinLevel && level <= scale.MaxLevel {
			return scale.Multiplier
		}
	}
	return 1.0
}
