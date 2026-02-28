package destruct

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem()
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.objects == nil {
		t.Fatal("objects map not initialized")
	}
}

func TestNewDestructible(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 5.0, 10.0)
	if d == nil {
		t.Fatal("NewDestructible returned nil")
	}
	if d.ID != "test1" {
		t.Fatalf("wrong ID: got %q", d.ID)
	}
	if d.Type != "barrel" {
		t.Fatalf("wrong type: got %q", d.Type)
	}
	if d.Health != 100 {
		t.Fatalf("wrong health: got %f", d.Health)
	}
	if d.MaxHealth != 100 {
		t.Fatalf("wrong max health: got %f", d.MaxHealth)
	}
	if d.Destroyed {
		t.Fatal("should not be destroyed")
	}
}

func TestSystem_Add(t *testing.T) {
	sys := NewSystem()
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	sys.Add(d)

	retrieved, ok := sys.Get("test1")
	if !ok {
		t.Fatal("object not found after add")
	}
	if retrieved.ID != "test1" {
		t.Fatalf("wrong object: got ID %q", retrieved.ID)
	}
}

func TestSystem_Remove(t *testing.T) {
	sys := NewSystem()
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	sys.Add(d)

	sys.Remove("test1")

	_, ok := sys.Get("test1")
	if ok {
		t.Fatal("object should be removed")
	}
}

func TestSystem_GetAll(t *testing.T) {
	sys := NewSystem()
	sys.Add(NewDestructible("1", "barrel", 100, 0, 0))
	sys.Add(NewDestructible("2", "crate", 50, 0, 0))

	all := sys.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(all))
	}
}

func TestDestructible_Damage(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)

	destroyed := d.Damage(30)
	if destroyed {
		t.Fatal("should not be destroyed yet")
	}
	if d.GetHealth() != 70 {
		t.Fatalf("wrong health: got %f", d.GetHealth())
	}
	if d.IsDestroyed() {
		t.Fatal("should not be destroyed")
	}
}

func TestDestructible_DamageDestroy(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)

	destroyed := d.Damage(150)
	if !destroyed {
		t.Fatal("should be destroyed")
	}
	if d.GetHealth() != 0 {
		t.Fatalf("health should be 0, got %f", d.GetHealth())
	}
	if !d.IsDestroyed() {
		t.Fatal("should be marked destroyed")
	}
}

func TestDestructible_DamageAlreadyDestroyed(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	d.Destroy()

	destroyed := d.Damage(50)
	if destroyed {
		t.Fatal("Damage should return false for already destroyed object")
	}
}

func TestDestructible_Destroy(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	d.Destroy()

	if !d.IsDestroyed() {
		t.Fatal("should be destroyed")
	}
	if d.GetHealth() != 0 {
		t.Fatalf("health should be 0, got %f", d.GetHealth())
	}
}

func TestDestructible_Repair(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	d.Damage(50)

	d.Repair(20)
	if d.GetHealth() != 70 {
		t.Fatalf("wrong health after repair: got %f", d.GetHealth())
	}
}

func TestDestructible_RepairOverMax(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	d.Damage(50)

	d.Repair(100)
	if d.GetHealth() != 100 {
		t.Fatalf("health should cap at max: got %f", d.GetHealth())
	}
}

func TestDestructible_RepairDestroyed(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	d.Destroy()

	d.Repair(50)
	// Should not repair destroyed object
	if d.GetHealth() != 0 {
		t.Fatal("destroyed object should not be repaired")
	}
}

func TestDestructible_AddDropItem(t *testing.T) {
	d := NewDestructible("test1", "barrel", 100, 0, 0)
	d.AddDropItem("ammo_bullets")
	d.AddDropItem("health_pack")

	items := d.GetDropItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 drop items, got %d", len(items))
	}
	if items[0] != "ammo_bullets" {
		t.Fatalf("wrong first item: got %q", items[0])
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}
