// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"testing"

	"github.com/opd-ai/violence/pkg/rng"
)

func TestGetRoleConfig(t *testing.T) {
	tests := []struct {
		name         string
		role         EnemyRole
		genre        string
		wantMinRange float64
		wantMaxRange float64
	}{
		{
			name:         "Tank fantasy",
			role:         RoleTank,
			genre:        "fantasy",
			wantMinRange: 40.0,
			wantMaxRange: 150.0,
		},
		{
			name:         "Ranged cyberpunk increased range",
			role:         RoleRanged,
			genre:        "cyberpunk",
			wantMinRange: 150.0,
			wantMaxRange: 360.0, // 300 * 1.2
		},
		{
			name:         "Healer fantasy boosted support",
			role:         RoleHealer,
			genre:        "fantasy",
			wantMinRange: 120.0,
			wantMaxRange: 250.0,
		},
		{
			name:         "Ambusher horror",
			role:         RoleAmbusher,
			genre:        "horror",
			wantMinRange: 30.0,
			wantMaxRange: 120.0,
		},
		{
			name:         "Scout scifi",
			role:         RoleScout,
			genre:        "scifi",
			wantMinRange: 100.0,
			wantMaxRange: 240.0, // 200 * 1.2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetRoleConfig(tt.role, tt.genre)
			if cfg.Role != tt.role {
				t.Errorf("Role = %v, want %v", cfg.Role, tt.role)
			}
			if cfg.MinRange != tt.wantMinRange {
				t.Errorf("MinRange = %v, want %v", cfg.MinRange, tt.wantMinRange)
			}
			if cfg.MaxRange != tt.wantMaxRange {
				t.Errorf("MaxRange = %v, want %v", cfg.MaxRange, tt.wantMaxRange)
			}
			if cfg.PreferredRange < cfg.MinRange || cfg.PreferredRange > cfg.MaxRange {
				t.Errorf("PreferredRange %v outside valid range [%v, %v]",
					cfg.PreferredRange, cfg.MinRange, cfg.MaxRange)
			}
		})
	}
}

func TestGetRoleConfig_GenreModifications(t *testing.T) {
	// Test that cyberpunk increases range
	baseCfg := GetRoleConfig(RoleRanged, "fantasy")
	cyberpunkCfg := GetRoleConfig(RoleRanged, "cyberpunk")
	if cyberpunkCfg.MaxRange <= baseCfg.MaxRange {
		t.Errorf("cyberpunk should increase MaxRange")
	}
	if !cyberpunkCfg.UsesCover {
		t.Errorf("cyberpunk should enable UsesCover")
	}

	// Test that fantasy boosts healer support
	fantasyHealer := GetRoleConfig(RoleHealer, "fantasy")
	baseHealer := GetRoleConfig(RoleHealer, "unknown")
	if fantasyHealer.SupportPriority <= baseHealer.SupportPriority {
		t.Errorf("fantasy should boost healer SupportPriority")
	}

	// Test that horror affects ambusher aggression
	horrorAmbush := GetRoleConfig(RoleAmbusher, "horror")
	if horrorAmbush.AggressionLevel < 0.5 {
		t.Errorf("horror ambusher should still be aggressive")
	}
}

func TestRoleConfig_Characteristics(t *testing.T) {
	// Tank should be aggressive and durable
	tank := GetRoleConfig(RoleTank, "fantasy")
	if tank.AggressionLevel < 0.7 {
		t.Errorf("Tank should be aggressive")
	}
	if tank.RetreatHealthPct > 0.2 {
		t.Errorf("Tank should retreat at low health only")
	}

	// Ranged should keep distance
	ranged := GetRoleConfig(RoleRanged, "fantasy")
	if ranged.PreferredRange < 150 {
		t.Errorf("Ranged should prefer distance")
	}
	if ranged.RetreatHealthPct < 0.3 {
		t.Errorf("Ranged should retreat earlier")
	}

	// Healer should prioritize support
	healer := GetRoleConfig(RoleHealer, "fantasy")
	if healer.SupportPriority < 0.8 {
		t.Errorf("Healer should have high support priority")
	}
	if healer.AggressionLevel > 0.4 {
		t.Errorf("Healer should be less aggressive")
	}

	// Ambusher should be fast and aggressive
	ambusher := GetRoleConfig(RoleAmbusher, "fantasy")
	if ambusher.MovementSpeed < 1.2 {
		t.Errorf("Ambusher should be fast")
	}
	if ambusher.AggressionLevel < 0.9 {
		t.Errorf("Ambusher should be very aggressive")
	}

	// Scout should be fastest
	scout := GetRoleConfig(RoleScout, "fantasy")
	if scout.MovementSpeed < 1.3 {
		t.Errorf("Scout should be fastest")
	}
	if !scout.AlertsOnPlayerSight {
		t.Errorf("Scout should alert squad")
	}
}

