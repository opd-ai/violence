package engine

import (
	"reflect"
	"testing"
)

func TestNewPlayerEntity(t *testing.T) {
	w := NewWorld()
	e := w.NewPlayerEntity(5.0, 10.0)

	t.Run("entity exists in world", func(t *testing.T) {
		_, exists := w.components[e]
		if !exists {
			t.Error("NewPlayerEntity() entity not found in world")
		}
	})

	t.Run("has Position component", func(t *testing.T) {
		comp, ok := w.GetComponent(e, reflect.TypeOf(&Position{}))
		if !ok {
			t.Fatal("Player missing Position component")
		}
		pos := comp.(*Position)
		if pos.X != 5.0 || pos.Y != 10.0 {
			t.Errorf("Position = (%v, %v), want (5.0, 10.0)", pos.X, pos.Y)
		}
	})

	t.Run("has Health component", func(t *testing.T) {
		comp, ok := w.GetComponent(e, reflect.TypeOf(&Health{}))
		if !ok {
			t.Fatal("Player missing Health component")
		}
		health := comp.(*Health)
		if health.Current != 100 || health.Max != 100 {
			t.Errorf("Health = (%v/%v), want (100/100)", health.Current, health.Max)
		}
	})

	t.Run("has Armor component", func(t *testing.T) {
		comp, ok := w.GetComponent(e, reflect.TypeOf(&Armor{}))
		if !ok {
			t.Fatal("Player missing Armor component")
		}
		armor := comp.(*Armor)
		if armor.Value != 0 {
			t.Errorf("Armor.Value = %v, want 0", armor.Value)
		}
	})

	t.Run("has Inventory component", func(t *testing.T) {
		comp, ok := w.GetComponent(e, reflect.TypeOf(&Inventory{}))
		if !ok {
			t.Fatal("Player missing Inventory component")
		}
		inv := comp.(*Inventory)
		if len(inv.Items) != 0 {
			t.Errorf("Inventory.Items length = %v, want 0", len(inv.Items))
		}
		if inv.Credits != 0 {
			t.Errorf("Inventory.Credits = %v, want 0", inv.Credits)
		}
	})

	t.Run("has Camera component", func(t *testing.T) {
		comp, ok := w.GetComponent(e, reflect.TypeOf(&Camera{}))
		if !ok {
			t.Fatal("Player missing Camera component")
		}
		cam := comp.(*Camera)
		if cam.DirX != 0 || cam.DirY != -1 {
			t.Errorf("Camera direction = (%v, %v), want (0, -1)", cam.DirX, cam.DirY)
		}
		if cam.FOV != 66.0 {
			t.Errorf("Camera.FOV = %v, want 66.0", cam.FOV)
		}
		if cam.PitchRadians != 0 {
			t.Errorf("Camera.PitchRadians = %v, want 0", cam.PitchRadians)
		}
	})

	t.Run("has Input component", func(t *testing.T) {
		comp, ok := w.GetComponent(e, reflect.TypeOf(&Input{}))
		if !ok {
			t.Fatal("Player missing Input component")
		}
		input := comp.(*Input)
		if input.Forward || input.Backward || input.Left || input.Right {
			t.Error("Input movement should be false initially")
		}
		if input.Fire || input.AltFire || input.Interact || input.Reload {
			t.Error("Input actions should be false initially")
		}
	})

	t.Run("has correct archetype bits", func(t *testing.T) {
		mask := w.GetArchetype(e)
		expectedBits := []ComponentID{
			ComponentIDPosition,
			ComponentIDHealth,
			ComponentIDArmor,
			ComponentIDInventory,
			ComponentIDCamera,
			ComponentIDInput,
		}
		for _, bit := range expectedBits {
			if mask&(1<<uint64(bit)) == 0 {
				t.Errorf("Archetype missing bit for component %v", bit)
			}
		}
	})
}

