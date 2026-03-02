package trap

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem(12345)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}

	if sys.genre != "fantasy" {
		t.Errorf("Default genre = %v, want fantasy", sys.genre)
	}

	if len(sys.traps) != 0 {
		t.Errorf("Initial trap count = %v, want 0", len(sys.traps))
	}
}

func TestSystemSetGenre(t *testing.T) {
	sys := NewSystem(12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genre != genre {
			t.Errorf("Genre = %v, want %v", sys.genre, genre)
		}
	}
}

func TestSystemAddTrap(t *testing.T) {
	sys := NewSystem(12345)

	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	sys.AddTrap(trap)

	if len(sys.traps) != 1 {
		t.Errorf("Trap count = %v, want 1", len(sys.traps))
	}

	traps := sys.GetTraps()
	if len(traps) != 1 {
		t.Errorf("GetTraps count = %v, want 1", len(traps))
	}

	if traps[0] != trap {
		t.Error("GetTraps returned wrong trap")
	}
}

func TestSystemRemoveTrap(t *testing.T) {
	sys := NewSystem(12345)

	trap1 := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	trap2 := NewTrap(TrapTypeTripwire, 10.0, 10.0, 67890)

	sys.AddTrap(trap1)
	sys.AddTrap(trap2)

	if len(sys.traps) != 2 {
		t.Fatalf("Trap count = %v, want 2", len(sys.traps))
	}

	sys.RemoveTrap(trap1)

	if len(sys.traps) != 1 {
		t.Errorf("Trap count after remove = %v, want 1", len(sys.traps))
	}

	if sys.traps[0] != trap2 {
		t.Error("Wrong trap remaining after removal")
	}
}

func TestSystemClearTraps(t *testing.T) {
	sys := NewSystem(12345)

	for i := 0; i < 5; i++ {
		trap := NewTrap(TrapTypePressurePlate, float64(i), float64(i), int64(i))
		sys.AddTrap(trap)
	}

	if len(sys.traps) != 5 {
		t.Fatalf("Trap count = %v, want 5", len(sys.traps))
	}

	sys.ClearTraps()

	if len(sys.traps) != 0 {
		t.Errorf("Trap count after clear = %v, want 0", len(sys.traps))
	}
}

func TestSystemUpdate(t *testing.T) {
	sys := NewSystem(12345)

	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	trap.State = StateCooldown
	trap.CooldownTimer = 1.0
	sys.AddTrap(trap)

	// Create ECS world
	w := engine.NewWorld()

	// Create entity far from trap
	entity := w.AddEntity()
	w.AddComponent(entity, &PositionComponent{PosX: 50.0, PosY: 50.0})

	// Multiple updates to reduce cooldown
	for i := 0; i < 60; i++ {
		sys.Update(w)
	}

	if trap.State != StateHidden {
		t.Errorf("State = %v, want %v after cooldown", trap.State, StateHidden)
	}
}

func TestSystemEntityTrigger(t *testing.T) {
	sys := NewSystem(12345)

	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	sys.AddTrap(trap)

	// Create ECS world
	w := engine.NewWorld()

	// Create entity on trap with health
	entity := w.AddEntity()
	w.AddComponent(entity, &PositionComponent{PosX: 5.0, PosY: 5.0})
	w.AddComponent(entity, &HealthComponent{Current: 100, Max: 100})

	healthType := reflect.TypeOf((*HealthComponent)(nil))
	initialHealthComp, _ := w.GetComponent(entity, healthType)
	initialHealth := initialHealthComp.(*HealthComponent).Current

	// Update should trigger trap
	sys.Update(w)

	if trap.State != StateTriggered {
		t.Error("Trap should be triggered by entity")
	}

	finalHealthComp, _ := w.GetComponent(entity, healthType)
	finalHealth := finalHealthComp.(*HealthComponent).Current
	if finalHealth >= initialHealth {
		t.Error("Trap should have damaged entity")
	}
}

