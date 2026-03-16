// Package trap provides interactive trap mechanics with pressure plates,
// tripwires, dart launchers, and other triggered hazards that require player
// awareness and skill to detect and disarm.
package trap

import (
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// TrapType represents the kind of trap mechanism.
type TrapType int

const (
	// TrapTypePressurePlate is a floor-triggered pressure plate.
	TrapTypePressurePlate TrapType = iota
	// TrapTypeTripwire is a wire that triggers when crossed.
	TrapTypeTripwire
	// TrapTypeProximity triggers when entities approach.
	TrapTypeProximity
	// TrapTypeButton is manually activated.
	TrapTypeButton
	// TrapTypeLever is a two-state switch mechanism.
	TrapTypeLever

	// TrapTypeDartWall shoots darts from the wall.
	TrapTypeDartWall
	// TrapTypeArrowSlit fires arrows through a slot.
	TrapTypeArrowSlit
	// TrapTypeFlameThrower emits a stream of fire.
	TrapTypeFlameThrower
	// TrapTypeSpikePit drops victims onto spikes.
	TrapTypeSpikePit
	// TrapTypeSwingingBlade is a pendulum blade trap.
	TrapTypeSwingingBlade
	// TrapTypeRollingBoulder releases a crushing boulder.
	TrapTypeRollingBoulder
	// TrapTypeElectricShock delivers an electric discharge.
	TrapTypeElectricShock
	// TrapTypePoisonDart shoots a poisoned dart.
	TrapTypePoisonDart
	// TrapTypeNetCatcher ensnares targets in a net.
	TrapTypeNetCatcher
	// TrapTypeBearTrap is a spring-loaded jaw trap.
	TrapTypeBearTrap
	// TrapTypeExplosive detonates on activation.
	TrapTypeExplosive
	// TrapTypeTeleporter transports victims elsewhere.
	TrapTypeTeleporter
	// TrapTypeIllusionWall hides a dangerous passage.
	TrapTypeIllusionWall
	// TrapTypeCollapseCeiling drops debris from above.
	TrapTypeCollapseCeiling
)

// TrapState represents the current state of a trap.
type TrapState int

const (
	StateHidden    TrapState = iota // StateHidden means the trap is not yet discovered.
	StateDetected                   // StateDetected means the trap has been spotted.
	StateTriggered                  // StateTriggered means the trap has been activated.
	StateDisarmed                   // StateDisarmed means the trap has been safely disabled.
	StateCooldown                   // StateCooldown means the trap is resetting.
	StateBroken                     // StateBroken means the trap is permanently disabled.
)

// Trap represents an interactive trap in the dungeon.
type Trap struct {
	Type            TrapType
	State           TrapState
	X, Y            float64
	TriggerRadius   float64
	EffectRadius    float64
	Damage          int
	StatusEffect    string
	Cooldown        float64
	CooldownTimer   float64
	DetectionDC     int
	DisarmDC        int
	ResetTime       float64
	ResetTimer      float64
	Retriggerable   bool
	Genre           string
	EffectDirection float64
	EffectVelocity  float64
	Seed            int64
}

// TriggerInfo contains information about what triggered a trap.
type TriggerInfo struct {
	EntityID    string
	X, Y        float64
	IsPlayer    bool
	DetectSkill int
	DisarmSkill int
}

// EffectResult represents the outcome of a trap triggering.
type EffectResult struct {
	Triggered       bool
	Detected        bool
	Disarmed        bool
	Damage          int
	StatusEffect    string
	KnockbackX      float64
	KnockbackY      float64
	TeleportX       float64
	TeleportY       float64
	SpawnProjectile bool
	ProjectileType  string
	ProjectileAngle float64
	ParticleEffect  string
}

// NewTrap creates a trap with default parameters.
func NewTrap(trapType TrapType, x, y float64, seed int64) *Trap {
	rngInst := rng.NewRNG(uint64(seed))

	t := &Trap{
		Type:            trapType,
		State:           StateHidden,
		X:               x,
		Y:               y,
		TriggerRadius:   1.0,
		EffectRadius:    3.0,
		Damage:          10,
		Cooldown:        2.0,
		DetectionDC:     12,
		DisarmDC:        15,
		ResetTime:       5.0,
		Retriggerable:   true,
		EffectDirection: rngInst.Float64() * 2 * math.Pi,
		EffectVelocity:  5.0,
		Seed:            seed,
	}

	t.configureByType(rngInst)
	return t
}

// configureByType sets trap-specific parameters based on trap type.
func (t *Trap) configureByType(rng *rng.RNG) {
	switch t.Type {
	case TrapTypePressurePlate:
		t.TriggerRadius = 0.8
		t.DetectionDC = 10
		t.DisarmDC = 12

	case TrapTypeTripwire:
		t.TriggerRadius = 0.3
		t.DetectionDC = 14
		t.DisarmDC = 10

	case TrapTypeProximity:
		t.TriggerRadius = 2.0
		t.DetectionDC = 16
		t.DisarmDC = 18

	case TrapTypeDartWall:
		t.Damage = 8
		t.StatusEffect = "poison"
		t.EffectRadius = 5.0
		t.Retriggerable = false

	case TrapTypeArrowSlit:
		t.Damage = 15
		t.EffectRadius = 8.0
		t.Retriggerable = false

	case TrapTypeFlameThrower:
		t.Damage = 20
		t.StatusEffect = "burning"
		t.EffectRadius = 4.0
		t.Cooldown = 3.0

	case TrapTypeSpikePit:
		t.Damage = 25
		t.Retriggerable = false
		t.DisarmDC = 18

	case TrapTypeSwingingBlade:
		t.Damage = 30
		t.EffectRadius = 2.0
		t.Cooldown = 4.0

	case TrapTypeRollingBoulder:
		t.Damage = 35
		t.EffectRadius = 10.0
		t.EffectVelocity = 8.0
		t.Retriggerable = false

	case TrapTypeElectricShock:
		t.Damage = 18
		t.StatusEffect = "stunned"
		t.EffectRadius = 2.5

	case TrapTypePoisonDart:
		t.Damage = 5
		t.StatusEffect = "poison_strong"
		t.EffectRadius = 6.0

	case TrapTypeNetCatcher:
		t.Damage = 0
		t.StatusEffect = "trapped"
		t.EffectRadius = 2.0

	case TrapTypeBearTrap:
		t.Damage = 20
		t.StatusEffect = "immobilized"
		t.Retriggerable = false

	case TrapTypeExplosive:
		t.Damage = 40
		t.EffectRadius = 4.0
		t.Retriggerable = false

	case TrapTypeTeleporter:
		t.Damage = 0
		t.DetectionDC = 18
		t.DisarmDC = 20

	case TrapTypeIllusionWall:
		t.Damage = 0
		t.DetectionDC = 20
		t.DisarmDC = 25

	case TrapTypeCollapseCeiling:
		t.Damage = 50
		t.EffectRadius = 3.0
		t.Retriggerable = false
		t.DetectionDC = 16
	}
}

// Update advances trap state by deltaTime.
func (t *Trap) Update(deltaTime float64) {
	if t.State == StateCooldown {
		t.CooldownTimer -= deltaTime
		if t.CooldownTimer <= 0 {
			t.State = StateHidden
			t.CooldownTimer = 0
		}
	}

	if t.State == StateTriggered && t.Retriggerable {
		t.ResetTimer -= deltaTime
		if t.ResetTimer <= 0 {
			t.State = StateCooldown
			t.CooldownTimer = t.Cooldown
			t.ResetTimer = 0
		}
	}
}

// CheckTrigger tests if an entity triggers the trap.
func (t *Trap) CheckTrigger(info *TriggerInfo) *EffectResult {
	result := &EffectResult{}

	if t.State == StateDisarmed || t.State == StateBroken {
		return result
	}

	if t.State == StateTriggered && !t.Retriggerable {
		return result
	}

	if t.State == StateCooldown {
		return result
	}

	dx := info.X - t.X
	dy := info.Y - t.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Check detection
	if t.State == StateHidden && info.DetectSkill > 0 {
		detectRoll := rng.NewRNG(uint64(t.Seed+int64(info.X*100+info.Y))).Intn(20) + 1 + info.DetectSkill
		if detectRoll >= t.DetectionDC {
			t.State = StateDetected
			result.Detected = true
			return result
		}
	}

	// Check disarm attempt (only if detected)
	if t.State == StateDetected && info.DisarmSkill > 0 && dist < 1.5 {
		disarmRoll := rng.NewRNG(uint64(t.Seed+int64(info.X*200+info.Y))).Intn(20) + 1 + info.DisarmSkill
		if disarmRoll >= t.DisarmDC {
			t.State = StateDisarmed
			result.Disarmed = true
			return result
		} else if disarmRoll < t.DisarmDC-5 {
			// Critical failure triggers the trap
			t.State = StateTriggered
			result.Triggered = true
			t.applyEffect(info, result)
			return result
		}
	}

	// Check trigger
	if dist <= t.TriggerRadius {
		t.State = StateTriggered
		t.ResetTimer = t.ResetTime
		result.Triggered = true
		t.applyEffect(info, result)
	}

	return result
}

// applyEffect applies the trap's effect to the trigger result.
func (t *Trap) applyEffect(info *TriggerInfo, result *EffectResult) {
	result.Damage = t.Damage
	result.StatusEffect = t.StatusEffect

	dx := info.X - t.X
	dy := info.Y - t.Y
	angle := math.Atan2(dy, dx)

	switch t.Type {
	case TrapTypeDartWall, TrapTypeArrowSlit, TrapTypePoisonDart:
		result.SpawnProjectile = true
		result.ProjectileType = "trap_projectile"
		result.ProjectileAngle = t.EffectDirection
		result.ParticleEffect = "dart_launch"

	case TrapTypeFlameThrower:
		result.ParticleEffect = "flame_burst"
		result.ProjectileAngle = t.EffectDirection

	case TrapTypeSpikePit:
		result.KnockbackX = 0
		result.KnockbackY = 0
		result.ParticleEffect = "spike_emerge"

	case TrapTypeSwingingBlade:
		// Knockback perpendicular to blade swing
		result.KnockbackX = math.Cos(t.EffectDirection+math.Pi/2) * 3.0
		result.KnockbackY = math.Sin(t.EffectDirection+math.Pi/2) * 3.0
		result.ParticleEffect = "blade_slash"

	case TrapTypeRollingBoulder:
		result.KnockbackX = math.Cos(t.EffectDirection) * t.EffectVelocity
		result.KnockbackY = math.Sin(t.EffectDirection) * t.EffectVelocity
		result.ParticleEffect = "boulder_roll"

	case TrapTypeElectricShock:
		result.ParticleEffect = "electric_arc"

	case TrapTypeNetCatcher:
		result.ParticleEffect = "net_spring"

	case TrapTypeBearTrap:
		result.ParticleEffect = "bear_snap"

	case TrapTypeExplosive:
		// Radial knockback from explosion center
		result.KnockbackX = math.Cos(angle) * 5.0
		result.KnockbackY = math.Sin(angle) * 5.0
		result.ParticleEffect = "explosion"

	case TrapTypeTeleporter:
		// Generate random teleport destination
		teleportRNG := rng.NewRNG(uint64(t.Seed + int64(info.X*300+info.Y)))
		result.TeleportX = t.X + (teleportRNG.Float64()*20 - 10)
		result.TeleportY = t.Y + (teleportRNG.Float64()*20 - 10)
		result.ParticleEffect = "teleport"

	case TrapTypeCollapseCeiling:
		result.ParticleEffect = "ceiling_collapse"
	}
}

// GetGenreTraps returns trap types appropriate for a genre.
func GetGenreTraps(genre string) []TrapType {
	switch genre {
	case "fantasy":
		return []TrapType{
			TrapTypePressurePlate, TrapTypeTripwire,
			TrapTypeDartWall, TrapTypeArrowSlit, TrapTypeSpikePit,
			TrapTypeSwingingBlade, TrapTypeRollingBoulder,
			TrapTypeBearTrap, TrapTypeCollapseCeiling,
		}
	case "scifi":
		return []TrapType{
			TrapTypeProximity, TrapTypeButton,
			TrapTypeElectricShock, TrapTypeExplosive,
			TrapTypeTeleporter, TrapTypeIllusionWall,
		}
	case "horror":
		return []TrapType{
			TrapTypeTripwire, TrapTypePressurePlate,
			TrapTypeSpikePit, TrapTypeBearTrap,
			TrapTypeCollapseCeiling, TrapTypePoisonDart,
		}
	case "cyberpunk":
		return []TrapType{
			TrapTypeProximity, TrapTypeLever,
			TrapTypeElectricShock, TrapTypeExplosive,
			TrapTypeTeleporter, TrapTypeFlameThrower,
		}
	default:
		return []TrapType{
			TrapTypePressurePlate, TrapTypeTripwire,
			TrapTypeDartWall, TrapTypeSpikePit,
			TrapTypeExplosive, TrapTypeBearTrap,
		}
	}
}
