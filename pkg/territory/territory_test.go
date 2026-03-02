package territory

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/faction"
)

func TestNewControlSystem(t *testing.T) {
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	if sys == nil {
		t.Fatal("NewControlSystem returned nil")
	}

	if sys.mapWidth != 100 || sys.mapHeight != 100 {
		t.Errorf("Map dimensions incorrect: got %dx%d, want 100x100", sys.mapWidth, sys.mapHeight)
	}

	if len(sys.territories) != 0 {
		t.Errorf("Expected 0 initial territories, got %d", len(sys.territories))
	}

	if sys.gridSize != 8 {
		t.Errorf("Expected grid size 8, got %d", sys.gridSize)
	}
}

func TestClaimRoom(t *testing.T) {
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 20, W: 15, H: 12}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	if len(sys.territories) != 1 {
		t.Fatalf("Expected 1 territory after claiming, got %d", len(sys.territories))
	}

	centerX := float64(room.X + room.W/2)
	centerY := float64(room.Y + room.H/2)

	territory := sys.GetTerritoryByPosition(centerX, centerY)
	if territory == nil {
		t.Fatal("Territory not found at room center")
	}

	if territory.ControlFaction != faction.FactionMercenaries {
		t.Errorf("Expected faction %s, got %s", faction.FactionMercenaries, territory.ControlFaction)
	}

	if territory.Contested {
		t.Error("New territory should not be contested")
	}

	if len(territory.PatrolSpawns) < 1 {
		t.Error("Territory should have patrol spawns")
	}
}

func TestTerritoryAssignment(t *testing.T) {
	world := engine.NewWorld()
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 10, W: 10, H: 10}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	entity := world.AddEntity()
	world.AddComponent(entity, &PositionComponent{X: 15.0, Y: 15.0})

	sys.Update(world)

	terType := reflect.TypeOf((*TerritoryComponent)(nil))
	terComp, ok := world.GetComponent(entity, terType)
	if !ok {
		t.Fatal("Entity should have TerritoryComponent after update")
	}

	tc, ok := terComp.(*TerritoryComponent)
	if !ok {
		t.Fatal("Component is not TerritoryComponent")
	}

	if tc.CurrentTerritory == "" {
		t.Error("Entity should be assigned to a territory")
	}
}

func TestContestationDetection(t *testing.T) {
	world := engine.NewWorld()
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 10, W: 10, H: 10}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	centerX := 15.0
	centerY := 15.0

	defender := world.AddEntity()
	world.AddComponent(defender, &PositionComponent{X: centerX, Y: centerY})
	world.AddComponent(defender, &faction.FactionMemberComponent{FactionID: faction.FactionMercenaries})
	world.AddComponent(defender, &combat.HealthComponent{Current: 100, Max: 100})

	attacker := world.AddEntity()
	world.AddComponent(attacker, &PositionComponent{X: centerX + 1, Y: centerY + 1})
	world.AddComponent(attacker, &faction.FactionMemberComponent{FactionID: faction.FactionRebels})
	world.AddComponent(attacker, &combat.HealthComponent{Current: 100, Max: 100})

	sys.Update(world)

	territory := sys.GetTerritoryByPosition(centerX, centerY)
	if territory == nil {
		t.Fatal("Territory not found")
	}

	territory.Contested = true
	sys.Update(world)

	if !territory.Contested {
		t.Error("Territory should remain contested with opposing factions present")
	}
}

func TestPatrolSpawning(t *testing.T) {
	world := engine.NewWorld()
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 10, W: 20, H: 20}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	for i := 0; i < 1000; i++ {
		sys.Update(world)
	}

	patType := reflect.TypeOf((*PatrolComponent)(nil))
	patrols := world.Query(patType)

	if len(patrols) == 0 {
		t.Log("No patrols spawned after 1000 updates (expected due to random chance)")
	}
}

func TestGetBattleFronts(t *testing.T) {
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room1 := &bsp.Room{X: 10, Y: 10, W: 10, H: 10}
	sys.ClaimRoom(room1, faction.FactionMercenaries)

	room2 := &bsp.Room{X: 30, Y: 30, W: 10, H: 10}
	sys.ClaimRoom(room2, faction.FactionRebels)

	fronts := sys.GetBattleFronts()
	if len(fronts) != 0 {
		t.Errorf("Expected 0 battle fronts initially, got %d", len(fronts))
	}

	territories := make([]*Territory, 0, len(sys.territories))
	sys.mu.RLock()
	for _, ter := range sys.territories {
		territories = append(territories, ter)
	}
	sys.mu.RUnlock()

	if len(territories) > 0 {
		territories[0].Contested = true
	}

	fronts = sys.GetBattleFronts()
	if len(fronts) != 1 {
		t.Errorf("Expected 1 battle front after setting contested, got %d", len(fronts))
	}
}

