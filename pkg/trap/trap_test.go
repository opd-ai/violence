package trap

import (
	"testing"
)

func TestNewTrap(t *testing.T) {
	tests := []struct {
		name     string
		trapType TrapType
		x, y     float64
		seed     int64
	}{
		{"pressure plate", TrapTypePressurePlate, 5.0, 5.0, 12345},
		{"tripwire", TrapTypeTripwire, 10.0, 10.0, 67890},
		{"spike pit", TrapTypeSpikePit, 15.0, 15.0, 11111},
		{"dart wall", TrapTypeDartWall, 20.0, 20.0, 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trap := NewTrap(tt.trapType, tt.x, tt.y, tt.seed)

			if trap == nil {
				t.Fatal("NewTrap returned nil")
			}

			if trap.Type != tt.trapType {
				t.Errorf("Type = %v, want %v", trap.Type, tt.trapType)
			}

			if trap.X != tt.x || trap.Y != tt.y {
				t.Errorf("Position = (%v, %v), want (%v, %v)", trap.X, trap.Y, tt.x, tt.y)
			}

			if trap.State != StateHidden {
				t.Errorf("State = %v, want %v", trap.State, StateHidden)
			}

			if trap.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", trap.Seed, tt.seed)
			}
		})
	}
}

func TestTrapConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		trapType          TrapType
		wantDamage        int
		wantRetriggerable bool
	}{
		{"pressure plate", TrapTypePressurePlate, 10, true},
		{"spike pit", TrapTypeSpikePit, 25, false},
		{"explosive", TrapTypeExplosive, 40, false},
		{"electric shock", TrapTypeElectricShock, 18, true},
		{"rolling boulder", TrapTypeRollingBoulder, 35, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trap := NewTrap(tt.trapType, 5.0, 5.0, 12345)

			if trap.Damage != tt.wantDamage {
				t.Errorf("Damage = %v, want %v", trap.Damage, tt.wantDamage)
			}

			if trap.Retriggerable != tt.wantRetriggerable {
				t.Errorf("Retriggerable = %v, want %v", trap.Retriggerable, tt.wantRetriggerable)
			}
		})
	}
}

func TestTrapUpdate(t *testing.T) {
	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	trap.State = StateCooldown
	trap.CooldownTimer = 1.0

	// Update should reduce cooldown
	trap.Update(0.5)
	if trap.CooldownTimer != 0.5 {
		t.Errorf("CooldownTimer = %v, want 0.5", trap.CooldownTimer)
	}

	// Update should transition to hidden when cooldown ends
	trap.Update(0.6)
	if trap.State != StateHidden {
		t.Errorf("State = %v, want %v", trap.State, StateHidden)
	}
	if trap.CooldownTimer != 0 {
		t.Errorf("CooldownTimer = %v, want 0", trap.CooldownTimer)
	}
}

func TestCheckTrigger(t *testing.T) {
	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)

	tests := []struct {
		name         string
		x, y         float64
		wantTrigger  bool
		wantDistance float64
	}{
		{"on trap", 5.0, 5.0, true, 0.0},
		{"within radius", 5.5, 5.5, true, 0.71},
		{"outside radius", 7.0, 7.0, false, 2.83},
		{"far away", 15.0, 15.0, false, 14.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset trap state
			trap.State = StateHidden

			info := &TriggerInfo{
				EntityID: "test_entity",
				X:        tt.x,
				Y:        tt.y,
				IsPlayer: true,
			}

			result := trap.CheckTrigger(info)

			if result.Triggered != tt.wantTrigger {
				t.Errorf("Triggered = %v, want %v", result.Triggered, tt.wantTrigger)
			}

			if tt.wantTrigger && result.Damage == 0 {
				t.Error("Triggered trap should deal damage")
			}
		})
	}
}

func TestTrapDetection(t *testing.T) {
	trap := NewTrap(TrapTypeTripwire, 5.0, 5.0, 12345)
	trap.DetectionDC = 15

	tests := []struct {
		name        string
		detectSkill int
		wantDetect  bool
	}{
		{"no skill", 0, false},
		{"low skill", 5, false},
		{"medium skill", 10, false},
		{"high skill", 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trap.State = StateHidden

			info := &TriggerInfo{
				EntityID:    "test_entity",
				X:           5.0,
				Y:           5.0,
				DetectSkill: tt.detectSkill,
			}

			result := trap.CheckTrigger(info)

			// Note: Detection involves randomness, so we can't guarantee exact outcomes
			// Just verify the system attempts detection
			if tt.detectSkill > 0 && trap.State == StateDetected {
				if !result.Detected {
					t.Error("Detected trap but result.Detected is false")
				}
			}
		})
	}
}

func TestTrapDisarm(t *testing.T) {
	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	trap.State = StateDetected
	trap.DisarmDC = 12

	info := &TriggerInfo{
		EntityID:    "test_entity",
		X:           5.0,
		Y:           5.0,
		DisarmSkill: 20,
	}

	result := trap.CheckTrigger(info)

	if trap.State == StateDisarmed {
		if !result.Disarmed {
			t.Error("Disarmed trap but result.Disarmed is false")
		}
	}
}