func TestNewSquadTactics(t *testing.T) {
	st := NewSquadTactics("squad-alpha")
	if st.SquadID != "squad-alpha" {
		t.Errorf("SquadID = %v, want squad-alpha", st.SquadID)
	}
	if len(st.Members) != 0 {
		t.Errorf("New squad should have no members")
	}
	if st.AlertLevel != 0.0 {
		t.Errorf("AlertLevel should start at 0")
	}
}

func TestSquadTactics_AddRemoveMember(t *testing.T) {
	st := NewSquadTactics("squad-1")

	// Add first member (becomes leader)
	st.AddMember("entity-1")
	if len(st.Members) != 1 {
		t.Errorf("Should have 1 member")
	}
	if st.LeaderID != "entity-1" {
		t.Errorf("First member should be leader")
	}

	// Add more members
	st.AddMember("entity-2")
	st.AddMember("entity-3")
	if len(st.Members) != 3 {
		t.Errorf("Should have 3 members")
	}

	// Adding duplicate does nothing
	st.AddMember("entity-2")
	if len(st.Members) != 3 {
		t.Errorf("Duplicate add should not increase count")
	}

	// Remove non-leader member
	st.RemoveMember("entity-2")
	if len(st.Members) != 2 {
		t.Errorf("Should have 2 members after removal")
	}
	if st.LeaderID != "entity-1" {
		t.Errorf("Leader should not change")
	}

	// Remove leader
	st.RemoveMember("entity-1")
	if len(st.Members) != 1 {
		t.Errorf("Should have 1 member after leader removal")
	}
	if st.LeaderID != "entity-3" {
		t.Errorf("New leader should be entity-3, got %v", st.LeaderID)
	}

	// Remove non-existent member
	st.RemoveMember("entity-999")
	if len(st.Members) != 1 {
		t.Errorf("Removing non-existent member should not change count")
	}
}

func TestSquadTactics_UpdateFormation(t *testing.T) {
	st := NewSquadTactics("squad-1")
	st.AddMember("entity-1")
	st.AddMember("entity-2")
	st.AddMember("entity-3")

	positions := map[string][2]float64{
		"entity-1": {100.0, 100.0},
		"entity-2": {150.0, 100.0},
		"entity-3": {125.0, 150.0},
	}

	rngSrc := rng.NewRNG(42)
	targetPos := [2]float64{200.0, 200.0}

	st.UpdateFormation(positions, targetPos, rngSrc)

	// Check formation center is calculated
	expectedCenterX := (100.0 + 150.0 + 125.0) / 3.0
	expectedCenterY := (100.0 + 100.0 + 150.0) / 3.0

	if st.FormationCenter[0] != expectedCenterX {
		t.Errorf("FormationCenter[0] = %v, want %v", st.FormationCenter[0], expectedCenterX)
	}
	if st.FormationCenter[1] != expectedCenterY {
		t.Errorf("FormationCenter[1] = %v, want %v", st.FormationCenter[1], expectedCenterY)
	}

	// Check that some members are assigned flank positions
	flankCount := 0
	for _, isFlanking := range st.FlankPositions {
		if isFlanking {
			flankCount++
		}
	}

	// 40% of 3 members = 1.2, rounded down to 1, but max(1, ...) ensures at least 1
	if flankCount < 1 {
		t.Errorf("At least 1 member should be flanking, got %v", flankCount)
	}
}

func TestSquadTactics_ShouldFlank(t *testing.T) {
	st := NewSquadTactics("squad-1")
	st.AddMember("entity-1")
	st.AddMember("entity-2")

	st.FlankPositions["entity-1"] = true
	st.FlankPositions["entity-2"] = false

	if !st.ShouldFlank("entity-1") {
		t.Errorf("entity-1 should be flanking")
	}
	if st.ShouldFlank("entity-2") {
		t.Errorf("entity-2 should not be flanking")
	}
	if st.ShouldFlank("entity-999") {
		t.Errorf("unknown entity should not be flanking")
	}
}

func TestSquadTactics_GetFlankVector(t *testing.T) {
	st := NewSquadTactics("squad-1")
	st.FormationCenter = [2]float64{100.0, 100.0}

	targetPos := [2]float64{200.0, 100.0}
	vec := st.GetFlankVector("entity-1", targetPos, 1.0)

	// Target is directly to the right of formation
	// Perpendicular should be up or down
	// Vector from target to formation: (-100, 0)
	// Perpendicular: (0, 100) or (0, -100)

	if vec[0] != 0.0 {
		t.Errorf("Flank vector X should be 0 for horizontal approach, got %v", vec[0])
	}
	if vec[1] != 120.0 && vec[1] != -120.0 {
		t.Errorf("Flank vector Y should be ±120, got %v", vec[1])
	}
}

