// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// EnemyRole defines distinct behavioral archetypes for enemies.
type EnemyRole int

const (
	// RoleTank - high HP, aggressive positioning, draws fire.
	RoleTank EnemyRole = iota
	// RoleRanged - keeps distance, kites when approached.
	RoleRanged
	// RoleHealer - supports allies, retreats from combat.
	RoleHealer
	// RoleAmbusher - hides and waits, burst damage.
	RoleAmbusher
	// RoleScout - fast movement, alerts squad.
	RoleScout
)

// RoleConfig holds behavior parameters for an enemy role.
type RoleConfig struct {
	Role                EnemyRole
	PreferredRange      float64 // Ideal distance from target
	MinRange            float64 // Minimum comfortable distance
	MaxRange            float64 // Maximum engagement distance
	AggressionLevel     float64 // 0.0 = passive, 1.0 = very aggressive
	RetreatHealthPct    float64 // Health % at which to retreat
	SupportPriority     float64 // How much to prioritize helping allies
	MovementSpeed       float64 // Speed multiplier
	UsesCover           bool
	AlertsOnPlayerSight bool
}

// GetRoleConfig returns behavior configuration for a role.
func GetRoleConfig(role EnemyRole, genre string) RoleConfig {
	base := getRoleBase(role)

	// Genre modifications
	switch genre {
	case "cyberpunk", "scifi":
		base.MaxRange *= 1.2
		base.UsesCover = true
	case "horror":
		base.AggressionLevel *= 0.8
		if role == RoleAmbusher {
			base.AggressionLevel *= 1.5
		}
	case "fantasy":
		if role == RoleHealer {
			base.SupportPriority *= 1.3
		}
	}

	return base
}

func getRoleBase(role EnemyRole) RoleConfig {
	switch role {
	case RoleTank:
		return RoleConfig{
			Role:                RoleTank,
			PreferredRange:      80.0,
			MinRange:            40.0,
			MaxRange:            150.0,
			AggressionLevel:     0.9,
			RetreatHealthPct:    0.15,
			SupportPriority:     0.3,
			MovementSpeed:       0.9,
			UsesCover:           false,
			AlertsOnPlayerSight: true,
		}
	case RoleRanged:
		return RoleConfig{
			Role:                RoleRanged,
			PreferredRange:      200.0,
			MinRange:            150.0,
			MaxRange:            300.0,
			AggressionLevel:     0.6,
			RetreatHealthPct:    0.4,
			SupportPriority:     0.2,
			MovementSpeed:       1.0,
			UsesCover:           true,
			AlertsOnPlayerSight: true,
		}
	case RoleHealer:
		return RoleConfig{
			Role:                RoleHealer,
			PreferredRange:      180.0,
			MinRange:            120.0,
			MaxRange:            250.0,
			AggressionLevel:     0.2,
			RetreatHealthPct:    0.6,
			SupportPriority:     0.95,
			MovementSpeed:       1.1,
			UsesCover:           true,
			AlertsOnPlayerSight: false,
		}
	case RoleAmbusher:
		return RoleConfig{
			Role:                RoleAmbusher,
			PreferredRange:      60.0,
			MinRange:            30.0,
			MaxRange:            120.0,
			AggressionLevel:     0.95,
			RetreatHealthPct:    0.3,
			SupportPriority:     0.1,
			MovementSpeed:       1.3,
			UsesCover:           true,
			AlertsOnPlayerSight: false,
		}
	case RoleScout:
		return RoleConfig{
			Role:                RoleScout,
			PreferredRange:      140.0,
			MinRange:            100.0,
			MaxRange:            200.0,
			AggressionLevel:     0.5,
			RetreatHealthPct:    0.5,
			SupportPriority:     0.4,
			MovementSpeed:       1.4,
			UsesCover:           false,
			AlertsOnPlayerSight: true,
		}
	default:
		return getRoleBase(RoleTank)
	}
}

// SquadTactics manages group coordination for enemy squads.
type SquadTactics struct {
	SquadID         string
	Members         []string // Entity IDs
	LeaderID        string
	FocusTargetID   string          // Coordinated target
	FlankPositions  map[string]bool // Member ID -> is flanking
	FormationCenter [2]float64
	LastUpdateTime  float64
	AlertLevel      float64 // 0.0 = unaware, 1.0 = full alert
}

