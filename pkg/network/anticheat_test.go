package network

import (
	"testing"
	"time"
)

func TestValidateMovement(t *testing.T) {
	tests := []struct {
		name         string
		oldPos       Vec2
		newPos       Vec2
		deltaTime    float64
		isSprinting  bool
		wantValid    bool
		wantSeverity int
	}{
		{
			name:        "normal walking speed",
			oldPos:      Vec2{0, 0},
			newPos:      Vec2{4, 0},
			deltaTime:   1.0,
			isSprinting: false,
			wantValid:   true,
		},
		{
			name:        "sprint speed",
			oldPos:      Vec2{0, 0},
			newPos:      Vec2{10, 0},
			deltaTime:   1.0,
			isSprinting: true,
			wantValid:   true,
		},
		{
			name:         "speed hack detected",
			oldPos:       Vec2{0, 0},
			newPos:       Vec2{30, 0},
			deltaTime:    1.0,
			isSprinting:  true,
			wantValid:    false,
			wantSeverity: SeverityKick,
		},
		{
			name:         "suspicious but not kick-worthy",
			oldPos:       Vec2{0, 0},
			newPos:       Vec2{15, 0},
			deltaTime:    1.0,
			isSprinting:  false,
			wantValid:    false,
			wantSeverity: SeverityWarning,
		},
		{
			name:         "zero delta time",
			oldPos:       Vec2{0, 0},
			newPos:       Vec2{5, 0},
			deltaTime:    0,
			isSprinting:  false,
			wantValid:    false,
			wantSeverity: SeverityWarning,
		},
		{
			name:        "small movement",
			oldPos:      Vec2{0, 0},
			newPos:      Vec2{0.1, 0},
			deltaTime:   1.0,
			isSprinting: false,
			wantValid:   true,
		},
		{
			name:        "diagonal movement",
			oldPos:      Vec2{0, 0},
			newPos:      Vec2{6, 8},
			deltaTime:   1.0,
			isSprinting: true,
			wantValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMovement(tt.oldPos, tt.newPos, tt.deltaTime, tt.isSprinting)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateMovement() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if !tt.wantValid && result.Severity != tt.wantSeverity {
				t.Errorf("ValidateMovement() severity = %v, want %v", result.Severity, tt.wantSeverity)
			}

			if !result.Valid && result.Violation == "" {
				t.Error("ValidateMovement() violation message should not be empty")
			}
		})
	}
}

func TestValidateDamage(t *testing.T) {
	pistol := WeaponDefinition{
		ID:           1,
		Name:         "Pistol",
		BaseDamage:   10,
		MaxRange:     50.0,
		HeadshotMult: 2.0,
		IsHitscan:    true,
	}

	shotgun := WeaponDefinition{
		ID:           2,
		Name:         "Shotgun",
		BaseDamage:   50,
		MaxRange:     15.0,
		HeadshotMult: 1.5,
		IsHitscan:    false,
	}

	tests := []struct {
		name         string
		weapon       WeaponDefinition
		damage       int
		distance     float64
		isHeadshot   bool
		wantValid    bool
		wantSeverity int
	}{
		{
			name:       "valid pistol body shot",
			weapon:     pistol,
			damage:     10,
			distance:   20.0,
			isHeadshot: false,
			wantValid:  true,
		},
		{
			name:       "valid pistol headshot",
			weapon:     pistol,
			damage:     20,
			distance:   20.0,
			isHeadshot: true,
			wantValid:  true,
		},
		{
			name:         "damage too high",
			weapon:       pistol,
			damage:       50,
			distance:     20.0,
			isHeadshot:   false,
			wantValid:    false,
			wantSeverity: SeverityKick,
		},
		{
			name:         "damage too low",
			weapon:       pistol,
			damage:       1,
			distance:     20.0,
			isHeadshot:   false,
			wantValid:    false,
			wantSeverity: SeverityKick,
		},
		{
			name:       "within tolerance",
			weapon:     pistol,
			damage:     11,
			distance:   20.0,
			isHeadshot: false,
			wantValid:  true,
		},
		{
			name:         "unreasonable distance",
			weapon:       pistol,
			damage:       10,
			distance:     150.0,
			isHeadshot:   false,
			wantValid:    false,
			wantSeverity: SeverityWarning,
		},
		{
			name:       "shotgun close range",
			weapon:     shotgun,
			damage:     40, // Slight falloff expected
			distance:   5.0,
			isHeadshot: false,
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDamage(tt.weapon, tt.damage, tt.distance, tt.isHeadshot)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateDamage() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if !tt.wantValid && result.Severity != tt.wantSeverity {
				t.Errorf("ValidateDamage() severity = %v, want %v", result.Severity, tt.wantSeverity)
			}
		})
	}
}