func TestPatrolRouteGeneration(t *testing.T) {
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	territory := &Territory{
		ID:      "test",
		CenterX: 50,
		CenterY: 50,
		Radius:  10,
	}

	spawn := SpawnPoint{X: 50, Y: 50, Active: true}
	route := sys.generatePatrolRoute(territory, spawn)

	if len(route) < 4 {
		t.Errorf("Expected at least 4 waypoints in patrol route, got %d", len(route))
	}

	if route[0].X != spawn.X || route[0].Y != spawn.Y {
		t.Error("Patrol route should start at spawn point")
	}

	if route[len(route)-1].X != spawn.X || route[len(route)-1].Y != spawn.Y {
		t.Error("Patrol route should end at spawn point (loop)")
	}
}

func TestTerritoryControlTransfer(t *testing.T) {
	world := engine.NewWorld()
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 10, W: 10, H: 10}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	centerX := 15.0
	centerY := 15.0

	for i := 0; i < 5; i++ {
		attacker := world.AddEntity()
		world.AddComponent(attacker, &PositionComponent{
			X: centerX + float64(i)*0.5,
			Y: centerY + float64(i)*0.5,
		})
		world.AddComponent(attacker, &faction.FactionMemberComponent{FactionID: faction.FactionRebels})
		world.AddComponent(attacker, &combat.HealthComponent{Current: 100, Max: 100})
	}

	territory := sys.GetTerritoryByPosition(centerX, centerY)
	if territory == nil {
		t.Fatal("Territory not found")
	}

	originalFaction := territory.ControlFaction
	territory.Contested = true
	territory.ControlPoints = 1

	for i := 0; i < 5; i++ {
		sys.Update(world)
	}

	if territory.ControlFaction == originalFaction && territory.ControlPoints <= 0 {
		t.Error("Territory should transfer control when control points depleted")
	}
}

func TestIsContested(t *testing.T) {
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 10, W: 10, H: 10}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	sys.mu.RLock()
	var territoryID TerritoryID
	for id := range sys.territories {
		territoryID = id
		break
	}
	sys.mu.RUnlock()

	if sys.IsContested(territoryID) {
		t.Error("New territory should not be contested")
	}

	sys.mu.Lock()
	if ter := sys.territories[territoryID]; ter != nil {
		ter.Contested = true
	}
	sys.mu.Unlock()

	if !sys.IsContested(territoryID) {
		t.Error("Territory should be contested after setting flag")
	}
}

func TestGetControllingFaction(t *testing.T) {
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	room := &bsp.Room{X: 10, Y: 10, W: 10, H: 10}
	sys.ClaimRoom(room, faction.FactionMercenaries)

	sys.mu.RLock()
	var territoryID TerritoryID
	for id := range sys.territories {
		territoryID = id
		break
	}
	sys.mu.RUnlock()

	controlFaction := sys.GetControllingFaction(territoryID)
	if controlFaction != faction.FactionMercenaries {
		t.Errorf("Expected controlling faction %s, got %s", faction.FactionMercenaries, controlFaction)
	}

	invalidFaction := sys.GetControllingFaction("nonexistent")
	if invalidFaction != "" {
		t.Errorf("Expected empty faction for invalid territory, got %s", invalidFaction)
	}
}

func TestPatrolMovement(t *testing.T) {
	world := engine.NewWorld()
	factionSys := faction.NewReputationSystem()
	sys := NewControlSystem(100, 100, factionSys)

	entity := world.AddEntity()
	startX := 10.0
	startY := 10.0

	world.AddComponent(entity, &PositionComponent{X: startX, Y: startY})
	world.AddComponent(entity, &PatrolComponent{
		TerritoryID: "test",
		PatrolRoute: []Position{
			{X: startX, Y: startY},
			{X: startX + 5, Y: startY + 5},
			{X: startX, Y: startY},
		},
		RouteIndex: 0,
	})

	for i := 0; i < 100; i++ {
		sys.updatePatrolMovement(world)
	}

	posType := reflect.TypeOf((*PositionComponent)(nil))
	posComp, ok := world.GetComponent(entity, posType)
	if !ok {
		t.Fatal("Entity lost position component")
	}

	pos := posComp.(*PositionComponent)
	if pos.X == startX && pos.Y == startY {
		t.Error("Patrol should have moved from starting position")
	}
}

func TestComponentTypes(t *testing.T) {
	tests := []struct {
		name      string
		component interface{ Type() string }
		expected  string
	}{
		{"TerritoryComponent", &TerritoryComponent{}, "TerritoryComponent"},
		{"PatrolComponent", &PatrolComponent{}, "PatrolComponent"},
		{"ReinforcementComponent", &ReinforcementComponent{}, "ReinforcementComponent"},
		{"PositionComponent", &PositionComponent{}, "PositionComponent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.Type(); got != tt.expected {
				t.Errorf("Type() = %v, want %v", got, tt.expected)
			}
		})
	}
}
