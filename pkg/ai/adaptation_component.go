// Package ai - ECS components for adaptive AI
package ai

// AdaptationComponent tracks AI behavioral adaptation state for an entity.
type AdaptationComponent struct {
	CurrentAdaptation Adaptation
	LastUpdateTime    float64
}

// Type returns the component type identifier.
func (c *AdaptationComponent) Type() string {
	return "AdaptationComponent"
}

// PlayerProfileComponent stores the player behavior profile (singleton on player entity).
type PlayerProfileComponent struct {
	Profile *PlayerBehaviorProfile
}

// Type returns the component type identifier.
func (c *PlayerProfileComponent) Type() string {
	return "PlayerProfileComponent"
}