func TestValidateFireRate(t *testing.T) {
	pistol := WeaponDefinition{
		ID:          1,
		MaxFireRate: 5.0, // 5 shots per second
	}

	machineGun := WeaponDefinition{
		ID:          2,
		MaxFireRate: 10.0, // 10 shots per second
	}

	tests := []struct {
		name         string
		weapon       WeaponDefinition
		shotCount    int
		interval     time.Duration
		wantValid    bool
		wantSeverity int
	}{
		{
			name:      "normal fire rate",
			weapon:    pistol,
			shotCount: 3,
			interval:  200 * time.Millisecond,
			wantValid: true,
		},
		{
			name:         "rapid fire hack",
			weapon:       pistol,
			shotCount:    10,
			interval:     50 * time.Millisecond,
			wantValid:    false,
			wantSeverity: SeverityKick,
		},
		{
			name:      "machine gun normal",
			weapon:    machineGun,
			shotCount: 8,
			interval:  100 * time.Millisecond,
			wantValid: true,
		},
		{
			name:      "single shot",
			weapon:    pistol,
			shotCount: 1,
			interval:  0,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &AntiCheatStats{}
			baseTime := time.Now()
			var lastResult ValidationResult

			for i := 0; i < tt.shotCount; i++ {
				shotTime := baseTime.Add(time.Duration(i) * tt.interval)
				lastResult = ValidateFireRate(stats, tt.weapon, shotTime)
			}

			if lastResult.Valid != tt.wantValid {
				t.Errorf("ValidateFireRate() valid = %v, want %v", lastResult.Valid, tt.wantValid)
			}

			if !tt.wantValid && lastResult.Severity != tt.wantSeverity {
				t.Errorf("ValidateFireRate() severity = %v, want %v", lastResult.Severity, tt.wantSeverity)
			}
		})
	}
}

func TestCheckStatisticalAnomaly(t *testing.T) {
	tests := []struct {
		name         string
		stats        *AntiCheatStats
		wantValid    bool
		wantSeverity int
	}{
		{
			name: "normal stats",
			stats: &AntiCheatStats{
				TotalShots: 100,
				Headshots:  30,
				BodyShots:  50,
				Misses:     20,
			},
			wantValid: true,
		},
		{
			name: "suspicious headshot ratio",
			stats: &AntiCheatStats{
				TotalShots: 100,
				Headshots:  85,
				BodyShots:  10,
				Misses:     5,
			},
			wantValid:    false,
			wantSeverity: SeverityWarning,
		},
		{
			name: "perfect aim (aimbot suspected)",
			stats: &AntiCheatStats{
				TotalShots: 100,
				Headshots:  100,
				BodyShots:  0,
				Misses:     0,
			},
			wantValid:    false,
			wantSeverity: SeverityWarning,
		},
		{
			name: "insufficient sample size",
			stats: &AntiCheatStats{
				TotalShots: 20,
				Headshots:  20,
				BodyShots:  0,
				Misses:     0,
			},
			wantValid: true,
		},
		{
			name:      "nil stats",
			stats:     nil,
			wantValid: false,
		},
		{
			name: "zero shots",
			stats: &AntiCheatStats{
				TotalShots: 0,
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckStatisticalAnomaly(tt.stats)

			if result.Valid != tt.wantValid {
				t.Errorf("CheckStatisticalAnomaly() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if !tt.wantValid && tt.stats != nil && result.Severity != tt.wantSeverity {
				t.Errorf("CheckStatisticalAnomaly() severity = %v, want %v", result.Severity, tt.wantSeverity)
			}
		})
	}
}

func TestRecordShot(t *testing.T) {
	tests := []struct {
		name          string
		shots         []struct{ hit, headshot bool }
		wantTotal     int
		wantHeadshots int
		wantBodyShots int
		wantMisses    int
	}{
		{
			name: "mixed shots",
			shots: []struct{ hit, headshot bool }{
				{true, true},   // Headshot
				{true, false},  // Body shot
				{false, false}, // Miss
				{true, true},   // Headshot
			},
			wantTotal:     4,
			wantHeadshots: 2,
			wantBodyShots: 1,
			wantMisses:    1,
		},
		{
			name: "all headshots",
			shots: []struct{ hit, headshot bool }{
				{true, true},
				{true, true},
				{true, true},
			},
			wantTotal:     3,
			wantHeadshots: 3,
			wantBodyShots: 0,
			wantMisses:    0,
		},
		{
			name: "all misses",
			shots: []struct{ hit, headshot bool }{
				{false, false},
				{false, false},
			},
			wantTotal:     2,
			wantHeadshots: 0,
			wantBodyShots: 0,
			wantMisses:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &AntiCheatStats{}
			shotTime := time.Now()

			for _, shot := range tt.shots {
				RecordShot(stats, shot.hit, shot.headshot, shotTime)
			}

			if stats.TotalShots != tt.wantTotal {
				t.Errorf("RecordShot() total = %v, want %v", stats.TotalShots, tt.wantTotal)
			}
			if stats.Headshots != tt.wantHeadshots {
				t.Errorf("RecordShot() headshots = %v, want %v", stats.Headshots, tt.wantHeadshots)
			}
			if stats.BodyShots != tt.wantBodyShots {
				t.Errorf("RecordShot() bodyshots = %v, want %v", stats.BodyShots, tt.wantBodyShots)
			}
			if stats.Misses != tt.wantMisses {
				t.Errorf("RecordShot() misses = %v, want %v", stats.Misses, tt.wantMisses)
			}
		})
	}
}

