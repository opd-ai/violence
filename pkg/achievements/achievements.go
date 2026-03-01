// Package achievements provides a local achievement tracking system with persistence.
package achievements

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// ErrAchievementNotFound indicates the requested achievement ID does not exist
	ErrAchievementNotFound = errors.New("achievement not found")
	// ErrInvalidStats indicates nil or invalid player stats provided
	ErrInvalidStats = errors.New("invalid player stats")
)

// Category represents achievement groupings
type Category string

const (
	CategoryCombat      Category = "combat"
	CategoryExploration Category = "exploration"
	CategorySurvival    Category = "survival"
	CategorySocial      Category = "social"
)

// PlayerStats contains all player statistics tracked for achievement conditions
type PlayerStats struct {
	// Combat stats
	Kills                int
	Deaths               int
	HeadshotKills        int
	MeleeKills           int
	ExplosiveKills       int
	WeaponKills          map[string]int // Kills per weapon type
	DamageDealt          int64
	DamageTaken          int64
	CriticalHits         int
	PerfectAccuracyShots int // Shots that hit without missing

	// Exploration stats
	TilesRevealed  int
	TotalTiles     int
	SecretsFound   int
	TotalSecrets   int
	DoorsOpened    int
	ItemsCollected int

	// Survival stats
	TotalDeaths         int
	CompletedLevels     int
	LevelCompletionTime time.Duration
	HealthLostTotal     int

	// Social stats
	CoopGamesPlayed  int
	CoopGamesWon     int
	DeathmatchWins   int
	DeathmatchLosses int
	MessagesReceived int
	MessagesSent     int
}

// Achievement defines a single achievement with unlock conditions
type Achievement struct {
	ID          string
	Name        string
	Description string
	Category    Category
	Hidden      bool // Hidden until unlocked

	// Condition checks if achievement should unlock based on stats
	Condition func(stats *PlayerStats) bool

	// Progress computes current progress towards achievement (current, target)
	Progress func(stats *PlayerStats) (int, int)
}

// UnlockedAchievement tracks when an achievement was unlocked
type UnlockedAchievement struct {
	ID         string    `json:"id"`
	UnlockedAt time.Time `json:"unlocked_at"`
}

// AchievementManager manages achievement definitions and unlock state
type AchievementManager struct {
	achievements map[string]Achievement
	unlocked     map[string]UnlockedAchievement
	savePath     string
	mu           sync.RWMutex
}

// NewAchievementManager creates a new achievement manager with default achievements
func NewAchievementManager(savePath string) (*AchievementManager, error) {
	am := &AchievementManager{
		achievements: make(map[string]Achievement),
		unlocked:     make(map[string]UnlockedAchievement),
		savePath:     savePath,
	}

	// Register default achievements
	am.registerDefaultAchievements()

	// Load existing unlocks
	if err := am.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load achievements: %w", err)
	}

	return am, nil
}

