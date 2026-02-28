package combat

import (
	"math"
	"testing"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem()
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if s.armorAbsorption != 0.5 {
		t.Errorf("Expected default armor absorption 0.5, got %f", s.armorAbsorption)
	}
	if s.difficulty != 1.0 {
		t.Errorf("Expected default difficulty 1.0, got %f", s.difficulty)
	}
	if s.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got %s", s.genreID)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name       string
		genre      string
		wantAbsorb float64
	}{
		{"fantasy", "fantasy", 0.5},
		{"scifi", "scifi", 0.6},
		{"horror", "horror", 0.4},
		{"cyberpunk", "cyberpunk", 0.55},
		{"postapoc", "postapoc", 0.45},
		{"unknown", "unknown", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem()
			s.SetGenre(tt.genre)
			if s.armorAbsorption != tt.wantAbsorb {
				t.Errorf("SetGenre(%s): absorption = %f, want %f", tt.genre, s.armorAbsorption, tt.wantAbsorb)
			}
			if s.genreID != tt.genre {
				t.Errorf("SetGenre(%s): genreID = %s, want %s", tt.genre, s.genreID, tt.genre)
			}
		})
	}
}

func TestSetDifficulty(t *testing.T) {
	tests := []struct {
		name       string
		difficulty float64
	}{
		{"easy", 0.5},
		{"normal", 1.0},
		{"hard", 1.5},
		{"nightmare", 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem()
			s.SetDifficulty(tt.difficulty)
			if s.difficulty != tt.difficulty {
				t.Errorf("SetDifficulty(%f): got %f", tt.difficulty, s.difficulty)
			}
		})
	}
}

func TestApplyDamage_NoArmor(t *testing.T) {
	s := NewSystem()
	result := s.ApplyDamage(100, 0, 50, 10, 10, 5, 5)

	if result.HealthDamage != 50 {
		t.Errorf("Expected health damage 50, got %f", result.HealthDamage)
	}
	if result.ArmorDamage != 0 {
		t.Errorf("Expected armor damage 0, got %f", result.ArmorDamage)
	}
	if result.Killed {
		t.Error("Expected not killed with 50 damage to 100 health")
	}
}

func TestApplyDamage_WithArmor_PartialAbsorb(t *testing.T) {
	s := NewSystem()
	// Armor absorption is 0.5, so 50 armor absorbs 25 damage
	// 50 damage - 25 absorbed = 25 health damage
	// Armor takes: 25 / 0.5 = 50 damage (depleted)
	result := s.ApplyDamage(100, 50, 50, 10, 10, 5, 5)

	if result.ArmorDamage != 50 {
		t.Errorf("Expected armor damage 50, got %f", result.ArmorDamage)
	}
	if result.HealthDamage != 25 {
		t.Errorf("Expected health damage 25, got %f", result.HealthDamage)
	}
	if result.Killed {
		t.Error("Expected not killed")
	}
}

func TestApplyDamage_WithArmor_FullAbsorb(t *testing.T) {
	s := NewSystem()
	// 100 armor absorbs 50 damage
	// 100 * 0.5 = 50 absorbed >= 20 damage
	result := s.ApplyDamage(100, 100, 20, 10, 10, 5, 5)

	if result.HealthDamage != 0 {
		t.Errorf("Expected health damage 0, got %f", result.HealthDamage)
	}
	if result.ArmorDamage != 40 {
		t.Errorf("Expected armor damage 40, got %f", result.ArmorDamage)
	}
	if result.Killed {
		t.Error("Expected not killed")
	}
}

func TestApplyDamage_Lethal(t *testing.T) {
	s := NewSystem()
	result := s.ApplyDamage(50, 0, 100, 10, 10, 5, 5)

	if !result.Killed {
		t.Error("Expected killed with 100 damage to 50 health")
	}
	if result.HealthDamage != 100 {
		t.Errorf("Expected health damage 100, got %f", result.HealthDamage)
	}
}

func TestApplyDamage_DifficultyScaling(t *testing.T) {
	tests := []struct {
		name           string
		difficulty     float64
		baseDamage     float64
		expectedDamage float64
	}{
		{"easy", 0.5, 100, 50},
		{"normal", 1.0, 100, 100},
		{"hard", 1.5, 100, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem()
			s.SetDifficulty(tt.difficulty)
			result := s.ApplyDamage(200, 0, tt.baseDamage, 0, 0, 0, 0)
			if result.HealthDamage != tt.expectedDamage {
				t.Errorf("Expected %f damage, got %f", tt.expectedDamage, result.HealthDamage)
			}
		})
	}
}

func TestApplyDamage_DirectionCalculation(t *testing.T) {
	s := NewSystem()
	result := s.ApplyDamage(100, 0, 50, 10, 10, 5, 5)

	// Direction from (5,5) to (10,10) should be normalized (0.707, 0.707)
	expectedDirX := 5.0 / math.Sqrt(50)
	expectedDirY := 5.0 / math.Sqrt(50)

	if math.Abs(result.DirectionX-expectedDirX) > 0.01 {
		t.Errorf("Expected directionX ~%f, got %f", expectedDirX, result.DirectionX)
	}
	if math.Abs(result.DirectionY-expectedDirY) > 0.01 {
		t.Errorf("Expected directionY ~%f, got %f", expectedDirY, result.DirectionY)
	}
}