func TestRecordViolation(t *testing.T) {
	stats := &AntiCheatStats{}

	RecordViolation(stats, SeverityWarning)
	if stats.SuspicionScore != SeverityWarning {
		t.Errorf("RecordViolation() score = %v, want %v", stats.SuspicionScore, SeverityWarning)
	}

	RecordViolation(stats, SeverityKick)
	expectedScore := SeverityWarning + SeverityKick
	if stats.SuspicionScore != expectedScore {
		t.Errorf("RecordViolation() score = %v, want %v", stats.SuspicionScore, expectedScore)
	}

	// Test nil safety
	RecordViolation(nil, SeverityBan)
}

func TestVec2Distance(t *testing.T) {
	tests := []struct {
		name string
		v1   Vec2
		v2   Vec2
		want float64
	}{
		{
			name: "horizontal distance",
			v1:   Vec2{0, 0},
			v2:   Vec2{3, 0},
			want: 3.0,
		},
		{
			name: "vertical distance",
			v1:   Vec2{0, 0},
			v2:   Vec2{0, 4},
			want: 4.0,
		},
		{
			name: "diagonal distance (3-4-5 triangle)",
			v1:   Vec2{0, 0},
			v2:   Vec2{3, 4},
			want: 5.0,
		},
		{
			name: "same position",
			v1:   Vec2{5, 5},
			v2:   Vec2{5, 5},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Distance(tt.v2)
			if got != tt.want {
				t.Errorf("Distance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAntiCheatConstants(t *testing.T) {
	// Verify constants are sensible
	if MaxPlayerSpeed >= MaxSprintSpeed {
		t.Error("MaxPlayerSpeed should be less than MaxSprintSpeed")
	}

	if SpeedHackThreshold <= MaxSprintSpeed {
		t.Error("SpeedHackThreshold should be greater than MaxSprintSpeed")
	}

	if MaxHeadshotRatio > 1.0 || MaxHeadshotRatio < 0.5 {
		t.Errorf("MaxHeadshotRatio = %v, expected between 0.5 and 1.0", MaxHeadshotRatio)
	}

	if MinSampleSize < 10 {
		t.Errorf("MinSampleSize = %v, expected at least 10", MinSampleSize)
	}
}

// BenchmarkValidateMovement measures movement validation performance
func BenchmarkValidateMovement(b *testing.B) {
	oldPos := Vec2{0, 0}
	newPos := Vec2{5, 5}
	deltaTime := 1.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateMovement(oldPos, newPos, deltaTime, false)
	}
}

// BenchmarkValidateDamage measures damage validation performance
func BenchmarkValidateDamage(b *testing.B) {
	weapon := WeaponDefinition{
		ID:           1,
		BaseDamage:   10,
		MaxRange:     50.0,
		HeadshotMult: 2.0,
		IsHitscan:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDamage(weapon, 10, 20.0, false)
	}
}