// registerDefaultAchievements registers all built-in achievements
func (am *AchievementManager) registerDefaultAchievements() {
	// Combat achievements
	am.Register(Achievement{
		ID:          "first_blood",
		Name:        "First Blood",
		Description: "Kill your first enemy",
		Category:    CategoryCombat,
		Condition:   func(s *PlayerStats) bool { return s.Kills >= 1 },
		Progress:    func(s *PlayerStats) (int, int) { return min(s.Kills, 1), 1 },
	})

	am.Register(Achievement{
		ID:          "centurion",
		Name:        "Centurion",
		Description: "Kill 100 enemies",
		Category:    CategoryCombat,
		Condition:   func(s *PlayerStats) bool { return s.Kills >= 100 },
		Progress:    func(s *PlayerStats) (int, int) { return s.Kills, 100 },
	})

	am.Register(Achievement{
		ID:          "pacifist",
		Name:        "Pacifist",
		Description: "Complete a level without killing any enemies",
		Category:    CategoryCombat,
		Condition:   func(s *PlayerStats) bool { return s.CompletedLevels > 0 && s.Kills == 0 },
		Progress: func(s *PlayerStats) (int, int) {
			if s.Kills == 0 && s.CompletedLevels > 0 {
				return 1, 1
			}
			return 0, 1
		},
	})

	am.Register(Achievement{
		ID:          "headhunter",
		Name:        "Headhunter",
		Description: "Get 50 headshot kills",
		Category:    CategoryCombat,
		Condition:   func(s *PlayerStats) bool { return s.HeadshotKills >= 50 },
		Progress:    func(s *PlayerStats) (int, int) { return s.HeadshotKills, 50 },
	})

	am.Register(Achievement{
		ID:          "demolition_expert",
		Name:        "Demolition Expert",
		Description: "Kill 25 enemies with explosives",
		Category:    CategoryCombat,
		Condition:   func(s *PlayerStats) bool { return s.ExplosiveKills >= 25 },
		Progress:    func(s *PlayerStats) (int, int) { return s.ExplosiveKills, 25 },
	})

	// Exploration achievements
	am.Register(Achievement{
		ID:          "cartographer",
		Name:        "Cartographer",
		Description: "Reveal 100% of the map",
		Category:    CategoryExploration,
		Condition: func(s *PlayerStats) bool {
			if s.TotalTiles == 0 {
				return false
			}
			return s.TilesRevealed >= s.TotalTiles
		},
		Progress: func(s *PlayerStats) (int, int) {
			if s.TotalTiles == 0 {
				return 0, 100
			}
			return s.TilesRevealed, s.TotalTiles
		},
	})

	am.Register(Achievement{
		ID:          "secret_hunter",
		Name:        "Secret Hunter",
		Description: "Find 10 secret areas",
		Category:    CategoryExploration,
		Condition:   func(s *PlayerStats) bool { return s.SecretsFound >= 10 },
		Progress:    func(s *PlayerStats) (int, int) { return s.SecretsFound, 10 },
	})

	am.Register(Achievement{
		ID:          "explorer",
		Name:        "Explorer",
		Description: "Open 100 doors",
		Category:    CategoryExploration,
		Condition:   func(s *PlayerStats) bool { return s.DoorsOpened >= 100 },
		Progress:    func(s *PlayerStats) (int, int) { return s.DoorsOpened, 100 },
	})

	// Survival achievements
	am.Register(Achievement{
		ID:          "iron_man",
		Name:        "Iron Man",
		Description: "Complete a level without dying",
		Category:    CategorySurvival,
		Condition: func(s *PlayerStats) bool {
			return s.CompletedLevels > 0 && s.TotalDeaths == 0
		},
		Progress: func(s *PlayerStats) (int, int) {
			if s.TotalDeaths == 0 && s.CompletedLevels > 0 {
				return 1, 1
			}
			return 0, 1
		},
	})

	am.Register(Achievement{
		ID:          "speed_demon",
		Name:        "Speed Demon",
		Description: "Complete a level in under 5 minutes",
		Category:    CategorySurvival,
		Condition: func(s *PlayerStats) bool {
			return s.CompletedLevels > 0 && s.LevelCompletionTime > 0 &&
				s.LevelCompletionTime < 5*time.Minute
		},
		Progress: func(s *PlayerStats) (int, int) {
			if s.LevelCompletionTime == 0 {
				return 0, 1
			}
			if s.LevelCompletionTime < 5*time.Minute {
				return 1, 1
			}
			return 0, 1
		},
	})

	am.Register(Achievement{
		ID:          "untouchable",
		Name:        "Untouchable",
		Description: "Complete a level without taking damage",
		Category:    CategorySurvival,
		Condition: func(s *PlayerStats) bool {
			return s.CompletedLevels > 0 && s.DamageTaken == 0
		},
		Progress: func(s *PlayerStats) (int, int) {
			if s.DamageTaken == 0 && s.CompletedLevels > 0 {
				return 1, 1
			}
			return 0, 1
		},
	})

	// Social achievements
	am.Register(Achievement{
		ID:          "team_player",
		Name:        "Team Player",
		Description: "Complete 10 co-op games",
		Category:    CategorySocial,
		Condition:   func(s *PlayerStats) bool { return s.CoopGamesPlayed >= 10 },
		Progress:    func(s *PlayerStats) (int, int) { return s.CoopGamesPlayed, 10 },
	})

	am.Register(Achievement{
		ID:          "dominator",
		Name:        "Dominator",
		Description: "Win 10 deathmatch games",
		Category:    CategorySocial,
		Condition:   func(s *PlayerStats) bool { return s.DeathmatchWins >= 10 },
		Progress:    func(s *PlayerStats) (int, int) { return s.DeathmatchWins, 10 },
	})

	am.Register(Achievement{
		ID:          "social_butterfly",
		Name:        "Social Butterfly",
		Description: "Send 100 chat messages",
		Category:    CategorySocial,
		Condition:   func(s *PlayerStats) bool { return s.MessagesSent >= 100 },
		Progress:    func(s *PlayerStats) (int, int) { return s.MessagesSent, 100 },
	})
}

// Register adds a new achievement definition
func (am *AchievementManager) Register(achievement Achievement) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.achievements[achievement.ID] = achievement
}

