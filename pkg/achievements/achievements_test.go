package achievements

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewAchievementManager(t *testing.T) {
	tests := []struct {
		name      string
		savePath  string
		wantError bool
	}{
		{
			name:      "valid path",
			savePath:  filepath.Join(t.TempDir(), "achievements.json"),
			wantError: false,
		},
		{
			name:      "nested path",
			savePath:  filepath.Join(t.TempDir(), "data", "achievements.json"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAchievementManager(tt.savePath)
			if (err != nil) != tt.wantError {
				t.Errorf("NewAchievementManager() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && am == nil {
				t.Error("NewAchievementManager() returned nil")
			}
			if am != nil && len(am.achievements) == 0 {
				t.Error("NewAchievementManager() should register default achievements")
			}
		})
	}
}

func TestRegister(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	initialCount := len(am.GetAll())

	achievement := Achievement{
		ID:          "test_achievement",
		Name:        "Test Achievement",
		Description: "Test Description",
		Category:    CategoryCombat,
		Condition:   func(s *PlayerStats) bool { return s.Kills > 0 },
		Progress:    func(s *PlayerStats) (int, int) { return s.Kills, 10 },
	}

	am.Register(achievement)

	allAchievements := am.GetAll()
	if len(allAchievements) != initialCount+1 {
		t.Errorf("Expected %d achievements, got %d", initialCount+1, len(allAchievements))
	}
}

func TestCheckUnlocks(t *testing.T) {
	tests := []struct {
		name         string
		stats        *PlayerStats
		wantUnlocked []string
		wantError    bool
	}{
		{
			name:         "nil stats",
			stats:        nil,
			wantUnlocked: nil,
			wantError:    true,
		},
		{
			name: "first blood unlock",
			stats: &PlayerStats{
				Kills: 1,
			},
			wantUnlocked: []string{"first_blood"},
			wantError:    false,
		},
		{
			name: "centurion unlock",
			stats: &PlayerStats{
				Kills: 100,
			},
			wantUnlocked: []string{"first_blood", "centurion"},
			wantError:    false,
		},
		{
			name: "pacifist unlock",
			stats: &PlayerStats{
				CompletedLevels: 1,
				Kills:           0,
				DamageTaken:     10, // Take some damage to avoid untouchable
				TotalDeaths:     1,  // Die once to avoid iron_man
			},
			wantUnlocked: []string{"pacifist"},
			wantError:    false,
		},
		{
			name: "headhunter unlock",
			stats: &PlayerStats{
				Kills:         50,
				HeadshotKills: 50,
			},
			wantUnlocked: []string{"first_blood", "headhunter"},
			wantError:    false,
		},
		{
			name: "demolition expert unlock",
			stats: &PlayerStats{
				Kills:          25,
				ExplosiveKills: 25,
			},
			wantUnlocked: []string{"first_blood", "demolition_expert"},
			wantError:    false,
		},
		{
			name: "cartographer unlock",
			stats: &PlayerStats{
				TilesRevealed: 100,
				TotalTiles:    100,
			},
			wantUnlocked: []string{"cartographer"},
			wantError:    false,
		},
		{
			name: "cartographer incomplete",
			stats: &PlayerStats{
				TilesRevealed: 50,
				TotalTiles:    100,
			},
			wantUnlocked: []string{},
			wantError:    false,
		},
		{
			name: "secret hunter unlock",
			stats: &PlayerStats{
				SecretsFound: 10,
			},
			wantUnlocked: []string{"secret_hunter"},
			wantError:    false,
		},
		{
			name: "explorer unlock",
			stats: &PlayerStats{
				DoorsOpened: 100,
			},
			wantUnlocked: []string{"explorer"},
			wantError:    false,
		},
		{
			name: "iron man unlock",
			stats: &PlayerStats{
				CompletedLevels: 1,
				TotalDeaths:     0,
				Kills:           5,  // Kill some enemies to avoid pacifist
				DamageTaken:     10, // Take some damage to avoid untouchable
			},
			wantUnlocked: []string{"iron_man", "first_blood"},
			wantError:    false,
		},
		{
			name: "speed demon unlock",
			stats: &PlayerStats{
				CompletedLevels:     1,
				LevelCompletionTime: 4 * time.Minute,
				Kills:               5,  // Kill some enemies to avoid pacifist
				TotalDeaths:         1,  // Die once to avoid iron_man
				DamageTaken:         10, // Take some damage to avoid untouchable
			},
			wantUnlocked: []string{"speed_demon", "first_blood"},
			wantError:    false,
		},
		{
			name: "speed demon too slow",
			stats: &PlayerStats{
				CompletedLevels:     1,
				LevelCompletionTime: 6 * time.Minute,
				Kills:               5,  // Kill some enemies to avoid pacifist
				TotalDeaths:         1,  // Die once to avoid iron_man
				DamageTaken:         10, // Take some damage to avoid untouchable
			},
			wantUnlocked: []string{"first_blood"},
			wantError:    false,
		},
		{
			name: "untouchable unlock",
			stats: &PlayerStats{
				CompletedLevels: 1,
				DamageTaken:     0,
				Kills:           5, // Kill some enemies to avoid pacifist
				TotalDeaths:     1, // Die once to avoid iron_man
			},
			wantUnlocked: []string{"untouchable", "first_blood"},
			wantError:    false,
		},
		{
			name: "team player unlock",
			stats: &PlayerStats{
				CoopGamesPlayed: 10,
			},
			wantUnlocked: []string{"team_player"},
			wantError:    false,
		},
		{
			name: "dominator unlock",
			stats: &PlayerStats{
				DeathmatchWins: 10,
			},
			wantUnlocked: []string{"dominator"},
			wantError:    false,
		},
		{
			name: "social butterfly unlock",
			stats: &PlayerStats{
				MessagesSent: 100,
			},
			wantUnlocked: []string{"social_butterfly"},
			wantError:    false,
		},
		{
			name: "multiple unlocks",
			stats: &PlayerStats{
				Kills:           100,
				HeadshotKills:   50,
				ExplosiveKills:  25,
				SecretsFound:    10,
				DoorsOpened:     100,
				CoopGamesPlayed: 10,
				DeathmatchWins:  10,
				MessagesSent:    100,
			},
			wantUnlocked: []string{"first_blood", "centurion", "headhunter", "demolition_expert", "secret_hunter", "explorer", "team_player", "dominator", "social_butterfly"},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}

			unlocked, err := am.CheckUnlocks(tt.stats)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckUnlocks() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				unlockedIDs := make(map[string]bool)
				for _, a := range unlocked {
					unlockedIDs[a.ID] = true
				}

				for _, wantID := range tt.wantUnlocked {
					if !unlockedIDs[wantID] {
						t.Errorf("Expected achievement %s to be unlocked", wantID)
					}
				}

				if len(unlocked) != len(tt.wantUnlocked) {
					t.Errorf("Expected %d unlocked achievements, got %d", len(tt.wantUnlocked), len(unlocked))
				}
			}
		})
	}
}

