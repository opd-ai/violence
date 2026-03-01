package economy

import (
	"sync"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}

	// Check base values (updated for balanced economy)
	if cfg.BaseKillReward != 20 {
		t.Errorf("BaseKillReward = %d, want 20", cfg.BaseKillReward)
	}
	if cfg.BaseMissionReward != 100 {
		t.Errorf("BaseMissionReward = %d, want 100", cfg.BaseMissionReward)
	}
	if cfg.BaseObjectiveReward != 50 {
		t.Errorf("BaseObjectiveReward = %d, want 50", cfg.BaseObjectiveReward)
	}

	// Check genre multipliers exist
	if len(cfg.GenreMultipliers) == 0 {
		t.Error("GenreMultipliers is empty")
	}

	// Check difficulty multipliers exist
	if len(cfg.DifficultyMultipliers) == 0 {
		t.Error("DifficultyMultipliers is empty")
	}

	// Check level scaling exists
	if len(cfg.LevelScaling) == 0 {
		t.Error("LevelScaling is empty")
	}
}

func TestCalculateKillReward(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		difficulty  string
		playerLevel int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "fantasy_normal_level1",
			genre:       "fantasy",
			difficulty:  "normal",
			playerLevel: 1,
			wantMin:     19,
			wantMax:     21,
		},
		{
			name:        "horror_hard_level5",
			genre:       "horror",
			difficulty:  "hard",
			playerLevel: 5,
			wantMin:     31,
			wantMax:     33,
		},
		{
			name:        "scifi_easy_level10",
			genre:       "scifi",
			difficulty:  "easy",
			playerLevel: 10,
			wantMin:     25,
			wantMax:     27,
		},
		{
			name:        "cyberpunk_nightmare_level8",
			genre:       "cyberpunk",
			difficulty:  "nightmare",
			playerLevel: 8,
			wantMin:     43,
			wantMax:     45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			reward := cfg.CalculateKillReward(tt.genre, tt.difficulty, tt.playerLevel)

			if reward < tt.wantMin || reward > tt.wantMax {
				t.Errorf("CalculateKillReward() = %d, want between %d and %d",
					reward, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateMissionReward(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		difficulty  string
		playerLevel int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "fantasy_normal_level1",
			genre:       "fantasy",
			difficulty:  "normal",
			playerLevel: 1,
			wantMin:     98,
			wantMax:     102,
		},
		{
			name:        "postapoc_hard_level10",
			genre:       "postapoc",
			difficulty:  "hard",
			playerLevel: 10,
			wantMin:     221,
			wantMax:     225,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			reward := cfg.CalculateMissionReward(tt.genre, tt.difficulty, tt.playerLevel)

			if reward < tt.wantMin || reward > tt.wantMax {
				t.Errorf("CalculateMissionReward() = %d, want between %d and %d",
					reward, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateObjectiveReward(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		name        string
		genre       string
		difficulty  string
		playerLevel int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "scifi_normal_level1",
			genre:       "scifi",
			difficulty:  "normal",
			playerLevel: 1,
			wantMin:     46,
			wantMax:     49,
		},
		{
			name:        "horror_nightmare_level7",
			genre:       "horror",
			difficulty:  "nightmare",
			playerLevel: 7,
			wantMin:     113,
			wantMax:     116,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := cfg.CalculateObjectiveReward(tt.genre, tt.difficulty, tt.playerLevel)

			if reward < tt.wantMin || reward > tt.wantMax {
				t.Errorf("CalculateObjectiveReward() = %d, want between %d and %d",
					reward, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateItemPrice(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		name        string
		genre       string
		playerLevel int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "fantasy_level1",
			genre:       "fantasy",
			playerLevel: 1,
			wantMin:     48,
			wantMax:     52,
		},
		{
			name:        "horror_level10",
			genre:       "horror",
			playerLevel: 10,
			wantMin:     88,
			wantMax:     90,
		},
		{
			name:        "scifi_level5",
			genre:       "scifi",
			playerLevel: 5,
			wantMin:     56,
			wantMax:     58,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := cfg.CalculateItemPrice(tt.genre, tt.playerLevel)

			if price < tt.wantMin || price > tt.wantMax {
				t.Errorf("CalculateItemPrice() = %d, want between %d and %d",
					price, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateWeaponPrice(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		name        string
		genre       string
		playerLevel int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "fantasy_level1",
			genre:       "fantasy",
			playerLevel: 1,
			wantMin:     98,
			wantMax:     102,
		},
		{
			name:        "cyberpunk_level10",
			genre:       "cyberpunk",
			playerLevel: 10,
			wantMin:     168,
			wantMax:     172,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := cfg.CalculateWeaponPrice(tt.genre, tt.playerLevel)

			if price < tt.wantMin || price > tt.wantMax {
				t.Errorf("CalculateWeaponPrice() = %d, want between %d and %d",
					price, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateArmorPrice(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		name        string
		genre       string
		playerLevel int
		wantMin     int
		wantMax     int
	}{
		{
			name:        "fantasy_level1",
			genre:       "fantasy",
			playerLevel: 1,
			wantMin:     78,
			wantMax:     82,
		},
		{
			name:        "postapoc_level8",
			genre:       "postapoc",
			playerLevel: 8,
			wantMin:     120,
			wantMax:     123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := cfg.CalculateArmorPrice(tt.genre, tt.playerLevel)

			if price < tt.wantMin || price > tt.wantMax {
				t.Errorf("CalculateArmorPrice() = %d, want between %d and %d",
					price, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGenreMultipliers(t *testing.T) {
	cfg := NewConfig()

	genres := []string{"horror", "scifi", "fantasy", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			mult, ok := cfg.GenreMultipliers[genre]
			if !ok {
				t.Errorf("genre %s missing from GenreMultipliers", genre)
			}

			if mult < 0.5 || mult > 2.0 {
				t.Errorf("genre %s multiplier %f out of reasonable range", genre, mult)
			}
		})
	}
}

func TestDifficultyMultipliers(t *testing.T) {
	cfg := NewConfig()

	difficulties := []string{"easy", "normal", "hard", "nightmare"}

	for _, diff := range difficulties {
		t.Run(diff, func(t *testing.T) {
			mult, ok := cfg.DifficultyMultipliers[diff]
			if !ok {
				t.Errorf("difficulty %s missing from DifficultyMultipliers", diff)
			}

			if mult < 0.5 || mult > 2.0 {
				t.Errorf("difficulty %s multiplier %f out of reasonable range", diff, mult)
			}
		})
	}

	// Verify ordering: easy < normal < hard < nightmare
	if cfg.DifficultyMultipliers["easy"] >= cfg.DifficultyMultipliers["normal"] {
		t.Error("easy multiplier should be less than normal")
	}
	if cfg.DifficultyMultipliers["normal"] >= cfg.DifficultyMultipliers["hard"] {
		t.Error("normal multiplier should be less than hard")
	}
	if cfg.DifficultyMultipliers["hard"] >= cfg.DifficultyMultipliers["nightmare"] {
		t.Error("hard multiplier should be less than nightmare")
	}
}

func TestLevelScaling(t *testing.T) {
	cfg := NewConfig()

	// Test that level ranges are covered
	testLevels := []int{1, 3, 5, 7, 9, 10, 15}

	for _, level := range testLevels {
		t.Run("level_"+string(rune('0'+level)), func(t *testing.T) {
			mult := cfg.getLevelMultiplier(level)

			if mult < 1.0 || mult > 2.0 {
				t.Errorf("level %d multiplier %f out of range", level, mult)
			}
		})
	}

	// Verify progression: higher levels = higher multipliers
	mult1 := cfg.getLevelMultiplier(1)
	mult5 := cfg.getLevelMultiplier(5)
	mult10 := cfg.getLevelMultiplier(10)

	if mult5 <= mult1 {
		t.Errorf("level 5 multiplier (%f) should be > level 1 (%f)", mult5, mult1)
	}
	if mult10 <= mult5 {
		t.Errorf("level 10 multiplier (%f) should be > level 5 (%f)", mult10, mult5)
	}
}

func TestSetGenreMultiplier(t *testing.T) {
	cfg := NewConfig()

	cfg.SetGenreMultiplier("fantasy", 1.5)

	if cfg.GenreMultipliers["fantasy"] != 1.5 {
		t.Errorf("SetGenreMultiplier failed: got %f, want 1.5", cfg.GenreMultipliers["fantasy"])
	}

	// Verify it affects calculations
	reward := cfg.CalculateKillReward("fantasy", "normal", 1)
	expected := int(20.0 * 1.5 * 1.0 * 1.0)

	if reward != expected {
		t.Errorf("reward after SetGenreMultiplier = %d, want %d", reward, expected)
	}
}

func TestSetDifficultyMultiplier(t *testing.T) {
	cfg := NewConfig()

	cfg.SetDifficultyMultiplier("hard", 2.0)

	if cfg.DifficultyMultipliers["hard"] != 2.0 {
		t.Errorf("SetDifficultyMultiplier failed: got %f, want 2.0", cfg.DifficultyMultipliers["hard"])
	}

	// Verify it affects calculations
	reward := cfg.CalculateKillReward("fantasy", "hard", 1)
	expected := int(20.0 * 1.0 * 2.0 * 1.0)

	if reward != expected {
		t.Errorf("reward after SetDifficultyMultiplier = %d, want %d", reward, expected)
	}
}

func TestUnknownGenre(t *testing.T) {
	cfg := NewConfig()

	// Unknown genre should use 1.0 multiplier
	reward := cfg.CalculateKillReward("unknown", "normal", 1)
	expected := int(20.0 * 1.0 * 1.0 * 1.0)

	if reward != expected {
		t.Errorf("reward for unknown genre = %d, want %d", reward, expected)
	}
}

func TestUnknownDifficulty(t *testing.T) {
	cfg := NewConfig()

	// Unknown difficulty should use 1.0 multiplier
	reward := cfg.CalculateKillReward("fantasy", "unknown", 1)
	expected := int(20.0 * 1.0 * 1.0 * 1.0)

	if reward != expected {
		t.Errorf("reward for unknown difficulty = %d, want %d", reward, expected)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cfg := NewConfig()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = cfg.CalculateKillReward("fantasy", "normal", 1)
			}
		}()
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cfg.SetGenreMultiplier("test", float64(id))
			}
		}(i)
	}

	wg.Wait()
}

func TestRewardProgression(t *testing.T) {
	cfg := NewConfig()

	// Verify that rewards increase with level
	reward1 := cfg.CalculateKillReward("fantasy", "normal", 1)
	reward5 := cfg.CalculateKillReward("fantasy", "normal", 5)
	reward10 := cfg.CalculateKillReward("fantasy", "normal", 10)

	if reward5 <= reward1 {
		t.Errorf("level 5 reward (%d) should be > level 1 (%d)", reward5, reward1)
	}
	if reward10 <= reward5 {
		t.Errorf("level 10 reward (%d) should be > level 5 (%d)", reward10, reward5)
	}
}

func TestPriceProgression(t *testing.T) {
	cfg := NewConfig()

	// Verify that prices increase with level
	price1 := cfg.CalculateWeaponPrice("fantasy", 1)
	price5 := cfg.CalculateWeaponPrice("fantasy", 5)
	price10 := cfg.CalculateWeaponPrice("fantasy", 10)

	if price5 <= price1 {
		t.Errorf("level 5 price (%d) should be > level 1 (%d)", price5, price1)
	}
	if price10 <= price5 {
		t.Errorf("level 10 price (%d) should be > level 5 (%d)", price10, price5)
	}
}

func TestDifficultyRewardScaling(t *testing.T) {
	cfg := NewConfig()

	easyReward := cfg.CalculateKillReward("fantasy", "easy", 1)
	normalReward := cfg.CalculateKillReward("fantasy", "normal", 1)
	hardReward := cfg.CalculateKillReward("fantasy", "hard", 1)
	nightmareReward := cfg.CalculateKillReward("fantasy", "nightmare", 1)

	if easyReward >= normalReward {
		t.Errorf("easy reward (%d) should be < normal (%d)", easyReward, normalReward)
	}
	if normalReward >= hardReward {
		t.Errorf("normal reward (%d) should be < hard (%d)", normalReward, hardReward)
	}
	if hardReward >= nightmareReward {
		t.Errorf("hard reward (%d) should be < nightmare (%d)", hardReward, nightmareReward)
	}
}

func BenchmarkCalculateKillReward(b *testing.B) {
	cfg := NewConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.CalculateKillReward("fantasy", "normal", 5)
	}
}

func BenchmarkCalculateMissionReward(b *testing.B) {
	cfg := NewConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.CalculateMissionReward("cyberpunk", "hard", 10)
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	cfg := NewConfig()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cfg.CalculateKillReward("fantasy", "normal", 1)
		}
	})
}