// CheckUnlocks checks player stats against all achievements and returns newly unlocked ones
func (am *AchievementManager) CheckUnlocks(stats *PlayerStats) ([]Achievement, error) {
	if stats == nil {
		return nil, ErrInvalidStats
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	var newlyUnlocked []Achievement

	for id, achievement := range am.achievements {
		// Skip already unlocked
		if _, unlocked := am.unlocked[id]; unlocked {
			continue
		}

		// Check condition
		if achievement.Condition != nil && achievement.Condition(stats) {
			unlock := UnlockedAchievement{
				ID:         id,
				UnlockedAt: time.Now(),
			}
			am.unlocked[id] = unlock
			newlyUnlocked = append(newlyUnlocked, achievement)

			logrus.WithFields(logrus.Fields{
				"achievement_id":   id,
				"achievement_name": achievement.Name,
				"category":         achievement.Category,
			}).Info("Achievement unlocked")
		}
	}

	// Persist unlocks (unlock mutex before saving to avoid deadlock)
	if len(newlyUnlocked) > 0 {
		am.mu.Unlock()
		err := am.Save()
		am.mu.Lock()
		if err != nil {
			return newlyUnlocked, fmt.Errorf("failed to save unlocks: %w", err)
		}
	}

	return newlyUnlocked, nil
}

// GetProgress returns current progress for a specific achievement
func (am *AchievementManager) GetProgress(achievementID string) (int, int, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	achievement, exists := am.achievements[achievementID]
	if !exists {
		return 0, 0, ErrAchievementNotFound
	}

	// If unlocked, return complete
	if _, unlocked := am.unlocked[achievementID]; unlocked {
		if achievement.Progress != nil {
			_, target := achievement.Progress(&PlayerStats{})
			return target, target, nil
		}
		return 1, 1, nil
	}

	return 0, 1, nil
}

// GetProgressWithStats returns current progress for an achievement based on provided stats
func (am *AchievementManager) GetProgressWithStats(achievementID string, stats *PlayerStats) (int, int, error) {
	if stats == nil {
		return 0, 0, ErrInvalidStats
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	achievement, exists := am.achievements[achievementID]
	if !exists {
		return 0, 0, ErrAchievementNotFound
	}

	if achievement.Progress == nil {
		// Binary achievement - either 0 or 1
		if _, unlocked := am.unlocked[achievementID]; unlocked {
			return 1, 1, nil
		}
		return 0, 1, nil
	}

	current, target := achievement.Progress(stats)
	return current, target, nil
}

// IsUnlocked checks if an achievement has been unlocked
func (am *AchievementManager) IsUnlocked(achievementID string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	_, unlocked := am.unlocked[achievementID]
	return unlocked
}

// GetUnlocked returns all unlocked achievements
func (am *AchievementManager) GetUnlocked() []UnlockedAchievement {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make([]UnlockedAchievement, 0, len(am.unlocked))
	for _, unlock := range am.unlocked {
		result = append(result, unlock)
	}
	return result
}

// GetAll returns all registered achievements
func (am *AchievementManager) GetAll() []Achievement {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make([]Achievement, 0, len(am.achievements))
	for _, achievement := range am.achievements {
		result = append(result, achievement)
	}
	return result
}

// GetByCategory returns all achievements in a specific category
func (am *AchievementManager) GetByCategory(category Category) []Achievement {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var result []Achievement
	for _, achievement := range am.achievements {
		if achievement.Category == category {
			result = append(result, achievement)
		}
	}
	return result
}

// Save persists unlocked achievements to disk
func (am *AchievementManager) Save() error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(am.savePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Convert map to slice for JSON
	unlocks := make([]UnlockedAchievement, 0, len(am.unlocked))
	for _, unlock := range am.unlocked {
		unlocks = append(unlocks, unlock)
	}

	data, err := json.MarshalIndent(unlocks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal unlocks: %w", err)
	}

	if err := os.WriteFile(am.savePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	return nil
}

// Load reads unlocked achievements from disk
func (am *AchievementManager) Load() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	data, err := os.ReadFile(am.savePath)
	if err != nil {
		return err
	}

	var unlocks []UnlockedAchievement
	if err := json.Unmarshal(data, &unlocks); err != nil {
		return fmt.Errorf("failed to unmarshal unlocks: %w", err)
	}

	// Convert slice to map
	am.unlocked = make(map[string]UnlockedAchievement)
	for _, unlock := range unlocks {
		am.unlocked[unlock.ID] = unlock
	}

	return nil
}

// Reset clears all unlocked achievements (for testing or game reset)
func (am *AchievementManager) Reset() error {
	am.mu.Lock()
	am.unlocked = make(map[string]UnlockedAchievement)
	am.mu.Unlock()

	return am.Save()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
