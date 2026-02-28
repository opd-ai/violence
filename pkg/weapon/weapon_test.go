package weapon

import (
	"math"
	"math/rand"
	"testing"
)

func TestNewArsenal(t *testing.T) {
	a := NewArsenal()
	if a == nil {
		t.Fatal("NewArsenal returned nil")
	}
	if len(a.Weapons) != 7 {
		t.Errorf("Expected 7 weapons, got %d", len(a.Weapons))
	}
	if a.CurrentSlot != 1 {
		t.Errorf("Expected CurrentSlot=1, got %d", a.CurrentSlot)
	}
	if a.Ammo == nil {
		t.Error("Ammo map is nil")
	}
	if a.Clips == nil {
		t.Error("Clips map is nil")
	}
}

func TestWeaponTypes(t *testing.T) {
	tests := []struct {
		slot         int
		expectedType WeaponType
		expectedName string
	}{
		{0, TypeMelee, "Fist"},
		{1, TypeHitscan, "Pistol"},
		{2, TypeHitscan, "Shotgun"},
		{3, TypeHitscan, "Chaingun"},
		{4, TypeProjectile, "Rocket Launcher"},
		{5, TypeProjectile, "Plasma Gun"},
		{6, TypeMelee, "Knife"},
	}

	a := NewArsenal()
	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			weapon := a.Weapons[tt.slot]
			if weapon.Type != tt.expectedType {
				t.Errorf("Weapon %s: expected type %d, got %d", weapon.Name, tt.expectedType, weapon.Type)
			}
			if weapon.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, weapon.Name)
			}
		})
	}
}

func TestHitscanFire(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1) // Pistol

	raycastCalled := false
	mockRaycast := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		raycastCalled = true
		// Verify parameters
		if x != 5.0 || y != 5.0 {
			t.Errorf("Expected pos (5, 5), got (%f, %f)", x, y)
		}
		if math.Abs(dx-1.0) > 0.01 || math.Abs(dy) > 0.01 {
			t.Errorf("Expected dir (1, 0), got (%f, %f)", dx, dy)
		}
		if maxDist != 100 {
			t.Errorf("Expected range 100, got %f", maxDist)
		}
		return true, 10.0, 15.0, 5.0, 42
	}

	results := a.Fire(5.0, 5.0, 1.0, 0.0, mockRaycast)

	if !raycastCalled {
		t.Error("Raycast function was not called")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 hit result, got %d", len(results))
	}
	if !results[0].Hit {
		t.Error("Expected hit=true")
	}
	if results[0].Distance != 10.0 {
		t.Errorf("Expected distance 10, got %f", results[0].Distance)
	}
	if results[0].Damage != 15.0 {
		t.Errorf("Expected damage 15, got %f", results[0].Damage)
	}
	if results[0].EntityID != 42 {
		t.Errorf("Expected entityID 42, got %d", results[0].EntityID)
	}

	// Check ammo consumed
	if a.Clips[1] != 11 {
		t.Errorf("Expected 11 bullets in clip after firing, got %d", a.Clips[1])
	}
}

func TestShotgunMultiRay(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(2) // Shotgun (7 rays)

	rayCount := 0
	mockRaycast := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		rayCount++
		return true, 5.0, 10.0, 5.0, 0
	}

	results := a.Fire(0, 0, 1.0, 0, mockRaycast)

	if rayCount != 7 {
		t.Errorf("Expected 7 raycasts for shotgun, got %d", rayCount)
	}
	if len(results) != 7 {
		t.Errorf("Expected 7 hit results, got %d", len(results))
	}

	// Check ammo consumed (1 shell for all rays)
	if a.Clips[2] != 7 {
		t.Errorf("Expected 7 shells in clip after firing, got %d", a.Clips[2])
	}
}

func TestFireCooldown(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1) // Pistol, FireRate=15

	mockRaycast := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		return true, 10, 10, 10, 0
	}

	// First fire should succeed
	results := a.Fire(0, 0, 1, 0, mockRaycast)
	if results == nil {
		t.Fatal("First fire failed")
	}

	// Second fire immediately should fail (cooldown)
	results = a.Fire(0, 0, 1, 0, mockRaycast)
	if results != nil {
		t.Error("Fire should be blocked by cooldown")
	}

	// Update frames to clear cooldown
	for i := 0; i < 15; i++ {
		a.Update()
	}

	// Third fire should succeed
	results = a.Fire(0, 0, 1, 0, mockRaycast)
	if results == nil {
		t.Error("Fire should succeed after cooldown")
	}
}