func TestSquadTactics_GetFlankVector_ZeroDistance(t *testing.T) {
	st := NewSquadTactics("squad-1")
	st.FormationCenter = [2]float64{100.0, 100.0}

	targetPos := [2]float64{100.0, 100.0} // Same as formation center
	vec := st.GetFlankVector("entity-1", targetPos, 1.0)

	// Should return default offset when distance is zero
	if vec[0] != 100.0 {
		t.Errorf("Default flank X = %v, want 100", vec[0])
	}
	if vec[1] != 0.0 {
		t.Errorf("Default flank Y = %v, want 0", vec[1])
	}
}

func TestSquadTactics_RaiseAlert(t *testing.T) {
	st := NewSquadTactics("squad-1")

	st.RaiseAlert(0.3)
	if st.AlertLevel != 0.3 {
		t.Errorf("AlertLevel = %v, want 0.3", st.AlertLevel)
	}

	st.RaiseAlert(0.5)
	if st.AlertLevel != 0.8 {
		t.Errorf("AlertLevel = %v, want 0.8", st.AlertLevel)
	}

	// Should cap at 1.0
	st.RaiseAlert(0.5)
	if st.AlertLevel != 1.0 {
		t.Errorf("AlertLevel = %v, want 1.0 (capped)", st.AlertLevel)
	}
}

func TestSquadTactics_DecayAlert(t *testing.T) {
	st := NewSquadTactics("squad-1")
	st.AlertLevel = 0.5

	st.DecayAlert(1.0) // 1 second at 10% decay rate
	expected := 0.5 - 0.1
	if st.AlertLevel != expected {
		t.Errorf("AlertLevel = %v, want %v", st.AlertLevel, expected)
	}

	// Should floor at 0.0
	st.AlertLevel = 0.05
	st.DecayAlert(1.0)
	if st.AlertLevel != 0.0 {
		t.Errorf("AlertLevel = %v, want 0.0 (floored)", st.AlertLevel)
	}
}

func TestSquadTactics_SelectFocusTarget(t *testing.T) {
	st := NewSquadTactics("squad-1")
	rngSrc := rng.NewRNG(99)

	// No candidates
	changed := st.SelectFocusTarget([]string{}, rngSrc)
	if changed {
		t.Errorf("Should not change with no candidates")
	}
	if st.FocusTargetID != "" {
		t.Errorf("FocusTargetID should be empty")
	}

	// One candidate
	candidates := []string{"player-1"}
	changed = st.SelectFocusTarget(candidates, rngSrc)
	if !changed {
		t.Errorf("Should change when setting first target")
	}
	if st.FocusTargetID != "player-1" {
		t.Errorf("FocusTargetID = %v, want player-1", st.FocusTargetID)
	}

	// Same candidate, no change
	changed = st.SelectFocusTarget(candidates, rngSrc)
	if changed {
		t.Errorf("Should not change when target is same")
	}

	// Multiple candidates
	candidates = []string{"player-1", "player-2", "player-3"}
	changed = st.SelectFocusTarget(candidates, rngSrc)
	// May or may not change depending on RNG
	if st.FocusTargetID != "player-1" && st.FocusTargetID != "player-2" && st.FocusTargetID != "player-3" {
		t.Errorf("FocusTargetID should be one of the candidates")
	}
}

func TestSquadTactics_SelectFocusTarget_ClearsOnEmpty(t *testing.T) {
	st := NewSquadTactics("squad-1")
	st.FocusTargetID = "player-1"

	changed := st.SelectFocusTarget([]string{}, nil)
	if !changed {
		t.Errorf("Should detect change when clearing target")
	}
	if st.FocusTargetID != "" {
		t.Errorf("FocusTargetID should be cleared")
	}
}

func BenchmarkGetRoleConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetRoleConfig(RoleTank, "fantasy")
	}
}

func BenchmarkSquadTactics_UpdateFormation(b *testing.B) {
	st := NewSquadTactics("squad-1")
	for i := 0; i < 10; i++ {
		st.AddMember("entity-" + string(rune(i)))
	}

	positions := make(map[string][2]float64)
	for i := 0; i < 10; i++ {
		positions["entity-"+string(rune(i))] = [2]float64{float64(i * 50), float64(i * 50)}
	}

	rngSrc := rng.NewRNG(42)
	targetPos := [2]float64{500.0, 500.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.UpdateFormation(positions, targetPos, rngSrc)
	}
}