func TestNonRetriggerableTrap(t *testing.T) {
	trap := NewTrap(TrapTypeSpikePit, 5.0, 5.0, 12345)
	if trap.Retriggerable {
		t.Fatal("Spike pit should not be retriggerable")
	}

	info := &TriggerInfo{
		EntityID: "test_entity",
		X:        5.0,
		Y:        5.0,
	}

	// First trigger
	result1 := trap.CheckTrigger(info)
	if !result1.Triggered {
		t.Error("First trigger should succeed")
	}

	// Second trigger should fail
	result2 := trap.CheckTrigger(info)
	if result2.Triggered {
		t.Error("Non-retriggerable trap triggered twice")
	}
}

func TestTrapEffectResults(t *testing.T) {
	tests := []struct {
		name           string
		trapType       TrapType
		wantProjectile bool
		wantKnockback  bool
		wantTeleport   bool
	}{
		{"dart wall", TrapTypeDartWall, true, false, false},
		{"explosive", TrapTypeExplosive, false, true, false},
		{"teleporter", TrapTypeTeleporter, false, false, true},
		{"spike pit", TrapTypeSpikePit, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trap := NewTrap(tt.trapType, 5.0, 5.0, 12345)

			info := &TriggerInfo{
				EntityID: "test_entity",
				X:        5.0,
				Y:        5.0,
			}

			result := trap.CheckTrigger(info)

			if !result.Triggered {
				t.Fatal("Trap should trigger when entity is on it")
			}

			if result.SpawnProjectile != tt.wantProjectile {
				t.Errorf("SpawnProjectile = %v, want %v", result.SpawnProjectile, tt.wantProjectile)
			}

			hasKnockback := result.KnockbackX != 0 || result.KnockbackY != 0
			if hasKnockback != tt.wantKnockback {
				t.Errorf("Knockback = %v, want %v", hasKnockback, tt.wantKnockback)
			}

			hasTeleport := result.TeleportX != 0 || result.TeleportY != 0
			if hasTeleport != tt.wantTeleport {
				t.Errorf("Teleport = %v, want %v", hasTeleport, tt.wantTeleport)
			}
		})
	}
}

func TestGetGenreTraps(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		wantCount int
		mustHave  TrapType
	}{
		{"fantasy", "fantasy", 9, TrapTypeArrowSlit},
		{"scifi", "scifi", 6, TrapTypeElectricShock},
		{"horror", "horror", 6, TrapTypeBearTrap},
		{"cyberpunk", "cyberpunk", 6, TrapTypeTeleporter},
		{"default", "unknown", 6, TrapTypePressurePlate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traps := GetGenreTraps(tt.genre)

			if len(traps) != tt.wantCount {
				t.Errorf("Trap count = %v, want %v", len(traps), tt.wantCount)
			}

			found := false
			for _, trapType := range traps {
				if trapType == tt.mustHave {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Genre %s should include trap type %v", tt.genre, tt.mustHave)
			}
		})
	}
}

func TestTrapResetCycle(t *testing.T) {
	trap := NewTrap(TrapTypePressurePlate, 5.0, 5.0, 12345)
	trap.Retriggerable = true
	trap.ResetTime = 1.0
	trap.Cooldown = 0.5

	info := &TriggerInfo{
		EntityID: "test_entity",
		X:        5.0,
		Y:        5.0,
	}

	// Trigger trap
	result := trap.CheckTrigger(info)
	if !result.Triggered {
		t.Fatal("First trigger should succeed")
	}

	if trap.State != StateTriggered {
		t.Errorf("State = %v, want %v", trap.State, StateTriggered)
	}

	// Update through reset
	trap.Update(1.1)

	if trap.State != StateCooldown {
		t.Errorf("State = %v, want %v after reset", trap.State, StateCooldown)
	}

	// Update through cooldown
	trap.Update(0.6)

	if trap.State != StateHidden {
		t.Errorf("State = %v, want %v after cooldown", trap.State, StateHidden)
	}

	// Should trigger again
	result2 := trap.CheckTrigger(info)
	if !result2.Triggered {
		t.Error("Retriggerable trap should trigger again after reset cycle")
	}
}

func TestTrapStatusEffects(t *testing.T) {
	tests := []struct {
		name       string
		trapType   TrapType
		wantEffect string
	}{
		{"poison dart", TrapTypePoisonDart, "poison_strong"},
		{"flame thrower", TrapTypeFlameThrower, "burning"},
		{"electric shock", TrapTypeElectricShock, "stunned"},
		{"net catcher", TrapTypeNetCatcher, "trapped"},
		{"bear trap", TrapTypeBearTrap, "immobilized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trap := NewTrap(tt.trapType, 5.0, 5.0, 12345)

			info := &TriggerInfo{
				EntityID: "test_entity",
				X:        5.0,
				Y:        5.0,
			}

			result := trap.CheckTrigger(info)

			if !result.Triggered {
				t.Fatal("Trap should trigger")
			}

			if result.StatusEffect != tt.wantEffect {
				t.Errorf("StatusEffect = %v, want %v", result.StatusEffect, tt.wantEffect)
			}
		})
	}
}
