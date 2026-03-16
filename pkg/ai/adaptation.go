// Package ai - Adaptive AI that learns from player behavior
package ai

import (
	"math"
)

// PlayerTactic represents a recognized player behavior pattern.
type PlayerTactic int

const (
	TacticUnknown    PlayerTactic = iota // TacticUnknown is an unrecognized tactic.
	TacticRushMelee                      // TacticRushMelee is aggressive melee rushing.
	TacticKiteRanged                     // TacticKiteRanged is kiting while using ranged.
	TacticCoverBased                     // TacticCoverBased is using cover strategically.
	TacticStealthy                       // TacticStealthy is stealthy approach.
	TacticExplosives                     // TacticExplosives is heavy explosive use.
	TacticHitAndRun                      // TacticHitAndRun is hit and run tactics.
)

// TacticObservation records a single observation of player behavior.
type TacticObservation struct {
	Tactic     PlayerTactic
	Timestamp  float64
	Confidence float64
}

// PlayerBehaviorProfile tracks aggregated player tactics over time.
type PlayerBehaviorProfile struct {
	// Tactic frequencies (0.0 to 1.0)
	MeleeFrequency     float64
	RangedFrequency    float64
	ExplosiveFrequency float64
	StealthFrequency   float64

	// Positional patterns
	AverageEngagementRange float64
	PrefersCover           float64
	FlankingAttempts       int

	// Temporal patterns
	AverageEngagementDuration float64
	HitAndRunCount            int

	// Observation count for confidence
	ObservationCount int
	LastUpdateTime   float64
}

// NewPlayerBehaviorProfile creates a new behavior profile.
func NewPlayerBehaviorProfile() *PlayerBehaviorProfile {
	return &PlayerBehaviorProfile{
		AverageEngagementRange: 10.0, // Default assumption
	}
}

// RecordObservation updates the profile with a new observation.
func (p *PlayerBehaviorProfile) RecordObservation(obs TacticObservation) {
	weight := 0.1 // New observations have 10% weight to allow gradual adaptation

	switch obs.Tactic {
	case TacticRushMelee:
		p.MeleeFrequency = lerp(p.MeleeFrequency, 1.0, weight)
		p.RangedFrequency = lerp(p.RangedFrequency, 0.0, weight)
	case TacticKiteRanged:
		p.RangedFrequency = lerp(p.RangedFrequency, 1.0, weight)
		p.MeleeFrequency = lerp(p.MeleeFrequency, 0.0, weight)
	case TacticCoverBased:
		p.PrefersCover = lerp(p.PrefersCover, 1.0, weight)
	case TacticStealthy:
		p.StealthFrequency = lerp(p.StealthFrequency, 1.0, weight)
	case TacticExplosives:
		p.ExplosiveFrequency = lerp(p.ExplosiveFrequency, 1.0, weight)
	case TacticHitAndRun:
		p.HitAndRunCount++
	}

	p.ObservationCount++
	p.LastUpdateTime = obs.Timestamp
}

// UpdateRange records a player engagement range observation.
func (p *PlayerBehaviorProfile) UpdateRange(distance float64) {
	if p.ObservationCount == 0 {
		p.AverageEngagementRange = distance
	} else {
		// Exponential moving average
		alpha := 0.15
		p.AverageEngagementRange = lerp(p.AverageEngagementRange, distance, alpha)
	}
}

// RecordFlankAttempt increments flanking counter.
func (p *PlayerBehaviorProfile) RecordFlankAttempt() {
	p.FlankingAttempts++
}