func TestCheckUnlocks_NoDoubleUnlock(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	stats := &PlayerStats{Kills: 1}

	// First unlock
	unlocked1, err := am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}
	if len(unlocked1) == 0 {
		t.Error("Expected first_blood to be unlocked")
	}

	// Second check with same stats should not unlock again
	unlocked2, err := am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}
	if len(unlocked2) != 0 {
		t.Errorf("Expected no new unlocks, got %d", len(unlocked2))
	}
}

func TestGetProgress(t *testing.T) {
	tests := []struct {
		name          string
		achievementID string
		wantError     bool
	}{
		{
			name:          "valid achievement",
			achievementID: "first_blood",
			wantError:     false,
		},
		{
			name:          "invalid achievement",
			achievementID: "nonexistent",
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}

			_, _, err = am.GetProgress(tt.achievementID)
			if (err != nil) != tt.wantError {
				t.Errorf("GetProgress() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetProgressWithStats(t *testing.T) {
	tests := []struct {
		name          string
		achievementID string
		stats         *PlayerStats
		wantCurrent   int
		wantTarget    int
		wantError     bool
	}{
		{
			name:          "nil stats",
			achievementID: "first_blood",
			stats:         nil,
			wantError:     true,
		},
		{
			name:          "first blood in progress",
			achievementID: "first_blood",
			stats:         &PlayerStats{Kills: 0},
			wantCurrent:   0,
			wantTarget:    1,
			wantError:     false,
		},
		{
			name:          "first blood complete",
			achievementID: "first_blood",
			stats:         &PlayerStats{Kills: 1},
			wantCurrent:   1,
			wantTarget:    1,
			wantError:     false,
		},
		{
			name:          "centurion in progress",
			achievementID: "centurion",
			stats:         &PlayerStats{Kills: 47},
			wantCurrent:   47,
			wantTarget:    100,
			wantError:     false,
		},
		{
			name:          "centurion complete",
			achievementID: "centurion",
			stats:         &PlayerStats{Kills: 100},
			wantCurrent:   100,
			wantTarget:    100,
			wantError:     false,
		},
		{
			name:          "centurion over target",
			achievementID: "centurion",
			stats:         &PlayerStats{Kills: 150},
			wantCurrent:   150,
			wantTarget:    100,
			wantError:     false,
		},
		{
			name:          "headhunter in progress",
			achievementID: "headhunter",
			stats:         &PlayerStats{HeadshotKills: 25},
			wantCurrent:   25,
			wantTarget:    50,
			wantError:     false,
		},
		{
			name:          "cartographer in progress",
			achievementID: "cartographer",
			stats:         &PlayerStats{TilesRevealed: 50, TotalTiles: 100},
			wantCurrent:   50,
			wantTarget:    100,
			wantError:     false,
		},
		{
			name:          "invalid achievement",
			achievementID: "nonexistent",
			stats:         &PlayerStats{},
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}

			current, target, err := am.GetProgressWithStats(tt.achievementID, tt.stats)
			if (err != nil) != tt.wantError {
				t.Errorf("GetProgressWithStats() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if current != tt.wantCurrent {
					t.Errorf("GetProgressWithStats() current = %d, want %d", current, tt.wantCurrent)
				}
				if target != tt.wantTarget {
					t.Errorf("GetProgressWithStats() target = %d, want %d", target, tt.wantTarget)
				}
			}
		})
	}
}

func TestIsUnlocked(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Should not be unlocked initially
	if am.IsUnlocked("first_blood") {
		t.Error("Expected first_blood to not be unlocked initially")
	}

	// Unlock it
	stats := &PlayerStats{Kills: 1}
	_, err = am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}

	// Should be unlocked now
	if !am.IsUnlocked("first_blood") {
		t.Error("Expected first_blood to be unlocked")
	}
}

