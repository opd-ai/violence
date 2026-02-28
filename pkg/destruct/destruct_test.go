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

func TestNewBreakableWall(t *testing.T) {
	wall := NewBreakableWall("wall1", 10.0, 20.0, 150.0, true)
	if wall == nil {
		t.Fatal("NewBreakableWall returned nil")
	}
	if wall.ID != "wall1" {
		t.Fatalf("wrong ID: got %q", wall.ID)
	}
	if wall.Health != 150.0 {
		t.Fatalf("wrong health: got %f", wall.Health)
	}
	if !wall.RevealsPath {
		t.Fatal("should reveal path")
	}
	if wall.Destroyed {
		t.Fatal("should not be destroyed")
	}
}

func TestBreakableWall_Damage(t *testing.T) {
	wall := NewBreakableWall("wall1", 10.0, 20.0, 100.0, true)

	destroyed := wall.Damage(50.0)
	if destroyed {
		t.Fatal("should not be destroyed yet")
	}
	if wall.GetHealth() != 50.0 {
		t.Fatalf("wrong health after damage: got %f", wall.GetHealth())
	}

	destroyed = wall.Damage(50.0)
	if !destroyed {
		t.Fatal("should be destroyed")
	}
	if !wall.IsDestroyed() {
		t.Fatal("IsDestroyed should return true")
	}
}

func TestBreakableWall_SetRevealedPath(t *testing.T) {
	wall := NewBreakableWall("wall1", 10.0, 20.0, 100.0, false)

	wall.SetRevealedPath(5, 7)

	x, y, reveals := wall.GetRevealedPath()
	if !reveals {
		t.Fatal("should reveal path after SetRevealedPath")
	}
	if x != 5 || y != 7 {
		t.Fatalf("wrong path coords: got (%d, %d)", x, y)
	}
}

func TestBreakableWall_DamageWhenDestroyed(t *testing.T) {
	wall := NewBreakableWall("wall1", 10.0, 20.0, 50.0, true)
	wall.Damage(50.0) // destroy it

	destroyed := wall.Damage(10.0) // try to damage again
	if destroyed {
		t.Fatal("damaging destroyed wall should return false")
	}
}

func TestNewDestructibleObject(t *testing.T) {
	obj := NewDestructibleObject("barrel1", "barrel", 100.0, 5.0, 10.0, true)
	if obj == nil {
		t.Fatal("NewDestructibleObject returned nil")
	}
	if obj.ID != "barrel1" {
		t.Fatalf("wrong ID: got %q", obj.ID)
	}
	if !obj.Explosive {
		t.Fatal("should be explosive")
	}
	if obj.ExplosionRange <= 0 {
		t.Fatal("explosion range not set")
	}
}

func TestDestructibleObject_GetExplosionTargets(t *testing.T) {
	obj1 := NewDestructibleObject("barrel1", "barrel", 100.0, 0.0, 0.0, true)
	obj2 := NewDestructibleObject("barrel2", "barrel", 100.0, 2.0, 0.0, true)  // within range
	obj3 := NewDestructibleObject("barrel3", "barrel", 100.0, 10.0, 0.0, true) // out of range

	allObjects := []*DestructibleObject{obj1, obj2, obj3}

	// Before destruction, no targets
	targets := obj1.GetExplosionTargets(allObjects)
	if len(targets) != 0 {
		t.Fatal("should have no targets before destruction")
	}

	// After destruction
	obj1.Damage(100.0)
	targets = obj1.GetExplosionTargets(allObjects)

	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].ID != "barrel2" {
		t.Fatalf("wrong target: got %q", targets[0].ID)
	}
}

func TestDestructibleObject_NonExplosive(t *testing.T) {
	obj := NewDestructibleObject("crate1", "crate", 100.0, 0.0, 0.0, false)

	if obj.Explosive {
		t.Fatal("should not be explosive")
	}
	if obj.ChainReaction {
		t.Fatal("should not have chain reaction")
	}
}