func TestApplyDamage_ZeroDistance(t *testing.T) {
	s := NewSystem()
	result := s.ApplyDamage(100, 0, 50, 5, 5, 5, 5)

	// Direction should be zero when source and target are at same position
	if result.DirectionX != 0 || result.DirectionY != 0 {
		t.Errorf("Expected zero direction for same position, got (%f, %f)", result.DirectionX, result.DirectionY)
	}
}

func TestShouldGib(t *testing.T) {
	s := NewSystem()

	tests := []struct {
		name        string
		finalHealth float64
		wantGib     bool
	}{
		{"alive", 10, false},
		{"just_dead", 0, false},
		{"overkill_small", -10, false},
		{"overkill_medium", -30, false},
		{"overkill_large", -60, true},
		{"overkill_huge", -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.ShouldGib(tt.finalHealth)
			if got != tt.wantGib {
				t.Errorf("ShouldGib(%f) = %v, want %v", tt.finalHealth, got, tt.wantGib)
			}
		})
	}
}

func TestScaleDamage(t *testing.T) {
	tests := []struct {
		name       string
		difficulty float64
		baseDamage float64
		want       float64
	}{
		{"easy", 0.5, 100, 50},
		{"normal", 1.0, 50, 50},
		{"hard", 1.5, 40, 60},
		{"nightmare", 2.0, 25, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem()
			s.SetDifficulty(tt.difficulty)
			got := s.ScaleDamage(tt.baseDamage)
			if got != tt.want {
				t.Errorf("ScaleDamage(%f) = %f, want %f", tt.baseDamage, got, tt.want)
			}
		})
	}
}

func TestGlobalSystemApply(t *testing.T) {
	SetDifficulty(1.0)
	SetGenre("fantasy")

	event := DamageEvent{
		Source:  1,
		Target:  2,
		Amount:  50,
		DmgType: DamagePhysical,
		PosX:    10,
		PosY:    10,
	}

	result := Apply(event)
	// Global Apply doesn't use position data, so result should have defaults
	if result.HealthDamage == 0 && result.ArmorDamage == 0 {
		t.Error("Expected some damage from Apply")
	}
}

func TestDamageTypes(t *testing.T) {
	// Verify all damage type constants are defined
	types := []DamageType{
		DamagePhysical,
		DamageFire,
		DamagePlasma,
		DamageEnergy,
		DamageExplosive,
	}

	for _, dt := range types {
		if dt == "" {
			t.Error("Damage type should not be empty")
		}
	}
}

func TestArmorAbsorptionByGenre(t *testing.T) {
	tests := []struct {
		genre    string
		health   float64
		armor    float64
		damage   float64
		wantDead bool
	}{
		// Fantasy: 0.5 absorption, 50 armor absorbs 25, damage 100-25=75 to health, 50-75=-25 -> dead
		{"fantasy", 50, 50, 100, true},
		// Scifi: 0.6 absorption, 50 armor absorbs 30, damage 100-30=70 to health, 50-70=-20 -> dead
		{"scifi", 50, 50, 100, true},
		// Horror: 0.4 absorption, 50 armor absorbs 20, damage 100-20=80 to health, 50-80=-30 -> dead
		{"horror", 50, 50, 100, true},
		// All should be dead with this damage/armor/health ratio
		{"cyberpunk", 50, 50, 100, true},
		{"postapoc", 50, 50, 100, true},
		// Test survival case: higher armor
		{"fantasy", 100, 100, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			s := NewSystem()
			s.SetGenre(tt.genre)
			result := s.ApplyDamage(tt.health, tt.armor, tt.damage, 0, 0, 0, 0)
			if result.Killed != tt.wantDead {
				t.Errorf("Genre %s: killed=%v, want %v (health=%f, armor=%f, damage=%f, healthDmg=%f)",
					tt.genre, result.Killed, tt.wantDead, tt.health, tt.armor, tt.damage, result.HealthDamage)
			}
		})
	}
}

func TestCombinedArmorAndDifficultyScaling(t *testing.T) {
	s := NewSystem()
	s.SetGenre("fantasy") // 0.5 absorption
	s.SetDifficulty(2.0)  // hard mode: 2x damage

	// 50 base damage * 2.0 difficulty = 100 damage
	// 50 armor absorbs 25 damage, leaving 75 health damage
	// Armor takes 50 damage (depleted)
	result := s.ApplyDamage(100, 50, 50, 0, 0, 0, 0)

	if result.ArmorDamage != 50 {
		t.Errorf("Expected armor damage 50, got %f", result.ArmorDamage)
	}
	if result.HealthDamage != 75 {
		t.Errorf("Expected health damage 75, got %f", result.HealthDamage)
	}
}

func TestEdgeCases(t *testing.T) {
	s := NewSystem()

	t.Run("zero_damage", func(t *testing.T) {
		result := s.ApplyDamage(100, 50, 0, 0, 0, 0, 0)
		if result.HealthDamage != 0 || result.ArmorDamage != 0 {
			t.Error("Zero damage should cause no damage")
		}
		if result.Killed {
			t.Error("Zero damage should not kill")
		}
	})

	t.Run("negative_health", func(t *testing.T) {
		result := s.ApplyDamage(-10, 0, 50, 0, 0, 0, 0)
		if !result.Killed {
			t.Error("Already dead entity should remain dead")
		}
	})

	t.Run("massive_armor", func(t *testing.T) {
		result := s.ApplyDamage(100, 1000, 50, 0, 0, 0, 0)
		if result.HealthDamage != 0 {
			t.Error("Massive armor should absorb all damage")
		}
	})
}