func TestGenerateTraps(t *testing.T) {
	sys := NewSystem(12345)

	// Create a map with corridors (0 = wall, 2 = floor)
	worldMap := make([][]int, 20)
	for y := 0; y < 20; y++ {
		worldMap[y] = make([]int, 20)
		for x := 0; x < 20; x++ {
			// Border walls
			if x == 0 || x == 19 || y == 0 || y == 19 {
				worldMap[y][x] = 1
			} else {
				worldMap[y][x] = 1 // Default to wall
			}
		}
	}

	// Create a corridor
	for x := 5; x < 15; x++ {
		worldMap[10][x] = 2 // Horizontal corridor
	}
	for y := 5; y < 15; y++ {
		worldMap[y][10] = 2 // Vertical corridor
	}

	sys.GenerateTraps(worldMap, 12345)

	if len(sys.traps) == 0 {
		t.Error("GenerateTraps should create at least one trap")
	}

	// Verify trap positions are valid
	for _, trap := range sys.traps {
		x := int(trap.X)
		y := int(trap.Y)

		if x < 0 || x >= 20 || y < 0 || y >= 20 {
			t.Errorf("Trap at (%v, %v) is outside map bounds", trap.X, trap.Y)
		}

		if worldMap[y][x] != 2 {
			t.Errorf("Trap at (%v, %v) is not on floor tile", trap.X, trap.Y)
		}
	}
}

func TestGenerateTrapsGenreVariety(t *testing.T) {
	// Create a map with corridors
	worldMap := make([][]int, 20)
	for y := 0; y < 20; y++ {
		worldMap[y] = make([]int, 20)
		for x := 0; x < 20; x++ {
			if x == 0 || x == 19 || y == 0 || y == 19 {
				worldMap[y][x] = 1
			} else {
				worldMap[y][x] = 1 // Default to wall
			}
		}
	}

	// Create corridors
	for x := 5; x < 15; x++ {
		worldMap[10][x] = 2
	}
	for y := 5; y < 15; y++ {
		worldMap[y][10] = 2
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(12345)
			sys.SetGenre(genre)
			sys.GenerateTraps(worldMap, 12345)

			if len(sys.traps) == 0 {
				t.Error("No traps generated")
			}

			// Verify genre-appropriate traps
			genreTypes := GetGenreTraps(genre)
			for _, trap := range sys.traps {
				found := false
				for _, validType := range genreTypes {
					if trap.Type == validType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Trap type %v not valid for genre %s", trap.Type, genre)
				}
			}
		})
	}
}

func TestGenerateTrapsEmptyMap(t *testing.T) {
	sys := NewSystem(12345)

	// Empty map
	worldMap := [][]int{}
	sys.GenerateTraps(worldMap, 12345)

	if len(sys.traps) != 0 {
		t.Error("Should not generate traps on empty map")
	}
}

func TestSystemMultipleEntities(t *testing.T) {
	sys := NewSystem(12345)

	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	sys.AddTrap(trap)

	// Create ECS world
	w := engine.NewWorld()

	// Entity far away
	entity1 := w.AddEntity()
	w.AddComponent(entity1, &PositionComponent{PosX: 50.0, PosY: 50.0})

	// Entity on trap with health
	entity2 := w.AddEntity()
	w.AddComponent(entity2, &PositionComponent{PosX: 5.0, PosY: 5.0})
	w.AddComponent(entity2, &HealthComponent{Current: 100, Max: 100})

	// Entity also far away
	entity3 := w.AddEntity()
	w.AddComponent(entity3, &PositionComponent{PosX: 10.0, PosY: 10.0})

	sys.Update(w)

	// Only entity2 should trigger the trap
	if trap.State != StateTriggered {
		t.Error("Trap should be triggered by nearby entity")
	}

	// entity2 should have taken damage
	healthType := reflect.TypeOf((*HealthComponent)(nil))
	healthComp, _ := w.GetComponent(entity2, healthType)
	health := healthComp.(*HealthComponent)
	if health.Current >= 100 {
		t.Error("Entity on trap should have taken damage")
	}
}

func TestComponent(t *testing.T) {
	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	comp := &Component{Trap: trap}

	if comp.Type() != "trap" {
		t.Errorf("Component type = %v, want trap", comp.Type())
	}

	if comp.Trap != trap {
		t.Error("Component should store trap reference")
	}
}