func TestNewDebris(t *testing.T) {
	debris := NewDebris("debris1", 5.0, 10.0, "rubble", true, 5.0)
	if debris == nil {
		t.Fatal("NewDebris returned nil")
	}
	if debris.ID != "debris1" {
		t.Fatalf("wrong ID: got %q", debris.ID)
	}
	if debris.Material != "rubble" {
		t.Fatalf("wrong material: got %q", debris.Material)
	}
	if !debris.BlocksPath {
		t.Fatal("should block path")
	}
	if debris.TimeRemaining != 5.0 {
		t.Fatalf("wrong time remaining: got %f", debris.TimeRemaining)
	}
}

func TestDebris_Update(t *testing.T) {
	debris := NewDebris("debris1", 5.0, 10.0, "rubble", true, 5.0)

	cleared := debris.Update(2.0)
	if cleared {
		t.Fatal("should not be cleared yet")
	}
	if debris.IsCleared() {
		t.Fatal("IsCleared should return false")
	}

	cleared = debris.Update(3.0)
	if !cleared {
		t.Fatal("should be cleared now")
	}
	if !debris.IsCleared() {
		t.Fatal("IsCleared should return true")
	}
}

func TestDebris_GetProgress(t *testing.T) {
	debris := NewDebris("debris1", 5.0, 10.0, "rubble", true, 10.0)

	if debris.GetProgress() != 0.0 {
		t.Fatal("initial progress should be 0")
	}

	debris.Update(5.0) // half time
	progress := debris.GetProgress()
	if progress < 0.4 || progress > 0.6 {
		t.Fatalf("progress should be around 0.5, got %f", progress)
	}

	debris.Update(5.0) // full time
	progress = debris.GetProgress()
	if progress != 1.0 {
		t.Fatalf("final progress should be 1.0, got %f", progress)
	}
}

func TestDebris_GetProgressZeroMaxTime(t *testing.T) {
	debris := NewDebris("debris1", 5.0, 10.0, "rubble", true, 0.0)

	// Should not panic, should return 1.0
	progress := debris.GetProgress()
	if progress != 1.0 {
		t.Fatalf("progress with zero max time should be 1.0, got %f", progress)
	}
}

func TestGetDebrisMaterial(t *testing.T) {
	tests := []struct {
		genre      string
		objectType string
		want       string
	}{
		{"fantasy", "wall", "stone rubble"},
		{"fantasy", "barrel", "wooden splinters"},
		{"scifi", "wall", "hull shards"},
		{"scifi", "crate", "alloy pieces"},
		{"horror", "barrel", "rotted wood"},
		{"cyberpunk", "wall", "shattered glass"},
		{"postapoc", "wall", "concrete chunks"},
		{"unknown", "wall", "stone rubble"}, // falls back to fantasy
		{"fantasy", "unknown", "debris"},    // unknown object type
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.objectType, func(t *testing.T) {
			got := GetDebrisMaterial(tt.genre, tt.objectType)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetDebrisMaterialAllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	objectTypes := []string{"wall", "barrel", "crate", "door", "unknown"}

	for _, genre := range genres {
		for _, objType := range objectTypes {
			t.Run(genre+"_"+objType, func(t *testing.T) {
				material := GetDebrisMaterial(genre, objType)
				if material == "" {
					t.Fatalf("genre %s, type %s: empty material", genre, objType)
				}
			})
		}
	}
}

func TestDestructibleObject_ExplosionChain(t *testing.T) {
	// Create a chain of explosive objects
	obj1 := NewDestructibleObject("barrel1", "barrel", 50.0, 0.0, 0.0, true)
	obj2 := NewDestructibleObject("barrel2", "barrel", 50.0, 2.0, 0.0, true)
	obj3 := NewDestructibleObject("barrel3", "barrel", 50.0, 4.0, 0.0, true)

	allObjects := []*DestructibleObject{obj1, obj2, obj3}

	// Destroy first barrel
	obj1.Damage(50.0)

	// Get targets for explosion
	targets := obj1.GetExplosionTargets(allObjects)
	if len(targets) == 0 {
		t.Fatal("should have at least one target")
	}

	// Damage targets
	for _, target := range targets {
		target.Damage(50.0) // destroy them
	}

	// obj2 should now be destroyed and have its own targets
	if !obj2.IsDestroyed() {
		t.Fatal("obj2 should be destroyed")
	}

	targets2 := obj2.GetExplosionTargets(allObjects)
	if len(targets2) == 0 {
		t.Fatal("obj2 should have explosion targets")
	}
}