// NewSquadTactics creates a squad tactics coordinator.
func NewSquadTactics(squadID string) *SquadTactics {
	return &SquadTactics{
		SquadID:        squadID,
		Members:        make([]string, 0, 6),
		FlankPositions: make(map[string]bool),
		AlertLevel:     0.0,
	}
}

// AddMember registers an entity as part of this squad.
func (st *SquadTactics) AddMember(entityID string) {
	for _, id := range st.Members {
		if id == entityID {
			return
		}
	}
	st.Members = append(st.Members, entityID)
	if st.LeaderID == "" {
		st.LeaderID = entityID
	}
}

// RemoveMember unregisters an entity from the squad.
func (st *SquadTactics) RemoveMember(entityID string) {
	for i, id := range st.Members {
		if id == entityID {
			st.Members = append(st.Members[:i], st.Members[i+1:]...)
			delete(st.FlankPositions, entityID)
			if st.LeaderID == entityID && len(st.Members) > 0 {
				st.LeaderID = st.Members[0]
			}
			return
		}
	}
}

// UpdateFormation calculates formation center and flank positions.
func (st *SquadTactics) UpdateFormation(memberPositions map[string][2]float64, targetPos [2]float64, rngSrc *rng.RNG) {
	if len(st.Members) == 0 {
		return
	}

	// Calculate center of mass
	var sumX, sumY float64
	count := 0
	for _, id := range st.Members {
		if pos, ok := memberPositions[id]; ok {
			sumX += pos[0]
			sumY += pos[1]
			count++
		}
	}

	if count > 0 {
		st.FormationCenter[0] = sumX / float64(count)
		st.FormationCenter[1] = sumY / float64(count)
	}

	// Assign flank positions - 40% of squad flanks
	numFlank := max(1, len(st.Members)*2/5)
	assignedFlank := 0

	// Reset flank positions
	for k := range st.FlankPositions {
		st.FlankPositions[k] = false
	}

	// Select flankers (prefer fast roles)
	shuffled := make([]string, len(st.Members))
	copy(shuffled, st.Members)
	if rngSrc != nil {
		// Fisher-Yates shuffle
		for i := len(shuffled) - 1; i > 0; i-- {
			j := rngSrc.Intn(i + 1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}
	}

	for _, id := range shuffled {
		if assignedFlank >= numFlank {
			break
		}
		st.FlankPositions[id] = true
		assignedFlank++
	}
}

// ShouldFlank returns whether this member should flank.
func (st *SquadTactics) ShouldFlank(entityID string) bool {
	return st.FlankPositions[entityID]
}

// GetFlankVector returns a vector for flanking the target.
// Returns perpendicular offset from target-to-formation vector.
func (st *SquadTactics) GetFlankVector(entityID string, targetPos [2]float64, side float64) [2]float64 {
	// Vector from target to formation center
	dx := st.FormationCenter[0] - targetPos[0]
	dy := st.FormationCenter[1] - targetPos[1]

	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.001 {
		return [2]float64{side * 100, 0}
	}

	// Perpendicular vector (rotate 90 degrees)
	perpX := -dy / dist
	perpY := dx / dist

	// Scale by flank distance and side
	flankDist := 120.0
	return [2]float64{perpX * flankDist * side, perpY * flankDist * side}
}

// RaiseAlert increases squad awareness.
func (st *SquadTactics) RaiseAlert(amount float64) {
	st.AlertLevel = math.Min(1.0, st.AlertLevel+amount)
}

// DecayAlert gradually reduces alert level over time.
func (st *SquadTactics) DecayAlert(deltaTime float64) {
	decayRate := 0.1 // 10% per second
	st.AlertLevel = math.Max(0.0, st.AlertLevel-decayRate*deltaTime)
}

// SelectFocusTarget picks a target for coordinated assault.
// Returns true if target changed.
func (st *SquadTactics) SelectFocusTarget(candidates []string, rngSrc *rng.RNG) bool {
	if len(candidates) == 0 {
		if st.FocusTargetID != "" {
			st.FocusTargetID = ""
			return true
		}
		return false
	}

	// Pick random target (in real impl, would factor threat/health/distance)
	idx := 0
	if rngSrc != nil && len(candidates) > 1 {
		idx = rngSrc.Intn(len(candidates))
	}

	newTarget := candidates[idx]
	changed := newTarget != st.FocusTargetID
	st.FocusTargetID = newTarget
	return changed
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