func TestFireOutOfAmmo(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1)  // Pistol
	a.Clips[1] = 0 // Empty clip

	mockRaycast := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		return true, 10, 10, 10, 0
	}

	results := a.Fire(0, 0, 1, 0, mockRaycast)
	if results != nil {
		t.Error("Fire should fail when out of ammo")
	}
}

func TestMeleeFire(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(6) // Knife

	mockRaycast := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		if maxDist != 1.5 {
			t.Errorf("Expected melee range 1.5, got %f", maxDist)
		}
		return true, 1.0, 5.0, 5.0, 10
	}

	results := a.Fire(0, 0, 1, 0, mockRaycast)
	if results == nil {
		t.Fatal("Melee fire failed")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 melee hit, got %d", len(results))
	}
	if results[0].Damage != 25.0 {
		t.Errorf("Expected knife damage 25, got %f", results[0].Damage)
	}

	// Melee should not consume ammo
	if a.Clips[6] != 0 {
		t.Errorf("Melee weapon should have 0 clip, got %d", a.Clips[6])
	}
}

func TestReload(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1)  // Pistol, ClipSize=12
	a.Clips[1] = 5 // Partially empty
	a.Ammo["bullets"] = 100

	success := a.Reload()
	if !success {
		t.Error("Reload should succeed")
	}
	if a.Clips[1] != 12 {
		t.Errorf("Expected full clip (12), got %d", a.Clips[1])
	}
	if a.Ammo["bullets"] != 93 {
		t.Errorf("Expected 93 bullets in pool, got %d", a.Ammo["bullets"])
	}
}

func TestReloadPartial(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1) // Pistol, ClipSize=12
	a.Clips[1] = 5
	a.Ammo["bullets"] = 3 // Not enough for full reload

	success := a.Reload()
	if !success {
		t.Error("Reload should succeed even if partial")
	}
	if a.Clips[1] != 8 {
		t.Errorf("Expected 8 bullets in clip, got %d", a.Clips[1])
	}
	if a.Ammo["bullets"] != 0 {
		t.Errorf("Expected 0 bullets in pool, got %d", a.Ammo["bullets"])
	}
}

func TestReloadNoAmmo(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1)
	a.Clips[1] = 5
	a.Ammo["bullets"] = 0

	success := a.Reload()
	if success {
		t.Error("Reload should fail with no ammo")
	}
	if a.Clips[1] != 5 {
		t.Errorf("Clip should remain at 5, got %d", a.Clips[1])
	}
}

func TestReloadMelee(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(0) // Fist

	success := a.Reload()
	if success {
		t.Error("Melee weapons should not reload")
	}
}

func TestSwitchWeapon(t *testing.T) {
	a := NewArsenal()

	tests := []struct {
		slot    int
		success bool
	}{
		{0, true},
		{3, true},
		{6, true},
		{-1, false},
		{7, false},
		{100, false},
	}

	for _, tt := range tests {
		success := a.SwitchTo(tt.slot)
		if success != tt.success {
			t.Errorf("SwitchTo(%d): expected success=%v, got %v", tt.slot, tt.success, success)
		}
		if success && a.CurrentSlot != tt.slot {
			t.Errorf("CurrentSlot should be %d, got %d", tt.slot, a.CurrentSlot)
		}
	}
}

func TestAddAmmo(t *testing.T) {
	a := NewArsenal()
	a.Ammo["bullets"] = 50

	a.AddAmmo("bullets", 20)
	if a.Ammo["bullets"] != 70 {
		t.Errorf("Expected 70 bullets, got %d", a.Ammo["bullets"])
	}

	a.AddAmmo("shells", 10)
	if a.Ammo["shells"] != 18 {
		t.Errorf("Expected 18 shells, got %d", a.Ammo["shells"])
	}
}