func TestIsPlayer(t *testing.T) {
	w := NewWorld()

	t.Run("entity with all player components is player", func(t *testing.T) {
		e := w.NewPlayerEntity(0, 0)
		if !w.IsPlayer(e) {
			t.Error("IsPlayer() = false for full player entity, want true")
		}
	})

	t.Run("entity missing components is not player", func(t *testing.T) {
		e := w.AddEntity()
		w.AddComponent(e, &Position{X: 0, Y: 0})
		w.AddArchetypeComponent(e, ComponentIDPosition)
		w.AddComponent(e, &Health{Current: 100, Max: 100})
		w.AddArchetypeComponent(e, ComponentIDHealth)

		if w.IsPlayer(e) {
			t.Error("IsPlayer() = true for incomplete entity, want false")
		}
	})

	t.Run("empty entity is not player", func(t *testing.T) {
		e := w.AddEntity()
		if w.IsPlayer(e) {
			t.Error("IsPlayer() = true for empty entity, want false")
		}
	})

	t.Run("entity with extra components is still player", func(t *testing.T) {
		e := w.NewPlayerEntity(0, 0)
		w.AddComponent(e, &struct{ Extra int }{Extra: 42})
		w.AddArchetypeComponent(e, ComponentIDWeapon)

		if !w.IsPlayer(e) {
			t.Error("IsPlayer() = false for player with extra components, want true")
		}
	})
}

func TestPlayerComponents_DefaultValues(t *testing.T) {
	t.Run("Position defaults", func(t *testing.T) {
		p := &Position{}
		if p.X != 0 || p.Y != 0 {
			t.Errorf("Position zero value = (%v, %v), want (0, 0)", p.X, p.Y)
		}
	})

	t.Run("Health defaults", func(t *testing.T) {
		h := &Health{}
		if h.Current != 0 || h.Max != 0 {
			t.Errorf("Health zero value = (%v/%v), want (0/0)", h.Current, h.Max)
		}
	})

	t.Run("Armor defaults", func(t *testing.T) {
		a := &Armor{}
		if a.Value != 0 {
			t.Errorf("Armor zero value = %v, want 0", a.Value)
		}
	})

	t.Run("Inventory defaults", func(t *testing.T) {
		inv := &Inventory{}
		if inv.Items != nil {
			t.Errorf("Inventory.Items zero value = %v, want nil", inv.Items)
		}
		if inv.Credits != 0 {
			t.Errorf("Inventory.Credits zero value = %v, want 0", inv.Credits)
		}
	})

	t.Run("Camera defaults", func(t *testing.T) {
		c := &Camera{}
		if c.DirX != 0 || c.DirY != 0 {
			t.Errorf("Camera direction zero value = (%v, %v), want (0, 0)", c.DirX, c.DirY)
		}
		if c.FOV != 0 {
			t.Errorf("Camera.FOV zero value = %v, want 0", c.FOV)
		}
	})

	t.Run("Input defaults", func(t *testing.T) {
		input := &Input{}
		if input.Forward || input.Fire {
			t.Error("Input zero value should have all flags false")
		}
	})
}

func TestPlayerEntityQuery(t *testing.T) {
	w := NewWorld()

	// Create multiple entities
	player1 := w.NewPlayerEntity(0, 0)
	player2 := w.NewPlayerEntity(10, 10)

	enemy := w.AddEntity()
	w.AddComponent(enemy, &Position{X: 5, Y: 5})
	w.AddArchetypeComponent(enemy, ComponentIDPosition)
	w.AddComponent(enemy, &Health{Current: 50, Max: 50})
	w.AddArchetypeComponent(enemy, ComponentIDHealth)

	t.Run("query for all entities with Position and Health", func(t *testing.T) {
		it := w.QueryWithBitmask(ComponentIDPosition, ComponentIDHealth)
		count := 0
		entities := make(map[Entity]bool)
		for it.Next() {
			count++
			entities[it.Entity()] = true
		}

		if count != 3 {
			t.Errorf("Query matched %d entities, want 3", count)
		}
		if !entities[player1] || !entities[player2] || !entities[enemy] {
			t.Error("Query missing expected entities")
		}
	})

	t.Run("query for player-specific components", func(t *testing.T) {
		it := w.QueryWithBitmask(ComponentIDPosition, ComponentIDCamera, ComponentIDInput)
		count := 0
		entities := make(map[Entity]bool)
		for it.Next() {
			count++
			entities[it.Entity()] = true
		}

		if count != 2 {
			t.Errorf("Query matched %d entities, want 2 (players only)", count)
		}
		if !entities[player1] || !entities[player2] {
			t.Error("Query missing player entities")
		}
		if entities[enemy] {
			t.Error("Query incorrectly matched enemy entity")
		}
	})
}

func BenchmarkNewPlayerEntity(b *testing.B) {
	w := NewWorld()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.NewPlayerEntity(0, 0)
	}
}

func BenchmarkIsPlayer(b *testing.B) {
	w := NewWorld()
	e := w.NewPlayerEntity(0, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.IsPlayer(e)
	}
}
