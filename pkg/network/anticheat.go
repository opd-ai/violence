// Package network provides anti-cheat validation for server-side game state.
package network

import (
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// Violation severity levels
const (
	SeverityWarning = 1 // Log warning
	SeverityKick    = 2 // Kick from match
	SeverityBan     = 3 // Ban from server
)

// Game physics constants
const (
	MaxPlayerSpeed        = 8.0   // Units per second (normal movement)
	MaxSprintSpeed        = 12.0  // Units per second (sprint)
	SpeedHackThreshold    = 24.0  // 2x sprint speed triggers kick
	MaxShotsPerSecond     = 20.0  // Rate limit for shooting
	MaxHeadshotRatio      = 0.80  // 80% headshot ratio triggers review
	MinSampleSize         = 50    // Minimum shots before checking headshot ratio
	DamageTolerancePct    = 0.10  // 10% tolerance for damage validation
	MaxReasonableDistance = 100.0 // Max distance for hitscan weapons
)

// ValidationResult describes the outcome of a validation check.
type ValidationResult struct {
	Valid     bool
	Violation string
	Severity  int
}

// Vec2 represents a 2D position vector.
type Vec2 struct {
	X float64
	Y float64
}

// Distance calculates Euclidean distance between two positions.
func (v Vec2) Distance(other Vec2) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// AntiCheatStats tracks player statistics for anomaly detection.
type AntiCheatStats struct {
	TotalShots      int
	Headshots       int
	BodyShots       int
	Misses          int
	DamageDealt     int64
	SuspicionScore  int // Cumulative violations
	LastShotTime    time.Time
	RecentShotTimes []time.Time // Last 20 shots for rate limiting
}

// WeaponDefinition defines expected weapon behavior.
type WeaponDefinition struct {
	ID           int
	Name         string
	BaseDamage   int
	MaxRange     float64
	HeadshotMult float64
	IsHitscan    bool
	MaxFireRate  float64 // Shots per second
}

// ValidateMovement checks if player movement is within acceptable bounds.
// Prevents speed hacks by validating distance traveled vs. delta time.
func ValidateMovement(oldPos, newPos Vec2, deltaTime float64, isSprinting bool) ValidationResult {
	if deltaTime <= 0 {
		return ValidationResult{
			Valid:     false,
			Violation: "invalid deltaTime",
			Severity:  SeverityWarning,
		}
	}

	distance := oldPos.Distance(newPos)
	speed := distance / deltaTime

	maxAllowed := MaxPlayerSpeed
	if isSprinting {
		maxAllowed = MaxSprintSpeed
	}

	if speed > SpeedHackThreshold {
		logrus.WithFields(logrus.Fields{
			"speed":      speed,
			"max_sprint": MaxSprintSpeed,
			"threshold":  SpeedHackThreshold,
			"distance":   distance,
			"deltaTime":  deltaTime,
		}).Warn("speed hack detected")

		return ValidationResult{
			Valid:     false,
			Violation: "speed exceeds threshold (possible speed hack)",
			Severity:  SeverityKick,
		}
	}

	if speed > maxAllowed*1.5 {
		return ValidationResult{
			Valid:     false,
			Violation: "suspicious movement speed",
			Severity:  SeverityWarning,
		}
	}

	return ValidationResult{Valid: true}
}

// ValidateDamage checks if damage output matches weapon definition.
// Prevents damage hacks by validating against weapon stats.
func ValidateDamage(weapon WeaponDefinition, damage int, distance float64, isHeadshot bool) ValidationResult {
	if weapon.BaseDamage <= 0 {
		return ValidationResult{
			Valid:     false,
			Violation: "invalid weapon definition",
			Severity:  SeverityWarning,
		}
	}

	expectedDamage := weapon.BaseDamage
	if isHeadshot {
		expectedDamage = int(float64(weapon.BaseDamage) * weapon.HeadshotMult)
	}

	// Apply range falloff for projectile weapons
	if !weapon.IsHitscan && weapon.MaxRange > 0 {
		if distance > weapon.MaxRange {
			expectedDamage = 0
		} else {
			falloff := 1.0 - (distance / weapon.MaxRange * 0.5)
			expectedDamage = int(float64(expectedDamage) * falloff)
		}
	}

	// Validate distance
	if weapon.IsHitscan && distance > MaxReasonableDistance {
		return ValidationResult{
			Valid:     false,
			Violation: "shot distance exceeds reasonable range",
			Severity:  SeverityWarning,
		}
	}

	// Allow 10% tolerance for rounding
	tolerance := int(float64(expectedDamage) * DamageTolerancePct)
	minDamage := expectedDamage - tolerance
	maxDamage := expectedDamage + tolerance

	if damage < minDamage || damage > maxDamage {
		logrus.WithFields(logrus.Fields{
			"weapon_id":       weapon.ID,
			"actual_damage":   damage,
			"expected_damage": expectedDamage,
			"min_allowed":     minDamage,
			"max_allowed":     maxDamage,
			"distance":        distance,
			"is_headshot":     isHeadshot,
		}).Warn("damage mismatch detected")

		return ValidationResult{
			Valid:     false,
			Violation: "damage does not match weapon definition",
			Severity:  SeverityKick,
		}
	}

	return ValidationResult{Valid: true}
}

// ValidateFireRate checks if shooting rate is within weapon limits.
// Prevents rapid-fire hacks by tracking shot timestamps.
func ValidateFireRate(stats *AntiCheatStats, weapon WeaponDefinition, shotTime time.Time) ValidationResult {
	if stats == nil {
		return ValidationResult{Valid: false, Violation: "nil stats", Severity: SeverityWarning}
	}

	// Initialize shot tracking
	if stats.RecentShotTimes == nil {
		stats.RecentShotTimes = make([]time.Time, 0, 20)
	}

	// Add current shot
	stats.RecentShotTimes = append(stats.RecentShotTimes, shotTime)
	if len(stats.RecentShotTimes) > 20 {
		stats.RecentShotTimes = stats.RecentShotTimes[1:]
	}

	// Need at least 2 shots to check rate
	if len(stats.RecentShotTimes) < 2 {
		return ValidationResult{Valid: true}
	}

	// Check shots in last second
	oneSecondAgo := shotTime.Add(-1 * time.Second)
	recentShots := 0
	for _, t := range stats.RecentShotTimes {
		if t.After(oneSecondAgo) {
			recentShots++
		}
	}

	maxRate := MaxShotsPerSecond
	if weapon.MaxFireRate > 0 && weapon.MaxFireRate < MaxShotsPerSecond {
		maxRate = weapon.MaxFireRate
	}

	if float64(recentShots) > maxRate {
		logrus.WithFields(logrus.Fields{
			"weapon_id":     weapon.ID,
			"shots_per_sec": recentShots,
			"max_allowed":   maxRate,
			"recent_shots":  len(stats.RecentShotTimes),
		}).Warn("fire rate exceeded")

		return ValidationResult{
			Valid:     false,
			Violation: "fire rate exceeds weapon maximum",
			Severity:  SeverityKick,
		}
	}

	return ValidationResult{Valid: true}
}

// CheckStatisticalAnomaly detects suspicious player statistics.
// Reviews headshot ratio and other behavioral patterns.
func CheckStatisticalAnomaly(stats *AntiCheatStats) ValidationResult {
	if stats == nil {
		return ValidationResult{Valid: false, Violation: "nil stats", Severity: SeverityWarning}
	}

	// Need minimum sample size
	if stats.TotalShots < MinSampleSize {
		return ValidationResult{Valid: true}
	}

	// Check headshot ratio
	headshotRatio := float64(stats.Headshots) / float64(stats.TotalShots)
	if headshotRatio > MaxHeadshotRatio {
		logrus.WithFields(logrus.Fields{
			"headshot_ratio": headshotRatio,
			"headshots":      stats.Headshots,
			"total_shots":    stats.TotalShots,
			"max_allowed":    MaxHeadshotRatio,
		}).Warn("suspicious headshot ratio")

		return ValidationResult{
			Valid:     false,
			Violation: "headshot ratio exceeds threshold (possible aimbot)",
			Severity:  SeverityWarning, // Warning only - requires review
		}
	}

	return ValidationResult{Valid: true}
}

// RecordShot updates player statistics after a shot.
func RecordShot(stats *AntiCheatStats, hit, isHeadshot bool, shotTime time.Time) {
	if stats == nil {
		return
	}

	stats.TotalShots++
	stats.LastShotTime = shotTime

	if !hit {
		stats.Misses++
		return
	}

	if isHeadshot {
		stats.Headshots++
	} else {
		stats.BodyShots++
	}
}

// RecordViolation increments suspicion score for tracking repeat offenders.
func RecordViolation(stats *AntiCheatStats, severity int) {
	if stats == nil {
		return
	}
	stats.SuspicionScore += severity
}
