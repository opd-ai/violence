package common

// Game timing constants for consistent frame-based calculations across all systems.
const (
	// TargetFPS is the target frame rate for the game.
	TargetFPS = 60

	// DeltaTime represents the time delta per frame at 60 FPS (1/60 seconds).
	// Use this constant instead of hardcoding 1.0/60.0 throughout the codebase.
	DeltaTime = 1.0 / float64(TargetFPS)

	// FrameDuration is the duration of a single frame in seconds (same as DeltaTime).
	FrameDuration = DeltaTime
)

// Common gameplay constants used across multiple packages.
const (
	// DefaultMapSize is the standard map dimension used for level generation.
	DefaultMapSize = 64

	// DefaultCellSize is the standard spatial grid cell size for collision detection.
	DefaultCellSize = 64.0

	// MaxParticles is the default particle pool capacity.
	MaxParticles = 1024

	// MaxCorpses is the default corpse pool capacity.
	MaxCorpses = 200

	// MaxDecals is the default combat decal pool capacity.
	MaxDecals = 500

	// MaxSquadMembers is the maximum number of squad companions.
	MaxSquadMembers = 3
)

// Animation timing constants.
const (
	// DefaultAnimationSpeed is the standard animation frames per second.
	DefaultAnimationSpeed = 8.0

	// FlickerFrequencyBase is the base frequency for flame/torch flicker effects.
	FlickerFrequencyBase = 3.0

	// PulseFrequency is the standard pulse rate for glowing effects.
	PulseFrequency = 2.0
)

// Combat constants shared across combat-related packages.
const (
	// DefaultDamageMultiplier is the baseline damage multiplier.
	DefaultDamageMultiplier = 1.0

	// CriticalHitMultiplier is the default critical hit damage multiplier.
	CriticalHitMultiplier = 1.5

	// BackstabMultiplier is the damage multiplier for backstab attacks.
	BackstabMultiplier = 2.0

	// FlankMultiplier is the damage multiplier for flank attacks.
	FlankMultiplier = 1.25

	// HeadshotMultiplier is the damage multiplier for headshots.
	HeadshotMultiplier = 2.0
)

// Visual effect constants.
const (
	// DefaultAlpha is the standard full opacity value (0-255 scale).
	DefaultAlpha = 255

	// HalfAlpha is the standard 50% opacity value.
	HalfAlpha = 128

	// QuarterAlpha is the standard 25% opacity value.
	QuarterAlpha = 64

	// FadeInDuration is the standard fade-in animation duration in seconds.
	FadeInDuration = 0.3

	// FadeOutDuration is the standard fade-out animation duration in seconds.
	FadeOutDuration = 0.5
)

// Physics constants.
const (
	// Gravity is the standard gravity acceleration (units per second squared).
	Gravity = 9.8

	// DefaultFriction is the standard friction coefficient.
	DefaultFriction = 0.8

	// DefaultBounciness is the standard bounce/restitution coefficient.
	DefaultBounciness = 0.3
)

// Color math constants for normalized color calculations.
const (
	// ColorMaxValue is the maximum value for 8-bit color channels (255).
	ColorMaxValue = 255.0

	// ColorMaxValueInt is the integer version of ColorMaxValue.
	ColorMaxValueInt = 255
)

// NormalizeColor converts a color channel value (0-255) to normalized (0.0-1.0).
func NormalizeColor(c uint8) float64 {
	return float64(c) / ColorMaxValue
}

// DenormalizeColor converts a normalized color value (0.0-1.0) to 8-bit (0-255).
func DenormalizeColor(c float64) uint8 {
	if c < 0 {
		return 0
	}
	if c > 1.0 {
		return ColorMaxValueInt
	}
	return uint8(c * ColorMaxValue)
}