// GetDominantTactic returns the most frequently observed player tactic.
func (p *PlayerBehaviorProfile) GetDominantTactic() PlayerTactic {
	if p.ObservationCount < 3 {
		return TacticUnknown
	}

	maxFreq := 0.0
	dominant := TacticUnknown

	if p.MeleeFrequency > maxFreq {
		maxFreq = p.MeleeFrequency
		dominant = TacticRushMelee
	}
	if p.RangedFrequency > maxFreq {
		maxFreq = p.RangedFrequency
		dominant = TacticKiteRanged
	}
	if p.ExplosiveFrequency > maxFreq {
		maxFreq = p.ExplosiveFrequency
		dominant = TacticExplosives
	}
	if p.StealthFrequency > maxFreq {
		maxFreq = p.StealthFrequency
		dominant = TacticStealthy
	}
	if p.PrefersCover > 0.6 {
		dominant = TacticCoverBased
	}
	if p.HitAndRunCount > 5 && p.ObservationCount > 10 {
		dominant = TacticHitAndRun
	}

	return dominant
}

// AIAdaptation represents behavioral adjustments based on player profile.
type AIAdaptation struct {
	// Spacing adjustments
	PreferredRangeMultiplier float64
	SpreadFormation          float64

	// Aggression modifiers
	RetreatThreshold  float64
	PursuitAggression float64

	// Tactical counters
	UseCover          bool
	FocusFirePriority float64
	FlankingPriority  float64

	// Defensive adjustments
	DodgeFrequency    float64
	AlertnessModifier float64
}

// ComputeAdaptation generates AI behavioral adjustments for a player profile.
func ComputeAdaptation(profile *PlayerBehaviorProfile) AIAdaptation {
	adapt := AIAdaptation{
		PreferredRangeMultiplier: 1.0,
		SpreadFormation:          0.5,
		RetreatThreshold:         0.3,
		PursuitAggression:        0.5,
		UseCover:                 false,
		FocusFirePriority:        0.5,
		FlankingPriority:         0.5,
		DodgeFrequency:           0.3,
		AlertnessModifier:        1.0,
	}

	dominant := profile.GetDominantTactic()

	// Adapt counters based on dominant player tactic
	switch dominant {
	case TacticRushMelee:
		// Counter melee rushers with kiting and spread formation
		adapt.PreferredRangeMultiplier = 1.5
		adapt.SpreadFormation = 0.8
		adapt.RetreatThreshold = 0.5
		adapt.DodgeFrequency = 0.6

	case TacticKiteRanged:
		// Counter kiting with pursuit and focus fire
		adapt.PursuitAggression = 0.8
		adapt.FocusFirePriority = 0.8
		adapt.FlankingPriority = 0.7
		adapt.UseCover = true

	case TacticCoverBased:
		// Counter cover users with flanking and explosives
		adapt.FlankingPriority = 0.9
		adapt.SpreadFormation = 0.7
		adapt.AlertnessModifier = 1.3

	case TacticStealthy:
		// Counter stealth with increased alertness and patrols
		adapt.AlertnessModifier = 1.5
		adapt.SpreadFormation = 0.3
		adapt.FocusFirePriority = 0.9

	case TacticExplosives:
		// Counter explosives with spread formation
		adapt.SpreadFormation = 0.9
		adapt.PreferredRangeMultiplier = 0.7
		adapt.DodgeFrequency = 0.7

	case TacticHitAndRun:
		// Counter hit-and-run with persistent pursuit
		adapt.PursuitAggression = 0.9
		adapt.RetreatThreshold = 0.2
		adapt.AlertnessModifier = 1.2
	}

	// Fine-tune based on engagement range
	if profile.AverageEngagementRange < 5.0 {
		// Player fights close - enemies spread and use range
		adapt.PreferredRangeMultiplier *= 1.3
		adapt.SpreadFormation = math.Min(1.0, adapt.SpreadFormation+0.2)
	} else if profile.AverageEngagementRange > 15.0 {
		// Player fights far - enemies close distance aggressively
		adapt.PreferredRangeMultiplier *= 0.7
		adapt.PursuitAggression = math.Min(1.0, adapt.PursuitAggression+0.2)
	}

	// Adjust for flanking tendency
	if profile.FlankingAttempts > 10 {
		// Player flanks - enemies watch their backs
		adapt.AlertnessModifier *= 1.2
		adapt.SpreadFormation = math.Min(1.0, adapt.SpreadFormation+0.1)
	}

	return adapt
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}