func TestGenreWeaponNames(t *testing.T) {
	tests := []struct {
		genre     string
		slot1Name string
		slot2Name string
		slot4Name string
	}{
		{"fantasy", "Crossbow", "Blunderbuss", "Explosive Orb"},
		{"scifi", "Blaster", "Scatter Cannon", "Missile Launcher"},
		{"horror", "Revolver", "Sawed-off Shotgun", "Grenade Launcher"},
		{"cyberpunk", "Smart Pistol", "Auto-Shotgun", "Rocket Pod"},
		{"postapoc", "Makeshift Pistol", "Pipe Shotgun", "Improvised Launcher"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			a := NewArsenal()
			a.SetGenre(tt.genre)

			if a.Weapons[1].Name != tt.slot1Name {
				t.Errorf("Genre %s slot 1: expected %s, got %s", tt.genre, tt.slot1Name, a.Weapons[1].Name)
			}
			if a.Weapons[2].Name != tt.slot2Name {
				t.Errorf("Genre %s slot 2: expected %s, got %s", tt.genre, tt.slot2Name, a.Weapons[2].Name)
			}
			if a.Weapons[4].Name != tt.slot4Name {
				t.Errorf("Genre %s slot 4: expected %s, got %s", tt.genre, tt.slot4Name, a.Weapons[4].Name)
			}
		})
	}
}

func TestFireProjectile(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(4) // Rocket launcher
	rng := rand.New(rand.NewSource(42))

	velX, velY, spawned := a.FireProjectile(0, 0, 1.0, 0, rng)

	if !spawned {
		t.Fatal("Projectile should spawn")
	}
	if math.Abs(velX-0.25) > 0.01 {
		t.Errorf("Expected velX ~0.25, got %f", velX)
	}
	if math.Abs(velY) > 0.01 {
		t.Errorf("Expected velY ~0, got %f", velY)
	}

	// Check ammo consumed
	if a.Clips[4] != 4 {
		t.Errorf("Expected 4 rockets after firing, got %d", a.Clips[4])
	}
}

func TestFireProjectileNotProjectileWeapon(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1) // Pistol (hitscan)
	rng := rand.New(rand.NewSource(42))

	_, _, spawned := a.FireProjectile(0, 0, 1, 0, rng)

	if spawned {
		t.Error("Hitscan weapon should not spawn projectile")
	}
}

func TestFireProjectileCooldown(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(4) // Rocket launcher, FireRate=45
	rng := rand.New(rand.NewSource(42))

	// First fire
	_, _, spawned := a.FireProjectile(0, 0, 1, 0, rng)
	if !spawned {
		t.Fatal("First fire should succeed")
	}

	// Immediate second fire (should fail)
	_, _, spawned = a.FireProjectile(0, 0, 1, 0, rng)
	if spawned {
		t.Error("Fire should be blocked by cooldown")
	}

	// Update frames
	for i := 0; i < 45; i++ {
		a.Update()
	}

	// Third fire should succeed
	_, _, spawned = a.FireProjectile(0, 0, 1, 0, rng)
	if !spawned {
		t.Error("Fire should succeed after cooldown")
	}
}

func TestGetCurrentWeapon(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(3)

	weapon := a.GetCurrentWeapon()
	if weapon.Name != "Chaingun" {
		t.Errorf("Expected Chaingun, got %s", weapon.Name)
	}
}

func TestHitResultBeyondRange(t *testing.T) {
	a := NewArsenal()
	a.SwitchTo(1) // Pistol, range=100

	mockRaycast := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		return true, 150.0, 200, 200, 5 // Hit beyond range
	}

	results := a.Fire(0, 0, 1, 0, mockRaycast)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Damage should be 0 if beyond range
	if results[0].Damage != 0 {
		t.Errorf("Expected 0 damage for hit beyond range, got %f", results[0].Damage)
	}
}

func TestUpdate(t *testing.T) {
	a := NewArsenal()
	a.FramesSinceFire[1] = 5

	a.Update()

	if a.FramesSinceFire[1] != 6 {
		t.Errorf("Expected FramesSinceFire to increment to 6, got %d", a.FramesSinceFire[1])
	}
}
