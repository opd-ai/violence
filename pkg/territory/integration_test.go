package territory

import (
	"testing"

	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/faction"
)

// TestIntegration verifies territory system integration without GUI dependencies.
func TestIntegration(t *testing.T) {
	// Create faction system
	factionSys := faction.NewReputationSystem()
	if factionSys == nil {
		t.Fatal("Failed to create faction system")
	}

	// Create territory system
	territorySys := NewControlSystem(128, 128, factionSys)
	if territorySys == nil {
		t.Fatal("Failed to create territory system")
	}

	// Create mock rooms
	rooms := []*bsp.Room{
		{X: 10, Y: 10, W: 15, H: 15},
		{X: 30, Y: 30, W: 20, H: 20},
		{X: 60, Y: 60, W: 12, H: 12},
	}

	// Claim territories
	activeFactions := []faction.FactionID{
		faction.FactionMercenaries,
		faction.FactionRebels,
		faction.FactionCult,
	}

	for i, room := range rooms {
		factionID := activeFactions[i%len(activeFactions)]
		territorySys.ClaimRoom(room, factionID)
	}

	// Verify territories were created
	sys := territorySys
	sys.mu.RLock()
	territoryCount := len(sys.territories)
	sys.mu.RUnlock()

	if territoryCount != len(rooms) {
		t.Errorf("Expected %d territories, got %d", len(rooms), territoryCount)
	}

	// Verify territory lookup
	for i, room := range rooms {
		centerX := float64(room.X + room.W/2)
		centerY := float64(room.Y + room.H/2)

		territory := territorySys.GetTerritoryByPosition(centerX, centerY)
		if territory == nil {
			t.Errorf("Territory not found for room %d at position (%.1f, %.1f)", i, centerX, centerY)
			continue
		}

		expectedFaction := activeFactions[i%len(activeFactions)]
		if territory.ControlFaction != expectedFaction {
			t.Errorf("Room %d: expected faction %s, got %s", i, expectedFaction, territory.ControlFaction)
		}
	}

	// Verify battle fronts query
	fronts := territorySys.GetBattleFronts()
	if len(fronts) != 0 {
		t.Errorf("Expected 0 battle fronts initially, got %d", len(fronts))
	}

	// Mark one territory as contested
	sys.mu.Lock()
	for _, ter := range sys.territories {
		ter.Contested = true
		break
	}
	sys.mu.Unlock()

	fronts = territorySys.GetBattleFronts()
	if len(fronts) != 1 {
		t.Errorf("Expected 1 battle front after marking contested, got %d", len(fronts))
	}

	t.Logf("Integration test passed: %d territories claimed, battle fronts working", territoryCount)
}