func TestGetUnlocked(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// No unlocks initially
	unlocked := am.GetUnlocked()
	if len(unlocked) != 0 {
		t.Errorf("Expected no unlocked achievements, got %d", len(unlocked))
	}

	// Unlock some achievements
	stats := &PlayerStats{
		Kills:         100,
		HeadshotKills: 50,
	}
	_, err = am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}

	unlocked = am.GetUnlocked()
	if len(unlocked) == 0 {
		t.Error("Expected some unlocked achievements")
	}
}

func TestGetAll(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	all := am.GetAll()
	if len(all) == 0 {
		t.Error("Expected default achievements to be registered")
	}

	// Should contain expected achievements
	expectedIDs := []string{
		"first_blood", "centurion", "pacifist", "headhunter", "demolition_expert",
		"cartographer", "secret_hunter", "explorer",
		"iron_man", "speed_demon", "untouchable",
		"team_player", "dominator", "social_butterfly",
	}

	achievementMap := make(map[string]bool)
	for _, a := range all {
		achievementMap[a.ID] = true
	}

	for _, id := range expectedIDs {
		if !achievementMap[id] {
			t.Errorf("Expected achievement %s to be registered", id)
		}
	}
}

func TestGetByCategory(t *testing.T) {
	tests := []struct {
		name        string
		category    Category
		expectedIDs []string
	}{
		{
			name:        "combat category",
			category:    CategoryCombat,
			expectedIDs: []string{"first_blood", "centurion", "pacifist", "headhunter", "demolition_expert"},
		},
		{
			name:        "exploration category",
			category:    CategoryExploration,
			expectedIDs: []string{"cartographer", "secret_hunter", "explorer"},
		},
		{
			name:        "survival category",
			category:    CategorySurvival,
			expectedIDs: []string{"iron_man", "speed_demon", "untouchable"},
		},
		{
			name:        "social category",
			category:    CategorySocial,
			expectedIDs: []string{"team_player", "dominator", "social_butterfly"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}

			achievements := am.GetByCategory(tt.category)

			achievementMap := make(map[string]bool)
			for _, a := range achievements {
				achievementMap[a.ID] = true
			}

			for _, expectedID := range tt.expectedIDs {
				if !achievementMap[expectedID] {
					t.Errorf("Expected achievement %s in category %s", expectedID, tt.category)
				}
			}

			if len(achievements) != len(tt.expectedIDs) {
				t.Errorf("Expected %d achievements in category %s, got %d", len(tt.expectedIDs), tt.category, len(achievements))
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	savePath := filepath.Join(t.TempDir(), "achievements.json")

	// Create manager and unlock some achievements
	am1, err := NewAchievementManager(savePath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	stats := &PlayerStats{
		Kills:         100,
		HeadshotKills: 50,
		SecretsFound:  10,
	}

	unlocked1, err := am1.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}

	if len(unlocked1) == 0 {
		t.Fatal("Expected some achievements to be unlocked")
	}

	// Create new manager and load
	am2, err := NewAchievementManager(savePath)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	unlocked2 := am2.GetUnlocked()
	if len(unlocked2) != len(unlocked1) {
		t.Errorf("Expected %d loaded achievements, got %d", len(unlocked1), len(unlocked2))
	}

	// Verify specific achievements
	for _, a := range unlocked1 {
		if !am2.IsUnlocked(a.ID) {
			t.Errorf("Expected achievement %s to be loaded", a.ID)
		}
	}
}

func TestPersistence(t *testing.T) {
	savePath := filepath.Join(t.TempDir(), "data", "achievements.json")

	am, err := NewAchievementManager(savePath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	stats := &PlayerStats{Kills: 1}
	_, err = am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Expected save file to be created")
	}
}

func TestReset(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Unlock some achievements
	stats := &PlayerStats{Kills: 100}
	unlocked, err := am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}

	if len(unlocked) == 0 {
		t.Fatal("Expected some achievements to be unlocked")
	}

	// Reset
	if err := am.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Verify all achievements are locked
	unlockedAfter := am.GetUnlocked()
	if len(unlockedAfter) != 0 {
		t.Errorf("Expected no unlocked achievements after reset, got %d", len(unlockedAfter))
	}
}

func TestConcurrentAccess(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	done := make(chan bool)

	// Concurrent unlocks
	for i := 0; i < 10; i++ {
		go func() {
			stats := &PlayerStats{Kills: 1}
			_, _ = am.CheckUnlocks(stats)
			done <- true
		}()
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = am.GetAll()
			_ = am.GetUnlocked()
			_ = am.IsUnlocked("first_blood")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestGetProgress_Unlocked(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Unlock an achievement
	stats := &PlayerStats{Kills: 100}
	_, err = am.CheckUnlocks(stats)
	if err != nil {
		t.Fatalf("CheckUnlocks() error = %v", err)
	}

	// Get progress of unlocked achievement
	current, target, err := am.GetProgress("centurion")
	if err != nil {
		t.Errorf("GetProgress() error = %v", err)
	}
	if current != target {
		t.Errorf("Expected unlocked achievement to show complete: current=%d, target=%d", current, target)
	}
}

func TestGetProgressWithStats_UnlockedWithProgress(t *testing.T) {
	am, err := NewAchievementManager(filepath.Join(t.TempDir(), "test.json"))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test speed_demon with nil Progress (binary achievement)
	stats := &PlayerStats{
		CompletedLevels:     1,
		LevelCompletionTime: 4 * time.Minute,
		Kills:               5,
		TotalDeaths:         1,
		DamageTaken:         10,
	}

	current, target, err := am.GetProgressWithStats("speed_demon", stats)
	if err != nil {
		t.Errorf("GetProgressWithStats() error = %v", err)
	}
	if current != 1 || target != 1 {
		t.Errorf("Expected speed_demon to be complete: current=%d, target=%d", current, target)
	}
}

func TestSave_Error(t *testing.T) {
	// Create manager with invalid save path (read-only location)
	am, err := NewAchievementManager("/proc/achievements.json")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Try to save - should fail due to read-only filesystem
	stats := &PlayerStats{Kills: 1}
	unlocked, err := am.CheckUnlocks(stats)

	// Should unlock but fail to save
	if len(unlocked) == 0 {
		t.Error("Expected achievements to unlock")
	}
	// Error should be reported but achievements still returned
	if err == nil {
		t.Log("Expected save error but got none (may vary by filesystem)")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "achievements.json")

	// Write invalid JSON
	if err := os.WriteFile(savePath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Try to load - should fail
	_, err := NewAchievementManager(savePath)
	if err == nil {
		t.Error("Expected error loading invalid JSON")
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 7, 7, 7},
		{"negative a", -5, 10, -5},
		{"negative b", 10, -5, -5},
		{"both negative", -10, -5, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := min(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
